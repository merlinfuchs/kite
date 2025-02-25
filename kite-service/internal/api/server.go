package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/kitecloud/kite/kite-service/internal/config"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/rs/cors"
)

type APIServerConfig struct {
	SecureCookies       bool
	StrictCookies       bool
	AppPublicBaseURL    string
	APIPublicBaseURL    string
	DiscordClientID     string
	DiscordClientSecret string
	UserLimits          APIUserLimitsConfig
	Billing             BillingConfig
}

type APIUserLimitsConfig struct {
	MaxAppsPerUser          int
	MaxCommandsPerApp       int
	MaxVariablesPerApp      int
	MaxMessagesPerApp       int
	MaxEventListenersPerApp int
	MaxAssetSize            int
}

type BillingConfig struct {
	LemonSqueezyAPIKey        string
	LemonSqueezySigningSecret string
	LemonSqueezyStoreID       string
	TestMode                  bool
	Plans                     []config.BillingPlanConfig
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
	usageStore store.UsageStore,
	commandStore store.CommandStore,
	variableStore store.VariableStore,
	variableValueStore store.VariableValueStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	eventListenerStore store.EventListenerStore,
	subscriptionStore store.SubscriptionStore,
	entitlementStore store.EntitlementStore,
	assetStore store.AssetStore,
	appStateManager store.AppStateManager,
	planManager *plan.PlanManager,
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
		usageStore,
		commandStore,
		variableStore,
		variableValueStore,
		messageStore,
		messageInstanceStore,
		eventListenerStore,
		subscriptionStore,
		entitlementStore,
		assetStore,
		appStateManager,
		planManager,
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
