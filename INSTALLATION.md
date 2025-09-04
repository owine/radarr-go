# Radarr Go Installation & Setup Guide

**Version**: v0.9.0-alpha (95% feature parity, near production-ready)
**Target**: Production deployment within 30 minutes for experienced users

## Quick Start Decision Matrix

| Scenario | Recommendation | Time |
|----------|---------------|------|
| **Testing/Development** | Docker Compose | 5 min |
| **Production (Linux)** | Binary + systemd | 15 min |
| **Home Lab** | Docker with external DB | 10 min |
| **High Performance** | Binary + PostgreSQL | 20 min |

---

## Table of Contents

1. [Docker Installation (Recommended)](#docker-installation-recommended)
2. [Binary Installation](#binary-installation)
3. [Database Setup](#database-setup)
4. [Configuration Guide](#configuration-guide)
5. [Migration from Original Radarr](#migration-from-original-radarr)
6. [Production Deployment](#production-deployment)
7. [Performance Tuning](#performance-tuning)
8. [Troubleshooting](#troubleshooting)

---

## Docker Installation (Recommended)

### Quick Start with Docker Compose

```bash
# Download the docker-compose file
wget https://raw.githubusercontent.com/radarr/radarr-go/main/docker-compose.yml

# Start with PostgreSQL (recommended)
docker-compose --profile postgres up -d

# Or start with basic setup (SQLite)
docker-compose up -d
```

Access Radarr at: http://localhost:7878

### Custom Docker Compose Setup

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  radarr:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha-postgres
    container_name: radarr-go
    restart: unless-stopped
    ports:
      - "7878:7878"
    volumes:
      - ./data:/data
      - /path/to/your/movies:/movies
      - /path/to/your/downloads:/downloads
    environment:
      - RADARR_DATABASE_TYPE=postgres
      - RADARR_DATABASE_HOST=postgres
      - RADARR_DATABASE_USERNAME=radarr
      - RADARR_DATABASE_PASSWORD=your_secure_password
      - RADARR_LOG_LEVEL=info
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:7878/ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  postgres:
    image: postgres:17-alpine
    container_name: radarr-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_DB=radarr
      - POSTGRES_USER=radarr
      - POSTGRES_PASSWORD=your_secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  postgres_data:
```

### Docker Image Tags

**Current Recommendations**:

- `v0.9.0-alpha`: Stable alpha version (recommended)
- `v0.9.0-alpha-postgres`: PostgreSQL optimized
- `v0.9.0-alpha-mariadb`: MariaDB optimized
- `testing`: Latest pre-release

**Production Planning**:

- `latest`: Available at v1.0.0 (Q2 2025)
- `stable`: Production stable releases

### Docker Run Commands

```bash
# Basic setup with SQLite
docker run -d \
  --name radarr-go \
  -p 7878:7878 \
  -v radarr_data:/data \
  -v /path/to/movies:/movies \
  ghcr.io/radarr/radarr-go:v0.9.0-alpha

# Production setup with PostgreSQL
docker run -d \
  --name radarr-go \
  -p 7878:7878 \
  -v radarr_data:/data \
  -v /path/to/movies:/movies \
  -e RADARR_DATABASE_TYPE=postgres \
  -e RADARR_DATABASE_HOST=your-postgres-host \
  -e RADARR_DATABASE_USERNAME=radarr \
  -e RADARR_DATABASE_PASSWORD=your_password \
  ghcr.io/radarr/radarr-go:v0.9.0-alpha-postgres
```

---

## Binary Installation

### Download Binaries

**Supported Platforms**: Linux, macOS, Windows, FreeBSD (amd64/arm64)

```bash
# Linux amd64 (most common)
wget https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-linux-amd64
chmod +x radarr-linux-amd64
sudo mv radarr-linux-amd64 /usr/local/bin/radarr

# Linux arm64 (Raspberry Pi, Apple Silicon)
wget https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-linux-arm64
chmod +x radarr-linux-arm64
sudo mv radarr-linux-arm64 /usr/local/bin/radarr

# macOS Intel
wget https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-darwin-amd64
chmod +x radarr-darwin-amd64
sudo mv radarr-darwin-amd64 /usr/local/bin/radarr

# macOS Apple Silicon
wget https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-darwin-arm64
chmod +x radarr-darwin-arm64
sudo mv radarr-darwin-arm64 /usr/local/bin/radarr

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-windows-amd64.exe" -OutFile "radarr.exe"

# FreeBSD
fetch https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-freebsd-amd64
chmod +x radarr-freebsd-amd64
sudo mv radarr-freebsd-amd64 /usr/local/bin/radarr
```

### Create User and Directories

```bash
# Create radarr user (Linux/FreeBSD)
sudo useradd --system --shell /bin/false --home-dir /var/lib/radarr --create-home radarr

# Create directories
sudo mkdir -p /etc/radarr /var/lib/radarr /var/log/radarr
sudo chown radarr:radarr /var/lib/radarr /var/log/radarr
sudo chmod 755 /etc/radarr
```

### Linux systemd Service

Create `/etc/systemd/system/radarr.service`:

```ini
[Unit]
Description=Radarr Go Movie Collection Manager
After=network.target postgresql.service mariadb.service
Wants=network.target

[Service]
Type=exec
User=radarr
Group=radarr
ExecStart=/usr/local/bin/radarr --data /var/lib/radarr --config /etc/radarr/config.yaml
Restart=on-failure
RestartSec=5
TimeoutStopSec=20

# Environment file (optional)
EnvironmentFile=-/etc/radarr/environment

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/radarr /var/log/radarr /path/to/movies /path/to/downloads

# Resource limits
LimitNOFILE=1048576
LimitNPROC=1048576

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl daemon-reload
sudo systemctl enable radarr
sudo systemctl start radarr
sudo systemctl status radarr
```

### FreeBSD rc.d Service

Create `/usr/local/etc/rc.d/radarr`:

```bash
#!/bin/sh

# PROVIDE: radarr
# REQUIRE: LOGIN cleanvar
# KEYWORD: shutdown

. /etc/rc.subr

name="radarr"
rcvar="radarr_enable"

load_rc_config $name

: ${radarr_enable:="NO"}
: ${radarr_user:="radarr"}
: ${radarr_group:="radarr"}
: ${radarr_dir:="/var/db/radarr"}
: ${radarr_config:="/usr/local/etc/radarr/config.yaml"}

pidfile="/var/run/radarr.pid"
command="/usr/sbin/daemon"
command_args="-c -f -P ${pidfile} -p ${radarr_dir}/radarr.pid /usr/local/bin/radarr --data ${radarr_dir} --config ${radarr_config}"

start_precmd="radarr_precmd"

radarr_precmd() {
    if [ ! -d ${radarr_dir} ]; then
        install -d -o ${radarr_user} -g ${radarr_group} ${radarr_dir}
    fi
}

run_rc_command "$1"
```

Enable:

```bash
sudo chmod +x /usr/local/etc/rc.d/radarr
echo 'radarr_enable="YES"' | sudo tee -a /etc/rc.conf
sudo service radarr start
```

### Windows Service Setup

1. **Download NSSM** (Non-Sucking Service Manager):

   ```powershell
   # Download from https://nssm.cc/download
   # Or via Chocolatey
   choco install nssm
   ```

2. **Install service**:

   ```cmd
   nssm install Radarr "C:\Program Files\Radarr\radarr.exe"
   nssm set Radarr Arguments "--data C:\ProgramData\Radarr --config C:\ProgramData\Radarr\config.yaml"
   nssm set Radarr DisplayName "Radarr Go Movie Manager"
   nssm set Radarr Description "Radarr Go Movie Collection Manager"
   nssm set Radarr Start SERVICE_AUTO_START
   nssm start Radarr
   ```

### Source Compilation

**Requirements**: Go 1.24+

```bash
# Clone repository
git clone https://github.com/radarr/radarr-go.git
cd radarr-go

# Install dependencies
make deps

# Build for current platform
make build

# Build for specific platform
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-linux-amd64 ./cmd/radarr

# Install system-wide
sudo make install
```

---

## Database Setup

### PostgreSQL (Recommended)

**Why PostgreSQL?**

- Best performance with Radarr-go
- Advanced features (JSON columns, complex queries)
- Excellent concurrency support
- Native Go driver (no CGO)

#### PostgreSQL Installation

**Ubuntu/Debian**:

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**CentOS/RHEL/Fedora**:

```bash
sudo dnf install postgresql postgresql-server postgresql-contrib
sudo postgresql-setup --initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**macOS**:

```bash
brew install postgresql@17
brew services start postgresql@17
```

**FreeBSD**:

```bash
sudo pkg install postgresql17-server
echo 'postgresql_enable="YES"' | sudo tee -a /etc/rc.conf
sudo service postgresql initdb
sudo service postgresql start
```

#### PostgreSQL Configuration

```bash
# Switch to postgres user
sudo -u postgres psql

-- Create database and user
CREATE DATABASE radarr;
CREATE USER radarr WITH PASSWORD 'your_secure_password_here';
GRANT ALL PRIVILEGES ON DATABASE radarr TO radarr;
GRANT ALL ON SCHEMA public TO radarr;

-- Enable extensions (optional but recommended)
\c radarr
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

\q
```

#### PostgreSQL Performance Tuning

Edit `/etc/postgresql/17/main/postgresql.conf`:

```ini
# Memory settings (adjust for your system)
shared_buffers = 256MB                  # 25% of RAM for dedicated systems
effective_cache_size = 1GB              # 75% of available RAM
work_mem = 4MB                          # Per-connection work memory
maintenance_work_mem = 64MB

# Checkpoint settings
checkpoint_completion_target = 0.9
wal_buffers = 16MB

# Connection settings
max_connections = 100                   # Default is usually fine
listen_addresses = 'localhost'          # Restrict to localhost for security

# Logging (optional)
log_destination = 'stderr'
logging_collector = on
log_directory = 'log'
log_filename = 'postgresql-%Y-%m-%d_%H%M%S.log'
log_min_duration_statement = 1000       # Log slow queries
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
```

### MariaDB/MySQL Alternative

**Ubuntu/Debian**:

```bash
sudo apt update
sudo apt install mariadb-server mariadb-client
sudo mysql_secure_installation
```

**CentOS/RHEL/Fedora**:

```bash
sudo dnf install mariadb-server mariadb
sudo systemctl start mariadb
sudo systemctl enable mariadb
sudo mysql_secure_installation
```

#### MariaDB Configuration

```bash
sudo mysql -u root -p

-- Create database and user
CREATE DATABASE radarr CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'radarr'@'localhost' IDENTIFIED BY 'your_secure_password_here';
GRANT ALL PRIVILEGES ON radarr.* TO 'radarr'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

#### MariaDB Performance Tuning

Edit `/etc/mysql/mariadb.conf.d/50-server.cnf`:

```ini
[mysqld]
# Basic settings
innodb_buffer_pool_size = 256M          # 70% of available RAM
innodb_log_file_size = 64M
innodb_flush_log_at_trx_commit = 2
innodb_file_per_table = 1

# Connection settings
max_connections = 100
wait_timeout = 300

# Character set
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# Query cache (if using MySQL 5.7)
query_cache_type = 1
query_cache_size = 32M
```

### Database Connection Testing

```bash
# Test PostgreSQL connection
psql -h localhost -U radarr -d radarr -c "SELECT version();"

# Test MariaDB connection
mysql -h localhost -u radarr -p radarr -e "SELECT version();"
```

---

## Configuration Guide

### Complete config.yaml Example

Create `/etc/radarr/config.yaml` (or `./data/config.yaml` for local install):

```yaml
---
# Server Configuration
server:
  port: 7878
  host: "0.0.0.0"                       # Bind to all interfaces
  url_base: ""                          # For reverse proxy: "/radarr"
  enable_ssl: false                     # Enable for HTTPS
  ssl_cert_path: "/etc/ssl/radarr.crt"  # SSL certificate path
  ssl_key_path: "/etc/ssl/radarr.key"   # SSL private key path

# Database Configuration
database:
  type: "postgres"                      # Options: postgres, mariadb, mysql
  host: "localhost"
  port: 5432                           # 3306 for MariaDB/MySQL
  database: "radarr"
  username: "radarr"
  password: "${RADARR_DATABASE_PASSWORD:-your_secure_password}"

  # Advanced connection settings
  max_connections: 25                   # Connection pool size
  max_idle_connections: 5              # Idle connections
  connection_max_lifetime: "1h"        # Connection lifetime

  # Optional: Full connection string override
  connection_url: ""                   # postgresql://user:pass@host:port/db

# Logging Configuration
log:
  level: "info"                        # Options: debug, info, warn, error
  format: "json"                       # Options: json, text
  output: "stdout"                     # Options: stdout, file
  file_path: "/var/log/radarr/radarr.log"  # If output is "file"
  max_size: 100                        # Max log file size (MB)
  max_backups: 5                       # Number of old log files
  max_age: 28                          # Days to retain logs

# Authentication
auth:
  method: "apikey"                     # Options: none, apikey
  api_key: "${RADARR_API_KEY:-}"       # Generate with: openssl rand -hex 32
  require_auth_for_static: false      # Require auth for static content

# Storage Directories
storage:
  data_directory: "/var/lib/radarr"    # Application data
  movie_directory: "/movies"           # Default movie storage
  backup_directory: "/var/lib/radarr/backups"
  logs_directory: "/var/log/radarr"

# External API Keys
tmdb:
  api_key: "${TMDB_API_KEY:-}"         # Get from https://www.themoviedb.org/settings/api

# CORS Configuration (for web UI)
cors:
  enabled: true
  allowed_origins: ["*"]               # Restrict in production: ["https://yourdomain.com"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["*"]

# Performance Tuning
performance:
  worker_count: 4                      # Background workers (CPU cores)
  queue_size: 1000                     # Task queue size
  request_timeout: "30s"               # HTTP request timeout
  max_request_size: "32MB"             # Max upload size

# Health Check Configuration
health:
  enabled: true
  check_interval: "5m"                 # Health check frequency
  disk_space_warning: 5                # GB free space warning
  disk_space_critical: 1               # GB free space critical
```

### Environment Variables

All configuration can be overridden with environment variables using `RADARR_` prefix:

```bash
# Database configuration
export RADARR_DATABASE_TYPE=postgres
export RADARR_DATABASE_HOST=localhost
export RADARR_DATABASE_PORT=5432
export RADARR_DATABASE_USERNAME=radarr
export RADARR_DATABASE_PASSWORD=your_secure_password
export RADARR_DATABASE_MAX_CONNECTIONS=25

# Server configuration
export RADARR_SERVER_PORT=7878
export RADARR_SERVER_HOST=0.0.0.0
export RADARR_SERVER_URL_BASE=""

# Logging
export RADARR_LOG_LEVEL=info
export RADARR_LOG_FORMAT=json

# Authentication
export RADARR_AUTH_METHOD=apikey
export RADARR_AUTH_API_KEY=$(openssl rand -hex 32)

# External APIs
export TMDB_API_KEY=your_tmdb_api_key
```

### Environment File Example

Create `/etc/radarr/environment`:

```bash
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST=localhost
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=your_secure_password_here
RADARR_LOG_LEVEL=info
RADARR_AUTH_METHOD=apikey
RADARR_AUTH_API_KEY=your_64_character_hex_api_key_here
TMDB_API_KEY=your_tmdb_api_key_here
```

### Security Configuration

#### API Key Generation

```bash
# Generate secure API key
openssl rand -hex 32

# Or use Python
python3 -c "import secrets; print(secrets.token_hex(32))"
```

#### SSL/HTTPS Setup

1. **Generate self-signed certificate** (development):

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/radarr.key \
  -out /etc/ssl/radarr.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

sudo chown radarr:radarr /etc/ssl/radarr.*
sudo chmod 600 /etc/ssl/radarr.key
```

2. **Update config**:

```yaml
server:
  enable_ssl: true
  ssl_cert_path: "/etc/ssl/radarr.crt"
  ssl_key_path: "/etc/ssl/radarr.key"
```

#### Reverse Proxy Configuration

**Nginx**:

```nginx
server {
    listen 80;
    server_name radarr.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name radarr.yourdomain.com;

    ssl_certificate /path/to/ssl/cert;
    ssl_certificate_key /path/to/ssl/key;

    location / {
        proxy_pass http://localhost:7878;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_redirect off;
    }
}
```

**Traefik**:

```yaml
services:
  radarr:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.radarr.rule=Host(`radarr.yourdomain.com`)"
      - "traefik.http.routers.radarr.entrypoints=websecure"
      - "traefik.http.routers.radarr.tls.certresolver=myresolver"
```

---

## Migration from Original Radarr

### Pre-Migration Checklist

- [ ] Backup original Radarr database and configuration
- [ ] Document current indexers, download clients, and notifications
- [ ] Note custom quality profiles and naming formats
- [ ] Identify import list configurations
- [ ] Test Radarr-go in parallel environment first

### Migration Strategy

⚠️ **Important**: Direct database migration is not supported. This is an API-compatible rewrite, not a drop-in database replacement.

#### Step 1: Environment Preparation

```bash
# Stop original Radarr
sudo systemctl stop radarr  # or docker stop radarr

# Backup original Radarr
sudo cp -r /home/radarr/.config/Radarr /home/radarr/.config/Radarr.backup
# or for Docker:
docker run --rm -v radarr_data:/data -v $(pwd):/backup alpine tar czf /backup/radarr-backup.tar.gz /data
```

#### Step 2: Install Radarr-go

Follow the [Docker Installation](#docker-installation-recommended) or [Binary Installation](#binary-installation) sections above.

#### Step 3: Configuration Migration

1. **Server Settings**:

```yaml
# Map from original Radarr UI settings
server:
  port: 7878              # Same port as original
  url_base: ""            # Same URL base if using reverse proxy
```

2. **Quality Profiles**:
   - Must be recreated manually in Radarr-go UI
   - Document existing profiles from original Radarr

3. **Indexers**:
   - Must be reconfigured manually
   - API keys and settings can be copied

#### Step 4: Movie Library Re-import

```bash
# Radarr-go will scan existing movie files
# Set movie directories in UI: Settings > Media Management > Root Folders
# Add: /path/to/your/movies

# Manual scan command
curl -X POST "http://localhost:7878/api/v3/command" \
     -H "X-Api-Key: your-api-key" \
     -H "Content-Type: application/json" \
     -d '{"name": "RescanMovie"}'
```

#### Step 5: Download Client Setup

Reconfigure in UI:

- Settings > Download Clients
- Add your existing clients (qBittorrent, Transmission, etc.)
- Test connections

#### Step 6: Notification Setup

Reconfigure in UI:

- Settings > Connect
- 11 providers supported: Discord, Slack, Email, Webhook, etc.

### Migration Validation

```bash
# Check API compatibility
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/system/status"

# Verify movie count
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/movie" | jq length

# Test search functionality
curl -X POST -H "X-Api-Key: your-api-key" \
     -H "Content-Type: application/json" \
     "http://localhost:7878/api/v3/release" \
     -d '{"term":"movie name"}'
```

### Parallel Testing Approach (Recommended)

1. **Install Radarr-go on different port**:

```yaml
server:
  port: 7879  # Different port
```

2. **Test functionality** while original Radarr runs
3. **Migrate gradually** once satisfied
4. **Switch ports** when ready

---

## Production Deployment

### High Availability Setup

#### Load Balancer Configuration

**HAProxy**:

```
backend radarr_backend
    balance roundrobin
    option httpchk GET /ping
    server radarr1 radarr1.local:7878 check
    server radarr2 radarr2.local:7878 check
```

#### Database High Availability

**PostgreSQL with Patroni**:

```yaml
# patroni.yml
scope: radarr-cluster
name: radarr-node1

restapi:
  listen: 0.0.0.0:8008
  connect_address: radarr-node1:8008

etcd:
  hosts: etcd1:2379,etcd2:2379,etcd3:2379

bootstrap:
  dcs:
    ttl: 30
    loop_wait: 10
    retry_timeout: 60
    postgresql:
      use_pg_rewind: true
      use_slots: true
      parameters:
        max_connections: 200
        shared_buffers: 256MB

postgresql:
  listen: 0.0.0.0:5432
  connect_address: radarr-node1:5432
  data_dir: /var/lib/postgresql/data
  authentication:
    replication:
      username: replicator
      password: password
```

### Monitoring Setup

#### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'radarr-go'
    static_configs:
      - targets: ['localhost:7878']
    scrape_interval: 30s
    metrics_path: /metrics
    params:
      api_key: ['your-api-key']
```

#### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Radarr Go Monitoring",
    "panels": [
      {
        "title": "API Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "http_request_duration_seconds{job=\"radarr-go\"}"
          }
        ]
      },
      {
        "title": "Movie Count",
        "type": "stat",
        "targets": [
          {
            "expr": "radarr_movies_total{job=\"radarr-go\"}"
          }
        ]
      }
    ]
  }
}
```

### Backup Strategy

#### Automated Backup Script

Create `/usr/local/bin/radarr-backup.sh`:

```bash
#!/bin/bash
set -euo pipefail

BACKUP_DIR="/var/backups/radarr"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Database backup
if [ "$RADARR_DATABASE_TYPE" = "postgres" ]; then
    pg_dump -h "$RADARR_DATABASE_HOST" -U "$RADARR_DATABASE_USERNAME" \
            -d "$RADARR_DATABASE_DATABASE" > "$BACKUP_DIR/radarr_db_$DATE.sql"
elif [ "$RADARR_DATABASE_TYPE" = "mariadb" ]; then
    mysqldump -h "$RADARR_DATABASE_HOST" -u "$RADARR_DATABASE_USERNAME" \
              -p"$RADARR_DATABASE_PASSWORD" "$RADARR_DATABASE_DATABASE" \
              > "$BACKUP_DIR/radarr_db_$DATE.sql"
fi

# Configuration backup
tar -czf "$BACKUP_DIR/radarr_config_$DATE.tar.gz" /etc/radarr /var/lib/radarr

# Cleanup old backups
find "$BACKUP_DIR" -name "*.sql" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: $DATE"
```

#### Cron Job

```bash
# Add to crontab
sudo crontab -e

# Daily backup at 2 AM
0 2 * * * /usr/local/bin/radarr-backup.sh >> /var/log/radarr-backup.log 2>&1
```

### Security Hardening

#### Firewall Configuration

**UFW (Ubuntu)**:

```bash
sudo ufw allow 7878/tcp comment 'Radarr Go'
sudo ufw allow from 192.168.1.0/24 to any port 7878 comment 'Radarr Go LAN only'
```

**firewalld (CentOS/RHEL)**:

```bash
sudo firewall-cmd --permanent --add-port=7878/tcp
sudo firewall-cmd --permanent --add-rich-rule='rule family="ipv4" source address="192.168.1.0/24" port protocol="tcp" port="7878" accept'
sudo firewall-cmd --reload
```

#### SELinux Configuration (CentOS/RHEL)

```bash
# Allow Radarr to bind to port
sudo setsebool -P httpd_can_network_connect 1

# Create custom policy if needed
sudo semanage port -a -t http_port_t -p tcp 7878
```

---

## Performance Tuning

### System Requirements

| Library Size | CPU | RAM | Storage | Database |
|-------------|-----|-----|---------|----------|
| < 1,000 movies | 1 core | 512MB | 10GB | SQLite |
| 1,000-5,000 movies | 2 cores | 1GB | 20GB | PostgreSQL |
| 5,000-10,000 movies | 4 cores | 2GB | 50GB | PostgreSQL |
| > 10,000 movies | 8 cores | 4GB | 100GB+ | PostgreSQL HA |

### Configuration Optimization

#### For Large Libraries (>5,000 movies)

```yaml
database:
  max_connections: 50
  max_idle_connections: 10
  connection_max_lifetime: "30m"

performance:
  worker_count: 8                      # Match CPU cores
  queue_size: 2000
  request_timeout: "60s"
  max_request_size: "64MB"

log:
  level: "warn"                        # Reduce log verbosity
```

#### Database Tuning for PostgreSQL

```sql
-- Analyze database regularly
ANALYZE;

-- Create indexes for common queries
CREATE INDEX CONCURRENTLY idx_movies_monitored ON movies (monitored) WHERE monitored = true;
CREATE INDEX CONCURRENTLY idx_movies_status ON movies (status);
CREATE INDEX CONCURRENTLY idx_movie_files_movie_id ON movie_files (movie_id);

-- Update statistics
UPDATE pg_stat_user_tables SET n_tup_ins = 0, n_tup_upd = 0, n_tup_del = 0;
```

### Monitoring Performance

#### Built-in Metrics

```bash
# Check system status
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/system/status"

# Health check
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/health"

# Performance metrics (if enabled)
curl "http://localhost:7878/metrics"
```

#### Custom Monitoring Script

```bash
#!/bin/bash
# radarr-monitor.sh

API_KEY="your-api-key"
BASE_URL="http://localhost:7878"

# Check API response time
response_time=$(curl -w "%{time_total}" -s -o /dev/null -H "X-Api-Key: $API_KEY" "$BASE_URL/api/v3/system/status")
echo "API Response Time: ${response_time}s"

# Check movie count
movie_count=$(curl -s -H "X-Api-Key: $API_KEY" "$BASE_URL/api/v3/movie" | jq length)
echo "Total Movies: $movie_count"

# Check queue size
queue_size=$(curl -s -H "X-Api-Key: $API_KEY" "$BASE_URL/api/v3/queue" | jq length)
echo "Queue Size: $queue_size"

# Check disk space
disk_usage=$(df -h /var/lib/radarr | awk 'NR==2 {print $5}')
echo "Disk Usage: $disk_usage"
```

---

## Troubleshooting

### Common Issues and Solutions

#### 1. Database Connection Failures

**Symptoms**: `failed to connect to database`, `connection refused`

**Diagnosis**:

```bash
# Test database connectivity
pg_isready -h localhost -p 5432 -U radarr  # PostgreSQL
mysql -h localhost -u radarr -p -e "SELECT 1"  # MariaDB

# Check if database is running
sudo systemctl status postgresql  # or mariadb
netstat -tlnp | grep :5432  # or :3306
```

**Solutions**:

1. **Start database service**:

   ```bash
   sudo systemctl start postgresql
   sudo systemctl enable postgresql
   ```

2. **Check credentials**:

   ```bash
   psql -h localhost -U radarr -d radarr  # Test manually
   ```

3. **Verify network binding**:

   ```bash
   # PostgreSQL: Edit /etc/postgresql/17/main/postgresql.conf
   listen_addresses = 'localhost'

   # MariaDB: Edit /etc/mysql/mariadb.conf.d/50-server.cnf
   bind-address = 127.0.0.1
   ```

4. **Check firewall**:

   ```bash
   sudo ufw allow 5432  # PostgreSQL
   sudo ufw allow 3306  # MariaDB
   ```

#### 2. Permission Errors

**Symptoms**: `permission denied`, `cannot write to directory`

**Solutions**:

```bash
# Fix ownership
sudo chown -R radarr:radarr /var/lib/radarr /var/log/radarr

# Fix permissions
sudo chmod 755 /var/lib/radarr
sudo chmod 644 /etc/radarr/config.yaml

# SELinux issues (CentOS/RHEL)
sudo restorecon -R /var/lib/radarr
sudo setsebool -P httpd_can_network_connect 1
```

#### 3. Port Binding Issues

**Symptoms**: `bind: address already in use`

**Diagnosis**:

```bash
# Check what's using port 7878
sudo lsof -i :7878
sudo netstat -tlnp | grep :7878

# Check if old Radarr process is running
ps aux | grep radarr
```

**Solutions**:

```bash
# Kill conflicting process
sudo pkill -f radarr

# Use different port temporarily
RADARR_SERVER_PORT=7879 ./radarr

# Or update configuration
server:
  port: 7879
```

#### 4. Configuration Not Loading

**Symptoms**: Application using defaults instead of config file

**Diagnosis**:

```bash
# Check config file exists and is readable
ls -la /etc/radarr/config.yaml
cat /etc/radarr/config.yaml

# Validate YAML syntax
python3 -c "import yaml; print('Valid YAML' if yaml.safe_load(open('/etc/radarr/config.yaml')) else 'Invalid YAML')"

# Check environment variables
env | grep RADARR_
```

**Solutions**:

1. **Specify config path explicitly**:

   ```bash
   ./radarr --config /etc/radarr/config.yaml
   ```

2. **Fix YAML syntax**:

   ```bash
   # Common issues:
   # - Incorrect indentation (use spaces, not tabs)
   # - Missing quotes around strings with special characters
   # - Wrong nesting levels
   ```

3. **Environment variable precedence**:

   ```bash
   # Unset conflicting env vars
   unset RADARR_SERVER_PORT
   ```

#### 5. Migration Issues

**Symptoms**: Movies not appearing, metadata missing

**Solutions**:

1. **Force library scan**:

   ```bash
   curl -X POST "http://localhost:7878/api/v3/command" \
        -H "X-Api-Key: your-api-key" \
        -H "Content-Type: application/json" \
        -d '{"name": "RescanMovie"}'
   ```

2. **Check root folder configuration**:
   - UI: Settings > Media Management > Root Folders
   - Verify paths are correct and accessible

3. **Clear cache and restart**:

   ```bash
   sudo systemctl stop radarr
   sudo rm -rf /var/lib/radarr/cache/*
   sudo systemctl start radarr
   ```

#### 6. SSL/HTTPS Issues

**Symptoms**: `certificate verify failed`, SSL handshake errors

**Solutions**:

1. **Check certificate validity**:

   ```bash
   openssl x509 -in /etc/ssl/radarr.crt -text -noout
   ```

2. **Verify file permissions**:

   ```bash
   sudo chown radarr:radarr /etc/ssl/radarr.*
   sudo chmod 600 /etc/ssl/radarr.key
   sudo chmod 644 /etc/ssl/radarr.crt
   ```

3. **Test without SSL first**:

   ```yaml
   server:
     enable_ssl: false
   ```

### Performance Issues

#### High Memory Usage

**Diagnosis**:

```bash
# Check memory usage
ps aux | grep radarr
free -h
top -p $(pgrep radarr)
```

**Solutions**:

1. **Reduce connection pool size**:

   ```yaml
   database:
     max_connections: 10
     max_idle_connections: 2
   ```

2. **Tune garbage collection**:

   ```bash
   export GOGC=50  # More aggressive GC
   ./radarr
   ```

3. **Check for memory leaks**:

   ```bash
   # Monitor over time
   watch 'ps aux | grep radarr | grep -v grep'
   ```

#### Slow API Responses

**Diagnosis**:

```bash
# Test API response times
time curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/movie"

# Check database performance
EXPLAIN ANALYZE SELECT * FROM movies LIMIT 10;
```

**Solutions**:

1. **Database optimization**:

   ```sql
   -- PostgreSQL
   VACUUM ANALYZE movies;
   REINDEX TABLE movies;

   -- Create missing indexes
   CREATE INDEX CONCURRENTLY idx_movies_title ON movies (title);
   ```

2. **Increase worker count**:

   ```yaml
   performance:
     worker_count: 8  # Match CPU cores
   ```

3. **Optimize logging**:

   ```yaml
   log:
     level: "warn"  # Reduce verbosity
   ```

### Diagnostic Commands

#### System Information

```bash
# Version check
./radarr --version

# System status
curl -s -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/system/status" | jq

# Health check
curl -s -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/health" | jq

# Log analysis
tail -f /var/log/radarr/radarr.log | jq -r '.msg'
```

#### Database Health

```bash
# PostgreSQL
psql -h localhost -U radarr -d radarr -c "
  SELECT schemaname,tablename,attname,n_distinct,correlation
  FROM pg_stats
  WHERE tablename IN ('movies','movie_files')
  ORDER BY tablename,attname;"

# MariaDB
mysql -h localhost -u radarr -p radarr -e "
  SELECT TABLE_NAME, ENGINE, TABLE_ROWS, DATA_LENGTH, INDEX_LENGTH
  FROM information_schema.TABLES
  WHERE TABLE_SCHEMA='radarr';"
```

### Getting Help

#### Log Collection for Support

```bash
#!/bin/bash
# collect-logs.sh - Gather diagnostic information

RADARR_VERSION=$(./radarr --version)
SYSTEM_INFO=$(uname -a)
MEMORY_INFO=$(free -h)
DISK_INFO=$(df -h)

echo "=== Radarr Go Diagnostic Report ===" > diagnostic.txt
echo "Version: $RADARR_VERSION" >> diagnostic.txt
echo "System: $SYSTEM_INFO" >> diagnostic.txt
echo "Memory: $MEMORY_INFO" >> diagnostic.txt
echo "Disk: $DISK_INFO" >> diagnostic.txt
echo "" >> diagnostic.txt

# Recent logs
echo "=== Recent Logs ===" >> diagnostic.txt
tail -100 /var/log/radarr/radarr.log >> diagnostic.txt

# Configuration (sanitized)
echo "=== Configuration (Sanitized) ===" >> diagnostic.txt
grep -v -i "password\|key\|secret" /etc/radarr/config.yaml >> diagnostic.txt

# System status
echo "=== System Status ===" >> diagnostic.txt
curl -s -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/system/status" >> diagnostic.txt

echo "Diagnostic report saved to diagnostic.txt"
```

#### Support Channels

- **GitHub Issues**: https://github.com/radarr/radarr-go/issues
- **Documentation**: README.md, VERSIONING.md, MIGRATION.md
- **Community**: Discord/Reddit (links in main repository)

When reporting issues, include:

1. Radarr-go version (`./radarr --version`)
2. Operating system and architecture
3. Database type and version
4. Configuration file (sanitized)
5. Recent log entries
6. Steps to reproduce

---

## Success Validation

### Quick Health Check

After installation, verify everything is working:

```bash
# 1. Check application is running
curl http://localhost:7878/ping
# Expected: "pong"

# 2. Verify database connection
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/system/status"
# Expected: JSON with database status

# 3. Test API functionality
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/movie" | jq length
# Expected: Number (movie count)

# 4. Check health status
curl -H "X-Api-Key: your-api-key" "http://localhost:7878/api/v3/health"
# Expected: Array of health checks (empty is good)
```

### Performance Benchmarks

Expected performance improvements over original Radarr:

- **Memory Usage**: 60-80% reduction
- **Startup Time**: 80-90% faster
- **API Response**: <100ms average
- **Docker Image**: 95% smaller

### Production Readiness Checklist

- [ ] Database properly configured and optimized
- [ ] SSL/HTTPS enabled (if needed)
- [ ] Firewall configured
- [ ] Backups automated
- [ ] Monitoring configured
- [ ] Log rotation setup
- [ ] Service auto-start enabled
- [ ] API key configured
- [ ] TMDB API key added
- [ ] Movie directories accessible
- [ ] Download clients configured
- [ ] Indexers configured
- [ ] Notifications configured

---

**Installation Complete!**

Access Radarr-go at: **http://localhost:7878**

For additional configuration and advanced features, see the project documentation in the repository.
