-- name: UpsertDeployment :one
INSERT INTO deployments (
    id,
    key, 
    name, 
    description, 
    guild_id, 
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
) ON CONFLICT (key, guild_id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    plugin_version_id = EXCLUDED.plugin_version_id,
    wasm_bytes = EXCLUDED.wasm_bytes,
    manifest = EXCLUDED.manifest,
    config = EXCLUDED.config,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetGuildIdsWithDeployments :many
SELECT DISTINCT guild_id FROM deployments;

-- name: GetDeploymentsForGuild :many
SELECT * FROM deployments WHERE guild_id = $1 ORDER BY updated_at DESC;

-- name: GetDeploymentForGuild :one
SELECT * FROM deployments WHERE id = $1 AND guild_id = $2;

-- name: DeleteDeployment :one
DELETE FROM deployments WHERE id = $1 AND guild_id = $2 RETURNING *;
