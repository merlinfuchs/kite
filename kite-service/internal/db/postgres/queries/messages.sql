-- name: GetMessage :one
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

-- message_instances has no app_id, so these reach the app through messages.

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
-- Apps sharing a bot token must not re-link each other's instances
WHERE message_instances.message_id IN (SELECT id FROM messages WHERE app_id = $10)
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

-- name: GetFlowMessageInstancesByMessage :many
SELECT message_instances.* FROM message_instances
JOIN messages ON messages.id = message_instances.message_id
WHERE message_instances.message_id = $1 AND messages.app_id = $2
  AND message_instances.hidden AND NOT message_instances.ephemeral
ORDER BY message_instances.created_at DESC LIMIT $3;

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

-- name: TouchMessageInstance :exec
UPDATE message_instances SET last_used_at = @used_at
FROM messages
WHERE messages.id = message_instances.message_id
  AND message_instances.id = @id
  AND messages.app_id = @app_id;

-- name: DeleteUnusedMessageInstances :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM message_instances WHERE id IN (
    SELECT unused.id FROM message_instances unused
    WHERE (unused.hidden AND unused.last_used_at < @flow_used_before)
       OR unused.last_used_at < @dashboard_used_before
    LIMIT @batch_size
);
