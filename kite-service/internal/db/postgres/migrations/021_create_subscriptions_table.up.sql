CREATE TABLE IF NOT EXISTS subscriptions (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    source TEXT NOT NULL, -- always "lemonsqueezy"
    status TEXT NOT NULL, -- "on_trial", "active", "paused", "past_due", "unpaid", "canceled", "expired"
    status_formatted TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    renews_at TIMESTAMP NOT NULL,
    trial_ends_at TIMESTAMP,
    ends_at TIMESTAMP,

    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    lemonsqueezy_subscription_id TEXT UNIQUE,
    lemonsqueezy_customer_id TEXT,
    lemonsqueezy_order_id TEXT,
    lemonsqueezy_product_id TEXT,
    lemonsqueezy_variant_id TEXT
);

CREATE INDEX IF NOT EXISTS subscriptions_user_id ON subscriptions (user_id);
