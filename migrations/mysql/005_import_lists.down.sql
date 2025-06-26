-- Rollback import lists migration for MySQL/MariaDB

-- Drop foreign key constraints first
ALTER TABLE import_list_movies DROP FOREIGN KEY IF EXISTS fk_import_list_movies_import_list;

-- Drop tables
DROP TABLE IF EXISTS import_list_exclusions;
DROP TABLE IF EXISTS import_list_movies;
DROP TABLE IF EXISTS import_lists;