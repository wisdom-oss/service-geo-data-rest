package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rs/zerolog/log"
	healthcheckServer "github.com/wisdom-oss/go-healthcheck/server"

	"microservice/internal"
	"microservice/internal/config"
	"microservice/internal/db"
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

	r := config.PrepareRouter()
	r.Use(middlewares.EnablePrivateLayers)

	r.GET("/", routes.LayerOverview)
	r.GET("/:layerID", middlewares.ResolveLayer, routes.LayerInformation)
	r.GET("/identify", routes.IdentifyObject)

	content := r.Group("/content", middlewares.ResolveLayer)
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
