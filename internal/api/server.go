package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/services"
)

const (
	// HTTPReadTimeout defines the maximum duration for reading the entire request
	HTTPReadTimeout = 15 * time.Second
	// HTTPWriteTimeout defines the maximum duration before timing out writes
	HTTPWriteTimeout = 15 * time.Second
	// HTTPIdleTimeout defines the maximum amount of time to wait for the next request
	HTTPIdleTimeout = 60 * time.Second
)

// Server represents the HTTP server for the Radarr API
type Server struct {
	config   *config.Config
	services *services.Container
	logger   *logger.Logger
	engine   *gin.Engine
	server   *http.Server
}

// NewServer creates a new HTTP server instance with the provided configuration and services
func NewServer(cfg *config.Config, services *services.Container, logger *logger.Logger) *Server {
	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(loggingMiddleware(logger))
	engine.Use(corsMiddleware())

	// API key middleware for protected routes
	if cfg.Auth.APIKey != "" {
		engine.Use(apiKeyMiddleware(cfg.Auth.APIKey))
	}

	server := &Server{
		config:   cfg,
		services: services,
		logger:   logger,
		engine:   engine,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	// Health check endpoint
	s.engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// API v3 routes (matching Radarr's API structure)
	v3 := s.engine.Group("/api/v3")
	s.setupAPIRoutes(v3)

	// Serve static files (if any)
	s.engine.Static("/static", "./web/static")
	s.setupTemplateRoutes()
}

func (s *Server) setupAPIRoutes(v3 *gin.RouterGroup) {
	// System info
	v3.GET("/system/status", s.handleSystemStatus)

	// Movies
	s.setupMovieRoutes(v3)

	// Quality management
	s.setupQualityRoutes(v3)

	// Indexers
	s.setupIndexerRoutes(v3)

	// Download clients
	s.setupDownloadClientRoutes(v3)

	// Import lists
	s.setupImportListRoutes(v3)

	// Queue management
	s.setupQueueRoutes(v3)

	// Task management
	s.setupTaskRoutes(v3)

	// History and Activity
	s.setupHistoryRoutes(v3)
	s.setupActivityRoutes(v3)

	// Configuration
	s.setupConfigRoutes(v3)

	// Notifications
	s.setupNotificationRoutes(v3)

	// File Organization and Import
	s.setupFileOrganizationRoutes(v3)

	// Search & Release Management
	s.setupSearchRoutes(v3)

	// Health monitoring
	s.setupHealthRoutes(v3)

	// Calendar
	s.setupCalendarRoutes(v3)

	// Wanted movies
	s.setupWantedRoutes(v3)

	// Search
	searchRoutes := v3.Group("/search")
	searchRoutes.GET("/movie", s.handleSearchMovies)

	// Collections
	s.setupCollectionRoutes(v3)

	// Parse functionality
	s.setupParseRoutes(v3)

	// Rename functionality
	s.setupRenameRoutes(v3)
}

func (s *Server) setupMovieRoutes(v3 *gin.RouterGroup) {
	movieRoutes := v3.Group("/movie")
	movieRoutes.GET("", s.handleGetMovies)
	movieRoutes.GET("/:id", s.handleGetMovie)
	movieRoutes.POST("", s.handleCreateMovie)
	movieRoutes.PUT("/:id", s.handleUpdateMovie)
	movieRoutes.DELETE("/:id", s.handleDeleteMovie)

	// Movie discovery and metadata endpoints
	movieRoutes.GET("/lookup", s.handleMovieLookup)
	movieRoutes.GET("/lookup/tmdb", s.handleMovieByTMDBID)
	movieRoutes.GET("/popular", s.handleMovieDiscoverPopular)
	movieRoutes.GET("/trending", s.handleMovieDiscoverTrending)
	movieRoutes.PUT("/:id/refresh", s.handleRefreshMovieMetadata)

	movieFileRoutes := v3.Group("/moviefile")
	movieFileRoutes.GET("", s.handleGetMovieFiles)
	movieFileRoutes.GET("/:id", s.handleGetMovieFile)
	movieFileRoutes.DELETE("/:id", s.handleDeleteMovieFile)
}

func (s *Server) setupQualityRoutes(v3 *gin.RouterGroup) {
	qualityProfileRoutes := v3.Group("/qualityprofile")
	qualityProfileRoutes.GET("", s.handleGetQualityProfiles)
	qualityProfileRoutes.GET("/:id", s.handleGetQualityProfile)
	qualityProfileRoutes.POST("", s.handleCreateQualityProfile)
	qualityProfileRoutes.PUT("/:id", s.handleUpdateQualityProfile)
	qualityProfileRoutes.DELETE("/:id", s.handleDeleteQualityProfile)

	qualityDefinitionRoutes := v3.Group("/qualitydefinition")
	qualityDefinitionRoutes.GET("", s.handleGetQualityDefinitions)
	qualityDefinitionRoutes.GET("/:id", s.handleGetQualityDefinition)
	qualityDefinitionRoutes.PUT("/:id", s.handleUpdateQualityDefinition)

	customFormatRoutes := v3.Group("/customformat")
	customFormatRoutes.GET("", s.handleGetCustomFormats)
	customFormatRoutes.GET("/:id", s.handleGetCustomFormat)
	customFormatRoutes.POST("", s.handleCreateCustomFormat)
	customFormatRoutes.PUT("/:id", s.handleUpdateCustomFormat)
	customFormatRoutes.DELETE("/:id", s.handleDeleteCustomFormat)
}

func (s *Server) setupIndexerRoutes(v3 *gin.RouterGroup) {
	indexerRoutes := v3.Group("/indexer")
	indexerRoutes.GET("", s.handleGetIndexers)
	indexerRoutes.GET("/:id", s.handleGetIndexer)
	indexerRoutes.POST("", s.handleCreateIndexer)
	indexerRoutes.PUT("/:id", s.handleUpdateIndexer)
	indexerRoutes.DELETE("/:id", s.handleDeleteIndexer)
	indexerRoutes.POST("/:id/test", s.handleTestIndexer)
}

func (s *Server) setupDownloadClientRoutes(v3 *gin.RouterGroup) {
	downloadClientRoutes := v3.Group("/downloadclient")
	downloadClientRoutes.GET("", s.handleGetDownloadClients)
	downloadClientRoutes.GET("/:id", s.handleGetDownloadClient)
	downloadClientRoutes.POST("", s.handleCreateDownloadClient)
	downloadClientRoutes.PUT("/:id", s.handleUpdateDownloadClient)
	downloadClientRoutes.DELETE("/:id", s.handleDeleteDownloadClient)
	downloadClientRoutes.POST("/test", s.handleTestDownloadClient)
	downloadClientRoutes.GET("/stats", s.handleGetDownloadClientStats)

	// Download history
	v3.GET("/downloadhistory", s.handleGetDownloadHistory)
}

func (s *Server) setupImportListRoutes(v3 *gin.RouterGroup) {
	importListRoutes := v3.Group("/importlist")
	importListRoutes.GET("", s.handleGetImportLists)
	importListRoutes.GET("/:id", s.handleGetImportList)
	importListRoutes.POST("", s.handleCreateImportList)
	importListRoutes.PUT("/:id", s.handleUpdateImportList)
	importListRoutes.DELETE("/:id", s.handleDeleteImportList)
	importListRoutes.POST("/test", s.handleTestImportList)
	importListRoutes.POST("/:id/sync", s.handleSyncImportList)
	importListRoutes.POST("/sync", s.handleSyncAllImportLists)
	importListRoutes.GET("/stats", s.handleGetImportListStats)

	// Import list movies
	v3.GET("/importlistmovies", s.handleGetImportListMovies)
}

func (s *Server) setupQueueRoutes(v3 *gin.RouterGroup) {
	queueRoutes := v3.Group("/queue")
	queueRoutes.GET("", s.handleGetQueue)
	queueRoutes.GET("/:id", s.handleGetQueueItem)
	queueRoutes.DELETE("/:id", s.handleRemoveQueueItem)
	queueRoutes.DELETE("/bulk", s.handleRemoveQueueItemsBulk)
	queueRoutes.GET("/stats", s.handleGetQueueStats)
}

func (s *Server) setupNotificationRoutes(v3 *gin.RouterGroup) {
	notificationRoutes := v3.Group("/notification")
	notificationRoutes.GET("", s.handleGetNotifications)
	notificationRoutes.GET("/:id", s.handleGetNotification)
	notificationRoutes.POST("", s.handleCreateNotification)
	notificationRoutes.PUT("/:id", s.handleUpdateNotification)
	notificationRoutes.DELETE("/:id", s.handleDeleteNotification)
	notificationRoutes.POST("/test", s.handleTestNotification)
	notificationRoutes.GET("/schema", s.handleGetNotificationProviders)
	notificationRoutes.GET("/schema/:type", s.handleGetNotificationProviderFields)
	notificationRoutes.GET("/history", s.handleGetNotificationHistory)
}

func (s *Server) setupTemplateRoutes() {
	if _, err := os.Stat("web/templates"); err == nil {
		s.engine.LoadHTMLGlob("web/templates/*")
		s.engine.NoRoute(func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
	} else {
		s.engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message":       "Radarr Go API Server",
				"version":       "1.0.0-go",
				"documentation": "Access /api/v3/system/status for system information",
			})
		})
	}
}

