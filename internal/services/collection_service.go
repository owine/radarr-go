// Package services provides business logic for the Radarr application.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// CollectionService handles movie collection operations
type CollectionService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewCollectionService creates a new collection service
func NewCollectionService(db *database.Database, logger *logger.Logger) *CollectionService {
	return &CollectionService{
		db:     db,
		logger: logger,
	}
}

// GetAll returns all collections with optional filtering
func (s *CollectionService) GetAll(ctx context.Context, monitored *bool) ([]*models.MovieCollectionV2, error) {
	var collections []*models.MovieCollectionV2

	query := s.db.GORM.WithContext(ctx).
		Order("title ASC")

	if monitored != nil {
		query = query.Where("monitored = ?", *monitored)
	}

	if err := query.Find(&collections).Error; err != nil {
		s.logger.Error("Failed to fetch collections", "error", err)
		return nil, fmt.Errorf("failed to fetch collections: %w", err)
	}

	s.logger.Info("Fetched collections", "count", len(collections))
	return collections, nil
}

// GetByID returns a collection by its ID
func (s *CollectionService) GetByID(ctx context.Context, id int) (*models.MovieCollectionV2, error) {
	var collection models.MovieCollectionV2

	if err := s.db.GORM.WithContext(ctx).
		First(&collection, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("collection with ID %d not found", id)
		}
		s.logger.Error("Failed to fetch collection by ID", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch collection: %w", err)
	}

	return &collection, nil
}

// GetByTmdbID returns a collection by its TMDB ID
func (s *CollectionService) GetByTmdbID(ctx context.Context, tmdbID int) (*models.MovieCollectionV2, error) {
	var collection models.MovieCollectionV2

	if err := s.db.GORM.WithContext(ctx).
		Where("tmdb_id = ?", tmdbID).
		First(&collection).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("collection with TMDB ID %d not found", tmdbID)
		}
		s.logger.Error("Failed to fetch collection by TMDB ID", "tmdbId", tmdbID, "error", err)
		return nil, fmt.Errorf("failed to fetch collection: %w", err)
	}

	return &collection, nil
}

