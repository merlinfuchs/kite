// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: usage_records.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createUsageRecord = `-- name: CreateUsageRecord :exec
INSERT INTO usage_records (
    type,
    app_id,
    command_id,
    event_listener_id,
    message_id,
    credits_used,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, type, app_id, command_id, event_listener_id, message_id, credits_used, created_at
`

type CreateUsageRecordParams struct {
	Type            string
	AppID           string
	CommandID       pgtype.Text
	EventListenerID pgtype.Text
	MessageID       pgtype.Text
	CreditsUsed     int32
	CreatedAt       pgtype.Timestamp
}

func (q *Queries) CreateUsageRecord(ctx context.Context, arg CreateUsageRecordParams) error {
	_, err := q.db.Exec(ctx, createUsageRecord,
		arg.Type,
		arg.AppID,
		arg.CommandID,
		arg.EventListenerID,
		arg.MessageID,
		arg.CreditsUsed,
		arg.CreatedAt,
	)
	return err
}

const getUsageCreditsUsedByAppBetween = `-- name: GetUsageCreditsUsedByAppBetween :one
SELECT SUM(credits_used) FROM usage_records WHERE app_id = $1 AND created_at BETWEEN $2 AND $3
`

type GetUsageCreditsUsedByAppBetweenParams struct {
	AppID       string
	CreatedAt   pgtype.Timestamp
	CreatedAt_2 pgtype.Timestamp
}

func (q *Queries) GetUsageCreditsUsedByAppBetween(ctx context.Context, arg GetUsageCreditsUsedByAppBetweenParams) (int64, error) {
	row := q.db.QueryRow(ctx, getUsageCreditsUsedByAppBetween, arg.AppID, arg.CreatedAt, arg.CreatedAt_2)
	var sum int64
	err := row.Scan(&sum)
	return sum, err
}

const getUsageRecordsByAppBetween = `-- name: GetUsageRecordsByAppBetween :many
SELECT id, type, app_id, command_id, event_listener_id, message_id, credits_used, created_at FROM usage_records WHERE app_id = $1 AND created_at BETWEEN $2 AND $3 ORDER BY created_at DESC
`

type GetUsageRecordsByAppBetweenParams struct {
	AppID       string
	CreatedAt   pgtype.Timestamp
	CreatedAt_2 pgtype.Timestamp
}

func (q *Queries) GetUsageRecordsByAppBetween(ctx context.Context, arg GetUsageRecordsByAppBetweenParams) ([]UsageRecord, error) {
	rows, err := q.db.Query(ctx, getUsageRecordsByAppBetween, arg.AppID, arg.CreatedAt, arg.CreatedAt_2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UsageRecord
	for rows.Next() {
		var i UsageRecord
		if err := rows.Scan(
			&i.ID,
			&i.Type,
			&i.AppID,
			&i.CommandID,
			&i.EventListenerID,
			&i.MessageID,
			&i.CreditsUsed,
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
