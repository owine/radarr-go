package testhelpers

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	// Database type constants for consistency with main database package
	postgresType = "postgres"
	mariadbType  = "mariadb"
)

// TestContext provides a comprehensive testing environment
type TestContext struct {
	T            *testing.T
	DB           *database.Database
	Logger       *logger.Logger
	Factory      *TestDataFactory
	SeedData     *SeedData
	TempDir      string
	Config       *config.Config
	CleanupFuncs []func()
}

// TestOptions configures test setup behavior
type TestOptions struct {
	DatabaseType    string        // postgres, mariadb, or auto
	SkipMigrations  bool          // Skip running migrations
	SkipCleanup     bool          // Skip cleanup (for debugging)
	IsolateDatabase bool          // Use isolated database schema
	Timeout         time.Duration // Test timeout
	TempDir         bool          // Create temporary directory
	LogLevel        string        // Log level for tests
}

// DefaultTestOptions returns sensible defaults for testing
func DefaultTestOptions() TestOptions {
	return TestOptions{
		DatabaseType:    "auto", // Auto-select available database
		SkipMigrations:  false,
		SkipCleanup:     false,
		IsolateDatabase: true,
		Timeout:         30 * time.Second,
		TempDir:         true,
		LogLevel:        "error",
	}
}

// NewTestContext creates a comprehensive test environment
func NewTestContext(t *testing.T, options TestOptions) *TestContext {
	t.Helper()

	// Set test timeout if specified
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		_, cancel = context.WithTimeout(context.Background(), options.Timeout)
		t.Cleanup(cancel)
	}

	ctx := &TestContext{
		T:            t,
		CleanupFuncs: make([]func(), 0),
	}

	// Determine database type
	dbType := options.DatabaseType
	if dbType == "auto" {
		dbType = selectBestAvailableDatabase(t)
	}

	// Create temporary directory if requested
	if options.TempDir {
		tempDir := t.TempDir()
		ctx.TempDir = tempDir
	}

	// Setup database
	ctx.setupDatabase(t, dbType, options)

	// Setup additional utilities
	ctx.Factory = NewTestDataFactory(ctx.DB.GORM)
	ctx.SeedData = NewSeedData(ctx.DB.GORM)

	// Register cleanup
	t.Cleanup(func() {
		ctx.Cleanup()
	})

	return ctx
}

// setupDatabase initializes the test database
func (ctx *TestContext) setupDatabase(t *testing.T, dbType string, options TestOptions) {
	t.Helper()

	var db *database.Database
	var log *logger.Logger

	if options.IsolateDatabase {
		// Use isolated database setup
		db, log = ctx.setupIsolatedDatabase(t, dbType, options)
	} else {
		// Use shared database setup
		db, log = SetupTestDatabaseWithManager(t, dbType)
		ctx.CleanupFuncs = append(ctx.CleanupFuncs, func() {
			CleanupTestDatabaseWithManager(t, db)
		})
	}

	ctx.DB = db
	ctx.Logger = log

	// Create config for reference
	databases := GetTestDatabases()
	for _, testDB := range databases {
		if testDB.Type == dbType {
			ctx.Config = &config.Config{
				Database: config.DatabaseConfig{
					Type:     testDB.Type,
					Host:     testDB.Host,
					Port:     testDB.Port,
					Database: testDB.Database,
					Username: testDB.Username,
					Password: testDB.Password,
				},
				Log: config.LogConfig{
					Level:  options.LogLevel,
					Format: "json",
				},
			}
			break
		}
	}
}

// setupIsolatedDatabase creates an isolated database schema for this test
func (ctx *TestContext) setupIsolatedDatabase(t *testing.T, dbType string,
	options TestOptions) (*database.Database, *logger.Logger) {
	t.Helper()

	// Generate unique schema name for this test
	schemaName := generateUniqueSchemaName(t)

	// Get base database configuration
	databases := GetTestDatabases()
	var baseTestDB TestDatabase
	for _, db := range databases {
		if db.Type == dbType {
			baseTestDB = db
			break
		}
	}

	if baseTestDB.Type == "" {
		t.Skipf("Test database %s not available", dbType)
	}

	// Create isolated schema/database
	if err := ctx.createIsolatedSchema(baseTestDB, schemaName); err != nil {
		t.Fatalf("Failed to create isolated schema: %v", err)
	}

	// Setup cleanup for isolated schema
	ctx.CleanupFuncs = append(ctx.CleanupFuncs, func() {
		_ = ctx.dropIsolatedSchema(baseTestDB, schemaName)
	})

	// Create database configuration for isolated schema
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     baseTestDB.Type,
			Host:     baseTestDB.Host,
			Port:     baseTestDB.Port,
			Database: schemaName, // Use isolated schema as database name
			Username: baseTestDB.Username,
			Password: baseTestDB.Password,
		},
		Log: config.LogConfig{
			Level:  options.LogLevel,
			Format: "json",
		},
	}

	// Create logger
	testLogger := logger.New(cfg.Log)

	// Create database connection to isolated schema
	db, err := database.New(&cfg.Database, testLogger)
	if err != nil {
		t.Fatalf("Failed to create isolated database connection: %v", err)
	}

	// Run migrations if not skipped
	if !options.SkipMigrations {
		if err := database.Migrate(db, testLogger); err != nil {
			t.Fatalf("Failed to run migrations on isolated schema: %v", err)
		}
	}

	return db, testLogger
}

