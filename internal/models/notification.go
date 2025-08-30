package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// NotificationType represents the type of notification provider
type NotificationType string

const (
	// NotificationTypeEmail represents email notifications
	NotificationTypeEmail NotificationType = "Email"
	// NotificationTypeDiscord represents Discord webhook notifications
	NotificationTypeDiscord NotificationType = "Discord"
	// NotificationTypeSlack represents Slack webhook notifications
	NotificationTypeSlack NotificationType = "Slack"
	// NotificationTypeTelegram represents Telegram bot notifications
	NotificationTypeTelegram NotificationType = "Telegram"
	// NotificationTypeWebhook represents generic webhook notifications
	NotificationTypeWebhook NotificationType = "Webhook"

	// NotificationTypePushover represents Pushover push notifications
	NotificationTypePushover NotificationType = "Pushover"
	// NotificationTypePushbullet represents Pushbullet push notifications
	NotificationTypePushbullet NotificationType = "Pushbullet"
	// NotificationTypeGotify represents Gotify push notifications
	NotificationTypeGotify NotificationType = "Gotify"
	// NotificationTypeJoin represents Join push notifications
	NotificationTypeJoin NotificationType = "Join"
	// NotificationTypeApprise represents Apprise multi-service notifications
	NotificationTypeApprise NotificationType = "Apprise"
	// NotificationTypeNotifiarr represents Notifiarr service notifications
	NotificationTypeNotifiarr NotificationType = "Notifiarr"

	// NotificationTypeMailgun represents Mailgun email service
	NotificationTypeMailgun NotificationType = "Mailgun"
	// NotificationTypeSendGrid represents SendGrid email service
	NotificationTypeSendGrid NotificationType = "SendGrid"

	// NotificationTypePlex represents Plex Media Server notifications
	NotificationTypePlex NotificationType = "Plex"
	// NotificationTypeEmby represents Emby Media Server notifications
	NotificationTypeEmby NotificationType = "Emby"
	// NotificationTypeJellyfin represents Jellyfin Media Server notifications
	NotificationTypeJellyfin NotificationType = "Jellyfin"
	// NotificationTypeKodi represents Kodi media center notifications
	NotificationTypeKodi NotificationType = "Kodi"

	// NotificationTypeCustomScript represents custom script notifications
	NotificationTypeCustomScript NotificationType = "CustomScript"
	// NotificationTypeSynologyIndexer represents Synology indexer notifications
	NotificationTypeSynologyIndexer NotificationType = "SynologyIndexer"
	// NotificationTypeTwitter represents Twitter notifications
	NotificationTypeTwitter NotificationType = "Twitter"
	// NotificationTypeSignal represents Signal messenger notifications
	NotificationTypeSignal NotificationType = "Signal"
	// NotificationTypeMatrix represents Matrix protocol notifications
	NotificationTypeMatrix NotificationType = "Matrix"
	// NotificationTypeNtfy represents Ntfy push notifications
	NotificationTypeNtfy NotificationType = "Ntfy"
)

// NotificationTrigger represents when notifications should be sent
type NotificationTrigger string

const (
	// NotificationTriggerOnGrab when a movie is grabbed from an indexer
	NotificationTriggerOnGrab NotificationTrigger = "onGrab"
	// NotificationTriggerOnDownload when a movie file is downloaded and imported
	NotificationTriggerOnDownload NotificationTrigger = "onDownload"
	// NotificationTriggerOnUpgrade when a movie file is upgraded to better quality
	NotificationTriggerOnUpgrade NotificationTrigger = "onUpgrade"
	// NotificationTriggerOnRename when a movie file is renamed
	NotificationTriggerOnRename NotificationTrigger = "onRename"
	// NotificationTriggerOnMovieDelete when a movie is deleted
	NotificationTriggerOnMovieDelete NotificationTrigger = "onMovieDelete"
	// NotificationTriggerOnMovieFileDelete when a movie file is deleted
	NotificationTriggerOnMovieFileDelete NotificationTrigger = "onMovieFileDelete"
	// NotificationTriggerOnHealth when health issues are detected
	NotificationTriggerOnHealth NotificationTrigger = "onHealth"
	// NotificationTriggerOnApplicationUpdate when application updates are available
	NotificationTriggerOnApplicationUpdate NotificationTrigger = "onApplicationUpdate"
)

