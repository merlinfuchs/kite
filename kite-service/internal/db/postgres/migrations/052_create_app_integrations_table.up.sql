-- The integrations an app set up, and whether they're enabled. Without a
-- row, the integration's default applies: enabled for those every app can use
-- or that are on by default, disabled for opt-in ones. The credential of an
-- integration that needs one is a row in app_secrets that references its row
-- here, so removing the integration removes its credential.
CREATE TABLE IF NOT EXISTS app_integrations (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    -- The integration defined in code, like cookie_api.
    integration_id TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    UNIQUE (app_id, integration_id)
);
