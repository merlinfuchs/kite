-- When the cleanup first found that no message template or flow uses the
-- asset, NULL while one does.
ALTER TABLE assets ADD COLUMN IF NOT EXISTS unused_since TIMESTAMP;

CREATE INDEX IF NOT EXISTS assets_unused_since ON assets (unused_since) WHERE unused_since IS NOT NULL;
