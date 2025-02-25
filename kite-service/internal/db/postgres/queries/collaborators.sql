-- name: GetCollaborator :one
SELECT sqlc.embed(collaborators), sqlc.embed(users) FROM collaborators
LEFT JOIN users ON collaborators.user_id = users.id
WHERE app_id = $1 AND user_id = $2;

-- name: GetCollaboratorsByApp :many
SELECT sqlc.embed(collaborators), sqlc.embed(users) FROM collaborators
LEFT JOIN users ON collaborators.user_id = users.id
WHERE app_id = $1;

-- name: CountCollaboratorsByApp :one
SELECT COUNT(*) FROM collaborators
WHERE app_id = $1;

-- name: CreateCollaborator :one
INSERT INTO collaborators (app_id, user_id, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateCollaborator :one
UPDATE collaborators
SET role = $3, updated_at = $4
WHERE app_id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteCollaborator :exec
DELETE FROM collaborators
WHERE app_id = $1 AND user_id = $2;

