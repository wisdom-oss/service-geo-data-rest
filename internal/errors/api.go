package errors

import "github.com/wisdom-oss/common-go/v2/types"

var ErrUnknownLayer = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5",
	Status: 404,
	Title:  "Unknown Layer ID",
	Detail: "The specified layer ID is not known. Please check your request",
}

var ErrUnknownObject = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5",
	Status: 404,
	Title:  "Unknown Object IDs",
	Detail: "None of the keys are resolvable into geometries",
}

var ErrUnsupportedSpatialRelation = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Unsupported Spatial Relation",
	Detail: "The selected spatial relation for the query is not supported",
}
