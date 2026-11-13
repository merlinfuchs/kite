-- name: CreateAppAuditLog :exec
INSERT INTO app_audit_logs (app_id, user_id, action, target_type, target_id, target_name, api_token_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: GetAppAuditLogs :many
SELECT app_audit_logs.*, users.display_name AS user_display_name, users.discord_username AS user_discord_username, users.discord_id AS user_discord_id, users.discord_avatar AS user_discord_avatar
FROM app_audit_logs
LEFT JOIN users ON users.id = app_audit_logs.user_id
WHERE app_audit_logs.app_id = @app_id
    AND (@before_id::bigint = 0 OR app_audit_logs.id < @before_id::bigint)
    AND (@user_id::text = '' OR app_audit_logs.user_id = @user_id::text)
    -- The kinds are the part of the action before the dot, like "command".
    AND (cardinality(@kinds::text[]) = 0 OR split_part(app_audit_logs.action, '.', 1) = ANY(@kinds::text[]))
ORDER BY app_audit_logs.id DESC
LIMIT @limit_count;

-- name: DeleteAppAuditLogsBefore :exec
DELETE FROM app_audit_logs WHERE created_at < @before;
