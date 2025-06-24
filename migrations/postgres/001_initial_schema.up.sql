-- Initial schema for Radarr Go (PostgreSQL version)

-- Movies table
CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    original_title VARCHAR(255),
    original_language VARCHAR(50),
    alternate_titles TEXT,
    secondary_year INTEGER,
    secondary_year_source_id INTEGER DEFAULT 0,
    sort_title VARCHAR(255),
    size_on_disk BIGINT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'tba',
    overview TEXT,
    in_cinemas TIMESTAMP,
    physical_release TIMESTAMP,
    digital_release TIMESTAMP,
    physical_release_note TEXT,
    images TEXT,
    website VARCHAR(500),
    year INTEGER,
    youtube_trailer_id VARCHAR(100),
    studio VARCHAR(255),
    path TEXT,
    quality_profile_id INTEGER,
    has_file BOOLEAN DEFAULT FALSE,
    movie_file_id INTEGER DEFAULT 0,
    monitored BOOLEAN DEFAULT TRUE,
    minimum_availability VARCHAR(50) DEFAULT 'tba',
    is_available BOOLEAN DEFAULT FALSE,
    folder_name VARCHAR(255),
    runtime INTEGER DEFAULT 0,
    clean_title VARCHAR(255),
    imdb_id VARCHAR(20),
    tmdb_id INTEGER UNIQUE,
    title_slug VARCHAR(255) UNIQUE,
    root_folder_path TEXT,
    certification VARCHAR(50),
    genres TEXT,
    tags TEXT,
    added TIMESTAMP,
    add_options TEXT,
    ratings TEXT,
    collection TEXT,
    popularity REAL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Movie files table
CREATE TABLE IF NOT EXISTS movie_files (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER,
    relative_path TEXT,
    path TEXT,
    size BIGINT DEFAULT 0,
    date_added TIMESTAMP,
    scene_name VARCHAR(255),
    indexer_flags INTEGER DEFAULT 0,
    quality TEXT,
    custom_formats TEXT,
    custom_format_score INTEGER DEFAULT 0,
    media_info TEXT,
    original_file_path TEXT,
    languages TEXT,
    release_group VARCHAR(255),
    edition VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_movies_tmdb_id ON movies(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_movies_title_slug ON movies(title_slug);
CREATE INDEX IF NOT EXISTS idx_movies_monitored ON movies(monitored);
CREATE INDEX IF NOT EXISTS idx_movies_has_file ON movies(has_file);
CREATE INDEX IF NOT EXISTS idx_movies_status ON movies(status);
CREATE INDEX IF NOT EXISTS idx_movie_files_movie_id ON movie_files(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_files_path ON movie_files(path);

-- Function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers to update timestamps
CREATE TRIGGER update_movies_updated_at
    BEFORE UPDATE ON movies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_movie_files_updated_at
    BEFORE UPDATE ON movie_files
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();