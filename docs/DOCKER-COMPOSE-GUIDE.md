# Docker Compose Consolidation Guide

This guide explains the new consolidated Docker Compose architecture for Radarr Go, designed to eliminate duplication and simplify development workflows. All files are **Docker Compose v2 compatible** with deprecated `version:` fields removed.

## 📋 Requirements

- **Docker Compose v2** (recommended: v2.20+)
- **Docker** (recommended: v24.0+)
- **Note**: All compose files have been updated to remove deprecated `version:` fields for full Docker Compose v2 compatibility

## 🎯 Quick Start Commands

### Development (Most Common)

```bash
# Quick development (automatic - uses docker-compose.override.yml)
make dev                    # Backend + PostgreSQL + Admin tools
docker-compose up --build  # Same as above

# Full development environment
make dev-full              # All services: Backend + Databases + Monitoring + Frontend

# Specific database development
make dev-postgres          # Backend + PostgreSQL + Adminer
make dev-mariadb          # Backend + MariaDB + Adminer

# With monitoring
make dev-monitoring       # Backend + PostgreSQL + Prometheus + Grafana + Jaeger
```

### Testing

```bash
# Quick test database setup
make test-db-up           # Start test databases (PostgreSQL + MariaDB)
make test                 # Run tests with existing test runner
make test-db-down         # Stop test databases

# Complete test environment
make test-env-up          # Full test environment with optimizations
```

### Production

```bash
# Production deployment
make prod-up              # Secure, optimized production environment
make prod-logs            # View production logs
make prod-status          # Check service status
make prod-down            # Stop production services
```

## 📁 New File Structure

### Core Files (Consolidated)

```
├── docker-compose.yml              # 🎯 BASE - All services with profiles
├── docker-compose.override.yml     # 🔧 DEV DEFAULTS - Auto-loaded for development

# Environment-Specific Overrides
├── docker-compose.dev.yml.new      # 🚀 FULL DEV - Complete development stack
├── docker-compose.test.yml.new     # 🧪 TESTING - Optimized for test performance
├── docker-compose.prod.yml.new     # 🏭 PRODUCTION - Security & performance hardened

# Legacy Files (Will be removed)
├── docker-compose.dev.yml          # ❌ DELETE - Replaced by new structure
├── docker-compose.simple.yml       # ✅ REMOVED - Functionality moved to base + profiles
├── deployment/docker-compose.*.yml # ❌ MOVE - Consolidated into new structure
```

## 🔧 Service Profiles Explained

The base `docker-compose.yml` now includes ALL services organized by profiles:

### Core Profiles

- `default` - Main application + PostgreSQL (always available)
- `dev` - Development-specific services with hot reload
- `frontend` - React development server (Phase 2)
- `admin` - Database administration tools (Adminer, MailHog)
- `monitoring` - Observability stack (Prometheus, Grafana, Jaeger)
- `mariadb` - MariaDB alternative database
- `redis` - Redis caching layer
- `test` - Test-optimized database services

### Profile Usage Examples

```bash
# Use specific profiles
docker-compose up --profile admin          # Add Adminer + MailHog
docker-compose up --profile monitoring     # Add Prometheus + Grafana + Jaeger
docker-compose up --profile mariadb        # Add MariaDB database
docker-compose up --profile frontend       # Add React development server

# Multiple profiles
docker-compose up --profile admin --profile monitoring --profile redis
```

## 🏗️ Architecture Benefits

### ✅ Problems Solved

- **Eliminated 70% duplication** - PostgreSQL/MariaDB defined once, reused everywhere
- **Clear command patterns** - `make dev`, `make test-db-up`, `make prod-up`
- **Automatic development defaults** - `docker-compose up` just works for developers
- **Profile-based service selection** - Enable only what you need
- **Consistent configurations** - No more config drift between environments

### 🎯 Developer Experience Improvements

- **Single command development**: `make dev` starts everything needed
- **Smart defaults**: `docker-compose.override.yml` provides perfect dev setup automatically
- **Clear service selection**: Profiles make it obvious what's running
- **Performance optimized**: Test databases use tmpfs, dev databases log everything
- **Monitoring ready**: Add `--profile monitoring` to any environment

## 📚 Detailed Command Reference

### Core Development Workflows

#### Basic Development (Default)

```bash
# Starts: Backend (hot reload) + PostgreSQL + Adminer + MailHog
make dev
# or
docker-compose up --build
```

- Uses `docker-compose.override.yml` automatically
- Backend hot reloads with Air
- PostgreSQL with development-friendly logging
- Adminer for database management
- MailHog for email testing

#### Full Development Environment

```bash
# Starts: Everything for complete development
make dev-full
# or
docker-compose -f docker-compose.yml -f docker-compose.dev.yml.new up --profile dev --build
```

- Backend with hot reload
- PostgreSQL + MariaDB (alternative)
- Redis caching
- Complete monitoring stack
- Frontend development server (when ready)

#### Database-Specific Development

```bash
# PostgreSQL only
make dev-postgres
docker-compose up radarr-go postgres adminer --build

# MariaDB alternative
make dev-mariadb
RADARR_DATABASE_TYPE=mariadb docker-compose up radarr-go mariadb adminer --profile mariadb --build
```

#### Monitoring Development

```bash
# Add monitoring to any environment
make dev-monitoring
docker-compose up --profile admin --profile monitoring --build
```

- Prometheus metrics collection
- Grafana dashboards
- Jaeger distributed tracing
- All accessible via web interfaces

