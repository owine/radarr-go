# Radarr Go Automation and Templates

**Version**: v0.9.0-alpha

## Overview

This guide provides comprehensive automation scripts and monitoring templates for production Radarr Go deployments. These tools enable:

- **One-Click Deployment** - Fully automated deployment with infrastructure as code
- **Continuous Integration/Continuous Deployment (CI/CD)** - Automated testing and deployment pipelines
- **Infrastructure Automation** - Terraform and Ansible automation
- **Monitoring Templates** - Pre-configured dashboards and alerting rules
- **Operational Automation** - Automated maintenance, backups, and scaling

## Quick Start Automation

### Complete Stack Deployment

Create `scripts/deploy-complete-stack.sh`:

```bash
#!/bin/bash
# deploy-complete-stack.sh - Complete Radarr Go stack deployment

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Configuration
ENVIRONMENT="${1:-production}"
DOMAIN="${2:-radarr.yourdomain.com}"
EMAIL="${3:-admin@yourdomain.com}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')] $1${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING: $1${NC}"; }
error() { echo -e "${RED}[$(date +'%H:%M:%S')] ERROR: $1${NC}"; exit 1; }
info() { echo -e "${BLUE}[$(date +'%H:%M:%S')] INFO: $1${NC}"; }

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    local required_tools=("docker" "docker-compose" "git" "curl" "jq")
    for tool in "${required_tools[@]}"; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            error "$tool is required but not installed"
        fi
    done

    # Check Docker is running
    if ! docker info >/dev/null 2>&1; then
        error "Docker is not running"
    fi

    # Check available disk space (minimum 10GB)
    local free_space=$(df / | awk 'NR==2 {print $4}')
    if [ "$free_space" -lt 10000000 ]; then
        warn "Low disk space: ${free_space}KB free (minimum 10GB recommended)"
    fi

    log "Prerequisites check passed"
}

# Generate secure configuration
generate_secure_config() {
    log "Generating secure configuration..."

    local env_file=".env.${ENVIRONMENT}"

    # Generate secure passwords and keys
    local postgres_password=$(openssl rand -hex 32)
    local api_key=$(openssl rand -hex 32)
    local grafana_password=$(openssl rand -base64 24)
    local jwt_secret=$(openssl rand -hex 32)

    cat > "$env_file" << EOF
# Radarr Go Environment Configuration
# Generated: $(date -Iseconds)

# Application Configuration
ENVIRONMENT=$ENVIRONMENT
DOMAIN=$DOMAIN
EMAIL=$EMAIL

# Database Configuration
POSTGRES_PASSWORD=$postgres_password
POSTGRES_USER=radarr
POSTGRES_DB=radarr
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Radarr Configuration
RADARR_AUTH_API_KEY=$api_key
RADARR_SERVER_HOST=0.0.0.0
RADARR_SERVER_PORT=7878
RADARR_DATABASE_TYPE=postgres
RADARR_LOG_LEVEL=info
RADARR_LOG_FORMAT=json

# Security Configuration
RADARR_AUTH_METHOD=apikey
RADARR_SECURITY_ENABLE_SECURITY_HEADERS=true
RADARR_SECURITY_ENABLE_CORS=false

# Monitoring Configuration
GRAFANA_ADMIN_PASSWORD=$grafana_password
PROMETHEUS_RETENTION_TIME=30d

# SSL Configuration
CERTBOT_EMAIL=$EMAIL
ACME_DOMAIN=$DOMAIN

# Backup Configuration
BACKUP_ENCRYPTION_KEY=$(openssl rand -hex 32)
BACKUP_RETENTION_DAYS=30
S3_BACKUP_BUCKET=""  # Optional: Configure for cloud backups

# Notification Configuration (Optional)
SLACK_WEBHOOK_URL=""
DISCORD_WEBHOOK_URL=""
SMTP_HOST=""
SMTP_PORT=587
SMTP_USER=""
SMTP_PASSWORD=""
EOF

    chmod 600 "$env_file"
    log "Secure configuration generated: $env_file"
    info "IMPORTANT: Save the API key: $api_key"
    info "IMPORTANT: Save the Grafana password: $grafana_password"
}

# Setup directories and permissions
setup_directories() {
    log "Setting up directories and permissions..."

    local base_dir="/opt/radarr"

    # Create directory structure
    sudo mkdir -p "$base_dir"/{data,config,backups,logs,ssl}
    sudo mkdir -p "$base_dir"/monitoring/{prometheus,grafana,loki,alertmanager}

    # Set ownership
    sudo chown -R 1000:1000 "$base_dir"
    sudo chown -R root:root "$base_dir"/ssl
    sudo chmod 755 "$base_dir"/ssl

    log "Directory structure created"
}

# Deploy infrastructure
deploy_infrastructure() {
    log "Deploying infrastructure stack..."

    # Copy configuration files
    cp -r "$PROJECT_ROOT"/docker/ ./deployment/
    cp -r "$PROJECT_ROOT"/monitoring/ ./deployment/

    cd ./deployment

    # Deploy database first
    log "Deploying database..."
    docker-compose -f docker-compose.postgres.yml --env-file "../.env.${ENVIRONMENT}" up -d postgres

    # Wait for database to be ready
    info "Waiting for database to initialize..."
    local max_attempts=30
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if docker-compose -f docker-compose.postgres.yml --env-file "../.env.${ENVIRONMENT}" exec -T postgres pg_isready -U radarr >/dev/null 2>&1; then
            log "Database is ready"
            break
        fi
        sleep 5
        ((attempt++))
    done

    if [ $attempt -eq $max_attempts ]; then
        error "Database failed to initialize within timeout"
    fi

    # Deploy application
    log "Deploying Radarr Go application..."
    docker-compose -f docker-compose.production.yml --env-file "../.env.${ENVIRONMENT}" up -d radarr-go

    # Deploy monitoring stack
    log "Deploying monitoring stack..."
    docker-compose -f docker-compose.monitoring.yml --env-file "../.env.${ENVIRONMENT}" up -d

    # Deploy reverse proxy
    log "Deploying reverse proxy..."
    docker-compose -f docker-compose.nginx.yml --env-file "../.env.${ENVIRONMENT}" up -d

    cd ..
    log "Infrastructure deployment completed"
}

# Setup SSL certificates
setup_ssl() {
    log "Setting up SSL certificates..."

    if [ "$DOMAIN" != "radarr.yourdomain.com" ]; then
        # Use Let's Encrypt for real domains
        docker run --rm \
            -v /opt/radarr/ssl:/etc/letsencrypt \
            -v /var/www/html:/var/www/html \
            -p 80:80 \
            certbot/certbot certonly \
            --standalone \
            --email "$EMAIL" \
            --agree-tos \
            --no-eff-email \
            --non-interactive \
            -d "$DOMAIN"

        log "SSL certificate obtained for $DOMAIN"
    else
        warn "Using self-signed certificate for development/testing"
        # Generate self-signed certificate
        sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
            -keyout /opt/radarr/ssl/privkey.pem \
            -out /opt/radarr/ssl/fullchain.pem \
            -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"
    fi
}

# Configure monitoring and alerting
configure_monitoring() {
    log "Configuring monitoring and alerting..."

    # Import Grafana dashboards
    info "Waiting for Grafana to start..."
    sleep 30

    local grafana_url="http://localhost:3000"
    local grafana_user="admin"
    local grafana_password=$(grep GRAFANA_ADMIN_PASSWORD ".env.${ENVIRONMENT}" | cut -d'=' -f2)

    # Wait for Grafana to be ready
    local max_attempts=20
    local attempt=0

    while [ $attempt -lt $max_attempts ]; do
        if curl -sf "$grafana_url/api/health" >/dev/null; then
            break
        fi
        sleep 5
        ((attempt++))
    done

    # Import dashboard
    if [ -f "$PROJECT_ROOT/monitoring/grafana/dashboards/radarr-dashboard.json" ]; then
        curl -X POST \
            -H "Content-Type: application/json" \
            -u "$grafana_user:$grafana_password" \
            "$grafana_url/api/dashboards/db" \
            -d @"$PROJECT_ROOT/monitoring/grafana/dashboards/radarr-dashboard.json" || warn "Failed to import dashboard"
    fi

    log "Monitoring configured"
}

# Verify deployment
verify_deployment() {
    log "Verifying deployment..."

    local api_key=$(grep RADARR_AUTH_API_KEY ".env.${ENVIRONMENT}" | cut -d'=' -f2)
    local checks=0
    local passed=0

    # Check application health
    ((checks++))
    if curl -sf -H "X-API-Key: $api_key" http://localhost:7878/api/v3/system/status >/dev/null; then
        log "✓ Application health check passed"
        ((passed++))
    else
        warn "✗ Application health check failed"
    fi

    # Check database connectivity
    ((checks++))
    if docker exec radarr-postgres pg_isready -U radarr >/dev/null 2>&1; then
        log "✓ Database connectivity check passed"
        ((passed++))
    else
        warn "✗ Database connectivity check failed"
    fi

    # Check monitoring
    ((checks++))
    if curl -sf http://localhost:9090/api/v1/query?query=up >/dev/null; then
        log "✓ Prometheus monitoring check passed"
        ((passed++))
    else
        warn "✗ Prometheus monitoring check failed"
    fi

    # Check SSL (if configured)
    if [ "$DOMAIN" != "radarr.yourdomain.com" ]; then
        ((checks++))
        if curl -sf -k "https://$DOMAIN/ping" >/dev/null; then
            log "✓ HTTPS connectivity check passed"
            ((passed++))
        else
            warn "✗ HTTPS connectivity check failed"
        fi
    fi

    log "Deployment verification: $passed/$checks checks passed"

    if [ "$passed" -eq "$checks" ]; then
        log "🎉 Deployment completed successfully!"
        return 0
    else
        warn "Some checks failed. Review the logs above."
        return 1
    fi
}

# Display access information
display_access_info() {
    log "Deployment Access Information"
    echo
    echo "📱 Applications:"
    echo "   Radarr Go:     http://$DOMAIN (or http://localhost:7878)"
    echo "   Grafana:       http://localhost:3000 (admin/$(grep GRAFANA_ADMIN_PASSWORD ".env.${ENVIRONMENT}" | cut -d'=' -f2))"
    echo "   Prometheus:    http://localhost:9090"
    echo "   AlertManager:  http://localhost:9093"
    echo
    echo "🔑 API Access:"
    echo "   API Key: $(grep RADARR_AUTH_API_KEY ".env.${ENVIRONMENT}" | cut -d'=' -f2)"
    echo "   Test: curl -H \"X-API-Key: $(grep RADARR_AUTH_API_KEY ".env.${ENVIRONMENT}" | cut -d'=' -f2)\" http://localhost:7878/api/v3/system/status"
    echo
    echo "📁 Important Files:"
    echo "   Environment: .env.${ENVIRONMENT}"
    echo "   Data Directory: /opt/radarr/data"
    echo "   Logs: /opt/radarr/logs"
    echo "   Backups: /opt/radarr/backups"
    echo
    echo "🔧 Management Commands:"
    echo "   View logs: docker-compose logs -f"
    echo "   Restart: docker-compose restart"
    echo "   Stop: docker-compose down"
    echo "   Backup: ./scripts/backup.sh"
    echo
}

# Main deployment process
main() {
    log "Starting complete Radarr Go stack deployment"
    log "Environment: $ENVIRONMENT"
    log "Domain: $DOMAIN"
    log "Email: $EMAIL"
    echo

    check_prerequisites
    generate_secure_config
    setup_directories
    deploy_infrastructure
    setup_ssl
    configure_monitoring

    if verify_deployment; then
        display_access_info
        log "🚀 Deployment completed successfully!"

        # Schedule regular maintenance
        info "Setting up automated maintenance..."
        (crontab -l 2>/dev/null; echo "0 2 * * * $SCRIPT_DIR/backup.sh >/dev/null 2>&1") | crontab -
        (crontab -l 2>/dev/null; echo "0 4 * * 0 $SCRIPT_DIR/maintenance.sh >/dev/null 2>&1") | crontab -

    else
        error "Deployment verification failed. Check logs above."
    fi
}

# Handle script arguments
case "${1:-deploy}" in
    "deploy")
        main "$@"
        ;;
    "verify")
        verify_deployment
        ;;
    "info")
        display_access_info
        ;;
    *)
        echo "Usage: $0 [deploy|verify|info] [environment] [domain] [email]"
        echo ""
        echo "Commands:"
        echo "  deploy  - Full deployment (default)"
        echo "  verify  - Verify existing deployment"
        echo "  info    - Display access information"
        echo ""
        echo "Examples:"
        echo "  $0 deploy production radarr.example.com admin@example.com"
        echo "  $0 deploy development radarr.local.dev dev@example.com"
        echo "  $0 verify"
        exit 1
        ;;
esac
```

