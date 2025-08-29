-- Health monitoring system tables for MySQL/MariaDB

-- Health issues table
CREATE TABLE IF NOT EXISTS health_issues (
    id INT AUTO_INCREMENT PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    source VARCHAR(100) NOT NULL,
    severity ENUM('info', 'warning', 'error', 'critical') NOT NULL,
    message TEXT NOT NULL,
    details JSON,
    wiki_url VARCHAR(255),
    first_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP NULL,
    is_resolved BOOLEAN DEFAULT false,
    is_dismissed BOOLEAN DEFAULT false,
    data JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_health_issues_type (type),
    INDEX idx_health_issues_severity (severity),
    INDEX idx_health_issues_source (source),
    INDEX idx_health_issues_resolved (is_resolved),
    INDEX idx_health_issues_dismissed (is_dismissed),
    INDEX idx_health_issues_created_at (created_at),
    INDEX idx_health_issues_resolved_at (resolved_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Performance metrics table
CREATE TABLE IF NOT EXISTS performance_metrics (
    id INT AUTO_INCREMENT PRIMARY KEY,
    cpu_usage_percent FLOAT NOT NULL DEFAULT 0,
    memory_usage_mb FLOAT NOT NULL DEFAULT 0,
    memory_total_mb FLOAT NOT NULL DEFAULT 0,
    disk_usage_percent FLOAT NOT NULL DEFAULT 0,
    disk_available_gb FLOAT NOT NULL DEFAULT 0,
    disk_total_gb FLOAT NOT NULL DEFAULT 0,
    database_latency_ms FLOAT NOT NULL DEFAULT 0,
    api_latency_ms FLOAT NOT NULL DEFAULT 0,
    active_connections INT NOT NULL DEFAULT 0,
    queue_size INT NOT NULL DEFAULT 0,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_performance_metrics_timestamp (timestamp),
    INDEX idx_performance_metrics_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add table comments
ALTER TABLE health_issues COMMENT = 'Stores health issues detected by the health monitoring system';
ALTER TABLE performance_metrics COMMENT = 'Stores system performance metrics collected over time';
