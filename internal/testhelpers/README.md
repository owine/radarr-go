# Test Helpers

This package provides comprehensive testing infrastructure for radarr-go, including containerized database testing, test data factories, and benchmarking support.

## Features

- **Containerized Test Databases**: Automatic setup of PostgreSQL and MariaDB test databases using Docker
- **Test Data Factories**: Consistent test data creation with realistic defaults
- **Cross-Database Testing**: Support for testing against both PostgreSQL and MariaDB
- **Benchmark Testing**: Performance testing with real database operations
- **Test Isolation**: Each test gets clean database state

## Usage

### Basic Test Setup

```go
func TestMyService(t *testing.T) {
    db, log := setupTestDB(t)
    defer cleanupTestDB(db)

    // Your test code here
    service := NewMyService(db, log)
    // ...
}
```

### Using Test Data Factories

```go
func TestMovieOperations(t *testing.T) {
    db, log := setupTestDB(t)
    defer cleanupTestDB(db)

    factory := testhelpers.NewTestDataFactory(db.GORM)
    defer factory.Cleanup()

    // Create test data with defaults
    profile := factory.CreateQualityProfile()
    movie := factory.CreateMovie(func(m *models.Movie) {
        m.QualityProfileID = profile.ID
        m.Title = "Custom Test Movie"
    })

    // Test operations
    service := NewMovieService(db, log)
    result, err := service.GetByID(context.Background(), movie.ID)
    assert.NoError(t, err)
    assert.Equal(t, movie.Title, result.Title)
}
```

### Benchmark Testing

```go
func BenchmarkMyOperation(b *testing.B) {
    db, log := setupTestDBForBenchmark(b)
    defer cleanupTestDB(db)

    factory := testhelpers.NewTestDataFactory(db.GORM)
    defer factory.Cleanup()

    // Setup test data
    // ...

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Operation to benchmark
    }
}
```

## Test Database Management

The test infrastructure automatically manages PostgreSQL and MariaDB test databases using Docker containers.

### Starting Test Databases

```bash
# Start both test databases
make test-db-up

# Check database logs
make test-db-logs

# Stop test databases
make test-db-down

# Clean up test databases and volumes
make test-db-clean
```

### Running Tests

```bash
# Run all tests (automatically starts databases)
make test

# Run tests with specific database
make test-postgres
make test-mariadb

# Run benchmark tests
make test-bench

# Run unit tests only (no database)
make test-unit

# Run tests with coverage
make test-coverage
```

## Environment Variables

- `RADARR_TEST_DATABASE_TYPE`: Preferred database type (`postgres` or `mariadb`)
- `POSTGRES_TEST_HOST`: PostgreSQL test host (default: localhost)
- `POSTGRES_TEST_PORT`: PostgreSQL test port (default: 15432)
- `MARIADB_TEST_HOST`: MariaDB test host (default: localhost)
- `MARIADB_TEST_PORT`: MariaDB test port (default: 13306)

## Test Data Factories

The `TestDataFactory` provides methods to create realistic test data:

### Available Factories

- `CreateMovie()`: Creates a movie with realistic metadata
- `CreateMovieFile()`: Creates a movie file with media info
- `CreateQualityProfile()`: Creates a quality profile with multiple quality items
- `CreateIndexer()`: Creates a configured indexer
- `CreateDownloadClient()`: Creates a download client
- `CreateNotification()`: Creates a notification configuration
- `CreateTask()`: Creates a scheduled task

### Customizing Test Data

All factory methods accept override functions to customize the created data:

```go
movie := factory.CreateMovie(func(m *models.Movie) {
    m.Title = "My Custom Movie"
    m.Year = 2024
    m.Monitored = false
})
```

## Database Connection Management

The test helpers automatically:

- Wait for databases to be ready before running tests
- Run database migrations
- Clean up test data between tests
- Handle connection pooling and timeouts
- Fall back to alternate database if primary is unavailable

## Performance Considerations

- Test databases use optimized settings for fast testing
- Data is automatically cleaned up after each test
- Connection pooling reduces setup overhead
- Benchmark tests include proper timer management

## Troubleshooting

### Test Databases Not Available

If tests are being skipped due to unavailable databases:

1. Start test databases: `make test-db-up`
2. Check container status: `docker ps`
3. Check logs: `make test-db-logs`
4. Try cleaning up: `make test-db-clean && make test-db-up`

### Connection Timeouts

Test databases have health checks and the helpers wait up to 30 seconds for readiness. If tests still fail:

1. Check Docker daemon is running
2. Ensure ports 15432 (PostgreSQL) and 13306 (MariaDB) are not in use
3. Check available memory for containers

### Migration Errors

If migrations fail during test setup:

1. Ensure migrations are in the correct format
2. Check database permissions
3. Clean test databases and restart: `make test-db-clean && make test-db-up`
