# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **Radarr Go**, a complete rewrite of the Radarr movie collection manager from C#/.NET to Go. It maintains 100% API compatibility with Radarr's v3 API while providing significant performance improvements and simplified deployment.

**Current Status: v0.9.0-alpha** - Near production-ready with 95% feature parity to original Radarr.

**Versioning Strategy**: Follows [Semantic Versioning 2.0.0](https://semver.org/) with automated Docker tag management. See [VERSIONING.md](VERSIONING.md) and [MIGRATION.md](MIGRATION.md) for complete details.

## Comprehensive Feature Set

### Core Movie Management
- **Movie Library**: Complete CRUD operations for movie management
- **Movie Discovery**: TMDB integration for popular/trending movie discovery
- **Movie Metadata**: Automatic metadata refresh and TMDB lookup
- **Movie Files**: File management with media info extraction and organization
- **Movie Collections**: Support for managing movie collections with TMDB sync
- **Quality Management**: Quality profiles, definitions, and custom formats
- **Root Folder Management**: Multi-folder support with statistics

### Advanced Search and Acquisition
- **Indexer Management**: Support for multiple search providers with testing capabilities
- **Release Management**: Release searching, filtering, and statistics
- **Interactive Search**: Manual search with release selection
- **Download Client Integration**: Support for multiple download clients with statistics
- **Queue Management**: Download queue monitoring and management
- **Wanted Movies Tracking**: Missing and cutoff unmet movie management with priority system

### File Organization and Management
- **File Organization System**: Automated file processing and organization
- **Manual Import Processing**: Manual import with override capabilities
- **Rename Operations**: File and folder renaming with preview functionality
- **Media Info Extraction**: Automatic media information detection
- **File Operation Tracking**: Comprehensive file operation monitoring
- **Parse Service**: Release name parsing with caching

### Notification System (11 Providers)
- **Discord**: Rich embed notifications
- **Slack**: Channel-based notifications
- **Email**: SMTP email notifications
- **Webhook**: Custom HTTP webhook integration
- **Pushover**: Mobile push notifications
- **Telegram**: Bot-based messaging
- **Pushbullet**: Cross-device notifications
- **Gotify**: Self-hosted push notifications
- **Mailgun**: Transactional email service
- **SendGrid**: Cloud email delivery
- **Custom Script**: Custom notification scripts

### Task Scheduling and Automation
- **Task Management**: Complete task scheduling system with status tracking
- **System Commands**: Health checks, cleanup operations
- **Movie Commands**: Refresh operations for individual or all movies
- **Import List Sync**: Automated list synchronization
- **Scheduled Tasks**: Background task execution with monitoring
- **Queue Status**: Real-time task queue monitoring

### Health Monitoring and Diagnostics
- **Health Dashboard**: Comprehensive system health overview
- **Health Issue Management**: Issue tracking, dismissal, and resolution
- **System Resource Monitoring**: CPU, memory, and disk space tracking
- **Performance Metrics**: Performance monitoring with time-based metrics
- **Health Checkers**: Multiple built-in health verification systems
- **Disk Space Monitoring**: Configurable thresholds and alerts

### Calendar and Event Tracking
- **Calendar Events**: Movie release date tracking with filtering
- **iCal Feed**: RFC 5545 compliant calendar feeds for external applications
- **Calendar Configuration**: Customizable calendar settings
- **Calendar Statistics**: Event statistics and metrics
- **Feed URL Generation**: Shareable calendar feed URLs

### Import and List Management
- **Import Lists**: Multiple import list provider support
- **List Synchronization**: Automated and manual list sync operations
- **Import List Statistics**: Provider performance metrics
- **Import List Movies**: Dedicated import candidate management
- **Bulk Operations**: Mass operations on import lists

### Configuration and Settings
- **Host Configuration**: Server and application settings management
- **Naming Configuration**: File naming patterns with token support
- **Media Management**: File handling and organization settings
- **Configuration Statistics**: System configuration metrics
- **Environment Integration**: Comprehensive environment variable support

### API and Integration
- **150+ API Endpoints**: Complete REST API with Radarr v3 compatibility
- **Authentication**: API key-based authentication with header/query support
- **CORS Support**: Configurable cross-origin resource sharing
- **Activity Tracking**: API activity logging and monitoring
- **History Management**: Comprehensive history tracking and statistics

### Database and Performance
- **Multi-Database Support**: PostgreSQL (default) and MariaDB with optimizations
- **GORM Integration**: Advanced ORM with prepared statements and transactions
- **Migration System**: Database schema management with rollback support
- **Performance Benchmarks**: Automated benchmark testing for regression monitoring
- **Connection Pooling**: Configurable database connection management

## Development Commands

### Quick Start (New Developers)
```bash
# Automated setup script (macOS/Linux)
./scripts/dev-setup.sh       # Complete environment setup with tool installation

# Manual setup
make check-env               # Check development environment prerequisites
make setup                   # Install dev tools (air, golangci-lint, migrate)
make dev-full               # Start complete development environment (Docker)
```

### Core Development
```bash
# Install dependencies and setup
make deps                    # Download Go modules
make setup                   # Install both backend and frontend development tools
make setup-backend           # Install only backend tools (air, golangci-lint, migrate)
make setup-frontend          # Prepare frontend structure for React (Phase 2)

# Pre-commit hooks setup (recommended for development)
pip install pre-commit       # Install pre-commit (requires Python)
pre-commit install          # Install git hooks
pre-commit run --all-files  # Run hooks on all files (initial setup)

# Development Environment Options
make dev                     # Backend with hot reload (local)
make dev-full               # Complete environment: backend + databases + monitoring (Docker)
make dev-frontend           # Frontend development server (Phase 2)

# Building
make build                   # Build binary for current platform
make build-linux            # Build for Linux (production)
make build-frontend         # Build React frontend for production (Phase 2)
make build-all              # Build all platforms (backend only)
make build-all-with-frontend # Build all platforms + frontend (Phase 2)

# Multi-platform building (matches CI and Makefile)
make build-all               # Build for all platforms (recommended)

# Individual platform builds
make build-linux-amd64       # Linux x86_64
make build-linux-arm64       # Linux ARM64
make build-darwin-amd64      # macOS Intel
make build-darwin-arm64      # macOS Apple Silicon
make build-windows-amd64     # Windows x86_64
make build-windows-arm64     # Windows ARM64
make build-freebsd-amd64     # FreeBSD x86_64
make build-freebsd-arm64     # FreeBSD ARM64

# Manual builds with version info (matches CI exactly)
export VERSION=v1.0.0
export COMMIT=$(git rev-parse --short HEAD)
export BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS="-w -s -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}' -X 'main.date=${BUILD_DATE}'"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-linux-amd64 ./cmd/radarr
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-linux-arm64 ./cmd/radarr
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-darwin-amd64 ./cmd/radarr
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-darwin-arm64 ./cmd/radarr
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-windows-amd64.exe ./cmd/radarr
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-windows-arm64.exe ./cmd/radarr
GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-freebsd-amd64 ./cmd/radarr
GOOS=freebsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$LDFLAGS" -o radarr-freebsd-arm64 ./cmd/radarr

# Running
make run                     # Build and run locally
make dev                     # Run with hot reload (requires air)
./radarr --data ./data       # Run with custom data directory

# Testing
make test                    # Run all tests
make test-coverage           # Run tests with HTML coverage report
make test-bench              # Run benchmark tests for performance monitoring
make test-examples           # Run example tests for documentation validation
go test ./internal/api -v    # Run specific package tests
go test -run TestPingHandler ./internal/api  # Run single test

# Database-specific testing (matches CI matrix)
RADARR_DATABASE_TYPE=postgres go test -v ./...   # Requires PostgreSQL server
RADARR_DATABASE_TYPE=mariadb go test -v ./...    # Requires MariaDB server

# Code Quality
make fmt                     # Format code
make lint                    # Run linter (requires golangci-lint)
make all                     # Format, lint, test, and build
make dev-all                 # Comprehensive development workflow
```

### Database Operations
```bash
# Migrations
make migrate-up              # Apply database migrations
make migrate-down            # Rollback migrations
migrate create -ext sql -dir migrations migration_name  # Create new migration

# Database switching
RADARR_DATABASE_TYPE=mariadb ./radarr     # Use MariaDB
RADARR_DATABASE_TYPE=postgres ./radarr    # Use PostgreSQL (default)
```

### Docker Operations
```bash
# Production Docker
make docker-build           # Build Docker image
make docker-run             # Start with docker-compose
make docker-stop            # Stop docker-compose
make docker-logs            # View container logs

# Development Docker Environment
make dev-full               # Start complete development environment
docker-compose -f docker-compose.dev.yml up -d  # Start all development services
docker-compose -f docker-compose.dev.yml --profile monitoring up -d  # Include monitoring
docker-compose -f docker-compose.dev.yml --profile mariadb up -d     # Use MariaDB instead
docker-compose -f docker-compose.dev.yml --profile frontend up -d    # Include frontend (Phase 2)

# Development Services (when running make dev-full)
# - Backend API: http://localhost:7878 (with hot reload)
# - Database Admin (Adminer): http://localhost:8081
# - Prometheus Metrics: http://localhost:9090
# - Grafana Dashboard: http://localhost:3001 (admin/admin)
# - Jaeger Tracing: http://localhost:16686
# - MailHog (Email Testing): http://localhost:8025
# - Frontend Dev Server: http://localhost:3000 (Phase 2)
```

### Versioning and Release Management
```bash
# Version Analysis and Validation
./.github/scripts/version-analyzer.sh v1.2.3 --env     # Analyze version and generate environment variables
./.github/scripts/version-analyzer.sh v1.2.3 --json   # Get JSON analysis output
./.github/scripts/version-analyzer.sh v1.2.3 --docker-tags ghcr.io/radarr/radarr-go  # Generate Docker tags

# Version Progression Validation
./.github/scripts/validate-version-progression.sh v1.2.3  # Validate version against Git history

# Build System Testing
./.github/scripts/validate-build-version.sh         # Test version injection system
./.github/scripts/validate-build-version.sh ./radarr  # Test specific binary

# Release Notes Generation
./.github/scripts/generate-release-notes.sh         # Generate comprehensive release notes

# Complete System Testing
./.github/scripts/test-versioning-system.sh         # Test all versioning components

# Version Information
./radarr --version              # Check application version information

# Release Process (Automated via GitHub Actions)
git tag v1.2.3                 # Create release tag (triggers automated workflow)
git push origin v1.2.3         # Push tag to trigger CI/CD pipeline

# Pre-release Process
git tag v1.2.3-alpha.1         # Create pre-release tag
git push origin v1.2.3-alpha.1 # Push pre-release tag
```

### Frontend Development (Phase 2 Preparation)
```bash
# Frontend structure setup
make setup-frontend         # Create React project structure
make install-frontend       # Install npm dependencies (when package.json exists)
make build-frontend         # Build React frontend for production
make dev-frontend           # Start React development server
make clean-frontend         # Clean frontend build artifacts

# Full-stack development workflow
make dev-full               # Start backend + databases + monitoring in Docker
# In another terminal:
make dev-frontend           # Start frontend development server (when React is ready)
```

## Architecture Overview

### Layered Architecture
The application follows a clean layered architecture with dependency injection:

1. **API Layer** (`internal/api/`): Gin-based HTTP server with middleware
2. **Service Layer** (`internal/services/`): Business logic and domain operations
3. **Data Layer** (`internal/database/`, `internal/models/`): Database access and models
4. **Configuration** (`internal/config/`): YAML + environment variable management
5. **Infrastructure** (`internal/logger/`): Logging, utilities

### Dependency Flow
```
main() → config → logger → database → services → api server
```

### Service Container Pattern
All services are managed through a `services.Container` that provides comprehensive dependency injection:

#### Core Business Services
- `MovieService`: Movie CRUD operations, search, discovery, and metadata management
- `MovieFileService`: File management, metadata extraction, and organization
- `QualityService`: Quality profiles, definitions, and custom format management
- `IndexerService`: Search provider management and configuration
- `ImportListService`: Import list management and synchronization
- `DownloadService`: Download client integration and queue management
- `SearchService`: Release search and interactive search capabilities

#### Task and Workflow Services
- `TaskService`: Task scheduling, execution, and status tracking
- `TaskHandlerService`: Specialized task handlers for different operations
- `FileOrganizationService`: Automated file processing and organization
- `RenameService`: File and folder renaming operations
- `ImportService`: Manual and automatic import processing
- `ParseService`: Release name parsing with intelligent caching

#### Monitoring and Health Services
- `HealthService`: Comprehensive health monitoring and issue management
- `HealthIssueService`: Health issue tracking, resolution, and notifications
- `PerformanceMonitor`: System performance metrics and monitoring
- `CalendarService`: Movie release date tracking and calendar event generation
- `ICalService`: RFC 5545 compliant iCal feed generation for external calendar integration

#### Notification and Communication
- `NotificationService`: Multi-provider notification system with 11+ providers
- `NotificationFactory`: Provider instantiation and management
- `TemplateService`: Notification template processing and customization

#### Specialized Services
- `WantedService`: Missing and cutoff unmet movie management
- `CollectionService`: Movie collection management with TMDB integration
- `MediaInfoService`: Media file information extraction and analysis
- `ConfigService`: Dynamic configuration management and validation
- `HistoryService`: Activity tracking and historical data management
- `MetadataService`: Movie metadata management and TMDB integration

### Database Architecture
- **GORM Optimized**: Enhanced with prepared statements, transactions, and validation hooks
- **Performance Features**: Index hints, connection pooling, and optimized query patterns
- **Hybrid Strategy**: GORM for complex operations, sqlc for performance-critical queries
- **Migration System**: golang-migrate for schema management
- **Multi-Database**: PostgreSQL (default) and MariaDB support with database-specific optimizations
- **Connection Management**: Configurable connection pooling with pgx for PostgreSQL
- **Pure Go Strategy**: PostgreSQL (pgx driver) and MariaDB with native Go drivers (CGO_ENABLED=0)
- **Data Integrity**: GORM validation hooks with business logic enforcement

### CI/CD Architecture
The project uses a structured CI pipeline with concurrent execution:

**Stage 1**: Concurrent quality checks (lint + security + workspace validation)
**Stage 2**: Multi-platform build (after quality checks pass) including Windows
**Stage 3**: Comprehensive matrix testing (PostgreSQL + MariaDB on amd64/arm64)
  - **Test Types**: Unit tests, benchmark tests, example tests
  - **Linux**: Service containers for both databases
  - **macOS/FreeBSD**: Native installation testing
  - **Performance Monitoring**: Automated benchmark execution
**Stage 4**: Publish (Docker images + artifacts after all tests pass)

Supported platforms: Linux, Darwin, Windows, FreeBSD on amd64/arm64 architectures.

### Versioning and Release Architecture

The project implements a comprehensive automated versioning system that eliminates manual processes:

#### Automated Versioning Workflow
- **Version Analysis**: Automatic semantic version validation and Docker tag generation via `.github/scripts/version-analyzer.sh`
- **Progression Validation**: Git history-based version progression validation via `.github/scripts/validate-version-progression.sh`
- **Build Integration**: Automated version injection during build process with validation via `.github/scripts/validate-build-version.sh`
- **Release Documentation**: Automated comprehensive release notes with Docker information via `.github/scripts/generate-release-notes.sh`

#### Docker Tag Strategy (Automated)
**Current Phase (Pre-1.0)**:
- `:testing`, `:prerelease` - Latest pre-release versions
- `:v0.9.0-alpha` - Specific version pinning (recommended)
- `:alpha`, `:beta`, `:rc` - Prerelease type tags
- Database-specific tags: `:v0.9.0-alpha-postgres`, `:v0.9.0-alpha-mariadb`

**Future Phase (v1.0.0+)**:
- `:latest` - Production releases (assigned at v1.0.0)
- `:stable` - Stable release pointer
- `:2025.04` - Calendar-based versioning

#### Version Command Support
```bash
# Check application version
./radarr --version
# Output: Radarr Go v0.9.0-alpha (commit: abc123, built: 2025-01-01_12:00:00)
```

#### Release Process Integration
- **Tag Creation**: `git tag v1.2.3` triggers full automated workflow
- **Validation**: Version format, progression rules, build system validation
- **Publication**: Multi-platform builds, Docker images, comprehensive documentation
- **Quality Assurance**: Integration tests, performance benchmarks, security scanning

### Go Workspace and Modern Practices
The project now includes Go 1.24+ workspace support and follows modern Go best practices:

- **Go Workspace**: `go.work` file for multi-module development support
- **Benchmark Testing**: Performance regression monitoring with `make test-bench`
- **Example Testing**: Documentation validation with `make test-examples`
- **Package Documentation**: Comprehensive `doc.go` files with usage examples
- **Multi-Platform Builds**: Full cross-compilation support including Windows
- **Pre-commit Hooks**: Automated quality checks with formatting, linting, and tests
- **GORM Best Practices**: Prepared statements, transactions, validation hooks, and index hints
- **Security Scanning**: Continuous vulnerability monitoring with govulncheck

## Key Patterns and Conventions

### API Design
- **Radarr v3 Compatibility**: All endpoints match original Radarr API structure
- **RESTful Routes**: Standard HTTP verbs with resource-based URLs
- **Middleware Stack**: Logging → CORS → API Key Authentication → Routes
- **Error Handling**: Consistent JSON error responses with proper HTTP status codes

### Configuration Management
- **YAML Primary**: `config.yaml` with nested structure
- **Environment Override**: `RADARR_` prefixed variables automatically override YAML
- **Validation**: Configuration validation at startup with helpful error messages
- **Defaults**: Sensible defaults for development and production

### Database Models
- **GORM Annotations**: Struct tags for database mapping and validation
- **JSON Serialization**: Custom Value/Scan methods for complex types (arrays, JSON fields)
- **Relationships**: Foreign key relationships with preloading support
- **Timestamps**: Automatic created_at/updated_at fields

### Testing Strategy
- **Unit Tests**: Service layer and individual components
- **API Tests**: HTTP endpoint testing with test server
- **Matrix Testing**: Comprehensive testing across multiple platforms (Linux, macOS, FreeBSD) and architectures (amd64, arm64)
- **Database Testing**: Both PostgreSQL and MariaDB testing on all supported platforms
- **Test Mode**: Gin test mode for reduced noise in tests
- **Mocking**: Interface-based dependency injection enables easy mocking

## Configuration System

The configuration system uses Viper for flexible config management:

### Key Configuration Sections
- **server**: HTTP server settings (port, SSL, URL base)
- **database**: Database type, connection, pooling
- **log**: Logging level, format, output
- **auth**: Authentication method and API key
- **storage**: Data directories and paths

### Environment Variables
All config keys can be overridden with `RADARR_` prefix:
- `RADARR_SERVER_PORT=7878`
- `RADARR_DATABASE_TYPE=postgres` (or mariadb)
- `RADARR_DATABASE_HOST=localhost`
- `RADARR_DATABASE_PORT=5432` (or 3306 for mariadb)
- `RADARR_DATABASE_USERNAME=radarr`
- `RADARR_DATABASE_PASSWORD=password`
- `RADARR_LOG_LEVEL=debug`

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
5. Add API endpoints if needed

### New Service
1. Create service struct in `internal/services/`
2. Add to `services.Container`
3. Initialize in `NewContainer()`
4. Follow dependency injection pattern

## Database Schema

### Core Tables
- **movies**: Main movie entities with metadata
- **movie_files**: Physical file information and media info
- **quality_profiles**: Quality settings and cutoff definitions
- **indexers**: Search provider configurations
- **download_clients**: Download automation settings
- **notifications**: Alert and notification configurations

### Migration Strategy
- **Sequential Numbering**: `001_initial_schema.up.sql`
- **Rollback Support**: Corresponding `.down.sql` files
- **Auto-Migration**: Runs automatically on startup
- **Schema Evolution**: Additive changes preferred

## API Compatibility

This implementation maintains strict compatibility with Radarr's v3 API:
- **Same Endpoints**: Identical URL patterns and HTTP methods
- **Same Responses**: JSON structure matches exactly
- **Same Behavior**: Pagination, filtering, sorting work identically
- **Authentication**: X-API-Key header and query parameter support

## Production Considerations

### Performance
- **Go Runtime**: Significantly lower memory usage than .NET
- **Gin Framework**: High-performance HTTP routing
- **Connection Pooling**: Configurable database connections
- **Structured Logging**: JSON logging with minimal overhead

### Deployment
- **Single Binary**: No runtime dependencies except database
- **Docker Ready**: Multi-stage builds with minimal Alpine base
- **Health Checks**: `/ping` endpoint for monitoring
- **Graceful Shutdown**: Proper signal handling and cleanup

### Security
- **API Key Auth**: Optional API key authentication
- **CORS**: Configurable cross-origin resource sharing
- **Input Validation**: Request validation and sanitization
- **No Root**: Docker container runs as non-root user

# Code Generation with sqlc

The project uses sqlc for generating type-safe Go code from SQL queries.

```bash
# Generate sqlc code (after modifying SQL queries)
sqlc generate

# Add new SQL queries
mkdir -p internal/database/queries/postgres internal/database/queries/mysql
# Add .sql files with query definitions
# Run sqlc generate to create Go code
```

### Query Organization
- **PostgreSQL queries**: `internal/database/queries/postgres/*.sql`
- **MySQL queries**: `internal/database/queries/mysql/*.sql`
- **Generated code**: `internal/database/generated/{postgres,mysql}/`

### Usage in Services
```go
// PostgreSQL
result, err := db.Postgres.GetMovieByID(ctx, movieID)

// MySQL
result, err := db.MySQL.GetMovieByID(ctx, movieID)
```

## Code Quality and Formatting Standards

### Required Quality Checks
**CRITICAL**: Before committing any code changes, ALWAYS run the linting tests to ensure code quality:

- **Pre-commit hooks (Recommended)**: Install pre-commit hooks to automatically run quality checks before each commit
  - Install: `pip install pre-commit && pre-commit install`
  - Hooks will automatically run: `go fmt`, `go mod tidy`, `golangci-lint`, `go test`, and security checks
- **Manual checks**: Run `make lint` to verify all code passes golangci-lint checks
- Fix any linting errors before committing
- Ensure code follows Go conventions and project standards
- All commits must pass CI pipeline quality checks

### Code Formatting Standards

#### Go Code Formatting
- **Use `gofmt`**: All Go code must be formatted with `gofmt -s` (run `make fmt`)
- **Import Organization**: Group imports in this order:
  1. Standard library packages
  2. Third-party packages
  3. Local project packages
- **Line Length**: Maximum 120 characters per line (enforced by `lll` linter)
- **Function Length**: Maximum 40 statements per function (enforced by `funlen` linter)

#### Naming Conventions
- **Variables**: Use camelCase (`userID`, `movieFile`)
- **Constants**: Use UPPER_SNAKE_CASE (`MAX_RETRIES`, `DEFAULT_PORT`)
- **Functions/Methods**: Use camelCase (`GetMovie`, `CreateIndexer`)
- **Types**: Use PascalCase (`MovieService`, `QualityProfile`)
- **Interfaces**: Use PascalCase with descriptive names (`MovieRepository`, `Logger`)
- **Packages**: Use lowercase, single words when possible (`api`, `models`, `services`)

#### Documentation Standards
- **Public Functions**: All exported functions must have documentation comments
- **Public Types**: All exported types must have documentation comments
- **Package Documentation**: Each package must have a package-level comment
- **Comment Format**: Use complete sentences, start with the function/type name
```go
// GetMovie retrieves a movie by its ID from the database.
func GetMovie(id int) (*Movie, error) {
    // Implementation
}
```

#### Error Handling
- **Error Wrapping**: Use `fmt.Errorf` with `%w` verb for error context
- **Error Messages**: Use lowercase, no punctuation at end
- **Error Variables**: Use `Err` prefix for error variables (`ErrNotFound`)
```go
if err != nil {
    return fmt.Errorf("failed to fetch movie with id %d: %w", id, err)
}
```

#### Code Structure
- **Single Responsibility**: Each function should do one thing
- **Small Functions**: Keep functions focused and testable
- **Clear Variable Names**: Use descriptive names over comments
- **Early Returns**: Use guard clauses to reduce nesting
```go
func ProcessMovie(movie *Movie) error {
    if movie == nil {
        return ErrInvalidMovie
    }

    if movie.ID == 0 {
        return ErrMissingID
    }

    // Process movie
    return nil
}
```

### Enabled Linters
The project uses golangci-lint with these enabled linters:
- **bodyclose**: Checks HTTP response body is closed
- **errcheck**: Checks unchecked errors
- **gosec**: Security analysis
- **govet**: Examines Go source code and reports suspicious constructs
- **ineffassign**: Detects ineffectual assignments
- **misspell**: Finds commonly misspelled English words
- **revive**: Replacement for golint with more rules
- **staticcheck**: Advanced Go linter
- **unused**: Checks for unused constants, variables, functions, and types
- **whitespace**: Checks for unnecessary whitespace

### SQL and Migration Standards
- **File Naming**: Use sequential numbering (`001_initial_schema.up.sql`)
- **Reversible Changes**: Always provide corresponding `.down.sql` files
- **Column Naming**: Use snake_case (`created_at`, `movie_id`)
- **Table Naming**: Use snake_case plural (`movies`, `quality_profiles`)
- **Foreign Keys**: Use `_id` suffix (`movie_id`, `quality_profile_id`)

### JSON and API Standards
- **JSON Tags**: Use camelCase for JSON field names
- **Struct Tags**: Include both `json` and `gorm` tags
```go
type Movie struct {
    ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
    Title       string    `json:"title" gorm:"not null;size:255"`
    ReleaseDate time.Time `json:"releaseDate" gorm:"not null"`
}
```

### Testing Standards
- **Test File Naming**: Use `_test.go` suffix
- **Test Function Naming**: Use `TestFunctionName` format
- **Test Coverage**: Aim for >80% test coverage
- **Test Organization**: Group related tests in subtests using `t.Run()`
- **Mock Usage**: Use interfaces for dependency injection and mocking

### Git Commit Standards
- **Commit Message Format**: Use conventional commits (`feat:`, `fix:`, `docs:`)
- **Scope**: Include scope when relevant (`feat(api):`, `fix(database):`)
- **Line Length**: Keep first line under 72 characters
- **Body**: Include detailed explanation for complex changes

## Documentation Maintenance

**CRITICAL**: When making changes to the codebase, ALWAYS update documentation to reflect those changes:

- Update CLAUDE.md with new development commands, architecture changes, or workflow modifications
- Update README.md with new features, installation steps, or usage instructions
- Update versioning documentation (VERSIONING.md, MIGRATION.md) for version-related changes
- Update inline code comments for significant logic changes
- Ensure all documentation remains accurate and current
- Documentation updates should be included in the same commit as the related code changes
- **Versioning Impact**: Consider if changes affect versioning strategy, Docker tags, or migration paths

## Development Best Practices

### Variable and Function Naming
- Use descriptive, self-documenting names for all variables and functions
- Prefer longer, clear names over abbreviations (e.g., `connectionString` vs `connStr`)
- Follow Go naming conventions: camelCase for unexported, PascalCase for exported
- Use constants for repeated string values to improve maintainability

### Test Management
- Always clean up test files and artifacts after test completion
- Use temporary directories for test data that get removed automatically
- Ensure tests don't leave persistent state that affects other tests
- Remove test databases and connections properly in teardown

### Repository Organization
- Keep the repository clean, organized, and human-readable
- Remove unused files, deprecated code, and temporary artifacts
- Maintain consistent file structure and naming conventions
- Update documentation immediately when making code changes

### Development Workflow
- **Quick Setup**: Use `./scripts/dev-setup.sh` for automated environment setup on new machines
- **Environment Check**: Run `make check-env` to verify all required tools are installed
- **Full Development**: Use `make dev-full` for complete development environment with monitoring
- **Frontend Preparation**: Use `make setup-frontend` to prepare structure for React (Phase 2)
- **Use pre-commit hooks**: Install and use pre-commit hooks for automatic quality checks (`pre-commit install`)
- **Alternative manual checks**: Run `make lint` before committing to ensure code quality
- Use `make all` for comprehensive quality checks (format, lint, test, build)
- Test both database backends (PostgreSQL and MariaDB) during development
- Maintain backwards compatibility with Radarr v3 API at all times
- **Documentation**: See `DEVELOPMENT.md` for comprehensive setup and troubleshooting guide
