-- The server and user a flow ran for, so credit limits can be enforced per
-- server and per user. Either is NULL when the execution had none, like a
-- scheduled event listener or a direct message for guild_id.
ALTER TABLE usage_records ADD COLUMN IF NOT EXISTS guild_id TEXT, ADD COLUMN IF NOT EXISTS user_id TEXT;
