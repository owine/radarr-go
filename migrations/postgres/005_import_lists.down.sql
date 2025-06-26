-- Rollback import lists migration for PostgreSQL

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_import_list_exclusions_updated_at ON import_list_exclusions;
DROP TRIGGER IF EXISTS trigger_import_list_movies_updated_at ON import_list_movies;
DROP TRIGGER IF EXISTS trigger_import_lists_updated_at ON import_lists;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_import_list_exclusions_updated_at();
DROP FUNCTION IF EXISTS update_import_list_movies_updated_at();
DROP FUNCTION IF EXISTS update_import_lists_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_import_list_exclusions_imdb_id;
DROP INDEX IF EXISTS idx_import_list_exclusions_tmdb_id;

DROP INDEX IF EXISTS idx_import_list_movies_discovered_at;
DROP INDEX IF EXISTS idx_import_list_movies_is_recommendation;
DROP INDEX IF EXISTS idx_import_list_movies_is_existing;
DROP INDEX IF EXISTS idx_import_list_movies_is_excluded;
DROP INDEX IF EXISTS idx_import_list_movies_year;
DROP INDEX IF EXISTS idx_import_list_movies_imdb_id;
DROP INDEX IF EXISTS idx_import_list_movies_tmdb_id;
DROP INDEX IF EXISTS idx_import_list_movies_import_list_id;

DROP INDEX IF EXISTS idx_import_lists_quality_profile;
DROP INDEX IF EXISTS idx_import_lists_enable_auto;
DROP INDEX IF EXISTS idx_import_lists_enabled;
DROP INDEX IF EXISTS idx_import_lists_implementation;
DROP INDEX IF EXISTS idx_import_lists_name;

-- Drop tables
DROP TABLE IF EXISTS import_list_exclusions;
DROP TABLE IF EXISTS import_list_movies;
DROP TABLE IF EXISTS import_lists;