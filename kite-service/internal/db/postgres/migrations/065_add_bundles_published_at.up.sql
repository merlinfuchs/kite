-- When a user published their bundle, so anyone with its link can see and
-- install it. Official bundles are always public. Publishing doesn't list a
-- bundle anywhere, that's what bundle_listings will be for.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS published_at TIMESTAMP;
