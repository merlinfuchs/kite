package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-common/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-common/model"
	"github.com/kitecloud/kite/kite-common/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) AppsByUser(ctx context.Context, userID string) ([]*model.App, error) {
	rows, err := c.Q.GetAppsByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	var apps []*model.App
	for _, row := range rows {
		apps = append(apps, rowToApp(row))
	}

	return apps, nil
}

func (c *Client) CountAppsByUser(ctx context.Context, userID string) (int, error) {
	res, err := c.Q.CountAppsByOwner(ctx, userID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) App(ctx context.Context, id string) (*model.App, error) {
	row, err := c.Q.GetApp(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToApp(row), nil
}

func (c *Client) AppCredentials(ctx context.Context, id string) (*model.AppCredentials, error) {
	row, err := c.Q.GetAppCredentials(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return &model.AppCredentials{
		DiscordID:    row.DiscordID,
		DiscordToken: row.DiscordToken,
	}, nil
}

func (c *Client) CreateApp(ctx context.Context, app *model.App) (*model.App, error) {
	row, err := c.Q.CreateApp(ctx, pgmodel.CreateAppParams{
		ID:   app.ID,
		Name: app.Name,
		Description: pgtype.Text{
			String: app.Description.String,
			Valid:  app.Description.Valid,
		},
		OwnerUserID:   app.OwnerUserID,
		CreatorUserID: app.CreatorUserID,
		DiscordToken:  app.DiscordToken,
		DiscordID:     app.DiscordID,
		CreatedAt:     pgtype.Timestamp{Time: app.CreatedAt.UTC(), Valid: true},
		UpdatedAt:     pgtype.Timestamp{Time: app.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToApp(row), nil
}

func (c *Client) UpdateApp(ctx context.Context, opts store.AppUpdateOpts) (*model.App, error) {
	row, err := c.Q.UpdateApp(ctx, pgmodel.UpdateAppParams{
		ID:   opts.ID,
		Name: opts.Name,
		Description: pgtype.Text{
			String: opts.Description.String,
			Valid:  opts.Description.Valid,
		},
		DiscordToken: opts.DiscordToken,
		UpdatedAt:    pgtype.Timestamp{Time: opts.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToApp(row), nil
}

func (c *Client) DeleteApp(ctx context.Context, id string) error {
	return c.Q.DeleteApp(ctx, id)
}

func (c *Client) AppIDs(ctx context.Context) ([]string, error) {
	return c.Q.GetAppIDs(ctx)
}

func (c *Client) AppsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.App, error) {
	rows, err := c.Q.GetAppsUpdatedSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	var apps []*model.App
	for _, row := range rows {
		apps = append(apps, rowToApp(row))
	}

	return apps, nil
}

func rowToApp(row pgmodel.App) *model.App {
	return &model.App{
		ID:            row.ID,
		Name:          row.Name,
		Description:   null.NewString(row.Description.String, row.Description.Valid),
		OwnerUserID:   row.OwnerUserID,
		CreatorUserID: row.CreatorUserID,
		DiscordToken:  row.DiscordToken,
		DiscordID:     row.DiscordID,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}
}
