-- Bundles are what apps install, for now the official templates, which the
-- service stores on startup. A version's content is the bundle as JSON, in the
-- format of internal/core/bundle. Installs point at a bundle and version by
-- bundle_installs.bundle_id and bundle_version.
CREATE TABLE IF NOT EXISTS bundles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    -- NULL for official bundles.
    owner_user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    -- The version apps install.
    version INT NOT NULL,

    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS bundles_owner_user_id ON bundles (owner_user_id);

-- How a bundle is shown, on its page and later in the marketplace. It can
-- change without a new version.
CREATE TABLE IF NOT EXISTS bundle_listings (
    bundle_id TEXT PRIMARY KEY REFERENCES bundles(id) ON DELETE CASCADE,
    icon TEXT NOT NULL,
    category TEXT NOT NULL,
    -- Paragraphs on what the bundle does.
    about JSONB NOT NULL,
    -- Questions and answers, each with question and answer.
    faq JSONB NOT NULL,
    -- The steps outside of the bundle that it needs to work.
    setup JSONB NOT NULL,
    -- Listings are shown in this order.
    position INT NOT NULL,

    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS bundle_versions (
    bundle_id TEXT NOT NULL REFERENCES bundles(id) ON DELETE CASCADE,
    version INT NOT NULL,
    content JSONB NOT NULL,

    created_at TIMESTAMP NOT NULL,

    PRIMARY KEY (bundle_id, version)
);