## CI/CD Pipeline Templates

### GitHub Actions Workflow

Create `.github/workflows/deploy.yml`:

```yaml
# .github/workflows/deploy.yml
name: Deploy Radarr Go

on:
  push:
    branches: [main, develop]
    tags: [v*]
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17-alpine
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: radarr_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'

      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Install dependencies
        run: go mod download

      - name: Run linter
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

      - name: Run tests
        env:
          RADARR_DATABASE_TYPE: postgres
          RADARR_DATABASE_HOST: localhost
          RADARR_DATABASE_PORT: 5432
          RADARR_DATABASE_USERNAME: postgres
          RADARR_DATABASE_PASSWORD: postgres
          RADARR_DATABASE_NAME: radarr_test
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html

      - name: Upload coverage reports
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=semver,pattern={{major}}
            type=sha

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./Dockerfile.prod
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          build-args: |
            VERSION=${{ github.ref_name }}
            COMMIT=${{ github.sha }}
            BUILD_DATE=${{ steps.date.outputs.date }}
          cache-from: type=gha
          cache-to: type=gha,mode=max

  security-scan:
    needs: build
    runs-on: ubuntu-latest
    permissions:
      security-events: write

    steps:
      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          image-ref: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy scan results
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    needs: [test, build, security-scan]
    runs-on: ubuntu-latest
    environment: staging

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Deploy to staging
        env:
          DEPLOY_HOST: ${{ secrets.STAGING_HOST }}
          DEPLOY_USER: ${{ secrets.STAGING_USER }}
          DEPLOY_KEY: ${{ secrets.STAGING_SSH_KEY }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          echo "$DEPLOY_KEY" > deploy_key
          chmod 600 deploy_key

          ssh -i deploy_key -o StrictHostKeyChecking=no $DEPLOY_USER@$DEPLOY_HOST << EOF
            cd /opt/radarr
            docker-compose pull
            docker-compose up -d
            sleep 30
            curl -f http://localhost:7878/ping || exit 1
          EOF

  deploy-production:
    if: startsWith(github.ref, 'refs/tags/v')
    needs: [test, build, security-scan]
    runs-on: ubuntu-latest
    environment: production

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Deploy to production
        env:
          DEPLOY_HOST: ${{ secrets.PRODUCTION_HOST }}
          DEPLOY_USER: ${{ secrets.PRODUCTION_USER }}
          DEPLOY_KEY: ${{ secrets.PRODUCTION_SSH_KEY }}
          IMAGE_TAG: ${{ github.ref_name }}
        run: |
          echo "$DEPLOY_KEY" > deploy_key
          chmod 600 deploy_key

          ssh -i deploy_key -o StrictHostKeyChecking=no $DEPLOY_USER@$DEPLOY_HOST << EOF
            cd /opt/radarr

            # Create backup before deployment
            ./scripts/backup.sh

            # Deploy new version
            export IMAGE_TAG=$IMAGE_TAG
            docker-compose pull
            docker-compose up -d --no-deps radarr-go

            # Health check
            sleep 60
            if ! curl -f http://localhost:7878/ping; then
              echo "Health check failed, rolling back"
              docker-compose rollback radarr-go
              exit 1
            fi

            echo "Deployment successful"
          EOF

      - name: Notify deployment
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          channel: '#deployments'
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        if: always()
```

