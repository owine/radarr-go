-- Rollback download clients migration for MySQL/MariaDB

-- Drop index first
DROP INDEX IF EXISTS idx_queue_items_download_client_id ON queue_items;

-- Remove foreign key constraints and columns from queue_items
ALTER TABLE queue_items 
DROP FOREIGN KEY IF EXISTS fk_queue_items_download_client,
DROP COLUMN IF EXISTS downloaded_info,
DROP COLUMN IF EXISTS download_client_id;

-- Drop tables
DROP TABLE IF EXISTS download_history;
DROP TABLE IF EXISTS download_clients;