// Start begins listening for HTTP requests on the configured address
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  HTTPReadTimeout,
		WriteTimeout: HTTPWriteTimeout,
		IdleTimeout:  HTTPIdleTimeout,
	}

	s.logger.Info("Starting HTTP server", "address", addr)

	if s.config.Server.EnableSSL {
		return s.server.ListenAndServeTLS(s.config.Server.SSLCertPath, s.config.Server.SSLKeyPath)
	}

	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// Middleware functions
func loggingMiddleware(logger *logger.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.Info("HTTP Request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"ip", param.ClientIP,
			"user-agent", param.Request.UserAgent(),
		)
		return ""
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func apiKeyMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip API key check for health endpoint
		if c.Request.URL.Path == "/ping" {
			c.Next()
			return
		}

		providedKey := c.GetHeader("X-API-Key")
		if providedKey == "" {
			providedKey = c.Query("apikey")
		}

		if providedKey != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (s *Server) setupHistoryRoutes(v3 *gin.RouterGroup) {
	historyRoutes := v3.Group("/history")
	historyRoutes.GET("", s.handleGetHistory)
	historyRoutes.GET("/:id", s.handleGetHistoryByID)
	historyRoutes.DELETE("/:id", s.handleDeleteHistoryRecord)
	historyRoutes.GET("/stats", s.handleGetHistoryStats)
}

func (s *Server) setupActivityRoutes(v3 *gin.RouterGroup) {
	activityRoutes := v3.Group("/activity")
	activityRoutes.GET("", s.handleGetActivity)
	activityRoutes.GET("/:id", s.handleGetActivityByID)
	activityRoutes.DELETE("/:id", s.handleDeleteActivity)
	activityRoutes.GET("/running", s.handleGetRunningActivities)
}

func (s *Server) setupConfigRoutes(v3 *gin.RouterGroup) {
	// Host configuration
	v3.GET("/config/host", s.handleGetHostConfig)
	v3.PUT("/config/host", s.handleUpdateHostConfig)

	// Naming configuration
	v3.GET("/config/naming", s.handleGetNamingConfig)
	v3.PUT("/config/naming", s.handleUpdateNamingConfig)
	v3.GET("/config/naming/tokens", s.handleGetNamingTokens)

	// Media management configuration
	v3.GET("/config/mediamanagement", s.handleGetMediaManagementConfig)
	v3.PUT("/config/mediamanagement", s.handleUpdateMediaManagementConfig)

	// Root folders
	rootFolderRoutes := v3.Group("/rootfolder")
	rootFolderRoutes.GET("", s.handleGetRootFolders)
	rootFolderRoutes.GET("/:id", s.handleGetRootFolder)
	rootFolderRoutes.POST("", s.handleCreateRootFolder)
	rootFolderRoutes.PUT("/:id", s.handleUpdateRootFolder)
	rootFolderRoutes.DELETE("/:id", s.handleDeleteRootFolder)

	// Configuration stats
	v3.GET("/config/stats", s.handleGetConfigStats)

	// Application settings
	v3.GET("/config/app", s.handleGetAppSettings)
	v3.PUT("/config/app", s.handleUpdateAppSettings)

	// Configuration validation
	v3.GET("/config/validate", s.handleValidateAllConfigurations)
	v3.GET("/config/validate/:type", s.handleTestConfiguration)

	// Configuration backup and restore
	v3.POST("/config/backup", s.handleCreateConfigurationBackup)
	v3.POST("/config/restore", s.handleRestoreConfigurationBackup)
	v3.POST("/config/reset", s.handleFactoryResetConfiguration)

	// Configuration export/import
	v3.GET("/config/export", s.handleExportConfiguration)
	v3.POST("/config/import", s.handleImportConfiguration)
}

func (s *Server) setupSearchRoutes(v3 *gin.RouterGroup) {
	// Release routes
	releaseRoutes := v3.Group("/release")
	releaseRoutes.GET("", s.handleGetReleases)
	releaseRoutes.GET("/:id", s.handleGetRelease)
	releaseRoutes.DELETE("/:id", s.handleDeleteRelease)
	releaseRoutes.GET("/stats", s.handleGetReleaseStats)
	releaseRoutes.POST("/grab", s.handleGrabRelease)

	// Search routes
	searchRoutes := v3.Group("/search")
	searchRoutes.GET("", s.handleSearchReleases)
	searchRoutes.GET("/movie/:id", s.handleSearchMovieReleases)
	searchRoutes.GET("/interactive", s.handleInteractiveSearch)
}

func (s *Server) setupTaskRoutes(v3 *gin.RouterGroup) {
	// Task management routes (Radarr command API)
	commandRoutes := v3.Group("/command")
	commandRoutes.GET("", s.handleGetTasks)
	commandRoutes.GET("/:id", s.handleGetTask)
	commandRoutes.POST("", s.handleQueueTask)
	commandRoutes.DELETE("/:id", s.handleCancelTask)

	// System task routes
	systemRoutes := v3.Group("/system/task")
	systemRoutes.GET("", s.handleGetScheduledTasks)
	systemRoutes.POST("", s.handleCreateScheduledTask)
	systemRoutes.PUT("/:id", s.handleUpdateScheduledTask)
	systemRoutes.DELETE("/:id", s.handleDeleteScheduledTask)
	systemRoutes.GET("/status", s.handleGetQueueStatus)

	// Command-specific task routes
	movieCommands := v3.Group("/movie")
	movieCommands.POST("/:id/refresh", s.handleRefreshMovie)
	movieCommands.POST("/refresh", s.handleRefreshAllMovies)

	// Import list sync commands are already handled in setupImportListRoutes()

	systemCommands := v3.Group("/system")
	systemCommands.POST("/health", s.handleRunHealthCheck)
	systemCommands.POST("/cleanup", s.handleRunCleanup)
}

func (s *Server) setupFileOrganizationRoutes(v3 *gin.RouterGroup) {
	// File Organization routes
	orgRoutes := v3.Group("/fileorganization")
	orgRoutes.GET("", s.handleGetFileOrganizations)
	orgRoutes.GET("/:id", s.handleGetFileOrganizationByID)
	orgRoutes.POST("/retry", s.handleRetryFailedOrganizations)
	orgRoutes.POST("/scan", s.handleScanDirectory)

	// Import routes
	importRoutes := v3.Group("/import")
	importRoutes.POST("/process", s.handleProcessImport)
	importRoutes.GET("/manual", s.handleGetManualImports)
	importRoutes.POST("/manual", s.handleProcessManualImport)

	// Additional naming routes (basic naming routes are in setupConfigRoutes)
	v3.GET("/config/naming/preview/:movieId", s.handlePreviewNaming)

	// File operation tracking routes
	operationRoutes := v3.Group("/fileoperation")
	operationRoutes.GET("", s.handleGetFileOperations)
	operationRoutes.GET("/:id", s.handleGetFileOperation)
	operationRoutes.DELETE("/:id", s.handleCancelFileOperation)
	operationRoutes.GET("/summary", s.handleGetFileOperationSummary)

	// Media info routes
	mediaInfoRoutes := v3.Group("/mediainfo")
	mediaInfoRoutes.POST("/extract", s.handleExtractMediaInfo)
}

// setupHealthRoutes configures health monitoring and diagnostics routes
func (s *Server) setupHealthRoutes(v3 *gin.RouterGroup) {
	// Health status routes
	healthRoutes := v3.Group("/health")
	healthRoutes.GET("", s.handleGetHealth)                    // Overall health status
	healthRoutes.GET("/dashboard", s.handleGetHealthDashboard) // Complete health dashboard
	healthRoutes.GET("/check/:name", s.handleGetHealthCheck)   // Run specific health check

	// Health issues management
	issueRoutes := healthRoutes.Group("/issue")
	issueRoutes.GET("", s.handleGetHealthIssues)                 // List health issues with filtering
	issueRoutes.GET("/:id", s.handleGetHealthIssue)              // Get specific health issue
	issueRoutes.POST("/:id/dismiss", s.handleDismissHealthIssue) // Dismiss a health issue
	issueRoutes.POST("/:id/resolve", s.handleResolveHealthIssue) // Mark health issue as resolved

	// System resource monitoring
	systemRoutes := healthRoutes.Group("/system")
	systemRoutes.GET("/resources", s.handleGetSystemResources) // Current system resources
	systemRoutes.GET("/diskspace", s.handleGetDiskSpace)       // Disk space information

	// Performance metrics
	metricsRoutes := healthRoutes.Group("/metrics")
	metricsRoutes.GET("", s.handleGetPerformanceMetrics)            // Performance metrics with time range
	metricsRoutes.POST("/record", s.handleRecordPerformanceMetrics) // Manually record metrics

	// Health monitoring control
	monitoringRoutes := healthRoutes.Group("/monitoring")
	monitoringRoutes.POST("/start", s.handleStartHealthMonitoring) // Start background monitoring
	monitoringRoutes.POST("/stop", s.handleStopHealthMonitoring)   // Stop background monitoring
	monitoringRoutes.POST("/cleanup", s.handleCleanupHealthData)   // Cleanup old health data
}

// setupCalendarRoutes configures calendar and event tracking routes
func (s *Server) setupCalendarRoutes(v3 *gin.RouterGroup) {
	// Calendar events routes
	calendarRoutes := v3.Group("/calendar")
	calendarRoutes.GET("", s.handleGetCalendar)                 // Get calendar events with filtering
	calendarRoutes.GET("/feed.ics", s.handleGetCalendarFeed)    // iCal feed for external applications
	calendarRoutes.GET("/feed/url", s.handleGetCalendarFeedURL) // Generate iCal feed URL
	calendarRoutes.POST("/refresh", s.handleRefreshCalendar)    // Force refresh calendar events
	calendarRoutes.GET("/stats", s.handleGetCalendarStats)      // Calendar statistics

	// Calendar configuration routes
	configRoutes := calendarRoutes.Group("/config")
	configRoutes.GET("", s.handleGetCalendarConfiguration)    // Get calendar configuration
	configRoutes.PUT("", s.handleUpdateCalendarConfiguration) // Update calendar configuration
}

// setupWantedRoutes configures wanted movies tracking and management routes
func (s *Server) setupWantedRoutes(v3 *gin.RouterGroup) {
	// Wanted movies routes
	wantedRoutes := v3.Group("/wanted")
	wantedRoutes.GET("/missing", s.handleGetMissingMovies)          // Get missing movies
	wantedRoutes.GET("/cutoff", s.handleGetCutoffUnmetMovies)       // Get cutoff unmet movies
	wantedRoutes.GET("", s.handleGetAllWantedMovies)                // Get all wanted movies with filters
	wantedRoutes.GET("/stats", s.handleGetWantedStats)              // Get wanted movies statistics
	wantedRoutes.GET("/:id", s.handleGetWantedMovie)                // Get specific wanted movie
	wantedRoutes.POST("/search", s.handleTriggerWantedSearch)       // Trigger searches for wanted movies
	wantedRoutes.POST("/bulk", s.handleWantedBulkOperation)         // Bulk operations on wanted movies
	wantedRoutes.POST("/refresh", s.handleRefreshWantedMovies)      // Refresh wanted movies analysis
	wantedRoutes.PUT("/:id/priority", s.handleUpdateWantedPriority) // Update wanted movie priority
	wantedRoutes.DELETE("/:id", s.handleRemoveWantedMovie)          // Remove from wanted list
}

// setupCollectionRoutes configures movie collection management routes
func (s *Server) setupCollectionRoutes(v3 *gin.RouterGroup) {
	collectionRoutes := v3.Group("/collection")
	collectionRoutes.GET("", s.handleGetCollections)                         // Get all collections
	collectionRoutes.GET("/:id", s.handleGetCollection)                      // Get specific collection
	collectionRoutes.POST("", s.handleCreateCollection)                      // Create new collection
	collectionRoutes.PUT("/:id", s.handleUpdateCollection)                   // Update collection
	collectionRoutes.DELETE("/:id", s.handleDeleteCollection)                // Delete collection
	collectionRoutes.POST("/:id/search", s.handleSearchCollectionMovies)     // Search for missing movies
	collectionRoutes.POST("/:id/sync", s.handleSyncCollectionFromTMDB)       // Sync from TMDB
	collectionRoutes.GET("/:id/statistics", s.handleGetCollectionStatistics) // Get collection statistics
}

// setupParseRoutes configures release name parsing routes
func (s *Server) setupParseRoutes(v3 *gin.RouterGroup) {
	parseRoutes := v3.Group("/parse")
	parseRoutes.GET("", s.handleParseReleaseTitle)        // Parse single release title
	parseRoutes.POST("", s.handleParseMultipleTitles)     // Parse multiple release titles
	parseRoutes.DELETE("/cache", s.handleClearParseCache) // Clear parse cache
}

// setupRenameRoutes configures file and folder renaming routes
func (s *Server) setupRenameRoutes(v3 *gin.RouterGroup) {
	renameRoutes := v3.Group("/rename")
	renameRoutes.GET("/preview", s.handlePreviewRename)                   // Preview file renames
	renameRoutes.POST("", s.handleRenameMovies)                           // Execute file renames
	renameRoutes.GET("/preview/folder", s.handlePreviewMovieFolderRename) // Preview folder renames
	renameRoutes.POST("/folder", s.handleRenameMovieFolders)              // Execute folder renames
}