// createIsolatedSchema creates a new schema/database for isolation
func (ctx *TestContext) createIsolatedSchema(baseDB TestDatabase, schemaName string) error {
	// This is a simplified implementation - in practice you might want more sophisticated schema creation
	switch baseDB.Type {
	case postgresType:
		return ctx.createPostgresSchema(baseDB, schemaName)
	case mariadbType, "mysql":
		return ctx.createMySQLDatabase(baseDB, schemaName)
	default:
		return fmt.Errorf("unsupported database type for schema isolation: %s", baseDB.Type)
	}
}

// createPostgresSchema creates a new PostgreSQL schema
func (ctx *TestContext) createPostgresSchema(baseDB TestDatabase, schemaName string) error {
	// For PostgreSQL, we can create a new database for full isolation
	// This requires connecting to the default postgres database first
	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		baseDB.Username, baseDB.Password, baseDB.Host, baseDB.Port)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return err
	}

	// Create new database
	createSQL := fmt.Sprintf("CREATE DATABASE %s", schemaName)
	if err := adminDB.Exec(createSQL).Error; err != nil {
		return err
	}

	// Close admin connection
	if sqlDB, err := adminDB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	return nil
}

// createMySQLDatabase creates a new MySQL database
func (ctx *TestContext) createMySQLDatabase(baseDB TestDatabase, databaseName string) error {
	// Similar to PostgreSQL but for MySQL
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		baseDB.Username, baseDB.Password, baseDB.Host, baseDB.Port)

	adminDB, err := gorm.Open(mysql.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return err
	}

	// Create new database
	createSQL := fmt.Sprintf("CREATE DATABASE %s", databaseName)
	if err := adminDB.Exec(createSQL).Error; err != nil {
		return err
	}

	// Close admin connection
	if sqlDB, err := adminDB.DB(); err == nil {
		_ = sqlDB.Close()
	}

	return nil
}

// dropIsolatedSchema removes the isolated schema/database
func (ctx *TestContext) dropIsolatedSchema(baseDB TestDatabase, schemaName string) error {
	switch baseDB.Type {
	case postgresType:
		return ctx.dropPostgresDatabase(baseDB, schemaName)
	case mariadbType, "mysql":
		return ctx.dropMySQLDatabase(baseDB, schemaName)
	default:
		return fmt.Errorf("unsupported database type for schema cleanup: %s", baseDB.Type)
	}
}

// dropPostgresDatabase drops a PostgreSQL database
func (ctx *TestContext) dropPostgresDatabase(baseDB TestDatabase, databaseName string) error {
	adminDSN := fmt.Sprintf("postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		baseDB.Username, baseDB.Password, baseDB.Host, baseDB.Port)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return err
	}
	defer func() {
		if sqlDB, err := adminDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	// Terminate connections to the database first
	terminateSQL := fmt.Sprintf(`
		SELECT pg_terminate_backend(pid)
		FROM pg_stat_activity
		WHERE datname = '%s' AND pid <> pg_backend_pid()`, databaseName)
	_ = adminDB.Exec(terminateSQL)

	// Drop database
	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s", databaseName)
	return adminDB.Exec(dropSQL).Error
}

// dropMySQLDatabase drops a MySQL database
func (ctx *TestContext) dropMySQLDatabase(baseDB TestDatabase, databaseName string) error {
	adminDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		baseDB.Username, baseDB.Password, baseDB.Host, baseDB.Port)

	adminDB, err := gorm.Open(mysql.Open(adminDSN), &gorm.Config{})
	if err != nil {
		return err
	}
	defer func() {
		if sqlDB, err := adminDB.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}()

	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS %s", databaseName)
	return adminDB.Exec(dropSQL).Error
}

// generateUniqueSchemaName creates a unique schema name for test isolation
func generateUniqueSchemaName(t *testing.T) string {
	// Create a unique name based on test name, timestamp, and random number
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, " ", "_")
	testName = strings.ToLower(testName)

	// Limit length and add randomness
	if len(testName) > 20 {
		testName = testName[:20]
	}

	timestamp := time.Now().Unix()
	// Use crypto/rand for secure random number generation
	randomNum, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		// Fallback to timestamp if crypto/rand fails
		randomNum = big.NewInt(timestamp % 10000)
	}

	return fmt.Sprintf("test_%s_%d_%d", testName, timestamp, randomNum.Int64())
}

