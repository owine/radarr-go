-- Rollback Task Management System Schema for PostgreSQL
-- This migration removes the task management and scheduling system

-- Drop indexes first
DROP INDEX IF EXISTS idx_tasks_command_name;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_priority;
DROP INDEX IF EXISTS idx_tasks_queued_at;
DROP INDEX IF EXISTS idx_tasks_started_at;
DROP INDEX IF EXISTS idx_tasks_ended_at;
DROP INDEX IF EXISTS idx_tasks_last_execution;
DROP INDEX IF EXISTS idx_tasks_next_execution;
DROP INDEX IF EXISTS idx_scheduled_tasks_next_run;
DROP INDEX IF EXISTS idx_scheduled_tasks_enabled;

-- Drop tables
DROP TABLE IF EXISTS task_queues;
DROP TABLE IF EXISTS scheduled_tasks;
DROP TABLE IF EXISTS tasks;
