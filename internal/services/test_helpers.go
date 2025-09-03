package services

import (
	"os"
	"testing"

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
