package routes

import (
	"archive/zip"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/goccy/go-json"
	"github.com/iancoleman/strcase"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/twpayne/go-shapefile"
	wisdomType "github.com/wisdom-oss/commonTypes/v2"
	errorMiddleware "github.com/wisdom-oss/microservice-middlewares/v5/error"
	"golang.org/x/exp/maps"

	"microservice/globals"
	"microservice/helpers"
	"microservice/types"
)

var ErrInvalidContentType = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Invalid Content Type",
	Detail: "The content type used in your request has not been recognized. Please check your request",
}

var ErrUnsupportedContentType = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.16",
	Status: 415,
	Title:  "Unsupported Content Type",
	Detail: "The creation of a new layer either required the usage `multipart/form-data` or 'application/json' as content type. Please refer to the documentation",
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

var ErrNoProjectionProvided = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "[Shapefile] No Projection Found",
	Detail: "The shapefile needs to be accompanied by a projection file to allow the database to handle the entries correct.",
}

var ErrNoEPSGCodeFound = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "[Shapefile] No EPSG Code Found",
	Detail: "The shapefile contained a projection which could not be resolved to an EPSG code. Please transform your shapes into a known EPSG code projection",
}

// NewLayer handles the creation of a new layer.
func NewLayer(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	// detect the mimetype of the request and exclude the parameters as they are
	// not needed for the route
	mimetype, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		if errors.Is(err, mime.ErrInvalidMediaParameter) {
			errorHandler <- ErrInvalidContentType
			return
		}
		errorHandler <- err
		return
	}

	switch mimetype {
	case "application/json", "text/json":
		createLayerFromView(w, r)
		return
	case "multipart/form-data":
		createLayerFromUpload(w, r)
		return
	default:
		errorHandler <- ErrUnsupportedContentType
		return
	}
}

func createLayerFromView(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	var layerConfiguration types.LayerConfiguration
	err := json.NewDecoder(r.Body).Decode(&layerConfiguration)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to decode layer configuration: %w", err)
		return
	}

	// now extract the query used to create the view
	query, err := layerConfiguration.ViewConfiguration.BuildCreateQuery()
	if errors.Is(err, types.ErrMissingField) {
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
	err = pgxscan.Get(r.Context(), globals.Db, &layer, query, layerConfiguration.Name, layerConfiguration.Description, layerConfiguration.ViewConfiguration.TableName(), nil)
	if err != nil {
		errorHandler <- err
		return
	}
	_ = json.NewEncoder(w).Encode(layer)
}

const FormKeyLayerName = "layer.name"
const FormKeyLayerDescription = "layer.description"
const FormKeyLayerAttribution = "layer.attribution"
const FormKeyLayerAdditionalAttribute = "additional-attribute"
const FormKeyNameAttribute = "attributes.name"
const FormKeyKeyAttribute = "attributes.key"
const FormKeyArchiveFile = "archive"
const FormKeyEsriShapesFile = "shp"
const FormKeyEsriAttributesFile = "dbf"
const FormKeyEsriShapeIndexFile = "shx"
const FormKeyEsriProjectionFile = "prj"
const FormKeyEsriCodePageFile = "cpg"

var ErrNoUploadMethodDetected = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "No Upload Method Detected",
	Detail: "The fields contained in the request match neither the required fields for an archive upload nor the required fields for a direct upload. Please check that you only use one of the upload methods",
}

var ErrEmptyLayerName = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Empty Layer Name set",
	Detail: "The request only contained an empty layer name. This is not allowed",
}

var ErrInvalidNameType = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Invalid Type for Name",
	Detail: "The upload only supports strings as name value. Please check your layer",
}

var ErrInvalidKeyType = wisdomType.WISdoMError{
	Type:   "https://www.rfc-editor.org/rfc/rfc9110#section-15.5.1",
	Status: 400,
	Title:  "Invalid Type for Key",
	Detail: "The upload only supports strings as key value. Please check your layer",
}

var alwaysOptionalKeys = []string{FormKeyLayerAdditionalAttribute, FormKeyLayerDescription, FormKeyLayerAttribution}
var archiveUploadRequiredKeys = []string{FormKeyLayerName, FormKeyArchiveFile, FormKeyNameAttribute, FormKeyKeyAttribute}
var archiveUploadOptionalKeys = []string{}
var directUploadRequiredKeys = []string{FormKeyLayerName, FormKeyEsriShapesFile, FormKeyEsriAttributesFile, FormKeyEsriShapeIndexFile, FormKeyNameAttribute, FormKeyKeyAttribute, FormKeyEsriProjectionFile}
var directUploadOptionalKeys = []string{FormKeyEsriCodePageFile}

