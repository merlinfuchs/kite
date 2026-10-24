-- name: GetBundleCurrentVersion :one
SELECT sqlc.embed(bundles), bundle_versions.content FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
WHERE bundles.id = $1;

-- name: GetListedBundleCurrentVersion :one
SELECT sqlc.embed(bundles), sqlc.embed(bundle_listings), bundle_versions.content FROM bundles
JOIN bundle_listings ON bundle_listings.bundle_id = bundles.id
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
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
