-- Queue system migration for MySQL/MariaDB

CREATE TABLE IF NOT EXISTS queue_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    languages TEXT,
    quality TEXT,
    size BIGINT DEFAULT 0,
    title VARCHAR(500) NOT NULL,
    size_left BIGINT DEFAULT 0,
    time_left BIGINT,  -- Store duration as nanoseconds
    estimated_completion_time DATETIME,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'queued',
    tracked_download_status VARCHAR(50),
    tracked_download_state VARCHAR(50),
    status_messages TEXT,
    download_id VARCHAR(255) NOT NULL,
    protocol VARCHAR(20) DEFAULT 'unknown',
    download_client VARCHAR(100),
    download_client_has_post_import_category TINYINT(1) DEFAULT 0,
    indexer VARCHAR(100),
    output_path TEXT,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes for better performance
CREATE INDEX idx_queue_items_movie_id ON queue_items(movie_id);
CREATE INDEX idx_queue_items_download_id ON queue_items(download_id);
CREATE INDEX idx_queue_items_status ON queue_items(status);
CREATE INDEX idx_queue_items_protocol ON queue_items(protocol);
CREATE INDEX idx_queue_items_added ON queue_items(added);