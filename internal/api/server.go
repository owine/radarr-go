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
	HTTPReadTimeout  = 15 * time.Second
	HTTPWriteTimeout = 15 * time.Second
	HTTPIdleTimeout  = 60 * time.Second
)

type Server struct {
	config   *config.Config
	services *services.Container
	logger   *logger.Logger
	engine   *gin.Engine
	server   *http.Server
}

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
	{
		// System info
		v3.GET("/system/status", s.handleSystemStatus)

		// Movies
		movieRoutes := v3.Group("/movie")
		movieRoutes.GET("", s.handleGetMovies)
		movieRoutes.GET("/:id", s.handleGetMovie)
		movieRoutes.POST("", s.handleCreateMovie)
		movieRoutes.PUT("/:id", s.handleUpdateMovie)
		movieRoutes.DELETE("/:id", s.handleDeleteMovie)

		// Movie files
		movieFileRoutes := v3.Group("/moviefile")
		movieFileRoutes.GET("", s.handleGetMovieFiles)
		movieFileRoutes.GET("/:id", s.handleGetMovieFile)
		movieFileRoutes.DELETE("/:id", s.handleDeleteMovieFile)

		// Quality profiles
		v3.GET("/qualityprofile", s.handleGetQualityProfiles)

		// Indexers
		v3.GET("/indexer", s.handleGetIndexers)

		// Download clients
		v3.GET("/downloadclient", s.handleGetDownloadClients)

		// Queue
		v3.GET("/queue", s.handleGetQueue)

		// History
		v3.GET("/history", s.handleGetHistory)

		// Search
		searchRoutes := v3.Group("/search")
		searchRoutes.GET("/movie", s.handleSearchMovies)
	}

	// Serve static files (if any)
	s.engine.Static("/static", "./web/static")

	// Try to load HTML templates, but don't fail if they don't exist
	if _, err := os.Stat("web/templates"); err == nil {
		s.engine.LoadHTMLGlob("web/templates/*")

		// Default route for SPA
		s.engine.NoRoute(func(c *gin.Context) {
			c.HTML(http.StatusOK, "index.html", nil)
		})
	} else {
		// Default route without templates
		s.engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message":       "Radarr Go API Server",
				"version":       "1.0.0-go",
				"documentation": "Access /api/v3/system/status for system information",
			})
		})
	}
}

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
