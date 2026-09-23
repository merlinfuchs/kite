-- name: GetMessage :one
SELECT * FROM messages WHERE id = $1;

-- name: GetMessageByApp :one
SELECT * FROM messages WHERE id = $1 AND app_id = $2;

-- name: GetMessagesByApp :many
SELECT * FROM messages WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountMessagesByApp :one
SELECT COUNT(*) FROM messages WHERE app_id = $1;

-- name: CreateMessage :one
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
) RETURNING *;

-- name: UpdateMessage :one
UPDATE messages SET
    name = $2,
    description = $3,
    data = $4,
    flow_sources = $5,
    updated_at = $6
WHERE id = $1 RETURNING *;

-- name: DeleteMessage :exec
DELETE FROM messages WHERE id = $1;

-- name: CreateMessageInstance :one
INSERT INTO message_instances (
    message_id,
    discord_guild_id,
    discord_channel_id,
    discord_message_id,
    ephemeral,
    hidden,
    flow_sources,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
-- Editing a message to a different template re-links it, so its new buttons resolve
ON CONFLICT (discord_message_id) DO UPDATE SET
    message_id = EXCLUDED.message_id,
    hidden = message_instances.hidden AND EXCLUDED.hidden,
    flow_sources = EXCLUDED.flow_sources,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetMessageInstance :one
SELECT * FROM message_instances WHERE id = $1 AND message_id = $2;

-- name: GetMessageInstancesByMessage :many
SELECT * FROM message_instances WHERE message_id = $1 AND NOT hidden ORDER BY created_at DESC;

-- name: GetFlowMessageInstancesByMessage :many
SELECT * FROM message_instances WHERE message_id = $1 AND hidden AND NOT ephemeral ORDER BY created_at DESC LIMIT $2;

-- name: GetMessageInstanceByDiscordMessageId :one
SELECT * FROM message_instances WHERE discord_message_id = $1;

-- name: UpdateMessageInstance :one
UPDATE message_instances SET
    flow_sources = $3,
    updated_at = $4
WHERE id = $1 AND message_id = $2 RETURNING *;

-- name: DeleteMessageInstance :exec
DELETE FROM message_instances WHERE id = $1 AND message_id = $2;

-- name: DeleteMessageInstanceByDiscordMessageId :exec
DELETE FROM message_instances WHERE discord_message_id = $1;