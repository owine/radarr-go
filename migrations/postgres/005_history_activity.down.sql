-- Rollback history and activity migration for PostgreSQL

-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_activity_updated_at ON activity;
DROP FUNCTION IF EXISTS update_activity_updated_at();

-- Drop tables
DROP TABLE IF EXISTS activity;
DROP TABLE IF EXISTS history;