func createLayerFromUpload(w http.ResponseWriter, r *http.Request) {
	errorHandler := r.Context().Value(errorMiddleware.ChannelName).(chan<- interface{})

	// now parse the multipart form (with max of 256 MiB stored in memory)
	err := r.ParseMultipartForm(256 * (1 << 20))
	if err != nil {
		errorHandler <- fmt.Errorf("unable to parse multipart form: %w", err)
		return
	}

	// now get the keys to check if the required keys for either upload method
	// are available
	f := r.MultipartForm
	availableKeys := maps.Keys(f.Value)
	availableKeys = append(availableKeys, maps.Keys(f.File)...)

	// now sort the keys
	slices.Sort(availableKeys)
	slices.Sort(archiveUploadRequiredKeys)
	slices.Sort(archiveUploadOptionalKeys)
	slices.Sort(directUploadRequiredKeys)
	slices.Sort(directUploadOptionalKeys)
	slices.Sort(alwaysOptionalKeys)

	// remove the optional keys
	recognitionKeys := slices.Clone(availableKeys)
	recognitionKeys = helpers.SubtractArrays(recognitionKeys, archiveUploadOptionalKeys)
	recognitionKeys = helpers.SubtractArrays(recognitionKeys, directUploadOptionalKeys)
	recognitionKeys = helpers.SubtractArrays(recognitionKeys, alwaysOptionalKeys)

	// now check if one of the arrays match the upload types
	isArchiveUpload := slices.Equal(recognitionKeys, archiveUploadRequiredKeys)
	isDirectUpload := slices.Equal(recognitionKeys, directUploadRequiredKeys)

	shpFile := &shapefile.Shapefile{}

	switch {
	case !isArchiveUpload && !isDirectUpload, isArchiveUpload && isDirectUpload:
		errorHandler <- ErrNoUploadMethodDetected
		return
	case isDirectUpload:
		// read the .shp file
		hdr := f.File[FormKeyEsriShapesFile][0]
		file, err := hdr.Open()
		if err != nil {
			errorHandler <- fmt.Errorf("unable to open shp file: %w", err)
			return
		}
		shp, err := shapefile.ReadSHP(file, hdr.Size, nil)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to read shp file: %w", err)
			return
		}
		shpFile.SHP = shp
		file.Close()

		// read the .dbf file
		hdr = f.File[FormKeyEsriAttributesFile][0]
		file, err = hdr.Open()
		if err != nil {
			errorHandler <- fmt.Errorf("unable to open dbf file: %w", err)
			return
		}
		dbf, err := shapefile.ReadDBF(file, hdr.Size, nil)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to read dbf file: %w", err)
			return
		}
		shpFile.DBF = dbf

		// read the .shx file
		hdr = f.File[FormKeyEsriShapeIndexFile][0]
		file, err = hdr.Open()
		if err != nil {
			errorHandler <- fmt.Errorf("unable to open shx file: %w", err)
			return
		}
		shx, err := shapefile.ReadSHX(file, hdr.Size)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to read shx file: %w", err)
			return
		}
		shpFile.SHX = shx

		// now check for the optional .prj and .cpg files
		if slices.Contains(availableKeys, FormKeyEsriProjectionFile) {
			hdr = f.File[FormKeyEsriProjectionFile][0]
			file, err = hdr.Open()
			if err != nil {
				errorHandler <- fmt.Errorf("unable to open prj file: %w", err)
				return
			}
			prj, err := shapefile.ReadPRJ(file, hdr.Size)
			if err != nil {
				errorHandler <- fmt.Errorf("unable to read prj file: %w", err)
				return
			}
			shpFile.PRJ = prj
		}

		if slices.Contains(availableKeys, FormKeyEsriCodePageFile) {
			hdr = f.File[FormKeyEsriCodePageFile][0]
			file, err = hdr.Open()
			if err != nil {
				errorHandler <- fmt.Errorf("unable to open cpg file: %w", err)
				return
			}
			cpg, err := shapefile.ReadCPG(file, hdr.Size)
			if err != nil {
				errorHandler <- fmt.Errorf("unable to read cpg file: %w", err)
				return
			}
			shpFile.CPG = cpg
		}
		break
	case isArchiveUpload:
		// open the archived file and create a ZIP reader on it
		archive := f.File[FormKeyArchiveFile][0]
		file, err := archive.Open()
		if err != nil {
			errorHandler <- fmt.Errorf("unable to open transmitted archive: %w", err)
			return
		}
		reader, err := zip.NewReader(file, archive.Size)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to create zip reader on archvie: %w", err)
			return
		}
		shpFile, err = shapefile.ReadZipReader(reader, nil)
		if err != nil {
			errorHandler <- fmt.Errorf("unable to read shapefile from archive: %w", err)
			return
		}
	}

	// now try to resolve the EPSG code for the projection
	if shpFile.PRJ == nil {
		errorHandler <- ErrNoProjectionProvided
		return
	}

	out, err := helpers.SpatialReferenceInformation(shpFile.PRJ.Projection, "epsg")
	if err != nil {
		errorHandler <- ErrNoEPSGCodeFound
		return
	}

	epsgCode := out.(int)

	// now extract the name and description from the request
	layerName := f.Value[FormKeyLayerName][0]
	if strings.TrimSpace(layerName) == "" {
		errorHandler <- ErrEmptyLayerName
		return
	}
	tableName := strcase.ToSnake(layerName)

	layerDescription := helpers.Pointer("")
	if slices.Contains(availableKeys, FormKeyLayerDescription) {
		*layerDescription = f.Value[FormKeyLayerDescription][0]
	}

	layerAttribution := helpers.Pointer("")
	if slices.Contains(availableKeys, FormKeyLayerAttribution) {
		*layerAttribution = f.Value[FormKeyLayerAttribution][0]
	}

	// now create the table for the new layer and prepare it
	tableCreationQuery, err := globals.SqlQueries.Raw("create-layer-table")
	if err != nil {
		errorHandler <- fmt.Errorf("unable to load layer creation query: %w", err)
		return
	}
	tableCreationQuery = fmt.Sprintf(tableCreationQuery, tableName)

	layerDefinitionQuery, err := globals.SqlQueries.Raw("crate-layer-definition")
	if err != nil {
		errorHandler <- err
		return
	}

	geometryUpdateQuery, err := globals.SqlQueries.Raw("update-geometry-srid")
	if err != nil {
		errorHandler <- err
		return
	}

	// now start a transaction to allow writing the whole shape in a controlled
	// manner
	tx, err := globals.Db.Begin(r.Context())
	if err != nil {
		errorHandler <- fmt.Errorf("unable to begin database transaction: %w", err)
		return
	}

	// now defer a rollback of the transaction here to remove handing this case
	// every time.
	// according to the pgx documentation rollback may even be called without an
	// error if tx.Commit has been called previously
	defer tx.Rollback(r.Context())

	_, err = tx.Exec(r.Context(), tableCreationQuery)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to prepare statement for layer creation query: %w", err)
		return
	}

	_, err = tx.Prepare(r.Context(), "define-layer", layerDefinitionQuery)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to prepare statement for layer definition query: %w", err)
		return
	}

	_, err = tx.Prepare(r.Context(), "update-srid", geometryUpdateQuery)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to prepare statement for geometry update: %w", err)
		return
	}

	var values [][]interface{}

	// make all additional attribute keys lower camel case
	var additionalAttributeKeys []string
	for _, key := range f.Value[FormKeyLayerAdditionalAttribute] {
		additionalAttributeKeys = append(additionalAttributeKeys, strcase.ToLowerCamel(key))
	}

	// now iterate over the shapefiles entries
	for i := 0; i < shpFile.NumRecords(); i++ {
		attributes, geometry := shpFile.Record(i)

		// create a map for the additional attributes
		additionalProperties := make(map[string]interface{})
		for key, value := range attributes {
			key = strcase.ToLowerCamel(key)
			if slices.Contains(additionalAttributeKeys, key) {
				additionalProperties[strcase.ToLowerCamel(key)] = value
			}
		}

		val := attributes[f.Value[FormKeyNameAttribute][0]]
		if reflect.TypeOf(val).Kind() != reflect.String {
			errorHandler <- ErrInvalidNameType
			return
		}
		name := val.(string)

		val = attributes[f.Value[FormKeyKeyAttribute][0]]
		if reflect.TypeOf(val).Kind() != reflect.String {
			errorHandler <- ErrInvalidKeyType
			return
		}
		key := val.(string)

		values = append(values, []interface{}{geometry, name, key, additionalProperties})
	}

	_, err = tx.CopyFrom(r.Context(), []string{"geodata", tableName}, []string{"geometry", "name", "key", "additional_properties"}, pgx.CopyFromRows(values))
	if err != nil {
		errorHandler <- fmt.Errorf("unable to insert geometrties into table: %w", err)
		return
	}

	_, err = tx.Exec(r.Context(), "update-srid", tableName, epsgCode)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to update geometry srid: %w", err)
		return
	}

	row, err := tx.Query(r.Context(), "define-layer", layerName, layerDescription, tableName, epsgCode, layerAttribution)
	var layer types.Layer
	err = pgxscan.ScanOne(&layer, row)
	if err != nil {
		errorHandler <- fmt.Errorf("unable to define layer: %w", err)
		return
	}

	err = tx.Commit(r.Context())
	if err != nil {
		errorHandler <- fmt.Errorf("unable to commit changes to the database: %w", err)
		return
	}

	_ = json.NewEncoder(w).Encode(layer)

}
