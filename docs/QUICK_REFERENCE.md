# Radarr-Go Quick Reference Guide

## Development Commands

### Setup and Dependencies
```bash
make deps                    # Download Go modules
make setup                   # Install dev tools (air, golangci-lint, migrate)
pip install pre-commit       # Install pre-commit hooks
pre-commit install          # Setup git hooks
```

### Building
```bash
make build                   # Build for current platform
make build-all               # Build for all platforms
make build-linux            # Build for Linux
make build-darwin-amd64     # Build for macOS Intel
make build-darwin-arm64     # Build for macOS Apple Silicon
```

### Development Workflow
```bash
make dev                     # Run with hot reload
make run                     # Build and run
make all                     # Format, lint, test, build
make fmt                     # Format code
make lint                    # Run linter
```

### Testing
```bash
make test                    # Run all tests
make test-coverage          # Run tests with HTML coverage
make test-bench             # Run benchmark tests
make test-examples          # Run example tests

# Database-specific testing
RADARR_DATABASE_TYPE=postgres go test ./...
RADARR_DATABASE_TYPE=mariadb go test ./...
```

### Database Operations
```bash
make migrate-up             # Apply migrations
make migrate-down           # Rollback migrations
migrate create -ext sql -dir migrations/postgres migration_name
```

## Architecture Quick Reference

### Service Container Pattern
```go
// All services managed through dependency injection
type Container struct {
    DB     *database.Database
    Config *config.Config
    Logger *logger.Logger

    // Domain services
    MovieService        *MovieService
    TaskService         *TaskService
    HealthService       *HealthService
    // ... more services
}
```

### Database Access Patterns
```go
// GORM for complex operations
s.db.GORM.Where("title ILIKE ?", "%"+title+"%").Find(&movies)

// SQLC for performance-critical queries
s.db.Postgres.GetMovieByTMDBID(ctx, tmdbID)
s.db.MySQL.GetMovieByTMDBID(ctx, tmdbID)
```

### Task Handler Pattern
```go
type TaskHandler interface {
    Execute(ctx context.Context, task *models.TaskV2, updateProgress func(percent int, message string)) error
    GetName() string
    GetDescription() string
}
```

## Common Code Patterns

### Service Constructor
```go
func NewMovieService(db *database.Database, logger *logger.Logger) *MovieService {
    return &MovieService{
        db:     db,
        logger: logger,
    }
}
```

### API Handler Pattern
```go
func (s *Server) handleGetMovies(c *gin.Context) {
    // Parse parameters
    limit := c.DefaultQuery("limit", "50")

    // Validate input
    if limit == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
        return
    }

    // Call service
    movies, err := s.services.MovieService.GetMovies(c.Request.Context(), limit)
    if err != nil {
        s.logger.Errorw("Failed to get movies", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // Return response
    c.JSON(http.StatusOK, movies)
}
```

### Error Handling Pattern
```go
if err != nil {
    return fmt.Errorf("failed to process movie with id %d: %w", id, err)
}
```

## Configuration

### Environment Variables
```bash
RADARR_SERVER_PORT=7878
RADARR_DATABASE_TYPE=postgres       # or mariadb
RADARR_DATABASE_HOST=localhost
RADARR_DATABASE_PORT=5432          # or 3306 for mariadb
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=password
RADARR_LOG_LEVEL=debug
RADARR_AUTH_API_KEY=your_api_key
```

### Config Structure
```yaml
server:
  port: 7878
  host: "0.0.0.0"

database:
  type: postgres
  host: localhost
  port: 5432
  username: radarr
  password: password
  max_connections: 10

log:
  level: info
  format: json
  output: stdout

auth:
  api_key: "your_secure_api_key"

health:
  enabled: true
  interval: "15m"
  disk_space_warning_threshold: 5368709120  # 5GB
```

## Testing Patterns

### Unit Test Setup
```go
func TestMovieService_GetMovie(t *testing.T) {
    // Setup
    db := setupTestDatabase(t)
    service := NewMovieService(db, setupTestLogger())

    // Create test data
    movie := createTestMovie()
    err := db.GORM.Create(movie).Error
    require.NoError(t, err)

    // Execute
    result, err := service.GetMovie(context.Background(), movie.ID)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, movie.Title, result.Title)
}
```

### API Test Pattern
```go
func TestMovieHandler_GetMovie(t *testing.T) {
    server := setupTestServer()

    resp, err := server.GET("/api/v3/movie/1")

    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.Code)
}
```

### Benchmark Test Pattern
```go
func BenchmarkMovieService_GetMovie(b *testing.B) {
    service := setupBenchmarkService()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetMovie(ctx, 1)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

## Extension Points

### Add New Task Handler
1. Implement `TaskHandler` interface
2. Register in container: `container.TaskService.RegisterHandler(handler)`

### Add New Health Checker
1. Implement `HealthChecker` interface
2. Register in service: `healthService.RegisterChecker(checker)`

### Add New Notification Provider
1. Implement provider with `Send()` and `Test()` methods
2. Register in notification service

### Add New API Endpoint
1. Create handler function
2. Add route in `setupAPIRoutes()`

## File Locations

```
radarr-go/
├── cmd/radarr/main.go              # Application entry point
├── internal/
│   ├── api/                        # HTTP API layer
│   │   ├── server.go              # Server setup and routing
│   │   └── handlers.go            # Request handlers
│   ├── services/                   # Business logic layer
│   │   ├── container.go           # Dependency injection
│   │   ├── movie_service.go       # Movie operations
│   │   ├── task_service.go        # Background tasks
│   │   └── health_service.go      # Health monitoring
│   ├── database/                   # Data access layer
│   │   ├── database.go            # Connection management
│   │   └── generated/             # SQLC generated code
│   ├── models/                     # Domain models
│   ├── config/                     # Configuration management
│   └── logger/                     # Structured logging
├── migrations/                     # Database migrations
│   ├── postgres/                  # PostgreSQL migrations
│   └── mysql/                     # MySQL/MariaDB migrations
└── docs/                          # Documentation
```

## Git Workflow

### Commit Message Format
```
feat(api): add movie search endpoint
fix(database): resolve connection pool exhaustion
docs(developer): update architecture diagrams
refactor(services): extract validation logic
```

### Branch Naming
- `feature/movie-search-api`
- `fix/database-connection-leak`
- `refactor/service-container-cleanup`

## Quality Checklist

Before committing:
- [ ] Code passes `make lint`
- [ ] All tests pass with `make test`
- [ ] Code is formatted with `make fmt`
- [ ] No security issues found
- [ ] Documentation updated if needed
- [ ] Commit message follows convention

## Performance Tips

1. Use appropriate database queries (GORM vs SQLC)
2. Implement proper connection pooling
3. Use context timeouts for external calls
4. Monitor goroutine usage with worker pools
5. Profile with `go test -bench -benchmem`

## Security Checklist

- [ ] Input validation on all API endpoints
- [ ] Parameterized database queries
- [ ] Secure error messages (no internal details)
- [ ] API key authentication
- [ ] Proper context usage
- [ ] Resource limits configured
