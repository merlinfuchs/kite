package app

import (
	"errors"
	"fmt"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

type AppHandler struct {
	appStore       store.AppStore
	userStore      store.UserStore
	maxAppsPerUser int

	tokenCrypt *util.SymmetricCrypt
	// appStateManager reaches the Kite bot for apps on it.
	appStateManager store.AppStateManager
}

func NewAppHandler(
	appStore store.AppStore,
	userStore store.UserStore,
	maxAppsPerUser int,
	tokenCrypt *util.SymmetricCrypt,
	appStateManager store.AppStateManager,
) *AppHandler {
	return &AppHandler{
		appStore:       appStore,
		userStore:      userStore,
		maxAppsPerUser: maxAppsPerUser,
		tokenCrypt:     tokenCrypt,

		appStateManager: appStateManager,
	}
}

func (h *AppHandler) HandleAppList(c *handler.Context) (*wire.AppListResponse, error) {
	apps, err := h.appStore.AppsByUser(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}

	res := make([]*wire.App, 0, len(apps))
	for _, app := range apps {
		// Apps on the Kite bot can't be used with tokens, see CustomBotOnly.
		if c.APIToken != nil && (app.IsShared() || !c.APIToken.CanUseApp(app.ID)) {
			continue
		}
		res = append(res, wire.AppToWire(app))
	}

	return &res, nil
}

func (h *AppHandler) HandleAppGet(c *handler.Context) (*wire.AppGetResponse, error) {
	return wire.AppToWire(c.App), nil
}

func (h *AppHandler) HandleAppCreate(c *handler.Context, req wire.AppCreateRequest) (*wire.AppCreateResponse, error) {
	appCount, err := h.appStore.CountAppsByUser(c.Context(), c.Session.UserID)
	if err != nil {
		slog.Error(
			"Failed to count apps",
			slog.String("user_id", c.Session.UserID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to count apps: %w", err)
	}

	if appCount >= h.maxAppsPerUser {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of apps (%d) reached", h.maxAppsPerUser))
	}

	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Discord didn't accept this token. Copy it again from Bot in the Discord Developer Portal.")
		}
		slog.Error(
			"Failed to get discord app info",
			slog.String("user_id", c.Session.UserID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	encryptedToken, err := h.tokenCrypt.EncryptString(req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt discord token: %w", err)
	}

	app, err := h.appStore.CreateApp(c.Context(), &model.App{
		ID:               util.UniqueID(),
		Name:             appInfo.Name,
		Description:      appInfo.Description,
		Enabled:          true,
		OwnerUserID:      c.Session.UserID,
		CreatorUserID:    c.Session.UserID,
		DiscordToken:     encryptedToken,
		DiscordID:        appInfo.ID,
		DiscordVerifyKey: null.NewString(appInfo.VerifyKey, appInfo.VerifyKey != ""),
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	})
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, handler.ErrBadRequest("app_exists", "This bot is already in Kite. If it's yours, open it from your apps and update its token in the settings. Otherwise ask the person who added it to invite you.")
		}
		slog.Error(
			"Failed to create app",
			slog.String("discord_id", appInfo.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppUpdate(c *handler.Context, req wire.AppUpdateRequest) (*wire.AppUpdateResponse, error) {
	if req.Enabled && !c.App.Enabled {
		if err := h.checkOwnerNotBanned(c); err != nil {
			return nil, err
		}
	}

	disabledReason := c.App.DisabledReason
	if req.Enabled {
		disabledReason = null.String{}
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:             c.App.ID,
		Name:           req.Name,
		Description:    req.Description,
		DiscordToken:   c.App.DiscordToken,
		DiscordStatus:  c.App.DiscordStatus,
		Enabled:        req.Enabled,
		DisabledReason: disabledReason,
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	if req.Enabled && !c.App.Enabled {
		h.retryFailedCommandsDeploy(c)
	}

	// The Kite bot's name is the same in every server, so a shared app's name
	// is only its name in Kite.
	if !app.IsShared() && (req.Name != c.App.Name || req.Description != c.App.Description) {
		if err := h.updateDiscordApp(c.Context(), app); err != nil {
			if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
				return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
			}

			slog.Error(
				"Failed to update discord app name",
				slog.String("app_id", c.App.ID),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to update discord app name: %w", err)
		}
	}

	if !app.IsShared() && req.Name != c.App.Name {
		if err := h.updateDiscordBotUser(c.Context(), app); err != nil {
			if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
				return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
			}

			slog.Error(
				"Failed to update discord bot user",
				slog.String("app_id", c.App.ID),
				slog.String("error", err.Error()),
			)
			return nil, fmt.Errorf("failed to update discord bot user: %w", err)
		}
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppStatusUpdate(c *handler.Context, req wire.AppStatusUpdateRequest) (*wire.AppStatusUpdateResponse, error) {
	var status *model.AppDiscordStatus
	if req.DiscordStatus != nil {
		entries := make([]model.AppDiscordStatusEntry, len(req.DiscordStatus.Statuses))
		for i, e := range req.DiscordStatus.Statuses {
			entries[i] = model.AppDiscordStatusEntry(e)
			if entries[i].ID == "" {
				entries[i].ID = util.UniqueID()
			}
		}

		status = &model.AppDiscordStatus{
			Statuses:      entries,
			ActiveID:      req.DiscordStatus.ActiveID,
			RotateEnabled: req.DiscordStatus.RotateEnabled && c.Features.RotatingStatus,
		}
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:             c.App.ID,
		Name:           c.App.Name,
		Description:    c.App.Description,
		DiscordToken:   c.App.DiscordToken,
		DiscordStatus:  status,
		Enabled:        c.App.Enabled,
		DisabledReason: c.App.DisabledReason,
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update app status: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppCreditNoticesUpdate(c *handler.Context, req wire.AppCreditNoticesUpdateRequest) (*wire.AppCreditNoticesUpdateResponse, error) {
	app, err := h.appStore.UpdateAppCreditNotices(c.Context(), c.App.ID, req.Enabled, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("failed to update app credit notices: %w", err)
	}

	return wire.AppToWire(app), nil
}

// HandleAppSwitchToCustom moves an app on the Kite bot to the user's own bot,
// Custom setup. It keeps its templates, which become editable, its stored
// variables, logs and plan. The Kite bot stays in the server, its commands
// for the app are removed.
// The user becomes its only owner, see design/kite-lite.md.
func (h *AppHandler) HandleAppSwitchToCustom(c *handler.Context, req wire.AppSwitchToCustomRequest) (*wire.AppSwitchToCustomResponse, error) {
	if !c.App.IsShared() {
		return nil, handler.ErrBadRequest("not_shared_app", "This app already has a bot of its own.")
	}
	// Taking the app over needs Manage Server now, not access through an old
	// record while the bot isn't in the server.
	if c.SharedAppMember == nil {
		return nil, handler.ErrForbidden("kite_bot_missing", "Add the Kite bot to your server again to switch this app.")
	}
	// So a banned owner can't get the app running again through someone else.
	if err := h.checkOwnerNotBanned(c); err != nil {
		return nil, err
	}
	if c.App.OwnerUserID != c.Session.UserID {
		appCount, err := h.appStore.CountAppsByUser(c.Context(), c.Session.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to count apps: %w", err)
		}
		if appCount >= h.maxAppsPerUser {
			return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("You already have the maximum number of apps (%d). Switching makes this app yours, so delete one of your apps first.", h.maxAppsPerUser))
		}
	}

	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Discord didn't accept this token. Copy it again from Bot in the Discord Developer Portal.")
		}
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	encryptedToken, err := h.tokenCrypt.EncryptString(req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt discord token: %w", err)
	}

	// Before the switch, as only the app on the Kite bot reaches them.
	h.removeSharedCommands(c)

	app, err := h.appStore.SwitchSharedAppToOwn(c.Context(), store.SharedAppSwitchOpts{
		ID:               c.App.ID,
		OwnerUserID:      c.Session.UserID,
		Name:             appInfo.Name,
		Description:      appInfo.Description,
		DiscordID:        appInfo.ID,
		DiscordToken:     encryptedToken,
		DiscordVerifyKey: null.NewString(appInfo.VerifyKey, appInfo.VerifyKey != ""),
		UpdatedAt:        time.Now().UTC(),
	})
	if err != nil {
		// The Kite bot sends the commands it had again.
		if err := h.appStore.QueueAppCommandsDeploy(c.Context(), c.App.ID, true); err != nil {
			slog.Error("Failed to queue commands deploy", slog.String("app_id", c.App.ID), slog.String("error", err.Error()))
		}
		if errors.Is(err, store.ErrAlreadyExists) {
			return nil, handler.ErrBadRequest("app_exists", "This bot is already an app in Kite. Create a new bot for this app in the Discord Developer Portal.")
		}
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrBadRequest("not_shared_app", "This app already has a bot of its own.")
		}
		return nil, fmt.Errorf("failed to switch app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppTokenUpdate(c *handler.Context, req wire.AppTokenUpdateRequest) (*wire.AppTokenUpdateResponse, error) {
	// A new token starts the app.
	if err := h.checkOwnerNotBanned(c); err != nil {
		return nil, err
	}

	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
		}

		slog.Error(
			"Failed to get discord app info",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	if appInfo.ID != c.App.DiscordID {
		return nil, handler.ErrBadRequest("wrong_discord_app", "This token belongs to a different bot. Copy the token of this bot in the Discord Developer Portal.")
	}

	encryptedToken, err := h.tokenCrypt.EncryptString(req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt discord token: %w", err)
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:            c.App.ID,
		Name:          c.App.Name,
		Description:   c.App.Description,
		DiscordToken:  encryptedToken,
		DiscordStatus: c.App.DiscordStatus,
		// A new token starts the app and resets why it was stopped.
		Enabled:        true,
		DisabledReason: null.String{},
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to update app",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	h.retryFailedCommandsDeploy(c)

	return wire.AppToWire(app), nil
}

// checkOwnerNotBanned refuses to start an app whose owner Kite banned, which
// its collaborators could otherwise do.
func (h *AppHandler) checkOwnerNotBanned(c *handler.Context) error {
	_, err := h.userStore.UserBan(c.Context(), c.App.OwnerUserID, model.UserBanFeatureKite)
	if errors.Is(err, store.ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get user ban: %w", err)
	}
	return handler.ErrForbidden("app_owner_banned", "Kite banned the owner of this app, so it can't be started")
}

func (h *AppHandler) HandleAppDelete(c *handler.Context) (*wire.AppDeleteResponse, error) {
	if !c.UserAppRole.CanDeleteApp() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to delete this app")
	}

	if c.App.IsShared() {
		h.removeSharedCommands(c)
	}

	if err := h.appStore.DeleteApp(c.Context(), c.App.ID); err != nil {
		return nil, fmt.Errorf("failed to delete app: %w", err)
	}

	return &wire.AppDeleteResponse{}, nil
}

// removeSharedCommands removes the commands of an app on the Kite bot from its
// server, where nothing would answer them anymore. The bot stays in the
// server, adding it again sets up a new app. A failure doesn't stop the app
// from being deleted.
func (h *AppHandler) removeSharedCommands(c *handler.Context) {
	err := func() error {
		client, err := h.appStateManager.AppClient(c.Context(), c.App.ID)
		if err != nil {
			return err
		}
		bot, err := client.CurrentApplication()
		if err != nil {
			return err
		}
		guildID, err := discord.ParseSnowflake(c.App.DiscordGuildID.String)
		if err != nil {
			return err
		}
		_, err = client.BulkOverwriteGuildCommands(bot.ID, discord.GuildID(guildID), []api.CreateCommandData{})
		return err
	}()
	if err != nil {
		slog.Error(
			"Failed to remove the commands of a deleted app on the Kite bot",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
	}
}

func (h *AppHandler) HandleAppEmojisList(c *handler.Context) (*wire.AppEmojiListResponse, error) {
	// The Kite bot's emojis aren't the app's.
	if c.App.IsShared() {
		return &wire.AppEmojiListResponse{}, nil
	}

	emojis, err := h.getAppEmojis(c.Context(), c.App)
	if err != nil {
		return nil, fmt.Errorf("failed to get app emojis: %w", err)
	}

	res := make([]*wire.AppEmoji, len(emojis))
	for i, emoji := range emojis {
		res[i] = &wire.AppEmoji{
			ID:        emoji.ID.String(),
			Name:      emoji.Name,
			Animated:  emoji.Animated,
			Available: emoji.Available,
		}
	}

	return &res, nil
}

func (h *AppHandler) HandleAppEntityList(c *handler.Context) (*wire.AppEntityListResponse, error) {
	entities, err := h.appStore.AppEntities(c.Context(), c.App.ID)
	if err != nil {
		slog.Error(
			"Failed to get app entities",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to get app entities: %w", err)
	}

	res := make([]*wire.AppEntity, len(entities))
	for i, entity := range entities {
		res[i] = &wire.AppEntity{
			ID:      entity.ID,
			Type:    string(entity.Type),
			Name:    entity.Name,
			Enabled: entity.Enabled,
		}
	}

	return &res, nil
}