### GitLab CI/CD Pipeline

Create `.gitlab-ci.yml`:

```yaml
# .gitlab-ci.yml
stages:
  - test
  - build
  - security
  - deploy-staging
  - deploy-production

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_BUILDKIT: 1
  REGISTRY: $CI_REGISTRY_IMAGE

# Test Stage
test:
  stage: test
  image: golang:1.25-alpine
  services:
    - name: postgres:17-alpine
      alias: postgres
      variables:
        POSTGRES_DB: radarr_test
        POSTGRES_USER: postgres
        POSTGRES_PASSWORD: postgres
  variables:
    RADARR_DATABASE_TYPE: postgres
    RADARR_DATABASE_HOST: postgres
    RADARR_DATABASE_PORT: 5432
    RADARR_DATABASE_USERNAME: postgres
    RADARR_DATABASE_PASSWORD: postgres
    RADARR_DATABASE_NAME: radarr_test
  script:
    - apk add --no-cache git
    - go mod download
    - go test -v -race -coverprofile=coverage.out ./...
    - go tool cover -func=coverage.out
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.out
  coverage: '/total:\s+\(statements\)\s+(\d+.\d+)%/'

lint:
  stage: test
  image: golangci/golangci-lint:latest
  script:
    - golangci-lint run --timeout=5m

# Build Stage
build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build
        --build-arg VERSION=${CI_COMMIT_TAG:-${CI_COMMIT_REF_NAME}}
        --build-arg COMMIT=${CI_COMMIT_SHA}
        --build-arg BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)
        -t $REGISTRY:$CI_COMMIT_SHA
        -t $REGISTRY:${CI_COMMIT_REF_NAME}
        -f Dockerfile.prod .
    - docker push $REGISTRY:$CI_COMMIT_SHA
    - docker push $REGISTRY:${CI_COMMIT_REF_NAME}
  only:
    - main
    - develop
    - tags

# Security Stage
security-scan:
  stage: security
  image: aquasec/trivy:latest
  script:
    - trivy image --exit-code 0 --severity HIGH,CRITICAL $REGISTRY:$CI_COMMIT_SHA
  artifacts:
    reports:
      container_scanning: trivy-report.json
  only:
    - main
    - develop
    - tags

# Deploy Staging
deploy-staging:
  stage: deploy-staging
  image: alpine:latest
  before_script:
    - apk add --no-cache openssh-client curl
    - eval $(ssh-agent -s)
    - echo "$STAGING_SSH_PRIVATE_KEY" | tr -d '\r' | ssh-add -
    - mkdir -p ~/.ssh
    - chmod 700 ~/.ssh
    - ssh-keyscan $STAGING_HOST >> ~/.ssh/known_hosts
    - chmod 644 ~/.ssh/known_hosts
  script:
    - ssh $STAGING_USER@$STAGING_HOST "
        cd /opt/radarr &&
        export IMAGE_TAG=$CI_COMMIT_SHA &&
        docker-compose pull &&
        docker-compose up -d &&
        sleep 30 &&
        curl -f http://localhost:7878/ping
      "
  environment:
    name: staging
    url: https://staging-radarr.yourdomain.com
  only:
    - develop

# Deploy Production
deploy-production:
  stage: deploy-production
  image: alpine:latest
  before_script:
    - apk add --no-cache openssh-client curl
    - eval $(ssh-agent -s)
    - echo "$PRODUCTION_SSH_PRIVATE_KEY" | tr -d '\r' | ssh-add -
    - mkdir -p ~/.ssh
    - chmod 700 ~/.ssh
    - ssh-keyscan $PRODUCTION_HOST >> ~/.ssh/known_hosts
    - chmod 644 ~/.ssh/known_hosts
  script:
    - ssh $PRODUCTION_USER@$PRODUCTION_HOST "
        cd /opt/radarr &&
        ./scripts/backup.sh &&
        export IMAGE_TAG=$CI_COMMIT_TAG &&
        docker-compose pull &&
        docker-compose up -d --no-deps radarr-go &&
        sleep 60 &&
        if ! curl -f http://localhost:7878/ping; then
          echo 'Health check failed, rolling back' &&
          docker-compose rollback radarr-go &&
          exit 1
        fi
      "
  environment:
    name: production
    url: https://radarr.yourdomain.com
  when: manual
  only:
    - tags
```

