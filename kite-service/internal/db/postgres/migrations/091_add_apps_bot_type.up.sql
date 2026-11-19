-- Own apps run on a bot of their own, with its token. Shared apps run on the
-- Kite bot in one server, see design/kite-lite.md, and have no token or
-- Discord application of their own.
ALTER TABLE apps
    ADD COLUMN IF NOT EXISTS bot_type TEXT NOT NULL DEFAULT 'own',
    ADD COLUMN IF NOT EXISTS discord_guild_id TEXT,
    ALTER COLUMN discord_token DROP NOT NULL,
    ALTER COLUMN discord_id DROP NOT NULL;

ALTER TABLE apps ADD CONSTRAINT apps_bot_type_check CHECK (
    (bot_type = 'own' AND discord_token IS NOT NULL AND discord_id IS NOT NULL AND discord_guild_id IS NULL)
    OR (bot_type = 'shared' AND discord_token IS NULL AND discord_id IS NULL AND discord_guild_id IS NOT NULL)
);

-- A server has at most one shared app.
CREATE UNIQUE INDEX IF NOT EXISTS apps_shared_discord_guild_id ON apps (discord_guild_id) WHERE bot_type = 'shared';
