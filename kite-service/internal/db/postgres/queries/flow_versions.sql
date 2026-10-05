-- name: CreateFlowVersion :exec
INSERT INTO flow_versions (
    id,
    app_id,
    command_id,
    event_listener_id,
    flow_source,
    auto_saved,
    creator_user_id,
    created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetFlowVersion :one
SELECT * FROM flow_versions WHERE id = @id AND app_id = @app_id;

-- name: GetFlowVersionsByCommand :many
-- Leaves out the flow so the list stays small, it's loaded on restore.
SELECT
    flow_versions.id,
    flow_versions.app_id,
    flow_versions.command_id,
    flow_versions.event_listener_id,
    flow_versions.auto_saved,
    flow_versions.creator_user_id,
    flow_versions.created_at,
    users.display_name AS creator_display_name
FROM flow_versions
LEFT JOIN users ON users.id = flow_versions.creator_user_id
WHERE flow_versions.app_id = @app_id AND flow_versions.command_id = @command_id
ORDER BY flow_versions.created_at DESC
LIMIT @max_count;

-- name: GetFlowVersionsByEventListener :many
SELECT
    flow_versions.id,
    flow_versions.app_id,
    flow_versions.command_id,
    flow_versions.event_listener_id,
    flow_versions.auto_saved,
    flow_versions.creator_user_id,
    flow_versions.created_at,
    users.display_name AS creator_display_name
FROM flow_versions
LEFT JOIN users ON users.id = flow_versions.creator_user_id
WHERE flow_versions.app_id = @app_id AND flow_versions.event_listener_id = @event_listener_id
ORDER BY flow_versions.created_at DESC
LIMIT @max_count;

-- name: CountFlowVersionsByCommand :one
SELECT COUNT(*) FROM flow_versions WHERE command_id = $1;

-- name: CountFlowVersionsByEventListener :one
SELECT COUNT(*) FROM flow_versions WHERE event_listener_id = $1;

-- name: DeleteOldFlowVersionsByCommand :exec
DELETE FROM flow_versions old WHERE old.command_id = @command_id AND old.id NOT IN (
    SELECT kept.id FROM flow_versions kept
    WHERE kept.command_id = @command_id
    ORDER BY kept.created_at DESC
    LIMIT @keep_count
);

-- name: DeleteOldFlowVersionsByEventListener :exec
DELETE FROM flow_versions old WHERE old.event_listener_id = @event_listener_id AND old.id NOT IN (
    SELECT kept.id FROM flow_versions kept
    WHERE kept.event_listener_id = @event_listener_id
    ORDER BY kept.created_at DESC
    LIMIT @keep_count
);
