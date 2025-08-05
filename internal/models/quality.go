// Package models provides data structures for the Radarr application.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// QualityLevel represents a quality level definition
type QualityLevel struct {
	ID            int     `json:"id" gorm:"primaryKey"`
	Title         string  `json:"title" gorm:"not null;size:255"`
	Weight        int     `json:"weight" gorm:"not null;default:1"`
	MinSize       float64 `json:"minSize" gorm:"default:0"`   // MB per minute
	MaxSize       float64 `json:"maxSize" gorm:"default:400"` // MB per minute
	PreferredSize float64 `json:"preferredSize,omitempty"`    // MB per minute
}

// TableName returns the database table name for the QualityLevel model
func (QualityLevel) TableName() string {
	return "quality_definitions"
}

// QualityProfileItem represents a quality within a profile
type QualityProfileItem struct {
	Quality *QualityLevel         `json:"quality,omitempty"`
	Items   []*QualityProfileItem `json:"items,omitempty"`
	Allowed bool                  `json:"allowed"`
	Name    string                `json:"name,omitempty"`
	ID      int                   `json:"id,omitempty"`
}

// QualityProfileItems represents a slice of quality profile items for JSON/DB storage
type QualityProfileItems []*QualityProfileItem

// Value implements the driver.Valuer interface for database storage
func (qpi QualityProfileItems) Value() (driver.Value, error) {
	if qpi == nil {
		return nil, nil
	}
	return json.Marshal(qpi)
}

// Scan implements the sql.Scanner interface for database retrieval
func (qpi *QualityProfileItems) Scan(value interface{}) error {
	if value == nil {
		*qpi = QualityProfileItems{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, qpi)
	case string:
		return json.Unmarshal([]byte(v), qpi)
	default:
		return fmt.Errorf("cannot scan %T into QualityProfileItems", value)
	}
}

