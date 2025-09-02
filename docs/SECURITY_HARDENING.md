# Radarr Go Security Hardening Guide

**Version**: v0.9.0-alpha

## Overview

This guide provides comprehensive security hardening strategies for production Radarr Go deployments. Security is built into Radarr Go from the ground up with:

- **Secure Defaults** - Production-ready security configuration out of the box
- **Container Security** - Rootless containers with minimal attack surface
- **API Security** - Strong authentication and rate limiting
- **Network Security** - TLS/SSL encryption and network isolation
- **Data Protection** - Encrypted storage and secure backup strategies

## Security Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Internet                             │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                 Reverse Proxy                           │
│              (TLS Termination)                          │
│             Rate Limiting/WAF                           │
└─────────────────────────┬───────────────────────────────┘
                          │
┌─────────────────────────▼───────────────────────────────┐
│                Application Network                      │
│            (Isolated Network Segment)                   │
│  ┌─────────────────┐    ┌─────────────────┐            │
│  │   Radarr Go     │    │    Database     │            │
│  │  (Non-root)     │    │   (Encrypted)   │            │
│  └─────────────────┘    └─────────────────┘            │
└─────────────────────────────────────────────────────────┘
```

## Container Security Hardening

### Rootless Container Configuration

```yaml
# docker-compose.secure.yml
version: '3.8'

services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha
    container_name: radarr-go-secure
    restart: unless-stopped

    # Security Context - Non-root user
    user: "1000:1000"

    # Security Options
    security_opt:
      - no-new-privileges:true
      - apparmor:docker-default
      - seccomp:./security/seccomp-profile.json

    # Read-only Root Filesystem
    read_only: true

    # Temporary Filesystem Mounts (writable areas)
    tmpfs:
      - /tmp:noexec,nosuid,size=100m
      - /var/cache:noexec,nosuid,size=50m

    # Capability Dropping
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - SETGID
      - SETUID

    # Resource Limits (DoS protection)
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1'
          pids: 100
        reservations:
          memory: 256M
          cpus: '0.5'

    # Environment Variables
    environment:
      # Security Configuration
      - RADARR_AUTH_METHOD=apikey
      - RADARR_AUTH_API_KEY=${API_KEY}
      - RADARR_AUTH_REQUIRE_AUTHENTICATION_FOR_PING=true

      # Network Security
      - RADARR_SERVER_HOST=127.0.0.1  # Bind to localhost only
      - RADARR_SECURITY_ENABLE_CORS=false
      - RADARR_SECURITY_ENABLE_SECURITY_HEADERS=true
      - RADARR_SECURITY_MAX_REQUEST_SIZE=10MB

      # Logging for Security Monitoring
      - RADARR_LOG_LEVEL=info
      - RADARR_LOG_FORMAT=json
      - RADARR_DEVELOPMENT_ENABLE_DEBUG_ENDPOINTS=false

    # Volumes with Security Options
    volumes:
      - type: bind
        source: /opt/radarr/data
        target: /data
        bind:
          propagation: rslave
        read_only: false

      - type: bind
        source: /opt/radarr/config
        target: /app/config
        bind:
          propagation: rslave
        read_only: true

      - type: bind
        source: /mnt/movies
        target: /movies
        bind:
          propagation: rslave
        read_only: true

    # Network Configuration
    networks:
      - radarr-internal

    # Health Check
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://127.0.0.1:7878/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 10s

networks:
  radarr-internal:
    driver: bridge
    internal: true  # No external access
    driver_opts:
      com.docker.network.bridge.name: radarr-sec
      com.docker.network.bridge.enable_ip_masquerade: 'false'
```

### Seccomp Security Profile

Create `security/seccomp-profile.json`:

```json
{
  "defaultAction": "SCMP_ACT_ERRNO",
  "architectures": ["SCMP_ARCH_X86_64", "SCMP_ARCH_X86", "SCMP_ARCH_X32"],
  "syscalls": [
    {
      "names": [
        "accept", "accept4", "access", "arch_prctl", "bind", "brk", "close",
        "connect", "dup", "dup2", "epoll_create", "epoll_create1", "epoll_ctl",
        "epoll_wait", "exit", "exit_group", "fcntl", "fcntl64", "fstat", "fstat64",
        "futex", "getcwd", "getdents", "getdents64", "getegid", "geteuid", "getgid",
        "getpeername", "getpid", "getppid", "getrandom", "getrlimit", "getsockname",
        "getsockopt", "getuid", "listen", "lseek", "_llseek", "lstat", "lstat64",
        "madvise", "mmap", "mmap2", "mprotect", "munmap", "nanosleep", "open",
        "openat", "pipe", "pipe2", "poll", "pread64", "pwrite64", "read", "readv",
        "recvfrom", "recvmsg", "rt_sigaction", "rt_sigprocmask", "rt_sigreturn",
        "sched_getaffinity", "select", "sendmsg", "sendto", "setrlimit", "setsockopt",
        "shutdown", "socket", "stat", "stat64", "uname", "unlink", "write", "writev"
      ],
      "action": "SCMP_ACT_ALLOW"
    }
  ]
}
```

### AppArmor Profile

Create `security/apparmor-profile`:

```bash
# /etc/apparmor.d/radarr-go
#include <tunables/global>

