CREATE TABLE IF NOT EXISTS entitlements (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL, -- "subscription", "manual"

    subscription_id TEXT UNIQUE REFERENCES subscriptions(id) ON DELETE CASCADE,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

    feature_usage_credits_per_month INTEGER NOT NULL,
    feature_max_collaborator INTEGER NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS entitlements_subscription_id ON entitlements (subscription_id);
CREATE INDEX IF NOT EXISTS entitlements_app_id ON entitlements (app_id);
