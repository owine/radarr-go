-- History and Activity tracking migration for PostgreSQL

-- Create history table
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

-- Create indexes for history table
CREATE INDEX IF NOT EXISTS idx_history_movie_id ON history(movie_id);
CREATE INDEX IF NOT EXISTS idx_history_event_type ON history(event_type);
CREATE INDEX IF NOT EXISTS idx_history_date ON history(date);
CREATE INDEX IF NOT EXISTS idx_history_download_id ON history(download_id);
CREATE INDEX IF NOT EXISTS idx_history_successful ON history(successful);

-- Create activity table
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

-- Create indexes for activity table
CREATE INDEX IF NOT EXISTS idx_activity_type ON activity(type);
CREATE INDEX IF NOT EXISTS idx_activity_status ON activity(status);
CREATE INDEX IF NOT EXISTS idx_activity_movie_id ON activity(movie_id);
CREATE INDEX IF NOT EXISTS idx_activity_start_time ON activity(start_time);

-- Create trigger to update updated_at on activity table
CREATE OR REPLACE FUNCTION update_activity_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS trigger_update_activity_updated_at ON activity;
CREATE TRIGGER trigger_update_activity_updated_at
    BEFORE UPDATE ON activity
    FOR EACH ROW
    EXECUTE FUNCTION update_activity_updated_at();