profile radarr-go flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>
  #include <abstractions/nameservice>
  #include <abstractions/openssl>

  # Allow network access
  network inet stream,
  network inet6 stream,
  network unix stream,

  # Binary execution
  /radarr r,

  # Configuration files
  /app/config/ r,
  /app/config/** r,
  /app/migrations/ r,
  /app/migrations/** r,

  # Data directory
  /data/ rw,
  /data/** rw,

  # Movies directory (read-only)
  /movies/ r,
  /movies/** r,

  # System files (minimal access)
  /etc/passwd r,
  /etc/group r,
  /etc/nsswitch.conf r,
  /etc/resolv.conf r,
  /etc/ssl/certs/ r,
  /etc/ssl/certs/** r,

  # Temporary files
  /tmp/ rw,
  /tmp/** rw,

  # Deny dangerous operations
  deny /proc/sys/kernel/core_pattern w,
  deny /sys/kernel/security/ r,
  deny mount,
  deny umount,
  deny pivot_root,
  deny ptrace,
  deny @{PROC}/sys/kernel/modprobe w,
  deny @{PROC}/sys/vm/panic_on_oom w,
}
```

## Network Security Configuration

### Firewall Rules (iptables)

```bash
#!/bin/bash
# firewall-setup.sh - Configure firewall rules

set -euo pipefail

# Flush existing rules
iptables -F
iptables -X
iptables -t nat -F
iptables -t nat -X

# Default policies
iptables -P INPUT DROP
iptables -P FORWARD DROP
iptables -P OUTPUT ACCEPT

# Allow loopback
iptables -A INPUT -i lo -j ACCEPT
iptables -A OUTPUT -o lo -j ACCEPT

# Allow established connections
iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

# SSH access (change port as needed)
iptables -A INPUT -p tcp --dport 22 -m conntrack --ctstate NEW -m limit --limit 3/min --limit-burst 3 -j ACCEPT

# HTTPS access (reverse proxy)
iptables -A INPUT -p tcp --dport 443 -m conntrack --ctstate NEW -j ACCEPT
iptables -A INPUT -p tcp --dport 80 -m conntrack --ctstate NEW -j ACCEPT

# Database access (only from application network)
iptables -A INPUT -p tcp --dport 5432 -s 172.20.0.0/16 -j ACCEPT

# Rate limiting for application port (if directly exposed)
iptables -A INPUT -p tcp --dport 7878 -m limit --limit 25/min --limit-burst 100 -j ACCEPT

# Drop invalid packets
iptables -A INPUT -m conntrack --ctstate INVALID -j DROP

# Log dropped packets
iptables -A INPUT -m limit --limit 5/min -j LOG --log-prefix "iptables denied: "

# Save rules
iptables-save > /etc/iptables/rules.v4

echo "Firewall rules configured"
```

### Network Segmentation

```yaml
# docker-compose.network-segmented.yml
version: '3.8'

services:
  radarr-go:
    networks:
      - app-network
      - db-network

  postgres:
    networks:
      - db-network

  nginx-proxy:
    networks:
      - app-network
      - public-network
    ports:
      - "80:80"
      - "443:443"

networks:
  # Public-facing network
  public-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.18.0.0/16

  # Application network
  app-network:
    driver: bridge
    internal: false
    ipam:
      config:
        - subnet: 172.19.0.0/16

  # Database network (internal only)
  db-network:
    driver: bridge
    internal: true
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

### TLS/SSL Configuration

#### Nginx SSL Configuration

```nginx
# nginx-ssl.conf - Production SSL configuration
server {
    listen 443 ssl http2;
    server_name radarr.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/ssl/certs/radarr.yourdomain.com.crt;
    ssl_certificate_key /etc/ssl/private/radarr.yourdomain.com.key;

    # Modern SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;

    # SSL session optimization
    ssl_session_cache shared:SSL:50m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/ssl/certs/chain.pem;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;

    # Security Headers
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';" always;
    add_header Permissions-Policy "geolocation=(), microphone=(), camera=(), payment=(), usb=(), interest-cohort=()" always;

    # Hide server information
    server_tokens off;
    more_clear_headers Server;

    # Request size limits
    client_max_body_size 50M;
    client_body_timeout 60s;
    client_header_timeout 60s;

    # Rate limiting
    limit_req zone=api burst=10 nodelay;
    limit_conn conn_limit_per_ip 10;

    location / {
        proxy_pass http://127.0.0.1:7878;

        # Security headers for proxied requests
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;

        # Remove server headers
        proxy_hide_header X-Powered-By;
        proxy_hide_header Server;

        # Timeout configuration
        proxy_connect_timeout 30s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # Buffer configuration
        proxy_buffering on;
        proxy_buffer_size 16k;
        proxy_buffers 8 16k;
    }

    # Security endpoint (health check bypass)
    location /ping {
        proxy_pass http://127.0.0.1:7878/ping;
        access_log off;
        allow 127.0.0.1;
        allow 10.0.0.0/8;
        allow 172.16.0.0/12;
        allow 192.168.0.0/16;
        deny all;
    }

    # Block access to sensitive paths
    location ~ /\. {
        deny all;
        access_log off;
        log_not_found off;
    }

    location ~ ^/(config|data|backup)/ {
        deny all;
        access_log off;
        log_not_found off;
    }
}

# Rate limiting zones
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=30r/m;
    limit_conn_zone $binary_remote_addr zone=conn_limit_per_ip:10m;
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    server_name radarr.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

#### SSL Certificate Automation

Create `scripts/ssl-automation.sh`:

```bash
#!/bin/bash
# ssl-automation.sh - Automated SSL certificate management

set -euo pipefail

DOMAIN="${1:-radarr.yourdomain.com}"
EMAIL="${2:-admin@yourdomain.com}"
WEBROOT="${3:-/var/www/html}"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }
error() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1"; exit 1; }

# Install certbot
install_certbot() {
    if ! command -v certbot >/dev/null 2>&1; then
        log "Installing certbot..."
        if command -v apt-get >/dev/null 2>&1; then
            apt-get update
            apt-get install -y certbot python3-certbot-nginx
        elif command -v yum >/dev/null 2>&1; then
            yum install -y certbot python3-certbot-nginx
        else
            error "Unsupported package manager"
        fi
    fi
}

# Generate certificate
generate_certificate() {
    log "Generating SSL certificate for $DOMAIN..."

    certbot certonly \
        --webroot \
        --webroot-path="$WEBROOT" \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        --non-interactive \
        -d "$DOMAIN"

    if [ $? -eq 0 ]; then
        log "Certificate generated successfully"
    else
        error "Certificate generation failed"
    fi
}

# Set up auto-renewal
setup_renewal() {
    log "Setting up automatic renewal..."

    # Create renewal script
    cat > /etc/cron.daily/certbot-renew << 'EOF'
#!/bin/bash
certbot renew --quiet --post-hook "systemctl reload nginx"
EOF

    chmod +x /etc/cron.daily/certbot-renew

    # Test renewal
    certbot renew --dry-run

    log "Auto-renewal configured"
}

# Configure strong DH parameters
setup_dhparam() {
    log "Generating strong DH parameters..."

    if [ ! -f /etc/ssl/certs/dhparam.pem ]; then
        openssl dhparam -out /etc/ssl/certs/dhparam.pem 2048
    fi
}

# Main execution
main() {
    install_certbot
    generate_certificate
    setup_renewal
    setup_dhparam

    log "SSL configuration completed for $DOMAIN"
    log "Certificate location: /etc/letsencrypt/live/$DOMAIN/"
}

main "$@"
```

## Authentication and Authorization

### API Key Management

```yaml
# config.yaml - Secure authentication configuration
auth:
  # Strong API key authentication
  method: "apikey"
  api_key: "${API_KEY}"  # Must be 32+ characters, use secure generator

  # Additional security settings
  require_authentication_for_ping: true
  session_timeout: "2h"  # Shorter session timeout
  max_failed_attempts: 5
  lockout_duration: "15m"

  # API key rotation
  enable_key_rotation: true
  key_rotation_interval: "30d"
  previous_keys_valid_duration: "1h"

# Rate limiting
security:
  # Request rate limiting
  api_rate_limit: 60  # requests per minute per API key
  global_rate_limit: 1000  # requests per minute globally

  # Request size limits
  max_request_size: "10MB"
  max_header_size: "8KB"

  # Connection limits
  max_connections: 100
  connection_timeout: "30s"

  # Security headers
  enable_security_headers: true
  content_security_policy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none';"

  # CORS configuration (disable in production)
  enable_cors: false
  cors_origins: []
```

### Secure API Key Generation

Create `scripts/generate-api-key.sh`:

```bash
#!/bin/bash
# generate-api-key.sh - Generate secure API keys

set -euo pipefail

KEY_LENGTH=${1:-64}

# Generate cryptographically secure API key
generate_api_key() {
    local key_length=$1

    # Method 1: OpenSSL (preferred)
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -hex $((key_length / 2))
        return
    fi

    # Method 2: /dev/urandom
    if [ -r /dev/urandom ]; then
        LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$key_length"
        echo
        return
    fi

    echo "ERROR: Unable to generate secure random key" >&2
    exit 1
}

# Validate key strength
validate_key() {
    local key=$1
    local min_length=32

    if [ ${#key} -lt $min_length ]; then
        echo "ERROR: Key too short (minimum $min_length characters)" >&2
        exit 1
    fi

    # Check for sufficient entropy (basic check)
    local unique_chars=$(echo "$key" | fold -w1 | sort -u | wc -l)
    if [ "$unique_chars" -lt 16 ]; then
        echo "WARNING: Key may have low entropy (only $unique_chars unique characters)" >&2
    fi
}

# Generate and validate key
main() {
    echo "Generating secure API key..."

    local api_key
    api_key=$(generate_api_key "$KEY_LENGTH")

    validate_key "$api_key"

    echo "Generated API key (save securely):"
    echo "$api_key"
    echo
    echo "Add to your .env file:"
    echo "RADARR_AUTH_API_KEY=$api_key"
    echo
    echo "Key strength: ${#api_key} characters"
}

main "$@"
```

### Multi-Factor Authentication (Future Enhancement)

```yaml
# config.yaml - MFA configuration (roadmap feature)
auth:
  method: "apikey"

  # Future: Multi-factor authentication
  mfa:
    enabled: false  # Not yet implemented
    methods: ["totp", "backup_codes"]
    totp:
      issuer: "Radarr Go"
      period: 30
    backup_codes:
      count: 10
      length: 8
```

## Database Security

### PostgreSQL Security Hardening

```sql
-- postgres-security.sql - Database security configuration

-- Create dedicated user with minimal privileges
CREATE USER radarr_app WITH PASSWORD 'strong-password-here';

-- Create database
CREATE DATABASE radarr OWNER radarr_app;

-- Grant only necessary privileges
GRANT CONNECT ON DATABASE radarr TO radarr_app;
GRANT USAGE ON SCHEMA public TO radarr_app;
GRANT CREATE ON SCHEMA public TO radarr_app;

-- Revoke public access
REVOKE ALL ON DATABASE radarr FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;

-- Enable row-level security (if needed for multi-tenancy)
ALTER TABLE movies ENABLE ROW LEVEL SECURITY;
ALTER TABLE movie_files ENABLE ROW LEVEL SECURITY;

-- Audit logging
ALTER SYSTEM SET log_connections = 'on';
ALTER SYSTEM SET log_disconnections = 'on';
ALTER SYSTEM SET log_checkpoints = 'on';
ALTER SYSTEM SET log_lock_waits = 'on';
ALTER SYSTEM SET log_temp_files = 0;
ALTER SYSTEM SET log_autovacuum_min_duration = 0;

-- Connection security
ALTER SYSTEM SET ssl = 'on';
ALTER SYSTEM SET ssl_prefer_server_ciphers = 'on';
ALTER SYSTEM SET ssl_ciphers = 'HIGH:MEDIUM:+3DES:!aNULL:!SSLv2:!SSLv3';

-- Reload configuration
SELECT pg_reload_conf();
```

### Database Connection Security

```yaml
# docker-compose.db-secure.yml
version: '3.8'
services:
  postgres:
    image: postgres:17-alpine
    environment:
      - POSTGRES_DB=radarr
      - POSTGRES_USER=radarr_app
      - POSTGRES_PASSWORD=${DB_PASSWORD}
      - POSTGRES_INITDB_ARGS=--auth-host=scram-sha-256 --auth-local=scram-sha-256
    command:
      - postgres
      - -c
      - ssl=on
      - -c
      - ssl_cert_file=/etc/ssl/certs/postgres.crt
      - -c
      - ssl_key_file=/etc/ssl/private/postgres.key
      - -c
      - ssl_prefer_server_ciphers=on
      - -c
      - log_connections=on
      - -c
      - log_disconnections=on
      - -c
      - log_statement=ddl
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./ssl/postgres.crt:/etc/ssl/certs/postgres.crt:ro
      - ./ssl/postgres.key:/etc/ssl/private/postgres.key:ro
    networks:
      - db-network

    # Security hardening
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp
      - /var/run/postgresql
```

### Database Encryption

Create `scripts/db-encryption-setup.sh`:

```bash
#!/bin/bash
# db-encryption-setup.sh - Database encryption setup

set -euo pipefail

DB_NAME="${1:-radarr}"
BACKUP_DIR="${2:-/opt/radarr/encrypted-backups}"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }

# Setup database encryption
setup_database_encryption() {
    log "Setting up database encryption..."

    # PostgreSQL: Enable TDE (Transparent Data Encryption) if available
    # This is more relevant for enterprise PostgreSQL distributions

    # For standard PostgreSQL, we use application-level encryption
    # and encrypted backups

    mkdir -p "$BACKUP_DIR"
    chmod 700 "$BACKUP_DIR"
}

# Create encrypted backup
create_encrypted_backup() {
    log "Creating encrypted database backup..."

    local backup_file="$BACKUP_DIR/backup_$(date +%Y%m%d_%H%M%S).sql.gpg"

    # Create backup and encrypt
    PGPASSWORD="$POSTGRES_PASSWORD" pg_dump \
        -h "${POSTGRES_HOST:-localhost}" \
        -U "${POSTGRES_USER:-radarr}" \
        -d "$DB_NAME" | \
    gpg --cipher-algo AES256 --compress-algo 1 --symmetric --output "$backup_file"

    log "Encrypted backup created: $backup_file"
}

# Setup backup encryption keys
setup_encryption_keys() {
    log "Setting up encryption keys..."

    # Generate encryption key for backups
    if [ ! -f /etc/radarr/backup.key ]; then
        openssl rand -hex 32 > /etc/radarr/backup.key
        chmod 600 /etc/radarr/backup.key
        log "Backup encryption key generated"
    fi

    # Setup GPG for backup encryption
    if ! gpg --list-keys radarr-backup >/dev/null 2>&1; then
        gpg --batch --gen-key << 'EOF'
%no-protection
Key-Type: RSA
Key-Length: 4096
Subkey-Type: RSA
Subkey-Length: 4096
Name-Real: Radarr Backup
Name-Email: backup@radarr.local
Expire-Date: 2y
%commit
EOF
        log "GPG key for backups generated"
    fi
}

main() {
    setup_encryption_keys
    setup_database_encryption
    create_encrypted_backup
}

main "$@"
```

## Logging and Security Monitoring

### Security Event Logging

```yaml
# config.yaml - Security logging configuration
log:
  level: "info"
  format: "json"  # Structured logging for SIEM integration
  output: "stdout"

  # Security-specific logging
  enable_security_audit: true
  security_log_file: "/var/log/radarr/security.log"

  # Log sensitive events
  log_authentication_attempts: true
  log_authorization_failures: true
  log_api_key_usage: true
  log_configuration_changes: true
  log_file_access: true

  # Log rotation
  max_size: "100MB"
  max_backups: 10
  max_age: 90  # days
  compress: true

# Security monitoring
security:
  # Intrusion detection
  enable_intrusion_detection: true
  max_failed_requests: 10
  suspicious_activity_threshold: 50  # requests per minute

  # Audit trail
  enable_audit_trail: true
  audit_log_retention: "1y"

  # Alerting
  enable_security_alerts: true
  alert_on_brute_force: true
  alert_on_unusual_activity: true
```

### Log Monitoring with Fail2Ban

Create `security/fail2ban-radarr.conf`:

```ini
# /etc/fail2ban/jail.d/radarr.conf
[radarr-auth]
enabled = true
port = http,https
filter = radarr-auth
logpath = /var/log/radarr/security.log
maxretry = 5
bantime = 1800  # 30 minutes
findtime = 600  # 10 minutes
action = iptables[name=radarr, port=http, protocol=tcp]
         iptables[name=radarr, port=https, protocol=tcp]

[radarr-dos]
enabled = true
port = http,https
filter = radarr-dos
logpath = /var/log/radarr/access.log
maxretry = 100
bantime = 3600  # 1 hour
findtime = 300  # 5 minutes
action = iptables[name=radarr-dos, port=http, protocol=tcp]
```

Create `security/fail2ban-radarr-auth.conf`:

```ini
# /etc/fail2ban/filter.d/radarr-auth.conf
[Definition]
failregex = ^.*"level":"error".*"msg":"Authentication failed".*"remote_ip":"<HOST>".*$
            ^.*"level":"warn".*"msg":"Invalid API key".*"remote_ip":"<HOST>".*$
            ^.*"level":"warn".*"msg":"Rate limit exceeded".*"remote_ip":"<HOST>".*$

ignoreregex =
```

### SIEM Integration

Create `scripts/siem-integration.sh`:

```bash
#!/bin/bash
# siem-integration.sh - Security Information and Event Management integration

set -euo pipefail

SIEM_ENDPOINT="${SIEM_ENDPOINT:-https://siem.yourdomain.com/api/events}"
SIEM_API_KEY="${SIEM_API_KEY:-}"
LOG_FILE="/var/log/radarr/security.log"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }

# Send security events to SIEM
send_to_siem() {
    local event_data=$1

    if [ -n "$SIEM_API_KEY" ]; then
        curl -X POST "$SIEM_ENDPOINT" \
             -H "Authorization: Bearer $SIEM_API_KEY" \
             -H "Content-Type: application/json" \
             -d "$event_data" \
             --max-time 30 \
             --retry 3 || log "Failed to send event to SIEM"
    fi
}

# Monitor security log
monitor_security_log() {
    tail -F "$LOG_FILE" | while read -r line; do
        # Parse security events
        if echo "$line" | jq -e '.level == "error" or .level == "warn"' >/dev/null 2>&1; then
            local event_json=$(echo "$line" | jq -c '{
                timestamp: .timestamp,
                level: .level,
                message: .msg,
                source: "radarr-go",
                remote_ip: .remote_ip // null,
                user_agent: .user_agent // null,
                event_type: .event_type // "unknown"
            }')

            send_to_siem "$event_json"
        fi
    done
}

# Check for security indicators
check_security_indicators() {
    log "Checking for security indicators..."

    # Check for suspicious patterns in logs
    local suspicious_ips=$(grep -E "(Authentication failed|Invalid API key)" "$LOG_FILE" | \
                          jq -r '.remote_ip' | sort | uniq -c | sort -nr | head -5)

    if [ -n "$suspicious_ips" ]; then
        log "Suspicious IP addresses detected:"
        echo "$suspicious_ips"
    fi

    # Check for unusual activity patterns
    local hourly_requests=$(grep "$(date +'%Y-%m-%d %H')" "$LOG_FILE" | wc -l)
    if [ "$hourly_requests" -gt 1000 ]; then
        log "WARNING: High request volume detected: $hourly_requests requests this hour"
    fi
}

# Main monitoring loop
main() {
    log "Starting SIEM integration for Radarr Go..."

    if [ ! -f "$LOG_FILE" ]; then
        log "Security log file not found: $LOG_FILE"
        exit 1
    fi

    # Run security checks every hour
    while true; do
        check_security_indicators
        sleep 3600
    done &

    # Monitor logs in real-time
    monitor_security_log
}

main "$@"
```

## Backup and Disaster Recovery Security

### Encrypted Backup Strategy

Create `scripts/secure-backup.sh`:

```bash
#!/bin/bash
# secure-backup.sh - Secure backup with encryption

set -euo pipefail

BACKUP_DIR="/opt/radarr/backups"
ENCRYPTION_KEY_FILE="/etc/radarr/backup.key"
RETENTION_DAYS=30
S3_BUCKET="${RADARR_BACKUP_S3_BUCKET:-}"
GPG_RECIPIENT="${RADARR_BACKUP_GPG_RECIPIENT:-backup@radarr.local}"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }
error() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1"; exit 1; }

# Create encrypted database backup
backup_database() {
    log "Creating encrypted database backup..."

    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/db_backup_$timestamp.sql"
    local encrypted_file="$backup_file.gpg"

    # Create database dump
    PGPASSWORD="$POSTGRES_PASSWORD" pg_dump \
        -h "${POSTGRES_HOST:-localhost}" \
        -U "${POSTGRES_USER:-radarr}" \
        -d "${POSTGRES_DB:-radarr}" \
        -f "$backup_file"

    # Encrypt backup
    gpg --trust-model always --encrypt --recipient "$GPG_RECIPIENT" --output "$encrypted_file" "$backup_file"

    # Remove unencrypted backup
    rm "$backup_file"

    # Verify backup integrity
    if ! gpg --quiet --decrypt "$encrypted_file" >/dev/null 2>&1; then
        error "Backup encryption verification failed"
    fi

    log "Database backup created: $encrypted_file"
    echo "$encrypted_file"
}

# Create encrypted application data backup
backup_application_data() {
    log "Creating encrypted application data backup..."

    local timestamp=$(date +%Y%m%d_%H%M%S)
    local archive_file="$BACKUP_DIR/app_data_$timestamp.tar"
    local encrypted_file="$archive_file.gpg"

    # Create archive
    tar -cf "$archive_file" -C /opt/radarr data config

    # Encrypt archive
    gpg --trust-model always --encrypt --recipient "$GPG_RECIPIENT" --output "$encrypted_file" "$archive_file"

    # Remove unencrypted archive
    rm "$archive_file"

    log "Application data backup created: $encrypted_file"
    echo "$encrypted_file"
}

# Upload to secure cloud storage
upload_to_cloud() {
    local file_path=$1

    if [ -n "$S3_BUCKET" ]; then
        log "Uploading backup to S3..."

        # Use server-side encryption
        aws s3 cp "$file_path" "s3://$S3_BUCKET/$(basename "$file_path")" \
            --server-side-encryption AES256 \
            --storage-class STANDARD_IA || log "S3 upload failed"
    fi
}

# Cleanup old backups
cleanup_old_backups() {
    log "Cleaning up old backups..."

    find "$BACKUP_DIR" -name "*.gpg" -type f -mtime +$RETENTION_DAYS -delete

    if [ -n "$S3_BUCKET" ]; then
        # Cleanup old S3 backups
        aws s3 ls "s3://$S3_BUCKET/" | while read -r line; do
            local file_date=$(echo "$line" | awk '{print $1" "$2}')
            local file_name=$(echo "$line" | awk '{print $4}')
            local days_old=$(( ($(date +%s) - $(date -d "$file_date" +%s)) / 86400 ))

            if [ "$days_old" -gt $RETENTION_DAYS ]; then
                aws s3 rm "s3://$S3_BUCKET/$file_name"
                log "Deleted old backup: $file_name"
            fi
        done
    fi
}

# Main backup process
main() {
    log "Starting secure backup process..."

    # Check prerequisites
    [ -d "$BACKUP_DIR" ] || mkdir -p "$BACKUP_DIR"
    [ -f "$ENCRYPTION_KEY_FILE" ] || error "Encryption key file not found"
    command -v gpg >/dev/null 2>&1 || error "GPG not installed"

    # Create backups
    local db_backup
    local data_backup

    db_backup=$(backup_database)
    data_backup=$(backup_application_data)

    # Upload to cloud storage
    if [ -n "$S3_BUCKET" ]; then
        upload_to_cloud "$db_backup"
        upload_to_cloud "$data_backup"
    fi

    # Cleanup old backups
    cleanup_old_backups

    log "Backup process completed successfully"
}

main "$@"
```

### Disaster Recovery Testing

Create `scripts/dr-test.sh`:

```bash
#!/bin/bash
# dr-test.sh - Disaster recovery testing

set -euo pipefail

BACKUP_DIR="/opt/radarr/backups"
TEST_DIR="/tmp/radarr-dr-test"
GPG_RECIPIENT="${RADARR_BACKUP_GPG_RECIPIENT:-backup@radarr.local}"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }
error() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1"; exit 1; }

# Test backup decryption
test_backup_decryption() {
    log "Testing backup decryption..."

    local latest_backup=$(ls -t "$BACKUP_DIR"/*.sql.gpg | head -1)

    if [ -z "$latest_backup" ]; then
        error "No encrypted backups found"
    fi

    # Test decryption
    if gpg --quiet --decrypt "$latest_backup" > "$TEST_DIR/test-restore.sql" 2>/dev/null; then
        log "✓ Backup decryption successful"
        rm "$TEST_DIR/test-restore.sql"
    else
        error "✗ Backup decryption failed"
    fi
}

# Test database restore
test_database_restore() {
    log "Testing database restore (dry run)..."

    local latest_backup=$(ls -t "$BACKUP_DIR"/*.sql.gpg | head -1)
    local test_db="radarr_dr_test"

    # Create test database
    PGPASSWORD="$POSTGRES_PASSWORD" createdb \
        -h "${POSTGRES_HOST:-localhost}" \
        -U "${POSTGRES_USER:-radarr}" \
        "$test_db" || error "Failed to create test database"

    # Restore to test database
    gpg --quiet --decrypt "$latest_backup" | \
    PGPASSWORD="$POSTGRES_PASSWORD" psql \
        -h "${POSTGRES_HOST:-localhost}" \
        -U "${POSTGRES_USER:-radarr}" \
        -d "$test_db" >/dev/null 2>&1

    if [ $? -eq 0 ]; then
        log "✓ Database restore test successful"
    else
        error "✗ Database restore test failed"
    fi

    # Cleanup test database
    PGPASSWORD="$POSTGRES_PASSWORD" dropdb \
        -h "${POSTGRES_HOST:-localhost}" \
        -U "${POSTGRES_USER:-radarr}" \
        "$test_db"
}

# Test application data restore
test_application_restore() {
    log "Testing application data restore..."

    local latest_backup=$(ls -t "$BACKUP_DIR"/*app_data*.tar.gpg | head -1)
    local test_restore_dir="$TEST_DIR/app-restore"

    mkdir -p "$test_restore_dir"

    # Decrypt and extract
    gpg --quiet --decrypt "$latest_backup" | tar -xf - -C "$test_restore_dir"

    if [ -d "$test_restore_dir/data" ] && [ -d "$test_restore_dir/config" ]; then
        log "✓ Application data restore test successful"
    else
        error "✗ Application data restore test failed"
    fi

    rm -rf "$test_restore_dir"
}

# Generate DR report
generate_dr_report() {
    log "Generating disaster recovery report..."

    local report_file="$TEST_DIR/dr-report-$(date +%Y%m%d).txt"

    cat > "$report_file" << EOF
Disaster Recovery Test Report
============================
Date: $(date)
Tested by: $(whoami)
Environment: Production Backup Testing

Test Results:
- Backup Decryption: PASSED
- Database Restore: PASSED
- Application Data Restore: PASSED

Backup Status:
- Latest Database Backup: $(ls -t "$BACKUP_DIR"/*.sql.gpg | head -1)
- Latest App Data Backup: $(ls -t "$BACKUP_DIR"/*app_data*.tar.gpg | head -1)
- Total Backups: $(ls "$BACKUP_DIR"/*.gpg | wc -l)

Recommendations:
- All backup tests passed successfully
- Backup integrity verified
- Recovery procedures validated

Next Test: $(date -d '+1 month')
EOF

    log "DR report generated: $report_file"
}

# Main DR test
main() {
    log "Starting disaster recovery test..."

    mkdir -p "$TEST_DIR"

    test_backup_decryption
    test_database_restore
    test_application_restore
    generate_dr_report

    rm -rf "$TEST_DIR"

    log "Disaster recovery test completed successfully"
}

main "$@"
```

## Security Compliance and Auditing

### Security Compliance Checklist

Create `security/compliance-checklist.md`:

```markdown
# Radarr Go Security Compliance Checklist

## Authentication and Access Control
- [ ] Strong API keys (32+ characters)
- [ ] API key rotation enabled
- [ ] Rate limiting configured
- [ ] Failed login attempt monitoring
- [ ] Session timeout configured
- [ ] Privilege separation implemented

## Network Security
- [ ] TLS/SSL encryption enabled
- [ ] Strong cipher suites configured
- [ ] HSTS headers enabled
- [ ] Network segmentation implemented
- [ ] Firewall rules configured
- [ ] Port exposure minimized

## Container Security
- [ ] Non-root containers
- [ ] Read-only root filesystem
- [ ] Capability dropping
- [ ] Seccomp profile applied
- [ ] AppArmor/SELinux enabled
- [ ] Resource limits set

## Data Protection
- [ ] Database encryption enabled
- [ ] Backup encryption configured
- [ ] Secure key management
- [ ] Data retention policies
- [ ] Secure data disposal
- [ ] Access logging enabled

## Monitoring and Logging
- [ ] Security event logging
- [ ] Log integrity protection
- [ ] SIEM integration
- [ ] Anomaly detection
- [ ] Incident response plan
- [ ] Regular security audits

## Backup and Recovery
- [ ] Encrypted backups
- [ ] Backup integrity verification
- [ ] Disaster recovery testing
- [ ] Off-site backup storage
- [ ] Recovery time objectives defined
- [ ] Recovery procedures documented

## Vulnerability Management
- [ ] Regular security updates
- [ ] Vulnerability scanning
- [ ] Penetration testing
- [ ] Security patch management
- [ ] Dependency monitoring
- [ ] Security advisory monitoring
```

### Automated Security Scanning

Create `scripts/security-scan.sh`:

```bash
#!/bin/bash
# security-scan.sh - Automated security scanning

set -euo pipefail

SCAN_RESULTS_DIR="/var/log/radarr/security-scans"
DOCKER_IMAGE="${RADARR_DOCKER_IMAGE:-ghcr.io/radarr/radarr-go:v0.9.0-alpha}"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }

# Container vulnerability scanning
scan_container_vulnerabilities() {
    log "Scanning container for vulnerabilities..."

    if command -v trivy >/dev/null 2>&1; then
        trivy image "$DOCKER_IMAGE" --format json --output "$SCAN_RESULTS_DIR/container-vulns.json"

        local high_vulns=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH")] | length' "$SCAN_RESULTS_DIR/container-vulns.json")
        local critical_vulns=$(jq '[.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL")] | length' "$SCAN_RESULTS_DIR/container-vulns.json")

        log "Container scan complete: $critical_vulns critical, $high_vulns high vulnerabilities"
    else
        log "Trivy not installed, skipping container scan"
    fi
}

# Network security scanning
scan_network_security() {
    log "Scanning network security..."

    if command -v nmap >/dev/null 2>&1; then
        nmap -sV -sC --script vuln localhost -oX "$SCAN_RESULTS_DIR/network-scan.xml"
        log "Network scan completed"
    else
        log "Nmap not installed, skipping network scan"
    fi
}

# SSL/TLS security testing
test_ssl_security() {
    log "Testing SSL/TLS security..."

    local domain="${RADARR_DOMAIN:-localhost}"

    if command -v testssl.sh >/dev/null 2>&1; then
        testssl.sh --jsonfile "$SCAN_RESULTS_DIR/ssl-test.json" "$domain:443"
        log "SSL/TLS test completed"
    elif command -v sslyze >/dev/null 2>&1; then
        sslyze --json_out "$SCAN_RESULTS_DIR/ssl-test.json" "$domain:443"
        log "SSL/TLS test completed"
    else
        log "SSL testing tools not installed"
    fi
}

# Configuration security audit
audit_configuration() {
    log "Auditing security configuration..."

    local audit_report="$SCAN_RESULTS_DIR/config-audit.json"

    # Check configuration security
    cat > "$audit_report" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "checks": [
    {
      "name": "API Authentication",
      "status": "$([ -n "${RADARR_AUTH_API_KEY:-}" ] && echo "PASS" || echo "FAIL")",
      "description": "API key authentication enabled"
    },
    {
      "name": "Debug Mode",
      "status": "$([ "${RADARR_DEVELOPMENT_ENABLE_DEBUG_ENDPOINTS:-false}" = "false" ] && echo "PASS" || echo "FAIL")",
      "description": "Debug endpoints disabled"
    },
    {
      "name": "CORS Security",
      "status": "$([ "${RADARR_SECURITY_ENABLE_CORS:-false}" = "false" ] && echo "PASS" || echo "FAIL")",
      "description": "CORS disabled for security"
    },
    {
      "name": "Security Headers",
      "status": "$([ "${RADARR_SECURITY_ENABLE_SECURITY_HEADERS:-true}" = "true" ] && echo "PASS" || echo "FAIL")",
      "description": "Security headers enabled"
    }
  ]
}
EOF

    log "Configuration audit completed"
}

# Generate security report
generate_security_report() {
    log "Generating comprehensive security report..."

    local report_file="$SCAN_RESULTS_DIR/security-report-$(date +%Y%m%d).html"

    cat > "$report_file" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Radarr Go Security Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .pass { color: green; }
        .fail { color: red; }
        .warn { color: orange; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>Radarr Go Security Report</h1>
    <p><strong>Generated:</strong> $(date)</p>

    <h2>Summary</h2>
    <ul>
        <li>Container vulnerabilities: See container-vulns.json</li>
        <li>Network security: See network-scan.xml</li>
        <li>SSL/TLS security: See ssl-test.json</li>
        <li>Configuration audit: See config-audit.json</li>
    </ul>

    <h2>Recommendations</h2>
    <ul>
        <li>Regular security updates</li>
        <li>Monitor security advisories</li>
        <li>Review access logs</li>
        <li>Test disaster recovery</li>
    </ul>

    <h2>Next Actions</h2>
    <ul>
        <li>Review scan results</li>
        <li>Address critical vulnerabilities</li>
        <li>Update security policies</li>
        <li>Schedule next security scan</li>
    </ul>
</body>
</html>
EOF

    log "Security report generated: $report_file"
}

# Main security scan
main() {
    log "Starting comprehensive security scan..."

    mkdir -p "$SCAN_RESULTS_DIR"

    scan_container_vulnerabilities
    scan_network_security
    test_ssl_security
    audit_configuration
    generate_security_report

    log "Security scan completed. Results in: $SCAN_RESULTS_DIR"
}

main "$@"
```

## Best Practices Summary

### Security Configuration Checklist

- [ ] **Container Security**: Non-root containers with minimal privileges
- [ ] **Network Security**: TLS encryption with strong ciphers
- [ ] **Authentication**: Strong API keys with rotation
- [ ] **Authorization**: Principle of least privilege
- [ ] **Data Protection**: Encrypted backups and data at rest
- [ ] **Monitoring**: Security event logging and SIEM integration
- [ ] **Vulnerability Management**: Regular scanning and updates
- [ ] **Disaster Recovery**: Tested backup and recovery procedures
- [ ] **Compliance**: Regular security audits and assessments
- [ ] **Incident Response**: Documented procedures and contacts

### Security Maintenance Schedule

| Task | Frequency | Responsibility |
|------|-----------|---------------|
| Security updates | Weekly | Operations Team |
| Vulnerability scans | Weekly | Security Team |
| Access log review | Daily | Operations Team |
| Backup testing | Monthly | Operations Team |
| Disaster recovery test | Quarterly | Operations Team |
| Security audit | Annually | Security Team |
| Penetration testing | Annually | External Auditor |

This security hardening guide provides comprehensive protection for production Radarr Go deployments with enterprise-grade security controls and monitoring capabilities.
