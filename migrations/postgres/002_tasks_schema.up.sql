-- Task Management System Schema for PostgreSQL
-- This migration adds the task management and scheduling system

-- Tasks table for tracking individual task executions
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    command_name VARCHAR(255) NOT NULL,
    message VARCHAR(1000),
    body TEXT DEFAULT '{}'::TEXT,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    queued_at TIMESTAMP NOT NULL DEFAULT NOW(),
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    duration BIGINT,
    exception TEXT,
    trigger VARCHAR(20) NOT NULL DEFAULT 'manual',
    progress TEXT DEFAULT '{}'::TEXT,
    interval_ms BIGINT,
    last_execution TIMESTAMP,
    next_execution TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Scheduled tasks table for recurring tasks
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    command_name VARCHAR(255) NOT NULL,
    body TEXT DEFAULT '{}'::TEXT,
    interval_ms BIGINT NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run TIMESTAMP,
    next_run TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Task queues table for managing worker pools
CREATE TABLE IF NOT EXISTS task_queues (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    max_workers INTEGER NOT NULL DEFAULT 3,
    active_workers INTEGER NOT NULL DEFAULT 0,
    queued_tasks INTEGER NOT NULL DEFAULT 0,
    running_tasks INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_tasks_command_name ON tasks(command_name);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);
CREATE INDEX IF NOT EXISTS idx_tasks_queued_at ON tasks(queued_at);
CREATE INDEX IF NOT EXISTS idx_tasks_started_at ON tasks(started_at);
CREATE INDEX IF NOT EXISTS idx_tasks_ended_at ON tasks(ended_at);
CREATE INDEX IF NOT EXISTS idx_tasks_last_execution ON tasks(last_execution);
CREATE INDEX IF NOT EXISTS idx_tasks_next_execution ON tasks(next_execution);

CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_next_run ON scheduled_tasks(next_run);
CREATE INDEX IF NOT EXISTS idx_scheduled_tasks_enabled ON scheduled_tasks(enabled);

-- Check constraints to ensure valid enum values
ALTER TABLE tasks ADD CONSTRAINT check_tasks_priority
    CHECK (priority IN ('high', 'normal', 'low'));

ALTER TABLE tasks ADD CONSTRAINT check_tasks_status
    CHECK (status IN ('queued', 'started', 'completed', 'failed', 'aborted', 'cancelling'));

ALTER TABLE tasks ADD CONSTRAINT check_tasks_trigger
    CHECK (trigger IN ('manual', 'scheduled', 'system', 'api'));

ALTER TABLE scheduled_tasks ADD CONSTRAINT check_scheduled_tasks_priority
    CHECK (priority IN ('high', 'normal', 'low'));

-- Insert default task queues
INSERT INTO task_queues (name, max_workers, enabled) VALUES
    ('default', 3, true),
    ('high-priority', 2, true),
    ('background', 1, true)
ON CONFLICT (name) DO NOTHING;

-- Insert default scheduled tasks
INSERT INTO scheduled_tasks (name, command_name, interval_ms, priority, enabled, next_run) VALUES
    ('Health Check', 'HealthCheck', 1800000, 'low', true, NOW() + INTERVAL '30 minutes'), -- Every 30 minutes
    ('Cleanup Tasks', 'Cleanup', 86400000, 'low', true, NOW() + INTERVAL '1 day'), -- Daily
    ('Refresh All Movies', 'RefreshAllMovies', 604800000, 'normal', false, NOW() + INTERVAL '7 days') -- Weekly (disabled by default)
ON CONFLICT (name) DO NOTHING;
