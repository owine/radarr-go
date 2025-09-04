package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// AppConfig represents application-wide configuration settings
type AppConfig struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Key         string    `json:"key" gorm:"not null;unique;size:255"`
	Value       JSON      `json:"value" gorm:"type:text"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the AppConfig model
func (AppConfig) TableName() string {
	return "app_config"
}

// AppSettings represents the consolidated application settings
type AppSettings struct {
	// Application Info
	Version     string `json:"version"`
	Initialized bool   `json:"initialized"`

	// UI Settings
	Theme           string `json:"theme"`
	Language        string `json:"language"`
	TimeFormat      string `json:"timeFormat"`
	DateFormat      string `json:"dateFormat"`
	EnableColorMode bool   `json:"enableColorImpairedMode"`

	// Security Settings
	APIKeyRequired        bool     `json:"apiKeyRequired"`
	EnableAutomaticBackup bool     `json:"enableAutomaticBackup"`
	BackupRetentionDays   int      `json:"backupRetentionDays"`
	AllowedHosts          []string `json:"allowedHosts"`

	// Performance Settings
	MaxConcurrentTasks    int  `json:"maxConcurrentTasks"`
	EnablePerformanceMode bool `json:"enablePerformanceMode"`
	CacheExpirationHours  int  `json:"cacheExpirationHours"`

	// Feature Toggles
	EnableAdvancedSettings     bool `json:"enableAdvancedSettings"`
	EnableExperimentalFeatures bool `json:"enableExperimentalFeatures"`
	EnableTelemetry            bool `json:"enableTelemetry"`

	// Maintenance Settings
	MaintenanceMode    bool      `json:"maintenanceMode"`
	MaintenanceMessage string    `json:"maintenanceMessage"`
	LastBackup         time.Time `json:"lastBackup"`
}

// GetDefaultAppSettings returns the default application settings
func GetDefaultAppSettings() *AppSettings {
	return &AppSettings{
		Version:     "1.0.0",
		Initialized: false,

		// UI Defaults
		Theme:           "dark",
		Language:        "en-US",
		TimeFormat:      "24h",
		DateFormat:      "MM/dd/yyyy",
		EnableColorMode: false,

		// Security Defaults
		APIKeyRequired:        false,
		EnableAutomaticBackup: true,
		BackupRetentionDays:   30,
		AllowedHosts:          []string{"*"},

		// Performance Defaults
		MaxConcurrentTasks:    5,
		EnablePerformanceMode: false,
		CacheExpirationHours:  24,

		// Feature Defaults
		EnableAdvancedSettings:     false,
		EnableExperimentalFeatures: false,
		EnableTelemetry:            true,

		// Maintenance Defaults
		MaintenanceMode:    false,
		MaintenanceMessage: "",
		LastBackup:         time.Time{},
	}
}

// ValidateSettings validates the application settings
func (as *AppSettings) ValidateSettings() []string {
	var errors []string

	// Validate backup retention
	if as.BackupRetentionDays < 1 || as.BackupRetentionDays > 365 {
		errors = append(errors, "Backup retention days must be between 1 and 365")
	}

	// Validate concurrent tasks
	if as.MaxConcurrentTasks < 1 || as.MaxConcurrentTasks > 50 {
		errors = append(errors, "Max concurrent tasks must be between 1 and 50")
	}

	// Validate cache expiration
	if as.CacheExpirationHours < 1 || as.CacheExpirationHours > 168 {
		errors = append(errors, "Cache expiration hours must be between 1 and 168 (1 week)")
	}

	// Validate theme
	validThemes := []string{"dark", "light", "auto"}
	validTheme := false
	for _, theme := range validThemes {
		if as.Theme == theme {
			validTheme = true
			break
		}
	}
	if !validTheme {
		errors = append(errors, "Theme must be one of: dark, light, auto")
	}

	// Validate time format
	validTimeFormats := []string{"12h", "24h"}
	validTimeFormat := false
	for _, format := range validTimeFormats {
		if as.TimeFormat == format {
			validTimeFormat = true
			break
		}
	}
	if !validTimeFormat {
		errors = append(errors, "Time format must be either 12h or 24h")
	}

	return errors
}

// ConfigurationBackup represents a backup of all configuration settings
type ConfigurationBackup struct {
	ID                    int                    `json:"id"`
	BackupName            string                 `json:"backupName"`
	Description           string                 `json:"description"`
	CreatedAt             time.Time              `json:"createdAt"`
	HostConfig            *HostConfig            `json:"hostConfig"`
	NamingConfig          *NamingConfig          `json:"namingConfig"`
	MediaManagementConfig *MediaManagementConfig `json:"mediaManagementConfig"`
	AppSettings           *AppSettings           `json:"appSettings"`
	QualityProfiles       []QualityProfile       `json:"qualityProfiles"`
	CustomFormats         []CustomFormat         `json:"customFormats"`
	RootFolders           []RootFolder           `json:"rootFolders"`
	Indexers              []Indexer              `json:"indexers"`
	DownloadClients       []DownloadClient       `json:"downloadClients"`
	ImportLists           []ImportList           `json:"importLists"`
	Notifications         []Notification         `json:"notifications"`
	BackupSize            int64                  `json:"backupSize"`
	ConfigurationCount    int                    `json:"configurationCount"`
}

// ConfigurationValidationResult represents validation results for configuration
type ConfigurationValidationResult struct {
	IsValid          bool                               `json:"isValid"`
	ValidationErrors map[string][]string                `json:"validationErrors"`
	Warnings         map[string][]string                `json:"warnings"`
	ComponentStatus  map[string]ConfigurationStatus     `json:"componentStatus"`
	TestResults      map[string]ConfigurationTestResult `json:"testResults"`
}

// ConfigurationStatus represents the status of a configuration component
type ConfigurationStatus string

const (
	// ConfigurationStatusOK indicates the configuration is valid
	ConfigurationStatusOK ConfigurationStatus = "ok"
	// ConfigurationStatusWarning indicates the configuration has warnings
	ConfigurationStatusWarning ConfigurationStatus = "warning"
	// ConfigurationStatusError indicates the configuration has errors
	ConfigurationStatusError ConfigurationStatus = "error"
	// ConfigurationStatusTesting indicates the configuration is being tested
	ConfigurationStatusTesting ConfigurationStatus = "testing"
)

// ConfigurationTestResult represents the result of testing a configuration
type ConfigurationTestResult struct {
	Success      bool      `json:"success"`
	Message      string    `json:"message"`
	ResponseTime int64     `json:"responseTime"` // in milliseconds
	TestedAt     time.Time `json:"testedAt"`
	Details      JSON      `json:"details"`
}

// JSON represents a JSON field that can be stored in the database
type JSON map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database retrieval
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
}

// ConfigurationExport represents exported configuration data
type ConfigurationExport struct {
	ExportedAt         time.Time `json:"exportedAt"`
	ApplicationVersion string    `json:"applicationVersion"`
	ExportVersion      string    `json:"exportVersion"`
	Data               JSON      `json:"data"`
	Checksum           string    `json:"checksum"`
}

// ConfigurationImportResult represents the result of importing configuration
type ConfigurationImportResult struct {
	Success             bool                `json:"success"`
	ImportedComponents  []string            `json:"importedComponents"`
	SkippedComponents   []string            `json:"skippedComponents"`
	ValidationErrors    map[string][]string `json:"validationErrors"`
	ConflictResolutions map[string]string   `json:"conflictResolutions"`
	ImportedAt          time.Time           `json:"importedAt"`
}
