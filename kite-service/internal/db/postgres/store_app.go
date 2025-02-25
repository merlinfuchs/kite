package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) AppsByUser(ctx context.Context, userID string) ([]*model.App, error) {
	rows, err := c.Q.GetAppsByCollaborator(ctx, userID)
	if err != nil {
		return nil, err
	}

	var apps []*model.App
	for _, row := range rows {
		app, err := rowToApp(row)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
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

	return rowToApp(row)
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

	return rowToApp(row)
}

func (c *Client) UpdateApp(ctx context.Context, opts store.AppUpdateOpts) (*model.App, error) {
	var rawStatus []byte
	if opts.DiscordStatus != nil {
		var err error
		rawStatus, err = json.Marshal(opts.DiscordStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal discord status: %w", err)
		}
	}

	row, err := c.Q.UpdateApp(ctx, pgmodel.UpdateAppParams{
		ID:   opts.ID,
		Name: opts.Name,
		Description: pgtype.Text{
			String: opts.Description.String,
			Valid:  opts.Description.Valid,
		},
		DiscordToken:  opts.DiscordToken,
		DiscordStatus: rawStatus,
		Enabled:       opts.Enabled,
		DisabledReason: pgtype.Text{
			String: opts.DisabledReason.String,
			Valid:  opts.DisabledReason.Valid,
		},
		UpdatedAt: pgtype.Timestamp{Time: opts.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToApp(row)
}

func (c *Client) DisableApp(ctx context.Context, opts store.AppDisableOpts) error {
	return c.Q.DisableApp(ctx, pgmodel.DisableAppParams{
		ID: opts.ID,
		DisabledReason: pgtype.Text{
			String: opts.DisabledReason.String,
			Valid:  opts.DisabledReason.Valid,
		},
		UpdatedAt: pgtype.Timestamp{Time: opts.UpdatedAt.UTC(), Valid: true},
	})
}

func (c *Client) DeleteApp(ctx context.Context, id string) error {
	return c.Q.DeleteApp(ctx, id)
}

func (c *Client) EnabledAppIDs(ctx context.Context) ([]string, error) {
	return c.Q.GetEnabledAppIDs(ctx)
}

func (c *Client) EnabledAppsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.App, error) {
	rows, err := c.Q.GetEnabledAppsUpdatedSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	var apps []*model.App
	for _, row := range rows {
		app, err := rowToApp(row)
		if err != nil {
			return nil, err
		}
		apps = append(apps, app)
	}

	return apps, nil
}

func (c *Client) AppEntities(ctx context.Context, appID string) ([]*model.AppEntity, error) {
	rows, err := c.Q.GetAppEntities(ctx, appID)
	if err != nil {
		return nil, err
	}

	var entities []*model.AppEntity
	for _, row := range rows {
		entities = append(entities, &model.AppEntity{
			ID:   row.ID,
			Type: model.AppEntityType(row.Type),
			Name: row.Name,
		})
	}

	return entities, nil
}

func rowToApp(row pgmodel.App) (*model.App, error) {
	var status *model.AppDiscordStatus
	if row.DiscordStatus != nil {
		status = &model.AppDiscordStatus{}
		if err := json.Unmarshal(row.DiscordStatus, status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal discord status: %w", err)
		}
	}

	return &model.App{
		ID:             row.ID,
		Name:           row.Name,
		Description:    null.NewString(row.Description.String, row.Description.Valid),
		Enabled:        row.Enabled,
		DisabledReason: null.NewString(row.DisabledReason.String, row.DisabledReason.Valid),
		OwnerUserID:    row.OwnerUserID,
		CreatorUserID:  row.CreatorUserID,
		DiscordToken:   row.DiscordToken,
		DiscordID:      row.DiscordID,
		DiscordStatus:  status,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}, nil
}
