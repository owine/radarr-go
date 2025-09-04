-- PostgreSQL initialization script for development
-- This script runs when the PostgreSQL container starts for the first time

-- Enable useful extensions for development
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create additional schemas if needed
-- CREATE SCHEMA IF NOT EXISTS radarr_analytics;

-- Set up logging and monitoring
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_min_duration_statement = 100;
ALTER SYSTEM SET log_connections = 'on';
ALTER SYSTEM SET log_disconnections = 'on';

-- Performance settings for development
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET track_activity_query_size = 2048;
ALTER SYSTEM SET pg_stat_statements.track = 'all';

-- Reload configuration
SELECT pg_reload_conf();

-- Create development-specific tables or data if needed
-- This will be managed by the application's migration system

-- Print success message
DO $$
BEGIN
    RAISE NOTICE 'PostgreSQL development database initialized successfully';
END $$;
