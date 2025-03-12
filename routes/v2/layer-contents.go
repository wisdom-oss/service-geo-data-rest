package v2

import (
	"net/http"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"

	"microservice/internal/db"
	"microservice/types"
)

func AttributedLayerContents(c *gin.Context) {
	layerIface, _ := c.Get("layer")
	layer, _ := layerIface.(types.Layer)

	query, err := layer.ContentQuery()
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	var layerContents []types.Object
	err = pgxscan.Select(c, db.Pool, &layerContents, query)
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, types.AttributedContents{
		Attribution: layer.Attribution,
		Contents:    layerContents,
	})
}
