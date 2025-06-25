package services

import (
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestQueueService_GetQueue(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	// Test with nil database
	_, err := service.GetQueue(nil, nil, nil, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueService_GetQueueByID(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	// Test with nil database
	_, err := service.GetQueueByID(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueService_AddQueueItem(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	queueItem := &models.QueueItem{
		Title:      "Test Movie",
		DownloadID: "test-download-123",
		Status:     models.QueueStatusQueued,
		Protocol:   models.DownloadProtocolTorrent,
	}

	// Test with nil database
	err := service.AddQueueItem(queueItem)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueService_UpdateQueueItemStatus(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	// Test with nil database
	err := service.UpdateQueueItemStatus(1, models.QueueStatusDownloading, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueService_RemoveQueueItem(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	// Test with nil database
	err := service.RemoveQueueItem(1, true, false, false, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueService_UpdateProgress(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	timeLeft := time.Hour
	estimated := time.Now().Add(time.Hour)

	// Test with nil database
	err := service.UpdateProgress("test-download-123", 1000, 500, &timeLeft, &estimated)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueService_GetQueueStats(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewQueueService(nil, logger)

	// Test with nil database
	_, err := service.GetQueueStats()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestQueueStatus_Constants(t *testing.T) {
	// Test that all queue status constants are properly defined
	assert.Equal(t, "unknown", string(models.QueueStatusUnknown))
	assert.Equal(t, "queued", string(models.QueueStatusQueued))
	assert.Equal(t, "downloading", string(models.QueueStatusDownloading))
	assert.Equal(t, "completed", string(models.QueueStatusCompleted))
	assert.Equal(t, "failed", string(models.QueueStatusFailed))
	assert.Equal(t, "paused", string(models.QueueStatusPaused))
}

func TestDownloadProtocol_Constants(t *testing.T) {
	// Test that all download protocol constants are properly defined
	assert.Equal(t, "unknown", string(models.DownloadProtocolUnknown))
	assert.Equal(t, "usenet", string(models.DownloadProtocolUsenet))
	assert.Equal(t, "torrent", string(models.DownloadProtocolTorrent))
}

func TestTrackedDownloadStatus_Constants(t *testing.T) {
	// Test that all tracked download status constants are properly defined
	assert.Equal(t, "ok", string(models.TrackedDownloadStatusOk))
	assert.Equal(t, "warning", string(models.TrackedDownloadStatusWarning))
	assert.Equal(t, "error", string(models.TrackedDownloadStatusError))
}

func TestStatusMessage_Structure(t *testing.T) {
	// Test status message structure
	message := models.StatusMessage{
		Title:    "Test Message",
		Messages: []string{"Detail 1", "Detail 2"},
		Type:     models.StatusMessageTypeWarning,
	}

	assert.Equal(t, "Test Message", message.Title)
	assert.Len(t, message.Messages, 2)
	assert.Equal(t, "warning", string(message.Type))
}

func TestQualityModel_Structure(t *testing.T) {
	// Test quality model structure
	quality := models.QualityModel{
		Quality: models.QualityDefinition{
			ID:     1,
			Name:   "HD-1080p",
			Source: "bluray",
		},
		Revision: models.Revision{
			Version:  1,
			Real:     0,
			IsRepack: false,
		},
	}

	assert.Equal(t, 1, quality.Quality.ID)
	assert.Equal(t, "HD-1080p", quality.Quality.Name)
	assert.Equal(t, "bluray", quality.Quality.Source)
	assert.Equal(t, 1, quality.Revision.Version)
	assert.False(t, quality.Revision.IsRepack)
}