-- Import lists migration for MySQL/MariaDB

-- Create import_lists table
CREATE TABLE IF NOT EXISTS import_lists (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT,
    enable_auto BOOLEAN DEFAULT TRUE,
    enabled BOOLEAN DEFAULT TRUE,
    enable_interactive BOOLEAN DEFAULT FALSE,
    list_type VARCHAR(20) DEFAULT 'program',
    list_order INT DEFAULT 0,
    min_refresh_interval BIGINT DEFAULT 1440,
    quality_profile_id INT NOT NULL,
    root_folder_path VARCHAR(500) NOT NULL,
    should_monitor BOOLEAN DEFAULT TRUE,
    minimum_availability VARCHAR(20) DEFAULT 'released',
    tags TEXT,
    fields TEXT,
    last_sync TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_import_lists_name (name),
    INDEX idx_import_lists_implementation (implementation),
    INDEX idx_import_lists_enabled (enabled),
    INDEX idx_import_lists_enable_auto (enable_auto),
    INDEX idx_import_lists_quality_profile (quality_profile_id)
);

-- Create import_list_movies table
CREATE TABLE IF NOT EXISTS import_list_movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    import_list_id INT NOT NULL,
    tmdb_id INT NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    original_title VARCHAR(500),
    year INT,
    overview TEXT,
    runtime INT,
    images TEXT,
    genres TEXT,
    ratings TEXT,
    certification VARCHAR(20),
    status VARCHAR(20),
    in_cinemas TIMESTAMP NULL,
    physical_release TIMESTAMP NULL,
    digital_release TIMESTAMP NULL,
    website VARCHAR(500),
    youtube_trailer_id VARCHAR(50),
    studio VARCHAR(255),
    minimum_availability VARCHAR(20),
    is_excluded BOOLEAN DEFAULT FALSE,
    is_existing BOOLEAN DEFAULT FALSE,
    is_recommendation BOOLEAN DEFAULT FALSE,
    list_position INT DEFAULT 0,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_import_list_movies_import_list_id (import_list_id),
    INDEX idx_import_list_movies_tmdb_id (tmdb_id),
    INDEX idx_import_list_movies_imdb_id (imdb_id),
    INDEX idx_import_list_movies_year (year),
    INDEX idx_import_list_movies_is_excluded (is_excluded),
    INDEX idx_import_list_movies_is_existing (is_existing),
    INDEX idx_import_list_movies_is_recommendation (is_recommendation),
    INDEX idx_import_list_movies_discovered_at (discovered_at),
    CONSTRAINT fk_import_list_movies_import_list 
        FOREIGN KEY (import_list_id) REFERENCES import_lists(id) ON DELETE CASCADE
);

-- Create import_list_exclusions table
CREATE TABLE IF NOT EXISTS import_list_exclusions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tmdb_id INT NOT NULL UNIQUE,
    movie_title VARCHAR(500) NOT NULL,
    movie_year INT NOT NULL,
    imdb_id VARCHAR(20),
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_import_list_exclusions_tmdb_id (tmdb_id),
    INDEX idx_import_list_exclusions_imdb_id (imdb_id)
);