-- name: GetEventListener :one
SELECT * FROM event_listeners WHERE id = $1;

-- name: GetEventListenersByApp :many
SELECT * FROM event_listeners WHERE app_id = $1 ORDER BY created_at DESC;

-- name: CreateEventListener :one
INSERT INTO event_listeners (
    id,
    source,
    type,
    description,
    enabled,
    app_id,
    module_id,
    creator_user_id,
    filter,
    flow_source,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: UpdateEventListener :one
UPDATE event_listeners SET
    enabled = $2,
    type = $3,
    filter = $4,
    description = $5,
    flow_source = $6,
    updated_at = $7
WHERE id = $1 RETURNING *;

-- name: GetEventListenersUpdatedSince :many
-- Includes disabled listeners so the engine can drop them right away. The
-- first load has nothing to drop, so it skips them.
SELECT * FROM event_listeners WHERE updated_at > @updated_since AND (enabled = TRUE OR @include_disabled::BOOLEAN);

-- name: GetEnabledScheduledEventListenerIDs :many
SELECT id FROM event_listeners WHERE enabled = TRUE AND source = 'schedule';

-- name: GetEnabledEventListenerIDs :many
SELECT id FROM event_listeners WHERE enabled = TRUE;

-- name: DeleteEventListener :exec
DELETE FROM event_listeners WHERE id = $1;

-- name: CountEventListenersByAppAndSource :one
SELECT COUNT(*) FROM event_listeners WHERE app_id = $1 AND source = $2;

-- name: UpdateEventListenersLastRunAt :exec
-- Doesn't touch updated_at, otherwise every run would make the engine reload the listener.
UPDATE event_listeners SET last_run_at = runs.last_run_at
FROM (
    SELECT UNNEST(@ids::TEXT[]) AS id, UNNEST(@last_run_ats::TIMESTAMP[]) AS last_run_at
) AS runs
WHERE event_listeners.id = runs.id;
