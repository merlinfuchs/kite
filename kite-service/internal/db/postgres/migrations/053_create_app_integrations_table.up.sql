-- Whether an app turned an integration on or off. Only integrations without a
-- credential are turned on and off like this, those with one are on while the
-- app has connected them. Without a row, the integration's default applies.
CREATE TABLE IF NOT EXISTS app_integrations (
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    integration_id TEXT NOT NULL,
    enabled BOOLEAN NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    PRIMARY KEY (app_id, integration_id)
);
