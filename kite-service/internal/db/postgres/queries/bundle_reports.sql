-- A new report of the same user replaces their open one.
-- name: UpsertOpenBundleReport :exec
INSERT INTO bundle_reports (
    id,
    bundle_id,
    reporter_user_id,
    category,
    details,
    created_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (bundle_id, reporter_user_id) WHERE resolved_at IS NULL DO UPDATE SET
    category = EXCLUDED.category,
    details = EXCLUDED.details,
    created_at = EXCLUDED.created_at;

-- name: GetOpenBundleReports :many
SELECT sqlc.embed(bundle_reports), sqlc.embed(bundles)
FROM bundle_reports
JOIN bundles ON bundles.id = bundle_reports.bundle_id
WHERE bundle_reports.resolved_at IS NULL
ORDER BY bundle_reports.created_at, bundle_reports.id;

-- name: ResolveBundleReport :execrows
UPDATE bundle_reports SET resolved_at = @resolved_at, resolver_user_id = @resolver_user_id
WHERE id = @id AND resolved_at IS NULL;

-- name: ResolveBundleReportsOfBundle :exec
UPDATE bundle_reports SET resolved_at = @resolved_at, resolver_user_id = @resolver_user_id
WHERE bundle_id = @bundle_id AND resolved_at IS NULL;

-- name: RemoveBundle :execrows
UPDATE bundles SET removed_at = @removed_at, removed_reason = @removed_reason
WHERE id = @id AND owner_user_id IS NOT NULL AND removed_at IS NULL;
