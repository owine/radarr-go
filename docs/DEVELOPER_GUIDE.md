# Radarr-Go Developer Guide

## Table of Contents

1. [Architecture Deep-Dive](#architecture-deep-dive)
2. [Code Contribution Guidelines](#code-contribution-guidelines)
3. [Testing Strategy](#testing-strategy)
4. [Extension Development Guide](#extension-development-guide)
5. [Performance and Security Considerations](#performance-and-security-considerations)

---

## Architecture Deep-Dive

### System Architecture Overview

Radarr-Go follows a sophisticated layered architecture with clear separation of concerns and dependency injection patterns:

```
┌─────────────────────────────────────────────────────────────┐
│                    HTTP API Layer                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │
│  │   Gin       │ │ Middleware  │ │   Route Handlers    │    │
│  │   Engine    │ │   Stack     │ │   (API Endpoints)   │    │
│  └─────────────┘ └─────────────┘ └─────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Service Layer                             │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │
│  │  Service    │ │   Task      │ │   Health & Perf     │    │
│  │ Container   │ │  Service    │ │    Monitoring       │    │
│  │ (DI System) │ │(Worker Pool)│ │                     │    │
│  └─────────────┘ └─────────────┘ └─────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                   Data Layer                                │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │
│  │    GORM     │ │    SQLC     │ │     Database        │    │
│  │  (Complex   │ │ (Performance│ │    Migrations       │    │
│  │ Operations) │ │  Critical)  │ │                     │    │
│  └─────────────┘ └─────────────┘ └─────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              Infrastructure Layer                           │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────────────┐    │
│  │ PostgreSQL  │ │   MariaDB   │ │     File System     │    │
│  │  (Primary)  │ │(Alternative)│ │    & External       │    │
│  │             │ │             │ │     Services        │    │
│  └─────────────┘ └─────────────┘ └─────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

### Dependency Injection Pattern

The core of the architecture is the **Service Container** pattern implemented in `/internal/services/container.go`:

```go
// Container holds all services and their dependencies for dependency injection
type Container struct {
    DB     *database.Database
    Config *config.Config
    Logger *logger.Logger

    // Core services with explicit dependencies
    MovieService        *MovieService
    MovieFileService    *MovieFileService
    QualityService      *QualityService
    // ... additional services
}
```

#### Key Architectural Benefits:

1. **Explicit Dependencies**: All service dependencies are declared explicitly
2. **Single Responsibility**: Each service has a focused domain responsibility
3. **Testability**: Interface-based design enables easy mocking
4. **Configuration Management**: Centralized configuration with environment overrides

### Database Architecture

The project implements a **hybrid database strategy** combining GORM and sqlc:

```
┌─────────────────┐    ┌──────────────────┐
│      GORM       │    │       SQLC       │
│                 │    │                  │
│ • Complex Ops   │    │ • Performance    │
│ • Relationships │    │ • Type Safety    │
│ • Migrations    │    │ • Raw SQL Power  │
│ • Validations   │    │ • Zero Allocs    │
└─────────────────┘    └──────────────────┘
         │                       │
         └───────────┬───────────┘
                     │
            ┌────────▼────────┐
            │  Database.go    │
            │                 │
            │ • Connection    │
            │   Management    │
            │ • Pool Config   │
            │ • Multi-DB      │
            │   Support       │
            └─────────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
   PostgreSQL    MariaDB    Connection
   (Primary)   (Alternative)   Pooling
```

#### Database Features:

- **Multi-Database Support**: PostgreSQL (primary) and MariaDB with database-specific optimizations
- **Connection Pooling**: Configurable pool sizes with efficient connection management
- **Migration System**: golang-migrate for schema evolution
- **Type Safety**: sqlc generates type-safe Go code from SQL queries
- **Performance**: Prepared statements, index hints, and optimized queries

### Worker Pool and Task Scheduling System

The task service implements a sophisticated worker pool pattern:

```
                    TaskService
                        │
        ┌───────────────┼───────────────┐
        │               │               │
    ┌───▼───┐     ┌─────▼─────┐   ┌─────▼──────┐
    │High   │     │  Default  │   │ Background │
    │Priority│     │   Pool    │   │    Pool    │
    │Pool   │     │(3 workers)│   │(1 worker)  │
    │(2 wrks)│     └───────────┘   └────────────┘
    └───────┘           │               │
        │               │               │
        └───────────────┼───────────────┘
                        │
              ┌─────────▼─────────┐
              │   Task Handlers   │
              │                   │
              │ • RefreshMovie    │
              │ • ImportList      │
              │ • HealthCheck     │
              │ • Cleanup         │
              │ • WantedSearch    │
              └───────────────────┘
```

#### Worker Pool Features:

1. **Prioritized Execution**: High, normal, and background priority queues
2. **Graceful Cancellation**: Context-based task cancellation with status monitoring
3. **Panic Recovery**: Automatic panic recovery with error logging
4. **Progress Tracking**: Task progress updates and status management
5. **Scheduled Tasks**: Recurring task scheduling with interval management

### Real-Time Update System

The project includes WebSocket integration for real-time updates:

```
Client Browser ←─── WebSocket ←─── API Server
      │                              │
      │                              │
      └─── HTTP REST API ────────────┘
                │
                ▼
         Service Container
                │
         ┌──────┼──────┐
         ▼      ▼      ▼
    Task    Health  Calendar
   Service Service  Service
```

---

## Code Contribution Guidelines

### Go Code Style and Conventions

#### Project-Specific Standards

1. **Package Structure**:
   ```
   internal/
   ├── api/          # HTTP handlers and routing
   ├── services/     # Business logic layer
   ├── database/     # Database access layer
   ├── models/       # Domain models and types
   ├── config/       # Configuration management
   └── logger/       # Structured logging
   ```

2. **Naming Conventions**:
   - **Services**: Always end with `Service` (e.g., `MovieService`)
   - **Handlers**: Use descriptive action names (e.g., `handleGetMovies`)
   - **Database Models**: Use singular nouns (e.g., `Movie`, `QualityProfile`)
   - **Interfaces**: Use descriptive names without "I" prefix (e.g., `TaskHandler`)

3. **Error Handling**:
   ```go
   // ✅ Good - Wrapped errors with context
   if err != nil {
       return fmt.Errorf("failed to fetch movie with id %d: %w", id, err)
   }

   // ❌ Bad - Raw error without context
   if err != nil {
       return err
   }
   ```

#### Code Quality Requirements

**CRITICAL**: All code must pass these quality checks before committing:

1. **Pre-commit Hooks** (Recommended):
   ```bash
   pip install pre-commit
   pre-commit install
   pre-commit run --all-files  # Initial setup
   ```

2. **Manual Quality Checks**:
   ```bash
   make lint    # Run golangci-lint
   make fmt     # Format code
   make test    # Run tests
   make all     # Complete quality workflow
   ```

3. **Enabled Linters**:
   - `gosec`: Security analysis
   - `govet`: Suspicious constructs
   - `staticcheck`: Advanced Go analysis
   - `revive`: Enhanced linting rules
   - `unused`: Unused code detection

### Git Workflow and Commit Standards

#### Branch Strategy
- **main**: Production-ready code
- **feature/**: New feature development
- **fix/**: Bug fixes
- **refactor/**: Code improvements without functionality changes

#### Commit Message Format
Follow conventional commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Examples:
```
feat(api): add movie search endpoint with TMDB integration
fix(database): resolve connection pool exhaustion under high load
docs(developer): update architecture diagrams and contribution guide
refactor(services): extract common validation logic into shared utilities
```

#### Code Review Criteria

1. **Functionality**: Does the code solve the intended problem?
2. **Architecture**: Does it follow project patterns and principles?
3. **Performance**: Are there any obvious performance issues?
4. **Security**: Are there security vulnerabilities?
5. **Testing**: Is the code adequately tested?
6. **Documentation**: Are complex parts documented?

### Development Environment Setup

#### Required Tools
```bash
# Core development tools
make setup                    # Install air, golangci-lint, migrate

# Pre-commit hooks (recommended)
pip install pre-commit
pre-commit install

# Database tools (optional)
brew install postgresql mariadb  # For local database testing
```

#### Development Workflow
```bash
# 1. Start development server with hot reload
make dev

# 2. Run tests during development
make test
make test-coverage           # With HTML coverage report

# 3. Quality checks before commit
make lint
make fmt

# 4. Build for testing
make build-all              # All platforms
make build                  # Current platform only
```

---

## Testing Strategy

### Testing Architecture

The project implements a comprehensive testing strategy across multiple layers:

```
┌─────────────────────────────────────────────────┐
│                Unit Tests                       │
│  • Service logic testing                       │
│  • Individual component testing                │
│  • Mock-based dependency isolation             │
└─────────────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────────────┐
│             Integration Tests                   │
│  • Database interaction testing                │
│  • Service integration testing                 │
│  • Container-based database testing            │
└─────────────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────────────┐
│           End-to-End API Tests                  │
│  • HTTP endpoint testing                       │
│  • Complete workflow testing                   │
│  • Authentication and authorization testing    │
└─────────────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────────────┐
│            Benchmark Tests                      │
│  • Performance regression detection            │
│  • Memory allocation tracking                  │
│  • Database query performance                  │
└─────────────────────────────────────────────────┘
```

### Unit Testing with Mocks and Interfaces

#### Interface-Based Design for Testability

All services implement interfaces to enable mocking:

```go
// Example service interface
type MovieServiceInterface interface {
    GetMovie(ctx context.Context, id int) (*models.Movie, error)
    CreateMovie(ctx context.Context, movie *models.Movie) error
    UpdateMovie(ctx context.Context, movie *models.Movie) error
    DeleteMovie(ctx context.Context, id int) error
}

// Implementation
type MovieService struct {
    db     *database.Database
    logger *logger.Logger
}

func (s *MovieService) GetMovie(ctx context.Context, id int) (*models.Movie, error) {
    // Implementation
}
```

#### Mock Generation and Usage

```go
// Test example with mocks
func TestMovieHandler_GetMovie(t *testing.T) {
    // Setup mock service
    mockService := &MockMovieService{}
    mockService.On("GetMovie", mock.Anything, 1).Return(&models.Movie{
        ID:    1,
        Title: "Test Movie",
    }, nil)

    // Setup test server
    server := setupTestServer(mockService)

    // Execute request
    resp, err := server.GET("/api/v3/movie/1")

    // Assertions
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.Code)
    mockService.AssertExpectations(t)
}
```

### Integration Testing with Database Containers

#### Testcontainers Integration

The project uses testcontainers for isolated database testing:

```go
// internal/testhelpers/containers.go
func SetupPostgresContainer(ctx context.Context) (*testcontainers.Container, *database.Database, error) {
    // Create PostgreSQL container
    req := testcontainers.ContainerRequest{
        Image:        "postgres:15-alpine",
        ExposedPorts: []string{"5432/tcp"},
        Env: map[string]string{
            "POSTGRES_DB":       "radarr_test",
            "POSTGRES_USER":     "test",
            "POSTGRES_PASSWORD": "test",
        },
        WaitingFor: wait.ForListeningPort("5432/tcp"),
    }

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: req,
        Started:         true,
    })

    // Setup database connection
    // Return container and database instance
}
```

#### Database Testing Patterns

1. **Test Isolation**: Each test uses a fresh database state
2. **Migration Testing**: Verify migrations work correctly
3. **Multi-Database Testing**: Test both PostgreSQL and MariaDB
4. **Performance Testing**: Benchmark database operations

### Benchmark Testing for Performance Regression Detection

#### Performance Benchmarks

```go
// Example benchmark test
func BenchmarkMovieService_GetMovie(b *testing.B) {
    // Setup
    service := setupMovieService()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.GetMovie(ctx, 1)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// Database query benchmark
func BenchmarkMovieService_SearchMovies(b *testing.B) {
    service := setupMovieService()
    ctx := context.Background()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, err := service.SearchMovies(ctx, "action", 10, 0)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### Running Benchmark Tests

```bash
# Run all benchmarks
make test-bench

# Run specific benchmarks with memory profiling
go test -bench=BenchmarkMovieService -benchmem ./internal/services

# Compare benchmark results over time
go test -bench=. -count=5 ./... > bench_results.txt
```

### Test Data Management and Fixtures

#### Test Data Patterns

```go
// internal/testhelpers/fixtures.go
func CreateTestMovie() *models.Movie {
    return &models.Movie{
        Title:         "Test Movie",
        Year:          2023,
        TMDBID:        12345,
        Status:        "released",
        QualityProfileID: 1,
    }
}

func CreateTestMovies(count int) []*models.Movie {
    movies := make([]*models.Movie, count)
    for i := 0; i < count; i++ {
        movies[i] = &models.Movie{
            Title:  fmt.Sprintf("Test Movie %d", i+1),
            Year:   2020 + i,
            TMDBID: 10000 + i,
        }
    }
    return movies
}
```

#### Test Database Management

```go
// Test cleanup pattern
func setupTestDatabase(t *testing.T) *database.Database {
    db := setupContainerDatabase(t)

    t.Cleanup(func() {
        db.Close()
    })

    return db
}
```

### Testing Commands

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests only
go test -tags=integration ./...

# Run database-specific tests
RADARR_DATABASE_TYPE=postgres go test ./internal/services
RADARR_DATABASE_TYPE=mariadb go test ./internal/services

# Run benchmark tests
make test-bench

# Run example tests (documentation validation)
make test-examples
```

---

## Extension Development Guide

### Adding New Notification Providers

#### 1. Create Provider Interface Implementation

```go
// internal/services/notifications/custom_provider.go
package notifications

import (
    "context"
    "fmt"
)

type CustomProvider struct {
    config CustomProviderConfig
    logger *logger.Logger
}

type CustomProviderConfig struct {
    WebhookURL string `json:"webhookUrl"`
    APIKey     string `json:"apiKey"`
    Channel    string `json:"channel"`
}

func NewCustomProvider(config CustomProviderConfig, logger *logger.Logger) *CustomProvider {
    return &CustomProvider{
        config: config,
        logger: logger,
    }
}

func (p *CustomProvider) Send(ctx context.Context, notification *models.Notification) error {
    // Implement notification sending logic
    payload := map[string]interface{}{
        "title":   notification.Title,
        "message": notification.Message,
        "level":   notification.Level,
        "channel": p.config.Channel,
    }

    return p.sendWebhook(ctx, payload)
}

func (p *CustomProvider) Test(ctx context.Context) error {
    testNotification := &models.Notification{
        Title:   "Test Notification",
        Message: "This is a test from Radarr-Go",
        Level:   "info",
    }
    return p.Send(ctx, testNotification)
}
```

#### 2. Register Provider in Service

```go
// internal/services/notification_service.go
func (ns *NotificationService) RegisterProvider(providerType string, factory ProviderFactory) {
    ns.providers[providerType] = factory
}

// In NewNotificationService()
func NewNotificationService(db *database.Database, logger *logger.Logger) *NotificationService {
    service := &NotificationService{
        db:        db,
        logger:    logger,
        providers: make(map[string]ProviderFactory),
    }

    // Register built-in providers
    service.RegisterProvider("discord", NewDiscordProvider)
    service.RegisterProvider("slack", NewSlackProvider)
    service.RegisterProvider("custom", NewCustomProvider) // Add your provider

    return service
}
```

### Creating Custom Task Handlers

#### 1. Implement TaskHandler Interface

```go
// internal/services/custom_task_handler.go
type CustomTaskHandler struct {
    movieService *MovieService
    logger       *logger.Logger
}

func NewCustomTaskHandler(movieService *MovieService) *CustomTaskHandler {
    return &CustomTaskHandler{
        movieService: movieService,
    }
}

func (h *CustomTaskHandler) Execute(ctx context.Context, task *models.TaskV2, updateProgress func(percent int, message string)) error {
    updateProgress(0, "Starting custom task")

    // Extract parameters from task body
    var params CustomTaskParams
    if err := json.Unmarshal(task.Body, &params); err != nil {
        return fmt.Errorf("invalid task parameters: %w", err)
    }

    // Perform task logic with progress updates
    updateProgress(25, "Processing movies")

    for i, movieID := range params.MovieIDs {
        select {
        case <-ctx.Done():
            return ctx.Err() // Handle cancellation
        default:
            // Process movie
            if err := h.processMovie(ctx, movieID); err != nil {
                h.logger.Errorw("Failed to process movie", "movieId", movieID, "error", err)
            }

            // Update progress
            percent := 25 + (50 * (i + 1) / len(params.MovieIDs))
            updateProgress(percent, fmt.Sprintf("Processed %d/%d movies", i+1, len(params.MovieIDs)))
        }
    }

    updateProgress(100, "Custom task completed")
    return nil
}

func (h *CustomTaskHandler) GetName() string {
    return "CustomTask"
}

func (h *CustomTaskHandler) GetDescription() string {
    return "Custom task for processing movies with specific logic"
}
```

#### 2. Register Handler

```go
// In services/container.go NewContainer()
container.TaskService.RegisterHandler(NewCustomTaskHandler(container.MovieService))
```

### Extending the Health Monitoring System

#### 1. Create Custom Health Checker

```go
// internal/services/custom_health_checker.go
type CustomHealthChecker struct {
    name   string
    logger *logger.Logger
}

func NewCustomHealthChecker(logger *logger.Logger) HealthChecker {
    return &CustomHealthChecker{
        name:   "custom-service",
        logger: logger,
    }
}

func (c *CustomHealthChecker) Name() string {
    return c.name
}

func (c *CustomHealthChecker) Check(ctx context.Context) *models.HealthStatus {
    // Implement your health check logic
    status := &models.HealthStatus{
        CheckName:   c.name,
        Status:      "healthy",
        LastChecked: time.Now(),
    }

    // Example: Check external service
    if err := c.checkExternalService(ctx); err != nil {
        status.Status = "unhealthy"
        status.Message = err.Error()
        status.Data = map[string]interface{}{
            "error": err.Error(),
        }
    }

    return status
}

func (c *CustomHealthChecker) checkExternalService(ctx context.Context) error {
    // Implement external service health check
    client := &http.Client{Timeout: 5 * time.Second}
    resp, err := client.Get("https://api.external-service.com/health")
    if err != nil {
        return fmt.Errorf("failed to reach external service: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("external service returned status %d", resp.StatusCode)
    }

    return nil
}
```

#### 2. Register Health Checker

```go
// In services/health_service.go
func (hs *HealthService) RegisterChecker(checker HealthChecker) {
    hs.checkers[checker.Name()] = checker
    hs.logger.Infow("Registered health checker", "name", checker.Name())
}

// In container initialization
container.HealthService.RegisterChecker(NewCustomHealthChecker(logger))
```

### Adding New API Endpoints

#### 1. Create Handler Function

```go
// internal/api/custom_handlers.go
func (s *Server) handleCustomEndpoint(c *gin.Context) {
    // Extract parameters
    param := c.Param("id")
    query := c.Query("filter")

    // Validate input
    if param == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "ID parameter required"})
        return
    }

    // Call service
    result, err := s.services.CustomService.ProcessRequest(c.Request.Context(), param, query)
    if err != nil {
        s.logger.Errorw("Custom endpoint failed", "error", err, "param", param)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // Return response
    c.JSON(http.StatusOK, result)
}
```

#### 2. Register Route

```go
// internal/api/server.go - in setupAPIRoutes()
func (s *Server) setupCustomRoutes(v3 *gin.RouterGroup) {
    customRoutes := v3.Group("/custom")
    customRoutes.GET("/:id", s.handleCustomEndpoint)
    customRoutes.POST("", s.handleCreateCustomResource)
    customRoutes.PUT("/:id", s.handleUpdateCustomResource)
    customRoutes.DELETE("/:id", s.handleDeleteCustomResource)
}

// Add to setupAPIRoutes()
func (s *Server) setupAPIRoutes(v3 *gin.RouterGroup) {
    // ... existing routes
    s.setupCustomRoutes(v3)
}
```

### Database Migration Best Practices

#### 1. Creating Migrations

```bash
# Create new migration
migrate create -ext sql -dir migrations/postgres add_custom_table
migrate create -ext sql -dir migrations/mysql add_custom_table
```

#### 2. Migration File Structure

```sql
-- migrations/postgres/001_add_custom_table.up.sql
CREATE TABLE custom_resources (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_custom_resources_name ON custom_resources(name);
```

```sql
-- migrations/postgres/001_add_custom_table.down.sql
DROP TABLE IF EXISTS custom_resources;
```

#### 3. Adding GORM Models

```go
// internal/models/custom_resource.go
type CustomResource struct {
    ID          int               `json:"id" gorm:"primaryKey;autoIncrement"`
    Name        string            `json:"name" gorm:"not null;size:255;uniqueIndex"`
    Description string            `json:"description" gorm:"type:text"`
    Config      JSONField         `json:"config" gorm:"type:jsonb"`
    CreatedAt   time.Time         `json:"createdAt" gorm:"autoCreateTime"`
    UpdatedAt   time.Time         `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (CustomResource) TableName() string {
    return "custom_resources"
}
```

---

## Performance and Security Considerations

### Performance Optimization Guidelines

#### 1. Database Performance

**Query Optimization:**
```go
// ✅ Good - Use indexes and limit results
func (s *MovieService) GetMoviesByTitle(ctx context.Context, title string, limit int) ([]*models.Movie, error) {
    var movies []*models.Movie
    return movies, s.db.GORM.Where("title ILIKE ?", "%"+title+"%").
        Order("created_at DESC").
        Limit(limit).
        Find(&movies).Error
}

// ❌ Bad - Full table scan without limits
func (s *MovieService) GetAllMoviesByTitle(ctx context.Context, title string) ([]*models.Movie, error) {
    var movies []*models.Movie
    return movies, s.db.GORM.Where("title LIKE ?", "%"+title+"%").Find(&movies).Error
}
```

**Connection Pool Management:**
```go
// Configure connection pools based on workload
func configureConnectionPool(cfg *config.DatabaseConfig, sqlDB *sql.DB, pgxPool *pgxpool.Pool) {
    if cfg.MaxConnections > 0 {
        sqlDB.SetMaxOpenConns(cfg.MaxConnections)
        sqlDB.SetMaxIdleConns(cfg.MaxConnections / 2)
        sqlDB.SetConnMaxLifetime(time.Hour)
    }
}
```

#### 2. Memory Management

**Efficient Data Structures:**
```go
// ✅ Good - Use appropriate data types
type MovieSummary struct {
    ID    int    `json:"id"`
    Title string `json:"title"`
    Year  int    `json:"year"`
}

// ❌ Bad - Loading full objects when summary is sufficient
func (s *MovieService) GetMovieSummaries(ctx context.Context) ([]*models.Movie, error) {
    // This loads all fields unnecessarily
}
```

**Context Usage:**
```go
// ✅ Good - Proper context usage with timeouts
func (s *MovieService) GetMovie(ctx context.Context, id int) (*models.Movie, error) {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    var movie models.Movie
    return &movie, s.db.GORM.WithContext(ctx).First(&movie, id).Error
}
```

#### 3. Goroutine Management

**Worker Pool Pattern:**
```go
// Use bounded worker pools to prevent resource exhaustion
type TaskWorkerPool struct {
    workers chan struct{} // Semaphore for max concurrent workers
    queue   chan *models.TaskV2
}

func (pool *TaskWorkerPool) worker(ctx context.Context, service *TaskService) {
    for {
        select {
        case <-ctx.Done():
            return
        case task := <-pool.queue:
            pool.workers <- struct{}{} // Acquire worker slot
            pool.executeTask(ctx, service, task)
            <-pool.workers // Release worker slot
        }
    }
}
```

### Security Best Practices

#### 1. Input Validation and Sanitization

**API Input Validation:**
```go
type CreateMovieRequest struct {
    Title           string `json:"title" validate:"required,min=1,max=255"`
    Year            int    `json:"year" validate:"required,min=1900,max=2100"`
    TMDBID          int    `json:"tmdbId" validate:"required,min=1"`
    QualityProfileID int    `json:"qualityProfileId" validate:"required,min=1"`
}

func (s *Server) handleCreateMovie(c *gin.Context) {
    var req CreateMovieRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    // Validate using go-playground/validator
    if err := validator.New().Struct(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Process validated request
}
```

#### 2. SQL Injection Prevention

**Parameterized Queries:**
```go
// ✅ Good - Parameterized queries (GORM automatically handles this)
func (s *MovieService) SearchMovies(ctx context.Context, title string) ([]*models.Movie, error) {
    var movies []*models.Movie
    return movies, s.db.GORM.Where("title ILIKE ?", "%"+title+"%").Find(&movies).Error
}

// ✅ Good - SQLC generates safe queries
func (s *MovieService) GetMovieByTMDBID(ctx context.Context, tmdbID int) (*models.Movie, error) {
    // This uses sqlc-generated type-safe queries
    return s.db.Postgres.GetMovieByTMDBID(ctx, int32(tmdbID))
}
```

#### 3. Authentication and Authorization

**API Key Middleware:**
```go
func apiKeyMiddleware(apiKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        providedKey := c.GetHeader("X-API-Key")
        if providedKey == "" {
            providedKey = c.Query("apikey")
        }

        // Use constant-time comparison to prevent timing attacks
        if !compareAPIKeys(providedKey, apiKey) {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }

        c.Next()
    }
}

func compareAPIKeys(provided, expected string) bool {
    return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
```

#### 4. Secure Configuration Management

**Environment Variable Handling:**
```go
// ✅ Good - Secure defaults and validation
func setDefaults(vip *viper.Viper, dataDir string) {
    vip.SetDefault("auth.api_key", generateRandomAPIKey()) // Generate secure default
    vip.SetDefault("server.enable_ssl", true)              // Secure by default
    vip.SetDefault("database.max_connections", 10)         // Reasonable limits
}

func validateConfig(config *Config) error {
    if config.Auth.APIKey != "" && len(config.Auth.APIKey) < 32 {
        return fmt.Errorf("API key must be at least 32 characters long")
    }
    return nil
}
```

#### 5. Error Handling Security

**Safe Error Messages:**
```go
func (s *Server) handleGetMovie(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        // Don't expose internal details
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
        return
    }

    movie, err := s.services.MovieService.GetMovie(c.Request.Context(), id)
    if err != nil {
        // Log detailed error internally
        s.logger.Errorw("Failed to get movie", "movieId", id, "error", err)

        // Return generic error to client
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        }
        return
    }

    c.JSON(http.StatusOK, movie)
}
```

### Monitoring and Observability

#### 1. Structured Logging

```go
// Use structured logging with context
func (s *MovieService) UpdateMovie(ctx context.Context, movie *models.Movie) error {
    logger := s.logger.With("movieId", movie.ID, "title", movie.Title)

    logger.Infow("Updating movie")

    if err := s.db.GORM.Save(movie).Error; err != nil {
        logger.Errorw("Failed to update movie", "error", err)
        return fmt.Errorf("failed to update movie: %w", err)
    }

    logger.Infow("Movie updated successfully")
    return nil
}
```

#### 2. Performance Metrics

```go
// Track performance metrics
func (s *MovieService) GetMovie(ctx context.Context, id int) (*models.Movie, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        s.logger.Infow("Movie lookup completed",
            "movieId", id,
            "duration", duration,
            "duration_ms", duration.Milliseconds())
    }()

    // Implementation
}
```

This comprehensive developer guide provides the foundation for maintaining high code quality while enabling rapid contributor onboarding to the radarr-go project. The architecture patterns, testing strategies, and security considerations ensure that new features can be added safely and efficiently while maintaining the project's excellent standards.
