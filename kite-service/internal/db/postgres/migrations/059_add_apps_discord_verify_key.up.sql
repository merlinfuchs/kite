-- The app's public key, which Discord signs interactions and webhook events
-- with. Stored when the app is created, for older apps the first time a flow
-- needs it.
ALTER TABLE apps ADD COLUMN discord_verify_key TEXT;
