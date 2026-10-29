-- credit_notices_enabled is whether the owner gets a DM from their bot when
-- the app has used 80% or all of its credits. credits_warned_at is when the
-- 80% DM was last sent, so it's sent once a month.
ALTER TABLE apps
    ADD COLUMN credit_notices_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN credits_warned_at TIMESTAMP;
