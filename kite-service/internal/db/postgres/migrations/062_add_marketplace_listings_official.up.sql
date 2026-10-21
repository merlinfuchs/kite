-- Moderators can mark community listings as official, so they show on the
-- Official tab next to the Kite team's templates.
ALTER TABLE marketplace_listings ADD COLUMN IF NOT EXISTS official BOOLEAN NOT NULL DEFAULT FALSE;
