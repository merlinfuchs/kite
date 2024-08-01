package api

import (
	"github.com/kitecloud/kite/kite-service/internal/api/access"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/app"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/auth"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/command"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/logs"
	"github.com/kitecloud/kite/kite-service/internal/api/handler/user"
	"github.com/kitecloud/kite/kite-service/internal/api/session"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (s *APIServer) RegisterRoutes(
	userStore store.UserStore,
	sessionStore store.SessionStore,
	appStore store.AppStore,
	logStore store.LogStore,
	commandStore store.CommandStore,
) {
	sessionManager := session.NewSessionManager(session.SessionManagerConfig{
		SecureCookies: s.config.SecureCookies,
	}, sessionStore)
	accessManager := access.NewAccessManager(appStore, commandStore)

	// 404 handler
	s.mux.Handle("/", handler.APIHandler(func(c *handler.Context) error {
		return handler.ErrNotFound("unknown_route", "Route not found")
	}))

	v1Group := handler.Group(s.mux, "/v1")

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

	appsGroup := v1Group.Group("/apps", sessionManager.RequireSession)
	appsGroup.Get("/", handler.Typed(appHandler.HandleAppList))
	appsGroup.Post("/", handler.TypedWithBody(appHandler.HandleAppCreate))

	appGroup := appsGroup.Group("/{appID}", accessManager.AppAccess)
	appGroup.Get("/", handler.Typed(appHandler.HandleAppGet))
	appGroup.Patch("/", handler.TypedWithBody(appHandler.HandleAppUpdate))
	appGroup.Put("/token", handler.TypedWithBody(appHandler.HandleAppTokenUpdate))
	appGroup.Delete("/", handler.Typed(appHandler.HandleAppDelete))

	// Log routes
	logHandler := logs.NewLogHandler(logStore)

	logsGroup := appGroup.Group("/logs", accessManager.AppAccess)
	logsGroup.Get("/", handler.Typed(logHandler.HandleLogEntryList))

	// Command routes
	commandsHandler := command.NewCommandHandler(commandStore, s.config.UserLimits.MaxCommandsPerApp)

	commandsGroup := appGroup.Group("/commands")
	commandsGroup.Get("/", handler.Typed(commandsHandler.HandleCommandList))
	commandsGroup.Post("/", handler.TypedWithBody(commandsHandler.HandleCommandCreate))

	commandGroup := commandsGroup.Group("/{commandID}", accessManager.CommandAccess)
	commandGroup.Get("/", handler.Typed(commandsHandler.HandleCommandGet))
	commandGroup.Patch("/", handler.TypedWithBody(commandsHandler.HandleCommandUpdate))
	commandGroup.Delete("/", handler.Typed(commandsHandler.HandleCommandDelete))
}
