package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/internal/config"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/ravener/discord-oauth2"
	"golang.org/x/oauth2"
	"gopkg.in/guregu/null.v4"
)

type AuthHandler struct {
	sessionManager  *session.SessionManager
	userStore       store.UserStore
	oauth2Config    *oauth2.Config
	cliOauth2Config *oauth2.Config
	cfg             *config.ServerConfig
}

func New(sessionManager *session.SessionManager, userStore store.UserStore, cfg *config.ServerConfig) *AuthHandler {
	return &AuthHandler{
		sessionManager: sessionManager,
		userStore:      userStore,
		oauth2Config: &oauth2.Config{
			RedirectURL:  cfg.AuthCallbackURL(),
			ClientID:     cfg.Discord.ClientID,
			ClientSecret: cfg.Discord.ClientSecret,
			Scopes:       []string{discord.ScopeIdentify, discord.ScopeEmail},
			Endpoint:     discord.Endpoint,
		},
		cliOauth2Config: &oauth2.Config{
			RedirectURL:  cfg.AuthCLICallbackURL(),
			ClientID:     cfg.Discord.ClientID,
			ClientSecret: cfg.Discord.ClientSecret,
			Scopes:       []string{discord.ScopeIdentify, discord.ScopeEmail},
			Endpoint:     discord.Endpoint,
		},
		cfg: cfg,
	}
}

func (h *AuthHandler) HandleAuthInviteRedirect(c *fiber.Ctx) error {
	oauth2Config := *h.oauth2Config
	oauth2Config.Scopes = append(oauth2Config.Scopes, discord.ScopeBot, discord.ScopeApplicationsCommands)

	state := setOauthStateCookie(c)
	url := oauth2Config.AuthCodeURL(state)
	// TODO: add permissions?

	return c.Redirect(url, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleAuthRedirect(c *fiber.Ctx) error {
	state := setOauthStateCookie(c)
	return c.Redirect(h.oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleAuthCallback(c *fiber.Ctx) error {
	state := getOauthStateCookie(c)
	if state == "" || c.Query("state") != state {
		slog.Error("Failed to login: Invalid state")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}

	accessToken, userID, err := h.ExchangeAccessToken(c.Context(), h.oauth2Config, c.Query("code"))
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to exchange access token")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}

	err = h.sessionManager.CreateSessionCookie(c, model.SessionTypeWebApp, userID, accessToken)
	if err != nil {
		return err
	}

	return c.Redirect(h.cfg.App.AuthCallbackURL()+"/app", http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleAuthLogout(c *fiber.Ctx) error {
	err := h.sessionManager.DeleteSession(c)
	if err != nil {
		return err
	}

	return c.Redirect(h.cfg.App.AuthCallbackURL(), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) ExchangeAccessToken(ctx context.Context, oauth2 *oauth2.Config, code string) (string, distype.Snowflake, error) {
	token, err := oauth2.Exchange(ctx, code)
	if err != nil {
		return "", "", fmt.Errorf("Failed to exchange token: %v", err)
	}

	client := oauth2.Client(ctx, token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return "", "", fmt.Errorf("Failed to get user info: %v", err)
	}

	user := struct {
		ID            distype.Snowflake `json:"id"`
		Username      string            `json:"username"`
		Email         string            `json:"email"`
		GlobalName    null.String       `json:"global_name"`
		Discriminator string            `json:"discriminator"`
		Avatar        null.String       `json:"avatar"`
		PublicFlags   int               `json:"public_flags"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		return "", "", fmt.Errorf("Failed to decode user info: %v", err)
	}
	resp.Body.Close()

	if user.Email == "" {
		return "", "", fmt.Errorf("User email is empty")
	}

	err = h.userStore.UpsertUser(ctx, &model.User{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Discriminator: null.NewString(user.Discriminator, user.Discriminator != "0"),
		GlobalName:    user.GlobalName,
		Avatar:        user.Avatar,
		PublicFlags:   user.PublicFlags,
	})
	if err != nil {
		return "", "", fmt.Errorf("Failed to upsert user: %v", err)
	}

	return token.AccessToken, distype.Snowflake(user.ID), nil
}

func getOauthStateCookie(c *fiber.Ctx) string {
	state := c.Cookies("kite_oauth_state")
	c.ClearCookie("kite_oauth_state")
	return state
}

func setOauthStateCookie(c *fiber.Ctx) string {
	b := make([]byte, 128)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.Cookie(&fiber.Cookie{
		Name:     "kite_oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   true,
	})
	return state
}
