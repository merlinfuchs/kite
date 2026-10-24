-- name: GetBundleCurrentVersion :one
SELECT sqlc.embed(bundles), bundle_versions.content FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
WHERE bundles.id = $1;

-- name: GetOfficialBundlesCurrentVersion :many
SELECT sqlc.embed(bundles), bundle_versions.content FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id AND bundle_versions.version = bundles.version
WHERE bundles.owner_user_id IS NULL
ORDER BY bundles.id;

-- name: UpsertOfficialBundle :execrows
INSERT INTO bundles (
    id,
    name,
    description,
    version,
    icon,
    category,
    about,
    faq,
    setup,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    version = EXCLUDED.version,
    icon = EXCLUDED.icon,
    category = EXCLUDED.category,
    about = EXCLUDED.about,
    faq = EXCLUDED.faq,
    setup = EXCLUDED.setup,
    updated_at = EXCLUDED.updated_at
WHERE bundles.owner_user_id IS NULL;

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
