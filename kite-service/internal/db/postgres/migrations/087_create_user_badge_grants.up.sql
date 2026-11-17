-- Badges Kite gives users by hand, like contributor. The others follow from
-- what users did, like listing templates or subscribing, and aren't stored.
CREATE TABLE IF NOT EXISTS user_badge_grants (
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge TEXT NOT NULL,
    granted_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (user_id, badge)
);
