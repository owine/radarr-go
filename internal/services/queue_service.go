package services

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// QueueService handles queue-related operations
type QueueService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewQueueService creates a new queue service
func NewQueueService(db *database.Database, logger *logger.Logger) *QueueService {
	return &QueueService{
		db:     db,
		logger: logger,
	}
}

// GetQueue retrieves all queue items with optional filtering
func (s *QueueService) GetQueue(
	movieIDs []int, protocol *models.DownloadProtocol, _ []int, _ []int,
	status []models.QueueStatus, includeUnknownMovieItems bool) ([]models.QueueItem, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var queue []models.QueueItem
	query := s.db.GORM.Preload("Movie")

	// Apply filters
	if len(movieIDs) > 0 {
		query = query.Where("movie_id IN ?", movieIDs)
	}

	if protocol != nil {
		query = query.Where("protocol = ?", *protocol)
	}

	if len(status) > 0 {
		statusStrings := make([]string, len(status))
		for i, st := range status {
			statusStrings[i] = string(st)
		}
		query = query.Where("status IN ?", statusStrings)
	}

	if !includeUnknownMovieItems {
		query = query.Where("movie_id IS NOT NULL AND movie_id > 0")
	}

	if err := query.Find(&queue).Error; err != nil {
		s.logger.Error("Failed to get queue items", "error", err)
		return nil, fmt.Errorf("failed to get queue items: %w", err)
	}

	s.logger.Debug("Retrieved queue items", "count", len(queue))
	return queue, nil
}

// GetQueueByID retrieves a specific queue item by ID
func (s *QueueService) GetQueueByID(id int) (*models.QueueItem, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var queueItem models.QueueItem
	if err := s.db.GORM.Preload("Movie").First(&queueItem, id).Error; err != nil {
		s.logger.Error("Failed to get queue item", "id", id, "error", err)
		return nil, fmt.Errorf("queue item not found: %w", err)
	}

	return &queueItem, nil
}

// AddQueueItem adds a new item to the queue
func (s *QueueService) AddQueueItem(queueItem *models.QueueItem) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	if err := s.db.GORM.Create(queueItem).Error; err != nil {
		s.logger.Error("Failed to add queue item", "error", err)
		return fmt.Errorf("failed to add queue item: %w", err)
	}

	s.logger.Info("Queue item added", "id", queueItem.ID, "title", queueItem.Title)
	return nil
}

// UpdateQueueItem updates an existing queue item
func (s *QueueService) UpdateQueueItem(queueItem *models.QueueItem) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	if err := s.db.GORM.Save(queueItem).Error; err != nil {
		s.logger.Error("Failed to update queue item", "id", queueItem.ID, "error", err)
		return fmt.Errorf("failed to update queue item: %w", err)
	}

	s.logger.Info("Queue item updated", "id", queueItem.ID, "status", queueItem.Status)
	return nil
}

// RemoveQueueItem removes a queue item by ID
func (s *QueueService) RemoveQueueItem(
	id int, removeFromClient bool, blocklist bool, _ bool, changeCategory bool) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	queueItem, err := s.GetQueueByID(id)
	if err != nil {
		return fmt.Errorf("queue item not found: %w", err)
	}

	// TODO: Implement download client removal logic
	if removeFromClient {
		s.logger.Info("Would remove from download client", "downloadId", queueItem.DownloadID,
			"client", queueItem.DownloadClient)
	}

	// TODO: Implement blocklist logic
	if blocklist {
		s.logger.Info("Would add to blocklist", "title", queueItem.Title)
	}

	// TODO: Implement change category logic
	if changeCategory {
		s.logger.Info("Would change category", "downloadId", queueItem.DownloadID)
	}

	if err := s.db.GORM.Delete(&models.QueueItem{}, id).Error; err != nil {
		s.logger.Error("Failed to remove queue item", "id", id, "error", err)
		return fmt.Errorf("failed to remove queue item: %w", err)
	}

	s.logger.Info("Queue item removed", "id", id, "title", queueItem.Title)
	return nil
}

