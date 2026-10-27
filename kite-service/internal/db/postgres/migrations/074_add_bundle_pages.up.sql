-- The page of a user's bundle as its owner last wrote it, next to its name and
-- description. Its link shows it, and the marketplace once a review approves
-- it. It changes without a new version.
ALTER TABLE bundles
    ADD COLUMN IF NOT EXISTS icon TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS about JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS setup JSONB NOT NULL DEFAULT '[]';

-- Owners wrote their page in their last submission.
UPDATE bundles SET
    icon = latest.listing->>'icon',
    category = latest.listing->>'category',
    about = COALESCE(latest.listing->'about', '[]'),
    setup = COALESCE(latest.listing->'setup', '[]')
FROM (
    SELECT DISTINCT ON (bundle_id) bundle_id, listing
    FROM bundle_reviews
    ORDER BY bundle_id, status = 'pending' DESC, decided_at DESC NULLS LAST, id
) latest
WHERE latest.bundle_id = bundles.id;

-- Listings and reviews have the name and description too, which were part of
-- the version before.
ALTER TABLE bundle_listings
    ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '';

UPDATE bundle_listings SET
    name = bundle_versions.content->>'name',
    description = bundle_versions.content->>'description'
FROM bundles
JOIN bundle_versions ON bundle_versions.bundle_id = bundles.id
    AND bundle_versions.version = COALESCE(bundles.approved_version, bundles.version)
WHERE bundles.id = bundle_listings.bundle_id;

ALTER TABLE bundle_listings
    ALTER COLUMN name DROP DEFAULT,
    ALTER COLUMN description DROP DEFAULT;

UPDATE bundle_reviews SET listing = bundle_reviews.listing || jsonb_build_object(
    'name', bundle_versions.content->>'name',
    'description', bundle_versions.content->>'description'
)
FROM bundle_versions
WHERE bundle_versions.bundle_id = bundle_reviews.bundle_id
    AND bundle_versions.version = bundle_reviews.version;

-- Old decided reviews can be of versions that were deleted since.
UPDATE bundle_reviews SET listing = bundle_reviews.listing || jsonb_build_object(
    'name', bundles.name,
    'description', bundles.description
)
FROM bundles
WHERE bundles.id = bundle_reviews.bundle_id
    AND bundle_reviews.listing->>'name' IS NULL;
