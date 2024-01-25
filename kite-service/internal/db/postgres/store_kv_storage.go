package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/merlinfuchs/kite/go-types/kvmodel"
	"github.com/merlinfuchs/kite/kite-service/internal/db/postgres/pgmodel"
	"github.com/merlinfuchs/kite/kite-service/pkg/model"
	"github.com/merlinfuchs/kite/kite-service/pkg/store"
)

func (c *Client) GetKVStorageNamespaces(ctx context.Context, guildID string) ([]model.KVStorageNamespace, error) {
	rows, err := c.Q.GetKVStorageNamespaces(ctx, guildID)
	if err != nil {
		return nil, err
	}

	res := make([]model.KVStorageNamespace, len(rows))
	for i, row := range rows {
		res[i] = model.KVStorageNamespace{
			GuildID:   guildID,
			Namespace: row.Namespace,
			KeyCount:  int(row.KeyCount),
		}
	}

	return res, nil
}

func (c *Client) GetKVStorageKeys(ctx context.Context, guildID, namespace string) ([]model.KVStorageValue, error) {
	rows, err := c.Q.GetKVStorageKeys(ctx, pgmodel.GetKVStorageKeysParams{
		GuildID:   guildID,
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	res := make([]model.KVStorageValue, len(rows))
	for i, row := range rows {
		var value kvmodel.TypedKVValue
		err = json.Unmarshal(row.Value, &value)
		if err != nil {
			return nil, err
		}

		res[i] = model.KVStorageValue{
			GuildID:   row.GuildID,
			Namespace: row.Namespace,
			Key:       row.Key,
			Value:     value,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		}
	}

	return res, nil
}

func (c *Client) SetKVStorageKey(ctx context.Context, guildID, namespace, key string, value kvmodel.TypedKVValue) error {
	rawValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = c.Q.SetKVStorageKey(ctx, pgmodel.SetKVStorageKeyParams{
		GuildID:   guildID,
		Namespace: namespace,
		Key:       key,
		Value:     rawValue,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetKVStorageKey(ctx context.Context, guildID, namespace, key string) (*model.KVStorageValue, error) {
	res, err := c.Q.GetKVStorageKey(ctx, pgmodel.GetKVStorageKeyParams{
		GuildID:   guildID,
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var value kvmodel.TypedKVValue
	err = json.Unmarshal(res.Value, &value)
	if err != nil {
		return nil, err
	}

	return &model.KVStorageValue{
		GuildID:   res.GuildID,
		Namespace: res.Namespace,
		Key:       res.Key,
		Value:     value,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil
}

func (c *Client) DeleteKVStorageKey(ctx context.Context, guildID, namespace, key string) (*model.KVStorageValue, error) {
	res, err := c.Q.DeleteKVStorageKey(ctx, pgmodel.DeleteKVStorageKeyParams{
		GuildID:   guildID,
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	var value kvmodel.TypedKVValue
	err = json.Unmarshal(res.Value, &value)
	if err != nil {
		return nil, err
	}

	return &model.KVStorageValue{
		GuildID:   res.GuildID,
		Namespace: res.Namespace,
		Key:       res.Key,
		Value:     value,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil
}

func (c *Client) IncreaseKVStorageKey(ctx context.Context, guildID, namespace, key string, increment int) (*model.KVStorageValue, error) {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		// We intentionally ignore the error because the transaction will already be commited
		_ = tx.Rollback()
	}()
	qtx := c.Q.WithTx(tx)

	var currentValue int
	res, err := qtx.GetKVStorageKey(ctx, pgmodel.GetKVStorageKeyParams{
		GuildID:   guildID,
		Namespace: namespace,
		Key:       key,
	})
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
	} else {
		var value kvmodel.TypedKVValue
		err = json.Unmarshal(res.Value, &value)
		if err != nil {
			return nil, err
		}

		currentValue = value.Value.Int()
	}

	newValue := kvmodel.TypedKVValue{
		Type:  kvmodel.KVValueTypeInt,
		Value: kvmodel.KVInt(currentValue + increment),
	}

	rawValue, err := json.Marshal(newValue)
	if err != nil {
		return nil, err
	}

	res, err = qtx.SetKVStorageKey(ctx, pgmodel.SetKVStorageKeyParams{
		GuildID:   guildID,
		Namespace: namespace,
		Key:       key,
		Value:     rawValue,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &model.KVStorageValue{
		GuildID:   guildID,
		Namespace: namespace,
		Key:       key,
		Value:     newValue,
		CreatedAt: res.CreatedAt,
		UpdatedAt: res.UpdatedAt,
	}, nil
}
