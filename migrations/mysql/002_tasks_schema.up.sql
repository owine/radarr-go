-- Task Management System Schema for MySQL/MariaDB
-- This migration adds the task management and scheduling system

-- Tasks table for tracking individual task executions
CREATE TABLE IF NOT EXISTS tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    command_name VARCHAR(255) NOT NULL,
    message VARCHAR(1000),
    body TEXT,
    priority ENUM('high', 'normal', 'low') NOT NULL DEFAULT 'normal',
    status ENUM('queued', 'started', 'completed', 'failed', 'aborted', 'cancelling') NOT NULL DEFAULT 'queued',
    queued_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP NULL,
    ended_at TIMESTAMP NULL,
    duration BIGINT NULL,
    exception TEXT,
    `trigger` ENUM('manual', 'scheduled', 'system', 'api') NOT NULL DEFAULT 'manual',
    progress TEXT,
    interval_ms BIGINT NULL,
    last_execution TIMESTAMP NULL,
    next_execution TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Scheduled tasks table for recurring tasks
CREATE TABLE IF NOT EXISTS scheduled_tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    command_name VARCHAR(255) NOT NULL,
    body TEXT,
    interval_ms BIGINT NOT NULL,
    priority ENUM('high', 'normal', 'low') NOT NULL DEFAULT 'normal',
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_run TIMESTAMP NULL,
    next_run TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Task queues table for managing worker pools
CREATE TABLE IF NOT EXISTS task_queues (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    max_workers INT NOT NULL DEFAULT 3,
    active_workers INT NOT NULL DEFAULT 0,
    queued_tasks INT NOT NULL DEFAULT 0,
    running_tasks INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Indexes for performance
CREATE INDEX idx_tasks_command_name ON tasks(command_name);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority ON tasks(priority);
CREATE INDEX idx_tasks_queued_at ON tasks(queued_at);
CREATE INDEX idx_tasks_started_at ON tasks(started_at);
CREATE INDEX idx_tasks_ended_at ON tasks(ended_at);
CREATE INDEX idx_tasks_last_execution ON tasks(last_execution);
CREATE INDEX idx_tasks_next_execution ON tasks(next_execution);

CREATE INDEX idx_scheduled_tasks_next_run ON scheduled_tasks(next_run);
CREATE INDEX idx_scheduled_tasks_enabled ON scheduled_tasks(enabled);

-- Insert default task queues
INSERT IGNORE INTO task_queues (name, max_workers, enabled) VALUES
    ('default', 3, true),
    ('high-priority', 2, true),
    ('background', 1, true);

-- Insert default scheduled tasks
INSERT IGNORE INTO scheduled_tasks (name, command_name, interval_ms, priority, enabled, next_run) VALUES
    ('Health Check', 'HealthCheck', 1800000, 'low', true, DATE_ADD(NOW(), INTERVAL 30 MINUTE)), -- Every 30 minutes
    ('Cleanup Tasks', 'Cleanup', 86400000, 'low', true, DATE_ADD(NOW(), INTERVAL 1 DAY)), -- Daily
    ('Refresh All Movies', 'RefreshAllMovies', 604800000, 'normal', false, DATE_ADD(NOW(), INTERVAL 7 DAY)); -- Weekly (disabled by default)
