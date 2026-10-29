-- The version apps install: the one a review approved last of a user's
-- bundle, otherwise the newest. The listing columns are null for bundles that
-- aren't listed.
-- name: GetBundleCurrentVersion :one
SELECT
    sqlc.embed(bundles),
    bundle_versions.content,
    (SELECT count(*) FROM bundle_installs WHERE bundle_installs.bundle_id = bundles.id)::int AS install_count,
    (SELECT count(*) FROM bundle_stars WHERE bundle_stars.bundle_id = bundles.id)::int AS star_count,
    EXISTS (
        SELECT 1 FROM bundle_reviews
        WHERE bundle_reviews.bundle_id = bundles.id AND bundle_reviews.status = 'pending'
    )::bool AS in_review,
    bundle_listings.name AS listing_name,
    bundle_listings.description AS listing_description,
    bundle_listings.icon AS listing_icon,
    bundle_listings.category AS listing_category,
    bundle_listings.about AS listing_about,
    bundle_listings.faq AS listing_faq,
    bundle_listings.setup AS listing_setup,
    bundle_listings.position AS listing_position,
    bundle_listings.updated_at AS listing_updated_at
FROM bundles
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id
    AND bundle_versions.version = COALESCE(bundles.approved_version, bundles.version)
WHERE bundles.id = $1;

-- name: GetOfficialBundlesCurrentVersion :many
SELECT
    sqlc.embed(bundles),
    sqlc.embed(bundle_listings),
    bundle_versions.content,
    (SELECT count(*) FROM bundle_installs WHERE bundle_installs.bundle_id = bundles.id)::int AS install_count,
    (SELECT count(*) FROM bundle_stars WHERE bundle_stars.bundle_id = bundles.id)::int AS star_count
FROM bundles
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
WHERE bundles.owner_user_id IS NULL
ORDER BY bundle_listings.position, bundles.id;

