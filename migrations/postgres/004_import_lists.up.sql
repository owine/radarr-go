-- Import lists migration for PostgreSQL

-- Create import_lists table
CREATE TABLE IF NOT EXISTS import_lists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT DEFAULT '{}'::TEXT,
    enable_auto BOOLEAN DEFAULT TRUE,
    enabled BOOLEAN DEFAULT TRUE,
    enable_interactive BOOLEAN DEFAULT FALSE,
    list_type VARCHAR(20) DEFAULT 'program',
    list_order INTEGER DEFAULT 0,
    min_refresh_interval BIGINT DEFAULT 1440, -- minutes
    quality_profile_id INTEGER NOT NULL,
    root_folder_path VARCHAR(500) NOT NULL,
    should_monitor BOOLEAN DEFAULT TRUE,
    minimum_availability VARCHAR(20) DEFAULT 'released',
    tags TEXT DEFAULT '[]'::TEXT,
    fields TEXT DEFAULT '[]'::TEXT,
    last_sync TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create import_list_movies table
CREATE TABLE IF NOT EXISTS import_list_movies (
    id SERIAL PRIMARY KEY,
    import_list_id INTEGER NOT NULL REFERENCES import_lists(id) ON DELETE CASCADE,
    tmdb_id INTEGER NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    original_title VARCHAR(500),
    year INTEGER,
    overview TEXT,
    runtime INTEGER,
    images TEXT DEFAULT '[]'::TEXT,
    genres TEXT DEFAULT '[]'::TEXT,
    ratings TEXT DEFAULT '{}'::TEXT,
    certification VARCHAR(20),
    status VARCHAR(20),
    in_cinemas TIMESTAMP,
    physical_release TIMESTAMP,
    digital_release TIMESTAMP,
    website VARCHAR(500),
    youtube_trailer_id VARCHAR(50),
    studio VARCHAR(255),
    minimum_availability VARCHAR(20),
    is_excluded BOOLEAN DEFAULT FALSE,
    is_existing BOOLEAN DEFAULT FALSE,
    is_recommendation BOOLEAN DEFAULT FALSE,
    list_position INTEGER DEFAULT 0,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create import_list_exclusions table
CREATE TABLE IF NOT EXISTS import_list_exclusions (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER NOT NULL UNIQUE,
    movie_title VARCHAR(500) NOT NULL,
    movie_year INTEGER NOT NULL,
    imdb_id VARCHAR(20),
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_import_lists_name ON import_lists(name);
CREATE INDEX IF NOT EXISTS idx_import_lists_implementation ON import_lists(implementation);
CREATE INDEX IF NOT EXISTS idx_import_lists_enabled ON import_lists(enabled);
CREATE INDEX IF NOT EXISTS idx_import_lists_enable_auto ON import_lists(enable_auto);
CREATE INDEX IF NOT EXISTS idx_import_lists_quality_profile ON import_lists(quality_profile_id);

CREATE INDEX IF NOT EXISTS idx_import_list_movies_import_list_id ON import_list_movies(import_list_id);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_tmdb_id ON import_list_movies(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_imdb_id ON import_list_movies(imdb_id);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_year ON import_list_movies(year);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_is_excluded ON import_list_movies(is_excluded);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_is_existing ON import_list_movies(is_existing);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_is_recommendation ON import_list_movies(is_recommendation);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_discovered_at ON import_list_movies(discovered_at);

CREATE INDEX IF NOT EXISTS idx_import_list_exclusions_tmdb_id ON import_list_exclusions(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_import_list_exclusions_imdb_id ON import_list_exclusions(imdb_id);

-- Create trigger to automatically update the updated_at timestamp for import_lists
CREATE OR REPLACE FUNCTION update_import_lists_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_import_lists_updated_at
    BEFORE UPDATE ON import_lists
    FOR EACH ROW
    EXECUTE FUNCTION update_import_lists_updated_at();

-- Create trigger to automatically update the updated_at timestamp for import_list_movies
CREATE OR REPLACE FUNCTION update_import_list_movies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_import_list_movies_updated_at
    BEFORE UPDATE ON import_list_movies
    FOR EACH ROW
    EXECUTE FUNCTION update_import_list_movies_updated_at();

-- Create trigger to automatically update the updated_at timestamp for import_list_exclusions
CREATE OR REPLACE FUNCTION update_import_list_exclusions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_import_list_exclusions_updated_at
    BEFORE UPDATE ON import_list_exclusions
    FOR EACH ROW
    EXECUTE FUNCTION update_import_list_exclusions_updated_at();