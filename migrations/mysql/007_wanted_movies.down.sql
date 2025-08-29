-- Rollback wanted movies tracking system for MySQL/MariaDB

-- Drop the view first
DROP VIEW IF EXISTS wanted_movies_with_details;

-- Drop indexes
DROP INDEX IF EXISTS idx_wanted_movies_search_eligible ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_available_searchable ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_status_priority ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_current_quality_id ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_target_quality_id ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_search_attempts ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_next_search_time ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_last_search_time ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_is_available ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_priority ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_status ON wanted_movies;
DROP INDEX IF EXISTS idx_wanted_movies_movie_id ON wanted_movies;

-- Drop the table
DROP TABLE IF EXISTS wanted_movies;
