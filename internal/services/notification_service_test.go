package services

import (
	"context"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/services/notifications"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationService_CreateNotification(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	notification := &models.Notification{
		Name:           "Test Discord Notification",
		Implementation: models.NotificationTypeDiscord,
		Settings: models.NotificationSettings{
			"webhookUrl": "https://discord.com/api/webhooks/test",
		},
		OnGrab:     true,
		OnDownload: true,
		Enabled:    true,
	}

	// Test with nil database
	err := service.CreateNotification(notification)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestNotificationService_GetNotifications(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	// Test with nil database
	_, err := service.GetNotifications()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestNotificationService_TestNotification(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	notification := &models.Notification{
		Name:           "Test Notification",
		Implementation: models.NotificationTypeDiscord,
		Settings: models.NotificationSettings{
			"webhookUrl": "https://discord.com/api/webhooks/test",
		},
		Enabled: true,
	}

	result, err := service.TestNotification(notification)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Since we're using a stub provider, the test should pass validation
	assert.True(t, result.IsValid)
}

func TestNotificationService_SendNotification(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	// Create test event
	event := &models.NotificationEvent{
		Type:      "grab",
		EventType: "grab",
		Movie: &models.Movie{
			ID:    1,
			Title: "Test Movie",
			Year:  2023,
		},
		Message: "Test movie grabbed",
	}

	// Test with nil database - should return error
	err := service.SendNotification(event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database not available")
}

func TestNotificationService_GetProviderInfo(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	providers, err := service.GetProviderInfo()
	assert.NoError(t, err)
	assert.Greater(t, len(providers), 0)

	// Check that Discord provider is available
	var discordProvider *notifications.ProviderInfo
	for _, provider := range providers {
		if provider.Type == models.NotificationTypeDiscord {
			discordProvider = provider
			break
		}
	}

	assert.NotNil(t, discordProvider)
	assert.Equal(t, "Discord", discordProvider.Name)
	assert.True(t, discordProvider.IsEnabled)
}

func TestNotificationService_GetProviderFields(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	fields, err := service.GetProviderFields(models.NotificationTypeDiscord)
	assert.NoError(t, err)
	assert.Greater(t, len(fields), 0)

	// Check that the enabled field exists (from stub provider)
	var enabledField *models.NotificationField
	for _, field := range fields {
		if field.Name == "enabled" {
			enabledField = &field
			break
		}
	}

	assert.NotNil(t, enabledField)
	assert.Equal(t, "checkbox", enabledField.Type)
}

func TestNotificationTemplateRendering(t *testing.T) {
	logger := logger.New(config.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})

	templateEngine := notifications.NewTemplateEngine(logger)

	// Test template rendering
	template := &notifications.NotificationTemplate{
		EventType: "grab",
		Subject:   "{movie.title} ({movie.year}) - Grabbed",
		Body:      "Movie '{movie.title} ({movie.year})' was grabbed.\n\nQuality: {quality.name}",
	}

	message := &notifications.NotificationMessage{
		EventType: "grab",
		Movie: &models.Movie{
			Title: "Test Movie",
			Year:  2023,
		},
		Quality: &models.QualityDefinition{
			Name: "HD-1080p",
		},
		Timestamp: time.Now(),
	}

	result, err := templateEngine.RenderTemplate(template, message)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	assert.Contains(t, result.Subject, "Test Movie (2023)")
	assert.Contains(t, result.Body, "Test Movie (2023)")
	assert.Contains(t, result.Body, "HD-1080p")
}

func TestNotificationProviderFactory(t *testing.T) {
	logger := logger.New(config.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})

	factory := notifications.NewProviderFactory(logger)

	// Test provider creation
	provider, err := factory.CreateProvider(models.NotificationTypeDiscord)
	assert.NoError(t, err)
	assert.NotNil(t, provider)
	assert.Equal(t, models.NotificationTypeDiscord, provider.GetType())
	assert.Equal(t, "Discord", provider.GetName())

	// Test unsupported provider
	_, err = factory.CreateProvider("UnsupportedProvider")
	assert.Error(t, err)

	// Test supported types
	types := factory.GetSupportedTypes()
	assert.Greater(t, len(types), 0)
	assert.Contains(t, types, models.NotificationTypeDiscord)
	assert.Contains(t, types, models.NotificationTypeSlack)
	assert.Contains(t, types, models.NotificationTypeEmail)
}

func TestNotificationValidation(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := NewNotificationService(nil, logger)

	// Since database operations fail with nil database, test provider validation instead

	// Test unsupported provider type
	notification := &models.Notification{
		Name:           "Test",
		Implementation: "UnsupportedProvider",
		Settings:       models.NotificationSettings{},
		Enabled:        true,
	}

	result, err := service.TestNotification(notification)
	assert.NoError(t, err) // Method doesn't return error, it returns result with error info
	assert.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Greater(t, len(result.Errors), 0)
	assert.Contains(t, result.Errors[0], "unsupported")
}

func TestStubProviderIntegration(t *testing.T) {
	logger := logger.New(config.LogConfig{
		Level:  "debug",
		Format: "json",
		Output: "stdout",
	})

	factory := notifications.NewProviderFactory(logger)
	provider, err := factory.CreateProvider(models.NotificationTypeDiscord)
	require.NoError(t, err)

	// Test provider capabilities
	capabilities := provider.GetCapabilities()
	assert.True(t, capabilities.OnGrab)
	assert.True(t, capabilities.OnDownload)
	assert.True(t, capabilities.OnUpgrade)

	// Test configuration validation
	settings := models.NotificationSettings{
		"enabled": true,
	}
	err = provider.ValidateConfig(settings)
	assert.NoError(t, err)

	// Test connection
	ctx := context.Background()
	err = provider.TestConnection(ctx, settings)
	assert.NoError(t, err)

	// Test sending notification
	message := &notifications.NotificationMessage{
		Subject:   "Test Subject",
		Body:      "Test Body",
		EventType: "test",
		IsTest:    true,
		Timestamp: time.Now(),
	}

	err = provider.SendNotification(ctx, settings, message)
	assert.NoError(t, err)
}
