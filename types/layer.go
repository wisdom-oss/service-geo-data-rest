package types

import "github.com/jackc/pgx/v5/pgtype"

type Layer struct {
	ID                        pgtype.UUID `json:"id" db:"id"`
	Name                      pgtype.Text `json:"name" db:"name"`
	Description               pgtype.Text `json:"description" db:"description"`
	TableName                 pgtype.Text `json:"-" db:"table"`
	CoordinateReferenceSystem *int        `json:"crs" db:"crs"`
}
