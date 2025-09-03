-- Migration 010: Fix tasks table schema to match Go models
-- This migration adds missing columns that were manually added to the database

-- Add missing columns to tasks table if they don't exist
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS result TEXT DEFAULT NULL;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS error_message TEXT DEFAULT '';

-- Update duration column to duration_ms if it doesn't exist
-- (Keep both for backward compatibility, but prefer duration_ms)
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS duration_ms BIGINT DEFAULT NULL;

-- Update existing duration column data to duration_ms if duration_ms is null
UPDATE tasks SET duration_ms = duration WHERE duration_ms IS NULL AND duration IS NOT NULL;

-- Convert body column to JSON type if it's currently TEXT
-- First check if it's already JSON type
DO $$
BEGIN
    -- Check if body column is TEXT and convert to JSON
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'tasks'
        AND column_name = 'body'
        AND data_type = 'text'
    ) THEN
        -- Convert TEXT to JSON, handling invalid JSON
        UPDATE tasks SET body = '{}' WHERE body IS NULL OR body = '';
        ALTER TABLE tasks ALTER COLUMN body TYPE JSON USING body::JSON;
    END IF;

    -- Ensure body has proper default
    ALTER TABLE tasks ALTER COLUMN body SET DEFAULT '{}'::JSON;
END $$;

-- Add indexes for new columns if they don't exist
CREATE INDEX IF NOT EXISTS idx_tasks_result ON tasks(result) WHERE result IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_error_message ON tasks(error_message) WHERE error_message != '';
CREATE INDEX IF NOT EXISTS idx_tasks_duration_ms ON tasks(duration_ms) WHERE duration_ms IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN tasks.result IS 'JSON result data from task execution';
COMMENT ON COLUMN tasks.error_message IS 'Error message if task failed';
COMMENT ON COLUMN tasks.duration_ms IS 'Task execution duration in milliseconds';
