-- name: CreateShareCode :one
INSERT INTO share_codes (code, data, created_at) VALUES ($1, $2, $3) RETURNING *;

-- name: GetShareCode :one
SELECT * FROM share_codes WHERE code = $1;
