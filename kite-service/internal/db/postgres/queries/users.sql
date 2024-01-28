-- name: UpsertUser :one
INSERT INTO users (
    id,
    username,
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
    $8
) ON CONFLICT (id) DO UPDATE SET
    username = $2,
    discriminator = $3,
    global_name = $4,
    avatar = $5,
    public_flags = $6,
    updated_at = $8
RETURNING *;

-- name: GetUser :one
SELECT * FROM users WHERE id = $1;