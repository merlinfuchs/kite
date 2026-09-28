package app

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/utils/httputil"
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
}

func NewAppHandler(
	appStore store.AppStore,
	userStore store.UserStore,
	maxAppsPerUser int,
	tokenCrypt *util.SymmetricCrypt,
) *AppHandler {
	return &AppHandler{
		appStore:       appStore,
		userStore:      userStore,
		maxAppsPerUser: maxAppsPerUser,
		tokenCrypt:     tokenCrypt,
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
		return nil, fmt.Errorf("discord token belongs to a different app")
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

// maxAvatarSize bounds the decoded avatar. Its base64 form has to fit in
// handler.MaxJSONBodySize.
const maxAvatarSize = 5 * 1024 * 1024

func (h *AppHandler) HandleAppAvatarGet(c *handler.Context) (*wire.AppAvatarGetResponse, error) {
	avatarURL, err := h.getDiscordBotAvatarURL(c.Context(), c.App)
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
		}
		return nil, fmt.Errorf("failed to get discord bot avatar: %w", err)
	}

	return &wire.AppAvatarGetResponse{AvatarURL: avatarURL}, nil
}

func (h *AppHandler) HandleAppAvatarUpdate(c *handler.Context, req wire.AppAvatarUpdateRequest) (*wire.AppAvatarUpdateResponse, error) {
	content, err := base64.StdEncoding.DecodeString(req.Avatar)
	if err != nil {
		return nil, handler.ErrBadRequest("invalid_avatar", "Avatar is not valid base64")
	}

	if len(content) > maxAvatarSize {
		return nil, handler.ErrBadRequest(
			"resource_limit",
			fmt.Sprintf("avatar size exceeds maximum allowed size (%d)", maxAvatarSize),
		)
	}

	// Sniff the type instead of trusting the client.
	contentType := http.DetectContentType(content)
	switch contentType {
	case "image/png", "image/jpeg", "image/gif":
	default:
		return nil, handler.ErrBadRequest("invalid_avatar", "Avatar must be a PNG, JPEG or GIF image")
	}

	avatarURL, err := h.updateDiscordBotAvatar(c.Context(), c.App, api.Image{
		ContentType: contentType,
		Content:     content,
	})
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
		}
		var restErr *httputil.HTTPError
		if errors.As(err, &restErr) && restErr.Status == http.StatusBadRequest {
			// Discord reports avatar rate limits as a form error, not a 429.
			if bytes.Contains(restErr.Errors, []byte("AVATAR_RATE_LIMIT")) {
				return nil, handler.ErrBadRequest("avatar_rate_limited", "The avatar was changed too often, try again later")
			}
			return nil, handler.ErrBadRequest("invalid_avatar", "Discord rejected the avatar")
		}

		slog.Error(
			"Failed to update discord bot avatar",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to update discord bot avatar: %w", err)
	}

	return &wire.AppAvatarUpdateResponse{AvatarURL: avatarURL}, nil
}

func (h *AppHandler) HandleAppAvatarDelete(c *handler.Context) (*wire.AppAvatarDeleteResponse, error) {
	avatarURL, err := h.updateDiscordBotAvatar(c.Context(), c.App, api.Image{})
	if err != nil {
		if util.IsDiscordRestStatusCode(err, http.StatusUnauthorized) {
			return nil, handler.ErrBadRequest("invalid_discord_token", "Invalid Discord token")
		}

		var restErr *httputil.HTTPError
		if errors.As(err, &restErr) && restErr.Status == http.StatusBadRequest &&
			bytes.Contains(restErr.Errors, []byte("AVATAR_RATE_LIMIT")) {
			return nil, handler.ErrBadRequest("avatar_rate_limited", "The avatar was changed too often, try again later")
		}

		slog.Error(
			"Failed to remove discord bot avatar",
			slog.String("app_id", c.App.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to remove discord bot avatar: %w", err)
	}

	return &wire.AppAvatarDeleteResponse{AvatarURL: avatarURL}, nil
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
