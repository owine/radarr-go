-- Migration 011 Down: Remove app_config table and related objects

-- Drop trigger and function
DROP TRIGGER IF EXISTS update_app_config_updated_at ON app_config;
DROP FUNCTION IF EXISTS update_app_config_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_app_config_key;
DROP INDEX IF EXISTS idx_app_config_created_at;
DROP INDEX IF EXISTS idx_app_config_updated_at;

-- Drop the table (commented out for safety to prevent data loss)
-- Uncomment only if you're sure you want to remove all configuration data
-- DROP TABLE IF EXISTS app_config;

-- For safety, we only remove the indexes and triggers in rollback
-- The table and its data are preserved to prevent configuration loss
