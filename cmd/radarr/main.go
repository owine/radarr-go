// Package main provides the entry point for the Radarr Go application.
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/radarr/radarr-go/internal/api"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/services"
)

// Build information - set by ldflags during build
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var configPath = flag.String("config", "config.yaml", "path to configuration file")
	var dataDir = flag.String("data", "./data", "path to data directory")
	var showVersion = flag.Bool("version", false, "show version information and exit")
	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("Radarr Go v%s (commit: %s, built: %s)\n", version, commit, date)
		return
	}

	// Check for --version argument (also support this format)
	for _, arg := range flag.Args() {
		if arg == "--version" {
			fmt.Printf("Radarr Go v%s (commit: %s, built: %s)\n", version, commit, date)
			return
		}
	}

	// Initialize configuration
	cfg, err := config.Load(*configPath, *dataDir)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Log)

	// Initialize database
	db, err := database.New(&cfg.Database, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}()

	// Run migrations
	if err := database.Migrate(db, logger); err != nil {
		logger.Fatal("Failed to run database migrations", "error", err)
	}

	// Initialize services
	serviceContainer := services.NewContainer(db, cfg, logger)

	// Initialize and start API server
	server := api.NewServer(cfg, serviceContainer, logger)

	// Log build information
	logger.Info("Starting Radarr Go",
		"version", version,
		"commit", commit,
		"built", date,
		"port", cfg.Server.Port)

	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
