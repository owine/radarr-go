package services

import (
	"fmt"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// IndexerService provides operations for managing movie indexers and search providers.
type IndexerService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewIndexerService creates a new instance of IndexerService with the provided database and logger.
func NewIndexerService(db *database.Database, logger *logger.Logger) *IndexerService {
	return &IndexerService{
		db:     db,
		logger: logger,
	}
}

// GetIndexers retrieves all configured indexers from the system.
func (s *IndexerService) GetIndexers() ([]*models.Indexer, error) {
	var indexers []*models.Indexer

	if err := s.db.GORM.Find(&indexers).Error; err != nil {
		s.logger.Error("Failed to fetch indexers", "error", err)
		return nil, fmt.Errorf("failed to fetch indexers: %w", err)
	}

	return indexers, nil
}

// GetIndexerByID retrieves a specific indexer by its ID.
func (s *IndexerService) GetIndexerByID(id int) (*models.Indexer, error) {
	var indexer models.Indexer

	if err := s.db.GORM.Where("id = ?", id).First(&indexer).Error; err != nil {
		s.logger.Error("Failed to fetch indexer", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch indexer with id %d: %w", id, err)
	}

	return &indexer, nil
}

// CreateIndexer creates a new indexer configuration.
func (s *IndexerService) CreateIndexer(indexer *models.Indexer) error {
	if err := s.db.GORM.Create(indexer).Error; err != nil {
		s.logger.Error("Failed to create indexer", "name", indexer.Name, "error", err)
		return fmt.Errorf("failed to create indexer: %w", err)
	}

	s.logger.Info("Created indexer", "id", indexer.ID, "name", indexer.Name, "type", indexer.Type)
	return nil
}

// UpdateIndexer updates an existing indexer configuration.
func (s *IndexerService) UpdateIndexer(indexer *models.Indexer) error {
	if err := s.db.GORM.Save(indexer).Error; err != nil {
		s.logger.Error("Failed to update indexer", "id", indexer.ID, "error", err)
		return fmt.Errorf("failed to update indexer: %w", err)
	}

	s.logger.Info("Updated indexer", "id", indexer.ID, "name", indexer.Name)
	return nil
}

// DeleteIndexer removes an indexer configuration.
func (s *IndexerService) DeleteIndexer(id int) error {
	result := s.db.GORM.Delete(&models.Indexer{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete indexer", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete indexer: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("indexer with id %d not found", id)
	}

	s.logger.Info("Deleted indexer", "id", id)
	return nil
}

// GetEnabledIndexers retrieves all enabled indexers.
func (s *IndexerService) GetEnabledIndexers() ([]*models.Indexer, error) {
	var indexers []*models.Indexer

	if err := s.db.GORM.Where("status = ?", models.IndexerStatusEnabled).Find(&indexers).Error; err != nil {
		s.logger.Error("Failed to fetch enabled indexers", "error", err)
		return nil, fmt.Errorf("failed to fetch enabled indexers: %w", err)
	}

	return indexers, nil
}

// TestIndexer tests the connection to an indexer.
func (s *IndexerService) TestIndexer(indexer *models.Indexer) (*models.IndexerTestResult, error) {
	// Basic validation
	errors := []string{}

	if indexer.Name == "" {
		errors = append(errors, "Name is required")
	}

	if indexer.BaseURL == "" {
		errors = append(errors, "Base URL is required")
	}

	if indexer.Type == models.IndexerTypeTorznab || indexer.Type == models.IndexerTypeNewznab {
		if indexer.APIKey == "" {
			errors = append(errors, "API Key is required for Torznab/Newznab indexers")
		}
	}

	result := &models.IndexerTestResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}

	if !result.IsValid {
		return result, nil
	}

	// TODO: Implement actual indexer connection testing
	// This would involve making HTTP requests to the indexer's capabilities endpoint

	s.logger.Info("Tested indexer connection", "name", indexer.Name, "valid", result.IsValid)
	return result, nil
}

// SearchMovies searches for movies across all enabled indexers.
func (s *IndexerService) SearchMovies(query string) ([]*models.Movie, error) {
	enabledIndexers, err := s.GetEnabledIndexers()
	if err != nil {
		return nil, err
	}

	if len(enabledIndexers) == 0 {
		s.logger.Warn("No enabled indexers found for search")
		return []*models.Movie{}, nil
	}

	// TODO: Implement actual search across indexers
	// This would involve:
	// 1. Making search requests to each indexer's search endpoint
	// 2. Parsing responses (Torznab/Newznab XML or RSS)
	// 3. Converting results to movie objects
	// 4. Deduplicating results
	// 5. Sorting by relevance/quality

	s.logger.Info("Searched movies across indexers", "query", query, "indexers", len(enabledIndexers))
	return []*models.Movie{}, nil
}

// GetIndexerCapabilities retrieves the capabilities of an indexer.
func (s *IndexerService) GetIndexerCapabilities(indexer *models.Indexer) (*models.IndexerCapabilities, error) {
	// TODO: Implement capabilities detection
	// This would involve querying the indexer's capabilities endpoint

	capabilities := &models.IndexerCapabilities{
		SupportsSearch:            indexer.SupportsSearch,
		SupportsRSS:               indexer.SupportsRSS,
		SupportsRedirect:          indexer.SupportsRedirect,
		SupportedSearchParameters: []string{"q", "cat", "imdbid", "tmdbid"},
		Categories:                []int{2000, 2010, 2020, 2030, 2040, 2050, 2060}, // Common movie categories
	}

	return capabilities, nil
}
