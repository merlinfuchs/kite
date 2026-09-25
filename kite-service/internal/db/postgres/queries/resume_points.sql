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
    expires_at,
    resume_at,
    interaction_token
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14);

-- name: DeleteResumePoint :exec
DELETE FROM resume_points WHERE id = $1;

-- name: ResumePoint :one
-- Scoped by app since the ID comes from a user-controlled custom_id
SELECT * FROM resume_points WHERE id = $1 AND app_id = $2;

-- name: DeleteStaleResumePoints :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM resume_points WHERE id IN (
    SELECT stale.id FROM resume_points stale
    -- Pending timers are never used before they resume, their expiry covers them.
    WHERE stale.expires_at < @now OR (stale.last_used_at < @used_before AND stale.resume_at IS NULL)
    LIMIT @batch_size
);

-- name: TouchResumePoint :exec
UPDATE resume_points SET last_used_at = @used_at WHERE id = @id AND app_id = @app_id;


-- name: CountPendingTimerResumePoints :one
SELECT COUNT(*) FROM resume_points WHERE app_id = $1 AND resume_at IS NOT NULL;

-- name: HasDueTimerResumePoints :one
SELECT EXISTS(SELECT 1 FROM resume_points WHERE resume_at <= @now);

-- name: LeaseDueTimerResumePoints :many
-- Moves resume_at to the end of the lease instead of deleting, so a timer that
-- fails to resume is retried once the lease is over.
UPDATE resume_points SET resume_at = @lease_until WHERE id IN (
    SELECT due.id FROM resume_points due
    WHERE due.resume_at <= @now AND due.app_id = ANY(@app_ids::TEXT[])
    ORDER BY due.resume_at
    LIMIT @batch_size
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: DeleteTimerResumePoint :execrows
-- Deleting is the commit point: a timer only resumes if this deleted it.
DELETE FROM resume_points WHERE id = $1 AND app_id = $2 AND resume_at IS NOT NULL;
