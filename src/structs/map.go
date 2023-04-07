package structs

import geojson "github.com/paulmach/go.geojson"

type Shape struct {
	Name    string           `json:"name"`
	Key     string           `json:"key"`
	NutsKey string           `json:"nutsKey"`
	GeoJSON geojson.Geometry `json:"geojson"`
}
