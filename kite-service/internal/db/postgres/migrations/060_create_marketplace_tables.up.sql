CREATE TABLE IF NOT EXISTS marketplace_listings (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    author_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- No foreign key so listings stay up after the source app is deleted.
    source_app_id TEXT,
    -- pending, approved, rejected or removed. Only approved listings are public.
    status TEXT NOT NULL,
    -- The commands and event listeners, see model.MarketplaceListingItem.
    items JSONB NOT NULL,
    command_count INT NOT NULL,
    event_listener_count INT NOT NULL,
    -- Distinct block types used by the items, so risky blocks can be shown
    -- without loading the items.
    block_types TEXT[] NOT NULL,
    import_count INT NOT NULL DEFAULT 0,

    review_note TEXT,
    reviewed_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS marketplace_listings_status_import_count
    ON marketplace_listings (status, import_count DESC);
CREATE INDEX IF NOT EXISTS marketplace_listings_author_user_id
    ON marketplace_listings (author_user_id);

CREATE TABLE IF NOT EXISTS marketplace_reports (
    id TEXT PRIMARY KEY,
    listing_id TEXT NOT NULL REFERENCES marketplace_listings(id) ON DELETE CASCADE,
    reporter_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    resolved_at TIMESTAMP,
    resolved_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL
);

-- A user can only have one open report per listing.
CREATE UNIQUE INDEX IF NOT EXISTS marketplace_reports_open_listing_id_reporter_user_id
    ON marketplace_reports (listing_id, reporter_user_id) WHERE resolved_at IS NULL;

-- Moderators are keyed by Discord user ID so they can be added before they
-- have ever logged in to Kite.
CREATE TABLE IF NOT EXISTS marketplace_moderators (
    discord_user_id TEXT PRIMARY KEY,
    added_by_user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL
);
