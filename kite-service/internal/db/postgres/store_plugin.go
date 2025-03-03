package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

func (c *Client) EnabledPluginInstanceIDs(ctx context.Context) (map[string][]string, error) {
	rows, err := c.Q.GetPartialEnabledPluginInstanceIDs(ctx)
	if err != nil {
		return nil, err
	}

	res := make(map[string][]string)
	for _, row := range rows {
		res[row.AppID] = append(res[row.AppID], row.PluginID)
	}
	return res, nil
}

func (c *Client) EnabledPluginInstancesUpdatedSince(ctx context.Context, lastUpdate time.Time) ([]*model.PluginInstance, error) {
	rows, err := c.Q.GetEnabledPluginInstancesUpdatedSince(ctx, pgtype.Timestamp{Time: lastUpdate, Valid: true})
	if err != nil {
		return nil, err
	}

	res := make([]*model.PluginInstance, len(rows))
	for i, row := range rows {
		res[i] = rowToPluginInstance(row)
	}

	return res, nil
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
	return rowToPluginInstance(row), nil
}

func (c *Client) UpsertPluginInstance(ctx context.Context, pluginInstance model.PluginInstance) (*model.PluginInstance, error) {
	row, err := c.Q.UpsertPluginInstance(ctx, pgmodel.UpsertPluginInstanceParams{
		AppID:     pluginInstance.AppID,
		PluginID:  pluginInstance.PluginID,
		Enabled:   pluginInstance.Enabled,
		Config:    pluginInstance.Config,
		CreatedAt: pgtype.Timestamp{Time: pluginInstance.CreatedAt, Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: pluginInstance.UpdatedAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToPluginInstance(row), nil
}

func (c *Client) DeletePluginInstance(ctx context.Context, appID string, pluginID string) error {
	return c.Q.DeletePluginInstance(ctx, pgmodel.DeletePluginInstanceParams{
		AppID:    appID,
		PluginID: pluginID,
	})
}

func rowToPluginInstance(row pgmodel.PluginInstance) *model.PluginInstance {
	return &model.PluginInstance{
		AppID:     row.AppID,
		PluginID:  row.PluginID,
		Enabled:   row.Enabled,
		Config:    row.Config,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
