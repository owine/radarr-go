// Package testhelpers provides utilities for setting up and managing test databases,
// including containerized PostgreSQL and MariaDB instances for integration testing.
package testhelpers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"

	// Database drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	// Container management timeouts
	containerStartTimeout = 60 * time.Second
	containerStopTimeout  = 30 * time.Second
	healthCheckTimeout    = 45 * time.Second
	dbHealthCheckInterval = 2 * time.Second

	// Default container configurations
	postgresTestPort = 15432
	mariadbTestPort  = 13306

	// Docker compose file path
	testComposeFile = "docker-compose.test.yml"
)

// DatabaseManager handles test database lifecycle management
type DatabaseManager struct {
	mu              sync.RWMutex
	runningServices map[string]bool
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewDatabaseManager creates a new database manager instance
func NewDatabaseManager() *DatabaseManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &DatabaseManager{
		runningServices: make(map[string]bool),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// Global database manager instance
var globalDBManager = NewDatabaseManager()

// StartTestDatabases starts the test database containers if they're not already running
func StartTestDatabases() error {
	return globalDBManager.StartDatabases()
}

// StopTestDatabases stops all test database containers
func StopTestDatabases() error {
	return globalDBManager.StopDatabases()
}

// CleanTestDatabases stops containers and removes volumes
func CleanTestDatabases() error {
	return globalDBManager.CleanDatabases()
}

// IsTestDatabaseRunning checks if a specific test database is running
func IsTestDatabaseRunning(dbType string) bool {
	return globalDBManager.IsServiceRunning(dbType)
}

// WaitForTestDatabase waits for a specific test database to be ready
func WaitForTestDatabase(dbType string, timeout time.Duration) error {
	return globalDBManager.WaitForDatabase(dbType, timeout)
}

// StartDatabases starts all test database containers
func (dm *DatabaseManager) StartDatabases() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	// Check if Docker is available
	if !isDockerAvailable() {
		return fmt.Errorf("Docker is not available or not running")
	}

	// Check if docker-compose.test.yml exists
	if _, err := os.Stat(testComposeFile); os.IsNotExist(err) {
		return fmt.Errorf("test compose file %s not found", testComposeFile)
	}

	// Start the database services
	cmd := exec.CommandContext(dm.ctx, "docker-compose", "-f", testComposeFile, "up", "-d", "postgres-test", "mariadb-test")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start test databases: %w", err)
	}

	// Wait for services to be healthy
	services := []string{"postgres-test", "mariadb-test"}
	for _, service := range services {
		if err := dm.waitForServiceHealthy(service); err != nil {
			log.Printf("Warning: Service %s may not be fully ready: %v", service, err)
		}
		dm.runningServices[service] = true
	}

	log.Println("Test databases started successfully")
	return nil
}

// StopDatabases stops all test database containers
func (dm *DatabaseManager) StopDatabases() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if !isDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	cmd := exec.CommandContext(dm.ctx, "docker-compose", "-f", testComposeFile, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop test databases: %w", err)
	}

	// Clear running services
	dm.runningServices = make(map[string]bool)

	log.Println("Test databases stopped successfully")
	return nil
}

// CleanDatabases stops containers and removes volumes
func (dm *DatabaseManager) CleanDatabases() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()

	if !isDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	cmd := exec.CommandContext(dm.ctx, "docker-compose", "-f", testComposeFile, "down", "-v", "--remove-orphans")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean test databases: %w", err)
	}

	// Clear running services
	dm.runningServices = make(map[string]bool)

	log.Println("Test databases cleaned successfully")
	return nil
}

// IsServiceRunning checks if a service is marked as running
func (dm *DatabaseManager) IsServiceRunning(serviceName string) bool {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.runningServices[serviceName]
}

// WaitForDatabase waits for a database to be ready for connections
func (dm *DatabaseManager) WaitForDatabase(dbType string, timeout time.Duration) error {
	var serviceName string
	var port int

	switch dbType {
	case "postgres":
		serviceName = "postgres-test"
		port = postgresTestPort
	case "mariadb", "mysql":
		serviceName = "mariadb-test"
		port = mariadbTestPort
	default:
		return fmt.Errorf("unsupported database type: %s", dbType)
	}

	// Check if service is running
	if !dm.IsServiceRunning(serviceName) {
		return fmt.Errorf("service %s is not running", serviceName)
	}

	ctx, cancel := context.WithTimeout(dm.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(dbHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for %s database to be ready", dbType)
		case <-ticker.C:
			if dm.isDatabaseReady(dbType, port) {
				return nil
			}
		}
	}
}

// waitForServiceHealthy waits for a Docker service to report as healthy
func (dm *DatabaseManager) waitForServiceHealthy(serviceName string) error {
	ctx, cancel := context.WithTimeout(dm.ctx, healthCheckTimeout)
	defer cancel()

	ticker := time.NewTicker(dbHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service %s to be healthy", serviceName)
		case <-ticker.C:
			if dm.isServiceHealthy(serviceName) {
				return nil
			}
		}
	}
}

