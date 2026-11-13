-- Without their columns, the tokens of OAuth clients would last until their
-- refresh token expires.
DELETE FROM api_tokens WHERE oauth_client_id IS NOT NULL;

ALTER TABLE api_tokens DROP COLUMN IF EXISTS access_expires_at;
ALTER TABLE api_tokens DROP COLUMN IF EXISTS refresh_key_hash;
ALTER TABLE api_tokens DROP COLUMN IF EXISTS oauth_resource;
ALTER TABLE api_tokens DROP COLUMN IF EXISTS oauth_client_id;

DROP TABLE IF EXISTS oauth_codes;
