UPDATE subscriptions SET renews_at = '0001-01-01 00:00:00' WHERE renews_at IS NULL;

ALTER TABLE subscriptions ALTER COLUMN renews_at SET NOT NULL;
