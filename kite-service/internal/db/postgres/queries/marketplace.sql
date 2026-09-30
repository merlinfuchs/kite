-- name: CreateMarketplaceListing :one
INSERT INTO marketplace_listings (
    id,
    name,
    description,
    author_user_id,
    source_app_id,
    status,
    items,
    command_count,
    event_listener_count,
    block_types,
    created_at,
    updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING *;

-- name: UpdateMarketplaceListing :one
-- Any change sends the listing back to the moderation queue.
UPDATE marketplace_listings SET
    name = $2,
    description = $3,
    source_app_id = $4,
    items = $5,
    command_count = $6,
    event_listener_count = $7,
    block_types = $8,
    status = $9,
    review_note = NULL,
    reviewed_by_user_id = NULL,
    reviewed_at = NULL,
    updated_at = $10
WHERE id = $1 RETURNING *;

-- name: ReviewMarketplaceListing :one
UPDATE marketplace_listings SET
    status = @status,
    review_note = @review_note,
    reviewed_by_user_id = @reviewed_by_user_id,
    reviewed_at = @reviewed_at
WHERE id = @id RETURNING *;

-- name: IncrementMarketplaceListingImportCount :exec
UPDATE marketplace_listings SET import_count = import_count + 1 WHERE id = $1;

-- name: DeleteMarketplaceListing :exec
DELETE FROM marketplace_listings WHERE id = $1;

-- name: MarketplaceListing :one
SELECT sqlc.embed(marketplace_listings), sqlc.embed(users)
FROM marketplace_listings
JOIN users ON users.id = marketplace_listings.author_user_id
WHERE marketplace_listings.id = $1;

-- name: MarketplaceListings :many
-- Search matches the name and description, the sort is by imports or by age.
SELECT sqlc.embed(marketplace_listings), sqlc.embed(users)
FROM marketplace_listings
JOIN users ON users.id = marketplace_listings.author_user_id
WHERE marketplace_listings.status = @status
    AND (
        sqlc.narg('search')::TEXT IS NULL
        OR marketplace_listings.name ILIKE '%' || sqlc.narg('search')::TEXT || '%'
        OR marketplace_listings.description ILIKE '%' || sqlc.narg('search')::TEXT || '%'
    )
    AND (
        @kind::TEXT = ''
        OR (@kind::TEXT = 'command' AND marketplace_listings.command_count > 0 AND marketplace_listings.event_listener_count = 0)
        OR (@kind::TEXT = 'event_listener' AND marketplace_listings.event_listener_count > 0 AND marketplace_listings.command_count = 0)
        OR (@kind::TEXT = 'module' AND marketplace_listings.command_count + marketplace_listings.event_listener_count > 1)
    )
ORDER BY
    CASE WHEN @sort::TEXT = 'popular' THEN marketplace_listings.import_count END DESC,
    marketplace_listings.updated_at DESC
LIMIT @page_limit OFFSET @page_offset;

-- name: MarketplaceListingsByAuthor :many
SELECT sqlc.embed(marketplace_listings), sqlc.embed(users)
FROM marketplace_listings
JOIN users ON users.id = marketplace_listings.author_user_id
WHERE marketplace_listings.author_user_id = $1
ORDER BY marketplace_listings.updated_at DESC;

-- name: CountMarketplaceListingsByAuthor :one
SELECT COUNT(*) FROM marketplace_listings WHERE author_user_id = $1;

-- name: CreateMarketplaceReport :one
INSERT INTO marketplace_reports (
    id,
    listing_id,
    reporter_user_id,
    reason,
    created_at
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (listing_id, reporter_user_id) WHERE resolved_at IS NULL DO UPDATE SET reason = EXCLUDED.reason
RETURNING *;

-- name: CountOpenMarketplaceReportsByListing :one
SELECT COUNT(*) FROM marketplace_reports WHERE listing_id = $1 AND resolved_at IS NULL;

-- name: OpenMarketplaceReports :many
SELECT sqlc.embed(marketplace_reports), sqlc.embed(users)
FROM marketplace_reports
JOIN users ON users.id = marketplace_reports.reporter_user_id
WHERE marketplace_reports.resolved_at IS NULL
ORDER BY marketplace_reports.created_at ASC
LIMIT 100;

-- name: ResolveMarketplaceReport :exec
UPDATE marketplace_reports SET
    resolved_at = @resolved_at,
    resolved_by_user_id = @resolved_by_user_id
WHERE id = @id AND resolved_at IS NULL;

-- name: ResolveMarketplaceReportsByListing :exec
UPDATE marketplace_reports SET
    resolved_at = @resolved_at,
    resolved_by_user_id = @resolved_by_user_id
WHERE listing_id = @listing_id AND resolved_at IS NULL;

-- name: MarketplaceModerators :many
SELECT * FROM marketplace_moderators ORDER BY created_at ASC;

-- name: MarketplaceModerator :one
SELECT * FROM marketplace_moderators WHERE discord_user_id = $1;

-- name: CreateMarketplaceModerator :one
INSERT INTO marketplace_moderators (
    discord_user_id,
    added_by_user_id,
    created_at
) VALUES ($1, $2, $3)
ON CONFLICT (discord_user_id) DO UPDATE SET discord_user_id = EXCLUDED.discord_user_id
RETURNING *;

-- name: DeleteMarketplaceModerator :exec
DELETE FROM marketplace_moderators WHERE discord_user_id = $1;
