-- Add calendar and scheduling system tables

-- Calendar events table for movie release tracking
CREATE TABLE IF NOT EXISTS calendar_events (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    original_title VARCHAR(255),
    event_type VARCHAR(50) NOT NULL,
    event_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) NOT NULL,
    monitored BOOLEAN NOT NULL DEFAULT true,
    has_file BOOLEAN NOT NULL DEFAULT false,
    downloaded BOOLEAN NOT NULL DEFAULT false,
    unmonitored BOOLEAN NOT NULL DEFAULT false,
    year INTEGER,
    runtime INTEGER,
    overview TEXT,
    images TEXT, -- JSON array of MediaCover
    genres TEXT, -- JSON array of strings
    quality_profile_id INTEGER,
    folder_name VARCHAR(255),
    path TEXT,
    tmdb_id INTEGER,
    imdb_id VARCHAR(20),
    tags TEXT, -- JSON array of integers

    -- Event-specific metadata
    event_description TEXT,
    all_day BOOLEAN NOT NULL DEFAULT true,
    end_date TIMESTAMP WITH TIME ZONE,
    location VARCHAR(255),
    reminder BIGINT, -- Duration in nanoseconds

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for calendar events
CREATE INDEX IF NOT EXISTS idx_calendar_events_movie_id ON calendar_events(movie_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_event_date ON calendar_events(event_date);
CREATE INDEX IF NOT EXISTS idx_calendar_events_event_type ON calendar_events(event_type);
CREATE INDEX IF NOT EXISTS idx_calendar_events_status ON calendar_events(status);
CREATE INDEX IF NOT EXISTS idx_calendar_events_monitored ON calendar_events(monitored);
CREATE INDEX IF NOT EXISTS idx_calendar_events_year ON calendar_events(year);
CREATE INDEX IF NOT EXISTS idx_calendar_events_tmdb_id ON calendar_events(tmdb_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_imdb_id ON calendar_events(imdb_id);

-- Calendar configuration table
CREATE TABLE IF NOT EXISTS calendar_configurations (
    id SERIAL PRIMARY KEY,
    enabled_event_types TEXT, -- JSON array of CalendarEventType
    default_view VARCHAR(50) NOT NULL DEFAULT 'month',
    first_day_of_week INTEGER NOT NULL DEFAULT 0,
    show_colored_events BOOLEAN NOT NULL DEFAULT true,
    show_movie_information BOOLEAN NOT NULL DEFAULT true,
    full_calendar_event_filter BOOLEAN NOT NULL DEFAULT false,
    collapse_multiple_events BOOLEAN NOT NULL DEFAULT false,

    -- iCal feed settings
    enable_ical_feed BOOLEAN NOT NULL DEFAULT true,
    ical_feed_auth BOOLEAN NOT NULL DEFAULT false,
    ical_feed_passkey VARCHAR(255),
    ical_days_in_future INTEGER NOT NULL DEFAULT 365,
    ical_days_in_past INTEGER NOT NULL DEFAULT 30,
    ical_tags TEXT, -- JSON array of integers

    -- Event display settings
    event_title_format VARCHAR(255) NOT NULL DEFAULT '{Movie Title}',
    event_description_format TEXT NOT NULL DEFAULT '{Movie Overview}',
    time_zone VARCHAR(100) NOT NULL DEFAULT 'UTC',

    -- Caching settings
    enable_event_caching BOOLEAN NOT NULL DEFAULT true,
    event_cache_duration INTEGER NOT NULL DEFAULT 60, -- minutes

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Calendar event cache table for performance optimization
CREATE TABLE IF NOT EXISTS calendar_event_cache (
    id VARCHAR(32) PRIMARY KEY, -- MD5 hash of cache key
    cache_key VARCHAR(255) UNIQUE NOT NULL,
    events TEXT NOT NULL, -- JSON array of CalendarEvent
    summary TEXT NOT NULL, -- JSON CalendarSummary
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create index for cache expiration
CREATE INDEX IF NOT EXISTS idx_calendar_event_cache_expires_at ON calendar_event_cache(expires_at);
CREATE INDEX IF NOT EXISTS idx_calendar_event_cache_cache_key ON calendar_event_cache(cache_key);

-- Insert default calendar configuration
INSERT INTO calendar_configurations (
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
) ON CONFLICT DO NOTHING;

-- Add trigger to update updated_at timestamp for calendar events
CREATE OR REPLACE FUNCTION update_calendar_events_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_calendar_events_updated_at
    BEFORE UPDATE ON calendar_events
    FOR EACH ROW
    EXECUTE FUNCTION update_calendar_events_updated_at();

-- Add trigger to update updated_at timestamp for calendar configurations
CREATE OR REPLACE FUNCTION update_calendar_configurations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_calendar_configurations_updated_at
    BEFORE UPDATE ON calendar_configurations
    FOR EACH ROW
    EXECUTE FUNCTION update_calendar_configurations_updated_at();

-- Add trigger to update updated_at timestamp for calendar event cache
CREATE OR REPLACE FUNCTION update_calendar_event_cache_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_calendar_event_cache_updated_at
    BEFORE UPDATE ON calendar_event_cache
    FOR EACH ROW
    EXECUTE FUNCTION update_calendar_event_cache_updated_at();
