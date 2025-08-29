package notifications

import (
	"context"
	"time"

	"github.com/radarr/radarr-go/internal/models"
)

// Provider is the interface that all notification providers must implement
type Provider interface {
	// GetName returns the human-readable name of the provider
	GetName() string

	// GetType returns the notification type this provider implements
	GetType() models.NotificationType

	// GetConfigFields returns the configuration fields required for this provider
	GetConfigFields() []models.NotificationField

	// ValidateConfig validates the provider configuration
	ValidateConfig(settings models.NotificationSettings) error

	// SendNotification sends a notification using the provider
	SendNotification(ctx context.Context, settings models.NotificationSettings, message *NotificationMessage) error

	// TestConnection tests the provider configuration with a test message
	TestConnection(ctx context.Context, settings models.NotificationSettings) error

	// GetCapabilities returns what events this provider supports
	GetCapabilities() ProviderCapabilities

	// SupportsRetry returns true if this provider supports retry logic
	SupportsRetry() bool

	// GetDefaultRetryConfig returns default retry configuration
	GetDefaultRetryConfig() RetryConfig
}

// ProviderCapabilities defines what events a provider can handle
type ProviderCapabilities struct {
	OnGrab                      bool
	OnDownload                  bool
	OnUpgrade                   bool
	OnRename                    bool
	OnMovieAdded                bool
	OnMovieDelete               bool
	OnMovieFileDelete           bool
	OnHealthIssue               bool
	OnApplicationUpdate         bool
	OnManualInteractionRequired bool
	SupportsCustomTemplates     bool
	SupportsRichContent         bool // HTML, embeds, etc.
}

// RetryConfig defines retry behavior for a provider
type RetryConfig struct {
	MaxRetries     int
	InitialDelay   time.Duration
	MaxDelay       time.Duration
	BackoffFactor  float64
	RetryCondition func(error) bool
}

// NotificationMessage represents a notification message with context
type NotificationMessage struct {
	// Core message fields
	Subject string `json:"subject"`
	Body    string `json:"body"`

	// Event context
	EventType      string                    `json:"eventType"`
	Movie          *models.Movie             `json:"movie,omitempty"`
	MovieFile      *models.MovieFile         `json:"movieFile,omitempty"`
	DeletedFiles   []models.MovieFile        `json:"deletedFiles,omitempty"`
	OldMovieFile   *models.MovieFile         `json:"oldMovieFile,omitempty"`
	SourceTitle    string                    `json:"sourceTitle,omitempty"`
	Quality        *models.QualityDefinition `json:"quality,omitempty"`
	QualityUpgrade bool                      `json:"qualityUpgrade"`
	DownloadClient string                    `json:"downloadClient,omitempty"`
	DownloadID     string                    `json:"downloadId,omitempty"`
	HealthCheck    *models.HealthCheck       `json:"healthCheck,omitempty"`

	// Additional context data
	Data map[string]interface{} `json:"data,omitempty"`

	// Template-specific fields
	IsTest     bool      `json:"isTest"`
	ServerURL  string    `json:"serverUrl,omitempty"`
	ServerName string    `json:"serverName,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// NotificationResult represents the result of sending a notification
type NotificationResult struct {
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
	Duration   time.Duration `json:"duration"`
	RetryCount int           `json:"retryCount"`
	MessageID  string        `json:"messageId,omitempty"` // For providers that return message IDs
}

// ProviderFactory creates notification providers
type ProviderFactory interface {
	// CreateProvider creates a provider instance for the given type
	CreateProvider(providerType models.NotificationType) (Provider, error)

	// GetSupportedTypes returns all supported notification types
	GetSupportedTypes() []models.NotificationType

	// GetProviderInfo returns information about a provider type
	GetProviderInfo(providerType models.NotificationType) (*ProviderInfo, error)
}

// ProviderInfo contains metadata about a notification provider
type ProviderInfo struct {
	Type         models.NotificationType    `json:"type"`
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	Website      string                     `json:"website,omitempty"`
	DocsURL      string                     `json:"docsUrl,omitempty"`
	Capabilities ProviderCapabilities       `json:"capabilities"`
	ConfigFields []models.NotificationField `json:"configFields"`
	IsEnabled    bool                       `json:"isEnabled"`
	Version      string                     `json:"version,omitempty"`
}

// NotificationTemplate represents a customizable notification template
type NotificationTemplate struct {
	EventType      string            `json:"eventType"`
	Subject        string            `json:"subject"`
	Body           string            `json:"body"`
	BodyHTML       string            `json:"bodyHtml,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
	DefaultSubject string            `json:"defaultSubject"`
	DefaultBody    string            `json:"defaultBody"`
}

