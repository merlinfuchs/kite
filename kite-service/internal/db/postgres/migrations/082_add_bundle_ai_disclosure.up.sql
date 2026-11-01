ALTER TABLE bundles
    ADD COLUMN ai_use TEXT NOT NULL DEFAULT 'unknown'
        CHECK (ai_use IN ('unknown', 'none', 'assisted', 'generated'));

ALTER TABLE bundle_listings
    ADD COLUMN ai_use TEXT NOT NULL DEFAULT 'unknown'
        CHECK (ai_use IN ('unknown', 'none', 'assisted', 'generated'));

UPDATE bundles SET ai_use = 'none' WHERE owner_user_id IS NULL;
UPDATE bundle_listings SET ai_use = 'none'
FROM bundles
WHERE bundle_listings.bundle_id = bundles.id AND bundles.owner_user_id IS NULL;

UPDATE bundle_reviews
SET listing = jsonb_set(listing, '{ai_use}', '"unknown"'::jsonb, true)
WHERE COALESCE(listing->>'ai_use', '') = '';
