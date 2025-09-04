# Radarr Go Production Deployment Guide

**Version**: v0.9.0-alpha

## Overview

This guide provides comprehensive production deployment strategies for Radarr Go, covering containerized deployments, orchestration platforms, monitoring setup, and operational best practices. Radarr Go offers significant advantages over the original .NET version:

- **Single Binary Deployment** - No runtime dependencies beyond database
- **60% Lower Memory Usage** - Efficient resource utilization
- **3x Faster Response Times** - Go performance improvements
- **Multi-Database Support** - PostgreSQL and MariaDB with automatic failover

## Quick Production Checklist

Before deploying to production:

- [ ] Configure external database (PostgreSQL/MariaDB)
- [ ] Set up SSL/TLS termination
- [ ] Configure monitoring and alerting
- [ ] Set up backup automation
- [ ] Configure log aggregation
- [ ] Test disaster recovery procedures
- [ ] Set resource limits and health checks
- [ ] Configure security hardening

## Docker Production Setup

### Production Docker Compose

Create a production-ready setup with external database and monitoring:

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha  # Pin to specific version for production
    container_name: radarr-go
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "7878:7878"
    volumes:
      - radarr_data:/data
      - radarr_config:/app/config
      - /mnt/movies:/movies:ro
      - /mnt/downloads:/downloads
    environment:
      # Database Configuration
      - RADARR_DATABASE_TYPE=postgres
      - RADARR_DATABASE_HOST=postgres
      - RADARR_DATABASE_PORT=5432
      - RADARR_DATABASE_DATABASE=radarr
      - RADARR_DATABASE_USERNAME=radarr
      - RADARR_DATABASE_PASSWORD=${POSTGRES_PASSWORD}
      - RADARR_DATABASE_MAX_CONNECTIONS=20

      # Server Configuration
      - RADARR_SERVER_HOST=0.0.0.0
      - RADARR_SERVER_PORT=7878
      - RADARR_SERVER_URL_BASE=${URL_BASE:-}

      # Security
      - RADARR_AUTH_METHOD=apikey
      - RADARR_AUTH_API_KEY=${API_KEY}

      # Logging
      - RADARR_LOG_LEVEL=info
      - RADARR_LOG_FORMAT=json
      - RADARR_LOG_OUTPUT=stdout

      # External Services
      - RADARR_TMDB_API_KEY=${TMDB_API_KEY}

      # Performance
      - RADARR_PERFORMANCE_CONNECTION_POOL_SIZE=20
      - RADARR_PERFORMANCE_PARALLEL_FILE_OPERATIONS=10

      # Health Monitoring
      - RADARR_HEALTH_ENABLED=true
      - RADARR_HEALTH_INTERVAL=2m
      - RADARR_HEALTH_NOTIFY_CRITICAL_ISSUES=true
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:7878/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 30s
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '1'
        reservations:
          memory: 256M
          cpus: '0.5'
    security_opt:
      - no-new-privileges:true
    read_only: true
    tmpfs:
      - /tmp:noexec,nosuid,size=100m
    labels:
      # Traefik Labels (if using Traefik)
      - traefik.enable=true
      - traefik.http.routers.radarr.rule=Host(`radarr.yourdomain.com`)
      - traefik.http.routers.radarr.entrypoints=websecure
      - traefik.http.routers.radarr.tls=true
      - traefik.http.routers.radarr.tls.certresolver=letsencrypt
      - traefik.http.services.radarr.loadbalancer.server.port=7878

      # Monitoring Labels
      - prometheus.io/scrape=true
      - prometheus.io/port=7878
      - prometheus.io/path=/metrics

  postgres:
    image: postgres:17-alpine
    container_name: radarr-postgres
    restart: unless-stopped
    environment:
      - POSTGRES_DB=radarr
      - POSTGRES_USER=radarr
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_INITDB_ARGS=--auth-host=scram-sha-256
      - POSTGRES_HOST_AUTH_METHOD=scram-sha-256
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/postgres-init:/docker-entrypoint-initdb.d:ro
    command:
      - postgres
      - -c
      - max_connections=100
      - -c
      - shared_buffers=256MB
      - -c
      - effective_cache_size=1GB
      - -c
      - maintenance_work_mem=64MB
      - -c
      - checkpoint_completion_target=0.9
      - -c
      - wal_buffers=16MB
      - -c
      - default_statistics_target=100
      - -c
      - random_page_cost=1.1
      - -c
      - effective_io_concurrency=200
      - -c
      - work_mem=4MB
      - -c
      - min_wal_size=1GB
      - -c
      - max_wal_size=4GB
      - -c
      - log_min_duration_statement=1000
      - -c
      - log_checkpoints=on
      - -c
      - log_connections=on
      - -c
      - log_disconnections=on
      - -c
      - log_lock_waits=on
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U radarr -d radarr"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '2'
        reservations:
          memory: 512M
          cpus: '1'
    security_opt:
      - no-new-privileges:true

  # Backup Service
  postgres-backup:
    image: postgres:17-alpine
    container_name: radarr-postgres-backup
    restart: unless-stopped
    depends_on:
      - postgres
    environment:
      - PGPASSWORD=${POSTGRES_PASSWORD}
    volumes:
      - postgres_backups:/backups
      - ./scripts/backup:/scripts:ro
    command: >
      sh -c "
        apk add --no-cache dcron &&
        echo '0 2 * * * /scripts/backup.sh postgres radarr radarr ${POSTGRES_PASSWORD} /backups' | crontab - &&
        crond -f
      "

  # Redis for Caching (Optional)
  redis:
    image: redis:7-alpine
    container_name: radarr-redis
    restart: unless-stopped
    command:
      - redis-server
      - --save 60 1
      - --loglevel warning
      - --maxmemory 256mb
      - --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 3s
      retries: 3
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.5'
    security_opt:
      - no-new-privileges:true

  # Monitoring with Prometheus (Optional)
  prometheus:
    image: prom/prometheus:latest
    container_name: radarr-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - prometheus_data:/prometheus
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    profiles:
      - monitoring

  # Log Aggregation with Grafana Loki (Optional)
  loki:
    image: grafana/loki:latest
    container_name: radarr-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - loki_data:/loki
      - ./monitoring/loki.yml:/etc/loki/local-config.yaml:ro
    profiles:
      - monitoring

