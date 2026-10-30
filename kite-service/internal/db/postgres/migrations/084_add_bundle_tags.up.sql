-- Tags are words that describe what a bundle does, like roblox or economy,
-- next to its category. People find bundles by them in the marketplace. On the
-- page of a user's bundle and on its listing, reviews have them in their
-- listing.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE bundle_listings ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';
