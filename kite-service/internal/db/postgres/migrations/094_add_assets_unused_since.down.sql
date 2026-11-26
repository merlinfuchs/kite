DROP INDEX IF EXISTS assets_unused_since;

ALTER TABLE assets DROP COLUMN IF EXISTS unused_since;
