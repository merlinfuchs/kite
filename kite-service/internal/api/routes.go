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
	"github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	eventlistener "github.com/kitecloud/kite/kite-service/internal/api/handler/event_listener"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/logs"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/message"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/usage"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/user"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/variable"
	"github.com/kitecloud/kite/kite-service/internal/api/session"
	"github.com/kitecloud/kite/kite-service/internal/store"
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
	assetStore store.AssetStore,
	appStateManager store.AppStateManager,
) {
	sessionManager := session.NewSessionManager(session.SessionManagerConfig{
		StrictCookies: s.config.StrictCookies,
		SecureCookies: s.config.SecureCookies,
	}, sessionStore)
	accessManager := access.NewAccessManager(appStore, commandStore, variableStore, messageStore, eventListenerStore)

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

	v1Group := handler.Group(s.mux, "/v1")

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
	appHandler := app.NewAppHandler(appStore, s.config.UserLimits.MaxAppsPerUser)

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
	appGroup.Get("/entities", handler.Typed(appHandler.HandleAppEntityList))

	// Log routes
	logHandler := logs.NewLogHandler(logStore)

	logsGroup := appGroup.Group("/logs", accessManager.AppAccess)
	logsGroup.Get("/", handler.Typed(logHandler.HandleLogEntryList))
	logsGroup.Get("/summary", handler.Typed(logHandler.HandleLogSummaryGet))

	// Usage routes
	usageHandler := usage.NewUsageHandler(usageStore, s.config.UserLimits.CreditsPerMonth)

	usageGroup := appGroup.Group("/usage")
	usageGroup.Get("/credits", handler.Typed(usageHandler.HandleUsageCreditsGet))
	usageGroup.Get("/by-day", handler.Typed(usageHandler.HandleUsageByDayList))
	usageGroup.Get("/by-type", handler.Typed(usageHandler.HandleUsageByTypeList))

	// Command routes
	commandsHandler := command.NewCommandHandler(commandStore, s.config.UserLimits.MaxCommandsPerApp)

	commandsGroup := appGroup.Group("/commands")
	commandsGroup.Get("/", handler.Typed(commandsHandler.HandleCommandList))
	commandsGroup.Post("/", handler.TypedWithBody(commandsHandler.HandleCommandCreate))

	commandGroup := commandsGroup.Group("/{commandID}", accessManager.CommandAccess)
	commandGroup.Get("/", handler.Typed(commandsHandler.HandleCommandGet))
	commandGroup.Patch("/", handler.TypedWithBody(commandsHandler.HandleCommandUpdate))
	commandGroup.Delete("/", handler.Typed(commandsHandler.HandleCommandDelete))

	// Event listener routes
	eventListenerHandler := eventlistener.NewEventListenerHandler(eventListenerStore, s.config.UserLimits.MaxEventListenersPerApp)

	eventListenersGroup := appGroup.Group("/event-listeners")
	eventListenersGroup.Get("/", handler.Typed(eventListenerHandler.HandleEventListenerList))
	eventListenersGroup.Post("/", handler.TypedWithBody(eventListenerHandler.HandleEventListenerCreate))

	eventListenerGroup := eventListenersGroup.Group("/{listenerID}", accessManager.EventListenerAccess)
	eventListenerGroup.Get("/", handler.Typed(eventListenerHandler.HandleEventListenerGet))
	eventListenerGroup.Patch("/", handler.TypedWithBody(eventListenerHandler.HandleEventListenerUpdate))
	eventListenerGroup.Delete("/", handler.Typed(eventListenerHandler.HandleEventListenerDelete))

	// Variable routes
	variablesHandler := variable.NewVariableHandler(variableStore, variableValueStore, s.config.UserLimits.MaxVariablesPerApp)

	variablesGroup := appGroup.Group("/variables")
	variablesGroup.Get("/", handler.Typed(variablesHandler.HandleVariableList))
	variablesGroup.Post("/", handler.TypedWithBody(variablesHandler.HandleVariableCreate))

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
		s.config.UserLimits.MaxMessagesPerApp,
	)

	messagesGroup := appGroup.Group("/messages")
	messagesGroup.Get("/", handler.Typed(messageHandler.HandleMessageList))
	messagesGroup.Post("/", handler.TypedWithBody(messageHandler.HandleMessageCreate))

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
