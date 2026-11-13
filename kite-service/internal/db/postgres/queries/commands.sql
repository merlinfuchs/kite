-- name: GetCommand :one
SELECT * FROM commands WHERE id = $1;

-- name: GetCommandsByApp :many
SELECT * FROM commands WHERE app_id = $1 ORDER BY created_at DESC, id;

-- name: CountCommandsByApp :one
SELECT
    COUNT(*) AS total,
    COUNT(*) FILTER (WHERE enabled) AS enabled
FROM commands WHERE app_id = $1;

-- name: CreateCommand :one
INSERT INTO commands (
    id,
    name,
    description,
    enabled,
    app_id,
    bundle_install_id,
    creator_user_id,
    flow_source,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: UpdateCommand :one
UPDATE commands SET
    name = $2,
    description = $3,
    flow_source = $4,
    updated_at = $5
WHERE id = $1 RETURNING *;

-- name: UpdateCommandEnabled :one
UPDATE commands SET
    enabled = $2,
    updated_at = $3
WHERE id = $1 RETURNING *;

-- name: GetEnabledCommandsUpdatesSince :many
SELECT * FROM commands WHERE enabled = TRUE AND updated_at > $1;

-- name: GetEnabledCommandIDs :many
SELECT id FROM commands WHERE enabled = TRUE;

-- name: DeleteCommand :one
DELETE FROM commands WHERE id = $1 RETURNING *;

-- name: DeleteCommandsByBundleInstall :many
DELETE FROM commands WHERE bundle_install_id = $1 RETURNING *;

-- name: GetCommandForUpdate :one
SELECT * FROM commands WHERE id = $1 FOR UPDATE;

-- Stopped apps are left alone, their commands aren't in Discord anyway.
-- name: GetAppsWithEnabledCommandsOver :many
SELECT c.app_id, COUNT(*) AS count FROM commands c
JOIN apps a ON a.id = c.app_id AND a.enabled
WHERE c.enabled
GROUP BY c.app_id
HAVING COUNT(*) > @min_count::INT;

-- Keeps the oldest enabled ones, so running it again changes nothing.
-- name: DisableNewestCommands :execrows
UPDATE commands SET enabled = FALSE, updated_at = @updated_at
WHERE id IN (
    SELECT c.id FROM commands c
    WHERE c.app_id = @app_id AND c.enabled
    ORDER BY c.created_at, c.id
    OFFSET @keep::INT
);
