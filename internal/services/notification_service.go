package services

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/services/notifications"
	"gorm.io/gorm"
)

// NotificationService provides operations for managing notifications and alerts.
type NotificationService struct {
	db               *database.Database
	logger           *logger.Logger
	httpClient       *http.Client
	factory          notifications.ProviderFactory
	templateEngine   notifications.TemplateEngine
	defaultTemplates map[string]*notifications.NotificationTemplate
}

// NewNotificationService creates a new instance of NotificationService with the provided database and logger.
func NewNotificationService(db *database.Database, logger *logger.Logger) *NotificationService {
	factory := notifications.NewProviderFactory(logger)
	templateEngine := notifications.NewTemplateEngine(logger)

	return &NotificationService{
		db:               db,
		logger:           logger,
		httpClient:       &http.Client{Timeout: 30 * time.Second},
		factory:          factory,
		templateEngine:   templateEngine,
		defaultTemplates: notifications.GetDefaultTemplates(),
	}
}

// GetNotifications retrieves all configured notifications from the system.
func (s *NotificationService) GetNotifications() ([]models.Notification, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var notifications []models.Notification

	if err := s.db.GORM.Find(&notifications).Error; err != nil {
		s.logger.Error("Failed to fetch notifications", "error", err)
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	s.logger.Debug("Retrieved notifications", "count", len(notifications))
	return notifications, nil
}

// GetNotificationByID retrieves a specific notification by its ID.
func (s *NotificationService) GetNotificationByID(id int) (*models.Notification, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var notification models.Notification

	if err := s.db.GORM.First(&notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("notification not found")
		}
		s.logger.Error("Failed to fetch notification", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch notification: %w", err)
	}

	return &notification, nil
}

// CreateNotification creates a new notification configuration.
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate required fields
	if err := s.validateNotification(notification); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.GORM.Create(notification).Error; err != nil {
		s.logger.Error("Failed to create notification", "name", notification.Name, "error", err)
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("Created notification", "id", notification.ID, "name", notification.Name,
		"type", notification.Implementation)
	return nil
}

// UpdateNotification updates an existing notification configuration.
func (s *NotificationService) UpdateNotification(notification *models.Notification) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Validate required fields
	if err := s.validateNotification(notification); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := s.db.GORM.Save(notification).Error; err != nil {
		s.logger.Error("Failed to update notification", "id", notification.ID, "error", err)
		return fmt.Errorf("failed to update notification: %w", err)
	}

	s.logger.Info("Updated notification", "id", notification.ID, "name", notification.Name)
	return nil
}

// DeleteNotification removes a notification configuration.
func (s *NotificationService) DeleteNotification(id int) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	// Check if notification exists first
	_, err := s.GetNotificationByID(id)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	result := s.db.GORM.Delete(&models.Notification{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete notification", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	s.logger.Info("Deleted notification", "id", id)
	return nil
}

// GetEnabledNotifications retrieves all enabled notifications.
func (s *NotificationService) GetEnabledNotifications() ([]models.Notification, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var notifications []models.Notification

	if err := s.db.GORM.Where("enabled = ?", true).Find(&notifications).Error; err != nil {
		s.logger.Error("Failed to fetch enabled notifications", "error", err)
		return nil, fmt.Errorf("failed to fetch enabled notifications: %w", err)
	}

	s.logger.Debug("Retrieved enabled notifications", "count", len(notifications))
	return notifications, nil
}

// TestNotification tests a notification configuration.
func (s *NotificationService) TestNotification(notification *models.Notification) (
	*models.NotificationTestResult, error) {
	// Basic validation
	errors := []string{}

	if notification.Name == "" {
		errors = append(errors, "Name is required")
	}

	if notification.Implementation == "" {
		errors = append(errors, "Implementation is required")
	}

	result := &models.NotificationTestResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}

	if !result.IsValid {
		return result, nil
	}

	// Create provider and test connection
	provider, err := s.factory.CreateProvider(notification.Implementation)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to create provider: %v", err))
		return result, nil
	}

	// Validate configuration
	if err := provider.ValidateConfig(notification.Settings); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Configuration validation failed: %v", err))
		return result, nil
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := provider.TestConnection(ctx, notification.Settings); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Connection test failed: %v", err))
	}

	s.logger.Info("Tested notification", "name", notification.Name, "valid", result.IsValid)
	return result, nil
}

