package notifications

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// DefaultProviderFactory is the default implementation of ProviderFactory
type DefaultProviderFactory struct {
	logger    *logger.Logger
	providers map[models.NotificationType]func() Provider
	mu        sync.RWMutex
}

// NewProviderFactory creates a new provider factory
func NewProviderFactory(logger *logger.Logger) *DefaultProviderFactory {
	factory := &DefaultProviderFactory{
		logger:    logger,
		providers: make(map[models.NotificationType]func() Provider),
	}

	// Register built-in providers
	factory.registerBuiltInProviders()

	return factory
}

// registerBuiltInProviders registers all built-in notification providers
func (f *DefaultProviderFactory) registerBuiltInProviders() {
	// Register stub providers for now - actual implementations will be added later
	providerTypes := []models.NotificationType{
		models.NotificationTypeDiscord,
		models.NotificationTypeSlack,
		models.NotificationTypeEmail,
		models.NotificationTypeWebhook,
		models.NotificationTypePushover,
		models.NotificationTypeTelegram,
		models.NotificationTypePushbullet,
		models.NotificationTypeGotify,
		models.NotificationTypeMailgun,
		models.NotificationTypeSendGrid,
		models.NotificationTypeCustomScript,
	}

	for _, providerType := range providerTypes {
		f.RegisterProvider(providerType, func(pt models.NotificationType) func() Provider {
			return func() Provider {
				return NewStubProvider(string(pt), pt, f.logger)
			}
		}(providerType))
	}
}

// RegisterProvider registers a new provider constructor
func (f *DefaultProviderFactory) RegisterProvider(providerType models.NotificationType, constructor func() Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.providers[providerType] = constructor
	f.logger.Debug("Registered notification provider", "type", providerType)
}

// CreateProvider creates a provider instance for the given type
func (f *DefaultProviderFactory) CreateProvider(providerType models.NotificationType) (Provider, error) {
	f.mu.RLock()
	constructor, exists := f.providers[providerType]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported notification provider type: %s", providerType)
	}

	provider := constructor()
	f.logger.Debug("Created notification provider", "type", providerType, "name", provider.GetName())

	return provider, nil
}

// GetSupportedTypes returns all supported notification types
func (f *DefaultProviderFactory) GetSupportedTypes() []models.NotificationType {
	f.mu.RLock()
	defer f.mu.RUnlock()

	types := make([]models.NotificationType, 0, len(f.providers))
	for providerType := range f.providers {
		types = append(types, providerType)
	}

	return types
}

