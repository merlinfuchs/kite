ALTER TABLE plugin_values ADD COLUMN metadata JSONB;

CREATE INDEX plugin_values_metadata ON plugin_values USING GIN(metadata);
