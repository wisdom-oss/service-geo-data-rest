package middlewares

import (
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"microservice/internal/db"
	apiErrors "microservice/internal/errors"
	"microservice/types"
)

func ResolveLayer(c *gin.Context) {
	layerID := c.Param("layerID")

	query, err := db.Queries.Raw("get-layer")
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	if err = uuid.Validate(layerID); err != nil {
		query, err = db.Queries.Raw("get-layer-by-url-key")
		if err != nil {
			c.Abort()
			_ = c.Error(err)
			return
		}
	}

	var layer types.Layer
	err = pgxscan.Get(c, db.Pool, &layer, query, layerID)
	if err != nil {
		c.Abort()
		if pgxscan.NotFound(err) {
			apiErrors.ErrUnknownLayer.Emit(c)
			return
		}
		_ = c.Error(err)
		return
	}

	c.Set("layer", layer)
}
