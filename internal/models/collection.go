// Package models defines data structures and database models for Radarr.
package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// MovieCollection represents a collection of movies with shared metadata
type MovieCollection struct {
	ID                  int          `json:"id" db:"id" gorm:"primaryKey;autoIncrement"`
	Title               string       `json:"title" db:"title" gorm:"not null;size:500"`
	CleanTitle          string       `json:"cleanTitle" db:"clean_title" gorm:"size:500"`
	SortTitle           string       `json:"sortTitle" db:"sort_title" gorm:"size:500"`
	TmdbID              int          `json:"tmdbId" db:"tmdb_id" gorm:"uniqueIndex;not null"`
	Overview            string       `json:"overview" db:"overview" gorm:"type:text"`
	Monitored           bool         `json:"monitored" db:"monitored" gorm:"default:true"`
	QualityProfileID    int          `json:"qualityProfileId" db:"quality_profile_id" gorm:"not null;default:1"`
	RootFolderPath      string       `json:"rootFolderPath" db:"root_folder_path" gorm:"size:500"`
	SearchOnAdd         bool         `json:"searchOnAdd" db:"search_on_add" gorm:"default:false"`
	MinimumAvailability Availability `json:"minimumAvailability" db:"minimum_availability" gorm:"default:'announced'"`
	LastInfoSync        *time.Time   `json:"lastInfoSync,omitempty" db:"last_info_sync"`
	Images              MediaCover   `json:"images" db:"images" gorm:"type:text"`
	Tags                IntArray     `json:"tags" db:"tags" gorm:"type:text"`
	Added               time.Time    `json:"added" db:"added" gorm:"autoCreateTime"`

	// Statistics (computed fields)
	MovieCount          int `json:"movieCount" gorm:"-"`
	MonitoredMovieCount int `json:"monitoredMovieCount" gorm:"-"`
	AvailableMovieCount int `json:"availableMovieCount" gorm:"-"`

	// Relationships
	Movies         []Movie         `json:"movies,omitempty" gorm:"foreignKey:CollectionTmdbID;references:TmdbID"`
	QualityProfile *QualityProfile `json:"qualityProfile,omitempty" gorm:"foreignKey:QualityProfileID"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// CollectionStatistics represents statistics for a collection
type CollectionStatistics struct {
	MovieCount          int     `json:"movieCount"`
	MonitoredMovieCount int     `json:"monitoredMovieCount"`
	AvailableMovieCount int     `json:"availableMovieCount"`
	HasFile             int     `json:"hasFile"`
	SizeOnDisk          int64   `json:"sizeOnDisk"`
	PercentOfMovies     float64 `json:"percentOfMovies"`
}

// BeforeCreate hook validates collection data before creation
func (c *MovieCollection) BeforeCreate(_ *gorm.DB) error {
	if c.TmdbID == 0 {
		return errors.New("tmdb_id is required for collection creation")
	}
	if c.Title == "" {
		return errors.New("title is required for collection creation")
	}
	if c.QualityProfileID == 0 {
		c.QualityProfileID = 1 // Set default quality profile
	}
	return nil
}

// BeforeUpdate hook validates collection data before updates
func (c *MovieCollection) BeforeUpdate(_ *gorm.DB) error {
	if c.TmdbID == 0 {
		return errors.New("tmdb_id cannot be empty")
	}
	if c.Title == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}

// AfterFind hook processes collection data after retrieval
func (c *MovieCollection) AfterFind(db *gorm.DB) error {
	// Load statistics if movies are not preloaded
	if len(c.Movies) == 0 {
		var stats CollectionStatistics
		err := db.Model(&Movie{}).
			Select(
				"COUNT(*) as movie_count",
				"COUNT(CASE WHEN monitored = true THEN 1 END) as monitored_movie_count",
				"COUNT(CASE WHEN has_file = true THEN 1 END) as available_movie_count",
			).
			Where("collection_tmdb_id = ?", c.TmdbID).
			Scan(&stats).Error

		if err == nil {
			c.MovieCount = stats.MovieCount
			c.MonitoredMovieCount = stats.MonitoredMovieCount
			c.AvailableMovieCount = stats.AvailableMovieCount
		}
	} else {
		// Compute statistics from loaded movies
		c.MovieCount = len(c.Movies)
		for _, movie := range c.Movies {
			if movie.Monitored {
				c.MonitoredMovieCount++
			}
			if movie.HasFile {
				c.AvailableMovieCount++
			}
		}
	}

	return nil
}

// ApplyChanges updates collection settings from another collection instance
func (c *MovieCollection) ApplyChanges(other *MovieCollection) {
	c.TmdbID = other.TmdbID
	c.Monitored = other.Monitored
	c.SearchOnAdd = other.SearchOnAdd
	c.QualityProfileID = other.QualityProfileID
	c.MinimumAvailability = other.MinimumAvailability
	c.RootFolderPath = other.RootFolderPath
	c.Tags = other.Tags
}

// TableName returns the table name for GORM
func (MovieCollection) TableName() string {
	return "collections"
}
