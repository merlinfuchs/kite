-- name: GetApp :one
SELECT * FROM apps WHERE id = $1;

-- name: GetAppCredentials :one
SELECT discord_id, discord_token FROM apps WHERE id = $1;

-- name: GetAppsByOwner :many
SELECT * FROM apps WHERE owner_user_id = $1 ORDER BY created_at DESC;

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
