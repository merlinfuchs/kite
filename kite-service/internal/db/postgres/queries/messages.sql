-- name: GetMessage :one
-- Scoped by app: the message ID reaches the engine from flow node config that
-- the tenant authored, so an ID on its own is not proof of ownership.
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

-- message_instances has no app_id of its own, so the queries below reach the app
-- through messages. The engine reaches these from Discord interactions, where
-- the message and instance IDs are supplied by whoever clicked.

-- name: CreateMessageInstance :one
-- INSERT ... SELECT rather than VALUES so the app check is part of the write:
-- if the message does not belong to the app the select is empty, nothing is
-- inserted, and :one surfaces it as ErrNoRows.
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
)
SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9
FROM messages WHERE messages.id = $1 AND messages.app_id = $10
RETURNING *;

-- name: GetMessageInstance :one
SELECT message_instances.* FROM message_instances
JOIN messages ON messages.id = message_instances.message_id
WHERE message_instances.id = $1
  AND message_instances.message_id = $2
  AND messages.app_id = $3;

-- name: GetMessageInstancesByMessage :many
SELECT message_instances.* FROM message_instances
JOIN messages ON messages.id = message_instances.message_id
WHERE message_instances.message_id = $1 AND messages.app_id = $2 AND NOT message_instances.hidden
ORDER BY message_instances.created_at DESC;

-- name: GetMessageInstancesByMessageWithHidden :many
SELECT message_instances.* FROM message_instances
JOIN messages ON messages.id = message_instances.message_id
WHERE message_instances.message_id = $1 AND messages.app_id = $2
ORDER BY message_instances.created_at DESC;

-- name: GetMessageInstanceByDiscordMessageId :one
SELECT message_instances.* FROM message_instances
JOIN messages ON messages.id = message_instances.message_id
WHERE message_instances.discord_message_id = $1 AND messages.app_id = $2;

-- name: UpdateMessageInstance :one
UPDATE message_instances SET
    flow_sources = $4,
    updated_at = $5
FROM messages
WHERE messages.id = message_instances.message_id
  AND message_instances.id = $1
  AND message_instances.message_id = $2
  AND messages.app_id = $3
RETURNING message_instances.*;

-- name: DeleteMessageInstance :exec
DELETE FROM message_instances
USING messages
WHERE messages.id = message_instances.message_id
  AND message_instances.id = $1
  AND message_instances.message_id = $2
  AND messages.app_id = $3;

-- name: DeleteMessageInstanceByDiscordMessageId :exec
DELETE FROM message_instances
USING messages
WHERE messages.id = message_instances.message_id
  AND message_instances.discord_message_id = $1
  AND messages.app_id = $2;