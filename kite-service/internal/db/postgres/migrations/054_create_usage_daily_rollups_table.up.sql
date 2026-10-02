-- usage_records only keeps 40 days. Before the cleanup deletes a day, it is
-- summed into this table so analytics can still show longer ranges.
CREATE TABLE IF NOT EXISTS usage_daily_rollups (
    date DATE NOT NULL,
    app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    -- The command, event listener or message ID, empty if the record had none.
    -- Not a foreign key so stats stay after the source is deleted.
    source_id TEXT NOT NULL DEFAULT '',

    executions BIGINT NOT NULL,
    credits_used BIGINT NOT NULL,

    -- The leading date serves MAX(date), which marks how far the rollup got.
    PRIMARY KEY (date, app_id, type, source_id)
);

-- Serves the analytics queries and the ON DELETE CASCADE from apps.
CREATE INDEX IF NOT EXISTS usage_daily_rollups_app_id_date ON usage_daily_rollups (app_id, date);
