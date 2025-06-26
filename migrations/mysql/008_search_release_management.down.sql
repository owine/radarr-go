-- Rollback Search & Release Management System Migration for MySQL/MariaDB

-- Drop indexes
DROP INDEX IF EXISTS idx_releases_created_at ON releases;
DROP INDEX IF EXISTS idx_releases_quality_weight ON releases;
DROP INDEX IF EXISTS idx_releases_status ON releases;
DROP INDEX IF EXISTS idx_releases_publish_date ON releases;
DROP INDEX IF EXISTS idx_releases_tmdb_id ON releases;
DROP INDEX IF EXISTS idx_releases_imdb_id ON releases;
DROP INDEX IF EXISTS idx_releases_indexer_id ON releases;
DROP INDEX IF EXISTS idx_releases_movie_id ON releases;
DROP INDEX IF EXISTS idx_releases_title ON releases;
DROP INDEX IF EXISTS idx_releases_guid_indexer_unique ON releases;
DROP INDEX IF EXISTS idx_releases_guid_indexer ON releases;

-- Drop table
DROP TABLE IF EXISTS releases;