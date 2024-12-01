-- name: CreateAPIKey :one
INSERT INTO api_keys (
    id, 
    type, 
    name, 
    key, 
    key_hash, 
    app_id, 
    creator_user_id, 
    created_at, 
    updated_at, 
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1;

-- name: GetAPIKeysByAppID :many
SELECT * FROM api_keys WHERE app_id = $1;
