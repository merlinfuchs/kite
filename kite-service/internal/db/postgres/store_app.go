package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) AllApps(ctx context.Context) ([]*model.App, error) {
	rows, err := c.Q.GetAllApps(ctx)
	if err != nil {
		return nil, err
	}

	apps := make([]*model.App, len(rows))
	for i, row := range rows {
		app, err := rowToApp(row)
		if err != nil {
			return nil, err
		}
		apps[i] = app
	}

	return apps, nil
}

func (c *Client) AppsByUser(ctx context.Context, userID string) ([]*model.App, error) {
	rows, err := c.Q.GetAppsByCollaborator(ctx, userID)
	if err != nil {
		return nil, err
	}

	apps := make([]*model.App, len(rows))
	for i, row := range rows {
		app, err := rowToApp(row)
		if err != nil {
			return nil, err
		}
		apps[i] = app
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
		DiscordVerifyKey: pgtype.Text{
			String: app.DiscordVerifyKey.String,
			Valid:  app.DiscordVerifyKey.Valid,
		},
		CreatedAt: pgtype.Timestamp{Time: app.CreatedAt.UTC(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: app.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		// apps.discord_id is unique, so each bot can only be added once.
		return nil, uniqueViolation(err)
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

func (c *Client) AppIDsWithGatewayRequirementsChangedSince(ctx context.Context, updatedSince time.Time) ([]string, error) {
	return c.Q.GetAppIDsWithGatewayRequirementsChangedSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
}

func (c *Client) AppGatewayRequirements(ctx context.Context, appID string) (*store.AppGatewayRequirementsRow, error) {
	row, err := c.Q.GetAppGatewayRequirements(ctx, appID)
	if err != nil {
		return nil, err
	}

	listenerTypes := make([]model.EventListenerType, len(row.EventListenerTypes))
	for i, t := range row.EventListenerTypes {
		listenerTypes[i] = model.EventListenerType(t)
	}

	return &store.AppGatewayRequirementsRow{
		EventListenerTypes:  listenerTypes,
		PluginResources:     row.PluginResources,
		HasMessageInstances: row.HasMessageInstances,
	}, nil
}

func (c *Client) DisabledAppIDsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]string, error) {
	return c.Q.GetDisabledAppIDsUpdatedSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
}

func (c *Client) EnabledAppsUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.App, error) {
	rows, err := c.Q.GetEnabledAppsUpdatedSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	return decodeRows(rows, rowToApp, func(row pgmodel.App) []any {
		return []any{slog.String("app_id", row.ID)}
	}), nil
}

func (c *Client) QueueAppCommandsDeploy(ctx context.Context, appID string, force bool) error {
	return c.Q.QueueAppCommandsDeploy(ctx, pgmodel.QueueAppCommandsDeployParams{
		ID:    appID,
		Force: force,
	})
}

func (c *Client) AppIDsWithDueCommandsDeploy(ctx context.Context, limit int) ([]string, error) {
	return c.Q.GetAppIDsWithDueCommandsDeploy(ctx, int32(limit))
}

func (c *Client) FinishAppCommandsDeploy(ctx context.Context, res store.AppCommandsDeployResult) error {
	rawError, err := marshalCommandsDeployError(res.Error)
	if err != nil {
		return err
	}

	return c.Q.FinishAppCommandsDeploy(ctx, pgmodel.FinishAppCommandsDeployParams{
		ID:           res.ID,
		DeployAt:     pgtype.Timestamp{Time: res.DeployAt, Valid: true},
		DeployedHash: pgtype.Text{String: res.DeployedHash.String, Valid: res.DeployedHash.Valid},
		Error:        rawError,
	})
}

func (c *Client) DelayAppCommandsDeploy(ctx context.Context, appID string, deployErr *model.CommandsDeployError) error {
	rawError, err := marshalCommandsDeployError(deployErr)
	if err != nil {
		return err
	}

	return c.Q.DelayAppCommandsDeploy(ctx, pgmodel.DelayAppCommandsDeployParams{
		ID:    appID,
		Error: rawError,
	})
}

func (c *Client) UpdateAppDiscordCounts(ctx context.Context, appID string, guildCount int, userInstallCount int) error {
	return c.Q.UpdateAppDiscordCounts(ctx, pgmodel.UpdateAppDiscordCountsParams{
		ID:                      appID,
		DiscordGuildCount:       pgtype.Int4{Int32: int32(guildCount), Valid: true},
		DiscordUserInstallCount: pgtype.Int4{Int32: int32(userInstallCount), Valid: true},
	})
}

func (c *Client) UpdateAppDiscordVerifyKey(ctx context.Context, appID string, verifyKey string) error {
	return c.Q.UpdateAppDiscordVerifyKey(ctx, pgmodel.UpdateAppDiscordVerifyKeyParams{
		ID:               appID,
		DiscordVerifyKey: pgtype.Text{String: verifyKey, Valid: true},
	})
}

func marshalCommandsDeployError(deployErr *model.CommandsDeployError) ([]byte, error) {
	if deployErr == nil {
		return nil, nil
	}
	raw, err := json.Marshal(deployErr)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal commands deploy error: %w", err)
	}
	return raw, nil
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

	var deployErr *model.CommandsDeployError
	if row.CommandsDeployError != nil {
		deployErr = &model.CommandsDeployError{}
		if err := json.Unmarshal(row.CommandsDeployError, deployErr); err != nil {
			return nil, fmt.Errorf("failed to unmarshal commands deploy error: %w", err)
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

		CommandsDeployAt:     null.NewTime(row.CommandsDeployAt.Time, row.CommandsDeployAt.Valid),
		CommandsDeployedHash: null.NewString(row.CommandsDeployedHash.String, row.CommandsDeployedHash.Valid),
		CommandsDeployError:  deployErr,

		DiscordGuildCount:       null.NewInt(int64(row.DiscordGuildCount.Int32), row.DiscordGuildCount.Valid),
		DiscordUserInstallCount: null.NewInt(int64(row.DiscordUserInstallCount.Int32), row.DiscordUserInstallCount.Valid),
		DiscordCountsUpdatedAt:  null.NewTime(row.DiscordCountsUpdatedAt.Time, row.DiscordCountsUpdatedAt.Valid),
		DiscordVerifyKey:        null.NewString(row.DiscordVerifyKey.String, row.DiscordVerifyKey.Valid),
	}, nil
}
