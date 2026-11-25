ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS paddle_subscription_id,
    DROP COLUMN IF EXISTS paddle_customer_id,
    DROP COLUMN IF EXISTS paddle_product_id,
    DROP COLUMN IF EXISTS paddle_price_id;
