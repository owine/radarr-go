package services

import (
	"os"
	"strconv"
	"testing"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
)

const ciEnvTrue = "true"

// setupTestDB creates a test database and logger for testing
func setupTestDB(t *testing.T) (*database.Database, *logger.Logger) {
	// Check if database environment variables are set (CI environment)
	dbType := os.Getenv("RADARR_DATABASE_TYPE")
	dbHost := os.Getenv("RADARR_DATABASE_HOST")
	dbPortStr := os.Getenv("RADARR_DATABASE_PORT")
	dbDatabase := os.Getenv("RADARR_DATABASE_DATABASE")
	dbUsername := os.Getenv("RADARR_DATABASE_USERNAME")
	dbPassword := os.Getenv("RADARR_DATABASE_PASSWORD")

	// If no database environment variables are set, skip database tests
	if dbType == "" || dbHost == "" || dbDatabase == "" || dbUsername == "" {
		t.Skip("Database tests require environment variables (RADARR_DATABASE_TYPE, RADARR_DATABASE_HOST, etc.) - skipping")
	}

	// Parse database port
	var dbPort int
	if dbPortStr != "" {
		var err error
		dbPort, err = strconv.Atoi(dbPortStr)
		if err != nil {
			t.Fatalf("Invalid database port: %v", err)
		}
	} else {
		// Default ports based on database type
		switch dbType {
		case "postgres":
			dbPort = 5432
		case "mariadb", "mysql":
			dbPort = 3306
		default:
			t.Fatalf("Unknown database type: %s", dbType)
		}
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     dbType,
			Host:     dbHost,
			Port:     dbPort,
			Database: dbDatabase,
			Username: dbUsername,
			Password: dbPassword,
		},
		Log: config.LogConfig{
			Level:  "error", // Reduce log noise during tests
			Format: "json",
		},
	}

	// Create logger
	testLogger := logger.New(cfg.Log)

	// Create database
	db, err := database.New(&cfg.Database, testLogger)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Run database migrations to create tables
	err = database.Migrate(db, testLogger)
	if err != nil {
		t.Fatalf("Failed to run database migrations: %v", err)
	}

	return db, testLogger
}

// cleanupTestDB closes the test database
func cleanupTestDB(db *database.Database) {
	sqlDB, err := db.GORM.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
