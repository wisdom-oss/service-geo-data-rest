package routes

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/georgysavva/scany/v2/pgxscan"

	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"

	"microservice/globals"
	"microservice/types"
)

// LayerInformation
func LayerInformation(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	query, err := globals.SqlQueries.Raw("get-layers")
	if err != nil {
		errorHandler <- err
		return
	}

	var layers []types.Layer
	err = pgxscan.Select(context.Background(), globals.Db, &layers, query)
	if err != nil {
		errorHandler <- err
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(layers)
}
