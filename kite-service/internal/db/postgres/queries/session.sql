-- name: CreateSession :one
INSERT INTO sessions (key_hash, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- With the reason if Kite banned the user, so checking needs no extra query.
-- name: GetSession :one
SELECT sqlc.embed(sessions), user_bans.reason AS kite_ban_reason
FROM sessions
LEFT JOIN user_bans ON user_bans.user_id = sessions.user_id AND user_bans.feature = 'kite'
WHERE sessions.key_hash = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE key_hash = $1;

-- name: UpdateSessionExpiry :execrows
UPDATE sessions SET expires_at = $2 WHERE key_hash = $1;
