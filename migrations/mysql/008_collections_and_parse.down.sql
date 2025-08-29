-- Rollback Collections and Parse/Rename System Migration for MySQL
-- Removes collections management and parse/rename functionality

-- Drop parse_cache table
DROP TABLE IF EXISTS parse_cache;

-- Remove foreign key constraints from movies table
ALTER TABLE movies DROP FOREIGN KEY IF EXISTS fk_movies_collection_tmdb_id;

-- Remove collection_tmdb_id column from movies table
ALTER TABLE movies DROP COLUMN IF EXISTS collection_tmdb_id;

-- Remove foreign key constraint from collections table
ALTER TABLE collections DROP FOREIGN KEY IF EXISTS fk_collections_quality_profile_id;

-- Drop collections table
DROP TABLE IF EXISTS collections;
