package types

type ShapefileInformation struct {
	FeatureCount int            `json:"featureCount"`
	Attributes   map[string]int `json:"attributes"`
}
