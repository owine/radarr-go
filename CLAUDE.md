# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is **Radarr Go**, a complete rewrite of the Radarr movie collection manager from C#/.NET to Go. It maintains 100% API compatibility with Radarr's v3 API while providing significant performance improvements and simplified deployment.

## Development Commands

### Core Development
```bash
# Install dependencies and setup
make deps                    # Download Go modules
make setup                   # Install dev tools (air, golangci-lint, migrate)

# Building
make build                   # Build binary for current platform
make build-linux            # Build for Linux (production)

# Multi-platform building (matches CI)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-linux-amd64 ./cmd/radarr
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-linux-arm64 ./cmd/radarr
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-darwin-amd64 ./cmd/radarr
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-darwin-arm64 ./cmd/radarr
GOOS=freebsd GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-freebsd-amd64 ./cmd/radarr
GOOS=freebsd GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-w -s" -o radarr-freebsd-arm64 ./cmd/radarr

# Running
make run                     # Build and run locally
make dev                     # Run with hot reload (requires air)
./radarr --data ./data       # Run with custom data directory

# Testing
make test                    # Run all tests
make test-coverage           # Run tests with HTML coverage report
go test ./internal/api -v    # Run specific package tests
go test -run TestPingHandler ./internal/api  # Run single test

# Database-specific testing (matches CI matrix)
RADARR_DATABASE_TYPE=postgres go test -v ./...   # Requires PostgreSQL server
RADARR_DATABASE_TYPE=mariadb go test -v ./...    # Requires MariaDB server

# Code Quality
make fmt                     # Format code
make lint                    # Run linter (requires golangci-lint)
make all                     # Format, lint, test, and build
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
make docker-build           # Build Docker image
make docker-run             # Start with docker-compose
make docker-stop            # Stop docker-compose
make docker-logs            # View container logs
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
All services are managed through a `services.Container` that provides dependency injection:
- `MovieService`: Movie CRUD operations and search
- `MovieFileService`: File management and metadata
- `QualityService`, `IndexerService`, etc.: Domain-specific operations

### Database Architecture
- **Dual ORM Strategy**: GORM for complex operations, sqlx for performance-critical queries
- **Migration System**: golang-migrate for schema management
- **Multi-Database**: PostgreSQL (default) and MariaDB support
- **Connection Management**: Configurable connection pooling
- **Pure Go Strategy**: PostgreSQL and MariaDB with native Go drivers (CGO_ENABLED=0)

### CI/CD Architecture
The project uses a structured CI pipeline with concurrent execution:

**Stage 1**: Concurrent quality checks (lint + security)
**Stage 2**: Multi-platform build (after quality checks pass)
**Stage 3**: Matrix testing (PostgreSQL + MariaDB on amd64/arm64)
  - **Linux**: Service containers for both databases
  - **macOS/FreeBSD**: Native installation testing
**Stage 4**: Publish (Docker images + artifacts after all tests pass)

Supported platforms: Linux, Darwin, FreeBSD on amd64/arm64 architectures.

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

## Code Quality and Formatting Standards

### Required Quality Checks
**CRITICAL**: Before committing any code changes, ALWAYS run the linting tests to ensure code quality:

- Run `make lint` to verify all code passes golangci-lint checks
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
- Update inline code comments for significant logic changes
- Ensure all documentation remains accurate and current
- Documentation updates should be included in the same commit as the related code changes

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
- Run `make lint` before committing to ensure code quality
- Use `make all` for comprehensive quality checks (format, lint, test, build)
- Test both database backends (PostgreSQL and MariaDB) during development
- Maintain backwards compatibility with Radarr v3 API at all times