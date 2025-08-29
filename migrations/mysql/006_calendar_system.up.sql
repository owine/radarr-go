-- Add calendar and scheduling system tables

-- Calendar events table for movie release tracking
CREATE TABLE IF NOT EXISTS calendar_events (
    id INT AUTO_INCREMENT PRIMARY KEY,
    movie_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    original_title VARCHAR(255),
    event_type VARCHAR(50) NOT NULL,
    event_date DATETIME NOT NULL,
    status VARCHAR(50) NOT NULL,
    monitored BOOLEAN NOT NULL DEFAULT true,
    has_file BOOLEAN NOT NULL DEFAULT false,
    downloaded BOOLEAN NOT NULL DEFAULT false,
    unmonitored BOOLEAN NOT NULL DEFAULT false,
    year INT,
    runtime INT,
    overview TEXT,
    images TEXT, -- JSON array of MediaCover
    genres TEXT, -- JSON array of strings
    quality_profile_id INT,
    folder_name VARCHAR(255),
    path TEXT,
    tmdb_id INT,
    imdb_id VARCHAR(20),
    tags TEXT, -- JSON array of integers

    -- Event-specific metadata
    event_description TEXT,
    all_day BOOLEAN NOT NULL DEFAULT true,
    end_date DATETIME,
    location VARCHAR(255),
    reminder BIGINT, -- Duration in nanoseconds

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_calendar_events_movie_id (movie_id),
    INDEX idx_calendar_events_event_date (event_date),
    INDEX idx_calendar_events_event_type (event_type),
    INDEX idx_calendar_events_status (status),
    INDEX idx_calendar_events_monitored (monitored),
    INDEX idx_calendar_events_year (year),
    INDEX idx_calendar_events_tmdb_id (tmdb_id),
    INDEX idx_calendar_events_imdb_id (imdb_id),

    FOREIGN KEY (movie_id) REFERENCES movies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Calendar configuration table
CREATE TABLE IF NOT EXISTS calendar_configurations (
    id INT AUTO_INCREMENT PRIMARY KEY,
    enabled_event_types TEXT, -- JSON array of CalendarEventType
    default_view VARCHAR(50) NOT NULL DEFAULT 'month',
    first_day_of_week INT NOT NULL DEFAULT 0,
    show_colored_events BOOLEAN NOT NULL DEFAULT true,
    show_movie_information BOOLEAN NOT NULL DEFAULT true,
    full_calendar_event_filter BOOLEAN NOT NULL DEFAULT false,
    collapse_multiple_events BOOLEAN NOT NULL DEFAULT false,

    -- iCal feed settings
    enable_ical_feed BOOLEAN NOT NULL DEFAULT true,
    ical_feed_auth BOOLEAN NOT NULL DEFAULT false,
    ical_feed_passkey VARCHAR(255),
    ical_days_in_future INT NOT NULL DEFAULT 365,
    ical_days_in_past INT NOT NULL DEFAULT 30,
    ical_tags TEXT, -- JSON array of integers

    -- Event display settings
    event_title_format VARCHAR(255) NOT NULL DEFAULT '{Movie Title}',
    event_description_format TEXT NOT NULL DEFAULT '{Movie Overview}',
    time_zone VARCHAR(100) NOT NULL DEFAULT 'UTC',

    -- Caching settings
    enable_event_caching BOOLEAN NOT NULL DEFAULT true,
    event_cache_duration INT NOT NULL DEFAULT 60, -- minutes

    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Calendar event cache table for performance optimization
CREATE TABLE IF NOT EXISTS calendar_event_cache (
    id VARCHAR(32) PRIMARY KEY, -- MD5 hash of cache key
    cache_key VARCHAR(255) UNIQUE NOT NULL,
    events LONGTEXT NOT NULL, -- JSON array of CalendarEvent
    summary TEXT NOT NULL, -- JSON CalendarSummary
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_calendar_event_cache_expires_at (expires_at),
    INDEX idx_calendar_event_cache_cache_key (cache_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default calendar configuration
INSERT IGNORE INTO calendar_configurations (
    enabled_event_types,
    default_view,
    first_day_of_week,
    show_colored_events,
    show_movie_information,
    full_calendar_event_filter,
    collapse_multiple_events,
    enable_ical_feed,
    ical_feed_auth,
    ical_days_in_future,
    ical_days_in_past,
    event_title_format,
    event_description_format,
    time_zone,
    enable_event_caching,
    event_cache_duration
) VALUES (
    '["cinemaRelease","physicalRelease","digitalRelease"]',
    'month',
    0,
    true,
    true,
    false,
    false,
    true,
    false,
    365,
    30,
    '{Movie Title}',
    '{Movie Overview}',
    'UTC',
    true,
    60
);
