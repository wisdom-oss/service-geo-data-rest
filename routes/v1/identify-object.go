package routes

import (
	"net/http"
	"sync"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gin-gonic/gin"

	"microservice/internal/db"
	apiErrors "microservice/internal/errors"
	"microservice/types"
)

func IdentifyObject(c *gin.Context) {
	var parameters struct {
		Keys []string `binding:"required" form:"key" json:"keys"`
	}

	if err := c.ShouldBind(&parameters); err != nil {
		c.Abort()
		apiErrors.ErrMissingParameter.Emit(c)
		return
	}

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

	objects := make(map[string]map[string]types.Object)
	var wg sync.WaitGroup
	var mapLock sync.Mutex
	for _, k := range parameters.Keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			for _, l := range layers {
				query, err = l.FilteredContentQuery()
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
				mapLock.Lock()
				if objects[l.TableName] == nil {
					objects[l.TableName] = make(map[string]types.Object)
				}
				objects[l.TableName][key] = object
				mapLock.Unlock()
			}
		}(k)

	}
	wg.Wait()
	if c.IsAborted() {
		return
	}
	if len(objects) == 0 {
		apiErrors.ErrUnknownObject.Emit(c)
		return
	}

	c.JSON(http.StatusOK, objects)
}
