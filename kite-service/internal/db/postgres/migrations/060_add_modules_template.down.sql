DROP INDEX IF EXISTS modules_app_id;

ALTER TABLE modules DROP COLUMN IF EXISTS settings;
ALTER TABLE modules DROP COLUMN IF EXISTS template_version;
ALTER TABLE modules DROP COLUMN IF EXISTS template_id;
