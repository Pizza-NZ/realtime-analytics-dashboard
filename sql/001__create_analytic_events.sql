CREATE TABLE IF NOT EXISTS analytics_events (
    EventType varchar(255),
    UserID bigint,
    Created_at timestamptz NOT NULL,
    EventData jsonb
);