// GetProviderInfo returns information about a provider type
func (f *DefaultProviderFactory) GetProviderInfo(providerType models.NotificationType) (*ProviderInfo, error) {
	provider, err := f.CreateProvider(providerType)
	if err != nil {
		return nil, err
	}

	info := &ProviderInfo{
		Type:         providerType,
		Name:         provider.GetName(),
		Capabilities: provider.GetCapabilities(),
		ConfigFields: provider.GetConfigFields(),
		IsEnabled:    true,
	}

	// Add provider-specific metadata
	switch providerType {
	case models.NotificationTypeDiscord:
		info.Description = "Send notifications to Discord channels via webhooks"
		info.Website = "https://discord.com"
		info.DocsURL = "https://support.discord.com/hc/en-us/articles/228383668"
		info.Version = "1.0.0"
	case models.NotificationTypeSlack:
		info.Description = "Send notifications to Slack channels via webhooks"
		info.Website = "https://slack.com"
		info.DocsURL = "https://api.slack.com/messaging/webhooks"
		info.Version = "1.0.0"
	case models.NotificationTypeEmail:
		info.Description = "Send email notifications via SMTP"
		info.Website = ""
		info.DocsURL = ""
		info.Version = "1.0.0"
	case models.NotificationTypeWebhook:
		info.Description = "Send HTTP POST notifications to any webhook endpoint"
		info.Website = ""
		info.DocsURL = ""
		info.Version = "1.0.0"
	case models.NotificationTypePushover:
		info.Description = "Send push notifications to mobile devices via Pushover"
		info.Website = "https://pushover.net"
		info.DocsURL = "https://pushover.net/api"
		info.Version = "1.0.0"
	case models.NotificationTypeTelegram:
		info.Description = "Send notifications via Telegram bot"
		info.Website = "https://telegram.org"
		info.DocsURL = "https://core.telegram.org/bots/api"
		info.Version = "1.0.0"
	case models.NotificationTypePushbullet:
		info.Description = "Send push notifications via Pushbullet"
		info.Website = "https://www.pushbullet.com"
		info.DocsURL = "https://docs.pushbullet.com"
		info.Version = "1.0.0"
	case models.NotificationTypeGotify:
		info.Description = "Send push notifications via Gotify server"
		info.Website = "https://gotify.net"
		info.DocsURL = "https://gotify.net/docs"
		info.Version = "1.0.0"
	case models.NotificationTypeMailgun:
		info.Description = "Send email notifications via Mailgun service"
		info.Website = "https://www.mailgun.com"
		info.DocsURL = "https://documentation.mailgun.com/en/latest/"
		info.Version = "1.0.0"
	case models.NotificationTypeSendGrid:
		info.Description = "Send email notifications via SendGrid service"
		info.Website = "https://sendgrid.com"
		info.DocsURL = "https://docs.sendgrid.com"
		info.Version = "1.0.0"
	case models.NotificationTypeCustomScript:
		info.Description = "Execute custom scripts with notification data"
		info.Website = ""
		info.DocsURL = ""
		info.Version = "1.0.0"
	case models.NotificationTypeJoin:
		info.Description = "Send push notifications via Join"
		info.Website = "https://joaoapps.com/join/"
		info.DocsURL = "https://joaoapps.com/join/api/"
		info.Version = "1.0.0"
	case models.NotificationTypeApprise:
		info.Description = "Send notifications via Apprise"
		info.Website = "https://github.com/caronc/apprise"
		info.DocsURL = "https://github.com/caronc/apprise/wiki"
		info.Version = "1.0.0"
	case models.NotificationTypeNotifiarr:
		info.Description = "Send notifications via Notifiarr"
		info.Website = "https://notifiarr.com"
		info.DocsURL = "https://notifiarr.com/docs"
		info.Version = "1.0.0"
	case models.NotificationTypePlex:
		info.Description = "Send notifications to Plex Media Server"
		info.Website = "https://www.plex.tv"
		info.DocsURL = "https://support.plex.tv/articles/115002267687-webhooks/"
		info.Version = "1.0.0"
	case models.NotificationTypeEmby:
		info.Description = "Send notifications to Emby Media Server"
		info.Website = "https://emby.media"
		info.DocsURL = "https://github.com/MediaBrowser/Emby/wiki/Webhooks"
		info.Version = "1.0.0"
	case models.NotificationTypeJellyfin:
		info.Description = "Send notifications to Jellyfin Media Server"
		info.Website = "https://jellyfin.org"
		info.DocsURL = "https://jellyfin.org/docs/general/server/webhooks"
		info.Version = "1.0.0"
	case models.NotificationTypeKodi:
		info.Description = "Send notifications to Kodi"
		info.Website = "https://kodi.tv"
		info.DocsURL = "https://kodi.wiki/view/JSON-RPC_API"
		info.Version = "1.0.0"
	case models.NotificationTypeSynologyIndexer:
		info.Description = "Send notifications via Synology Indexer"
		info.Website = "https://www.synology.com"
		info.DocsURL = ""
		info.Version = "1.0.0"
	case models.NotificationTypeTwitter:
		info.Description = "Send notifications via Twitter"
		info.Website = "https://twitter.com"
		info.DocsURL = "https://developer.twitter.com/en/docs"
		info.Version = "1.0.0"
	case models.NotificationTypeSignal:
		info.Description = "Send notifications via Signal"
		info.Website = "https://signal.org"
		info.DocsURL = "https://github.com/bbernhard/signal-cli-rest-api"
		info.Version = "1.0.0"
	case models.NotificationTypeMatrix:
		info.Description = "Send notifications via Matrix"
		info.Website = "https://matrix.org"
		info.DocsURL = "https://matrix.org/docs/guides/client-server-api"
		info.Version = "1.0.0"
	case models.NotificationTypeNtfy:
		info.Description = "Send notifications via Ntfy"
		info.Website = "https://ntfy.sh"
		info.DocsURL = "https://docs.ntfy.sh/"
		info.Version = "1.0.0"
	default:
		info.Description = fmt.Sprintf("Notification provider for %s", providerType)
		info.Version = "1.0.0"
	}

	return info, nil
}

