-- Migration 011: Create app_config table
-- This migration creates the app_config table that was manually created but never formalized

-- Create app_config table if it doesn't exist
CREATE TABLE IF NOT EXISTS app_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    value JSON DEFAULT '{}'::JSON,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_app_config_key ON app_config(key);
CREATE INDEX IF NOT EXISTS idx_app_config_created_at ON app_config(created_at);
CREATE INDEX IF NOT EXISTS idx_app_config_updated_at ON app_config(updated_at);

-- Create trigger for automatic updated_at timestamp
CREATE OR REPLACE FUNCTION update_app_config_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger only if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger
        WHERE tgname = 'update_app_config_updated_at'
    ) THEN
        CREATE TRIGGER update_app_config_updated_at
            BEFORE UPDATE ON app_config
            FOR EACH ROW
            EXECUTE FUNCTION update_app_config_updated_at();
    END IF;
END $$;

-- Add comments for documentation
COMMENT ON TABLE app_config IS 'Application configuration key-value storage';
COMMENT ON COLUMN app_config.key IS 'Configuration key (unique)';
COMMENT ON COLUMN app_config.value IS 'Configuration value stored as JSON';
COMMENT ON COLUMN app_config.description IS 'Human-readable description of the configuration';

-- Insert default configuration values if they don't exist
INSERT INTO app_config (key, value, description) VALUES
    ('app.version', '"1.0.0"', 'Application version'),
    ('app.initialized', 'true', 'Whether the application has been initialized'),
    ('ui.theme', '"dark"', 'Default UI theme preference'),
    ('security.api_key_required', 'false', 'Whether API key authentication is required')
ON CONFLICT (key) DO NOTHING;
