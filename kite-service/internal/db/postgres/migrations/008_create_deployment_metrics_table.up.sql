CREATE TABLE IF NOT EXISTS deployment_metrics (
   id BIGSERIAL PRIMARY KEY,
   deployment_id TEXT NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
   -- EVENT_RECEIVED, EVENT_HANDLED, CALL_EXECUTED
   type TEXT NOT NULL,
   metadata JSONB,

   -- For EVENT_RECEIVED & EVENT_HANDLED
   event_id BIGINT NOT NULL,
   event_type TEXT NOT NULL,

   -- For EVENT_HANDLED
   event_success BOOLEAN NOT NULL,
   event_execution_time INT NOT NULL,
   event_total_time INT NOT NULL,

   -- For CALL_EXECUTED
   call_type TEXT NOT NULL,
   call_success BOOLEAN NOT NULL,
   call_total_time INT NOT NULL,
   
   timestamp TIMESTAMP NOT NULL
);
