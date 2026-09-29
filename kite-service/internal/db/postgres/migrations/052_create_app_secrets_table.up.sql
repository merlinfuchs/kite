-- Values flows use without them being in the flow data. A secret either has a
-- name, which HTTP blocks reference as {{secrets.NAME}}, or belongs to an
-- integration, whose requests use it as their credential.
CREATE TABLE IF NOT EXISTS app_secrets (
    id TEXT PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    name TEXT,
    integration_id TEXT,
    -- Encrypted like apps.discord_token, and never sent back to the client.
    value_encrypted TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,

    CHECK ((name IS NULL) <> (integration_id IS NULL))
);

CREATE UNIQUE INDEX IF NOT EXISTS app_secrets_app_id_name ON app_secrets (app_id, name) WHERE name IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS app_secrets_app_id_integration_id ON app_secrets (app_id, integration_id) WHERE integration_id IS NOT NULL;
