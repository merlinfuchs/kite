-- name: GetStats :one
SELECT
    (SELECT count(*) FROM users) AS users,
    (SELECT count(*) FROM apps) AS apps,
    (SELECT count(*) FROM commands) AS commands,
    (SELECT count(*) FROM event_listeners) AS event_listeners,
    (SELECT count(*) FROM usage_records WHERE created_at > sqlc.arg(executions_since)::timestamp) AS executions;

-- Only bundles in the marketplace, the IDs of others are their share links.
-- name: GetListedBundleInstallCounts :many
SELECT bundle_installs.bundle_id, count(*) AS installs FROM bundle_installs
JOIN bundles ON bundles.id = bundle_installs.bundle_id
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
WHERE bundles.removed_at IS NULL
GROUP BY bundle_installs.bundle_id;
