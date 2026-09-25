ALTER TABLE resume_points ADD COLUMN resume_at TIMESTAMP;
-- Encrypted, only set for timers that resume an interaction flow.
ALTER TABLE resume_points ADD COLUMN interaction_token TEXT;

-- Not CONCURRENTLY, which can't run in this transaction. Building it still
-- scans the table and blocks writes to resume_points while it does.
CREATE INDEX IF NOT EXISTS resume_points_resume_at ON resume_points (resume_at) WHERE resume_at IS NOT NULL;
