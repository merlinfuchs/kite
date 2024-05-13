package app

import (
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/gofiber/fiber/v2"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/api/access"
	"github.com/merlinfuchs/kite/kite-service/internal/api/helpers"
	"github.com/merlinfuchs/kite/kite-service/internal/api/session"
	"github.com/merlinfuchs/kite/kite-service/internal/util"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
	"github.com/merlinfuchs/kite/kite-service/pkg/wire"
	"gopkg.in/guregu/null.v4"
)

type AppHandler struct {
	apps            store.AppStore
	appUsages       store.AppUsageStore
	appEntitlements store.AppEntitlementStore
	accessManager   *access.AccessManager
}

func NewHandler(apps store.AppStore, appUsages store.AppUsageStore, appEntitlements store.AppEntitlementStore, accessManager *access.AccessManager) *AppHandler {
	return &AppHandler{
		apps:            apps,
		appUsages:       appUsages,
		appEntitlements: appEntitlements,
		accessManager:   accessManager,
	}
}

func (h *AppHandler) HandleAppCreate(c *fiber.Ctx, req wire.AppCreateRequest) error {
	session := session.GetSession(c)

	client := rest.New(rest.NewClient(req.Token))

	disApp, err := client.GetCurrentApplication()
	if err != nil {
		return fmt.Errorf("failed to get current application: %w", err)
	}

	_, err = h.apps.GetApp(c.Context(), distype.Snowflake(disApp.ID.String()))
	if err == nil {
		return helpers.BadRequest("bot_exists", "The bot already exists.")
	} else if err != store.ErrNotFound {
		return fmt.Errorf("failed to get app: %w", err)
	}

	app := appFromDiscordApp(disApp, session.UserID, req.Token)
	err = h.apps.CreateApp(c.Context(), app)
	if err != nil {
		return fmt.Errorf("failed to create app: %w", err)
	}

	_, err = h.appEntitlements.UpsertAppEntitlement(c.Context(), model.AppEntitlement{
		ID:       util.UniqueID(),
		AppID:    app.ID,
		Source:   model.AppEntitlementSourceDefault,
		SourceID: null.NewString("initial_0", true),
		Features: model.AppEntitlementFeatures{
			MonthlyExecutionTimeLimit:    100_000 * time.Millisecond,
			MonthlyExecutionTimeAdditive: false,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert app entitlement: %w", err)
	}

	return c.JSON(wire.AppCreateResponse{
		Success: true,
		Data:    wire.AppToWire(app),
	})
}

func (h *AppHandler) HandleAppTokenUpdate(c *fiber.Ctx, req wire.AppTokenUpdateRequest) error {
	session := session.GetSession(c)
	appID := c.Params("appID")

	client := rest.New(rest.NewClient(req.Token))

	disApp, err := client.GetCurrentApplication()
	if err != nil {
		return fmt.Errorf("failed to get current application: %w", err)
	}

	if disApp.ID.String() != appID {
		return helpers.BadRequest("token_invalid", "The provided token doesn't belong to the bot.")
	}

	app := appFromDiscordApp(disApp, session.UserID, req.Token)
	app, err = h.apps.UpdateApp(c.Context(), app)
	if err != nil {
		return err
	}

	return c.JSON(wire.AppTokenUpdateResponse{
		Success: true,
		Data:    wire.AppToWire(app),
	})
}

func (h *AppHandler) HandleAppStatusUpdate(c *fiber.Ctx, req wire.AppStatusUpdateRequest) error {
	appID := c.Params("appID")

	app, err := h.apps.UpdateAppStatus(c.Context(), &model.App{
		ID:                  distype.Snowflake(appID),
		StatusType:          req.StatusType,
		StatusActivityType:  req.ActivityType,
		StatusActivityName:  req.ActivityName,
		StatusActivityState: req.ActivityState,
		StatusActivityUrl:   req.ActivityUrl,
	})
	if err != nil {
		return err
	}

	return c.JSON(wire.AppStatusUpdateResponse{
		Success: true,
		Data:    wire.AppToWire(app),
	})
}

func (h *AppHandler) HandleAppList(c *fiber.Ctx) error {
	session := session.GetSession(c)

	rows, err := h.apps.GetAppsForOwnerUser(c.Context(), session.UserID)
	if err != nil {
		return err
	}

	res := make([]wire.App, len(rows))
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
	app, err := h.apps.GetAppForOwnerUser(c.Context(), appID, session.UserID)
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

func appFromDiscordApp(app *discord.Application, userID distype.Snowflake, token string) *model.App {
	var avatar null.String
	if app.Bot.Avatar != nil {
		avatar = null.NewString(*app.Bot.Avatar, true)
	}

	var banner null.String
	if app.Bot.Banner != nil {
		banner = null.NewString(*app.Bot.Banner, true)
	}

	return &model.App{
		ID:                distype.Snowflake(app.ID.String()),
		OwnerUserID:       userID,
		Token:             token,
		TokenInvalid:      false,
		PublicKey:         app.VerifyKey,
		UserID:            distype.Snowflake(app.Bot.ID.String()),
		UserName:          app.Bot.Username,
		UserDiscriminator: app.Bot.Discriminator,
		UserAvatar:        avatar,
		UserBanner:        banner,
		UserBio:           null.NewString(app.Description, app.Description == ""),
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
}
