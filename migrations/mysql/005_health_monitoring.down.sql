-- Rollback health monitoring system tables for MySQL/MariaDB

-- Drop tables
DROP TABLE IF EXISTS performance_metrics;
DROP TABLE IF EXISTS health_issues;
