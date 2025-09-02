# Radarr Go Configuration Reference

**Version**: v0.9.0-alpha
**Configuration Format**: YAML + Environment Variables
**Location**: `config.yaml` or custom path via `--config` flag

This document provides a comprehensive reference for all configuration options available in Radarr Go. The configuration system supports hierarchical YAML configuration files with environment variable overrides for flexible deployment scenarios.

## Configuration Loading Order

1. **Default Values**: Built-in sensible defaults for all settings
2. **YAML Configuration**: Values from `config.yaml` file
3. **Environment Variables**: `RADARR_*` prefixed environment variables (highest priority)

## Quick Start Examples

### Minimal Configuration
```yaml
# config.yaml - Minimal setup for local development
server:
  port: 7878

database:
  type: "postgres"
  password: "your_secure_password"

tmdb:
  api_key: "your_tmdb_api_key"
```

### Production Configuration
```yaml
# config.yaml - Production ready setup
server:
  port: 7878
  host: "0.0.0.0"
  url_base: "/radarr"
  enable_ssl: true
  ssl_cert_path: "/etc/ssl/certs/radarr.crt"
  ssl_key_path: "/etc/ssl/private/radarr.key"

database:
  type: "postgres"
  host: "postgres.example.com"
  port: 5432
  database: "radarr_prod"
  username: "radarr_user"
  password: "${RADARR_DATABASE_PASSWORD}"
  max_connections: 25

log:
  level: "info"
  format: "json"
  output: "/var/log/radarr-go/radarr.log"

auth:
  method: "apikey"
  api_key: "${RADARR_API_KEY}"

storage:
  data_directory: "/var/lib/radarr-go"
  movie_directory: "/media/movies"
  backup_directory: "/var/lib/radarr-go/backups"

tmdb:
  api_key: "${RADARR_TMDB_API_KEY}"

health:
  enabled: true
  interval: "10m"
  disk_space_warning_threshold: 10737418240  # 10GB
  disk_space_critical_threshold: 2147483648  # 2GB
  notify_critical_issues: true
```

## Configuration Sections

### Server Configuration

Controls HTTP server behavior, SSL/TLS settings, and network binding.

```yaml
server:
  port: 7878                    # Port to bind HTTP server
  host: "0.0.0.0"              # Host/interface to bind to
  url_base: ""                 # URL base path (e.g., "/radarr")
  enable_ssl: false            # Enable HTTPS/SSL
  ssl_cert_path: ""            # Path to SSL certificate file
  ssl_key_path: ""             # Path to SSL private key file
```

#### Server Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `port` | int | `7878` | HTTP server port | `RADARR_SERVER_PORT` |
| `host` | string | `"0.0.0.0"` | Network interface to bind | `RADARR_SERVER_HOST` |
| `url_base` | string | `""` | URL base path for reverse proxy | `RADARR_SERVER_URL_BASE` |
| `enable_ssl` | bool | `false` | Enable HTTPS/SSL | `RADARR_SERVER_ENABLE_SSL` |
| `ssl_cert_path` | string | `""` | SSL certificate file path | `RADARR_SERVER_SSL_CERT_PATH` |
| `ssl_key_path` | string | `""` | SSL private key file path | `RADARR_SERVER_SSL_KEY_PATH` |

#### Server Examples

**Reverse Proxy Setup (Nginx/Apache)**:
```yaml
server:
  port: 7878
  host: "127.0.0.1"  # Only local connections
  url_base: "/radarr"
```

**Direct SSL/HTTPS**:
```yaml
server:
  port: 7878
  enable_ssl: true
  ssl_cert_path: "/path/to/certificate.crt"
  ssl_key_path: "/path/to/private.key"
```

**Docker/Container Deployment**:
```yaml
server:
  port: 7878
  host: "0.0.0.0"  # Accept connections from any interface
```

### Database Configuration

Configures database connections, connection pooling, and database-specific options. Supports PostgreSQL (recommended) and MariaDB/MySQL.

