CREATE TABLE IF NOT EXISTS deployment_metrics (
   id BIGSERIAL PRIMARY KEY,
   deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
   -- EVENT_HANDLED, CALL_EXECUTED
   type TEXT NOT NULL,
   metadata JSONB,

   -- For EVENT_HANDLED
   event_type TEXT NOT NULL,
   event_success BOOLEAN NOT NULL,
   event_execution_time BIGINT NOT NULL, -- in microseconds
   event_total_time BIGINT NOT NULL, -- in microseconds

   -- For CALL_EXECUTED
   call_type TEXT NOT NULL,
   call_success BOOLEAN NOT NULL,
   call_total_time BIGINT NOT NULL, -- in microseconds
   
   timestamp TIMESTAMP NOT NULL
);
