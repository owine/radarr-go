package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
	"gorm.io/hints"
)

// WantedMoviesService provides operations for managing wanted movies
type WantedMoviesService struct {
	db             *database.Database
	logger         *logger.Logger
	movieService   *MovieService
	qualityService *QualityService
}

// NewWantedMoviesService creates a new instance of WantedMoviesService
func NewWantedMoviesService(db *database.Database, logger *logger.Logger,
	movieService *MovieService, qualityService *QualityService) *WantedMoviesService {
	return &WantedMoviesService{
		db:             db,
		logger:         logger,
		movieService:   movieService,
		qualityService: qualityService,
	}
}

// GetMissingMovies retrieves all monitored movies that don't have files
func (s *WantedMoviesService) GetMissingMovies(filter *models.WantedMovieFilter) (*models.WantedMoviesResponse, error) {
	if filter == nil {
		filter = &models.WantedMovieFilter{
			Page:     1,
			PageSize: 20,
		}
	}

	// Set default values
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	query := s.db.GORM.Model(&models.WantedMovie{}).
		Clauses(hints.UseIndex("idx_wanted_movies_status")).
		Preload("Movie").
		Preload("Movie.MovieFile").
		Preload("TargetQuality").
		Where("status = ?", models.WantedStatusMissing)

	query = s.applyFilters(query, filter)

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.logger.Error("Failed to count missing movies", "error", err)
		return nil, fmt.Errorf("failed to count missing movies: %w", err)
	}

	// Apply sorting
	query = s.applySorting(query, filter)

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	var wantedMovies []models.WantedMovie
	if err := query.Find(&wantedMovies).Error; err != nil {
		s.logger.Error("Failed to get missing movies", "error", err)
		return nil, fmt.Errorf("failed to get missing movies: %w", err)
	}

	return &models.WantedMoviesResponse{
		Records:       wantedMovies,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		SortKey:       filter.SortBy,
		SortDirection: filter.SortDir,
		TotalRecords:  totalCount,
		FilteredCount: totalCount,
	}, nil
}

// GetCutoffUnmetMovies retrieves all movies with files below quality cutoff
func (s *WantedMoviesService) GetCutoffUnmetMovies(
	filter *models.WantedMovieFilter) (*models.WantedMoviesResponse, error) {
	if filter == nil {
		filter = &models.WantedMovieFilter{
			Page:     1,
			PageSize: 20,
		}
	}

	// Set default values
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	query := s.db.GORM.Model(&models.WantedMovie{}).
		Clauses(hints.UseIndex("idx_wanted_movies_status")).
		Preload("Movie").
		Preload("Movie.MovieFile").
		Preload("CurrentQuality").
		Preload("TargetQuality").
		Where("status IN ?", []models.WantedStatus{models.WantedStatusCutoffUnmet, models.WantedStatusUpgrade})

	query = s.applyFilters(query, filter)

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.logger.Error("Failed to count cutoff unmet movies", "error", err)
		return nil, fmt.Errorf("failed to count cutoff unmet movies: %w", err)
	}

	// Apply sorting
	query = s.applySorting(query, filter)

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	var wantedMovies []models.WantedMovie
	if err := query.Find(&wantedMovies).Error; err != nil {
		s.logger.Error("Failed to get cutoff unmet movies", "error", err)
		return nil, fmt.Errorf("failed to get cutoff unmet movies: %w", err)
	}

	return &models.WantedMoviesResponse{
		Records:       wantedMovies,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		SortKey:       filter.SortBy,
		SortDirection: filter.SortDir,
		TotalRecords:  totalCount,
		FilteredCount: totalCount,
	}, nil
}

// GetAllWanted retrieves all wanted movies with optional filtering
func (s *WantedMoviesService) GetAllWanted(filter *models.WantedMovieFilter) (*models.WantedMoviesResponse, error) {
	if filter == nil {
		filter = &models.WantedMovieFilter{
			Page:     1,
			PageSize: 20,
		}
	}

	// Set default values
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}

	query := s.db.GORM.Model(&models.WantedMovie{}).
		Preload("Movie").
		Preload("Movie.MovieFile").
		Preload("CurrentQuality").
		Preload("TargetQuality")

	query = s.applyFilters(query, filter)

	// Get total count
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		s.logger.Error("Failed to count wanted movies", "error", err)
		return nil, fmt.Errorf("failed to count wanted movies: %w", err)
	}

	// Apply sorting
	query = s.applySorting(query, filter)

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	var wantedMovies []models.WantedMovie
	if err := query.Find(&wantedMovies).Error; err != nil {
		s.logger.Error("Failed to get wanted movies", "error", err)
		return nil, fmt.Errorf("failed to get wanted movies: %w", err)
	}

	return &models.WantedMoviesResponse{
		Records:       wantedMovies,
		Page:          filter.Page,
		PageSize:      filter.PageSize,
		SortKey:       filter.SortBy,
		SortDirection: filter.SortDir,
		TotalRecords:  totalCount,
		FilteredCount: totalCount,
	}, nil
}