// RemoveQueueItems removes multiple queue items
func (s *QueueService) RemoveQueueItems(ids []int, removeFromClient bool, blocklist bool,
	skipRedownload bool, changeCategory bool) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	for _, id := range ids {
		if err := s.RemoveQueueItem(id, removeFromClient, blocklist, skipRedownload, changeCategory); err != nil {
			s.logger.Error("Failed to remove queue item", "id", id, "error", err)
			// Continue with other items even if one fails
		}
	}

	s.logger.Info("Bulk queue item removal completed", "count", len(ids))
	return nil
}

// GetQueueByDownloadID retrieves a queue item by download ID
func (s *QueueService) GetQueueByDownloadID(downloadID string) (*models.QueueItem, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var queueItem models.QueueItem
	if err := s.db.GORM.Preload("Movie").Where("download_id = ?", downloadID).First(&queueItem).Error; err != nil {
		return nil, fmt.Errorf("queue item not found with download ID %s: %w", downloadID, err)
	}

	return &queueItem, nil
}

// UpdateQueueItemStatus updates the status of a queue item
func (s *QueueService) UpdateQueueItemStatus(id int, status models.QueueStatus, errorMessage string) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	updates := map[string]interface{}{
		"status": status,
	}

	if errorMessage != "" {
		updates["error_message"] = errorMessage
	}

	if err := s.db.GORM.Model(&models.QueueItem{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		s.logger.Error("Failed to update queue item status", "id", id, "error", err)
		return fmt.Errorf("failed to update queue item status: %w", err)
	}

	s.logger.Debug("Queue item status updated", "id", id, "status", status)
	return nil
}

// GetQueueItemsByStatus retrieves queue items by status
func (s *QueueService) GetQueueItemsByStatus(status models.QueueStatus) ([]models.QueueItem, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var queue []models.QueueItem
	if err := s.db.GORM.Preload("Movie").Where("status = ?", status).Find(&queue).Error; err != nil {
		s.logger.Error("Failed to get queue items by status", "status", status, "error", err)
		return nil, fmt.Errorf("failed to get queue items by status: %w", err)
	}

	return queue, nil
}

// UpdateProgress updates download progress for a queue item
func (s *QueueService) UpdateProgress(downloadID string, size int64, sizeLeft int64,
	timeLeft *time.Duration, estimatedCompletion *time.Time) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	updates := map[string]interface{}{
		"size":      size,
		"size_left": sizeLeft,
	}

	if timeLeft != nil {
		updates["time_left"] = *timeLeft
	}

	if estimatedCompletion != nil {
		updates["estimated_completion_time"] = *estimatedCompletion
	}

	err := s.db.GORM.Model(&models.QueueItem{}).Where("download_id = ?", downloadID).Updates(updates).Error
	if err != nil {
		s.logger.Error("Failed to update queue item progress", "downloadId", downloadID, "error", err)
		return fmt.Errorf("failed to update queue item progress: %w", err)
	}

	s.logger.Debug("Queue item progress updated", "downloadId", downloadID, "sizeLeft", sizeLeft)
	return nil
}

// GetQueueStats returns basic statistics about the queue
func (s *QueueService) GetQueueStats() (map[string]int, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := make(map[string]int)

	// Count total items
	var total int64
	if err := s.db.GORM.Model(&models.QueueItem{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total queue items: %w", err)
	}
	stats["total"] = int(total)

	// Count by status
	statuses := []models.QueueStatus{
		models.QueueStatusQueued,
		models.QueueStatusDownloading,
		models.QueueStatusCompleted,
		models.QueueStatusFailed,
		models.QueueStatusPaused,
	}

	for _, status := range statuses {
		var count int64
		if err := s.db.GORM.Model(&models.QueueItem{}).Where("status = ?", status).Count(&count).Error; err != nil {
			s.logger.Warn("Failed to count queue items by status", "status", status, "error", err)
			continue
		}
		stats[string(status)] = int(count)
	}

	return stats, nil
}