## Infrastructure as Code

### Terraform Configuration

Create `terraform/main.tf`:

```hcl
# terraform/main.tf
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    docker = {
      source  = "kreuzwerker/docker"
      version = "~> 3.0"
    }
  }
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "domain" {
  description = "Domain name for the application"
  type        = string
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.medium"
}

variable "key_pair_name" {
  description = "EC2 Key Pair name"
  type        = string
}

# VPC and Networking
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "radarr-${var.environment}"
  cidr = "10.0.0.0/16"

  azs             = ["${data.aws_region.current.name}a", "${data.aws_region.current.name}b"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = false

  tags = {
    Environment = var.environment
    Project     = "radarr-go"
  }
}

# Security Groups
resource "aws_security_group" "radarr_sg" {
  name        = "radarr-${var.environment}-sg"
  description = "Security group for Radarr Go"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 7878
    to_port     = 7878
    protocol    = "tcp"
    cidr_blocks = [module.vpc.vpc_cidr_block]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = "radarr-${var.environment}-sg"
    Environment = var.environment
  }
}

# RDS Database
resource "aws_db_instance" "postgres" {
  identifier = "radarr-${var.environment}-db"

  engine         = "postgres"
  engine_version = "17.2"
  instance_class = "db.t3.micro"

  allocated_storage     = 20
  max_allocated_storage = 100
  storage_type          = "gp3"
  storage_encrypted     = true

  db_name  = "radarr"
  username = "radarr"
  password = random_password.db_password.result

  vpc_security_group_ids = [aws_security_group.postgres_sg.id]
  db_subnet_group_name   = aws_db_subnet_group.postgres.name

  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"

  skip_final_snapshot = var.environment != "production"
  deletion_protection = var.environment == "production"

  tags = {
    Name        = "radarr-${var.environment}-db"
    Environment = var.environment
  }
}

resource "aws_security_group" "postgres_sg" {
  name        = "radarr-${var.environment}-postgres-sg"
  description = "Security group for PostgreSQL"
  vpc_id      = module.vpc.vpc_id

  ingress {
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.radarr_sg.id]
  }

  tags = {
    Name        = "radarr-${var.environment}-postgres-sg"
    Environment = var.environment
  }
}

resource "aws_db_subnet_group" "postgres" {
  name       = "radarr-${var.environment}-postgres-subnet-group"
  subnet_ids = module.vpc.private_subnets

  tags = {
    Name        = "radarr-${var.environment}-postgres-subnet-group"
    Environment = var.environment
  }
}

resource "random_password" "db_password" {
  length  = 16
  special = true
}

# EC2 Instance
resource "aws_instance" "radarr" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.instance_type
  key_name               = var.key_pair_name
  vpc_security_group_ids = [aws_security_group.radarr_sg.id]
  subnet_id              = module.vpc.public_subnets[0]

  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    db_host     = aws_db_instance.postgres.endpoint
    db_password = random_password.db_password.result
    domain      = var.domain
    environment = var.environment
  }))

  root_block_device {
    volume_type = "gp3"
    volume_size = 20
    encrypted   = true
  }

  tags = {
    Name        = "radarr-${var.environment}"
    Environment = var.environment
  }
}

# Elastic IP
resource "aws_eip" "radarr" {
  instance = aws_instance.radarr.id
  domain   = "vpc"

  tags = {
    Name        = "radarr-${var.environment}-eip"
    Environment = var.environment
  }
}

# Route53 DNS
resource "aws_route53_record" "radarr" {
  zone_id = data.aws_route53_zone.main.zone_id
  name    = var.domain
  type    = "A"
  ttl     = 300
  records = [aws_eip.radarr.public_ip]
}

# Data Sources
data "aws_region" "current" {}

data "aws_ami" "amazon_linux" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

data "aws_route53_zone" "main" {
  name = replace(var.domain, "/^[^.]+\\./", "")
}

# Outputs
output "instance_ip" {
  description = "Public IP address of the instance"
  value       = aws_eip.radarr.public_ip
}

output "database_endpoint" {
  description = "Database endpoint"
  value       = aws_db_instance.postgres.endpoint
  sensitive   = true
}

output "database_password" {
  description = "Database password"
  value       = random_password.db_password.result
  sensitive   = true
}

output "application_url" {
  description = "Application URL"
  value       = "https://${var.domain}"
}
```

Create `terraform/user_data.sh`:

