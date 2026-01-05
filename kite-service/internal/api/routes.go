package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/access"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/app"
	appstate "github.com/kitecloud/kite/kite-service/internal/api/handler/app_state"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/asset"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/auth"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/billing"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	eventlistener "github.com/kitecloud/kite/kite-service/internal/api/handler/event_listener"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/logs"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/message"
	pluginhandler "github.com/kitecloud/kite/kite-service/internal/api/handler/plugin"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/usage"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/user"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/variable"
	"github.com/kitecloud/kite/kite-service/internal/api/session"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	kiteweb "github.com/merlinfuchs/kite/kite-web"
)

func (s *APIServer) RegisterRoutes(
	userStore store.UserStore,
	sessionStore store.SessionStore,
	appStore store.AppStore,
	logStore store.LogStore,
	usageStore store.UsageStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
	variableValueStore store.VariableValueStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	eventListenerStore store.EventListenerStore,
	pluginInstanceStore store.PluginInstanceStore,
	subscriptionStore store.SubscriptionStore,
	entitlementStore store.EntitlementStore,
	assetStore store.AssetStore,
	appStateManager store.AppStateManager,
	planManager *plan.PlanManager,
	pluginRegistry *plugin.Registry,
	tokenCrypt *util.SymmetricCrypt,
) {
	sessionManager := session.NewSessionManager(session.SessionManagerConfig{
		StrictCookies: s.config.StrictCookies,
		SecureCookies: s.config.SecureCookies,
	}, sessionStore)
	accessManager := access.NewAccessManager(
		appStore,
		commandStore,
		variableStore,
		messageStore,
		eventListenerStore,
		pluginInstanceStore,
		planManager,
	)

	cacheManager, err := handler.NewCacheManager(10000)
	if err != nil {
		panic(err)
	}

	webHandler, err := kiteweb.NewHandler()
	if err == nil {
		slog.Info("Website embedded")
		s.mux.Handle("/", webHandler)
	} else {
		// 404 handler
		slog.Info("Website not embedded, set 'embedweb' build tag to embed it.")
		s.mux.Handle("/", handler.APIHandler(func(c *handler.Context) error {
			return handler.ErrNotFound("unknown_route", "Route not found")
		}))
	}

	v1Group := handler.Group(
		s.mux, "/v1",
		handler.ClusterIndexProvider(s.config.ClusterCount, s.config.ClusterIndex),
	)

	v1Group.Get("/health", func(c *handler.Context) error {
		return c.Send(http.StatusOK, []byte("OK"))
	})

	// Auth routes
	authHandler := auth.NewAuthHandler(auth.AuthHandlerConfig{
		SecureCookies:       s.config.SecureCookies,
		AppPublicBaseURL:    s.config.AppPublicBaseURL,
		APIPublicBaseURL:    s.config.APIPublicBaseURL,
		DiscordClientID:     s.config.DiscordClientID,
		DiscordClientSecret: s.config.DiscordClientSecret,
	}, userStore, sessionManager)

	authGroup := v1Group.Group("/auth")
	authGroup.Get("/login", authHandler.HandleAuthLogin)
	authGroup.Get("/login/callback", authHandler.HandleAuthLoginCallback)
	authGroup.Post("/logout", handler.Typed(authHandler.HandleAuthLogout))

	// User routes
	userHandler := user.NewUserHandler(userStore)

	usersGroup := v1Group.Group("/users", sessionManager.RequireSession)
	usersGroup.Get("/{userID}", handler.Typed(userHandler.HandlerUserGet))

	// App routes
	appHandler := app.NewAppHandler(
		appStore,
		userStore,
		s.config.UserLimits.MaxAppsPerUser,
		tokenCrypt,
	)

	appsGroup := v1Group.Group("/apps",
		sessionManager.RequireSession,
		handler.RateLimitByUser(60, time.Minute),
	)
	appsGroup.Get("/", handler.Typed(appHandler.HandleAppList))
	appsGroup.Post("/", handler.TypedWithBody(appHandler.HandleAppCreate))

	appGroup := appsGroup.Group("/{appID}", accessManager.AppAccess)
	appGroup.Get("/", handler.Typed(appHandler.HandleAppGet))
	appGroup.Put("/",
		handler.TypedWithBody(appHandler.HandleAppUpdate),
		handler.RateLimitByUser(2, time.Minute),
	)
	appGroup.Put("/status", handler.TypedWithBody(appHandler.HandleAppStatusUpdate))
	appGroup.Put("/token",
		handler.TypedWithBody(appHandler.HandleAppTokenUpdate),
		handler.RateLimitByUser(2, time.Minute),
	)
	appGroup.Delete("/", handler.Typed(appHandler.HandleAppDelete))
	appGroup.Get("/emojis",
		handler.Typed(appHandler.HandleAppEmojisList),
		handler.CacheByUser(cacheManager, time.Minute),
	)
	appGroup.Get("/entities", handler.Typed(appHandler.HandleAppEntityList))
	appGroup.Get("/collaborators", handler.Typed(appHandler.HandleAppCollaboratorsList))
	appGroup.Post("/collaborators", handler.TypedWithBody(appHandler.HandleAppCollaboratorCreate))
	appGroup.Delete("/collaborators/{userID}", handler.Typed(appHandler.HandleAppCollaboratorDelete))

	// Billing routes
	billingHandler := billing.NewBillingHandler(billing.BillingHandlerConfig{
		LemonSqueezyAPIKey:        s.config.Billing.LemonSqueezyAPIKey,
		LemonSqueezySigningSecret: s.config.Billing.LemonSqueezySigningSecret,
		LemonSqueezyStoreID:       s.config.Billing.LemonSqueezyStoreID,
		TestMode:                  s.config.Billing.TestMode,
		AppPublicBaseURL:          s.config.AppPublicBaseURL,
	}, userStore, subscriptionStore, entitlementStore, planManager)

	v1Group.Post("/billing/webhook", handler.TypedWithBody(billingHandler.HandleBillingWebhook))
	v1Group.Get("/billing/plans", handler.Typed(billingHandler.HandleBillingPlanList))

	userBillingGroup := v1Group.Group("/billing", sessionManager.RequireSession)
	userBillingGroup.Post("/subscriptions/{subscriptionID}/manage", handler.Typed(billingHandler.HandleSubscriptionManage))

	appBillingGroup := appGroup.Group("/billing")
	appBillingGroup.Get("/subscriptions", handler.Typed(billingHandler.HandleAppSubscriptionList))
	appBillingGroup.Post("/checkout", handler.TypedWithBody(billingHandler.HandleAppCheckout))
	appBillingGroup.Get("/features", handler.Typed(billingHandler.HandleFeaturesGet))

	// Log routes
	logHandler := logs.NewLogHandler(logStore)

	logsGroup := appGroup.Group("/logs", accessManager.AppAccess)
	logsGroup.Get("/", handler.Typed(logHandler.HandleLogEntryList))
	logsGroup.Get("/summary", handler.Typed(logHandler.HandleLogSummaryGet))

	// Usage routes
	usageHandler := usage.NewUsageHandler(usageStore)

	usageGroup := appGroup.Group("/usage")
	usageGroup.Get("/credits", handler.Typed(usageHandler.HandleUsageCreditsGet))
	usageGroup.Get("/by-day", handler.Typed(usageHandler.HandleUsageByDayList))
	usageGroup.Get("/by-type", handler.Typed(usageHandler.HandleUsageByTypeList))

	// Command routes
	commandsHandler := command.NewCommandHandler(commandStore)

	commandsGroup := appGroup.Group("/commands")
	commandsGroup.Get("/", handler.Typed(commandsHandler.HandleCommandList))
	commandsGroup.Post("/", handler.TypedWithBody(commandsHandler.HandleCommandCreate))
	commandsGroup.Post("/import", handler.TypedWithBody(commandsHandler.HandleCommandsImport))

	commandGroup := commandsGroup.Group("/{commandID}", accessManager.CommandAccess)
	commandGroup.Get("/", handler.Typed(commandsHandler.HandleCommandGet))
	commandGroup.Patch("/", handler.TypedWithBody(commandsHandler.HandleCommandUpdate))
	commandGroup.Delete("/", handler.Typed(commandsHandler.HandleCommandDelete))
	commandGroup.Put("/enabled", handler.TypedWithBody(commandsHandler.HandleCommandUpdateEnabled))

	// Event listener routes
	eventListenerHandler := eventlistener.NewEventListenerHandler(eventListenerStore)

	eventListenersGroup := appGroup.Group("/event-listeners")
	eventListenersGroup.Get("/", handler.Typed(eventListenerHandler.HandleEventListenerList))
	eventListenersGroup.Post("/", handler.TypedWithBody(eventListenerHandler.HandleEventListenerCreate))
	eventListenersGroup.Post("/import", handler.TypedWithBody(eventListenerHandler.HandleEventListenersImport))

	eventListenerGroup := eventListenersGroup.Group("/{listenerID}", accessManager.EventListenerAccess)
	eventListenerGroup.Get("/", handler.Typed(eventListenerHandler.HandleEventListenerGet))
	eventListenerGroup.Patch("/", handler.TypedWithBody(eventListenerHandler.HandleEventListenerUpdate))
	eventListenerGroup.Delete("/", handler.Typed(eventListenerHandler.HandleEventListenerDelete))
	eventListenerGroup.Put("/enabled", handler.TypedWithBody(eventListenerHandler.HandleEventListenerUpdateEnabled))

	// Plugin instance routes
	pluginHandler := pluginhandler.NewPluginHandler(pluginRegistry, pluginInstanceStore)

	pluginsGroup := v1Group.Group("/plugins")
	pluginsGroup.Get("/", handler.Typed(pluginHandler.HandlePluginList))

	pluginInstancesGroup := appGroup.Group("/plugins")
	pluginInstancesGroup.Get("/", handler.Typed(pluginHandler.HandlePluginInstanceList))
	pluginInstancesGroup.Post("/", handler.TypedWithBody(pluginHandler.HandlePluginInstanceCreate))

	pluginInstanceGroup := pluginInstancesGroup.Group("/{pluginID}", accessManager.PluginInstanceAccess)
	pluginInstanceGroup.Get("/", handler.Typed(pluginHandler.HandlePluginInstanceGet))
	pluginInstanceGroup.Patch("/", handler.TypedWithBody(pluginHandler.HandlePluginInstanceUpdate))
	pluginInstanceGroup.Delete("/", handler.Typed(pluginHandler.HandlePluginInstanceDelete))
	pluginInstanceGroup.Put("/enabled", handler.TypedWithBody(pluginHandler.HandlePluginInstanceUpdateEnabled))

	// Variable routes
	variablesHandler := variable.NewVariableHandler(variableStore, variableValueStore)

	variablesGroup := appGroup.Group("/variables")
	variablesGroup.Get("/", handler.Typed(variablesHandler.HandleVariableList))
	variablesGroup.Post("/", handler.TypedWithBody(variablesHandler.HandleVariableCreate))
	variablesGroup.Post("/import", handler.TypedWithBody(variablesHandler.HandleVariablesImport))

	variableGroup := variablesGroup.Group("/{variableID}", accessManager.VariableAccess)
	variableGroup.Get("/", handler.Typed(variablesHandler.HandleVariableGet))
	variableGroup.Patch("/", handler.TypedWithBody(variablesHandler.HandleVariableUpdate))
	variableGroup.Delete("/", handler.Typed(variablesHandler.HandleVariableDelete))

	// Message routes
	messageHandler := message.NewMessageHandler(
		messageStore,
		messageInstanceStore,
		assetStore,
		appStateManager,
	)

	messagesGroup := appGroup.Group("/messages")
	messagesGroup.Get("/", handler.Typed(messageHandler.HandleMessageList))
	messagesGroup.Post("/", handler.TypedWithBody(messageHandler.HandleMessageCreate))
	messagesGroup.Post("/import", handler.TypedWithBody(messageHandler.HandleMessagesImport))

	messageGroup := messagesGroup.Group("/{messageID}", accessManager.MessageAccess)
	messageGroup.Get("/", handler.Typed(messageHandler.HandleMessageGet))
	messageGroup.Patch("/", handler.TypedWithBody(messageHandler.HandleMessageUpdate))
	messageGroup.Delete("/", handler.Typed(messageHandler.HandleMessageDelete))
	messageGroup.Get("/instances", handler.Typed(messageHandler.HandleMessageInstanceList))
	messageGroup.Post("/instances", handler.TypedWithBody(messageHandler.HandleMessageInstanceCreate))
	messageGroup.Put("/instances/{instanceID}", handler.Typed(messageHandler.HandleMessageInstanceUpdate))
	messageGroup.Delete("/instances/{instanceID}", handler.Typed(messageHandler.HandleMessageInstanceDelete))

	// Asset routes
	assetHandler := asset.NewAssetHandler(assetStore, asset.AssetHandlerConfig{
		APIPublicBaseURL: s.config.APIPublicBaseURL,
		MaxAssetSize:     int64(s.config.UserLimits.MaxAssetSize),
	})

	assetsGroup := appGroup.Group("/assets")
	assetsGroup.Post("/", handler.Typed(assetHandler.HandleAssetCreate))
	assetsGroup.Get("/{assetID}", handler.Typed(assetHandler.HandleAssetGet))
	v1Group.Get(
		"/assets/{assetID}",
		assetHandler.HandleAssetDownload,
		sessionManager.OptionalSession,
	)

	// State routes
	stateHandler := appstate.NewAppStateHandler(appStateManager)

	stateGroup := appGroup.Group("/state")
	stateGroup.Get("/", handler.Typed(stateHandler.HandleStateStatusGet))
	stateGroup.Get("/guilds", handler.Typed(stateHandler.HandleStateGuildList))
	stateGroup.Delete("/guilds/{guildID}", handler.Typed(stateHandler.HandleStateGuildLeave))
	stateGroup.Get("/guilds/{guildID}/channels", handler.Typed(stateHandler.HandleStateGuildChannelList))
}
