-- Migration 009: Fix quality definitions dependency for wanted_movies table (MySQL/MariaDB)
-- This migration ensures quality_definitions table exists and is properly configured

-- Create quality_definitions table if it doesn't exist (should already exist from 001)
-- This is a safety measure to ensure wanted_movies foreign keys work
CREATE TABLE IF NOT EXISTS quality_definitions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(255) NOT NULL UNIQUE,
    weight INT NOT NULL DEFAULT 1,
    min_size DECIMAL(8,3) DEFAULT 0,
    max_size DECIMAL(8,3) DEFAULT 400,
    preferred_size DECIMAL(8,3),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Ensure indexes exist for performance
CREATE INDEX IF NOT EXISTS idx_quality_definitions_title ON quality_definitions(title);
CREATE INDEX IF NOT EXISTS idx_quality_definitions_weight ON quality_definitions(weight);

-- Insert default quality definitions if not present
-- This ensures wanted_movies has valid quality_ids to reference
INSERT IGNORE INTO quality_definitions (id, title, weight, min_size, max_size) VALUES
    (0, 'Unknown', 1, 0, 199.9),
    (1, 'SDTV', 8, 0, 199.9),
    (2, 'DVD', 9, 0, 199.9),
    (3, 'WEBDL-1080p', 20, 2, 137.3),
    (4, 'HDTV-720p', 15, 0.8, 137.3),
    (5, 'WEBDL-720p', 16, 0.8, 137.3),
    (6, 'Bluray-720p', 18, 4.3, 137.3),
    (7, 'Bluray-1080p', 22, 4.3, 258.1),
    (8, 'WEBDL-480p', 11, 0, 199.9),
    (9, 'HDTV-1080p', 19, 2, 137.3),
    (12, 'WEBRip-480p', 12, 0, 199.9),
    (14, 'WEBRip-720p', 17, 0.8, 137.3),
    (15, 'WEBRip-1080p', 21, 2, 137.3),
    (16, 'HDTV-2160p', 24, 4.7, 199.9),
    (17, 'WEBRip-2160p', 26, 4.7, 258.1),
    (18, 'WEBDL-2160p', 25, 4.7, 258.1),
    (19, 'Bluray-2160p', 27, 4.3, 258.1),
    (20, 'Bluray-480p', 13, 0, 199.9),
    (21, 'Bluray-576p', 14, 0, 199.9),
    (23, 'DVD-R', 10, 0, 199.9),
    (24, 'WORKPRINT', 2, 0, 199.9),
    (25, 'CAM', 3, 0, 199.9),
    (26, 'TELESYNC', 4, 0, 199.9),
    (27, 'TELECINE', 5, 0, 199.9),
    (28, 'DVDSCR', 7, 0, 199.9),
    (29, 'REGIONAL', 6, 0, 199.9),
    (30, 'Remux-1080p', 23, 0, 0),
    (31, 'Remux-2160p', 28, 0, 0)
ON DUPLICATE KEY UPDATE
    title = VALUES(title),
    weight = VALUES(weight),
    min_size = VALUES(min_size),
    max_size = VALUES(max_size);

-- Validate that foreign key constraints will work
-- Check that quality_definitions has records
SELECT 'Quality definitions migration completed successfully' as status;
SELECT COUNT(*) as quality_count FROM quality_definitions;
