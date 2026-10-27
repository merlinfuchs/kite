-- name: UpsertUserBan :exec
INSERT INTO user_bans (
    user_id,
    feature,
    reason,
    banned_by_user_id,
    created_at
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (user_id, feature) DO UPDATE SET
    reason = EXCLUDED.reason,
    banned_by_user_id = EXCLUDED.banned_by_user_id,
    created_at = EXCLUDED.created_at;

-- name: GetUserBan :one
SELECT * FROM user_bans WHERE user_id = $1 AND feature = $2;

-- name: GetUserBans :many
SELECT sqlc.embed(user_bans), sqlc.embed(users)
FROM user_bans
JOIN users ON users.id = user_bans.user_id
ORDER BY user_bans.created_at DESC, user_bans.user_id, user_bans.feature;

-- name: DeleteUserBan :execrows
DELETE FROM user_bans WHERE user_id = $1 AND feature = $2;
