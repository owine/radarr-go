-- MariaDB initialization script for development
-- This script runs when the MariaDB container starts for the first time

-- Enable useful settings for development
SET GLOBAL general_log = 'ON';
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 1;

-- Performance settings for development
SET GLOBAL innodb_buffer_pool_size = 268435456; -- 256MB
SET GLOBAL query_cache_size = 67108864; -- 64MB
SET GLOBAL query_cache_type = 1;

-- Create development-specific configurations
-- Increase connection limits for development
SET GLOBAL max_connections = 200;

-- Enable performance schema for monitoring
SET GLOBAL performance_schema = ON;

-- Create a development user with broader permissions if needed
-- This will be handled by the container's environment variables

-- Print success message (MariaDB doesn't have RAISE NOTICE, so we use SELECT)
SELECT 'MariaDB development database initialized successfully' AS message;