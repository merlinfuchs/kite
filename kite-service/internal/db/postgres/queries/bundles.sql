-- The version apps install: the listed one of a listed user's bundle,
-- otherwise the newest. The listing columns are null for bundles that aren't
-- listed.
-- name: GetBundleCurrentVersion :one
SELECT
    sqlc.embed(bundles),
    bundle_versions.content,
    bundle_listings.icon AS listing_icon,
    bundle_listings.category AS listing_category,
    bundle_listings.about AS listing_about,
    bundle_listings.faq AS listing_faq,
    bundle_listings.setup AS listing_setup,
    bundle_listings.position AS listing_position,
    bundle_listings.listed_version AS listing_listed_version,
    bundle_listings.updated_at AS listing_updated_at
FROM bundles
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id
    AND bundle_versions.version = COALESCE(bundle_listings.listed_version, bundles.version)
WHERE bundles.id = $1;

-- name: GetOfficialBundlesCurrentVersion :many
SELECT sqlc.embed(bundles), sqlc.embed(bundle_listings), bundle_versions.content FROM bundles
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
    icon,
    category,
    about,
    faq,
    setup,
    position,
    listed_version,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
ON CONFLICT (bundle_id) DO UPDATE SET
    icon = EXCLUDED.icon,
    category = EXCLUDED.category,
    about = EXCLUDED.about,
    faq = EXCLUDED.faq,
    setup = EXCLUDED.setup,
    position = EXCLUDED.position,
    listed_version = EXCLUDED.listed_version,
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

-- Without the content, only how many items the newest version has.
-- name: GetBundlesByOwner :many
SELECT
    sqlc.embed(bundles),
    bundle_listings.listed_version AS listed_version,
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
-- name: UpdateBundleVersion :execrows
UPDATE bundles SET
    name = @name,
    description = @description,
    version = @version,
    updated_at = @updated_at
WHERE id = @id AND version = @from_version;

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

-- name: DeleteBundle :exec
DELETE FROM bundles WHERE id = $1;

-- Locks the owner's bundles, so two bundles created at once can't both pass
-- the limit.
-- name: LockBundleOwner :exec
SELECT pg_advisory_xact_lock(hashtext('bundles:' || @owner_user_id::text));

-- Keeps the listed version, which apps still install.
-- name: DeleteOldBundleVersions :exec
DELETE FROM bundle_versions
WHERE bundle_versions.bundle_id = @bundle_id AND bundle_versions.version <= @before_version
    AND bundle_versions.version <> COALESCE((SELECT bundle_listings.listed_version FROM bundle_listings WHERE bundle_listings.bundle_id = @bundle_id), 0);

-- name: SetBundlePublishedAt :exec
UPDATE bundles SET published_at = $2 WHERE id = $1;

-- name: GetBundleVersionNotes :many
SELECT version, notes, created_at FROM bundle_versions
WHERE bundle_id = $1 AND notes <> ''
ORDER BY version DESC;

-- Users' bundles listed in the marketplace, with their listed version and who
-- made them, newest first. A bundle that isn't shared, like one that was taken
-- down, has no listing, but the filter makes sure.
-- name: GetListedUserBundles :many
SELECT
    sqlc.embed(bundles),
    sqlc.embed(bundle_listings),
    bundle_versions.content,
    sqlc.embed(users)
FROM bundles
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundle_listings.listed_version
JOIN users ON users.id = bundles.owner_user_id
WHERE bundles.published_at IS NOT NULL
ORDER BY bundles.created_at DESC, bundles.id;

-- The bundles without their content, with the icon and listed version of
-- their listing and who made them, for lists like the installs of an app.
-- name: GetBundleSummaries :many
SELECT
    sqlc.embed(bundles),
    bundle_listings.icon AS listing_icon,
    bundle_listings.listed_version AS listed_version,
    users.display_name AS owner_display_name,
    users.discord_id AS owner_discord_id,
    users.discord_username AS owner_discord_username,
    users.discord_avatar AS owner_discord_avatar
FROM bundles
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
LEFT JOIN users ON users.id = bundles.owner_user_id
WHERE bundles.id = ANY(@ids::text[]);
