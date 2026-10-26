package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kitecloud/kite/kite-service/internal/api/session"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/urfave/cli/v2"
)

var adminCMD = cli.Command{
	Name:  "admin",
	Usage: "Admin commands used for debugging and administration",
	Subcommands: []*cli.Command{
		{
			Name:   "impersonate",
			Usage:  "Impersonate a user by creating a session token that can be injected into the session cookie",
			Action: adminImpersonateCMD,
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "user_id",
					Usage: "The ID of the user to impersonate.",
				},
			},
		},
	},
}

func adminImpersonateCMD(c *cli.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logging.SetupLogger(cfg.Logging)

	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres), 1)
	if err != nil {
		slog.Error(
			"Failed to create postgres client",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	sessionManager := session.NewSessionManager(session.SessionManagerConfig{}, pg)

	key, _, err := sessionManager.CreateSession(context.Background(), c.String("user_id"))
	if err != nil {
		slog.Error(
			"Failed to create session",
			slog.String("error", err.Error()),
			slog.String("user_id", c.String("user_id")),
		)
		return fmt.Errorf("failed to create session: %w", err)
	}

	fmt.Println(key)
	return nil
}
