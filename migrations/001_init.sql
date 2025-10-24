CREATE TABLE IF NOT EXISTS servers (
                                       id SERIAL PRIMARY KEY,
                                       name TEXT NOT NULL,
                                       ip TEXT,
                                       description TEXT,
                                       created_at TIMESTAMP DEFAULT NOW()
    );

CREATE TABLE IF NOT EXISTS metrics (
                                       id SERIAL PRIMARY KEY,
                                       object_type TEXT NOT NULL,   -- 'server', 'database', 'service'
                                       object_id INT REFERENCES servers(id) ON DELETE CASCADE,               -- optionally references servers.id etc.
                                       metric_name TEXT NOT NULL,
                                       value DOUBLE PRECISION NOT NULL,
                                       timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- optional index for faster queries by metric and time
CREATE INDEX IF NOT EXISTS idx_metrics_metric_time ON metrics(metric_name, timestamp DESC);