-- name: CreateBundleStar :exec
INSERT INTO bundle_stars (bundle_id, user_id, created_at)
VALUES ($1, $2, $3)
ON CONFLICT (bundle_id, user_id) DO NOTHING;

-- name: DeleteBundleStar :exec
DELETE FROM bundle_stars WHERE bundle_id = $1 AND user_id = $2;

-- name: GetStarredBundleIDs :many
SELECT bundle_id FROM bundle_stars WHERE user_id = $1 ORDER BY created_at DESC, bundle_id;

-- name: CountBundleStars :one
SELECT count(*)::int FROM bundle_stars WHERE bundle_id = $1;
