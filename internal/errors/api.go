package errors

import (
	"net/http"

	"github.com/wisdom-oss/common-go/v3/types"
)

var ErrUnknownLayer = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5",
	Status: http.StatusNotFound,
	Title:  "Unknown Layer ID",
	Detail: "The specified layer ID is not known. Please check your request",
}

var ErrUnknownObject = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.5",
	Status: http.StatusNotFound,
	Title:  "Unknown Object IDs",
	Detail: "None of the keys are resolvable into geometries",
}

var ErrUnsupportedSpatialRelation = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusBadRequest,
	Title:  "Unsupported Spatial Relation",
	Detail: "The selected spatial relation for the query is not supported",
}

var ErrMissingParameter = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusBadRequest,
	Title:  "Request Missing Parameter",
	Detail: "The request is missing a required parameter. Check the error field for more information",
}

var ErrUnknownTopLayer = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusBadRequest,
	Title:  "Unknown Top Layer",
	Detail: "The specified top layer is unknown.",
}

var ErrLayerPrivate = types.ServiceError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: http.StatusForbidden,
	Title:  "Forbidden Layer Accessed",
	Detail: "The specified layer requires more access priviliges to be displayed.",
}
