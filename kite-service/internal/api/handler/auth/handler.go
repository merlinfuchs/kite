package auth

import (
	"fmt"
	"net/http"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/session"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/ravener/discord-oauth2"
	"golang.org/x/oauth2"
)

type AuthHandlerConfig struct {
	SecureCookies       bool
	AppPublicBaseURL    string
	APIPublicBaseURL    string
	DiscordClientID     string
	DiscordClientSecret string
}

type AuthHandler struct {
	config         AuthHandlerConfig
	userStore      store.UserStore
	sessionManager *session.SessionManager
	oauth2Config   *oauth2.Config
}

func NewAuthHandler(config AuthHandlerConfig, userStore store.UserStore, sessionManager *session.SessionManager) *AuthHandler {
	conf := &oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/v1/auth/login/callback", config.APIPublicBaseURL),
		ClientID:     config.DiscordClientID,
		ClientSecret: config.DiscordClientSecret,
		Scopes:       []string{discord.ScopeIdentify, discord.ScopeEmail},
		Endpoint:     discord.Endpoint,
	}

	return &AuthHandler{
		config:         config,
		userStore:      userStore,
		sessionManager: sessionManager,
		oauth2Config:   conf,
	}
}

func (h *AuthHandler) HandleAuthLogin(c *handler.Context) error {
	state := h.setOauthStateCookie(c)
	h.setOauthRedirectCookie(c)
	c.Redirect(h.oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
	return nil
}

func (h *AuthHandler) HandleAuthLoginCallback(c *handler.Context) error {
	state := h.getOauthStateCookie(c)
	if state == "" || c.Query("state") != state {
		return fmt.Errorf("invalid state")
	}

	_, _, err := h.authenticateWithCode(c, c.Query("code"))
	if err != nil {
		return fmt.Errorf("failed to authenticate with code: %w", err)
	}

	redirectURL := h.getOauthRedirectURL(c)
	c.Redirect(redirectURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *AuthHandler) HandleAuthLogout(c *handler.Context) (*wire.AuthLogoutResponse, error) {
	if err := h.sessionManager.DeleteSession(c); err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}
	return &wire.Empty{}, nil
}
