-- name: UpdateEditorSessionTime :exec
-- The editor reports its total active time. It can be at most the time
-- since the session started plus slack_seconds, and never goes down. A session ID sent for another app, user or item changes nothing.
INSERT INTO editor_sessions (
    id,
    app_id,
    user_id,
    command_id,
    event_listener_id,
    active_seconds,
    created_at,
    updated_at
) VALUES (
    @id,
    @app_id,
    @user_id,
    @command_id,
    @event_listener_id,
    LEAST(sqlc.arg(active_seconds)::integer, sqlc.arg(slack_seconds)::integer),
    @now,
    @now
)
ON CONFLICT (id) DO UPDATE SET
    active_seconds = GREATEST(editor_sessions.active_seconds, LEAST(
        sqlc.arg(active_seconds)::integer,
        EXTRACT(EPOCH FROM sqlc.arg(now)::timestamp - editor_sessions.created_at)::integer + sqlc.arg(slack_seconds)::integer
    )),
    updated_at = @now
WHERE editor_sessions.app_id = EXCLUDED.app_id
    AND editor_sessions.user_id = EXCLUDED.user_id
    AND editor_sessions.command_id IS NOT DISTINCT FROM EXCLUDED.command_id
    AND editor_sessions.event_listener_id IS NOT DISTINCT FROM EXCLUDED.event_listener_id;

-- name: RecordEditorSessionSave :exec
INSERT INTO editor_sessions (
    id,
    app_id,
    user_id,
    command_id,
    event_listener_id,
    save_count,
    created_at,
    updated_at,
    last_saved_at
) VALUES (
    @id,
    @app_id,
    @user_id,
    @command_id,
    @event_listener_id,
    1,
    @now,
    @now,
    @now
)
ON CONFLICT (id) DO UPDATE SET
    save_count = editor_sessions.save_count + 1,
    updated_at = @now,
    last_saved_at = @now
WHERE editor_sessions.app_id = EXCLUDED.app_id
    AND editor_sessions.user_id = EXCLUDED.user_id
    AND editor_sessions.command_id IS NOT DISTINCT FROM EXCLUDED.command_id
    AND editor_sessions.event_listener_id IS NOT DISTINCT FROM EXCLUDED.event_listener_id;

-- name: GetEditorUsageByApp :many
SELECT
    command_id,
    event_listener_id,
    SUM(active_seconds)::bigint AS active_seconds,
    SUM(save_count)::bigint AS save_count,
    MAX(last_saved_at)::timestamp AS last_saved_at
FROM editor_sessions
WHERE app_id = @app_id AND created_at >= @since
GROUP BY command_id, event_listener_id;

-- name: DeleteOldEditorSessions :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM editor_sessions WHERE id IN (
    SELECT old.id FROM editor_sessions old
    WHERE old.created_at < @created_before
    LIMIT @batch_size
);
