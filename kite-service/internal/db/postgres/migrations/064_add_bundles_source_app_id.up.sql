-- The app a user's bundle is made from. Its versions are exported from it, so
-- a bundle whose app was deleted keeps working but can't get new versions.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS source_app_id TEXT REFERENCES apps(id) ON DELETE SET NULL;
