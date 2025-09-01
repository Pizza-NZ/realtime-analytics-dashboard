CREATE TABLE IF NOT EXISTS analytics_events (
    event_type varchar(255),
    user_id bigint,
    created_at timestamptz NOT NULL,
    event_data jsonb,
    event_id UUID DEFAULT gen_random_UUID()
);

ALTER TABLE analytics_events 
ADD PRIMARY KEY (event_id, created_at);

SELECT create_hypertable('analytics_events', 'created_at');