-- The version of a user's bundle that apps install once a review approved one,
-- also after it leaves the marketplace, so new versions always need a review.
-- Bundles that were never approved leave it NULL and install their newest
-- version.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS approved_version INT;

-- Submissions of users' bundles to the marketplace, and new versions of listed
-- ones. A bundle has at most one pending review, of its newest version.
CREATE TABLE IF NOT EXISTS bundle_reviews (
    id TEXT PRIMARY KEY,
    bundle_id TEXT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    version INT NOT NULL,
    -- The listing to show once it's approved, like bundle_listings.
    listing JSONB NOT NULL,
    -- pending, approved or rejected.
    status TEXT NOT NULL,
    -- Why it was rejected, shown to the user.
    reason TEXT NOT NULL DEFAULT '',
    reviewer_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,

    created_at TIMESTAMP NOT NULL,
    decided_at TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS bundle_reviews_pending ON bundle_reviews (bundle_id) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS bundle_reviews_bundle_id ON bundle_reviews (bundle_id, created_at);
