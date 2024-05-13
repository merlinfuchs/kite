-- name: UpsertUser :one
INSERT INTO users (
    id,
    username,
    email,
    discriminator,
    global_name,
    avatar,
    public_flags,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9
) ON CONFLICT (id) DO UPDATE SET
    username = EXCLUDED.username,
    email = EXCLUDED.email,
    discriminator = EXCLUDED.discriminator,
    global_name = EXCLUDED.global_name,
    avatar = EXCLUDED.avatar,
    public_flags = EXCLUDED.public_flags,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;