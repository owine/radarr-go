package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Storage  StorageConfig  `mapstructure:"storage"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	URLBase      string `mapstructure:"url_base"`
	EnableSSL    bool   `mapstructure:"enable_ssl"`
	SSLCertPath  string `mapstructure:"ssl_cert_path"`
	SSLKeyPath   string `mapstructure:"ssl_key_path"`
}

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

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

type AuthConfig struct {
	Method   string `mapstructure:"method"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	APIKey   string `mapstructure:"api_key"`
}

type StorageConfig struct {
	DataDirectory  string `mapstructure:"data_directory"`
	MovieDirectory string `mapstructure:"movie_directory"`
	BackupDir      string `mapstructure:"backup_directory"`
}

func Load(configPath, dataDir string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	
	// Add config search paths
	viper.AddConfigPath(".")
	viper.AddConfigPath(dataDir)
	viper.AddConfigPath(filepath.Dir(configPath))
	
	// Set defaults
	setDefaults(dataDir)
	
	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}
	
	// Override with environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("RADARR")
	
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}
	
	// Ensure directories exist
	if err := ensureDirectories(&config); err != nil {
		return nil, fmt.Errorf("error creating directories: %w", err)
	}
	
	return &config, nil
}

func setDefaults(dataDir string) {
	viper.SetDefault("server.port", 7878)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.url_base", "")
	viper.SetDefault("server.enable_ssl", false)
	
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.connection_url", filepath.Join(dataDir, "radarr.db"))
	viper.SetDefault("database.max_connections", 10)
	
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")
	
	viper.SetDefault("auth.method", "none")
	viper.SetDefault("auth.api_key", "")
	
	viper.SetDefault("storage.data_directory", dataDir)
	viper.SetDefault("storage.movie_directory", filepath.Join(dataDir, "movies"))
	viper.SetDefault("storage.backup_directory", filepath.Join(dataDir, "backups"))
}

func ensureDirectories(config *Config) error {
	dirs := []string{
		config.Storage.DataDirectory,
		config.Storage.MovieDirectory,
		config.Storage.BackupDir,
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}