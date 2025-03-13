package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"microservice/types"
)

func LayerInformation(c *gin.Context) {
	layerInterface, _ := c.Get("layer")
	layer, _ := layerInterface.(types.Layer)

	c.JSON(http.StatusOK, layer)
}
