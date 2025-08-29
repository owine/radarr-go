-- Complete Radarr Go Database Schema Rollback for PostgreSQL
-- This migration removes all tables and functions

-- Drop all triggers first
DROP TRIGGER IF EXISTS update_movies_updated_at ON movies;
DROP TRIGGER IF EXISTS update_movie_files_updated_at ON movie_files;
DROP TRIGGER IF EXISTS update_quality_definitions_updated_at ON quality_definitions;
DROP TRIGGER IF EXISTS update_quality_profiles_updated_at ON quality_profiles;
DROP TRIGGER IF EXISTS update_indexers_updated_at ON indexers;
DROP TRIGGER IF EXISTS update_queue_items_updated_at ON queue_items;
DROP TRIGGER IF EXISTS trigger_download_clients_updated_at ON download_clients;
DROP TRIGGER IF EXISTS trigger_download_history_updated_at ON download_history;
DROP TRIGGER IF EXISTS trigger_import_lists_updated_at ON import_lists;
DROP TRIGGER IF EXISTS trigger_import_list_movies_updated_at ON import_list_movies;
DROP TRIGGER IF EXISTS trigger_import_list_exclusions_updated_at ON import_list_exclusions;
DROP TRIGGER IF EXISTS trigger_update_activity_updated_at ON activity;
DROP TRIGGER IF EXISTS update_host_config_updated_at ON host_config;
DROP TRIGGER IF EXISTS update_naming_config_updated_at ON naming_config;
DROP TRIGGER IF EXISTS update_media_management_config_updated_at ON media_management_config;
DROP TRIGGER IF EXISTS update_root_folders_updated_at ON root_folders;
DROP TRIGGER IF EXISTS trigger_releases_updated_at ON releases;
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;

-- Drop all functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS update_download_clients_updated_at();
DROP FUNCTION IF EXISTS update_download_history_updated_at();
DROP FUNCTION IF EXISTS update_import_lists_updated_at();
DROP FUNCTION IF EXISTS update_import_list_movies_updated_at();
DROP FUNCTION IF EXISTS update_import_list_exclusions_updated_at();
DROP FUNCTION IF EXISTS update_activity_updated_at();
DROP FUNCTION IF EXISTS update_releases_updated_at();

-- Drop all tables in correct order (considering foreign key constraints)
DROP TABLE IF EXISTS releases;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS root_folders;
DROP TABLE IF EXISTS media_management_config;
DROP TABLE IF EXISTS naming_config;
DROP TABLE IF EXISTS host_config;
DROP TABLE IF EXISTS activity;
DROP TABLE IF EXISTS history;
DROP TABLE IF EXISTS import_list_exclusions;
DROP TABLE IF EXISTS import_list_movies;
DROP TABLE IF EXISTS import_lists;
DROP TABLE IF EXISTS download_history;
DROP TABLE IF EXISTS download_clients;
DROP TABLE IF EXISTS queue_items;
DROP TABLE IF EXISTS indexers;
DROP TABLE IF EXISTS quality_profiles;
DROP TABLE IF EXISTS quality_definitions;
DROP TABLE IF EXISTS movie_files;
DROP TABLE IF EXISTS movies;
