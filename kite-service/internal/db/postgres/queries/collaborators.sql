-- name: GetCollaborator :one
SELECT sqlc.embed(collaborators), sqlc.embed(users) FROM collaborators
LEFT JOIN users ON collaborators.user_id = users.id
WHERE app_id = $1 AND user_id = $2;

-- name: GetCollaboratorsByApp :many
SELECT sqlc.embed(collaborators), sqlc.embed(users) FROM collaborators
LEFT JOIN users ON collaborators.user_id = users.id
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


-- name: GetCollaboratorInvitesByApp :many
SELECT * FROM collaborator_invites
WHERE app_id = @app_id AND expires_at > @now
ORDER BY created_at;

-- name: CountCollaboratorsAndInvitesByApp :one
SELECT
    (SELECT COUNT(*) FROM collaborators WHERE collaborators.app_id = @app_id) +
    (SELECT COUNT(*) FROM collaborator_invites WHERE collaborator_invites.app_id = @app_id AND expires_at > @now);

-- name: UpsertCollaboratorInvite :one
INSERT INTO collaborator_invites (app_id, discord_username, role, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (app_id, discord_username) DO UPDATE SET
    role = EXCLUDED.role,
    created_at = EXCLUDED.created_at,
    expires_at = EXCLUDED.expires_at
RETURNING *;

-- name: DeleteCollaboratorInvite :exec
DELETE FROM collaborator_invites
WHERE app_id = $1 AND discord_username = $2;

-- name: ClaimCollaboratorInvites :exec
WITH claimed AS (
    DELETE FROM collaborator_invites
    WHERE discord_username = lower(@discord_username)
    RETURNING *
)
INSERT INTO collaborators (app_id, user_id, role, created_at, updated_at)
SELECT claimed.app_id, @user_id::text, claimed.role, @now::timestamp, @now::timestamp FROM claimed
JOIN apps ON apps.id = claimed.app_id
WHERE claimed.expires_at > @now::timestamp AND apps.owner_user_id != @user_id::text
ON CONFLICT DO NOTHING;
