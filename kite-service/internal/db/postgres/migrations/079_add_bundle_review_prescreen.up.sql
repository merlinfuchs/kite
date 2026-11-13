-- What the AI pre-screen made of the version in review, for the admins. A
-- review it rejected has no reviewer and the verdict reject.
ALTER TABLE bundle_reviews
    -- review, reject, failed or skipped, empty while it wasn't screened yet.
    ADD COLUMN IF NOT EXISTS ai_verdict TEXT NOT NULL DEFAULT '',
    -- How likely the admins approve it, from 0 to 100.
    ADD COLUMN IF NOT EXISTS ai_rating INT,
    ADD COLUMN IF NOT EXISTS ai_summary TEXT NOT NULL DEFAULT '',
    -- What the admins should look at, as [{"severity": ..., "text": ...}].
    ADD COLUMN IF NOT EXISTS ai_concerns JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS ai_model TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS ai_screened_at TIMESTAMP;

-- The pre-screen picks up the pending reviews it didn't screen yet.
CREATE INDEX IF NOT EXISTS bundle_reviews_unscreened
    ON bundle_reviews (created_at) WHERE status = 'pending' AND ai_screened_at IS NULL;
