-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS resume_points_expires_at ON resume_points (expires_at) WHERE expires_at IS NOT NULL;
