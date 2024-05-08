-- name: GetKVStorageNamespaces :many
SELECT  namespace, COUNT(key) as key_count FROM kv_storage WHERE app_id = $1 GROUP BY namespace;

-- name: GetKVStorageKeys :many
SELECT * FROM kv_storage WHERE app_id = $1 AND namespace = $2;

-- name: SetKVStorageKey :one
INSERT INTO kv_storage (
    app_id, 
    namespace, 
    key, 
    value, 
    created_at, 
    updated_at
) VALUES (
    $1, 
    $2, 
    $3, 
    $4, 
    $5, 
    $6
) ON CONFLICT (app_id, namespace, key) DO UPDATE SET 
    value = $4, 
    updated_at = $6
RETURNING *;

-- name: GetKVStorageKey :one
SELECT * FROM kv_storage WHERE app_id = $1 AND namespace = $2 AND key = $3;

-- name: DeleteKVStorageKey :one
DELETE FROM kv_storage WHERE app_id = $1 AND namespace = $2 AND key = $3 RETURNING *;
