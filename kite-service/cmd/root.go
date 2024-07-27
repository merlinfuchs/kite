package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var app = cli.App{
	Name:        "kite",
	Description: "Kite CLI",
	Commands: []*cli.Command{
		&serverCMD,
		&databaseCMD,
	},
}

func Execute() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
