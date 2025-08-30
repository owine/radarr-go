# Radarr Go Configuration Reference

**Version**: v0.9.0-alpha

## Overview

Radarr Go uses a hierarchical configuration system that prioritizes environment variables over YAML file settings. This allows for flexible deployment across different environments while maintaining configuration as code.

## Configuration Priority (Highest to Lowest)

1. **Environment Variables** - `RADARR_` prefixed variables
2. **YAML Configuration File** - `config.yaml` (default) or custom path
3. **Built-in Defaults** - Sensible defaults for all settings

## Configuration File Structure

### Complete Configuration Example

```yaml
# Radarr Go Configuration File
# All settings shown with their default values

server:
  # HTTP server bind address (0.0.0.0 for all interfaces)
  host: "0.0.0.0"

  # HTTP server port
  port: 7878

  # URL base path (for reverse proxy setups)
  url_base: ""

  # SSL/TLS configuration
  enable_ssl: false
  ssl_cert_path: ""
  ssl_key_path: ""

database:
  # Database type: "postgres" or "mariadb"
  type: "postgres"

  # Database connection details
  host: "localhost"
  port: 5432  # 3306 for MariaDB
  database: "radarr"
  username: "radarr"
  password: "password"

  # Advanced database settings
  max_connections: 10
  connection_timeout: "30s"
  idle_timeout: "10m"
  max_lifetime: "1h"

  # Optional: Complete connection URL (overrides individual settings)
  connection_url: ""

  # Database-specific optimizations
  enable_prepared_statements: true
  enable_query_logging: false
  slow_query_threshold: "1s"

log:
  # Log level: "debug", "info", "warn", "error"
  level: "info"

  # Log format: "json", "text"
  format: "json"

  # Log output: "stdout", "stderr", or file path
  output: "stdout"

  # Advanced logging options
  enable_caller: false
  enable_stacktrace: false
  max_size: "100MB"
  max_backups: 5
  max_age: 30  # days

auth:
  # Authentication method: "none", "basic", "apikey"
  method: "apikey"

  # API key for authentication (generated if not provided)
  api_key: ""

  # Basic auth credentials (if using basic auth)
  username: ""
  password: ""

  # Session configuration
  session_timeout: "24h"
  require_authentication_for_ping: false

storage:
  # Primary data directory for databases, logs, etc.
  data_directory: "./data"

  # Default movie storage directory
  movie_directory: "./movies"

  # Backup directory for database backups
  backup_directory: "./data/backups"

  # Temporary directory for processing
  temp_directory: "/tmp"

  # File permissions for created directories
  directory_permissions: 0755
  file_permissions: 0644

tmdb:
  # The Movie Database API key (required for movie discovery)
  api_key: ""

  # API request configuration
  request_timeout: "30s"
  rate_limit: 40  # requests per 10 seconds
  language: "en-US"
  region: "US"

  # Cache settings
  cache_duration: "6h"
  enable_caching: true

health:
  # Enable health monitoring system
  enabled: true

  # Health check interval
  interval: "5m"

  # Disk space monitoring thresholds (in bytes)
  disk_space_warning_threshold: 2147483648   # 2GB
  disk_space_critical_threshold: 1073741824  # 1GB

  # Database performance thresholds
  database_timeout_threshold: "5s"
  database_slow_query_threshold: "1s"

  # External service timeouts
  external_service_timeout: "30s"
  tmdb_timeout: "15s"
  indexer_timeout: "30s"

  # Metrics and data retention
  metrics_retention_days: 30
  issue_retention_days: 90

  # Notification preferences for health issues
  notify_critical_issues: true
  notify_warning_issues: false
  notification_cooldown: "1h"

  # Health check configuration
  checks:
    disk_space: true
    database_connectivity: true
    tmdb_connectivity: true
    indexer_connectivity: true
    download_client_connectivity: true
    file_permissions: true
    system_resources: true

# Task scheduling and execution
tasks:
  # Maximum concurrent tasks
  max_concurrent_tasks: 5

  # Task timeout settings
  default_timeout: "30m"
  import_timeout: "2h"
  search_timeout: "15m"

  # Task cleanup settings
  completed_task_retention: "7d"
  failed_task_retention: "30d"

  # Automatic task scheduling
  auto_refresh_movies: true
  refresh_interval: "12h"
  auto_import_lists_sync: true
  import_lists_sync_interval: "6h"

# File organization and management
file_organization:
  # Enable automatic file organization
  enabled: true

  # File watching and processing
  watch_library_for_changes: true
  rescan_after_refresh: true

  # Import configuration
  skip_free_space_check: false
  minimum_free_space_when_importing: "100MB"
  use_hardlinks_instead_of_copy: false

  # File handling
  import_extra_files: true
  extra_file_extensions: ".srt,.sub,.nfo,.jpg,.png"

  # Organization behavior
  create_empty_movie_folders: false
  delete_empty_folders: true

  # Permissions and ownership
  set_permissions_linux: false
  chmod_folder: "755"
  chmod_file: "644"

# Naming and formatting
naming:
  # Movie naming format
  standard_movie_format: "{Movie Title} ({Release Year}) {Quality Full}"
  movie_folder_format: "{Movie Title} ({Release Year})"

  # Special handling
  rename_movies: false
  replace_illegal_characters: true
  colon_replacement: " - "

  # Multi-episode handling
  multi_part_format: "Part {Part}"

  # Case handling
  movie_title_case: "title"  # "title", "upper", "lower"

# Media management
media_management:
  # File management
  unmonitor_deleted_movies: true
  download_propers: true
  enable_media_info: true

  # Quality and format preferences
  allow_fingerprinting: false
  skip_redownload: true

  # Import and processing
  rescan_after_refresh: true
  auto_rename_folders: false

  # Metadata and extras
  create_movie_folders: true
  extra_file_extensions: ".srt,.sub,.idx,.nfo"

# Quality profiles and definitions
quality:
  # Default quality profile settings
  default_profile_id: 1

  # Quality definition defaults
  quality_definitions:
    - name: "Unknown"
      weight: 1
      min_size: 0
      max_size: null
    - name: "SDTV"
      weight: 2
      min_size: 0
      max_size: 1000
    - name: "DVD"
      weight: 3
      min_size: 0
      max_size: 2000
    # Additional quality definitions...

# Notification system
notifications:
  # Global notification settings
  enabled: true
  include_health_warnings: true
  include_movie_information: true

  # Notification triggers
  on_grab: true
  on_download: true
  on_upgrade: true
  on_rename: true
  on_movie_delete: false
  on_movie_file_delete: false
  on_health_issue: true

  # Message formatting
  message_title_format: "Radarr - {EventType}"
  message_body_format: "{MovieTitle} - {Quality}"

# Calendar and scheduling
calendar:
  # Calendar configuration
  enabled: true

  # Calendar feed settings
  enable_ical_feed: true
  ical_feed_url_base: ""
  ical_secret_key: ""  # Auto-generated if not provided

  # Event filtering
  include_unmonitored: false
  past_days: 7
  future_days: 30

  # iCal feed customization
  feed_title: "Radarr Movie Calendar"
  feed_description: "Radarr movie release dates"

# Import lists and discovery
import_lists:
  # Global import list settings
  enabled: true
  auto_sync_interval: "6h"

  # Import behavior
  monitor_new_items: true
  search_on_add: true

  # List management
  minimum_availability: "inCinemas"  # "announced", "inCinemas", "released"

# Performance and caching
performance:
  # General performance settings
  enable_response_caching: true
  cache_duration: "5m"

  # Database performance
  enable_query_optimization: true
  connection_pool_size: 10

  # File system performance
  parallel_file_operations: 5
  io_timeout: "30s"

  # API performance
  api_rate_limit: 100  # requests per minute
  enable_request_compression: true

# Security settings
security:
  # CORS configuration
  enable_cors: true
  cors_origins: ["*"]
  cors_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  cors_headers: ["Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"]

  # Security headers
  enable_security_headers: true
  content_security_policy: "default-src 'self'"

  # Request security
  max_request_size: "50MB"
  request_timeout: "30s"

# Development and debugging
development:
  # Development mode settings
  enable_debug_endpoints: false
  enable_profiling: false
  log_sql_queries: false

  # Testing configuration
  enable_test_mode: false
  mock_external_services: false
```

