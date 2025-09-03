-- Migration 010 Down: Rollback tasks table schema fixes

-- Remove indexes added in migration 010
DROP INDEX IF EXISTS idx_tasks_result;
DROP INDEX IF EXISTS idx_tasks_error_message;
DROP INDEX IF EXISTS idx_tasks_duration_ms;

-- Remove comments
COMMENT ON COLUMN tasks.result IS NULL;
COMMENT ON COLUMN tasks.error_message IS NULL;
COMMENT ON COLUMN tasks.duration_ms IS NULL;

-- Note: We don't remove the columns or revert JSON type changes
-- as this could cause data loss and break the application.
-- This is intentionally a safe rollback that only removes the additions
-- that don't affect existing functionality.

-- Drop the new columns (commented out for safety)
-- ALTER TABLE tasks DROP COLUMN IF EXISTS result;
-- ALTER TABLE tasks DROP COLUMN IF EXISTS error_message;
-- ALTER TABLE tasks DROP COLUMN IF EXISTS duration_ms;
