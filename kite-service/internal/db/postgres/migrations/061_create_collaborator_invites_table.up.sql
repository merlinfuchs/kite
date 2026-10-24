-- Collaborators added by Discord username before they logged in to Kite. They
-- become collaborators when a user with that username logs in for the first
-- time.
CREATE TABLE IF NOT EXISTS collaborator_invites (
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    discord_username TEXT NOT NULL,
    role TEXT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,

    PRIMARY KEY (app_id, discord_username)
);

CREATE INDEX IF NOT EXISTS collaborator_invites_discord_username ON collaborator_invites (discord_username);
CREATE INDEX IF NOT EXISTS users_discord_username ON users (lower(discord_username));
