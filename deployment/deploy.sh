#!/bin/bash
# deployment/deploy.sh - Production deployment automation script
# Comprehensive deployment script for Radarr Go with monitoring, backups, and rollback capabilities

set -euo pipefail

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEPLOYMENT_DIR="$SCRIPT_DIR"
CONFIG_DIR="$DEPLOYMENT_DIR/config"
COMPOSE_FILE="$DEPLOYMENT_DIR/docker-compose.prod.yml"
MONITORING_COMPOSE_FILE="$DEPLOYMENT_DIR/docker-compose.monitoring.yml"
ENV_FILE="$DEPLOYMENT_DIR/.env"
BACKUP_DIR="/opt/radarr/backups"
LOG_FILE="$DEPLOYMENT_DIR/deploy.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp=$(date +'%Y-%m-%d %H:%M:%S')
    echo -e "${GREEN}[$timestamp] [$level]${NC} $message" | tee -a "$LOG_FILE"
}

info() { log "INFO" "$@"; }
warn() { log "WARN" "${YELLOW}$*${NC}"; }
error() { log "ERROR" "${RED}$*${NC}"; exit 1; }
success() { log "SUCCESS" "${GREEN}$*${NC}"; }

# Version detection
detect_version() {
    if [ -n "${RADARR_VERSION:-}" ]; then
        echo "$RADARR_VERSION"
        return
    fi

    # Try to get version from git
    if git -C "$PROJECT_ROOT" describe --tags --exact-match HEAD 2>/dev/null; then
        return
    fi

    # Fallback to latest commit
    if git -C "$PROJECT_ROOT" rev-parse --short HEAD 2>/dev/null; then
        return
    fi

    echo "latest"
}

VERSION=$(detect_version)

