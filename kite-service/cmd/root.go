package cmd

import (
	"log"
	"os"

	"github.com/merlinfuchs/kite/kite-service/cmd/config"
	"github.com/merlinfuchs/kite/kite-service/cmd/plugin"
	"github.com/merlinfuchs/kite/kite-service/cmd/server"
	"github.com/urfave/cli/v2"
)

func Entry() {
	app := &cli.App{
		Name:  "kite",
		Usage: "Kite CLI tool for managing plugins and running the Kite server",
		Commands: []*cli.Command{
			server.CMD(),
			plugin.CMD(),
			config.CMD(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
