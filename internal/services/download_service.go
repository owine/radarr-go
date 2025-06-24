package services

import (
	"fmt"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// DownloadService provides operations for managing download clients and download queue.
type DownloadService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewDownloadService creates a new instance of DownloadService with the provided database and logger.
func NewDownloadService(db *database.Database, logger *logger.Logger) *DownloadService {
	return &DownloadService{
		db:     db,
		logger: logger,
	}
}

// GetDownloadClients retrieves all configured download clients.
func (s *DownloadService) GetDownloadClients() ([]*models.DownloadClient, error) {
	var clients []*models.DownloadClient

	if err := s.db.GORM.Find(&clients).Error; err != nil {
		s.logger.Error("Failed to fetch download clients", "error", err)
		return nil, fmt.Errorf("failed to fetch download clients: %w", err)
	}

	return clients, nil
}

// GetDownloadClientByID retrieves a specific download client by its ID.
func (s *DownloadService) GetDownloadClientByID(id int) (*models.DownloadClient, error) {
	var client models.DownloadClient

	if err := s.db.GORM.Where("id = ?", id).First(&client).Error; err != nil {
		s.logger.Error("Failed to fetch download client", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch download client with id %d: %w", id, err)
	}

	return &client, nil
}

// CreateDownloadClient creates a new download client configuration.
func (s *DownloadService) CreateDownloadClient(client *models.DownloadClient) error {
	if err := s.db.GORM.Create(client).Error; err != nil {
		s.logger.Error("Failed to create download client", "name", client.Name, "error", err)
		return fmt.Errorf("failed to create download client: %w", err)
	}

	s.logger.Info("Created download client", "id", client.ID, "name", client.Name, "type", client.Type)
	return nil
}

// UpdateDownloadClient updates an existing download client configuration.
func (s *DownloadService) UpdateDownloadClient(client *models.DownloadClient) error {
	if err := s.db.GORM.Save(client).Error; err != nil {
		s.logger.Error("Failed to update download client", "id", client.ID, "error", err)
		return fmt.Errorf("failed to update download client: %w", err)
	}

	s.logger.Info("Updated download client", "id", client.ID, "name", client.Name)
	return nil
}

// DeleteDownloadClient removes a download client configuration.
func (s *DownloadService) DeleteDownloadClient(id int) error {
	result := s.db.GORM.Delete(&models.DownloadClient{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete download client", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete download client: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("download client with id %d not found", id)
	}

	s.logger.Info("Deleted download client", "id", id)
	return nil
}

// GetEnabledDownloadClients retrieves all enabled download clients.
func (s *DownloadService) GetEnabledDownloadClients() ([]*models.DownloadClient, error) {
	var clients []*models.DownloadClient

	if err := s.db.GORM.Where("enable = ?", true).Find(&clients).Error; err != nil {
		s.logger.Error("Failed to fetch enabled download clients", "error", err)
		return nil, fmt.Errorf("failed to fetch enabled download clients: %w", err)
	}

	return clients, nil
}

// TestDownloadClient tests the connection to a download client.
func (s *DownloadService) TestDownloadClient(client *models.DownloadClient) (*models.DownloadClientTestResult, error) {
	// Basic validation
	errors := []string{}

	if client.Name == "" {
		errors = append(errors, "Name is required")
	}

	if client.Host == "" {
		errors = append(errors, "Host is required")
	}

	if client.Port <= 0 || client.Port > 65535 {
		errors = append(errors, "Port must be between 1 and 65535")
	}

	result := &models.DownloadClientTestResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}

	if !result.IsValid {
		return result, nil
	}

	// TODO: Implement actual download client connection testing
	// This would involve making HTTP requests to the client's API endpoint

	s.logger.Info("Tested download client connection", "name", client.Name, "valid", result.IsValid)
	return result, nil
}

// GetQueue retrieves all items in the download queue.
func (s *DownloadService) GetQueue() ([]*models.QueueItem, error) {
	var items []*models.QueueItem

	if err := s.db.GORM.Preload("Movie").Preload("DownloadClient").Find(&items).Error; err != nil {
		s.logger.Error("Failed to fetch queue items", "error", err)
		return nil, fmt.Errorf("failed to fetch queue items: %w", err)
	}

	return items, nil
}

// GetQueueItemByID retrieves a specific queue item by its ID.
func (s *DownloadService) GetQueueItemByID(id int) (*models.QueueItem, error) {
	var item models.QueueItem

	if err := s.db.GORM.Preload("Movie").Preload("DownloadClient").Where("id = ?", id).First(&item).Error; err != nil {
		s.logger.Error("Failed to fetch queue item", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch queue item with id %d: %w", id, err)
	}

	return &item, nil
}

// AddToQueue adds a new item to the download queue.
func (s *DownloadService) AddToQueue(item *models.QueueItem) error {
	if err := s.db.GORM.Create(item).Error; err != nil {
		s.logger.Error("Failed to add item to queue", "title", item.Title, "error", err)
		return fmt.Errorf("failed to add item to queue: %w", err)
	}

	s.logger.Info("Added item to queue", "id", item.ID, "title", item.Title)
	return nil
}

// RemoveFromQueue removes an item from the download queue.
func (s *DownloadService) RemoveFromQueue(id int) error {
	result := s.db.GORM.Delete(&models.QueueItem{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to remove item from queue", "id", id, "error", result.Error)
		return fmt.Errorf("failed to remove item from queue: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("queue item with id %d not found", id)
	}

	s.logger.Info("Removed item from queue", "id", id)
	return nil
}

// GetDownloadHistory retrieves download history.
func (s *DownloadService) GetDownloadHistory(limit int) ([]*models.DownloadHistory, error) {
	var history []*models.DownloadHistory

	query := s.db.GORM.Preload("Movie").Preload("DownloadClient").Order("date DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&history).Error; err != nil {
		s.logger.Error("Failed to fetch download history", "error", err)
		return nil, fmt.Errorf("failed to fetch download history: %w", err)
	}

	return history, nil
}

// AddToHistory adds a download to the history.
func (s *DownloadService) AddToHistory(history *models.DownloadHistory) error {
	if err := s.db.GORM.Create(history).Error; err != nil {
		s.logger.Error("Failed to add to download history", "title", history.SourceTitle, "error", err)
		return fmt.Errorf("failed to add to download history: %w", err)
	}

	s.logger.Info("Added to download history", "id", history.ID, "title", history.SourceTitle,
		"successful", history.Successful)
	return nil
}

// GetDownloads retrieves all active downloads from the download queue (legacy method for compatibility).
func (s *DownloadService) GetDownloads() ([]*models.QueueItem, error) {
	return s.GetQueue()
}
