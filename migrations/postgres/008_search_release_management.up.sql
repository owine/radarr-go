-- Search & Release Management System Migration for PostgreSQL
-- This migration adds support for release searching and management

-- Releases table to store found releases from indexers
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

-- Indexes for optimal query performance
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

-- Trigger to update updated_at timestamp
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

-- Add comments for documentation
COMMENT ON TABLE releases IS 'Stores movie releases found by indexer searches';
COMMENT ON COLUMN releases.guid IS 'Unique identifier from the indexer';
COMMENT ON COLUMN releases.title IS 'Original release title';
COMMENT ON COLUMN releases.sort_title IS 'Lowercase title for sorting';
COMMENT ON COLUMN releases.quality IS 'JSON structure containing quality information';
COMMENT ON COLUMN releases.quality_weight IS 'Calculated weight for quality comparison';
COMMENT ON COLUMN releases.age IS 'Age in days since publication';
COMMENT ON COLUMN releases.protocol IS 'Download protocol: torrent or usenet';
COMMENT ON COLUMN releases.status IS 'Release status: available, grabbed, rejected, failed';
COMMENT ON COLUMN releases.source IS 'Where the release was found: search, rss, interactive';
COMMENT ON COLUMN releases.release_info IS 'JSON structure containing detailed release metadata';
COMMENT ON COLUMN releases.categories IS 'JSON array of indexer category IDs';
COMMENT ON COLUMN releases.rejection_reasons IS 'JSON array of reasons why release was rejected';