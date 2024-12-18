// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: logs.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createLogEntry = `-- name: CreateLogEntry :exec
INSERT INTO logs (app_id, message, level, created_at) VALUES ($1, $2, $3, $4) RETURNING id, app_id, message, level, created_at
`

type CreateLogEntryParams struct {
	AppID     string
	Message   string
	Level     string
	CreatedAt pgtype.Timestamp
}

func (q *Queries) CreateLogEntry(ctx context.Context, arg CreateLogEntryParams) error {
	_, err := q.db.Exec(ctx, createLogEntry,
		arg.AppID,
		arg.Message,
		arg.Level,
		arg.CreatedAt,
	)
	return err
}

const getLogEntriesByApp = `-- name: GetLogEntriesByApp :many
SELECT id, app_id, message, level, created_at FROM logs WHERE app_id = $1 ORDER BY created_at DESC LIMIT $2
`

type GetLogEntriesByAppParams struct {
	AppID string
	Limit int32
}

func (q *Queries) GetLogEntriesByApp(ctx context.Context, arg GetLogEntriesByAppParams) ([]Log, error) {
	rows, err := q.db.Query(ctx, getLogEntriesByApp, arg.AppID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Log
	for rows.Next() {
		var i Log
		if err := rows.Scan(
			&i.ID,
			&i.AppID,
			&i.Message,
			&i.Level,
			&i.CreatedAt,
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

const getLogEntriesByAppBefore = `-- name: GetLogEntriesByAppBefore :many
SELECT id, app_id, message, level, created_at FROM logs WHERE app_id = $1 AND id < $2 ORDER BY created_at DESC LIMIT $3
`

type GetLogEntriesByAppBeforeParams struct {
	AppID string
	ID    int64
	Limit int32
}

func (q *Queries) GetLogEntriesByAppBefore(ctx context.Context, arg GetLogEntriesByAppBeforeParams) ([]Log, error) {
	rows, err := q.db.Query(ctx, getLogEntriesByAppBefore, arg.AppID, arg.ID, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Log
	for rows.Next() {
		var i Log
		if err := rows.Scan(
			&i.ID,
			&i.AppID,
			&i.Message,
			&i.Level,
			&i.CreatedAt,
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

const getLogSummary = `-- name: GetLogSummary :one
SELECT COUNT(*) AS total_entries,
       SUM(CASE WHEN level = 'error' THEN 1 ELSE 0 END) AS total_errors,
       SUM(CASE WHEN level = 'warn' THEN 1 ELSE 0 END) AS total_warnings,
       SUM(CASE WHEN level = 'info' THEN 1 ELSE 0 END) AS total_infos,
       SUM(CASE WHEN level = 'debug' THEN 1 ELSE 0 END) AS total_debugs
FROM logs WHERE app_id = $1 AND created_at >= $2 AND created_at < $3
`

type GetLogSummaryParams struct {
	AppID   string
	StartAt pgtype.Timestamp
	EndAt   pgtype.Timestamp
}

type GetLogSummaryRow struct {
	TotalEntries  int64
	TotalErrors   int64
	TotalWarnings int64
	TotalInfos    int64
	TotalDebugs   int64
}

func (q *Queries) GetLogSummary(ctx context.Context, arg GetLogSummaryParams) (GetLogSummaryRow, error) {
	row := q.db.QueryRow(ctx, getLogSummary, arg.AppID, arg.StartAt, arg.EndAt)
	var i GetLogSummaryRow
	err := row.Scan(
		&i.TotalEntries,
		&i.TotalErrors,
		&i.TotalWarnings,
		&i.TotalInfos,
		&i.TotalDebugs,
	)
	return i, err
}