## Environment Variable Mapping

All configuration options can be overridden using environment variables with the `RADARR_` prefix. Nested configuration uses underscores to separate levels.

### Server Configuration
```bash
RADARR_SERVER_HOST=0.0.0.0
RADARR_SERVER_PORT=7878
RADARR_SERVER_URL_BASE=""
RADARR_SERVER_ENABLE_SSL=false
RADARR_SERVER_SSL_CERT_PATH=/path/to/cert.pem
RADARR_SERVER_SSL_KEY_PATH=/path/to/key.pem
```

### Database Configuration
```bash
RADARR_DATABASE_TYPE=postgres          # or mariadb
RADARR_DATABASE_HOST=localhost
RADARR_DATABASE_PORT=5432              # 3306 for MariaDB
RADARR_DATABASE_DATABASE=radarr
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=password
RADARR_DATABASE_MAX_CONNECTIONS=10
RADARR_DATABASE_CONNECTION_URL="postgres://user:pass@host:port/db"
```

### Logging Configuration
```bash
RADARR_LOG_LEVEL=info                  # debug, info, warn, error
RADARR_LOG_FORMAT=json                 # json, text
RADARR_LOG_OUTPUT=stdout               # stdout, stderr, or file path
```

### Authentication Configuration
```bash
RADARR_AUTH_METHOD=apikey              # none, basic, apikey
RADARR_AUTH_API_KEY=your-api-key-here
RADARR_AUTH_USERNAME=admin
RADARR_AUTH_PASSWORD=password
```

### Storage Configuration
```bash
RADARR_STORAGE_DATA_DIRECTORY=./data
RADARR_STORAGE_MOVIE_DIRECTORY=./movies
RADARR_STORAGE_BACKUP_DIRECTORY=./data/backups
RADARR_STORAGE_TEMP_DIRECTORY=/tmp
```

