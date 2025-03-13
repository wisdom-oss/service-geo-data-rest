package v2Routes

import (
	"net/http"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"

	"microservice/internal/db"
	"microservice/types"
	v2 "microservice/types/v2"
)

func AttributedLayerContents(c *gin.Context) {
	layerIface, _ := c.Get("layer")
	layer, _ := layerIface.(v2.Layer)

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

	c.JSON(http.StatusOK, v2.AttributedContents{
		Attribution:    layer.Attribution,
		AttributionURL: layer.AttributionURL,
		Contents:       layerContents,
	})
}