```yaml
database:
  type: "postgres"              # Database type: postgres, mariadb, mysql
  host: "localhost"             # Database server hostname
  port: 5432                    # Database server port
  database: "radarr"            # Database name
  username: "radarr"            # Database username
  password: "password"          # Database password
  connection_url: ""            # Full connection string (optional)
  max_connections: 10           # Maximum connection pool size
```

#### Database Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `type` | string | `"postgres"` | Database type (postgres, mariadb, mysql) | `RADARR_DATABASE_TYPE` |
| `host` | string | `"localhost"` | Database server hostname | `RADARR_DATABASE_HOST` |
| `port` | int | `5432` | Database server port (3306 for MariaDB) | `RADARR_DATABASE_PORT` |
| `database` | string | `"radarr"` | Database name | `RADARR_DATABASE_DATABASE` |
| `username` | string | `"radarr"` | Database username | `RADARR_DATABASE_USERNAME` |
| `password` | string | `"password"` | Database password | `RADARR_DATABASE_PASSWORD` |
| `connection_url` | string | `""` | Full connection string override | `RADARR_DATABASE_CONNECTION_URL` |
| `max_connections` | int | `10` | Connection pool size | `RADARR_DATABASE_MAX_CONNECTIONS` |

#### Database Examples

**PostgreSQL (Recommended)**:
```yaml
database:
  type: "postgres"
  host: "postgres-server.example.com"
  port: 5432
  database: "radarr_production"
  username: "radarr_user"
  password: "${POSTGRES_PASSWORD}"
  max_connections: 20
```

**MariaDB/MySQL**:
```yaml
database:
  type: "mariadb"
  host: "mariadb-server.example.com"
  port: 3306
  database: "radarr_db"
  username: "radarr_user"
  password: "${MYSQL_PASSWORD}"
  max_connections: 15
```

**Connection String Override**:
```yaml
database:
  type: "postgres"
  connection_url: "postgres://user:pass@localhost/radarr?sslmode=require"
```

**Docker Compose Integration**:
```yaml
database:
  type: "postgres"
  host: "radarr-postgres"  # Docker service name
  port: 5432
  database: "radarr"
  username: "radarr"
  password: "${POSTGRES_PASSWORD}"
```

### Logging Configuration

Controls logging behavior, output format, and log levels for debugging and monitoring.

```yaml
log:
  level: "info"                 # Log level: debug, info, warn, error
  format: "json"                # Log format: json, text
  output: "stdout"              # Output: stdout, stderr, or file path
```

#### Logging Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `level` | string | `"info"` | Minimum log level | `RADARR_LOG_LEVEL` |
| `format` | string | `"json"` | Log output format | `RADARR_LOG_FORMAT` |
| `output` | string | `"stdout"` | Log output destination | `RADARR_LOG_OUTPUT` |

#### Log Levels

- **`debug`**: Detailed debugging information, SQL queries, HTTP requests
- **`info`**: General information, startup messages, important operations
- **`warn`**: Warning messages, non-critical issues
- **`error`**: Error messages, critical failures

#### Log Formats

- **`json`**: Structured JSON logging (recommended for production)
- **`text`**: Human-readable text format (good for development)

#### Logging Examples

**Development Logging**:
```yaml
log:
  level: "debug"
  format: "text"
  output: "stdout"
```

**Production Logging**:
```yaml
log:
  level: "info"
  format: "json"
  output: "/var/log/radarr-go/radarr.log"
```

**Container Logging**:
```yaml
log:
  level: "info"
  format: "json"
  output: "stdout"  # Captured by container runtime
```

### Authentication Configuration

Configures authentication methods and API security settings.

```yaml
auth:
  method: "none"                # Authentication method: none, apikey
  username: ""                  # Basic auth username (future)
  password: ""                  # Basic auth password (future)
  api_key: ""                   # API key for authentication
```

#### Authentication Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `method` | string | `"none"` | Authentication method | `RADARR_AUTH_METHOD` |
| `username` | string | `""` | Username (reserved for future use) | `RADARR_AUTH_USERNAME` |
| `password` | string | `""` | Password (reserved for future use) | `RADARR_AUTH_PASSWORD` |
| `api_key` | string | `""` | API key for request authentication | `RADARR_AUTH_API_KEY` |

#### Authentication Methods

