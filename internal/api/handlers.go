// Package api provides HTTP handlers and server functionality for the Radarr API.
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/services"
)

const (
	// DebugLevel represents the debug log level
	DebugLevel = "debug"
	// DefaultPageSize is the default number of items per page for paginated responses
	DefaultPageSize = 20
	// DefaultPage is the default page number for paginated responses
	DefaultPage = 1
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
		"runtimeVersion":         "go1.24",
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
		s.logger.Error("Failed to get quality profiles", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quality profiles"})
		return
	}
	c.JSON(http.StatusOK, profiles)
}

func (s *Server) handleGetIndexers(c *gin.Context) {
	indexers, err := s.services.IndexerService.GetIndexers()
	if err != nil {
		s.logger.Error("Failed to get indexers", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve indexers"})
		return
	}
	c.JSON(http.StatusOK, indexers)
}

func (s *Server) handleGetDownloadClients(c *gin.Context) {
	clients, err := s.services.DownloadService.GetDownloadClients()
	if err != nil {
		s.logger.Error("Failed to get download clients", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve download clients"})
		return
	}
	c.JSON(http.StatusOK, clients)
}

func (s *Server) handleGetDownloadClient(c *gin.Context) {
	s.handleGetByID(c, "download client", func(id int) (any, error) {
		return s.services.DownloadService.GetDownloadClientByID(id)
	})
}

func (s *Server) handleCreateDownloadClient(c *gin.Context) {
	var client models.DownloadClient
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid download client data"})
		return
	}

	if err := s.services.DownloadService.CreateDownloadClient(&client); err != nil {
		s.logger.Error("Failed to create download client", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create download client"})
		return
	}

	c.JSON(http.StatusCreated, client)
}

func (s *Server) handleUpdateDownloadClient(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var client models.DownloadClient
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid download client data"})
		return
	}

	client.ID = id
	if err := s.services.DownloadService.UpdateDownloadClient(&client); err != nil {
		s.logger.Error("Failed to update download client", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update download client"})
		return
	}

	c.JSON(http.StatusOK, client)
}

func (s *Server) handleDeleteDownloadClient(c *gin.Context) {
	s.handleDeleteByID(c, "download client", s.services.DownloadService.DeleteDownloadClient)
}

func (s *Server) handleTestDownloadClient(c *gin.Context) {
	var client models.DownloadClient
	if err := c.ShouldBindJSON(&client); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid download client data"})
		return
	}

	result, err := s.services.DownloadService.TestDownloadClient(&client)
	if err != nil {
		s.logger.Error("Failed to test download client", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test download client"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleGetDownloadClientStats(c *gin.Context) {
	stats, err := s.services.DownloadService.GetDownloadClientStats()
	if err != nil {
		s.logger.Error("Failed to get download client stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get download client statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleGetDownloadHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	history, err := s.services.DownloadService.GetDownloadHistory(limit)
	if err != nil {
		s.logger.Error("Failed to get download history", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve download history"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// Import List Handlers

func (s *Server) handleGetImportLists(c *gin.Context) {
	lists, err := s.services.ImportListService.GetImportLists()
	if err != nil {
		s.logger.Error("Failed to get import lists", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve import lists"})
		return
	}
	c.JSON(http.StatusOK, lists)
}

func (s *Server) handleGetImportList(c *gin.Context) {
	s.handleGetByID(c, "import list", func(id int) (any, error) {
		return s.services.ImportListService.GetImportListByID(id)
	})
}

func (s *Server) handleCreateImportList(c *gin.Context) {
	var list models.ImportList
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid import list data"})
		return
	}

	if err := s.services.ImportListService.CreateImportList(&list); err != nil {
		s.logger.Error("Failed to create import list", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create import list"})
		return
	}

	c.JSON(http.StatusCreated, list)
}

func (s *Server) handleUpdateImportList(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var list models.ImportList
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid import list data"})
		return
	}

	list.ID = id
	if err := s.services.ImportListService.UpdateImportList(&list); err != nil {
		s.logger.Error("Failed to update import list", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update import list"})
		return
	}

	c.JSON(http.StatusOK, list)
}

func (s *Server) handleDeleteImportList(c *gin.Context) {
	s.handleDeleteByID(c, "import list", s.services.ImportListService.DeleteImportList)
}

func (s *Server) handleTestImportList(c *gin.Context) {
	var list models.ImportList
	if err := c.ShouldBindJSON(&list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid import list data"})
		return
	}

	result, err := s.services.ImportListService.TestImportList(&list)
	if err != nil {
		s.logger.Error("Failed to test import list", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test import list"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleSyncImportList(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.services.ImportListService.SyncImportList(id)
	if err != nil {
		s.logger.Error("Failed to sync import list", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync import list"})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleSyncAllImportLists(c *gin.Context) {
	results, err := s.services.ImportListService.SyncAllImportLists()
	if err != nil {
		s.logger.Error("Failed to sync all import lists", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync import lists"})
		return
	}

	c.JSON(http.StatusOK, results)
}

func (s *Server) handleGetImportListStats(c *gin.Context) {
	stats, err := s.services.ImportListService.GetImportListStats()
	if err != nil {
		s.logger.Error("Failed to get import list stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get import list statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (s *Server) handleGetImportListMovies(c *gin.Context) {
	var listID *int
	if listIDStr := c.Query("importListId"); listIDStr != "" {
		if id, err := strconv.Atoi(listIDStr); err == nil {
			listID = &id
		}
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	movies, err := s.services.ImportListService.GetImportListMovies(listID, limit)
	if err != nil {
		s.logger.Error("Failed to get import list movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve import list movies"})
		return
	}

	c.JSON(http.StatusOK, movies)
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

// Quality Profile handlers
func (s *Server) handleGetQualityProfile(c *gin.Context) {
	s.handleGetByID(c, "quality profile", func(id int) (any, error) {
		return s.services.QualityService.GetQualityProfileByID(id)
	})
}

func (s *Server) handleCreateQualityProfile(c *gin.Context) {
	var profile models.QualityProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quality profile data"})
		return
	}

	if err := s.services.QualityService.CreateQualityProfile(&profile); err != nil {
		s.logger.Error("Failed to create quality profile", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create quality profile"})
		return
	}

	c.JSON(http.StatusCreated, profile)
}

func (s *Server) handleUpdateQualityProfile(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var profile models.QualityProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quality profile data"})
		return
	}

	profile.ID = id
	if err := s.services.QualityService.UpdateQualityProfile(&profile); err != nil {
		s.logger.Error("Failed to update quality profile", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quality profile"})
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (s *Server) handleDeleteQualityProfile(c *gin.Context) {
	s.handleDeleteByID(c, "quality profile", s.services.QualityService.DeleteQualityProfile)
}

// Quality Definition handlers
func (s *Server) handleGetQualityDefinitions(c *gin.Context) {
	definitions, err := s.services.QualityService.GetQualityDefinitions()
	if err != nil {
		s.logger.Error("Failed to get quality definitions", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve quality definitions"})
		return
	}
	c.JSON(http.StatusOK, definitions)
}

func (s *Server) handleGetQualityDefinition(c *gin.Context) {
	s.handleGetByID(c, "quality definition", func(id int) (any, error) {
		return s.services.QualityService.GetQualityDefinitionByID(id)
	})
}

func (s *Server) handleUpdateQualityDefinition(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var definition models.QualityLevel
	if err := c.ShouldBindJSON(&definition); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid quality definition data"})
		return
	}

	definition.ID = id
	if err := s.services.QualityService.UpdateQualityDefinition(&definition); err != nil {
		s.logger.Error("Failed to update quality definition", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update quality definition"})
		return
	}

	c.JSON(http.StatusOK, definition)
}

// Custom Format handlers
func (s *Server) handleGetCustomFormats(c *gin.Context) {
	formats, err := s.services.QualityService.GetCustomFormats()
	if err != nil {
		s.logger.Error("Failed to get custom formats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve custom formats"})
		return
	}
	c.JSON(http.StatusOK, formats)
}

func (s *Server) handleGetCustomFormat(c *gin.Context) {
	s.handleGetByID(c, "custom format", func(id int) (any, error) {
		return s.services.QualityService.GetCustomFormatByID(id)
	})
}

func (s *Server) handleCreateCustomFormat(c *gin.Context) {
	var format models.CustomFormat
	if err := c.ShouldBindJSON(&format); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom format data"})
		return
	}

	if err := s.services.QualityService.CreateCustomFormat(&format); err != nil {
		s.logger.Error("Failed to create custom format", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom format"})
		return
	}

	c.JSON(http.StatusCreated, format)
}

func (s *Server) handleUpdateCustomFormat(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var format models.CustomFormat
	if err := c.ShouldBindJSON(&format); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom format data"})
		return
	}

	format.ID = id
	if err := s.services.QualityService.UpdateCustomFormat(&format); err != nil {
		s.logger.Error("Failed to update custom format", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update custom format"})
		return
	}

	c.JSON(http.StatusOK, format)
}

func (s *Server) handleDeleteCustomFormat(c *gin.Context) {
	s.handleDeleteByID(c, "custom format", s.services.QualityService.DeleteCustomFormat)
}

// Indexer handlers
func (s *Server) handleGetIndexer(c *gin.Context) {
	s.handleGetByID(c, "indexer", func(id int) (any, error) {
		return s.services.IndexerService.GetIndexerByID(id)
	})
}

func (s *Server) handleCreateIndexer(c *gin.Context) {
	var indexer models.Indexer
	if err := c.ShouldBindJSON(&indexer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid indexer data"})
		return
	}

	if err := s.services.IndexerService.CreateIndexer(&indexer); err != nil {
		s.logger.Error("Failed to create indexer", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create indexer"})
		return
	}

	c.JSON(http.StatusCreated, indexer)
}

func (s *Server) handleUpdateIndexer(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var indexer models.Indexer
	if err := c.ShouldBindJSON(&indexer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid indexer data"})
		return
	}

	indexer.ID = id
	if err := s.services.IndexerService.UpdateIndexer(&indexer); err != nil {
		s.logger.Error("Failed to update indexer", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update indexer"})
		return
	}

	c.JSON(http.StatusOK, indexer)
}

func (s *Server) handleDeleteIndexer(c *gin.Context) {
	s.handleDeleteByID(c, "indexer", s.services.IndexerService.DeleteIndexer)
}

func (s *Server) handleTestIndexer(c *gin.Context) {
	var indexer models.Indexer
	if err := c.ShouldBindJSON(&indexer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid indexer data"})
		return
	}

	result, err := s.services.IndexerService.TestIndexer(&indexer)
	if err != nil {
		s.logger.Error("Failed to test indexer", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to test indexer"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Movie discovery and metadata handlers
func (s *Server) handleMovieLookup(c *gin.Context) {
	term := c.Query("term")
	if term == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search term is required"})
		return
	}

	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	response, err := s.services.MetadataService.SearchMovies(term, page)
	if err != nil {
		s.logger.Error("Failed to lookup movies", "term", term, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup movies"})
		return
	}

	c.JSON(http.StatusOK, response.Results)
}

func (s *Server) handleMovieDiscoverPopular(c *gin.Context) {
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	response, err := s.services.MetadataService.GetPopularMovies(page)
	if err != nil {
		s.logger.Error("Failed to get popular movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get popular movies"})
		return
	}

	c.JSON(http.StatusOK, response.Results)
}

func (s *Server) handleMovieDiscoverTrending(c *gin.Context) {
	timeWindow := c.DefaultQuery("timeWindow", "week")
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsed, err := strconv.Atoi(pageStr); err == nil && parsed > 0 {
			page = parsed
		}
	}

	response, err := s.services.MetadataService.GetTrendingMovies(timeWindow, page)
	if err != nil {
		s.logger.Error("Failed to get trending movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get trending movies"})
		return
	}

	c.JSON(http.StatusOK, response.Results)
}

func (s *Server) handleMovieByTMDBID(c *gin.Context) {
	tmdbIDStr := c.Param("tmdbId")
	tmdbID, err := strconv.Atoi(tmdbIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid TMDB ID"})
		return
	}

	movie, err := s.services.MetadataService.LookupMovieByTMDBID(tmdbID)
	if err != nil {
		s.logger.Error("Failed to get movie by TMDB ID", "tmdbId", tmdbID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	c.JSON(http.StatusOK, movie)
}

func (s *Server) handleRefreshMovieMetadata(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.MetadataService.RefreshMovieMetadata(id); err != nil {
		s.logger.Error("Failed to refresh movie metadata", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh movie metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie metadata refreshed successfully"})
}

// Queue handlers
func (s *Server) handleGetQueue(c *gin.Context) {
	params := s.parseQueueQueryParams(c)

	queue, err := s.services.QueueService.GetQueue(
		params.MovieIDs, params.Protocol, params.Languages,
		params.Quality, params.Status, params.IncludeUnknownMovieItems)
	if err != nil {
		s.logger.Error("Failed to get queue", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve queue"})
		return
	}

	paginatedQueue := s.applyQueuePagination(queue, params)
	if !params.IncludeMovie {
		s.removeMovieDetails(paginatedQueue)
	}

	response := s.buildQueueResponse(paginatedQueue, params, len(queue))
	c.JSON(http.StatusOK, response)
}

// queueQueryParams holds parsed query parameters for queue endpoints
type queueQueryParams struct {
	MovieIDs                 []int
	Protocol                 *models.DownloadProtocol
	Languages                []int
	Quality                  []int
	Status                   []models.QueueStatus
	IncludeUnknownMovieItems bool
	IncludeMovie             bool
	Page                     int
	PageSize                 int
	SortKey                  string
	SortDirection            string
}

const (
	trueBoolString  = "true"
	falseBoolString = "false"
	httpsScheme     = "https"
)

func (s *Server) parseQueueQueryParams(c *gin.Context) queueQueryParams {
	var params queueQueryParams

	// Parse protocol
	if protocolStr := c.Query("protocol"); protocolStr != "" {
		p := models.DownloadProtocol(protocolStr)
		params.Protocol = &p
	}

	// Parse status
	if statusStr := c.Query("status"); statusStr != "" {
		st := models.QueueStatus(statusStr)
		params.Status = []models.QueueStatus{st}
	}

	params.IncludeUnknownMovieItems = c.DefaultQuery("includeUnknownMovieItems", falseBoolString) == trueBoolString
	params.IncludeMovie = c.DefaultQuery("includeMovie", falseBoolString) == trueBoolString
	params.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	params.PageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	params.SortKey = c.DefaultQuery("sortKey", "timeleft")
	params.SortDirection = c.DefaultQuery("sortDirection", "ascending")

	return params
}

func (s *Server) applyQueuePagination(queue []models.QueueItem, params queueQueryParams) []models.QueueItem {
	startIndex := (params.Page - 1) * params.PageSize
	endIndex := startIndex + params.PageSize
	if endIndex > len(queue) {
		endIndex = len(queue)
	}
	if startIndex < len(queue) {
		return queue[startIndex:endIndex]
	}
	return []models.QueueItem{}
}

func (s *Server) removeMovieDetails(queue []models.QueueItem) {
	for i := range queue {
		queue[i].Movie = nil
	}
}

func (s *Server) buildQueueResponse(queue []models.QueueItem, params queueQueryParams, totalRecords int) gin.H {
	return gin.H{
		"page":          params.Page,
		"pageSize":      params.PageSize,
		"sortKey":       params.SortKey,
		"sortDirection": params.SortDirection,
		"totalRecords":  totalRecords,
		"records":       queue,
	}
}

func (s *Server) handleGetQueueItem(c *gin.Context) {
	s.handleGetByID(c, "queue item", func(id int) (any, error) {
		return s.services.QueueService.GetQueueByID(id)
	})
}

func (s *Server) handleRemoveQueueItem(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	removeFromClient := c.DefaultQuery("removeFromClient", trueBoolString) == trueBoolString
	blocklist := c.DefaultQuery("blocklist", falseBoolString) == trueBoolString
	skipRedownload := c.DefaultQuery("skipRedownload", falseBoolString) == trueBoolString
	changeCategory := c.DefaultQuery("changeCategory", falseBoolString) == trueBoolString

	err = s.services.QueueService.RemoveQueueItem(id, removeFromClient, blocklist, skipRedownload, changeCategory)
	if err != nil {
		s.logger.Error("Failed to remove queue item", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove queue item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Queue item removed successfully"})
}

func (s *Server) handleRemoveQueueItemsBulk(c *gin.Context) {
	var bulkRequest models.QueueBulkResource
	if err := c.ShouldBindJSON(&bulkRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bulk request data"})
		return
	}

	removeFromClient := c.DefaultQuery("removeFromClient", trueBoolString) == trueBoolString
	blocklist := c.DefaultQuery("blocklist", falseBoolString) == trueBoolString
	skipRedownload := c.DefaultQuery("skipRedownload", falseBoolString) == trueBoolString
	changeCategory := c.DefaultQuery("changeCategory", falseBoolString) == trueBoolString

	err := s.services.QueueService.RemoveQueueItems(
		bulkRequest.IDs, removeFromClient, blocklist, skipRedownload, changeCategory)
	if err != nil {
		s.logger.Error("Failed to remove queue items", "count", len(bulkRequest.IDs), "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove queue items"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Queue items removed successfully"})
}

func (s *Server) handleGetQueueStats(c *gin.Context) {
	stats, err := s.services.QueueService.GetQueueStats()
	if err != nil {
		s.logger.Error("Failed to get queue stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get queue statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// History handlers

func (s *Server) handleGetHistory(c *gin.Context) {
	var req models.HistoryRequest

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			req.PageSize = ps
		}
	}

	req.SortKey = c.DefaultQuery("sortKey", "date")
	req.SortDir = c.DefaultQuery("sortDir", "desc")

	if movieID := c.Query("movieId"); movieID != "" {
		if mid, err := strconv.Atoi(movieID); err == nil {
			req.MovieID = &mid
		}
	}

	if eventType := c.Query("eventType"); eventType != "" {
		et := models.HistoryEventType(eventType)
		req.EventType = &et
	}

	if successful := c.Query("successful"); successful != "" {
		if s := successful == trueBoolString; successful == trueBoolString || successful == "false" {
			req.Successful = &s
		}
	}

	req.DownloadID = c.Query("downloadId")

	if since := c.Query("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			req.Since = &t
		}
	}

	if until := c.Query("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			req.Until = &t
		}
	}

	response, err := s.services.HistoryService.GetHistory(req)
	if err != nil {
		s.logger.Error("Failed to get history", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleGetHistoryByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid history ID"})
		return
	}

	history, err := s.services.HistoryService.GetHistoryByID(id)
	if err != nil {
		if err.Error() == "history record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "History record not found"})
			return
		}
		s.logger.Error("Failed to get history record", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history record"})
		return
	}

	c.JSON(http.StatusOK, history)
}

func (s *Server) handleDeleteHistoryRecord(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid history ID"})
		return
	}

	err = s.services.HistoryService.DeleteHistoryRecord(id)
	if err != nil {
		if err.Error() == "history record not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "History record not found"})
			return
		}
		s.logger.Error("Failed to delete history record", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete history record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "History record deleted successfully"})
}

func (s *Server) handleGetHistoryStats(c *gin.Context) {
	stats, err := s.services.HistoryService.GetHistoryStats()
	if err != nil {
		s.logger.Error("Failed to get history stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get history statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Activity handlers

func (s *Server) handleGetActivity(c *gin.Context) {
	var req models.ActivityRequest

	// Parse query parameters
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}

	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			req.PageSize = ps
		}
	}

	if activityType := c.Query("type"); activityType != "" {
		at := models.ActivityType(activityType)
		req.Type = &at
	}

	if status := c.Query("status"); status != "" {
		as := models.ActivityStatus(status)
		req.Status = &as
	}

	if movieID := c.Query("movieId"); movieID != "" {
		if mid, err := strconv.Atoi(movieID); err == nil {
			req.MovieID = &mid
		}
	}

	if since := c.Query("since"); since != "" {
		if t, err := time.Parse(time.RFC3339, since); err == nil {
			req.Since = &t
		}
	}

	if until := c.Query("until"); until != "" {
		if t, err := time.Parse(time.RFC3339, until); err == nil {
			req.Until = &t
		}
	}

	response, err := s.services.HistoryService.GetActivities(req)
	if err != nil {
		s.logger.Error("Failed to get activities", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve activities"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleGetActivityByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	activity, err := s.services.HistoryService.GetActivityByID(id)
	if err != nil {
		if err.Error() == "activity not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Activity not found"})
			return
		}
		s.logger.Error("Failed to get activity", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve activity"})
		return
	}

	c.JSON(http.StatusOK, activity)
}

func (s *Server) handleDeleteActivity(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid activity ID"})
		return
	}

	err = s.services.HistoryService.DeleteActivity(id)
	if err != nil {
		if err.Error() == "activity not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Activity not found"})
			return
		}
		s.logger.Error("Failed to delete activity", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete activity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity deleted successfully"})
}

func (s *Server) handleGetRunningActivities(c *gin.Context) {
	activities, err := s.services.HistoryService.GetRunningActivities()
	if err != nil {
		s.logger.Error("Failed to get running activities", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve running activities"})
		return
	}

	c.JSON(http.StatusOK, activities)
}

// Configuration handlers

// Host Configuration handlers
func (s *Server) handleGetHostConfig(c *gin.Context) {
	config, err := s.services.ConfigService.GetHostConfig()
	if err != nil {
		s.logger.Error("Failed to get host config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve host configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (s *Server) handleUpdateHostConfig(c *gin.Context) {
	var config models.HostConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid host configuration data"})
		return
	}

	err := s.services.ConfigService.UpdateHostConfig(&config)
	if err != nil {
		s.logger.Error("Failed to update host config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update host configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Naming Configuration handlers
func (s *Server) handleGetNamingConfig(c *gin.Context) {
	config, err := s.services.ConfigService.GetNamingConfig()
	if err != nil {
		s.logger.Error("Failed to get naming config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve naming configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (s *Server) handleUpdateNamingConfig(c *gin.Context) {
	var config models.NamingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid naming configuration data"})
		return
	}

	err := s.services.ConfigService.UpdateNamingConfig(&config)
	if err != nil {
		s.logger.Error("Failed to update naming config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update naming configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (s *Server) handleGetNamingTokens(c *gin.Context) {
	tokens := s.services.ConfigService.GetNamingTokens()
	c.JSON(http.StatusOK, tokens)
}

// Media Management Configuration handlers
func (s *Server) handleGetMediaManagementConfig(c *gin.Context) {
	config, err := s.services.ConfigService.GetMediaManagementConfig()
	if err != nil {
		s.logger.Error("Failed to get media management config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve media management configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

func (s *Server) handleUpdateMediaManagementConfig(c *gin.Context) {
	var config models.MediaManagementConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid media management configuration data"})
		return
	}

	err := s.services.ConfigService.UpdateMediaManagementConfig(&config)
	if err != nil {
		s.logger.Error("Failed to update media management config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update media management configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// Root Folder handlers
func (s *Server) handleGetRootFolders(c *gin.Context) {
	rootFolders, err := s.services.ConfigService.GetRootFolders()
	if err != nil {
		s.logger.Error("Failed to get root folders", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve root folders"})
		return
	}

	c.JSON(http.StatusOK, rootFolders)
}

func (s *Server) handleGetRootFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid root folder ID"})
		return
	}

	rootFolder, err := s.services.ConfigService.GetRootFolderByID(id)
	if err != nil {
		if err.Error() == "root folder not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Root folder not found"})
			return
		}
		s.logger.Error("Failed to get root folder", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve root folder"})
		return
	}

	c.JSON(http.StatusOK, rootFolder)
}

func (s *Server) handleCreateRootFolder(c *gin.Context) {
	var rootFolder models.RootFolder
	if err := c.ShouldBindJSON(&rootFolder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid root folder data"})
		return
	}

	err := s.services.ConfigService.CreateRootFolder(&rootFolder)
	if err != nil {
		s.logger.Error("Failed to create root folder", "path", rootFolder.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create root folder"})
		return
	}

	c.JSON(http.StatusCreated, rootFolder)
}

func (s *Server) handleUpdateRootFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid root folder ID"})
		return
	}

	var rootFolder models.RootFolder
	if err := c.ShouldBindJSON(&rootFolder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid root folder data"})
		return
	}

	rootFolder.ID = id
	err = s.services.ConfigService.UpdateRootFolder(&rootFolder)
	if err != nil {
		s.logger.Error("Failed to update root folder", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update root folder"})
		return
	}

	c.JSON(http.StatusOK, rootFolder)
}

func (s *Server) handleDeleteRootFolder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid root folder ID"})
		return
	}

	err = s.services.ConfigService.DeleteRootFolder(id)
	if err != nil {
		if err.Error() == "root folder not found: root folder not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Root folder not found"})
			return
		}
		s.logger.Error("Failed to delete root folder", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete root folder"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Root folder deleted successfully"})
}

// Configuration stats handler
func (s *Server) handleGetConfigStats(c *gin.Context) {
	stats, err := s.services.ConfigService.GetConfigurationStats()
	if err != nil {
		s.logger.Error("Failed to get configuration stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configuration statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Search & Release Management handlers

// Release handlers
func (s *Server) handleGetReleases(c *gin.Context) {
	var filter models.ReleaseFilter

	// Parse query parameters for filtering
	if movieIDStr := c.Query("movieId"); movieIDStr != "" {
		if movieID, err := strconv.Atoi(movieIDStr); err == nil {
			filter.MovieIDs = []int{movieID}
		}
	}

	if indexerIDStr := c.Query("indexerId"); indexerIDStr != "" {
		if indexerID, err := strconv.Atoi(indexerIDStr); err == nil {
			filter.IndexerIDs = []int{indexerID}
		}
	}

	if protocol := c.Query("protocol"); protocol != "" {
		p := models.Protocol(protocol)
		filter.Protocol = []models.Protocol{p}
	}

	if status := c.Query("status"); status != "" {
		s := models.ReleaseStatus(status)
		filter.Status = []models.ReleaseStatus{s}
	}

	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	releases, total, err := s.services.SearchService.GetReleases(&filter, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get releases", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve releases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"releases": releases,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (s *Server) handleGetRelease(c *gin.Context) {
	s.handleGetByID(c, "release", func(id int) (any, error) {
		return s.services.SearchService.GetReleaseByID(id)
	})
}

func (s *Server) handleDeleteRelease(c *gin.Context) {
	s.handleDeleteByID(c, "release", s.services.SearchService.DeleteRelease)
}

func (s *Server) handleGetReleaseStats(c *gin.Context) {
	stats, err := s.services.SearchService.GetReleaseStats()
	if err != nil {
		s.logger.Error("Failed to get release stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get release statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Search handlers
func (s *Server) handleSearchReleases(c *gin.Context) {
	var request models.SearchRequest

	// Parse search parameters
	if movieIDStr := c.Query("movieId"); movieIDStr != "" {
		if movieID, err := strconv.Atoi(movieIDStr); err == nil {
			request.MovieID = &movieID
		}
	}

	if tmdbIDStr := c.Query("tmdbId"); tmdbIDStr != "" {
		if tmdbID, err := strconv.Atoi(tmdbIDStr); err == nil {
			request.TmdbID = &tmdbID
		}
	}

	request.ImdbID = c.Query("imdbId")
	request.Title = c.Query("title")

	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			request.Year = &year
		}
	}

	if protocol := c.Query("protocol"); protocol != "" {
		p := models.Protocol(protocol)
		request.Protocol = &p
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			request.Limit = limit
		}
	} else {
		request.Limit = 50
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			request.Offset = offset
		}
	}

	request.SortBy = c.DefaultQuery("sortBy", "qualityWeight")
	request.SortOrder = c.DefaultQuery("sortOrder", "desc")
	request.Source = models.ReleaseSourceSearch

	forceSearch := c.DefaultQuery("forceSearch", "false") == trueBoolString

	response, err := s.services.SearchService.SearchReleases(&request, forceSearch)
	if err != nil {
		s.logger.Error("Failed to search releases", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search releases"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleSearchMovieReleases(c *gin.Context) {
	movieID, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	forceSearch := c.DefaultQuery("forceSearch", "false") == trueBoolString

	response, err := s.services.SearchService.SearchMovieReleases(movieID, forceSearch)
	if err != nil {
		s.logger.Error("Failed to search movie releases", "movieId", movieID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search movie releases"})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleInteractiveSearch(c *gin.Context) {
	var request models.SearchRequest

	// Parse search parameters
	if movieIDStr := c.Query("movieId"); movieIDStr != "" {
		if movieID, err := strconv.Atoi(movieIDStr); err == nil {
			request.MovieID = &movieID
		}
	}

	if tmdbIDStr := c.Query("tmdbId"); tmdbIDStr != "" {
		if tmdbID, err := strconv.Atoi(tmdbIDStr); err == nil {
			request.TmdbID = &tmdbID
		}
	}

	request.ImdbID = c.Query("imdbId")
	request.Title = c.Query("title")

	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			request.Year = &year
		}
	}

	if protocol := c.Query("protocol"); protocol != "" {
		p := models.Protocol(protocol)
		request.Protocol = &p
	}

	request.Limit = 100
	request.SortBy = c.DefaultQuery("sortBy", "qualityWeight")
	request.SortOrder = c.DefaultQuery("sortOrder", "desc")

	response, err := s.services.SearchService.InteractiveSearch(&request)
	if err != nil {
		s.logger.Error("Failed to perform interactive search", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform interactive search"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Grab handler
func (s *Server) handleGrabRelease(c *gin.Context) {
	var request models.GrabRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid grab request data"})
		return
	}

	response, err := s.services.SearchService.GrabRelease(&request)
	if err != nil {
		s.logger.Error("Failed to grab release", "guid", request.GUID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to grab release"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// Task Management Handlers

// handleGetTasks retrieves tasks with optional filtering
func (s *Server) handleGetTasks(c *gin.Context) {
	// Parse query parameters
	status := c.Query("status")
	commandName := c.Query("command")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit parameter"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset parameter"})
		return
	}

	tasks, total, err := s.services.TaskService.ListTasks(models.TaskStatus(status), commandName, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tasks"})
		return
	}

	response := gin.H{
		"records":    tasks,
		"total":      total,
		"page":       (offset / limit) + 1,
		"pageSize":   limit,
		"totalPages": (int(total) + limit - 1) / limit,
	}

	c.JSON(http.StatusOK, response)
}

// handleGetTask retrieves a single task by ID
func (s *Server) handleGetTask(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := s.services.TaskService.GetTask(id)
	if err != nil {
		s.logger.Error("Failed to get task", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// handleQueueTask queues a new task for execution
func (s *Server) handleQueueTask(c *gin.Context) {
	var request struct {
		Name        string              `json:"name" binding:"required"`
		CommandName string              `json:"commandName" binding:"required"`
		Body        models.TaskBody     `json:"body"`
		Priority    models.TaskPriority `json:"priority"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task request data"})
		return
	}

	// Set default priority if not specified
	if request.Priority == "" {
		request.Priority = models.TaskPriorityNormal
	}

	task, err := s.services.TaskService.QueueTask(
		request.Name,
		request.CommandName,
		request.Body,
		request.Priority,
		models.TaskTriggerAPI,
	)
	if err != nil {
		s.logger.Error("Failed to queue task", "command", request.CommandName, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// handleCancelTask cancels a running or queued task
func (s *Server) handleCancelTask(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.TaskService.CancelTask(id); err != nil {
		s.logger.Error("Failed to cancel task", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task cancelled"})
}

// handleGetQueueStatus returns the current status of all task queues
func (s *Server) handleGetQueueStatus(c *gin.Context) {
	status := s.services.TaskService.GetQueueStatus()
	c.JSON(http.StatusOK, status)
}

// handleGetScheduledTasks retrieves all scheduled tasks
func (s *Server) handleGetScheduledTasks(c *gin.Context) {
	scheduledTasks, err := s.services.TaskService.GetScheduledTasks()
	if err != nil {
		s.logger.Error("Failed to get scheduled tasks", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve scheduled tasks"})
		return
	}

	c.JSON(http.StatusOK, scheduledTasks)
}

// handleCreateScheduledTask creates a new scheduled task
func (s *Server) handleCreateScheduledTask(c *gin.Context) {
	var request struct {
		Name        string              `json:"name" binding:"required"`
		CommandName string              `json:"commandName" binding:"required"`
		Body        models.TaskBody     `json:"body"`
		Interval    int64               `json:"interval" binding:"required"` // milliseconds
		Priority    models.TaskPriority `json:"priority"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled task data"})
		return
	}

	// Set default priority if not specified
	if request.Priority == "" {
		request.Priority = models.TaskPriorityNormal
	}

	interval := time.Duration(request.Interval) * time.Millisecond
	scheduledTask, err := s.services.TaskService.CreateScheduledTask(
		request.Name,
		request.CommandName,
		request.Body,
		interval,
		request.Priority,
	)
	if err != nil {
		s.logger.Error("Failed to create scheduled task", "name", request.Name, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create scheduled task"})
		return
	}

	c.JSON(http.StatusCreated, scheduledTask)
}

// handleUpdateScheduledTask updates a scheduled task
func (s *Server) handleUpdateScheduledTask(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var request struct {
		Name     *string              `json:"name,omitempty"`
		Body     *models.TaskBody     `json:"body,omitempty"`
		Interval *int64               `json:"interval,omitempty"` // milliseconds
		Priority *models.TaskPriority `json:"priority,omitempty"`
		Enabled  *bool                `json:"enabled,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid scheduled task data"})
		return
	}

	updates := make(map[string]interface{})
	if request.Name != nil {
		updates["name"] = *request.Name
	}
	if request.Body != nil {
		updates["body"] = *request.Body
	}
	if request.Interval != nil {
		updates["interval"] = time.Duration(*request.Interval) * time.Millisecond
	}
	if request.Priority != nil {
		updates["priority"] = *request.Priority
	}
	if request.Enabled != nil {
		updates["enabled"] = *request.Enabled
	}

	if err := s.services.TaskService.UpdateScheduledTask(id, updates); err != nil {
		s.logger.Error("Failed to update scheduled task", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update scheduled task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled task updated"})
}

// handleDeleteScheduledTask deletes a scheduled task
func (s *Server) handleDeleteScheduledTask(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.TaskService.DeleteScheduledTask(id); err != nil {
		s.logger.Error("Failed to delete scheduled task", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete scheduled task"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Scheduled task deleted"})
}

// Command handlers for specific task operations

// handleRefreshMovie queues a refresh task for a specific movie
func (s *Server) handleRefreshMovie(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body := models.TaskBody{
		"movieId": id,
	}

	task, err := s.services.TaskService.QueueTask(
		fmt.Sprintf("Refresh Movie - ID %d", id),
		"RefreshMovie",
		body,
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)
	if err != nil {
		s.logger.Error("Failed to queue movie refresh task", "movieId", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue movie refresh"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// handleRefreshAllMovies queues a task to refresh all movies
func (s *Server) handleRefreshAllMovies(c *gin.Context) {
	task, err := s.services.TaskService.QueueTask(
		"Refresh All Movies",
		"RefreshAllMovies",
		models.TaskBody{},
		models.TaskPriorityNormal,
		models.TaskTriggerAPI,
	)
	if err != nil {
		s.logger.Error("Failed to queue refresh all movies task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue refresh all movies"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// handleRunHealthCheck queues a system health check task
func (s *Server) handleRunHealthCheck(c *gin.Context) {
	task, err := s.services.TaskService.QueueTask(
		"System Health Check",
		"HealthCheck",
		models.TaskBody{},
		models.TaskPriorityHigh,
		models.TaskTriggerAPI,
	)
	if err != nil {
		s.logger.Error("Failed to queue health check task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue health check"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// handleRunCleanup queues a cleanup task
func (s *Server) handleRunCleanup(c *gin.Context) {
	task, err := s.services.TaskService.QueueTask(
		"System Cleanup",
		"Cleanup",
		models.TaskBody{},
		models.TaskPriorityLow,
		models.TaskTriggerAPI,
	)
	if err != nil {
		s.logger.Error("Failed to queue cleanup task", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue cleanup"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// File Organization and Import Handlers

// handleGetFileOrganizations returns file organization records
func (s *Server) handleGetFileOrganizations(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	organizations, err := s.services.FileOrganizationService.GetFileOrganizations(limit, offset)
	if err != nil {
		s.logger.Error("Failed to get file organizations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file organizations"})
		return
	}

	c.JSON(http.StatusOK, organizations)
}

// handleGetFileOrganizationByID returns a specific file organization record
func (s *Server) handleGetFileOrganizationByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	organization, err := s.services.FileOrganizationService.GetFileOrganizationByID(id)
	if err != nil {
		s.logger.Error("Failed to get file organization", "id", id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File organization not found"})
		return
	}

	c.JSON(http.StatusOK, organization)
}

// handleRetryFailedOrganizations retries failed file organization operations
func (s *Server) handleRetryFailedOrganizations(c *gin.Context) {
	if err := s.services.FileOrganizationService.RetryFailedOrganizations(); err != nil {
		s.logger.Error("Failed to retry file organizations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retry file organizations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Retry initiated for failed file organizations"})
}

// handleScanDirectory scans a directory for importable files
func (s *Server) handleScanDirectory(c *gin.Context) {
	var request struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	files, err := s.services.FileOrganizationService.ScanDirectory(request.Path)
	if err != nil {
		s.logger.Error("Failed to scan directory", "path", request.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan directory"})
		return
	}

	c.JSON(http.StatusOK, files)
}

// handleProcessImport processes files for import
func (s *Server) handleProcessImport(c *gin.Context) {
	var request struct {
		Path       string                    `json:"path" binding:"required"`
		ImportMode models.ImportDecisionType `json:"importMode"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	options := &services.ImportOptions{
		ImportMode: request.ImportMode,
	}

	if options.ImportMode == "" {
		options.ImportMode = models.ImportDecisionApproved
	}

	result, err := s.services.ImportService.ProcessImport(c.Request.Context(), request.Path, options)
	if err != nil {
		s.logger.Error("Failed to process import", "path", request.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process import"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleGetManualImports returns manual import candidates
func (s *Server) handleGetManualImports(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path parameter is required"})
		return
	}

	manualImports, err := s.services.ImportService.GetManualImports(path)
	if err != nil {
		s.logger.Error("Failed to get manual imports", "path", path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get manual imports"})
		return
	}

	c.JSON(http.StatusOK, manualImports)
}

// handleProcessManualImport processes a manual import
func (s *Server) handleProcessManualImport(c *gin.Context) {
	var manualImport models.ManualImport

	if err := c.ShouldBindJSON(&manualImport); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manual import data"})
		return
	}

	if err := s.services.ImportService.ProcessManualImport(c.Request.Context(), &manualImport); err != nil {
		s.logger.Error("Failed to process manual import", "path", manualImport.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process manual import"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Manual import processed successfully"})
}

// handlePreviewNaming generates a preview of file naming
func (s *Server) handlePreviewNaming(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid movie ID"})
		return
	}

	movie, err := s.services.MovieService.GetByID(movieID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Movie not found"})
		return
	}

	config, err := s.services.NamingService.GetNamingConfig()
	if err != nil {
		s.logger.Error("Failed to get naming config", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get naming config"})
		return
	}

	preview, err := s.services.NamingService.PreviewNaming(movie, config)
	if err != nil {
		s.logger.Error("Failed to generate naming preview", "movieId", movieID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate naming preview"})
		return
	}

	c.JSON(http.StatusOK, preview)
}

// handleGetFileOperations returns file operations
func (s *Server) handleGetFileOperations(c *gin.Context) {
	status := models.FileOperationStatus(c.Query("status"))
	operationType := models.FileOperationType(c.Query("type"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	operations, err := s.services.FileOperationService.GetOperations(status, operationType, limit, offset)
	if err != nil {
		s.logger.Error("Failed to get file operations", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file operations"})
		return
	}

	c.JSON(http.StatusOK, operations)
}

// handleGetFileOperation returns a specific file operation
func (s *Server) handleGetFileOperation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	operation, err := s.services.FileOperationService.GetOperationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File operation not found"})
		return
	}

	c.JSON(http.StatusOK, operation)
}

// handleCancelFileOperation cancels a file operation
func (s *Server) handleCancelFileOperation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation ID"})
		return
	}

	if err := s.services.FileOperationService.CancelOperation(id); err != nil {
		s.logger.Error("Failed to cancel file operation", "id", id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File operation canceled"})
}

// handleGetFileOperationSummary returns file operation summary
func (s *Server) handleGetFileOperationSummary(c *gin.Context) {
	summary, err := s.services.FileOperationService.GetOperationSummary()
	if err != nil {
		s.logger.Error("Failed to get file operation summary", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get operation summary"})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// handleExtractMediaInfo extracts media info from a file
func (s *Server) handleExtractMediaInfo(c *gin.Context) {
	var request struct {
		Path string `json:"path" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	mediaInfo, err := s.services.MediaInfoService.ExtractMediaInfo(c.Request.Context(), request.Path)
	if err != nil {
		s.logger.Error("Failed to extract media info", "path", request.Path, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract media info"})
		return
	}

	c.JSON(http.StatusOK, mediaInfo)
}

// Notification handlers
func (s *Server) handleGetNotifications(c *gin.Context) {
	notifications, err := s.services.NotificationService.GetNotifications()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

func (s *Server) handleGetNotification(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	notification, err := s.services.NotificationService.GetNotificationByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}
	c.JSON(http.StatusOK, notification)
}

func (s *Server) handleCreateNotification(c *gin.Context) {
	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.NotificationService.CreateNotification(&notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

func (s *Server) handleUpdateNotification(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification.ID = id
	if err := s.services.NotificationService.UpdateNotification(&notification); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

func (s *Server) handleDeleteNotification(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}

	if err := s.services.NotificationService.DeleteNotification(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (s *Server) handleTestNotification(c *gin.Context) {
	var notification models.Notification
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := s.services.NotificationService.TestNotification(&notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *Server) handleGetNotificationProviders(c *gin.Context) {
	providers, err := s.services.NotificationService.GetProviderInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, providers)
}

func (s *Server) handleGetNotificationProviderFields(c *gin.Context) {
	providerType := models.NotificationType(c.Param("type"))
	fields, err := s.services.NotificationService.GetProviderFields(providerType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fields)
}

func (s *Server) handleGetNotificationHistory(c *gin.Context) {
	limit := DefaultPageSize
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	history, err := s.services.NotificationService.GetNotificationHistory(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// Calendar API handlers

// handleGetCalendar retrieves calendar events based on query parameters
func (s *Server) handleGetCalendar(c *gin.Context) {
	request, err := s.parseCalendarRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := s.services.CalendarService.GetCalendarEvents(request)
	if err != nil {
		s.logger.Error("Failed to get calendar events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve calendar events"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleGetCalendarFeed generates an iCal feed for external calendar applications
func (s *Server) handleGetCalendarFeed(c *gin.Context) {
	// Parse feed configuration from query parameters
	config := s.services.ICalService.ParseICalFeedParams(s.extractQueryParams(c))

	// Validate configuration
	if err := s.services.ICalService.ValidateICalFeedConfig(config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check authentication if required
	if config.RequireAuth {
		providedKey := c.Query("passKey")
		if providedKey == "" || providedKey != config.PassKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing pass key"})
			return
		}
	}

	// Generate base URL for event links
	baseURL := fmt.Sprintf("%s://%s", s.getScheme(c), c.Request.Host)
	if s.config.Server.URLBase != "" {
		baseURL += s.config.Server.URLBase
	}

	// Generate iCal feed
	icalData, err := s.services.ICalService.GenerateICalFeed(config, baseURL)
	if err != nil {
		s.logger.Error("Failed to generate iCal feed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate calendar feed"})
		return
	}

	// Set appropriate headers for iCal content
	c.Header("Content-Type", "text/calendar; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\"radarr-calendar.ics\"")
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	c.String(http.StatusOK, icalData)
}

// handleGetCalendarConfiguration retrieves calendar configuration settings
func (s *Server) handleGetCalendarConfiguration(c *gin.Context) {
	config, err := s.services.CalendarService.GetCalendarConfiguration()
	if err != nil {
		s.logger.Error("Failed to get calendar configuration", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve calendar configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// handleUpdateCalendarConfiguration updates calendar configuration settings
func (s *Server) handleUpdateCalendarConfiguration(c *gin.Context) {
	var config models.CalendarConfiguration
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid calendar configuration data"})
		return
	}

	if err := s.services.CalendarService.UpdateCalendarConfiguration(&config); err != nil {
		s.logger.Error("Failed to update calendar configuration", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update calendar configuration"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// handleGetCalendarStats retrieves calendar statistics
func (s *Server) handleGetCalendarStats(c *gin.Context) {
	stats, err := s.services.CalendarService.GetCalendarStats()
	if err != nil {
		s.logger.Error("Failed to get calendar statistics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve calendar statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleRefreshCalendar forces a refresh of calendar events and clears cache
func (s *Server) handleRefreshCalendar(c *gin.Context) {
	if err := s.services.CalendarService.RefreshCalendarEvents(); err != nil {
		s.logger.Error("Failed to refresh calendar events", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh calendar events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Calendar events refreshed successfully"})
}

// handleGetCalendarFeedURL generates a URL for accessing the iCal feed
func (s *Server) handleGetCalendarFeedURL(c *gin.Context) {
	config := s.services.ICalService.ParseICalFeedParams(s.extractQueryParams(c))

	baseURL := fmt.Sprintf("%s://%s", s.getScheme(c), c.Request.Host)
	if s.config.Server.URLBase != "" {
		baseURL += s.config.Server.URLBase
	}

	feedURL := s.services.ICalService.GenerateICalFeedURL(baseURL, config)

	c.JSON(http.StatusOK, gin.H{
		"url":    feedURL,
		"config": config,
	})
}

// Helper functions for calendar handlers

// parseCalendarRequest parses query parameters into a calendar request
func (s *Server) parseCalendarRequest(c *gin.Context) (*models.CalendarRequest, error) {
	request := &models.CalendarRequest{}

	if err := s.parseCalendarDateRange(c, request); err != nil {
		return nil, err
	}

	s.parseCalendarViewAndTypes(c, request)
	s.parseCalendarFilters(c, request)
	s.parseCalendarIncludeFlags(c, request)

	return request, nil
}

// parseCalendarDateRange parses start and end date parameters
func (s *Server) parseCalendarDateRange(c *gin.Context, request *models.CalendarRequest) error {
	if startStr := c.Query("start"); startStr != "" {
		start, err := s.parseCalendarDate(startStr)
		if err != nil {
			return fmt.Errorf("invalid start date format: %s", startStr)
		}
		request.Start = start
	}

	if endStr := c.Query("end"); endStr != "" {
		end, err := s.parseCalendarDate(endStr)
		if err != nil {
			return fmt.Errorf("invalid end date format: %s", endStr)
		}
		request.End = end
	}

	return nil
}

// parseCalendarDate parses a date string in multiple formats
func (s *Server) parseCalendarDate(dateStr string) (*time.Time, error) {
	// Try RFC3339 format first
	if date, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return &date, nil
	}

	// Try date-only format
	if date, err := time.Parse("2006-01-02", dateStr); err == nil {
		return &date, nil
	}

	return nil, fmt.Errorf("unsupported date format")
}

// parseCalendarViewAndTypes parses view type and event types
func (s *Server) parseCalendarViewAndTypes(c *gin.Context, request *models.CalendarRequest) {
	// Parse view type
	if viewStr := c.Query("view"); viewStr != "" {
		request.View = models.CalendarViewType(viewStr)
	}

	// Parse event types
	if eventTypesStr := c.Query("eventTypes"); eventTypesStr != "" {
		eventTypeStrs := strings.Split(eventTypesStr, ",")
		for _, eventTypeStr := range eventTypeStrs {
			eventType := models.CalendarEventType(strings.TrimSpace(eventTypeStr))
			request.EventTypes = append(request.EventTypes, eventType)
		}
	}
}

// parseCalendarFilters parses movie IDs, tags, and monitored filter
func (s *Server) parseCalendarFilters(c *gin.Context, request *models.CalendarRequest) {
	// Parse movie IDs
	if movieIDsStr := c.Query("movieIds"); movieIDsStr != "" {
		movieIDStrs := strings.Split(movieIDsStr, ",")
		for _, movieIDStr := range movieIDStrs {
			if movieID, err := strconv.Atoi(strings.TrimSpace(movieIDStr)); err == nil {
				request.MovieIDs = append(request.MovieIDs, movieID)
			}
		}
	}

	// Parse tags
	if tagsStr := c.Query("tags"); tagsStr != "" {
		tagStrs := strings.Split(tagsStr, ",")
		for _, tagStr := range tagStrs {
			if tag, err := strconv.Atoi(strings.TrimSpace(tagStr)); err == nil {
				request.Tags = append(request.Tags, tag)
			}
		}
	}

	// Parse monitored flag
	if monitoredStr := c.Query("monitored"); monitoredStr != "" {
		monitored := strings.ToLower(monitoredStr) == trueBoolString
		request.Monitored = &monitored
	}
}

// parseCalendarIncludeFlags parses include flags for calendar request
func (s *Server) parseCalendarIncludeFlags(c *gin.Context, request *models.CalendarRequest) {
	request.IncludeUnmonitored = c.DefaultQuery("includeUnmonitored", "false") == trueBoolString
	request.IncludeMovieInformation = c.DefaultQuery("includeMovieInformation", trueBoolString) == trueBoolString
}

// extractQueryParams extracts all query parameters as a map
func (s *Server) extractQueryParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

// getScheme determines the HTTP scheme (http or https)
func (s *Server) getScheme(c *gin.Context) string {
	if s.config.Server.EnableSSL {
		return httpsScheme
	}

	if c.GetHeader("X-Forwarded-Proto") == httpsScheme {
		return httpsScheme
	}

	return "http"
}
