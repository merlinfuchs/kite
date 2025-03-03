-- name: GetPartialEnabledPluginInstanceIDs :many
SELECT app_id, plugin_id FROM plugin_instances WHERE enabled = TRUE;

-- name: GetPluginInstancesUpdatedSince :many
SELECT * FROM plugin_instances WHERE updated_at > $1;

-- name: UpsertPluginInstance :one
INSERT INTO plugin_instances (
    app_id, 
    plugin_id, 
    enabled, 
    config, 
    created_at, 
    updated_at
) VALUES (
    $1, 
    $2, 
    $3, 
    $4, 
    $5, 
    $6
) ON CONFLICT (app_id, plugin_id) 
DO UPDATE SET 
    enabled = $3,
    config = $4, 
    updated_at = $6 
RETURNING *;

-- name: GetPluginInstance :one
SELECT * FROM plugin_instances WHERE app_id = $1 AND plugin_id = $2;

-- name: DeletePluginInstance :exec
DELETE FROM plugin_instances WHERE app_id = $1 AND plugin_id = $2;