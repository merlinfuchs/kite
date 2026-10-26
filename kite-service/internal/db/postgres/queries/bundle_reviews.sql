-- Submits the bundle to the marketplace, replacing a review that's still
-- pending. It gets a new ID, so a decision on what an admin saw before fails.
-- name: UpsertPendingBundleReview :exec
INSERT INTO bundle_reviews (
    id,
    bundle_id,
    version,
    listing,
    status,
    created_at
) VALUES (
    $1, $2, $3, $4, 'pending', $5
)
ON CONFLICT (bundle_id) WHERE status = 'pending' DO UPDATE SET
    id = EXCLUDED.id,
    version = EXCLUDED.version,
    listing = EXCLUDED.listing,
    created_at = EXCLUDED.created_at;

-- A new version replaces the version of the pending review, which keeps its
-- place in the queue. It gets a new ID, like when it's submitted again.
-- name: SetPendingBundleReviewVersion :execrows
UPDATE bundle_reviews SET id = @id, version = @version WHERE bundle_id = @bundle_id AND status = 'pending';

-- Reviews lock their bundle first, so a review can't be submitted for an old
-- version or approved while the bundle is taken down.
-- name: LockBundle :one
SELECT * FROM bundles WHERE id = $1 FOR UPDATE;

-- name: GetPendingBundleReviewBundleID :one
SELECT bundle_id FROM bundle_reviews WHERE id = $1 AND status = 'pending';

-- name: GetBundleListing :one
SELECT * FROM bundle_listings WHERE bundle_id = $1;

-- Without the content, which only the review of one is loaded with.
-- name: GetPendingBundleReviews :many
SELECT
    sqlc.embed(bundle_reviews),
    sqlc.embed(bundles),
    (bundle_listings.bundle_id IS NOT NULL)::bool AS listed,
    sqlc.embed(users)
FROM bundle_reviews
JOIN bundles ON bundles.id = bundle_reviews.bundle_id
JOIN users ON users.id = bundles.owner_user_id
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundle_reviews.bundle_id
WHERE bundle_reviews.status = 'pending'
ORDER BY bundle_reviews.created_at, bundle_reviews.id;

-- name: GetPendingBundleReview :one
SELECT
    sqlc.embed(bundle_reviews),
    sqlc.embed(bundles),
    bundle_versions.content,
    bundle_versions.notes,
    (bundle_listings.bundle_id IS NOT NULL)::bool AS listed,
    sqlc.embed(users)
FROM bundle_reviews
JOIN bundles ON bundles.id = bundle_reviews.bundle_id
JOIN users ON users.id = bundles.owner_user_id
JOIN bundle_versions ON bundle_versions.bundle_id = bundle_reviews.bundle_id AND bundle_versions.version = bundle_reviews.version
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundle_reviews.bundle_id
WHERE bundle_reviews.id = $1 AND bundle_reviews.status = 'pending';

-- name: GetPendingBundleReviewForUpdate :one
SELECT * FROM bundle_reviews WHERE id = $1 AND status = 'pending' FOR UPDATE;

-- name: DecideBundleReview :exec
UPDATE bundle_reviews SET
    status = @status,
    reason = @reason,
    reviewer_user_id = @reviewer_user_id,
    decided_at = @decided_at
WHERE id = @id;

-- The pending review of each of the user's bundles, or the last decided one.
-- name: GetLatestBundleReviewsByOwner :many
SELECT DISTINCT ON (bundle_reviews.bundle_id) bundle_reviews.*
FROM bundle_reviews
JOIN bundles ON bundles.id = bundle_reviews.bundle_id
WHERE bundles.owner_user_id = $1
ORDER BY
    bundle_reviews.bundle_id,
    bundle_reviews.status = 'pending' DESC,
    bundle_reviews.decided_at DESC NULLS LAST,
    bundle_reviews.id;

-- Takes a bundle out of the marketplace, when its link stops working.
-- name: DeleteBundleListing :exec
DELETE FROM bundle_listings WHERE bundle_id = $1;

-- name: DeletePendingBundleReviews :exec
DELETE FROM bundle_reviews WHERE bundle_id = $1 AND status = 'pending';
