package types

import (
	"encoding/json"

	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/geojson"
)

const geometryDecimalPrecision = 15

var defaultEncodingParameters = []geojson.EncodeGeometryOption{
	geojson.EncodeGeometryWithBBox(),
	geojson.EncodeGeometryWithMaxDecimalDigits(geometryDecimalPrecision),
}

// Object represents a single object contained in a Layer
// As the marshaling of the object is a bit complex due to the geometry
// requiring an extra marshaling step the MarshalJSON function has been
// implemented.
type Object struct {
	ID                   uint64                 `db:"id"                    json:"id"`
	Geometry             geom.T                 `db:"geometry"              json:"geometry"`
	Name                 *string                `db:"name"                  json:"name"`
	Key                  string                 `db:"key"                   json:"key"`
	AdditionalProperties map[string]interface{} `db:"additional_properties" json:"additionalProperties"`
}

// _object is used as the marshaling object for the Object as the geometry
// is manually encoded and decoded as the geom.T interface doesn't implement
// the [json.Marshaler] interface.
type _object struct {
	ID                   uint64                 `json:"id"`
	Geometry             json.RawMessage        `json:"geometry"`
	Name                 *string                `json:"name"`
	Key                  string                 `json:"key"`
	AdditionalProperties map[string]interface{} `json:"additionalProperties"`
}

// MarshalJSON implements the [json.Marshaler] interface as the geom.T interface
// doesn't implement/requires it. It creates a new instance of a _object and
// manually converts the Geometry into a GeoJSON output and adds it to the
// output object before.
func (o Object) MarshalJSON() ([]byte, error) {
	output := _object{
		ID:                   o.ID,
		Name:                 o.Name,
		Key:                  o.Key,
		AdditionalProperties: o.AdditionalProperties,
	}
	output.Geometry, _ = geojson.Marshal(o.Geometry, defaultEncodingParameters...)
	return json.Marshal(output)
}
