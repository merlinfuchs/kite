package app

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/plan"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"gopkg.in/guregu/null.v4"
)

type AppHandler struct {
	appStore       store.AppStore
	userStore      store.UserStore
	planManager    *plan.PlanManager
	maxAppsPerUser int
}

func NewAppHandler(
	appStore store.AppStore,
	userStore store.UserStore,
	planManager *plan.PlanManager,
	maxAppsPerUser int,
) *AppHandler {
	return &AppHandler{
		appStore:       appStore,
		userStore:      userStore,
		planManager:    planManager,
		maxAppsPerUser: maxAppsPerUser,
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
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	app, err := h.appStore.CreateApp(c.Context(), &model.App{
		ID:            util.UniqueID(),
		Name:          appInfo.Name,
		Description:   appInfo.Description,
		Enabled:       true,
		OwnerUserID:   c.Session.UserID,
		CreatorUserID: c.Session.UserID,
		DiscordToken:  req.DiscordToken,
		DiscordID:     appInfo.ID,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	})
	if err != nil {
		slog.Error(
			"Failed to create app",
			slog.String("app_id", app.ID),
			slog.String("error", err.Error()),
		)
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppUpdate(c *handler.Context, req wire.AppUpdateRequest) (*wire.AppUpdateResponse, error) {
	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:             c.App.ID,
		Name:           req.Name,
		Description:    req.Description,
		DiscordToken:   c.App.DiscordToken,
		DiscordStatus:  c.App.DiscordStatus,
		Enabled:        req.Enabled,
		DisabledReason: c.App.DisabledReason,
		UpdatedAt:      time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	if req.Name != c.App.Name || req.Description != c.App.Description {
		if err := h.updateDiscordApp(c.Context(), app); err != nil {
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
		status = &model.AppDiscordStatus{
			Status:        req.DiscordStatus.Status,
			ActivityType:  req.DiscordStatus.ActivityType,
			ActivityName:  req.DiscordStatus.ActivityName,
			ActivityState: req.DiscordStatus.ActivityState,
			ActivityURL:   req.DiscordStatus.ActivityURL,
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

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:             c.App.ID,
		Name:           c.App.Name,
		Description:    c.App.Description,
		DiscordToken:   req.DiscordToken,
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
