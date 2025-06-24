package services

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// NotificationService provides operations for managing notifications and alerts.
type NotificationService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewNotificationService creates a new instance of NotificationService with the provided database and logger.
func NewNotificationService(db *database.Database, logger *logger.Logger) *NotificationService {
	return &NotificationService{
		db:     db,
		logger: logger,
	}
}

// GetNotifications retrieves all configured notifications from the system.
func (s *NotificationService) GetNotifications() ([]*models.Notification, error) {
	var notifications []*models.Notification

	if err := s.db.GORM.Find(&notifications).Error; err != nil {
		s.logger.Error("Failed to fetch notifications", "error", err)
		return nil, fmt.Errorf("failed to fetch notifications: %w", err)
	}

	return notifications, nil
}

// GetNotificationByID retrieves a specific notification by its ID.
func (s *NotificationService) GetNotificationByID(id int) (*models.Notification, error) {
	var notification models.Notification

	if err := s.db.GORM.Where("id = ?", id).First(&notification).Error; err != nil {
		s.logger.Error("Failed to fetch notification", "id", id, "error", err)
		return nil, fmt.Errorf("failed to fetch notification with id %d: %w", id, err)
	}

	return &notification, nil
}

// CreateNotification creates a new notification configuration.
func (s *NotificationService) CreateNotification(notification *models.Notification) error {
	if err := s.db.GORM.Create(notification).Error; err != nil {
		s.logger.Error("Failed to create notification", "name", notification.Name, "error", err)
		return fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Info("Created notification", "id", notification.ID, "name", notification.Name, "type", notification.Type)
	return nil
}

// UpdateNotification updates an existing notification configuration.
func (s *NotificationService) UpdateNotification(notification *models.Notification) error {
	if err := s.db.GORM.Save(notification).Error; err != nil {
		s.logger.Error("Failed to update notification", "id", notification.ID, "error", err)
		return fmt.Errorf("failed to update notification: %w", err)
	}

	s.logger.Info("Updated notification", "id", notification.ID, "name", notification.Name)
	return nil
}

// DeleteNotification removes a notification configuration.
func (s *NotificationService) DeleteNotification(id int) error {
	result := s.db.GORM.Delete(&models.Notification{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete notification", "id", id, "error", result.Error)
		return fmt.Errorf("failed to delete notification: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("notification with id %d not found", id)
	}

	s.logger.Info("Deleted notification", "id", id)
	return nil
}

// GetEnabledNotifications retrieves all enabled notifications.
func (s *NotificationService) GetEnabledNotifications() ([]*models.Notification, error) {
	var notifications []*models.Notification

	if err := s.db.GORM.Where("enable = ?", true).Find(&notifications).Error; err != nil {
		s.logger.Error("Failed to fetch enabled notifications", "error", err)
		return nil, fmt.Errorf("failed to fetch enabled notifications: %w", err)
	}

	return notifications, nil
}

// TestNotification tests a notification configuration.
func (s *NotificationService) TestNotification(
	notification *models.Notification,
) (*models.NotificationTestResult, error) {
	// Basic validation
	errors := []string{}

	if notification.Name == "" {
		errors = append(errors, "Name is required")
	}

	if notification.Type == "" {
		errors = append(errors, "Type is required")
	}

	result := &models.NotificationTestResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}

	if !result.IsValid {
		return result, nil
	}

	// TODO: Implement actual notification testing
	// This would involve sending a test notification based on the type

	s.logger.Info("Tested notification", "name", notification.Name, "type", notification.Type, "valid", result.IsValid)
	return result, nil
}

// SendNotification sends a notification message.
func (s *NotificationService) SendNotification(
	trigger models.NotificationTrigger,
	message *models.NotificationMessage,
) error {
	notifications, err := s.GetEnabledNotifications()
	if err != nil {
		return fmt.Errorf("failed to get enabled notifications: %w", err)
	}

	for _, notification := range notifications {
		if !notification.ShouldNotifyFor(trigger) {
			continue
		}

		// Send the notification
		err := s.sendNotificationToProvider(notification, message)

		// Record the notification in history
		history := &models.NotificationHistory{
			NotificationID: notification.ID,
			Trigger:        trigger,
			Subject:        message.Subject,
			Body:           message.Body,
			Successful:     err == nil,
			SentAt:         time.Now(),
		}

		if message.Movie != nil {
			history.MovieID = &message.Movie.ID
		}

		if err != nil {
			history.ErrorMessage = err.Error()
			s.logger.Error("Failed to send notification", "notification", notification.Name, "error", err)
		} else {
			s.logger.Info("Sent notification", "notification", notification.Name, "trigger", trigger)
		}

		// Save to history
		if histErr := s.db.GORM.Create(history).Error; histErr != nil {
			s.logger.Error("Failed to save notification history", "error", histErr)
		}
	}

	return nil
}

// sendNotificationToProvider sends a notification to a specific provider.
func (s *NotificationService) sendNotificationToProvider(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement actual notification sending based on provider type
	// This would involve HTTP requests to various notification services

	switch notification.Type {
	case models.NotificationTypeDiscord:
		return s.sendDiscordNotification(notification, message)
	case models.NotificationTypeSlack:
		return s.sendSlackNotification(notification, message)
	case models.NotificationTypeEmail:
		return s.sendEmailNotification(notification, message)
	case models.NotificationTypeWebhook:
		return s.sendWebhookNotification(notification, message)
	case models.NotificationTypePushbullet:
		return s.sendPushbulletNotification(notification, message)
	case models.NotificationTypePushover:
		return s.sendPushoverNotification(notification, message)
	case models.NotificationTypeTelegram:
		return s.sendTelegramNotification(notification, message)
	case models.NotificationTypePlex:
		return s.sendPlexNotification(notification, message)
	case models.NotificationTypeEmby:
		return s.sendEmbyNotification(notification, message)
	case models.NotificationTypeJellyfin:
		return s.sendJellyfinNotification(notification, message)
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type)
	}
}

// sendDiscordNotification sends a notification to Discord.
func (s *NotificationService) sendDiscordNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Discord webhook notification
	s.logger.Debug("Would send Discord notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendSlackNotification sends a notification to Slack.
func (s *NotificationService) sendSlackNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Slack webhook notification
	s.logger.Debug("Would send Slack notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendEmailNotification sends an email notification.
func (s *NotificationService) sendEmailNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement SMTP email notification
	s.logger.Debug("Would send email notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendWebhookNotification sends a generic webhook notification.
func (s *NotificationService) sendWebhookNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement generic webhook notification
	s.logger.Debug("Would send webhook notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendPushbulletNotification sends a notification to Pushbullet.
func (s *NotificationService) sendPushbulletNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Pushbullet notification
	s.logger.Debug("Would send Pushbullet notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendPushoverNotification sends a notification to Pushover.
func (s *NotificationService) sendPushoverNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Pushover notification
	s.logger.Debug("Would send Pushover notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendTelegramNotification sends a notification to Telegram.
func (s *NotificationService) sendTelegramNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Telegram notification
	s.logger.Debug("Would send Telegram notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendPlexNotification sends a notification to Plex.
func (s *NotificationService) sendPlexNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Plex notification
	s.logger.Debug("Would send Plex notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendEmbyNotification sends a notification to Emby.
func (s *NotificationService) sendEmbyNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Emby notification
	s.logger.Debug("Would send Emby notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// sendJellyfinNotification sends a notification to Jellyfin.
func (s *NotificationService) sendJellyfinNotification(
	notification *models.Notification,
	message *models.NotificationMessage,
) error {
	// TODO: Implement Jellyfin notification
	s.logger.Debug("Would send Jellyfin notification", "notification", notification.Name, "subject", message.Subject)
	return nil
}

// GetNotificationHistory retrieves notification history.
func (s *NotificationService) GetNotificationHistory(limit int) ([]*models.NotificationHistory, error) {
	var history []*models.NotificationHistory

	query := s.db.GORM.Preload("Notification").Preload("Movie").Order("sent_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&history).Error; err != nil {
		s.logger.Error("Failed to fetch notification history", "error", err)
		return nil, fmt.Errorf("failed to fetch notification history: %w", err)
	}

	return history, nil
}

// GetHealthChecks retrieves system health checks.
func (s *NotificationService) GetHealthChecks() ([]*models.HealthCheck, error) {
	var checks []*models.HealthCheck

	if err := s.db.GORM.Order("time DESC").Find(&checks).Error; err != nil {
		s.logger.Error("Failed to fetch health checks", "error", err)
		return nil, fmt.Errorf("failed to fetch health checks: %w", err)
	}

	return checks, nil
}

// AddHealthCheck adds a new health check result.
func (s *NotificationService) AddHealthCheck(check *models.HealthCheck) error {
	if err := s.db.GORM.Create(check).Error; err != nil {
		s.logger.Error("Failed to add health check", "source", check.Source, "error", err)
		return fmt.Errorf("failed to add health check: %w", err)
	}

	// Send notification if it's an error or warning
	if check.IsError() || check.IsWarning() {
		message := &models.NotificationMessage{
			Type:    models.NotificationTriggerOnHealth,
			Subject: fmt.Sprintf("Health Check: %s", check.Source),
			Body:    check.Message,
			Data: map[string]interface{}{
				"source": check.Source,
				"type":   check.Type,
				"status": check.Status,
			},
		}

		if err := s.SendNotification(models.NotificationTriggerOnHealth, message); err != nil {
			s.logger.Error("Failed to send health notification", "error", err)
		}
	}

	s.logger.Info("Added health check", "id", check.ID, "source", check.Source, "status", check.Status)
	return nil
}
