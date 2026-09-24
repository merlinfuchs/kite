-- The gateway manager polls these tables by updated_at to notice when an app's
-- intent requirements change, so both need an index on it.
CREATE INDEX IF NOT EXISTS event_listeners_updated_at ON event_listeners (updated_at);
CREATE INDEX IF NOT EXISTS plugin_instances_updated_at ON plugin_instances (updated_at);
