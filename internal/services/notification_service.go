package services

import (
	"fmt"
	"net/http"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"gorm.io/gorm"
)

// NotificationService provides operations for managing notifications and alerts.
type NotificationService struct {
	db         *database.Database
	logger     *logger.Logger
	httpClient *http.Client
}

// NewNotificationService creates a new instance of NotificationService with the provided database and logger.
func NewNotificationService(db *database.Database, logger *logger.Logger) *NotificationService {
	return &NotificationService{
		db:     db,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
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

	// Test actual notification sending
	testMessage := &models.NotificationMessage{
		Subject: "Test Notification from Radarr",
		Body:    "This is a test notification to verify your configuration is working correctly.",
		Data: map[string]interface{}{
			"test": true,
		},
	}

	if err := s.sendNotificationMessage(notification, testMessage); err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, err.Error())
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

	for _, notification := range notifications {
		if !s.shouldSendNotification(&notification, eventType) {
			continue
		}

		message := s.buildNotificationMessage(&notification, event)
		if err := s.sendAndRecordNotification(&notification, message, event); err != nil {
			s.logger.Error("Failed to send notification", "id", notification.ID, 
				"name", notification.Name, "error", err)
		}
	}

	return nil
}

// shouldSendNotification checks if a notification should be sent for an event type
func (s *NotificationService) shouldSendNotification(notification *models.Notification, eventType string) bool {
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
	default:
		return false
	}
}

// Add essential helper methods
func (s *NotificationService) validateNotification(notification *models.Notification) error {
	if notification.Name == "" {
		return fmt.Errorf("name is required")
	}
	if notification.Implementation == "" {
		return fmt.Errorf("implementation is required")
	}
	return nil
}

func (s *NotificationService) buildNotificationMessage(
	_ *models.Notification,
	event *models.NotificationEvent) *models.NotificationMessage {
	return &models.NotificationMessage{
		Subject: fmt.Sprintf("Radarr - %s", event.EventType),
		Body:    event.Message,
		Movie:   event.Movie,
		Data:    event.Data,
	}
}

func (s *NotificationService) sendAndRecordNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
	event *models.NotificationEvent) error {
	startTime := time.Now()
	err := s.sendNotificationMessage(notification, message)

	// Record the notification attempt
	history := &models.NotificationHistory{
		NotificationID: notification.ID,
		EventType:      event.EventType,
		Subject:        message.Subject,
		Body:           message.Body,
		Successful:     err == nil,
		SentAt:         startTime,
	}

	if event.Movie != nil {
		history.MovieID = &event.Movie.ID
	}

	if err != nil {
		history.ErrorMessage = err.Error()
	}

	if dbErr := s.db.GORM.Create(history).Error; dbErr != nil {
		s.logger.Error("Failed to record notification history", "error", dbErr)
	}

	return err
}

func (s *NotificationService) sendNotificationMessage(
	notification *models.Notification,
	message *models.NotificationMessage) error {
	switch notification.Implementation {
	case models.NotificationTypeDiscord:
		return s.sendDiscordNotification(notification, message)
	case models.NotificationTypeSlack:
		return s.sendSlackNotification(notification, message)
	case models.NotificationTypeEmail:
		return s.sendEmailNotification(notification, message)
	case models.NotificationTypeWebhook:
		return s.sendWebhookNotification(notification, message)
	case models.NotificationTypeTelegram,
		models.NotificationTypePushover,
		models.NotificationTypePushbullet,
		models.NotificationTypeGotify,
		models.NotificationTypeJoin,
		models.NotificationTypeApprise,
		models.NotificationTypeNotifiarr,
		models.NotificationTypeMailgun,
		models.NotificationTypeSendGrid,
		models.NotificationTypePlex,
		models.NotificationTypeEmby,
		models.NotificationTypeJellyfin,
		models.NotificationTypeKodi,
		models.NotificationTypeCustomScript,
		models.NotificationTypeSynologyIndexer,
		models.NotificationTypeTwitter,
		models.NotificationTypeSignal,
		models.NotificationTypeMatrix,
		models.NotificationTypeNtfy:
		s.logger.Debug("Would send notification", "type", notification.Implementation, "subject", message.Subject)
		return nil // Placeholder - would implement actual sending
	default:
		s.logger.Debug("Unknown notification type", "type", notification.Implementation)
		return nil
	}
}

// Simplified placeholder methods for notification providers
func (s *NotificationService) sendDiscordNotification(
	notification *models.Notification,
	message *models.NotificationMessage) error {
	s.logger.Debug("Would send Discord notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// Simplified placeholder methods for notification providers
func (s *NotificationService) sendSlackNotification(
	notification *models.Notification,
	message *models.NotificationMessage) error {
	s.logger.Debug("Would send Slack notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

func (s *NotificationService) sendEmailNotification(
	notification *models.Notification,
	message *models.NotificationMessage) error {
	s.logger.Debug("Would send email notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

func (s *NotificationService) sendWebhookNotification(
	notification *models.Notification,
	message *models.NotificationMessage) error {
	s.logger.Debug("Would send webhook notification", "notification", notification.Name, "subject", message.Subject)
	return nil
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
