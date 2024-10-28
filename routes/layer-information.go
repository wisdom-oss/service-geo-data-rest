package routes

import (
	"github.com/gin-gonic/gin"

	"microservice/types"
)

func LayerInformation(c *gin.Context) {
	layerInterface, _ := c.Get("layer")
	layer, _ := layerInterface.(types.Layer)

	c.JSON(200, layer)
}
