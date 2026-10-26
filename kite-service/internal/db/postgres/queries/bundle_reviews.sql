-- Submits the bundle to the marketplace, replacing a review that's still
-- pending.
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
    version = EXCLUDED.version,
    listing = EXCLUDED.listing,
    created_at = EXCLUDED.created_at;

-- A new version replaces the version of the pending review, which keeps its
-- place in the queue.
-- name: SetPendingBundleReviewVersion :exec
UPDATE bundle_reviews SET version = @version WHERE bundle_id = @bundle_id AND status = 'pending';

-- A new version of a listed bundle is reviewed with the listing it has.
-- name: CreateListedBundleReview :exec
INSERT INTO bundle_reviews (
    id,
    bundle_id,
    version,
    listing,
    status,
    created_at
)
SELECT
    @id,
    bundle_listings.bundle_id,
    @version,
    jsonb_build_object('icon', icon, 'category', category, 'about', about, 'faq', faq, 'setup', setup),
    'pending',
    @created_at
FROM bundle_listings
WHERE bundle_listings.bundle_id = @bundle_id AND bundle_listings.listed_version IS NOT NULL
ON CONFLICT (bundle_id) WHERE status = 'pending' DO NOTHING;

-- name: GetPendingBundleReviews :many
SELECT
    sqlc.embed(bundle_reviews),
    sqlc.embed(bundles),
    bundle_versions.content,
    bundle_listings.listed_version AS listed_version
FROM bundle_reviews
JOIN bundles ON bundles.id = bundle_reviews.bundle_id
JOIN bundle_versions ON bundle_versions.bundle_id = bundle_reviews.bundle_id AND bundle_versions.version = bundle_reviews.version
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundle_reviews.bundle_id
WHERE bundle_reviews.status = 'pending'
ORDER BY bundle_reviews.created_at, bundle_reviews.id;

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
