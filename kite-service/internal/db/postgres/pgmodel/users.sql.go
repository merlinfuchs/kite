// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.24.0
// source: users.sql

package pgmodel

import (
	"context"
	"database/sql"
	"time"
)

const getUser = `-- name: GetUser :one
SELECT id, username, discriminator, global_name, avatar, public_flags, created_at, updated_at FROM users WHERE id = $1
`

func (q *Queries) GetUser(ctx context.Context, id string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Discriminator,
		&i.GlobalName,
		&i.Avatar,
		&i.PublicFlags,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const upsertUser = `-- name: UpsertUser :one
INSERT INTO users (
    id,
    username,
    discriminator,
    global_name,
    avatar,
    public_flags,
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
) ON CONFLICT (id) DO UPDATE SET
    username = $2,
    discriminator = $3,
    global_name = $4,
    avatar = $5,
    public_flags = $6,
    updated_at = $8
RETURNING id, username, discriminator, global_name, avatar, public_flags, created_at, updated_at
`

type UpsertUserParams struct {
	ID            string
	Username      string
	Discriminator sql.NullString
	GlobalName    sql.NullString
	Avatar        sql.NullString
	PublicFlags   int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (q *Queries) UpsertUser(ctx context.Context, arg UpsertUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, upsertUser,
		arg.ID,
		arg.Username,
		arg.Discriminator,
		arg.GlobalName,
		arg.Avatar,
		arg.PublicFlags,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.Username,
		&i.Discriminator,
		&i.GlobalName,
		&i.Avatar,
		&i.PublicFlags,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}