// NotificationSettings represents flexible configuration for different notification types
type NotificationSettings map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (ns NotificationSettings) Value() (driver.Value, error) {
	if ns == nil {
		return nil, nil
	}
	return json.Marshal(ns)
}

// Scan implements the sql.Scanner interface for database retrieval
func (ns *NotificationSettings) Scan(value interface{}) error {
	if value == nil {
		*ns = make(NotificationSettings)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ns)
	case string:
		return json.Unmarshal([]byte(v), ns)
	default:
		return fmt.Errorf("cannot scan %T into NotificationSettings", value)
	}
}

// NotificationField represents a configuration field for a notification
type NotificationField struct {
	Name          string         `json:"name"`
	Label         string         `json:"label"`
	Value         interface{}    `json:"value"`
	Type          string         `json:"type"`
	Advanced      bool           `json:"advanced"`
	Privacy       string         `json:"privacy"`
	SelectOptions []SelectOption `json:"selectOptions,omitempty"`
	HelpText      string         `json:"helpText,omitempty"`
	HelpLink      string         `json:"helpLink,omitempty"`
	Order         int            `json:"order"`
	Hidden        bool           `json:"hidden"`
}

// NotificationFieldsArray is a custom type for handling JSON arrays of notification fields
type NotificationFieldsArray []NotificationField

// Value implements the driver.Valuer interface for database storage
func (f NotificationFieldsArray) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements the sql.Scanner interface for database retrieval
func (f *NotificationFieldsArray) Scan(value interface{}) error {
	if value == nil {
		*f = NotificationFieldsArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, f)
}

