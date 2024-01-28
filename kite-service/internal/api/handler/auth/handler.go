package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/ravener/discord-oauth2"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	sessionManager *session.SessionManager
	userStore      store.UserStore
	oauth2Config   *oauth2.Config
	cfg            *config.ServerConfig
}

func New(sessionManager *session.SessionManager, userStore store.UserStore, cfg *config.ServerConfig) *AuthHandler {
	conf := &oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/v1/auth/callback", cfg.PublicURL),
		ClientID:     cfg.Discord.ClientID,
		ClientSecret: cfg.Discord.ClientSecret,
		Scopes:       []string{discord.ScopeIdentify, discord.ScopeGuilds},
		Endpoint:     discord.Endpoint,
	}

	return &AuthHandler{
		sessionManager: sessionManager,
		userStore:      userStore,
		oauth2Config:   conf,
		cfg:            cfg,
	}
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

	token, err := h.oauth2Config.Exchange(c.Context(), c.Query("code"))
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to exchange token")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}

	client := h.oauth2Config.Client(c.Context(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to get user info")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}

	user := struct {
		ID            string `json:"id"`
		Username      string `json:"username"`
		GlobalName    string `json:"global_name"`
		Discriminator string `json:"discriminator"`
		Avatar        string `json:"avatar"`
		PublicFlags   int    `json:"public_flags"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&user)
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to decode user info")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}
	resp.Body.Close()

	err = h.userStore.UpsertUser(c.Context(), &model.User{
		ID:            user.ID,
		Username:      user.Username,
		Discriminator: user.Discriminator,
		GlobalName:    user.GlobalName,
		Avatar:        user.Avatar,
		PublicFlags:   user.PublicFlags,
	})
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to upsert user")
		return err
	}

	resp, err = client.Get("https://discord.com/api/users/@me/guilds")
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to get guilds")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}

	guilds := []struct {
		ID string `json:"id"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&guilds)
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to decode guilds")
		// TODO: redirect to error page
		return h.HandleAuthRedirect(c)
	}
	resp.Body.Close()

	guildIDs := make([]string, len(guilds))
	for i, guild := range guilds {
		guildIDs[i] = guild.ID
	}

	err = h.sessionManager.CreateSessionCookie(c, user.ID, guildIDs, token.AccessToken)
	if err != nil {
		return err
	}

	return c.Redirect(h.cfg.AppPublicURL, http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleAuthLogout(c *fiber.Ctx) error {
	err := h.sessionManager.DeleteSession(c)
	if err != nil {
		return err
	}

	return c.Redirect(h.cfg.AppPublicURL, http.StatusTemporaryRedirect)
}

func getOauthStateCookie(c *fiber.Ctx) string {
	state := c.Cookies("oauth_state")
	c.ClearCookie("oauth_state")
	return state
}

func setOauthStateCookie(c *fiber.Ctx) string {
	b := make([]byte, 128)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HTTPOnly: true,
		Secure:   true,
	})
	return state
}
