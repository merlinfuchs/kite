-- name: GetStats :one
SELECT
    (SELECT count(*) FROM users) AS users,
    (SELECT count(*) FROM apps) AS apps,
    (SELECT count(*) FROM commands) AS commands,
    (SELECT count(*) FROM event_listeners) AS event_listeners,
    (SELECT count(*) FROM usage_records WHERE created_at > sqlc.arg(executions_since)::timestamp) AS executions;

-- name: GetOfficialBundleInstallCounts :many
SELECT bundle_installs.bundle_id, count(*) AS installs FROM bundle_installs
JOIN bundles ON bundles.id = bundle_installs.bundle_id
WHERE bundles.owner_user_id IS NULL
GROUP BY bundle_installs.bundle_id;
