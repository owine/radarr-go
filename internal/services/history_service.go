package services

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// HistoryService provides operations for managing history and activity tracking
type HistoryService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewHistoryService creates a new instance of HistoryService
func NewHistoryService(db *database.Database, logger *logger.Logger) *HistoryService {
	return &HistoryService{
		db:     db,
		logger: logger,
	}
}

// GetHistory retrieves history records with filtering and pagination
func (s *HistoryService) GetHistory(req models.HistoryRequest) (*models.HistoryResponse, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	s.setHistoryRequestDefaults(&req)
	query := s.buildHistoryQuery(req)

	// Count total records
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		s.logger.Error("Failed to count history records", "error", err)
		return nil, fmt.Errorf("failed to count history records: %w", err)
	}

	records, err := s.fetchHistoryRecords(query, req)
	if err != nil {
		return nil, err
	}

	response := s.buildHistoryResponse(req, totalRecords, records)
	s.logger.Debug("Retrieved history records", "count", len(records), "total", totalRecords)
	return response, nil
}

// setHistoryRequestDefaults sets default values for the history request
func (s *HistoryService) setHistoryRequestDefaults(req *models.HistoryRequest) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Limit to prevent excessive load
	}
	if req.SortKey == "" {
		req.SortKey = "date"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
}

// buildHistoryQuery builds the database query with filters
func (s *HistoryService) buildHistoryQuery(req models.HistoryRequest) *gorm.DB {
	query := s.db.GORM.Model(&models.History{}).Preload("Movie")

	if req.MovieID != nil {
		query = query.Where("movie_id = ?", *req.MovieID)
	}
	if req.EventType != nil {
		query = query.Where("event_type = ?", *req.EventType)
	}
	if req.Successful != nil {
		query = query.Where("successful = ?", *req.Successful)
	}
	if req.DownloadID != "" {
		query = query.Where("download_id = ?", req.DownloadID)
	}
	if req.Since != nil {
		query = query.Where("date >= ?", *req.Since)
	}
	if req.Until != nil {
		query = query.Where("date <= ?", *req.Until)
	}

	return query
}

// fetchHistoryRecords fetches the actual history records with pagination and sorting
func (s *HistoryService) fetchHistoryRecords(query *gorm.DB, req models.HistoryRequest) ([]models.History, error) {
	sortField := req.SortKey
	if req.SortDir == "desc" {
		sortField += " DESC"
	} else {
		sortField += " ASC"
	}

	offset := (req.Page - 1) * req.PageSize
	var records []models.History

	if err := query.Order(sortField).Offset(offset).Limit(req.PageSize).Find(&records).Error; err != nil {
		s.logger.Error("Failed to fetch history records", "error", err)
		return nil, fmt.Errorf("failed to fetch history records: %w", err)
	}

	return records, nil
}

// buildHistoryResponse builds the response object
func (s *HistoryService) buildHistoryResponse(req models.HistoryRequest, totalRecords int64,
	records []models.History) *models.HistoryResponse {
	return &models.HistoryResponse{
		Page:         req.Page,
		PageSize:     req.PageSize,
		SortKey:      req.SortKey,
		SortDir:      req.SortDir,
		TotalRecords: totalRecords,
		Records:      records,
	}
}

// GetHistoryByID retrieves a specific history record by its ID
func (s *HistoryService) GetHistoryByID(id int) (*models.History, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var history models.History

	if err := s.db.GORM.Preload("Movie").First(&history, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("history record not found")
		}
		s.logger.Error("Failed to fetch history record", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch history record: %w", err)
	}

	return &history, nil
}

// CreateHistoryRecord creates a new history record
func (s *HistoryService) CreateHistoryRecord(history *models.History) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Set default date if not provided
	if history.Date.IsZero() {
		history.Date = time.Now()
	}

	if err := s.db.GORM.Create(history).Error; err != nil {
		s.logger.Error("Failed to create history record", "eventType", history.EventType, "error", err)
		return fmt.Errorf("failed to create history record: %w", err)
	}

	s.logger.Info("Created history record", "id", history.ID, "eventType", history.EventType,
		"movieId", history.MovieID, "successful", history.Successful)
	return nil
}