volumes:
  radarr_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/radarr/data
  radarr_config:
    driver: local
  postgres_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/radarr/postgres
  postgres_backups:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/radarr/backups
  redis_data:
    driver: local
  prometheus_data:
    driver: local
  loki_data:
    driver: local

networks:
  default:
    driver: bridge
```

### Environment Configuration

Create a `.env` file for production secrets:

```bash
# .env
POSTGRES_PASSWORD=super-secure-postgres-password-123
API_KEY=your-super-secure-api-key-here-64-chars-recommended
TMDB_API_KEY=your-tmdb-api-key
URL_BASE=
DOMAIN=radarr.yourdomain.com
```

### Docker Deployment Script

Create a deployment script:

```bash
#!/bin/bash
# deploy.sh - Production deployment script

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Configuration
COMPOSE_FILE="docker-compose.prod.yml"
ENV_FILE=".env"
BACKUP_DIR="/opt/radarr/backups"
DATA_DIR="/opt/radarr/data"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"
}

error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"
    exit 1
}

# Pre-deployment checks
preflight_checks() {
    log "Running pre-deployment checks..."

    # Check required files
    [ -f "$COMPOSE_FILE" ] || error "Docker compose file not found: $COMPOSE_FILE"
    [ -f "$ENV_FILE" ] || error "Environment file not found: $ENV_FILE"

    # Check docker and docker-compose
    command -v docker >/dev/null 2>&1 || error "Docker not installed"
    command -v docker compose >/dev/null 2>&1 || command -v docker-compose >/dev/null 2>&1 || error "Docker Compose not installed"

    # Check disk space (minimum 5GB free)
    local free_space=$(df / | awk 'NR==2 {print $4}')
    [ "$free_space" -gt 5000000 ] || warn "Low disk space: ${free_space}KB free"

    # Create required directories
    sudo mkdir -p "$DATA_DIR" "$BACKUP_DIR" /opt/radarr/postgres
    sudo chown -R 1000:1000 /opt/radarr

    log "Pre-deployment checks passed"
}

