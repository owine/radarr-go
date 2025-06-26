-- Download clients migration for MySQL/MariaDB

CREATE TABLE IF NOT EXISTS download_clients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL DEFAULT 8080,
    username VARCHAR(255),
    password VARCHAR(255),
    api_key VARCHAR(255),
    category VARCHAR(100),
    recent_movie_priority VARCHAR(20) DEFAULT 'Normal',
    older_movie_priority VARCHAR(20) DEFAULT 'Normal',
    add_paused BOOLEAN DEFAULT FALSE,
    use_ssl BOOLEAN DEFAULT FALSE,
    enable BOOLEAN DEFAULT TRUE,
    remove_completed_downloads BOOLEAN DEFAULT TRUE,
    remove_failed_downloads BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 1,
    fields TEXT,
    tags TEXT,
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_download_clients_name (name),
    INDEX idx_download_clients_protocol (protocol),
    INDEX idx_download_clients_enable (enable),
    INDEX idx_download_clients_priority (priority)
);

-- Add download_client_id column to queue_items table
ALTER TABLE queue_items 
ADD COLUMN download_client_id INT,
ADD CONSTRAINT fk_queue_items_download_client 
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL;

-- Add downloaded_info column to queue_items table
ALTER TABLE queue_items 
ADD COLUMN downloaded_info TEXT;

-- Create download_history table
CREATE TABLE IF NOT EXISTS download_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    download_client_id INT,
    source_title VARCHAR(500) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    indexer_name VARCHAR(255),
    download_id VARCHAR(255),
    successful BOOLEAN NOT NULL DEFAULT FALSE,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_download_history_movie_id (movie_id),
    INDEX idx_download_history_download_client_id (download_client_id),
    INDEX idx_download_history_date (date),
    INDEX idx_download_history_successful (successful),
    INDEX idx_download_history_protocol (protocol),
    CONSTRAINT fk_download_history_movie 
        FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    CONSTRAINT fk_download_history_download_client 
        FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL
);

-- Create index for queue_items download_client_id
CREATE INDEX idx_queue_items_download_client_id ON queue_items(download_client_id);