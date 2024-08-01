package app

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
)

type AppHandler struct {
	appStore       store.AppStore
	maxAppsPerUser int
}

func NewAppHandler(appStore store.AppStore, maxAppsPerUser int) *AppHandler {
	return &AppHandler{
		appStore:       appStore,
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
		return nil, fmt.Errorf("failed to count apps: %w", err)
	}

	if appCount >= h.maxAppsPerUser {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf("maximum number of apps (%d) reached", h.maxAppsPerUser))
	}

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
		return nil, fmt.Errorf("failed to create app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppUpdate(c *handler.Context, req wire.AppUpdateRequest) (*wire.AppUpdateResponse, error) {
	if req.Name != c.App.Name {
		if err := h.updateDiscordAppName(c.Context(), c.App.DiscordToken, req.Name); err != nil {
			return nil, fmt.Errorf("failed to update discord app name: %w", err)
		}
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:           c.App.ID,
		Name:         req.Name,
		Description:  req.Description,
		DiscordToken: c.App.DiscordToken,
		Enabled:      req.Enabled,
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppTokenUpdate(c *handler.Context, req wire.AppTokenUpdateRequest) (*wire.AppTokenUpdateResponse, error) {
	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	if appInfo.ID != c.App.DiscordID {
		return nil, fmt.Errorf("discord token belongs to a different app")
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:           c.App.ID,
		Name:         c.App.Name,
		Description:  c.App.Description,
		DiscordToken: req.DiscordToken,
		UpdatedAt:    time.Now().UTC(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return wire.AppToWire(app), nil
}

func (h *AppHandler) HandleAppDelete(c *handler.Context) (*wire.AppDeleteResponse, error) {
	if err := h.appStore.DeleteApp(c.Context(), c.App.ID); err != nil {
		return nil, fmt.Errorf("failed to delete app: %w", err)
	}

	return &wire.AppDeleteResponse{}, nil
}
