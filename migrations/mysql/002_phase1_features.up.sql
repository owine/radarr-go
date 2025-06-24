-- Phase 1 features: Indexers, Download Clients, Quality Profiles, Custom Formats, Notifications (MySQL/MariaDB version)

-- Quality definitions table
CREATE TABLE IF NOT EXISTS quality_definitions (
    id INT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    weight INT NOT NULL DEFAULT 1,
    min_size DECIMAL(10,3) DEFAULT 0,
    max_size DECIMAL(10,3) DEFAULT 400,
    preferred_size DECIMAL(10,3),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quality profiles table
CREATE TABLE IF NOT EXISTS quality_profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    cutoff INT NOT NULL,
    items TEXT,
    language VARCHAR(50) DEFAULT 'english',
    upgrade_allowed TINYINT(1) DEFAULT 1,
    min_format_score INT DEFAULT 0,
    cutoff_format_score INT DEFAULT 0,
    format_items TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Custom formats table
CREATE TABLE IF NOT EXISTS custom_formats (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    include_custom_format_when_renaming TINYINT(1) DEFAULT 0,
    specifications TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Download clients table
CREATE TABLE IF NOT EXISTS download_clients (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    protocol VARCHAR(50) NOT NULL,
    host VARCHAR(255) NOT NULL,
    port INT DEFAULT 8080,
    username VARCHAR(255),
    password VARCHAR(255),
    api_key VARCHAR(255),
    category VARCHAR(100),
    recent_movie_priority VARCHAR(50) DEFAULT 'Normal',
    older_movie_priority VARCHAR(50) DEFAULT 'Normal',
    add_paused TINYINT(1) DEFAULT 0,
    use_ssl TINYINT(1) DEFAULT 0,
    enable TINYINT(1) DEFAULT 1,
    remove_completed_downloads TINYINT(1) DEFAULT 1,
    remove_failed_downloads TINYINT(1) DEFAULT 1,
    priority INT DEFAULT 1,
    settings TEXT,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Indexers table
CREATE TABLE IF NOT EXISTS indexers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    api_key VARCHAR(255),
    username VARCHAR(255),
    password VARCHAR(255),
    categories TEXT,
    priority INT DEFAULT 25,
    status VARCHAR(50) DEFAULT 'enabled',
    settings TEXT,
    supports_search TINYINT(1) DEFAULT 1,
    supports_rss TINYINT(1) DEFAULT 1,
    download_client_id INT,
    last_rss_sync DATETIME,
    enable_rss TINYINT(1) DEFAULT 1,
    enable_automatic_search TINYINT(1) DEFAULT 1,
    enable_interactive_search TINYINT(1) DEFAULT 1,
    supports_redirect TINYINT(1) DEFAULT 0,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Queue items table
CREATE TABLE IF NOT EXISTS queue_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    download_client_id INT,
    download_id VARCHAR(255),
    title VARCHAR(500) NOT NULL,
    size BIGINT DEFAULT 0,
    size_left BIGINT DEFAULT 0,
    status VARCHAR(100) NOT NULL,
    tracked_download_status VARCHAR(100),
    status_messages TEXT,
    downloaded_info TEXT,
    error_message TEXT,
    time_left VARCHAR(100),
    estimated_completion_time DATETIME,
    protocol VARCHAR(50),
    output_path TEXT,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Download history table
CREATE TABLE IF NOT EXISTS download_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    download_client_id INT,
    source_title VARCHAR(500) NOT NULL,
    date DATETIME NOT NULL,
    protocol VARCHAR(50) NOT NULL,
    indexer_name VARCHAR(255),
    download_id VARCHAR(255),
    successful TINYINT(1) NOT NULL,
    data TEXT,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    settings TEXT,
    tags TEXT,
    on_grab TINYINT(1) DEFAULT 0,
    on_download TINYINT(1) DEFAULT 0,
    on_upgrade TINYINT(1) DEFAULT 0,
    on_rename TINYINT(1) DEFAULT 0,
    on_movie_delete TINYINT(1) DEFAULT 0,
    on_movie_file_delete TINYINT(1) DEFAULT 0,
    on_health TINYINT(1) DEFAULT 0,
    on_application_update TINYINT(1) DEFAULT 0,
    include_health_warnings TINYINT(1) DEFAULT 0,
    enable TINYINT(1) DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Notification history table
CREATE TABLE IF NOT EXISTS notification_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    notification_id INT,
    movie_id INT,
    event_type VARCHAR(100) NOT NULL,
    subject VARCHAR(500),
    message TEXT,
    successful TINYINT(1) NOT NULL,
    error_message TEXT,
    date DATETIME NOT NULL,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Health checks table
CREATE TABLE IF NOT EXISTS health_checks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    wiki_url VARCHAR(500),
    status VARCHAR(50) NOT NULL,
    time DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create indexes for better performance
CREATE INDEX idx_quality_profiles_name ON quality_profiles(name);
CREATE INDEX idx_custom_formats_name ON custom_formats(name);
CREATE INDEX idx_indexers_status ON indexers(status);
CREATE INDEX idx_indexers_type ON indexers(type);
CREATE INDEX idx_download_clients_enable ON download_clients(enable);
CREATE INDEX idx_download_clients_protocol ON download_clients(protocol);
CREATE INDEX idx_queue_items_movie_id ON queue_items(movie_id);
CREATE INDEX idx_queue_items_download_client_id ON queue_items(download_client_id);
CREATE INDEX idx_queue_items_download_id ON queue_items(download_id);
CREATE INDEX idx_queue_items_status ON queue_items(status);
CREATE INDEX idx_download_history_movie_id ON download_history(movie_id);
CREATE INDEX idx_download_history_date ON download_history(date);
CREATE INDEX idx_download_history_successful ON download_history(successful);
CREATE INDEX idx_notifications_enable ON notifications(enable);
CREATE INDEX idx_notification_history_notification_id ON notification_history(notification_id);
CREATE INDEX idx_notification_history_movie_id ON notification_history(movie_id);
CREATE INDEX idx_notification_history_date ON notification_history(date);
CREATE INDEX idx_health_checks_status ON health_checks(status);
CREATE INDEX idx_health_checks_source ON health_checks(source);

-- Insert default quality definitions
INSERT IGNORE INTO quality_definitions (id, title, weight, min_size, max_size) VALUES
(0, 'Unknown', 1, 0, 199.9),
(24, 'WORKPRINT', 2, 0, 199.9),
(25, 'CAM', 3, 0, 199.9),
(26, 'TELESYNC', 4, 0, 199.9),
(27, 'TELECINE', 5, 0, 199.9),
(29, 'REGIONAL', 6, 0, 199.9),
(28, 'DVDSCR', 7, 0, 199.9),
(1, 'SDTV', 8, 0, 199.9),
(2, 'DVD', 9, 0, 199.9),
(23, 'DVD-R', 10, 0, 199.9),
(8, 'WEBDL-480p', 11, 0, 199.9),
(12, 'WEBRip-480p', 12, 0, 199.9),
(20, 'Bluray-480p', 13, 0, 199.9),
(21, 'Bluray-576p', 14, 0, 199.9),
(4, 'HDTV-720p', 15, 0.8, 137.3),
(5, 'WEBDL-720p', 16, 0.8, 137.3),
(14, 'WEBRip-720p', 17, 0.8, 137.3),
(6, 'Bluray-720p', 18, 4.3, 137.3),
(9, 'HDTV-1080p', 19, 2, 137.3),
(3, 'WEBDL-1080p', 20, 2, 137.3),
(15, 'WEBRip-1080p', 21, 2, 137.3),
(7, 'Bluray-1080p', 22, 4.3, 258.1),
(30, 'Remux-1080p', 23, 0, 0),
(16, 'HDTV-2160p', 24, 4.7, 199.9),
(18, 'WEBDL-2160p', 25, 4.7, 258.1),
(17, 'WEBRip-2160p', 26, 4.7, 258.1),
(19, 'Bluray-2160p', 27, 4.3, 258.1),
(31, 'Remux-2160p', 28, 0, 0);

-- Insert default quality profile
INSERT IGNORE INTO quality_profiles (id, name, cutoff, items, language) VALUES
(1, 'Any', 20, '[{"quality":{"id":0,"name":"Unknown","source":"unknown","resolution":0},"items":[],"allowed":false},{"name":"WEB 480p","items":[{"quality":{"id":8,"name":"WEBDL-480p","source":"webdl","resolution":480},"items":[],"allowed":true},{"quality":{"id":12,"name":"WEBRip-480p","source":"webrip","resolution":480},"items":[],"allowed":true}],"allowed":true,"id":1000},{"name":"WEB 720p","items":[{"quality":{"id":5,"name":"WEBDL-720p","source":"webdl","resolution":720},"items":[],"allowed":true},{"quality":{"id":14,"name":"WEBRip-720p","source":"webrip","resolution":720},"items":[],"allowed":true}],"allowed":true,"id":1001},{"name":"WEB 1080p","items":[{"quality":{"id":3,"name":"WEBDL-1080p","source":"webdl","resolution":1080},"items":[],"allowed":true},{"quality":{"id":15,"name":"WEBRip-1080p","source":"webrip","resolution":1080},"items":[],"allowed":true}],"allowed":true,"id":1002}]', 'english');