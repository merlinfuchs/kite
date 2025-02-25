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
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/core/usage"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/db/s3"
	"github.com/kitecloud/kite/kite-service/internal/logging"
	"github.com/kitecloud/kite/kite-service/internal/model"
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
		engine.Env{
			Config: engine.EngineConfig{
				MaxStackDepth: cfg.Engine.MaxStackDepth,
				MaxOperations: cfg.Engine.MaxOperations,
				MaxCredits:    cfg.Engine.MaxCredits,
			},
			AppStore:             pg,
			LogStore:             pg,
			UsageStore:           pg,
			MessageStore:         pg,
			MessageInstanceStore: pg,
			CommandStore:         pg,
			EventListenerStore:   pg,
			VariableValueStore:   pg,
			ResumePointStore:     pg,
			HttpClient:           &http.Client{}, // TODO: think about proxying http requests
			OpenaiClient:         openaiClient,
		},
	)
	engine.Run(ctx)

	handler := event.NewEventHandlerWrapper(engine, pg)

	billingPlans := make([]model.Plan, len(cfg.Billing.Plans))
	for i, plan := range cfg.Billing.Plans {
		billingPlans[i] = model.Plan(plan)
	}

	planManager := plan.NewPlanManager(pg, pg, billingPlans, plan.PlanManagerConfig{
		DiscordBotToken: cfg.Discord.BotToken,
		DiscordGuildID:  cfg.Discord.GuildID,
	})
	planManager.Run(ctx)

	gateway := gateway.NewGatewayManager(pg, pg, planManager, handler)
	gateway.Run(ctx)

	usage := usage.NewUsageManager(pg, pg, planManager)
	usage.Run(ctx)

	apiServer := api.NewAPIServer(api.APIServerConfig{
		SecureCookies:       cfg.API.SecureCookies,
		StrictCookies:       cfg.API.StrictCookies,
		APIPublicBaseURL:    cfg.API.PublicBaseURL,
		AppPublicBaseURL:    cfg.App.PublicBaseURL,
		DiscordClientID:     cfg.Discord.ClientID,
		DiscordClientSecret: cfg.Discord.ClientSecret,
		UserLimits: api.APIUserLimitsConfig{
			MaxAppsPerUser:          cfg.UserLimits.MaxAppsPerUser,
			MaxCommandsPerApp:       cfg.UserLimits.MaxCommandsPerApp,
			MaxVariablesPerApp:      cfg.UserLimits.MaxVariablesPerApp,
			MaxMessagesPerApp:       cfg.UserLimits.MaxMessagesPerApp,
			MaxEventListenersPerApp: cfg.UserLimits.MaxEventListenersPerApp,
			MaxAssetSize:            cfg.UserLimits.MaxAssetSize,
		},
		Billing: api.BillingConfig{
			LemonSqueezyAPIKey:        cfg.Billing.LemonSqueezyAPIKey,
			LemonSqueezySigningSecret: cfg.Billing.LemonSqueezySigningSecret,
			LemonSqueezyStoreID:       cfg.Billing.LemonSqueezyStoreID,
			TestMode:                  cfg.Billing.TestMode,
			Plans:                     cfg.Billing.Plans,
		},
	}, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, assetStore, gateway, planManager)
	address := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	if err := apiServer.Serve(ctx, address); err != nil {
		slog.With("error", err).Error("Failed to start API server")
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}
