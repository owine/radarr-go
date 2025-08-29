-- Add wanted movies tracking system for MySQL/MariaDB

-- Wanted movies table for tracking missing movies and cutoff unmet movies
CREATE TABLE IF NOT EXISTS wanted_movies (
    id INT PRIMARY KEY AUTO_INCREMENT,
    movie_id INT NOT NULL,
    status VARCHAR(50) NOT NULL,
    reason TEXT,
    current_quality_id INT NULL,
    target_quality_id INT NOT NULL,
    is_available BOOLEAN NOT NULL DEFAULT FALSE,
    last_search_time TIMESTAMP NULL,
    next_search_time TIMESTAMP NULL,
    search_attempts INT NOT NULL DEFAULT 0,
    max_search_attempts INT NOT NULL DEFAULT 10,
    priority INT NOT NULL DEFAULT 3,
    search_failures TEXT, -- JSON array of SearchFailure
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT uq_wanted_movies_movie_id UNIQUE (movie_id),
    CONSTRAINT fk_wanted_movies_movie_id FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE,
    CONSTRAINT fk_wanted_movies_current_quality_id FOREIGN KEY (current_quality_id) REFERENCES quality_definitions(id),
    CONSTRAINT fk_wanted_movies_target_quality_id FOREIGN KEY (target_quality_id) REFERENCES quality_definitions(id)
);

-- Create indexes for wanted movies
CREATE INDEX idx_wanted_movies_movie_id ON wanted_movies(movie_id);
CREATE INDEX idx_wanted_movies_status ON wanted_movies(status);
CREATE INDEX idx_wanted_movies_priority ON wanted_movies(priority);
CREATE INDEX idx_wanted_movies_is_available ON wanted_movies(is_available);
CREATE INDEX idx_wanted_movies_last_search_time ON wanted_movies(last_search_time);
CREATE INDEX idx_wanted_movies_next_search_time ON wanted_movies(next_search_time);
CREATE INDEX idx_wanted_movies_search_attempts ON wanted_movies(search_attempts);
CREATE INDEX idx_wanted_movies_target_quality_id ON wanted_movies(target_quality_id);
CREATE INDEX idx_wanted_movies_current_quality_id ON wanted_movies(current_quality_id);

-- Composite indexes for common queries
CREATE INDEX idx_wanted_movies_status_priority ON wanted_movies(status, priority DESC);
CREATE INDEX idx_wanted_movies_available_searchable ON wanted_movies(is_available, search_attempts, next_search_time);
CREATE INDEX idx_wanted_movies_search_eligible ON wanted_movies(search_attempts, max_search_attempts, next_search_time);

-- Add check constraints for valid enum values (MySQL 8.0+)
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
