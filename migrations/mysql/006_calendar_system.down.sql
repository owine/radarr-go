-- Remove calendar and scheduling system tables

-- Drop tables (in reverse order due to dependencies)
DROP TABLE IF EXISTS calendar_event_cache;
DROP TABLE IF EXISTS calendar_configurations;
DROP TABLE IF EXISTS calendar_events;
