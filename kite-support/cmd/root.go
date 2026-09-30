package cmd

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var app = cli.App{
	Name:        "kite-support",
	Description: "Discord support bot for Kite.",
	Commands: []*cli.Command{
		&botCMD,
		&indexCMD,
	},
}

func Execute() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