### TMDB Configuration
```bash
RADARR_TMDB_API_KEY=your-tmdb-api-key
RADARR_TMDB_LANGUAGE=en-US
RADARR_TMDB_REGION=US
```

### Health Monitoring Configuration
```bash
RADARR_HEALTH_ENABLED=true
RADARR_HEALTH_INTERVAL=5m
RADARR_HEALTH_DISK_SPACE_WARNING_THRESHOLD=2147483648
RADARR_HEALTH_DISK_SPACE_CRITICAL_THRESHOLD=1073741824
RADARR_HEALTH_DATABASE_TIMEOUT_THRESHOLD=5s
RADARR_HEALTH_METRICS_RETENTION_DAYS=30
RADARR_HEALTH_NOTIFY_CRITICAL_ISSUES=true
```

## Database-Specific Configuration

### PostgreSQL Configuration
```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  database: "radarr"
  username: "radarr"
  password: "password"
  max_connections: 10

  # PostgreSQL-specific settings
  ssl_mode: "prefer"  # disable, require, verify-ca, verify-full
  connect_timeout: "30s"
  statement_timeout: "60s"
  lock_timeout: "30s"
```

### MariaDB Configuration
```yaml
database:
  type: "mariadb"
  host: "localhost"
  port: 3306
  database: "radarr"
  username: "radarr"
  password: "password"
  max_connections: 10

  # MariaDB-specific settings
  charset: "utf8mb4"
  collation: "utf8mb4_unicode_ci"
  max_allowed_packet: "64MB"
  sql_mode: "TRADITIONAL"
```

## Docker Environment Variables

When running in Docker, these environment variables are commonly used:

```bash
# Basic Docker configuration
RADARR_SERVER_PORT=7878
RADARR_LOG_LEVEL=info
RADARR_DATA_DIRECTORY=/data
RADARR_MOVIE_DIRECTORY=/movies

# Database configuration for external database
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST=postgres-server
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_DATABASE=radarr
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=secure-password

# Security
RADARR_AUTH_API_KEY=your-secure-api-key

# External service integration
RADARR_TMDB_API_KEY=your-tmdb-api-key
```

## Configuration Validation

The application validates configuration at startup and will:
- Report validation errors with helpful messages
- Create missing directories with appropriate permissions
- Generate secure API keys if not provided
- Apply sensible defaults for missing values

## Configuration File Location

The configuration file is searched in the following locations:
1. Path specified by `--config` command line flag
2. `./config.yaml` (current directory)
3. `./data/config.yaml`
4. `/etc/radarr/config.yaml`

## Performance Tuning

### Database Performance
```yaml
database:
  max_connections: 25          # Increase for high load
  connection_timeout: "10s"    # Reduce for faster failover
  idle_timeout: "5m"          # Reduce to free connections
  enable_prepared_statements: true  # Always enable
  slow_query_threshold: "500ms"     # Monitor performance
```

### File System Performance
```yaml
file_organization:
  use_hardlinks_instead_of_copy: true   # Faster, less disk usage
  skip_free_space_check: false          # Keep enabled for safety
  minimum_free_space_when_importing: "500MB"
```

### Health Monitoring Performance
```yaml
health:
  interval: "10m"                       # Reduce frequency for less overhead
  metrics_retention_days: 14            # Reduce retention for less storage
  checks:
    system_resources: false             # Disable if not needed
```

## Security Configuration

### API Security
```yaml
auth:
  method: "apikey"                      # Always use API key authentication
  api_key: "strong-random-key"          # Use strong, unique API keys
  require_authentication_for_ping: true # Require auth for health checks
```

### Network Security
```yaml
server:
  host: "127.0.0.1"                    # Bind to localhost only if behind reverse proxy
  enable_ssl: true                     # Enable SSL in production

security:
  enable_cors: false                   # Disable CORS if not needed
  cors_origins: ["https://yourdomain.com"]  # Restrict origins
  max_request_size: "10MB"            # Limit request size
```

## Troubleshooting Configuration

### Common Issues

1. **Database Connection Failures**
   - Check host, port, username, password
   - Verify database exists and user has permissions
   - Test connection with database client

2. **File Permission Issues**
   - Ensure data directory is writable
   - Check file/directory permissions
   - Verify user has access to movie directories

3. **API Key Authentication Issues**
   - Verify API key is configured
   - Check header format: `X-API-Key: your-key`
   - Ensure key matches configuration

4. **Health Monitoring Issues**
   - Check disk space thresholds are reasonable
   - Verify external service connectivity
   - Review health check intervals

### Debug Configuration
```yaml
log:
  level: "debug"
  format: "text"          # Easier to read during debugging

development:
  enable_debug_endpoints: true
  log_sql_queries: true   # Log all database queries
```

This configuration reference covers all available options in Radarr Go v0.9.0-alpha. For the most up-to-date information, consult the application's built-in help or generate a sample configuration file.
