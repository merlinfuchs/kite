-- name: CreateFlowVersion :exec
INSERT INTO flow_versions (
    id,
    app_id,
    target_type,
    target_id,
    source,
    creator_user_id,
    flow_source,
    target_data,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: UpdateFlowVersion :exec
UPDATE flow_versions SET
    flow_source = @flow_source,
    created_at = @created_at
WHERE id = @id;

-- name: GetFlowVersion :one
SELECT * FROM flow_versions WHERE id = @id AND app_id = @app_id;

-- name: GetNewestFlowVersion :one
SELECT id, source, creator_user_id, created_at FROM flow_versions
WHERE target_id = @target_id
ORDER BY created_at DESC
LIMIT 1;

-- name: GetFlowVersionsByTarget :many
-- Leaves out the flow so the list stays small, it's loaded on restore.
SELECT
    flow_versions.id,
    flow_versions.app_id,
    flow_versions.target_type,
    flow_versions.target_id,
    flow_versions.source,
    flow_versions.creator_user_id,
    flow_versions.created_at,
    users.display_name AS creator_display_name
FROM flow_versions
LEFT JOIN users ON users.id = flow_versions.creator_user_id
WHERE flow_versions.app_id = @app_id AND flow_versions.target_id = @target_id
ORDER BY flow_versions.created_at DESC
LIMIT @max_count;

-- name: DeleteOldFlowVersions :exec
DELETE FROM flow_versions old WHERE old.target_id = @target_id AND old.id NOT IN (
    SELECT kept.id FROM flow_versions kept
    WHERE kept.target_id = @target_id
    ORDER BY kept.created_at DESC
    LIMIT @keep_count
);

-- name: DeleteDeletedFlowVersions :execrows
-- Deletes the history of commands and event listeners deleted before
-- deleted_before. A newer version means it was restored, so it's kept.
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM flow_versions WHERE id IN (
    SELECT stale.id FROM flow_versions stale
    JOIN flow_versions deleted ON deleted.target_id = stale.target_id
    WHERE deleted.source = 'delete'
        AND deleted.created_at < @deleted_before
        AND NOT EXISTS (
            SELECT 1 FROM flow_versions newer
            WHERE newer.target_id = deleted.target_id
                AND newer.created_at > deleted.created_at
        )
    LIMIT @batch_size
);
