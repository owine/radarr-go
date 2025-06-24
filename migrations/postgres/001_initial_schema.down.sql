-- Rollback initial schema for Radarr Go (PostgreSQL version)

-- Drop triggers
DROP TRIGGER IF EXISTS update_movies_updated_at ON movies;
DROP TRIGGER IF EXISTS update_movie_files_updated_at ON movie_files;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_movies_tmdb_id;
DROP INDEX IF EXISTS idx_movies_title_slug;
DROP INDEX IF EXISTS idx_movies_monitored;
DROP INDEX IF EXISTS idx_movies_has_file;
DROP INDEX IF EXISTS idx_movies_status;
DROP INDEX IF EXISTS idx_movie_files_movie_id;
DROP INDEX IF EXISTS idx_movie_files_path;

-- Drop tables
DROP TABLE IF EXISTS movie_files;
DROP TABLE IF EXISTS movies;