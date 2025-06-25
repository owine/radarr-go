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
	trueBoolString = "true"
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

	params.IncludeUnknownMovieItems = c.DefaultQuery("includeUnknownMovieItems", "false") == trueBoolString
	params.IncludeMovie = c.DefaultQuery("includeMovie", "false") == trueBoolString
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
	blocklist := c.DefaultQuery("blocklist", "false") == trueBoolString
	skipRedownload := c.DefaultQuery("skipRedownload", "false") == trueBoolString
	changeCategory := c.DefaultQuery("changeCategory", "false") == trueBoolString

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
	blocklist := c.DefaultQuery("blocklist", "false") == trueBoolString
	skipRedownload := c.DefaultQuery("skipRedownload", "false") == trueBoolString
	changeCategory := c.DefaultQuery("changeCategory", "false") == trueBoolString

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
		if s := successful == "true"; successful == "true" || successful == "false" {
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
