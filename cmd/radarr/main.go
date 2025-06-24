// Package main provides the entry point for the Radarr Go application.
package main

import (
	"flag"
	"log"

	"github.com/radarr/radarr-go/internal/api"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/services"
)

func main() {
	var configPath = flag.String("config", "config.yaml", "path to configuration file")
	var dataDir = flag.String("data", "./data", "path to data directory")
	flag.Parse()

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

	logger.Info("Starting Radarr server", "port", cfg.Server.Port)

	if err := server.Start(); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}
