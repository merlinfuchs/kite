-- name: GetBundleInstallsByApp :many
SELECT * FROM bundle_installs WHERE app_id = $1 ORDER BY created_at DESC, id;

-- name: CreateBundleInstall :one
INSERT INTO bundle_installs (
    id,
    name,
    description,
    enabled,
    app_id,
    creator_user_id,
    bundle_id,
    bundle_version,
    settings,
    resources,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: GetBundleInstall :one
SELECT * FROM bundle_installs WHERE id = $1 AND app_id = $2;

-- Only updates an install that wasn't changed since from_updated_at, so two
-- changes at once can't both apply.
-- name: UpdateBundleInstall :one
UPDATE bundle_installs SET
    bundle_version = @bundle_version,
    settings = @settings,
    resources = @resources,
    updated_at = @updated_at
WHERE id = @id AND updated_at = @from_updated_at RETURNING *;

-- name: DeleteBundleInstall :exec
DELETE FROM bundle_installs WHERE id = $1;

-- name: GetInstalledCommandsByApp :many
SELECT * FROM commands WHERE app_id = $1 AND bundle_install_id IS NOT NULL;

-- name: GetInstalledEventListenersByApp :many
SELECT * FROM event_listeners WHERE app_id = $1 AND bundle_install_id IS NOT NULL;

-- name: GetInstalledMessagesByApp :many
SELECT * FROM messages WHERE app_id = $1 AND bundle_install_id IS NOT NULL;

-- name: GetInstalledVariablesByApp :many
SELECT * FROM variables WHERE app_id = $1 AND bundle_install_id IS NOT NULL;
