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
    path VARCHAR(500),
    popularity DOUBLE DEFAULT 0.0
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quality_definitions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL UNIQUE,
    weight INT NOT NULL DEFAULT 1,
    min_size DOUBLE DEFAULT 0,
    max_size DOUBLE DEFAULT 400,
    preferred_size DOUBLE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS quality_profiles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    cutoff INT NOT NULL,
    items TEXT NOT NULL,
    language VARCHAR(50) DEFAULT 'english',
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

-- Insert default quality definitions
INSERT INTO quality_definitions (id, title, weight, min_size, max_size) VALUES
(0, 'Unknown', 1, 0, 199.9),
(1, 'SDTV', 8, 0, 199.9),
(2, 'DVD', 9, 0, 199.9),
(3, 'WEBDL-1080p', 20, 2, 137.3),
(4, 'HDTV-720p', 15, 0.8, 137.3),
(5, 'WEBDL-720p', 16, 0.8, 137.3),
(6, 'Bluray-720p', 18, 4.3, 137.3),
(7, 'Bluray-1080p', 22, 4.3, 258.1),
(8, 'WEBDL-480p', 11, 0, 199.9),
(9, 'HDTV-1080p', 19, 2, 137.3),
(12, 'WEBRip-480p', 12, 0, 199.9),
(14, 'WEBRip-720p', 17, 0.8, 137.3),
(15, 'WEBRip-1080p', 21, 2, 137.3),
(16, 'HDTV-2160p', 24, 4.7, 199.9),
(17, 'WEBRip-2160p', 26, 4.7, 258.1),
(18, 'WEBDL-2160p', 25, 4.7, 258.1),
(19, 'Bluray-2160p', 27, 4.3, 258.1),
(20, 'Bluray-480p', 13, 0, 199.9),
(21, 'Bluray-576p', 14, 0, 199.9),
(23, 'DVD-R', 10, 0, 199.9),
(24, 'WORKPRINT', 2, 0, 199.9),
(25, 'CAM', 3, 0, 199.9),
(26, 'TELESYNC', 4, 0, 199.9),
(27, 'TELECINE', 5, 0, 199.9),
(28, 'DVDSCR', 7, 0, 199.9),
(29, 'REGIONAL', 6, 0, 199.9),
(30, 'Remux-1080p', 23, 0, 0),
(31, 'Remux-2160p', 28, 0, 0)
ON DUPLICATE KEY UPDATE id=id;

-- Insert default quality profile
INSERT INTO quality_profiles (id, name, cutoff, items, language) VALUES
(1, 'Any', 1, '[{"quality": {"id": 1, "name": "Unknown"}, "allowed": true}]', 'english')
ON DUPLICATE KEY UPDATE id=id;
