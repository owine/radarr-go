-- File Organization and Import Management Tables

-- Create file_organizations table
CREATE TABLE IF NOT EXISTS file_organizations (
    id SERIAL PRIMARY KEY,
    source_path VARCHAR(500) NOT NULL,
    destination_path VARCHAR(500) NOT NULL,
    movie_id INTEGER REFERENCES movies(id) ON DELETE SET NULL,
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
    processed_at TIMESTAMP,
    error_message TEXT,
    attempt_count INTEGER DEFAULT 0,
    last_attempt_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes on file_organizations
CREATE INDEX IF NOT EXISTS idx_file_organizations_status ON file_organizations(status);
CREATE INDEX IF NOT EXISTS idx_file_organizations_movie_id ON file_organizations(movie_id);
CREATE INDEX IF NOT EXISTS idx_file_organizations_created_at ON file_organizations(created_at);

-- Create manual_imports table
CREATE TABLE IF NOT EXISTS manual_imports (
    id SERIAL PRIMARY KEY,
    path VARCHAR(500) NOT NULL,
    name VARCHAR(255) NOT NULL,
    size BIGINT DEFAULT 0,
    quality TEXT NOT NULL,
    languages TEXT,
    movie_id INTEGER REFERENCES movies(id) ON DELETE SET NULL,
    download_id VARCHAR(255),
    folder_name VARCHAR(255),
    scene_name VARCHAR(255),
    release_group VARCHAR(255),
    edition VARCHAR(255),
    rejections TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes on manual_imports
CREATE INDEX IF NOT EXISTS idx_manual_imports_movie_id ON manual_imports(movie_id);
CREATE INDEX IF NOT EXISTS idx_manual_imports_created_at ON manual_imports(created_at);

-- Create file_operations table for tracking file operations
CREATE TABLE IF NOT EXISTS file_operations (
    id SERIAL PRIMARY KEY,
    operation_type VARCHAR(20) NOT NULL,
    source_path VARCHAR(500) NOT NULL,
    destination_path VARCHAR(500),
    movie_id INTEGER REFERENCES movies(id) ON DELETE SET NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    progress DECIMAL(5,2) DEFAULT 0.0,
    size BIGINT DEFAULT 0,
    bytes_processed BIGINT DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes on file_operations
CREATE INDEX IF NOT EXISTS idx_file_operations_status ON file_operations(status);
CREATE INDEX IF NOT EXISTS idx_file_operations_type ON file_operations(operation_type);
CREATE INDEX IF NOT EXISTS idx_file_operations_movie_id ON file_operations(movie_id);
CREATE INDEX IF NOT EXISTS idx_file_operations_created_at ON file_operations(created_at);

-- Create naming_config table
CREATE TABLE IF NOT EXISTS naming_config (
    id SERIAL PRIMARY KEY,
    rename_movies BOOLEAN DEFAULT false,
    replace_illegal_characters BOOLEAN DEFAULT true,
    colon_replacement_format VARCHAR(20) DEFAULT 'delete',
    standard_movie_format TEXT DEFAULT '{Movie Title} ({Release Year}) {Quality Full}',
    movie_folder_format TEXT DEFAULT '{Movie Title} ({Release Year})',
    create_empty_folders BOOLEAN DEFAULT false,
    delete_empty_folders BOOLEAN DEFAULT false,
    skip_free_space_check BOOLEAN DEFAULT false,
    minimum_free_space BIGINT DEFAULT 100,
    use_hardlinks BOOLEAN DEFAULT true,
    import_extra_files BOOLEAN DEFAULT false,
    extra_file_extensions TEXT,
    enable_media_info BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default naming configuration
INSERT INTO naming_config (
    id,
    rename_movies,
    replace_illegal_characters,
    colon_replacement_format,
    standard_movie_format,
    movie_folder_format,
    create_empty_folders,
    delete_empty_folders,
    skip_free_space_check,
    minimum_free_space,
    use_hardlinks,
    import_extra_files,
    extra_file_extensions,
    enable_media_info
) VALUES (
    1,
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
) ON CONFLICT (id) DO NOTHING;

-- Add triggers for updating the updated_at columns
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_file_organizations_updated_at
    BEFORE UPDATE ON file_organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_manual_imports_updated_at
    BEFORE UPDATE ON manual_imports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_file_operations_updated_at
    BEFORE UPDATE ON file_operations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_naming_config_updated_at
    BEFORE UPDATE ON naming_config
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