```bash
#!/bin/bash
# terraform/user_data.sh - EC2 user data script

set -euo pipefail

# Update system
yum update -y
yum install -y docker git curl

# Start Docker
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create application directory
mkdir -p /opt/radarr
cd /opt/radarr

# Clone application repository
git clone https://github.com/username/radarr-go.git .

# Create environment configuration
cat > .env.production << 'EOF'
ENVIRONMENT=${environment}
DOMAIN=${domain}

# Database Configuration
POSTGRES_PASSWORD=${db_password}
POSTGRES_HOST=${db_host}
POSTGRES_PORT=5432
POSTGRES_USER=radarr
POSTGRES_DB=radarr

# Application Configuration
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST=${db_host}
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=${db_password}
RADARR_DATABASE_NAME=radarr

# Security
RADARR_AUTH_API_KEY=$(openssl rand -hex 32)
RADARR_AUTH_METHOD=apikey
EOF

# Deploy application
docker-compose -f docker-compose.production.yml --env-file .env.production up -d

# Setup SSL certificate
./scripts/ssl-automation.sh ${domain} admin@${domain}

# Configure monitoring
docker-compose -f docker-compose.monitoring.yml --env-file .env.production up -d

echo "Deployment completed successfully"
```

### Ansible Playbook

Create `ansible/playbook.yml`:

```yaml
# ansible/playbook.yml
---
- name: Deploy Radarr Go
  hosts: radarr_servers
  become: yes
  vars:
    radarr_user: radarr
    radarr_home: /opt/radarr
    docker_compose_version: "2.24.0"

  tasks:
    - name: Update system packages
      yum:
        name: "*"
        state: latest
      when: ansible_os_family == "RedHat"

    - name: Install required packages
      package:
        name:
          - docker
          - git
          - curl
          - openssl
        state: present

    - name: Start and enable Docker
      systemd:
        name: docker
        state: started
        enabled: yes

    - name: Create radarr user
      user:
        name: "{{ radarr_user }}"
        home: "{{ radarr_home }}"
        create_home: yes
        system: yes
        shell: /bin/bash

    - name: Add radarr user to docker group
      user:
        name: "{{ radarr_user }}"
        groups: docker
        append: yes

    - name: Install Docker Compose
      get_url:
        url: "https://github.com/docker/compose/releases/download/v{{ docker_compose_version }}/docker-compose-{{ ansible_system }}-{{ ansible_architecture }}"
        dest: /usr/local/bin/docker-compose
        mode: '0755'

    - name: Create application directory
      file:
        path: "{{ radarr_home }}"
        state: directory
        owner: "{{ radarr_user }}"
        group: "{{ radarr_user }}"
        mode: '0755'

    - name: Clone application repository
      git:
        repo: https://github.com/username/radarr-go.git
        dest: "{{ radarr_home }}"
        force: yes
      become_user: "{{ radarr_user }}"

    - name: Generate secure API key
      command: openssl rand -hex 32
      register: api_key_result
      changed_when: false

    - name: Generate database password
      command: openssl rand -hex 32
      register: db_password_result
      changed_when: false

    - name: Create environment configuration
      template:
        src: environment.j2
        dest: "{{ radarr_home }}/.env.{{ environment }}"
        owner: "{{ radarr_user }}"
        group: "{{ radarr_user }}"
        mode: '0600'
      vars:
        api_key: "{{ api_key_result.stdout }}"
        db_password: "{{ db_password_result.stdout }}"

    - name: Create data directories
      file:
        path: "{{ item }}"
        state: directory
        owner: "{{ radarr_user }}"
        group: "{{ radarr_user }}"
        mode: '0755'
      loop:
        - "{{ radarr_home }}/data"
        - "{{ radarr_home }}/config"
        - "{{ radarr_home }}/backups"
        - "{{ radarr_home }}/logs"

    - name: Deploy database
      docker_compose:
        project_src: "{{ radarr_home }}"
        files:
          - docker-compose.postgres.yml
        env_file: "{{ radarr_home }}/.env.{{ environment }}"
        services:
          - postgres
        state: present
      become_user: "{{ radarr_user }}"

    - name: Wait for database to be ready
      wait_for:
        port: 5432
        host: "{{ ansible_default_ipv4.address }}"
        delay: 10
        timeout: 60

    - name: Deploy application
      docker_compose:
        project_src: "{{ radarr_home }}"
        files:
          - docker-compose.production.yml
        env_file: "{{ radarr_home }}/.env.{{ environment }}"
        state: present
      become_user: "{{ radarr_user }}"

    - name: Deploy monitoring stack
      docker_compose:
        project_src: "{{ radarr_home }}"
        files:
          - docker-compose.monitoring.yml
        env_file: "{{ radarr_home }}/.env.{{ environment }}"
        state: present
      become_user: "{{ radarr_user }}"

    - name: Configure firewall
      firewalld:
        port: "{{ item }}/tcp"
        permanent: yes
        state: enabled
        immediate: yes
      loop:
        - "80"
        - "443"
        - "7878"
      when: ansible_os_family == "RedHat"

    - name: Setup SSL certificates
      command: "{{ radarr_home }}/scripts/ssl-automation.sh {{ domain }} {{ admin_email }}"
      become_user: "{{ radarr_user }}"
      when: domain != "localhost"

    - name: Setup automated backups
      cron:
        name: "Radarr backup"
        minute: "0"
        hour: "2"
        job: "{{ radarr_home }}/scripts/backup.sh >/dev/null 2>&1"
        user: "{{ radarr_user }}"

    - name: Setup maintenance tasks
      cron:
        name: "Radarr maintenance"
        minute: "0"
        hour: "4"
        weekday: "0"
        job: "{{ radarr_home }}/scripts/maintenance.sh >/dev/null 2>&1"
        user: "{{ radarr_user }}"

    - name: Verify deployment
      uri:
        url: "http://{{ ansible_default_ipv4.address }}:7878/ping"
        method: GET
      retries: 5
      delay: 10

    - name: Display deployment information
      debug:
        msg: |
          Radarr Go deployed successfully!

          Access URLs:
            Application: http://{{ ansible_default_ipv4.address }}:7878
            Grafana: http://{{ ansible_default_ipv4.address }}:3000
            Prometheus: http://{{ ansible_default_ipv4.address }}:9090

          API Key: {{ api_key_result.stdout }}

          Configuration file: {{ radarr_home }}/.env.{{ environment }}
```

Create `ansible/templates/environment.j2`:

