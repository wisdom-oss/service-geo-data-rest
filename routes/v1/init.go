package routes

import (
	"github.com/georgysavva/scany/v2/dbscan"
	"github.com/georgysavva/scany/v2/pgxscan"
)

var scanner *pgxscan.API

func init() {
	api, err := dbscan.NewAPI(dbscan.WithAllowUnknownColumns(true))
	if err != nil {
		panic(err)
	}

	scanner, err = pgxscan.NewAPI(api)
	if err != nil {
		panic(err)
	}
}
