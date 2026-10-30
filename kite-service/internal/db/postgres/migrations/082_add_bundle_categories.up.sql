-- A bundle can be in up to 3 categories of the marketplace, the first is its
-- main one, which category still has. On the page of a user's bundle and on its
-- listing, reviews have them in their listing.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS categories TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE bundle_listings ADD COLUMN IF NOT EXISTS categories TEXT[] NOT NULL DEFAULT '{}';

UPDATE bundles SET categories = ARRAY[category] WHERE category <> '';
UPDATE bundle_listings SET categories = ARRAY[category] WHERE category <> '';
