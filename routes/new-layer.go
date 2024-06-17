package routes

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"slices"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/twpayne/go-shapefile"
	wisdomType "github.com/wisdom-oss/commonTypes/v2"
	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"

	"microservice/globals"
	"microservice/types"
)

var ErrUnsupportedContentType = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.16",
	Status: 415,
	Title:  "Unsupported Content Type",
	Detail: "The creation of a new layer requires the usage of `multipart/form-data` as content type",
}

var ErrMissingField = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Missing Field in View Configuration",
	Detail: "",
}

var ErrUnknownRelation = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "[DB] Unknown Relation",
	Detail: "The relation you specified does not exist. Therefore no layer could be created.",
}

var ErrLayerNameMissing = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Layer Name missing",
	Detail: "The request did not contain a name for the new layer",
}

var ErrRecordNameFieldMissing = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Record Name Identifier not set",
	Detail: "The request did not contain a value which identifies the records name",
}

var ErrRecordKeyFieldMissing = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Record Key Identifier not set",
	Detail: "The request did not contain a value which identifies the records key",
}

var ErrUnknownCreationMethod = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Unknown Creation Method",
	Detail: "The request did specify the method of the layer creation or contained multiple options",
}

const LayerNameFormKey = "layer-name"
const LayerDescriptionFormKey = "layer-description"
const LayerShapefileKey = "shape-file"
const LayerViewKey = "view-configuration"
const RecordNameKey = "record-name-field"
const RecordKeyKey = "record-key-field"
const AdditionalPropertiesKey = "additional-property"

// NewLayer handles the creation of a new layer.
func NewLayer(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	// read the multipart request
	contentType := strings.TrimSpace(r.Header.Get("Content-Type"))
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		errorHandler <- ErrUnsupportedContentType
		return
	}

	// parse the multipart form with a max amount of 128MiB stored in memory
	err := r.ParseMultipartForm(128 * (1 << 20))
	if err != nil {
		errorHandler <- fmt.Errorf("unable to parse multipart form without errors: %w", err)
		return
	}

	form := r.MultipartForm

	var layerName string
	layerNames, isSet := form.Value[LayerNameFormKey]
	if !isSet {
		errorHandler <- ErrLayerNameMissing
		return
	} else {
		layerName = layerNames[0]
	}
	var layerDescription *string
	layerDescriptions, isSet := form.Value[LayerDescriptionFormKey]
	if !isSet {
		layerDescription = nil
	} else {
		layerDescription = &layerDescriptions[0]
	}

	// now check if the layer is being created by using either using a view or
	// a shapefile
	rawLayerConfigurations, layerConfigurationsSent := form.Value[LayerViewKey]
	shapeFileHeaders, shapeFilesSent := form.File[LayerShapefileKey]

	switch {
	case layerConfigurationsSent && shapeFilesSent, !layerConfigurationsSent && !shapeFilesSent:
		errorHandler <- ErrUnknownCreationMethod
		return
	case layerConfigurationsSent:
		createLayerFromView(w, r, layerName, layerDescription, rawLayerConfigurations[0])
		return
	default:
		recordNameIdentifiers, recordNameIdentifierSet := form.Value[RecordNameKey]
		if !recordNameIdentifierSet {
			errorHandler <- ErrRecordNameFieldMissing
			return
		}
		recordKeyIdentifiers, recordKeyIdentifiersSet := form.Value[RecordKeyKey]
		if !recordKeyIdentifiersSet {
			errorHandler <- ErrRecordKeyFieldMissing
			return
		}
		additionalProperties, _ := form.Value[AdditionalPropertiesKey]

		createLayerFromShapefile(w, r, layerName, layerDescription, recordNameIdentifiers[0], recordKeyIdentifiers[0], additionalProperties, shapeFileHeaders[0])
		return
	}
}

func createLayerFromView(w http.ResponseWriter, r *http.Request, layerName string, layerDescription *string, rawViewConfiguration string) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	var viewConfiguration types.ViewConfiguration
	err := json.Unmarshal([]byte(rawViewConfiguration), &viewConfiguration)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to parse view configuration: %w", err)
		return
	}

	query, err := viewConfiguration.BuildCreateQuery()
	if err != nil {
		e := ErrMissingField
		e.Detail = err.Error()
		errorHandler <- e
		return
	}

	_, err = globals.Db.Exec(r.Context(), query)
	if err != nil {
		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "42P01":
				errorHandler <- ErrUnknownRelation
				return
			default:
				errorHandler <- err
				return
			}
		}
		errorHandler <- err
		return
	}

	query, err = globals.SqlQueries.Raw("crate-layer-definition")
	if err != nil {
		errorHandler <- err
		return
	}

	var layer types.Layer
	err = pgxscan.Get(r.Context(), globals.Db, &layer, query, layerName, layerDescription, viewConfiguration.TableName(), nil)
	if err != nil {
		errorHandler <- err
		return
	}
	_ = json.NewEncoder(w).Encode(layer)
}

