-- Initial schema for Radarr Go

-- Movies table
CREATE TABLE IF NOT EXISTS movies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    original_title TEXT,
    original_language TEXT,
    alternate_titles TEXT,
    secondary_year INTEGER,
    secondary_year_source_id INTEGER DEFAULT 0,
    sort_title TEXT,
    size_on_disk INTEGER DEFAULT 0,
    status TEXT DEFAULT 'tba',
    overview TEXT,
    in_cinemas DATETIME,
    physical_release DATETIME,
    digital_release DATETIME,
    physical_release_note TEXT,
    images TEXT,
    website TEXT,
    year INTEGER,
    youtube_trailer_id TEXT,
    studio TEXT,
    path TEXT,
    quality_profile_id INTEGER,
    has_file INTEGER DEFAULT 0,
    movie_file_id INTEGER DEFAULT 0,
    monitored INTEGER DEFAULT 1,
    minimum_availability TEXT DEFAULT 'tba',
    is_available INTEGER DEFAULT 0,
    folder_name TEXT,
    runtime INTEGER DEFAULT 0,
    clean_title TEXT,
    imdb_id TEXT,
    tmdb_id INTEGER UNIQUE,
    title_slug TEXT UNIQUE,
    root_folder_path TEXT,
    certification TEXT,
    genres TEXT,
    tags TEXT,
    added DATETIME,
    add_options TEXT,
    ratings TEXT,
    collection TEXT,
    popularity REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Movie files table
CREATE TABLE IF NOT EXISTS movie_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id INTEGER,
    relative_path TEXT,
    path TEXT,
    size INTEGER DEFAULT 0,
    date_added DATETIME,
    scene_name TEXT,
    indexer_flags INTEGER DEFAULT 0,
    quality TEXT,
    custom_formats TEXT,
    custom_format_score INTEGER DEFAULT 0,
    media_info TEXT,
    original_file_path TEXT,
    languages TEXT,
    release_group TEXT,
    edition TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
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

-- Triggers to update timestamps
CREATE TRIGGER IF NOT EXISTS update_movies_updated_at
    AFTER UPDATE ON movies
    FOR EACH ROW
BEGIN
    UPDATE movies SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_movie_files_updated_at
    AFTER UPDATE ON movie_files
    FOR EACH ROW
BEGIN
    UPDATE movie_files SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;