package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/core/sharedbot"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	discordoauth "github.com/ravener/discord-oauth2"
	"golang.org/x/oauth2"
)

// KiteBotConfig is what adding the Kite bot to a server needs. See
// design/kite-lite.md.
type KiteBotConfig struct {
	// Bot is nil without a Kite bot, which turns adding it off.
	Bot *sharedbot.Bot
	// ClientSecret is the OAuth2 secret of the Kite bot's application.
	ClientSecret   string
	MaxAppsPerUser int
}

// setupCookie keeps what the setup page in the dashboard should start with,
// like the templates and settings in a link from Xenon, while the user is on
// Discord's page.
const setupCookie = "kite_bot_setup"

// maxSetupLength bounds what's kept of it.
const maxSetupLength = 2000

func (h *AuthHandler) kiteBotOAuth2Config(c *handler.Context) (*oauth2.Config, error) {
	if h.kiteBot.Bot == nil {
		return nil, handler.ErrNotFound("no_kite_bot", "This Kite has no Kite bot")
	}
	appID, err := h.kiteBot.Bot.AppID(c.Context())
	if err != nil {
		return nil, err
	}

	return &oauth2.Config{
		RedirectURL:  fmt.Sprintf("%s/v1/auth/kite-bot/callback", h.config.APIPublicBaseURL),
		ClientID:     appID.String(),
		ClientSecret: h.kiteBot.ClientSecret,
		Scopes:       []string{discordoauth.ScopeIdentify, discordoauth.ScopeEmail, "bot", "applications.commands"},
		Endpoint:     discordoauth.Endpoint,
	}, nil
}

// HandleKiteBotAdd logs the user in and adds the Kite bot to a server on the
// same page of Discord. guild_id picks the server, the rest of the query goes
// to the setup page afterwards.
func (h *AuthHandler) HandleKiteBotAdd(c *handler.Context) error {
	conf, err := h.kiteBotOAuth2Config(c)
	if err != nil {
		return err
	}

	state := h.setOauthStateCookie(c)
	if setup := c.QueryValues().Encode(); setup != "" && len(setup) <= maxSetupLength {
		c.SetCookie(&http.Cookie{
			Name:     setupCookie,
			Value:    url.QueryEscape(setup),
			HttpOnly: true,
			Secure:   h.config.SecureCookies,
		})
	} else {
		c.DeleteCookie(setupCookie)
	}

	params := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("permissions", strconv.FormatUint(uint64(sharedbot.Permissions), 10)),
		oauth2.SetAuthURLParam("integration_type", "0"),
	}
	if guildID, err := discord.ParseSnowflake(c.Query("guild_id")); err == nil && guildID.IsValid() {
		params = append(params,
			oauth2.SetAuthURLParam("guild_id", guildID.String()),
			oauth2.SetAuthURLParam("disable_guild_select", "true"),
		)
	}

	c.Redirect(conf.AuthCodeURL(state, params...), http.StatusTemporaryRedirect)
	return nil
}

// HandleKiteBotCallback gets the server the user added the Kite bot to from
// Discord, gives it an app on the Kite bot, and sends the user to set it up.
// Discord only lets users with Manage Server add a bot, so they can change the
// server's app: the first one owns it, the others become collaborators.
func (h *AuthHandler) HandleKiteBotCallback(c *handler.Context) error {
	base := strings.TrimSuffix(h.config.AppPublicBaseURL, "/")

	state := h.getOauthStateCookie(c)
	if state == "" || c.Query("state") != state {
		return fmt.Errorf("invalid state")
	}
	setup, _ := url.QueryUnescape(c.Cookie(setupCookie))
	c.DeleteCookie(setupCookie)

	// Like when the user cancels on Discord's page.
	if c.Query("error") != "" {
		c.Redirect(base+"/apps", http.StatusTemporaryRedirect)
		return nil
	}

	conf, err := h.kiteBotOAuth2Config(c)
	if err != nil {
		return err
	}
	tokenData, err := conf.Exchange(c.Context(), c.Query("code"))
	if err != nil {
		return fmt.Errorf("failed to exchange token: %w", err)
	}

	_, _, user, err := h.authenticate(c, conf.Client(c.Context(), tokenData))
	if err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Discord says which server the bot was added to in the token, which only
	// it can make, unlike the guild_id of the redirect.
	guild, _ := tokenData.Extra("guild").(map[string]any)
	guildID, _ := guild["id"].(string)
	guildName, _ := guild["name"].(string)
	if _, err := discord.ParseSnowflake(guildID); err != nil {
		return fmt.Errorf("token has no server")
	}

	app, errCode, err := h.kiteBotApp(c, user, guildID, guildName)
	if err != nil {
		return err
	}
	if errCode != "" {
		c.Redirect(base+"/apps?kite_bot_error="+errCode, http.StatusTemporaryRedirect)
		return nil
	}

	target := fmt.Sprintf("%s/apps/%s/kite-bot", base, app.ID)
	if setup != "" {
		target += "?" + setup
	}
	c.Redirect(target, http.StatusTemporaryRedirect)
	return nil
}

// kiteBotApp returns the server's app on the Kite bot for the user, or the
// code of why they can't have it.
func (h *AuthHandler) kiteBotApp(c *handler.Context, user *model.User, guildID string, guildName string) (*model.App, string, error) {
	ctx := c.Context()
	now := time.Now().UTC()

	_, err := h.userStore.UserBan(ctx, user.ID, model.UserBanFeatureKite)
	if err == nil {
		return nil, "banned", nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return nil, "", fmt.Errorf("failed to get user ban: %w", err)
	}

	app, err := h.appStore.SharedAppByGuild(ctx, guildID)
	if errors.Is(err, store.ErrNotFound) {
		count, err := h.appStore.CountAppsByUser(ctx, user.ID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to count apps: %w", err)
		}
		if count >= h.kiteBot.MaxAppsPerUser {
			return nil, "app_limit", nil
		}

		if guildName == "" {
			guildName = "Kite bot"
		}
		app, err = h.appStore.CreateSharedApp(ctx, store.SharedAppCreateOpts{
			ID:             util.UniqueID(),
			Name:           guildName,
			OwnerUserID:    user.ID,
			DiscordGuildID: guildID,
			CreatedAt:      now,
		})
		if errors.Is(err, store.ErrAlreadyExists) {
			// Someone else set the server up at the same time.
			app, err = h.appStore.SharedAppByGuild(ctx, guildID)
		}
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to get app of server: %w", err)
	}

	// Adding the bot again starts the app, like when the bot sees itself
	// join, which may come first.
	if _, err := h.appStore.EnableSharedAppDisabledWithReason(ctx, guildID, model.AppDisabledReasonKiteBotRemoved, now); err != nil {
		return nil, "", fmt.Errorf("failed to start app: %w", err)
	}

	if app.OwnerUserID == user.ID {
		return app, "", nil
	}
	_, err = h.appStore.Collaborator(ctx, app.ID, user.ID)
	if errors.Is(err, store.ErrNotFound) {
		_, err = h.appStore.CreateCollaborator(ctx, &model.AppCollaborator{
			AppID:     app.ID,
			UserID:    user.ID,
			Role:      model.AppCollaboratorRoleAdmin,
			CreatedAt: now,
			UpdatedAt: now,
		})
		if err == nil {
			slog.Info("Added a user with Manage Server to the app of their server", slog.String("app_id", app.ID), slog.String("user_id", user.ID))
		}
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to add collaborator: %w", err)
	}

	return app, "", nil
}
