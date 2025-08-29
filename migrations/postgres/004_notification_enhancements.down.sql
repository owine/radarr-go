-- Rollback enhanced notification system migration for PostgreSQL

-- Drop triggers
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Remove added columns from notifications table
ALTER TABLE notifications DROP COLUMN IF EXISTS supports_on_manual_interaction_required;
ALTER TABLE notifications DROP COLUMN IF EXISTS supports_on_movie_added;
ALTER TABLE notifications DROP COLUMN IF EXISTS include_health_warnings;
ALTER TABLE notifications DROP COLUMN IF EXISTS on_manual_interaction_required;
ALTER TABLE notifications DROP COLUMN IF EXISTS on_movie_added;

-- Drop health_checks table
DROP TABLE IF EXISTS health_checks;

-- Drop notification_history table
DROP TABLE IF EXISTS notification_history;
