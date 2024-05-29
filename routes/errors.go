package routes

import wisdomType "github.com/wisdom-oss/commonTypes/v2"

var ErrEmptyLayerID = wisdomType.WISdoMError{
	Type:   "https://https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Empty Layer ID",
	Detail: "Your request did not contain a layer ID. Please check your request",
}

var ErrUnknownLayerID = wisdomType.WISdoMError{
	Type:   "https://https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 404,
	Title:  "Unknown Layer ID",
	Detail: "The specified layer ID is not known. Please check your request",
}
