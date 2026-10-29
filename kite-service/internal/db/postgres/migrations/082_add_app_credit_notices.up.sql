-- credits_paused_at is when the app was paused for using all of its credits
-- for the month. Paused apps stay enabled and connected, so their bot can tell
-- users why it doesn't respond, but run no flows.
--
-- credit_notices_enabled is whether the owner gets a DM from their bot when
-- the app has used 80% or all of its credits. credits_warned_at is when the
-- 80% DM was last sent, so it's sent once a month.
ALTER TABLE apps
    ADD COLUMN credits_paused_at TIMESTAMP,
    ADD COLUMN credit_notices_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN credits_warned_at TIMESTAMP;

-- Apps stopped for credits were disabled with this reason before.
UPDATE apps SET
    enabled = TRUE,
    disabled_reason = NULL,
    credits_paused_at = updated_at
WHERE NOT enabled AND disabled_reason = 'No credits remaining';