// DeleteHistoryRecord removes a history record
func (s *HistoryService) DeleteHistoryRecord(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	result := s.db.GORM.Delete(&models.History{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete history record", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete history record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("history record not found")
	}

	s.logger.Info("Deleted history record", "id", id)
	return nil
}

// CleanupOldHistory removes history records older than the specified duration
func (s *HistoryService) CleanupOldHistory(olderThan time.Duration) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	cutoffDate := time.Now().Add(-olderThan)

	result := s.db.GORM.Where("date < ?", cutoffDate).Delete(&models.History{})
	if result.Error != nil {
		s.logger.Error("Failed to cleanup old history records", "error", result.Error)
		return 0, fmt.Errorf("failed to cleanup old history records: %w", result.Error)
	}

	s.logger.Info("Cleaned up old history records", "deleted", result.RowsAffected,
		"cutoffDate", cutoffDate)
	return result.RowsAffected, nil
}

// GetActivities retrieves current system activities with filtering and pagination
func (s *HistoryService) GetActivities(req models.ActivityRequest) (*models.ActivityResponse, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	query := s.db.GORM.Model(&models.Activity{}).Preload("Movie")

	// Apply filters
	if req.Type != nil {
		query = query.Where("type = ?", *req.Type)
	}

	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.MovieID != nil {
		query = query.Where("movie_id = ?", *req.MovieID)
	}

	if req.Since != nil {
		query = query.Where("start_time >= ?", *req.Since)
	}

	if req.Until != nil {
		query = query.Where("start_time <= ?", *req.Until)
	}

	// Count total records
	var totalRecords int64
	if err := query.Count(&totalRecords).Error; err != nil {
		s.logger.Error("Failed to count activity records", "error", err)
		return nil, fmt.Errorf("failed to count activity records: %w", err)
	}

	// Apply pagination and ordering (most recent first)
	offset := (req.Page - 1) * req.PageSize
	var records []models.Activity

	err := query.Order("start_time DESC").Offset(offset).Limit(req.PageSize).Find(&records).Error
	if err != nil {
		s.logger.Error("Failed to fetch activity records", "error", err)
		return nil, fmt.Errorf("failed to fetch activity records: %w", err)
	}

	response := &models.ActivityResponse{
		Page:         req.Page,
		PageSize:     req.PageSize,
		TotalRecords: totalRecords,
		Records:      records,
	}

	s.logger.Debug("Retrieved activity records", "count", len(records), "total", totalRecords)
	return response, nil
}

// GetActivityByID retrieves a specific activity by its ID
func (s *HistoryService) GetActivityByID(id int) (*models.Activity, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var activity models.Activity

	if err := s.db.GORM.Preload("Movie").First(&activity, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("activity not found")
		}
		s.logger.Error("Failed to fetch activity", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch activity: %w", err)
	}

	return &activity, nil
}

// CreateActivity creates a new activity record
func (s *HistoryService) CreateActivity(activity *models.Activity) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Set default values
	if activity.StartTime.IsZero() {
		activity.StartTime = time.Now()
	}
	if activity.Status == "" {
		activity.Status = models.ActivityStatusQueued
	}

	if err := s.db.GORM.Create(activity).Error; err != nil {
		s.logger.Error("Failed to create activity", "type", activity.Type, "error", err)
		return fmt.Errorf("failed to create activity: %w", err)
	}

	s.logger.Info("Created activity", "id", activity.ID, "type", activity.Type,
		"title", activity.Title, "status", activity.Status)
	return nil
}

// UpdateActivity updates an existing activity record
func (s *HistoryService) UpdateActivity(activity *models.Activity) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	if err := s.db.GORM.Save(activity).Error; err != nil {
		s.logger.Error("Failed to update activity", "id", activity.ID, "error", err)
		return fmt.Errorf("failed to update activity: %w", err)
	}

	s.logger.Debug("Updated activity", "id", activity.ID, "status", activity.Status,
		"progress", activity.Progress)
	return nil
}

// DeleteActivity removes an activity record
func (s *HistoryService) DeleteActivity(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	result := s.db.GORM.Delete(&models.Activity{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete activity", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete activity: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("activity not found")
	}

	s.logger.Info("Deleted activity", "id", id)
	return nil
}

// GetRunningActivities retrieves all currently running activities
func (s *HistoryService) GetRunningActivities() ([]models.Activity, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var activities []models.Activity

	err := s.db.GORM.Where("status = ?", models.ActivityStatusRunning).
		Preload("Movie").
		Order("start_time DESC").
		Find(&activities).Error

	if err != nil {
		s.logger.Error("Failed to fetch running activities", "error", err)
		return nil, fmt.Errorf("failed to fetch running activities: %w", err)
	}

	s.logger.Debug("Retrieved running activities", "count", len(activities))
	return activities, nil
}

// CleanupCompletedActivities removes completed activities older than specified duration
func (s *HistoryService) CleanupCompletedActivities(olderThan time.Duration) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not available")
	}

	cutoffDate := time.Now().Add(-olderThan)

	result := s.db.GORM.Where("end_time < ? AND status IN ?", cutoffDate,
		[]models.ActivityStatus{
			models.ActivityStatusCompleted,
			models.ActivityStatusFailed,
			models.ActivityStatusCancelled,
		}).Delete(&models.Activity{})

	if result.Error != nil {
		s.logger.Error("Failed to cleanup completed activities", "error", result.Error)
		return 0, fmt.Errorf("failed to cleanup completed activities: %w", result.Error)
	}

	s.logger.Info("Cleaned up completed activities", "deleted", result.RowsAffected,
		"cutoffDate", cutoffDate)
	return result.RowsAffected, nil
}

