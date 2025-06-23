// Package api provides HTTP handlers and server functionality for the Radarr API.
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/models"
)

const (
	DebugLevel      = "debug"
	DefaultPageSize = 20
	DefaultPage     = 1
)

// System handlers
func (s *Server) handleSystemStatus(c *gin.Context) {
	status := gin.H{
		"version":                "1.0.0-go",
		"buildTime":              time.Now().Format(time.RFC3339),
		"isDebug":                s.config.Log.Level == DebugLevel,
		"isProduction":           s.config.Log.Level != DebugLevel,
		"isAdmin":                true,
		"isUserInteractive":      false,
		"startupPath":            s.config.Storage.DataDirectory,
		"appData":                s.config.Storage.DataDirectory,
		"osName":                 "linux",
		"osVersion":              "unknown",
		"isMonoRuntime":          false,
		"isMono":                 false,
		"isLinux":                true,
		"isOsx":                  false,
		"isWindows":              false,
		"mode":                   "console",
		"branch":                 "develop",
		"authentication":         s.config.Auth.Method,
		"sqliteVersion":          "3.x",
		"migrationVersion":       1,
		"urlBase":                s.config.Server.URLBase,
		"runtimeVersion":         "go1.21",
		"databaseType":           s.config.Database.Type,
		"databaseVersion":        "unknown",
		"packageVersion":         "1.0.0-go",
		"packageAuthor":          "Radarr Go Team",
		"packageUpdateMechanism": "docker",
	}

	c.JSON(http.StatusOK, status)
}

// Helper function to parse ID from URL parameter
func (s *Server) parseIDParam(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid ID: %w", err)
	}
	return id, nil
}

// Helper function to handle get-by-ID operations
func (s *Server) handleGetByID(c *gin.Context, resourceName string, getFunc func(int) (any, error)) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resource, err := getFunc(id)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get %s", resourceName), "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("%s not found", resourceName)})
		return
	}

	c.JSON(http.StatusOK, resource)
}

// Helper function to handle delete operations
func (s *Server) handleDeleteByID(c *gin.Context, resourceName string, deleteFunc func(int) error) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := deleteFunc(id); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to delete %s", resourceName), "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete %s", resourceName)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s deleted successfully", resourceName)})
}

// Movie handlers
func (s *Server) handleGetMovies(c *gin.Context) {
	movies, err := s.services.MovieService.GetAll()
	if err != nil {
		s.logger.Error("Failed to get movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (s *Server) handleGetMovie(c *gin.Context) {
	s.handleGetByID(c, "movie", func(id int) (any, error) {
		return s.services.MovieService.GetByID(id)
	})
}

func (s *Server) handleCreateMovie(c *gin.Context) {
	var movie models.Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie data"})
		return
	}

	movie.Added = time.Now()

	if err := s.services.MovieService.Create(&movie); err != nil {
		s.logger.Error("Failed to create movie", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create movie"})
		return
	}

	c.JSON(http.StatusCreated, movie)
}

func (s *Server) handleUpdateMovie(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var movie models.Movie
	if err := c.ShouldBindJSON(&movie); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie data"})
		return
	}

	movie.ID = id

	if err := s.services.MovieService.Update(&movie); err != nil {
		s.logger.Error("Failed to update movie", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update movie"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (s *Server) handleDeleteMovie(c *gin.Context) {
	s.handleDeleteByID(c, "movie", s.services.MovieService.Delete)
}

// MovieFile handlers
func (s *Server) handleGetMovieFiles(c *gin.Context) {
	movieFiles, err := s.services.MovieFileService.GetAll()
	if err != nil {
		s.logger.Error("Failed to get movie files", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve movie files"})
		return
	}

	c.JSON(http.StatusOK, movieFiles)
}

func (s *Server) handleGetMovieFile(c *gin.Context) {
	s.handleGetByID(c, "movie file", func(id int) (any, error) {
		return s.services.MovieFileService.GetByID(id)
	})
}

func (s *Server) handleDeleteMovieFile(c *gin.Context) {
	s.handleDeleteByID(c, "movie file", s.services.MovieFileService.Delete)
}

// Placeholder handlers for other endpoints
func (s *Server) handleGetQualityProfiles(c *gin.Context) {
	profiles, err := s.services.QualityService.GetQualityProfiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quality profiles"})
		return
	}
	c.JSON(http.StatusOK, profiles)
}

func (s *Server) handleGetIndexers(c *gin.Context) {
	indexers, err := s.services.IndexerService.GetIndexers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve indexers"})
		return
	}
	c.JSON(http.StatusOK, indexers)
}

func (s *Server) handleGetDownloadClients(c *gin.Context) {
	downloads, err := s.services.DownloadService.GetDownloads()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve download clients"})
		return
	}
	c.JSON(http.StatusOK, downloads)
}

func (s *Server) handleGetQueue(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"page":          DefaultPage,
		"pageSize":      DefaultPageSize,
		"sortKey":       "timeleft",
		"sortDirection": "ascending",
		"totalRecords":  0,
		"records":       []any{},
	})
}

func (s *Server) handleGetHistory(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"page":          DefaultPage,
		"pageSize":      DefaultPageSize,
		"sortKey":       "date",
		"sortDirection": "descending",
		"totalRecords":  0,
		"records":       []any{},
	})
}

func (s *Server) handleSearchMovies(c *gin.Context) {
	query := c.Query("term")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
		return
	}

	movies, err := s.services.MovieService.Search(query)
	if err != nil {
		s.logger.Error("Failed to search movies", "query", query, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
}
