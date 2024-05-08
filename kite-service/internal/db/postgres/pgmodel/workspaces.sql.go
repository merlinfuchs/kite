// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: workspaces.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createWorkspace = `-- name: CreateWorkspace :one
INSERT INTO workspaces (
    id,
    app_id,
    type,
    name,
    description,
    files,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
) RETURNING id, app_id, type, name, description, files, created_at, updated_at
`

type CreateWorkspaceParams struct {
	ID          string
	AppID       string
	Type        string
	Name        string
	Description string
	Files       []byte
	CreatedAt   pgtype.Timestamp
	UpdatedAt   pgtype.Timestamp
}

func (q *Queries) CreateWorkspace(ctx context.Context, arg CreateWorkspaceParams) (Workspace, error) {
	row := q.db.QueryRow(ctx, createWorkspace,
		arg.ID,
		arg.AppID,
		arg.Type,
		arg.Name,
		arg.Description,
		arg.Files,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i Workspace
	err := row.Scan(
		&i.ID,
		&i.AppID,
		&i.Type,
		&i.Name,
		&i.Description,
		&i.Files,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteWorkspace = `-- name: DeleteWorkspace :one
DELETE FROM workspaces WHERE id = $1 AND app_id = $2 RETURNING id, app_id, type, name, description, files, created_at, updated_at
`

type DeleteWorkspaceParams struct {
	ID    string
	AppID string
}

func (q *Queries) DeleteWorkspace(ctx context.Context, arg DeleteWorkspaceParams) (Workspace, error) {
	row := q.db.QueryRow(ctx, deleteWorkspace, arg.ID, arg.AppID)
	var i Workspace
	err := row.Scan(
		&i.ID,
		&i.AppID,
		&i.Type,
		&i.Name,
		&i.Description,
		&i.Files,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getWorkspaceForApp = `-- name: GetWorkspaceForApp :one
SELECT id, app_id, type, name, description, files, created_at, updated_at FROM workspaces WHERE id = $1 AND app_id = $2
`

type GetWorkspaceForAppParams struct {
	ID    string
	AppID string
}

func (q *Queries) GetWorkspaceForApp(ctx context.Context, arg GetWorkspaceForAppParams) (Workspace, error) {
	row := q.db.QueryRow(ctx, getWorkspaceForApp, arg.ID, arg.AppID)
	var i Workspace
	err := row.Scan(
		&i.ID,
		&i.AppID,
		&i.Type,
		&i.Name,
		&i.Description,
		&i.Files,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getWorkspacesForApp = `-- name: GetWorkspacesForApp :many
SELECT id, app_id, type, name, description, files, created_at, updated_at FROM workspaces WHERE app_id = $1 ORDER BY updated_at DESC
`

func (q *Queries) GetWorkspacesForApp(ctx context.Context, appID string) ([]Workspace, error) {
	rows, err := q.db.Query(ctx, getWorkspacesForApp, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Workspace
	for rows.Next() {
		var i Workspace
		if err := rows.Scan(
			&i.ID,
			&i.AppID,
			&i.Type,
			&i.Name,
			&i.Description,
			&i.Files,
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

const updateWorkspace = `-- name: UpdateWorkspace :one
UPDATE workspaces SET
    name = $3,
    description = $4,
    files = $5,
    updated_at = $6
WHERE 
    id = $1 AND 
    app_id = $2 
RETURNING id, app_id, type, name, description, files, created_at, updated_at
`

type UpdateWorkspaceParams struct {
	ID          string
	AppID       string
	Name        string
	Description string
	Files       []byte
	UpdatedAt   pgtype.Timestamp
}

func (q *Queries) UpdateWorkspace(ctx context.Context, arg UpdateWorkspaceParams) (Workspace, error) {
	row := q.db.QueryRow(ctx, updateWorkspace,
		arg.ID,
		arg.AppID,
		arg.Name,
		arg.Description,
		arg.Files,
		arg.UpdatedAt,
	)
	var i Workspace
	err := row.Scan(
		&i.ID,
		&i.AppID,
		&i.Type,
		&i.Name,
		&i.Description,
		&i.Files,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
