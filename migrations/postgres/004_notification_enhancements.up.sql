-- Enhanced notification system migration for PostgreSQL
-- This migration adds notification history, health checks, and additional fields

-- Check if notification_history table exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.tables
                   WHERE table_schema = 'public'
                   AND table_name = 'notification_history') THEN
        -- Create notification_history table
        CREATE TABLE notification_history (
            id SERIAL PRIMARY KEY,
            notification_id INTEGER NOT NULL,
            movie_id INTEGER,
            event_type VARCHAR(50) NOT NULL,
            subject VARCHAR(500),
            message TEXT,
            successful BOOLEAN NOT NULL DEFAULT FALSE,
            error_message TEXT,
            date TIMESTAMP WITH TIME ZONE NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

            -- Foreign key constraints
            CONSTRAINT fk_notification_history_notification
                FOREIGN KEY (notification_id)
                REFERENCES notifications(id)
                ON DELETE CASCADE,
            CONSTRAINT fk_notification_history_movie
                FOREIGN KEY (movie_id)
                REFERENCES movies(id)
                ON DELETE SET NULL
        );

        -- Create indexes for notification_history
        CREATE INDEX idx_notification_history_notification_id ON notification_history(notification_id);
        CREATE INDEX idx_notification_history_movie_id ON notification_history(movie_id);
        CREATE INDEX idx_notification_history_event_type ON notification_history(event_type);
        CREATE INDEX idx_notification_history_date ON notification_history(date DESC);
        CREATE INDEX idx_notification_history_successful ON notification_history(successful);
    END IF;
END $$;

-- Check if health_checks table exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.tables
                   WHERE table_schema = 'public'
                   AND table_name = 'health_checks') THEN
        -- Create health_checks table
        CREATE TABLE health_checks (
            id SERIAL PRIMARY KEY,
            source VARCHAR(100) NOT NULL,
            type VARCHAR(50) NOT NULL,
            message VARCHAR(1000) NOT NULL,
            wiki_url VARCHAR(500),
            status VARCHAR(20) NOT NULL DEFAULT 'error',
            time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

            -- Ensure valid status values
            CONSTRAINT chk_health_status CHECK (status IN ('ok', 'warning', 'error'))
        );

        -- Create indexes for health_checks
        CREATE INDEX idx_health_checks_source ON health_checks(source);
        CREATE INDEX idx_health_checks_type ON health_checks(type);
        CREATE INDEX idx_health_checks_status ON health_checks(status);
        CREATE INDEX idx_health_checks_time ON health_checks(time DESC);
    END IF;
END $$;

-- Add new fields to notifications table if they don't exist
DO $$
BEGIN
    -- Check and add on_movie_added column
    IF NOT EXISTS (SELECT column_name FROM information_schema.columns
                   WHERE table_name = 'notifications'
                   AND column_name = 'on_movie_added') THEN
        ALTER TABLE notifications ADD COLUMN on_movie_added BOOLEAN DEFAULT FALSE;
    END IF;

    -- Check and add on_manual_interaction_required column
    IF NOT EXISTS (SELECT column_name FROM information_schema.columns
                   WHERE table_name = 'notifications'
                   AND column_name = 'on_manual_interaction_required') THEN
        ALTER TABLE notifications ADD COLUMN on_manual_interaction_required BOOLEAN DEFAULT FALSE;
    END IF;

    -- Check and add include_health_warnings column
    IF NOT EXISTS (SELECT column_name FROM information_schema.columns
                   WHERE table_name = 'notifications'
                   AND column_name = 'include_health_warnings') THEN
        ALTER TABLE notifications ADD COLUMN include_health_warnings BOOLEAN DEFAULT FALSE;
    END IF;

    -- Check and add supports_on_movie_added column
    IF NOT EXISTS (SELECT column_name FROM information_schema.columns
                   WHERE table_name = 'notifications'
                   AND column_name = 'supports_on_movie_added') THEN
        ALTER TABLE notifications ADD COLUMN supports_on_movie_added BOOLEAN DEFAULT TRUE;
    END IF;

    -- Check and add supports_on_manual_interaction_required column
    IF NOT EXISTS (SELECT column_name FROM information_schema.columns
                   WHERE table_name = 'notifications'
                   AND column_name = 'supports_on_manual_interaction_required') THEN
        ALTER TABLE notifications ADD COLUMN supports_on_manual_interaction_required BOOLEAN DEFAULT TRUE;
    END IF;
END $$;

-- Create or replace trigger function for updating timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for notifications table if it doesn't exist
DROP TRIGGER IF EXISTS update_notifications_updated_at ON notifications;
CREATE TRIGGER update_notifications_updated_at
    BEFORE UPDATE ON notifications
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
