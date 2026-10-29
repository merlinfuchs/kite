ALTER TABLE apps
    DROP COLUMN IF EXISTS credit_notices_enabled,
    DROP COLUMN IF EXISTS credits_warned_at;
