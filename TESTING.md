# Testing Guide for Radarr Go

This document provides comprehensive guidance on testing in the Radarr Go project, including unit tests, integration tests, benchmark tests, and best practices.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Test Infrastructure](#test-infrastructure)
- [Database Testing](#database-testing)
- [Writing Tests](#writing-tests)
- [Running Tests](#running-tests)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Overview

Radarr Go uses a comprehensive testing strategy that includes:

- **Unit Tests**: Fast, isolated tests with no external dependencies
- **Integration Tests**: Tests that require database connections and external services
- **Benchmark Tests**: Performance testing to prevent regressions
- **Example Tests**: Documentation validation through runnable examples

### Test Infrastructure Features

- ✅ **Containerized Test Databases**: Isolated PostgreSQL and MariaDB containers
- ✅ **Automatic Database Management**: Setup, migration, and cleanup
- ✅ **Test Data Factories**: Consistent, reusable test data generation
- ✅ **Database Isolation**: Schema-level isolation for parallel testing
- ✅ **Performance Testing**: Comprehensive benchmark suite
- ✅ **CI/CD Integration**: Automated testing across multiple platforms

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Make

### Run All Tests

```bash
# Start test databases and run all tests
make test

# Or use the test runner directly
./scripts/test-runner.sh --mode all
```

### Run Specific Test Types

```bash
# Unit tests only (no database required)
make test-unit

# Integration tests only
make test-integration

# Benchmark tests
make test-bench

# Quick tests (short mode)
make test-quick
```

### Database-Specific Testing

```bash
# Test with PostgreSQL only
make test-postgres

# Test with MariaDB only
make test-mariadb

# Auto-select available database
./scripts/test-runner.sh --database auto
```

## Test Infrastructure

### Database Containers

The project uses Docker Compose to manage test databases:

- **PostgreSQL**: Available on `localhost:15432`
- **MariaDB**: Available on `localhost:13306`

Test databases are automatically started and stopped by the test runner.

#### Manual Database Management

```bash
# Start test databases
make test-db-up

# View database logs
make test-db-logs

# Stop and clean databases
make test-db-clean
```

### Test Helpers Package

The `internal/testhelpers` package provides comprehensive testing utilities:

#### TestContext

Provides a complete testing environment with automatic cleanup:

```go
func TestMyService(t *testing.T) {
    ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

    // ctx.DB - Database connection
    // ctx.Logger - Configured logger
    // ctx.Factory - Test data factory
    // ctx.SeedData - Data seeding utilities

    // Automatic cleanup when test ends
}
```

#### Test Options

Customize test behavior with `TestOptions`:

```go
options := testhelpers.TestOptions{
    DatabaseType:    "postgres",  // Force specific database
    IsolateDatabase: true,        // Use isolated schema
    TempDir:         true,        // Create temp directory
    LogLevel:        "debug",     // Set log level
    SkipMigrations:  false,       // Run migrations
    Timeout:         30 * time.Second,
}

ctx := testhelpers.NewTestContext(t, options)
```

#### Database Requirements

Skip tests gracefully when databases aren't available:

```go
func TestDatabaseFeature(t *testing.T) {
    testhelpers.RequireDatabase(t, "postgres")
    // Test continues only if PostgreSQL is available

    // Or require any database
    testhelpers.RequireAnyDatabase(t)
}
```

#### Test Conditions

Skip tests based on conditions:

```go
func TestLongRunning(t *testing.T) {
    testhelpers.SkipInShortMode(t)     // Skip with -short flag
    testhelpers.SkipInCI(t)            // Skip in CI environment
    testhelpers.RequireCI(t)           // Only run in CI
}
```

### Test Data Factory

Create consistent test data with the `TestDataFactory`:

```go
factory := testhelpers.NewTestDataFactory(db.GORM)
defer factory.Cleanup()

// Create test movie with default values
movie := factory.CreateMovie()

// Create with custom values
movie := factory.CreateMovie(func(m *models.Movie) {
    m.Title = "Custom Title"
    m.Year = 2024
})

// Create related data
profile := factory.CreateQualityProfile()
movieFile := factory.CreateMovieFile(movie.ID)
```

### Data Seeding

Use `SeedData` for comprehensive test datasets:

```go
seedData := testhelpers.NewSeedData(db.GORM)

// Basic dataset for integration tests
dataset, err := seedData.SeedBasicDataset()
// - 4 movies with different states
// - Quality profiles
// - Indexers, download clients, notifications
// - Tasks and movie files

// Performance dataset
dataset, err := seedData.SeedPerformanceDataset(1000)
// - 1000+ movies for performance testing
// - Associated movie files
// - Multiple quality profiles and indexers

// Stress test dataset
dataset, err := seedData.SeedStressTestDataset()
// - Comprehensive data for stress testing
// - Multiple configurations and tasks
```

## Database Testing

### Isolation Strategies

#### 1. Shared Database with Cleanup

Default approach - uses shared test database with data cleanup:

```go
func TestWithSharedDB(t *testing.T) {
    ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

    // Use shared database, automatic cleanup
    movie := ctx.Factory.CreateMovie()

    // Test operations
}
```

#### 2. Isolated Database Schema

Creates unique database/schema per test for perfect isolation:

```go
options := testhelpers.TestOptions{
    IsolateDatabase: true, // Each test gets its own schema
}
ctx := testhelpers.NewTestContext(t, options)
```

#### 3. Transactional Testing

Use database transactions for atomic test isolation:

```go
func TestWithTransaction(t *testing.T) {
    ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

    ctx.WithTransaction(func(tx *gorm.DB) {
        // All operations within transaction
        // Automatically rolled back at end
        movie := &models.Movie{/* ... */}
        tx.Create(movie)

        // Test operations
    })
}
```

### Database Assertions

Verify database state with helper methods:

```go
func TestDatabaseState(t *testing.T) {
    ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

    ctx.Factory.CreateMovie()
    ctx.Factory.CreateMovie()

    // Assert record counts
    ctx.AssertDatabaseCount("movies", 2)

    ctx.Factory.Cleanup()

    // Assert tables are empty
    ctx.AssertDatabaseEmpty("movies", "quality_profiles")
}
```

## Writing Tests

### Unit Tests

Focus on single components without external dependencies:

```go
package models

import (
    "testing"
    "github.com/radarr/radarr-go/internal/models"
)

func TestMovie_Validation(t *testing.T) {
    movie := &models.Movie{
        Title: "Test Movie",
        Year:  2024,
    }

    if err := movie.Validate(); err != nil {
        t.Errorf("Valid movie should not fail validation: %v", err)
    }
}
```

### Integration Tests

Test components that interact with the database:

```go
func TestMovieService_Create(t *testing.T) {
    testhelpers.RunWithTestDatabase(t, "postgres", func(t *testing.T, db *database.Database, log *logger.Logger) {
        factory := testhelpers.NewTestDataFactory(db.GORM)
        defer factory.Cleanup()

        // Create required dependencies
        profile := factory.CreateQualityProfile()

        // Test the service
        service := NewMovieService(db, log)
        movie := &models.Movie{
            Title:            "Integration Test Movie",
            TmdbID:           12345,
            QualityProfileID: profile.ID,
        }

        err := service.Create(movie)
        if err != nil {
            t.Fatalf("Failed to create movie: %v", err)
        }

        // Verify creation
        created, err := service.GetByID(movie.ID)
        if err != nil {
            t.Fatalf("Failed to retrieve movie: %v", err)
        }

        if created.Title != movie.Title {
            t.Errorf("Expected title %s, got %s", movie.Title, created.Title)
        }
    })
}
```

### Benchmark Tests

Performance testing to prevent regressions:

```go
func BenchmarkMovieService_Search(b *testing.B) {
    testhelpers.SkipInShortMode(b)
    testhelpers.RequireAnyDatabase(b)

    testhelpers.RunBenchmarkWithTestDatabase(b, testhelpers.GetTestDatabaseType(), func(b *testing.B, db *database.Database, log *logger.Logger) {
        // Setup benchmark data outside timing
        factory := testhelpers.NewTestDataFactory(db.GORM)
        defer factory.Cleanup()

        profile := factory.CreateQualityProfile()

        // Create test movies for searching
        for i := 0; i < 100; i++ {
            factory.CreateMovie(func(m *models.Movie) {
                m.TmdbID = 10000 + i
                m.Title = fmt.Sprintf("Benchmark Movie %d", i)
                m.QualityProfileID = profile.ID
            })
        }

        service := NewMovieService(db, log)

        b.ResetTimer()
        for i := 0; i < b.N; i++ {
            _, _ = service.Search("benchmark")
        }
    })
}
```

### Table-Driven Tests

Use table-driven tests for comprehensive coverage:

```go
func TestMovie_StatusValidation(t *testing.T) {
    tests := []struct {
        name     string
        status   models.MovieStatus
        expected bool
    }{
        {"valid announced", models.MovieStatusAnnounced, true},
        {"valid released", models.MovieStatusReleased, true},
        {"invalid empty", "", false},
        {"invalid unknown", "unknown", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            movie := &models.Movie{Status: tt.status}
            err := movie.Validate()

            if tt.expected && err != nil {
                t.Errorf("Expected valid status, got error: %v", err)
            }
            if !tt.expected && err == nil {
                t.Errorf("Expected validation error for status: %s", tt.status)
            }
        })
    }
}
```

## Running Tests

### Test Runner Script

The `scripts/test-runner.sh` provides comprehensive test management:

```bash
# Show all options
./scripts/test-runner.sh --help

# Run all tests with coverage
./scripts/test-runner.sh --mode all --coverage

# Integration tests with PostgreSQL, verbose output
./scripts/test-runner.sh --mode integration --database postgres --verbose

# Benchmark tests in parallel
./scripts/test-runner.sh --mode benchmark --benchmarks --parallel

# CI mode (parallel, coverage, validation)
./scripts/test-runner.sh --ci

# Keep databases running for debugging
./scripts/test-runner.sh --keep-db
```

### Makefile Targets

Use Make targets for common operations:

```bash
# Standard test commands
make test               # All tests
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-bench         # Benchmark tests
make test-coverage      # Tests with coverage report

# Database-specific tests
make test-postgres      # PostgreSQL only
make test-mariadb       # MariaDB only

# CI/Development
make test-ci            # CI mode
make test-quick         # Quick unit tests
```

### Manual Test Execution

Run Go tests directly with environment variables:

```bash
# Set database type
export RADARR_TEST_DATABASE_TYPE=postgres

# Run specific tests
go test -v ./internal/services -run TestMovieService

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./internal/services
```

### Parallel Testing

Enable parallel test execution:

```bash
# Using test runner
./scripts/test-runner.sh --parallel

# Using go test directly
go test -parallel=4 ./...
```

## CI/CD Integration

### GitHub Actions Integration

The testing infrastructure integrates with GitHub Actions:

```yaml
# Example workflow step
- name: Run tests
  run: |
    make test-ci
  env:
    CI: true
    RADARR_TEST_DATABASE_TYPE: postgres
```

### CI-Specific Features

- **Automatic database startup**: Containers start automatically in CI
- **Parallel execution**: Tests run in parallel for faster feedback
- **Coverage reporting**: Automatic coverage report generation
- **Multi-database testing**: Tests run against both PostgreSQL and MariaDB
- **Performance monitoring**: Benchmark tests track performance regressions

### Environment Variables

Configure testing behavior with environment variables:

```bash
# Database configuration
RADARR_TEST_DATABASE_TYPE=postgres    # postgres, mariadb, auto
POSTGRES_TEST_HOST=localhost
POSTGRES_TEST_PORT=15432
MARIADB_TEST_HOST=localhost
MARIADB_TEST_PORT=13306

# Test behavior
CI=true                               # Enable CI mode
RADARR_TEST_TIMEOUT=300               # Test timeout in seconds
RADARR_TEST_PARALLEL=4                # Parallel test count
```

## Troubleshooting

### Common Issues

#### 1. Database Connection Failures

```bash
# Check if databases are running
docker-compose -f docker-compose.test.yml ps

# View database logs
docker-compose -f docker-compose.test.yml logs postgres-test
docker-compose -f docker-compose.test.yml logs mariadb-test

# Restart databases
make test-db-clean
make test-db-up
```

#### 2. Port Conflicts

Test databases use non-standard ports to avoid conflicts:
- PostgreSQL: `15432` (instead of 5432)
- MariaDB: `13306` (instead of 3306)

#### 3. Permission Errors

Make sure test runner is executable:

```bash
chmod +x scripts/test-runner.sh
```

#### 4. Docker Issues

```bash
# Check Docker is running
docker info

# Check Docker Compose version
docker-compose version

# Clean up Docker resources
docker system prune
```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Verbose test output
./scripts/test-runner.sh --verbose

# Debug log level in tests
go test -v ./internal/services -args -log-level=debug
```

### Test Isolation Issues

If tests interfere with each other:

```bash
# Use database isolation
./scripts/test-runner.sh --mode integration

# Or run tests sequentially
go test -p=1 ./...
```

### Performance Issues

For slow tests:

```bash
# Run only quick tests
make test-quick

# Skip long-running tests
go test -short ./...

# Profile test performance
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/services
```

## Best Practices

### Test Organization

1. **Separate unit and integration tests** in the same package but different files
2. **Use descriptive test names** that explain what is being tested
3. **Group related tests** using subtests with `t.Run()`
4. **Clean up resources** using defer statements and cleanup functions

### Database Testing

1. **Use transactions** for fast test isolation when possible
2. **Create minimal test data** - only what's needed for the test
3. **Test database constraints** and relationships
4. **Verify both success and failure cases**

### Performance Testing

1. **Establish baselines** for benchmark tests
2. **Test with realistic data sizes**
3. **Monitor memory allocations** with `-benchmem`
4. **Run benchmarks in CI** to catch regressions

### Test Data

1. **Use factories** for consistent test data generation
2. **Avoid hardcoded values** that might conflict
3. **Create realistic but minimal** datasets
4. **Document complex test scenarios**

This testing infrastructure provides a solid foundation for maintaining high code quality and preventing regressions in the Radarr Go project. The combination of unit tests, integration tests, and performance tests ensures comprehensive coverage across all components.
