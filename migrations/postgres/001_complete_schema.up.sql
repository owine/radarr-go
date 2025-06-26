-- Complete Radarr Go Database Schema for PostgreSQL
-- This migration contains the complete schema for all Phase 1 features

-- Core movies table
CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER UNIQUE NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    title_slug VARCHAR(500) UNIQUE NOT NULL,
    original_title VARCHAR(500),
    original_language VARCHAR(10),
    overview TEXT,
    website VARCHAR(500),
    in_cinemas TIMESTAMP,
    physical_release TIMESTAMP,
    digital_release TIMESTAMP,
    release_date TIMESTAMP,
    year INTEGER,
    runtime INTEGER,
    images TEXT DEFAULT '[]'::TEXT,
    genres TEXT DEFAULT '[]'::TEXT,
    tags TEXT DEFAULT '[]'::TEXT,
    certification VARCHAR(20),
    ratings TEXT DEFAULT '{}'::TEXT,
    movie_file_id INTEGER,
    quality_profile_id INTEGER NOT NULL DEFAULT 1,
    path VARCHAR(500),
    root_folder_path VARCHAR(500),
    folder_name VARCHAR(255),
    monitored BOOLEAN DEFAULT TRUE,
    minimum_availability VARCHAR(20) DEFAULT 'announced',
    has_file BOOLEAN DEFAULT FALSE,
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    studio VARCHAR(255),
    youtube_trailer_id VARCHAR(50),
    last_info_sync TIMESTAMP,
    status VARCHAR(20) DEFAULT 'tba',
    collection_tmdb_id INTEGER,
    collection_title VARCHAR(500),
    secondary_year INTEGER,
    secondary_year_source_id INTEGER DEFAULT 0,
    sort_title VARCHAR(500),
    size_on_disk BIGINT DEFAULT 0,
    popularity DOUBLE PRECISION DEFAULT 0.0
);

-- Movie files table
CREATE TABLE IF NOT EXISTS movie_files (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    relative_path VARCHAR(500) NOT NULL,
    path VARCHAR(500) NOT NULL,
    size BIGINT NOT NULL,
    date_added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    scene_name VARCHAR(500),
    media_info TEXT DEFAULT '{}'::TEXT,
    quality TEXT DEFAULT '{}'::TEXT,
    language TEXT DEFAULT '[]'::TEXT,
    release_group VARCHAR(255),
    edition VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    original_file_path VARCHAR(500),
    indexer_flags INTEGER DEFAULT 0
);

-- Quality profiles table
CREATE TABLE IF NOT EXISTS quality_profiles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    cutoff INTEGER NOT NULL,
    items TEXT NOT NULL DEFAULT '[]'::TEXT,
    min_format_score INTEGER DEFAULT 0,
    cutoff_format_score INTEGER DEFAULT 0,
    format_items TEXT DEFAULT '[]'::TEXT,
    upgrade_allowed BOOLEAN DEFAULT TRUE,
    language_id INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexers table
CREATE TABLE IF NOT EXISTS indexers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    settings TEXT DEFAULT '{}'::TEXT,
    config_contract VARCHAR(100),
    enable_rss BOOLEAN DEFAULT TRUE,
    enable_automatic_search BOOLEAN DEFAULT TRUE,
    enable_interactive_search BOOLEAN DEFAULT TRUE,
    supports_rss BOOLEAN DEFAULT TRUE,
    supports_search BOOLEAN DEFAULT TRUE,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    priority INTEGER DEFAULT 25,
    season_search_max_page_size INTEGER DEFAULT 100,
    download_client_id INTEGER DEFAULT 0,
    redirect_to_magnet BOOLEAN DEFAULT FALSE,
    tags TEXT DEFAULT '[]'::TEXT,
    fields TEXT DEFAULT '[]'::TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Download clients table
