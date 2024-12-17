//go:build docker

// This file contains a default configuration for local development.
// It automatically disables the requirement for authenticated calls and sets
// the default logging level to Debug to make debugging easier.
// Furthermore, the default listen address is set to localhost only as the
// authentication has been turned off to minimize the risk of data leaks

package config

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	errorHandler "github.com/wisdom-oss/common-go/v3/middleware/gin/error-handler"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/jwt"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/recoverer"

	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/requestid"
)

const ListenAddress = "0.0.0.0:8000"

func init() {
	// set gin to the production mode (aka release mode)
	gin.SetMode(gin.ReleaseMode)
}

// Middlewares configures and outputs the middlewares used in the configuration.
// The contained middlewares are the following:
//   - gin.Logger
func Middlewares() []gin.HandlerFunc {
	var middlewares []gin.HandlerFunc

	middlewares = append(middlewares,
		logger.SetLogger(
			logger.WithDefaultLevel(zerolog.DebugLevel),
			logger.WithUTC(false),
		))

	middlewares = append(middlewares, requestid.New())
	middlewares = append(middlewares, errorHandler.Handler)
	middlewares = append(middlewares, gin.CustomRecovery(recoverer.RecoveryHandler))

	// read the OIDC authority from the environment
	oidcAuthority, isSet := os.LookupEnv("OIDC_AUTHORITY")
	if !isSet {
		oidcAuthority = "http://backend/api/auth/"
	}

	validator := jwt.Validator{}
	err := validator.Discover(oidcAuthority)
	if err != nil {
		panic(err)
	}

	middlewares = append(middlewares, validator.Handler)

	return middlewares
}

func PrepareRouter() *gin.Engine {
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.ForwardedByClientIP = true
	_ = router.SetTrustedProxies([]string{"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"})
	router.Use(Middlewares()...)

	router.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, MethodNotAllowed)
	})
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, NotFound)

	})

	return router
}
