-- name: GetUserBadgeGrants :many
SELECT sqlc.embed(user_badge_grants), sqlc.embed(users)
FROM user_badge_grants
JOIN users ON users.id = user_badge_grants.user_id
ORDER BY user_badge_grants.created_at DESC, user_badge_grants.user_id, user_badge_grants.badge;

-- name: GetUserBadgeGrantsByUserIDs :many
SELECT * FROM user_badge_grants WHERE user_id = ANY(@user_ids::text[]);

-- name: UpsertUserBadgeGrant :exec
INSERT INTO user_badge_grants (
    user_id,
    badge,
    granted_by_user_id,
    created_at
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (user_id, badge) DO NOTHING;

-- name: DeleteUserBadgeGrant :execrows
DELETE FROM user_badge_grants WHERE user_id = $1 AND badge = $2;
