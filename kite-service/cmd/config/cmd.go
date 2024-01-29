package config

import (
	"github.com/urfave/cli/v2"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Configure the Kite CLI",
		Subcommands: []*cli.Command{
			loginCMD(),
		},
	}
}
