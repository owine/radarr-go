-- Complete Radarr Go Database Schema for MySQL/MariaDB
-- This migration contains the complete schema for all Phase 1 features

-- Core movies table
CREATE TABLE IF NOT EXISTS movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tmdb_id INT UNIQUE NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    title_slug VARCHAR(500) UNIQUE NOT NULL,
    original_title VARCHAR(500),
    original_language VARCHAR(10),
    overview TEXT,
    website VARCHAR(500),
    in_cinemas DATETIME,
    physical_release DATETIME,
    digital_release DATETIME,
    release_date DATETIME,
    year INT,
    runtime INT,
    images TEXT,
    genres TEXT,
    tags TEXT,
    certification VARCHAR(20),
    ratings TEXT,
    movie_file_id INT,
    quality_profile_id INT NOT NULL DEFAULT 1,
    path VARCHAR(500),
    root_folder_path VARCHAR(500),
    folder_name VARCHAR(255),
    monitored TINYINT(1) DEFAULT 1,
    minimum_availability VARCHAR(20) DEFAULT 'announced',
    has_file TINYINT(1) DEFAULT 0,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    studio VARCHAR(255),
    youtube_trailer_id VARCHAR(50),
    last_info_sync DATETIME,
    status VARCHAR(20) DEFAULT 'tba',
    collection_tmdb_id INT,
    collection_title VARCHAR(500),
    secondary_year INT,
    secondary_year_source_id INT DEFAULT 0,
    sort_title VARCHAR(500),
    size_on_disk BIGINT DEFAULT 0,
    popularity DOUBLE DEFAULT 0.0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Movie files table
CREATE TABLE IF NOT EXISTS movie_files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    relative_path VARCHAR(500) NOT NULL,
    path VARCHAR(500) NOT NULL,
    size BIGINT NOT NULL,
    date_added DATETIME DEFAULT CURRENT_TIMESTAMP,
    scene_name VARCHAR(500),
    media_info TEXT,
    quality TEXT,
    language TEXT,
    release_group VARCHAR(255),
    edition VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    original_file_path VARCHAR(500),
    indexer_flags INT DEFAULT 0,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Quality profiles table
CREATE TABLE IF NOT EXISTS quality_profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    cutoff INT NOT NULL,
    items TEXT NOT NULL,
    min_format_score INT DEFAULT 0,
    cutoff_format_score INT DEFAULT 0,
    format_items TEXT,
    upgrade_allowed TINYINT(1) DEFAULT 1,
    language_id INT DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Indexers table
CREATE TABLE IF NOT EXISTS indexers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    settings TEXT,
    config_contract VARCHAR(100),
    enable_rss TINYINT(1) DEFAULT 1,
    enable_automatic_search TINYINT(1) DEFAULT 1,
    enable_interactive_search TINYINT(1) DEFAULT 1,
    supports_rss TINYINT(1) DEFAULT 1,
    supports_search TINYINT(1) DEFAULT 1,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    priority INT DEFAULT 25,
    season_search_max_page_size INT DEFAULT 100,
    download_client_id INT DEFAULT 0,
    redirect_to_magnet TINYINT(1) DEFAULT 0,
    tags TEXT,
    fields TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Queue items table - CONSOLIDATED VERSION
CREATE TABLE IF NOT EXISTS queue_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    title VARCHAR(500) NOT NULL,
    size BIGINT DEFAULT 0,
    sizeleft BIGINT DEFAULT 0,
    timeleft TIME,
    estimated_completion_time DATETIME,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    tracked_download_status VARCHAR(50) DEFAULT 'ok',
    tracked_download_state VARCHAR(50) DEFAULT 'downloading',
    status_messages TEXT,
    error_message TEXT,
    download_id VARCHAR(255),
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    indexer_name VARCHAR(255),
    output_path VARCHAR(500),
    download_client_id INT,
    downloaded_info TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Download clients table
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
    add_paused TINYINT(1) DEFAULT 0,
    use_ssl TINYINT(1) DEFAULT 0,
    enable TINYINT(1) DEFAULT 1,
    remove_completed_downloads TINYINT(1) DEFAULT 1,
    remove_failed_downloads TINYINT(1) DEFAULT 1,
    priority INT DEFAULT 1,
    fields TEXT,
    tags TEXT,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Download history table
