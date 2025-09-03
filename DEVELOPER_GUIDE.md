# Radarr-Go Developer Guide

This comprehensive guide enables new contributors to understand the sophisticated architecture of radarr-go and contribute effectively while maintaining the high code quality standards.

## Table of Contents

1. [Architecture Deep-Dive](#architecture-deep-dive)
2. [Code Contribution Guidelines](#code-contribution-guidelines)
3. [Testing Strategy Documentation](#testing-strategy-documentation)
4. [Extension Development Guide](#extension-development-guide)
5. [Performance and Security Considerations](#performance-and-security-considerations)

---

## Architecture Deep-Dive

### System Architecture Overview

Radarr-go implements a sophisticated multi-tier enterprise architecture designed for high performance, maintainability, and extensibility:

```
┌─────────────────────────────────────────────────────────────────┐
│                     HTTP API Layer (Gin)                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │     Middleware Stack                                     │   │
│  │  Logging → CORS → API Key Auth → Rate Limiting          │   │
│  └──────────────────────────────────────────────────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                  Service Container Layer                       │
│  ┌──────────────┬──────────────┬─────────────┬───────────────┐   │
│  │ Core Business│ File Mgmt    │ Health Mon. │ Communication │   │
│  │ Services     │ Services     │ Services    │ Services      │   │
│  │              │              │             │               │   │
│  │ • Movies     │ • File Ops   │ • Health    │ • Notifications│  │
│  │ • Quality    │ • Import     │ • Perf Mon  │ • Calendar    │   │
│  │ • Search     │ • Rename     │ • Metrics   │ • iCal        │   │
│  │ • Tasks      │ • Media Info │             │               │   │
│  └──────────────┴──────────────┴─────────────┴───────────────┘   │
├─────────────────────────────────────────────────────────────────┤
│                  Data Access Layer                             │
│  ┌────────────────┬────────────────┬───────────────────────┐    │
│  │ GORM ORM       │ sqlc Queries   │ Database Abstraction  │    │
│  │ • Relationships│ • Performance  │ • PostgreSQL/MariaDB  │    │
│  │ • Migrations   │ • Type Safety  │ • Connection Pooling  │    │
│  │ • Transactions │ • Raw SQL      │ • Health Monitoring   │    │
│  └────────────────┴────────────────┴───────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Dependency Injection Container Architecture

The service container (`/Users/owine/Git/radarr-go/internal/services/container.go`) implements sophisticated dependency injection patterns:

#### Service Container Features

1. **Hierarchical Dependency Resolution**: Services are initialized in dependency order
2. **Singleton Pattern**: All services are registered as singletons with shared database connections
3. **Interface-Based Design**: Services depend on interfaces for easy testing and mocking
4. **Lifecycle Management**: Container handles service startup, configuration, and shutdown

#### Service Categories

```go
// Core Business Services
type Container struct {
    // Database and Infrastructure
    DB     *database.Database
    Config *config.Config
    Logger *logger.Logger

    // Core Movie Management
    MovieService        *MovieService
    MovieFileService    *MovieFileService
    QualityService      *QualityService

    // Search and Acquisition
    IndexerService      *IndexerService
    DownloadService     *DownloadService
    SearchService       *SearchService

    // Task and Workflow Management
    TaskService         *TaskService
    ImportListService   *ImportListService

    // Health and Performance Monitoring
    HealthService       *HealthService
    PerformanceMonitor  *PerformanceMonitor

    // Communication and Integration
    NotificationService *NotificationService
    CalendarService     *CalendarService
}
```

### Database Architecture with Multi-Backend Support

#### Hybrid Query Strategy

The application uses both GORM and sqlc for optimal performance:

```go
type Database struct {
    GORM     *gorm.DB                    // Complex business logic queries
    Postgres *sqlcPostgres.Queries      // High-performance PostgreSQL queries
    MySQL    *sqlcMySQL.Queries         // High-performance MySQL queries
    DB       *sql.DB                    // Direct SQL access for migrations
    PgxPool  *pgxpool.Pool             // PostgreSQL connection pool
    DbType   string                     // Database type identifier
}
```

#### Database Features

1. **Connection Pooling**: Configurable pool sizes with health monitoring
2. **Prepared Statements**: Automatic caching for performance optimization
3. **Migration System**: Database-specific migrations with rollback support
4. **Multi-Database Support**: PostgreSQL (recommended) and MariaDB/MySQL

### Task Scheduling and Worker Pool System

The task system (`/Users/owine/Git/radarr-go/internal/services/task_service.go`) implements enterprise-grade job processing:

#### Worker Pool Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Task Scheduler                               │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │ Cron-based Scheduling │ Dynamic Intervals │ Timezone Support│ │
│  └───────────────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                    Worker Pool Manager                         │
│  ┌──────────────┬──────────────┬─────────────────────────────┐  │
│  │High Priority │   Default    │      Background             │  │
│  │Pool (2)      │   Pool (3)   │      Pool (1)               │  │
│  │              │              │                             │  │
│  │• Critical    │• Normal      │• Cleanup                    │  │
│  │  tasks       │  operations  │• Maintenance                │  │
│  │• Health      │• Metadata    │• Log rotation               │  │
│  │  checks      │  refresh     │                             │  │
│  └──────────────┴──────────────┴─────────────────────────────┘  │
├─────────────────────────────────────────────────────────────────┤
│                    Task Handlers                               │
│  • Movie refresh handlers     • Health check handlers          │
│  • Import list sync handlers  • Performance monitoring         │
│  • Search automation handlers • Cleanup and maintenance        │
└─────────────────────────────────────────────────────────────────┘
```

#### Task Handler Interface

```go
type TaskHandler interface {
    Execute(ctx context.Context, task *models.TaskV2, updateProgress func(int, string)) error
    GetName() string
    GetDescription() string
}
```

#### Key Features

1. **Priority Queues**: Tasks routed to appropriate worker pools based on priority
2. **Cancellation Support**: Context-based cancellation with graceful shutdown
3. **Progress Tracking**: Real-time task progress reporting
4. **Error Recovery**: Automatic retry logic with exponential backoff
5. **Panic Recovery**: Task panics don't crash the application

### Health Monitoring System Architecture

The health system provides comprehensive application monitoring:

#### Health Checker Interface

```go
type HealthChecker interface {
    Name() string
    Type() models.HealthCheckType
    Check(ctx context.Context) models.HealthCheckExecution
    IsEnabled() bool
    GetInterval() time.Duration
}
```

#### Health Check Categories

1. **System Health**: CPU, memory, disk space monitoring
2. **Database Health**: Connection pooling, query performance
3. **External Services**: Indexer connectivity, download client status
4. **Application Health**: Service health, configuration validation

### Real-time Update System with WebSocket Integration

Future implementation will include:

1. **Event-Driven Architecture**: Domain events for real-time updates
2. **WebSocket Gateway**: Real-time client notifications
3. **State Synchronization**: Client-server state consistency
4. **Connection Management**: Auto-reconnection and scaling support

---

## Code Contribution Guidelines

### Go Code Style and Conventions

#### Project Standards

Radarr-go follows strict coding standards enforced by `golangci-lint` with 15+ enabled linters:

```yaml
# Required linters (enforced in CI)
- bodyclose      # HTTP response body closure
- errcheck       # Unchecked error handling
- gosec          # Security analysis
- govet          # Suspicious constructs
- ineffassign    # Ineffectual assignments
- misspell       # Common misspellings
- revive         # Advanced Go linting
- staticcheck    # Advanced static analysis
- unused         # Unused code detection
- whitespace     # Unnecessary whitespace
```

#### Naming Conventions

```go
// Variables: camelCase
var movieID int
var qualityProfile *QualityProfile

// Constants: UPPER_SNAKE_CASE
const (
    MaxRetries     = 3
    DefaultTimeout = 30 * time.Second
)

// Functions/Methods: PascalCase (exported), camelCase (private)
func GetMovieByID(id int) (*Movie, error) { }
func validateMovieData(movie *Movie) error { }

// Types: PascalCase
type MovieService struct {
    db     *database.Database
    logger *logger.Logger
}

// Interfaces: PascalCase with descriptive names
type MovieRepository interface {
    GetMovie(ctx context.Context, id int) (*Movie, error)
    CreateMovie(ctx context.Context, movie *Movie) error
}
```

#### Error Handling Patterns

```go
// Error wrapping with context
func (s *MovieService) GetMovie(id int) (*Movie, error) {
    movie, err := s.repository.FindByID(id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("movie not found: %d", id)
        }
        return nil, fmt.Errorf("failed to fetch movie %d: %w", id, err)
    }
    return movie, nil
}

// Error variables with Err prefix
var (
    ErrMovieNotFound     = errors.New("movie not found")
    ErrInvalidQuality    = errors.New("invalid quality profile")
    ErrDuplicateMovie    = errors.New("movie already exists")
)
```

#### Function Structure and Documentation

```go
// GetMovie retrieves a movie by its ID from the database.
// It returns ErrMovieNotFound if the movie doesn't exist.
// Example:
//   movie, err := service.GetMovie(123)
//   if err != nil {
//       if errors.Is(err, ErrMovieNotFound) {
//           // Handle not found case
//       }
//       return fmt.Errorf("failed to get movie: %w", err)
//   }
func (s *MovieService) GetMovie(ctx context.Context, id int) (*Movie, error) {
    if id <= 0 {
        return nil, fmt.Errorf("invalid movie ID: %d", id)
    }

    movie, err := s.repository.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch movie %d: %w", id, err)
    }

    return movie, nil
}
```

### Testing Requirements and Patterns

#### Testing Standards

1. **Test Coverage**: Maintain >80% code coverage for all packages
2. **Test File Naming**: Use `_test.go` suffix for all test files
3. **Test Function Naming**: Use `TestFunction_Scenario` pattern
4. **Benchmarks**: Include benchmark tests for performance-critical code

#### Test Structure Pattern

```go
func TestMovieService_GetMovie_Success(t *testing.T) {
    // Arrange
    db, logger := setupTestDB(t)
    service := NewMovieService(db, logger)

    movie := &Movie{
        ID:    1,
        Title: "Test Movie",
    }
    require.NoError(t, db.GORM.Create(movie).Error)

    // Act
    result, err := service.GetMovie(context.Background(), 1)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, movie.Title, result.Title)
    assert.Equal(t, movie.ID, result.ID)
}

func TestMovieService_GetMovie_NotFound(t *testing.T) {
    // Arrange
    db, logger := setupTestDB(t)
    service := NewMovieService(db, logger)

    // Act
    result, err := service.GetMovie(context.Background(), 999)

    // Assert
    assert.Error(t, err)
    assert.Nil(t, result)
    assert.Contains(t, err.Error(), "not found")
}
```

### Code Review Process and Criteria

#### Pre-commit Requirements

```bash
# Install pre-commit hooks (recommended)
pip install pre-commit
pre-commit install

# Manual quality checks
make lint          # Run all linters
make test          # Run test suite
make test-coverage # Generate coverage report
make fmt           # Format code
```

#### Review Checklist

1. **Code Quality**
   - [ ] All linters pass without warnings
   - [ ] Test coverage maintained or improved
   - [ ] Error handling follows project patterns
   - [ ] Documentation updated for public APIs

2. **Architecture Compliance**
   - [ ] Follows dependency injection patterns
   - [ ] Uses appropriate service layer
   - [ ] Database queries use GORM or sqlc appropriately
   - [ ] Proper context usage for cancellation

3. **Performance Considerations**
   - [ ] Database queries optimized with indexes
   - [ ] Memory allocations minimized in hot paths
   - [ ] Goroutine leaks prevented
   - [ ] Context timeouts implemented

### Git Workflow and Commit Message Standards

#### Commit Message Format (Conventional Commits)

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

#### Examples

```bash
feat(api): add movie collection management endpoints

- Add CRUD operations for movie collections
- Implement TMDB collection sync functionality
- Add collection statistics tracking

Closes #123

fix(database): resolve connection pool exhaustion

The connection pool was not properly releasing connections
after failed transactions, causing pool exhaustion under
high load conditions.

Breaking-change: Database configuration format updated
```

#### Commit Types

- `feat`: New feature implementation
- `fix`: Bug fix
- `docs`: Documentation updates
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring without feature changes
- `perf`: Performance improvements
- `test`: Test additions or improvements
- `chore`: Build process or tooling changes

---

## Testing Strategy Documentation

### Unit Testing with Mocks and Interfaces

#### Test Database Setup

```go
func setupTestDB(t *testing.T) (*database.Database, *logger.Logger) {
    // Use in-memory SQLite for fast tests
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
        Logger: gormLogger.Default.LogMode(gormLogger.Silent),
    })
    require.NoError(t, err)

    // Auto-migrate test schema
    err = db.AutoMigrate(&models.Movie{}, &models.QualityProfile{})
    require.NoError(t, err)

    logger := logger.NewTest()
    return &database.Database{GORM: db}, logger
}
```

#### Interface-Based Mocking

```go
// Define interfaces for dependencies
type MovieRepository interface {
    GetByID(ctx context.Context, id int) (*Movie, error)
    Create(ctx context.Context, movie *Movie) error
}

// Mock implementation for testing
type MockMovieRepository struct {
    movies map[int]*Movie
}

func (m *MockMovieRepository) GetByID(ctx context.Context, id int) (*Movie, error) {
    if movie, exists := m.movies[id]; exists {
        return movie, nil
    }
    return nil, gorm.ErrRecordNotFound
}
```

### Integration Testing with Database Containers

#### Database Integration Tests

```go
func TestMovieService_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup test database (PostgreSQL in CI, SQLite locally)
    db := setupIntegrationDB(t)
    defer cleanupTestDB(t, db)

    service := NewMovieService(db, logger.NewTest())

    // Test complete workflow
    movie := &Movie{Title: "Integration Test Movie"}
    err := service.CreateMovie(context.Background(), movie)
    require.NoError(t, err)

    retrieved, err := service.GetMovie(context.Background(), movie.ID)
    require.NoError(t, err)
    assert.Equal(t, movie.Title, retrieved.Title)
}
```

### Benchmark Testing for Performance Regression Detection

#### Performance Benchmarks

```go
func BenchmarkMovieService_GetMovie(b *testing.B) {
    db, logger := setupTestDB(b)
    service := NewMovieService(db, logger)

    // Setup test data
    movie := &Movie{ID: 1, Title: "Benchmark Movie"}
    db.GORM.Create(movie)

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        _, err := service.GetMovie(context.Background(), 1)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDatabase_QualityProfiles(b *testing.B) {
    db := setupBenchmarkDB(b)

    // Benchmark GORM vs sqlc performance
    b.Run("GORM", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var profiles []QualityProfile
            db.GORM.Find(&profiles)
        }
    })

    b.Run("sqlc", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            profiles, _ := db.Postgres.GetAllQualityProfiles(context.Background())
            _ = profiles
        }
    })
}
```

### End-to-End Testing Strategies

#### API Integration Tests

```go
func TestMovieAPI_EndToEnd(t *testing.T) {
    // Setup test server
    router := setupTestRouter(t)
    server := httptest.NewServer(router)
    defer server.Close()

    // Test complete movie creation workflow
    movieData := `{"title": "Test Movie", "tmdbId": 12345}`

    // Create movie
    resp, err := http.Post(server.URL+"/api/v3/movie",
        "application/json", strings.NewReader(movieData))
    require.NoError(t, err)
    assert.Equal(t, http.StatusCreated, resp.StatusCode)

    // Parse response
    var movie Movie
    err = json.NewDecoder(resp.Body).Decode(&movie)
    require.NoError(t, err)

    // Verify movie was created
    resp, err = http.Get(fmt.Sprintf("%s/api/v3/movie/%d", server.URL, movie.ID))
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

### Test Data Management and Fixtures

#### Test Fixtures

```go
// Test data factory
func CreateTestMovie(t *testing.T, db *gorm.DB, options ...func(*Movie)) *Movie {
    movie := &Movie{
        Title:     "Test Movie " + strconv.Itoa(rand.Int()),
        TMDbID:    rand.Int31(),
        Year:      2024,
        Status:    "announced",
    }

    // Apply options
    for _, option := range options {
        option(movie)
    }

    require.NoError(t, db.Create(movie).Error)
    return movie
}

// Usage in tests
func TestMovieSearch(t *testing.T) {
    db, _ := setupTestDB(t)

    // Create test data with options
    actionMovie := CreateTestMovie(t, db, func(m *Movie) {
        m.Genres = []string{"Action", "Thriller"}
        m.Year = 2023
    })

    comedyMovie := CreateTestMovie(t, db, func(m *Movie) {
        m.Genres = []string{"Comedy"}
        m.Year = 2024
    })

    // Test search functionality
    // ...
}
```

---

## Extension Development Guide

### How to Add New Notification Providers

#### Notification Provider Interface

```go
// Implement this interface for new providers
type NotificationProvider interface {
    Send(ctx context.Context, notification *Notification) error
    Test(ctx context.Context, config map[string]interface{}) error
    GetConfigSchema() map[string]interface{}
    GetName() string
    GetDescription() string
}
```

#### Example: Adding a Teams Provider

1. **Create Provider Implementation**

```go
// /internal/services/notifications/teams.go
package notifications

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type TeamsProvider struct {
    webhookURL string
    client     *http.Client
}

func NewTeamsProvider(config map[string]interface{}) (*TeamsProvider, error) {
    webhookURL, ok := config["webhook_url"].(string)
    if !ok || webhookURL == "" {
        return nil, fmt.Errorf("webhook_url is required")
    }

    return &TeamsProvider{
        webhookURL: webhookURL,
        client:     &http.Client{Timeout: 30 * time.Second},
    }, nil
}

func (t *TeamsProvider) Send(ctx context.Context, notification *Notification) error {
    message := map[string]interface{}{
        "@type":      "MessageCard",
        "@context":   "https://schema.org/extensions",
        "themeColor": t.getColorForType(notification.Type),
        "summary":    notification.Title,
        "sections": []map[string]interface{}{
            {
                "activityTitle": notification.Title,
                "text":         notification.Message,
                "facts": []map[string]string{
                    {"name": "Movie", "value": notification.MovieTitle},
                    {"name": "Quality", "value": notification.Quality},
                },
            },
        },
    }

    payload, err := json.Marshal(message)
    if err != nil {
        return fmt.Errorf("failed to marshal Teams message: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, "POST", t.webhookURL, bytes.NewBuffer(payload))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := t.client.Do(req)
    if err != nil {
        return fmt.Errorf("failed to send Teams notification: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("Teams API returned status %d", resp.StatusCode)
    }

    return nil
}

func (t *TeamsProvider) Test(ctx context.Context, config map[string]interface{}) error {
    provider, err := NewTeamsProvider(config)
    if err != nil {
        return err
    }

    testNotification := &Notification{
        Type:       "test",
        Title:      "Radarr Test Notification",
        Message:    "This is a test notification from Radarr-go",
        MovieTitle: "Test Movie",
        Quality:    "1080p",
    }

    return provider.Send(ctx, testNotification)
}

func (t *TeamsProvider) GetConfigSchema() map[string]interface{} {
    return map[string]interface{}{
        "webhook_url": map[string]interface{}{
            "type":        "string",
            "required":    true,
            "description": "Microsoft Teams webhook URL",
            "placeholder": "https://outlook.office.com/webhook/...",
        },
        "mention_users": map[string]interface{}{
            "type":        "array",
            "required":    false,
            "description": "Users to mention in notifications",
            "items":       map[string]string{"type": "string"},
        },
    }
}

func (t *TeamsProvider) GetName() string {
    return "teams"
}

func (t *TeamsProvider) GetDescription() string {
    return "Send notifications to Microsoft Teams channels"
}
```

2. **Register Provider in Factory**

```go
// /internal/services/notifications/factory.go
func CreateProvider(providerType string, config map[string]interface{}) (NotificationProvider, error) {
    switch strings.ToLower(providerType) {
    case "discord":
        return NewDiscordProvider(config)
    case "slack":
        return NewSlackProvider(config)
    case "teams": // Add new provider
        return NewTeamsProvider(config)
    // ... other providers
    default:
        return nil, fmt.Errorf("unknown notification provider: %s", providerType)
    }
}
```

3. **Add Provider Tests**

```go
func TestTeamsProvider_Send(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

        var payload map[string]interface{}
        err := json.NewDecoder(r.Body).Decode(&payload)
        require.NoError(t, err)

        assert.Equal(t, "MessageCard", payload["@type"])
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    provider, err := NewTeamsProvider(map[string]interface{}{
        "webhook_url": server.URL,
    })
    require.NoError(t, err)

    notification := &Notification{
        Type:       "grab",
        Title:      "Movie Downloaded",
        Message:    "Test Movie has been downloaded",
        MovieTitle: "Test Movie",
    }

    err = provider.Send(context.Background(), notification)
    assert.NoError(t, err)
}
```

### Creating Custom Task Handlers

#### Task Handler Implementation

```go
// /internal/services/task_handlers.go
type CustomMaintenanceHandler struct {
    healthService *HealthService
    logger        *logger.Logger
}

func NewCustomMaintenanceHandler(healthService *HealthService) *CustomMaintenanceHandler {
    return &CustomMaintenanceHandler{
        healthService: healthService,
        logger:        logger.NewWith("handler", "custom_maintenance"),
    }
}

func (h *CustomMaintenanceHandler) Execute(
    ctx context.Context,
    task *models.TaskV2,
    updateProgress func(percent int, message string),
) error {
    h.logger.Infow("Starting custom maintenance task", "taskId", task.ID)
    updateProgress(0, "Starting maintenance")

    // Step 1: Clean old logs (25% complete)
    if err := h.cleanOldLogs(ctx); err != nil {
        return fmt.Errorf("failed to clean logs: %w", err)
    }
    updateProgress(25, "Old logs cleaned")

    // Step 2: Optimize database (50% complete)
    if err := h.optimizeDatabase(ctx); err != nil {
        return fmt.Errorf("failed to optimize database: %w", err)
    }
    updateProgress(50, "Database optimized")

    // Step 3: Clean temporary files (75% complete)
    if err := h.cleanTempFiles(ctx); err != nil {
        return fmt.Errorf("failed to clean temp files: %w", err)
    }
    updateProgress(75, "Temporary files cleaned")

    // Step 4: Generate health report (100% complete)
    if err := h.generateHealthReport(ctx); err != nil {
        return fmt.Errorf("failed to generate health report: %w", err)
    }
    updateProgress(100, "Maintenance completed")

    h.logger.Infow("Custom maintenance task completed", "taskId", task.ID)
    return nil
}

func (h *CustomMaintenanceHandler) GetName() string {
    return "CustomMaintenance"
}

func (h *CustomMaintenanceHandler) GetDescription() string {
    return "Performs comprehensive system maintenance including log cleanup, database optimization, and health reporting"
}
```

#### Register Handler in Container

```go
// In /internal/services/container.go NewContainer function
func NewContainer(db *database.Database, cfg *config.Config, logger *logger.Logger) *Container {
    // ... existing initialization ...

    // Register custom task handlers
    container.TaskService.RegisterHandler(NewCustomMaintenanceHandler(container.HealthService))

    return container
}
```

### Extending the Health Monitoring System

#### Custom Health Checker Implementation

```go
// /internal/services/health_checkers.go
type IndexerHealthChecker struct {
    indexerService *IndexerService
    logger         *logger.Logger
}

func NewIndexerHealthChecker(indexerService *IndexerService) *IndexerHealthChecker {
    return &IndexerHealthChecker{
        indexerService: indexerService,
        logger:         logger.NewWith("checker", "indexer"),
    }
}

func (c *IndexerHealthChecker) Name() string {
    return "Indexer Connectivity"
}

func (c *IndexerHealthChecker) Type() models.HealthCheckType {
    return models.HealthCheckTypeExternal
}

func (c *IndexerHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
    start := time.Now()
    execution := models.HealthCheckExecution{
        Type:      c.Type(),
        Source:    c.Name(),
        Timestamp: start,
    }

    indexers, err := c.indexerService.GetEnabledIndexers()
    if err != nil {
        execution.Status = models.HealthStatusCritical
        execution.Message = fmt.Sprintf("Failed to get indexers: %v", err)
        execution.Duration = time.Since(start)
        return execution
    }

    var issues []models.HealthIssue
    healthyCount := 0

    for _, indexer := range indexers {
        if err := c.testIndexer(ctx, indexer); err != nil {
            issues = append(issues, models.HealthIssue{
                Type:     c.Type(),
                Source:   c.Name(),
                Severity: models.HealthSeverityWarning,
                Message:  fmt.Sprintf("Indexer '%s' is not responding: %v", indexer.Name, err),
                Data: map[string]interface{}{
                    "indexer_id":   indexer.ID,
                    "indexer_name": indexer.Name,
                    "error":        err.Error(),
                },
            })
        } else {
            healthyCount++
        }
    }

    execution.Duration = time.Since(start)
    execution.Issues = issues

    if len(indexers) == 0 {
        execution.Status = models.HealthStatusWarning
        execution.Message = "No indexers configured"
    } else if healthyCount == 0 {
        execution.Status = models.HealthStatusCritical
        execution.Message = "All indexers are failing"
    } else if len(issues) > 0 {
        execution.Status = models.HealthStatusWarning
        execution.Message = fmt.Sprintf("%d of %d indexers failing", len(issues), len(indexers))
    } else {
        execution.Status = models.HealthStatusHealthy
        execution.Message = fmt.Sprintf("All %d indexers healthy", len(indexers))
    }

    return execution
}

func (c *IndexerHealthChecker) IsEnabled() bool {
    return true // Could be configurable
}

func (c *IndexerHealthChecker) GetInterval() time.Duration {
    return 10 * time.Minute // Check every 10 minutes
}

func (c *IndexerHealthChecker) testIndexer(ctx context.Context, indexer *models.Indexer) error {
    // Implementation depends on indexer type
    testCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()

    return c.indexerService.TestIndexer(testCtx, indexer.ID)
}
```

### Adding New API Endpoints Following Project Patterns

#### API Handler Implementation

```go
// /internal/api/statistics_handlers.go
func (s *Server) setupStatisticsRoutes(v3 *gin.RouterGroup) {
    statsRoutes := v3.Group("/statistics")
    statsRoutes.GET("/overview", s.handleGetStatisticsOverview)
    statsRoutes.GET("/movies", s.handleGetMovieStatistics)
    statsRoutes.GET("/quality", s.handleGetQualityStatistics)
    statsRoutes.GET("/performance", s.handleGetPerformanceStatistics)
}

func (s *Server) handleGetStatisticsOverview(c *gin.Context) {
    ctx := c.Request.Context()

    stats, err := s.services.StatisticsService.GetOverview(ctx)
    if err != nil {
        s.logger.Errorw("Failed to get statistics overview", "error", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get statistics"})
        return
    }

    c.JSON(http.StatusOK, stats)
}
```

#### Service Implementation

```go
// /internal/services/statistics_service.go
type StatisticsService struct {
    db     *database.Database
    logger *logger.Logger
}

type StatisticsOverview struct {
    Movies struct {
        Total     int `json:"total"`
        Available int `json:"available"`
        Missing   int `json:"missing"`
    } `json:"movies"`

    Storage struct {
        TotalSize int64 `json:"totalSize"`
        FreeSpace int64 `json:"freeSpace"`
    } `json:"storage"`

    Quality struct {
        Profiles []QualityProfileStats `json:"profiles"`
    } `json:"quality"`
}

func NewStatisticsService(db *database.Database, logger *logger.Logger) *StatisticsService {
    return &StatisticsService{
        db:     db,
        logger: logger,
    }
}

func (s *StatisticsService) GetOverview(ctx context.Context) (*StatisticsOverview, error) {
    var overview StatisticsOverview

    // Get movie statistics
    if err := s.db.GORM.WithContext(ctx).Model(&models.Movie{}).Count(&overview.Movies.Total).Error; err != nil {
        return nil, fmt.Errorf("failed to count total movies: %w", err)
    }

    if err := s.db.GORM.WithContext(ctx).Model(&models.Movie{}).
        Where("status = ?", "available").Count(&overview.Movies.Available).Error; err != nil {
        return nil, fmt.Errorf("failed to count available movies: %w", err)
    }

    overview.Movies.Missing = overview.Movies.Total - overview.Movies.Available

    // Get storage statistics
    // Implementation depends on storage backend

    return &overview, nil
}
```

### Database Migration Best Practices

#### Migration File Structure

```sql
-- /migrations/postgres/003_add_statistics_table.up.sql
CREATE TABLE IF NOT EXISTS movie_statistics (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
    download_count INTEGER DEFAULT 0,
    search_count INTEGER DEFAULT 0,
    last_searched_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(movie_id)
);

CREATE INDEX idx_movie_statistics_movie_id ON movie_statistics(movie_id);
CREATE INDEX idx_movie_statistics_last_searched ON movie_statistics(last_searched_at);

-- /migrations/postgres/003_add_statistics_table.down.sql
DROP INDEX IF EXISTS idx_movie_statistics_last_searched;
DROP INDEX IF EXISTS idx_movie_statistics_movie_id;
DROP TABLE IF EXISTS movie_statistics;
```

#### Migration Testing

```go
func TestMigration_003_StatisticsTable(t *testing.T) {
    db := setupTestDB(t)

    // Test migration up
    err := runMigration(db, "003_add_statistics_table.up.sql")
    require.NoError(t, err)

    // Verify table exists
    var count int
    err = db.GORM.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'movie_statistics'").Scan(&count).Error
    require.NoError(t, err)
    assert.Equal(t, 1, count)

    // Test migration down
    err = runMigration(db, "003_add_statistics_table.down.sql")
    require.NoError(t, err)

    // Verify table is dropped
    err = db.GORM.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'movie_statistics'").Scan(&count).Error
    require.NoError(t, err)
    assert.Equal(t, 0, count)
}
```

---

## Performance and Security Considerations

### Performance Optimization Guidelines

#### Database Query Optimization

```go
// Bad: N+1 query problem
func (s *MovieService) GetMoviesWithFiles() ([]*Movie, error) {
    movies, err := s.db.GORM.Find(&movies).Error
    if err != nil {
        return nil, err
    }

    for _, movie := range movies {
        // This creates N additional queries
        s.db.GORM.Where("movie_id = ?", movie.ID).Find(&movie.Files)
    }
    return movies, nil
}

// Good: Use eager loading
func (s *MovieService) GetMoviesWithFiles() ([]*Movie, error) {
    var movies []*Movie
    err := s.db.GORM.Preload("Files").Find(&movies).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get movies with files: %w", err)
    }
    return movies, nil
}

// Better: Use joins for performance-critical queries
func (s *MovieService) GetMoviesWithFileCount() ([]MovieWithFileCount, error) {
    var results []MovieWithFileCount
    err := s.db.GORM.Table("movies").
        Select("movies.*, COUNT(movie_files.id) as file_count").
        Joins("LEFT JOIN movie_files ON movies.id = movie_files.movie_id").
        Group("movies.id").
        Scan(&results).Error

    if err != nil {
        return nil, fmt.Errorf("failed to get movies with file count: %w", err)
    }
    return results, nil
}
```

#### Memory Management

```go
// Use object pools for frequently allocated objects
var moviePool = sync.Pool{
    New: func() interface{} {
        return &Movie{}
    },
}

func (s *MovieService) processMovies(movies []Movie) {
    for _, movie := range movies {
        // Get object from pool
        processedMovie := moviePool.Get().(*Movie)
        defer moviePool.Put(processedMovie)

        // Reset and reuse
        *processedMovie = movie

        // Process movie...
    }
}
```

#### Goroutine Management

```go
// Use worker pools instead of unlimited goroutines
func (s *SearchService) searchAllIndexers(ctx context.Context, query string) ([]*Release, error) {
    indexers, err := s.indexerService.GetEnabledIndexers()
    if err != nil {
        return nil, err
    }

    // Limit concurrent searches
    semaphore := make(chan struct{}, 5) // Max 5 concurrent searches
    results := make(chan []*Release, len(indexers))

    var wg sync.WaitGroup
    for _, indexer := range indexers {
        wg.Add(1)
        go func(idx *Indexer) {
            defer wg.Done()

            semaphore <- struct{}{} // Acquire
            defer func() { <-semaphore }() // Release

            releases, err := s.searchIndexer(ctx, idx, query)
            if err != nil {
                s.logger.Errorw("Indexer search failed", "indexer", idx.Name, "error", err)
                results <- nil
                return
            }
            results <- releases
        }(indexer)
    }

    // Close results channel when all goroutines complete
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    var allReleases []*Release
    for releases := range results {
        if releases != nil {
            allReleases = append(allReleases, releases...)
        }
    }

    return allReleases, nil
}
```

### Security Best Practices

#### Input Validation

```go
func (s *Server) handleCreateMovie(c *gin.Context) {
    var req CreateMovieRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
        return
    }

    // Validate input
    if err := s.validateCreateMovieRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Sanitize input
    req.Title = html.EscapeString(strings.TrimSpace(req.Title))

    // Process request...
}

func (s *Server) validateCreateMovieRequest(req *CreateMovieRequest) error {
    if strings.TrimSpace(req.Title) == "" {
        return fmt.Errorf("title is required")
    }

    if len(req.Title) > 255 {
        return fmt.Errorf("title too long (max 255 characters)")
    }

    if req.TMDbID <= 0 {
        return fmt.Errorf("valid TMDb ID is required")
    }

    return nil
}
```

#### SQL Injection Prevention

```go
// Always use parameterized queries
func (s *MovieService) SearchMoviesByTitle(title string, limit int) ([]*Movie, error) {
    var movies []*Movie

    // Good: Parameterized query
    err := s.db.GORM.Where("title ILIKE ?", "%"+title+"%").
        Limit(limit).
        Find(&movies).Error

    if err != nil {
        return nil, fmt.Errorf("failed to search movies: %w", err)
    }

    return movies, nil
}

// For raw queries, always use parameters
func (s *MovieService) GetMovieStatistics() (*MovieStats, error) {
    var stats MovieStats

    query := `
        SELECT
            COUNT(*) as total_movies,
            COUNT(CASE WHEN status = $1 THEN 1 END) as available_movies,
            AVG(rating) as average_rating
        FROM movies
        WHERE created_at >= $2
    `

    err := s.db.GORM.Raw(query, "available", time.Now().AddDate(0, -1, 0)).
        Scan(&stats).Error

    if err != nil {
        return nil, fmt.Errorf("failed to get movie statistics: %w", err)
    }

    return &stats, nil
}
```

#### Authentication and Authorization

```go
// Rate limiting middleware
func rateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(time.Minute), 60) // 60 requests per minute

    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// API key validation with timing attack prevention
func apiKeyMiddleware(validAPIKey string) gin.HandlerFunc {
    return func(c *gin.Context) {
        providedKey := c.GetHeader("X-API-Key")
        if providedKey == "" {
            providedKey = c.Query("apikey")
        }

        // Use constant-time comparison to prevent timing attacks
        if subtle.ConstantTimeCompare([]byte(providedKey), []byte(validAPIKey)) != 1 {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

This comprehensive developer guide provides new contributors with the knowledge needed to understand radarr-go's sophisticated architecture and contribute effectively while maintaining the project's high code quality standards.