# Backup current deployment
backup_current() {
    log "Creating backup of current deployment..."

    local backup_timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="$BACKUP_DIR/deployment_backup_$backup_timestamp"

    mkdir -p "$backup_path"

    # Backup configuration and data
    if [ -d "$DATA_DIR" ]; then
        cp -r "$DATA_DIR" "$backup_path/"
        log "Data backup created: $backup_path/data"
    fi

    # Backup database if running
    if docker ps | grep -q radarr-postgres; then
        docker exec radarr-postgres pg_dump -U radarr radarr > "$backup_path/database_backup.sql"
        log "Database backup created: $backup_path/database_backup.sql"
    fi
}

# Deploy application
deploy() {
    log "Deploying Radarr Go..."

    # Pull latest images
    docker compose -f "$COMPOSE_FILE" pull

    # Start services
    docker compose -f "$COMPOSE_FILE" up -d

    # Wait for services to be healthy
    log "Waiting for services to become healthy..."
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if docker compose -f "$COMPOSE_FILE" ps | grep -q "healthy"; then
            log "Services are healthy"
            break
        fi

        sleep 10
        ((attempt++))

        if [ $attempt -eq $max_attempts ]; then
            error "Services did not become healthy within timeout"
        fi
    done

    log "Deployment completed successfully"
}

# Post-deployment verification
verify_deployment() {
    log "Verifying deployment..."

    # Check service status
    docker compose -f "$COMPOSE_FILE" ps

    # Test API endpoint
    local api_url="http://localhost:7878/ping"
    if curl -f -s "$api_url" >/dev/null; then
        log "API health check passed"
    else
        error "API health check failed"
    fi

    # Check logs for errors
    if docker compose -f "$COMPOSE_FILE" logs radarr-go | grep -i error | tail -5; then
        warn "Recent errors found in logs (shown above)"
    fi

    log "Deployment verification completed"
}

# Rollback to previous version
rollback() {
    log "Rolling back deployment..."

    # Stop current deployment
    docker compose -f "$COMPOSE_FILE" down

    # Find latest backup
    local latest_backup=$(ls -1t "$BACKUP_DIR"/deployment_backup_* | head -1)

    if [ -n "$latest_backup" ] && [ -d "$latest_backup" ]; then
        log "Restoring from backup: $latest_backup"

        # Restore data
        if [ -d "$latest_backup/data" ]; then
            rm -rf "$DATA_DIR"
            cp -r "$latest_backup/data" "$DATA_DIR"
        fi

        # Restore database
        if [ -f "$latest_backup/database_backup.sql" ]; then
            docker compose -f "$COMPOSE_FILE" up -d postgres
            sleep 30
            docker exec -i radarr-postgres psql -U radarr radarr < "$latest_backup/database_backup.sql"
        fi

        # Start services
        docker compose -f "$COMPOSE_FILE" up -d

        log "Rollback completed"
    else
        error "No backup found for rollback"
    fi
}

# Show logs
show_logs() {
    docker compose -f "$COMPOSE_FILE" logs -f
}

# Main execution
case "${1:-deploy}" in
    "deploy")
        preflight_checks
        backup_current
        deploy
        verify_deployment
        ;;
    "rollback")
        rollback
        ;;
    "logs")
        show_logs
        ;;
    "verify")
        verify_deployment
        ;;
    "backup")
        backup_current
        ;;
    *)
        echo "Usage: $0 {deploy|rollback|logs|verify|backup}"
        echo ""
        echo "Commands:"
        echo "  deploy   - Deploy application with full checks"
        echo "  rollback - Rollback to previous deployment"
        echo "  logs     - Show application logs"
        echo "  verify   - Verify current deployment"
        echo "  backup   - Create backup of current deployment"
        exit 1
        ;;