CREATE TABLE IF NOT EXISTS download_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    download_client_id INT,
    source_title VARCHAR(500) NOT NULL,
    date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    indexer_name VARCHAR(255),
    download_id VARCHAR(255),
    successful TINYINT(1) NOT NULL DEFAULT 0,
    data TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Import lists table
CREATE TABLE IF NOT EXISTS import_lists (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT,
    enable_auto TINYINT(1) DEFAULT 1,
    enabled TINYINT(1) DEFAULT 1,
    enable_interactive TINYINT(1) DEFAULT 0,
    list_type VARCHAR(20) DEFAULT 'program',
    list_order INT DEFAULT 0,
    min_refresh_interval BIGINT DEFAULT 1440,
    quality_profile_id INT NOT NULL,
    root_folder_path VARCHAR(500) NOT NULL,
    should_monitor TINYINT(1) DEFAULT 1,
    minimum_availability VARCHAR(20) DEFAULT 'released',
    tags TEXT,
    fields TEXT,
    last_sync DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Import list movies table
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
    in_cinemas DATETIME,
    physical_release DATETIME,
    digital_release DATETIME,
    website VARCHAR(500),
    youtube_trailer_id VARCHAR(50),
    studio VARCHAR(255),
    minimum_availability VARCHAR(20),
    is_excluded TINYINT(1) DEFAULT 0,
    is_existing TINYINT(1) DEFAULT 0,
    is_recommendation TINYINT(1) DEFAULT 0,
    list_position INT DEFAULT 0,
    discovered_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (import_list_id) REFERENCES import_lists(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Import list exclusions table
CREATE TABLE IF NOT EXISTS import_list_exclusions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tmdb_id INT NOT NULL UNIQUE,
    movie_title VARCHAR(500) NOT NULL,
    movie_year INT NOT NULL,
    imdb_id VARCHAR(20),
    reason VARCHAR(255),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- History table for activity tracking
CREATE TABLE IF NOT EXISTS history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT,
    event_type VARCHAR(50) NOT NULL,
    date DATETIME NOT NULL,
    quality TEXT,
    source_title VARCHAR(500),
    language TEXT,
    download_id VARCHAR(100),
    data TEXT,
    message TEXT,
    successful TINYINT(1) DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Activity table for real-time activity tracking
CREATE TABLE IF NOT EXISTS activity (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    movie_id INT,
    progress DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    data TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Host configuration table
CREATE TABLE IF NOT EXISTS host_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bind_address VARCHAR(255) NOT NULL DEFAULT '*',
    port INT NOT NULL DEFAULT 7878,
    url_base VARCHAR(255) DEFAULT '',
    enable_ssl TINYINT(1) DEFAULT 0,
    ssl_port INT DEFAULT 6969,
    ssl_cert_path VARCHAR(500) DEFAULT '',
    ssl_key_path VARCHAR(500) DEFAULT '',
    username VARCHAR(255) DEFAULT '',
    password VARCHAR(255) DEFAULT '',
    authentication_method VARCHAR(50) DEFAULT 'none',
    authentication_required VARCHAR(50) DEFAULT 'enabled',
    log_level VARCHAR(20) DEFAULT 'info',
    launch_browser TINYINT(1) DEFAULT 1,
    enable_color_impared TINYINT(1) DEFAULT 0,
    proxy_settings TEXT,
    update_mechanism VARCHAR(50) DEFAULT 'builtin',
    update_branch VARCHAR(100) DEFAULT 'master',
    update_automatically TINYINT(1) DEFAULT 0,
    update_script_path VARCHAR(500) DEFAULT '',
    analytics_enabled TINYINT(1) DEFAULT 1,
    backup_folder VARCHAR(500) DEFAULT '',
    backup_interval INT DEFAULT 7,
    backup_retention INT DEFAULT 28,
    certificate_validation VARCHAR(50) DEFAULT 'enabled',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Naming configuration table
CREATE TABLE IF NOT EXISTS naming_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    rename_movies TINYINT(1) DEFAULT 0,
    replace_illegal_characters TINYINT(1) DEFAULT 1,
    colon_replacement_format VARCHAR(50) DEFAULT 'delete',
    standard_movie_format VARCHAR(500) NOT NULL DEFAULT '{Movie Title} ({Release Year}) {Quality Full}',
    movie_folder_format VARCHAR(500) NOT NULL DEFAULT '{Movie Title} ({Release Year})',
    create_empty_folders TINYINT(1) DEFAULT 0,
    delete_empty_folders TINYINT(1) DEFAULT 0,
    skip_free_space_check TINYINT(1) DEFAULT 0,
    minimum_free_space BIGINT DEFAULT 100,
    use_hardlinks TINYINT(1) DEFAULT 1,
    import_extra_files TINYINT(1) DEFAULT 0,
    extra_file_extensions TEXT,
    enable_media_info TINYINT(1) DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Media management configuration table
CREATE TABLE IF NOT EXISTS media_management_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    auto_unmonitor_previous_movie TINYINT(1) DEFAULT 0,
    recycle_bin VARCHAR(500) DEFAULT '',
    recycle_bin_cleanup INT DEFAULT 7,
    download_propers_and_repacks VARCHAR(50) DEFAULT 'preferAndUpgrade',
    create_empty_folders TINYINT(1) DEFAULT 0,
    delete_empty_folders TINYINT(1) DEFAULT 0,
    file_date VARCHAR(50) DEFAULT 'none',
    rescan_after_refresh VARCHAR(50) DEFAULT 'always',
    allow_fingerprinting VARCHAR(50) DEFAULT 'newFiles',
    set_permissions TINYINT(1) DEFAULT 0,
    chmod_folder VARCHAR(10) DEFAULT '755',
    chown_group VARCHAR(100) DEFAULT '',
    skip_free_space_check TINYINT(1) DEFAULT 0,
    minimum_free_space BIGINT DEFAULT 100,
    copy_using_hardlinks TINYINT(1) DEFAULT 1,
    use_script_import TINYINT(1) DEFAULT 0,
    script_import_path VARCHAR(500) DEFAULT '',
    import_extra_files TINYINT(1) DEFAULT 0,
    extra_file_extensions TEXT,
    enable_media_info TINYINT(1) DEFAULT 1,
    import_mechanism VARCHAR(50) DEFAULT 'move',
    watch_library_for_changes TINYINT(1) DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Root folders table
CREATE TABLE IF NOT EXISTS root_folders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    path VARCHAR(500) NOT NULL UNIQUE,
    accessible TINYINT(1) DEFAULT 1,
    free_space BIGINT DEFAULT 0,
    total_space BIGINT DEFAULT 0,
    unmapped_folders TEXT,
    default_tags TEXT,
    default_quality_profile_id INT DEFAULT 0,
    default_monitor_option VARCHAR(50) DEFAULT 'movieOnly',
    default_search_for_missing_movie TINYINT(1) DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Releases table for search and release management
CREATE TABLE IF NOT EXISTS releases (
    id INT AUTO_INCREMENT PRIMARY KEY,
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
    quality_revision_is_repack TINYINT(1) DEFAULT 0,
    quality JSON,
    quality_weight INT DEFAULT 0,
    age INT DEFAULT 0,
    age_hours DOUBLE DEFAULT 0,
    age_minutes DOUBLE DEFAULT 0,
    size BIGINT DEFAULT 0,
    indexer_id INT NOT NULL,
    movie_id INT,
    imdb_id VARCHAR(20),
    tmdb_id INT,
    protocol VARCHAR(20) NOT NULL DEFAULT 'torrent',
    download_url VARCHAR(2000) NOT NULL,
    info_url VARCHAR(2000),
    comment_url VARCHAR(2000),
    seeders INT,
    leechers INT,
    peer_count INT DEFAULT 0,
    publish_date DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'available',
    source VARCHAR(20) NOT NULL DEFAULT 'search',
    release_info JSON,
    categories JSON,
    download_client_id INT,
    rejection_reasons JSON,
    indexer_flags INT DEFAULT 0,
    scene_mapping TINYINT(1) DEFAULT 0,
    magnet_url VARCHAR(2000),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    grabbed_at DATETIME,
    failed_at DATETIME,
    FOREIGN KEY (indexer_id) REFERENCES indexers(id) ON DELETE CASCADE,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT,
    on_grab TINYINT(1) DEFAULT 0,
    on_import TINYINT(1) DEFAULT 0,
    on_upgrade TINYINT(1) DEFAULT 1,
    on_rename TINYINT(1) DEFAULT 0,
    on_movie_delete TINYINT(1) DEFAULT 0,
    on_movie_file_delete TINYINT(1) DEFAULT 0,
    on_movie_file_delete_for_upgrade TINYINT(1) DEFAULT 1,
    on_health_issue TINYINT(1) DEFAULT 0,
    on_health_restored TINYINT(1) DEFAULT 0,
    on_application_update TINYINT(1) DEFAULT 0,
    on_manual_interaction_required TINYINT(1) DEFAULT 0,
    include_health_warnings TINYINT(1) DEFAULT 0,
    enabled TINYINT(1) DEFAULT 1,
    tags TEXT,
    fields TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create all indexes for optimal performance

-- Movies indexes
CREATE INDEX idx_movies_tmdb_id ON movies(tmdb_id);
CREATE INDEX idx_movies_imdb_id ON movies(imdb_id);
CREATE INDEX idx_movies_title_slug ON movies(title_slug);
CREATE INDEX idx_movies_monitored ON movies(monitored);
CREATE INDEX idx_movies_has_file ON movies(has_file);
CREATE INDEX idx_movies_quality_profile_id ON movies(quality_profile_id);
CREATE INDEX idx_movies_year ON movies(year);
CREATE INDEX idx_movies_in_cinemas ON movies(in_cinemas);
CREATE INDEX idx_movies_physical_release ON movies(physical_release);

-- Movie files indexes
CREATE INDEX idx_movie_files_movie_id ON movie_files(movie_id);
CREATE INDEX idx_movie_files_path ON movie_files(path);

-- Quality profiles indexes
CREATE INDEX idx_quality_profiles_name ON quality_profiles(name);

-- Indexers indexes
CREATE INDEX idx_indexers_name ON indexers(name);
CREATE INDEX idx_indexers_implementation ON indexers(implementation);
CREATE INDEX idx_indexers_enable_rss ON indexers(enable_rss);
CREATE INDEX idx_indexers_enable_automatic_search ON indexers(enable_automatic_search);
CREATE INDEX idx_indexers_protocol ON indexers(protocol);

-- Queue items indexes
CREATE INDEX idx_queue_items_movie_id ON queue_items(movie_id);
CREATE INDEX idx_queue_items_status ON queue_items(status);
CREATE INDEX idx_queue_items_download_id ON queue_items(download_id);
CREATE INDEX idx_queue_items_download_client_id ON queue_items(download_client_id);

-- Download clients indexes
CREATE INDEX idx_download_clients_name ON download_clients(name);
CREATE INDEX idx_download_clients_protocol ON download_clients(protocol);
CREATE INDEX idx_download_clients_enable ON download_clients(enable);
CREATE INDEX idx_download_clients_priority ON download_clients(priority);

-- Download history indexes
CREATE INDEX idx_download_history_movie_id ON download_history(movie_id);
CREATE INDEX idx_download_history_download_client_id ON download_history(download_client_id);
CREATE INDEX idx_download_history_date ON download_history(date);
CREATE INDEX idx_download_history_successful ON download_history(successful);
CREATE INDEX idx_download_history_protocol ON download_history(protocol);

-- Import lists indexes
CREATE INDEX idx_import_lists_name ON import_lists(name);
CREATE INDEX idx_import_lists_implementation ON import_lists(implementation);
CREATE INDEX idx_import_lists_enabled ON import_lists(enabled);
CREATE INDEX idx_import_lists_enable_auto ON import_lists(enable_auto);
CREATE INDEX idx_import_lists_quality_profile ON import_lists(quality_profile_id);

-- Import list movies indexes
CREATE INDEX idx_import_list_movies_import_list_id ON import_list_movies(import_list_id);
CREATE INDEX idx_import_list_movies_tmdb_id ON import_list_movies(tmdb_id);
CREATE INDEX idx_import_list_movies_imdb_id ON import_list_movies(imdb_id);
CREATE INDEX idx_import_list_movies_year ON import_list_movies(year);
CREATE INDEX idx_import_list_movies_is_excluded ON import_list_movies(is_excluded);
CREATE INDEX idx_import_list_movies_is_existing ON import_list_movies(is_existing);
CREATE INDEX idx_import_list_movies_is_recommendation ON import_list_movies(is_recommendation);
CREATE INDEX idx_import_list_movies_discovered_at ON import_list_movies(discovered_at);

-- Import list exclusions indexes
CREATE INDEX idx_import_list_exclusions_tmdb_id ON import_list_exclusions(tmdb_id);
CREATE INDEX idx_import_list_exclusions_imdb_id ON import_list_exclusions(imdb_id);

-- History indexes
CREATE INDEX idx_history_movie_id ON history(movie_id);
CREATE INDEX idx_history_event_type ON history(event_type);
CREATE INDEX idx_history_date ON history(date);
CREATE INDEX idx_history_download_id ON history(download_id);
CREATE INDEX idx_history_successful ON history(successful);

-- Activity indexes
CREATE INDEX idx_activity_type ON activity(type);
CREATE INDEX idx_activity_status ON activity(status);
CREATE INDEX idx_activity_movie_id ON activity(movie_id);
CREATE INDEX idx_activity_start_time ON activity(start_time);

-- Configuration indexes
CREATE INDEX idx_host_config_auth ON host_config(authentication_method);
CREATE INDEX idx_naming_config_rename ON naming_config(rename_movies);
CREATE INDEX idx_media_config_recycle ON media_management_config(recycle_bin);
CREATE INDEX idx_root_folders_accessible ON root_folders(accessible);
CREATE INDEX idx_root_folders_path ON root_folders(path);

-- Releases indexes
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

-- Notifications indexes
CREATE INDEX idx_notifications_name ON notifications(name);
CREATE INDEX idx_notifications_implementation ON notifications(implementation);
CREATE INDEX idx_notifications_enabled ON notifications(enabled);

-- Insert default data

-- Insert default quality profile
INSERT INTO quality_profiles (id, name, cutoff, items) VALUES 
(1, 'Any', 1, '[{"quality": {"id": 1, "name": "Unknown"}, "allowed": true}]')
ON DUPLICATE KEY UPDATE id=id;

-- Insert default host configuration
INSERT INTO host_config (id) VALUES (1) ON DUPLICATE KEY UPDATE id=id;

-- Insert default naming configuration
INSERT INTO naming_config (id) VALUES (1) ON DUPLICATE KEY UPDATE id=id;

-- Insert default media management configuration
INSERT INTO media_management_config (id) VALUES (1) ON DUPLICATE KEY UPDATE id=id;