// applyFilters applies filtering criteria to the query
func (s *WantedMoviesService) applyFilters(query *gorm.DB, filter *models.WantedMovieFilter) *gorm.DB {
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if filter.Priority != nil {
		query = query.Where("priority = ?", *filter.Priority)
	}

	if filter.MinPriority != nil {
		query = query.Where("priority >= ?", *filter.MinPriority)
	}

	if filter.MaxPriority != nil {
		query = query.Where("priority <= ?", *filter.MaxPriority)
	}

	if filter.IsAvailable != nil {
		query = query.Where("is_available = ?", *filter.IsAvailable)
	}

	if filter.SearchRequired != nil {
		if *filter.SearchRequired {
			query = query.Where(
				"search_attempts < max_search_attempts AND (next_search_time IS NULL OR next_search_time <= ?)",
				time.Now())
		} else {
			query = query.Where(
				"search_attempts >= max_search_attempts OR (next_search_time IS NOT NULL AND next_search_time > ?)",
				time.Now())
		}
	}

	if filter.LastSearchBefore != nil {
		query = query.Where("last_search_time < ? OR last_search_time IS NULL", *filter.LastSearchBefore)
	}

	if filter.LastSearchAfter != nil {
		query = query.Where("last_search_time > ?", *filter.LastSearchAfter)
	}

	if filter.QualityProfileID != nil {
		query = query.Joins("JOIN movies ON movies.id = wanted_movies.movie_id").
			Where("movies.quality_profile_id = ?", *filter.QualityProfileID)
	}

	if filter.Monitored != nil {
		query = query.Joins("JOIN movies ON movies.id = wanted_movies.movie_id").
			Where("movies.monitored = ?", *filter.Monitored)
	}

	if filter.Year != nil {
		query = query.Joins("JOIN movies ON movies.id = wanted_movies.movie_id").
			Where("movies.year = ?", *filter.Year)
	}

	if filter.Genre != nil {
		query = query.Joins("JOIN movies ON movies.id = wanted_movies.movie_id").
			Where("JSON_CONTAINS(movies.genres, ?)", fmt.Sprintf(`"%s"`, *filter.Genre))
	}

	return query
}

// applySorting applies sorting to the query
func (s *WantedMoviesService) applySorting(query *gorm.DB, filter *models.WantedMovieFilter) *gorm.DB {
	sortBy := "priority"
	sortDir := "DESC"

	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}
	if filter.SortDir != "" {
		sortDir = filter.SortDir
	}

	switch sortBy {
	case "title":
		query = query.Joins("JOIN movies ON movies.id = wanted_movies.movie_id").
			Order(fmt.Sprintf("movies.title %s", sortDir))
	case "year":
		query = query.Joins("JOIN movies ON movies.id = wanted_movies.movie_id").
			Order(fmt.Sprintf("movies.year %s", sortDir))
	case "added":
		query = query.Order(fmt.Sprintf("wanted_movies.created_at %s", sortDir))
	case "lastSearchTime":
		query = query.Order(fmt.Sprintf("wanted_movies.last_search_time %s", sortDir))
	case "searchAttempts":
		query = query.Order(fmt.Sprintf("wanted_movies.search_attempts %s", sortDir))
	case "priority":
		query = query.Order(fmt.Sprintf("wanted_movies.priority %s", sortDir))
	default:
		query = query.Order(fmt.Sprintf("wanted_movies.priority %s, wanted_movies.created_at DESC", sortDir))
	}

	return query
}

// RefreshWantedMovies analyzes all monitored movies and updates wanted status
func (s *WantedMoviesService) RefreshWantedMovies() error {
	s.logger.Info("Starting wanted movies refresh")

	// Get all monitored movies
	movies, err := s.movieService.GetMonitored()
	if err != nil {
		return fmt.Errorf("failed to get monitored movies: %w", err)
	}

	var created, updated, removed int

	for _, movie := range movies {
		if err := s.analyzeMovie(&movie, &created, &updated, &removed); err != nil {
			s.logger.Error("Failed to analyze movie for wanted status", "movieId", movie.ID, "title", movie.Title, "error", err)
			continue
		}
	}

	s.logger.Info("Wanted movies refresh completed",
		"created", created, "updated", updated, "removed", removed)

	return nil
}

