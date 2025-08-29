-- Drop file organization and import management tables

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS naming_config;
DROP TABLE IF EXISTS file_operations;
DROP TABLE IF EXISTS manual_imports;
DROP TABLE IF EXISTS file_organizations;
