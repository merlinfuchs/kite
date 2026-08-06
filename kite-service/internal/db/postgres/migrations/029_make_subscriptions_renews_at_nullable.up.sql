-- LemonSqueezy leaves renews_at null for subscriptions that will not renew,
-- such as paused or expired ones. The column was NOT NULL, so those webhooks
-- stored a zero timestamp instead.
ALTER TABLE subscriptions ALTER COLUMN renews_at DROP NOT NULL;

-- Zero timestamps written before the column became nullable are not real dates.
UPDATE subscriptions SET renews_at = NULL WHERE renews_at = '0001-01-01 00:00:00';
