package types

import (
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

type Object struct {
	ID                   pgtype.Int8            `json:"id" db:"id"`
	Geometry             geom.T                 `json:"geometry" db:"geometry"`
	Name                 pgtype.Text            `json:"name" db:"name"`
	Key                  pgtype.Text            `json:"key" db:"key"`
	AdditionalProperties map[string]interface{} `json:"additionalProperties" db:"additional_properties"`
}

func (o Object) MarshalJSON() ([]byte, error) {
	type obj struct {
		ID                   pgtype.Int8            `json:"-" db:"id"`
		Geometry             json.RawMessage        `json:"geometry" db:"geometry"`
		Name                 pgtype.Text            `json:"name" db:"name"`
		Key                  pgtype.Text            `json:"key" db:"key"`
		AdditionalProperties map[string]interface{} `json:"additionalProperties" db:"additional_properties"`
	}
	out := obj{
		ID:                   o.ID,
		Name:                 o.Name,
		Key:                  o.Key,
		AdditionalProperties: o.AdditionalProperties,
	}
	out.Geometry, _ = geojson.Marshal(o.Geometry)
	return json.Marshal(out)
}
