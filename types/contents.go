package types

import "github.com/jackc/pgx/v5/pgtype"

type AttributedContents struct {
	Attribution pgtype.Text `json:"attribution"`
	Contents    []Object    `json:"data"`
}
