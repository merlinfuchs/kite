-- Users Kite banned from a feature, like marketplace, which keeps them from
-- submitting bundles and gets new versions of their bundles no review. What
-- they listed before stays listed until it's taken down.
CREATE TABLE IF NOT EXISTS user_bans (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature TEXT NOT NULL,
    -- Why they were banned, shown to them.
    reason TEXT NOT NULL,
    banned_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, feature)
);