// NotificationTestResult represents the result of testing a notification configuration
type NotificationTestResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"errors"`
}

// Notification represents a notification provider configuration
type Notification struct {
	ID             int                  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name           string               `json:"name" gorm:"not null;size:255;uniqueIndex"`
	Implementation NotificationType     `json:"implementation" gorm:"not null;size:50"`
	ConfigContract string               `json:"configContract" gorm:"size:100"`
	Settings       NotificationSettings `json:"settings" gorm:"type:text"`
	Tags           IntArray             `json:"tags" gorm:"type:text"`

	// Event triggers
	OnGrab                      bool `json:"onGrab" gorm:"default:false"`
	OnDownload                  bool `json:"onDownload" gorm:"default:true"`
	OnUpgrade                   bool `json:"onUpgrade" gorm:"default:true"`
	OnRename                    bool `json:"onRename" gorm:"default:false"`
	OnMovieAdded                bool `json:"onMovieAdded" gorm:"default:false"`
	OnMovieDelete               bool `json:"onMovieDelete" gorm:"default:false"`
	OnMovieFileDelete           bool `json:"onMovieFileDelete" gorm:"default:false"`
	OnHealthIssue               bool `json:"onHealthIssue" gorm:"default:false"`
	OnApplicationUpdate         bool `json:"onApplicationUpdate" gorm:"default:false"`
	OnManualInteractionRequired bool `json:"onManualInteractionRequired" gorm:"default:false"`
	IncludeHealthWarnings       bool `json:"includeHealthWarnings" gorm:"default:false"`

	// Provider capabilities
	SupportsOnGrab                      bool `json:"supportsOnGrab" gorm:"default:true"`
	SupportsOnDownload                  bool `json:"supportsOnDownload" gorm:"default:true"`
	SupportsOnUpgrade                   bool `json:"supportsOnUpgrade" gorm:"default:true"`
	SupportsOnRename                    bool `json:"supportsOnRename" gorm:"default:true"`
	SupportsOnMovieAdded                bool `json:"supportsOnMovieAdded" gorm:"default:true"`
	SupportsOnMovieDelete               bool `json:"supportsOnMovieDelete" gorm:"default:true"`
	SupportsOnMovieFileDelete           bool `json:"supportsOnMovieFileDelete" gorm:"default:true"`
	SupportsOnHealthIssue               bool `json:"supportsOnHealthIssue" gorm:"default:true"`
	SupportsOnApplicationUpdate         bool `json:"supportsOnApplicationUpdate" gorm:"default:true"`
	SupportsOnManualInteractionRequired bool `json:"supportsOnManualInteractionRequired" gorm:"default:true"`

	Enabled   bool                    `json:"enabled" gorm:"default:true"`
	Fields    NotificationFieldsArray `json:"fields" gorm:"type:text"`
	CreatedAt time.Time               `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time               `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the Notification model
func (Notification) TableName() string {
	return "notifications"
}

// IsEnabled returns true if the notification is enabled
func (n *Notification) IsEnabled() bool {
	return n.Enabled
}

// ShouldNotifyFor returns true if this notification should be sent for the given trigger
func (n *Notification) ShouldNotifyFor(trigger NotificationTrigger) bool {
	if !n.IsEnabled() {
		return false
	}

	switch trigger {
	case NotificationTriggerOnGrab:
		return n.SupportsOnGrab && n.OnGrab
	case NotificationTriggerOnDownload:
		return n.SupportsOnDownload && n.OnDownload
	case NotificationTriggerOnUpgrade:
		return n.SupportsOnUpgrade && n.OnUpgrade
	case NotificationTriggerOnRename:
		return n.SupportsOnRename && n.OnRename
	case NotificationTriggerOnMovieDelete:
		return n.SupportsOnMovieDelete && n.OnMovieDelete
	case NotificationTriggerOnMovieFileDelete:
		return n.SupportsOnMovieFileDelete && n.OnMovieFileDelete
	case NotificationTriggerOnHealth:
		return n.SupportsOnHealthIssue && n.OnHealthIssue
	case NotificationTriggerOnApplicationUpdate:
		return n.SupportsOnApplicationUpdate && n.OnApplicationUpdate
	default:
		return false
	}
}

// RequiresAuthentication checks if the notification type requires authentication
func (n *Notification) RequiresAuthentication() bool {
	switch n.Implementation {
	case NotificationTypeEmail, NotificationTypeMailgun, NotificationTypeSendGrid:
		return true
	case NotificationTypeDiscord, NotificationTypeSlack, NotificationTypeWebhook:
		return false // Uses webhook URLs
	case NotificationTypeTelegram, NotificationTypePushover, NotificationTypePushbullet,
		NotificationTypeGotify, NotificationTypeJoin, NotificationTypeApprise,
		NotificationTypeNotifiarr, NotificationTypePlex, NotificationTypeEmby,
		NotificationTypeJellyfin, NotificationTypeKodi, NotificationTypeTwitter,
		NotificationTypeSignal, NotificationTypeMatrix, NotificationTypeNtfy:
		return true
	case NotificationTypeCustomScript, NotificationTypeSynologyIndexer:
		return false
	default:
		return false
	}
}

// notificationProviderNames maps notification types to display names
var notificationProviderNames = map[NotificationType]string{
	NotificationTypeEmail:           "Email",
	NotificationTypeDiscord:         "Discord",
	NotificationTypeSlack:           "Slack",
	NotificationTypeTelegram:        "Telegram",
	NotificationTypePushover:        "Pushover",
	NotificationTypePushbullet:      "Pushbullet",
	NotificationTypeGotify:          "Gotify",
	NotificationTypeWebhook:         "Webhook",
	NotificationTypeCustomScript:    "Custom Script",
	NotificationTypePlex:            "Plex Media Server",
	NotificationTypeEmby:            "Emby Media Server",
	NotificationTypeJellyfin:        "Jellyfin Media Server",
	NotificationTypeJoin:            "Join",
	NotificationTypeApprise:         "Apprise",
	NotificationTypeNotifiarr:       "Notifiarr",
	NotificationTypeMailgun:         "Mailgun",
	NotificationTypeSendGrid:        "SendGrid",
	NotificationTypeKodi:            "Kodi",
	NotificationTypeSynologyIndexer: "Synology Indexer",
	NotificationTypeTwitter:         "Twitter",
	NotificationTypeSignal:          "Signal",
	NotificationTypeMatrix:          "Matrix",
	NotificationTypeNtfy:            "Ntfy",
}

// GetProviderName returns the display name for the notification provider
func (n *Notification) GetProviderName() string {
	if name, exists := notificationProviderNames[n.Implementation]; exists {
		return name
	}
	return string(n.Implementation)
}

// NotificationMessage represents a notification message to be sent
type NotificationMessage struct {
	Type           NotificationTrigger    `json:"type"`
	Subject        string                 `json:"subject"`
	Body           string                 `json:"body"`
	Movie          *Movie                 `json:"movie,omitempty"`
	MovieFile      *MovieFile             `json:"movieFile,omitempty"`
	DeletedFiles   []MovieFile            `json:"deletedFiles,omitempty"`
	OldMovieFile   *MovieFile             `json:"oldMovieFile,omitempty"`
	SourceTitle    string                 `json:"sourceTitle,omitempty"`
	Quality        *QualityDefinition     `json:"quality,omitempty"`
	QualityUpgrade bool                   `json:"qualityUpgrade"`
	DownloadClient string                 `json:"downloadClient,omitempty"`
	DownloadID     string                 `json:"downloadId,omitempty"`
	Message        string                 `json:"message,omitempty"`
	HealthCheck    *HealthCheck           `json:"healthCheck,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
}