// selectBestAvailableDatabase chooses the best available database for testing
func selectBestAvailableDatabase(t *testing.T) string {
	t.Helper()

	// Check environment preference first
	if envType := os.Getenv("RADARR_TEST_DATABASE_TYPE"); envType != "" {
		if IsTestDatabaseRunning(envType) {
			return envType
		}
	}

	// Check available databases in order of preference
	preferred := []string{postgresType, mariadbType}
	for _, dbType := range preferred {
		if IsTestDatabaseRunning(dbType) {
			return dbType
		}
	}

	// No databases available
	t.Skip("No test databases available. Start test containers with: make test-db-up")
	return ""
}

// RequireDatabase ensures a specific database type is available or skips the test
func RequireDatabase(t *testing.T, dbType string) {
	t.Helper()

	if !IsTestDatabaseRunning(dbType) {
		t.Skipf("Test database %s not available. Start test containers with: make test-db-up", dbType)
	}
}

// RequireAnyDatabase ensures at least one test database is available
func RequireAnyDatabase(t TestingT) {
	t.Helper()

	available := GetAvailableTestDatabases()
	if len(available) == 0 {
		t.Skip("No test databases available. Start test containers with: make test-db-up")
	}
}

// TestingT is an interface that describes the methods used by testing functions
// This allows the same helper functions to work with both *testing.T and *testing.B
type TestingT interface {
	Helper()
	Skip(args ...interface{})
	Skipf(format string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
}

// SkipInShortMode skips the test if running in short mode
func SkipInShortMode(t TestingT) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SkipInCI skips the test if running in CI environment
func SkipInCI(t *testing.T) {
	t.Helper()
	if isCI() {
		t.Skip("Skipping test in CI environment")
	}
}

// RequireCI skips the test if NOT running in CI environment
func RequireCI(t *testing.T) {
	t.Helper()
	if !isCI() {
		t.Skip("Test only runs in CI environment")
	}
}

// isCI checks if we're running in a CI environment
func isCI() bool {
	ci := os.Getenv("CI")
	return ci == "true" || ci == "1"
}

// GetTestDatabaseType returns the database type being used for tests
func GetTestDatabaseType() string {
	if envType := os.Getenv("RADARR_TEST_DATABASE_TYPE"); envType != "" {
		return envType
	}
	return postgresType // default
}

// IsRunningInParallel checks if tests are running in parallel
func IsRunningInParallel() bool {
	return runtime.GOMAXPROCS(0) > 1
}

// SetMaxParallelism sets the maximum number of tests that can run in parallel
func SetMaxParallelism(t *testing.T, maxParallelism int) {
	t.Helper()
	if IsRunningInParallel() {
		t.Parallel()
		// Create a buffered channel to limit concurrency
		sem := make(chan struct{}, maxParallelism)
		t.Cleanup(func() {
			<-sem
		})
		sem <- struct{}{}
	}
}

// Cleanup runs all registered cleanup functions
func (ctx *TestContext) Cleanup() {
	// Run cleanup functions in reverse order
	for i := len(ctx.CleanupFuncs) - 1; i >= 0; i-- {
		ctx.CleanupFuncs[i]()
	}

	// Clean up seed data and factory
	if ctx.SeedData != nil {
		ctx.SeedData.Cleanup()
	}
	if ctx.Factory != nil {
		ctx.Factory.Cleanup()
	}
}

// AddCleanup registers a cleanup function to be run at test end
func (ctx *TestContext) AddCleanup(cleanup func()) {
	ctx.CleanupFuncs = append(ctx.CleanupFuncs, cleanup)
}

// WithTransaction runs the provided function within a database transaction
// The transaction is rolled back automatically, providing perfect isolation
func (ctx *TestContext) WithTransaction(fn func(*gorm.DB)) {
	tx := ctx.DB.GORM.Begin()
	defer tx.Rollback()

	fn(tx)
}

// AssertDatabaseCount checks that a table has the expected number of records
func (ctx *TestContext) AssertDatabaseCount(tableName string, expected int64) {
	ctx.T.Helper()

	var count int64
	err := ctx.DB.GORM.Table(tableName).Count(&count).Error
	if err != nil {
		ctx.T.Fatalf("Failed to count records in table %s: %v", tableName, err)
	}

	if count != expected {
		ctx.T.Errorf("Expected %d records in table %s, got %d", expected, tableName, count)
	}
}

// AssertDatabaseEmpty checks that specified tables are empty
func (ctx *TestContext) AssertDatabaseEmpty(tableNames ...string) {
	ctx.T.Helper()

	for _, tableName := range tableNames {
		ctx.AssertDatabaseCount(tableName, 0)
	}
}
