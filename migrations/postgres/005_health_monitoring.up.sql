-- Health monitoring system tables for PostgreSQL

-- Health issues table
CREATE TABLE IF NOT EXISTS health_issues (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    source VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    message TEXT NOT NULL,
    details JSONB,
    wiki_url VARCHAR(255),
    first_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    is_resolved BOOLEAN DEFAULT false,
    is_dismissed BOOLEAN DEFAULT false,
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Performance metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id SERIAL PRIMARY KEY,
    cpu_usage_percent FLOAT NOT NULL DEFAULT 0,
    memory_usage_mb FLOAT NOT NULL DEFAULT 0,
    memory_total_mb FLOAT NOT NULL DEFAULT 0,
    disk_usage_percent FLOAT NOT NULL DEFAULT 0,
    disk_available_gb FLOAT NOT NULL DEFAULT 0,
    disk_total_gb FLOAT NOT NULL DEFAULT 0,
    database_latency_ms FLOAT NOT NULL DEFAULT 0,
    api_latency_ms FLOAT NOT NULL DEFAULT 0,
    active_connections INTEGER NOT NULL DEFAULT 0,
    queue_size INTEGER NOT NULL DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_health_issues_type ON health_issues(type);
CREATE INDEX IF NOT EXISTS idx_health_issues_severity ON health_issues(severity);
CREATE INDEX IF NOT EXISTS idx_health_issues_source ON health_issues(source);
CREATE INDEX IF NOT EXISTS idx_health_issues_resolved ON health_issues(is_resolved);
CREATE INDEX IF NOT EXISTS idx_health_issues_dismissed ON health_issues(is_dismissed);
CREATE INDEX IF NOT EXISTS idx_health_issues_created_at ON health_issues(created_at);
CREATE INDEX IF NOT EXISTS idx_health_issues_resolved_at ON health_issues(resolved_at);

CREATE INDEX IF NOT EXISTS idx_performance_metrics_timestamp ON performance_metrics(timestamp);
CREATE INDEX IF NOT EXISTS idx_performance_metrics_created_at ON performance_metrics(created_at);

-- Function to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at for health_issues
DROP TRIGGER IF EXISTS update_health_issues_updated_at ON health_issues;
CREATE TRIGGER update_health_issues_updated_at
    BEFORE UPDATE ON health_issues
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE health_issues IS 'Stores health issues detected by the health monitoring system';
COMMENT ON TABLE performance_metrics IS 'Stores system performance metrics collected over time';

COMMENT ON COLUMN health_issues.type IS 'Type of health check (database, diskSpace, system, etc.)';
COMMENT ON COLUMN health_issues.source IS 'Source component that detected the issue';
COMMENT ON COLUMN health_issues.severity IS 'Severity level: info, warning, error, critical';
COMMENT ON COLUMN health_issues.message IS 'Human-readable description of the issue';
COMMENT ON COLUMN health_issues.details IS 'Additional structured details about the issue';
COMMENT ON COLUMN health_issues.wiki_url IS 'URL to documentation for resolving the issue';
COMMENT ON COLUMN health_issues.first_seen IS 'When this issue was first detected';
COMMENT ON COLUMN health_issues.last_seen IS 'When this issue was last detected';
COMMENT ON COLUMN health_issues.resolved_at IS 'When this issue was resolved';
COMMENT ON COLUMN health_issues.is_resolved IS 'Whether the issue has been resolved';
COMMENT ON COLUMN health_issues.is_dismissed IS 'Whether the issue has been dismissed by user';
COMMENT ON COLUMN health_issues.data IS 'Additional contextual data';

COMMENT ON COLUMN performance_metrics.cpu_usage_percent IS 'CPU usage as percentage';
COMMENT ON COLUMN performance_metrics.memory_usage_mb IS 'Memory usage in megabytes';
COMMENT ON COLUMN performance_metrics.memory_total_mb IS 'Total memory in megabytes';
COMMENT ON COLUMN performance_metrics.disk_usage_percent IS 'Disk usage as percentage';
COMMENT ON COLUMN performance_metrics.disk_available_gb IS 'Available disk space in gigabytes';
COMMENT ON COLUMN performance_metrics.disk_total_gb IS 'Total disk space in gigabytes';
COMMENT ON COLUMN performance_metrics.database_latency_ms IS 'Database query latency in milliseconds';
COMMENT ON COLUMN performance_metrics.api_latency_ms IS 'API response latency in milliseconds';
COMMENT ON COLUMN performance_metrics.active_connections IS 'Number of active database connections';
COMMENT ON COLUMN performance_metrics.queue_size IS 'Current queue size';
COMMENT ON COLUMN performance_metrics.timestamp IS 'When these metrics were collected';
