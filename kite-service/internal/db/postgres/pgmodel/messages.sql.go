// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: messages.sql

package pgmodel

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const countMessagesByApp = `-- name: CountMessagesByApp :one
SELECT COUNT(*) FROM messages WHERE app_id = $1
`

func (q *Queries) CountMessagesByApp(ctx context.Context, appID string) (int64, error) {
	row := q.db.QueryRow(ctx, countMessagesByApp, appID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createMessage = `-- name: CreateMessage :one
INSERT INTO messages (
    id,
    name,
    description,
    app_id,
    module_id,
    creator_user_id,
    data,
    flow_sources,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING id, name, description, data, flow_sources, app_id, module_id, creator_user_id, created_at, updated_at
`

type CreateMessageParams struct {
	ID            string
	Name          string
	Description   pgtype.Text
	AppID         string
	ModuleID      pgtype.Text
	CreatorUserID string
	Data          []byte
	FlowSources   []byte
	CreatedAt     pgtype.Timestamp
	UpdatedAt     pgtype.Timestamp
}

func (q *Queries) CreateMessage(ctx context.Context, arg CreateMessageParams) (Message, error) {
	row := q.db.QueryRow(ctx, createMessage,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.AppID,
		arg.ModuleID,
		arg.CreatorUserID,
		arg.Data,
		arg.FlowSources,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i Message
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Data,
		&i.FlowSources,
		&i.AppID,
		&i.ModuleID,
		&i.CreatorUserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const createMessageInstance = `-- name: CreateMessageInstance :one
INSERT INTO message_instances (
    message_id,
    discord_guild_id,
    discord_channel_id,
    discord_message_id,
    flow_sources,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, message_id, hidden, ephemeral, discord_guild_id, discord_channel_id, discord_message_id, flow_sources, created_at, updated_at
`

type CreateMessageInstanceParams struct {
	MessageID        string
	DiscordGuildID   string
	DiscordChannelID string
	DiscordMessageID string
	FlowSources      []byte
	CreatedAt        pgtype.Timestamp
	UpdatedAt        pgtype.Timestamp
}

func (q *Queries) CreateMessageInstance(ctx context.Context, arg CreateMessageInstanceParams) (MessageInstance, error) {
	row := q.db.QueryRow(ctx, createMessageInstance,
		arg.MessageID,
		arg.DiscordGuildID,
		arg.DiscordChannelID,
		arg.DiscordMessageID,
		arg.FlowSources,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	var i MessageInstance
	err := row.Scan(
		&i.ID,
		&i.MessageID,
		&i.Hidden,
		&i.Ephemeral,
		&i.DiscordGuildID,
		&i.DiscordChannelID,
		&i.DiscordMessageID,
		&i.FlowSources,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteMessage = `-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1
`

func (q *Queries) DeleteMessage(ctx context.Context, id string) error {
	_, err := q.db.Exec(ctx, deleteMessage, id)
	return err
}

const deleteMessageInstance = `-- name: DeleteMessageInstance :exec
DELETE FROM message_instances WHERE id = $1 AND message_id = $2
`

type DeleteMessageInstanceParams struct {
	ID        int64
	MessageID string
}

func (q *Queries) DeleteMessageInstance(ctx context.Context, arg DeleteMessageInstanceParams) error {
	_, err := q.db.Exec(ctx, deleteMessageInstance, arg.ID, arg.MessageID)
	return err
}

const deleteMessageInstanceByDiscordMessageId = `-- name: DeleteMessageInstanceByDiscordMessageId :exec
DELETE FROM message_instances WHERE discord_message_id = $1
`

func (q *Queries) DeleteMessageInstanceByDiscordMessageId(ctx context.Context, discordMessageID string) error {
	_, err := q.db.Exec(ctx, deleteMessageInstanceByDiscordMessageId, discordMessageID)
	return err
}

const getMessage = `-- name: GetMessage :one
SELECT id, name, description, data, flow_sources, app_id, module_id, creator_user_id, created_at, updated_at FROM messages WHERE id = $1
`

func (q *Queries) GetMessage(ctx context.Context, id string) (Message, error) {
	row := q.db.QueryRow(ctx, getMessage, id)
	var i Message
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Data,
		&i.FlowSources,
		&i.AppID,
		&i.ModuleID,
		&i.CreatorUserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getMessageInstance = `-- name: GetMessageInstance :one
SELECT id, message_id, hidden, ephemeral, discord_guild_id, discord_channel_id, discord_message_id, flow_sources, created_at, updated_at FROM message_instances WHERE id = $1 AND message_id = $2
`

type GetMessageInstanceParams struct {
	ID        int64
	MessageID string
}

func (q *Queries) GetMessageInstance(ctx context.Context, arg GetMessageInstanceParams) (MessageInstance, error) {
	row := q.db.QueryRow(ctx, getMessageInstance, arg.ID, arg.MessageID)
	var i MessageInstance
	err := row.Scan(
		&i.ID,
		&i.MessageID,
		&i.Hidden,
		&i.Ephemeral,
		&i.DiscordGuildID,
		&i.DiscordChannelID,
		&i.DiscordMessageID,
		&i.FlowSources,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getMessageInstanceByDiscordMessageId = `-- name: GetMessageInstanceByDiscordMessageId :one
SELECT id, message_id, hidden, ephemeral, discord_guild_id, discord_channel_id, discord_message_id, flow_sources, created_at, updated_at FROM message_instances WHERE discord_message_id = $1
`

func (q *Queries) GetMessageInstanceByDiscordMessageId(ctx context.Context, discordMessageID string) (MessageInstance, error) {
	row := q.db.QueryRow(ctx, getMessageInstanceByDiscordMessageId, discordMessageID)
	var i MessageInstance
	err := row.Scan(
		&i.ID,
		&i.MessageID,
		&i.Hidden,
		&i.Ephemeral,
		&i.DiscordGuildID,
		&i.DiscordChannelID,
		&i.DiscordMessageID,
		&i.FlowSources,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getMessageInstancesByMessage = `-- name: GetMessageInstancesByMessage :many
SELECT id, message_id, hidden, ephemeral, discord_guild_id, discord_channel_id, discord_message_id, flow_sources, created_at, updated_at FROM message_instances WHERE message_id = $1 ORDER BY created_at DESC
`

func (q *Queries) GetMessageInstancesByMessage(ctx context.Context, messageID string) ([]MessageInstance, error) {
	rows, err := q.db.Query(ctx, getMessageInstancesByMessage, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MessageInstance
	for rows.Next() {
		var i MessageInstance
		if err := rows.Scan(
			&i.ID,
			&i.MessageID,
			&i.Hidden,
			&i.Ephemeral,
			&i.DiscordGuildID,
			&i.DiscordChannelID,
			&i.DiscordMessageID,
			&i.FlowSources,
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

const getMessagesByApp = `-- name: GetMessagesByApp :many
SELECT id, name, description, data, flow_sources, app_id, module_id, creator_user_id, created_at, updated_at FROM messages WHERE app_id = $1 ORDER BY created_at DESC
`

func (q *Queries) GetMessagesByApp(ctx context.Context, appID string) ([]Message, error) {
	rows, err := q.db.Query(ctx, getMessagesByApp, appID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Message
	for rows.Next() {
		var i Message
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Data,
			&i.FlowSources,
			&i.AppID,
			&i.ModuleID,
			&i.CreatorUserID,
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

const updateMessage = `-- name: UpdateMessage :one
UPDATE messages SET
    name = $2,
    description = $3,
    data = $4,
    flow_sources = $5,
    updated_at = $6
WHERE id = $1 RETURNING id, name, description, data, flow_sources, app_id, module_id, creator_user_id, created_at, updated_at
`

type UpdateMessageParams struct {
	ID          string
	Name        string
	Description pgtype.Text
	Data        []byte
	FlowSources []byte
	UpdatedAt   pgtype.Timestamp
}

func (q *Queries) UpdateMessage(ctx context.Context, arg UpdateMessageParams) (Message, error) {
	row := q.db.QueryRow(ctx, updateMessage,
		arg.ID,
		arg.Name,
		arg.Description,
		arg.Data,
		arg.FlowSources,
		arg.UpdatedAt,
	)
	var i Message
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Data,
		&i.FlowSources,
		&i.AppID,
		&i.ModuleID,
		&i.CreatorUserID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateMessageInstance = `-- name: UpdateMessageInstance :one
UPDATE message_instances SET
    flow_sources = $3,
    updated_at = $4
WHERE id = $1 AND message_id = $2 RETURNING id, message_id, hidden, ephemeral, discord_guild_id, discord_channel_id, discord_message_id, flow_sources, created_at, updated_at
`

type UpdateMessageInstanceParams struct {
	ID          int64
	MessageID   string
	FlowSources []byte
	UpdatedAt   pgtype.Timestamp
}

func (q *Queries) UpdateMessageInstance(ctx context.Context, arg UpdateMessageInstanceParams) (MessageInstance, error) {
	row := q.db.QueryRow(ctx, updateMessageInstance,
		arg.ID,
		arg.MessageID,
		arg.FlowSources,
		arg.UpdatedAt,
	)
	var i MessageInstance
	err := row.Scan(
		&i.ID,
		&i.MessageID,
		&i.Hidden,
		&i.Ephemeral,
		&i.DiscordGuildID,
		&i.DiscordChannelID,
		&i.DiscordMessageID,
		&i.FlowSources,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
