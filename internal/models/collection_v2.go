// Package models defines data structures and database models for Radarr.
// This file contains the refactored MovieCollection model with simplified validation.
package models

import (
	"time"
)

// MovieCollectionV2 represents a collection of movies with simplified structure
// This version eliminates problematic GORM hooks and complex relationships
type MovieCollectionV2 struct {
	ID       int    `json:"id" gorm:"primaryKey;autoIncrement"`
	TmdbID   int    `json:"tmdbId" gorm:"uniqueIndex;not null"`
	Title    string `json:"title" gorm:"not null;size:500"`
	Overview string `json:"overview,omitempty" gorm:"type:text"`

	// Configuration
	Monitored           bool   `json:"monitored" gorm:"default:true"`
	QualityProfileID    int    `json:"qualityProfileId" gorm:"not null;default:1"`
	MinimumAvailability string `json:"minimumAvailability" gorm:"default:'announced';size:20"`

	// Metadata as JSON (simplified)
	Images JSONField `json:"images" gorm:"type:json"`
	Tags   JSONField `json:"tags" gorm:"type:json"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	// Computed fields (not stored in database)
	MovieCount          int `json:"movieCount" gorm:"-"`
	MonitoredMovieCount int `json:"monitoredMovieCount" gorm:"-"`
	AvailableMovieCount int `json:"availableMovieCount" gorm:"-"`
}

// TableName returns the table name for GORM
func (MovieCollectionV2) TableName() string {
	return "collections"
}

// Validate performs basic validation without GORM hooks
func (c *MovieCollectionV2) Validate() error {
	if c.TmdbID == 0 {
		return ValidationError{Field: "tmdb_id", Message: "TMDB ID is required"}
	}
	if c.Title == "" {
		return ValidationError{Field: "title", Message: "Title is required"}
	}
	if c.QualityProfileID == 0 {
		c.QualityProfileID = 1 // Set default
	}
	return nil
}

// SetImages sets the images JSON field
func (c *MovieCollectionV2) SetImages(images []map[string]interface{}) {
	if c.Images == nil {
		c.Images = make(JSONField)
	}
	c.Images["images"] = images
}

// GetImages retrieves the images from the JSON field
func (c *MovieCollectionV2) GetImages() []map[string]interface{} {
	if c.Images == nil {
		return []map[string]interface{}{}
	}
	if images, ok := c.Images["images"].([]map[string]interface{}); ok {
		return images
	}
	return []map[string]interface{}{}
}

// SetTags sets the tags JSON field
func (c *MovieCollectionV2) SetTags(tags []int) {
	if c.Tags == nil {
		c.Tags = make(JSONField)
	}
	c.Tags["tags"] = tags
}

// GetTags retrieves the tags from the JSON field
func (c *MovieCollectionV2) GetTags() []int {
	if c.Tags == nil {
		return []int{}
	}
	if tags, ok := c.Tags["tags"].([]int); ok {
		return tags
	}
	return []int{}
}

// ApplyChanges updates collection settings from another collection instance
func (c *MovieCollectionV2) ApplyChanges(other *MovieCollectionV2) {
	c.TmdbID = other.TmdbID
	c.Monitored = other.Monitored
	c.QualityProfileID = other.QualityProfileID
	c.MinimumAvailability = other.MinimumAvailability
	c.Images = other.Images
	c.Tags = other.Tags
}

// ToV1 converts MovieCollectionV2 to MovieCollection (for API compatibility)
func (c *MovieCollectionV2) ToV1() *MovieCollection {
	v1 := &MovieCollection{
		ID:                  c.ID,
		Title:               c.Title,
		CleanTitle:          "", // Not stored in V2, would need to be computed
		SortTitle:           "", // Not stored in V2, would need to be computed
		TmdbID:              c.TmdbID,
		Overview:            c.Overview,
		Monitored:           c.Monitored,
		QualityProfileID:    c.QualityProfileID,
		RootFolderPath:      "",    // Not stored in V2
		SearchOnAdd:         false, // Default value
		MinimumAvailability: Availability(c.MinimumAvailability),
		LastInfoSync:        nil, // Not tracked in V2
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
		MovieCount:          c.MovieCount,
		MonitoredMovieCount: c.MonitoredMovieCount,
		AvailableMovieCount: c.AvailableMovieCount,
	}

	// Convert images from JSON to MediaCover if available
	if c.Images != nil {
		if imagesData, ok := c.Images["images"].([]interface{}); ok {
			images := make([]MediaCoverImage, 0, len(imagesData))
			for _, img := range imagesData {
				if imgMap, ok := img.(map[string]interface{}); ok {
					image := MediaCoverImage{}
					if coverType, ok := imgMap["coverType"].(string); ok {
						image.CoverType = coverType
					}
					if url, ok := imgMap["url"].(string); ok {
						image.URL = url
					}
					if remoteURL, ok := imgMap["remoteUrl"].(string); ok {
						image.RemoteURL = remoteURL
					}
					images = append(images, image)
				}
			}
			v1.Images = images
		}
	}

	// Convert tags from JSON to IntArray if available
	if c.Tags != nil {
		if tagsData, ok := c.Tags["tags"].([]interface{}); ok {
			tags := make(IntArray, 0, len(tagsData))
			for _, tag := range tagsData {
				if tagInt, ok := tag.(int); ok {
					tags = append(tags, tagInt)
				} else if tagFloat, ok := tag.(float64); ok {
					tags = append(tags, int(tagFloat))
				}
			}
			v1.Tags = tags
		}
	}

	return v1
}

// FromV1 converts MovieCollection to MovieCollectionV2
func (c *MovieCollectionV2) FromV1(v1 *MovieCollection) {
	c.ID = v1.ID
	c.Title = v1.Title
	c.TmdbID = v1.TmdbID
	c.Overview = v1.Overview
	c.Monitored = v1.Monitored
	c.QualityProfileID = v1.QualityProfileID
	c.MinimumAvailability = string(v1.MinimumAvailability)
	c.CreatedAt = v1.CreatedAt
	c.UpdatedAt = v1.UpdatedAt
	c.MovieCount = v1.MovieCount
	c.MonitoredMovieCount = v1.MonitoredMovieCount
	c.AvailableMovieCount = v1.AvailableMovieCount

	// Convert images from MediaCover to JSON
	if len(v1.Images) > 0 {
		if c.Images == nil {
			c.Images = make(JSONField)
		}
		imagesData := make([]map[string]interface{}, 0, len(v1.Images))
		for _, img := range v1.Images {
			imgMap := map[string]interface{}{
				"coverType": img.CoverType,
				"url":       img.URL,
				"remoteUrl": img.RemoteURL,
			}
			imagesData = append(imagesData, imgMap)
		}
		c.Images["images"] = imagesData
	}

	// Convert tags from IntArray to JSON
	if len(v1.Tags) > 0 {
		if c.Tags == nil {
			c.Tags = make(JSONField)
		}
		tagsData := make([]int, 0, len(v1.Tags))
		for _, tag := range v1.Tags {
			tagsData = append(tagsData, tag)
		}
		c.Tags["tags"] = tagsData
	}
}

// NewMovieCollectionV2FromV1 creates a new MovieCollectionV2 from MovieCollection
func NewMovieCollectionV2FromV1(v1 *MovieCollection) *MovieCollectionV2 {
	v2 := &MovieCollectionV2{}
	v2.FromV1(v1)
	return v2
}

// CollectionStatisticsV2 represents statistics for a collection (simplified)
type CollectionStatisticsV2 struct {
	MovieCount          int     `json:"movieCount"`
	MonitoredMovieCount int     `json:"monitoredMovieCount"`
	AvailableMovieCount int     `json:"availableMovieCount"`
	HasFile             int     `json:"hasFile"`
	SizeOnDisk          int64   `json:"sizeOnDisk"`
	PercentOfMovies     float64 `json:"percentOfMovies"`
}
