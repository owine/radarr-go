// Package models defines data structures and database models for Radarr.
// This file contains the refactored Movie model with simplified validation.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// MovieV2 represents a movie in the refactored Radarr database
// This version eliminates problematic GORM hooks and complex relationships
type MovieV2 struct {
	ID            int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TmdbID        int    `json:"tmdbId" gorm:"uniqueIndex;not null"`
	ImdbID        string `json:"imdbId,omitempty" gorm:"size:20"`
	Title         string `json:"title" gorm:"not null;size:500"`
	TitleSlug     string `json:"titleSlug" gorm:"uniqueIndex;not null;size:500"`
	OriginalTitle string `json:"originalTitle,omitempty" gorm:"size:500"`
	Overview      string `json:"overview,omitempty" gorm:"type:text"`
	Year          int    `json:"year,omitempty"`
	Runtime       int    `json:"runtime,omitempty"`
	Status        string `json:"status" gorm:"default:'tba';size:20"`

	// File information (simplified)
	HasFile  bool   `json:"hasFile" gorm:"default:false"`
	FilePath string `json:"filePath,omitempty" gorm:"size:1000"`
	FileSize int64  `json:"fileSize" gorm:"default:0"`

	// Configuration
	Monitored           bool   `json:"monitored" gorm:"default:true"`
	QualityProfileID    int    `json:"qualityProfileId" gorm:"not null;default:1"`
	MinimumAvailability string `json:"minimumAvailability" gorm:"default:'announced';size:20"`

	// Collection relationship (simplified - just ID reference)
	CollectionID *int `json:"collectionId,omitempty"`

	// Metadata as JSON (eliminates complex custom types)
	Images  JSONField `json:"images" gorm:"type:json"`
	Genres  JSONField `json:"genres" gorm:"type:json"`
	Tags    JSONField `json:"tags" gorm:"type:json"`
	Ratings JSONField `json:"ratings" gorm:"type:json"`

	// Dates
	InCinemas       *time.Time `json:"inCinemas,omitempty"`
	PhysicalRelease *time.Time `json:"physicalRelease,omitempty"`
	DigitalRelease  *time.Time `json:"digitalRelease,omitempty"`
	Added           time.Time  `json:"added" gorm:"autoCreateTime"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// JSONField is a generic JSON field type that handles any JSON data
type JSONField map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database retrieval
func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONField)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// TableName returns the table name for GORM
func (MovieV2) TableName() string {
	return "movies"
}

// IsAvailable determines if the movie is available based on its status and dates
func (m *MovieV2) IsAvailable() bool {
	now := time.Now()

	switch m.MinimumAvailability {
	case "tba":
		return false
	case "announced":
		return m.Status != "tba"
	case "inCinemas":
		return m.InCinemas != nil && m.InCinemas.Before(now)
	case "released":
		return (m.PhysicalRelease != nil && m.PhysicalRelease.Before(now)) ||
			(m.DigitalRelease != nil && m.DigitalRelease.Before(now))
	default:
		return false
	}
}

// SetImages sets the images JSON field
func (m *MovieV2) SetImages(images []map[string]interface{}) {
	if m.Images == nil {
		m.Images = make(JSONField)
	}
	m.Images["images"] = images
}

// GetImages retrieves the images from the JSON field
func (m *MovieV2) GetImages() []map[string]interface{} {
	if m.Images == nil {
		return []map[string]interface{}{}
	}
	if images, ok := m.Images["images"].([]map[string]interface{}); ok {
		return images
	}
	return []map[string]interface{}{}
}

// SetGenres sets the genres JSON field
func (m *MovieV2) SetGenres(genres []string) {
	if m.Genres == nil {
		m.Genres = make(JSONField)
	}
	m.Genres["genres"] = genres
}

// GetGenres retrieves the genres from the JSON field
func (m *MovieV2) GetGenres() []string {
	if m.Genres == nil {
		return []string{}
	}
	if genres, ok := m.Genres["genres"].([]string); ok {
		return genres
	}
	return []string{}
}

// Validate performs basic validation without GORM hooks
func (m *MovieV2) Validate() error {
	if m.TmdbID == 0 {
		return ErrInvalidTmdbID
	}
	if m.Title == "" {
		return ErrInvalidTitle
	}
	if m.TitleSlug == "" {
		return ErrInvalidTitleSlug
	}
	if m.QualityProfileID == 0 {
		m.QualityProfileID = 1 // Set default
	}
	return nil
}

// Custom validation errors
var (
	ErrInvalidTmdbID    = ValidationError{Field: "tmdb_id", Message: "TMDB ID is required"}
	ErrInvalidTitle     = ValidationError{Field: "title", Message: "Title is required"}
	ErrInvalidTitleSlug = ValidationError{Field: "title_slug", Message: "Title slug is required"}
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
