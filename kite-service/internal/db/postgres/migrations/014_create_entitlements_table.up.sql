CREATE TABLE IF NOT EXISTS entitlements (
    id TEXT PRIMARY KEY,
    sku_id TEXT NOT NULL,
    starts_at TIMESTAMP,
    ends_at TIMESTAMP,
    deleted BOOLEAN NOT NULL DEFAULT FALSE,

    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    creator_user_id TEXT NOT NULL,
   
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS entitlements_app_id ON entitlements (app_id);
