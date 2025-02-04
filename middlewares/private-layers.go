package middlewares

import (
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/jwt"

	"microservice/internal"
)

// EnablePrivateLayers checks if an access token has been validated and has
// the required permissions to access the private layers of the service.
func EnablePrivateLayers(c *gin.Context) {
	tokenValidated := c.GetBool(jwt.KeyTokenValidated)
	isAdmin := c.GetBool(jwt.KeyAdministrator)
	permissions := c.GetStringSlice(jwt.KeyTokenPermissions)

	c.Set("AccessPrivateLayers", (isAdmin || tokenValidated && slices.Contains(permissions, internal.ServiceName+":read")))
	c.Next()
}
