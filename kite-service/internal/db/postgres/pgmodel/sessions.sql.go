// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.22.0
// source: sessions.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createPendingSession = `-- name: CreatePendingSession :exec
INSERT INTO sessions_pending (
    code,
    created_at,
    expires_at
) VALUES (
    $1,
    $2,
    $3
)
`

type CreatePendingSessionParams struct {
	Code      string
	CreatedAt pgtype.Timestamp
	ExpiresAt pgtype.Timestamp
}

func (q *Queries) CreatePendingSession(ctx context.Context, arg CreatePendingSessionParams) error {
	_, err := q.db.Exec(ctx, createPendingSession, arg.Code, arg.CreatedAt, arg.ExpiresAt)
	return err
}

const createSession = `-- name: CreateSession :exec
INSERT INTO sessions (
    token_hash,
    type,
    user_id,
    access_token,
    created_at,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
`

type CreateSessionParams struct {
	TokenHash   string
	Type        string
	UserID      string
	AccessToken string
	CreatedAt   pgtype.Timestamp
	ExpiresAt   pgtype.Timestamp
}

func (q *Queries) CreateSession(ctx context.Context, arg CreateSessionParams) error {
	_, err := q.db.Exec(ctx, createSession,
		arg.TokenHash,
		arg.Type,
		arg.UserID,
		arg.AccessToken,
		arg.CreatedAt,
		arg.ExpiresAt,
	)
	return err
}

const deleteExpiredPendingSessions = `-- name: DeleteExpiredPendingSessions :exec
DELETE FROM sessions_pending WHERE expires_at < $1
`

func (q *Queries) DeleteExpiredPendingSessions(ctx context.Context, expiresAt pgtype.Timestamp) error {
	_, err := q.db.Exec(ctx, deleteExpiredPendingSessions, expiresAt)
	return err
}

const deleteSession = `-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1
`

func (q *Queries) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := q.db.Exec(ctx, deleteSession, tokenHash)
	return err
}

const getPendingSession = `-- name: GetPendingSession :one
SELECT code, token, created_at, expires_at FROM sessions_pending WHERE code = $1
`

func (q *Queries) GetPendingSession(ctx context.Context, code string) (SessionsPending, error) {
	row := q.db.QueryRow(ctx, getPendingSession, code)
	var i SessionsPending
	err := row.Scan(
		&i.Code,
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const getSession = `-- name: GetSession :one
SELECT token_hash, type, user_id, access_token, revoked, created_at, expires_at FROM sessions WHERE token_hash = $1
`

func (q *Queries) GetSession(ctx context.Context, tokenHash string) (Session, error) {
	row := q.db.QueryRow(ctx, getSession, tokenHash)
	var i Session
	err := row.Scan(
		&i.TokenHash,
		&i.Type,
		&i.UserID,
		&i.AccessToken,
		&i.Revoked,
		&i.CreatedAt,
		&i.ExpiresAt,
	)
	return i, err
}

const updatePendingSession = `-- name: UpdatePendingSession :one
UPDATE sessions_pending SET token = $1, expires_at = $2 WHERE code = $3 RETURNING code, token, created_at, expires_at
`

type UpdatePendingSessionParams struct {
	Token     pgtype.Text
	ExpiresAt pgtype.Timestamp
	Code      string
}

func (q *Queries) UpdatePendingSession(ctx context.Context, arg UpdatePendingSessionParams) (SessionsPending, error) {
	row := q.db.QueryRow(ctx, updatePendingSession, arg.Token, arg.ExpiresAt, arg.Code)
	var i SessionsPending
	err := row.Scan(
		&i.Code,
		&i.Token,
		&i.CreatedAt,
		&i.ExpiresAt,
	)
	return i, err
}
