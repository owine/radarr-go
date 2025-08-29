-- Rollback Task Management System Schema for MySQL/MariaDB
-- This migration removes the task management and scheduling system

-- Drop indexes first
DROP INDEX IF EXISTS idx_tasks_command_name ON tasks;
DROP INDEX IF EXISTS idx_tasks_status ON tasks;
DROP INDEX IF EXISTS idx_tasks_priority ON tasks;
DROP INDEX IF EXISTS idx_tasks_queued_at ON tasks;
DROP INDEX IF EXISTS idx_tasks_started_at ON tasks;
DROP INDEX IF EXISTS idx_tasks_ended_at ON tasks;
DROP INDEX IF EXISTS idx_tasks_last_execution ON tasks;
DROP INDEX IF EXISTS idx_tasks_next_execution ON tasks;
DROP INDEX IF EXISTS idx_scheduled_tasks_next_run ON scheduled_tasks;
DROP INDEX IF EXISTS idx_scheduled_tasks_enabled ON scheduled_tasks;

-- Drop tables
DROP TABLE IF EXISTS task_queues;
DROP TABLE IF EXISTS scheduled_tasks;
DROP TABLE IF EXISTS tasks;
