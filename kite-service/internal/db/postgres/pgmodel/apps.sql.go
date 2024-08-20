// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: apps.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countAppsByOwner = `-- name: CountAppsByOwner :one
SELECT COUNT(*) FROM apps WHERE owner_user_id = $1
`

func (q *Queries) CountAppsByOwner(ctx context.Context, ownerUserID string) (int64, error) {
	row := q.db.QueryRow(ctx, countAppsByOwner, ownerUserID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createApp = `-- name: CreateApp :one
INSERT INTO apps (
    id,
    name,
    description,
    owner_user_id,
    creator_user_id,
    discord_token,
    discord_id,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING id, name, description, enabled, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at
`

type CreateAppParams struct {
	ID            string
	Name          string
	Description   pgtype.Text
	OwnerUserID   string
	CreatorUserID string
	DiscordToken  string
	DiscordID     string
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

func (q *Queries) CreateApp(ctx context.Context, arg CreateAppParams) (App, error) {
	row := q.db.QueryRow(ctx, createApp,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.OwnerUserID,
		arg.CreatorUserID,
		arg.DiscordToken,
		arg.DiscordID,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i App
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Enabled,
		&i.OwnerUserID,
		&i.CreatorUserID,
		&i.DiscordToken,
		&i.DiscordID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteApp = `-- name: DeleteApp :exec
DELETE FROM apps WHERE id = $1
`

func (q *Queries) DeleteApp(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteApp, id)
	return err
}

const getApp = `-- name: GetApp :one
SELECT id, name, description, enabled, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at FROM apps WHERE id = $1
`

func (q *Queries) GetApp(ctx context.Context, id string) (App, error) {
	row := q.db.QueryRow(ctx, getApp, id)
	var i App
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Enabled,
		&i.OwnerUserID,
		&i.CreatorUserID,
		&i.DiscordToken,
		&i.DiscordID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getAppCredentials = `-- name: GetAppCredentials :one
SELECT discord_id, discord_token FROM apps WHERE id = $1
`

type GetAppCredentialsRow struct {
	DiscordID    string
	DiscordToken string
}

func (q *Queries) GetAppCredentials(ctx context.Context, id string) (GetAppCredentialsRow, error) {
	row := q.db.QueryRow(ctx, getAppCredentials, id)
	var i GetAppCredentialsRow
	err := row.Scan(&i.DiscordID, &i.DiscordToken)
	return i, err
}

const getAppsByOwner = `-- name: GetAppsByOwner :many
SELECT id, name, description, enabled, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at FROM apps WHERE owner_user_id = $1 ORDER BY created_at DESC
`

func (q *Queries) GetAppsByOwner(ctx context.Context, ownerUserID string) ([]App, error) {
	rows, err := q.db.Query(ctx, getAppsByOwner, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []App
	for rows.Next() {
		var i App
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Enabled,
			&i.OwnerUserID,
			&i.CreatorUserID,
			&i.DiscordToken,
			&i.DiscordID,
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

const getEnabledAppIDs = `-- name: GetEnabledAppIDs :many
SELECT id FROM apps WHERE enabled = TRUE
`

func (q *Queries) GetEnabledAppIDs(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, getEnabledAppIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEnabledAppsUpdatedSince = `-- name: GetEnabledAppsUpdatedSince :many
SELECT id, name, description, enabled, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at FROM apps WHERE enabled = TRUE AND updated_at > $1
`

func (q *Queries) GetEnabledAppsUpdatedSince(ctx context.Context, updatedAt pgtype.Timestamp) ([]App, error) {
	rows, err := q.db.Query(ctx, getEnabledAppsUpdatedSince, updatedAt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []App
	for rows.Next() {
		var i App
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Enabled,
			&i.OwnerUserID,
			&i.CreatorUserID,
			&i.DiscordToken,
			&i.DiscordID,
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

const updateApp = `-- name: UpdateApp :one
UPDATE apps SET
    name = $2,
    description = $3,
    discord_token = $4,
    enabled = $5,
    updated_at = $6
WHERE id = $1 RETURNING id, name, description, enabled, owner_user_id, creator_user_id, discord_token, discord_id, created_at, updated_at
`

type UpdateAppParams struct {
	ID           string
	Name         string
	Description  pgtype.Text
	DiscordToken string
	Enabled      bool
	UpdatedAt    pgtype.Timestamp
}

func (q *Queries) UpdateApp(ctx context.Context, arg UpdateAppParams) (App, error) {
	row := q.db.QueryRow(ctx, updateApp,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.DiscordToken,
		arg.Enabled,
		arg.UpdatedAt,
	)
	var i App
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Enabled,
		&i.OwnerUserID,
		&i.CreatorUserID,
		&i.DiscordToken,
		&i.DiscordID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