// SendNotification sends a notification for a specific event
func (s *NotificationService) SendNotification(event *models.NotificationEvent) error {
	if s.db == nil {
		return fmt.Errorf("database not available")
	}

	notifications, err := s.GetEnabledNotifications()
	if err != nil {
		return fmt.Errorf("failed to get enabled notifications: %w", err)
	}

	eventType := event.EventType
	if eventType == "" {
		eventType = event.Type
	}

	// Convert to notification message format
	notificationMessage := s.convertEventToMessage(event)

	// Process notifications concurrently
	var wg sync.WaitGroup
	for _, notification := range notifications {
		if !s.shouldSendNotificationForEvent(&notification, eventType) {
			continue
		}

		wg.Add(1)
		go func(n models.Notification) {
			defer wg.Done()
			if err := s.sendNotificationWithRetry(&n, notificationMessage); err != nil {
				s.logger.Error("Failed to send notification", "id", n.ID,
					"name", n.Name, "error", err)
			}
		}(notification)
	}

	wg.Wait()
	return nil
}

// shouldSendNotificationForEvent checks if a notification should be sent for an event type
func (s *NotificationService) shouldSendNotificationForEvent(notification *models.Notification, eventType string) bool {
	if !notification.IsEnabled() {
		return false
	}

	switch eventType {
	case "grab":
		return notification.SupportsOnGrab && notification.OnGrab
	case "download":
		return notification.SupportsOnDownload && notification.OnDownload
	case "upgrade":
		return notification.SupportsOnUpgrade && notification.OnUpgrade
	case "rename":
		return notification.SupportsOnRename && notification.OnRename
	case "movieAdded":
		return notification.SupportsOnMovieAdded && notification.OnMovieAdded
	case "movieDelete":
		return notification.SupportsOnMovieDelete && notification.OnMovieDelete
	case "movieFileDelete":
		return notification.SupportsOnMovieFileDelete && notification.OnMovieFileDelete
	case "health":
		return notification.SupportsOnHealthIssue && notification.OnHealthIssue
	case "applicationUpdate":
		return notification.SupportsOnApplicationUpdate && notification.OnApplicationUpdate
	case "manualInteractionRequired":
		return notification.SupportsOnManualInteractionRequired && notification.OnManualInteractionRequired
	default:
		return false
	}
}

// validateNotification validates a notification configuration
func (s *NotificationService) validateNotification(notification *models.Notification) error {
	if notification.Name == "" {
		return fmt.Errorf("name is required")
	}
	if notification.Implementation == "" {
		return fmt.Errorf("implementation is required")
	}

	// Validate using the provider
	provider, err := s.factory.CreateProvider(notification.Implementation)
	if err != nil {
		return fmt.Errorf("invalid provider type: %w", err)
	}

	if err := provider.ValidateConfig(notification.Settings); err != nil {
		return fmt.Errorf("provider configuration validation failed: %w", err)
	}

	return nil
}

// convertEventToMessage converts a notification event to a notification message
func (s *NotificationService) convertEventToMessage(event *models.NotificationEvent) *notifications.NotificationMessage {
	return &notifications.NotificationMessage{
		Subject:        s.buildSubject(event),
		Body:           s.buildBody(event),
		EventType:      event.EventType,
		Movie:          event.Movie,
		MovieFile:      event.MovieFile,
		DeletedFiles:   event.DeletedFiles,
		OldMovieFile:   event.OldMovieFile,
		SourceTitle:    event.SourceTitle,
		Quality:        event.Quality,
		QualityUpgrade: event.QualityUpgrade,
		DownloadClient: event.DownloadClient,
		DownloadID:     event.DownloadID,
		HealthCheck:    event.HealthCheck,
		Data:           event.Data,
		IsTest:         false,
		ServerName:     "Radarr",
		Timestamp:      time.Now(),
	}
}

