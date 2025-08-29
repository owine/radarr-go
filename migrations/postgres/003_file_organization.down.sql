-- Drop file organization and import management tables
-- Note: naming_config table and update_updated_at_column() function are from migration 001, not this migration

-- Drop triggers for the tables created in this migration only
DROP TRIGGER IF EXISTS update_file_operations_updated_at ON file_operations;
DROP TRIGGER IF EXISTS update_manual_imports_updated_at ON manual_imports;
DROP TRIGGER IF EXISTS update_file_organizations_updated_at ON file_organizations;

-- Drop indexes for the tables created in this migration
DROP INDEX IF EXISTS idx_file_organizations_created_at;
DROP INDEX IF EXISTS idx_file_organizations_movie_id;
DROP INDEX IF EXISTS idx_file_organizations_status;

DROP INDEX IF EXISTS idx_manual_imports_created_at;
DROP INDEX IF EXISTS idx_manual_imports_movie_id;

DROP INDEX IF EXISTS idx_file_operations_created_at;
DROP INDEX IF EXISTS idx_file_operations_movie_id;
DROP INDEX IF EXISTS idx_file_operations_type;
DROP INDEX IF EXISTS idx_file_operations_status;

-- Drop only the tables created in this migration
DROP TABLE IF EXISTS file_operations;
DROP TABLE IF EXISTS manual_imports;
DROP TABLE IF EXISTS file_organizations;
