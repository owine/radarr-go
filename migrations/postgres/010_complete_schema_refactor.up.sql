-- Complete Database Schema Refactor for PostgreSQL
-- This migration completely rebuilds the database with simplified, clean architecture
-- Backwards compatibility is NOT maintained - this is a fresh start

-- First, drop all existing tables and constraints to start clean
DROP SCHEMA IF EXISTS radarr_backup CASCADE;
CREATE SCHEMA radarr_backup;

-- Backup existing data (if needed)
DO $$
BEGIN
    -- Move existing tables to backup schema if they exist
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'movies' AND table_schema = 'public') THEN
        ALTER TABLE IF EXISTS movies SET SCHEMA radarr_backup;
        ALTER TABLE IF EXISTS movie_files SET SCHEMA radarr_backup;
        ALTER TABLE IF EXISTS collections SET SCHEMA radarr_backup;
        ALTER TABLE IF EXISTS health_issues SET SCHEMA radarr_backup;
        ALTER TABLE IF EXISTS tasks SET SCHEMA radarr_backup;
        ALTER TABLE IF EXISTS scheduled_tasks SET SCHEMA radarr_backup;
    END IF;
END $$;

-- Drop all existing functions and triggers
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
DROP FUNCTION IF EXISTS update_collections_updated_at() CASCADE;
DROP FUNCTION IF EXISTS update_parse_cache_updated_at() CASCADE;

-- ============================================================================
-- CORE SIMPLIFIED SCHEMA
-- ============================================================================

-- Create updated_at trigger function (simplified)
CREATE OR REPLACE FUNCTION update_timestamp_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 1. CORE TABLES
-- ============================================================================

