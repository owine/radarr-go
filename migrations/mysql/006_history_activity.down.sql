-- Rollback history and activity migration for MySQL/MariaDB

-- Drop foreign key constraints first
ALTER TABLE history DROP FOREIGN KEY IF EXISTS fk_history_movie;
ALTER TABLE activity DROP FOREIGN KEY IF EXISTS fk_activity_movie;

-- Drop tables
DROP TABLE IF EXISTS activity;
DROP TABLE IF EXISTS history;