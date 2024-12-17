package resources

import "embed"

//go:embed *.sql
var QueryFiles embed.FS

//go:embed migrations/*.sql
var MigrationFiles embed.FS
