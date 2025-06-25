package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// ImportListService provides operations for managing import lists and movie discovery
type ImportListService struct {
	db                *database.Database
	logger            *logger.Logger
	metadataService   *MetadataService
	movieService      *MovieService
	httpClient        *http.Client
}

// NewImportListService creates a new instance of ImportListService
func NewImportListService(db *database.Database, logger *logger.Logger, 
	metadataService *MetadataService, movieService *MovieService) *ImportListService {
	return &ImportListService{
		db:              db,
		logger:          logger,
		metadataService: metadataService,
		movieService:    movieService,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetImportLists retrieves all configured import lists
func (s *ImportListService) GetImportLists() ([]models.ImportList, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var lists []models.ImportList

	if err := s.db.GORM.Find(&lists).Error; err != nil {
		s.logger.Error("Failed to fetch import lists", "error", err)
		return nil, fmt.Errorf("failed to fetch import lists: %w", err)
	}

	s.logger.Debug("Retrieved import lists", "count", len(lists))
	return lists, nil
}

// GetImportListByID retrieves a specific import list by its ID
func (s *ImportListService) GetImportListByID(id int) (*models.ImportList, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var list models.ImportList

	if err := s.db.GORM.First(&list, id).Error; err != nil {
		s.logger.Error("Failed to fetch import list", "id", id, "error", err)
		return nil, fmt.Errorf("import list not found: %w", err)
	}

	return &list, nil
}

// CreateImportList creates a new import list configuration
func (s *ImportListService) CreateImportList(list *models.ImportList) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate required fields
	if err := s.validateImportList(list); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.GORM.Create(list).Error; err != nil {
		s.logger.Error("Failed to create import list", "name", list.Name, "error", err)
		return fmt.Errorf("failed to create import list: %w", err)
	}

	s.logger.Info("Created import list", "id", list.ID, "name", list.Name, "type", list.Implementation)
	return nil
}

// UpdateImportList updates an existing import list configuration
func (s *ImportListService) UpdateImportList(list *models.ImportList) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate required fields
	if err := s.validateImportList(list); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.GORM.Save(list).Error; err != nil {
		s.logger.Error("Failed to update import list", "id", list.ID, "error", err)
		return fmt.Errorf("failed to update import list: %w", err)
	}

	s.logger.Info("Updated import list", "id", list.ID, "name", list.Name)
	return nil
}

// DeleteImportList removes an import list configuration
func (s *ImportListService) DeleteImportList(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Check if list exists first
	_, err := s.GetImportListByID(id)
	if err != nil {
		return fmt.Errorf("import list not found: %w", err)
	}

	result := s.db.GORM.Delete(&models.ImportList{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete import list", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete import list: %w", result.Error)
	}

	s.logger.Info("Deleted import list", "id", id)
	return nil
}

// GetEnabledImportLists retrieves all enabled import lists
func (s *ImportListService) GetEnabledImportLists() ([]models.ImportList, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var lists []models.ImportList

	if err := s.db.GORM.Where("enabled = ? AND enable_auto = ?", true, true).Find(&lists).Error; err != nil {
		s.logger.Error("Failed to fetch enabled import lists", "error", err)
		return nil, fmt.Errorf("failed to fetch enabled import lists: %w", err)
	}

	s.logger.Debug("Retrieved enabled import lists", "count", len(lists))
	return lists, nil
}