// TemplateEngine handles template rendering and variable substitution
type TemplateEngine interface {
	// RenderTemplate renders a template with the given context
	RenderTemplate(template *NotificationTemplate, message *NotificationMessage) (*RenderedTemplate, error)

	// GetAvailableVariables returns all available template variables for an event type
	GetAvailableVariables(eventType string) map[string]string

	// ValidateTemplate validates a template for syntax errors
	ValidateTemplate(template *NotificationTemplate) error
}

// RenderedTemplate represents a rendered notification template
type RenderedTemplate struct {
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	BodyHTML string `json:"bodyHtml,omitempty"`
}

// NotificationQueue handles queuing and retry logic for failed notifications
type NotificationQueue interface {
	// Enqueue adds a notification to the queue
	Enqueue(item *QueueItem) error

	// Dequeue gets the next notification from the queue
	Dequeue() (*QueueItem, error)

	// MarkSuccess marks a notification as successfully sent
	MarkSuccess(itemID string) error

	// MarkFailed marks a notification as failed and schedules retry if applicable
	MarkFailed(itemID string, err error) error

	// GetQueueStats returns statistics about the queue
	GetQueueStats() *QueueStats
}

// QueueItem represents a queued notification
type QueueItem struct {
	ID             string                      `json:"id"`
	NotificationID int                         `json:"notificationId"`
	Message        *NotificationMessage        `json:"message"`
	Settings       models.NotificationSettings `json:"settings"`
	ProviderType   models.NotificationType     `json:"providerType"`
	AttemptCount   int                         `json:"attemptCount"`
	NextAttempt    time.Time                   `json:"nextAttempt"`
	MaxRetries     int                         `json:"maxRetries"`
	CreatedAt      time.Time                   `json:"createdAt"`
	LastError      string                      `json:"lastError,omitempty"`
}

// QueueStats provides statistics about the notification queue
type QueueStats struct {
	Pending    int `json:"pending"`
	Processing int `json:"processing"`
	Failed     int `json:"failed"`
	Completed  int `json:"completed"`
	Total      int `json:"total"`
}

// HealthChecker performs health checks for notification providers
type HealthChecker interface {
	// CheckHealth performs a health check for a provider
	CheckHealth(ctx context.Context, provider Provider, settings models.NotificationSettings) *HealthCheckResult

	// GetHealthStatus returns the overall health status of all providers
	GetHealthStatus() map[string]*HealthCheckResult
}

// HealthCheckResult represents the result of a provider health check
type HealthCheckResult struct {
	Healthy      bool                   `json:"healthy"`
	ResponseTime time.Duration          `json:"responseTime"`
	Error        string                 `json:"error,omitempty"`
	LastChecked  time.Time              `json:"lastChecked"`
	StatusCode   int                    `json:"statusCode,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// EventBus handles distributing notification events to providers
type EventBus interface {
	// Subscribe subscribes to notification events
	Subscribe(eventType string, handler EventHandler)

	// Unsubscribe unsubscribes from notification events
	Unsubscribe(eventType string, handler EventHandler)

	// Publish publishes a notification event
	Publish(event *NotificationEvent) error
}

// EventHandler handles notification events
type EventHandler func(event *NotificationEvent) error

// NotificationEvent represents an event that can trigger notifications
type NotificationEvent struct {
	Type              string                    `json:"type"`
	EventType         string                    `json:"eventType"`
	Movie             *models.Movie             `json:"movie,omitempty"`
	MovieFile         *models.MovieFile         `json:"movieFile,omitempty"`
	DeletedFiles      []models.MovieFile        `json:"deletedFiles,omitempty"`
	OldMovieFile      *models.MovieFile         `json:"oldMovieFile,omitempty"`
	SourceTitle       string                    `json:"sourceTitle,omitempty"`
	Quality           *models.QualityDefinition `json:"quality,omitempty"`
	QualityUpgrade    bool                      `json:"qualityUpgrade"`
	DownloadClient    string                    `json:"downloadClient,omitempty"`
	DownloadID        string                    `json:"downloadId,omitempty"`
	Message           string                    `json:"message,omitempty"`
	HealthCheck       *models.HealthCheck       `json:"healthCheck,omitempty"`
	ApplicationUpdate *models.ApplicationUpdate `json:"updateChanges,omitempty"`
	ManualInteraction *models.ManualInteraction `json:"manualInteraction,omitempty"`
	Data              map[string]interface{}    `json:"data,omitempty"`
	Timestamp         time.Time                 `json:"timestamp"`
}
