package middlewares_test

import (
	"testing"

	"github.com/gin-gonic/gin"

	"microservice/middlewares"
	"microservice/types"
)

func Test_LayerResolver(t *testing.T) {
	router := gin.New()
	router.Use(middlewares.ResolveLayer)

	router.Any("/", func(c *gin.Context) {
		layerInterface, isSet := c.Get("layer")
		if !isSet {
			t.Fail()
			t.Logf("Layer not set in context")
			return
		}

		_, isLayer := layerInterface.(types.Layer)
		if !isLayer {
			t.Fail()
			t.Logf("Layer is not a layer")
			return
		}
	})
}
