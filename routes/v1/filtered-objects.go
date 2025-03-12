package routes

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"microservice/internal/db"
	apiErrors "microservice/internal/errors"
	"microservice/types"
)

const queryString = `st_%s(st_transform(geometry, 4326), (SELECT st_transform(geometry, 4236) FROM geodata.%s WHERE key = $%d))`

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
	case "within", "overlaps", "contains":
		bottomLayerIface, _ := c.Get("layer")
		bottomLayer, _ := bottomLayerIface.(types.Layer)

		queryParts := make([]string, len(c.Keys))
		queryParameters := make([]string, len(c.Keys))

		for idx, key := range parameters.Keys {
			queryParts[idx] = fmt.Sprintf(queryString, parameters.Relation, bottomLayer.TableName, idx+1)
			queryParameters[idx] = key
		}

		queryClause := strings.Join(queryParts, ` OR `)

		baseLayerQuery, err := bottomLayer.ContentQuery()
		if err != nil {
			c.Abort()
			_ = c.Error(err)
			return
		}

		query := fmt.Sprintf(`%s WHERE %s;`, strings.TrimSuffix(baseLayerQuery, ";"), queryClause)
		err = pgxscan.Select(c, db.Pool, &objects, query, queryParameters)
		if err != nil {
			c.Abort()
			_ = c.Error(err)
			return
		}

	default:
		c.Abort()
		apiErrors.ErrUnsupportedSpatialRelation.Emit(c)
		return
	}

	if len(objects) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, objects)
}
