package db

import (
	"context"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/qustavo/dotsql"
	"github.com/rs/zerolog/log"
	pgxgeom "github.com/twpayne/pgx-geom"

	"microservice/resources"
)

func init() {
	l := log.With().Str("package", "internal/db").Logger()
	l.Debug().Msg("connecting to the database")

	var err error
	config, err := pgxpool.ParseConfig("")
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		if err := pgxgeom.Register(ctx, conn); err != nil {
			return err
		}
		return nil
	}
	Pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		l.Fatal().Err(err).Msg("could not connect to database")
	}
	err = Pool.Ping(context.Background())
	if err != nil {
		l.Fatal().Err(err).Msg("could not ping database")
	}
	l.Debug().Msg("connected to the database")

	l.Debug().Msg("loading prepared sql queries")
	files, err := fs.ReadDir(resources.QueryFiles, ".")
	if err != nil {
		l.Fatal().Err(err).Msg("could not load queries")
	}
	var instances []*dotsql.DotSql
	for _, queryFile := range files {
		fd, err := resources.QueryFiles.Open(queryFile.Name())
		if err != nil {
			l.Fatal().Err(err).Msg("could not open query file")
		}
		instance, err := dotsql.Load(fd)
		if err != nil {
			l.Fatal().Err(err).Msg("could not load query file")
		}
		instances = append(instances, instance)
	}
	Queries = dotsql.Merge(instances...)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load prepared queries")
	}
}
