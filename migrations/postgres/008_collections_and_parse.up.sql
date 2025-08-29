-- Collections and Parse/Rename System Migration for PostgreSQL
-- Adds collections management and parse/rename functionality

-- Create collections table
CREATE TABLE IF NOT EXISTS collections (
    id SERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    clean_title VARCHAR(500),
    sort_title VARCHAR(500),
    tmdb_id INTEGER UNIQUE NOT NULL,
    overview TEXT,
    monitored BOOLEAN DEFAULT TRUE,
    quality_profile_id INTEGER NOT NULL DEFAULT 1,
    root_folder_path VARCHAR(500),
    search_on_add BOOLEAN DEFAULT FALSE,
    minimum_availability VARCHAR(20) DEFAULT 'announced',
    last_info_sync TIMESTAMP,
    images TEXT DEFAULT '[]'::TEXT,
    tags TEXT DEFAULT '[]'::TEXT,
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for collections
CREATE INDEX IF NOT EXISTS idx_collections_tmdb_id ON collections(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_collections_monitored ON collections(monitored);
CREATE INDEX IF NOT EXISTS idx_collections_quality_profile_id ON collections(quality_profile_id);
CREATE INDEX IF NOT EXISTS idx_collections_title ON collections(title);

-- Add collection_tmdb_id column to movies table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'movies'
        AND column_name = 'collection_tmdb_id'
    ) THEN
        ALTER TABLE movies ADD COLUMN collection_tmdb_id INTEGER;
    END IF;
END $$;

-- Create index for collection relationships
CREATE INDEX IF NOT EXISTS idx_movies_collection_tmdb_id ON movies(collection_tmdb_id);

-- Add foreign key constraint for collection relationships (optional, can be nullable)
-- This links movies to collections via TMDB ID
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_movies_collection_tmdb_id'
        AND table_name = 'movies'
    ) THEN
        ALTER TABLE movies
        ADD CONSTRAINT fk_movies_collection_tmdb_id
        FOREIGN KEY (collection_tmdb_id) REFERENCES collections(tmdb_id)
        ON DELETE SET NULL;
    END IF;
EXCEPTION
    WHEN others THEN
        -- Ignore errors if constraint already exists or other issues
        NULL;
END $$;

-- Create parse_cache table for caching parse results
CREATE TABLE IF NOT EXISTS parse_cache (
    id SERIAL PRIMARY KEY,
    release_title VARCHAR(1000) NOT NULL,
    parsed_info TEXT NOT NULL, -- JSON containing ParsedMovieInfo
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create unique index on release title for parse cache
CREATE UNIQUE INDEX IF NOT EXISTS idx_parse_cache_release_title ON parse_cache(release_title);

-- Create index for cleanup by created_at
CREATE INDEX IF NOT EXISTS idx_parse_cache_created_at ON parse_cache(created_at);

-- Add any missing quality profile foreign key constraint for collections
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_collections_quality_profile_id'
        AND table_name = 'collections'
    ) THEN
        ALTER TABLE collections
        ADD CONSTRAINT fk_collections_quality_profile_id
        FOREIGN KEY (quality_profile_id) REFERENCES quality_profiles(id)
        ON DELETE RESTRICT;
    END IF;
EXCEPTION
    WHEN others THEN
        -- Ignore errors if constraint already exists or quality_profiles table doesn't exist yet
        NULL;
END $$;

-- Create trigger to update updated_at for collections
CREATE OR REPLACE FUNCTION update_collections_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_collections_updated_at ON collections;
CREATE TRIGGER trigger_collections_updated_at
    BEFORE UPDATE ON collections
    FOR EACH ROW
    EXECUTE FUNCTION update_collections_updated_at();

-- Create trigger to update updated_at for parse_cache
CREATE OR REPLACE FUNCTION update_parse_cache_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_parse_cache_updated_at ON parse_cache;
CREATE TRIGGER trigger_parse_cache_updated_at
    BEFORE UPDATE ON parse_cache
    FOR EACH ROW
    EXECUTE FUNCTION update_parse_cache_updated_at();