- **`none`**: No authentication required (development only)
- **`apikey`**: API key authentication via header or query parameter

#### Authentication Examples

**No Authentication (Development)**:
```yaml
auth:
  method: "none"
```

**API Key Authentication**:
```yaml
auth:
  method: "apikey"
  api_key: "your-secure-32-character-api-key-here"
```

**Environment Variable API Key**:
```yaml
auth:
  method: "apikey"
  api_key: "${RADARR_API_KEY}"
```

### Storage Configuration

Configures file system paths for data storage, movie files, and backups.

```yaml
storage:
  data_directory: "./data"      # Application data directory
  movie_directory: "./data/movies"  # Default movie storage path
  backup_directory: "./data/backups"  # Backup files location
```

#### Storage Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `data_directory` | string | `"./data"` | Main application data directory | `RADARR_STORAGE_DATA_DIRECTORY` |
| `movie_directory` | string | `"./data/movies"` | Default movie files location | `RADARR_STORAGE_MOVIE_DIRECTORY` |
| `backup_directory` | string | `"./data/backups"` | Database and config backups | `RADARR_STORAGE_BACKUP_DIRECTORY` |

#### Storage Examples

**Local Development**:
```yaml
storage:
  data_directory: "./data"
  movie_directory: "./data/movies"
  backup_directory: "./data/backups"
```

**Production Server**:
```yaml
storage:
  data_directory: "/var/lib/radarr-go"
  movie_directory: "/media/movies"
  backup_directory: "/var/backups/radarr-go"
```

**NAS/Network Storage**:
```yaml
storage:
  data_directory: "/var/lib/radarr-go"
  movie_directory: "/mnt/nas/movies"
  backup_directory: "/mnt/nas/backups/radarr"
```

### TMDB Configuration

Configures integration with The Movie Database (TMDB) for movie metadata.

```yaml
tmdb:
  api_key: ""                   # TMDB API key for metadata retrieval
```

#### TMDB Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `api_key` | string | `""` | TMDB API key | `RADARR_TMDB_API_KEY` |

#### Getting a TMDB API Key