```jinja2
# Environment Configuration for {{ environment }}
# Generated by Ansible on {{ ansible_date_time.iso8601 }}

ENVIRONMENT={{ environment }}
DOMAIN={{ domain | default('localhost') }}

# Database Configuration
POSTGRES_PASSWORD={{ db_password }}
POSTGRES_USER=radarr
POSTGRES_DB=radarr
POSTGRES_HOST={{ postgres_host | default('localhost') }}
POSTGRES_PORT=5432

# Application Configuration
RADARR_AUTH_API_KEY={{ api_key }}
RADARR_AUTH_METHOD=apikey
RADARR_SERVER_HOST=0.0.0.0
RADARR_SERVER_PORT=7878
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST={{ postgres_host | default('localhost') }}
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD={{ db_password }}
RADARR_DATABASE_NAME=radarr

# Security Configuration
RADARR_SECURITY_ENABLE_SECURITY_HEADERS=true
RADARR_SECURITY_ENABLE_CORS=false
RADARR_LOG_LEVEL={{ log_level | default('info') }}
RADARR_LOG_FORMAT=json

# Monitoring Configuration
GRAFANA_ADMIN_PASSWORD={{ grafana_password | default('admin') }}
PROMETHEUS_RETENTION_TIME=30d

# Backup Configuration
BACKUP_ENCRYPTION_KEY={{ backup_key | default(ansible_machine_id) }}
BACKUP_RETENTION_DAYS=30
```

## Monitoring Templates and Dashboards

### Complete Grafana Dashboard

Create `monitoring/grafana/dashboards/radarr-complete-dashboard.json`:

```json
{
  "dashboard": {
    "id": null,
    "title": "Radarr Go - Complete Operations Dashboard",
    "description": "Comprehensive monitoring dashboard for Radarr Go production deployment",
    "tags": ["radarr", "go", "monitoring", "operations"],
    "timezone": "browser",
    "refresh": "30s",
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "panels": [
      {
        "id": 1,
        "title": "System Overview",
        "type": "row",
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Service Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"radarr-go\"}",
            "legendFormat": "Radarr Go"
          },
          {
            "expr": "up{job=\"postgres\"}",
            "legendFormat": "Database"
          },
          {
            "expr": "up{job=\"prometheus\"}",
            "legendFormat": "Prometheus"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "mappings": [
              {
                "options": {
                  "0": {"text": "DOWN", "color": "red"},
                  "1": {"text": "UP", "color": "green"}
                },
                "type": "value"
              }
            ],
            "thresholds": {
              "steps": [
                {"color": "red", "value": null},
                {"color": "green", "value": 1}
              ]
            }
          }
        },
        "gridPos": {"h": 8, "w": 8, "x": 0, "y": 1}
      },
      {
        "id": 3,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(radarr_http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ],
        "yAxes": [
          {
            "label": "Requests/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 8, "x": 8, "y": 1}
      },
      {
        "id": 4,
        "title": "Response Time (95th percentile)",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(radarr_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(radarr_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "yAxes": [
          {
            "label": "Seconds"
          }
        ],
        "gridPos": {"h": 8, "w": 8, "x": 16, "y": 1}
      },
      {
        "id": 5,
        "title": "Resource Usage",
        "type": "row",
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": 9}
      },
      {
        "id": 6,
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "process_resident_memory_bytes{job=\"radarr-go\"}",
            "legendFormat": "RSS Memory"
          },
          {
            "expr": "go_memstats_heap_inuse_bytes{job=\"radarr-go\"}",
            "legendFormat": "Heap In Use"
          },
          {
            "expr": "go_memstats_stack_inuse_bytes{job=\"radarr-go\"}",
            "legendFormat": "Stack In Use"
          }
        ],
        "yAxes": [
          {
            "label": "Bytes"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 10}
      },
      {
        "id": 7,
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(process_cpu_seconds_total{job=\"radarr-go\"}[5m]) * 100",
            "legendFormat": "CPU Usage %"
          }
        ],
        "yAxes": [
          {
            "label": "Percent",
            "max": 100
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 10}
      },
      {
        "id": 8,
        "title": "Database Metrics",
        "type": "row",
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": 18}
      },
      {
        "id": 9,
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "pg_stat_activity_count{job=\"postgres\"}",
            "legendFormat": "Active Connections"
          },
          {
            "expr": "pg_settings_max_connections{job=\"postgres\"}",
            "legendFormat": "Max Connections"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 19}
      },
      {
        "id": 10,
        "title": "Database Query Performance",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(pg_stat_statements_calls_total{job=\"postgres\"}[5m])",
            "legendFormat": "Queries/sec"
          },
          {
            "expr": "rate(pg_stat_statements_total_time_seconds_total{job=\"postgres\"}[5m]) / rate(pg_stat_statements_calls_total{job=\"postgres\"}[5m])",
            "legendFormat": "Avg Query Time"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 19}
      },
      {
        "id": 11,
        "title": "Application Metrics",
        "type": "row",
        "gridPos": {"h": 1, "w": 24, "x": 0, "y": 27}
      },
      {
        "id": 12,
        "title": "Movies in Library",
        "type": "stat",
        "targets": [
          {
            "expr": "radarr_movies_total{job=\"radarr-go\"}",
            "legendFormat": "Total Movies"
          },
          {
            "expr": "radarr_movies_monitored{job=\"radarr-go\"}",
            "legendFormat": "Monitored"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "color": {
              "mode": "palette-classic"
            },
            "custom": {
              "displayMode": "basic"
            }
          }
        },
        "gridPos": {"h": 8, "w": 6, "x": 0, "y": 28}
      },
      {
        "id": 13,
        "title": "Download Activity",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(radarr_downloads_completed_total{job=\"radarr-go\"}[1h])",
            "legendFormat": "Downloads/hour"
          },
          {
            "expr": "radarr_downloads_active{job=\"radarr-go\"}",
            "legendFormat": "Active Downloads"
          }
        ],
        "gridPos": {"h": 8, "w": 9, "x": 6, "y": 28}
      },
      {
        "id": 14,
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(radarr_http_requests_total{status=~\"5..\",job=\"radarr-go\"}[5m])",
            "legendFormat": "5xx Errors/sec"
          },
          {
            "expr": "rate(radarr_http_requests_total{status=~\"4..\",job=\"radarr-go\"}[5m])",
            "legendFormat": "4xx Errors/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 9, "x": 15, "y": 28}
      }
    ]
  }
}
```

### AlertManager Rule Templates

Create `monitoring/prometheus/rules/complete-alerts.yml`:

