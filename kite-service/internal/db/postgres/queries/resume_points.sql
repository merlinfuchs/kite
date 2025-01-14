-- name: CreateResumePoint :one
INSERT INTO resume_points (
    id, 
    type,
    app_id, 
    command_id, 
    event_listener_id, 
    message_id, 
    message_instance_id,
    flow_source_id, 
    flow_node_id, 
    flow_state, 
    created_at, 
    expires_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: DeleteResumePoint :exec
DELETE FROM resume_points WHERE id = $1;

-- name: DeleteExpiredResumePoints :exec
DELETE FROM resume_points WHERE expires_at < $1;

-- name: ResumePoint :one
SELECT * FROM resume_points WHERE id = $1;
