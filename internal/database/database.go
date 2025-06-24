// Package database provides database connection and migration functionality.
package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	// Import for golang-migrate file source support
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// Import for PostgreSQL driver support
	_ "github.com/lib/pq"
	// Import for MariaDB/MySQL driver support
	_ "github.com/go-sql-driver/mysql"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	gormMariaDB "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Database provides access to the Radarr database
type Database struct {
	DB   *sqlx.DB
	GORM *gorm.DB
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig, _ *logger.Logger) (*Database, error) {
	var db *sqlx.DB
	var gormDB *gorm.DB
	var err error

	switch cfg.Type {
	case "postgres", "postgresql":
		connectionString := buildPostgresConnectionString(cfg)
		db, err = sqlx.Connect("postgres", connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}

		gormDB, err = gorm.Open(gormPostgres.Open(connectionString), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to open gorm postgres connection: %w", err)
		}

	case "mariadb", "mysql", "":
		connectionString := buildMariaDBConnectionString(cfg)
		db, err = sqlx.Connect("mysql", connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to mariadb: %w", err)
		}

		gormDB, err = gorm.Open(gormMariaDB.Open(connectionString), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to open gorm mariadb connection: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, mariadb)", cfg.Type)
	}

	// Set connection pool settings
	if cfg.MaxConnections > 0 {
		db.SetMaxOpenConns(cfg.MaxConnections)
		db.SetMaxIdleConns(cfg.MaxConnections / 2) //nolint:mnd // Use half of max connections for idle
	}

	return &Database{
		DB:   db,
		GORM: gormDB,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

const (
	defaultDatabaseName = "radarr"
	defaultUsername     = "radarr"
)

func buildPostgresConnectionString(cfg *config.DatabaseConfig) string {
	if cfg.ConnectionURL != "" {
		return cfg.ConnectionURL
	}

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}

	port := cfg.Port
	if port == 0 {
		port = 5432
	}

	database := cfg.Database
	if database == "" {
		database = defaultDatabaseName
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, cfg.Username, cfg.Password, database)
}

func buildMariaDBConnectionString(cfg *config.DatabaseConfig) string {
	if cfg.ConnectionURL != "" {
		return cfg.ConnectionURL
	}

	host := cfg.Host
	if host == "" {
		host = "localhost"
	}

	port := cfg.Port
	if port == 0 {
		port = 3306
	}

	database := cfg.Database
	if database == "" {
		database = defaultDatabaseName
	}

	username := cfg.Username
	if username == "" {
		username = defaultUsername
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, cfg.Password, host, port, database)
}

// Migrate runs database migrations to update the schema
func Migrate(db *Database, logger *logger.Logger) error {
	var driver database.Driver
	var sourceURL string
	var err error

	// Use database-specific migration paths
	switch db.DB.DriverName() {
	case "postgres":
		sourceURL = "file://./migrations/postgres"
		driver, err = postgres.WithInstance(db.DB.DB, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
	case "mysql":
		sourceURL = "file://./migrations/mysql"
		driver, err = mysql.WithInstance(db.DB.DB, &mysql.Config{})
		if err != nil {
			return fmt.Errorf("failed to create mariadb driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s (supported: postgres, mysql)", db.DB.DriverName())
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, db.DB.DriverName(), driver)
	if err != nil {
		logger.Warn("No migrations found, skipping migration step", "error", err)
		return nil
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}
