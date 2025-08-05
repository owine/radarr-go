// Package services provides the business logic layer for the Radarr application.
//
// This package contains all service implementations that handle domain-specific
// operations and business rules. Services act as the middle layer between the
// API handlers and the data access layer, implementing the core functionality
// of the Radarr movie management system.
//
// Architecture:
//
// The services follow a dependency injection pattern using a Container that
// manages all service instances and their dependencies. Each service is
// responsible for a specific domain area and communicates with the database
// through GORM and sqlc.
//
// Core Services:
//
//   - MovieService: Movie CRUD operations, search, and metadata management
//   - QualityService: Quality profile and definition management
//   - IndexerService: Search provider configuration and operations
//   - DownloadService: Download client management and automation
//   - NotificationService: Alert and notification handling
//   - ImportListService: Automatic movie discovery and import
//   - HistoryService: Event tracking and audit logging
//   - MetadataService: External metadata provider integration (TMDB)
//   - QueueService: Background task and job management
//   - ConfigService: Application configuration management
//   - SearchService: Movie search and discovery operations
//
// Service Container:
//
// All services are managed through a centralized Container that handles
// dependency injection and service lifecycle management:
//
//	container := services.NewContainer(db, config, logger)
//	movieService := container.MovieService
//
// Transaction Support:
//
// Services support atomic operations through GORM transactions:
//
//	err := movieService.CreateWithFile(movie, movieFile)
//	// Both movie and file are created atomically
//
// Error Handling:
//
// All services implement consistent error handling with contextual logging
// and proper error wrapping for debugging and monitoring.
//
// Testing:
//
// Services include comprehensive unit tests, benchmark tests, and example
// tests to ensure reliability and performance. Mock implementations are
// available for testing dependent services.
package services
