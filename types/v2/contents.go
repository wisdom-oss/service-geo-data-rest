package v2

import (
	"github.com/jackc/pgx/v5/pgtype"

	"microservice/types"
)

type AttributedContents struct {
	Attribution pgtype.Text    `json:"attribution"`
	Contents    []types.Object `json:"data"`
}
