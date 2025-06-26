-- Configuration Management migration for MySQL/MariaDB

-- Create host configuration table
CREATE TABLE IF NOT EXISTS host_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bind_address VARCHAR(255) NOT NULL DEFAULT '*',
    port INT NOT NULL DEFAULT 7878,
    url_base VARCHAR(255) DEFAULT '',
    enable_ssl BOOLEAN DEFAULT FALSE,
    ssl_port INT DEFAULT 6969,
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
    backup_interval INT DEFAULT 7,
    backup_retention INT DEFAULT 28,
    certificate_validation VARCHAR(50) DEFAULT 'enabled',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create naming configuration table
CREATE TABLE IF NOT EXISTS naming_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create media management configuration table
CREATE TABLE IF NOT EXISTS media_management_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    auto_unmonitor_previous_movie BOOLEAN DEFAULT FALSE,
    recycle_bin VARCHAR(500) DEFAULT '',
    recycle_bin_cleanup INT DEFAULT 7,
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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Create root folders table
CREATE TABLE IF NOT EXISTS root_folders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    path VARCHAR(500) NOT NULL,
    accessible BOOLEAN DEFAULT TRUE,
    free_space BIGINT DEFAULT 0,
    total_space BIGINT DEFAULT 0,
    unmapped_folders TEXT,
    default_tags TEXT,
    default_quality_profile_id INT DEFAULT 0,
    default_monitor_option VARCHAR(50) DEFAULT 'movieOnly',
    default_search_for_missing_movie BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_path (path)
);

-- Create indexes for better performance
CREATE INDEX idx_host_config_auth ON host_config(authentication_method);
CREATE INDEX idx_naming_config_rename ON naming_config(rename_movies);
CREATE INDEX idx_media_config_recycle ON media_management_config(recycle_bin(255));
CREATE INDEX idx_root_folders_accessible ON root_folders(accessible);
CREATE INDEX idx_root_folders_path ON root_folders(path(255));