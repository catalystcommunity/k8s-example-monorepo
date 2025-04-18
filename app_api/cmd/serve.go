package cmd

import (
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/config"
	"github.com/urfave/cli/v2"
)

var ServeCommand = &cli.Command{
	Name:  "serve",
	Usage: "Run the Server",
	Flags: flags,
	Action: func(ctx *cli.Context) error {
		return Serve()
	},
}

var flags = []cli.Flag{
	&cli.StringFlag{
		Name:        "db-uri",
		Aliases:     []string{"db"},
		Value:       "postgresql://devuser:devpass@monodemo-postgresql:5432/monodemopg?sslmode=disable",
		Usage:       "The uri to use to connect to the db",
		Destination: &config.DbUri,
		EnvVars:     []string{"DB_URI"},
	},
	&cli.IntFlag{
		Name:        "port",
		Aliases:     []string{"gp"},
		Value:       6080,
		Usage:       "Port to expose the web API on",
		EnvVars:     []string{"PORT"},
		Destination: &config.Port,
	},
}
