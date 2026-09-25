DROP INDEX IF EXISTS resume_points_resume_at;
ALTER TABLE resume_points DROP COLUMN interaction_token;
ALTER TABLE resume_points DROP COLUMN resume_at;
