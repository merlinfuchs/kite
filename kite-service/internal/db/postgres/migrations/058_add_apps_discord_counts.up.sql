-- Discord's approximate counts of the servers and users that installed the
-- app, refreshed by the gateway about once a day. NULL until first fetched.
ALTER TABLE apps ADD COLUMN discord_guild_count INTEGER;
ALTER TABLE apps ADD COLUMN discord_user_install_count INTEGER;
ALTER TABLE apps ADD COLUMN discord_counts_updated_at TIMESTAMP;
