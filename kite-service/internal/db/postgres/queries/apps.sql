-- name: CreateApp :exec
INSERT INTO apps (
    id,
    owner_user_id,
    token,
    token_invalid,
    public_key,
    user_id,
    user_name,
    user_discriminator,
    user_avatar,
    user_banner,
    user_bio,
    status_type,
    status_activity_type,
    status_activity_name,
    status_activity_state,
    status_activity_url,
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
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18
);

-- name: UpdateApp :one
UPDATE apps SET 
    token = $2, 
    token_invalid = false,
    public_key = $3,
    user_id = $4,
    user_name = $5,
    user_discriminator = $6,
    user_avatar = $7,
    user_banner = $8,
    user_bio = $9,
    updated_at = $10
WHERE id = $1 
RETURNING *;

-- name: UpdateAppStatus :one
UPDATE apps SET 
    status_type = $2, 
    status_activity_type = $3, 
    status_activity_name = $4, 
    status_activity_state = $5, 
    status_activity_url = $6,
    updated_at = $7
WHERE id = $1 
RETURNING *;

-- name: GetApp :one
SELECT * FROM apps WHERE id = $1;

-- name: GetAppForOwnerUser :one
SELECT * FROM apps WHERE id = $1 AND owner_user_id = $2;

-- name: DeleteApp :exec
DELETE FROM apps WHERE id = $1;

-- name: GetAppsWithValidToken :many
SELECT * FROM apps WHERE token_invalid = false;

-- name: GetAppsForOwnerUser :many
SELECT * FROM apps WHERE owner_user_id = $1;

-- name: GetDistinctAppIDs :many
SELECT DISTINCT id FROM apps;

-- name: CheckUserIsOwnerOfApp :one
SELECT true FROM apps WHERE id = $1 AND owner_user_id = $2;

-- name: SetAppTokenInvalid :exec
UPDATE apps SET token_invalid = true WHERE id = $1;