// GetAllProviderInfo returns information about all supported providers
func (f *DefaultProviderFactory) GetAllProviderInfo() ([]*ProviderInfo, error) {
	types := f.GetSupportedTypes()
	infos := make([]*ProviderInfo, 0, len(types))

	for _, providerType := range types {
		info, err := f.GetProviderInfo(providerType)
		if err != nil {
			f.logger.Error("Failed to get provider info", "type", providerType, "error", err)
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// IsProviderSupported checks if a provider type is supported
func (f *DefaultProviderFactory) IsProviderSupported(providerType models.NotificationType) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	_, exists := f.providers[providerType]
	return exists
}

// GetProviderCount returns the number of registered providers
func (f *DefaultProviderFactory) GetProviderCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.providers)
}

// StubProvider implements a basic stub for notification providers
type StubProvider struct {
	name         string
	providerType models.NotificationType
	logger       *logger.Logger
}

// NewStubProvider creates a new stub provider
func NewStubProvider(name string, providerType models.NotificationType, logger *logger.Logger) *StubProvider {
	return &StubProvider{
		name:         name,
		providerType: providerType,
		logger:       logger,
	}
}

// GetName returns the human-readable name of the provider
func (p *StubProvider) GetName() string {
	return p.name
}

// GetType returns the notification type this provider implements
func (p *StubProvider) GetType() models.NotificationType {
	return p.providerType
}

// GetConfigFields returns basic configuration fields
func (p *StubProvider) GetConfigFields() []models.NotificationField {
	return []models.NotificationField{
		{
			Name:     "enabled",
			Label:    "Enabled",
			Type:     "checkbox",
			Advanced: false,
			Privacy:  "normal",
			Value:    true,
			HelpText: fmt.Sprintf("Enable %s notifications (implementation coming soon)", p.name),
			Order:    1,
		},
	}
}

// ValidateConfig validates the provider configuration
func (p *StubProvider) ValidateConfig(settings models.NotificationSettings) error {
	p.logger.Debug("Validating stub provider config", "provider", p.name)
	return nil
}

// SendNotification logs that it would send a notification
func (p *StubProvider) SendNotification(ctx context.Context, settings models.NotificationSettings, message *NotificationMessage) error {
	p.logger.Debug("Would send notification",
		"provider", p.name,
		"eventType", message.EventType,
		"subject", message.Subject)
	return nil
}

// TestConnection simulates a successful test
func (p *StubProvider) TestConnection(ctx context.Context, settings models.NotificationSettings) error {
	p.logger.Debug("Testing stub provider connection", "provider", p.name)
	return nil
}

// GetCapabilities returns basic capabilities
func (p *StubProvider) GetCapabilities() ProviderCapabilities {
	return ProviderCapabilities{
		OnGrab:                      true,
		OnDownload:                  true,
		OnUpgrade:                   true,
		OnRename:                    false,
		OnMovieAdded:                true,
		OnMovieDelete:               false,
		OnMovieFileDelete:           false,
		OnHealthIssue:               true,
		OnApplicationUpdate:         true,
		OnManualInteractionRequired: false,
		SupportsCustomTemplates:     false,
		SupportsRichContent:         false,
	}
}

// SupportsRetry returns false for stubs
func (p *StubProvider) SupportsRetry() bool {
	return false
}

// GetDefaultRetryConfig returns basic retry config
func (p *StubProvider) GetDefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     1,
		InitialDelay:   time.Second,
		MaxDelay:       time.Second,
		BackoffFactor:  1.0,
		RetryCondition: func(error) bool { return false },
	}
}
