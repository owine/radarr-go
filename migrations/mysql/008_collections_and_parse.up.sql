-- Collections and Parse/Rename System Migration for MySQL
-- Adds collections management and parse/rename functionality

-- Create collections table
CREATE TABLE IF NOT EXISTS collections (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    clean_title VARCHAR(500),
    sort_title VARCHAR(500),
    tmdb_id INT UNIQUE NOT NULL,
    overview TEXT,
    monitored BOOLEAN DEFAULT TRUE,
    quality_profile_id INT NOT NULL DEFAULT 1,
    root_folder_path VARCHAR(500),
    search_on_add BOOLEAN DEFAULT FALSE,
    minimum_availability VARCHAR(20) DEFAULT 'announced',
    last_info_sync TIMESTAMP NULL,
    images TEXT DEFAULT ('[]'),
    tags TEXT DEFAULT ('[]'),
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_collections_tmdb_id (tmdb_id),
    INDEX idx_collections_monitored (monitored),
    INDEX idx_collections_quality_profile_id (quality_profile_id),
    INDEX idx_collections_title (title)
);

-- Add collection_tmdb_id column to movies table if it doesn't exist
SET @sql = CONCAT('ALTER TABLE movies ADD COLUMN collection_tmdb_id INT NULL');
SET @stmt = IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
     WHERE TABLE_SCHEMA = DATABASE()
     AND TABLE_NAME = 'movies'
     AND COLUMN_NAME = 'collection_tmdb_id') = 0,
    @sql,
    'SELECT 1'
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Create index for collection relationships
CREATE INDEX IF NOT EXISTS idx_movies_collection_tmdb_id ON movies(collection_tmdb_id);

-- Add foreign key constraint for collection relationships (optional, can be nullable)
SET @sql = CONCAT('ALTER TABLE movies ADD CONSTRAINT fk_movies_collection_tmdb_id FOREIGN KEY (collection_tmdb_id) REFERENCES collections(tmdb_id) ON DELETE SET NULL');
SET @stmt = IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
     AND CONSTRAINT_NAME = 'fk_movies_collection_tmdb_id'
     AND TABLE_NAME = 'movies') = 0,
    @sql,
    'SELECT 1'
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Create parse_cache table for caching parse results
CREATE TABLE IF NOT EXISTS parse_cache (
    id INT AUTO_INCREMENT PRIMARY KEY,
    release_title VARCHAR(1000) NOT NULL,
    parsed_info TEXT NOT NULL, -- JSON containing ParsedMovieInfo
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE INDEX idx_parse_cache_release_title (release_title),
    INDEX idx_parse_cache_created_at (created_at)
);

-- Add quality profile foreign key constraint for collections
SET @sql = CONCAT('ALTER TABLE collections ADD CONSTRAINT fk_collections_quality_profile_id FOREIGN KEY (quality_profile_id) REFERENCES quality_profiles(id) ON DELETE RESTRICT');
SET @stmt = IF(
    (SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS
     WHERE CONSTRAINT_SCHEMA = DATABASE()
     AND CONSTRAINT_NAME = 'fk_collections_quality_profile_id'
     AND TABLE_NAME = 'collections') = 0,
    @sql,
    'SELECT 1'
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
