package routes

import (
	"net/http"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"

	"microservice/internal/db"
	apiErrors "microservice/internal/errors"
	"microservice/types"
)

var _identifyObjectRequest struct {
	Keys []string `form:"key"`
}

func IdentifyObject(c *gin.Context) {

	objects := make(map[string]map[string]types.Object)

	if err := c.ShouldBind(&_identifyObjectRequest); err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	query, err := db.Queries.Raw("get-layers")
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	var layers []types.Layer
	err = pgxscan.Select(c, db.Pool, &layers, query)
	if err != nil {
		c.Abort()
		_ = c.Error(err)
		return
	}

	for _, layer := range layers {
		for _, key := range _identifyObjectRequest.Keys {
			query, err = layer.FilteredContentQuery()
			if err != nil {
				_ = c.Error(err)
				continue
			}
			var object types.Object
			err = pgxscan.Get(c, db.Pool, &object, query, key)
			if err != nil {
				if pgxscan.NotFound(err) {
					continue
				}
				c.Abort()
				_ = c.Error(err)
				continue
			}
			if objects[layer.TableName] == nil {
				objects[layer.TableName] = make(map[string]types.Object)
			}
			objects[layer.TableName][key] = object
		}
	}

	if c.IsAborted() {
		panic("error in layer identification")
	}
	if len(objects) == 0 {
		apiErrors.ErrUnknownObject.Emit(c)
		return
	}

	c.JSON(http.StatusOK, objects)
}