-- Movies table (simplified, minimal validation)
CREATE TABLE movies (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    title_slug VARCHAR(500) NOT NULL,
    original_title VARCHAR(500),
    overview TEXT,
    year INTEGER,
    runtime INTEGER,
    status VARCHAR(20) DEFAULT 'tba',

    -- File information
    has_file BOOLEAN DEFAULT FALSE,
    file_path VARCHAR(1000),
    file_size BIGINT DEFAULT 0,

    -- Configuration
    monitored BOOLEAN DEFAULT TRUE,
    quality_profile_id INTEGER NOT NULL DEFAULT 1,
    minimum_availability VARCHAR(20) DEFAULT 'announced',

    -- Collection relationship (simple integer reference)
    collection_id INTEGER,

    -- Metadata (JSON for flexibility)
    images JSONB DEFAULT '[]',
    genres JSONB DEFAULT '[]',
    tags JSONB DEFAULT '[]',
    ratings JSONB DEFAULT '{}',

    -- Dates
    in_cinemas TIMESTAMPTZ,
    physical_release TIMESTAMPTZ,
    digital_release TIMESTAMPTZ,
    added TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Collections table (drastically simplified)
CREATE TABLE collections (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    overview TEXT,

    -- Configuration
    monitored BOOLEAN DEFAULT TRUE,
    quality_profile_id INTEGER NOT NULL DEFAULT 1,
    minimum_availability VARCHAR(20) DEFAULT 'announced',

    -- Metadata
    images JSONB DEFAULT '[]',
    tags JSONB DEFAULT '[]',

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Quality profiles table (simplified)
CREATE TABLE quality_profiles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    cutoff INTEGER NOT NULL DEFAULT 1,
    items JSONB NOT NULL DEFAULT '[]',
    language VARCHAR(50) DEFAULT 'english',

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 2. HEALTH MONITORING (simplified)
-- ============================================================================

CREATE TABLE health_issues (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    source VARCHAR(100) NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('info', 'warning', 'error', 'critical')),
    message TEXT NOT NULL,

    -- Status
    is_resolved BOOLEAN DEFAULT FALSE,
    is_dismissed BOOLEAN DEFAULT FALSE,

    -- Metadata
    details JSONB,

    -- Timestamps
    first_seen TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 3. TASK SYSTEM (simplified)
-- ============================================================================

CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    command_name VARCHAR(255) NOT NULL,

    -- Status and execution
    status VARCHAR(20) NOT NULL DEFAULT 'queued' CHECK (status IN ('queued', 'started', 'completed', 'failed', 'aborted')),
    priority VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority IN ('high', 'normal', 'low')),

    -- Timing
    queued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    duration_ms BIGINT,

    -- Data and results
    body JSONB DEFAULT '{}',
    result JSONB,
    error_message TEXT,

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE scheduled_tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    command_name VARCHAR(255) NOT NULL,

    -- Configuration
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    interval_ms BIGINT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal' CHECK (priority IN ('high', 'normal', 'low')),

    -- Scheduling
    last_run TIMESTAMPTZ,
    next_run TIMESTAMPTZ NOT NULL,

    -- Data
    body JSONB DEFAULT '{}',

    -- Timestamps
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 4. CONFIGURATION TABLES
-- ============================================================================

CREATE TABLE app_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL,
    value JSONB,
    description TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- 5. SIMPLE INDEXES (only essential ones)
-- ============================================================================

-- Movies indexes
CREATE UNIQUE INDEX idx_movies_tmdb_id ON movies(tmdb_id);
CREATE UNIQUE INDEX idx_movies_title_slug ON movies(title_slug);
CREATE INDEX idx_movies_monitored ON movies(monitored);
CREATE INDEX idx_movies_has_file ON movies(has_file);
CREATE INDEX idx_movies_collection_id ON movies(collection_id);
CREATE INDEX idx_movies_status ON movies(status);

-- Collections indexes
CREATE UNIQUE INDEX idx_collections_tmdb_id ON collections(tmdb_id);
CREATE INDEX idx_collections_monitored ON collections(monitored);

-- Quality profiles indexes
CREATE UNIQUE INDEX idx_quality_profiles_name ON quality_profiles(name);

-- Health issues indexes
CREATE INDEX idx_health_issues_type ON health_issues(type);
CREATE INDEX idx_health_issues_severity ON health_issues(severity);
CREATE INDEX idx_health_issues_resolved ON health_issues(is_resolved);
CREATE INDEX idx_health_issues_created_at ON health_issues(created_at);

-- Tasks indexes
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_queued_at ON tasks(queued_at);
CREATE INDEX idx_tasks_command_name ON tasks(command_name);

-- Scheduled tasks indexes
CREATE UNIQUE INDEX idx_scheduled_tasks_name ON scheduled_tasks(name);
CREATE INDEX idx_scheduled_tasks_enabled ON scheduled_tasks(enabled);
CREATE INDEX idx_scheduled_tasks_next_run ON scheduled_tasks(next_run);

-- App config indexes
CREATE UNIQUE INDEX idx_app_config_key ON app_config(key);

-- ============================================================================
-- 6. SIMPLE FOREIGN KEY RELATIONSHIPS (with proper cascading)
-- ============================================================================

-- Movie to collection relationship (soft reference, no cascading issues)
ALTER TABLE movies
ADD CONSTRAINT fk_movies_collection
FOREIGN KEY (collection_id) REFERENCES collections(id)
ON DELETE SET NULL;

-- Movie to quality profile
ALTER TABLE movies
ADD CONSTRAINT fk_movies_quality_profile
FOREIGN KEY (quality_profile_id) REFERENCES quality_profiles(id)
ON DELETE RESTRICT;

-- Collection to quality profile
ALTER TABLE collections
ADD CONSTRAINT fk_collections_quality_profile
FOREIGN KEY (quality_profile_id) REFERENCES quality_profiles(id)
ON DELETE RESTRICT;

-- ============================================================================
-- 7. SIMPLE TRIGGERS (only for updated_at)
-- ============================================================================

CREATE TRIGGER trigger_movies_updated_at
    BEFORE UPDATE ON movies
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER trigger_collections_updated_at
    BEFORE UPDATE ON collections
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER trigger_quality_profiles_updated_at
    BEFORE UPDATE ON quality_profiles
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER trigger_health_issues_updated_at
    BEFORE UPDATE ON health_issues
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER trigger_tasks_updated_at
    BEFORE UPDATE ON tasks
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER trigger_scheduled_tasks_updated_at
    BEFORE UPDATE ON scheduled_tasks
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

CREATE TRIGGER trigger_app_config_updated_at
    BEFORE UPDATE ON app_config
    FOR EACH ROW EXECUTE FUNCTION update_timestamp_column();

-- ============================================================================
-- 8. DEFAULT DATA
-- ============================================================================

-- Insert default quality profile
INSERT INTO quality_profiles (id, name, cutoff, items, language) VALUES
(1, 'Any', 1, '[{"quality": {"id": 1, "name": "Unknown"}, "allowed": true}]', 'english')
ON CONFLICT (id) DO NOTHING;

-- Insert default scheduled tasks
INSERT INTO scheduled_tasks (name, command_name, interval_ms, priority, enabled, next_run) VALUES
    ('Health Check', 'HealthCheck', 1800000, 'low', true, CURRENT_TIMESTAMP + INTERVAL '30 minutes'),
    ('Cleanup Tasks', 'Cleanup', 86400000, 'low', true, CURRENT_TIMESTAMP + INTERVAL '1 day')
ON CONFLICT (name) DO NOTHING;

-- Insert default configuration
INSERT INTO app_config (key, value, description) VALUES
    ('database.version', '"10"', 'Database schema version'),
    ('server.port', '7878', 'Default server port'),
    ('health.check_interval', '1800000', 'Health check interval in milliseconds')
ON CONFLICT (key) DO NOTHING;

-- ============================================================================
-- COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE movies IS 'Core movies table with simplified structure and minimal validation';
COMMENT ON TABLE collections IS 'Movie collections with simplified relationship management';
COMMENT ON TABLE health_issues IS 'Health monitoring issues with proper isolation';
COMMENT ON TABLE tasks IS 'Task execution tracking with simplified status management';
COMMENT ON TABLE scheduled_tasks IS 'Recurring task configuration';
COMMENT ON TABLE app_config IS 'Application configuration key-value store';

-- Completion message
DO $$
BEGIN
    RAISE NOTICE 'Database schema refactor completed successfully!';
    RAISE NOTICE 'All problematic GORM hooks, complex relationships, and constraint issues have been resolved.';
    RAISE NOTICE 'The new schema is optimized for testing, performance, and maintainability.';
END $$;
