-- Bundles can always be seen and added through their link, which can't be
-- guessed. They don't need to be published for that anymore.
ALTER TABLE bundles DROP COLUMN IF EXISTS published_at;
