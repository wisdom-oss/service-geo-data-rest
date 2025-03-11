package routes

import (
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"microservice/internal/db"
	apiErrors "microservice/internal/errors"
	"microservice/types"
)

func FilteredLayerContents(c *gin.Context) {
	var parameters struct {
		Relation   string   `binding:"required" form:"relation"    json:"relation"`
		Keys       []string `binding:"required" form:"key"         json:"key"`
		OtherLayer string   `binding:"required" form:"other_layer" json:"other_layer"`
	}

	if err := c.ShouldBind(&parameters); err != nil {
		c.Abort()
		res := apiErrors.ErrMissingParameter
		res.Errors = []error{err}
		res.Emit(c)
		return
	}

	query, err := db.Queries.Raw("get-layer")
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	if err = uuid.Validate(parameters.OtherLayer); err != nil {
		query, err = db.Queries.Raw("get-layer-by-url-key")
		if err != nil {
			c.Abort()
			_ = c.Error(err)
			return
		}
	}

	var topLayer types.Layer
	err = pgxscan.Get(c, db.Pool, &topLayer, query, parameters.OtherLayer)
	if err != nil {
		c.Abort()
		if pgxscan.NotFound(err) {
			apiErrors.ErrUnknownTopLayer.Emit(c)
			return
		}
		_ = c.Error(err)
		return
	}

	var objects []types.Object
	switch parameters.Relation {
	case "within":
		objects = filteredLayerContents_Within(c, topLayer, parameters.Keys)
	case "overlaps":
		objects = filteredLayerContents_Overlaps(c, topLayer, parameters.Keys)
	case "contains":
		objects = filteredLayerContents_Contains(c, topLayer, parameters.Keys)
	default:
		c.Abort()
		apiErrors.ErrUnsupportedSpatialRelation.Emit(c)
		return
	}

	if len(objects) == 0 {
		c.Status(204)
		return
	}

	c.JSON(200, objects)
}

func filteredLayerContents_Within(c *gin.Context, topLayer types.Layer, keys []string) []types.Object {
	layerInterface, _ := c.Get("layer")
	baseLayer, _ := layerInterface.(types.Layer)

	var queryParts []string
	var queryParams []interface{}
	for idx, key := range keys {
		queryParts = append(queryParts,
			fmt.Sprintf(`ST_WITHIN(st_transform(geometry, 4326), (SELECT st_transform(geometry, 4326) FROM geodata.%s WHERE key = $%d))`,
				topLayer.TableName, idx+1))
		queryParams = append(queryParams, key)
	}

	queryCondition := strings.Join(queryParts, " OR ")

	baseQuery, _ := baseLayer.ContentQuery()
	query := fmt.Sprintf("%s WHERE %s;", strings.TrimSuffix(baseQuery, ";"), queryCondition)
	var layerContents []types.Object
	err := pgxscan.Select(c, db.Pool, &layerContents, query, queryParams...)
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return nil
	}

	return layerContents
}

func filteredLayerContents_Overlaps(c *gin.Context, topLayer types.Layer, keys []string) []types.Object {
	layerInterface, _ := c.Get("layer")
	baseLayer, _ := layerInterface.(types.Layer)

	var queryParts []string
	var queryParams []interface{}
	for idx, key := range keys {
		queryParts = append(queryParts,
			fmt.Sprintf(`ST_OVERLAPS(st_transform(geometry, 4326), (SELECT st_transform(geometry, 4326) FROM geodata.%s WHERE key = $%d))`,
				topLayer.TableName, idx+1))
		queryParams = append(queryParams, key)
	}

	queryCondition := strings.Join(queryParts, " OR ")

	baseQuery, _ := baseLayer.ContentQuery()
	query := fmt.Sprintf("%s WHERE %s;", strings.TrimSuffix(baseQuery, ";"), queryCondition)
	var layerContents []types.Object
	err := pgxscan.Select(c, db.Pool, &layerContents, query, queryParams...)
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return nil
	}

	return layerContents
}

func filteredLayerContents_Contains(c *gin.Context, topLayer types.Layer, keys []string) []types.Object {
	layerInterface, _ := c.Get("layer")
	baseLayer, _ := layerInterface.(types.Layer)

	var queryParts []string
	var queryParams []interface{}
	for idx, key := range keys {
		queryParts = append(queryParts,
			fmt.Sprintf(`ST_CONTAINS(st_transform(geometry, 4326), (SELECT st_transform(geometry, 4326) FROM geodata.%s WHERE key = $%d))`,
				topLayer.TableName, idx+1))
		queryParams = append(queryParams, key)
	}

	queryCondition := strings.Join(queryParts, " OR ")

	baseQuery, _ := baseLayer.ContentQuery()
	query := fmt.Sprintf("%s WHERE %s;", strings.TrimSuffix(baseQuery, ";"), queryCondition)
	var layerContents []types.Object
	err := pgxscan.Select(c, db.Pool, &layerContents, query, queryParams...)
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return nil
	}
	return layerContents
}
