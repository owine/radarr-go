# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Radarr Go** is a complete rewrite of the Radarr movie collection manager from C#/.NET to Go. It maintains 100% API compatibility with Radarr's v3 API while providing significant performance improvements and simplified deployment.

**Current Status: v0.9.0-alpha** - Near production-ready with 95% feature parity to original Radarr.

**Versioning Strategy**: Follows [Semantic Versioning 2.0.0](https://semver.org/) with automated Docker tag management. See [VERSIONING.md](VERSIONING.md) and [MIGRATION.md](MIGRATION.md) for complete details.

## Key Features

### Core Functionality

- **Movie Library**: Complete CRUD operations, TMDB integration, metadata management
- **Search & Acquisition**: Multiple indexer support, download client integration, interactive search
- **File Management**: Automated organization, media info extraction, rename operations
- **Quality Management**: Quality profiles, custom formats, upgrade decisions
- **Task Scheduling**: Cron-based scheduling, distributed execution, monitoring
- **Health Monitoring**: System diagnostics, performance monitoring, automated alerts

### Notification System (21+ Providers)

Discord, Slack, Email, Telegram, Pushover, Webhooks, Plex/Emby/Jellyfin integration, and specialized providers for comprehensive alerting.

### Calendar & Import Management

- **Calendar System**: Multi-view calendar with iCal/CalDAV integration
- **Import Lists**: 20+ providers (IMDb, Trakt, TMDb, Letterboxd) with smart sync

### API & Integration

- **150+ API Endpoints**: Complete REST API with Radarr v3 compatibility
- **Authentication**: API key-based with header/query support
- **Database Support**: PostgreSQL (default) and MariaDB with optimizations

## Development Commands

### Quick Start

```bash
# Automated setup (macOS/Linux)
./scripts/dev-setup.sh       # Complete environment setup
make check-env               # Validate prerequisites
make setup                   # Install development tools
make dev-full               # Launch complete dev environment with Docker

# Basic development
make deps                    # Download Go modules
make build                   # Build for current platform
make dev                     # Run with hot reload (requires air)
make test                    # Run all tests
make lint                    # Run linter
make all                     # Format, lint, test, and build
```

### Multi-Platform Building

```bash
# Build all platforms
make build-all               # Recommended approach

# Individual platforms
make build-linux-amd64       # Linux x86_64
make build-linux-arm64       # Linux ARM64
make build-darwin-amd64      # macOS Intel
make build-darwin-arm64      # macOS Apple Silicon
make build-windows-amd64     # Windows x86_64
make build-freebsd-amd64     # FreeBSD x86_64
```

### Database Operations

```bash
make migrate-up              # Apply migrations
make migrate-down            # Rollback migrations
RADARR_DATABASE_TYPE=mariadb ./radarr    # Use MariaDB
RADARR_DATABASE_TYPE=postgres ./radarr   # Use PostgreSQL (default)
```

### Docker Development

```bash
make docker-build           # Build Docker image
make docker-run             # Start with docker-compose
make dev-full               # Complete development environment

# Development services available at:
# - Backend API: http://localhost:7878
# - Database Admin: http://localhost:8081
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3001 (admin/admin)
```

### Testing & Quality

```bash
make test                    # Run all tests
make test-coverage           # Generate coverage report
make test-bench              # Run benchmarks
make fmt                     # Format code
make lint                    # Run golangci-lint

# Database-specific testing
RADARR_DATABASE_TYPE=postgres go test -v ./...
RADARR_DATABASE_TYPE=mariadb go test -v ./...
```

## Architecture Overview

### Layered Architecture

1. **API Layer** (`internal/api/`): Gin-based HTTP server with Radarr v3 compatibility
2. **Service Layer** (`internal/services/`): Business logic with dependency injection
3. **Data Layer** (`internal/database/`, `internal/models/`): Multi-database support with GORM
4. **Configuration** (`internal/config/`): YAML configuration with environment overrides

### Key Services

- **MovieService**: CRUD operations, TMDB integration, duplicate detection
- **SearchService**: Multi-provider search with result aggregation
- **TaskService**: Distributed task execution with scheduling
- **HealthService**: System monitoring and diagnostics
- **NotificationService**: Multi-provider notification delivery

### Database Architecture

- **Primary**: PostgreSQL with native Go drivers (pgx)
- **Secondary**: MariaDB/MySQL support
- **ORM**: GORM with prepared statements and connection pooling
- **Migrations**: Version-controlled schema with rollback support

## Configuration System

Uses Viper for flexible configuration management:

### Environment Variables

Override any config with `RADARR_` prefix:

- `RADARR_SERVER_PORT=7878`
- `RADARR_DATABASE_TYPE=postgres` (or mariadb)
- `RADARR_DATABASE_HOST=localhost`
- `RADARR_LOG_LEVEL=debug`

### Key Sections

- **server**: HTTP settings (port, SSL, URL base)
- **database**: Connection, pooling, type
- **log**: Level, format, output
- **auth**: API key configuration
- **storage**: Data directories

## Adding New Features

### New API Endpoint

1. Add handler to `internal/api/handlers.go`
2. Register route in `internal/api/server.go:setupRoutes()`
3. Create service method in appropriate service
4. Add tests in `internal/api/*_test.go`

### New Database Model

1. Define struct in `internal/models/`
2. Add GORM annotations and JSON serialization
3. Create migration in `migrations/`
4. Add service methods for CRUD operations

### New Service

1. Create service struct in `internal/services/`
2. Add to `services.Container`
3. Initialize in `NewContainer()`
4. Follow dependency injection pattern

## Code Quality Standards

### Comprehensive Linting System

**CRITICAL**: Always run comprehensive linting before committing:

```bash
# Recommended: Install pre-commit hooks for automatic linting
pip install pre-commit && pre-commit install

# Install all linting tools
make setup-lint-tools

# Comprehensive linting (all file types)
make lint-all               # Run all linting checks
make lint-fix               # Auto-fix issues where possible

# Individual linting by file type
make lint-go                # Go code (golangci-lint)
make lint-frontend          # TypeScript/React (ESLint)
make lint-yaml              # YAML files (yamllint)
make lint-json              # JSON files (jsonlint/python)
make lint-markdown          # Markdown files (markdownlint)
make lint-shell             # Shell scripts (ShellCheck)

# Check linting tool installation
make check-lint-tools

# Legacy commands still supported
make lint                    # Alias for lint-go
make fmt                     # Format code
make test                    # Run tests
```

### Formatting Standards

- **Go Code**: Use `gofmt -s` (run `make fmt`)
- **Import Order**: Standard library → Third-party → Local
- **Naming**: camelCase variables, PascalCase types, UPPER_SNAKE_CASE constants
- **Line Length**: Maximum 120 characters
- **Documentation**: All exported functions/types must have comments

### Error Handling

```go
if err != nil {
    return fmt.Errorf("failed to fetch movie with id %d: %w", id, err)
}
```

### Enabled Linters

golangci-lint with: bodyclose, errcheck, gosec, govet, ineffassign, misspell, revive, staticcheck, unused, whitespace

## Database Standards

- **Migrations**: Sequential numbering (`001_initial_schema.up.sql`)
- **Tables**: snake_case plural (`movies`, `quality_profiles`)
- **Columns**: snake_case (`created_at`, `movie_id`)
- **Foreign Keys**: `_id` suffix

## API Compatibility

- **100% Radarr v3 Compatible**: All 150+ endpoints match original structure
- **Authentication**: X-API-Key header and query parameter support
- **Performance**: 3-5x faster response times vs original
- **Error Handling**: Consistent JSON responses with proper HTTP status codes

## Production Features

- **Performance**: 60-80% lower memory usage vs .NET version
- **Deployment**: Single binary, container-first, Kubernetes-ready
- **Monitoring**: Prometheus metrics, distributed tracing, health checks
- **Security**: CORS, input validation, rate limiting, container hardening

## Development Best Practices

### Workflow

- Use `./scripts/dev-setup.sh` for new machine setup
- Run `make check-env` to verify prerequisites
- Use `make dev-full` for complete development environment
- Install pre-commit hooks for automated quality checks
- Test both PostgreSQL and MariaDB during development
- Maintain Radarr v3 API compatibility

### Repository Organization

- Keep codebase clean and organized
- Remove test artifacts after completion
- Update documentation with code changes
- Use descriptive variable/function names
- Follow Go naming conventions

### Testing

- Aim for >80% test coverage
- Use temporary directories for test data
- Clean up test resources in teardown
- Test across multiple platforms and databases

## Documentation Maintenance

**CRITICAL**: Update documentation when making code changes:

- Update CLAUDE.md for development workflow changes
- Update README.md for feature/installation changes
- Update VERSIONING.md/MIGRATION.md for version-related changes
- Keep inline documentation current
- Include documentation updates in the same commit as code changes
