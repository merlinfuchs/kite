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
INSERT INTO collaborator_invites (app_id, discord_username, role, user_id, invited_by_user_id, created_at, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (app_id, discord_username) DO UPDATE SET
    role = EXCLUDED.role,
    user_id = EXCLUDED.user_id,
    invited_by_user_id = EXCLUDED.invited_by_user_id,
    created_at = EXCLUDED.created_at,
    expires_at = EXCLUDED.expires_at
RETURNING *;

-- name: DeleteCollaboratorInvite :exec
DELETE FROM collaborator_invites
WHERE app_id = $1 AND discord_username = $2;

-- name: GetCollaboratorInvite :one
SELECT * FROM collaborator_invites
WHERE app_id = $1 AND discord_username = $2;

-- Invites of users that already used Kite are matched by their ID, the others
-- by the Discord username they were invited with.

-- name: GetCollaboratorInvitesByUser :many
SELECT sqlc.embed(collaborator_invites), sqlc.embed(apps) FROM collaborator_invites
JOIN apps ON apps.id = collaborator_invites.app_id
WHERE (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)))
    AND collaborator_invites.expires_at > @now::timestamp
    AND apps.owner_user_id != @user_id::text
ORDER BY collaborator_invites.created_at DESC;

-- name: AcceptCollaboratorInvite :one
WITH accepted AS (
    DELETE FROM collaborator_invites
    WHERE collaborator_invites.app_id = @app_id::text
        AND (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)))
        AND collaborator_invites.expires_at > @now::timestamp
    RETURNING *
)
INSERT INTO collaborators (app_id, user_id, role, created_at, updated_at)
SELECT accepted.app_id, @user_id::text, accepted.role, @now::timestamp, @now::timestamp FROM accepted
JOIN apps ON apps.id = accepted.app_id
WHERE apps.owner_user_id != @user_id::text
LIMIT 1
ON CONFLICT (user_id, app_id) DO UPDATE SET role = EXCLUDED.role, updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: DeclineCollaboratorInvite :execrows
DELETE FROM collaborator_invites
WHERE collaborator_invites.app_id = @app_id::text
    AND (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)));
