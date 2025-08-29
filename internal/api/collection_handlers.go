// Package api provides HTTP handlers for collection, parse, and rename functionality.
package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/radarr/radarr-go/internal/models"
)

// Collection handlers

// handleGetCollections retrieves all collections
func (s *Server) handleGetCollections(c *gin.Context) {
	var monitored *bool
	if monitoredStr := c.Query("monitored"); monitoredStr != "" {
		if m := monitoredStr == "true"; true {
			monitored = &m
		}
	}

	collections, err := s.services.CollectionService.GetAll(c.Request.Context(), monitored)
	if err != nil {
		s.logger.Error("Failed to fetch collections", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
		return
	}

	c.JSON(http.StatusOK, collections)
}

// handleGetCollection retrieves a specific collection
func (s *Server) handleGetCollection(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection, err := s.services.CollectionService.GetByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}
		s.logger.Error("Failed to fetch collection", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collection"})
		return
	}

	c.JSON(http.StatusOK, collection)
}

// handleCreateCollection creates a new collection
func (s *Server) handleCreateCollection(c *gin.Context) {
	var collection models.MovieCollection
	if err := c.ShouldBindJSON(&collection); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	createdCollection, err := s.services.CollectionService.Create(c.Request.Context(), &collection)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		s.logger.Error("Failed to create collection", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
		return
	}

	c.JSON(http.StatusCreated, createdCollection)
}

// handleUpdateCollection updates an existing collection
func (s *Server) handleUpdateCollection(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var updates models.MovieCollection
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	updatedCollection, err := s.services.CollectionService.Update(c.Request.Context(), id, &updates)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}
		s.logger.Error("Failed to update collection", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collection"})
		return
	}

	c.JSON(http.StatusOK, updatedCollection)
}

// handleDeleteCollection deletes a collection
func (s *Server) handleDeleteCollection(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	deleteMovies := c.Query("deleteMovies") == "true"

	if err := s.services.CollectionService.Delete(c.Request.Context(), id, deleteMovies); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}
		s.logger.Error("Failed to delete collection", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection deleted successfully"})
}

// handleSearchCollectionMovies searches for missing movies in a collection
func (s *Server) handleSearchCollectionMovies(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movieIDs, err := s.services.CollectionService.SearchMissing(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}
		s.logger.Error("Failed to search collection movies", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search collection movies"})
		return
	}

	response := gin.H{
		"collectionId": id,
		"movieIds":     movieIDs,
		"count":        len(movieIDs),
	}

	c.JSON(http.StatusOK, response)
}

// handleSyncCollectionFromTMDB syncs collection metadata from TMDB
func (s *Server) handleSyncCollectionFromTMDB(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.services.CollectionService.SyncFromTMDB(c.Request.Context(), id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}
		s.logger.Error("Failed to sync collection from TMDB", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync collection from TMDB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collection synced successfully"})
}

// handleGetCollectionStatistics retrieves statistics for a collection
func (s *Server) handleGetCollectionStatistics(c *gin.Context) {
	id, err := s.parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.services.CollectionService.GetCollectionStatistics(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
			return
		}
		s.logger.Error("Failed to get collection statistics", "id", id, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collection statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Parse handlers

// handleParseReleaseTitle parses a single release title
func (s *Server) handleParseReleaseTitle(c *gin.Context) {
	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title query parameter is required"})
		return
	}

	result, err := s.services.ParseService.ParseReleaseTitle(c.Request.Context(), title)
	if err != nil {
		s.logger.Error("Failed to parse release title", "title", title, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse release title"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleParseMultipleTitles parses multiple release titles
func (s *Server) handleParseMultipleTitles(c *gin.Context) {
	var request models.ParseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if len(request.Titles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one title is required"})
		return
	}

	results, err := s.services.ParseService.ParseMultipleTitles(c.Request.Context(), request.Titles)
	if err != nil {
		s.logger.Error("Failed to parse multiple titles", "count", len(request.Titles), "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse titles"})
		return
	}

	response := models.ParseResponse{
		Results: make([]models.ParseResult, len(results)),
	}
	for i, result := range results {
		response.Results[i] = *result
	}
	c.JSON(http.StatusOK, response)
}

// handleClearParseCache clears the parse cache
func (s *Server) handleClearParseCache(c *gin.Context) {
	if err := s.services.ParseService.ClearCache(c.Request.Context()); err != nil {
		s.logger.Error("Failed to clear parse cache", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear parse cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parse cache cleared successfully"})
}

// Rename handlers

// parseMovieIDsFromQuery parses movie IDs from the movieIds query parameter
func (s *Server) parseMovieIDsFromQuery(c *gin.Context) ([]int, error) {
	movieIDsParam := c.Query("movieIds")
	if movieIDsParam == "" {
		return nil, fmt.Errorf("movieIds query parameter is required")
	}

	var movieIDs []int
	for _, idStr := range strings.Split(movieIDsParam, ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(idStr)); err == nil {
			movieIDs = append(movieIDs, id)
		}
	}

	if len(movieIDs) == 0 {
		return nil, fmt.Errorf("no valid movie IDs provided")
	}

	return movieIDs, nil
}

// handlePreviewRename previews file renames for movies
func (s *Server) handlePreviewRename(c *gin.Context) {
	movieIDs, err := s.parseMovieIDsFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	previews, err := s.services.RenameService.PreviewRename(c.Request.Context(), movieIDs)
	if err != nil {
		s.logger.Error("Failed to preview rename", "movieIds", movieIDs, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preview rename"})
		return
	}

	c.JSON(http.StatusOK, previews)
}

// handleRenameRequest processes a rename request using the provided operation function
func (s *Server) handleRenameRequest(
	c *gin.Context,
	operation func(context.Context, []int) error,
	operationType, logAction string,
) {
	var request models.RenameMovieRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	if len(request.MovieIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No movie IDs provided"})
		return
	}

	if err := operation(c.Request.Context(), request.MovieIDs); err != nil {
		s.logger.Error(logAction, "movieIds", request.MovieIDs, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": logAction})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": operationType + " renamed successfully"})
}

// handleRenameMovies performs actual file renaming for movies
func (s *Server) handleRenameMovies(c *gin.Context) {
	s.handleRenameRequest(c, s.services.RenameService.RenameMovies, "Movies", "Failed to rename movies")
}

// handlePreviewMovieFolderRename previews folder renames for movies
func (s *Server) handlePreviewMovieFolderRename(c *gin.Context) {
	movieIDs, err := s.parseMovieIDsFromQuery(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	previews, err := s.services.RenameService.PreviewMovieFolderRename(c.Request.Context(), movieIDs)
	if err != nil {
		s.logger.Error("Failed to preview folder rename", "movieIds", movieIDs, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to preview folder rename"})
		return
	}

	c.JSON(http.StatusOK, previews)
}

// handleRenameMovieFolders performs actual folder renaming for movies
func (s *Server) handleRenameMovieFolders(c *gin.Context) {
	s.handleRenameRequest(
		c,
		s.services.RenameService.RenameMovieFolders,
		"Movie folders",
		"Failed to rename movie folders",
	)
}
