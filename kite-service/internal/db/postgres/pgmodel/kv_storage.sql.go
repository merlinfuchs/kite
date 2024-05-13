// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: kv_storage.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteKVStorageKey = `-- name: DeleteKVStorageKey :one
DELETE FROM kv_storage WHERE app_id = $1 AND namespace = $2 AND key = $3 RETURNING app_id, namespace, key, value, created_at, updated_at
`

type DeleteKVStorageKeyParams struct {
	AppID     string
	Namespace string
	Key       string
}

func (q *Queries) DeleteKVStorageKey(ctx context.Context, arg DeleteKVStorageKeyParams) (KvStorage, error) {
	row := q.db.QueryRow(ctx, deleteKVStorageKey, arg.AppID, arg.Namespace, arg.Key)
	var i KvStorage
	err := row.Scan(
		&i.AppID,
		&i.Namespace,
		&i.Key,
		&i.Value,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getKVStorageKey = `-- name: GetKVStorageKey :one
SELECT app_id, namespace, key, value, created_at, updated_at FROM kv_storage WHERE app_id = $1 AND namespace = $2 AND key = $3
`

type GetKVStorageKeyParams struct {
	AppID     string
	Namespace string
	Key       string
}

func (q *Queries) GetKVStorageKey(ctx context.Context, arg GetKVStorageKeyParams) (KvStorage, error) {
	row := q.db.QueryRow(ctx, getKVStorageKey, arg.AppID, arg.Namespace, arg.Key)
	var i KvStorage
	err := row.Scan(
		&i.AppID,
		&i.Namespace,
		&i.Key,
		&i.Value,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getKVStorageKeys = `-- name: GetKVStorageKeys :many
SELECT app_id, namespace, key, value, created_at, updated_at FROM kv_storage WHERE app_id = $1 AND namespace = $2
`

type GetKVStorageKeysParams struct {
	AppID     string
	Namespace string
}

func (q *Queries) GetKVStorageKeys(ctx context.Context, arg GetKVStorageKeysParams) ([]KvStorage, error) {
	rows, err := q.db.Query(ctx, getKVStorageKeys, arg.AppID, arg.Namespace)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []KvStorage
	for rows.Next() {
		var i KvStorage
		if err := rows.Scan(
			&i.AppID,
			&i.Namespace,
			&i.Key,
			&i.Value,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getKVStorageNamespaces = `-- name: GetKVStorageNamespaces :many
SELECT  namespace, COUNT(key) as key_count FROM kv_storage WHERE app_id = $1 GROUP BY namespace
`

type GetKVStorageNamespacesRow struct {
	Namespace string
	KeyCount  int64
}

func (q *Queries) GetKVStorageNamespaces(ctx context.Context, appID string) ([]GetKVStorageNamespacesRow, error) {
	rows, err := q.db.Query(ctx, getKVStorageNamespaces, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetKVStorageNamespacesRow
	for rows.Next() {
		var i GetKVStorageNamespacesRow
		if err := rows.Scan(&i.Namespace, &i.KeyCount); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setKVStorageKey = `-- name: SetKVStorageKey :one
INSERT INTO kv_storage (
    app_id, 
    namespace, 
    key, 
    value, 
    created_at, 
    updated_at
) VALUES (
    $1, 
    $2, 
    $3, 
    $4, 
    $5, 
    $6
) ON CONFLICT (app_id, namespace, key) DO UPDATE SET 
    value = $4, 
    updated_at = $6
RETURNING app_id, namespace, key, value, created_at, updated_at
`

type SetKVStorageKeyParams struct {
	AppID     string
	Namespace string
	Key       string
	Value     []byte
	CreatedAt pgtype.Timestamp
	UpdatedAt pgtype.Timestamp
}

func (q *Queries) SetKVStorageKey(ctx context.Context, arg SetKVStorageKeyParams) (KvStorage, error) {
	row := q.db.QueryRow(ctx, setKVStorageKey,
		arg.AppID,
		arg.Namespace,
		arg.Key,
		arg.Value,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i KvStorage
	err := row.Scan(
		&i.AppID,
		&i.Namespace,
		&i.Key,
		&i.Value,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
