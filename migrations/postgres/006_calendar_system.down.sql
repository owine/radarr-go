-- Remove calendar and scheduling system tables

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_update_calendar_event_cache_updated_at ON calendar_event_cache;
DROP TRIGGER IF EXISTS trigger_update_calendar_configurations_updated_at ON calendar_configurations;
DROP TRIGGER IF EXISTS trigger_update_calendar_events_updated_at ON calendar_events;

-- Drop functions
DROP FUNCTION IF EXISTS update_calendar_event_cache_updated_at();
DROP FUNCTION IF EXISTS update_calendar_configurations_updated_at();
DROP FUNCTION IF EXISTS update_calendar_events_updated_at();

-- Drop tables (in reverse order due to dependencies)
DROP TABLE IF EXISTS calendar_event_cache;
DROP TABLE IF EXISTS calendar_configurations;
DROP TABLE IF EXISTS calendar_events;
