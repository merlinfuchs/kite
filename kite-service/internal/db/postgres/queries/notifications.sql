-- Does nothing if a notification with the same dedupe_key was created before.
-- name: CreateNotification :execrows
INSERT INTO notifications (id, app_id, user_id, type, dedupe_key, data, created_at)
VALUES (@id, @app_id, @user_id, @type, @dedupe_key, @data, @created_at)
ON CONFLICT DO NOTHING;

-- name: DeleteNotificationsBefore :execrows
-- Batched so a large backlog doesn't hold one long transaction.
DELETE FROM notifications WHERE id IN (
    SELECT old.id FROM notifications old
    WHERE old.created_at < @created_before
    LIMIT @batch_size
);
