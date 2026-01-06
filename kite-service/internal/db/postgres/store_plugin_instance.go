package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
	"gopkg.in/guregu/null.v4"
)

func (c *Client) PluginInstancesByApp(ctx context.Context, appID string) ([]*model.PluginInstance, error) {
	rows, err := c.Q.GetPluginInstancesByApp(ctx, appID)
	if err != nil {
		return nil, err
	}

	instances := make([]*model.PluginInstance, len(rows))
	for i, row := range rows {
		instance, err := rowToPluginInstance(row)
		if err != nil {
			return nil, err
		}

		instances[i] = instance
	}

	return instances, nil
}

func (c *Client) CountPluginInstancesByApp(ctx context.Context, appID string) (int, error) {
	res, err := c.Q.CountPluginInstancesByApp(ctx, appID)
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

func (c *Client) PluginInstance(ctx context.Context, appID string, pluginID string) (*model.PluginInstance, error) {
	row, err := c.Q.GetPluginInstance(ctx, pgmodel.GetPluginInstanceParams{
		AppID:    appID,
		PluginID: pluginID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToPluginInstance(row)
}

func (c *Client) CreatePluginInstance(ctx context.Context, instance *model.PluginInstance) (*model.PluginInstance, error) {
	config, err := json.Marshal(instance.Config)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.CreatePluginInstance(ctx, pgmodel.CreatePluginInstanceParams{
		ID:                 instance.ID,
		PluginID:           instance.PluginID,
		Enabled:            instance.Enabled,
		AppID:              instance.AppID,
		CreatorUserID:      instance.CreatorUserID,
		Config:             config,
		EnabledResourceIds: instance.EnabledResourceIDs,
		CreatedAt:          pgtype.Timestamp{Time: instance.CreatedAt.UTC(), Valid: true},
		UpdatedAt:          pgtype.Timestamp{Time: instance.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToPluginInstance(row)
}

func (c *Client) UpdatePluginInstance(ctx context.Context, instance *model.PluginInstance) (*model.PluginInstance, error) {
	config, err := json.Marshal(instance.Config)
	if err != nil {
		return nil, err
	}

	row, err := c.Q.UpdatePluginInstance(ctx, pgmodel.UpdatePluginInstanceParams{
		AppID:              instance.AppID,
		PluginID:           instance.PluginID,
		Enabled:            instance.Enabled,
		Config:             config,
		EnabledResourceIds: instance.EnabledResourceIDs,
		UpdatedAt:          pgtype.Timestamp{Time: instance.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToPluginInstance(row)
}

func (c *Client) UpdatePluginInstancesLastDeployedAt(ctx context.Context, appID string, lastDeployedAt time.Time) error {
	return c.Q.UpdatePluginInstancesLastDeployedAt(ctx, pgmodel.UpdatePluginInstancesLastDeployedAtParams{
		AppID:          appID,
		LastDeployedAt: pgtype.Timestamp{Time: lastDeployedAt.UTC(), Valid: true},
	})
}

func (c *Client) EnabledPluginInstancesUpdatedSince(ctx context.Context, updatedSince time.Time) ([]*model.PluginInstance, error) {
	rows, err := c.Q.GetEnabledPluginInstancesUpdatesSince(ctx, pgtype.Timestamp{
		Time:  updatedSince.UTC(),
		Valid: true,
	})
	if err != nil {
		return nil, err
	}

	instances := make([]*model.PluginInstance, len(rows))
	for i, row := range rows {
		instance, err := rowToPluginInstance(row)
		if err != nil {
			return nil, err
		}

		instances[i] = instance
	}

	return instances, nil
}

func (c *Client) EnabledPluginInstanceIDs(ctx context.Context) ([]string, error) {
	return c.Q.GetEnabledPluginInstanceIDs(ctx)
}

func (c *Client) DeletePluginInstance(ctx context.Context, appID string, pluginID string) error {
	err := c.Q.DeletePluginInstance(ctx, pgmodel.DeletePluginInstanceParams{
		AppID:    appID,
		PluginID: pluginID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.ErrNotFound
		}

		return err
	}

	return nil
}

func (c *Client) DinstinctAppIDsWithUndeployedPluginInstances(ctx context.Context) ([]string, error) {
	return c.Q.DinstinctAppIDsWithUndeployedPluginInstances(ctx)
}

func rowToPluginInstance(row pgmodel.PluginInstance) (*model.PluginInstance, error) {
	var config plugin.ConfigValues
	if err := json.Unmarshal(row.Config, &config); err != nil {
		return nil, err
	}

	return &model.PluginInstance{
		ID:                 row.ID,
		PluginID:           row.PluginID,
		Enabled:            row.Enabled,
		AppID:              row.AppID,
		CreatorUserID:      row.CreatorUserID,
		Config:             config,
		EnabledResourceIDs: row.EnabledResourceIds,
		CreatedAt:          row.CreatedAt.Time,
		UpdatedAt:          row.UpdatedAt.Time,
		LastDeployedAt:     null.NewTime(row.LastDeployedAt.Time, row.LastDeployedAt.Valid),
	}, nil
}
