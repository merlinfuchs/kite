package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/kitecloud/kite/kite-service/internal/api"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/core/event"
	"github.com/kitecloud/kite/kite-service/internal/core/gateway"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/db/s3"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/sashabaranov/go-openai"
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

	s3Client, err := s3.New(cfg.Database.S3)
	if err != nil {
		slog.With("error", err).Error("Failed to create S3 client")
		return fmt.Errorf("failed to create S3 client: %w", err)
	}

	assetStore, err := postgres.NewAssetStore(context.Background(), pg, s3Client)
	if err != nil {
		slog.With("error", err).Warn("Failed to create asset store, continuing without support for assets")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var openaiClient *openai.Client
	if cfg.OpenAI.APIKey != "" {
		openaiClient = openai.NewClient(cfg.OpenAI.APIKey)
	}

	engine := engine.NewEngine(
		engine.EngineConfig{
			MaxStackDepth: cfg.Engine.MaxStackDepth,
			MaxOperations: cfg.Engine.MaxOperations,
			MaxActions:    cfg.Engine.MaxActions,
		},
		pg,
		pg,
		pg,
		pg,
		pg,
		pg,
		&http.Client{}, // TODO: think about proxying http requests
		openaiClient,
	)
	engine.Run(ctx)

	handler := event.NewEventHandlerWrapper(engine, pg)

	gateway := gateway.NewGatewayManager(pg, pg, handler)
	gateway.Run(ctx)

	apiServer := api.NewAPIServer(api.APIServerConfig{
		SecureCookies:       cfg.API.SecureCookies,
		StrictCookies:       cfg.API.StrictCookies,
		APIPublicBaseURL:    cfg.API.PublicBaseURL,
		AppPublicBaseURL:    cfg.App.PublicBaseURL,
		DiscordClientID:     cfg.Discord.ClientID,
		DiscordClientSecret: cfg.Discord.ClientSecret,
		UserLimits: api.APIUserLimitsConfig{
			MaxAppsPerUser:     cfg.API.UserLimits.MaxAppsPerUser,
			MaxCommandsPerApp:  cfg.API.UserLimits.MaxCommandsPerApp,
			MaxVariablesPerApp: cfg.API.UserLimits.MaxVariablesPerApp,
			MaxMessagesPerApp:  cfg.API.UserLimits.MaxMessagesPerApp,
			MaxAssetSize:       cfg.API.UserLimits.MaxAssetSize,
		},
	}, pg, pg, pg, pg, pg, pg, pg, pg, pg, assetStore, gateway)
	address := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	if err := apiServer.Serve(context.Background(), address); err != nil {
		slog.With("error", err).Error("Failed to start API server")
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}
