# Radarr Go Development Guide

This comprehensive guide helps developers set up and work with the Radarr Go development environment, including both backend and frontend development workflows.

## Table of Contents

- [Quick Start](#quick-start)
- [Prerequisites](#prerequisites)
- [Development Environment Setup](#development-environment-setup)
- [Development Workflows](#development-workflows)
- [Frontend Development](#frontend-development)
- [Database Development](#database-development)
- [Testing](#testing)
- [Debugging and Monitoring](#debugging-and-monitoring)
- [Docker Development](#docker-development)
- [Troubleshooting](#troubleshooting)
- [Contributing Guidelines](#contributing-guidelines)

## Quick Start

For new developers, the fastest way to get started:

```bash
# Automated setup (macOS/Linux)
./scripts/dev-setup.sh

# Manual setup
make check-env        # Check prerequisites
make setup           # Install development tools
make dev-full        # Start complete development environment
```

Access your development environment:

- **Backend API**: http://localhost:7878
- **Database Admin**: http://localhost:8081
- **Frontend Dev**: http://localhost:3000 (when React is ready)
- **Monitoring**: http://localhost:9090 (Prometheus)
- **Grafana Dashboard**: http://localhost:3001
- **Tracing**: http://localhost:16686 (Jaeger)

## Prerequisites

### Required Tools

- **Go 1.21+**: Backend development
- **Node.js 20+**: Frontend development (Phase 2)
- **Docker & Docker Compose**: Development databases and services
- **Make**: Build automation
- **Git**: Version control

### Optional Tools (Recommended)

- **air**: Hot reload for Go development
- **golangci-lint**: Code quality and linting
- **migrate**: Database migration management
- **pre-commit**: Git hooks for quality assurance

### Quick Environment Check

```bash
make check-env
```

This command verifies all required and optional tools are installed.

## Development Environment Setup

### Automated Setup (Recommended)

The development setup script installs all required tools and configures the environment:

```bash
./scripts/dev-setup.sh
```

This script:

- Detects your OS (macOS/Linux)
- Installs missing tools via package managers
- Sets up Go development tools
- Creates development directories
- Generates secure development configuration
- Sets up git hooks (if pre-commit is available)

### Manual Setup

If you prefer manual setup or need to customize:

```bash
# Install backend development tools
make setup-backend

# Create frontend structure for Phase 2
make setup-frontend

# Check everything is working
make check-env
```

### Development Configuration

The setup script creates a `config.yaml` with secure defaults. Key settings:

```yaml
server:
  port: 7878
  host: "0.0.0.0"

database:
  type: "postgres"  # or "mariadb"
  host: "localhost"
  port: 5432

log:
  level: "debug"
  format: "json"
```

Override any setting with environment variables:

```bash
export RADARR_SERVER_PORT=8080
export RADARR_LOG_LEVEL=info
export RADARR_DATABASE_TYPE=mariadb
```

## Development Workflows

### Backend Development

#### Basic Backend Development

```bash
# Start backend with hot reload
make dev

# Build and run locally
make run

# Run specific tests
go test ./internal/services -v
```

#### Full Development Environment

```bash
# Start complete environment (backend + databases + monitoring)
make dev-full

# This starts:
# - Go backend with hot reload
# - PostgreSQL database
# - Redis cache
# - Monitoring tools (Prometheus, Grafana, Jaeger)
# - Database admin interface (Adminer)
# - Email testing (MailHog)
```

#### Database-Specific Development

```bash
# Use MariaDB instead of PostgreSQL
docker-compose -f docker-compose.dev.yml --profile mariadb up -d

# Or set environment variable
export RADARR_DATABASE_TYPE=mariadb
make dev
```

### Build Workflows

#### Multi-Platform Building

```bash
# Build for all supported platforms
make build-all

# Build individual platforms
make build-linux-amd64
make build-darwin-arm64
make build-windows-amd64

# Build with frontend (when ready)
make build-all-with-frontend
```

#### Development Builds

```bash
# Quick local build
make build

# Linux build (for Docker testing)
make build-linux

# Clean builds
make clean
make build
```

### Code Quality Workflows

#### Pre-commit Quality Checks

```bash
# Install pre-commit hooks (recommended)
pip install pre-commit
pre-commit install

# Run on all files
pre-commit run --all-files

# Manual quality checks
make fmt         # Format code
make lint        # Run linter
make test        # Run tests
make all         # Format + lint + test + build
```

#### Comprehensive Development Workflow

```bash
make dev-all     # Format + lint + test + examples + bench + coverage + build
```

## Frontend Development

**Note**: Frontend development (React) will be available in Phase 2. Current setup prepares the infrastructure.

### Current State (Phase 1)

```bash
# Create frontend structure
make setup-frontend

# Build placeholder frontend
make build-frontend

# Start placeholder frontend dev server
make dev-frontend
```

### Future State (Phase 2)

When React is implemented, the workflow will be:

```bash
# Install frontend dependencies
make install-frontend

# Start React development server
make dev-frontend

# Build for production
make build-frontend

# Full-stack development
make dev-full           # Backend + databases + monitoring
# In another terminal:
make dev-frontend       # React development server
```

### Frontend Development Environment

The development environment includes:

- **React Development Server**: http://localhost:3000
- **API Proxy**: Configured to proxy API requests to backend
- **Hot Module Replacement**: For fast development cycles
- **Storybook**: http://localhost:3001 (when implemented)

## Database Development

### Database Options

Radarr Go supports multiple databases with easy switching:

```bash
# PostgreSQL (default)
export RADARR_DATABASE_TYPE=postgres
make dev

# MariaDB
export RADARR_DATABASE_TYPE=mariadb
docker-compose -f docker-compose.dev.yml --profile mariadb up -d
make dev
```

### Database Administration

Access database admin interface at http://localhost:8081 (Adminer) when running `make dev-full`.

**PostgreSQL Connection**:

- Server: `postgres-dev`
- Username: `radarr_dev`
- Password: `dev_password`
- Database: `radarr_dev`

**MariaDB Connection**:

- Server: `mariadb-dev`
- Username: `radarr_dev`
- Password: `dev_password`
- Database: `radarr_dev`

### Database Migrations

```bash
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
migrate create -ext sql -dir migrations migration_name
```

### Database Testing

```bash
# Test with PostgreSQL
make test-postgres

# Test with MariaDB
make test-mariadb

# Test both databases
make test
```

## Testing

### Test Types

#### Unit Tests

```bash
make test-unit          # Fast unit tests only
go test -short ./...    # Skip integration tests
```

#### Integration Tests

```bash
make test               # Full integration tests
make test-postgres      # PostgreSQL integration tests
make test-mariadb       # MariaDB integration tests
```

#### Performance Tests

```bash
make test-bench         # Benchmark tests
make test-examples      # Example tests
```

#### Coverage Analysis

```bash
make test-coverage      # Generate HTML coverage report
```

### Test Database Management

```bash
# Start test databases
make test-db-up

# Stop test databases
make test-db-down

# View test database logs
make test-db-logs

# Clean test databases
make test-db-clean
```

### Docker Testing

```bash
# Run tests in Docker (full isolation)
make test-docker
```

## Debugging and Monitoring

### Development Monitoring Stack

When running `make dev-full`, you get a complete monitoring stack:

#### Prometheus Metrics

- **URL**: http://localhost:9090
- **Purpose**: Application metrics and monitoring
- **Configuration**: `scripts/prometheus.yml`

#### Grafana Dashboards

- **URL**: http://localhost:3001
- **Login**: admin/admin
- **Purpose**: Visualization and alerting
- **Configuration**: `scripts/grafana-datasources.yml`

#### Distributed Tracing

- **URL**: http://localhost:16686
- **Purpose**: Request tracing and performance analysis
- **Service**: Jaeger all-in-one

#### Email Testing

- **SMTP**: localhost:1025
- **Web UI**: http://localhost:8025
- **Purpose**: Test notification emails
- **Service**: MailHog

### Application Debugging

#### Debug Port

The backend exposes a debug port (8080) for profiling:

```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine analysis
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

#### Logs

```bash
# View backend logs
docker-compose -f docker-compose.dev.yml logs -f radarr-backend

# View all service logs
docker-compose -f docker-compose.dev.yml logs -f

# View specific service logs
docker-compose -f docker-compose.dev.yml logs -f postgres-dev
```

## Docker Development

### Development Compose Profiles

The development environment uses profiles for different setups:

```bash
# Basic environment (default)
make dev-full

# With MariaDB instead of PostgreSQL
docker-compose -f docker-compose.dev.yml --profile mariadb up -d

# With monitoring tools
docker-compose -f docker-compose.dev.yml --profile monitoring up -d

# With frontend (when React is ready)
docker-compose -f docker-compose.dev.yml --profile frontend up -d

# With Redis caching
docker-compose -f docker-compose.dev.yml --profile redis up -d
```

### Development Services

| Service | Port | Purpose | Profile |
|---------|------|---------|---------|
| radarr-backend | 7878, 8080 | Main API + Debug | default |
| postgres-dev | 5432 | Primary database | default |
| mariadb-dev | 3306 | Alternative database | mariadb |
| redis-dev | 6379 | Caching | redis |
| radarr-frontend | 3000 | React dev server | frontend |
| adminer | 8081 | Database admin | monitoring |
| prometheus | 9090 | Metrics | monitoring |
| grafana | 3001 | Dashboards | monitoring |
| jaeger | 16686 | Tracing | monitoring |
| mailhog | 8025 | Email testing | monitoring |

### Docker Build Development

```bash
# Build development Docker image
docker build -f Dockerfile.dev -t radarr-go-dev .

# Build production Docker image
make docker-build

# Run with Docker Compose
make docker-run
```

### Volume Management

The development environment uses volumes for persistence:

- **Database data**: Persisted across restarts
- **Go module cache**: Speeds up builds
- **Frontend node_modules**: Faster npm installs
- **Monitoring data**: Persistent dashboards and metrics

## Troubleshooting

### Common Issues

#### Port Conflicts

```bash
# Check for conflicting services
lsof -i :7878
lsof -i :3000
lsof -i :5432

# Stop conflicting services or change ports in config
```

#### Database Connection Issues

```bash
# Check database status
make test-db-up
docker-compose -f docker-compose.dev.yml ps

# View database logs
docker-compose -f docker-compose.dev.yml logs postgres-dev
docker-compose -f docker-compose.dev.yml logs mariadb-dev
```

#### Hot Reload Not Working

```bash
# Ensure air is installed
go install github.com/cosmtrek/air@latest

# Check air configuration
cat .air.toml

# Restart development environment
make dev
```

#### Frontend Development Issues

```bash
# Ensure Node.js and npm are installed
node --version
npm --version

# Create frontend structure if missing
make setup-frontend

# Check if React is implemented
ls -la web/frontend/
```

#### Docker Issues

```bash
# Check Docker daemon
docker info

# Clean up development environment
make clean-all

# Rebuild containers
docker-compose -f docker-compose.dev.yml up --build --force-recreate
```

### Performance Issues

#### Slow Database Queries

- Check database logs in Adminer (http://localhost:8081)
- Use Prometheus metrics (http://localhost:9090)
- Enable query logging in development databases

#### High Memory Usage

- Use pprof for memory profiling: `go tool pprof http://localhost:8080/debug/pprof/heap`
- Monitor with Grafana dashboard (http://localhost:3001)

#### Build Performance

```bash
# Use build cache
export GOCACHE=/tmp/go-build-cache

# Parallel builds
make -j$(nproc) build-all

# Clean build cache if corrupted
go clean -cache
```

### Environment Reset

```bash
# Complete environment reset
make clean-all
docker-compose -f docker-compose.dev.yml down -v --remove-orphans
docker system prune -f
./scripts/dev-setup.sh
make dev-full
```

## Contributing Guidelines

### Code Quality Standards

#### Pre-commit Checks

Always run quality checks before committing:

```bash
# Automated (recommended)
pre-commit install
# Hooks will run automatically on git commit

# Manual
make lint           # Code quality checks
make test           # All tests
make all           # Complete quality workflow
```

#### Code Standards

- Follow Go conventions and idioms
- Use `gofmt` for formatting
- Maximum line length: 120 characters
- Maximum function length: 40 statements
- Comprehensive error handling
- Clear, descriptive variable names

#### Documentation Standards

- All exported functions must have documentation
- Package-level documentation required
- Update CLAUDE.md for significant changes
- Include examples in documentation

### Testing Requirements

#### Test Coverage

- Aim for >80% test coverage
- Unit tests for all business logic
- Integration tests for API endpoints
- Benchmark tests for performance-critical code

#### Test Organization

```bash
# Test structure
internal/
  api/
    handlers_test.go      # API endpoint tests
  services/
    movie_service_test.go # Business logic tests
  models/
    movie_test.go         # Model tests
```

### Git Workflow

#### Commit Message Format

Use conventional commits:

```
feat(api): add movie search endpoint
fix(database): resolve connection pool leak
docs(readme): update installation instructions
test(services): add movie service benchmarks
```

#### Branch Naming

- `feature/description`: New features
- `fix/description`: Bug fixes
- `docs/description`: Documentation updates
- `test/description`: Test improvements

### Pull Request Process

1. Create feature branch from `main`
2. Ensure all tests pass: `make ci`
3. Update documentation if needed
4. Create PR with clear description
5. Address review feedback
6. Merge after approval

### Development Best Practices

#### Local Development

- Use `make dev-full` for complete environment
- Test both PostgreSQL and MariaDB
- Run benchmarks for performance changes
- Use pre-commit hooks for quality

#### Code Organization

- Follow clean architecture principles
- Use dependency injection
- Implement interfaces for testability
- Keep functions focused and small

#### Performance Considerations

- Profile memory usage with pprof
- Monitor database query performance
- Use appropriate data structures
- Cache expensive operations

---

## Additional Resources

- **API Documentation**: See Radarr v3 API compatibility notes in CLAUDE.md
- **Architecture Guide**: Detailed architecture overview in CLAUDE.md
- **Versioning Guide**: VERSIONING.md for release management
- **Migration Guide**: MIGRATION.md for upgrade procedures

For questions or support, refer to the project documentation or create an issue in the repository.
