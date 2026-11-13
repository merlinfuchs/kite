DROP INDEX IF EXISTS apps_over_limit_since;
ALTER TABLE apps DROP COLUMN IF EXISTS over_limit_since;
