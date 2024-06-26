package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/merlinfuchs/kite/kite-sdk-go/kv"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

var _ store.KVStorageStore = (*Client)(nil)

func (c *Client) GetKVStorageNamespaces(ctx context.Context, appID string) ([]model.KVStorageNamespace, error) {
	rows, err := c.Q.GetKVStorageNamespaces(ctx, appID)
	if err != nil {
		return nil, err
	}

	res := make([]model.KVStorageNamespace, len(rows))
	for i, row := range rows {
		res[i] = model.KVStorageNamespace{
			AppID:     appID,
			Namespace: row.Namespace,
			KeyCount:  int(row.KeyCount),
		}
	}

	return res, nil
}

func (c *Client) GetKVStorageKeys(ctx context.Context, appID, namespace string) ([]model.KVStorageValue, error) {
	rows, err := c.Q.GetKVStorageKeys(ctx, pgmodel.GetKVStorageKeysParams{
		AppID:     appID,
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.KVStorageValue, len(rows))
	for i, row := range rows {
		var value kv.TypedKVValue
		err = json.Unmarshal(row.Value, &value)
		if err != nil {
			return nil, err
		}

		res[i] = model.KVStorageValue{
			AppID:     row.AppID,
			Namespace: row.Namespace,
			Key:       row.Key,
			Value:     value,
			CreatedAt: row.CreatedAt.Time,
			UpdatedAt: row.UpdatedAt.Time,
		}
	}

	return res, nil
}

func (c *Client) SetKVStorageKey(ctx context.Context, appID, namespace, key string, value kv.TypedKVValue) error {
	rawValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.Q.SetKVStorageKey(ctx, pgmodel.SetKVStorageKeyParams{
		AppID:     appID,
		Namespace: namespace,
		Key:       key,
		Value:     rawValue,
		CreatedAt: timeToTimestamp(time.Now().UTC()),
		UpdatedAt: timeToTimestamp(time.Now().UTC()),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetKVStorageKey(ctx context.Context, appID, namespace, key string) (*model.KVStorageValue, error) {
	res, err := c.Q.GetKVStorageKey(ctx, pgmodel.GetKVStorageKeyParams{
		AppID:     appID,
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var value kv.TypedKVValue
	err = json.Unmarshal(res.Value, &value)
	if err != nil {
		return nil, err
	}

	return &model.KVStorageValue{
		AppID:     res.AppID,
		Namespace: res.Namespace,
		Key:       res.Key,
		Value:     value,
		CreatedAt: res.CreatedAt.Time,
		UpdatedAt: res.UpdatedAt.Time,
	}, nil
}

func (c *Client) DeleteKVStorageKey(ctx context.Context, appID, namespace, key string) (*model.KVStorageValue, error) {
	res, err := c.Q.DeleteKVStorageKey(ctx, pgmodel.DeleteKVStorageKeyParams{
		AppID:     appID,
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var value kv.TypedKVValue
	err = json.Unmarshal(res.Value, &value)
	if err != nil {
		return nil, err
	}

	return &model.KVStorageValue{
		AppID:     res.AppID,
		Namespace: res.Namespace,
		Key:       res.Key,
		Value:     value,
		CreatedAt: res.CreatedAt.Time,
		UpdatedAt: res.UpdatedAt.Time,
	}, nil
}

func (c *Client) IncreaseKVStorageKey(ctx context.Context, appID, namespace, key string, increment int) (*model.KVStorageValue, error) {
	tx, err := c.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() {
		rerr := tx.Rollback(ctx)
		if rerr != nil && !errors.Is(rerr, pgx.ErrTxClosed) {
			slog.With(logattr.Error(rerr)).Error("failed to rollback kv increase transaction")
		}
	}()
	qtx := c.Q.WithTx(tx)

	var currentValue int
	res, err := qtx.GetKVStorageKey(ctx, pgmodel.GetKVStorageKeyParams{
		AppID:     appID,
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
	} else {
		var value kv.TypedKVValue
		err = json.Unmarshal(res.Value, &value)
		if err != nil {
			return nil, err
		}

		currentValue = value.Value.Int()
	}

	newValue := kv.TypedKVValue{
		Type:  kv.KVValueTypeInt,
		Value: kv.KVInt(currentValue + increment),
	}

	rawValue, err := json.Marshal(newValue)
	if err != nil {
		return nil, err
	}

	res, err = qtx.SetKVStorageKey(ctx, pgmodel.SetKVStorageKeyParams{
		AppID:     appID,
		Namespace: namespace,
		Key:       key,
		Value:     rawValue,
		CreatedAt: timeToTimestamp(time.Now().UTC()),
		UpdatedAt: timeToTimestamp(time.Now().UTC()),
	})

	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &model.KVStorageValue{
		AppID:     appID,
		Namespace: namespace,
		Key:       key,
		Value:     newValue,
		CreatedAt: res.CreatedAt.Time,
		UpdatedAt: res.UpdatedAt.Time,
	}, nil
}
