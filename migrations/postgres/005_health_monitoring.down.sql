-- Rollback health monitoring system tables for PostgreSQL

-- Drop triggers first
DROP TRIGGER IF EXISTS update_health_issues_updated_at ON health_issues;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_health_issues_type;
DROP INDEX IF EXISTS idx_health_issues_severity;
DROP INDEX IF EXISTS idx_health_issues_source;
DROP INDEX IF EXISTS idx_health_issues_resolved;
DROP INDEX IF EXISTS idx_health_issues_dismissed;
DROP INDEX IF EXISTS idx_health_issues_created_at;
DROP INDEX IF EXISTS idx_health_issues_resolved_at;

DROP INDEX IF EXISTS idx_performance_metrics_timestamp;
DROP INDEX IF EXISTS idx_performance_metrics_created_at;

-- Drop tables
DROP TABLE IF EXISTS performance_metrics;
DROP TABLE IF EXISTS health_issues;
