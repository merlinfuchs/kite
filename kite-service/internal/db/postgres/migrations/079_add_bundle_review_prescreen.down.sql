DROP INDEX IF EXISTS bundle_reviews_unscreened;

ALTER TABLE bundle_reviews
    DROP COLUMN IF EXISTS ai_verdict,
    DROP COLUMN IF EXISTS ai_rating,
    DROP COLUMN IF EXISTS ai_summary,
    DROP COLUMN IF EXISTS ai_concerns,
    DROP COLUMN IF EXISTS ai_model,
    DROP COLUMN IF EXISTS ai_screened_at;
