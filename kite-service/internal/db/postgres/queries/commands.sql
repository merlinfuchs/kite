-- name: GetCommand :one
SELECT * FROM commands WHERE id = $1;

-- name: GetCommandsByApp :many
SELECT * FROM commands WHERE app_id = $1 ORDER BY created_at DESC, id;

-- name: CountCommandsByApp :one
SELECT COUNT(*) FROM commands WHERE app_id = $1;

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
    enabled = $4,
    flow_source = $5,
    updated_at = $6
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
