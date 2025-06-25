-- Queue system migration for PostgreSQL

CREATE TABLE IF NOT EXISTS queue_items (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    languages TEXT DEFAULT '[]'::TEXT,
    quality TEXT DEFAULT '{}'::TEXT,
    size BIGINT DEFAULT 0,
    title VARCHAR(500) NOT NULL,
    size_left BIGINT DEFAULT 0,
    time_left INTERVAL,
    estimated_completion_time TIMESTAMP,
    added TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'queued',
    tracked_download_status VARCHAR(50),
    tracked_download_state VARCHAR(50),
    status_messages TEXT DEFAULT '[]'::TEXT,
    download_id VARCHAR(255) NOT NULL,
    protocol VARCHAR(20) DEFAULT 'unknown',
    download_client VARCHAR(100),
    download_client_has_post_import_category BOOLEAN DEFAULT FALSE,
    indexer VARCHAR(100),
    output_path TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_queue_items_movie_id ON queue_items(movie_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_download_id ON queue_items(download_id);
CREATE INDEX IF NOT EXISTS idx_queue_items_status ON queue_items(status);
CREATE INDEX IF NOT EXISTS idx_queue_items_protocol ON queue_items(protocol);
CREATE INDEX IF NOT EXISTS idx_queue_items_added ON queue_items(added);

-- Create trigger to automatically update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_queue_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_queue_items_updated_at
    BEFORE UPDATE ON queue_items
    FOR EACH ROW
    EXECUTE FUNCTION update_queue_updated_at();