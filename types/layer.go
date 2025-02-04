package types

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"microservice/internal/db"
)

// Layer represents an entry in the "layers" table.
type Layer struct {
	ID                        pgtype.UUID `db:"id"          json:"id"`
	Name                      string      `db:"name"        json:"name"`
	Description               pgtype.Text `db:"description" json:"description"`
	TableName                 string      `db:"table"       json:"key"`
	Attribution               pgtype.Text `db:"attribution" json:"attribution"`
	CoordinateReferenceSystem pgtype.Int4 `db:"crs"         json:"crs"`
	Private                   bool        `db:"private"     json:"private"`
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
