package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/chi/v5"
	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"

	"microservice/globals"
	"microservice/types"
)

func SingleLayerInformation(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	layerID := chi.URLParam(r, LayerIdUrlKey)
	if layerID == "" {
		errorHandler <- ErrEmptyLayerID
		return
	}

	query, err := globals.SqlQueries.Raw("get-layer")
	if err != nil {
		errorHandler <- err
		return
	}

	var layer types.Layer
	err = pgxscan.Get(context.Background(), globals.Db, &layer, query, layerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorHandler <- ErrUnknownLayerID
			return
		}
		errorHandler <- err
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(layer)
}
