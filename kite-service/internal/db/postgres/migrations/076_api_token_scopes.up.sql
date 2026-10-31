-- Scopes are "<resource>:<access>" now. The old "all" scope is kite:write.
UPDATE api_tokens SET scopes = array_replace(scopes, 'all', 'kite:write');

-- The apps a token can use, null for all apps of its user.
ALTER TABLE api_tokens ADD COLUMN IF NOT EXISTS app_ids TEXT[];