// analyzeMovie analyzes a single movie and updates its wanted status
func (s *WantedMoviesService) analyzeMovie(movie *models.Movie, created, updated, removed *int) error {
	// Get quality profile for the movie
	profile, err := s.qualityService.GetQualityProfileByID(movie.QualityProfileID)
	if err != nil {
		return fmt.Errorf("failed to get quality profile: %w", err)
	}

	// Check current wanted status
	var existingWanted models.WantedMovie
	err = s.db.GORM.Where("movie_id = ?", movie.ID).First(&existingWanted).Error
	hasExisting := err == nil

	wantedStatus, currentQualityID, targetQualityID, reason := s.determineWantedStatus(movie, profile)

	if wantedStatus == "" {
		// Movie is not wanted anymore
		if hasExisting {
			if err := s.db.GORM.Delete(&existingWanted).Error; err != nil {
				return fmt.Errorf("failed to remove from wanted: %w", err)
			}
			(*removed)++
		}
		return nil
	}

	// Movie is wanted
	wantedMovie := models.WantedMovie{
		MovieID:           movie.ID,
		Status:            wantedStatus,
		Reason:            reason,
		CurrentQualityID:  currentQualityID,
		TargetQualityID:   targetQualityID,
		IsAvailable:       movie.IsAvailable,
		Priority:          s.calculatePriority(movie, wantedStatus),
		MaxSearchAttempts: 10,
	}

	if hasExisting {
		// Update existing wanted movie
		wantedMovie.ID = existingWanted.ID
		wantedMovie.SearchAttempts = existingWanted.SearchAttempts
		wantedMovie.LastSearchTime = existingWanted.LastSearchTime
		wantedMovie.NextSearchTime = existingWanted.NextSearchTime
		wantedMovie.SearchFailures = existingWanted.SearchFailures
		wantedMovie.CreatedAt = existingWanted.CreatedAt

		if err := s.db.GORM.Save(&wantedMovie).Error; err != nil {
			return fmt.Errorf("failed to update wanted movie: %w", err)
		}
		(*updated)++
	} else {
		// Create new wanted movie
		if err := s.db.GORM.Create(&wantedMovie).Error; err != nil {
			return fmt.Errorf("failed to create wanted movie: %w", err)
		}
		(*created)++
	}

	return nil
}

// determineWantedStatus determines if a movie is wanted and why
func (s *WantedMoviesService) determineWantedStatus(movie *models.Movie,
	profile *models.QualityProfile) (models.WantedStatus, *int, int, string) {
	// Movie must be monitored and available to be wanted
	if !movie.Monitored || !movie.IsAvailable {
		return "", nil, 0, ""
	}

	// Get target quality (cutoff quality)
	targetQualityID := profile.Cutoff

	// If movie has no file, it's missing
	if !movie.HasFile || movie.MovieFile == nil {
		return models.WantedStatusMissing, nil, targetQualityID, "Movie has no file"
	}

	// If movie has file, check if it meets cutoff
	currentQualityID := movie.MovieFile.Quality.Quality.ID
	if currentQualityID < targetQualityID && profile.IsUpgradeAllowed() {
		return models.WantedStatusCutoffUnmet, &currentQualityID, targetQualityID,
			fmt.Sprintf("Current quality (%d) is below cutoff (%d)", currentQualityID, targetQualityID)
	}

	// Check if there's a better quality available within the profile
	if profile.IsUpgradeAllowed() {
		allowedQualities := profile.GetAllowedQualities()
		for _, quality := range allowedQualities {
			if quality.ID > currentQualityID {
				return models.WantedStatusUpgrade, &currentQualityID, quality.ID,
					fmt.Sprintf("Better quality available (%s)", quality.Title)
			}
		}
	}

	return "", nil, 0, ""
}

