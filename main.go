package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/wisdom-oss/common-go/v2/middleware"
	"github.com/wisdom-oss/common-go/v2/types"
	healthcheckServer "github.com/wisdom-oss/go-healthcheck/server"

	"microservice/internal"
	"microservice/internal/config"
	"microservice/internal/db"
	"microservice/internal/errors"
	"microservice/middlewares"
	"microservice/routes"
)

// the main function bootstraps the http server and handlers used for this
// microservice
func main() {
	// create a new logger for the main function
	l := log.Logger
	l.Info().Msgf("configuring %s service", internal.ServiceName)

	// create the healthcheck server
	hcServer := healthcheckServer.HealthcheckServer{}
	hcServer.InitWithFunc(func() error {
		// test if the database is reachable
		return db.Pool.Ping(context.Background())
	})
	err := hcServer.Start()
	if err != nil {
		l.Fatal().Err(err).Msg("unable to start healthcheck server")
	}
	go hcServer.Run()

	// shorthand for requiring scope
	requireScope := middleware.RequireScope{}.Gin

	r := gin.New()
	r.HandleMethodNotAllowed = true
	_ = r.SetTrustedProxies(nil)
	r.Use(config.Middlewares()...)
	r.NoMethod(func(c *gin.Context) {
		errors.MethodNotAllowed.Emit(c)
	})
	r.NoRoute(func(c *gin.Context) {
		errors.NotFound.Emit(c)
	})

	r.GET("/", routes.LayerOverview)
	r.GET("/:layerID", middlewares.ResolveLayer, routes.LayerInformation)
	r.GET("/identify", routes.IdentifyObject)

	content := r.Group("/content", requireScope(internal.ServiceName, types.ScopeRead), middlewares.ResolveLayer)
	{
		content.GET("/:layerID", routes.LayerContents)
		content.GET("/:layerID/filtered", routes.FilteredLayerContents)
	}

	l.Info().Msg("finished service configuration")
	l.Info().Msg("starting http server")

	// Start the server and log errors that happen while running it
	go func() {
		if err := r.Run(config.ListenAddress); err != nil {
			l.Fatal().Err(err).Msg("An error occurred while starting the http server")
		}
	}()

	// Set up the signal handling to allow the server to shut down gracefully

	cancelSignal := make(chan os.Signal, 1)
	signal.Notify(cancelSignal, os.Interrupt)

	// Block further code execution until the shutdown signal was received
	l.Info().Msg("server ready to accept connections")
	<-cancelSignal

}
