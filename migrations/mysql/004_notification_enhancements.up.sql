-- Enhanced notification system migration for MySQL/MariaDB
-- This migration adds notification history, health checks, and additional fields

-- Create notification_history table if it doesn't exist
CREATE TABLE IF NOT EXISTS notification_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    notification_id INT NOT NULL,
    movie_id INT NULL,
    event_type VARCHAR(50) NOT NULL,
    subject VARCHAR(500),
    message TEXT,
    successful BOOLEAN NOT NULL DEFAULT FALSE,
    error_message TEXT,
    date TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Foreign key constraints
    CONSTRAINT fk_notification_history_notification
        FOREIGN KEY (notification_id)
        REFERENCES notifications(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_notification_history_movie
        FOREIGN KEY (movie_id)
        REFERENCES movies(id)
        ON DELETE SET NULL,

    -- Indexes
    INDEX idx_notification_history_notification_id (notification_id),
    INDEX idx_notification_history_movie_id (movie_id),
    INDEX idx_notification_history_event_type (event_type),
    INDEX idx_notification_history_date (date DESC),
    INDEX idx_notification_history_successful (successful)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create health_checks table if it doesn't exist
CREATE TABLE IF NOT EXISTS health_checks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    message VARCHAR(1000) NOT NULL,
    wiki_url VARCHAR(500),
    status ENUM('ok', 'warning', 'error') NOT NULL DEFAULT 'error',
    time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Indexes
    INDEX idx_health_checks_source (source),
    INDEX idx_health_checks_type (type),
    INDEX idx_health_checks_status (status),
    INDEX idx_health_checks_time (time DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add new columns to notifications table if they don't exist
-- Note: MySQL doesn't have IF NOT EXISTS for columns, so we use a procedure

DELIMITER //
CREATE PROCEDURE AddNotificationColumns()
BEGIN
    DECLARE CONTINUE HANDLER FOR 1060 BEGIN END; -- Ignore duplicate column errors

    -- Only add columns that aren't already in the base table
    ALTER TABLE notifications ADD COLUMN on_movie_added BOOLEAN DEFAULT FALSE;
    ALTER TABLE notifications ADD COLUMN supports_on_movie_added BOOLEAN DEFAULT TRUE;
    ALTER TABLE notifications ADD COLUMN supports_on_manual_interaction_required BOOLEAN DEFAULT TRUE;
END //
DELIMITER ;

CALL AddNotificationColumns();
DROP PROCEDURE AddNotificationColumns;
