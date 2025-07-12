-- name: GetPluginInstance :one
SELECT * FROM plugin_instances WHERE id = $1;

-- name: GetPluginInstancesByApp :many
SELECT * FROM plugin_instances WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountPluginInstancesByApp :one
SELECT COUNT(*) FROM plugin_instances WHERE app_id = $1;

-- name: CreatePluginInstance :one
INSERT INTO plugin_instances (
    id,
    plugin_id,
    enabled,
    app_id,
    creator_user_id,
    config,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdatePluginInstance :one
UPDATE plugin_instances SET
    enabled = $2,
    config = $3,
    updated_at = $4
WHERE id = $1 RETURNING *;

-- name: UpdatePluginInstancesLastDeployedAt :exec
UPDATE plugin_instances SET
    last_deployed_at = $2
WHERE app_id = $1;

-- name: GetEnabledPluginInstancesUpdatesSince :many
SELECT * FROM plugin_instances WHERE enabled = TRUE AND updated_at > $1;

-- name: GetEnabledPluginInstanceIDs :many
SELECT id FROM plugin_instances WHERE enabled = TRUE;

-- name: DeletePluginInstance :exec
DELETE FROM plugin_instances WHERE id = $1;