package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/dismod/distype"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.AppStore = (*Client)(nil)

func (c *Client) CreateApp(ctx context.Context, app *model.App) error {
	err := c.Q.CreateApp(ctx, pgmodel.CreateAppParams{
		ID:                  string(app.ID),
		OwnerUserID:         string(app.OwnerUserID),
		Token:               app.Token,
		TokenInvalid:        app.TokenInvalid,
		PublicKey:           app.PublicKey,
		UserID:              string(app.UserID),
		UserName:            app.UserName,
		UserDiscriminator:   app.UserDiscriminator,
		UserAvatar:          nullStringToText(app.UserAvatar),
		UserBanner:          nullStringToText(app.UserBanner),
		UserBio:             nullStringToText(app.UserBio),
		StatusType:          string(app.StatusType),
		StatusActivityType:  nullIntToInt4(app.StatusActivityType),
		StatusActivityName:  nullStringToText(app.StatusActivityName),
		StatusActivityState: nullStringToText(app.StatusActivityState),
		StatusActivityUrl:   nullStringToText(app.StatusActivityUrl),
		CreatedAt:           timeToTimestamp(app.CreatedAt),
		UpdatedAt:           timeToTimestamp(app.UpdatedAt),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetAppsForOwnerUser(ctx context.Context, ownerUserID distype.Snowflake) ([]model.App, error) {
	apps, err := c.Q.GetAppsForOwnerUser(ctx, string(ownerUserID))
	if err != nil {
		return nil, err
	}

	result := make([]model.App, len(apps))
	for i, app := range apps {
		result[i] = appToModel(app)
	}

	return result, nil
}

func (c *Client) GetApp(ctx context.Context, id distype.Snowflake, ownerUserID distype.Snowflake) (*model.App, error) {
	app, err := c.Q.GetApp(ctx, pgmodel.GetAppParams{
		ID:          string(id),
		OwnerUserID: string(ownerUserID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}

		return nil, err
	}

	res := appToModel(app)
	return &res, nil
}

func (c *Client) GetAppsWithValidToken(ctx context.Context) ([]model.App, error) {
	apps, err := c.Q.GetAppsWithValidToken(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.App, len(apps))
	for i, app := range apps {
		result[i] = appToModel(app)
	}

	return result, nil
}

func (c *Client) GetDistinctAppIDs(ctx context.Context) ([]distype.Snowflake, error) {
	ids, err := c.Q.GetDistinctAppIDs(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]distype.Snowflake, len(ids))
	for i, id := range ids {
		result[i] = distype.Snowflake(id)
	}

	return result, nil
}

func (c *Client) UpdateApp(ctx context.Context, app *model.App) (*model.App, error) {
	res, err := c.Q.UpdateApp(ctx, pgmodel.UpdateAppParams{
		ID:                string(app.ID),
		Token:             app.Token,
		PublicKey:         app.PublicKey,
		UserID:            string(app.UserID),
		UserName:          app.UserName,
		UserDiscriminator: app.UserDiscriminator,
		UserAvatar:        nullStringToText(app.UserAvatar),
		UserBanner:        nullStringToText(app.UserBanner),
		UserBio:           nullStringToText(app.UserBio),
		UpdatedAt:         timeToTimestamp(app.UpdatedAt),
	})
	if err != nil {
		return nil, err
	}

	a := appToModel(res)
	return &a, nil
}

func (c *Client) UpdateAppStatus(ctx context.Context, bot *model.App) (*model.App, error) {
	res, err := c.Q.UpdateAppStatus(ctx, pgmodel.UpdateAppStatusParams{
		ID:                  string(bot.ID),
		StatusType:          string(bot.StatusType),
		StatusActivityType:  nullIntToInt4(bot.StatusActivityType),
		StatusActivityName:  nullStringToText(bot.StatusActivityName),
		StatusActivityState: nullStringToText(bot.StatusActivityState),
		StatusActivityUrl:   nullStringToText(bot.StatusActivityUrl),
		UpdatedAt:           timeToTimestamp(bot.UpdatedAt),
	})
	if err != nil {
		return nil, err
	}

	a := appToModel(res)
	return &a, nil
}

func (c *Client) DeleteApp(ctx context.Context, appID distype.Snowflake) error {
	err := c.Q.DeleteApp(ctx, string(appID))
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) CheckUserIsOwnerOfApp(ctx context.Context, appID distype.Snowflake, ownerUserID distype.Snowflake) (bool, error) {
	res, err := c.Q.CheckUserIsOwnerOfApp(ctx, pgmodel.CheckUserIsOwnerOfAppParams{
		ID:          string(appID),
		OwnerUserID: string(ownerUserID),
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return res, nil

}

func appToModel(row pgmodel.App) model.App {
	return model.App{
		ID:                  distype.Snowflake(row.ID),
		OwnerUserID:         distype.Snowflake(row.OwnerUserID),
		Token:               row.Token,
		TokenInvalid:        row.TokenInvalid,
		PublicKey:           row.PublicKey,
		UserID:              distype.Snowflake(row.UserID),
		UserName:            row.UserName,
		UserDiscriminator:   row.UserDiscriminator,
		UserAvatar:          textToNullString(row.UserAvatar),
		UserBanner:          textToNullString(row.UserBanner),
		UserBio:             textToNullString(row.UserBio),
		StatusType:          row.StatusType,
		StatusActivityType:  int4ToNullInt(row.StatusActivityType),
		StatusActivityName:  textToNullString(row.StatusActivityName),
		StatusActivityState: textToNullString(row.StatusActivityState),
		StatusActivityUrl:   textToNullString(row.StatusActivityUrl),
		CreatedAt:           row.CreatedAt.Time,
		UpdatedAt:           row.UpdatedAt.Time,
	}
}
