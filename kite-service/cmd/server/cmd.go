package server

import (
	"fmt"

	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal"
	"github.com/merlinfuchs/kite/kite-service/internal/logging"
	"github.com/urfave/cli/v2"
)

func CMD() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "Start and manage the Kite server",
		Subcommands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Start the Kite server",
				Action: func(c *cli.Context) error {
					cfg, err := config.LoadServerConfig(".")
					if err != nil {
						return fmt.Errorf("Failed to load server config: %v", err)
					}

					logging.SetupLogger(cfg.Log)
					return internal.RunServer(cfg)
				},
			},
			migrateCMD(),
		},
	}
}
