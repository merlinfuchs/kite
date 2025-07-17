package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kitecloud/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func (c *Client) GetPluginValue(ctx context.Context, pluginInstanceID, key string) (*model.PluginValue, error) {
	row, err := c.Q.GetPluginValue(ctx, pgmodel.GetPluginValueParams{
		PluginInstanceID: pluginInstanceID,
		Key:              key,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToPluginValue(row)
}

func (c *Client) DeletePluginValue(ctx context.Context, pluginInstanceID, key string) error {
	err := c.Q.DeletePluginValue(ctx, pgmodel.DeletePluginValueParams{
		PluginInstanceID: pluginInstanceID,
		Key:              key,
	})

	return err
}

func (c *Client) SetPluginValue(ctx context.Context, value model.PluginValue) error {
	_, err := c.setPluginValueWithTx(ctx, nil, value)
	return err
}

func (c *Client) UpdatePluginValue(ctx context.Context, operation model.PluginValueOperation, value model.PluginValue) (*model.PluginValue, error) {
	if operation == provider.VariableOperationOverwrite {
		return c.setPluginValueWithTx(ctx, nil, value)
	}

	tx, err := c.DB.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	currentValue, err := c.pluginValueWithTx(ctx, tx, value.PluginInstanceID, value.Key)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// Current trasaction is rolled back, we set the value outside of the transaction
			return c.setPluginValueWithTx(ctx, nil, value)
		}
		return nil, fmt.Errorf("failed to get current plugin value: %w", err)
	}

	switch operation {
	case provider.VariableOperationAppend:
		value.Value = currentValue.Value.Append(value.Value)
	case provider.VariableOperationPrepend:
		value.Value = value.Value.Append(currentValue.Value)
	case provider.VariableOperationIncrement:
		value.Value = currentValue.Value.Add(value.Value)
	case provider.VariableOperationDecrement:
		value.Value = currentValue.Value.Sub(value.Value)
	}

	newValue, err := c.setPluginValueWithTx(ctx, tx, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set plugin value: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newValue, nil
}

func (c *Client) pluginValueWithTx(ctx context.Context, tx pgx.Tx, pluginInstanceID string, key string) (*model.PluginValue, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	row, err := q.GetPluginValueForUpdate(ctx, pgmodel.GetPluginValueForUpdateParams{
		PluginInstanceID: pluginInstanceID,
		Key:              key,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return rowToPluginValue(row)
}

func (c *Client) setPluginValueWithTx(ctx context.Context, tx pgx.Tx, value model.PluginValue) (*model.PluginValue, error) {
	q := c.Q
	if tx != nil {
		q = c.Q.WithTx(tx)
	}

	data, err := json.Marshal(value.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin value: %w", err)
	}

	row, err := q.SetPluginValue(ctx, pgmodel.SetPluginValueParams{
		PluginInstanceID: value.PluginInstanceID,
		Key:              value.Key,
		Value:            data,
		CreatedAt:        pgtype.Timestamp{Time: value.CreatedAt, Valid: true},
		UpdatedAt:        pgtype.Timestamp{Time: value.UpdatedAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return rowToPluginValue(row)
}

func rowToPluginValue(row pgmodel.PluginValue) (*model.PluginValue, error) {
	var data thing.Thing
	err := json.Unmarshal(row.Value, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal plugin value: %w", err)
	}

	return &model.PluginValue{
		ID:               uint64(row.ID),
		PluginInstanceID: row.PluginInstanceID,
		Key:              row.Key,
		Value:            data,
		CreatedAt:        row.CreatedAt.Time,
		UpdatedAt:        row.UpdatedAt.Time,
	}, nil
}
