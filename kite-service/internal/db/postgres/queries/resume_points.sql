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

-- name: ResumePoint :one
-- Scoped by app since the ID comes from a user-controlled custom_id
SELECT * FROM resume_points WHERE id = $1 AND app_id = $2;

-- name: DeleteStaleResumePoints :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM resume_points WHERE id IN (
    SELECT stale.id FROM resume_points stale
    WHERE stale.expires_at < @now OR stale.last_used_at < @used_before
    LIMIT @batch_size
);

-- name: TouchResumePoint :exec
UPDATE resume_points SET last_used_at = @used_at WHERE id = @id AND app_id = @app_id;
