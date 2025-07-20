package server

import (
	"context"
	"fmt"
	"log/slog"

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
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/plugin/counting"
	"github.com/kitecloud/kite/kite-service/pkg/plugin/starboard"
	"github.com/sashabaranov/go-openai"
)

func StartServer(c context.Context) error {
	cfg, err := config.LoadConfig(".")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logging.SetupLogger(cfg.Logging)

	patchDiscordProxyURL(cfg)

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

	tokenCrypt, err := util.NewSymmetricCrypt(cfg.Encryption.TokenEncryptionKey)
	if err != nil {
		slog.With("error", err).Error("Failed to create token crypt")
		return fmt.Errorf("failed to create token crypt: %w", err)
	}

	err = encryptExistingTokens(pg, tokenCrypt)
	if err != nil {
		slog.With("error", err).Error("Failed to encrypt existing tokens")
		return fmt.Errorf("failed to encrypt existing tokens: %w", err)
	}

	var openaiClient *openai.Client
	if cfg.OpenAI.APIKey != "" {
		openaiClient = openai.NewClient(cfg.OpenAI.APIKey)
	}

	pluginRegistry := plugin.NewRegistry()
	pluginRegistry.Register(
		counting.NewCountingPlugin(),
		starboard.NewStarboardPlugin(),
	)

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
			PluginInstanceStore:  pg,
			PluginValueStore:     pg,
			PluginRegistry:       pluginRegistry,
			VariableValueStore:   pg,
			ResumePointStore:     pg,
			HttpClient:           engineHTTPClient(cfg),
			OpenaiClient:         openaiClient,
			TokenCrypt:           tokenCrypt,
		},
	)
	engine.Run(ctx)

	handler := event.NewEventHandlerWrapper(engine, pg)

	billingPlans := make([]model.Plan, len(cfg.Billing.Plans))
	for i, plan := range cfg.Billing.Plans {
		billingPlans[i] = model.Plan(plan)
	}

	planManager := plan.NewPlanManager(pg, pg, pg, billingPlans, plan.PlanManagerConfig{
		DiscordBotToken: cfg.Discord.BotToken,
		DiscordGuildID:  cfg.Discord.GuildID,
	})
	planManager.Run(ctx)

	gateway := gateway.NewGatewayManager(pg, pg, planManager, handler, tokenCrypt)
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
	},
		pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg,
		assetStore, gateway, planManager, pluginRegistry, tokenCrypt,
	)
	address := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	if err := apiServer.Serve(ctx, address); err != nil {
		slog.With("error", err).Error("Failed to start API server")
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}

func encryptExistingTokens(pg *postgres.Client, tokenCrypt *util.SymmetricCrypt) error {
	apps, err := pg.AllApps(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get apps: %w", err)
	}

	for _, app := range apps {
		_, err := tokenCrypt.DecryptString(app.DiscordToken)
		if err == nil {
			slog.With("app_id", app.ID).Info("Skipping app with already encrypted token")
			continue
		}

		token, err := tokenCrypt.EncryptString(app.DiscordToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt token: %w", err)
		}
		app.DiscordToken = token

		_, err = pg.UpdateApp(context.Background(), store.AppUpdateOpts{
			ID:             app.ID,
			Name:           app.Name,
			Description:    app.Description,
			DiscordToken:   token,
			DiscordStatus:  app.DiscordStatus,
			Enabled:        app.Enabled,
			DisabledReason: app.DisabledReason,
			UpdatedAt:      app.UpdatedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to update app: %w", err)
		}
	}

	return nil
}
