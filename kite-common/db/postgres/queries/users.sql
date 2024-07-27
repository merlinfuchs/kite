-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

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
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) ON CONFLICT (discord_id) 
DO UPDATE SET 
    email = EXCLUDED.email,
    display_name = EXCLUDED.display_name,
    discord_id = EXCLUDED.discord_id,
    discord_username = EXCLUDED.discord_username,
    discord_avatar = EXCLUDED.discord_avatar,
    updated_at = EXCLUDED.updated_at
RETURNING *;
