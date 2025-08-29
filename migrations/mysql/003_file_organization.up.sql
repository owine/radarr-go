-- File Organization and Import Management Tables

-- Create file_organizations table
CREATE TABLE file_organizations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source_path VARCHAR(500) NOT NULL,
    destination_path VARCHAR(500) NOT NULL,
    movie_id INT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    status_message TEXT,
    operation VARCHAR(20) NOT NULL DEFAULT 'move',
    size BIGINT DEFAULT 0,
    quality TEXT,
    languages TEXT,
    release_group VARCHAR(255),
    edition VARCHAR(255),
    original_file_name VARCHAR(255) NOT NULL,
    organized_file_name VARCHAR(255),
    backup_path VARCHAR(500),
    processed_at TIMESTAMP NULL,
    error_message TEXT,
    attempt_count INT DEFAULT 0,
    last_attempt_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL,
    INDEX idx_file_organizations_status (status),
    INDEX idx_file_organizations_movie_id (movie_id),
    INDEX idx_file_organizations_created_at (created_at)
) ENGINE=InnoDB;

-- Create manual_imports table
CREATE TABLE manual_imports (
    id INT AUTO_INCREMENT PRIMARY KEY,
    path VARCHAR(500) NOT NULL,
    name VARCHAR(255) NOT NULL,
    size BIGINT DEFAULT 0,
    quality TEXT NOT NULL,
    languages TEXT,
    movie_id INT,
    download_id VARCHAR(255),
    folder_name VARCHAR(255),
    scene_name VARCHAR(255),
    release_group VARCHAR(255),
    edition VARCHAR(255),
    rejections TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL,
    INDEX idx_manual_imports_movie_id (movie_id),
    INDEX idx_manual_imports_created_at (created_at)
) ENGINE=InnoDB;

-- Create file_operations table for tracking file operations
CREATE TABLE file_operations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    operation_type VARCHAR(20) NOT NULL,
    source_path VARCHAR(500) NOT NULL,
    destination_path VARCHAR(500),
    movie_id INT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    progress DECIMAL(5,2) DEFAULT 0.0,
    size BIGINT DEFAULT 0,
    bytes_processed BIGINT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE SET NULL,
    INDEX idx_file_operations_status (status),
    INDEX idx_file_operations_type (operation_type),
    INDEX idx_file_operations_movie_id (movie_id),
    INDEX idx_file_operations_created_at (created_at)
) ENGINE=InnoDB;

-- Create naming_config table
CREATE TABLE naming_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    rename_movies BOOLEAN DEFAULT false,
    replace_illegal_characters BOOLEAN DEFAULT true,
    colon_replacement_format VARCHAR(20) DEFAULT 'delete',
    standard_movie_format TEXT DEFAULT '{Movie Title} ({Release Year}) {Quality Full}',
    movie_folder_format TEXT DEFAULT '{Movie Title} ({Release Year})',
    create_empty_movie_folders BOOLEAN DEFAULT false,
    delete_empty_folders BOOLEAN DEFAULT false,
    skip_free_space_check_when_importing BOOLEAN DEFAULT false,
    minimum_free_space_when_importing BIGINT DEFAULT 100,
    copy_using_hardlinks BOOLEAN DEFAULT true,
    import_extra_files BOOLEAN DEFAULT false,
    extra_file_extensions TEXT,
    enable_media_info BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB;

-- Insert default naming configuration
INSERT INTO naming_config (
    rename_movies,
    replace_illegal_characters,
    colon_replacement_format,
    standard_movie_format,
    movie_folder_format,
    create_empty_movie_folders,
    delete_empty_folders,
    skip_free_space_check_when_importing,
    minimum_free_space_when_importing,
    copy_using_hardlinks,
    import_extra_files,
    extra_file_extensions,
    enable_media_info
) VALUES (
    false,
    true,
    'delete',
    '{Movie Title} ({Release Year}) {Quality Full}',
    '{Movie Title} ({Release Year})',
    false,
    false,
    false,
    100,
    true,
    false,
    '["srt", "nfo"]',
    true
);
