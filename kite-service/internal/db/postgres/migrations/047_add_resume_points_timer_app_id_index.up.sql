-- CONCURRENTLY can't run in a transaction, so this file must stay a single statement.
CREATE INDEX CONCURRENTLY IF NOT EXISTS resume_points_timer_app_id ON resume_points (app_id, resume_at) WHERE resume_at IS NOT NULL;
