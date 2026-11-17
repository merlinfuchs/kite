-- When Kite picked a user's bundle, which the marketplace shows with a badge.
-- NULL if it isn't picked.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS staff_picked_at TIMESTAMP;
