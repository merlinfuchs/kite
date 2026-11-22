-- Whether an AI check found the listed version of a user's bundle good enough
-- for the Featured row of the marketplace. NULL while it wasn't checked, which
-- it is again after every approved review.
ALTER TABLE bundles ADD COLUMN IF NOT EXISTS featurable BOOLEAN;