// calculatePriority determines the priority for a wanted movie
func (s *WantedMoviesService) calculatePriority(movie *models.Movie, status models.WantedStatus) models.WantedPriority {
	priority := models.PriorityNormal

	// Missing movies get higher priority than upgrades
	if status == models.WantedStatusMissing {
		priority = models.PriorityHigh
	}

	// Newer movies get higher priority
	if movie.Year >= time.Now().Year() {
		if priority < models.PriorityHigh {
			priority = models.PriorityHigh
		}
	} else if movie.Year >= time.Now().Year()-1 {
		if priority < models.PriorityNormal {
			priority = models.PriorityNormal
		}
	}

	// Popular movies get higher priority (based on popularity score)
	if movie.Popularity > 50.0 {
		if priority < models.PriorityHigh {
			priority = models.PriorityHigh
		}
	} else if movie.Popularity > 20.0 {
		if priority < models.PriorityNormal {
			priority = models.PriorityNormal
		}
	}

	return priority
}

// GetWantedStats returns statistics about wanted movies
func (s *WantedMoviesService) GetWantedStats() (*models.WantedMoviesStats, error) {
	var stats models.WantedMoviesStats

	// Total wanted count
	if err := s.db.GORM.Model(&models.WantedMovie{}).Count(&stats.TotalWanted).Error; err != nil {
		return nil, fmt.Errorf("failed to get total wanted count: %w", err)
	}

	// Missing count
	if err := s.db.GORM.Model(&models.WantedMovie{}).
		Where("status = ?", models.WantedStatusMissing).
		Count(&stats.MissingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get missing count: %w", err)
	}

	// Cutoff unmet count
	if err := s.db.GORM.Model(&models.WantedMovie{}).
		Where("status = ?", models.WantedStatusCutoffUnmet).
		Count(&stats.CutoffUnmetCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get cutoff unmet count: %w", err)
	}

	// Upgrade count
	if err := s.db.GORM.Model(&models.WantedMovie{}).
		Where("status = ?", models.WantedStatusUpgrade).
		Count(&stats.UpgradeCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get upgrade count: %w", err)
	}

	// Available count
	if err := s.db.GORM.Model(&models.WantedMovie{}).
		Where("is_available = ?", true).
		Count(&stats.AvailableCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get available count: %w", err)
	}

	// Searching count (eligible for search)
	if err := s.db.GORM.Model(&models.WantedMovie{}).
		Where("search_attempts < max_search_attempts AND (next_search_time IS NULL OR next_search_time <= ?)", time.Now()).
		Count(&stats.SearchingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get searching count: %w", err)
	}

	// High priority count
	if err := s.db.GORM.Model(&models.WantedMovie{}).
		Where("priority >= ?", models.PriorityHigh).
		Count(&stats.HighPriorityCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get high priority count: %w", err)
	}

	return &stats, nil
}

// GetByID retrieves a wanted movie by its ID
func (s *WantedMoviesService) GetByID(id int) (*models.WantedMovie, error) {
	var wantedMovie models.WantedMovie
	err := s.db.GORM.
		Preload("Movie").
		Preload("Movie.MovieFile").
		Preload("CurrentQuality").
		Preload("TargetQuality").
		First(&wantedMovie, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("wanted movie not found")
		}
		return nil, fmt.Errorf("failed to get wanted movie: %w", err)
	}

	return &wantedMovie, nil
}

// GetByMovieID retrieves a wanted movie by its movie ID
func (s *WantedMoviesService) GetByMovieID(movieID int) (*models.WantedMovie, error) {
	var wantedMovie models.WantedMovie
	err := s.db.GORM.
		Preload("Movie").
		Preload("Movie.MovieFile").
		Preload("CurrentQuality").
		Preload("TargetQuality").
		Where("movie_id = ?", movieID).
		First(&wantedMovie).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("wanted movie not found for movie ID %d", movieID)
		}
		return nil, fmt.Errorf("failed to get wanted movie: %w", err)
	}

	return &wantedMovie, nil
}

// UpdateSearchAttempt records a search attempt for a wanted movie
func (s *WantedMoviesService) UpdateSearchAttempt(id int, success bool, reason, indexer, errorCode string) error {
	return s.db.GORM.Transaction(func(tx *gorm.DB) error {
		var wantedMovie models.WantedMovie
		if err := tx.First(&wantedMovie, id).Error; err != nil {
			return fmt.Errorf("failed to get wanted movie: %w", err)
		}

		wantedMovie.IncrementSearchAttempts()

		if !success {
			wantedMovie.SearchFailures.AddFailure(reason, indexer, errorCode)
		}

		if err := tx.Save(&wantedMovie).Error; err != nil {
			return fmt.Errorf("failed to update wanted movie: %w", err)
		}

		s.logger.Info("Updated search attempt", "wantedMovieId", id, "success", success,
			"attempts", wantedMovie.SearchAttempts)

		return nil
	})
}

