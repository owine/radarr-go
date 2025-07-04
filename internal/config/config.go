// Package config provides configuration loading and management for Radarr.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// DefaultRadarrPort is the default port for Radarr server
	DefaultRadarrPort = 7878
	// DefaultMaxConnections is the default maximum number of database connections
	DefaultMaxConnections = 10
	// DefaultDirectoryPerm is the default permission for created directories
	DefaultDirectoryPerm = 0755
)

// Config represents the main configuration structure for Radarr
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
	TMDB     TMDBConfig     `mapstructure:"tmdb"`
}

// ServerConfig contains HTTP server configuration settings
type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Host        string `mapstructure:"host"`
	URLBase     string `mapstructure:"url_base"`
	EnableSSL   bool   `mapstructure:"enable_ssl"`
	SSLCertPath string `mapstructure:"ssl_cert_path"`
	SSLKeyPath  string `mapstructure:"ssl_key_path"`
}

// DatabaseConfig contains database connection and configuration settings
type DatabaseConfig struct {
	Type           string `mapstructure:"type"`
	ConnectionURL  string `mapstructure:"connection_url"`
	Host           string `mapstructure:"host"`
	Port           int    `mapstructure:"port"`
	Database       string `mapstructure:"database"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	MaxConnections int    `mapstructure:"max_connections"`
}

// LogConfig contains logging configuration settings
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// AuthConfig contains authentication and authorization settings
type AuthConfig struct {
	Method   string `mapstructure:"method"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	APIKey   string `mapstructure:"api_key"`
}

// StorageConfig contains file and directory path settings
type StorageConfig struct {
	DataDirectory  string `mapstructure:"data_directory"`
	MovieDirectory string `mapstructure:"movie_directory"`
	BackupDir      string `mapstructure:"backup_directory"`
}

// TMDBConfig contains TheMovieDB API configuration
type TMDBConfig struct {
	APIKey string `mapstructure:"api_key"`
}

// Load reads and parses the configuration from file and environment variables
func Load(configPath, dataDir string) (*Config, error) {
	vip := viper.New()
	vip.SetConfigFile(configPath)
	vip.SetConfigType("yaml")

	// Set defaults
	setDefaults(vip, dataDir)

	// Read config file if it exists
	if err := vip.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	// Override with environment variables
	vip.AutomaticEnv()
	vip.SetEnvPrefix("RADARR")

	// Map nested config keys to environment variables
	_ = vip.BindEnv("database.type", "RADARR_DATABASE_TYPE")
	_ = vip.BindEnv("database.host", "RADARR_DATABASE_HOST")
	_ = vip.BindEnv("database.port", "RADARR_DATABASE_PORT")
	_ = vip.BindEnv("database.database", "RADARR_DATABASE_DATABASE")
	_ = vip.BindEnv("database.username", "RADARR_DATABASE_USERNAME")
	_ = vip.BindEnv("database.password", "RADARR_DATABASE_PASSWORD")
	_ = vip.BindEnv("database.max_connections", "RADARR_DATABASE_MAX_CONNECTIONS")
	_ = vip.BindEnv("server.port", "RADARR_SERVER_PORT")
	_ = vip.BindEnv("log.level", "RADARR_LOG_LEVEL")

	var config Config
	if err := vip.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Ensure directories exist
	if err := ensureDirectories(&config); err != nil {
		return nil, fmt.Errorf("error creating directories: %w", err)
	}

	return &config, nil
}

func setDefaults(vip *viper.Viper, dataDir string) {
	vip.SetDefault("server.port", DefaultRadarrPort)
	vip.SetDefault("server.host", "0.0.0.0")
	vip.SetDefault("server.url_base", "")
	vip.SetDefault("server.enable_ssl", false)

	vip.SetDefault("database.type", "postgres")
	vip.SetDefault("database.host", "localhost")
	vip.SetDefault("database.port", 5432)
	vip.SetDefault("database.database", "radarr")
	vip.SetDefault("database.username", "radarr")
	vip.SetDefault("database.password", "password")
	vip.SetDefault("database.max_connections", DefaultMaxConnections)

	vip.SetDefault("log.level", "info")
	vip.SetDefault("log.format", "json")
	vip.SetDefault("log.output", "stdout")

	vip.SetDefault("auth.method", "none")
	vip.SetDefault("auth.api_key", "")

	vip.SetDefault("storage.data_directory", dataDir)
	vip.SetDefault("storage.movie_directory", filepath.Join(dataDir, "movies"))
	vip.SetDefault("storage.backup_directory", filepath.Join(dataDir, "backups"))
}

func ensureDirectories(config *Config) error {
	dirs := []string{
		config.Storage.DataDirectory,
		config.Storage.MovieDirectory,
		config.Storage.BackupDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, DefaultDirectoryPerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