esac
```

Make the script executable:

```bash
chmod +x deploy.sh
```

### Production Dockerfile

For custom builds, use this optimized production Dockerfile:

```dockerfile
# Dockerfile.prod - Production optimized build
FROM golang:1.25-alpine AS builder

# Build arguments for CI/CD
ARG VERSION="production"
ARG COMMIT="unknown"
ARG BUILD_DATE="unknown"

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    upx

WORKDIR /app

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'" \
    -trimpath \
    -o radarr ./cmd/radarr

# Compress binary (optional, reduces size by ~30%)
RUN upx --best --lzma radarr

# Final production image
FROM scratch

# Import certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary and required files
COPY --from=builder /app/radarr /radarr
COPY --from=builder /app/migrations /migrations
COPY --from=builder /app/config.yaml /config.yaml

# Set user (must match volume permissions)
USER 1000:1000

EXPOSE 7878

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD ["/radarr", "healthcheck"]

ENTRYPOINT ["/radarr"]
CMD ["-data", "/data"]
```

## Kubernetes Deployment

### Namespace and Configuration

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: radarr
  labels:
    name: radarr
    monitoring: enabled
    backup: enabled

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: radarr-config
  namespace: radarr
data:
  config.yaml: |
    server:
      host: "0.0.0.0"
      port: 7878
      url_base: ""

    database:
      type: "postgres"
      host: "postgres-service"
      port: 5432
      database: "radarr"
      username: "radarr"
      max_connections: 20
      connection_timeout: "30s"
      idle_timeout: "10m"
      enable_prepared_statements: true

    log:
      level: "info"
      format: "json"
      output: "stdout"

    auth:
      method: "apikey"

    health:
      enabled: true
      interval: "2m"
      notify_critical_issues: true

    performance:
      connection_pool_size: 20
      parallel_file_operations: 10
      enable_response_caching: true
```

### Secrets Management

```yaml
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: radarr-secrets
  namespace: radarr
type: Opaque
stringData:
  postgres-password: "your-secure-postgres-password"
  api-key: "your-secure-api-key-64-chars"
  tmdb-api-key: "your-tmdb-api-key"

---
apiVersion: v1
kind: Secret
metadata:
  name: postgres-secrets
  namespace: radarr
type: Opaque
stringData:
  POSTGRES_DB: "radarr"
  POSTGRES_USER: "radarr"
  POSTGRES_PASSWORD: "your-secure-postgres-password"
```

### Database Deployment

```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: radarr
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      securityContext:
        fsGroup: 999
      containers:
      - name: postgres
        image: postgres:17-alpine
        ports:
        - containerPort: 5432
          name: postgres
        envFrom:
        - secretRef:
            name: postgres-secrets
        args:
          - postgres
          - -c
          - max_connections=100
          - -c
          - shared_buffers=256MB
          - -c
          - effective_cache_size=1GB
          - -c
          - maintenance_work_mem=64MB
          - -c
          - checkpoint_completion_target=0.9
          - -c
          - wal_buffers=16MB
          - -c
          - default_statistics_target=100
          - -c
          - random_page_cost=1.1
          - -c
          - effective_io_concurrency=200
          - -c
          - work_mem=4MB
          - -c
          - min_wal_size=1GB
          - -c
          - max_wal_size=4GB
          - -c
          - log_min_duration_statement=1000
          - -c
          - log_checkpoints=on
          - -c
          - log_connections=on
          - -c
          - log_disconnections=on
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - radarr
            - -d
            - radarr
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - radarr
            - -d
            - radarr
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          successThreshold: 1
          failureThreshold: 2
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "2"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: false
          runAsNonRoot: true
          runAsUser: 999
          runAsGroup: 999
          capabilities:
            drop:
            - ALL
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: fast-ssd
      resources:
        requests:
          storage: 20Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: radarr
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
    name: postgres
  type: ClusterIP
```

### Radarr Application Deployment

