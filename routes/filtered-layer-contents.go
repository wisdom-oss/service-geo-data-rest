package routes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"

	"microservice/filters"
	"microservice/globals"
	"microservice/types"
)

var filterMapping = map[types.FilterType]filters.Filter{
	types.FilterOverlaps: filters.Overlaps{},
	types.FilterContains: filters.Contains{},
	types.FilterWithin:   filters.Within{},
}

// FilteredLayerContents handles requests which use geospatial relations
// between layers to exclude or include geometries from the response
func FilteredLayerContents(w http.ResponseWriter, r *http.Request) {
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
		if errors.Is(err, pgx.ErrNoRows) {
			errorHandler <- ErrUnknownLayerID
			return
		}
		errorHandler <- err
		return
	}

	var filterConfiguration map[types.FilterType]map[string][]string
	err = json.NewDecoder(r.Body).Decode(&filterConfiguration)
	if err != nil {
		if err != io.EOF {
			errorHandler <- err
			return
		}
	}

	var queryParts []string
	for filterType, layerInformation := range filterConfiguration {
		for layerID, keys := range layerInformation {
			parts, err := filterMapping[filterType].BuildQueryPart(layerID, keys...)
			if err != nil {
				errorHandler <- err
				return
			}
			queryParts = append(queryParts, parts)

		}
	}

	query, err = globals.SqlQueries.Raw("get-layer-contents")
	if err != nil {
		errorHandler <- err
		return
	}
	query = fmt.Sprintf(query, layer.TableName.String)
	query = strings.ReplaceAll(query, `;`, ``)
	query += " WHERE ("
	query += strings.Join(queryParts, ") AND (")
	query += ")"

	var objects []types.Object
	err = pgxscan.Select(context.Background(), globals.Db, &objects, query)
	if err != nil {
		errorHandler <- err
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(objects)

}
