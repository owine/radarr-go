-- Initial schema for Radarr Go (MySQL/MariaDB version)

-- Movies table
CREATE TABLE IF NOT EXISTS movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    original_title VARCHAR(255),
    original_language VARCHAR(50),
    alternate_titles TEXT,
    secondary_year INT,
    secondary_year_source_id INT DEFAULT 0,
    sort_title VARCHAR(255),
    size_on_disk BIGINT DEFAULT 0,
    status VARCHAR(50) DEFAULT 'tba',
    overview TEXT,
    in_cinemas DATETIME,
    physical_release DATETIME,
    digital_release DATETIME,
    physical_release_note TEXT,
    images TEXT,
    website VARCHAR(500),
    year INT,
    youtube_trailer_id VARCHAR(100),
    studio VARCHAR(255),
    path TEXT,
    quality_profile_id INT,
    has_file TINYINT(1) DEFAULT 0,
    movie_file_id INT DEFAULT 0,
    monitored TINYINT(1) DEFAULT 1,
    minimum_availability VARCHAR(50) DEFAULT 'tba',
    is_available TINYINT(1) DEFAULT 0,
    folder_name VARCHAR(255),
    runtime INT DEFAULT 0,
    clean_title VARCHAR(255),
    imdb_id VARCHAR(20),
    tmdb_id INT UNIQUE,
    title_slug VARCHAR(255) UNIQUE,
    root_folder_path TEXT,
    certification VARCHAR(50),
    genres TEXT,
    tags TEXT,
    added DATETIME,
    add_options TEXT,
    ratings TEXT,
    collection TEXT,
    popularity DECIMAL(10,3) DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ----

-- Movie files table
CREATE TABLE IF NOT EXISTS movie_files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    relative_path TEXT,
    path TEXT,
    size BIGINT DEFAULT 0,
    date_added DATETIME,
    scene_name VARCHAR(255),
    indexer_flags INT DEFAULT 0,
    quality TEXT,
    custom_formats TEXT,
    custom_format_score INT DEFAULT 0,
    media_info TEXT,
    original_file_path TEXT,
    languages TEXT,
    release_group VARCHAR(255),
    edition VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes
CREATE INDEX idx_movies_tmdb_id ON movies(tmdb_id);
CREATE INDEX idx_movies_title_slug ON movies(title_slug);
CREATE INDEX idx_movies_monitored ON movies(monitored);
CREATE INDEX idx_movies_has_file ON movies(has_file);
CREATE INDEX idx_movies_status ON movies(status);
CREATE INDEX idx_movie_files_movie_id ON movie_files(movie_id);
CREATE INDEX idx_movie_files_path ON movie_files(path(255));