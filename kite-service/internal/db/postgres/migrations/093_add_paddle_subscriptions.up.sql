-- Paddle replaces Lemon Squeezy. Until every subscription has moved, a
-- subscription is from either, see source.
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS paddle_subscription_id TEXT UNIQUE,
    ADD COLUMN IF NOT EXISTS paddle_customer_id TEXT,
    ADD COLUMN IF NOT EXISTS paddle_product_id TEXT,
    ADD COLUMN IF NOT EXISTS paddle_price_id TEXT;