1. Create account at [themoviedb.org](https://www.themoviedb.org)
2. Go to [API Settings](https://www.themoviedb.org/settings/api)
3. Request an API key
4. Copy the "API Read Access Token" (v4 auth)

#### TMDB Examples

```yaml
tmdb:
  api_key: "your_tmdb_api_key_here"
```

**With Environment Variable**:
```yaml
tmdb:
  api_key: "${TMDB_API_KEY}"
```

### Health Monitoring Configuration

Configures comprehensive health monitoring, alerting, and performance tracking.

```yaml
health:
  enabled: true                           # Enable health monitoring
  interval: "15m"                         # Health check interval
  disk_space_warning_threshold: 5368709120   # 5GB warning threshold
  disk_space_critical_threshold: 1073741824  # 1GB critical threshold
  database_timeout_threshold: "5s"        # Database timeout threshold
  external_service_timeout: "10s"         # External service timeout
  metrics_retention_days: 30               # Metrics retention period
  notify_critical_issues: true            # Notify on critical issues
  notify_warning_issues: false            # Notify on warnings
```

#### Health Monitoring Options

| Option | Type | Default | Description | Environment Variable |
|--------|------|---------|-------------|---------------------|
| `enabled` | bool | `true` | Enable health monitoring | `RADARR_HEALTH_ENABLED` |
| `interval` | string | `"15m"` | Health check frequency | `RADARR_HEALTH_INTERVAL` |
| `disk_space_warning_threshold` | int64 | `5368709120` | Disk space warning (bytes) | `RADARR_HEALTH_DISK_SPACE_WARNING_THRESHOLD` |
| `disk_space_critical_threshold` | int64 | `1073741824` | Disk space critical (bytes) | `RADARR_HEALTH_DISK_SPACE_CRITICAL_THRESHOLD` |
| `database_timeout_threshold` | string | `"5s"` | Database timeout limit | `RADARR_HEALTH_DATABASE_TIMEOUT_THRESHOLD` |
| `external_service_timeout` | string | `"10s"` | External service timeout | `RADARR_HEALTH_EXTERNAL_SERVICE_TIMEOUT` |
| `metrics_retention_days` | int | `30` | Keep metrics for N days | `RADARR_HEALTH_METRICS_RETENTION_DAYS` |
| `notify_critical_issues` | bool | `true` | Send critical notifications | `RADARR_HEALTH_NOTIFY_CRITICAL_ISSUES` |
| `notify_warning_issues` | bool | `false` | Send warning notifications | `RADARR_HEALTH_NOTIFY_WARNING_ISSUES` |

#### Health Check Types

- **Disk Space**: Monitors free disk space on movie and data directories
- **Database**: Checks database connectivity and query performance
- **External Services**: Validates TMDB API and indexer connections
- **System Resources**: CPU, memory, and system load monitoring
- **Application Health**: Service status and internal component health

#### Health Monitoring Examples

**Basic Health Monitoring**:
```yaml
health:
  enabled: true
  interval: "10m"
```

**Advanced Health Configuration**:
```yaml
health:
  enabled: true
  interval: "5m"
  disk_space_warning_threshold: 21474836480   # 20GB
  disk_space_critical_threshold: 5368709120   # 5GB
  database_timeout_threshold: "3s"
  external_service_timeout: "15s"
  metrics_retention_days: 90
  notify_critical_issues: true
  notify_warning_issues: true
```

**Minimal Health Monitoring**:
```yaml
health:
  enabled: true
  interval: "30m"
  notify_critical_issues: false
```

## Environment Variable Reference

All configuration options can be overridden using environment variables with the `RADARR_` prefix. Nested configuration uses underscores.

### Core Environment Variables

```bash
# Server Configuration
RADARR_SERVER_PORT=7878
RADARR_SERVER_HOST=0.0.0.0
RADARR_SERVER_URL_BASE=/radarr
RADARR_SERVER_ENABLE_SSL=true

# Database Configuration
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST=postgres.example.com
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_DATABASE=radarr_prod
RADARR_DATABASE_USERNAME=radarr_user
RADARR_DATABASE_PASSWORD=secure_password_here
RADARR_DATABASE_MAX_CONNECTIONS=20

# Logging Configuration
RADARR_LOG_LEVEL=info
RADARR_LOG_FORMAT=json
RADARR_LOG_OUTPUT=/var/log/radarr.log

# Authentication Configuration
RADARR_AUTH_METHOD=apikey
RADARR_AUTH_API_KEY=your_api_key_here

# Storage Configuration
RADARR_STORAGE_DATA_DIRECTORY=/var/lib/radarr-go
RADARR_STORAGE_MOVIE_DIRECTORY=/media/movies
RADARR_STORAGE_BACKUP_DIRECTORY=/var/backups/radarr

# TMDB Configuration
RADARR_TMDB_API_KEY=your_tmdb_api_key

# Health Monitoring Configuration
RADARR_HEALTH_ENABLED=true
RADARR_HEALTH_INTERVAL=10m
RADARR_HEALTH_NOTIFY_CRITICAL_ISSUES=true
```

### Docker Environment File

Create a `.env` file for Docker deployments:

```bash
# .env file for Docker Compose
RADARR_DATABASE_PASSWORD=secure_db_password
RADARR_API_KEY=secure_32_char_api_key_here
RADARR_TMDB_API_KEY=your_tmdb_api_key_here
RADARR_LOG_LEVEL=info

# Optional overrides
RADARR_SERVER_PORT=7878
RADARR_DATABASE_HOST=radarr-postgres
RADARR_DATABASE_MAX_CONNECTIONS=15
```

## Configuration Validation

Radarr Go validates configuration at startup and provides detailed error messages for invalid settings.

### Common Validation Errors

**Invalid Database Type**:
```
Error: unsupported database type 'sqlite', supported types: postgres, mariadb, mysql
```

**Missing Required Fields**:
```
Error: TMDB API key is required for movie metadata retrieval
```

**Invalid Port Range**:
```
Error: server port must be between 1 and 65535, got: 70000
```

**Invalid Directory Permissions**:
```
Error: cannot create data directory '/protected/path': permission denied
```

### Configuration Testing

Test configuration without starting the server:

```bash
# Test configuration file
./radarr --config config.yaml --test-config

# Test with environment variables
RADARR_DATABASE_PASSWORD=test ./radarr --test-config
```

## Performance Tuning

### Database Performance

**PostgreSQL Optimization**:
```yaml
database:
  type: "postgres"
  max_connections: 20  # Increase for high-load scenarios
  connection_url: "postgres://user:pass@host/db?pool_max_conns=20&pool_min_conns=5"
```

**MariaDB Optimization**:
```yaml
database:
  type: "mariadb"
  max_connections: 15
  connection_url: "user:pass@tcp(host:3306)/db?charset=utf8mb4&parseTime=True&loc=Local&maxAllowedPacket=4194304"
```

### Health Monitoring Performance

**High-Frequency Monitoring**:
```yaml
health:
  enabled: true
  interval: "2m"  # More frequent checks
  metrics_retention_days: 14  # Shorter retention
```

**Resource-Conscious Monitoring**:
```yaml
health:
  enabled: true
  interval: "30m"  # Less frequent checks
  metrics_retention_days: 7   # Minimal retention
```

## Security Considerations

### API Security

**Strong API Key**:
```yaml
auth:
  method: "apikey"
  api_key: "32-character-random-string-here"
```

**Environment Variable Security**:
```bash
# Store sensitive data in environment variables
export RADARR_API_KEY="$(openssl rand -base64 24)"
export RADARR_DATABASE_PASSWORD="$(openssl rand -base64 32)"
```

### SSL/TLS Configuration

**Self-Signed Certificate** (Development):
```bash
# Generate self-signed certificate
openssl req -x509 -newkey rsa:4096 -keyout radarr.key -out radarr.crt -days 365 -nodes
```

**Let's Encrypt Certificate** (Production):
```yaml
server:
  enable_ssl: true
  ssl_cert_path: "/etc/letsencrypt/live/radarr.example.com/fullchain.pem"
  ssl_key_path: "/etc/letsencrypt/live/radarr.example.com/privkey.pem"
```

### File System Security

```yaml
storage:
  data_directory: "/var/lib/radarr-go"  # Dedicated user directory
  movie_directory: "/media/movies"      # Read-only media directory
  backup_directory: "/var/backups/radarr"  # Secure backup location
```

## Migration from Original Radarr

### Configuration Mapping

Original Radarr configurations can be migrated to Radarr Go format:

**Original Radarr `config.xml`**:
```xml
<Config>
  <Port>7878</Port>
  <UrlBase>/radarr</UrlBase>
  <ApiKey>your-api-key</ApiKey>
</Config>
```

**Radarr Go `config.yaml`**:
```yaml
server:
  port: 7878
  url_base: "/radarr"

auth:
  method: "apikey"
  api_key: "your-api-key"
```

### Database Migration

Radarr Go can import existing Radarr databases:

```bash
# Export from original Radarr (SQLite)
sqlite3 /path/to/radarr.db .dump > radarr_export.sql

# Import to PostgreSQL
./radarr --migrate-from-sqlite radarr_export.sql
```

## Troubleshooting

### Common Issues

**Database Connection Failed**:
```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  # Check: Is PostgreSQL running?
  # Check: Are credentials correct?
  # Check: Is database accessible?
```

**Permission Denied on Directories**:
```bash
# Fix directory permissions
sudo chown -R radarr:radarr /var/lib/radarr-go
sudo chmod -R 755 /var/lib/radarr-go
```

**TMDB API Errors**:
```yaml
tmdb:
  api_key: "your_valid_api_key"  # Verify key is valid
```

### Debug Configuration

**Enable Debug Logging**:
```yaml
log:
  level: "debug"
  format: "text"
  output: "stdout"
```

**Health Check Debugging**:
```yaml
health:
  enabled: true
  interval: "1m"  # Frequent checks for debugging
```

### Configuration Validation

**Validate Configuration File**:
```bash
# Check YAML syntax
yamllint config.yaml

# Test Radarr Go configuration
./radarr --config config.yaml --test-config
```

---

For additional configuration examples and deployment scenarios, see the [examples](../examples) directory in the repository.
