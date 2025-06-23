-- Drop triggers
DROP TRIGGER IF EXISTS update_movies_updated_at;
DROP TRIGGER IF EXISTS update_movie_files_updated_at;

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