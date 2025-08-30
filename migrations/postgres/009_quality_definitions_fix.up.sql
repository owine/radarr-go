-- Migration 009: Fix quality definitions dependency for wanted_movies table
-- This migration ensures quality_definitions table exists and is properly configured

-- Create quality_definitions table if it doesn't exist (should already exist from 001)
-- This is a safety measure to ensure wanted_movies foreign keys work
CREATE TABLE IF NOT EXISTS quality_definitions (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL UNIQUE,
    weight INTEGER NOT NULL DEFAULT 1,
    min_size DOUBLE PRECISION DEFAULT 0,
    max_size DOUBLE PRECISION DEFAULT 400,
    preferred_size DOUBLE PRECISION,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ensure indexes exist for performance
CREATE INDEX IF NOT EXISTS idx_quality_definitions_title ON quality_definitions(title);
CREATE INDEX IF NOT EXISTS idx_quality_definitions_weight ON quality_definitions(weight);

-- Ensure trigger exists for updated_at
CREATE OR REPLACE FUNCTION update_quality_definitions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger only if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger 
        WHERE tgname = 'update_quality_definitions_updated_at'
    ) THEN
        CREATE TRIGGER update_quality_definitions_updated_at 
            BEFORE UPDATE ON quality_definitions
            FOR EACH ROW 
            EXECUTE FUNCTION update_quality_definitions_updated_at();
    END IF;
END $$;

-- Insert default quality definitions if not present
-- This ensures wanted_movies has valid quality_ids to reference
INSERT INTO quality_definitions (id, title, weight, min_size, max_size) VALUES
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
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    weight = EXCLUDED.weight,
    min_size = EXCLUDED.min_size,
    max_size = EXCLUDED.max_size;

-- Validate foreign key constraints work correctly
DO $$
BEGIN
    -- Test that quality_definitions table can be referenced
    IF NOT EXISTS (
        SELECT 1 FROM quality_definitions LIMIT 1
    ) THEN
        RAISE EXCEPTION 'Quality definitions table is empty - wanted_movies foreign keys will fail';
    END IF;
    
    -- Log success
    RAISE NOTICE 'Quality definitions migration completed successfully';
    RAISE NOTICE 'Foreign key dependencies for wanted_movies are now safe';
END $$;