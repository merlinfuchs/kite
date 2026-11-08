-- What changed in the version, which apps that update to it see.
ALTER TABLE bundle_versions ADD COLUMN IF NOT EXISTS notes TEXT NOT NULL DEFAULT '';
