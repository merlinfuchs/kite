package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/rs/cors"
)

type APIServerConfig struct {
	SecureCookies       bool
	AppPublicBaseURL    string
	APIPublicBaseURL    string
	DiscordClientID     string
	DiscordClientSecret string
	UserLimits          APIUserLimitsConfig
}

type APIUserLimitsConfig struct {
	MaxAppsPerUser     int
	MaxCommandsPerApp  int
	MaxVariablesPerApp int
}

type APIServer struct {
	config APIServerConfig
	mux    *http.ServeMux
	server *http.Server
}

func NewAPIServer(
	config APIServerConfig,
	userStore store.UserStore,
	sessionStore store.SessionStore,
	appStore store.AppStore,
	logStore store.LogStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
	variableValueStore store.VariableValueStore,
) *APIServer {
	s := &APIServer{
		config: config,
		mux:    http.NewServeMux(),
	}
	s.RegisterRoutes(
		userStore,
		sessionStore,
		appStore,
		logStore,
		commandStore,
		variableStore,
		variableValueStore,
	)
	return s
}

func (s *APIServer) Serve(ctx context.Context, address string) error {
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{s.config.AppPublicBaseURL},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}).Handler(s.mux)

	s.server = &http.Server{
		Addr:    address,
		Handler: handler,
	}

	slog.With("address", address).Info("Starting API server")
	if err := s.server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func (s *APIServer) Shutdown(ctx context.Context) error {
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