// Create creates a new collection
func (s *CollectionService) Create(
	ctx context.Context, collection *models.MovieCollectionV2,
) (*models.MovieCollectionV2, error) {
	// Check if collection already exists by TMDB ID
	existing, err := s.GetByTmdbID(ctx, collection.TmdbID)
	if err == nil {
		return nil, fmt.Errorf("collection with TMDB ID %d already exists with ID %d",
			collection.TmdbID, existing.ID)
	}

	// Validate the collection
	if err := collection.Validate(); err != nil {
		return nil, fmt.Errorf("collection validation failed: %w", err)
	}

	if err := s.db.GORM.WithContext(ctx).Create(collection).Error; err != nil {
		s.logger.Error("Failed to create collection", "title", collection.Title, "error", err)
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	s.logger.Info("Created collection", "id", collection.ID, "title", collection.Title, "tmdbId", collection.TmdbID)
	return collection, nil
}

// Update updates an existing collection
func (s *CollectionService) Update(
	ctx context.Context, id int, updates *models.MovieCollectionV2,
) (*models.MovieCollectionV2, error) {
	collection, err := s.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply changes
	collection.ApplyChanges(updates)

	if err := s.db.GORM.WithContext(ctx).Save(collection).Error; err != nil {
		s.logger.Error("Failed to update collection", "id", id, "error", err)
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	s.logger.Info("Updated collection", "id", id, "title", collection.Title)
	return collection, nil
}

// Delete deletes a collection and optionally its movies
func (s *CollectionService) Delete(ctx context.Context, id int, deleteMovies bool) error {
	collection, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if deleteMovies {
		// Delete all movies in the collection
		if err := s.db.GORM.WithContext(ctx).
			Where("collection_id = ?", collection.ID).
			Delete(&models.MovieV2{}).Error; err != nil {
			s.logger.Error("Failed to delete movies from collection", "id", id, "error", err)
			return fmt.Errorf("failed to delete movies from collection: %w", err)
		}
	} else {
		// Remove collection reference from movies
		if err := s.db.GORM.WithContext(ctx).
			Model(&models.MovieV2{}).
			Where("collection_id = ?", collection.ID).
			Select("collection_id").
			Update("collection_id", nil).Error; err != nil {
			s.logger.Error("Failed to remove collection reference from movies", "id", id, "error", err)
			return fmt.Errorf("failed to update movies: %w", err)
		}
	}

	if err := s.db.GORM.WithContext(ctx).Delete(collection).Error; err != nil {
		s.logger.Error("Failed to delete collection", "id", id, "error", err)
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	s.logger.Info("Deleted collection", "id", id, "title", collection.Title, "deleteMovies", deleteMovies)
	return nil
}

// AddMovie adds a movie to a collection
func (s *CollectionService) AddMovie(ctx context.Context, collectionID, movieID int) error {
	collection, err := s.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}

	// Update movie's collection reference
	if err := s.db.GORM.WithContext(ctx).
		Model(&models.MovieV2{}).
		Where("id = ?", movieID).
		Update("collection_id", collection.ID).Error; err != nil {
		s.logger.Error("Failed to add movie to collection", "collectionId", collectionID, "movieId", movieID, "error", err)
		return fmt.Errorf("failed to add movie to collection: %w", err)
	}

	s.logger.Info("Added movie to collection", "collectionId", collectionID, "movieId", movieID)
	return nil
}

// RemoveMovie removes a movie from a collection
func (s *CollectionService) RemoveMovie(ctx context.Context, collectionID, movieID int) error {
	if err := s.db.GORM.WithContext(ctx).
		Model(&models.MovieV2{}).
		Where("id = ?", movieID).
		Update("collection_id", nil).Error; err != nil {
		s.logger.Error("Failed to remove movie from collection",
			"collectionId", collectionID, "movieId", movieID, "error", err)
		return fmt.Errorf("failed to remove movie from collection: %w", err)
	}

	s.logger.Info("Removed movie from collection", "collectionId", collectionID, "movieId", movieID)
	return nil
}

// SearchMissing searches for missing movies in monitored collections
func (s *CollectionService) SearchMissing(ctx context.Context, collectionID int) ([]int, error) {
	collection, err := s.GetByID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	if !collection.Monitored {
		return nil, fmt.Errorf("collection %d is not monitored", collectionID)
	}

	// Find movies in collection without files
	var movieIDs []int
	if err := s.db.GORM.WithContext(ctx).
		Model(&models.MovieV2{}).
		Select("id").
		Where("collection_id = ? AND monitored = ? AND has_file = ?", collection.ID, true, false).
		Find(&movieIDs).Error; err != nil {
		s.logger.Error("Failed to find missing movies in collection", "collectionId", collectionID, "error", err)
		return nil, fmt.Errorf("failed to find missing movies: %w", err)
	}

	s.logger.Info("Found missing movies in collection", "collectionId", collectionID, "count", len(movieIDs))
	return movieIDs, nil
}

// SyncFromTMDB syncs collection metadata from TMDB
func (s *CollectionService) SyncFromTMDB(ctx context.Context, collectionID int) error {
	collection, err := s.GetByID(ctx, collectionID)
	if err != nil {
		return err
	}

	if err := s.enrichFromTMDB(ctx, collection); err != nil {
		return fmt.Errorf("failed to sync from TMDB: %w", err)
	}

	// Note: LastInfoSync removed in V2 simplified model

	if err := s.db.GORM.WithContext(ctx).Save(collection).Error; err != nil {
		s.logger.Error("Failed to save collection after TMDB sync", "id", collectionID, "error", err)
		return fmt.Errorf("failed to save collection: %w", err)
	}

	s.logger.Info("Synced collection from TMDB", "id", collectionID, "tmdbId", collection.TmdbID)
	return nil
}

// GetCollectionStatistics returns statistics for a collection
func (s *CollectionService) GetCollectionStatistics(
	ctx context.Context, collectionID int,
) (*models.CollectionStatisticsV2, error) {
	collection, err := s.GetByID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	var stats models.CollectionStatisticsV2
	if err := s.db.GORM.WithContext(ctx).
		Model(&models.MovieV2{}).
		Select(
			"COUNT(*) as movie_count",
			"COUNT(CASE WHEN monitored = true THEN 1 END) as monitored_movie_count",
			"COUNT(CASE WHEN has_file = true THEN 1 END) as has_file",
			"SUM(CASE WHEN has_file = true THEN file_size ELSE 0 END) as size_on_disk",
		).
		Where("collection_id = ?", collection.ID).
		Scan(&stats).Error; err != nil {
		s.logger.Error("Failed to calculate collection statistics", "collectionId", collectionID, "error", err)
		return nil, fmt.Errorf("failed to calculate statistics: %w", err)
	}

	stats.AvailableMovieCount = stats.HasFile
	if stats.MovieCount > 0 {
		stats.PercentOfMovies = float64(stats.HasFile) / float64(stats.MovieCount) * 100
	}

	return &stats, nil
}

// enrichFromTMDB fetches additional collection metadata from TMDB
func (s *CollectionService) enrichFromTMDB(ctx context.Context, collection *models.MovieCollectionV2) error {
	tmdbData, err := s.fetchTMDBData(ctx, collection.TmdbID)
	if err != nil {
		return err
	}

	s.updateCollectionFromTMDB(collection, tmdbData)
	return nil
}

// fetchTMDBData retrieves collection data from TMDB API
func (s *CollectionService) fetchTMDBData(ctx context.Context, tmdbID int) (*tmdbCollectionData, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/collection/%d?api_key=YOUR_API_KEY", tmdbID)

	resp, err := s.makeTMDBRequest(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			s.logger.Warnw("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDB API returned status %d", resp.StatusCode)
	}

	return s.parseTMDBResponse(resp.Body)
}

// makeTMDBRequest creates and executes HTTP request to TMDB
func (s *CollectionService) makeTMDBRequest(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create TMDB request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from TMDB: %w", err)
	}

	return resp, nil
}

// parseTMDBResponse parses TMDB API response into structured data
func (s *CollectionService) parseTMDBResponse(body io.Reader) (*tmdbCollectionData, error) {
	responseBody, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read TMDB response: %w", err)
	}

	var tmdbData tmdbCollectionData
	if err := json.Unmarshal(responseBody, &tmdbData); err != nil {
		return nil, fmt.Errorf("failed to parse TMDB response: %w", err)
	}

	return &tmdbData, nil
}

