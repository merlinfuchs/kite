-- name: CreateAsset :one
INSERT INTO assets (
    id,
    name,
    content_hash,
    content_type,
    content_size,
    app_id,
    module_id,
    creator_user_id,
    created_at,
    updated_at,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetAsset :one
SELECT * FROM assets WHERE id = $1;

-- name: DeleteAsset :one
DELETE FROM assets WHERE id = $1 RETURNING *;

-- name: GetExpiredAssets :many
SELECT * FROM assets WHERE expires_at < $1;

-- name: CountAssetsByContentHash :one
SELECT COUNT(*) FROM assets WHERE content_hash = $1;