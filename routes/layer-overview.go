package routes

import (
	"net/http"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"

	"microservice/internal/db"
	"microservice/types"
)

func LayerOverview(c *gin.Context) {
	query, err := db.Queries.Raw("get-layers")
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	var layers []types.Layer
	err = pgxscan.Select(c, db.Pool, &layers, query, c.GetBool("AccessPrivateLayers"))
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, layers)
}
