-- Invites now wait for the user to accept them instead of being claimed on
-- login. Invites of users that already use Kite have their user ID, so a
-- changed Discord username doesn't lose them.
ALTER TABLE collaborator_invites ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES users(id) ON DELETE CASCADE;
ALTER TABLE collaborator_invites ADD COLUMN IF NOT EXISTS invited_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS collaborator_invites_user_id ON collaborator_invites (user_id);

-- What collaborators did in an app, for its audit log.
CREATE TABLE IF NOT EXISTS app_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    -- NULL once the user deleted their account.
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL DEFAULT '',
    target_id TEXT NOT NULL DEFAULT '',
    target_name TEXT NOT NULL DEFAULT '',
    -- Set when the change was made with an API token, like by an MCP client.
    api_token_id TEXT,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS app_audit_logs_app_id_id ON app_audit_logs (app_id, id DESC);
