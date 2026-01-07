//nolint:revive // "api" is a standard package name for API layers
package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/models"
)

// === WANTED MOVIES HANDLERS ===

// handleGetMissingMovies handles GET /api/v3/wanted/missing
func (s *Server) handleGetMissingMovies(c *gin.Context) {
	filter := s.parseWantedMovieFilter(c)

	response, err := s.services.WantedMoviesService.GetMissingMovies(filter)
	if err != nil {
		s.logger.Error("Failed to get missing movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get missing movies"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleGetCutoffUnmetMovies handles GET /api/v3/wanted/cutoff
func (s *Server) handleGetCutoffUnmetMovies(c *gin.Context) {
	filter := s.parseWantedMovieFilter(c)

	response, err := s.services.WantedMoviesService.GetCutoffUnmetMovies(filter)
	if err != nil {
		s.logger.Error("Failed to get cutoff unmet movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get cutoff unmet movies"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleGetAllWantedMovies handles GET /api/v3/wanted
func (s *Server) handleGetAllWantedMovies(c *gin.Context) {
	filter := s.parseWantedMovieFilter(c)

	response, err := s.services.WantedMoviesService.GetAllWanted(filter)
	if err != nil {
		s.logger.Error("Failed to get wanted movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wanted movies"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleGetWantedStats handles GET /api/v3/wanted/stats
func (s *Server) handleGetWantedStats(c *gin.Context) {
	stats, err := s.services.WantedMoviesService.GetWantedStats()
	if err != nil {
		s.logger.Error("Failed to get wanted movies stats", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wanted movies statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// handleGetWantedMovie handles GET /api/v3/wanted/:id
func (s *Server) handleGetWantedMovie(c *gin.Context) {
	s.handleGetByID(c, "wanted movie", func(id int) (any, error) {
		return s.services.WantedMoviesService.GetByID(id)
	})
}

// handleTriggerWantedSearch handles POST /api/v3/wanted/search
func (s *Server) handleTriggerWantedSearch(c *gin.Context) {
	var searchTrigger models.WantedSearchTrigger
	if err := c.ShouldBindJSON(&searchTrigger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %s", err.Error())})
		return
	}

	moviesToSearch, err := s.getMoviesToSearch(&searchTrigger, c)
	if err != nil {
		return // Error response already sent
	}

	searchedCount := s.executeMovieSearches(moviesToSearch, searchTrigger.ForceSearch)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Search triggered successfully",
		"searchedCount": searchedCount,
		"totalEligible": len(moviesToSearch),
	})
}

// getMoviesToSearch determines which movies should be searched based on the trigger criteria
func (s *Server) getMoviesToSearch(
	searchTrigger *models.WantedSearchTrigger, c *gin.Context,
) ([]models.WantedMovie, error) {
	if len(searchTrigger.MovieIDs) > 0 {
		return s.getSpecificMoviesToSearch(searchTrigger)
	}
	return s.getAllEligibleMoviesToSearch(searchTrigger, c)
}

// getSpecificMoviesToSearch gets specific movies to search by ID
func (s *Server) getSpecificMoviesToSearch(searchTrigger *models.WantedSearchTrigger) ([]models.WantedMovie, error) {
	var moviesToSearch []models.WantedMovie

	for _, movieID := range searchTrigger.MovieIDs {
		wantedMovie, err := s.services.WantedMoviesService.GetByMovieID(movieID)
		if err != nil {
			s.logger.Warn("Movie not in wanted list", "movieId", movieID)
			continue
		}
		if searchTrigger.ForceSearch || wantedMovie.IsEligibleForSearch() {
			moviesToSearch = append(moviesToSearch, *wantedMovie)
		}
	}

	return moviesToSearch, nil
}

// getAllEligibleMoviesToSearch gets all eligible movies with optional filtering
func (s *Server) getAllEligibleMoviesToSearch(
	searchTrigger *models.WantedSearchTrigger, c *gin.Context,
) ([]models.WantedMovie, error) {
	eligible, err := s.services.WantedMoviesService.GetEligibleForSearch(100) // Limit to 100 for performance
	if err != nil {
		s.logger.Error("Failed to get eligible wanted movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get eligible movies for search"})
		return nil, err
	}

	var moviesToSearch []models.WantedMovie
	for _, wantedMovie := range eligible {
		if s.shouldIncludeMovie(wantedMovie, searchTrigger) {
			moviesToSearch = append(moviesToSearch, wantedMovie)
		}
	}

	return moviesToSearch, nil
}

// shouldIncludeMovie determines if a movie should be included based on filters
func (s *Server) shouldIncludeMovie(wantedMovie models.WantedMovie, searchTrigger *models.WantedSearchTrigger) bool {
	if searchTrigger.FilterMissing && wantedMovie.Status != models.WantedStatusMissing {
		return false
	}
	if searchTrigger.FilterCutoff && wantedMovie.Status != models.WantedStatusCutoffUnmet {
		return false
	}
	return true
}

// executeMovieSearches executes searches for the selected movies
func (s *Server) executeMovieSearches(moviesToSearch []models.WantedMovie, forceSearch bool) int {
	searchedCount := 0
	for _, wantedMovie := range moviesToSearch {
		if _, err := s.services.SearchService.SearchMovieReleases(
			wantedMovie.MovieID, forceSearch); err != nil {
			s.logger.Error("Failed to trigger search for wanted movie", "movieId", wantedMovie.MovieID, "error", err)
			continue
		}
		searchedCount++

		// Update search attempt tracking
		if err := s.services.WantedMoviesService.UpdateSearchAttempt(
			wantedMovie.ID, false, "Search triggered", "", ""); err != nil {
			s.logger.Error("Failed to update search attempt tracking", "wantedMovieId", wantedMovie.ID, "error", err)
		}
	}
	return searchedCount
}

// handleWantedBulkOperation handles POST /api/v3/wanted/bulk
func (s *Server) handleWantedBulkOperation(c *gin.Context) {
	var operation models.WantedMoviesBulkOperation
	if err := c.ShouldBindJSON(&operation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %s", err.Error())})
		return
	}

	if len(operation.MovieIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No movie IDs provided"})
		return
	}

	err := s.services.WantedMoviesService.BulkOperation(&operation)
	if err != nil {
		s.logger.Error("Failed to perform bulk operation", "operation", operation.Operation, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to perform bulk operation: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Bulk operation completed successfully",
		"operation":      operation.Operation,
		"affectedMovies": len(operation.MovieIDs),
	})
}

// handleRefreshWantedMovies handles POST /api/v3/wanted/refresh
func (s *Server) handleRefreshWantedMovies(c *gin.Context) {
	err := s.services.WantedMoviesService.RefreshWantedMovies()
	if err != nil {
		s.logger.Error("Failed to refresh wanted movies", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh wanted movies"})
		return
	}

	stats, err := s.services.WantedMoviesService.GetWantedStats()
	if err != nil {
		s.logger.Error("Failed to get updated stats", "error", err)
		c.JSON(http.StatusOK, gin.H{"message": "Wanted movies refreshed successfully"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Wanted movies refreshed successfully",
		"stats":   stats,
	})
}

// handleUpdateWantedPriority handles PUT /api/v3/wanted/:id/priority
func (s *Server) handleUpdateWantedPriority(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var priorityUpdate struct {
		Priority models.WantedPriority `json:"priority" binding:"required"`
	}

	if err := c.ShouldBindJSON(&priorityUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %s", err.Error())})
		return
	}

	// Validate priority range
	if priorityUpdate.Priority < models.PriorityVeryLow || priorityUpdate.Priority > models.PriorityVeryHigh {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority value. Must be between 1 and 5"})
		return
	}

	// Get the wanted movie first to get the movie ID
	wantedMovie, err := s.services.WantedMoviesService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wanted movie not found"})
		return
	}

	err = s.services.WantedMoviesService.BulkOperation(&models.WantedMoviesBulkOperation{
		MovieIDs:  []int{wantedMovie.MovieID},
		Operation: models.BulkOpSetPriority,
		Options: models.WantedMoviesBulkOpOptions{
			Priority: &priorityUpdate.Priority,
		},
	})

	if err != nil {
		s.logger.Error("Failed to update wanted movie priority", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update priority"})
		return
	}

	// Return the updated wanted movie
	updated, err := s.services.WantedMoviesService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Priority updated successfully"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// handleRemoveWantedMovie handles DELETE /api/v3/wanted/:id
func (s *Server) handleRemoveWantedMovie(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the wanted movie first to get the movie ID
	wantedMovie, err := s.services.WantedMoviesService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wanted movie not found"})
		return
	}

	err = s.services.WantedMoviesService.BulkOperation(&models.WantedMoviesBulkOperation{
		MovieIDs:  []int{wantedMovie.MovieID},
		Operation: models.BulkOpRemove,
	})

	if err != nil {
		s.logger.Error("Failed to remove wanted movie", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove wanted movie"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wanted movie removed successfully"})
}

// parseWantedMovieFilter parses query parameters into a WantedMovieFilter
func (s *Server) parseWantedMovieFilter(c *gin.Context) *models.WantedMovieFilter {
	filter := &models.WantedMovieFilter{
		Page:     1,
		PageSize: DefaultPageSize,
		SortBy:   "priority",
		SortDir:  "DESC",
	}

	s.parsePaginationParams(c, filter)
	s.parseSortingParams(c, filter)
	s.parseStatusAndPriorityParams(c, filter)
	s.parseBooleanParams(c, filter)
	s.parseFilterParams(c, filter)
	s.parseTimeParams(c, filter)

	return filter
}

func (s *Server) parsePaginationParams(c *gin.Context, filter *models.WantedMovieFilter) {
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			if pageSize > 100 {
				pageSize = 100 // Cap at 100
			}
			filter.PageSize = pageSize
		}
	}
}

func (s *Server) parseSortingParams(c *gin.Context, filter *models.WantedMovieFilter) {
	if sortBy := c.Query("sortKey"); sortBy != "" {
		filter.SortBy = sortBy
	}
	if sortDir := c.Query("sortDirection"); sortDir != "" {
		filter.SortDir = strings.ToUpper(sortDir)
	}
}

func (s *Server) parseStatusAndPriorityParams(c *gin.Context, filter *models.WantedMovieFilter) {
	if statusStr := c.Query("status"); statusStr != "" {
		status := models.WantedStatus(statusStr)
		filter.Status = &status
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		if priority, err := strconv.Atoi(priorityStr); err == nil {
			p := models.WantedPriority(priority)
			filter.Priority = &p
		}
	}

	if minPriorityStr := c.Query("minPriority"); minPriorityStr != "" {
		if minPriority, err := strconv.Atoi(minPriorityStr); err == nil {
			p := models.WantedPriority(minPriority)
			filter.MinPriority = &p
		}
	}

	if maxPriorityStr := c.Query("maxPriority"); maxPriorityStr != "" {
		if maxPriority, err := strconv.Atoi(maxPriorityStr); err == nil {
			p := models.WantedPriority(maxPriority)
			filter.MaxPriority = &p
		}
	}
}

func (s *Server) parseBooleanParams(c *gin.Context, filter *models.WantedMovieFilter) {
	if isAvailableStr := c.Query("isAvailable"); isAvailableStr != "" {
		isAvailable := strings.ToLower(isAvailableStr) == trueBoolString
		filter.IsAvailable = &isAvailable
	}

	if searchRequiredStr := c.Query("searchRequired"); searchRequiredStr != "" {
		searchRequired := strings.ToLower(searchRequiredStr) == trueBoolString
		filter.SearchRequired = &searchRequired
	}

	if monitoredStr := c.Query("monitored"); monitoredStr != "" {
		monitored := strings.ToLower(monitoredStr) == trueBoolString
		filter.Monitored = &monitored
	}
}

func (s *Server) parseFilterParams(c *gin.Context, filter *models.WantedMovieFilter) {
	if qualityProfileIDStr := c.Query("qualityProfileId"); qualityProfileIDStr != "" {
		if qualityProfileID, err := strconv.Atoi(qualityProfileIDStr); err == nil {
			filter.QualityProfileID = &qualityProfileID
		}
	}

	if yearStr := c.Query("year"); yearStr != "" {
		if year, err := strconv.Atoi(yearStr); err == nil {
			filter.Year = &year
		}
	}

	if genre := c.Query("genre"); genre != "" {
		filter.Genre = &genre
	}
}

func (s *Server) parseTimeParams(c *gin.Context, filter *models.WantedMovieFilter) {
	if lastSearchBeforeStr := c.Query("lastSearchBefore"); lastSearchBeforeStr != "" {
		if lastSearchBefore, err := time.Parse(time.RFC3339, lastSearchBeforeStr); err == nil {
			filter.LastSearchBefore = &lastSearchBefore
		}
	}

	if lastSearchAfterStr := c.Query("lastSearchAfter"); lastSearchAfterStr != "" {
		if lastSearchAfter, err := time.Parse(time.RFC3339, lastSearchAfterStr); err == nil {
			filter.LastSearchAfter = &lastSearchAfter
		}
	}
}
