-- Existing rows count as used now, so nothing expires until a full expiry
-- period after this ships.
ALTER TABLE resume_points ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC');
ALTER TABLE message_instances ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP NOT NULL DEFAULT (now() AT TIME ZONE 'UTC');
