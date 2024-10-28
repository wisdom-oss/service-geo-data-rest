package types

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"microservice/internal/db"
)

// Layer represents an entry in the "layers" table
type Layer struct {
	ID                        pgtype.UUID `json:"id" db:"id"`
	Name                      string      `json:"name" db:"name"`
	Description               pgtype.Text `json:"description" db:"description"`
	TableName                 string      `json:"key" db:"table"`
	Attribution               pgtype.Text `json:"attribution" db:"attribution"`
	CoordinateReferenceSystem pgtype.Int4 `json:"crs" db:"crs"`
}

func (l Layer) ContentQuery() (string, error) {
	rawQuery, err := db.Queries.Raw("get-layer-contents")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(rawQuery, l.TableName), nil
}

func (l Layer) FilteredContentQuery() (string, error) {
	rawQuery, err := db.Queries.Raw("get-layer-object-by-key")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(rawQuery, l.TableName), nil
}