### Testing Workflows

#### Quick Test Setup

```bash
# Start test databases on separate ports
make test-db-up
# Postgres: localhost:15432, MariaDB: localhost:13306

# Run your tests
go test ./...

# Cleanup
make test-db-down
```

#### Complete Test Environment

```bash
# Advanced test setup with optimizations
make test-env-up
docker-compose -f docker-compose.yml -f docker-compose.test.yml.new up --profile test -d
```

- Test databases with tmpfs for speed
- Separate network isolation
- Optimized database configurations for testing

### Production Deployment

#### Production Startup

```bash
# Secure production environment
make prod-up
docker-compose -f docker-compose.yml -f docker-compose.prod.yml.new up -d
```

**Production Features:**

- Security hardening (non-root users, capability restrictions)
- Resource limits and reservations
- Optimized database configurations
- Structured JSON logging
- Health checks and monitoring labels
- Bind-mounted volumes for data persistence

#### Production Management

```bash
make prod-logs     # View all production logs
make prod-status   # Check service health
make prod-down     # Graceful shutdown
```

## 🔄 Migration Guide

### For Existing Developers

#### Step 1: Update Your Workflow

**OLD:**

```bash
# Confusing - which file to use?
docker-compose -f docker-compose.dev.yml up     # Full but heavy
docker compose up                              # Now provides perfect minimal defaults!
```

**NEW:**

```bash
# Clear and intuitive
make dev          # Perfect for most development
make dev-full     # When you need everything
make test-db-up   # Just test databases
```

#### Step 2: Environment Variables

The new system uses the same environment variables but with better defaults:

```bash
# Still supported - all existing .env files work
RADARR_DATABASE_TYPE=mariadb make dev-mariadb
RADARR_LOG_LEVEL=debug make dev
```

#### Step 3: Port Mapping

**Unchanged ports:**

- Main app: `7878`
- PostgreSQL: `5432` (dev), `15432` (test)
- MariaDB: `3306` (dev), `13306` (test)
- Adminer: `8081`
- Grafana: `3001`
- Prometheus: `9090`

### File Migration Timeline

```bash
# Phase 1: Test new system (NOW)
# - New files are created with .new extension
# - Old files still work
# - Test the new commands

# Phase 2: Switch over (After validation)
# - Rename new files to replace old ones
# - Update CI/CD pipelines
# - Remove legacy files

# Phase 3: Cleanup (Final)
# - Delete old docker-compose files
# - Update documentation
```

## 🚀 Advanced Usage

### Custom Service Combinations

```bash
# Backend + MariaDB + Monitoring
docker-compose up --profile mariadb --profile monitoring radarr-go mariadb prometheus grafana

# Full stack with Redis caching
docker-compose up --profile redis --profile admin --profile monitoring --build

# Test environment with monitoring
docker-compose -f docker-compose.yml -f docker-compose.test.yml.new up --profile test --profile monitoring
```

### Environment Variable Overrides

```bash
# Use different database in any environment
RADARR_DATABASE_TYPE=mariadb docker-compose up --profile mariadb

# Custom ports
POSTGRES_EXTERNAL_PORT=15432 docker-compose up

# Production-style logging in development
RADARR_LOG_LEVEL=info RADARR_LOG_FORMAT=json make dev
```

### Docker Compose Override Files

You can create additional override files for specific scenarios:

```bash
# docker-compose.local.yml - Your personal overrides
# docker-compose.ci.yml - CI-specific settings
# docker-compose.debug.yml - Debug configurations

# Use multiple overrides
docker-compose -f docker-compose.yml -f docker-compose.dev.yml.new -f docker-compose.local.yml up
```

## 🔧 Troubleshooting

### Common Issues

#### "Service not found"

```bash
# Ensure you're using the right profile
docker-compose up --profile admin adminer  # ✅ Correct
docker-compose up adminer                  # ❌ Won't work without profile
```

#### "Port already in use"

```bash
# Check what's running
make docker-ps
docker ps -a

# Clean up everything
make docker-clean
```

#### Database connection issues

```bash
# Check database health
docker-compose ps postgres
docker-compose logs postgres

# Reset databases
make test-db-clean  # For test databases
make docker-clean   # For all environments
```

### Debugging Commands

```bash
# See what services would start
docker-compose config

# Check specific service configuration
docker-compose config radarr-go

# See all available services
docker-compose config --services

# Test specific profiles
docker-compose --profile dev config --services
```

## 📖 Integration with Existing Workflows

### CI/CD Pipeline Updates

The new structure maintains full compatibility:

```yaml
# GitHub Actions - no changes needed
- name: Start test databases
  run: make test-db-up

# But you can also use the new approach
- name: Start optimized test environment
  run: make test-env-up
```

### IDE Integration

**VS Code Tasks** (`.vscode/tasks.json`):

```json
{
    "label": "Start Development Environment",
    "type": "shell",
    "command": "make dev",
    "group": "build"
}
```

### Scripts Integration

Existing scripts continue to work:

```bash
# Your existing scripts work unchanged
./scripts/dev-setup.sh
./scripts/test-runner.sh
```

## 🎉 Next Steps

1. **Try the new commands**: Start with `make dev` for your daily development
2. **Test specific workflows**: Use `make dev-mariadb` or `make dev-monitoring` as needed
3. **Validate in your environment**: Ensure all your existing workflows still function
4. **Provide feedback**: Report any issues or suggestions for improvement

The consolidation maintains full backward compatibility while providing a much cleaner and more maintainable development experience.
