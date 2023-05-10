package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	geojson "github.com/paulmach/go.geojson"
	"microservice/enums"
	requestErrors "microservice/request/error"
	"microservice/structs"
	"microservice/vars/globals"
	"microservice/vars/globals/connections"
	"net/http"
	"strings"
)

var l = globals.HttpLogger

func GetShapes(w http.ResponseWriter, r *http.Request) {
	l.Info().Msg("new request for geo shapes")

	// check the query parameters and store the values
	var shapeKeysSet, resolutionSet bool
	var shapeKeys []string
	var resolution enums.Resolution

	if shapeKeysSet = r.URL.Query().Has("key"); shapeKeysSet {
		shapeKeys = r.URL.Query()["key"]
	}

	if resolutionSet = r.URL.Query().Has("resolution"); resolutionSet {
		resolution = enums.Resolution(r.URL.Query().Get("resolution"))
	}

	// now select the needed query based on the query parameters and execute it
	var shapeRows *sql.Rows
	var queryError error
	var shapeBoxRow *sql.Row

	switch {
	case !shapeKeysSet && !resolutionSet:
		l.Warn().Msg("no query parameters provided. query may take a long time to execute")
		shapeRows, queryError = globals.Queries.Query(connections.DbConnection, "get-all-shapes")
		shapeBoxRow, _ = globals.Queries.QueryRow(connections.DbConnection, "get-box-for-all-shapes")
		break
	case shapeKeysSet && !resolutionSet:
		shapeRows, queryError = globals.Queries.Query(connections.DbConnection, "get-shapes-by-key", pq.Array(shapeKeys))
		shapeBoxRow, _ = globals.Queries.QueryRow(connections.DbConnection, "get-box-for-shapes-by-key", pq.Array(shapeKeys))
		break
	case !shapeKeysSet && resolutionSet:
		shapeRows, queryError = globals.Queries.Query(connections.DbConnection, "get-shapes-by-resolution", resolution.GetKeyLength())
		shapeBoxRow, _ = globals.Queries.QueryRow(connections.DbConnection, "get-box-for-shapes-by-resolution", resolution.GetKeyLength())
		break
	case shapeKeysSet && resolutionSet:
		// create a regular expression to match the keys to their internal ids
		var regex string
		for _, key := range shapeKeys {
			if len(key) < resolution.GetKeyLength() {
				regex += fmt.Sprintf(`^%s\d+$|`, key)
			} else if len(key) == resolution.GetKeyLength() {
				regex += fmt.Sprintf(`^%s$|`, key)
			} else {
				l.Warn().Str("key", key).Msg("key is longer than the resolution key length. ignoring key")
			}
		}
		// remove the last pipe character
		regex = strings.Trim(regex, "|")
		// now query the database
		shapeRows, queryError = globals.Queries.Query(connections.DbConnection, "get-shapes-by-key-resolution", resolution.GetKeyLength(), regex)
		shapeBoxRow, _ = globals.Queries.QueryRow(connections.DbConnection, "get-box-for-shapes-by-key-resolution", resolution.GetKeyLength(), regex)
		break
	default:
		l.Warn().Msg("something went wrong with the query parameters. returning all shapes")
		l.Warn().Msg("no query parameters provided. query may take a long time to execute")
		shapeRows, queryError = globals.Queries.Query(connections.DbConnection, "get-all-shapes")
		shapeBoxRow, _ = globals.Queries.QueryRow(connections.DbConnection, "get-shape-box")
		break
	}

	// check if there was an error with the queries
	if queryError != nil {
		l.Error().Err(queryError).Msg("error with shape query")
		e, _ := requestErrors.WrapInternalError(queryError)
		requestErrors.SendError(e, w)
		return
	}

	if shapeBoxRow.Err() != nil {
		l.Error().Err(shapeBoxRow.Err()).Msg("error with box query")
		e, _ := requestErrors.WrapInternalError(shapeBoxRow.Err())
		requestErrors.SendError(e, w)
		return
	}

	// now build the shapes for the response
	var shapes []structs.Shape
	for shapeRows.Next() {
		var shape structs.Shape
		scanError := shapeRows.Scan(&shape.Name, &shape.Key, &shape.NutsKey, &shape.GeoJSON)
		if scanError != nil {
			l.Error().Err(scanError).Msg("error with shape scan")
			e, _ := requestErrors.WrapInternalError(scanError)
			requestErrors.SendError(e, w)
			return
		}
		shapes = append(shapes, shape)
	}

	var boundingBox geojson.Geometry
	scanError := shapeBoxRow.Scan(&boundingBox)
	if scanError != nil {
		l.Error().Err(scanError).Msg("error with bounding box scan")
		e, _ := requestErrors.WrapInternalError(scanError)
		requestErrors.SendError(e, w)
		return
	}

	// now build the response
	type response struct {
		Shapes      []structs.Shape `json:"shapes"`
		BoundingBox [][]float64     `json:"box"`
	}

	res := response{
		Shapes:      shapes,
		BoundingBox: boundingBox.Polygon[0][:4],
	}

	// send the response
	w.Header().Set("Content-Type", "application/json")
	encodingErr := json.NewEncoder(w).Encode(res)
	if encodingErr != nil {
		l.Error().Err(encodingErr).Msg("error encoding response")
		e, _ := requestErrors.WrapInternalError(encodingErr)
		requestErrors.SendError(e, w)
		return
	}

}
