package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
)

func (h *AuthHandler) HandleAuthRedirect(c *fiber.Ctx) error {
	state := setOauthStateCookie(c)
	return c.Redirect(h.oauth2Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) HandleAuthCallback(c *fiber.Ctx) error {
	state := getOauthStateCookie(c)
	if state == "" || c.Query("state") != state {
		slog.Error("Failed to login: Invalid state")
		// TODO: redirect to error page, same for all following errors
		return h.HandleAuthRedirect(c)
	}

	token, err := h.oauth2Config.Exchange(c.Context(), c.Query("code"))
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to exchange token")
		return h.HandleAuthRedirect(c)
	}

	client := h.oauth2Config.Client(c.Context(), token)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to get user info")
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
		return h.HandleAuthRedirect(c)
	}

	guilds := []struct {
		ID string `json:"id"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&guilds)
	if err != nil {
		slog.With(logattr.Error(err)).Error("Failed to decode guilds")
		return h.HandleAuthRedirect(c)
	}
	resp.Body.Close()

	guildIDs := make([]string, len(guilds))
	for i, guild := range guilds {
		guildIDs[i] = guild.ID
	}

	err = h.sessionManager.CreateSessionCookie(c, model.SessionTypeWebApp, user.ID, guildIDs, token.AccessToken)
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
