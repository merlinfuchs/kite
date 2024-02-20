CREATE TABLE IF NOT EXISTS guild_usage (
   id BIGSERIAL PRIMARY KEY,
   guild_id TEXT NOT NULL REFERENCES guilds(id) ON DELETE CASCADE,

   total_event_execution_time BIGINT NOT NULL, -- in microseconds
   avg_event_execution_time BIGINT NOT NULL, -- in microseconds
   total_event_total_time BIGINT NOT NULL, -- in microseconds
   avg_event_total_time BIGINT NOT NULL, -- in microseconds
   total_call_total_time BIGINT NOT NULL, -- in microseconds
   avg_call_total_time BIGINT NOT NULL, -- in microseconds
   
   period_starts_at TIMESTAMP NOT NULL,
   period_ends_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL
);
