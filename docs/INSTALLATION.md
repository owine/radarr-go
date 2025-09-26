# Radarr Go Installation and Setup Guide

**Version**: v0.9.0-alpha
**Last Updated**: September 2025

**Important**: This guide covers installation of Radarr Go v0.9.0-alpha and later versions. For version migration guidance, see [MIGRATION.md](../MIGRATION.md). For complete versioning information, see [VERSIONING.md](../VERSIONING.md).

## Table of Contents

- [Quick Start](#quick-start)
- [System Requirements](#system-requirements)
- [Installation Methods](#installation-methods)
  - [Docker Installation (Recommended)](#docker-installation-recommended)
  - [Binary Installation](#binary-installation)
  - [Source Installation](#source-installation)
- [Database Setup](#database-setup)
  - [PostgreSQL Setup](#postgresql-setup-recommended)
  - [MariaDB Setup](#mariadb-setup)
- [Initial Configuration](#initial-configuration)
- [System Service Setup](#system-service-setup)
  - [Linux Systemd](#linux-systemd)
  - [Windows Service](#windows-service)
  - [macOS LaunchDaemon](#macos-launchdaemon)
- [Migration from Original Radarr](#migration-from-original-radarr)
- [Security Configuration](#security-configuration)
- [Performance Tuning](#performance-tuning)
- [Troubleshooting](#troubleshooting)

## Quick Start

The fastest way to get Radarr Go running:

```bash
# Option 1: Docker (Recommended)
curl -o docker-compose.yml https://raw.githubusercontent.com/radarr/radarr-go/main/docker-compose.yml
docker-compose up -d

# Option 2: Binary
wget https://github.com/radarr/radarr-go/releases/latest/download/radarr-linux-amd64.tar.gz
tar -xzf radarr-linux-amd64.tar.gz
./radarr-linux-amd64 --data ./data
```

Access the application at: `http://localhost:7878`

## System Requirements

### Minimum Requirements

- **CPU**: 1 core, 1GHz
- **RAM**: 512MB
- **Disk**: 1GB free space (plus space for movies)
- **Network**: Internet connection for movie metadata

### Recommended Requirements

- **CPU**: 2+ cores, 2GHz+
- **RAM**: 2GB+
- **Disk**: 10GB+ free space (SSD preferred)
- **Network**: Broadband internet connection

### Supported Platforms

**Operating Systems**:

- Linux (Ubuntu 20.04+, RHEL 8+, Debian 11+, Alpine 3.15+)
- macOS 11.0+ (Big Sur)
- Windows 10/11, Windows Server 2019/2022
- FreeBSD 13.0+

**Architectures**:

- x86_64 (AMD64)
- ARM64 (AArch64)

**Database Support**:

- PostgreSQL 16+ (Recommended)
- MariaDB 10.5+ / MySQL 8.0+

## Installation Methods

### Docker Installation (Recommended)

Docker provides the easiest deployment with automatic dependency management and consistent behavior across platforms.

#### Simple Docker Run

```bash
# Create directories
mkdir -p ./data ./movies

# Run with PostgreSQL (recommended)
docker run -d \
  --name radarr-go \
  --restart unless-stopped \
  -p 7878:7878 \
  -v $(pwd)/data:/data \
  -v $(pwd)/movies:/movies \
  -e RADARR_DATABASE_TYPE=postgres \
  -e RADARR_DATABASE_HOST=your-postgres-host \
  -e RADARR_DATABASE_DATABASE=radarr \
  -e RADARR_DATABASE_USERNAME=radarr \
  -e RADARR_DATABASE_PASSWORD=your-password \
  -e RADARR_TMDB_API_KEY=your-tmdb-key \
  ghcr.io/radarr/radarr-go:v0.9.0-alpha
```

#### Docker Compose (Recommended)

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha  # Recommended: pin to specific version
    container_name: radarr-go
    restart: unless-stopped
    ports:
      - "7878:7878"
    volumes:
      - radarr_data:/data
      - ./movies:/movies
      - ./downloads:/downloads
    environment:
      # Database Configuration
      - RADARR_DATABASE_TYPE=postgres
      - RADARR_DATABASE_HOST=postgres
      - RADARR_DATABASE_DATABASE=radarr
      - RADARR_DATABASE_USERNAME=radarr
      - RADARR_DATABASE_PASSWORD=secure_password_here

      # TMDB Integration
      - RADARR_TMDB_API_KEY=your_tmdb_api_key_here

      # Security
      - RADARR_AUTH_METHOD=apikey
      - RADARR_AUTH_API_KEY=your_secure_api_key_here

      # Logging
      - RADARR_LOG_LEVEL=info

      # Performance
      - RADARR_DATABASE_MAX_CONNECTIONS=15
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:7878/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s

  postgres:
    image: postgres:17-alpine
    container_name: radarr-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_DB=radarr
      - POSTGRES_USER=radarr
      - POSTGRES_PASSWORD=secure_password_here
      - POSTGRES_INITDB_ARGS=--encoding=UTF-8 --lc-collate=C --lc-ctype=C
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U radarr"]
      interval: 10s
      timeout: 5s
      retries: 5

volumes:
  radarr_data:
  postgres_data:
```

Start the services:

```bash
# Start all services
docker-compose up -d

# Check logs
docker-compose logs -f radarr-go

# Stop services
docker-compose down
```

#### Docker with MariaDB

```yaml
version: '3.8'

services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha  # Recommended: pin to specific version
    container_name: radarr-go
    restart: unless-stopped
    ports:
      - "7878:7878"
    volumes:
      - radarr_data:/data
      - ./movies:/movies
    environment:
      - RADARR_DATABASE_TYPE=mariadb
      - RADARR_DATABASE_HOST=mariadb
      - RADARR_DATABASE_PORT=3306
      - RADARR_DATABASE_DATABASE=radarr
      - RADARR_DATABASE_USERNAME=radarr
      - RADARR_DATABASE_PASSWORD=secure_password_here
      - RADARR_TMDB_API_KEY=your_tmdb_api_key_here
    depends_on:
      - mariadb

  mariadb:
    image: mariadb:12.0.2
    container_name: radarr-mariadb
    restart: unless-stopped
    environment:
      - MYSQL_ROOT_PASSWORD=root_password_here
      - MYSQL_DATABASE=radarr
      - MYSQL_USER=radarr
      - MYSQL_PASSWORD=secure_password_here
      - MYSQL_CHARSET=utf8mb4
      - MYSQL_COLLATION=utf8mb4_unicode_ci
    volumes:
      - mariadb_data:/var/lib/mysql
    ports:
      - "3306:3306"
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

volumes:
  radarr_data:
  mariadb_data:
```

#### Docker Tag Strategy and Versioning

Radarr Go follows a comprehensive [versioning strategy](../VERSIONING.md) with automated Docker tag management:

##### Current Phase (Pre-1.0)

**Recommended for Testing/Development**:

- `:v0.9.0-alpha` - **Specific alpha version** (recommended for stability)
- `:testing` - Latest pre-release version (may include newer alphas/betas)
- `:alpha` - Latest alpha release

**Database-Optimized Tags**:

- `:v0.9.0-alpha-postgres` - PostgreSQL optimized (recommended)
- `:v0.9.0-alpha-mariadb` - MariaDB/MySQL optimized
- `:postgres` - Latest with PostgreSQL optimizations
- `:mariadb` - Latest with MariaDB optimizations

##### Future Production Tags (v1.0.0+)

Once v1.0.0 is released (Q2 2025), production tags will be available:

- `:latest` - Latest stable release (assigned starting with v1.0.0)
- `:stable` - Stable release pointer
- `:v1.0.0` - Immutable version pinning
- `:2025.04` - Calendar-based releases

##### Version Migration Strategy

- **Current users**: Use `:v0.9.0-alpha` for stability
- **Existing v0.0.x users**: See [MIGRATION.md](../MIGRATION.md) for upgrade path
- **Production deployment**: Wait for v1.0.0 or use `:v0.9.0-alpha` with caution

##### Security Recommendations

```bash
# Pin to specific digest for production-style deployment
docker pull ghcr.io/radarr/radarr-go@sha256:abc123...

# Check image digest
docker buildx imagetools inspect ghcr.io/radarr/radarr-go:v0.9.0-alpha
```

### Binary Installation

Binary installation provides maximum performance and minimal resource usage.

#### Download and Install

**Linux (x86_64):**

```bash
# Download latest release
wget https://github.com/radarr/radarr-go/releases/latest/download/radarr-linux-amd64.tar.gz
tar -xzf radarr-linux-amd64.tar.gz
sudo mv radarr-linux-amd64 /usr/local/bin/radarr
sudo chmod +x /usr/local/bin/radarr

# Create user and directories
sudo useradd --system --shell /bin/false --home-dir /var/lib/radarr --create-home radarr
sudo mkdir -p /etc/radarr /var/lib/radarr /var/log/radarr
sudo chown -R radarr:radarr /var/lib/radarr /var/log/radarr
```

**Linux (ARM64):**

```bash
wget https://github.com/radarr/radarr-go/releases/latest/download/radarr-linux-arm64.tar.gz
tar -xzf radarr-linux-arm64.tar.gz
sudo mv radarr-linux-arm64 /usr/local/bin/radarr
```

**macOS (Intel):**

```bash
wget https://github.com/radarr/radarr-go/releases/latest/download/radarr-darwin-amd64.tar.gz
tar -xzf radarr-darwin-amd64.tar.gz
sudo mv radarr-darwin-amd64 /usr/local/bin/radarr
```

**macOS (Apple Silicon):**

```bash
wget https://github.com/radarr/radarr-go/releases/latest/download/radarr-darwin-arm64.tar.gz
tar -xzf radarr-darwin-arm64.tar.gz
sudo mv radarr-darwin-arm64 /usr/local/bin/radarr
```

**Windows:**

```powershell
# Download from releases page or use PowerShell
Invoke-WebRequest -Uri "https://github.com/radarr/radarr-go/releases/latest/download/radarr-windows-amd64.zip" -OutFile "radarr-windows-amd64.zip"
Expand-Archive -Path "radarr-windows-amd64.zip" -DestinationPath "C:\Program Files\Radarr"
```

**FreeBSD:**

```bash
wget https://github.com/radarr/radarr-go/releases/latest/download/radarr-freebsd-amd64.tar.gz
tar -xzf radarr-freebsd-amd64.tar.gz
sudo mv radarr-freebsd-amd64 /usr/local/bin/radarr
```

#### Verify Installation

```bash
# Check version
radarr --version

# Check help
radarr --help
```

### Source Installation

Building from source provides the latest features and allows customization.

#### Prerequisites

```bash
# Install Go 1.25+
wget https://golang.org/dl/go1.25.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.25.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install build tools
sudo apt update && sudo apt install -y git make build-essential

# For macOS
brew install git make go
```

#### Build Process

```bash
# Clone repository
git clone https://github.com/radarr/radarr-go.git
cd radarr-go

# Install dependencies
make deps

# Build for production
make build-all

# Or build for current platform
make build

# Install binary
sudo cp radarr /usr/local/bin/
```

#### Development Build

```bash
# Setup development environment
make setup

# Install pre-commit hooks (optional but recommended)
pip install pre-commit
pre-commit install

# Run with hot reload
make dev

# Run tests
make test

# Run comprehensive quality checks
make all
```

## Database Setup

### PostgreSQL Setup (Recommended)

PostgreSQL is the recommended database for production deployments due to its reliability, performance, and advanced features.

#### Ubuntu/Debian Installation

```bash
# Install PostgreSQL
sudo apt update
sudo apt install -y postgresql postgresql-contrib

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres psql << EOF
CREATE DATABASE radarr;
CREATE USER radarr WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE radarr TO radarr;
ALTER USER radarr CREATEDB;
\q
EOF
```

#### CentOS/RHEL Installation

```bash
# Install PostgreSQL repository
sudo dnf install -y postgresql-server postgresql-contrib

# Initialize database
sudo postgresql-setup --initdb

# Start and enable PostgreSQL
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user (same as above)
```

#### macOS Installation

```bash
# Using Homebrew
brew install postgresql@17
brew services start postgresql@17

# Create database and user
createdb radarr
psql radarr << EOF
CREATE USER radarr WITH ENCRYPTED PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE radarr TO radarr;
\q
EOF
```

#### PostgreSQL Configuration

Edit `/etc/postgresql/17/main/postgresql.conf`:

```ini
# Performance tuning
shared_buffers = 256MB                  # 25% of RAM for dedicated server
effective_cache_size = 1GB              # 75% of RAM
work_mem = 4MB                          # For complex queries
maintenance_work_mem = 64MB             # For maintenance operations
random_page_cost = 1.1                  # For SSD storage
effective_io_concurrency = 200          # For SSD storage

# Connection settings
max_connections = 100
listen_addresses = '*'                  # Only if remote access needed
port = 5432

# Logging (for debugging)
log_statement = 'none'                  # Set to 'all' for debugging
log_min_duration_statement = 1000       # Log slow queries > 1s
```

Edit `/etc/postgresql/17/main/pg_hba.conf` for authentication:

```ini
# Local connections
local   all             radarr                                  md5
host    radarr          radarr          127.0.0.1/32           md5
host    radarr          radarr          ::1/128                md5

# Remote connections (if needed)
host    radarr          radarr          10.0.0.0/8             md5
```

Restart PostgreSQL:

```bash
sudo systemctl restart postgresql
```

#### Test PostgreSQL Connection

```bash
# Test local connection
psql -h localhost -U radarr -d radarr -c "SELECT version();"

# Test with Radarr Go
radarr --config config.yaml --test-db-connection
```

### MariaDB Setup

MariaDB offers excellent performance and is a suitable alternative to PostgreSQL.

#### Ubuntu/Debian Installation

```bash
# Install MariaDB
sudo apt update
sudo apt install -y mariadb-server mariadb-client

# Secure installation
sudo mysql_secure_installation

# Create database and user
sudo mysql << EOF
CREATE DATABASE radarr CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'radarr'@'localhost' IDENTIFIED BY 'your_secure_password';
GRANT ALL PRIVILEGES ON radarr.* TO 'radarr'@'localhost';
FLUSH PRIVILEGES;
EXIT;
EOF
```

#### CentOS/RHEL Installation

```bash
# Install MariaDB
sudo dnf install -y mariadb-server mariadb

# Start and enable MariaDB
sudo systemctl start mariadb
sudo systemctl enable mariadb

# Secure installation and create database (same as above)
```

#### MariaDB Configuration

Edit `/etc/mysql/mariadb.conf.d/50-server.cnf`:

```ini
[mysqld]
# Performance tuning
innodb_buffer_pool_size = 256M          # 70-80% of RAM for dedicated server
innodb_log_file_size = 64M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# Character set and collation
character-set-server = utf8mb4
collation-server = utf8mb4_unicode_ci

# Connection settings
max_connections = 100
connect_timeout = 60
wait_timeout = 600

# Query cache (disable for modern versions)
query_cache_size = 0
query_cache_type = 0

# Binary logging (optional, for replication/backup)
log-bin = mysql-bin
binlog_format = ROW
expire_logs_days = 7

# Slow query log (for debugging)
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
```

Restart MariaDB:

```bash
sudo systemctl restart mariadb
```

#### Test MariaDB Connection

```bash
# Test connection
mysql -h localhost -u radarr -p radarr -e "SELECT VERSION();"

# Test with Radarr Go
RADARR_DATABASE_TYPE=mariadb radarr --config config.yaml --test-db-connection
```

## Initial Configuration

### Basic Configuration File

Create `/etc/radarr/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 7878
  url_base: ""

database:
  type: "postgres"  # or "mariadb"
  host: "localhost"
  port: 5432        # 3306 for MariaDB
  database: "radarr"
  username: "radarr"
  password: "your_secure_password"
  max_connections: 15

log:
  level: "info"
  format: "json"
  output: "/var/log/radarr/radarr.log"

auth:
  method: "apikey"
  api_key: "your_secure_32_character_api_key"

storage:
  data_directory: "/var/lib/radarr"
  movie_directory: "/media/movies"
  backup_directory: "/var/lib/radarr/backups"

tmdb:
  api_key: "your_tmdb_api_key_here"
  language: "en-US"
  region: "US"

health:
  enabled: true
  interval: "5m"
  disk_space_warning_threshold: 2147483648   # 2GB
  disk_space_critical_threshold: 1073741824  # 1GB

performance:
  enable_response_caching: true
  cache_duration: "5m"
  connection_pool_size: 15

security:
  enable_cors: true
  cors_origins: ["*"]  # Restrict in production
  enable_security_headers: true
  max_request_size: "50MB"
```

### Environment Variables (Alternative)

Create `/etc/radarr/environment`:

```bash
# Server Configuration
RADARR_SERVER_HOST=0.0.0.0
RADARR_SERVER_PORT=7878

# Database Configuration
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST=localhost
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_DATABASE=radarr
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=your_secure_password
RADARR_DATABASE_MAX_CONNECTIONS=15

# Authentication
RADARR_AUTH_METHOD=apikey
RADARR_AUTH_API_KEY=your_secure_32_character_api_key

# TMDB Integration
RADARR_TMDB_API_KEY=your_tmdb_api_key_here

# Storage
RADARR_STORAGE_DATA_DIRECTORY=/var/lib/radarr
RADARR_STORAGE_MOVIE_DIRECTORY=/media/movies

# Logging
RADARR_LOG_LEVEL=info
RADARR_LOG_FORMAT=json
RADARR_LOG_OUTPUT=/var/log/radarr/radarr.log
```

### First Run

```bash
# Create necessary directories
sudo mkdir -p /var/lib/radarr /var/log/radarr /media/movies
sudo chown -R radarr:radarr /var/lib/radarr /var/log/radarr

# Test configuration
sudo -u radarr radarr --config /etc/radarr/config.yaml --test-config

# Start Radarr Go
sudo -u radarr radarr --config /etc/radarr/config.yaml --data /var/lib/radarr
```

Access the web interface at `http://your-server:7878`

## System Service Setup

### Linux Systemd

Create `/etc/systemd/system/radarr.service`:

```ini
[Unit]
Description=Radarr Go Movie Collection Manager
After=network.target postgresql.service mariadb.service
Wants=postgresql.service

[Service]
Type=simple
User=radarr
Group=radarr

# Binary location
ExecStart=/usr/local/bin/radarr --config /etc/radarr/config.yaml --data /var/lib/radarr

# Restart behavior
Restart=on-failure
RestartSec=10
StartLimitBurst=3
StartLimitIntervalSec=60

# Resource limits
LimitNOFILE=65536
MemoryHigh=1G
MemoryMax=2G

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/radarr /var/log/radarr /media/movies
CapabilityBoundingSet=
AmbientCapabilities=
ProtectKernelTunables=yes
ProtectKernelModules=yes
ProtectControlGroups=yes
RestrictRealtime=yes
RestrictNamespaces=yes
LockPersonality=yes
RemoveIPC=yes

# Environment
Environment=GOMAXPROCS=2
EnvironmentFile=-/etc/radarr/environment

# Working directory
WorkingDirectory=/var/lib/radarr

# Standard output/error
StandardOutput=journal
StandardError=journal
SyslogIdentifier=radarr

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable radarr.service

# Start service
sudo systemctl start radarr.service

# Check status
sudo systemctl status radarr.service

# View logs
sudo journalctl -u radarr.service -f
```

#### Service Management Commands

```bash
# Start/stop/restart
sudo systemctl start radarr.service
sudo systemctl stop radarr.service
sudo systemctl restart radarr.service

# Enable/disable auto-start
sudo systemctl enable radarr.service
sudo systemctl disable radarr.service

# Check service status
sudo systemctl status radarr.service
sudo systemctl is-active radarr.service
sudo systemctl is-enabled radarr.service

# View logs
sudo journalctl -u radarr.service -f          # Follow logs
sudo journalctl -u radarr.service --since today
sudo journalctl -u radarr.service -n 100      # Last 100 lines
```

### Windows Service

#### Using NSSM (Non-Sucking Service Manager)

1. **Download and Install NSSM**:

   ```powershell
   # Download from https://nssm.cc/download
   # Extract to C:\nssm
   ```

2. **Create Configuration File**:
   Create `C:\Program Files\Radarr\config.yaml`:

   ```yaml
   server:
     host: "0.0.0.0"
     port: 7878

   database:
     type: "postgres"
     host: "localhost"
     port: 5432
     database: "radarr"
     username: "radarr"
     password: "your_password"

   storage:
     data_directory: "C:\\ProgramData\\Radarr"
     movie_directory: "D:\\Movies"

   log:
     level: "info"
     format: "json"
     output: "C:\\ProgramData\\Radarr\\logs\\radarr.log"
   ```

3. **Install Service**:

   ```powershell
   # Run as Administrator
   C:\nssm\nssm.exe install Radarr

   # Configure service
   C:\nssm\nssm.exe set Radarr Application "C:\Program Files\Radarr\radarr.exe"
   C:\nssm\nssm.exe set Radarr AppParameters "--config C:\Program Files\Radarr\config.yaml --data C:\ProgramData\Radarr"
   C:\nssm\nssm.exe set Radarr AppDirectory "C:\Program Files\Radarr"
   C:\nssm\nssm.exe set Radarr DisplayName "Radarr Go"
   C:\nssm\nssm.exe set Radarr Description "Radarr Go Movie Collection Manager"
   C:\nssm\nssm.exe set Radarr Start SERVICE_AUTO_START

   # Set log files
   C:\nssm\nssm.exe set Radarr AppStdout "C:\ProgramData\Radarr\logs\service-output.log"
   C:\nssm\nssm.exe set Radarr AppStderr "C:\ProgramData\Radarr\logs\service-error.log"

   # Start service
   net start Radarr
   ```

#### Using sc.exe (Built-in Windows Tool)

```powershell
# Create service
sc.exe create Radarr binpath="C:\Program Files\Radarr\radarr.exe --config C:\Program Files\Radarr\config.yaml --data C:\ProgramData\Radarr" start=auto

# Configure service
sc.exe config Radarr displayname="Radarr Go"
sc.exe description Radarr "Radarr Go Movie Collection Manager"

# Start service
sc.exe start Radarr
```

#### Service Management

```powershell
# Start/stop service
net start Radarr
net stop Radarr

# Check service status
sc.exe query Radarr

# Remove service (if needed)
sc.exe delete Radarr
```

### macOS LaunchDaemon

Create `/Library/LaunchDaemons/com.radarr.radarr.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.radarr.radarr</string>

    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/radarr</string>
        <string>--config</string>
        <string>/usr/local/etc/radarr/config.yaml</string>
        <string>--data</string>
        <string>/usr/local/var/radarr</string>
    </array>

    <key>UserName</key>
    <string>radarr</string>
    <key>GroupName</key>
    <string>radarr</string>

    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>

    <key>StandardOutPath</key>
    <string>/usr/local/var/log/radarr/radarr.log</string>
    <key>StandardErrorPath</key>
    <string>/usr/local/var/log/radarr/radarr.error.log</string>

    <key>WorkingDirectory</key>
    <string>/usr/local/var/radarr</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin</string>
    </dict>
</dict>
</plist>
```

Load and start the service:

```bash
# Create user and directories
sudo dscl . -create /Users/radarr
sudo dscl . -create /Users/radarr UserShell /usr/bin/false
sudo dscl . -create /Users/radarr RealName "Radarr Service"
sudo dscl . -create /Users/radarr UniqueID 502
sudo dscl . -create /Users/radarr PrimaryGroupID 502

sudo mkdir -p /usr/local/var/radarr /usr/local/var/log/radarr /usr/local/etc/radarr
sudo chown -R radarr:radarr /usr/local/var/radarr /usr/local/var/log/radarr

# Load service
sudo launchctl load /Library/LaunchDaemons/com.radarr.radarr.plist

# Start service
sudo launchctl start com.radarr.radarr

# Check status
sudo launchctl list | grep radarr
```

## Migration from Original Radarr

### Pre-Migration Preparation

1. **Backup Original Radarr Data**:

   ```bash
   # Stop original Radarr service
   sudo systemctl stop radarr

   # Backup database and config
   cp -r ~/.config/Radarr ~/.config/Radarr.backup
   # OR for system installs
   cp -r /var/lib/radarr /var/lib/radarr.backup
   ```

2. **Extract Configuration**:
   - API Key from `config.xml`
   - Quality Profiles
   - Indexers configuration
   - Download Clients
   - Notification settings
   - Root folders and paths

### Database Migration

#### From SQLite to PostgreSQL

1. **Export SQLite data**:

   ```bash
   # Install sqlite3 tools
   sudo apt install sqlite3

   # Export schema and data
   sqlite3 ~/.config/Radarr/radarr.db .schema > radarr_schema.sql
   sqlite3 ~/.config/Radarr/radarr.db .dump > radarr_data.sql
   ```

2. **Convert to PostgreSQL**:

   ```bash
   # Use conversion tool (install first)
   pip install sqlite3-to-postgres

   # Convert database
   sqlite3-to-postgres \
     --sqlite-file ~/.config/Radarr/radarr.db \
     --postgres-dsn postgres://radarr:password@localhost:5432/radarr
   ```

3. **Manual Configuration Transfer**:

   ```bash
   # Extract important configuration
   sqlite3 ~/.config/Radarr/radarr.db "SELECT * FROM Config;" > config_export.csv
   sqlite3 ~/.config/Radarr/radarr.db "SELECT * FROM QualityProfiles;" > quality_profiles.csv
   sqlite3 ~/.config/Radarr/radarr.db "SELECT * FROM Indexers;" > indexers.csv
   ```

#### Configuration Mapping

**Original Radarr `config.xml` → Radarr Go `config.yaml`**:

```xml
<!-- Original config.xml -->
<Config>
  <Port>7878</Port>
  <UrlBase></UrlBase>
  <BindAddress>*</BindAddress>
  <LaunchBrowser>False</LaunchBrowser>
  <ApiKey>your-api-key</ApiKey>
</Config>
```

```yaml
# Radarr Go config.yaml
server:
  host: "0.0.0.0"
  port: 7878
  url_base: ""

auth:
  method: "apikey"
  api_key: "your-api-key"
```

### Step-by-Step Migration Process

1. **Setup Radarr Go with New Database**:

   ```bash
   # Install and configure Radarr Go (as per installation section)
   # Start with fresh PostgreSQL/MariaDB database
   radarr --config /etc/radarr/config.yaml --data /var/lib/radarr
   ```

2. **Configure Basic Settings**:
   - Set API Key (same as original)
   - Configure root folders
   - Set up media management settings

3. **Migrate Quality Profiles**:
   - Recreate quality profiles manually through API or web UI
   - Import quality definitions if needed

4. **Migrate Indexers**:

   ```bash
   # Use API to add indexers
   curl -X POST http://localhost:7878/api/v3/indexer \
     -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Your Indexer",
       "implementation": "TorrentImplementation",
       "settings": {
         "baseUrl": "https://indexer.example.com",
         "apiKey": "indexer-api-key"
       }
     }'
   ```

5. **Import Movie Library**:

   ```bash
   # Trigger library scan
   curl -X POST http://localhost:7878/api/v3/command \
     -H "X-API-Key: your-api-key" \
     -H "Content-Type: application/json" \
     -d '{"name": "RescanMovie"}'
   ```

6. **Verify Migration**:
   - Check all movies are imported
   - Verify quality profiles work
   - Test indexer connectivity
   - Confirm download client integration

### Migration Scripts

Create a migration helper script `/tmp/migrate-radarr.sh`:

```bash
#!/bin/bash
set -e

# Configuration
OLD_RADARR_DB="$HOME/.config/Radarr/radarr.db"
NEW_RADARR_API="http://localhost:7878/api/v3"
API_KEY="your-api-key-here"

# Check if old database exists
if [ ! -f "$OLD_RADARR_DB" ]; then
    echo "Original Radarr database not found at $OLD_RADARR_DB"
    exit 1
fi

echo "Starting migration from original Radarr to Radarr Go..."

# Extract movies list
sqlite3 "$OLD_RADARR_DB" "SELECT Title, ImdbId, TmdbId, Path FROM Movies;" > movies_list.csv

# Extract quality profiles
sqlite3 "$OLD_RADARR_DB" "SELECT Name, Items FROM QualityProfiles;" > quality_profiles.csv

# Extract indexers (be careful with sensitive data)
sqlite3 "$OLD_RADARR_DB" "SELECT Name, Implementation, Settings FROM Indexers;" > indexers.csv

echo "Data extracted. Please manually configure Radarr Go using the extracted data."
echo "Files created:"
echo "  - movies_list.csv"
echo "  - quality_profiles.csv"
echo "  - indexers.csv"
echo ""
echo "After configuring Radarr Go, run a library scan to import movies."
```

### Post-Migration Checklist

- [ ] All movies imported and properly tagged
- [ ] Quality profiles configured and working
- [ ] Indexers connected and searching properly
- [ ] Download clients configured and tested
- [ ] Notification systems working
- [ ] Calendar integration functional
- [ ] Backup original Radarr data (don't delete yet)
- [ ] Monitor logs for any errors
- [ ] Test full download workflow
- [ ] Update any external integrations (Sonarr, Overseerr, etc.)

## Security Configuration

### API Security

1. **Generate Strong API Key**:

   ```bash
   # Generate secure API key
   openssl rand -hex 16  # 32-character hex string

   # Or use Radarr Go's built-in generator
   radarr --generate-api-key
   ```

2. **Configure Authentication**:

   ```yaml
   auth:
     method: "apikey"
     api_key: "your-generated-secure-api-key"
     require_authentication_for_ping: true
   ```

3. **Restrict CORS Origins**:

   ```yaml
   security:
     enable_cors: true
     cors_origins:
       - "https://your-domain.com"
       - "http://localhost:8080"  # For local development
     cors_methods: ["GET", "POST", "PUT", "DELETE"]
   ```

### Network Security

1. **Bind to Specific Interface**:

   ```yaml
   server:
     host: "127.0.0.1"  # Localhost only
     # OR
     host: "10.0.0.10"  # Specific internal IP
   ```

2. **Enable SSL/TLS**:

   ```yaml
   server:
     enable_ssl: true
     ssl_cert_path: "/etc/ssl/certs/radarr.pem"
     ssl_key_path: "/etc/ssl/private/radarr.key"
   ```

3. **Generate SSL Certificate**:

   ```bash
   # Self-signed certificate (for testing)
   openssl req -x509 -newkey rsa:4096 -keyout radarr.key -out radarr.pem -days 365 -nodes

   # Let's Encrypt (for production)
   certbot certonly --standalone -d your-domain.com
   ```

### Firewall Configuration

**Ubuntu/Debian (ufw)**:

```bash
# Allow only specific IP ranges
sudo ufw allow from 10.0.0.0/8 to any port 7878
sudo ufw allow from 192.168.0.0/16 to any port 7878

# Or allow specific IPs
sudo ufw allow from 192.168.1.100 to any port 7878
```

**CentOS/RHEL (firewalld)**:

```bash
# Create custom service
sudo firewall-cmd --permanent --new-service=radarr
sudo firewall-cmd --permanent --service=radarr --set-short="Radarr Go"
sudo firewall-cmd --permanent --service=radarr --add-port=7878/tcp

# Allow from specific sources
sudo firewall-cmd --permanent --zone=internal --add-source=192.168.1.0/24
sudo firewall-cmd --permanent --zone=internal --add-service=radarr
sudo firewall-cmd --reload
```

### Reverse Proxy Configuration

#### Nginx

```nginx
server {
    listen 80;
    server_name radarr.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name radarr.yourdomain.com;

    ssl_certificate /etc/ssl/certs/radarr.pem;
    ssl_certificate_key /etc/ssl/private/radarr.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'" always;

    location / {
        proxy_pass http://127.0.0.1:7878;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;

        proxy_cache_bypass $http_upgrade;
        proxy_redirect off;
    }
}
```

#### Apache

```apache
<VirtualHost *:80>
    ServerName radarr.yourdomain.com
    Redirect permanent / https://radarr.yourdomain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName radarr.yourdomain.com

    SSLEngine on
    SSLCertificateFile /etc/ssl/certs/radarr.pem
    SSLCertificateKeyFile /etc/ssl/private/radarr.key
    SSLProtocol -all +TLSv1.2 +TLSv1.3

    # Security headers
    Header always set X-Frame-Options "SAMEORIGIN"
    Header always set X-XSS-Protection "1; mode=block"
    Header always set X-Content-Type-Options "nosniff"

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:7878/
    ProxyPassReverse / http://127.0.0.1:7878/

    ProxyPassReverse / http://127.0.0.1:7878/
    RequestHeader set "X-Forwarded-Proto" expr=%{REQUEST_SCHEME}
</VirtualHost>
```

#### Traefik (Docker)

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v3.0
    command:
      - "--api.insecure=true"
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.myresolver.acme.httpchallenge=true"
      - "--certificatesresolvers.myresolver.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.myresolver.acme.email=you@yourdomain.com"
      - "--certificatesresolvers.myresolver.acme.storage=/letsencrypt/acme.json"
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "./letsencrypt:/letsencrypt"

  radarr-go:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha  # Recommended: pin to specific version
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.radarr.rule=Host(`radarr.yourdomain.com`)"
      - "traefik.http.routers.radarr.entrypoints=websecure"
      - "traefik.http.routers.radarr.tls.certresolver=myresolver"
      - "traefik.http.services.radarr.loadbalancer.server.port=7878"
      # Redirect HTTP to HTTPS
      - "traefik.http.routers.radarr-http.rule=Host(`radarr.yourdomain.com`)"
      - "traefik.http.routers.radarr-http.entrypoints=web"
      - "traefik.http.routers.radarr-http.middlewares=redirect-to-https"
      - "traefik.http.middlewares.redirect-to-https.redirectscheme.scheme=https"
```

## Performance Tuning

### Database Performance

#### PostgreSQL Optimization

1. **Connection Pooling**:

   ```yaml
   database:
     type: "postgres"
     max_connections: 25
     connection_timeout: "10s"
     idle_timeout: "5m"
     max_lifetime: "1h"
   ```

2. **Query Optimization**:

   ```yaml
   database:
     enable_prepared_statements: true
     slow_query_threshold: "500ms"
     enable_query_logging: false  # Enable only for debugging
   ```

3. **PostgreSQL Server Settings** (`postgresql.conf`):

   ```ini
   # Memory settings (adjust based on available RAM)
   shared_buffers = 512MB                  # 25% of RAM
   effective_cache_size = 2GB              # 75% of RAM
   work_mem = 8MB                          # Per query
   maintenance_work_mem = 128MB            # For maintenance

   # Checkpoint settings
   checkpoint_segments = 32
   checkpoint_completion_target = 0.9
   wal_buffers = 16MB

   # Query planner
   random_page_cost = 1.1                  # For SSD
   effective_io_concurrency = 200          # For SSD
   ```

#### MariaDB Optimization

1. **Configuration**:

   ```yaml
   database:
     type: "mariadb"
     max_connections: 25
     charset: "utf8mb4"
     collation: "utf8mb4_unicode_ci"
   ```

2. **Server Settings** (`my.cnf`):

   ```ini
   [mysqld]
   # Memory settings
   innodb_buffer_pool_size = 512M          # 70-80% of RAM
   innodb_log_buffer_size = 32M
   query_cache_size = 0                    # Disable query cache
   tmp_table_size = 64M
   max_heap_table_size = 64M

   # InnoDB settings
   innodb_flush_method = O_DIRECT
   innodb_log_file_size = 128M
   innodb_flush_log_at_trx_commit = 2
   innodb_file_per_table = 1

   # Connection settings
   max_connections = 100
   connect_timeout = 60
   wait_timeout = 600
   ```

### Application Performance

1. **Response Caching**:

   ```yaml
   performance:
     enable_response_caching: true
     cache_duration: "5m"
     api_rate_limit: 200  # requests per minute
   ```

2. **File System Optimization**:

   ```yaml
   file_organization:
     use_hardlinks_instead_of_copy: true
     parallel_file_operations: 8
     skip_free_space_check: false
     minimum_free_space_when_importing: "1GB"
   ```

3. **Health Monitoring Optimization**:

   ```yaml
   health:
     interval: "10m"                       # Reduce frequency
     metrics_retention_days: 14            # Reduce storage
     checks:
       system_resources: false             # Disable if not needed
   ```

### System Level Optimization

1. **File System**:

   ```bash
   # For movie storage, use XFS or ext4
   mkfs.xfs -f /dev/sdb1
   mount -o noatime,nodiratime /dev/sdb1 /media/movies

   # Add to /etc/fstab
   /dev/sdb1 /media/movies xfs noatime,nodiratime 0 2
   ```

2. **Network Optimization**:

   ```bash
   # Increase network buffer sizes
   echo 'net.core.rmem_max = 134217728' >> /etc/sysctl.conf
   echo 'net.core.wmem_max = 134217728' >> /etc/sysctl.conf
   echo 'net.ipv4.tcp_rmem = 4096 16384 134217728' >> /etc/sysctl.conf
   echo 'net.ipv4.tcp_wmem = 4096 65536 134217728' >> /etc/sysctl.conf
   sysctl -p
   ```

3. **Memory Management**:

   ```bash
   # Adjust swappiness for better performance
   echo 'vm.swappiness = 10' >> /etc/sysctl.conf

   # Optimize file system cache
   echo 'vm.vfs_cache_pressure = 50' >> /etc/sysctl.conf
   ```

### Monitoring Performance

1. **Enable Metrics**:

   ```yaml
   health:
     enabled: true
     metrics_retention_days: 30
   ```

2. **Performance Monitoring Script**:

   ```bash
   #!/bin/bash
   # Monitor Radarr Go performance

   API_KEY="your-api-key"
   RADARR_URL="http://localhost:7878"

   # Check response time
   response_time=$(curl -w "%{time_total}" -s -o /dev/null \
     -H "X-API-Key: $API_KEY" \
     "$RADARR_URL/api/v3/system/status")

   echo "API Response Time: ${response_time}s"

   # Check database performance
   curl -s -H "X-API-Key: $API_KEY" \
     "$RADARR_URL/api/v3/health" | jq '.[] | select(.type == "database")'

   # Check disk usage
   df -h /media/movies
   ```

## Troubleshooting

### Common Installation Issues

#### 1. Database Connection Issues

**Symptoms**:

- "Connection refused" errors
- "Authentication failed" messages
- Application fails to start

**Solutions**:

```bash
# Test PostgreSQL connection
psql -h localhost -U radarr -d radarr -c "SELECT version();"

# Test MariaDB connection
mysql -h localhost -u radarr -p radarr -e "SELECT VERSION();"

# Check if database service is running
systemctl status postgresql  # or mariadb

# Verify configuration
radarr --config /etc/radarr/config.yaml --test-db-connection
```

**PostgreSQL specific**:

```bash
# Check PostgreSQL logs
sudo tail -f /var/log/postgresql/postgresql-17-main.log

# Verify pg_hba.conf allows connections
sudo grep -E "^(local|host)" /etc/postgresql/17/main/pg_hba.conf

# Test connection with detailed error
psql -h localhost -U radarr -d radarr -v ON_ERROR_STOP=1
```

#### 2. Permission Issues

**Symptoms**:

- "Permission denied" errors
- Files cannot be created or moved
- Database migration failures

**Solutions**:

```bash
# Fix ownership and permissions
sudo chown -R radarr:radarr /var/lib/radarr /var/log/radarr /media/movies
sudo chmod 755 /var/lib/radarr /media/movies
sudo chmod 644 /etc/radarr/config.yaml

# Check SELinux (CentOS/RHEL)
sestatus
sudo setsebool -P httpd_can_network_connect 1

# For movie directories on mounted drives
sudo mount -o uid=radarr,gid=radarr /dev/sdb1 /media/movies
```

#### 3. Port Binding Issues

**Symptoms**:

- "Address already in use" error
- Cannot start server on port 7878

**Solutions**:

```bash
# Check what's using the port
sudo netstat -tulnp | grep :7878
sudo lsof -i :7878

# Kill conflicting process
sudo kill -9 PID_NUMBER

# Use different port
RADARR_SERVER_PORT=7879 radarr --config config.yaml
```

#### 4. Memory Issues

**Symptoms**:

- Application crashes with "out of memory"
- Slow performance during large operations
- Database connection timeouts

**Solutions**:

```bash
# Check system memory
free -h
top -p $(pgrep radarr)

# Increase system service memory limit
# Edit /etc/systemd/system/radarr.service
MemoryHigh=2G
MemoryMax=4G

# Reduce database connections
RADARR_DATABASE_MAX_CONNECTIONS=5 radarr --config config.yaml
```

### Runtime Issues

#### 1. API Errors

**Symptoms**:

- 401 Unauthorized responses
- 500 Internal Server Error
- Timeout errors

**Diagnostics**:

```bash
# Check API health
curl -H "X-API-Key: your-api-key" http://localhost:7878/api/v3/health

# Test authentication
curl -v -H "X-API-Key: your-api-key" http://localhost:7878/api/v3/system/status

# Check logs for detailed errors
tail -f /var/log/radarr/radarr.log | jq '.'
```

**Solutions**:

```bash
# Verify API key configuration
grep -i api_key /etc/radarr/config.yaml

# Test with query parameter instead of header
curl "http://localhost:7878/api/v3/system/status?apikey=your-api-key"

# Restart service if needed
sudo systemctl restart radarr.service
```

#### 2. Movie Import Issues

**Symptoms**:

- Movies not found during scan
- Import failures
- Metadata not updated

**Diagnostics**:

```bash
# Check file permissions on movie directory
ls -la /media/movies/

# Test TMDB connectivity
curl "https://api.themoviedb.org/3/configuration?api_key=your-tmdb-key"

# Check import logs
grep -i "import" /var/log/radarr/radarr.log
```

**Solutions**:

```bash
# Verify TMDB API key
RADARR_TMDB_API_KEY=your-key radarr --test-tmdb

# Fix file permissions
sudo chown -R radarr:radarr /media/movies
sudo chmod -R 755 /media/movies

# Force library rescan
curl -X POST -H "X-API-Key: your-api-key" \
  "http://localhost:7878/api/v3/command" \
  -d '{"name": "RescanMovie"}'
```

#### 3. Performance Issues

**Symptoms**:

- Slow web interface
- Database query timeouts
- High CPU usage

**Diagnostics**:

```bash
# Check system resources
top -p $(pgrep radarr)
iostat -x 1

# Monitor database performance
# PostgreSQL:
sudo -u postgres psql -d radarr -c "SELECT * FROM pg_stat_activity;"

# MariaDB:
mysql -u radarr -p -e "SHOW PROCESSLIST;"
```

**Solutions**:

```bash
# Optimize database connections
RADARR_DATABASE_MAX_CONNECTIONS=10 radarr --config config.yaml

# Enable caching
RADARR_PERFORMANCE_ENABLE_RESPONSE_CACHING=true radarr --config config.yaml

# Reduce health check frequency
RADARR_HEALTH_INTERVAL=15m radarr --config config.yaml
```

### Docker-Specific Issues

#### 1. Container Startup Issues

**Symptoms**:

- Container exits immediately
- "No such file or directory" errors
- Volume mount issues

**Solutions**:

```bash
# Check container logs
docker logs radarr-go

# Verify volume mounts
docker inspect radarr-go | jq '.[0].Mounts'

# Fix volume permissions
sudo chown -R 1000:1000 ./data ./movies

# Use specific user ID
docker run -u 1000:1000 ghcr.io/radarr/radarr-go:latest
```

#### 2. Network Issues

**Symptoms**:

- Cannot connect to database container
- DNS resolution failures
- External API timeouts

**Solutions**:

```bash
# Test container networking
docker exec radarr-go ping postgres

# Check Docker network
docker network ls
docker network inspect bridge

# Use host networking for testing
docker run --network host ghcr.io/radarr/radarr-go:latest
```

### Log Analysis

#### Enable Debug Logging

```yaml
log:
  level: "debug"
  format: "text"  # Easier to read
  output: "/var/log/radarr/radarr-debug.log"

development:
  log_sql_queries: true
  enable_debug_endpoints: true
```

#### Log Analysis Scripts

```bash
#!/bin/bash
# Radarr Go log analysis

LOG_FILE="/var/log/radarr/radarr.log"

echo "=== Error Summary ==="
grep -i "error" "$LOG_FILE" | tail -10

echo -e "\n=== Database Issues ==="
grep -i "database\|sql" "$LOG_FILE" | grep -i "error\|timeout\|connection" | tail -5

echo -e "\n=== API Issues ==="
grep -i "api\|http" "$LOG_FILE" | grep -i "error\|timeout\|failed" | tail -5

echo -e "\n=== Performance Issues ==="
grep -i "slow\|timeout\|performance" "$LOG_FILE" | tail -5

echo -e "\n=== Recent Activity ==="
tail -20 "$LOG_FILE"
```

### Getting Help

1. **Check Documentation**:
   - [Configuration Reference](CONFIGURATION.md)
   - [API Documentation](API_ENDPOINTS.md)

2. **Enable Debug Mode**:

   ```yaml
   log:
     level: "debug"
   development:
     enable_debug_endpoints: true
   ```

3. **Collect System Information**:

   ```bash
   # System info
   uname -a
   cat /etc/os-release

   # Radarr Go version
   radarr --version

   # Configuration test
   radarr --config config.yaml --test-config

   # Resource usage
   free -h
   df -h

   # Network connectivity
   curl -I http://localhost:7878/ping
   ```

4. **Report Issues**:
   When reporting issues, include:
   - Radarr Go version
   - Operating system and version
   - Database type and version
   - Configuration file (remove sensitive data)
   - Relevant log entries
   - Steps to reproduce the issue

---

This comprehensive installation guide should enable users to successfully deploy radarr-go in production environments within 30 minutes for experienced users. The guide covers all major platforms, provides detailed troubleshooting information, and includes best practices for security and performance optimization.
