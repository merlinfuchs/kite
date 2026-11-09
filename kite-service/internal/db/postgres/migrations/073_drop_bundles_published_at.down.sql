-- All bundles that weren't taken down were reachable through their link.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS published_at TIMESTAMP;
UPDATE bundles SET published_at = created_at WHERE owner_user_id IS NOT NULL AND removed_at IS NULL;
