-- Rollback download clients migration for PostgreSQL

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_download_history_updated_at ON download_history;
DROP TRIGGER IF EXISTS trigger_download_clients_updated_at ON download_clients;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_download_history_updated_at();
DROP FUNCTION IF EXISTS update_download_clients_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_download_history_protocol;
DROP INDEX IF EXISTS idx_download_history_successful;
DROP INDEX IF EXISTS idx_download_history_date;
DROP INDEX IF EXISTS idx_download_history_download_client_id;
DROP INDEX IF EXISTS idx_download_history_movie_id;

DROP INDEX IF EXISTS idx_queue_items_download_client_id;

DROP INDEX IF EXISTS idx_download_clients_priority;
DROP INDEX IF EXISTS idx_download_clients_enable;
DROP INDEX IF EXISTS idx_download_clients_protocol;
DROP INDEX IF EXISTS idx_download_clients_name;

-- Remove columns from queue_items
ALTER TABLE queue_items DROP COLUMN IF EXISTS downloaded_info;
ALTER TABLE queue_items DROP COLUMN IF EXISTS download_client_id;

-- Drop tables
DROP TABLE IF EXISTS download_history;
DROP TABLE IF EXISTS download_clients;