package types

// Shapefile contains statistical information about an introspected shapefile
// that's been uploaded.
type Shapefile struct {
	// FeatureCount contains the number of features in the shapefile
	FeatureCount int `json:"featureCount"`

	// Attributes is a map which maps attribute names to the number of shapes
	// on which this attribute exists
	Attributes map[string]int `json:"attributes"`

	// EPSGCode contains the EPSG (European Petroleum Survey Group) Geodetic
	// Parameter Dataset Code to identify the coordinate system and projection
	// used within the shapefile
	//
	// More information: https://epsg.io/
	EPSGCode int `json:"epsg"`

	// Proj4String contains the shapefiles coordinate systems proj4 representation
	Proj4String string `json:"proj4"`
}