// sendNotificationWithRetry sends a notification with retry logic
func (s *NotificationService) sendNotificationWithRetry(
	notification *models.Notification,
	message *notifications.NotificationMessage) error {
	// Create provider
	provider, err := s.factory.CreateProvider(notification.Implementation)
	if err != nil {
		return fmt.Errorf("failed to create provider: %w", err)
	}

	// Get retry configuration
	retryConfig := provider.GetDefaultRetryConfig()
	if !provider.SupportsRetry() {
		retryConfig.MaxRetries = 0
	}

	// Try sending with retry logic
	var lastError error
	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		startTime := time.Now()

		// Apply template if available
		renderedMessage, err := s.applyTemplate(notification, message)
		if err != nil {
			s.logger.Warn("Failed to apply template, using original message", "error", err)
			renderedMessage = message
		}

		// Send notification
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		err = provider.SendNotification(ctx, notification.Settings, renderedMessage)
		cancel()

		// Record attempt
		s.recordNotificationHistory(notification, renderedMessage, err, startTime)

		if err == nil {
			// Success
			s.logger.Debug("Notification sent successfully",
				"provider", provider.GetName(),
				"attempt", attempt+1)
			return nil
		}

		lastError = err

		// Check if we should retry
		if attempt < retryConfig.MaxRetries && provider.SupportsRetry() &&
			retryConfig.RetryCondition != nil && retryConfig.RetryCondition(err) {
			// Calculate delay with exponential backoff
			delay := time.Duration(float64(retryConfig.InitialDelay) *
				pow(retryConfig.BackoffFactor, float64(attempt)))
			if delay > retryConfig.MaxDelay {
				delay = retryConfig.MaxDelay
			}

			s.logger.Warn("Notification failed, retrying",
				"provider", provider.GetName(),
				"attempt", attempt+1,
				"error", err,
				"delay", delay)

			time.Sleep(delay)
		} else {
			break
		}
	}

	return fmt.Errorf("notification failed after %d attempts: %w",
		retryConfig.MaxRetries+1, lastError)
}

// applyTemplate applies template rendering to a notification message
func (s *NotificationService) applyTemplate(
	notification *models.Notification,
	message *notifications.NotificationMessage) (*notifications.NotificationMessage, error) {
	// Get default template for event type
	template, exists := s.defaultTemplates[message.EventType]
	if !exists {
		return message, nil // No template available
	}

	// Render template
	rendered, err := s.templateEngine.RenderTemplate(template, message)
	if err != nil {
		return message, err
	}

	// Create new message with rendered content
	renderedMessage := *message // Copy
	renderedMessage.Subject = rendered.Subject
	renderedMessage.Body = rendered.Body

	return &renderedMessage, nil
}

// recordNotificationHistory records a notification attempt in the database
func (s *NotificationService) recordNotificationHistory(
	notification *models.Notification,
	message *notifications.NotificationMessage,
	err error,
	startTime time.Time) {
	history := &models.NotificationHistory{
		NotificationID: notification.ID,
		EventType:      message.EventType,
		Subject:        message.Subject,
		Body:           message.Body,
		Successful:     err == nil,
		SentAt:         startTime,
	}

	if message.Movie != nil {
		history.MovieID = &message.Movie.ID
	}

	if err != nil {
		history.ErrorMessage = err.Error()
	}

	if dbErr := s.db.GORM.Create(history).Error; dbErr != nil {
		s.logger.Error("Failed to record notification history", "error", dbErr)
	}
}

// GetNotificationStats returns statistics about notifications
func (s *NotificationService) GetNotificationStats() (map[string]interface{}, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	stats := make(map[string]interface{})

	// Total notifications
	var totalNotifications int64
	if err := s.db.GORM.Model(&models.Notification{}).Count(&totalNotifications).Error; err != nil {
		return nil, fmt.Errorf("failed to count total notifications: %w", err)
	}
	stats["total"] = totalNotifications

	// Enabled notifications
	var enabledNotifications int64
	err := s.db.GORM.Model(&models.Notification{}).Where("enabled = ?", true).Count(&enabledNotifications).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count enabled notifications: %w", err)
	}
	stats["enabled"] = enabledNotifications

	return stats, nil
}

