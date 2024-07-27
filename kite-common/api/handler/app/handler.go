package app

import (
	"fmt"
	"time"

	"github.com/kitecloud/kite/kite-common/api/handler"
	"github.com/kitecloud/kite/kite-common/api/wire"
	"github.com/kitecloud/kite/kite-common/model"
	"github.com/kitecloud/kite/kite-common/store"
	"github.com/kitecloud/kite/kite-common/util"
	"gopkg.in/guregu/null.v4"
)

type AppHandler struct {
	appStore store.AppStore
}

func NewAppHandler(appStore store.AppStore) *AppHandler {
	return &AppHandler{
		appStore: appStore,
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
	appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get discord app info: %w", err)
	}

	app, err := h.appStore.CreateApp(c.Context(), &model.App{
		ID:            util.UniqueID(),
		Name:          appInfo.Name,
		Description:   appInfo.Description,
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
	name := c.App.Name
	if req.Name.Valid {
		name = req.Name.String
	}

	description := c.App.Description
	if req.Description.Valid {
		description = null.NewString(req.Description.String, req.Description.String != "")
	}

	discordToken := c.App.DiscordToken

	if req.DiscordToken.Valid {
		discordToken = req.DiscordToken.String

		appInfo, err := h.getDiscordAppInfo(c.Context(), req.DiscordToken.String)
		if err != nil {
			return nil, fmt.Errorf("failed to get discord app info: %w", err)
		}

		if appInfo.ID != c.App.DiscordID {
			return nil, fmt.Errorf("discord token belongs to a different app")
		}
	}

	if name != c.App.Name {
		if err := h.updateDiscordAppName(c.Context(), discordToken, name); err != nil {
			return nil, fmt.Errorf("failed to update discord app name: %w", err)
		}
	}

	app, err := h.appStore.UpdateApp(c.Context(), store.AppUpdateOpts{
		ID:           c.App.ID,
		Name:         name,
		Description:  description,
		DiscordToken: discordToken,
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
