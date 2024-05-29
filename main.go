package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	healthcheckServer "github.com/wisdom-oss/go-healthcheck/server"
	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"
	securityMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/security"

	"microservice/config"
	"microservice/routes"

	"microservice/globals"
)

// the main function bootstraps the http server and handlers used for this
// microservice
func main() {
	// create a new logger for the main function
	l := log.With().Str("step", "main").Logger()
	l.Info().Msgf("starting %s service", globals.ServiceName)

	// create the healthcheck server
	hcServer := healthcheckServer.HealthcheckServer{}
	hcServer.InitWithFunc(func() error {
		// test if the database is reachable
		return globals.Db.Ping(context.Background())
	})
	err := hcServer.Start()
	if err != nil {
		l.Fatal().Err(err).Msg("unable to start healthcheck server")
	}
	go hcServer.Run()

	// create a new router
	router := chi.NewRouter()
	router.Use(config.Middlewares...)
	router.NotFound(errorMiddleware.NotFoundError)
	// now mount the routes as some examples
	router.
		With(securityMiddleware.RequireScope(globals.ServiceName, securityMiddleware.ScopeRead)).
		HandleFunc("/", routes.LayerInformation)
	router.
		With(securityMiddleware.RequireScope(globals.ServiceName, securityMiddleware.ScopeRead)).
		HandleFunc(fmt.Sprintf("/content/{%s}", routes.LayerIdUrlKey), routes.LayerContents)

	// now boot up the service
	// Configure the HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%s", globals.Environment["LISTEN_PORT"]),
		WriteTimeout: time.Second * 600,
		ReadTimeout:  time.Second * 600,
		IdleTimeout:  time.Second * 600,
		Handler:      router,
	}

	// Start the server and log errors that happen while running it
	go func() {
		if err := server.ListenAndServe(); err != nil {
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