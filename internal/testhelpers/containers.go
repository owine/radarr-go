// Package testhelpers provides utilities for setting up and managing test databases,
// including containerized PostgreSQL and MariaDB instances for integration testing.
package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"

	// Database drivers required for test database connections
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// Test database connection timeouts
	defaultConnectionTimeout  = 30 * time.Second
	defaultHealthCheckRetries = 30
	healthCheckInterval       = 1 * time.Second

	// Database type constants
	dbTypePostgres = "postgres"
	dbTypeMariaDB  = "mariadb"
	dbTypeMySQL    = "mysql"
)

// TestDatabase represents a test database configuration
type TestDatabase struct {
	Type     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	DSN      string
}

// GetTestDatabases returns available test database configurations
func GetTestDatabases() []TestDatabase {
	var databases []TestDatabase

	// Check for PostgreSQL test database
	if pgHost := os.Getenv("POSTGRES_TEST_HOST"); pgHost != "" || isTestContainerRunning("postgres-test") {
		host := pgHost
		if host == "" {
			host = "localhost"
		}

		port := 15432
		if portStr := os.Getenv("POSTGRES_TEST_PORT"); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				port = p
			}
		}

		databases = append(databases, TestDatabase{
			Type:     dbTypePostgres,
			Host:     host,
			Port:     port,
			Database: "radarr_test",
			Username: "radarr_test",
			Password: "test_password",
			DSN:      fmt.Sprintf("postgres://radarr_test:test_password@%s:%d/radarr_test?sslmode=disable", host, port),
		})
	}

	// Check for MariaDB test database
	if mariaHost := os.Getenv("MARIADB_TEST_HOST"); mariaHost != "" || isTestContainerRunning("mariadb-test") {
		host := mariaHost
		if host == "" {
			host = "localhost"
		}

		port := 13306
		if portStr := os.Getenv("MARIADB_TEST_PORT"); portStr != "" {
			if p, err := strconv.Atoi(portStr); err == nil {
				port = p
			}
		}

		databases = append(databases, TestDatabase{
			Type:     dbTypeMariaDB,
			Host:     host,
			Port:     port,
			Database: "radarr_test",
			Username: "radarr_test",
			Password: "test_password",
			DSN: fmt.Sprintf(
				"radarr_test:test_password@tcp(%s:%d)/radarr_test?charset=utf8mb4&parseTime=True&loc=Local",
				host, port,
			),
		})
	}

	return databases
}

// SetupTestDatabase creates a test database connection and runs migrations
func SetupTestDatabase(t *testing.T, dbType string) (*database.Database, *logger.Logger) {
	t.Helper()

	databases := GetTestDatabases()

	var testDB *TestDatabase
	for _, db := range databases {
		if db.Type == dbType {
			testDB = &db
			break
		}
	}

	if testDB == nil {
		t.Skipf("Test database %s not available. Start test containers with: make test-db-up", dbType)
		return nil, nil
	}

	// Wait for database to be ready
	if !waitForDatabase(t, *testDB) {
		t.Fatalf("Test database %s not ready after timeout", dbType)
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
		t.Fatalf("Failed to create test database connection: %v", err)
	}

	// Run migrations
	err = database.Migrate(db, testLogger)
	if err != nil {
		t.Fatalf("Failed to run database migrations: %v", err)
	}

	// Clean up any existing test data
	cleanupTestData(t, db, testDB.Type)

	return db, testLogger
}

// CleanupTestDatabase closes the database connection and cleans up test data
func CleanupTestDatabase(t *testing.T, db *database.Database) {
	t.Helper()

	if db != nil {
		if sqlDB, err := db.GORM.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

// waitForDatabase waits for the test database to be ready
func waitForDatabase(t *testing.T, testDB TestDatabase) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), defaultConnectionTimeout)
	defer cancel()

	var driverName, dsn string
	switch testDB.Type {
	case dbTypePostgres:
		driverName = "pgx"
		dsn = testDB.DSN
	case dbTypeMariaDB, dbTypeMySQL:
		driverName = dbTypeMySQL
		dsn = testDB.DSN
	default:
		t.Fatalf("Unsupported database type: %s", testDB.Type)
	}

	for i := 0; i < defaultHealthCheckRetries; i++ {
		select {
		case <-ctx.Done():
			t.Logf("Database connection timeout for %s", testDB.Type)
			return false
		default:
		}

		db, err := sql.Open(driverName, dsn)
		if err == nil {
			if err := db.PingContext(ctx); err == nil {
				if closeErr := db.Close(); closeErr != nil {
					t.Logf("Warning: failed to close database connection: %v", closeErr)
				}
				t.Logf("Test database %s is ready", testDB.Type)
				return true
			}
			if closeErr := db.Close(); closeErr != nil {
				t.Logf("Warning: failed to close database connection: %v", closeErr)
			}
		}

		t.Logf("Waiting for test database %s to be ready (attempt %d/%d)...", testDB.Type, i+1, defaultHealthCheckRetries)
		time.Sleep(healthCheckInterval)
	}

	return false
}

// cleanupTestData removes all test data from the database
func cleanupTestData(t *testing.T, db *database.Database, dbType string) {
	t.Helper()

	// List of tables to clean up (in order to respect foreign key constraints)
	tables := []string{
		"movie_files",
		"movies",
		"quality_profiles",
		"indexers",
		"download_clients",
		"notifications",
		"health_checks",
		"health_issues",
		"tasks",
		"collections",
		"wanted_missing",
		"wanted_cutoff_unmet",
	}

	for _, table := range tables {
		var query string
		switch dbType {
		case dbTypePostgres:
			query = fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		case dbTypeMariaDB, dbTypeMySQL:
			query = fmt.Sprintf("TRUNCATE TABLE %s", table)
		}

		result := db.GORM.Exec(query)
		if result.Error != nil {
			// Log the error but don't fail the test - table might not exist yet
			t.Logf("Warning: Could not clean table %s: %v", table, result.Error)
		}
	}
}

// isTestContainerRunning checks if a test container is running
func isTestContainerRunning(containerName string) bool {
	// Try to connect to the default test ports to see if containers are running
	switch containerName {
	case "postgres-test":
		return isPortOpen("localhost", 15432)
	case "mariadb-test":
		return isPortOpen("localhost", 13306)
	}
	return false
}

// isPortOpen checks if a TCP port is open
func isPortOpen(host string, port int) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return false
	}
	if closeErr := conn.Close(); closeErr != nil {
		// Log error but still return true since we successfully connected
		return true
	}
	return true
}
