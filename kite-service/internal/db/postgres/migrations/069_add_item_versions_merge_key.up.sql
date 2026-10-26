-- Changes with the same merge key, like the saves of one editor session or of
-- one API token, update the newest version for a while instead of adding one.
ALTER TABLE item_versions ADD COLUMN IF NOT EXISTS merge_key TEXT;
