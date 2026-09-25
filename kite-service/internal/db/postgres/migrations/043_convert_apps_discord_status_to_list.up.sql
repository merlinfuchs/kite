-- Apps used to store a single status directly in discord_status. Turn it into a
-- list with one active entry.
UPDATE apps
SET discord_status = jsonb_build_object(
    'statuses', jsonb_build_array(discord_status || '{"id": "default"}'::jsonb),
    'active_id', 'default'
)
WHERE discord_status IS NOT NULL
  AND jsonb_typeof(discord_status) = 'object'
  AND NOT discord_status ? 'statuses';
