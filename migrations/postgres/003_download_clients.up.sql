-- Download clients migration for PostgreSQL

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

-- Add download_client_id column to queue_items table
ALTER TABLE queue_items 
ADD COLUMN IF NOT EXISTS download_client_id INTEGER REFERENCES download_clients(id) ON DELETE SET NULL;

-- Add downloaded_info column to queue_items table
ALTER TABLE queue_items 
ADD COLUMN IF NOT EXISTS downloaded_info TEXT DEFAULT '{}'::TEXT;

-- Create download_history table
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

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_download_clients_name ON download_clients(name);
CREATE INDEX IF NOT EXISTS idx_download_clients_protocol ON download_clients(protocol);
CREATE INDEX IF NOT EXISTS idx_download_clients_enable ON download_clients(enable);
CREATE INDEX IF NOT EXISTS idx_download_clients_priority ON download_clients(priority);

CREATE INDEX IF NOT EXISTS idx_queue_items_download_client_id ON queue_items(download_client_id);

CREATE INDEX IF NOT EXISTS idx_download_history_movie_id ON download_history(movie_id);
CREATE INDEX IF NOT EXISTS idx_download_history_download_client_id ON download_history(download_client_id);
CREATE INDEX IF NOT EXISTS idx_download_history_date ON download_history(date);
CREATE INDEX IF NOT EXISTS idx_download_history_successful ON download_history(successful);
CREATE INDEX IF NOT EXISTS idx_download_history_protocol ON download_history(protocol);

-- Create trigger to automatically update the updated_at timestamp for download_clients
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

-- Create trigger to automatically update the updated_at timestamp for download_history
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