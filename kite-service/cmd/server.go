package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	disapi "github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
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
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"github.com/kitecloud/kite/kite-service/pkg/plugin/counting"
	"github.com/kitecloud/kite/kite-service/pkg/plugin/starboard"
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
	},
		pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg, pg,
		assetStore, gateway, planManager, pluginRegistry,
	)
	address := fmt.Sprintf("%s:%d", cfg.API.Host, cfg.API.Port)
	if err := apiServer.Serve(ctx, address); err != nil {
		slog.With("error", err).Error("Failed to start API server")
		return fmt.Errorf("failed to start API server: %w", err)
	}

	return nil
}

func patchDiscordProxyURL(cfg *config.Config) {
	if cfg.Discord.ProxyURL == "" {
		return
	}

	slog.Info("Using Proxy for Discord API", "url", cfg.Discord.ProxyURL)

	httputil.Retries = 10

	disapi.BaseEndpoint = cfg.Discord.ProxyURL
	disapi.Endpoint = disapi.BaseEndpoint + disapi.Path + "/"
	disapi.EndpointGateway = disapi.Endpoint + "gateway"
	disapi.EndpointGatewayBot = disapi.EndpointGateway + "/bot"
	disapi.EndpointApplications = disapi.Endpoint + "applications/"
	disapi.EndpointChannels = disapi.Endpoint + "channels/"
	disapi.EndpointGuilds = disapi.Endpoint + "guilds/"
	disapi.EndpointUsers = disapi.Endpoint + "users/"
	disapi.EndpointWebhooks = disapi.Endpoint + "webhooks/"
	disapi.EndpointInvites = disapi.Endpoint + "invites/"
	disapi.EndpointInteractions = disapi.Endpoint + "interactions/"
	disapi.EndpointStageInstances = disapi.Endpoint + "stage-instances/"
	disapi.EndpointMe = disapi.Endpoint + "users/@me"
	disapi.EndpointAuth = disapi.Endpoint + "auth/"
	disapi.EndpointLogin = disapi.EndpointAuth + "login"
	disapi.EndpointTOTP = disapi.EndpointAuth + "mfa/totp"
}

func engineHTTPClient(cfg *config.Config) *http.Client {
	if cfg.Engine.HTTPProxyURL != "" {
		proxyURL, err := url.Parse(cfg.Engine.HTTPProxyURL)
		if err != nil {
			slog.With("error", err).Error("Failed to parse proxy URL")
			return nil
		}

		slog.Info("Using HTTP proxy for Engine", "url", cfg.Engine.HTTPProxyURL)

		return &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}
	}

	return &http.Client{}
}
