-- Drop file organization and import management tables

-- Drop triggers first
DROP TRIGGER IF EXISTS update_naming_config_updated_at ON naming_config;
DROP TRIGGER IF EXISTS update_file_operations_updated_at ON file_operations;
DROP TRIGGER IF EXISTS update_manual_imports_updated_at ON manual_imports;
DROP TRIGGER IF EXISTS update_file_organizations_updated_at ON file_organizations;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_file_organizations_created_at;
DROP INDEX IF EXISTS idx_file_organizations_movie_id;
DROP INDEX IF EXISTS idx_file_organizations_status;

DROP INDEX IF EXISTS idx_manual_imports_created_at;
DROP INDEX IF EXISTS idx_manual_imports_movie_id;

DROP INDEX IF EXISTS idx_file_operations_created_at;
DROP INDEX IF EXISTS idx_file_operations_movie_id;
DROP INDEX IF EXISTS idx_file_operations_type;
DROP INDEX IF EXISTS idx_file_operations_status;

-- Drop tables
DROP TABLE IF EXISTS naming_config;
DROP TABLE IF EXISTS file_operations;
DROP TABLE IF EXISTS manual_imports;
DROP TABLE IF EXISTS file_organizations;
