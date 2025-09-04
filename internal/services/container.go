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

	// Initialize services in logical groups
	container.initializeCoreServices(db, cfg, logger)
	container.initializeFileServices(db, logger)
	container.initializeMonitoringServices(db, cfg, logger)
	container.initializeCalendarServices(db, logger)
	container.initializeCollectionServices(db, logger)

	// Set service container reference for ConfigService
	container.ConfigService.SetServiceContainer(container)

	// Register all task handlers
	container.registerTaskHandlers()

	return container
}

// initializeCoreServices initializes the core business logic services
func (c *Container) initializeCoreServices(db *database.Database, cfg *config.Config, logger *logger.Logger) {
	c.MovieService = NewMovieService(db, logger)
	c.MovieFileService = NewMovieFileService(db, logger)
	c.QualityService = NewQualityService(db, logger)
	c.IndexerService = NewIndexerService(db, logger)
	c.DownloadService = NewDownloadService(db, logger)
	c.NotificationService = NewNotificationService(db, logger)
	c.MetadataService = NewMetadataService(db, cfg, logger)
	c.QueueService = NewQueueService(db, logger)
	c.ImportListService = NewImportListService(db, logger, c.MetadataService, c.MovieService)
	c.HistoryService = NewHistoryService(db, logger)
	c.ConfigService = NewConfigService(db, logger)
	c.SearchService = NewSearchService(db, logger, c.IndexerService, c.QualityService,
		c.MovieService, c.DownloadService, c.NotificationService)
	c.WantedMoviesService = NewWantedMoviesService(db, logger, c.MovieService, c.QualityService)
}

// initializeFileServices initializes file management and organization services
func (c *Container) initializeFileServices(db *database.Database, logger *logger.Logger) {
	c.NamingService = NewNamingService(db, logger)
	c.MediaInfoService = NewMediaInfoService(db, logger)
	c.FileOperationService = NewFileOperationService(db, logger)
	c.FileOrganizationService = NewFileOrganizationService(db, logger, c.NamingService, c.MediaInfoService)
	c.ImportService = NewImportService(db, logger, c.MovieService, c.MovieFileService,
		c.FileOrganizationService, c.MediaInfoService, c.NamingService)
}

// initializeMonitoringServices initializes health monitoring and performance services
func (c *Container) initializeMonitoringServices(db *database.Database, cfg *config.Config, logger *logger.Logger) {
	c.TaskService = NewTaskService(db, logger)
	c.PerformanceMonitor = NewPerformanceMonitor(db, logger)
	c.HealthIssueService = NewHealthIssueService(db, logger)
	c.HealthService = NewHealthService(db, cfg, logger)
}

// initializeCalendarServices initializes calendar and scheduling services
func (c *Container) initializeCalendarServices(db *database.Database, logger *logger.Logger) {
	c.CalendarService = NewCalendarService(db, logger)
	c.ICalService = NewICalService(db, logger, c.CalendarService)
}

// initializeCollectionServices initializes collection management and parsing services
func (c *Container) initializeCollectionServices(db *database.Database, logger *logger.Logger) {
	c.CollectionService = NewCollectionService(db, logger)
	c.ParseService = NewParseService(db, logger)
	c.RenameService = NewRenameService(db, logger, c.NamingService)
}

// registerTaskHandlers registers all task handlers with the task service
func (c *Container) registerTaskHandlers() {
	// Register movie and metadata task handlers
	c.TaskService.RegisterHandler(NewRefreshMovieHandler(c.MovieService, c.MetadataService))
	c.TaskService.RegisterHandler(NewRefreshAllMoviesHandler(c.MovieService, c.MetadataService))
	c.TaskService.RegisterHandler(NewSyncImportListHandler(c.ImportListService))
	c.TaskService.RegisterHandler(NewRefreshWantedMoviesHandler(c.WantedMoviesService))
	c.TaskService.RegisterHandler(NewAutoWantedSearchHandler(c.WantedMoviesService, c.SearchService))

	// Register health monitoring task handlers
	c.TaskService.RegisterHandler(NewHealthCheckTaskHandler(c.HealthService))
	c.TaskService.RegisterHandler(NewPerformanceMetricsTaskHandler(c.HealthService))
	c.TaskService.RegisterHandler(NewHealthMaintenanceTaskHandler(c.HealthService, nil))
	c.TaskService.RegisterHandler(NewHealthReportTaskHandler(c.HealthService, nil))

	// Register legacy handlers for compatibility
	c.TaskService.RegisterHandler(NewHealthCheckHandler(c))
	c.TaskService.RegisterHandler(NewCleanupHandler(c))
}