CREATE TABLE IF NOT EXISTS download_clients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    host VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 8080,
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
    priority INTEGER DEFAULT 1,
    fields TEXT DEFAULT '{}'::TEXT,
    tags TEXT DEFAULT '[]'::TEXT,
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Queue items table - CONSOLIDATED VERSION
CREATE TABLE IF NOT EXISTS queue_items (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    size BIGINT DEFAULT 0,
    sizeleft BIGINT DEFAULT 0,
    timeleft INTERVAL,
    estimated_completion_time TIMESTAMP,
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(20) NOT NULL DEFAULT 'queued',
    tracked_download_status VARCHAR(50) DEFAULT 'ok',
    tracked_download_state VARCHAR(50) DEFAULT 'downloading',
    status_messages TEXT DEFAULT '[]'::TEXT,
    error_message TEXT,
    download_id VARCHAR(255),
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    indexer_name VARCHAR(255),
    output_path VARCHAR(500),
    download_client_id INTEGER REFERENCES download_clients(id) ON DELETE SET NULL,
    downloaded_info TEXT DEFAULT '{}'::TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Download history table
CREATE TABLE IF NOT EXISTS download_history (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    download_client_id INTEGER REFERENCES download_clients(id) ON DELETE SET NULL,
    source_title VARCHAR(500) NOT NULL,
    date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    protocol VARCHAR(20) NOT NULL DEFAULT 'unknown',
    indexer_name VARCHAR(255),
    download_id VARCHAR(255),
    successful BOOLEAN NOT NULL DEFAULT FALSE,
    data TEXT DEFAULT '{}'::TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Import lists table
CREATE TABLE IF NOT EXISTS import_lists (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT DEFAULT '{}'::TEXT,
    enable_auto BOOLEAN DEFAULT TRUE,
    enabled BOOLEAN DEFAULT TRUE,
    enable_interactive BOOLEAN DEFAULT FALSE,
    list_type VARCHAR(20) DEFAULT 'program',
    list_order INTEGER DEFAULT 0,
    min_refresh_interval BIGINT DEFAULT 1440, -- minutes
    quality_profile_id INTEGER NOT NULL,
    root_folder_path VARCHAR(500) NOT NULL,
    should_monitor BOOLEAN DEFAULT TRUE,
    minimum_availability VARCHAR(20) DEFAULT 'released',
    tags TEXT DEFAULT '[]'::TEXT,
    fields TEXT DEFAULT '[]'::TEXT,
    last_sync TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Import list movies table
CREATE TABLE IF NOT EXISTS import_list_movies (
    id SERIAL PRIMARY KEY,
    import_list_id INTEGER NOT NULL REFERENCES import_lists(id) ON DELETE CASCADE,
    tmdb_id INTEGER NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    original_title VARCHAR(500),
    year INTEGER,
    overview TEXT,
    runtime INTEGER,
    images TEXT DEFAULT '[]'::TEXT,
    genres TEXT DEFAULT '[]'::TEXT,
    ratings TEXT DEFAULT '{}'::TEXT,
    certification VARCHAR(20),
    status VARCHAR(20),
    in_cinemas TIMESTAMP,
    physical_release TIMESTAMP,
    digital_release TIMESTAMP,
    website VARCHAR(500),
    youtube_trailer_id VARCHAR(50),
    studio VARCHAR(255),
    minimum_availability VARCHAR(20),
    is_excluded BOOLEAN DEFAULT FALSE,
    is_existing BOOLEAN DEFAULT FALSE,
    is_recommendation BOOLEAN DEFAULT FALSE,
    list_position INTEGER DEFAULT 0,
    discovered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Import list exclusions table
CREATE TABLE IF NOT EXISTS import_list_exclusions (
    id SERIAL PRIMARY KEY,
    tmdb_id INTEGER NOT NULL UNIQUE,
    movie_title VARCHAR(500) NOT NULL,
    movie_year INTEGER NOT NULL,
    imdb_id VARCHAR(20),
    reason VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- History table for activity tracking
CREATE TABLE IF NOT EXISTS history (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    date TIMESTAMP NOT NULL,
    quality TEXT,
    source_title VARCHAR(500),
    language TEXT,
    download_id VARCHAR(100),
    data TEXT,
    message TEXT,
    successful BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Activity table for real-time activity tracking
CREATE TABLE IF NOT EXISTS activity (
    id SERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    progress DECIMAL(5,2) DEFAULT 0,
    status VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Host configuration table
CREATE TABLE IF NOT EXISTS host_config (
    id SERIAL PRIMARY KEY,
    bind_address VARCHAR(255) NOT NULL DEFAULT '*',
    port INTEGER NOT NULL DEFAULT 7878,
    url_base VARCHAR(255) DEFAULT '',
    enable_ssl BOOLEAN DEFAULT FALSE,
    ssl_port INTEGER DEFAULT 6969,
    ssl_cert_path VARCHAR(500) DEFAULT '',
    ssl_key_path VARCHAR(500) DEFAULT '',
    username VARCHAR(255) DEFAULT '',
    password VARCHAR(255) DEFAULT '',
    authentication_method VARCHAR(50) DEFAULT 'none',
    authentication_required VARCHAR(50) DEFAULT 'enabled',
    log_level VARCHAR(20) DEFAULT 'info',
    launch_browser BOOLEAN DEFAULT TRUE,
    enable_color_impared BOOLEAN DEFAULT FALSE,
    proxy_settings TEXT,
    update_mechanism VARCHAR(50) DEFAULT 'builtin',
    update_branch VARCHAR(100) DEFAULT 'master',
    update_automatically BOOLEAN DEFAULT FALSE,
    update_script_path VARCHAR(500) DEFAULT '',
    analytics_enabled BOOLEAN DEFAULT TRUE,
    backup_folder VARCHAR(500) DEFAULT '',
    backup_interval INTEGER DEFAULT 7,
    backup_retention INTEGER DEFAULT 28,
    certificate_validation VARCHAR(50) DEFAULT 'enabled',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Naming configuration table
CREATE TABLE IF NOT EXISTS naming_config (
    id SERIAL PRIMARY KEY,
    rename_movies BOOLEAN DEFAULT FALSE,
    replace_illegal_characters BOOLEAN DEFAULT TRUE,
    colon_replacement_format VARCHAR(50) DEFAULT 'delete',
    standard_movie_format VARCHAR(500) NOT NULL DEFAULT '{Movie Title} ({Release Year}) {Quality Full}',
    movie_folder_format VARCHAR(500) NOT NULL DEFAULT '{Movie Title} ({Release Year})',
    create_empty_folders BOOLEAN DEFAULT FALSE,
    delete_empty_folders BOOLEAN DEFAULT FALSE,
    skip_free_space_check BOOLEAN DEFAULT FALSE,
    minimum_free_space BIGINT DEFAULT 100,
    use_hardlinks BOOLEAN DEFAULT TRUE,
    import_extra_files BOOLEAN DEFAULT FALSE,
    extra_file_extensions TEXT,
    enable_media_info BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Media management configuration table
CREATE TABLE IF NOT EXISTS media_management_config (
    id SERIAL PRIMARY KEY,
    auto_unmonitor_previous_movie BOOLEAN DEFAULT FALSE,
    recycle_bin VARCHAR(500) DEFAULT '',
    recycle_bin_cleanup INTEGER DEFAULT 7,
    download_propers_and_repacks VARCHAR(50) DEFAULT 'preferAndUpgrade',
    create_empty_folders BOOLEAN DEFAULT FALSE,
    delete_empty_folders BOOLEAN DEFAULT FALSE,
    file_date VARCHAR(50) DEFAULT 'none',
    rescan_after_refresh VARCHAR(50) DEFAULT 'always',
    allow_fingerprinting VARCHAR(50) DEFAULT 'newFiles',
    set_permissions BOOLEAN DEFAULT FALSE,
    chmod_folder VARCHAR(10) DEFAULT '755',
    chown_group VARCHAR(100) DEFAULT '',
    skip_free_space_check BOOLEAN DEFAULT FALSE,
    minimum_free_space BIGINT DEFAULT 100,
    copy_using_hardlinks BOOLEAN DEFAULT TRUE,
    use_script_import BOOLEAN DEFAULT FALSE,
    script_import_path VARCHAR(500) DEFAULT '',
    import_extra_files BOOLEAN DEFAULT FALSE,
    extra_file_extensions TEXT,
    enable_media_info BOOLEAN DEFAULT TRUE,
    import_mechanism VARCHAR(50) DEFAULT 'move',
    watch_library_for_changes BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Root folders table
CREATE TABLE IF NOT EXISTS root_folders (
    id SERIAL PRIMARY KEY,
    path VARCHAR(500) NOT NULL UNIQUE,
    accessible BOOLEAN DEFAULT TRUE,
    free_space BIGINT DEFAULT 0,
    total_space BIGINT DEFAULT 0,
    unmapped_folders TEXT,
    default_tags TEXT,
    default_quality_profile_id INTEGER DEFAULT 0,
    default_monitor_option VARCHAR(50) DEFAULT 'movieOnly',
    default_search_for_missing_movie BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Releases table for search and release management
CREATE TABLE IF NOT EXISTS releases (
    id SERIAL PRIMARY KEY,
    guid VARCHAR(500) NOT NULL,
    title VARCHAR(500) NOT NULL,
    sort_title VARCHAR(500),
    overview TEXT,
    quality_id INTEGER DEFAULT 1,
    quality_name VARCHAR(50) DEFAULT 'Unknown',
    quality_source VARCHAR(50) DEFAULT 'unknown',
    quality_resolution INTEGER DEFAULT 0,
    quality_revision_version INTEGER DEFAULT 1,
    quality_revision_real INTEGER DEFAULT 0,
    quality_revision_is_repack BOOLEAN DEFAULT FALSE,
    quality JSONB,
    quality_weight INTEGER DEFAULT 0,
    age INTEGER DEFAULT 0,
    age_hours DOUBLE PRECISION DEFAULT 0,
    age_minutes DOUBLE PRECISION DEFAULT 0,
    size BIGINT DEFAULT 0,
    indexer_id INTEGER NOT NULL REFERENCES indexers(id) ON DELETE CASCADE,
    movie_id INTEGER REFERENCES movies(id) ON DELETE SET NULL,
    imdb_id VARCHAR(20),
    tmdb_id INTEGER,
    protocol VARCHAR(20) NOT NULL DEFAULT 'torrent',
    download_url VARCHAR(2000) NOT NULL,
    info_url VARCHAR(2000),
    comment_url VARCHAR(2000),
    seeders INTEGER,
    leechers INTEGER,
    peer_count INTEGER DEFAULT 0,
    publish_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'available',
    source VARCHAR(20) NOT NULL DEFAULT 'search',
    release_info JSONB DEFAULT '{}',
    categories JSONB DEFAULT '[]',
    download_client_id INTEGER REFERENCES download_clients(id) ON DELETE SET NULL,
    rejection_reasons JSONB DEFAULT '[]',
    indexer_flags INTEGER DEFAULT 0,
    scene_mapping BOOLEAN DEFAULT FALSE,
    magnet_url VARCHAR(2000),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    grabbed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ
);

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT DEFAULT '{}'::TEXT,
    on_grab BOOLEAN DEFAULT FALSE,
    on_import BOOLEAN DEFAULT FALSE,
    on_upgrade BOOLEAN DEFAULT TRUE,
    on_rename BOOLEAN DEFAULT FALSE,
    on_movie_delete BOOLEAN DEFAULT FALSE,
    on_movie_file_delete BOOLEAN DEFAULT FALSE,
    on_movie_file_delete_for_upgrade BOOLEAN DEFAULT TRUE,
    on_health_issue BOOLEAN DEFAULT FALSE,
    on_health_restored BOOLEAN DEFAULT FALSE,
    on_application_update BOOLEAN DEFAULT FALSE,
    on_manual_interaction_required BOOLEAN DEFAULT FALSE,
    include_health_warnings BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    tags TEXT DEFAULT '[]'::TEXT,
    fields TEXT DEFAULT '[]'::TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create all indexes for optimal performance

-- Movies indexes
CREATE INDEX IF NOT EXISTS idx_movies_tmdb_id ON movies(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_movies_imdb_id ON movies(imdb_id);
CREATE INDEX IF NOT EXISTS idx_movies_title_slug ON movies(title_slug);
CREATE INDEX IF NOT EXISTS idx_movies_monitored ON movies(monitored);
CREATE INDEX IF NOT EXISTS idx_movies_has_file ON movies(has_file);
CREATE INDEX IF NOT EXISTS idx_movies_quality_profile_id ON movies(quality_profile_id);
CREATE INDEX IF NOT EXISTS idx_movies_year ON movies(year);
CREATE INDEX IF NOT EXISTS idx_movies_in_cinemas ON movies(in_cinemas);
CREATE INDEX IF NOT EXISTS idx_movies_physical_release ON movies(physical_release);

-- Movie files indexes
CREATE INDEX IF NOT EXISTS idx_movie_files_movie_id ON movie_files(movie_id);
CREATE INDEX IF NOT EXISTS idx_movie_files_path ON movie_files(path);

-- Quality profiles indexes
CREATE INDEX IF NOT EXISTS idx_quality_profiles_name ON quality_profiles(name);

-- Indexers indexes
CREATE INDEX IF NOT EXISTS idx_indexers_name ON indexers(name);
CREATE INDEX IF NOT EXISTS idx_indexers_implementation ON indexers(implementation);
CREATE INDEX IF NOT EXISTS idx_indexers_enable_rss ON indexers(enable_rss);
CREATE INDEX IF NOT EXISTS idx_indexers_enable_automatic_search ON indexers(enable_automatic_search);
CREATE INDEX IF NOT EXISTS idx_indexers_protocol ON indexers(protocol);

-- Queue items indexes
CREATE INDEX IF NOT EXISTS idx_queue_items_movie_id ON queue_items(movie_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_status ON queue_items(status);
CREATE INDEX IF NOT EXISTS idx_queue_items_download_id ON queue_items(download_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_download_client_id ON queue_items(download_client_id);

-- Download clients indexes
CREATE INDEX IF NOT EXISTS idx_download_clients_name ON download_clients(name);
CREATE INDEX IF NOT EXISTS idx_download_clients_protocol ON download_clients(protocol);
CREATE INDEX IF NOT EXISTS idx_download_clients_enable ON download_clients(enable);
CREATE INDEX IF NOT EXISTS idx_download_clients_priority ON download_clients(priority);

-- Download history indexes
CREATE INDEX IF NOT EXISTS idx_download_history_movie_id ON download_history(movie_id);
CREATE INDEX IF NOT EXISTS idx_download_history_download_client_id ON download_history(download_client_id);
CREATE INDEX IF NOT EXISTS idx_download_history_date ON download_history(date);
CREATE INDEX IF NOT EXISTS idx_download_history_successful ON download_history(successful);
CREATE INDEX IF NOT EXISTS idx_download_history_protocol ON download_history(protocol);

-- Import lists indexes
CREATE INDEX IF NOT EXISTS idx_import_lists_name ON import_lists(name);
CREATE INDEX IF NOT EXISTS idx_import_lists_implementation ON import_lists(implementation);
CREATE INDEX IF NOT EXISTS idx_import_lists_enabled ON import_lists(enabled);
CREATE INDEX IF NOT EXISTS idx_import_lists_enable_auto ON import_lists(enable_auto);
CREATE INDEX IF NOT EXISTS idx_import_lists_quality_profile ON import_lists(quality_profile_id);

-- Import list movies indexes
CREATE INDEX IF NOT EXISTS idx_import_list_movies_import_list_id ON import_list_movies(import_list_id);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_tmdb_id ON import_list_movies(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_imdb_id ON import_list_movies(imdb_id);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_year ON import_list_movies(year);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_is_excluded ON import_list_movies(is_excluded);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_is_existing ON import_list_movies(is_existing);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_is_recommendation ON import_list_movies(is_recommendation);
CREATE INDEX IF NOT EXISTS idx_import_list_movies_discovered_at ON import_list_movies(discovered_at);

-- Import list exclusions indexes
CREATE INDEX IF NOT EXISTS idx_import_list_exclusions_tmdb_id ON import_list_exclusions(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_import_list_exclusions_imdb_id ON import_list_exclusions(imdb_id);

-- History indexes
CREATE INDEX IF NOT EXISTS idx_history_movie_id ON history(movie_id);
CREATE INDEX IF NOT EXISTS idx_history_event_type ON history(event_type);
CREATE INDEX IF NOT EXISTS idx_history_date ON history(date);
CREATE INDEX IF NOT EXISTS idx_history_download_id ON history(download_id);
CREATE INDEX IF NOT EXISTS idx_history_successful ON history(successful);

-- Activity indexes
CREATE INDEX IF NOT EXISTS idx_activity_type ON activity(type);
CREATE INDEX IF NOT EXISTS idx_activity_status ON activity(status);
CREATE INDEX IF NOT EXISTS idx_activity_movie_id ON activity(movie_id);
CREATE INDEX IF NOT EXISTS idx_activity_start_time ON activity(start_time);

-- Configuration indexes
CREATE INDEX IF NOT EXISTS idx_host_config_auth ON host_config(authentication_method);
CREATE INDEX IF NOT EXISTS idx_naming_config_rename ON naming_config(rename_movies);
CREATE INDEX IF NOT EXISTS idx_media_config_recycle ON media_management_config(recycle_bin);
CREATE INDEX IF NOT EXISTS idx_root_folders_accessible ON root_folders(accessible);
CREATE INDEX IF NOT EXISTS idx_root_folders_path ON root_folders(path);

-- Releases indexes
CREATE INDEX IF NOT EXISTS idx_releases_guid_indexer ON releases(guid, indexer_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_releases_guid_indexer_unique ON releases(guid, indexer_id);
CREATE INDEX IF NOT EXISTS idx_releases_title ON releases(sort_title);
CREATE INDEX IF NOT EXISTS idx_releases_movie_id ON releases(movie_id);
CREATE INDEX IF NOT EXISTS idx_releases_indexer_id ON releases(indexer_id);
CREATE INDEX IF NOT EXISTS idx_releases_imdb_id ON releases(imdb_id);
CREATE INDEX IF NOT EXISTS idx_releases_tmdb_id ON releases(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_releases_publish_date ON releases(publish_date);
CREATE INDEX IF NOT EXISTS idx_releases_status ON releases(status);
CREATE INDEX IF NOT EXISTS idx_releases_quality_weight ON releases(quality_weight);
CREATE INDEX IF NOT EXISTS idx_releases_created_at ON releases(created_at);

-- Notifications indexes
CREATE INDEX IF NOT EXISTS idx_notifications_name ON notifications(name);
CREATE INDEX IF NOT EXISTS idx_notifications_implementation ON notifications(implementation);
CREATE INDEX IF NOT EXISTS idx_notifications_enabled ON notifications(enabled);

-- Create all triggers for updated_at timestamp management

-- Helper function for updating updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Movie triggers
CREATE TRIGGER update_movies_updated_at BEFORE UPDATE ON movies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_movie_files_updated_at BEFORE UPDATE ON movie_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Quality profile triggers
CREATE TRIGGER update_quality_profiles_updated_at BEFORE UPDATE ON quality_profiles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Indexer triggers
CREATE TRIGGER update_indexers_updated_at BEFORE UPDATE ON indexers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Queue triggers
CREATE TRIGGER update_queue_items_updated_at BEFORE UPDATE ON queue_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Download client triggers
CREATE OR REPLACE FUNCTION update_download_clients_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_download_clients_updated_at
    BEFORE UPDATE ON download_clients
    FOR EACH ROW
    EXECUTE FUNCTION update_download_clients_updated_at();

-- Download history triggers
CREATE OR REPLACE FUNCTION update_download_history_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_download_history_updated_at
    BEFORE UPDATE ON download_history
    FOR EACH ROW
    EXECUTE FUNCTION update_download_history_updated_at();

-- Import list triggers
CREATE OR REPLACE FUNCTION update_import_lists_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_import_lists_updated_at
    BEFORE UPDATE ON import_lists
    FOR EACH ROW
    EXECUTE FUNCTION update_import_lists_updated_at();

CREATE OR REPLACE FUNCTION update_import_list_movies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_import_list_movies_updated_at
    BEFORE UPDATE ON import_list_movies
    FOR EACH ROW
    EXECUTE FUNCTION update_import_list_movies_updated_at();

CREATE OR REPLACE FUNCTION update_import_list_exclusions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_import_list_exclusions_updated_at
    BEFORE UPDATE ON import_list_exclusions
    FOR EACH ROW
    EXECUTE FUNCTION update_import_list_exclusions_updated_at();

-- Activity trigger
CREATE OR REPLACE FUNCTION update_activity_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_activity_updated_at
    BEFORE UPDATE ON activity
    FOR EACH ROW
    EXECUTE FUNCTION update_activity_updated_at();

-- Configuration triggers
CREATE TRIGGER update_host_config_updated_at BEFORE UPDATE ON host_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_naming_config_updated_at BEFORE UPDATE ON naming_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_media_management_config_updated_at BEFORE UPDATE ON media_management_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_root_folders_updated_at BEFORE UPDATE ON root_folders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Release triggers
CREATE OR REPLACE FUNCTION update_releases_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_releases_updated_at
    BEFORE UPDATE ON releases
    FOR EACH ROW
    EXECUTE FUNCTION update_releases_updated_at();

-- Notification triggers
CREATE TRIGGER update_notifications_updated_at BEFORE UPDATE ON notifications
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default data

-- Insert default quality profile
INSERT INTO quality_profiles (id, name, cutoff, items) VALUES 
(1, 'Any', 1, '[{"quality": {"id": 1, "name": "Unknown"}, "allowed": true}]')
ON CONFLICT (id) DO NOTHING;

-- Insert default host configuration
INSERT INTO host_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- Insert default naming configuration
INSERT INTO naming_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;

-- Insert default media management configuration
INSERT INTO media_management_config (id) VALUES (1) ON CONFLICT (id) DO NOTHING;