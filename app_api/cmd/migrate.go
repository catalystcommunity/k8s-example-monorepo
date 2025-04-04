package cmd

import (
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/catalystcommunity/k8s-example-monorepo/coredb"
	"github.com/urfave/cli/v2"

	"github.com/catalystcommunity/app-utils-go/errorutils"
	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var migrations = coredb.Migrations

var MigrateCommand = &cli.Command{
	Name:  "migrate",
	Usage: "Runs database migrations",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "db-uri",
			Aliases:     []string{"db"},
			Value:       "postgresql://devuser:devpass@monodemo-postgresql:5432/monodemopg?sslmode=disable",
			Usage:       "The uri to use to connect to the db",
			Destination: &config.DbUri,
			EnvVars:     []string{"DB_URI"},
		},
	},
	Action: func(ctx *cli.Context) error {
		return RunMigrations()
	},
}

func RunMigrations() error {
	db, err := gorm.Open(postgres.Open(config.DbUri), &gorm.Config{})
	errorutils.LogOnErr(nil, "error opening database connection", err)
	if err != nil {
		return err
	}
	sqldb, err := db.DB()
	errorutils.LogOnErr(nil, "error getting database connection", err)
	if err != nil {
		return err
	}
	// set goose file system to use the embedded migrations
	goose.SetBaseFS(migrations)
	logging.Log.Info("Running migrations")
	err = goose.Up(sqldb, "migrations")
	errorutils.LogOnErr(nil, "error running migrations", err)
	if err != nil {
		return err
	}

	return nil
}
