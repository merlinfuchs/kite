-- The app's public key, which Discord signs interactions and webhook events
-- with. Fetched with the install counts, NULL until then.
ALTER TABLE apps ADD COLUMN discord_verify_key TEXT;
