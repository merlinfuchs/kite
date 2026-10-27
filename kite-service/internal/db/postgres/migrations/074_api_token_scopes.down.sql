ALTER TABLE api_tokens DROP COLUMN IF EXISTS app_ids;

UPDATE api_tokens SET scopes = array_replace(scopes, 'kite:write', 'all');
