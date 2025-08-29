-- Rollback Collections and Parse/Rename System Migration for PostgreSQL
-- Removes collections management and parse/rename functionality

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_collections_updated_at ON collections;
DROP TRIGGER IF EXISTS trigger_parse_cache_updated_at ON parse_cache;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_collections_updated_at();
DROP FUNCTION IF EXISTS update_parse_cache_updated_at();

-- Drop parse_cache table
DROP TABLE IF EXISTS parse_cache;

-- Remove foreign key constraints from movies table
ALTER TABLE movies DROP CONSTRAINT IF EXISTS fk_movies_collection_tmdb_id;

-- Remove collection_tmdb_id column from movies table
ALTER TABLE movies DROP COLUMN IF EXISTS collection_tmdb_id;

-- Drop foreign key constraint from collections table
ALTER TABLE collections DROP CONSTRAINT IF EXISTS fk_collections_quality_profile_id;

-- Drop indexes
DROP INDEX IF EXISTS idx_collections_tmdb_id;
DROP INDEX IF EXISTS idx_collections_monitored;
DROP INDEX IF EXISTS idx_collections_quality_profile_id;
DROP INDEX IF EXISTS idx_collections_title;
DROP INDEX IF EXISTS idx_movies_collection_tmdb_id;
DROP INDEX IF EXISTS idx_parse_cache_release_title;
DROP INDEX IF EXISTS idx_parse_cache_created_at;

-- Drop collections table
DROP TABLE IF EXISTS collections;
