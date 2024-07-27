package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kitecloud/kite/kite-common/api"
	"github.com/kitecloud/kite/kite-common/config"
	"github.com/kitecloud/kite/kite-common/core/engine"
	"github.com/kitecloud/kite/kite-common/core/gateway"
	"github.com/kitecloud/kite/kite-common/db/postgres"
	"github.com/kitecloud/kite/kite-common/logging"
	"github.com/urfave/cli/v2"
)

var serverCMD = cli.Command{
	Name:  "server",
	Usage: "Manage the Kite server.",
	Subcommands: []*cli.Command{
		{
			Name:   "start",
			Usage:  "Start the Kite server.",
			Action: serverStartCMD,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "api",
					Usage: "The Kite API URL to authenticate with.",
				},
			},
		},
	},
}

func serverStartCMD(c *cli.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logging.SetupLogger(cfg.Logging)

	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres))
	if err != nil {
		slog.With("error", err).Error("Failed to create postgres client")
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine := engine.NewEngine(pg, pg, pg)
	engine.Run(ctx)

	gateway := gateway.NewGatewayManager(pg, engine)
	gateway.Run(ctx)

	apiServer := api.NewAPIServer(api.APIServerConfig{
		SecureCookies:       cfg.API.SecureCookies,
		APIPublicBaseURL:    cfg.API.PublicBaseURL,
		AppPublicBaseURL:    cfg.App.PublicBaseURL,
		DiscordClientID:     cfg.Discord.ClientID,
		DiscordClientSecret: cfg.Discord.ClientSecret,
	}, pg, pg, pg, pg, pg)
	address := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	if err := apiServer.Serve(context.Background(), address); err != nil {
		slog.With("error", err).Error("Failed to start API server")
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}
