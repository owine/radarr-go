-- Add wanted movies tracking system

-- Wanted movies table for tracking missing movies and cutoff unmet movies
CREATE TABLE IF NOT EXISTS wanted_movies (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL,
    reason TEXT,
    current_quality_id INTEGER REFERENCES quality_definitions(id),
    target_quality_id INTEGER NOT NULL REFERENCES quality_definitions(id),
    is_available BOOLEAN NOT NULL DEFAULT false,
    last_search_time TIMESTAMP WITH TIME ZONE,
    next_search_time TIMESTAMP WITH TIME ZONE,
    search_attempts INTEGER NOT NULL DEFAULT 0,
    max_search_attempts INTEGER NOT NULL DEFAULT 10,
    priority INTEGER NOT NULL DEFAULT 3,
    search_failures TEXT, -- JSON array of SearchFailure
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT uq_wanted_movies_movie_id UNIQUE (movie_id)
);

-- Create indexes for wanted movies
CREATE INDEX IF NOT EXISTS idx_wanted_movies_movie_id ON wanted_movies(movie_id);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_status ON wanted_movies(status);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_priority ON wanted_movies(priority);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_is_available ON wanted_movies(is_available);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_last_search_time ON wanted_movies(last_search_time);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_next_search_time ON wanted_movies(next_search_time);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_search_attempts ON wanted_movies(search_attempts);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_target_quality_id ON wanted_movies(target_quality_id);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_current_quality_id ON wanted_movies(current_quality_id);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_wanted_movies_status_priority ON wanted_movies(status, priority DESC);
CREATE INDEX IF NOT EXISTS idx_wanted_movies_available_searchable ON wanted_movies(is_available, search_attempts, next_search_time) WHERE is_available = true;
CREATE INDEX IF NOT EXISTS idx_wanted_movies_search_eligible ON wanted_movies(search_attempts, max_search_attempts, next_search_time) WHERE search_attempts < max_search_attempts;

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_wanted_movies_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_wanted_movies_updated_at
    BEFORE UPDATE ON wanted_movies
    FOR EACH ROW
    EXECUTE FUNCTION update_wanted_movies_updated_at();

-- Add check constraints for valid enum values
ALTER TABLE wanted_movies ADD CONSTRAINT chk_wanted_movies_status
    CHECK (status IN ('missing', 'cutoffUnmet', 'upgrade'));

ALTER TABLE wanted_movies ADD CONSTRAINT chk_wanted_movies_priority
    CHECK (priority >= 1 AND priority <= 5);

ALTER TABLE wanted_movies ADD CONSTRAINT chk_wanted_movies_search_attempts
    CHECK (search_attempts >= 0);

ALTER TABLE wanted_movies ADD CONSTRAINT chk_wanted_movies_max_search_attempts
    CHECK (max_search_attempts > 0);

-- Create a view for wanted movies with movie details
CREATE OR REPLACE VIEW wanted_movies_with_details AS
SELECT
    wm.id,
    wm.movie_id,
    wm.status,
    wm.reason,
    wm.current_quality_id,
    wm.target_quality_id,
    wm.is_available,
    wm.last_search_time,
    wm.next_search_time,
    wm.search_attempts,
    wm.max_search_attempts,
    wm.priority,
    wm.search_failures,
    wm.created_at,
    wm.updated_at,

    -- Movie details
    m.title,
    m.original_title,
    m.year,
    m.overview,
    m.status as movie_status,
    m.monitored,
    m.has_file,
    m.quality_profile_id,
    m.tmdb_id,
    m.imdb_id,
    m.path,
    m.folder_name,
    m.popularity,
    m.added as movie_added,

    -- Current quality details
    cq.title as current_quality_title,
    cq.weight as current_quality_weight,

    -- Target quality details
    tq.title as target_quality_title,
    tq.weight as target_quality_weight

FROM wanted_movies wm
JOIN movies m ON wm.movie_id = m.id
LEFT JOIN quality_definitions cq ON wm.current_quality_id = cq.id
LEFT JOIN quality_definitions tq ON wm.target_quality_id = tq.id;

-- Create indexes on the view (PostgreSQL will use the base table indexes)
COMMENT ON VIEW wanted_movies_with_details IS 'Comprehensive view of wanted movies with movie and quality details for reporting and API responses';
