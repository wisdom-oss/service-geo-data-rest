package v2

import (
	"github.com/jackc/pgx/v5/pgtype"

	"microservice/types"
)

type Layer struct {
	types.Layer    `db:""`
	AttributionURL pgtype.Text `db:"attribution_url" json:"attributionURL"`
}
