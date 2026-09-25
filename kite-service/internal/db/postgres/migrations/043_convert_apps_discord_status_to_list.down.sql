UPDATE apps
SET discord_status = COALESCE(discord_status -> 'statuses' -> 0, '{}'::jsonb) - 'id' - 'label'
WHERE discord_status ? 'statuses';
