-- credits_paused_at is when the app was paused for using all of its credits
-- for the month. Paused apps stay enabled and connected, so their bot can tell
-- users why it doesn't respond, but run no flows.
--
-- credit_notices_enabled is whether the owner gets a DM from their bot when
-- the app has used 80% or all of its credits.
ALTER TABLE apps
    ADD COLUMN credits_paused_at TIMESTAMP,
    ADD COLUMN credit_notices_enabled BOOLEAN NOT NULL DEFAULT TRUE;

-- Apps stopped for credits were disabled with this reason before.
UPDATE apps SET
    enabled = TRUE,
    disabled_reason = NULL,
    credits_paused_at = updated_at
WHERE NOT enabled AND disabled_reason = 'No credits remaining';

-- Notifications are what Kite told the owner of an app or a user, for the
-- dashboard. Ones with a dedupe_key are sent once per key, like once a month.
CREATE TABLE IF NOT EXISTS notifications (
    id TEXT PRIMARY KEY,
    app_id TEXT REFERENCES apps(id) ON DELETE CASCADE,
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    dedupe_key TEXT,
    data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL,
    CHECK (app_id IS NOT NULL OR user_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS notifications_app_id_created_at ON notifications (app_id, created_at);
CREATE INDEX IF NOT EXISTS notifications_user_id_created_at ON notifications (user_id, created_at);
CREATE INDEX IF NOT EXISTS notifications_created_at ON notifications (created_at);
CREATE UNIQUE INDEX IF NOT EXISTS notifications_app_dedupe ON notifications (app_id, type, dedupe_key)
    WHERE app_id IS NOT NULL AND dedupe_key IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS notifications_user_dedupe ON notifications (user_id, type, dedupe_key)
    WHERE app_id IS NULL AND dedupe_key IS NOT NULL;
