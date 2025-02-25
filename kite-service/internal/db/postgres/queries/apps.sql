-- name: GetApp :one
SELECT * FROM apps WHERE id = $1;

-- name: GetAppCredentials :one
SELECT discord_id, discord_token FROM apps WHERE id = $1;

-- name: GetAppsByOwner :many
SELECT * FROM apps WHERE owner_user_id = $1 ORDER BY created_at DESC;

-- name: GetAppsByCollaborator :many
SELECT a.* FROM apps a
LEFT JOIN collaborators c ON a.id = c.app_id
WHERE a.owner_user_id = @user_id OR c.user_id = @user_id
ORDER BY a.created_at DESC;

-- name: CountAppsByOwner :one
SELECT COUNT(*) FROM apps WHERE owner_user_id = $1;

-- name: CreateApp :one
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
) RETURNING *;

-- name: UpdateApp :one
UPDATE apps SET
    name = $2,
    description = $3,
    discord_token = $4,
    discord_status = $5,
    enabled = $6,
    disabled_reason = $7,
    updated_at = $8
WHERE id = $1 RETURNING *;

-- name: DisableApp :exec
UPDATE apps SET
    enabled = FALSE,
    disabled_reason = $2,
    updated_at = $3
WHERE id = $1;

-- name: DeleteApp :exec
DELETE FROM apps WHERE id = $1;

-- name: GetEnabledAppIDs :many
SELECT id FROM apps WHERE enabled = TRUE;

-- name: GetEnabledAppsUpdatedSince :many
SELECT * FROM apps WHERE enabled = TRUE AND updated_at > $1;

-- name: GetAppEntities :many
SELECT 
    id,
    'command' AS type,
    name
FROM commands
WHERE commands.app_id = $1
UNION ALL
SELECT 
    id,
    'event_listener' AS type,
    type as name
FROM event_listeners
WHERE event_listeners.app_id = $1
UNION ALL
SELECT 
    id,
    'message' AS type,
    name
FROM messages
WHERE messages.app_id = $1
UNION ALL
SELECT 
    id,
    'variable' AS type,
    name
FROM variables
WHERE variables.app_id = $1;