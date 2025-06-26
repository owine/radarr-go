-- Drop queue system migration for PostgreSQL

DROP TRIGGER IF EXISTS trigger_queue_items_updated_at ON queue_items;
DROP FUNCTION IF EXISTS update_queue_updated_at();
DROP TABLE IF EXISTS queue_items;