-- History and Activity tracking migration for MySQL/MariaDB

-- Create history table
CREATE TABLE IF NOT EXISTS history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    event_type VARCHAR(50) NOT NULL,
    date TIMESTAMP NOT NULL,
    quality TEXT,
    source_title VARCHAR(500),
    language TEXT,
    download_id VARCHAR(100),
    data TEXT,
    message TEXT,
    successful BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_history_movie_id (movie_id),
    INDEX idx_history_event_type (event_type),
    INDEX idx_history_date (date),
    INDEX idx_history_download_id (download_id),
    INDEX idx_history_successful (successful),
    CONSTRAINT fk_history_movie FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
);

-- Create activity table
CREATE TABLE IF NOT EXISTS activity (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    movie_id INT,
    progress DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NULL,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_activity_type (type),
    INDEX idx_activity_status (status),
    INDEX idx_activity_movie_id (movie_id),
    INDEX idx_activity_start_time (start_time),
    CONSTRAINT fk_activity_movie FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
);