-- name: CreateSuspendPoint :one
INSERT INTO suspend_points (
    id, 
    type,
    app_id, 
    command_id, 
    event_listener_id, 
    message_id, 
    flow_source_id, 
    flow_node_id, 
    flow_state, 
    created_at, 
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: DeleteSuspendPoint :exec
DELETE FROM suspend_points WHERE id = $1;

-- name: DeleteExpiredSuspendPoints :exec
DELETE FROM suspend_points WHERE expires_at < $1;

-- name: SuspendPoint :one
SELECT * FROM suspend_points WHERE id = $1;
