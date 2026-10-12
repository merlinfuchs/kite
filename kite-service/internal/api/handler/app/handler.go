package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/command"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

type AppHandler struct {
	appStore            store.AppStore
	userStore           store.UserStore
	commandStore        store.CommandStore
	pluginInstanceStore store.PluginInstanceStore
	commandManager      *command.CommandManager
	maxAppsPerUser      int

	tokenCrypt *util.SymmetricCrypt
}

func NewAppHandler(
	appStore store.AppStore,
	userStore store.UserStore,
	commandStore store.CommandStore,
	pluginInstanceStore store.PluginInstanceStore,
	commandManager *command.CommandManager,
	maxAppsPerUser int,
	tokenCrypt *util.SymmetricCrypt,
) *AppHandler {
	return &AppHandler{
		appStore:            appStore,
		userStore:           userStore,
		commandStore:        commandStore,
		pluginInstanceStore: pluginInstanceStore,
		commandManager:      commandManager,
		maxAppsPerUser:      maxAppsPerUser,
		tokenCrypt:          tokenCrypt,
	}
}

func (h *AppHandler) HandleAppList(c *handler.Context) (*wire.AppListResponse, error) {
	apps, err := h.appStore.AppsByUser(c.Context(), c.Session.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apps: %w", err)
	}

	res := make([]*wire.App, len(apps))
	for i, app := range apps {
		res[i] = wire.AppToWire(app)
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

	// TODO: existingApp, err := h.appStore.AppByDiscordID()

	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
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
		ID:            util.UniqueID(),
		Name:          appInfo.Name,
		Description:   appInfo.Description,
		Enabled:       true,
		OwnerUserID:   c.Session.UserID,
		CreatorUserID: c.Session.UserID,
		DiscordToken:  encryptedToken,
		DiscordID:     appInfo.ID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
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

	if req.Name != c.App.Name || req.Description != c.App.Description {
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

	if req.Name != c.App.Name {
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

func (h *AppHandler) HandleAppTokenUpdate(c *handler.Context, req wire.AppTokenUpdateRequest) (*wire.AppTokenUpdateResponse, error) {
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
		return h.changeDiscordApp(c, req, appInfo)
	}

	encryptedToken, err := h.tokenCrypt.EncryptString(req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt discord token: %w", err)
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:             c.App.ID,
		Name:           c.App.Name,
		Description:    c.App.Description,
		DiscordToken:   encryptedToken,
		DiscordStatus:  c.App.DiscordStatus,
		Enabled:        true,
		DisabledReason: null.String{}, // We reset the disabled reason when the app token is updated
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

	return wire.AppToWire(app), nil
}

// changeDiscordApp moves the app to the Discord app the new token belongs to.
// Everything stored in Kite stays, but the commands have to be deployed to the
// new Discord app, and anything tied to the old one (servers, emojis) doesn't
// carry over.
func (h *AppHandler) changeDiscordApp(c *handler.Context, req wire.AppTokenUpdateRequest, appInfo *DiscordAppInfo) (*wire.AppTokenUpdateResponse, error) {
	if !c.UserAppRole.CanChangeDiscordApp() {
		return nil, handler.ErrForbidden("missing_permissions", "Only the owner can switch the app to a different Discord app")
	}

	if !req.ChangeApp {
		return nil, handler.ErrBadRequest("discord_app_changed", fmt.Sprintf("This token belongs to a different Discord app (%s)", appInfo.Name))
	}

	existingApp, err := h.appStore.AppByDiscordID(c.Context(), appInfo.ID)
	if err != nil && !errors.Is(err, store.ErrNotFound) {
		return nil, fmt.Errorf("failed to get app by discord id: %w", err)
	}
	if existingApp != nil {
		return nil, handler.ErrBadRequest("discord_app_in_use", "This Discord app is already used by another app on Kite")
	}

	encryptedToken, err := h.tokenCrypt.EncryptString(req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt discord token: %w", err)
	}

	oldApp := c.App

	app, err := h.appStore.UpdateAppDiscordApp(c.Context(), store.AppDiscordAppUpdateOpts{
		ID:           c.App.ID,
		Name:         appInfo.Name,
		Description:  appInfo.Description,
		DiscordID:    appInfo.ID,
		DiscordToken: encryptedToken,
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to change discord app",
			slog.String("app_id", c.App.ID),
			slog.String("discord_id", appInfo.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to change discord app: %w", err)
	}

	// The new Discord app has none of the commands yet. Mark them as not
	// deployed first, so they show up as such if deploying below fails.
	if err := h.commandStore.ResetCommandsLastDeployedAt(c.Context(), app.ID); err != nil {
		return nil, fmt.Errorf("failed to reset commands last deployed at: %w", err)
	}
	if err := h.pluginInstanceStore.ResetPluginInstancesLastDeployedAt(c.Context(), app.ID); err != nil {
		return nil, fmt.Errorf("failed to reset plugin instances last deployed at: %w", err)
	}

	if err := h.commandManager.DeployCommandsForApp(c.Context(), app.ID); err != nil {
		slog.Warn(
			"Failed to deploy commands after changing discord app",
			slog.String("app_id", app.ID),
			slog.String("error", err.Error()),
		)
	}

	// Best effort: the old bot would otherwise keep showing commands that no
	// longer do anything. Its token may already be reset, which is fine.
	if err := h.clearDiscordAppCommands(c.Context(), oldApp); err != nil {
		slog.Info(
			"Failed to clear commands of old discord app",
			slog.String("app_id", app.ID),
			slog.String("discord_id", oldApp.DiscordID),
			slog.String("error", err.Error()),
		)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppDelete(c *handler.Context) (*wire.AppDeleteResponse, error) {
	if !c.UserAppRole.CanDeleteApp() {
		return nil, handler.ErrForbidden("missing_permissions", "You don't have permissions to delete this app")
	}

	if err := h.appStore.DeleteApp(c.Context(), c.App.ID); err != nil {
		return nil, fmt.Errorf("failed to delete app: %w", err)
	}

	return &wire.AppDeleteResponse{}, nil
}

func (h *AppHandler) HandleAppEmojisList(c *handler.Context) (*wire.AppEmojiListResponse, error) {
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
			ID:   entity.ID,
			Type: string(entity.Type),
			Name: entity.Name,
		}
	}

	return &res, nil
}
