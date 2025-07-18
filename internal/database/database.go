// Package database provides database connection and migration functionality.
package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	migrateMySQL "github.com/golang-migrate/migrate/v4/database/mysql"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"

	// Import for golang-migrate file source support
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// Import for PostgreSQL driver support (pgx)
	_ "github.com/jackc/pgx/v5/stdlib"
	// Import for MariaDB/MySQL driver support
	_ "github.com/go-sql-driver/mysql"
	"github.com/radarr/radarr-go/internal/config"
	sqlcMySQL "github.com/radarr/radarr-go/internal/database/generated/mysql"
	sqlcPostgres "github.com/radarr/radarr-go/internal/database/generated/postgres"
	"github.com/radarr/radarr-go/internal/logger"
	gormMariaDB "gorm.io/driver/mysql"
	gormPostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Database provides access to the Radarr database
type Database struct {
	GORM     *gorm.DB
	Postgres *sqlcPostgres.Queries
	MySQL    *sqlcMySQL.Queries
	DB       *sql.DB
	PgxPool  *pgxpool.Pool
	DbType   string
}

// New creates a new database connection
func New(cfg *config.DatabaseConfig, _ *logger.Logger) (*Database, error) {
	switch cfg.Type {
	case postgresType, "postgresql":
		return newPostgresDatabase(cfg)
	case "mariadb", mysqlType, "":
		return newMySQLDatabase(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type: %s (supported: postgres, mariadb)", cfg.Type)
	}
}

func newPostgresDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	connectionString := buildPostgresConnectionString(cfg)

	// Open direct SQL connection for migration
	sqlDB, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	// Open pgx pool connection for sqlc
	pgxPool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	gormDB, err := gorm.Open(gormPostgres.Open(connectionString), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm postgres connection: %w", err)
	}

	configureConnectionPool(cfg, sqlDB, pgxPool)
	postgresQueries := sqlcPostgres.New(pgxPool)

	return &Database{
		GORM:     gormDB,
		Postgres: postgresQueries,
		DB:       sqlDB,
		PgxPool:  pgxPool,
		DbType:   postgresType,
	}, nil
}

func newMySQLDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	connectionString := buildMariaDBConnectionString(cfg)

	// Open direct SQL connection for sqlc
	sqlDB, err := sql.Open(mysqlType, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}

	gormDB, err := gorm.Open(gormMariaDB.Open(connectionString), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm mariadb connection: %w", err)
	}

	configureConnectionPool(cfg, sqlDB, nil)
	mysqlQueries := sqlcMySQL.New(sqlDB)

	return &Database{
		GORM:   gormDB,
		MySQL:  mysqlQueries,
		DB:     sqlDB,
		DbType: mysqlType,
	}, nil
}

func configureConnectionPool(cfg *config.DatabaseConfig, sqlDB *sql.DB, pgxPool *pgxpool.Pool) {
	if cfg.MaxConnections > 0 {
		if sqlDB != nil {
			sqlDB.SetMaxOpenConns(cfg.MaxConnections)
			sqlDB.SetMaxIdleConns(cfg.MaxConnections / 2) //nolint:mnd // Use half of max connections for idle
		}
		if pgxPool != nil {
			// Safely convert int to int32 with bounds checking
			maxConns := cfg.MaxConnections
			if maxConns > 2147483647 { // int32 max value
				maxConns = 2147483647
			}
			// #nosec G115 -- Safe conversion with bounds checking above
			pgxPool.Config().MaxConns = int32(maxConns)
		}
	}
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.PgxPool != nil {
		d.PgxPool.Close()
	}
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

const (
	defaultDatabaseName = "radarr"
	defaultUsername     = "radarr"
	postgresType        = "postgres"
	mysqlType           = "mysql"
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

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true",
		username, cfg.Password, host, port, database)
}

// Migrate runs database migrations to update the schema
func Migrate(db *Database, logger *logger.Logger) error {
	var driver database.Driver
	var sourceURL string
	var err error

	// Get underlying sql.DB from GORM
	sqlDB, err := db.GORM.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Determine database type from GORM dialector
	var driverName string
	dialectorName := db.GORM.Name()
	switch dialectorName {
	case postgresType:
		sourceURL = "file://./migrations/postgres"
		driverName = postgresType
		driver, err = migratePostgres.WithInstance(sqlDB, &migratePostgres.Config{})
		if err != nil {
			return fmt.Errorf("failed to create postgres driver: %w", err)
		}
	case mysqlType:
		sourceURL = "file://./migrations/mysql"
		driverName = mysqlType
		driver, err = migrateMySQL.WithInstance(sqlDB, &migrateMySQL.Config{})
		if err != nil {
			return fmt.Errorf("failed to create mariadb driver: %w", err)
		}
	default:
		return fmt.Errorf("unsupported database type: %s (supported: postgres, mysql)", dialectorName)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, driverName, driver)
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