-- name: UpsertOfficialBundle :execrows
INSERT INTO bundles (
    id,
    name,
    description,
    version,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    version = EXCLUDED.version,
    updated_at = EXCLUDED.updated_at
WHERE bundles.owner_user_id IS NULL;

-- name: UpsertBundleListing :exec
INSERT INTO bundle_listings (
    bundle_id,
    name,
    description,
    icon,
    category,
    about,
    faq,
    setup,
    position,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
ON CONFLICT (bundle_id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    icon = EXCLUDED.icon,
    category = EXCLUDED.category,
    about = EXCLUDED.about,
    faq = EXCLUDED.faq,
    setup = EXCLUDED.setup,
    position = EXCLUDED.position,
    updated_at = EXCLUDED.updated_at;

-- name: UpsertBundleVersion :exec
INSERT INTO bundle_versions (
    bundle_id,
    version,
    content,
    notes,
    created_at
) VALUES (
    $1, $2, $3, $4, $5
)
ON CONFLICT (bundle_id, version) DO UPDATE SET content = EXCLUDED.content, notes = EXCLUDED.notes;

-- name: SetBundleVersionNotes :exec
UPDATE bundle_versions SET notes = $3 WHERE bundle_id = $1 AND version = $2;

-- name: GetOfficialBundleVersionNotes :many
SELECT bundle_versions.bundle_id, bundle_versions.version, bundle_versions.notes, bundle_versions.created_at
FROM bundle_versions
JOIN bundles ON bundles.id = bundle_versions.bundle_id
WHERE bundles.owner_user_id IS NULL AND bundle_versions.notes <> ''
ORDER BY bundle_versions.version DESC;

-- Without the content, only how many items the newest version has, and how
-- many apps added it.
-- name: GetBundlesByOwner :many
SELECT
    sqlc.embed(bundles),
    (bundle_listings.bundle_id IS NOT NULL)::bool AS listed,
    (SELECT count(*) FROM bundle_installs WHERE bundle_installs.bundle_id = bundles.id)::int AS install_count,
    (
        jsonb_array_length(COALESCE(bundle_versions.content->'commands', '[]'))
        + jsonb_array_length(COALESCE(bundle_versions.content->'event_listeners', '[]'))
        + jsonb_array_length(COALESCE(bundle_versions.content->'messages', '[]'))
        + jsonb_array_length(COALESCE(bundle_versions.content->'variables', '[]'))
    )::int AS item_count
FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
WHERE bundles.owner_user_id = $1
ORDER BY bundles.created_at DESC, bundles.id;

-- name: CountBundlesByOwner :one
SELECT COUNT(*) FROM bundles WHERE owner_user_id = $1;

-- name: CreateBundle :exec
INSERT INTO bundles (
    id,
    name,
    description,
    owner_user_id,
    source_app_id,
    version,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- Only updates a bundle that is still at from_version, so two versions
-- published at once can't both apply.
-- name: UpdateBundleVersion :one
UPDATE bundles SET
    version = @version,
    updated_at = @updated_at
WHERE id = @id AND version = @from_version
RETURNING *;

-- name: UpdateBundlePage :execrows
UPDATE bundles SET
    name = @name,
    description = @description,
    icon = @icon,
    category = @category,
    about = @about,
    setup = @setup
WHERE id = @id AND removed_at IS NULL;

-- name: CreateBundleVersion :exec
INSERT INTO bundle_versions (
    bundle_id,
    version,
    content,
    notes,
    created_at
) VALUES (
    $1, $2, $3, $4, $5
);

-- A bundle Kite took down stays, with its reports.
-- name: DeleteBundle :execrows
DELETE FROM bundles WHERE id = $1 AND removed_at IS NULL;

-- Locks the owner's bundles, so two bundles created at once can't both pass
-- the limit.
-- name: LockBundleOwner :exec
SELECT pg_advisory_xact_lock(hashtext('bundles:' || @owner_user_id::text));

-- Keeps the approved version, which apps still install.
-- name: DeleteOldBundleVersions :exec
DELETE FROM bundle_versions
WHERE bundle_versions.bundle_id = @bundle_id AND bundle_versions.version <= @before_version
    AND bundle_versions.version <> COALESCE((SELECT bundles.approved_version FROM bundles WHERE bundles.id = @bundle_id), 0);

-- name: GetBundleVersionNotes :many
SELECT version, notes, created_at FROM bundle_versions
WHERE bundle_id = $1 AND notes <> ''
ORDER BY version DESC;

-- Users' bundles listed in the marketplace, with their approved version and
-- who made them, newest first. A bundle that was taken down has no listing, but
-- the filter makes sure.
-- name: GetListedUserBundles :many
SELECT
    sqlc.embed(bundles),
    sqlc.embed(bundle_listings),
    bundle_versions.content,
    sqlc.embed(users),
    (SELECT count(*) FROM bundle_installs WHERE bundle_installs.bundle_id = bundles.id)::int AS install_count,
    (SELECT count(*) FROM bundle_stars WHERE bundle_stars.bundle_id = bundles.id)::int AS star_count
FROM bundles
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.approved_version
JOIN users ON users.id = bundles.owner_user_id
WHERE bundles.removed_at IS NULL
ORDER BY bundles.created_at DESC, bundles.id;

-- The bundles of the marketplace with only the ID, owner, source app and
-- the name, description and category of their listing: Kite's in their
-- order, then the users' ones the most apps added.
-- name: GetMarketplaceTemplates :many
SELECT
    bundles.id,
    bundles.owner_user_id,
    bundles.source_app_id,
    bundle_listings.name,
    bundle_listings.description,
    bundle_listings.category
FROM bundles
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
WHERE bundles.owner_user_id IS NULL OR (bundles.approved_version > 0 AND bundles.removed_at IS NULL)
ORDER BY
    bundles.owner_user_id IS NOT NULL,
    CASE WHEN bundles.owner_user_id IS NULL THEN bundle_listings.position END,
    (SELECT count(*) FROM bundle_installs WHERE bundle_installs.bundle_id = bundles.id) DESC,
    bundles.id
LIMIT $1;

-- The Discord IDs of the users with a bundle listed in the marketplace that
-- apps of at least min_installers other users installed, who get the creator
-- role in the Kite server. Users banned from the marketplace or Kite don't,
-- even while what they listed before is still listed.
-- name: GetCreatorDiscordIDs :many
SELECT DISTINCT users.discord_id
FROM bundles
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN users ON users.id = bundles.owner_user_id
WHERE bundles.removed_at IS NULL
    AND NOT EXISTS (
        SELECT 1 FROM user_bans
        WHERE user_bans.user_id = users.id AND user_bans.feature IN ('marketplace', 'kite')
    )
    AND (
        SELECT count(DISTINCT apps.owner_user_id)
        FROM bundle_installs
        JOIN apps ON apps.id = bundle_installs.app_id
        WHERE bundle_installs.bundle_id = bundles.id
            AND apps.owner_user_id != bundles.owner_user_id
    ) >= sqlc.arg(min_installers)::int;

-- The bundles without their content, with the icon of their listing and who
-- made them, for lists like the installs of an app.
-- name: GetBundleSummaries :many
SELECT
    sqlc.embed(bundles),
    bundle_listings.icon AS listing_icon,
    users.display_name AS owner_display_name,
    users.discord_id AS owner_discord_id,
    users.discord_username AS owner_discord_username,
    users.discord_avatar AS owner_discord_avatar
FROM bundles
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
LEFT JOIN users ON users.id = bundles.owner_user_id
WHERE bundles.id = ANY(@ids::text[]);

-- name: SetBundleApprovedVersion :exec
UPDATE bundles SET approved_version = @approved_version WHERE id = @id;
