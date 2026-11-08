DROP TABLE IF EXISTS bundle_reports;
ALTER TABLE bundles DROP COLUMN IF EXISTS removed_reason;
ALTER TABLE bundles DROP COLUMN IF EXISTS removed_at;