// BulkOperation performs bulk operations on wanted movies
func (s *WantedMoviesService) BulkOperation(operation *models.WantedMoviesBulkOperation) error {
	if len(operation.MovieIDs) == 0 {
		return errors.New("no movie IDs provided for bulk operation")
	}

	switch operation.Operation {
	case models.BulkOpSetPriority:
		if operation.Options.Priority == nil {
			return errors.New("priority is required for setPriority operation")
		}
		return s.bulkSetPriority(operation.MovieIDs, *operation.Options.Priority)

	case models.BulkOpResetSearchAttempts:
		return s.bulkResetSearchAttempts(operation.MovieIDs)

	case models.BulkOpRemove:
		return s.bulkRemove(operation.MovieIDs)

	case models.BulkOpSearch:
		return s.bulkSearch(operation.MovieIDs)

	default:
		return fmt.Errorf("unknown bulk operation: %s", operation.Operation)
	}
}

// bulkSetPriority sets priority for multiple wanted movies
func (s *WantedMoviesService) bulkSetPriority(movieIDs []int, priority models.WantedPriority) error {
	result := s.db.GORM.Model(&models.WantedMovie{}).
		Where("movie_id IN ?", movieIDs).
		Update("priority", priority)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk set priority: %w", result.Error)
	}

	s.logger.Info("Bulk set priority completed", "affectedRows", result.RowsAffected, "priority", priority)
	return nil
}

// bulkResetSearchAttempts resets search attempts for multiple wanted movies
func (s *WantedMoviesService) bulkResetSearchAttempts(movieIDs []int) error {
	updates := map[string]interface{}{
		"search_attempts":  0,
		"next_search_time": nil,
		"search_failures":  models.SearchFailures{},
	}

	result := s.db.GORM.Model(&models.WantedMovie{}).
		Where("movie_id IN ?", movieIDs).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed to bulk reset search attempts: %w", result.Error)
	}

	s.logger.Info("Bulk reset search attempts completed", "affectedRows", result.RowsAffected)
	return nil
}

// bulkRemove removes wanted movies (typically when they now have adequate files)
func (s *WantedMoviesService) bulkRemove(movieIDs []int) error {
	result := s.db.GORM.Where("movie_id IN ?", movieIDs).Delete(&models.WantedMovie{})

	if result.Error != nil {
		return fmt.Errorf("failed to bulk remove wanted movies: %w", result.Error)
	}

	s.logger.Info("Bulk remove wanted movies completed", "affectedRows", result.RowsAffected)
	return nil
}

func (s *WantedMoviesService) bulkSearch(movieIDs []int) error {
	// Trigger search for specified movies
	var wantedMovies []models.WantedMovie
	result := s.db.GORM.Where("movie_id IN ?", movieIDs).Find(&wantedMovies)
	if result.Error != nil {
		return fmt.Errorf("failed to fetch wanted movies for search: %w", result.Error)
	}

	// Update last search time for all movies
	now := time.Now()
	updateResult := s.db.GORM.Model(&models.WantedMovie{}).
		Where("movie_id IN ?", movieIDs).
		Updates(map[string]interface{}{
			"last_search_time": &now,
			"search_attempts":  gorm.Expr("search_attempts + 1"),
		})

	if updateResult.Error != nil {
		return fmt.Errorf("failed to update search metadata: %w", updateResult.Error)
	}

	s.logger.Info("Bulk search initiated for wanted movies", "count", len(wantedMovies))
	return nil
}

// GetEligibleForSearch returns wanted movies that are eligible for search
func (s *WantedMoviesService) GetEligibleForSearch(limit int) ([]models.WantedMovie, error) {
	if limit <= 0 {
		limit = 50 // Default limit
	}

	var wantedMovies []models.WantedMovie
	err := s.db.GORM.
		Preload("Movie").
		Preload("Movie.MovieFile").
		Preload("TargetQuality").
		Where("is_available = ? AND search_attempts < max_search_attempts", true).
		Where("next_search_time IS NULL OR next_search_time <= ?", time.Now()).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&wantedMovies).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get movies eligible for search: %w", err)
	}

	return wantedMovies, nil
}

// MarkSearchCompleted marks a wanted movie search as completed (successful)
func (s *WantedMoviesService) MarkSearchCompleted(movieID int) error {
	// Remove from wanted list if movie now has adequate file
	return s.db.GORM.Where("movie_id = ?", movieID).Delete(&models.WantedMovie{}).Error
}
