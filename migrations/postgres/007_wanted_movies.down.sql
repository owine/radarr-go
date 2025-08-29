-- Rollback wanted movies tracking system

-- Drop the view first
DROP VIEW IF EXISTS wanted_movies_with_details;

-- Drop triggers and functions
DROP TRIGGER IF EXISTS trigger_update_wanted_movies_updated_at ON wanted_movies;
DROP FUNCTION IF EXISTS update_wanted_movies_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_wanted_movies_search_eligible;
DROP INDEX IF EXISTS idx_wanted_movies_available_searchable;
DROP INDEX IF EXISTS idx_wanted_movies_status_priority;
DROP INDEX IF EXISTS idx_wanted_movies_current_quality_id;
DROP INDEX IF EXISTS idx_wanted_movies_target_quality_id;
DROP INDEX IF EXISTS idx_wanted_movies_search_attempts;
DROP INDEX IF EXISTS idx_wanted_movies_next_search_time;
DROP INDEX IF EXISTS idx_wanted_movies_last_search_time;
DROP INDEX IF EXISTS idx_wanted_movies_is_available;
DROP INDEX IF EXISTS idx_wanted_movies_priority;
DROP INDEX IF EXISTS idx_wanted_movies_status;
DROP INDEX IF EXISTS idx_wanted_movies_movie_id;

-- Drop the table
DROP TABLE IF EXISTS wanted_movies;
