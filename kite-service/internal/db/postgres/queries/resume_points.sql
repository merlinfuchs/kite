-- name: CreateResumePoint :exec
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
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: DeleteResumePoint :exec
DELETE FROM resume_points WHERE id = $1;

-- name: DeleteExpiredResumePoints :exec
DELETE FROM resume_points WHERE expires_at < $1;

-- name: ResumePoint :one
-- Scoped by app since the ID comes from a user-controlled custom_id
SELECT * FROM resume_points WHERE id = $1 AND app_id = $2;

-- name: DeleteUnusedResumePoints :exec
DELETE FROM resume_points WHERE last_used_at < @used_before;

-- name: TouchResumePoint :exec
-- Only writes once a day per resume point so busy buttons don't write on every click.
UPDATE resume_points SET last_used_at = @used_at
WHERE id = @id AND app_id = @app_id AND last_used_at < sqlc.arg(used_at)::timestamp - INTERVAL '1 day';
