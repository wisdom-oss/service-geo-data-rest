package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/gin-contrib/gzip"
	"github.com/rs/zerolog/log"

	healthcheckServer "github.com/wisdom-oss/go-healthcheck/server"

	"microservice/internal"
	"microservice/internal/config"
	"microservice/internal/db"
	"microservice/middlewares"
	v1Routes "microservice/routes/v1"
	v2Routes "microservice/routes/v2"
)

// the main function bootstraps the http server and handlers used for this
// microservice.
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

	r := config.PrepareRouter()
	r.Use(gzip.Gzip(gzip.BestCompression))
	r.Use(middlewares.EnablePrivateLayers)

	v1 := r.Group("/v1")
	{
		v1.GET("/", v1Routes.LayerOverview)
		v1.GET("/:layerID", middlewares.ResolveLayer, v1Routes.LayerInformation)
		v1.GET("/identify", v1Routes.IdentifyObject)

		content := v1.Group("/content", middlewares.ResolveLayer)
		{
			content.GET("/:layerID", v1Routes.LayerContents)
			content.GET("/:layerID/filtered", v1Routes.FilteredLayerContents)
		}
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/", v2Routes.LayerOverview)

		content := v2.Group("/content/:layerID", middlewares.ResolveV2Layer)
		{
			content.GET("/", v2Routes.AttributedLayerContents)
		}
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
