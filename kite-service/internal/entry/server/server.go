package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api"
	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/command"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/core/event"
	"github.com/kitecloud/kite/kite-service/internal/core/gateway"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/core/usage"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres"
	"github.com/kitecloud/kite/kite-service/internal/db/s3"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/plugin/counting"
	"github.com/kitecloud/kite/kite-service/pkg/plugin/starboard"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func StartServer(c context.Context, cfg *config.Config) error {
	patchDiscordProxyURL(cfg)

	pg, err := postgres.New(postgres.BuildConnectionDSN(cfg.Database.Postgres), cfg.ClusterCount)
	if err != nil {
		slog.With("error", err).Error("Failed to create postgres client")
		return fmt.Errorf("failed to create postgres client: %w", err)
	}

	// S3 is optional. Without it Kite runs fine, just without support for
	// assets, so a failure here must not stop the server from starting.
	var objectStore store.ObjectStore = store.DisabledObjectStore{}
	s3Client, err := s3.New(cfg.Database.S3)
	if err != nil {
		slog.With("error", err).Warn("Failed to create S3 client, continuing without support for assets")
	} else {
		objectStore = s3Client
	}

	// Bounded: this is the first call that actually reaches S3, and it blocks
	// startup. Without a deadline a configured but unreachable endpoint stalls
	// the whole server for minutes on minio's dial timeout and retries.
	assetCtx, cancelAssetCtx := context.WithTimeout(context.Background(), 10*time.Second)
	assetStore, err := postgres.NewAssetStore(assetCtx, pg, objectStore)
	cancelAssetCtx()
	if err != nil && !errors.Is(err, store.ErrObjectStoreDisabled) {
		slog.With("error", err).Warn("Failed to create asset store, continuing without support for assets")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configureHTTPTransport(cfg)
	warnUnverifiedBillingWebhook(cfg)
	startDebugServer(ctx, cfg.Debug)

	tokenCrypt, err := util.NewSymmetricCrypt(cfg.Encryption.TokenEncryptionKey)
	if err != nil {
		slog.With("error", err).Error("Failed to create token crypt")
		return fmt.Errorf("failed to create token crypt: %w", err)
	}

	var openaiClient openai.Client
	if cfg.OpenAI.APIKey != "" {
		openaiClient = openai.NewClient(option.WithAPIKey(cfg.OpenAI.APIKey))
	}

	pluginRegistry := plugin.NewRegistry()
	pluginRegistry.Register(
		counting.NewCountingPlugin(),
		starboard.NewStarboardPlugin(),
	)

	engine := engine.NewEngine(
		engine.Env{
			Config: engine.EngineConfig{
				MaxStackDepth:          cfg.Engine.MaxStackDepth,
				MaxOperations:          cfg.Engine.MaxOperations,
				MaxCredits:             cfg.Engine.MaxCredits,
				ClusterCount:           cfg.ClusterCount,
				ClusterIndex:           cfg.ClusterIndex,
				PopulateInterval:       cfg.Engine.PopulateInterval,
				RemoveDanglingInterval: cfg.Engine.RemoveDanglingInterval,
				PopulateOverlap:        cfg.Engine.PopulateOverlap,
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
			OpenaiClient:         &openaiClient,
			TokenCrypt:           tokenCrypt,
		},
	)
	engine.Run(ctx)

	commandManager := command.NewCommandManager(pg, pg, pg, pluginRegistry, tokenCrypt)

	handler := event.NewEventHandlerWrapper(engine, pg)

	planManager := plan.NewPlanManager(pg, pg, pg, plan.PlansFromConfig(cfg.Billing.Plans), plan.PlanManagerConfig{
		DiscordBotToken: cfg.Discord.BotToken,
		DiscordGuildID:  cfg.Discord.GuildID,
	})

	gateway := gateway.NewGatewayManager(pg, pg, planManager, handler, tokenCrypt, pluginRegistry, gateway.GatewayManagerConfig{
		ClusterCount:           cfg.ClusterCount,
		ClusterIndex:           cfg.ClusterIndex,
		PopulateInterval:       cfg.Gateway.PopulateInterval,
		RemoveDanglingInterval: cfg.Gateway.RemoveDanglingInterval,
		PopulateOverlap:        cfg.Gateway.PopulateOverlap,
		StartInterval:          cfg.Gateway.StartInterval,
	})
	gateway.Run(ctx)

	usage := usage.NewUsageManager(pg, pg, pg, planManager)

	if cfg.IsPrimaryCluster() {
		planManager.Run(ctx)
		usage.Run(ctx)
	}

	apiServer := api.NewAPIServer(api.APIServerConfig{
		ClusterCount:        cfg.ClusterCount,
		ClusterIndex:        cfg.ClusterIndex,
		SecureCookies:       cfg.API.SecureCookies,
		StrictCookies:       cfg.API.StrictCookies,
		APIPublicBaseURL:    cfg.API.PublicBaseURL,
		AppPublicBaseURL:    cfg.App.PublicBaseURL,
		DiscordClientID:     cfg.Discord.ClientID,
		DiscordClientSecret: cfg.Discord.ClientSecret,
		UserLimits: api.APIUserLimitsConfig{
			MaxAppsPerUser: cfg.UserLimits.MaxAppsPerUser,
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
		assetStore, gateway, planManager, pluginRegistry, tokenCrypt, commandManager,
	)
	address := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	if err := apiServer.Serve(ctx, address); err != nil {
		slog.With("error", err).Error("Failed to start API server")
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}
