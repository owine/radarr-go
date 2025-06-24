-- Rollback Phase 1 features (PostgreSQL version)

-- Drop triggers
DROP TRIGGER IF EXISTS update_quality_profiles_updated_at ON quality_profiles;
DROP TRIGGER IF EXISTS update_custom_formats_updated_at ON custom_formats;
DROP TRIGGER IF EXISTS update_indexers_updated_at ON indexers;
DROP TRIGGER IF EXISTS update_download_clients_updated_at ON download_clients;
DROP TRIGGER IF EXISTS update_queue_items_updated_at ON queue_items;
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;

-- Drop indexes
DROP INDEX IF EXISTS idx_quality_profiles_name;
DROP INDEX IF EXISTS idx_custom_formats_name;
DROP INDEX IF EXISTS idx_indexers_status;
DROP INDEX IF EXISTS idx_indexers_type;
DROP INDEX IF EXISTS idx_download_clients_enable;
DROP INDEX IF EXISTS idx_download_clients_protocol;
DROP INDEX IF EXISTS idx_queue_items_movie_id;
DROP INDEX IF EXISTS idx_queue_items_download_client_id;
DROP INDEX IF EXISTS idx_queue_items_download_id;
DROP INDEX IF EXISTS idx_queue_items_status;
DROP INDEX IF EXISTS idx_download_history_movie_id;
DROP INDEX IF EXISTS idx_download_history_date;
DROP INDEX IF EXISTS idx_download_history_successful;
DROP INDEX IF EXISTS idx_notifications_enable;
DROP INDEX IF EXISTS idx_notification_history_notification_id;
DROP INDEX IF EXISTS idx_notification_history_movie_id;
DROP INDEX IF EXISTS idx_notification_history_date;
DROP INDEX IF EXISTS idx_health_checks_status;
DROP INDEX IF EXISTS idx_health_checks_source;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS health_checks;
DROP TABLE IF EXISTS notification_history;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS download_history;
DROP TABLE IF EXISTS queue_items;
DROP TABLE IF EXISTS indexers;
DROP TABLE IF EXISTS download_clients;
DROP TABLE IF EXISTS custom_formats;
DROP TABLE IF EXISTS quality_profiles;
DROP TABLE IF EXISTS quality_definitions;