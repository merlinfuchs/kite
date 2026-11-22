-- Shared apps can't exist without the columns, so they're deleted.
DELETE FROM apps WHERE bot_type = 'shared';

DROP INDEX IF EXISTS apps_shared_discord_guild_id;
ALTER TABLE apps DROP CONSTRAINT IF EXISTS apps_bot_type_check;
ALTER TABLE apps
    DROP COLUMN IF EXISTS discord_guild_id,
    DROP COLUMN IF EXISTS bot_type,
    ALTER COLUMN discord_token SET NOT NULL,
    ALTER COLUMN discord_id SET NOT NULL;
