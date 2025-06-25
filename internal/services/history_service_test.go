package services

import (
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestHistoryService_GetHistory(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	req := models.HistoryRequest{
		Page:     1,
		PageSize: 10,
	}
	_, err := service.GetHistory(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_GetHistoryByID(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	_, err := service.GetHistoryByID(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_CreateHistoryRecord(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	movieID := 1
	history := &models.History{
		MovieID:     &movieID,
		EventType:   models.HistoryEventTypeGrabbed,
		SourceTitle: "Test Movie 2023",
		Successful:  true,
	}

	// Test with nil database
	err := service.CreateHistoryRecord(history)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_DeleteHistoryRecord(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	err := service.DeleteHistoryRecord(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_CleanupOldHistory(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	_, err := service.CleanupOldHistory(30 * 24 * time.Hour) // 30 days
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_GetActivities(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	req := models.ActivityRequest{
		Page:     1,
		PageSize: 10,
	}
	_, err := service.GetActivities(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_GetActivityByID(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	_, err := service.GetActivityByID(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_CreateActivity(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	activity := &models.Activity{
		Type:  models.ActivityTypeMovieSearch,
		Title: "Searching for Test Movie",
	}

	// Test with nil database
	err := service.CreateActivity(activity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_UpdateActivity(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	activity := &models.Activity{
		ID:       1,
		Type:     models.ActivityTypeMovieSearch,
		Title:    "Searching for Test Movie",
		Progress: 50.0,
		Status:   models.ActivityStatusRunning,
	}

	// Test with nil database
	err := service.UpdateActivity(activity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_DeleteActivity(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	err := service.DeleteActivity(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_GetRunningActivities(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	_, err := service.GetRunningActivities()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_CleanupCompletedActivities(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	_, err := service.CleanupCompletedActivities(7 * 24 * time.Hour) // 7 days
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_RecordGrab(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	quality := models.QualityDefinition{
		ID:   1,
		Name: "HDTV-720p",
	}

	data := models.HistoryEventData{
		Indexer:     "Test Indexer",
		DownloadURL: "https://example.com/download",
		Size:        1024 * 1024 * 1024, // 1GB
		Protocol:    "torrent",
	}

	// Test with nil database
	err := service.RecordGrab(1, "Test Movie 2023", "download123", quality, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_RecordDownloadFailed(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	data := models.HistoryEventData{
		Indexer:  "Test Indexer",
		Reason:   "Connection timeout",
		Protocol: "torrent",
	}

	// Test with nil database
	err := service.RecordDownloadFailed(1, "Test Movie 2023", "download123",
		"Download failed: Connection timeout", data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_RecordMovieImported(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	quality := models.QualityDefinition{
		ID:   1,
		Name: "HDTV-720p",
	}

	data := models.HistoryEventData{
		ImportedPath: "/movies/Test Movie (2023)/Test Movie (2023) HDTV-720p.mkv",
		Size:         1024 * 1024 * 1024, // 1GB
	}

	// Test with nil database
	err := service.RecordMovieImported(1, "Test Movie 2023", quality, data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_StartActivity(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	movieID := 1

	// Test with nil database
	_, err := service.StartActivity(models.ActivityTypeMovieSearch, "Searching for Test Movie", &movieID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_CompleteActivity(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	err := service.CompleteActivity(1, true, "Search completed successfully")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestHistoryService_GetHistoryStats(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewHistoryService(nil, logger)

	// Test with nil database
	_, err := service.GetHistoryStats()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestActivity_Methods(t *testing.T) {
	activity := &models.Activity{
		Type:      models.ActivityTypeMovieSearch,
		Title:     "Test Activity",
		Status:    models.ActivityStatusRunning,
		StartTime: time.Now().Add(-10 * time.Minute),
		Progress:  25.0,
	}

	// Test IsRunning
	assert.True(t, activity.IsRunning())
	assert.False(t, activity.IsCompleted())

	// Test UpdateProgress
	activity.UpdateProgress(50, 100)
	assert.Equal(t, 50.0, activity.Progress)
	assert.Equal(t, 50, activity.Data.ProcessedItems)
	assert.Equal(t, 100, activity.Data.TotalItems)

	// Test Complete
	activity.Complete(true)
	assert.Equal(t, models.ActivityStatusCompleted, activity.Status)
	assert.Equal(t, 100.0, activity.Progress)
	assert.NotNil(t, activity.EndTime)
	assert.True(t, activity.IsCompleted())
	assert.False(t, activity.IsRunning())

	// Test Fail
	activity2 := &models.Activity{
		Type:      models.ActivityTypeDownload,
		Title:     "Test Download",
		Status:    models.ActivityStatusRunning,
		StartTime: time.Now(),
	}

	activity2.Fail("Download failed: Network error")
	assert.Equal(t, models.ActivityStatusFailed, activity2.Status)
	assert.NotNil(t, activity2.EndTime)
	assert.Contains(t, activity2.Data.Errors, "Download failed: Network error")
	assert.Equal(t, "Download failed: Network error", activity2.Message)

	// Test Cancel
	activity3 := &models.Activity{
		Type:      models.ActivityTypeImport,
		Title:     "Test Import",
		Status:    models.ActivityStatusRunning,
		StartTime: time.Now(),
	}

	activity3.Cancel()
	assert.Equal(t, models.ActivityStatusCancelled, activity3.Status)
	assert.NotNil(t, activity3.EndTime)

	// Test GetDuration
	duration := activity.GetDuration()
	assert.True(t, duration > 0)
}

func TestHistoryEventType_Constants(t *testing.T) {
	// Test that all event type constants are properly defined
	assert.Equal(t, "grabbed", string(models.HistoryEventTypeGrabbed))
	assert.Equal(t, "downloadFolderImported", string(models.HistoryEventTypeDownloadFolderImported))
	assert.Equal(t, "downloadFailed", string(models.HistoryEventTypeDownloadFailed))
	assert.Equal(t, "movieFileDeleted", string(models.HistoryEventTypeMovieFileDeleted))
	assert.Equal(t, "movieFileRenamed", string(models.HistoryEventTypeMovieFileRenamed))
	assert.Equal(t, "movieAdded", string(models.HistoryEventTypeMovieAdded))
	assert.Equal(t, "movieDeleted", string(models.HistoryEventTypeMovieDeleted))
	assert.Equal(t, "movieSearched", string(models.HistoryEventTypeMovieSearched))
	assert.Equal(t, "movieRefreshed", string(models.HistoryEventTypeMovieRefreshed))
	assert.Equal(t, "qualityUpgraded", string(models.HistoryEventTypeQualityUpgraded))
	assert.Equal(t, "movieImported", string(models.HistoryEventTypeMovieImported))
	assert.Equal(t, "movieUnmonitored", string(models.HistoryEventTypeMovieUnmonitored))
	assert.Equal(t, "movieMonitored", string(models.HistoryEventTypeMovieMonitored))
	assert.Equal(t, "ignoredDownload", string(models.HistoryEventTypeIgnoredDownload))
}

func TestActivityType_Constants(t *testing.T) {
	// Test that all activity type constants are properly defined
	assert.Equal(t, "movieSearch", string(models.ActivityTypeMovieSearch))
	assert.Equal(t, "movieRefresh", string(models.ActivityTypeMovieRefresh))
	assert.Equal(t, "download", string(models.ActivityTypeDownload))
	assert.Equal(t, "import", string(models.ActivityTypeImport))
	assert.Equal(t, "rename", string(models.ActivityTypeRename))
	assert.Equal(t, "metadataRefresh", string(models.ActivityTypeMetadataRefresh))
	assert.Equal(t, "healthCheck", string(models.ActivityTypeHealthCheck))
	assert.Equal(t, "backup", string(models.ActivityTypeBackup))
	assert.Equal(t, "importListSync", string(models.ActivityTypeImportListSync))
	assert.Equal(t, "indexerTest", string(models.ActivityTypeIndexerTest))
	assert.Equal(t, "queueProcess", string(models.ActivityTypeQueueProcess))
	assert.Equal(t, "systemUpdate", string(models.ActivityTypeSystemUpdate))
}

func TestActivityStatus_Constants(t *testing.T) {
	// Test that all activity status constants are properly defined
	assert.Equal(t, "running", string(models.ActivityStatusRunning))
	assert.Equal(t, "completed", string(models.ActivityStatusCompleted))
	assert.Equal(t, "failed", string(models.ActivityStatusFailed))
	assert.Equal(t, "cancelled", string(models.ActivityStatusCancelled))
	assert.Equal(t, "queued", string(models.ActivityStatusQueued))
}

func TestHistoryRequest_Defaults(t *testing.T) {
	req := models.HistoryRequest{}

	// Test that empty request has sensible defaults when processed
	assert.Equal(t, 0, req.Page)
	assert.Equal(t, 0, req.PageSize)
	assert.Equal(t, "", req.SortKey)
	assert.Equal(t, "", req.SortDir)
	assert.Nil(t, req.MovieID)
	assert.Nil(t, req.EventType)
	assert.Nil(t, req.Successful)
	assert.Equal(t, "", req.DownloadID)
}

func TestActivityRequest_Defaults(t *testing.T) {
	req := models.ActivityRequest{}

	// Test that empty request has sensible defaults when processed
	assert.Equal(t, 0, req.Page)
	assert.Equal(t, 0, req.PageSize)
	assert.Nil(t, req.Type)
	assert.Nil(t, req.Status)
	assert.Nil(t, req.MovieID)
}
