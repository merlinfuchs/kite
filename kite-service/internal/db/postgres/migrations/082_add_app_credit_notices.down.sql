UPDATE apps SET
    enabled = FALSE,
    disabled_reason = 'No credits remaining'
WHERE enabled AND credits_paused_at IS NOT NULL;

ALTER TABLE apps
    DROP COLUMN IF EXISTS credits_paused_at,
    DROP COLUMN IF EXISTS credit_notices_enabled,
    DROP COLUMN IF EXISTS credits_warned_at;
