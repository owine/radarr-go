-- Simplified MySQL migration for CI compatibility
-- Core tables only, focused on essential functionality

CREATE TABLE IF NOT EXISTS movies (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tmdb_id INT UNIQUE NOT NULL,
    imdb_id VARCHAR(20),
    title VARCHAR(500) NOT NULL,
    title_slug VARCHAR(500) UNIQUE NOT NULL,
    original_title VARCHAR(500),
    year INT,
    runtime INT,
    overview TEXT,
    quality_profile_id INT NOT NULL DEFAULT 1,
    monitored TINYINT(1) DEFAULT 1,
    has_file TINYINT(1) DEFAULT 0,
    added DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    status VARCHAR(20) DEFAULT 'tba',
    sort_title VARCHAR(500),
    folder_name VARCHAR(255),
    path VARCHAR(500)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quality_profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    cutoff INT NOT NULL,
    items TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    implementation VARCHAR(50) NOT NULL,
    config_contract VARCHAR(100),
    settings TEXT DEFAULT '{}',
    on_grab BOOLEAN DEFAULT FALSE,
    on_import BOOLEAN DEFAULT FALSE,
    on_upgrade BOOLEAN DEFAULT TRUE,
    on_rename BOOLEAN DEFAULT FALSE,
    on_movie_delete BOOLEAN DEFAULT FALSE,
    on_movie_file_delete BOOLEAN DEFAULT FALSE,
    on_movie_file_delete_for_upgrade BOOLEAN DEFAULT TRUE,
    on_health_issue BOOLEAN DEFAULT FALSE,
    on_health_restored BOOLEAN DEFAULT FALSE,
    on_application_update BOOLEAN DEFAULT FALSE,
    on_manual_interaction_required BOOLEAN DEFAULT FALSE,
    include_health_warnings BOOLEAN DEFAULT FALSE,
    enabled BOOLEAN DEFAULT TRUE,
    tags TEXT DEFAULT '[]',
    fields TEXT DEFAULT '[]',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default quality profile
INSERT INTO quality_profiles (id, name, cutoff, items) VALUES
(1, 'Any', 1, '[{"quality": {"id": 1, "name": "Unknown"}, "allowed": true}]')
ON DUPLICATE KEY UPDATE id=id;