// isServiceHealthy checks if a Docker service is healthy
func (dm *DatabaseManager) isServiceHealthy(serviceName string) bool {
	cmd := exec.CommandContext(dm.ctx, "docker-compose", "-f", testComposeFile, "ps", "-q", serviceName)
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	if len(output) == 0 {
		return false
	}

	// Check container health status
	containerID := string(output)
	containerID = containerID[:len(containerID)-1] // Remove newline

	healthCmd := exec.CommandContext(dm.ctx, "docker", "inspect", "--format={{.State.Health.Status}}", containerID)
	healthOutput, err := healthCmd.Output()
	if err != nil {
		// If no health check, assume it's healthy if running
		return true
	}

	healthStatus := string(healthOutput)
	healthStatus = healthStatus[:len(healthStatus)-1] // Remove newline
	return healthStatus == "healthy"
}

// isDatabaseReady checks if database is ready by attempting a connection
func (dm *DatabaseManager) isDatabaseReady(dbType string, port int) bool {
	var driverName, dsn string

	switch dbType {
	case "postgres":
		driverName = "pgx"
		dsn = fmt.Sprintf("postgres://radarr_test:test_password@localhost:%d/radarr_test?sslmode=disable", port)
	case "mariadb", "mysql":
		driverName = "mysql"
		dsn = fmt.Sprintf("radarr_test:test_password@tcp(localhost:%d)/radarr_test?charset=utf8mb4&parseTime=True&loc=Local", port)
	default:
		return false
	}

	ctx, cancel := context.WithTimeout(dm.ctx, 5*time.Second)
	defer cancel()

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close test database connection: %v", closeErr)
		}
	}()

	return db.PingContext(ctx) == nil
}

// isDockerAvailable checks if Docker is installed and running
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

// SetupTestDatabaseWithManager creates a test database connection using the database manager
func SetupTestDatabaseWithManager(t *testing.T, dbType string) (*database.Database, *logger.Logger) {
	t.Helper()

	// Start databases if not already running
	if !globalDBManager.IsServiceRunning(dbType + "-test") {
		if err := globalDBManager.StartDatabases(); err != nil {
			t.Skipf("Could not start test databases: %v", err)
		}
	}

	// Wait for the specific database to be ready
	if err := globalDBManager.WaitForDatabase(dbType, 30*time.Second); err != nil {
		t.Skipf("Test database %s not ready: %v", dbType, err)
	}

	// Get database configuration
	var testDB TestDatabase
	databases := GetTestDatabases()
	for _, db := range databases {
		if db.Type == dbType {
			testDB = db
			break
		}
	}

	if testDB.Type == "" {
		t.Skipf("Test database configuration for %s not found", dbType)
	}

	// Create configuration
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

// CleanupTestDatabaseWithManager closes the database connection properly
func CleanupTestDatabaseWithManager(t *testing.T, db *database.Database) {
	t.Helper()

	if db != nil {
		// Clean up test data before closing
		if db.GORM != nil {
			// Use the existing cleanup function
			CleanupTestDatabase(t, db)
		}
	}
}

// GetAvailableTestDatabases returns a list of currently available test databases
func GetAvailableTestDatabases() []string {
	var available []string

	databases := GetTestDatabases()
	for _, db := range databases {
		serviceName := db.Type + "-test"
		if globalDBManager.IsServiceRunning(serviceName) {
			available = append(available, db.Type)
		}
	}

	return available
}

// RunWithTestDatabase is a helper function that ensures a test database is available
// before running the test function, and cleans up afterward
func RunWithTestDatabase(t *testing.T, dbType string, testFunc func(*testing.T, *database.Database, *logger.Logger)) {
	t.Helper()

	db, log := SetupTestDatabaseWithManager(t, dbType)
	defer CleanupTestDatabaseWithManager(t, db)

	testFunc(t, db, log)
}

// RunBenchmarkWithTestDatabase is like RunWithTestDatabase but for benchmarks
func RunBenchmarkWithTestDatabase(b *testing.B, dbType string, benchFunc func(*testing.B, *database.Database, *logger.Logger)) {
	b.Helper()

	// For benchmarks, we need to set up the database outside the benchmark timing
	if !globalDBManager.IsServiceRunning(dbType + "-test") {
		if err := globalDBManager.StartDatabases(); err != nil {
			b.Skipf("Could not start test databases: %v", err)
		}
	}

	if err := globalDBManager.WaitForDatabase(dbType, 30*time.Second); err != nil {
		b.Skipf("Test database %s not ready: %v", dbType, err)
	}

	// Get database configuration
	var testDB TestDatabase
	databases := GetTestDatabases()
	for _, db := range databases {
		if db.Type == dbType {
			testDB = db
			break
		}
	}

	if testDB.Type == "" {
		b.Skipf("Test database configuration for %s not found", dbType)
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
			Level:  "error",
			Format: "json",
		},
	}

	testLogger := logger.New(cfg.Log)
	db, err := database.New(&cfg.Database, testLogger)
	if err != nil {
		b.Fatalf("Failed to create test database connection: %v", err)
	}

	if err := database.Migrate(db, testLogger); err != nil {
		b.Fatalf("Failed to run database migrations: %v", err)
	}

	defer func() {
		if db != nil {
			if sqlDB, err := db.GORM.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	}()

	benchFunc(b, db, testLogger)
}
