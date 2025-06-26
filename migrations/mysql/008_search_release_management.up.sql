-- Search & Release Management System Migration for MySQL/MariaDB
-- This migration adds support for release searching and management

-- Releases table to store found releases from indexers
CREATE TABLE IF NOT EXISTS releases (
    id INT PRIMARY KEY AUTO_INCREMENT,
    guid VARCHAR(500) NOT NULL,
    title VARCHAR(500) NOT NULL,
    sort_title VARCHAR(500),
    overview TEXT,
    quality_id INT DEFAULT 1,
    quality_name VARCHAR(50) DEFAULT 'Unknown',
    quality_source VARCHAR(50) DEFAULT 'unknown',
    quality_resolution INT DEFAULT 0,
    quality_revision_version INT DEFAULT 1,
    quality_revision_real INT DEFAULT 0,
    quality_revision_is_repack BOOLEAN DEFAULT FALSE,
    quality JSON,
    quality_weight INT DEFAULT 0,
    age INT DEFAULT 0,
    age_hours DOUBLE DEFAULT 0,
    age_minutes DOUBLE DEFAULT 0,
    size BIGINT DEFAULT 0,
    indexer_id INT NOT NULL,
    movie_id INT NULL,
    imdb_id VARCHAR(20),
    tmdb_id INT,
    protocol VARCHAR(20) NOT NULL DEFAULT 'torrent',
    download_url VARCHAR(2000) NOT NULL,
    info_url VARCHAR(2000),
    comment_url VARCHAR(2000),
    seeders INT,
    leechers INT,
    peer_count INT DEFAULT 0,
    publish_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'available',
    source VARCHAR(20) NOT NULL DEFAULT 'search',
    release_info JSON,
    categories JSON,
    download_client_id INT NULL,
    rejection_reasons JSON,
    indexer_flags INT DEFAULT 0,
    scene_mapping BOOLEAN DEFAULT FALSE,
    magnet_url VARCHAR(2000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    grabbed_at TIMESTAMP NULL,
    failed_at TIMESTAMP NULL,
    
    -- Foreign key constraints
    FOREIGN KEY (indexer_id) REFERENCES indexers(id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL
);

-- Indexes for optimal query performance
CREATE INDEX idx_releases_guid_indexer ON releases(guid, indexer_id);
CREATE UNIQUE INDEX idx_releases_guid_indexer_unique ON releases(guid, indexer_id);
CREATE INDEX idx_releases_title ON releases(sort_title);
CREATE INDEX idx_releases_movie_id ON releases(movie_id);
CREATE INDEX idx_releases_indexer_id ON releases(indexer_id);
CREATE INDEX idx_releases_imdb_id ON releases(imdb_id);
CREATE INDEX idx_releases_tmdb_id ON releases(tmdb_id);
CREATE INDEX idx_releases_publish_date ON releases(publish_date);
CREATE INDEX idx_releases_status ON releases(status);
CREATE INDEX idx_releases_quality_weight ON releases(quality_weight);
CREATE INDEX idx_releases_created_at ON releases(created_at);