func createLayerFromShapefile(w http.ResponseWriter, r *http.Request, layerName string, layerDescription *string, recordNameKey string, recordKeyKey string, additionalPropertyKeys []string, shapeFileHeader *multipart.FileHeader) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	shapeFileIncoming, err := shapeFileHeader.Open()
	if err != nil {
		errorHandler <- fmt.Errorf("unable to open shapefile: %w", err)
		return
	}
	defer shapeFileIncoming.Close()
	reader, err := zip.NewReader(shapeFileIncoming, shapeFileHeader.Size)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to open zip reader: %w", err)
	}
	shapeFile, err := shapefile.ReadZipReader(reader, nil)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to read zipped shapefile: %w", err)
		return
	}

	tx, err := globals.Db.Begin(r.Context())
	if err != nil {
		errorHandler <- fmt.Errorf("unable to create transaction for shape file insertion: %w", err)
		return
	}
	tableCreationQuery, err := globals.SqlQueries.Raw("create-layer-table")
	if err != nil {
		errorHandler <- fmt.Errorf("unable to load table creation query: %w", err)
		return
	}
	tableCreationQuery = fmt.Sprintf(tableCreationQuery, strcase.ToSnake(layerName))
	_, err = tx.Exec(r.Context(), tableCreationQuery)
	if err != nil {
		rbErr := tx.Rollback(r.Context())
		if rbErr != nil {
			errorHandler <- fmt.Errorf("unable to create tabe for shape file: %w | rollback of database failed: %w", err, rbErr)
		} else {
			errorHandler <- fmt.Errorf("unable to create table for shape file: %w", err)
		}
		return
	}
	shapeInsertQuery, err := globals.SqlQueries.Raw("insert-shape-object")
	if err != nil {
		errorHandler <- fmt.Errorf("unable to load table creation query: %w", err)
		return
	}
	shapeInsertQuery = fmt.Sprintf(shapeInsertQuery, strcase.ToSnake(layerName))
	for i := 0; i < shapeFile.NumRecords(); i++ {
		fields, geometry := shapeFile.Record(i)
		additionalProperties := make(map[string]interface{})
		for name, value := range fields {
			if slices.Contains(additionalPropertyKeys, name) {
				additionalProperties[strcase.ToLowerCamel(name)] = value
			}
		}
		name, fieldSet := fields[recordNameKey]
		if !fieldSet {
			rbErr := tx.Rollback(r.Context())
			if rbErr != nil {
				errorHandler <- fmt.Errorf("unable to resolve object name: %w | rollback of database failed: %w", err, rbErr)
			} else {
				errorHandler <- fmt.Errorf("unable to resolve object name: %w", err)
			}
			return
		}
		key, fieldSet := fields[recordKeyKey]
		if !fieldSet {
			rbErr := tx.Rollback(r.Context())
			if rbErr != nil {
				errorHandler <- fmt.Errorf("unable to get key field: %w | rollback of database failed: %w", err, rbErr)
			} else {
				errorHandler <- fmt.Errorf("unable to get key field: %w", err)
			}
			return
		}
		keyString := fmt.Sprintf("%d", key.(int))
		_, err = tx.Exec(r.Context(), shapeInsertQuery, geometry, keyString, name, additionalProperties)
		if err != nil {
			rbErr := tx.Rollback(r.Context())
			if rbErr != nil {
				errorHandler <- fmt.Errorf("unable to insert record %w | rollback of database failed: %w", err, rbErr)
			} else {
				errorHandler <- fmt.Errorf("unable to insert record: %w", err)
			}
			return
		}
	}
	query, err := globals.SqlQueries.Raw("crate-layer-definition")
	if err != nil {
		errorHandler <- err
		return
	}
	row, err := tx.Query(r.Context(), query, layerName, layerDescription, strcase.ToSnake(layerName), 0)
	if err != nil {
		rbErr := tx.Rollback(r.Context())
		if rbErr != nil {
			errorHandler <- fmt.Errorf("unable to create layer entry: %w | rollback of database failed: %w", err, rbErr)
		} else {
			errorHandler <- fmt.Errorf("unable to create layer entry: %w", err)
		}
		return
	}
	var layer types.Layer
	err = pgxscan.ScanOne(&layer, row)
	if err != nil {
		rbErr := tx.Rollback(r.Context())
		if rbErr != nil {
			errorHandler <- fmt.Errorf("unable to scan new layer entry: %w | rollback of database failed: %w", err, rbErr)
		} else {
			errorHandler <- fmt.Errorf("unable to scan new layer entry: %w", err)
		}
		return
	}
	err = tx.Commit(r.Context())
	if err != nil {
		errorHandler <- fmt.Errorf("unable to commit database changes: %w", err)
		return
	}
}