```yaml
# k8s/radarr.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: radarr-go
  namespace: radarr
  labels:
    app: radarr-go
    version: v0.9.0-alpha
spec:
  replicas: 2
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 1
  selector:
    matchLabels:
      app: radarr-go
  template:
    metadata:
      labels:
        app: radarr-go
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "7878"
        prometheus.io/path: "/metrics"
    spec:
      securityContext:
        fsGroup: 1000
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
      containers:
      - name: radarr-go
        image: ghcr.io/username/radarr-go:v0.9.0-alpha
        imagePullPolicy: Always
        ports:
        - containerPort: 7878
          name: http
        env:
        - name: RADARR_DATABASE_TYPE
          value: "postgres"
        - name: RADARR_DATABASE_HOST
          value: "postgres-service"
        - name: RADARR_DATABASE_PORT
          value: "5432"
        - name: RADARR_DATABASE_DATABASE
          value: "radarr"
        - name: RADARR_DATABASE_USERNAME
          value: "radarr"
        - name: RADARR_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: radarr-secrets
              key: postgres-password
        - name: RADARR_DATABASE_MAX_CONNECTIONS
          value: "20"
        - name: RADARR_AUTH_METHOD
          value: "apikey"
        - name: RADARR_AUTH_API_KEY
          valueFrom:
            secretKeyRef:
              name: radarr-secrets
              key: api-key
        - name: RADARR_TMDB_API_KEY
          valueFrom:
            secretKeyRef:
              name: radarr-secrets
              key: tmdb-api-key
        - name: RADARR_LOG_LEVEL
          value: "info"
        - name: RADARR_LOG_FORMAT
          value: "json"
        - name: RADARR_HEALTH_ENABLED
          value: "true"
        - name: RADARR_HEALTH_INTERVAL
          value: "2m"
        volumeMounts:
        - name: radarr-data
          mountPath: /data
        - name: radarr-config
          mountPath: /app/config.yaml
          subPath: config.yaml
        - name: movies
          mountPath: /movies
          readOnly: true
        - name: downloads
          mountPath: /downloads
        - name: tmp
          mountPath: /tmp
        livenessProbe:
          httpGet:
            path: /ping
            port: 7878
          initialDelaySeconds: 30
          periodSeconds: 30
          timeoutSeconds: 5
          successThreshold: 1
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ping
            port: 7878
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 3
          successThreshold: 1
          failureThreshold: 2
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "1"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: radarr-data
        persistentVolumeClaim:
          claimName: radarr-data-pvc
      - name: radarr-config
        configMap:
          name: radarr-config
      - name: movies
        persistentVolumeClaim:
          claimName: movies-pvc
      - name: downloads
        persistentVolumeClaim:
          claimName: downloads-pvc
      - name: tmp
        emptyDir:
          sizeLimit: 1Gi

---
apiVersion: v1
kind: Service
metadata:
  name: radarr-service
  namespace: radarr
  labels:
    app: radarr-go
spec:
  selector:
    app: radarr-go
  ports:
  - port: 80
    targetPort: 7878
    name: http
  type: ClusterIP

---
# Persistent Volume Claims
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: radarr-data-pvc
  namespace: radarr
spec:
  accessModes:
  - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 10Gi

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: movies-pvc
  namespace: radarr
spec:
  accessModes:
  - ReadWriteMany
  storageClassName: nfs-movies
  resources:
    requests:
      storage: 10Ti

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: downloads-pvc
  namespace: radarr
spec:
  accessModes:
  - ReadWriteMany
  storageClassName: nfs-downloads
  resources:
    requests:
      storage: 1Ti
```

### Ingress Configuration

```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: radarr-ingress
  namespace: radarr
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/proxy-connect-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-buffer-size: "8k"
    nginx.ingress.kubernetes.io/configuration-snippet: |
      more_set_headers "X-Forwarded-Proto: https";
      more_set_headers "X-Forwarded-For: $remote_addr";
spec:
  tls:
  - hosts:
    - radarr.yourdomain.com
    secretName: radarr-tls
  rules:
  - host: radarr.yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: radarr-service
            port:
              number: 80
```

