-- Code before scopes ignores them and app IDs, so only tokens that could do
-- everything keep working instead of getting more access than they had.
DELETE FROM api_tokens WHERE scopes <> ARRAY['kite:write'] OR app_ids IS NOT NULL;

ALTER TABLE api_tokens DROP COLUMN IF EXISTS app_ids;

UPDATE api_tokens SET scopes = ARRAY['all'];
