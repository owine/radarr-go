-- Phase 1 features: Indexers, Download Clients, Quality Profiles, Custom Formats, Notifications

-- Quality definitions table
CREATE TABLE IF NOT EXISTS quality_definitions (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    min_size REAL DEFAULT 0,
    max_size REAL DEFAULT 400,
    preferred_size REAL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Quality profiles table
CREATE TABLE IF NOT EXISTS quality_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    cutoff INTEGER NOT NULL,
    items TEXT,
    language TEXT DEFAULT 'english',
    upgrade_allowed INTEGER DEFAULT 1,
    min_format_score INTEGER DEFAULT 0,
    cutoff_format_score INTEGER DEFAULT 0,
    format_items TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Custom formats table
CREATE TABLE IF NOT EXISTS custom_formats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    include_custom_format_when_renaming INTEGER DEFAULT 0,
    specifications TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexers table
CREATE TABLE IF NOT EXISTS indexers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    base_url TEXT NOT NULL,
    api_key TEXT,
    username TEXT,
    password TEXT,
    categories TEXT,
    priority INTEGER DEFAULT 25,
    status TEXT DEFAULT 'enabled',
    settings TEXT,
    supports_search INTEGER DEFAULT 1,
    supports_rss INTEGER DEFAULT 1,
    download_client_id INTEGER,
    last_rss_sync DATETIME,
    enable_rss INTEGER DEFAULT 1,
    enable_automatic_search INTEGER DEFAULT 1,
    enable_interactive_search INTEGER DEFAULT 1,
    supports_redirect INTEGER DEFAULT 0,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id)
);

-- Download clients table
CREATE TABLE IF NOT EXISTS download_clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    protocol TEXT NOT NULL,
    host TEXT NOT NULL,
    port INTEGER DEFAULT 8080,
    username TEXT,
    password TEXT,
    api_key TEXT,
    category TEXT,
    recent_movie_priority TEXT DEFAULT 'Normal',
    older_movie_priority TEXT DEFAULT 'Normal',
    add_paused INTEGER DEFAULT 0,
    use_ssl INTEGER DEFAULT 0,
    enable INTEGER DEFAULT 1,
    remove_completed_downloads INTEGER DEFAULT 1,
    remove_failed_downloads INTEGER DEFAULT 1,
    priority INTEGER DEFAULT 1,
    settings TEXT,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Queue items table
CREATE TABLE IF NOT EXISTS queue_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id INTEGER,
    download_client_id INTEGER,
    download_id TEXT,
    title TEXT NOT NULL,
    size INTEGER DEFAULT 0,
    size_left INTEGER DEFAULT 0,
    status TEXT NOT NULL,
    tracked_download_status TEXT,
    status_messages TEXT,
    downloaded_info TEXT,
    error_message TEXT,
    time_left TEXT,
    estimated_completion_time DATETIME,
    protocol TEXT,
    output_path TEXT,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id)
);

-- Download history table
CREATE TABLE IF NOT EXISTS download_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    movie_id INTEGER,
    download_client_id INTEGER,
    source_title TEXT NOT NULL,
    date DATETIME NOT NULL,
    protocol TEXT NOT NULL,
    indexer_name TEXT,
    download_id TEXT,
    successful INTEGER NOT NULL,
    data TEXT,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id)
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    settings TEXT,
    tags TEXT,
    on_grab INTEGER DEFAULT 0,
    on_download INTEGER DEFAULT 0,
    on_upgrade INTEGER DEFAULT 0,
    on_rename INTEGER DEFAULT 0,
    on_movie_delete INTEGER DEFAULT 0,
    on_movie_file_delete INTEGER DEFAULT 0,
    on_health INTEGER DEFAULT 0,
    on_application_update INTEGER DEFAULT 0,
    include_health_warnings INTEGER DEFAULT 0,
    enable INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Notification history table
CREATE TABLE IF NOT EXISTS notification_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    notification_id INTEGER,
    movie_id INTEGER,
    event_type TEXT NOT NULL,
    subject TEXT,
    message TEXT,
    successful INTEGER NOT NULL,
    error_message TEXT,
    date DATETIME NOT NULL,
    FOREIGN KEY (notification_id) REFERENCES notifications(id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
);