// RecordGrab records a movie grab event in history
func (s *HistoryService) RecordGrab(movieID int, sourceTitle, downloadID string,
	quality models.QualityDefinition, data models.HistoryEventData) error {
	history := &models.History{
		MovieID:     &movieID,
		EventType:   models.HistoryEventTypeGrabbed,
		Date:        time.Now(),
		Quality:     quality,
		SourceTitle: sourceTitle,
		DownloadID:  downloadID,
		Data:        data,
		Successful:  true,
	}

	return s.CreateHistoryRecord(history)
}

// RecordDownloadFailed records a download failure event
func (s *HistoryService) RecordDownloadFailed(movieID int, sourceTitle, downloadID,
	message string, data models.HistoryEventData) error {
	history := &models.History{
		MovieID:     &movieID,
		EventType:   models.HistoryEventTypeDownloadFailed,
		Date:        time.Now(),
		SourceTitle: sourceTitle,
		DownloadID:  downloadID,
		Data:        data,
		Message:     message,
		Successful:  false,
	}

	return s.CreateHistoryRecord(history)
}

// RecordMovieImported records a movie import event
func (s *HistoryService) RecordMovieImported(movieID int, sourceTitle string,
	quality models.QualityDefinition, data models.HistoryEventData) error {
	history := &models.History{
		MovieID:     &movieID,
		EventType:   models.HistoryEventTypeMovieImported,
		Date:        time.Now(),
		Quality:     quality,
		SourceTitle: sourceTitle,
		Data:        data,
		Successful:  true,
	}

	return s.CreateHistoryRecord(history)
}

// StartActivity creates and starts a new activity
func (s *HistoryService) StartActivity(activityType models.ActivityType, title string,
	movieID *int) (*models.Activity, error) {
	activity := &models.Activity{
		Type:      activityType,
		Title:     title,
		MovieID:   movieID,
		Status:    models.ActivityStatusRunning,
		StartTime: time.Now(),
		Progress:  0,
	}

	if err := s.CreateActivity(activity); err != nil {
		return nil, err
	}

	return activity, nil
}

// CompleteActivity marks an activity as completed and creates history record if needed
func (s *HistoryService) CompleteActivity(activityID int, successful bool,
	message string) error {
	activity, err := s.GetActivityByID(activityID)
	if err != nil {
		return fmt.Errorf("failed to get activity: %w", err)
	}

	activity.Complete(successful)
	if message != "" {
		activity.Message = message
	}

	return s.UpdateActivity(activity)
}

// GetHistoryStats returns statistics about history records
func (s *HistoryService) GetHistoryStats() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := make(map[string]interface{})

	// Total history records
	var totalHistory int64
	if err := s.db.GORM.Model(&models.History{}).Count(&totalHistory).Error; err != nil {
		return nil, fmt.Errorf("failed to count total history: %w", err)
	}
	stats["totalHistory"] = totalHistory

	// History by event type
	var eventTypeStats []struct {
		EventType string `json:"eventType"`
		Count     int64  `json:"count"`
	}

	err := s.db.GORM.Model(&models.History{}).
		Select("event_type, COUNT(*) as count").
		Group("event_type").
		Scan(&eventTypeStats).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get event type stats: %w", err)
	}
	stats["eventTypes"] = eventTypeStats

	// Recent activity count (last 24 hours)
	var recentActivity int64
	yesterday := time.Now().Add(-24 * time.Hour)
	err = s.db.GORM.Model(&models.History{}).
		Where("date >= ?", yesterday).
		Count(&recentActivity).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count recent activity: %w", err)
	}
	stats["recentActivity"] = recentActivity

	// Running activities
	var runningActivities int64
	err = s.db.GORM.Model(&models.Activity{}).
		Where("status = ?", models.ActivityStatusRunning).
		Count(&runningActivities).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count running activities: %w", err)
	}
	stats["runningActivities"] = runningActivities

	return stats, nil
}
