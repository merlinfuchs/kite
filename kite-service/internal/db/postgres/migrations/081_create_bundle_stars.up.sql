-- Users star templates of the marketplace to find them again, and the
-- marketplace shows the ones with the most stars.
CREATE TABLE IF NOT EXISTS bundle_stars (
    bundle_id TEXT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (bundle_id, user_id)
);

CREATE INDEX IF NOT EXISTS bundle_stars_user_id ON bundle_stars (user_id, created_at DESC);