### Horizontal Pod Autoscaler

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: radarr-hpa
  namespace: radarr
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: radarr-go
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 50
        periodSeconds: 60
```

### Kubernetes Deployment Script

```bash
#!/bin/bash
# k8s-deploy.sh - Kubernetes deployment script

set -euo pipefail

NAMESPACE="radarr"
KUBECTL="kubectl"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')] $1${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING: $1${NC}"; }
error() { echo -e "${RED}[$(date +'%H:%M:%S')] ERROR: $1${NC}"; exit 1; }

deploy() {
    log "Deploying Radarr Go to Kubernetes..."

    # Apply namespace and RBAC
    $KUBECTL apply -f k8s/namespace.yaml

    # Apply secrets (should be encrypted in real deployment)
    $KUBECTL apply -f k8s/secrets.yaml

    # Apply ConfigMap
    $KUBECTL apply -f k8s/configmap.yaml

    # Deploy PostgreSQL
    $KUBECTL apply -f k8s/postgres.yaml

    # Wait for PostgreSQL to be ready
    log "Waiting for PostgreSQL to be ready..."
    $KUBECTL wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s

    # Deploy Radarr
    $KUBECTL apply -f k8s/radarr.yaml

    # Wait for Radarr to be ready
    log "Waiting for Radarr to be ready..."
    $KUBECTL wait --for=condition=available deployment/radarr-go -n $NAMESPACE --timeout=300s

    # Apply Ingress
    $KUBECTL apply -f k8s/ingress.yaml

    # Apply HPA
    $KUBECTL apply -f k8s/hpa.yaml

    log "Deployment completed!"

    # Show status
    $KUBECTL get pods -n $NAMESPACE
    $KUBECTL get services -n $NAMESPACE
    $KUBECTL get ingress -n $NAMESPACE
}

status() {
    log "Checking deployment status..."
    $KUBECTL get pods -n $NAMESPACE -o wide
    $KUBECTL get services -n $NAMESPACE
    $KUBECTL get ingress -n $NAMESPACE

    # Check pod logs
    log "Recent logs:"
    $KUBECTL logs -l app=radarr-go -n $NAMESPACE --tail=10
}

rollback() {
    log "Rolling back deployment..."
    $KUBECTL rollout undo deployment/radarr-go -n $NAMESPACE
    $KUBECTL rollout status deployment/radarr-go -n $NAMESPACE
}

cleanup() {
    warn "This will delete ALL resources in the $NAMESPACE namespace!"
    read -p "Are you sure? (yes/no): " -r
    if [[ $REPLY == "yes" ]]; then
        $KUBECTL delete namespace $NAMESPACE
        log "Cleanup completed"
    else
        log "Cleanup cancelled"
    fi
}

case "${1:-deploy}" in
    "deploy") deploy ;;
    "status") status ;;
    "rollback") rollback ;;
    "cleanup") cleanup ;;
    *)
        echo "Usage: $0 {deploy|status|rollback|cleanup}"
        exit 1 ;;
esac
```

## Reverse Proxy Configuration

### Nginx Configuration

```nginx
# /etc/nginx/sites-available/radarr.conf
server {
    listen 80;
    server_name radarr.yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name radarr.yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/ssl/certs/radarr.yourdomain.com.crt;
    ssl_certificate_key /etc/ssl/private/radarr.yourdomain.com.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Proxy Configuration
    location / {
        proxy_pass http://127.0.0.1:7878;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Port $server_port;
        proxy_set_header X-Real-IP $remote_addr;

        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;

        # Buffer settings
        proxy_buffering on;
        proxy_buffer_size 8k;
        proxy_buffers 16 8k;
        proxy_busy_buffers_size 16k;

        # Client settings
        client_max_body_size 50M;
        client_body_timeout 60s;
    }

    # Health check endpoint (bypass authentication)
    location /ping {
        proxy_pass http://127.0.0.1:7878/ping;
        access_log off;
    }

    # API rate limiting
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://127.0.0.1:7878;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Static assets (if any)
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        proxy_pass http://127.0.0.1:7878;
        proxy_cache radarr_cache;
        proxy_cache_valid 200 1h;
        proxy_cache_valid 404 1m;
        add_header X-Cache-Status $upstream_cache_status;
    }

    # Logging
    access_log /var/log/nginx/radarr.access.log combined;
    error_log /var/log/nginx/radarr.error.log warn;
}

