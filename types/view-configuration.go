package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/iancoleman/strcase"
)

var ErrMissingField = errors.New("missing required field")

type LayerConfiguration struct {
	// Name contains the name of the new layer
	Name string `json:"name"`
	// Description contains the optional description of the new layer
	Description *string `json:"description"`

	ViewConfiguration ViewConfiguration `json:"viewConfiguration"`
}

type ViewConfiguration struct {
	// Schema contains the name of the DB schema in which Table lays
	Schema string `json:"schema"`

	// Table represents the name of the table used to create the view from
	Table string `json:"table"`

	// ID contains the name of the column that should be used as [Object.ID]
	// column (which is an internal id)
	ID string `json:"id"`

	// Key contains the name of the column that specifies the keys that are used
	// as another identifier for external views
	Key string `json:"key"`

	// Geometry contains the name of the geometry column
	Geometry string `json:"geometry"`

	// Name contains the name of the name column
	Name string `json:"name"`

	// AdditionalPropertyKeys contains the keys that will be used to build the
	// [Object.AdditionalProperties] value
	AdditionalPropertyKeys []string `json:"additionalPropertyKeys"`

	// WhereCondition contains a plain sql where condition that may prefilter the
	// view
	WhereCondition string `json:"whereCondition"`
}

func (vc *ViewConfiguration) validate() error {
	if strings.TrimSpace(vc.Schema) == "" {
		return fmt.Errorf("%w: schema", ErrMissingField)
	}
	if strings.TrimSpace(vc.Table) == "" {
		return fmt.Errorf("%w: table", ErrMissingField)
	}
	if strings.TrimSpace(vc.ID) == "" {
		return fmt.Errorf("%w: id", ErrMissingField)
	}
	if strings.TrimSpace(vc.Name) == "" {
		return fmt.Errorf("%w: name", ErrMissingField)
	}
	if strings.TrimSpace(vc.Key) == "" {
		return fmt.Errorf("%w: key", ErrMissingField)
	}
	if strings.TrimSpace(vc.Geometry) == "" {
		return fmt.Errorf("%w: geometry", ErrMissingField)
	}
	*vc = ViewConfiguration{
		Schema:                 strings.TrimSpace(vc.Schema),
		Table:                  strings.TrimSpace(vc.Table),
		ID:                     strings.TrimSpace(vc.ID),
		Key:                    strings.TrimSpace(vc.Key),
		Geometry:               strings.TrimSpace(vc.Geometry),
		Name:                   strings.TrimSpace(vc.Name),
		AdditionalPropertyKeys: vc.AdditionalPropertyKeys,
		WhereCondition:         strings.TrimSpace(vc.WhereCondition),
	}
	return nil
}

func (vc ViewConfiguration) BuildCreateQuery() (string, error) {
	if err := vc.validate(); err != nil {
		return "", err
	}
	const QueryPattern = `CREATE VIEW geodata.view_%s AS SELECT %s as id, %s as key, %s as geometry, %s as name, %s as additional_properties FROM %s.%s %s;`
	jsonBuildObjectCommand := "json_build_object("
	for _, propertyKey := range vc.AdditionalPropertyKeys {
		jsonBuildObjectCommand += fmt.Sprintf(`'%s', %s,`, strcase.ToLowerCamel(strings.TrimSpace(propertyKey)), strings.TrimSpace(propertyKey))
	}
	jsonBuildObjectCommand = strings.TrimSuffix(jsonBuildObjectCommand, ",")
	jsonBuildObjectCommand += ")"

	return fmt.Sprintf(QueryPattern, vc.Table, vc.ID, vc.Key, vc.Geometry, vc.Name, jsonBuildObjectCommand, vc.Schema, vc.Table, vc.WhereCondition), nil
}

func (vc ViewConfiguration) TableName() string {
	if err := vc.validate(); err != nil {
		return ""
	}
	return fmt.Sprintf("view_%s", vc.Table)
}
