package services

import (
	"testing"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
)

const ciEnvTrue = "true"

// setupTestDB creates a test database and logger for testing
func setupTestDB(t *testing.T) (*database.Database, *logger.Logger) {
	// Create test config - skip tests if no database available
	t.Skip("Database tests require PostgreSQL or MariaDB setup - skipping")

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Type:     "postgres",
			Host:     "localhost",
			Port:     5432,
			Database: "radarr_test",
			Username: "postgres",
			Password: "password",
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

	return db, testLogger
}

// cleanupTestDB closes the test database
func cleanupTestDB(db *database.Database) {
	sqlDB, err := db.GORM.DB()
	if err == nil {
		_ = sqlDB.Close()
	}
}
