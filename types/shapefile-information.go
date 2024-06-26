package types

type ShapefileInformation struct {
	FeatureCount int            `json:"featureCount"`
	Attributes   map[string]int `json:"attributes"`
	EPSGCode     int            `json:"epsg"`
	Proj4String  string         `json:"proj4"`
}
