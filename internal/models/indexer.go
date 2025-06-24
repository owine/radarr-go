// Package models provides data structures for the Radarr application.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// IndexerType represents the type of indexer
type IndexerType string

const (
	// IndexerTypeTorznab represents a Torznab-compatible indexer
	IndexerTypeTorznab IndexerType = "torznab"
	// IndexerTypeNewznab represents a Newznab-compatible indexer
	IndexerTypeNewznab IndexerType = "newznab"
	// IndexerTypeRSS represents an RSS-based indexer
	IndexerTypeRSS IndexerType = "rss"
)

// IndexerStatus represents the current status of an indexer
type IndexerStatus string

const (
	// IndexerStatusEnabled represents an active indexer
	IndexerStatusEnabled IndexerStatus = "enabled"
	// IndexerStatusDisabled represents a disabled indexer
	IndexerStatusDisabled IndexerStatus = "disabled"
	// IndexerStatusError represents an indexer with errors
	IndexerStatusError IndexerStatus = "error"
)

// IndexerSettings represents flexible configuration for different indexer types
type IndexerSettings map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (is IndexerSettings) Value() (driver.Value, error) {
	if is == nil {
		return nil, nil
	}
	return json.Marshal(is)
}

// Scan implements the sql.Scanner interface for database retrieval
func (is *IndexerSettings) Scan(value interface{}) error {
	if value == nil {
		*is = make(IndexerSettings)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, is)
	case string:
		return json.Unmarshal([]byte(v), is)
	default:
		return fmt.Errorf("cannot scan %T into IndexerSettings", value)
	}
}

// Indexer represents a movie indexer/search provider
type Indexer struct {
	ID                      int             `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                    string          `json:"name" gorm:"not null;size:255"`
	Type                    IndexerType     `json:"implementation" gorm:"not null;size:50"`
	BaseURL                 string          `json:"baseUrl" gorm:"not null;size:500"`
	APIKey                  string          `json:"apiKey" gorm:"size:255"`
	Username                string          `json:"username" gorm:"size:255"`
	Password                string          `json:"password" gorm:"size:255"`
	Categories              string          `json:"categories" gorm:"size:500"` // Comma-separated category IDs
	Priority                int             `json:"priority" gorm:"default:25"`
	Status                  IndexerStatus   `json:"enable" gorm:"default:'enabled'"`
	Settings                IndexerSettings `json:"fields" gorm:"type:text"`
	SupportsSearch          bool            `json:"supportsSearch" gorm:"default:true"`
	SupportsRSS             bool            `json:"supportsRss" gorm:"default:true"`
	DownloadClientID        *int            `json:"downloadClientId,omitempty"`
	CreatedAt               time.Time       `json:"added" gorm:"autoCreateTime"`
	UpdatedAt               time.Time       `json:"updated" gorm:"autoUpdateTime"`
	LastRSSSync             *time.Time      `json:"lastRssSync,omitempty"`
	EnableRSS               bool            `json:"enableRss" gorm:"default:true"`
	EnableAutomaticSearch   bool            `json:"enableAutomaticSearch" gorm:"default:true"`
	EnableInteractiveSearch bool            `json:"enableInteractiveSearch" gorm:"default:true"`
	SupportsRedirect        bool            `json:"supportsRedirect" gorm:"default:false"`
	Tags                    IntArray        `json:"tags" gorm:"type:text"`
}

// TableName returns the database table name for the Indexer model
func (Indexer) TableName() string {
	return "indexers"
}

// BeforeCreate hook to set default values
func (i *Indexer) BeforeCreate(_ *gorm.DB) error {
	if i.Settings == nil {
		i.Settings = make(IndexerSettings)
	}
	if i.Tags == nil {
		i.Tags = IntArray{}
	}
	return nil
}

// IsEnabled returns true if the indexer is enabled
func (i *Indexer) IsEnabled() bool {
	return i.Status == IndexerStatusEnabled
}

// CanSearch returns true if the indexer supports search functionality
func (i *Indexer) CanSearch() bool {
	return i.SupportsSearch && i.IsEnabled() && i.EnableAutomaticSearch
}

// CanInteractiveSearch returns true if the indexer supports interactive search
func (i *Indexer) CanInteractiveSearch() bool {
	return i.SupportsSearch && i.IsEnabled() && i.EnableInteractiveSearch
}

// CanRSS returns true if the indexer supports RSS functionality
func (i *Indexer) CanRSS() bool {
	return i.SupportsRSS && i.IsEnabled() && i.EnableRSS
}

// IndexerTestResult represents the result of testing an indexer connection
type IndexerTestResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"validationFailures"`
}

// IndexerCapabilities represents what an indexer supports
type IndexerCapabilities struct {
	SupportsSearch            bool     `json:"supportsSearch"`
	SupportsRSS               bool     `json:"supportsRss"`
	SupportsRedirect          bool     `json:"supportsRedirect"`
	SupportedSearchParameters []string `json:"supportedSearchParameters"`
	Categories                []int    `json:"categories"`
}

// IndexerStats represents statistics for an indexer
type IndexerStats struct {
	IndexerID           int        `json:"indexerId"`
	TotalQueries        int        `json:"totalQueries"`
	SuccessfulQueries   int        `json:"successfulQueries"`
	FailedQueries       int        `json:"failedQueries"`
	AverageResponseTime float64    `json:"averageResponseTime"`
	LastSuccess         *time.Time `json:"lastSuccess,omitempty"`
	LastFailure         *time.Time `json:"lastFailure,omitempty"`
	LastErrorMessage    string     `json:"lastErrorMessage,omitempty"`
}
