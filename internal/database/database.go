// Package database provides database connection and migration functionality.
package database

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	migrateSqlite "github.com/golang-migrate/migrate/v4/database/sqlite"

	// Import for golang-migrate file source support
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"

	// Import for PostgreSQL driver support
	_ "github.com/lib/pq"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	gormPostgres "gorm.io/driver/postgres"
	gormSqlite "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	// Import for pure-Go SQLite driver support (no CGO required)
	_ "modernc.org/sqlite"
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

	const postgresType = "postgres"
	switch cfg.Type {
	case postgresType, "postgresql":
		connectionString := buildPostgresConnectionString(cfg)
		db, err = sqlx.Connect(postgresType, connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to postgres: %w", err)
		}

		gormDB, err = gorm.Open(gormPostgres.Open(connectionString), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to open gorm postgres connection: %w", err)
		}

	case "sqlite", "":
		connectionString := cfg.ConnectionURL
		if connectionString == "" {
			connectionString = "radarr.db"
		}

		db, err = sqlx.Connect("sqlite", connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to sqlite: %w", err)
		}

		gormDB, err = gorm.Open(gormSqlite.Open(connectionString), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to open gorm sqlite connection: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
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
		database = "radarr"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, cfg.Username, cfg.Password, database)
}

// Migrate runs database migrations to update the schema
func Migrate(db *Database, logger *logger.Logger) error {
	var driver database.Driver
	var sourceURL string
	var err error

	// Determine the migration source path
	sourceURL = "file://./migrations"

	switch db.DB.DriverName() {
	case "postgres":
		driver, err = postgres.WithInstance(db.DB.DB, &postgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
	case "sqlite":
		driver, err = migrateSqlite.WithInstance(db.DB.DB, &migrateSqlite.Config{})
		if err != nil {
			return fmt.Errorf("failed to create sqlite driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database driver: %s", db.DB.DriverName())
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
