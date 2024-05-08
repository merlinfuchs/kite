package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
)

type AppHandler struct {
	apps            store.AppStore
	appUsages       store.AppUsageStore
	AppEntitlements store.AppEntitlementStore
	accessManager   *access.AccessManager
}

func NewHandler(apps store.AppStore, appUsages store.AppUsageStore, appEntitlements store.AppEntitlementStore, accessManager *access.AccessManager) *AppHandler {
	return &AppHandler{
		apps:            apps,
		appUsages:       appUsages,
		AppEntitlements: appEntitlements,
		accessManager:   accessManager,
	}
}

func (h *AppHandler) HandleAppList(c *fiber.Ctx) error {
	session := session.GetSession(c)

	rows, err := h.apps.GetAppsForOwnerUser(c.Context(), session.UserID)
	if err != nil {
		return err
	}

	res := make([]wire.App, 0, len(rows))
	for i, app := range rows {
		res[i] = wire.AppToWire(&app)
	}

	return c.JSON(wire.AppListResponse{
		Success: true,
		Data:    res,
	})
}

func (h *AppHandler) HandleAppGet(c *fiber.Ctx) error {
	session := session.GetSession(c)

	appID := distype.Snowflake(c.Params("appID"))
	app, err := h.apps.GetApp(c.Context(), appID, session.UserID)
	if err != nil {
		if err == store.ErrNotFound {
			return helpers.NotFound("unknown_app", "App not found")
		}
		return err
	}

	return c.JSON(wire.AppGetResponse{
		Success: true,
		Data:    wire.AppToWire(app),
	})
}
