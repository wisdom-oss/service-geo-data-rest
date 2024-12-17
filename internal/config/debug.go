//go:build !docker

// This file contains a default configuration for local development.
// It automatically disables the requirement for authenticated calls and sets
// the default logging level to Debug to make debugging easier.
// Furthermore, the default listen address is set to localhost only as the
// authentication has been turned off to minimize the risk of data leaks

package config

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/gin-contrib/logger"
	"github.com/gin-contrib/requestid"

	errorHandler "github.com/wisdom-oss/common-go/v3/middleware/gin/error-handler"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/jwt"
	"github.com/wisdom-oss/common-go/v3/middleware/gin/recoverer"
)

const ListenAddress = "127.0.0.1:8000"

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

	// this middleware allows all access during the local debugging and
	// development
	middlewares = append(middlewares, func(ctx *gin.Context) {
		ctx.Set(jwt.KeyAdministrator, true)
	})
	return middlewares
}

func PrepareRouter() *gin.Engine {
	router := gin.New()
	router.HandleMethodNotAllowed = true
	router.Use(Middlewares()...)

	router.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusMethodNotAllowed, MethodNotAllowed)
	})
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, NotFound)

	})

	return router
}
