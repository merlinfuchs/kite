CREATE TABLE IF NOT EXISTS app_usage (
   id BIGSERIAL PRIMARY KEY,
   app_id TEXT NOT NULL REFERENCES apps(id) ON DELETE CASCADE,

   total_event_count INT NOT NULL,
   success_event_count INT NOT NULL,
   total_event_execution_time BIGINT NOT NULL, -- in microseconds
   avg_event_execution_time BIGINT NOT NULL, -- in microseconds
   total_event_total_time BIGINT NOT NULL, -- in microseconds
   avg_event_total_time BIGINT NOT NULL, -- in microseconds
   total_call_count INT NOT NULL,
   success_call_count INT NOT NULL,
   total_call_total_time BIGINT NOT NULL, -- in microseconds
   avg_call_total_time BIGINT NOT NULL, -- in microseconds
   
   period_starts_at TIMESTAMP NOT NULL,
   period_ends_at TIMESTAMP NOT NULL
);

CREATE VIEW app_usage_month_view AS 
SELECT 
    app_id, 
    SUM(total_event_count) AS total_event_count,
    SUM(success_event_count) AS success_event_count,
    SUM(total_event_execution_time) AS total_event_execution_time,
    AVG(avg_event_execution_time) AS avg_event_execution_time,
    SUM(total_event_total_time) AS total_event_total_time,
    AVG(avg_event_total_time) AS avg_event_total_time,
    SUM(total_call_count) AS total_call_count,
    SUM(success_call_count) AS success_call_count,
    SUM(total_call_total_time) AS total_call_total_time,
    AVG(avg_call_total_time) AS avg_call_total_time
FROM app_usage 
WHERE 
   period_starts_at >= DATE_TRUNC('month', NOW()) AND 
   period_ends_at <= (DATE_TRUNC('month', NOW())  + interval '1 month')
GROUP BY app_id;