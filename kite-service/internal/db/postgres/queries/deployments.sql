-- name: UpsertDeployment :one
INSERT INTO deployments (
    id,
    key, 
    name, 
    description, 
    app_id, 
    plugin_version_id, 
    wasm_bytes, 
    manifest,
    config, 
    created_at, 
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
) ON CONFLICT (key, app_id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    plugin_version_id = EXCLUDED.plugin_version_id,
    wasm_bytes = EXCLUDED.wasm_bytes,
    manifest = EXCLUDED.manifest,
    config = EXCLUDED.config,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetAppIdsWithDeployments :many
SELECT DISTINCT app_id FROM deployments;

-- name: GetDeployments :many
SELECT * FROM deployments;

-- name: GetDeploymentsForApp :many
SELECT * FROM deployments WHERE app_id = $1 ORDER BY updated_at DESC;

-- name: GetDeploymentsWithUndeployedChanges :many
SELECT * FROM deployments WHERE deployed_at IS NULL OR updated_at > deployed_at;

-- name: GetDeploymentIDs :many
SELECT id, app_id FROM deployments;

-- name: GetDeploymentForApp :one
SELECT * FROM deployments WHERE id = $1 AND app_id = $2;

-- name: UpdateDeploymentsDeployedAtForApp :one
UPDATE deployments SET deployed_at = $1 WHERE app_id = $2 RETURNING *;

-- name: DeleteDeployment :one
DELETE FROM deployments WHERE id = $1 AND app_id = $2 RETURNING *;
