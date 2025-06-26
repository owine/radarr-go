-- Rollback configuration management migration for PostgreSQL

-- Drop triggers first
DROP TRIGGER IF EXISTS update_root_folders_updated_at ON root_folders;
DROP TRIGGER IF EXISTS update_media_management_config_updated_at ON media_management_config;
DROP TRIGGER IF EXISTS update_naming_config_updated_at ON naming_config;
DROP TRIGGER IF EXISTS update_host_config_updated_at ON host_config;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_root_folders_path;
DROP INDEX IF EXISTS idx_root_folders_accessible;
DROP INDEX IF EXISTS idx_media_config_recycle;
DROP INDEX IF EXISTS idx_naming_config_rename;
DROP INDEX IF EXISTS idx_host_config_auth;

-- Drop tables
DROP TABLE IF EXISTS root_folders;
DROP TABLE IF EXISTS media_management_config;
DROP TABLE IF EXISTS naming_config;
DROP TABLE IF EXISTS host_config;