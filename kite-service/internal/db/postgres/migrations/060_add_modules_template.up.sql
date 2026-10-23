-- A module is a template installed into an app. The commands and event
-- listeners it created point at it with module_id, and its resources list them
-- with the hash of their flow at install, so updates can tell what was edited.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS template_id TEXT NOT NULL DEFAULT '';
ALTER TABLE modules ADD COLUMN IF NOT EXISTS template_version INT NOT NULL DEFAULT 0;
-- The values the user filled in for the template's settings.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}';

CREATE INDEX IF NOT EXISTS modules_app_id ON modules (app_id);