// NotificationEvent represents an event that can trigger notifications
type NotificationEvent struct {
	Type              string                 `json:"type"`
	EventType         string                 `json:"eventType"`
	Movie             *Movie                 `json:"movie,omitempty"`
	MovieFile         *MovieFile             `json:"movieFile,omitempty"`
	DeletedFiles      []MovieFile            `json:"deletedFiles,omitempty"`
	OldMovieFile      *MovieFile             `json:"oldMovieFile,omitempty"`
	SourceTitle       string                 `json:"sourceTitle,omitempty"`
	Quality           *QualityDefinition     `json:"quality,omitempty"`
	QualityUpgrade    bool                   `json:"qualityUpgrade"`
	DownloadClient    string                 `json:"downloadClient,omitempty"`
	DownloadID        string                 `json:"downloadId,omitempty"`
	Message           string                 `json:"message,omitempty"`
	HealthCheck       *HealthCheck           `json:"healthCheck,omitempty"`
	ApplicationUpdate *ApplicationUpdate     `json:"updateChanges,omitempty"`
	ManualInteraction *ManualInteraction     `json:"manualInteraction,omitempty"`
	Data              map[string]interface{} `json:"data,omitempty"`
}

// ApplicationUpdate represents an application update
type ApplicationUpdate struct {
	Version     string    `json:"version"`
	Branch      string    `json:"branch"`
	ReleaseDate time.Time `json:"releaseDate"`
	FileName    string    `json:"fileName"`
	URL         string    `json:"url"`
	Changes     []string  `json:"changes"`
}

// ManualInteraction represents a manual interaction requirement
type ManualInteraction struct {
	DownloadID string `json:"downloadId"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

// NotificationHistory represents a sent notification
type NotificationHistory struct {
	ID             int           `json:"id" gorm:"primaryKey;autoIncrement"`
	NotificationID int           `json:"notificationId" gorm:"index"`
	Notification   *Notification `json:"notification,omitempty" gorm:"foreignKey:NotificationID"`
	MovieID        *int          `json:"movieId,omitempty" gorm:"index"`
	Movie          *Movie        `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	EventType      string        `json:"eventType" gorm:"not null;size:50"`
	Subject        string        `json:"subject" gorm:"size:500"`
	Body           string        `json:"message" gorm:"type:text"`
	Successful     bool          `json:"successful" gorm:"not null"`
	ErrorMessage   string        `json:"errorMessage,omitempty" gorm:"type:text"`
	SentAt         time.Time     `json:"date" gorm:"not null;index"`
	CreatedAt      time.Time     `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the database table name for the NotificationHistory model
func (NotificationHistory) TableName() string {
	return "notification_history"
}

// HealthCheck represents a system health check
type HealthCheck struct {
	ID        int          `json:"id" gorm:"primaryKey;autoIncrement"`
	Source    string       `json:"source" gorm:"not null;size:100"`
	Type      string       `json:"type" gorm:"not null;size:50"`
	Message   string       `json:"message" gorm:"not null;size:1000"`
	WikiURL   string       `json:"wikiUrl" gorm:"size:500"`
	Status    HealthStatus `json:"status" gorm:"not null;size:20"`
	CreatedAt time.Time    `json:"time" gorm:"autoCreateTime"`
}

// TableName returns the database table name for the HealthCheck model
func (HealthCheck) TableName() string {
	return "health_checks"
}

// IsWarning returns true if this health check is a warning
func (hc *HealthCheck) IsWarning() bool {
	return hc.Status == HealthStatusWarning
}

// IsError returns true if this health check is an error
func (hc *HealthCheck) IsError() bool {
	return hc.Status == HealthStatusError
}

// IsOK returns true if this health check is OK
func (hc *HealthCheck) IsOK() bool {
	return hc.Status == HealthStatusHealthy
}
