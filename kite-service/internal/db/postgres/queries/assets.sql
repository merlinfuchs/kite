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

-- name: GetUsedAssetIDs :many
-- The IDs of assets that message templates, flows and their saved versions use,
-- by asset_id like attachments and Play audio blocks, or by their URL like
-- images in embeds. Only apps with assets can use them by ID, but their URLs
-- also reach other apps through share codes and templates, which are read in
-- full.
WITH asset_apps AS (
    SELECT DISTINCT app_id FROM assets
), contents AS (
    SELECT flow_source::TEXT AS content FROM commands
    WHERE app_id IN (SELECT app_id FROM asset_apps)
    UNION ALL
    SELECT flow_source::TEXT FROM event_listeners
    WHERE app_id IN (SELECT app_id FROM asset_apps)
    UNION ALL
    SELECT data::TEXT || flow_sources::TEXT FROM messages
    WHERE app_id IN (SELECT app_id FROM asset_apps)
    UNION ALL
    -- Buttons of messages sent before an edit still run their old flows.
    SELECT message_instances.flow_sources::TEXT FROM message_instances
    JOIN messages ON messages.id = message_instances.message_id
    WHERE messages.app_id IN (SELECT app_id FROM asset_apps)
    UNION ALL
    -- Saved versions can be restored, including of deleted items.
    SELECT content::TEXT FROM item_versions
    WHERE app_id IN (SELECT app_id FROM asset_apps)
    UNION ALL
    SELECT data::TEXT FROM share_codes
    UNION ALL
    SELECT content::TEXT FROM bundle_versions
)
SELECT DISTINCT (regexp_matches(
    content,
    '(?:asset_id": ?"|/v1/assets/)([0-9a-z]+)',
    'g'
))[1]::TEXT AS asset_id FROM contents;

-- name: MarkAssetsUsed :exec
UPDATE assets SET unused_since = NULL
WHERE unused_since IS NOT NULL AND id = ANY(@used_ids::TEXT[]);

-- name: MarkAssetsUnused :exec
UPDATE assets SET unused_since = @unused_since
WHERE unused_since IS NULL AND NOT (id = ANY(@used_ids::TEXT[]));

-- name: GetAssetsUnusedBefore :many
SELECT * FROM assets
WHERE unused_since < @unused_before
ORDER BY unused_since
LIMIT @max_count;

-- name: IsAssetUsedByApp :one
-- Checks the rows of the asset's own app again right before it's deleted, in
-- case it was added back since GetUsedAssetIDs read them.
SELECT (
    EXISTS (
        SELECT 1 FROM commands
        WHERE commands.app_id = @app_id AND strpos(flow_source::TEXT, @asset_id::TEXT) > 0
    ) OR EXISTS (
        SELECT 1 FROM event_listeners
        WHERE event_listeners.app_id = @app_id AND strpos(flow_source::TEXT, @asset_id::TEXT) > 0
    ) OR EXISTS (
        SELECT 1 FROM messages
        WHERE messages.app_id = @app_id AND strpos(data::TEXT || flow_sources::TEXT, @asset_id::TEXT) > 0
    ) OR EXISTS (
        SELECT 1 FROM item_versions
        WHERE item_versions.app_id = @app_id AND strpos(content::TEXT, @asset_id::TEXT) > 0
    )
)::BOOLEAN;