# Environment validation
validate_environment() {
    info "Validating deployment environment..."

    # Check required commands
    local required_commands=("docker" "docker-compose" "curl" "jq")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            error "Required command not found: $cmd"
        fi
    done

    # Check Docker daemon
    if ! docker info >/dev/null 2>&1; then
        error "Docker daemon is not running or accessible"
    fi

    # Check Docker Compose version
    local compose_version=$(docker-compose --version | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' | head -1)
    if [ -z "$compose_version" ]; then
        # Try docker compose (newer syntax)
        compose_version=$(docker compose version --short 2>/dev/null || echo "")
    fi

    if [ -n "$compose_version" ]; then
        info "Docker Compose version: $compose_version"
    else
        warn "Could not detect Docker Compose version"
    fi

    # Check available disk space
    local free_space=$(df "$DEPLOYMENT_DIR" | awk 'NR==2 {print $4}')
    if [ "$free_space" -lt 5000000 ]; then # 5GB
        warn "Low disk space: $(($free_space / 1024 / 1024))GB available"
    fi

    # Check memory
    local available_memory=$(free -m | awk 'NR==2{print $7}')
    if [ "$available_memory" -lt 2048 ]; then # 2GB
        warn "Low available memory: ${available_memory}MB"
    fi

    # Validate environment file
    if [ ! -f "$ENV_FILE" ]; then
        warn "Environment file not found, creating template"
        create_env_template
    else
        source "$ENV_FILE"
        validate_env_variables
    fi

    success "Environment validation completed"
}

# Create environment template
create_env_template() {
    info "Creating environment file template..."

    cat > "$ENV_FILE" << 'EOF'
# Radarr Go Production Environment Configuration

# Application Configuration
RADARR_VERSION=latest
RADARR_API_KEY=your-secure-api-key-64-chars-recommended-please-change-this
RADARR_URL_BASE=

# Database Configuration
POSTGRES_VERSION=15-alpine
POSTGRES_PASSWORD=secure-postgres-password-please-change-this
POSTGRES_DB=radarr
POSTGRES_USER=radarr

# Domain and SSL
DOMAIN=radarr.yourdomain.com
ACME_EMAIL=your-email@domain.com

# TMDB API Key (required for movie metadata)
TMDB_API_KEY=your-tmdb-api-key

# External Services
REDIS_PASSWORD=secure-redis-password-please-change-this

# Monitoring (optional)
ENABLE_MONITORING=true
PROMETHEUS_RETENTION=30d
GRAFANA_ADMIN_PASSWORD=secure-grafana-password-please-change-this

# Backup Configuration
BACKUP_RETENTION_DAYS=30
BACKUP_ENCRYPTION_PASSWORD=secure-backup-password-please-change-this

# Network Configuration
DOCKER_SUBNET=172.20.0.0/16
APP_SUBNET=172.20.1.0/24
DB_SUBNET=172.20.2.0/24
MONITORING_SUBNET=172.20.3.0/24

# Storage Paths
DATA_ROOT=/opt/radarr
MOVIES_PATH=/media/movies
DOWNLOADS_PATH=/media/downloads
EOF

    warn "Please edit $ENV_FILE with your configuration before deploying"
    exit 1
}

# Validate environment variables
validate_env_variables() {
    local required_vars=(
        "RADARR_API_KEY"
        "POSTGRES_PASSWORD"
        "TMDB_API_KEY"
        "DOMAIN"
        "ACME_EMAIL"
    )

    local missing_vars=()

    for var in "${required_vars[@]}"; do
        if [ -z "${!var:-}" ]; then
            missing_vars+=("$var")
        fi
    done

    if [ ${#missing_vars[@]} -gt 0 ]; then
        error "Missing required environment variables: ${missing_vars[*]}"
    fi

    # Check for default/insecure values
    if [[ "${RADARR_API_KEY:-}" == *"please-change-this"* ]]; then
        error "Please change the default RADARR_API_KEY in $ENV_FILE"
    fi

    if [[ "${POSTGRES_PASSWORD:-}" == *"please-change-this"* ]]; then
        error "Please change the default POSTGRES_PASSWORD in $ENV_FILE"
    fi
}

# Pre-deployment checks
pre_deployment_checks() {
    info "Running pre-deployment checks..."

    # Check if services are already running
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        warn "Some services are already running. Consider stopping them first."

        read -p "Continue with deployment? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            info "Deployment cancelled by user"
            exit 0
        fi
    fi

    # Check for required directories
    local required_dirs=(
        "${DATA_ROOT:-/opt/radarr}"
        "${MOVIES_PATH:-/media/movies}"
        "${DOWNLOADS_PATH:-/media/downloads}"
        "$BACKUP_DIR"
    )

    for dir in "${required_dirs[@]}"; do
        if [ ! -d "$dir" ]; then
            info "Creating directory: $dir"
            sudo mkdir -p "$dir"
            sudo chown "${USER}:${USER}" "$dir" 2>/dev/null || true
        fi
    done

    # Test database connectivity (if upgrading)
    if docker ps | grep -q "radarr-postgres"; then
        info "Testing database connectivity..."
        if ! docker exec radarr-postgres pg_isready -U "${POSTGRES_USER:-radarr}" >/dev/null 2>&1; then
            warn "Database connectivity test failed"
        fi
    fi

    success "Pre-deployment checks completed"
}

# Backup existing deployment
backup_current_deployment() {
    info "Creating backup of current deployment..."

    local backup_timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_path="$BACKUP_DIR/deployment_backup_$backup_timestamp"

    mkdir -p "$backup_path"

    # Backup configuration
    if [ -d "${DATA_ROOT:-/opt/radarr}/config" ]; then
        cp -r "${DATA_ROOT:-/opt/radarr}/config" "$backup_path/"
        info "Configuration backed up"
    fi

    # Backup database
    if docker ps | grep -q "radarr-postgres"; then
        info "Backing up database..."
        docker exec radarr-postgres pg_dump \
            -U "${POSTGRES_USER:-radarr}" \
            -h localhost \
            "${POSTGRES_DB:-radarr}" | \
            gzip > "$backup_path/database_backup.sql.gz"
        info "Database backed up"
    fi

    # Backup Docker volumes
    if docker volume ls | grep -q "radarr"; then
        info "Backing up Docker volumes..."
        docker run --rm \
            -v radarr_data:/source:ro \
            -v "$backup_path:/backup" \
            alpine \
            tar czf /backup/volumes_backup.tar.gz -C /source .
        info "Volumes backed up"
    fi

    # Store deployment info
    cat > "$backup_path/deployment_info.txt" << EOF
Backup created: $(date)
Version: $VERSION
Git commit: $(git -C "$PROJECT_ROOT" rev-parse HEAD 2>/dev/null || echo "unknown")
Environment: $(cat "$ENV_FILE" | grep -v "PASSWORD\|KEY\|SECRET" || true)
Docker info: $(docker version --format json | jq -r '.Server.Version' 2>/dev/null || echo "unknown")
EOF

    success "Backup created: $backup_path"

    # Cleanup old backups
    cleanup_old_backups
}

# Cleanup old backups
cleanup_old_backups() {
    local retention_days="${BACKUP_RETENTION_DAYS:-30}"

    info "Cleaning up backups older than $retention_days days..."

    find "$BACKUP_DIR" -name "deployment_backup_*" -type d -mtime "+$retention_days" -exec rm -rf {} \; 2>/dev/null || true

    local remaining_backups=$(find "$BACKUP_DIR" -name "deployment_backup_*" -type d | wc -l)
    info "Remaining backups: $remaining_backups"
}

# Deploy application
deploy_application() {
    info "Deploying Radarr Go version $VERSION..."

    # Set version in environment
    export RADARR_VERSION="$VERSION"

    # Pull latest images
    info "Pulling Docker images..."
    docker-compose -f "$COMPOSE_FILE" pull

    # Create networks
    info "Creating Docker networks..."
    docker network create radarr-app-network --subnet="${APP_SUBNET:-172.20.1.0/24}" 2>/dev/null || true
    docker network create radarr-db-network --subnet="${DB_SUBNET:-172.20.2.0/24}" --internal 2>/dev/null || true

    # Start services
    info "Starting services..."
    docker-compose -f "$COMPOSE_FILE" up -d

    # Wait for services to be healthy
    wait_for_services

    success "Application deployment completed"
}

# Wait for services to be healthy
wait_for_services() {
    info "Waiting for services to become healthy..."

    local max_attempts=60
    local attempt=0
    local services=("radarr-go" "postgres")

    if [ "${ENABLE_MONITORING:-false}" = "true" ]; then
        services+=("prometheus" "grafana")
    fi

    while [ $attempt -lt $max_attempts ]; do
        local all_healthy=true

        for service in "${services[@]}"; do
            local health_status=$(docker inspect "$service" --format='{{.State.Health.Status}}' 2>/dev/null || echo "none")

            if [ "$health_status" != "healthy" ]; then
                all_healthy=false
                break
            fi
        done

        if $all_healthy; then
            success "All services are healthy"
            return 0
        fi

        echo -n "."
        sleep 5
        ((attempt++))
    done

    error "Services did not become healthy within timeout"
}

# Deploy monitoring stack
deploy_monitoring() {
    if [ "${ENABLE_MONITORING:-false}" != "true" ]; then
        info "Monitoring disabled, skipping deployment"
        return 0
    fi

    info "Deploying monitoring stack..."

    # Create monitoring network
    docker network create radarr-monitoring-network --subnet="${MONITORING_SUBNET:-172.20.3.0/24}" 2>/dev/null || true

    # Deploy monitoring services
    docker-compose -f "$MONITORING_COMPOSE_FILE" up -d

    info "Monitoring stack deployed"

    # Display access URLs
    cat << EOF

Monitoring Access URLs:
======================
Prometheus: http://localhost:9090
Grafana: http://localhost:3000 (admin / ${GRAFANA_ADMIN_PASSWORD:-admin})
AlertManager: http://localhost:9093

EOF
}

# Post-deployment verification
post_deployment_verification() {
    info "Running post-deployment verification..."

    # Test API endpoint
    local api_url="http://localhost:7878"
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -f -s "$api_url/ping" >/dev/null 2>&1; then
            success "API endpoint is responding"
            break
        fi

        ((attempt++))
        if [ $attempt -eq $max_attempts ]; then
            error "API endpoint is not responding after $max_attempts attempts"
        fi

        sleep 2
    done

    # Test API authentication
    if curl -f -s -H "X-Api-Key: ${RADARR_API_KEY}" "$api_url/api/v3/system/status" >/dev/null 2>&1; then
        success "API authentication is working"
    else
        error "API authentication failed"
    fi

    # Check database connectivity
    if docker exec radarr-postgres pg_isready -U "${POSTGRES_USER:-radarr}" >/dev/null 2>&1; then
        success "Database is accessible"
    else
        error "Database is not accessible"
    fi

    # Check service logs for errors
    info "Checking service logs for errors..."
    local error_count=$(docker-compose -f "$COMPOSE_FILE" logs --tail=100 radarr-go | grep -i error | wc -l)

    if [ "$error_count" -gt 0 ]; then
        warn "Found $error_count error messages in logs (this may be normal during startup)"
        docker-compose -f "$COMPOSE_FILE" logs --tail=20 radarr-go | grep -i error | tail -5
    else
        success "No errors found in recent logs"
    fi

    success "Post-deployment verification completed"
}

# Display deployment status
show_deployment_status() {
    info "Current deployment status:"

    echo
    docker-compose -f "$COMPOSE_FILE" ps
    echo

    # Show resource usage
    echo "Resource Usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}"
    echo

    # Show service URLs
    cat << EOF
Service Access:
==============
Radarr Go: http://localhost:7878
          https://${DOMAIN}
API Key: ${RADARR_API_KEY:0:8}...

Database: localhost:5432 (internal only)
Redis: localhost:6379 (internal only)

EOF

    if [ "${ENABLE_MONITORING:-false}" = "true" ]; then
        cat << EOF
Monitoring:
==========
Prometheus: http://localhost:9090
Grafana: http://localhost:3000
AlertManager: http://localhost:9093

EOF
    fi

    # Show recent logs
    echo "Recent Logs (last 10 lines):"
    docker-compose -f "$COMPOSE_FILE" logs --tail=10 radarr-go | sed 's/^/  /'
}

# Rollback deployment
rollback_deployment() {
    info "Rolling back deployment..."

    # Find latest backup
    local latest_backup=$(find "$BACKUP_DIR" -name "deployment_backup_*" -type d | sort -r | head -1)

    if [ -z "$latest_backup" ]; then
        error "No backup found for rollback"
    fi

    info "Rolling back to: $latest_backup"

    # Stop current services
    docker-compose -f "$COMPOSE_FILE" down

    # Restore configuration
    if [ -d "$latest_backup/config" ]; then
        rm -rf "${DATA_ROOT:-/opt/radarr}/config"
        cp -r "$latest_backup/config" "${DATA_ROOT:-/opt/radarr}/"
    fi

    # Restore database
    if [ -f "$latest_backup/database_backup.sql.gz" ]; then
        info "Restoring database..."

        # Start database
        docker-compose -f "$COMPOSE_FILE" up -d postgres
        sleep 30

        # Restore data
        zcat "$latest_backup/database_backup.sql.gz" | \
            docker exec -i radarr-postgres psql -U "${POSTGRES_USER:-radarr}" "${POSTGRES_DB:-radarr}"
    fi

    # Restore volumes
    if [ -f "$latest_backup/volumes_backup.tar.gz" ]; then
        info "Restoring volumes..."
        docker run --rm \
            -v radarr_data:/target \
            -v "$latest_backup:/backup:ro" \
            alpine \
            tar xzf /backup/volumes_backup.tar.gz -C /target
    fi

    # Start all services
    docker-compose -f "$COMPOSE_FILE" up -d

    success "Rollback completed"
    wait_for_services
}

# Stop deployment
stop_deployment() {
    info "Stopping Radarr Go deployment..."

    docker-compose -f "$COMPOSE_FILE" down

    if [ "${ENABLE_MONITORING:-false}" = "true" ]; then
        docker-compose -f "$MONITORING_COMPOSE_FILE" down
    fi

    success "Deployment stopped"
}

# Show logs
show_logs() {
    local service="${1:-radarr-go}"
    local lines="${2:-100}"

    info "Showing logs for $service (last $lines lines):"
    docker-compose -f "$COMPOSE_FILE" logs --tail="$lines" -f "$service"
}

# Update deployment
update_deployment() {
    local new_version="${1:-latest}"

    info "Updating deployment to version $new_version..."

    # Backup current deployment
    backup_current_deployment

    # Update version
    export RADARR_VERSION="$new_version"

    # Pull new images
    docker-compose -f "$COMPOSE_FILE" pull

    # Recreate containers with new image
    docker-compose -f "$COMPOSE_FILE" up -d --force-recreate radarr-go

    # Wait for services
    wait_for_services

    # Verify deployment
    post_deployment_verification

    success "Update completed to version $new_version"
}

# Health check
health_check() {
    info "Running health check..."

    local exit_code=0

    # Check service status
    local services=$(docker-compose -f "$COMPOSE_FILE" ps --services)
    for service in $services; do
        local status=$(docker-compose -f "$COMPOSE_FILE" ps "$service" | tail -n1 | awk '{print $4}')

        if [[ "$status" == "Up"* ]]; then
            echo "✓ $service: $status"
        else
            echo "✗ $service: $status"
            exit_code=1
        fi
    done

    # Check API health
    if curl -f -s "http://localhost:7878/ping" >/dev/null 2>&1; then
        echo "✓ API: Healthy"
    else
        echo "✗ API: Unhealthy"
        exit_code=1
    fi

    # Check database
    if docker exec radarr-postgres pg_isready -U "${POSTGRES_USER:-radarr}" >/dev/null 2>&1; then
        echo "✓ Database: Healthy"
    else
        echo "✗ Database: Unhealthy"
        exit_code=1
    fi

    # Check disk space
    local disk_usage=$(df "${DATA_ROOT:-/opt/radarr}" | awk 'NR==2 {print int($5)}')
    if [ "$disk_usage" -lt 90 ]; then
        echo "✓ Disk Space: ${disk_usage}% used"
    else
        echo "⚠ Disk Space: ${disk_usage}% used (high)"
        exit_code=1
    fi

    if [ $exit_code -eq 0 ]; then
        success "Health check passed"
    else
        error "Health check failed"
    fi

    return $exit_code
}

# Generate deployment report
generate_deployment_report() {
    local report_file="$DEPLOYMENT_DIR/deployment_report_$(date +%Y%m%d_%H%M%S).txt"

    {
        echo "Radarr Go Deployment Report"
        echo "==========================="
        echo "Generated: $(date)"
        echo "Version: $VERSION"
        echo "Deployment Directory: $DEPLOYMENT_DIR"
        echo ""

        echo "System Information:"
        echo "==================="
        echo "OS: $(uname -a)"
        echo "Docker Version: $(docker --version)"
        echo "Docker Compose Version: $(docker-compose --version 2>/dev/null || docker compose version)"
        echo "Available Memory: $(free -h | awk 'NR==2{print $7}')"
        echo "Available Disk: $(df -h "$DEPLOYMENT_DIR" | awk 'NR==2{print $4}')"
        echo ""

        echo "Service Status:"
        echo "==============="
        docker-compose -f "$COMPOSE_FILE" ps
        echo ""

        echo "Container Resource Usage:"
        echo "========================"
        docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"
        echo ""

        echo "Network Configuration:"
        echo "====================="
        docker network ls | grep radarr
        echo ""

        echo "Volume Information:"
        echo "=================="
        docker volume ls | grep radarr
        echo ""

        echo "Configuration Summary:"
        echo "====================="
        cat "$ENV_FILE" | grep -v "PASSWORD\|KEY\|SECRET" | sort
        echo ""

        echo "Recent Application Logs:"
        echo "======================="
        docker-compose -f "$COMPOSE_FILE" logs --tail=50 radarr-go

    } > "$report_file"

    info "Deployment report generated: $report_file"
}

# Main function
main() {
    # Initialize
    mkdir -p "$BACKUP_DIR"
    touch "$LOG_FILE"

    info "Starting Radarr Go deployment script"
    info "Version: $VERSION"
    info "Deployment directory: $DEPLOYMENT_DIR"

    case "${1:-deploy}" in
        "deploy")
            validate_environment
            pre_deployment_checks
            backup_current_deployment
            deploy_application
            deploy_monitoring
            post_deployment_verification
            show_deployment_status
            generate_deployment_report
            ;;
        "update")
            validate_environment
            update_deployment "${2:-latest}"
            ;;
        "rollback")
            rollback_deployment
            ;;
        "stop")
            stop_deployment
            ;;
        "start")
            docker-compose -f "$COMPOSE_FILE" up -d
            if [ "${ENABLE_MONITORING:-false}" = "true" ]; then
                docker-compose -f "$MONITORING_COMPOSE_FILE" up -d
            fi
            wait_for_services
            ;;
        "restart")
            stop_deployment
            sleep 5
            docker-compose -f "$COMPOSE_FILE" up -d
            wait_for_services
            ;;
        "status")
            show_deployment_status
            ;;
        "logs")
            show_logs "${2:-radarr-go}" "${3:-100}"
            ;;
        "health")
            health_check
            ;;
        "backup")
            backup_current_deployment
            ;;
        "report")
            generate_deployment_report
            ;;
        *)
            cat << EOF
Usage: $0 {deploy|update|rollback|stop|start|restart|status|logs|health|backup|report}

Commands:
  deploy     - Deploy Radarr Go with full setup and verification
  update     - Update to a new version (optional: specify version)
  rollback   - Rollback to the previous backup
  stop       - Stop all services
  start      - Start all services
  restart    - Restart all services
  status     - Show current deployment status
  logs       - Show service logs (optional: service name and line count)
  health     - Run health check
  backup     - Create a backup of current deployment
  report     - Generate a comprehensive deployment report

Examples:
  $0 deploy                    # Deploy with current configuration
  $0 update v1.0.0            # Update to specific version
  $0 logs radarr-go 200       # Show last 200 lines of radarr-go logs
  $0 health                   # Check deployment health

Configuration:
  Edit $ENV_FILE to customize your deployment
EOF
            exit 1
            ;;
    esac

    success "Operation completed successfully"
}

# Error handling
trap 'error "Script failed at line $LINENO"' ERR

# Run main function
main "$@"
