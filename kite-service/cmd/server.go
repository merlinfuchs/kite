package cmd

import (
	"github.com/kitecloud/kite/kite-service/internal/entry/server"
	"github.com/urfave/cli/v2"
)

var serverCMD = cli.Command{
	Name:  "server",
	Usage: "Manage the Kite server.",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "Start the Kite server.",
			Action: func(ctx *cli.Context) error {
				return server.StartServer(ctx.Context)
			},
		},
	},
}
