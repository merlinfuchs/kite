-- name: UpserGuild :one
INSERT INTO guilds (
    id, 
    name, 
    icon, 
    description, 
    created_at, 
    updated_at
) 
VALUES (
    $1, 
    $2, 
    $3, 
    $4, 
    $5, 
    $6
) 
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name, 
    icon = EXCLUDED.icon,
    description = EXCLUDED.description, 
    updated_at = $6 
RETURNING *;

-- name: GetGuilds :many
SELECT * FROM guilds ORDER BY name DESC;

-- name: GetGuild :one
SELECT * FROM guilds WHERE id = $1;

-- name: GetDistinctGuildIDs :many
SELECT DISTINCT id FROM guilds;