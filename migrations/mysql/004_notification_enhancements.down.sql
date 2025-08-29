-- Rollback enhanced notification system migration for MySQL/MariaDB

-- Remove added columns from notifications table
-- Only drop columns that were added in this migration
ALTER TABLE notifications DROP COLUMN IF EXISTS supports_on_manual_interaction_required;
ALTER TABLE notifications DROP COLUMN IF EXISTS supports_on_movie_added;
ALTER TABLE notifications DROP COLUMN IF EXISTS on_movie_added;

-- Drop health_checks table
DROP TABLE IF EXISTS health_checks;

-- Drop notification_history table
DROP TABLE IF EXISTS notification_history;
