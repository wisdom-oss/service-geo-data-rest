package routes

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/twpayne/go-shapefile"
	wisdomType "github.com/wisdom-oss/commonTypes/v2"
	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"

	"microservice/types"
)

var ErrNoRequestBody = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Empty Request Body",
	Detail: "This endpoint requires a request body",
}

func InspectShapefile(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	// read the multipart request
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		errorHandler <- ErrUnsupportedContentType
		return
	}

	err := r.ParseMultipartForm(128 * (1 << 20))
	if err != nil {
		errorHandler <- fmt.Errorf("unable to parse multipart form without errors: %w", err)
		return
	}

	incomingFiles, fileSent := r.MultipartForm.File["file"]
	if !fileSent {
		errorHandler <- ErrNoRequestBody
		return
	}
	incomingFile := incomingFiles[0]
	compressedShapefile, err := incomingFile.Open()
	defer compressedShapefile.Close()
	if err != nil {
		errorHandler <- fmt.Errorf("unable to open the shapefile for reading: %w", err)
		return
	}
	reader, err := zip.NewReader(compressedShapefile, incomingFile.Size)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to create zip reader for shapefile: %w", err)
		return
	}

	shp, err := shapefile.ReadZipReader(reader, nil)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to read shapefile: %w", err)
	}
	var shapefileInformation types.ShapefileInformation
	shapefileInformation.FeatureCount = shp.NumRecords()
	attributes := make(map[string]int)
	for i := 0; i < shp.NumRecords(); i++ {
		fields, _ := shp.Record(i)
		for key, _ := range fields {
			attributes[key] += 1
		}
	}
	shapefileInformation.Attributes = attributes

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(shapefileInformation)
}
