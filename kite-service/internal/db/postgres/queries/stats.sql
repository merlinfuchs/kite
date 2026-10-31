-- name: GetStats :one
SELECT
    (SELECT count(*) FROM users) AS users,
    (SELECT count(*) FROM apps) AS apps,
    (SELECT count(*) FROM commands) AS commands,
    (SELECT count(*) FROM event_listeners) AS event_listeners,
    (SELECT count(*) FROM usage_records WHERE created_at > sqlc.arg(executions_since)::timestamp) AS executions;
