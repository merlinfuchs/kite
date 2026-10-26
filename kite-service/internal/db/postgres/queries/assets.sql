-- name: CreateAsset :one
INSERT INTO assets (
    id,
    name,
    content_hash,
    content_type,
    content_size,
    app_id,
    bundle_install_id,
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

-- name: GetAssetContentHashes :many
SELECT DISTINCT content_hash FROM assets;

-- name: GetAssetStorageByApp :one
-- Files are stored once per content, so uploading the same one again doesn't
-- count twice.
SELECT COALESCE(SUM(content_size), 0)::BIGINT FROM (
    SELECT DISTINCT ON (content_hash) content_size FROM assets WHERE app_id = $1
) files;
