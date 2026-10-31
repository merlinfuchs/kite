-- Kite took the bundle down for breaking the rules. Only its owner can still
-- see it, with the reason, and it can't be shared again. Taking it down is
-- final, appeals are handled by hand.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS removed_at TIMESTAMP;
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS removed_reason TEXT NOT NULL DEFAULT '';

-- Reports of users' bundles that break the rules, for the admins.
CREATE TABLE IF NOT EXISTS bundle_reports (
    id TEXT PRIMARY KEY,
    bundle_id TEXT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    reporter_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- harmful, scam, inappropriate, spam or other.
    category TEXT NOT NULL,
    details TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    -- When it was dismissed or the bundle was taken down.
    resolved_at TIMESTAMP,
    resolver_user_id TEXT REFERENCES users(id) ON DELETE SET NULL
);

-- A user has at most one open report of a bundle.
CREATE UNIQUE INDEX IF NOT EXISTS bundle_reports_open ON bundle_reports (bundle_id, reporter_user_id) WHERE resolved_at IS NULL;
