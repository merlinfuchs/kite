-- name: UpsertDeployment :one
INSERT INTO deployments (
    id,
    key, 
    name, 
    description, 
    guild_id, 
    plugin_version_id, 
    wasm_bytes, 
    manifest_default_config, 
    manifest_events, 
    manifest_commands, 
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
    $11,
    $12,
    $13
) ON CONFLICT (key, guild_id) DO UPDATE SET 
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    plugin_version_id = EXCLUDED.plugin_version_id,
    wasm_bytes = EXCLUDED.wasm_bytes,
    manifest_default_config = EXCLUDED.manifest_default_config,
    manifest_events = EXCLUDED.manifest_events,
    manifest_commands = EXCLUDED.manifest_commands,
    config = EXCLUDED.config,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetGuildIdsWithDeployments :many
SELECT DISTINCT guild_id FROM deployments;

-- name: GetDeploymentsForGuild :many
SELECT * FROM deployments WHERE guild_id = $1;