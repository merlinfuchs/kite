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
