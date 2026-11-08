UPDATE bundle_reviews SET listing = listing - 'name' - 'description';

ALTER TABLE bundle_listings
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS description;

ALTER TABLE bundles
    DROP COLUMN IF EXISTS icon,
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS about,
    DROP COLUMN IF EXISTS setup;
