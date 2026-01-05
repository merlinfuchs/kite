package cmd

import (
	"fmt"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/entry/server"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/urfave/cli/v2"
)

var serverCMD = cli.Command{
	Name:  "server",
	Usage: "Manage the Kite server.",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "Start the Kite server.",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name:  "cluster_count",
					Usage: "The number of clusters to use.",
					Value: 1,
				},
				&cli.IntFlag{
					Name:  "cluster_index",
					Usage: "The index of the cluster to use.",
					Value: 0,
				},
				&cli.IntFlag{
					Name:  "port",
					Usage: "The port to use for the API server.",
					Value: 8080,
				},
			},
			Action: func(ctx *cli.Context) error {
				cfg, err := config.LoadConfig(".")
				if err != nil {
					return fmt.Errorf("failed to load config: %w", err)
				}

				logging.SetupLogger(cfg.Logging)

				if ctx.IsSet("cluster_count") {
					cfg.ClusterCount = ctx.Int("cluster_count")
				}
				if ctx.IsSet("cluster_index") {
					cfg.ClusterIndex = ctx.Int("cluster_index")
				}
				if ctx.IsSet("port") {
					cfg.API.Port = ctx.Int("port")
				}

				return server.StartServer(ctx.Context, cfg)
			},
		},
	},
}
