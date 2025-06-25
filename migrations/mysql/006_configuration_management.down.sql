-- Rollback configuration management migration for MySQL/MariaDB

-- Drop indexes first
DROP INDEX IF EXISTS idx_root_folders_path ON root_folders;
DROP INDEX IF EXISTS idx_root_folders_accessible ON root_folders;
DROP INDEX IF EXISTS idx_media_config_recycle ON media_management_config;
DROP INDEX IF EXISTS idx_naming_config_rename ON naming_config;
DROP INDEX IF EXISTS idx_host_config_auth ON host_config;

-- Drop tables
DROP TABLE IF EXISTS root_folders;
DROP TABLE IF EXISTS media_management_config;
DROP TABLE IF EXISTS naming_config;
DROP TABLE IF EXISTS host_config;