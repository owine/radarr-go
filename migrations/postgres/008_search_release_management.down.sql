-- Rollback Search & Release Management System Migration for PostgreSQL

-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_releases_updated_at ON releases;
DROP FUNCTION IF EXISTS update_releases_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_releases_created_at;
DROP INDEX IF EXISTS idx_releases_quality_weight;
DROP INDEX IF EXISTS idx_releases_status;
DROP INDEX IF EXISTS idx_releases_publish_date;
DROP INDEX IF EXISTS idx_releases_tmdb_id;
DROP INDEX IF EXISTS idx_releases_imdb_id;
DROP INDEX IF EXISTS idx_releases_indexer_id;
DROP INDEX IF EXISTS idx_releases_movie_id;
DROP INDEX IF EXISTS idx_releases_title;
DROP INDEX IF EXISTS idx_releases_guid_indexer_unique;
DROP INDEX IF EXISTS idx_releases_guid_indexer;

-- Drop table
DROP TABLE IF EXISTS releases;