// updateCollectionFromTMDB updates collection fields with TMDB data
func (s *CollectionService) updateCollectionFromTMDB(
	collection *models.MovieCollectionV2, tmdbData *tmdbCollectionData,
) {
	s.updateCollectionBasicFields(collection, tmdbData)
	s.updateCollectionImages(collection, tmdbData)
}

// updateCollectionBasicFields updates basic collection fields from TMDB data
func (s *CollectionService) updateCollectionBasicFields(
	collection *models.MovieCollectionV2, tmdbData *tmdbCollectionData,
) {
	if collection.Title == "" {
		collection.Title = tmdbData.Name
	}
	if collection.Overview == "" {
		collection.Overview = tmdbData.Overview
	}
}

// updateCollectionImages updates collection images from TMDB data
func (s *CollectionService) updateCollectionImages(collection *models.MovieCollectionV2, tmdbData *tmdbCollectionData) {
	if collection.Images != nil && len(collection.GetImages()) > 0 {
		return // Already has images
	}

	var images []map[string]interface{}
	if tmdbData.PosterPath != "" {
		images = append(images, map[string]interface{}{
			"coverType": "poster",
			"url":       fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", tmdbData.PosterPath),
			"remoteUrl": fmt.Sprintf("https://image.tmdb.org/t/p/original%s", tmdbData.PosterPath),
		})
	}
	if tmdbData.BackdropPath != "" {
		images = append(images, map[string]interface{}{
			"coverType": "fanart",
			"url":       fmt.Sprintf("https://image.tmdb.org/t/p/w500%s", tmdbData.BackdropPath),
			"remoteUrl": fmt.Sprintf("https://image.tmdb.org/t/p/original%s", tmdbData.BackdropPath),
		})
	}
	collection.SetImages(images)
}

// tmdbCollectionData represents TMDB collection response structure
type tmdbCollectionData struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}
