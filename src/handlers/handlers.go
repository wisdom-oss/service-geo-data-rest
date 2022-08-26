package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/lib/pq"
	geojson "github.com/paulmach/go.geojson"
	log "github.com/sirupsen/logrus"
	e "microservice/errors"
	"microservice/helpers"
	"microservice/structs"
	"microservice/vars"
)

func AuthorizationCheck(nextHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.WithFields(log.Fields{
			"middleware": true,
			"title":      "AuthorizationCheck",
		})
		logger.Debug("Checking the incoming request for authorization information set by the gateway")

		// Get the scopes the requesting user has
		scopes := r.Header.Get("X-Authenticated-Scope")
		// Check if the string is empty
		if strings.TrimSpace(scopes) == "" {
			logger.Warning("Unauthorized request detected. The required header had no content or was not set")
			requestError := e.NewRequestError(e.UnauthorizedRequest)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}

		scopeList := strings.Split(scopes, ",")
		if !helpers.StringArrayContains(scopeList, vars.Scope.ScopeValue) {
			logger.Error("Request rejected. The user is missing the scope needed for accessing this service")
			requestError := e.NewRequestError(e.MissingScope)
			w.Header().Set("Content-Type", "text/json")
			w.WriteHeader(requestError.HttpStatus)
			encodingError := json.NewEncoder(w).Encode(requestError)
			if encodingError != nil {
				logger.WithError(encodingError).Error("Unable to encode request error response")
			}
			return
		}
		// Call the next handler which will continue handling the request
		nextHandler.ServeHTTP(w, r)
	})
}

/*
PingHandler

This handler is used to test if the service is able to ping itself. This is done to run a healthcheck on the container
*/
func PingHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

/*
MapDataHandler

This handler shows how a basic handler works and how to send back a message
*/
func MapDataHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.WithFields(log.Fields{
		"middleware": false,
		"title":      "MapDataHandler",
		"origin":     r.RemoteAddr,
	})
	// Set the response content type
	w.Header().Set("Content-Type", "text/json")
	logger.Info("New incoming request")
	// Check what query parameters have been set
	shapeKeysSet := r.URL.Query().Has("key")
	resolutionSet := r.URL.Query().Has("resolution")
	var shapeRows *sql.Rows
	var queryError error
	var boxRow *sql.Row
	switch {
	case !shapeKeysSet && !resolutionSet:
		logger.Warning("Request has no filter options. The service may become unresponsive")
		shapeQuery := `SELECT name, key, nuts_key, ST_ASGeoJSON(geom) FROM geodata.shapes`
		boxQuery := `SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom))) FROM geodata.shapes`
		shapeRows, queryError = vars.PostgresConnection.Query(shapeQuery)
		boxRow = vars.PostgresConnection.QueryRow(boxQuery)
		break
	case !shapeKeysSet && resolutionSet:
		resolution := r.URL.Query().Get("resolution")
		shapeQuery := `SELECT name, key, nuts_key, ST_ASGeoJSON(geom) FROM geodata.shapes
					   WHERE length(key) = $1`
		boxQuery := `SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom))) FROM geodata.shapes
                     WHERE length(key) = $1`
		shapeRows, queryError = vars.PostgresConnection.Query(shapeQuery, getShapeKeyLength(resolution))
		boxRow = vars.PostgresConnection.QueryRow(boxQuery, getShapeKeyLength(resolution))
		break
	case shapeKeysSet && !resolutionSet:
		// Get the shape keys
		shapeKeys := r.URL.Query()["key"]
		shapeQuery := `SELECT name, key, nuts_key, ST_ASGeoJSON(geom) FROM geodata.shapes
					   WHERE key = any($1)`
		boxQuery := `SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom))) FROM geodata.shapes
					 WHERE key = any($1)`
		shapeRows, queryError = vars.PostgresConnection.Query(shapeQuery, pq.Array(shapeKeys))
		boxRow = vars.PostgresConnection.QueryRow(boxQuery, pq.Array(shapeKeys))
		break
	case shapeKeysSet && resolutionSet:
		// Get both parameters
		shapeKeys := r.URL.Query()["key"]
		resolution := r.URL.Query().Get("resolution")
		var regexString string
		for _, shapeKey := range shapeKeys {
			if len(shapeKey) < getShapeKeyLength(resolution) {
				regexString += fmt.Sprintf(`^%s\d+$|`, shapeKey)
			} else {
				regexString += fmt.Sprintf(`^%s$|`, shapeKey)
			}
		}
		regexString = strings.Trim(regexString, "|")
		shapeQuery := `SELECT name, key, nuts_key, ST_ASGeoJSON(geom) FROM geodata.shapes
					   WHERE length(key) = $1
					   AND key ~ $2`
		boxQuery := `SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom))) FROM geodata.shapes
					 WHERE length(key) = $1
					 AND key ~ $2`
		shapeRows, queryError = vars.PostgresConnection.Query(shapeQuery, getShapeKeyLength(resolution), regexString)
		boxRow = vars.PostgresConnection.QueryRow(boxQuery, getShapeKeyLength(resolution), regexString)
		break
	default:
		logger.Warning("Request has no detectable filter options. The service may become unresponsive")
		shapeQuery := `SELECT name, key, nuts_key, ST_ASGeoJSON(geom) FROM geodata.shapes`
		shapeRows, queryError = vars.PostgresConnection.Query(shapeQuery)
		boxQuery := `SELECT ST_ASGeoJson(ST_FlipCoordinates(ST_Extent(geom))) FROM geodata.shapes`
		boxRow = vars.PostgresConnection.QueryRow(boxQuery)
		break
	}
	if queryError != nil {
		logger.WithError(queryError).Error("An error occurred while executing the query in the database")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	// Build the shape data
	var shapes []structs.ShapeData
	for shapeRows.Next() {
		var Name, ShapeKey, NutsKey string
		var GeoJSON geojson.Geometry

		scanError := shapeRows.Scan(&Name, &ShapeKey, &NutsKey, &GeoJSON)
		if scanError != nil {
			logger.WithError(scanError).Error("An error occurred while scanning the result rows")
			helpers.SendRequestError(e.InternalServiceError, w)
			return
		}

		shapes = append(shapes, structs.ShapeData{
			Name:    Name,
			Key:     ShapeKey,
			NutsKey: NutsKey,
			GeoJSON: GeoJSON,
		})
	}
	// Get the bounding box data
	boundingBoxQueryError := boxRow.Err()
	if boundingBoxQueryError != nil {
		logger.WithError(queryError).Error("An error occurred while executing the query in the database")
		helpers.SendRequestError(e.DatabaseQueryError, w)
		return
	}
	var boundingBox geojson.Geometry
	scanError := boxRow.Scan(&boundingBox)
	if scanError != nil {
		logger.WithError(scanError).Error("An error occurred while scanning the bounding box row")
		helpers.SendRequestError(e.InternalServiceError, w)
		return
	}

	type responseData struct {
		Box    [][]float64         `json:"box"`
		Shapes []structs.ShapeData `json:"shapes"`
	}
	response := responseData{
		Box:    boundingBox.Polygon[0],
		Shapes: shapes,
	}

	encodingError := json.NewEncoder(w).Encode(&response)
	if encodingError != nil {
		logger.WithError(scanError).Error("An error occurred while returning the response")
		helpers.SendRequestError(e.InternalServiceError, w)
		return
	}

}
