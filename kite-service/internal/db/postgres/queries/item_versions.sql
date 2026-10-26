-- name: CreateItemVersion :exec
INSERT INTO item_versions (
    id,
    app_id,
    target_type,
    target_id,
    source,
    creator_user_id,
    content,
    target_data,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateItemVersion :exec
UPDATE item_versions SET
    content = @content,
    created_at = @created_at
WHERE id = @id;

-- name: GetItemVersion :one
SELECT * FROM item_versions WHERE id = @id AND app_id = @app_id;

-- name: GetNewestItemVersion :one
SELECT id, source, creator_user_id, created_at FROM item_versions
WHERE target_id = @target_id
ORDER BY created_at DESC
LIMIT 1;

-- name: GetItemVersionsByTarget :many
-- Leaves out the content so the list stays small, it's loaded on restore.
SELECT
    item_versions.id,
    item_versions.app_id,
    item_versions.target_type,
    item_versions.target_id,
    item_versions.source,
    item_versions.creator_user_id,
    item_versions.created_at,
    users.display_name AS creator_display_name
FROM item_versions
LEFT JOIN users ON users.id = item_versions.creator_user_id
WHERE item_versions.app_id = @app_id AND item_versions.target_id = @target_id
ORDER BY item_versions.created_at DESC
LIMIT @max_count;

-- name: DeleteOldItemVersions :exec
DELETE FROM item_versions old WHERE old.target_id = @target_id AND old.id NOT IN (
    SELECT kept.id FROM item_versions kept
    WHERE kept.target_id = @target_id
    ORDER BY kept.created_at DESC
    LIMIT @keep_count
);

-- name: DeleteDeletedItemVersions :execrows
-- Deletes the history of items deleted before deleted_before. A newer version
-- means the item was restored, so it's kept. Batched so a large backlog
-- doesn't hold one long transaction.
DELETE FROM item_versions WHERE id IN (
    SELECT stale.id FROM item_versions stale
    JOIN item_versions deleted ON deleted.target_id = stale.target_id
    WHERE deleted.source = 'delete'
        AND deleted.created_at < @deleted_before
        AND NOT EXISTS (
            SELECT 1 FROM item_versions newer
            WHERE newer.target_id = deleted.target_id
                AND newer.created_at > deleted.created_at
        )
    LIMIT @batch_size
);
