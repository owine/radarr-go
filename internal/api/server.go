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

	// History and Activity
	s.setupHistoryRoutes(v3)
	s.setupActivityRoutes(v3)

	// Configuration
	s.setupConfigRoutes(v3)

	// Search
	searchRoutes := v3.Group("/search")
	searchRoutes.GET("/movie", s.handleSearchMovies)
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
}
