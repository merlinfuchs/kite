-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByDiscordID :one
SELECT * FROM users WHERE discord_id = $1;

-- name: UpsertUser :one
INSERT INTO users (
    id,
    email,
    display_name, 
    discord_id,
    discord_username,
    discord_avatar,
    signup_referrer,
    signup_utm_source,
    signup_utm_medium,
    signup_utm_campaign,
    signup_landing_page,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) ON CONFLICT (discord_id) 
DO UPDATE SET 
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    discord_id = EXCLUDED.discord_id,
    discord_username = EXCLUDED.discord_username,
    discord_avatar = EXCLUDED.discord_avatar,
    updated_at = EXCLUDED.updated_at
RETURNING *;