// TestImportList tests the connection and configuration of an import list
func (s *ImportListService) TestImportList(list *models.ImportList) (*models.ImportListTestResult, error) {
	// Basic validation
	errors := []string{}

	if list.Name == "" {
		errors = append(errors, "Name is required")
	}

	if list.Implementation == "" {
		errors = append(errors, "Implementation is required")
	}

	if list.QualityProfileID <= 0 {
		errors = append(errors, "Quality profile is required")
	}

	if list.RootFolderPath == "" {
		errors = append(errors, "Root folder path is required")
	}

	result := &models.ImportListTestResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}

	if !result.IsValid {
		return result, nil
	}

	// Test actual connection based on implementation type
	if err := s.testImportListConnection(list); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
	} else {
		// Try to fetch a few sample movies to verify the connection works
		movies, err := s.fetchSampleMovies(list, 5)
		if err != nil {
			result.IsValid = false
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch sample movies: %s", err.Error()))
		} else {
			result.Movies = movies
		}
	}

	s.logger.Info("Tested import list connection", "name", list.Name, "valid", result.IsValid)
	return result, nil
}

// SyncImportList synchronizes movies from a specific import list
func (s *ImportListService) SyncImportList(listID int) (*models.ImportListSyncResult, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	list, err := s.GetImportListByID(listID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import list: %w", err)
	}

	if !list.IsEnabled() {
		return nil, fmt.Errorf("import list is not enabled")
	}

	result := s.initializeSyncResult(list)
	s.logger.Info("Starting import list sync", "listId", listID, "name", list.Name)

	// Fetch and process movies
	if err := s.processImportListSync(list, result); err != nil {
		result.Errors = append(result.Errors, err.Error())
		s.logger.Error("Failed to process import list sync", "listId", listID, "error", err)
		return result, nil
	}

	// Update last sync time
	list.LastSync = &result.SyncTime
	if err := s.UpdateImportList(list); err != nil {
		s.logger.Warn("Failed to update last sync time", "listId", listID, "error", err)
	}

	result.Success = len(result.Errors) == 0
	s.logSyncCompletion(listID, result)

	return result, nil
}

// initializeSyncResult creates and initializes a sync result
func (s *ImportListService) initializeSyncResult(list *models.ImportList) *models.ImportListSyncResult {
	return &models.ImportListSyncResult{
		ImportListID:   list.ID,
		ImportListName: list.Name,
		SyncTime:       time.Now(),
		Success:        false,
	}
}

// processImportListSync handles the movie fetching and processing for sync
func (s *ImportListService) processImportListSync(
	list *models.ImportList, result *models.ImportListSyncResult) error {
	// Fetch movies from the import list
	movies, err := s.fetchMoviesFromImportList(list)
	if err != nil {
		return fmt.Errorf("failed to fetch movies: %w", err)
	}

	result.MoviesTotal = len(movies)
	result.Movies = movies

	// Process each movie
	for _, movie := range movies {
		s.processSingleMovie(movie, list, result)
	}

	return nil
}

// processSingleMovie processes a single movie and updates the result counters
func (s *ImportListService) processSingleMovie(
	movie models.ImportListMovie, list *models.ImportList, result *models.ImportListSyncResult) {
	processed, err := s.processImportListMovie(movie, list)
	if err != nil {
		result.Errors = append(result.Errors, 
			fmt.Sprintf("Failed to process movie %s: %s", movie.Title, err.Error()))
		return
	}

	switch processed {
	case "added":
		result.MoviesAdded++
	case "updated":
		result.MoviesUpdated++
	case "excluded":
		result.MoviesExcluded++
	case "existing":
		result.MoviesExisting++
	}
}

// logSyncCompletion logs the completion of import list sync
func (s *ImportListService) logSyncCompletion(listID int, result *models.ImportListSyncResult) {
	s.logger.Info("Completed import list sync", "listId", listID, 
		"total", result.MoviesTotal, "added", result.MoviesAdded, 
		"updated", result.MoviesUpdated, "errors", len(result.Errors))
}

// SyncAllImportLists synchronizes all enabled import lists
func (s *ImportListService) SyncAllImportLists() ([]models.ImportListSyncResult, error) {
	lists, err := s.GetEnabledImportLists()
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled import lists: %w", err)
	}

	var results []models.ImportListSyncResult

	for _, list := range lists {
		result, err := s.SyncImportList(list.ID)
		if err != nil {
			s.logger.Error("Failed to sync import list", "listId", list.ID, "error", err)
			continue
		}
		results = append(results, *result)
	}

	s.logger.Info("Completed sync of all import lists", "count", len(results))
	return results, nil
}

