package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/config"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
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
