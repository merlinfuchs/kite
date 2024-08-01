package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

func (h *AuthHandler) authenticateWithCode(c *handler.Context, code string) (string, *model.Session, error) {
	tokenData, err := h.oauth2Config.Exchange(c.Context(), code)
	if err != nil {
		return "", nil, fmt.Errorf("Failed to exchange token: %w", err)
	}

	client := h.oauth2Config.Client(c.Context(), tokenData)
	resp, err := client.Get("https://discord.com/api/users/@me")
	if err != nil {
		return "", nil, fmt.Errorf("Failed to get user info: %w", err)
	}

	discordUser := struct {
		ID            string      `json:"id"`
		Username      string      `json:"username"`
		Discriminator string      `json:"discriminator"`
		GlobalName    string      `json:"global_name"`
		Avatar        null.String `json:"avatar"`
		Email         string      `json:"email"`
		Verified      bool        `json:"verified"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&discordUser)
	if err != nil {
		return "", nil, fmt.Errorf("Failed to decode user info: %w", err)
	}
	resp.Body.Close()

	if discordUser.Email == "" {
		return "", nil, fmt.Errorf("User has no email")
	}

	if !discordUser.Verified {
		return "", nil, fmt.Errorf("User email is not verified")
	}

	displayName := discordUser.GlobalName
	if displayName == "" {
		displayName = fmt.Sprintf("%s#%s", discordUser.Username, discordUser.Discriminator)
	}

	user, err := h.userStore.UpsertUser(c.Context(), &model.User{
		ID:              util.UniqueID(),
		Email:           discordUser.Email,
		DisplayName:     displayName,
		DiscordID:       discordUser.ID,
		DiscordUsername: discordUser.Username,
		DiscordAvatar:   discordUser.Avatar,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).Error("Failed to upsert user")
		return "", nil, err
	}

	resp, err = client.Get("https://discord.com/api/users/@me/guilds")
	if err != nil {
		slog.With("error", err).Error("Failed to get guilds")
		return "", nil, fmt.Errorf("Failed to get guilds: %w", err)
	}

	guilds := []struct {
		ID string `json:"id"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&guilds)
	if err != nil {
		return "", nil, fmt.Errorf("Failed to decode guilds: %w", err)
	}
	resp.Body.Close()

	guildIDs := make([]string, len(guilds))
	for i, guild := range guilds {
		guildIDs[i] = guild.ID
	}

	token, session, err := h.sessionManager.CreateSession(c, user.ID)
	if err != nil {
		return "", nil, err
	}

	return token, session, nil
}

func (h *AuthHandler) getOauthStateCookie(c *handler.Context) string {
	state := c.Cookie("oauth_state")
	c.DeleteCookie("oauth_state")
	return state
}

func (h *AuthHandler) setOauthStateCookie(c *handler.Context) string {
	b := make([]byte, 128)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	c.SetCookie(&http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		Secure:   h.config.SecureCookies,
	})
	return state
}

func (h *AuthHandler) getOauthRedirectURL(c *handler.Context) string {
	redirectURL := h.config.AppPublicBaseURL

	path := c.Cookie("oauth_redirect")
	if path != "" {
		redirectURL += path
	}

	c.DeleteCookie("oauth_redirect")
	return redirectURL
}

func (h *AuthHandler) setOauthRedirectCookie(c *handler.Context) {
	redirectURL := c.Query("redirect")
	if redirectURL != "" {
		c.SetCookie(&http.Cookie{
			Name:     "oauth_redirect",
			Value:    redirectURL,
			HttpOnly: true,
			Secure:   h.config.SecureCookies,
		})
	} else {
		c.DeleteCookie("oauth_redirect")
	}
}