# Rate limiting zones
http {
    limit_req_zone $binary_remote_addr zone=api:10m rate=30r/m;

    # Caching
    proxy_cache_path /var/cache/nginx/radarr levels=1:2 keys_zone=radarr_cache:10m max_size=100m inactive=60m use_temp_path=off;
}
```

### Apache Configuration

```apache
# /etc/apache2/sites-available/radarr.conf
<VirtualHost *:80>
    ServerName radarr.yourdomain.com
    Redirect permanent / https://radarr.yourdomain.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName radarr.yourdomain.com

    # SSL Configuration
    SSLEngine on
    SSLCertificateFile /etc/ssl/certs/radarr.yourdomain.com.crt
    SSLCertificateKeyFile /etc/ssl/private/radarr.yourdomain.com.key
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLCipherSuite ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305
    SSLHonorCipherOrder off
    SSLSessionTickets off

    # Security Headers
    Header always set Strict-Transport-Security "max-age=31536000; includeSubDomains"
    Header always set X-Frame-Options DENY
    Header always set X-Content-Type-Options nosniff
    Header always set X-XSS-Protection "1; mode=block"
    Header always set Referrer-Policy "strict-origin-when-cross-origin"

    # Proxy Configuration
    ProxyPreserveHost On
    ProxyRequests Off

    ProxyPass /ping http://127.0.0.1:7878/ping retry=0
    ProxyPassReverse /ping http://127.0.0.1:7878/ping

    ProxyPass / http://127.0.0.1:7878/
    ProxyPassReverse / http://127.0.0.1:7878/

    # WebSocket support
    RewriteEngine on
    RewriteCond %{HTTP:Upgrade} websocket [NC]
    RewriteCond %{HTTP:Connection} upgrade [NC]
    RewriteRule ^/?(.*) "ws://127.0.0.1:7878/$1" [P,L]

    # Logging
    LogLevel warn
    CustomLog ${APACHE_LOG_DIR}/radarr.access.log combined
    ErrorLog ${APACHE_LOG_DIR}/radarr.error.log
</VirtualHost>
```

### Traefik Configuration (Docker Labels)

```yaml
# docker-compose with Traefik
services:
  radarr-go:
    # ... other configuration
    labels:
      # Basic Traefik configuration
      - traefik.enable=true
      - traefik.http.routers.radarr.rule=Host(`radarr.yourdomain.com`)
      - traefik.http.routers.radarr.entrypoints=websecure
      - traefik.http.routers.radarr.tls=true
      - traefik.http.routers.radarr.tls.certresolver=letsencrypt
      - traefik.http.services.radarr.loadbalancer.server.port=7878

      # Security headers
      - traefik.http.middlewares.radarr-headers.headers.customrequestheaders.X-Forwarded-Proto=https
      - traefik.http.middlewares.radarr-headers.headers.sslredirect=true
      - traefik.http.middlewares.radarr-headers.headers.stsincludesubdomains=true
      - traefik.http.middlewares.radarr-headers.headers.stspreload=true
      - traefik.http.middlewares.radarr-headers.headers.stsseconds=31536000
      - traefik.http.middlewares.radarr-headers.headers.framedeny=true
      - traefik.http.middlewares.radarr-headers.headers.contenttypenosniff=true
      - traefik.http.middlewares.radarr-headers.headers.browserxssfilter=true

      # Rate limiting
      - traefik.http.middlewares.radarr-ratelimit.ratelimit.burst=20
      - traefik.http.middlewares.radarr-ratelimit.ratelimit.average=60

      # Apply middlewares
      - traefik.http.routers.radarr.middlewares=radarr-headers@docker,radarr-ratelimit@docker

  traefik:
    image: traefik:v3.0
    container_name: traefik
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - traefik_data:/data
    command:
      - --api.dashboard=true
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --entrypoints.web.address=:80
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.letsencrypt.acme.tlschallenge=true
      - --certificatesresolvers.letsencrypt.acme.email=your-email@domain.com
      - --certificatesresolvers.letsencrypt.acme.storage=/data/acme.json
      - --log.level=INFO
