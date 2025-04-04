package main

import (
	"os"

	"github.com/catalystcommunity/app-utils-go/logging"
	"github.com/catalystcommunity/k8s-example-monorepo/app_api/cmd"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			cmd.ServeCommand,
			cmd.MigrateCommand,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		// log fatal so we exit with the proper exit code, this is important for containerized deployment health checks
		logging.Log.WithError(err).Fatal("runtime error")
	}
}