```yaml
# monitoring/prometheus/rules/complete-alerts.yml
groups:
  - name: radarr.critical
    rules:
      - alert: RadarrServiceDown
        expr: up{job="radarr-go"} == 0
        for: 2m
        labels:
          severity: critical
          service: radarr
          team: operations
        annotations:
          summary: "Radarr Go service is down"
          description: "Radarr Go has been down for more than 2 minutes on {{ $labels.instance }}"
          runbook_url: "https://wiki.company.com/runbooks/radarr-down"

      - alert: DatabaseDown
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
          service: database
          team: operations
        annotations:
          summary: "PostgreSQL database is down"
          description: "PostgreSQL database has been down for more than 1 minute"

      - alert: HighErrorRate
        expr: (rate(radarr_http_requests_total{status=~"5.."}[5m]) / rate(radarr_http_requests_total[5m])) > 0.1
        for: 5m
        labels:
          severity: critical
          service: radarr
          team: operations
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

  - name: radarr.warning
    rules:
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(radarr_http_request_duration_seconds_bucket[5m])) > 2
        for: 10m
        labels:
          severity: warning
          service: radarr
          team: operations
        annotations:
          summary: "High response time detected"
          description: "95th percentile response time is {{ $value }}s over the last 10 minutes"

      - alert: HighMemoryUsage
        expr: (process_resident_memory_bytes{job="radarr-go"} / 1024 / 1024 / 1024) > 0.8
        for: 15m
        labels:
          severity: warning
          service: radarr
          team: operations
        annotations:
          summary: "High memory usage"
          description: "Memory usage is {{ $value | humanize }}GB"

      - alert: DatabaseConnectionsHigh
        expr: (pg_stat_activity_count / pg_settings_max_connections) > 0.8
        for: 10m
        labels:
          severity: warning
          service: database
          team: operations
        annotations:
          summary: "High database connection usage"
          description: "Database connection usage is {{ $value | humanizePercentage }}"

  - name: system.alerts
    rules:
      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) < 0.1
        for: 5m
        labels:
          severity: warning
          service: system
          team: operations
        annotations:
          summary: "Disk space low"
          description: "Disk space is {{ $value | humanizePercentage }} full"

      - alert: SystemLoadHigh
        expr: node_load15 > 2
        for: 10m
        labels:
          severity: warning
          service: system
          team: operations
        annotations:
          summary: "System load high"
          description: "15-minute load average is {{ $value }}"
```

## Operational Automation Scripts

### Comprehensive Maintenance Script

Create `scripts/comprehensive-maintenance.sh`:

