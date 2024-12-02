-- name: GetEventListener :one
SELECT * FROM event_listeners WHERE id = $1;

-- name: GetEventListenersByApp :many
SELECT * FROM event_listeners WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountEventListenersByApp :one
SELECT COUNT(*) FROM event_listeners WHERE app_id = $1;

-- name: CreateEventListener :one
INSERT INTO event_listeners (
    id,
    name,
    description,
    enabled,
    app_id,
    module_id,
    creator_user_id,
    integration,
    type,
    filter,
    flow_source,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: UpdateEventListener :one
UPDATE event_listeners SET
    name = $2,
    description = $3,
    enabled = $4,
    type = $5,
    filter = $6,
    flow_source = $7,
    updated_at = $8
WHERE id = $1 RETURNING *;

-- name: GetEnabledEventListenersUpdatesSince :many
SELECT * FROM event_listeners WHERE enabled = TRUE AND updated_at > $1;

-- name: GetEnabledEventListenerIDs :many
SELECT id FROM event_listeners WHERE enabled = TRUE;

-- name: DeleteEventListener :exec
DELETE FROM event_listeners WHERE id = $1;