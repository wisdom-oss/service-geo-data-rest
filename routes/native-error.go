package routes

import (
	"errors"
	"net/http"

	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"
)

func BasicWithErrorHandling(w http.ResponseWriter, r *http.Request) {
	// access the error handlers
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})
	// now publish an error to each of the wisdom errors
	errorHandler <- errors.New("native test error")
	// now wait for the error to be handled
	return
}