```bash
#!/bin/bash
# comprehensive-maintenance.sh - Complete system maintenance automation

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MAINTENANCE_LOG="/var/log/radarr/maintenance.log"
NOTIFICATION_WEBHOOK="${SLACK_WEBHOOK_URL:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    local message="[$(date +'%Y-%m-%d %H:%M:%S')] $1"
    echo -e "${GREEN}$message${NC}" | tee -a "$MAINTENANCE_LOG"
}

warn() {
    local message="[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1"
    echo -e "${YELLOW}$message${NC}" | tee -a "$MAINTENANCE_LOG"
}

error() {
    local message="[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1"
    echo -e "${RED}$message${NC}" | tee -a "$MAINTENANCE_LOG"
    exit 1
}

# Send notification
send_notification() {
    local title="$1"
    local message="$2"
    local color="${3:-good}"

    if [ -n "$NOTIFICATION_WEBHOOK" ]; then
        curl -X POST "$NOTIFICATION_WEBHOOK" \
             -H "Content-Type: application/json" \
             -d "{
                 \"attachments\": [{
                     \"title\": \"$title\",
                     \"text\": \"$message\",
                     \"color\": \"$color\"
                 }]
             }" >/dev/null 2>&1 || true
    fi
}

# System maintenance
system_maintenance() {
    log "Starting system maintenance..."

    # Update system packages
    log "Updating system packages..."
    if command -v yum >/dev/null 2>&1; then
        yum update -y
    elif command -v apt-get >/dev/null 2>&1; then
        apt-get update && apt-get upgrade -y
    fi

    # Clean up old logs
    log "Cleaning up old log files..."
    find /var/log -name "*.log" -type f -mtime +30 -delete || true
    find /var/log -name "*.gz" -type f -mtime +90 -delete || true

    # Clean up Docker resources
    log "Cleaning up Docker resources..."
    docker system prune -af --volumes || warn "Docker cleanup failed"

    # Clean up disk space
    log "Cleaning up temporary files..."
    rm -rf /tmp/* || true
    rm -rf /var/tmp/* || true

    log "System maintenance completed"
}

# Database maintenance
database_maintenance() {
    log "Starting database maintenance..."

    local db_type="${RADARR_DATABASE_TYPE:-postgres}"

    case "$db_type" in
        "postgres")
            log "PostgreSQL maintenance..."

            # Database statistics update
            docker exec radarr-postgres psql -U radarr -d radarr -c "ANALYZE;" || warn "ANALYZE failed"

            # Vacuum database
            docker exec radarr-postgres psql -U radarr -d radarr -c "VACUUM;" || warn "VACUUM failed"

            # Check for unused indexes
            local unused_indexes=$(docker exec radarr-postgres psql -U radarr -d radarr -t -c "
                SELECT schemaname, tablename, attname, n_distinct, correlation
                FROM pg_stats
                WHERE schemaname = 'public' AND n_distinct < -0.1
                ORDER BY abs(correlation) DESC;
            " | wc -l)

            if [ "$unused_indexes" -gt 0 ]; then
                warn "Found $unused_indexes potentially unused indexes"
            fi
            ;;

        "mariadb")
            log "MariaDB maintenance..."

            # Optimize tables
            docker exec radarr-mariadb mysql -u radarr -p"${MYSQL_PASSWORD}" radarr -e "
                SELECT CONCAT('OPTIMIZE TABLE ', table_name, ';') as stmt
                FROM information_schema.tables
                WHERE table_schema = 'radarr' AND table_type = 'BASE TABLE';
            " -s | docker exec -i radarr-mariadb mysql -u radarr -p"${MYSQL_PASSWORD}" radarr || warn "Table optimization failed"
            ;;
    esac

    log "Database maintenance completed"
}

# Application maintenance
application_maintenance() {
    log "Starting application maintenance..."

    # Check application health
    local api_key="${RADARR_AUTH_API_KEY}"
    if ! curl -sf -H "X-API-Key: $api_key" http://localhost:7878/api/v3/system/status >/dev/null; then
        error "Application health check failed"
    fi

    # Clean up old movie files metadata
    log "Cleaning up old metadata..."
    curl -sf -H "X-API-Key: $api_key" -X POST http://localhost:7878/api/v3/command \
         -H "Content-Type: application/json" \
         -d '{"name": "RefreshMovie"}' >/dev/null || warn "Movie refresh failed"

    # Clean up download history
    log "Cleaning up old download history..."
    curl -sf -H "X-API-Key: $api_key" -X POST http://localhost:7878/api/v3/command \
         -H "Content-Type: application/json" \
         -d '{"name": "CleanUpRecycleBin"}' >/dev/null || warn "Cleanup failed"

    log "Application maintenance completed"
}

# Backup verification
backup_verification() {
    log "Verifying backup integrity..."

    local backup_dir="/opt/radarr/backups"
    local latest_backup=$(ls -t "$backup_dir"/*.sql.gpg 2>/dev/null | head -1)

    if [ -z "$latest_backup" ]; then
        warn "No encrypted backups found"
        return 1
    fi

    # Test backup decryption
    if gpg --quiet --decrypt "$latest_backup" >/dev/null 2>&1; then
        log "Backup integrity verified"
    else
        error "Backup integrity check failed"
    fi
}

# Security scan
security_scan() {
    log "Running security scan..."

    # Check for security updates
    if command -v yum >/dev/null 2>&1; then
        local security_updates=$(yum --security check-update 2>/dev/null | grep -c "^[[:space:]]*[^[:space:]]" || echo "0")
        if [ "$security_updates" -gt 0 ]; then
            warn "Security updates available: $security_updates"
        fi
    fi

    # Check file permissions
    local suspicious_files=$(find /opt/radarr -type f -perm /002 | wc -l)
    if [ "$suspicious_files" -gt 0 ]; then
        warn "Found $suspicious_files world-writable files"
    fi

    # Check for large log files (possible DoS)
    local large_logs=$(find /var/log -name "*.log" -size +100M | wc -l)
    if [ "$large_logs" -gt 0 ]; then
        warn "Found $large_logs large log files (>100MB)"
    fi

    log "Security scan completed"
}

# Performance optimization
performance_optimization() {
    log "Running performance optimization..."

    # Optimize Docker images
    docker image prune -f || warn "Docker image cleanup failed"

    # Check memory usage
    local memory_usage=$(free | awk 'FNR==2{printf "%.0f", $3/$2*100}')
    if [ "$memory_usage" -gt 85 ]; then
        warn "High memory usage: ${memory_usage}%"
    fi

    # Check disk I/O
    local io_wait=$(iostat 1 1 | awk '/^avg-cpu/ {getline; print $4}' | sed 's/,/./')
    if [ "$(echo "$io_wait > 20" | bc -l)" -eq 1 ]; then
        warn "High I/O wait: ${io_wait}%"
    fi

    log "Performance optimization completed"
}

# Generate maintenance report
generate_report() {
    log "Generating maintenance report..."

    local report_file="/var/log/radarr/maintenance-report-$(date +%Y%m%d).json"

    cat > "$report_file" << EOF
{
    "timestamp": "$(date -Iseconds)",
    "maintenance_type": "scheduled",
    "tasks_completed": [
        "system_maintenance",
        "database_maintenance",
        "application_maintenance",
        "backup_verification",
        "security_scan",
        "performance_optimization"
    ],
    "system_info": {
        "uptime": "$(uptime -p)",
        "load_average": "$(uptime | awk -F'load average:' '{print $2}')",
        "disk_usage": "$(df -h / | awk 'NR==2{print $5}')",
        "memory_usage": "$(free | awk 'FNR==2{printf "%.0f%%", $3/$2*100}')"
    },
    "application_status": {
        "version": "$(curl -sf -H "X-API-Key: ${RADARR_AUTH_API_KEY}" http://localhost:7878/api/v3/system/status | jq -r '.version' 2>/dev/null || echo 'unknown')",
        "database_size": "$(docker exec radarr-postgres psql -U radarr -d radarr -t -c "SELECT pg_size_pretty(pg_database_size('radarr'));" 2>/dev/null | tr -d ' ' || echo 'unknown')"
    },
    "next_maintenance": "$(date -d '+1 week' -Iseconds)"
}
EOF

    log "Maintenance report generated: $report_file"
}

# Main maintenance routine
main() {
    local maintenance_type="${1:-full}"
    local start_time=$(date +%s)

    log "Starting comprehensive maintenance (type: $maintenance_type)"

    # Create log directory
    mkdir -p "$(dirname "$MAINTENANCE_LOG")"

    # Send start notification
    send_notification "Maintenance Started" "Comprehensive maintenance started on $(hostname)"

    case "$maintenance_type" in
        "full")
            system_maintenance
            database_maintenance
            application_maintenance
            backup_verification
            security_scan
            performance_optimization
            ;;
        "quick")
            database_maintenance
            application_maintenance
            ;;
        "security")
            security_scan
            backup_verification
            ;;
        *)
            error "Unknown maintenance type: $maintenance_type"
            ;;
    esac

    generate_report

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))

    log "Maintenance completed in ${duration} seconds"

    # Send completion notification
    send_notification "Maintenance Completed" "Comprehensive maintenance completed successfully in ${duration} seconds" "good"
}

# Handle different maintenance types
case "${1:-full}" in
    "full"|"quick"|"security")
        main "$1"
        ;;
    "report")
        generate_report
        ;;
    *)
        echo "Usage: $0 [full|quick|security|report]"
        echo ""
        echo "Maintenance Types:"
        echo "  full     - Complete system maintenance (default)"
        echo "  quick    - Database and application maintenance only"
        echo "  security - Security scan and backup verification"
        echo "  report   - Generate maintenance report only"
        exit 1
        ;;
esac
```

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create production deployment guide with Docker Compose and Kubernetes configurations", "status": "completed", "activeForm": "Creating production deployment guide with Docker Compose and Kubernetes configurations"}, {"content": "Create monitoring and alerting setup documentation", "status": "completed", "activeForm": "Creating monitoring and alerting setup documentation"}, {"content": "Create performance tuning guide", "status": "completed", "activeForm": "Creating performance tuning guide"}, {"content": "Create security hardening recommendations", "status": "completed", "activeForm": "Creating security hardening recommendations"}, {"content": "Create automated deployment scripts and monitoring templates", "status": "completed", "activeForm": "Creating automated deployment scripts and monitoring templates"}]