// buildSubject builds a subject line for a notification event
func (s *NotificationService) buildSubject(event *models.NotificationEvent) string {
	switch event.EventType {
	case "grab":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - Grabbed", event.Movie.Title, event.Movie.Year)
		}
		return "Movie Grabbed"
	case "download":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - Downloaded", event.Movie.Title, event.Movie.Year)
		}
		return "Movie Downloaded"
	case "upgrade":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - Upgraded", event.Movie.Title, event.Movie.Year)
		}
		return "Movie Upgraded"
	case "rename":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - Renamed", event.Movie.Title, event.Movie.Year)
		}
		return "Movie Renamed"
	case "movieAdded":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - Added", event.Movie.Title, event.Movie.Year)
		}
		return "Movie Added"
	case "movieDelete":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - Deleted", event.Movie.Title, event.Movie.Year)
		}
		return "Movie Deleted"
	case "movieFileDelete":
		if event.Movie != nil {
			return fmt.Sprintf("%s (%d) - File Deleted", event.Movie.Title, event.Movie.Year)
		}
		return "Movie File Deleted"
	case "health":
		if event.HealthCheck != nil {
			return fmt.Sprintf("Radarr Health Issue - %s", event.HealthCheck.Type)
		}
		return "Radarr Health Issue"
	case "applicationUpdate":
		return "Radarr Application Update Available"
	default:
		return fmt.Sprintf("Radarr - %s", event.EventType)
	}
}

// buildBody builds a message body for a notification event
func (s *NotificationService) buildBody(event *models.NotificationEvent) string {
	if event.Message != "" {
		return event.Message
	}

	switch event.EventType {
	case "grab":
		if event.Movie != nil {
			body := fmt.Sprintf("Movie '%s (%d)' was grabbed", event.Movie.Title, event.Movie.Year)
			if event.DownloadClient != "" {
				body += fmt.Sprintf(" from %s", event.DownloadClient)
			}
			if event.Quality != nil {
				body += fmt.Sprintf(".\n\nQuality: %s", event.Quality.Name)
			}
			if event.SourceTitle != "" {
				body += fmt.Sprintf("\nSource: %s", event.SourceTitle)
			}
			return body
		}
		return "A movie was grabbed from an indexer."
	case "download":
		if event.Movie != nil {
			body := fmt.Sprintf("Movie '%s (%d)' has been downloaded and imported", event.Movie.Title, event.Movie.Year)
			if event.Quality != nil {
				body += fmt.Sprintf(".\n\nQuality: %s", event.Quality.Name)
			}
			if event.MovieFile != nil {
				body += fmt.Sprintf("\nFile: %s", event.MovieFile.RelativePath)
				if event.MovieFile.Size > 0 {
					body += fmt.Sprintf("\nSize: %s", s.formatBytes(event.MovieFile.Size))
				}
			}
			return body
		}
		return "A movie has been downloaded and imported."
	case "health":
		if event.HealthCheck != nil {
			body := fmt.Sprintf("A health issue has been detected in Radarr.\n\nType: %s\nStatus: %s\nMessage: %s",
				event.HealthCheck.Type, event.HealthCheck.Status, event.HealthCheck.Message)
			if event.HealthCheck.WikiURL != "" {
				body += fmt.Sprintf("\n\nMore info: %s", event.HealthCheck.WikiURL)
			}
			return body
		}
		return "A health issue has been detected."
	default:
		return fmt.Sprintf("A %s event occurred.", event.EventType)
	}
}

// formatBytes formats byte size into human-readable format
func (s *NotificationService) formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// pow calculates base^exp for float64
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	if exp == 1 {
		return base
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}

// GetProviderInfo returns information about available notification providers
func (s *NotificationService) GetProviderInfo() ([]*notifications.ProviderInfo, error) {
	if factory, ok := s.factory.(*notifications.DefaultProviderFactory); ok {
		return factory.GetAllProviderInfo()
	}
	return nil, fmt.Errorf("factory does not support GetAllProviderInfo")
}

// GetProviderFields returns configuration fields for a specific provider type
func (s *NotificationService) GetProviderFields(providerType models.NotificationType) ([]models.NotificationField, error) {
	provider, err := s.factory.CreateProvider(providerType)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}
	return provider.GetConfigFields(), nil
}

// GetNotificationHistory retrieves notification history with optional filters
func (s *NotificationService) GetNotificationHistory(limit int, offset int) ([]models.NotificationHistory, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	var history []models.NotificationHistory
	query := s.db.GORM.Preload("Notification").Preload("Movie")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Order("sent_at DESC").Find(&history).Error; err != nil {
		s.logger.Error("Failed to fetch notification history", "error", err)
		return nil, fmt.Errorf("failed to fetch notification history: %w", err)
	}

	return history, nil
}
