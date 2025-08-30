package services

import (
	"os"
	"testing"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/testhelpers"
)

const (
	ciEnvTrue = "true"

	// Database type constants for services package
	serviceDbTypePostgres = "postgres"
	serviceDbTypeMariaDB  = "mariadb"
	serviceDbTypeMySQL    = "mysql"
)

// setupTestDB creates a test database and logger for testing
// It now uses containerized test databases instead of requiring manual setup
func setupTestDB(t *testing.T) (*database.Database, *logger.Logger) {
	t.Helper()

	// Check if a specific database type is requested
	dbType := os.Getenv("RADARR_TEST_DATABASE_TYPE")
	if dbType == "" {
		// Default to PostgreSQL for tests
		dbType = serviceDbTypePostgres
	}

	// Try to set up the requested database type first
	db, log := testhelpers.SetupTestDatabase(t, dbType)
	if db != nil && log != nil {
		return db, log
	}

	// If the requested type failed, try the other database type
	var fallbackType string
	switch dbType {
	case serviceDbTypePostgres:
		fallbackType = serviceDbTypeMariaDB
	case serviceDbTypeMariaDB, serviceDbTypeMySQL:
		fallbackType = serviceDbTypePostgres
	default:
		fallbackType = serviceDbTypePostgres
	}

	db, log = testhelpers.SetupTestDatabase(t, fallbackType)
	if db != nil && log != nil {
		t.Logf("Using fallback database type: %s", fallbackType)
		return db, log
	}

	// If both database types failed, skip the test
	t.Skip("No test databases available. Start test containers with: make test-db-up")
	return nil, nil
}

// cleanupTestDB closes the test database
func cleanupTestDB(db *database.Database) {
	if db != nil {
		testhelpers.CleanupTestDatabase(nil, db)
	}
}

// setupTestDBForBenchmark creates a test database for benchmark tests
func setupTestDBForBenchmark(b *testing.B) (*database.Database, *logger.Logger) {
	b.Helper()

	// Check if a specific database type is requested
	dbType := os.Getenv("RADARR_TEST_DATABASE_TYPE")
	if dbType == "" {
		// Default to PostgreSQL for benchmarks
		dbType = serviceDbTypePostgres
	}

	// Try to set up the requested database type first
	db, log := setupTestDatabaseForBenchmark(b, dbType)
	if db != nil && log != nil {
		return db, log
	}

	// If the requested type failed, try the other database type
	var fallbackType string
	switch dbType {
	case serviceDbTypePostgres:
		fallbackType = serviceDbTypeMariaDB
	case serviceDbTypeMariaDB, serviceDbTypeMySQL:
		fallbackType = serviceDbTypePostgres
	default:
		fallbackType = serviceDbTypePostgres
	}

	db, log = setupTestDatabaseForBenchmark(b, fallbackType)
	if db != nil && log != nil {
		b.Logf("Using fallback database type: %s", fallbackType)
		return db, log
	}

	// If both database types failed, skip the benchmark
	b.Skip("No test databases available. Start test containers with: make test-db-up")
	return nil, nil
}

// setupTestDatabaseForBenchmark is a helper function specifically for benchmark testing
func setupTestDatabaseForBenchmark(b *testing.B, dbType string) (*database.Database, *logger.Logger) {
	b.Helper()

	databases := testhelpers.GetTestDatabases()

	var testDB *testhelpers.TestDatabase
	for _, db := range databases {
		if db.Type == dbType {
			testDB = &db
			break
		}
	}

	if testDB == nil {
		b.Skipf("Test database %s not available. Start test containers with: make test-db-up", dbType)
		return nil, nil
	}

	// Simple wait approach for benchmarks
	if !waitForDatabaseSimple() {
		b.Fatalf("Test database %s not ready after timeout", dbType)
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     testDB.Type,
			Host:     testDB.Host,
			Port:     testDB.Port,
			Database: testDB.Database,
			Username: testDB.Username,
			Password: testDB.Password,
		},
		Log: config.LogConfig{
			Level:  "error", // Reduce log noise during tests
			Format: "json",
		},
	}

	// Create logger
	testLogger := logger.New(cfg.Log)

	// Create database connection
	db, err := database.New(&cfg.Database, testLogger)
	if err != nil {
		b.Fatalf("Failed to create test database connection: %v", err)
	}

	// Run migrations
	err = database.Migrate(db, testLogger)
	if err != nil {
		b.Fatalf("Failed to run database migrations: %v", err)
	}

	return db, testLogger
}

// waitForDatabaseSimple is a simple database ready check without fancy interfaces
func waitForDatabaseSimple() bool {
	// For benchmarks, just assume databases are ready after starting containers
	return true
}
