-- Rollback Complete Database Schema Refactor for PostgreSQL
-- This migration rollback restores the original complex schema

-- Drop the refactored schema
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

-- Restore from backup if it exists
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'radarr_backup') THEN
        -- Restore tables from backup
        ALTER TABLE IF EXISTS radarr_backup.movies SET SCHEMA public;
        ALTER TABLE IF EXISTS radarr_backup.movie_files SET SCHEMA public;
        ALTER TABLE IF EXISTS radarr_backup.collections SET SCHEMA public;
        ALTER TABLE IF EXISTS radarr_backup.health_issues SET SCHEMA public;
        ALTER TABLE IF EXISTS radarr_backup.tasks SET SCHEMA public;
        ALTER TABLE IF EXISTS radarr_backup.scheduled_tasks SET SCHEMA public;

        RAISE NOTICE 'Restored original schema from backup';
    ELSE
        RAISE NOTICE 'No backup schema found - manual data restoration may be required';
    END IF;
END $$;

-- Clean up backup schema
DROP SCHEMA IF EXISTS radarr_backup CASCADE;

COMMENT ON SCHEMA public IS 'Restored original schema - refactor rolled back';
