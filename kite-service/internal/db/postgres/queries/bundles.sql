-- The listing columns are null for bundles that weren't published.
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
    bundle_listings.updated_at AS listing_updated_at
FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
LEFT JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
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
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
ON CONFLICT (bundle_id) DO UPDATE SET
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
    created_at
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (bundle_id, version) DO UPDATE SET content = EXCLUDED.content;

-- Without the content, only how many items the current version has.
-- name: GetBundlesByOwner :many
SELECT
    sqlc.embed(bundles),
    (
        jsonb_array_length(COALESCE(bundle_versions.content->'commands', '[]'))
        + jsonb_array_length(COALESCE(bundle_versions.content->'event_listeners', '[]'))
        + jsonb_array_length(COALESCE(bundle_versions.content->'messages', '[]'))
        + jsonb_array_length(COALESCE(bundle_versions.content->'variables', '[]'))
    )::int AS item_count
FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
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
    created_at
) VALUES (
    $1, $2, $3, $4
);

-- name: DeleteBundle :exec
DELETE FROM bundles WHERE id = $1;

-- Locks the owner's bundles, so two bundles created at once can't both pass
-- the limit.
-- name: LockBundleOwner :exec
SELECT pg_advisory_xact_lock(hashtext('bundles:' || @owner_user_id::text));

-- name: DeleteOldBundleVersions :exec
DELETE FROM bundle_versions WHERE bundle_id = @bundle_id AND version <= @before_version;
