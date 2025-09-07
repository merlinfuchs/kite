-- name: SetPluginValue :one
INSERT INTO plugin_values (
    plugin_instance_id,
    key,
    value,
    metadata,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
) ON CONFLICT (plugin_instance_id, key) DO UPDATE SET
    value = EXCLUDED.value,
    metadata = EXCLUDED.metadata,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetPluginValue :one
SELECT * FROM plugin_values WHERE plugin_instance_id = $1 AND key = $2;

-- name: GetPluginValueForUpdate :one
SELECT * FROM plugin_values WHERE plugin_instance_id = $1 AND key = $2 FOR UPDATE;

-- name: DeletePluginValue :exec
DELETE FROM plugin_values WHERE plugin_instance_id = $1 AND key = $2;

-- name: SearchPluginValues :many
SELECT * FROM plugin_values WHERE plugin_instance_id = $1 AND metadata @> $2;
