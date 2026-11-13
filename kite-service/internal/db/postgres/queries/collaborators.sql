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

-- name: DeleteExpiredCollaboratorInvites :exec
DELETE FROM collaborator_invites WHERE expires_at <= @now;

-- name: GetCollaboratorInvite :one
SELECT * FROM collaborator_invites
WHERE app_id = $1 AND discord_username = $2;

-- Invites of users that already used Kite are matched by their ID, the others
-- by the Discord username they were invited with.
-- name: GetCollaboratorInvitesByUser :many
SELECT sqlc.embed(collaborator_invites), sqlc.embed(apps),
    inviter.display_name AS inviter_display_name,
    inviter.discord_username AS inviter_discord_username,
    inviter.discord_id AS inviter_discord_id,
    inviter.discord_avatar AS inviter_discord_avatar
FROM collaborator_invites
JOIN apps ON apps.id = collaborator_invites.app_id
LEFT JOIN users AS inviter ON inviter.id = collaborator_invites.invited_by_user_id
WHERE (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)))
    AND collaborator_invites.expires_at > @now::timestamp
    AND apps.owner_user_id != @user_id::text
ORDER BY collaborator_invites.created_at DESC;

-- A user can match two invites to the app, one by their ID and one by a
-- username they took later. The one by ID wins, and both are deleted, as
-- they're a collaborator now. An existing collaborator keeps their role.
-- name: AcceptCollaboratorInvite :one
WITH invite AS (
    SELECT collaborator_invites.app_id, collaborator_invites.role FROM collaborator_invites
    JOIN apps ON apps.id = collaborator_invites.app_id
    WHERE collaborator_invites.app_id = @app_id::text
        AND (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)))
        AND collaborator_invites.expires_at > @now::timestamp
        AND apps.owner_user_id != @user_id::text
    ORDER BY collaborator_invites.user_id IS NULL, collaborator_invites.created_at DESC
    LIMIT 1
), deleted AS (
    DELETE FROM collaborator_invites
    WHERE collaborator_invites.app_id IN (SELECT invite.app_id FROM invite)
        AND (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)))
)
INSERT INTO collaborators (app_id, user_id, role, created_at, updated_at)
SELECT invite.app_id, @user_id::text, invite.role, @now::timestamp, @now::timestamp FROM invite
ON CONFLICT (user_id, app_id) DO NOTHING
RETURNING *;

-- name: DeclineCollaboratorInvite :execrows
DELETE FROM collaborator_invites
WHERE collaborator_invites.app_id = @app_id::text
    AND (collaborator_invites.user_id = @user_id::text OR (collaborator_invites.user_id IS NULL AND collaborator_invites.discord_username = lower(@discord_username::text)));
