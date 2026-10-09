-- Listings can contain message templates too.
ALTER TABLE marketplace_listings ADD COLUMN IF NOT EXISTS message_count INT NOT NULL DEFAULT 0;