-- Health checks table
CREATE TABLE IF NOT EXISTS health_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,
    type TEXT NOT NULL,
    message TEXT NOT NULL,
    wiki_url TEXT,
    status TEXT NOT NULL,
    time DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_quality_profiles_name ON quality_profiles(name);
CREATE INDEX IF NOT EXISTS idx_custom_formats_name ON custom_formats(name);
CREATE INDEX IF NOT EXISTS idx_indexers_status ON indexers(status);
CREATE INDEX IF NOT EXISTS idx_indexers_type ON indexers(type);
CREATE INDEX IF NOT EXISTS idx_download_clients_enable ON download_clients(enable);
CREATE INDEX IF NOT EXISTS idx_download_clients_protocol ON download_clients(protocol);
CREATE INDEX IF NOT EXISTS idx_queue_items_movie_id ON queue_items(movie_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_download_client_id ON queue_items(download_client_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_download_id ON queue_items(download_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_status ON queue_items(status);
CREATE INDEX IF NOT EXISTS idx_download_history_movie_id ON download_history(movie_id);
CREATE INDEX IF NOT EXISTS idx_download_history_date ON download_history(date);
CREATE INDEX IF NOT EXISTS idx_download_history_successful ON download_history(successful);
CREATE INDEX IF NOT EXISTS idx_notifications_enable ON notifications(enable);
CREATE INDEX IF NOT EXISTS idx_notification_history_notification_id ON notification_history(notification_id);
CREATE INDEX IF NOT EXISTS idx_notification_history_movie_id ON notification_history(movie_id);
CREATE INDEX IF NOT EXISTS idx_notification_history_date ON notification_history(date);
CREATE INDEX IF NOT EXISTS idx_health_checks_status ON health_checks(status);
CREATE INDEX IF NOT EXISTS idx_health_checks_source ON health_checks(source);

-- Insert default quality definitions
INSERT OR IGNORE INTO quality_definitions (id, title, weight, min_size, max_size) VALUES
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
INSERT OR IGNORE INTO quality_profiles (id, name, cutoff, items, language) VALUES
(1, 'Any', 20, '[{"quality":{"id":0,"name":"Unknown","source":"unknown","resolution":0},"items":[],"allowed":false},{"name":"WEB 480p","items":[{"quality":{"id":8,"name":"WEBDL-480p","source":"webdl","resolution":480},"items":[],"allowed":true},{"quality":{"id":12,"name":"WEBRip-480p","source":"webrip","resolution":480},"items":[],"allowed":true}],"allowed":true,"id":1000},{"name":"WEB 720p","items":[{"quality":{"id":5,"name":"WEBDL-720p","source":"webdl","resolution":720},"items":[],"allowed":true},{"quality":{"id":14,"name":"WEBRip-720p","source":"webrip","resolution":720},"items":[],"allowed":true}],"allowed":true,"id":1001},{"name":"WEB 1080p","items":[{"quality":{"id":3,"name":"WEBDL-1080p","source":"webdl","resolution":1080},"items":[],"allowed":true},{"quality":{"id":15,"name":"WEBRip-1080p","source":"webrip","resolution":1080},"items":[],"allowed":true}],"allowed":true,"id":1002}]', 'english');

-- Triggers to update timestamps
CREATE TRIGGER IF NOT EXISTS update_quality_profiles_updated_at
    AFTER UPDATE ON quality_profiles
    FOR EACH ROW
BEGIN
    UPDATE quality_profiles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_custom_formats_updated_at
    AFTER UPDATE ON custom_formats
    FOR EACH ROW
BEGIN
    UPDATE custom_formats SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_indexers_updated_at
    AFTER UPDATE ON indexers
    FOR EACH ROW
BEGIN
    UPDATE indexers SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_download_clients_updated_at
    AFTER UPDATE ON download_clients
    FOR EACH ROW
BEGIN
    UPDATE download_clients SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_queue_items_updated_at
    AFTER UPDATE ON queue_items
    FOR EACH ROW
BEGIN
    UPDATE queue_items SET updated = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_notifications_updated_at
    AFTER UPDATE ON notifications
    FOR EACH ROW
BEGIN
    UPDATE notifications SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;