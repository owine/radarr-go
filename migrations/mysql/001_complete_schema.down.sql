-- Complete Radarr Go Database Schema Rollback for MySQL/MariaDB
-- This migration removes all tables

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
DROP TABLE IF EXISTS movie_files;
DROP TABLE IF EXISTS movies;