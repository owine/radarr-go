// Package models provides data structures for the Radarr application.
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
	// NotificationTypeDiscord represents Discord notifications
	NotificationTypeDiscord NotificationType = "discord"
	// NotificationTypeSlack represents Slack notifications
	NotificationTypeSlack NotificationType = "slack"
	// NotificationTypeEmail represents email notifications
	NotificationTypeEmail NotificationType = "email"
	// NotificationTypePushbullet represents Pushbullet notifications
	NotificationTypePushbullet NotificationType = "pushbullet"
	// NotificationTypePushover represents Pushover notifications
	NotificationTypePushover NotificationType = "pushover"
	// NotificationTypeWebhook represents generic webhook notifications
	NotificationTypeWebhook NotificationType = "webhook"
	// NotificationTypeTelegram represents Telegram notifications
	NotificationTypeTelegram NotificationType = "telegram"
	// NotificationTypePlex represents Plex notifications
	NotificationTypePlex NotificationType = "plex"
	// NotificationTypeEmby represents Emby notifications
	NotificationTypeEmby NotificationType = "emby"
	// NotificationTypeJellyfin represents Jellyfin notifications
	NotificationTypeJellyfin NotificationType = "jellyfin"
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

// Notification represents a notification provider configuration
type Notification struct {
	ID                    int                  `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                  string               `json:"name" gorm:"not null;size:255"`
	Type                  NotificationType     `json:"implementation" gorm:"not null;size:50"`
	Settings              NotificationSettings `json:"fields" gorm:"type:text"`
	Tags                  IntArray             `json:"tags" gorm:"type:text"`
	OnGrab                bool                 `json:"onGrab" gorm:"default:false"`
	OnDownload            bool                 `json:"onDownload" gorm:"default:false"`
	OnUpgrade             bool                 `json:"onUpgrade" gorm:"default:false"`
	OnRename              bool                 `json:"onRename" gorm:"default:false"`
	OnMovieDelete         bool                 `json:"onMovieDelete" gorm:"default:false"`
	OnMovieFileDelete     bool                 `json:"onMovieFileDelete" gorm:"default:false"`
	OnHealth              bool                 `json:"onHealth" gorm:"default:false"`
	OnApplicationUpdate   bool                 `json:"onApplicationUpdate" gorm:"default:false"`
	IncludeHealthWarnings bool                 `json:"includeHealthWarnings" gorm:"default:false"`
	Enable                bool                 `json:"enable" gorm:"default:true"`
	CreatedAt             time.Time            `json:"added" gorm:"autoCreateTime"`
	UpdatedAt             time.Time            `json:"updated" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the Notification model
func (Notification) TableName() string {
	return "notifications"
}

// IsEnabled returns true if the notification is enabled
func (n *Notification) IsEnabled() bool {
	return n.Enable
}

// ShouldNotifyFor returns true if this notification should be sent for the given trigger
func (n *Notification) ShouldNotifyFor(trigger NotificationTrigger) bool {
	if !n.IsEnabled() {
		return false
	}

	switch trigger {
	case NotificationTriggerOnGrab:
		return n.OnGrab
	case NotificationTriggerOnDownload:
		return n.OnDownload
	case NotificationTriggerOnUpgrade:
		return n.OnUpgrade
	case NotificationTriggerOnRename:
		return n.OnRename
	case NotificationTriggerOnMovieDelete:
		return n.OnMovieDelete
	case NotificationTriggerOnMovieFileDelete:
		return n.OnMovieFileDelete
	case NotificationTriggerOnHealth:
		return n.OnHealth
	case NotificationTriggerOnApplicationUpdate:
		return n.OnApplicationUpdate
	default:
		return false
	}
}

// NotificationMessage represents a notification message to be sent
type NotificationMessage struct {
	Type    NotificationTrigger    `json:"type"`
	Subject string                 `json:"subject"`
	Body    string                 `json:"body"`
	Movie   *Movie                 `json:"movie,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// NotificationTestResult represents the result of testing a notification
type NotificationTestResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"validationFailures"`
}

// NotificationHistory represents a sent notification
type NotificationHistory struct {
	ID             int                 `json:"id" gorm:"primaryKey;autoIncrement"`
	NotificationID int                 `json:"notificationId" gorm:"index"`
	Notification   *Notification       `json:"notification,omitempty" gorm:"foreignKey:NotificationID"`
	MovieID        *int                `json:"movieId,omitempty" gorm:"index"`
	Movie          *Movie              `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	Trigger        NotificationTrigger `json:"eventType" gorm:"not null;size:50"`
	Subject        string              `json:"subject" gorm:"size:500"`
	Body           string              `json:"message" gorm:"type:text"`
	Successful     bool                `json:"successful" gorm:"not null"`
	ErrorMessage   string              `json:"errorMessage,omitempty" gorm:"type:text"`
	SentAt         time.Time           `json:"date" gorm:"not null;index"`
}

// TableName returns the database table name for the NotificationHistory model
func (NotificationHistory) TableName() string {
	return "notification_history"
}

// HealthStatus represents the health status of the system
type HealthStatus string

const (
	// HealthStatusOK represents a healthy system
	HealthStatusOK HealthStatus = "ok"
	// HealthStatusWarning represents a system with warnings
	HealthStatusWarning HealthStatus = "warning"
	// HealthStatusError represents a system with errors
	HealthStatusError HealthStatus = "error"
)

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
	return hc.Status == HealthStatusOK
}