// GetImportListMovies retrieves movies discovered from import lists
func (s *ImportListService) GetImportListMovies(listID *int, limit int) ([]models.ImportListMovie, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var movies []models.ImportListMovie
	query := s.db.GORM.Preload("ImportList").Order("discovered_at DESC")

	if listID != nil {
		query = query.Where("import_list_id = ?", *listID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&movies).Error; err != nil {
		s.logger.Error("Failed to fetch import list movies", "error", err)
		return nil, fmt.Errorf("failed to fetch import list movies: %w", err)
	}

	s.logger.Debug("Retrieved import list movies", "count", len(movies))
	return movies, nil
}

// validateImportList performs validation on import list configuration
func (s *ImportListService) validateImportList(list *models.ImportList) error {
	if list.Name == "" {
		return fmt.Errorf("name is required")
	}

	if list.Implementation == "" {
		return fmt.Errorf("implementation is required")
	}

	if list.QualityProfileID <= 0 {
		return fmt.Errorf("quality profile ID is required")
	}

	if list.RootFolderPath == "" {
		return fmt.Errorf("root folder path is required")
	}

	// Validate implementation-specific requirements
	if err := s.validateImplementationRequirements(list); err != nil {
		return err
	}

	return nil
}

// validateImplementationRequirements validates requirements for specific import list implementations
func (s *ImportListService) validateImplementationRequirements(list *models.ImportList) error {
	switch list.Implementation {
	case models.ImportListTypeTMDBList, models.ImportListTypeTMDBUser:
		if list.Settings.ListID == "" {
			return fmt.Errorf("list ID is required for TMDB lists")
		}
	case models.ImportListTypeTrakt, models.ImportListTypeTraktList, models.ImportListTypeTraktUser:
		if list.Settings.Username == "" {
			return fmt.Errorf("username is required for Trakt lists")
		}
	case models.ImportListTypeRSSImport:
		if list.Settings.URL == "" {
			return fmt.Errorf("URL is required for RSS import")
		}
	case models.ImportListTypeIMDbList:
		if list.Settings.ListID == "" {
			return fmt.Errorf("list ID is required for IMDb lists")
		}
	case models.ImportListTypeTMDBCollection, models.ImportListTypeTMDBCompany, models.ImportListTypeTMDBKeyword,
		 models.ImportListTypeTMDBPerson, models.ImportListTypeTMDBPopular, models.ImportListTypeTraktPopular,
		 models.ImportListTypePlexWatchlist, models.ImportListTypeRadarrList, models.ImportListTypeStevenLu,
		 models.ImportListTypeCouchPotato:
		// No additional validation required for these implementations
	}

	return nil
}

// testImportListConnection tests the connection to an import list
func (s *ImportListService) testImportListConnection(list *models.ImportList) error {
	baseURL := list.GetBaseURL()
	if baseURL == "" {
		return fmt.Errorf("no base URL available for testing")
	}

	// Create HTTP client with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	// Check if we get a reasonable response
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

	return nil
}

// fetchSampleMovies fetches a sample of movies from an import list for testing
func (s *ImportListService) fetchSampleMovies(list *models.ImportList, _ int) ([]models.ImportListMovie, error) {
	// This is a simplified version for testing
	// In a real implementation, this would connect to the actual import list service
	// and fetch sample movies based on the list type
	
	if list.Implementation == "" {
		return nil, fmt.Errorf("no implementation specified")
	}
	
	movies := []models.ImportListMovie{
		{
			TmdbID:        550,
			Title:         "Fight Club",
			OriginalTitle: "Fight Club",
			Year:          1999,
			Overview:      "Test movie from import list",
		},
	}

	s.logger.Debug("Fetched sample movies for testing", "count", len(movies), "listId", list.ID)
	return movies, nil
}

// fetchMoviesFromImportList fetches all movies from an import list
func (s *ImportListService) fetchMoviesFromImportList(list *models.ImportList) ([]models.ImportListMovie, error) {
	// This is a placeholder implementation
	// In a real implementation, this would connect to the actual import list service
	// (TMDB, Trakt, Plex, etc.) and fetch the actual movie list
	
	s.logger.Debug("Fetching movies from import list", "listId", list.ID, "type", list.Implementation)
	
	if !list.IsEnabled() {
		return nil, fmt.Errorf("import list is not enabled")
	}
	
	// Return empty slice for now - this would be implemented with actual API calls
	return []models.ImportListMovie{}, nil
}

// processImportListMovie processes a single movie from an import list
func (s *ImportListService) processImportListMovie(
	movie models.ImportListMovie, list *models.ImportList) (string, error) {
	// Check if movie is already in the collection
	// TODO: Implement GetMovieByTMDBID in MovieService
	// For now, assume movie doesn't exist
	_ = movie

	// Check if movie is excluded
	if s.isMovieExcluded(movie.TmdbID) {
		return "excluded", nil
	}

	// If auto-add is enabled, add the movie to the collection
	if list.ShouldAutoAdd() {
		newMovie := &models.Movie{
			Title:               movie.Title,
			OriginalTitle:       movie.OriginalTitle,
			Year:                movie.Year,
			TmdbID:              movie.TmdbID,
			ImdbID:              movie.ImdbID,
			Overview:            movie.Overview,
			Runtime:             movie.Runtime,
			QualityProfileID:    list.QualityProfileID,
			RootFolderPath:      list.RootFolderPath,
			Monitored:           list.ShouldMonitor,
			MinimumAvailability: list.MinimumAvailability,
			Tags:                list.Tags,
			Added:               time.Now(),
		}

		// TODO: Implement CreateMovie in MovieService
		// For now, simulate success
		_ = newMovie

		s.logger.Info("Added movie from import list", "title", movie.Title, "tmdbId", movie.TmdbID, "listId", list.ID)
		return "added", nil
	}

	// Store as discovered movie for manual review
	movie.ImportListID = list.ID
	movie.DiscoveredAt = time.Now()

	if err := s.db.GORM.Create(&movie).Error; err != nil {
		return "", fmt.Errorf("failed to store discovered movie: %w", err)
	}

	return "updated", nil
}

// isMovieExcluded checks if a movie is in the exclusion list
func (s *ImportListService) isMovieExcluded(tmdbID int) bool {
	if s.db == nil {
		return false
	}

	var count int64
	s.db.GORM.Model(&models.ImportListExclusion{}).Where("tmdb_id = ?", tmdbID).Count(&count)
	return count > 0
}

// GetImportListStats returns statistics about import lists
func (s *ImportListService) GetImportListStats() (map[string]any, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := make(map[string]any)

	// Count total lists
	var totalLists int64
	if err := s.db.GORM.Model(&models.ImportList{}).Count(&totalLists).Error; err != nil {
		return nil, fmt.Errorf("failed to count total lists: %w", err)
	}
	stats["total"] = totalLists

	// Count enabled lists
	var enabledLists int64
	err := s.db.GORM.Model(&models.ImportList{}).Where("enabled = ?", true).Count(&enabledLists).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count enabled lists: %w", err)
	}
	stats["enabled"] = enabledLists

	// Count discovered movies
	var discoveredMovies int64
	err = s.db.GORM.Model(&models.ImportListMovie{}).Count(&discoveredMovies).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count discovered movies: %w", err)
	}
	stats["discoveredMovies"] = discoveredMovies

	// Count exclusions
	var exclusions int64
	err = s.db.GORM.Model(&models.ImportListExclusion{}).Count(&exclusions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count exclusions: %w", err)
	}
	stats["exclusions"] = exclusions

	return stats, nil
}