-- The app's public key, which Discord signs interactions and webhook events
-- with. Stored when the gateway connects, NULL until then.
ALTER TABLE apps ADD COLUMN discord_verify_key TEXT;
