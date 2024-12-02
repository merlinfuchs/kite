-- name: GetEventListener :one
SELECT * FROM event_listeners WHERE id = $1;

-- name: GetEventListenersByApp :many
SELECT * FROM event_listeners WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CountEventListenersByApp :one
SELECT COUNT(*) FROM event_listeners WHERE app_id = $1;

-- name: CreateEventListener :one
INSERT INTO event_listeners (
    id,
    enabled,
    app_id,
    module_id,
    creator_user_id,
    source,
    type,
    filter,
    flow_source,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: UpdateEventListener :one
UPDATE event_listeners SET
    enabled = $2,
    type = $3,
    filter = $4,
    flow_source = $5,
    updated_at = $6
WHERE id = $1 RETURNING *;

-- name: GetEnabledEventListenersUpdatesSince :many
SELECT * FROM event_listeners WHERE enabled = TRUE AND updated_at > $1;

-- name: GetEnabledEventListenerIDs :many
SELECT id FROM event_listeners WHERE enabled = TRUE;

-- name: DeleteEventListener :exec
DELETE FROM event_listeners WHERE id = $1;