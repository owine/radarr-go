// Package services provides business logic and domain services for Radarr.
package services

import (
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
)

// Container holds all services and their dependencies for dependency injection
type Container struct {
	DB     *database.Database
	Config *config.Config
	Logger *logger.Logger

	// Services
	MovieService        *MovieService
	MovieFileService    *MovieFileService
	QualityService      *QualityService
	IndexerService      *IndexerService
	DownloadService     *DownloadService
	NotificationService *NotificationService
	MetadataService     *MetadataService
	QueueService        *QueueService
	ImportListService   *ImportListService
	HistoryService      *HistoryService
	ConfigService       *ConfigService
	SearchService       *SearchService
	TaskService         *TaskService
	WantedMoviesService *WantedMoviesService

	// File management services
	NamingService           *NamingService
	MediaInfoService        *MediaInfoService
	FileOrganizationService *FileOrganizationService
	ImportService           *ImportService
	FileOperationService    *FileOperationService

	// Health monitoring services
	HealthService      *HealthService
	HealthIssueService *HealthIssueService
	PerformanceMonitor *PerformanceMonitor

	// Calendar services
	CalendarService *CalendarService
	ICalService     *ICalService

	// Collection and parse services
	CollectionService *CollectionService
	ParseService      *ParseService
	RenameService     *RenameService
}

// NewContainer creates a new service container with all dependencies initialized
func NewContainer(db *database.Database, cfg *config.Config, logger *logger.Logger) *Container {
	container := &Container{
		DB:     db,
		Config: cfg,
		Logger: logger,
	}

	// Initialize core services
	container.MovieService = NewMovieService(db, logger)
	container.MovieFileService = NewMovieFileService(db, logger)
	container.QualityService = NewQualityService(db, logger)
	container.IndexerService = NewIndexerService(db, logger)
	container.DownloadService = NewDownloadService(db, logger)
	container.NotificationService = NewNotificationService(db, logger)
	container.MetadataService = NewMetadataService(db, cfg, logger)
	container.QueueService = NewQueueService(db, logger)
	container.ImportListService = NewImportListService(db, logger, container.MetadataService, container.MovieService)
	container.HistoryService = NewHistoryService(db, logger)
	container.ConfigService = NewConfigService(db, logger)
	container.SearchService = NewSearchService(db, logger, container.IndexerService, container.QualityService,
		container.MovieService, container.DownloadService, container.NotificationService)
	container.WantedMoviesService = NewWantedMoviesService(db, logger, container.MovieService, container.QualityService)

	// Initialize file management services
	container.NamingService = NewNamingService(db, logger)
	container.MediaInfoService = NewMediaInfoService(db, logger)
	container.FileOperationService = NewFileOperationService(db, logger)
	container.FileOrganizationService = NewFileOrganizationService(db, logger, container.NamingService, container.MediaInfoService)
	container.ImportService = NewImportService(db, logger, container.MovieService, container.MovieFileService,
		container.FileOrganizationService, container.MediaInfoService, container.NamingService)

	// Initialize task service and register handlers
	container.TaskService = NewTaskService(db, logger)

	// Initialize health monitoring services
	container.PerformanceMonitor = NewPerformanceMonitor(db, logger)
	container.HealthIssueService = NewHealthIssueService(db, logger)
	container.HealthService = NewHealthService(db, cfg, logger)

	// Initialize calendar services
	container.CalendarService = NewCalendarService(db, logger)
	container.ICalService = NewICalService(db, logger, container.CalendarService)

	// Initialize collection and parse services
	container.CollectionService = NewCollectionService(db, logger)
	container.ParseService = NewParseService(db, logger)
	container.RenameService = NewRenameService(db, logger, container.NamingService)

	// Register task handlers
	container.TaskService.RegisterHandler(NewRefreshMovieHandler(container.MovieService, container.MetadataService))
	container.TaskService.RegisterHandler(NewRefreshAllMoviesHandler(container.MovieService, container.MetadataService))
	container.TaskService.RegisterHandler(NewSyncImportListHandler(container.ImportListService))
	container.TaskService.RegisterHandler(NewRefreshWantedMoviesHandler(container.WantedMoviesService))
	container.TaskService.RegisterHandler(NewAutoWantedSearchHandler(container.WantedMoviesService, container.SearchService))

	// Register health monitoring task handlers
	container.TaskService.RegisterHandler(NewHealthCheckTaskHandler(container.HealthService))
	container.TaskService.RegisterHandler(NewPerformanceMetricsTaskHandler(container.HealthService))
	container.TaskService.RegisterHandler(NewHealthMaintenanceTaskHandler(container.HealthService, container.HealthIssueService))
	container.TaskService.RegisterHandler(NewHealthReportTaskHandler(container.HealthService, container.HealthIssueService))

	// Keep the legacy health check handler for compatibility
	container.TaskService.RegisterHandler(NewHealthCheckHandler(container))
	container.TaskService.RegisterHandler(NewCleanupHandler(container))

	return container
}
