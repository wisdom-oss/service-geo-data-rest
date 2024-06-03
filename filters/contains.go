package filters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"

	"microservice/globals"
	"microservice/types"
)

type Contains struct{}

func (f Contains) BuildQueryPart(layerID string, keys ...string) (string, error) {
	query, err := globals.SqlQueries.Raw("get-layer")
	if err != nil {
		return "", nil
	}

	var layer types.Layer
	err = pgxscan.Get(context.Background(), globals.Db, &layer, query, layerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUnknownLayerID
		}
		return "", err
	}

	var parts []string
	for _, key := range keys {
		parts = append(parts,
			fmt.Sprintf(`ST_CONTAINS(geometry, (SELECT geometry FROM geodata.%s WHERE key = '%s'))`, layer.TableName.String, key),
		)
	}
	return strings.Join(parts, " OR "), nil
}