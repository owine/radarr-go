package services

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// FileOperationService tracks and manages file operations
type FileOperationService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewFileOperationService creates a new instance of FileOperationService
func NewFileOperationService(db *database.Database, logger *logger.Logger) *FileOperationService {
	return &FileOperationService{
		db:     db,
		logger: logger,
	}
}

// CreateOperation creates a new file operation record
func (s *FileOperationService) CreateOperation(
	operationType models.FileOperationType,
	sourcePath string,
	destinationPath string,
	movieID *int,
	size int64,
) (*models.FileOperationRecord, error) {
	operation := &models.FileOperationRecord{
		OperationType:   operationType,
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
		MovieID:         movieID,
		Size:            size,
		Status:          models.FileOperationStatusPending,
		Progress:        0.0,
	}

	if err := s.db.GORM.Create(operation).Error; err != nil {
		return nil, fmt.Errorf("failed to create file operation: %w", err)
	}

	s.logger.Info("Created file operation",
		"id", operation.ID,
		"type", operationType,
		"source", sourcePath,
		"destination", destinationPath)

	return operation, nil
}

// GetOperationByID retrieves a file operation by its ID
func (s *FileOperationService) GetOperationByID(id int) (*models.FileOperationRecord, error) {
	var operation models.FileOperationRecord

	err := s.db.GORM.Preload("Movie").Where("id = ?", id).First(&operation).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get file operation: %w", err)
	}

	return &operation, nil
}

// GetOperations retrieves file operations with optional filtering
func (s *FileOperationService) GetOperations(
	status models.FileOperationStatus,
	operationType models.FileOperationType,
	limit int,
	offset int,
) ([]models.FileOperationRecord, error) {
	query := s.db.GORM.Preload("Movie").Order("created_at desc")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if operationType != "" {
		query = query.Where("operation_type = ?", operationType)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	var operations []models.FileOperationRecord
	if err := query.Find(&operations).Error; err != nil {
		return nil, fmt.Errorf("failed to get file operations: %w", err)
	}

	return operations, nil
}

// GetActiveOperations retrieves currently active (processing) operations
func (s *FileOperationService) GetActiveOperations() ([]models.FileOperationRecord, error) {
	return s.GetOperations(models.FileOperationStatusProcessing, "", 0, 0)
}

// GetPendingOperations retrieves pending operations
func (s *FileOperationService) GetPendingOperations() ([]models.FileOperationRecord, error) {
	return s.GetOperations(models.FileOperationStatusPending, "", 0, 0)
}

// UpdateOperation updates a file operation record
func (s *FileOperationService) UpdateOperation(operation *models.FileOperationRecord) error {
	if err := s.db.GORM.Save(operation).Error; err != nil {
		return fmt.Errorf("failed to update file operation: %w", err)
	}

	s.logger.Debug("Updated file operation",
		"id", operation.ID,
		"status", operation.Status,
		"progress", operation.Progress)

	return nil
}

// StartOperation marks an operation as started
func (s *FileOperationService) StartOperation(id int) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	if operation.Status != models.FileOperationStatusPending {
		return fmt.Errorf("operation is not in pending status")
	}

	operation.MarkAsStarted()
	return s.UpdateOperation(operation)
}

// CompleteOperation marks an operation as completed
func (s *FileOperationService) CompleteOperation(id int) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	operation.MarkAsCompleted()
	return s.UpdateOperation(operation)
}

// FailOperation marks an operation as failed
func (s *FileOperationService) FailOperation(id int, errorMsg string) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	operation.MarkAsFailed(errorMsg)
	return s.UpdateOperation(operation)
}

// CancelOperation cancels a pending or processing operation
func (s *FileOperationService) CancelOperation(id int) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	if !operation.CanCancel() {
		return fmt.Errorf("operation cannot be canceled in current status: %s", operation.Status)
	}

	operation.MarkAsCanceled()
	return s.UpdateOperation(operation)
}

// UpdateProgress updates the progress of a file operation
func (s *FileOperationService) UpdateProgress(id int, bytesProcessed int64) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	operation.UpdateProgress(bytesProcessed)
	return s.UpdateOperation(operation)
}

// GetOperationSummary returns a summary of file operations by status
func (s *FileOperationService) GetOperationSummary() (*models.FileOperationSummary, error) {
	summary := &models.FileOperationSummary{}

	// Count total operations
	var totalCount int64
	if err := s.db.GORM.Model(&models.FileOperationRecord{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total operations: %w", err)
	}
	summary.Total = int(totalCount)

	// Count by status
	statusCounts := []struct {
		Status models.FileOperationStatus
		Count  *int
	}{
		{models.FileOperationStatusPending, &summary.Pending},
		{models.FileOperationStatusProcessing, &summary.Processing},
		{models.FileOperationStatusCompleted, &summary.Completed},
		{models.FileOperationStatusFailed, &summary.Failed},
		{models.FileOperationStatusCanceled, &summary.Canceled},
	}

	for _, sc := range statusCounts {
		var count int64
		if err := s.db.GORM.Model(&models.FileOperationRecord{}).
			Where("status = ?", sc.Status).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count operations by status: %w", err)
		}
		*sc.Count = int(count)
	}

	return summary, nil
}

// CleanupOldOperations removes old completed/failed operations
func (s *FileOperationService) CleanupOldOperations(olderThanDays int) error {
	cutoff := time.Now().AddDate(0, 0, -olderThanDays)

	result := s.db.GORM.Where(
		"(status = ? OR status = ? OR status = ?) AND completed_at < ?",
		models.FileOperationStatusCompleted,
		models.FileOperationStatusFailed,
		models.FileOperationStatusCanceled,
		cutoff,
	).Delete(&models.FileOperationRecord{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old operations: %w", result.Error)
	}

	s.logger.Info("Cleaned up old file operations",
		"deleted", result.RowsAffected,
		"cutoff", cutoff)

	return nil
}

// GetOperationsForMovie retrieves all file operations for a specific movie
func (s *FileOperationService) GetOperationsForMovie(movieID int) ([]models.FileOperationRecord, error) {
	var operations []models.FileOperationRecord

	err := s.db.GORM.Where("movie_id = ?", movieID).
		Order("created_at desc").Find(&operations).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get operations for movie: %w", err)
	}

	return operations, nil
}

// RetryFailedOperation retries a failed file operation
func (s *FileOperationService) RetryFailedOperation(id int) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	if operation.Status != models.FileOperationStatusFailed {
		return fmt.Errorf("operation is not in failed status")
	}

	// Reset operation status
	operation.Status = models.FileOperationStatusPending
	operation.Progress = 0.0
	operation.BytesProcessed = 0
	operation.ErrorMessage = ""
	operation.StartedAt = nil
	operation.CompletedAt = nil

	return s.UpdateOperation(operation)
}

// DeleteOperation removes a file operation record
func (s *FileOperationService) DeleteOperation(id int) error {
	operation, err := s.GetOperationByID(id)
	if err != nil {
		return err
	}

	if operation.IsProcessing() {
		return fmt.Errorf("cannot delete operation in processing status")
	}

	if err := s.db.GORM.Delete(operation).Error; err != nil {
		return fmt.Errorf("failed to delete operation: %w", err)
	}

	s.logger.Info("Deleted file operation", "id", id)
	return nil
}