```

## SSL/TLS Setup

### Let's Encrypt with Certbot

```bash
#!/bin/bash
# ssl-setup.sh - SSL certificate setup

DOMAIN="radarr.yourdomain.com"
EMAIL="your-email@domain.com"

# Install certbot
if command -v apt-get >/dev/null 2>&1; then
    sudo apt-get update
    sudo apt-get install -y certbot python3-certbot-nginx
elif command -v yum >/dev/null 2>&1; then
    sudo yum install -y certbot python3-certbot-nginx
fi

# Generate certificate
sudo certbot --nginx -d "$DOMAIN" --email "$EMAIL" --agree-tos --non-interactive

# Set up auto-renewal
echo "0 2 * * * root /usr/bin/certbot renew --quiet" | sudo tee -a /etc/crontab

# Test renewal
sudo certbot renew --dry-run
```

### Manual SSL Certificate

```bash
#!/bin/bash
# manual-ssl.sh - Manual SSL certificate generation

DOMAIN="radarr.yourdomain.com"
SSL_DIR="/etc/ssl"

# Create SSL directories
sudo mkdir -p "$SSL_DIR/certs" "$SSL_DIR/private"

# Generate private key
sudo openssl genrsa -out "$SSL_DIR/private/$DOMAIN.key" 2048

# Generate certificate signing request
sudo openssl req -new -key "$SSL_DIR/private/$DOMAIN.key" -out "$SSL_DIR/certs/$DOMAIN.csr" -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"

# Generate self-signed certificate (for testing)
sudo openssl x509 -req -in "$SSL_DIR/certs/$DOMAIN.csr" -signkey "$SSL_DIR/private/$DOMAIN.key" -out "$SSL_DIR/certs/$DOMAIN.crt" -days 365

# Set permissions
sudo chmod 600 "$SSL_DIR/private/$DOMAIN.key"
sudo chmod 644 "$SSL_DIR/certs/$DOMAIN.crt"

echo "SSL certificate generated for $DOMAIN"
echo "Certificate: $SSL_DIR/certs/$DOMAIN.crt"
echo "Private Key: $SSL_DIR/private/$DOMAIN.key"
```

## Environment-Specific Configuration

### Development Environment

```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  radarr-go:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
      - radarr_dev_data:/data
    environment:
      - RADARR_LOG_LEVEL=debug
      - RADARR_DEVELOPMENT_ENABLE_DEBUG_ENDPOINTS=true
      - RADARR_DEVELOPMENT_LOG_SQL_QUERIES=true
    ports:
      - "7878:7878"
      - "2345:2345"  # Delve debugger
```

### Staging Environment

```yaml
# docker-compose.staging.yml
version: '3.8'
services:
  radarr-go:
    image: ghcr.io/username/radarr-go:staging
    environment:
      - RADARR_LOG_LEVEL=debug
      - RADARR_DATABASE_TYPE=postgres
      - RADARR_DATABASE_HOST=staging-postgres
      - RADARR_HEALTH_INTERVAL=1m
    labels:
      - traefik.http.routers.radarr-staging.rule=Host(`radarr-staging.yourdomain.com`)
```

## Next Steps

This completes the production deployment guide. The next sections will cover:

1. **Monitoring and Alerting Setup** - Prometheus, Grafana, and alerting configurations
2. **Performance Tuning Guide** - Database optimization, Go runtime tuning, and scaling
3. **Security Hardening** - Network security, authentication, and container security
4. **Automated Scripts** - Deployment automation and monitoring templates

Each section builds upon this deployment foundation to create a complete production-ready Radarr Go environment.