// QualityProfile represents a quality profile configuration
type QualityProfile struct {
	ID                int                 `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string              `json:"name" gorm:"not null;unique;size:255"`
	Cutoff            int                 `json:"cutoff" gorm:"not null"`
	Items             QualityProfileItems `json:"items" gorm:"type:text"`
	Language          string              `json:"language" gorm:"default:'english';size:50"`
	UpgradeAllowed    bool                `json:"upgradeAllowed" gorm:"default:true"`
	MinFormatScore    int                 `json:"minFormatScore" gorm:"default:0"`
	CutoffFormatScore int                 `json:"cutoffFormatScore" gorm:"default:0"`
	FormatItems       CustomFormatItems   `json:"formatItems" gorm:"type:text"`
	CreatedAt         time.Time           `json:"added" gorm:"autoCreateTime"`
	UpdatedAt         time.Time           `json:"updated" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the QualityProfile model
func (QualityProfile) TableName() string {
	return "quality_profiles"
}

// IsUpgradeAllowed returns true if upgrades are allowed for this profile
func (qp *QualityProfile) IsUpgradeAllowed() bool {
	return qp.UpgradeAllowed
}

// GetAllowedQualities returns all allowed qualities in this profile
func (qp *QualityProfile) GetAllowedQualities() []*QualityLevel {
	var allowed []*QualityLevel

	var collectQualities func([]*QualityProfileItem)
	collectQualities = func(items []*QualityProfileItem) {
		for _, item := range items {
			if item.Allowed {
				if item.Quality != nil {
					allowed = append(allowed, item.Quality)
				}
				if item.Items != nil {
					collectQualities(item.Items)
				}
			}
		}
	}

	collectQualities(qp.Items)
	return allowed
}

// CustomFormat represents a custom quality format
type CustomFormat struct {
	ID                              int               `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                            string            `json:"name" gorm:"not null;unique;size:255"`
	IncludeCustomFormatWhenRenaming bool              `json:"includeCustomFormatWhenRenaming" gorm:"default:false"`
	Specifications                  CustomFormatSpecs `json:"specifications" gorm:"type:text"`
	CreatedAt                       time.Time         `json:"added" gorm:"autoCreateTime"`
	UpdatedAt                       time.Time         `json:"updated" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the CustomFormat model
func (CustomFormat) TableName() string {
	return "custom_formats"
}

// CustomFormatSpec represents a specification for a custom format
type CustomFormatSpec struct {
	Name           string                 `json:"name"`
	Implementation string                 `json:"implementation"`
	Negate         bool                   `json:"negate"`
	Required       bool                   `json:"required"`
	Fields         map[string]interface{} `json:"fields"`
}

// CustomFormatSpecs represents a slice of custom format specifications
type CustomFormatSpecs []*CustomFormatSpec

// Value implements the driver.Valuer interface for database storage
func (cfs CustomFormatSpecs) Value() (driver.Value, error) {
	if cfs == nil {
		return nil, nil
	}
	return json.Marshal(cfs)
}

// Scan implements the sql.Scanner interface for database retrieval
func (cfs *CustomFormatSpecs) Scan(value interface{}) error {
	if value == nil {
		*cfs = CustomFormatSpecs{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, cfs)
	case string:
		return json.Unmarshal([]byte(v), cfs)
	default:
		return fmt.Errorf("cannot scan %T into CustomFormatSpecs", value)
	}
}

// CustomFormatItem represents a custom format item in a quality profile
type CustomFormatItem struct {
	Format *CustomFormat `json:"format,omitempty"`
	Name   string        `json:"name,omitempty"`
	Score  int           `json:"score"`
}

// CustomFormatItems represents a slice of custom format items
type CustomFormatItems []*CustomFormatItem

// Value implements the driver.Valuer interface for database storage
func (cfi CustomFormatItems) Value() (driver.Value, error) {
	if cfi == nil {
		return nil, nil
	}
	return json.Marshal(cfi)
}

// Scan implements the sql.Scanner interface for database retrieval
func (cfi *CustomFormatItems) Scan(value interface{}) error {
	if value == nil {
		*cfi = CustomFormatItems{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, cfi)
	case string:
		return json.Unmarshal([]byte(v), cfi)
	default:
		return fmt.Errorf("cannot scan %T into CustomFormatItems", value)
	}
}

// DefaultQualityDefinitions returns the standard quality definitions
func DefaultQualityDefinitions() []*QualityLevel {
	return []*QualityLevel{
		{ID: 0, Title: "Unknown", Weight: 1, MinSize: 0, MaxSize: 199.9},
		{ID: 24, Title: "WORKPRINT", Weight: 2, MinSize: 0, MaxSize: 199.9},
		{ID: 25, Title: "CAM", Weight: 3, MinSize: 0, MaxSize: 199.9},
		{ID: 26, Title: "TELESYNC", Weight: 4, MinSize: 0, MaxSize: 199.9},
		{ID: 27, Title: "TELECINE", Weight: 5, MinSize: 0, MaxSize: 199.9},
		{ID: 29, Title: "REGIONAL", Weight: 6, MinSize: 0, MaxSize: 199.9},
		{ID: 28, Title: "DVDSCR", Weight: 7, MinSize: 0, MaxSize: 199.9},
		{ID: 1, Title: "SDTV", Weight: 8, MinSize: 0, MaxSize: 199.9},
		{ID: 2, Title: "DVD", Weight: 9, MinSize: 0, MaxSize: 199.9},
		{ID: 23, Title: "DVD-R", Weight: 10, MinSize: 0, MaxSize: 199.9},
		{ID: 8, Title: "WEBDL-480p", Weight: 11, MinSize: 0, MaxSize: 199.9},
		{ID: 12, Title: "WEBRip-480p", Weight: 12, MinSize: 0, MaxSize: 199.9},
		{ID: 20, Title: "Bluray-480p", Weight: 13, MinSize: 0, MaxSize: 199.9},
		{ID: 21, Title: "Bluray-576p", Weight: 14, MinSize: 0, MaxSize: 199.9},
		{ID: 4, Title: "HDTV-720p", Weight: 15, MinSize: 0.8, MaxSize: 137.3},
		{ID: 5, Title: "WEBDL-720p", Weight: 16, MinSize: 0.8, MaxSize: 137.3},
		{ID: 14, Title: "WEBRip-720p", Weight: 17, MinSize: 0.8, MaxSize: 137.3},
		{ID: 6, Title: "Bluray-720p", Weight: 18, MinSize: 4.3, MaxSize: 137.3},
		{ID: 9, Title: "HDTV-1080p", Weight: 19, MinSize: 2, MaxSize: 137.3},
		{ID: 3, Title: "WEBDL-1080p", Weight: 20, MinSize: 2, MaxSize: 137.3},
		{ID: 15, Title: "WEBRip-1080p", Weight: 21, MinSize: 2, MaxSize: 137.3},
		{ID: 7, Title: "Bluray-1080p", Weight: 22, MinSize: 4.3, MaxSize: 258.1},
		{ID: 30, Title: "Remux-1080p", Weight: 23, MinSize: 0, MaxSize: 0},
		{ID: 16, Title: "HDTV-2160p", Weight: 24, MinSize: 4.7, MaxSize: 199.9},
		{ID: 18, Title: "WEBDL-2160p", Weight: 25, MinSize: 4.7, MaxSize: 258.1},
		{ID: 17, Title: "WEBRip-2160p", Weight: 26, MinSize: 4.7, MaxSize: 258.1},
		{ID: 19, Title: "Bluray-2160p", Weight: 27, MinSize: 4.3, MaxSize: 258.1},
		{ID: 31, Title: "Remux-2160p", Weight: 28, MinSize: 0, MaxSize: 0},
	}
}

// BeforeCreate hook validates quality profile data before creation
func (qp *QualityProfile) BeforeCreate(_ *gorm.DB) error {
	if qp.Name == "" {
		return errors.New("quality profile name is required")
	}
	if qp.Cutoff == 0 {
		return errors.New("quality profile cutoff is required")
	}
	if len(qp.Items) == 0 {
		return errors.New("quality profile must have at least one quality item")
	}
	return nil
}

// BeforeUpdate hook validates quality profile data before updates
func (qp *QualityProfile) BeforeUpdate(_ *gorm.DB) error {
	if qp.Name == "" {
		return errors.New("quality profile name cannot be empty")
	}
	if qp.Cutoff == 0 {
		return errors.New("quality profile cutoff cannot be zero")
	}
	return nil
}

// BeforeCreate hook validates quality level data before creation
func (ql *QualityLevel) BeforeCreate(_ *gorm.DB) error {
	if ql.Title == "" {
		return errors.New("quality level title is required")
	}
	if ql.Weight <= 0 {
		ql.Weight = 1 // Set default weight
	}
	if ql.MaxSize < 0 {
		return errors.New("quality level max size cannot be negative")
	}
	if ql.MinSize < 0 {
		return errors.New("quality level min size cannot be negative")
	}
	if ql.MinSize > ql.MaxSize && ql.MaxSize > 0 {
		return errors.New("quality level min size cannot be greater than max size")
	}
	return nil
}

// BeforeUpdate hook validates quality level data before updates
func (ql *QualityLevel) BeforeUpdate(_ *gorm.DB) error {
	if ql.Title == "" {
		return errors.New("quality level title cannot be empty")
	}
	if ql.MaxSize < 0 {
		return errors.New("quality level max size cannot be negative")
	}
	if ql.MinSize < 0 {
		return errors.New("quality level min size cannot be negative")
	}
	if ql.MinSize > ql.MaxSize && ql.MaxSize > 0 {
		return errors.New("quality level min size cannot be greater than max size")
	}
	return nil
}
