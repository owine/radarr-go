// Package models defines data structures and database models for Radarr.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Movie represents a movie in the Radarr database
type Movie struct {
	ID                    int          `json:"id" db:"id" gorm:"primaryKey"`
	Title                 string       `json:"title" db:"title" gorm:"not null"`
	OriginalTitle         string       `json:"originalTitle" db:"original_title"`
	OriginalLanguage      Language     `json:"originalLanguage" db:"original_language" gorm:"type:text"`
	AlternateTitles       StringArray  `json:"alternateTitles" db:"alternate_titles" gorm:"type:text"`
	SecondaryYear         *int         `json:"secondaryYear,omitempty" db:"secondary_year"`
	SecondaryYearSourceID int          `json:"secondaryYearSourceId" db:"secondary_year_source_id"`
	SortTitle             string       `json:"sortTitle" db:"sort_title"`
	SizeOnDisk            int64        `json:"sizeOnDisk" db:"size_on_disk"`
	Status                MovieStatus  `json:"status" db:"status"`
	Overview              string       `json:"overview" db:"overview" gorm:"type:text"`
	InCinemas             *time.Time   `json:"inCinemas,omitempty" db:"in_cinemas"`
	PhysicalRelease       *time.Time   `json:"physicalRelease,omitempty" db:"physical_release"`
	DigitalRelease        *time.Time   `json:"digitalRelease,omitempty" db:"digital_release"`
	PhysicalReleaseNote   string       `json:"physicalReleaseNote" db:"physical_release_note"`
	Images                MediaCover   `json:"images" db:"images" gorm:"type:text"`
	Website               string       `json:"website" db:"website"`
	Year                  int          `json:"year" db:"year"`
	YouTubeTrailerID      string       `json:"youTubeTrailerId" db:"youtube_trailer_id"`
	Studio                string       `json:"studio" db:"studio"`
	Path                  string       `json:"path" db:"path"`
	QualityProfileID      int          `json:"qualityProfileId" db:"quality_profile_id"`
	HasFile               bool         `json:"hasFile" db:"has_file"`
	MovieFileID           int          `json:"movieFileId" db:"movie_file_id"`
	Monitored             bool         `json:"monitored" db:"monitored"`
	MinimumAvailability   Availability `json:"minimumAvailability" db:"minimum_availability"`
	IsAvailable           bool         `json:"isAvailable" db:"is_available"`
	FolderName            string       `json:"folderName" db:"folder_name"`
	Runtime               int          `json:"runtime" db:"runtime"`
	CleanTitle            string       `json:"cleanTitle" db:"clean_title"`
	ImdbID                string       `json:"imdbId" db:"imdb_id"`
	TmdbID                int          `json:"tmdbId" db:"tmdb_id" gorm:"uniqueIndex"`
	TitleSlug             string       `json:"titleSlug" db:"title_slug" gorm:"uniqueIndex"`
	RootFolderPath        string       `json:"rootFolderPath" db:"root_folder_path"`
	Certification         string       `json:"certification" db:"certification"`
	Genres                StringArray  `json:"genres" db:"genres" gorm:"type:text"`
	Tags                  IntArray     `json:"tags" db:"tags" gorm:"type:text"`
	Added                 time.Time    `json:"added" db:"added"`
	AddOptions            AddOptions   `json:"addOptions" db:"add_options" gorm:"type:text"`
	Ratings               Ratings      `json:"ratings" db:"ratings" gorm:"type:text"`
	MovieFile             *MovieFile   `json:"movieFile,omitempty" gorm:"foreignKey:MovieFileID"`
	Collection            *Collection  `json:"collection,omitempty" db:"collection" gorm:"type:text"`
	Popularity            float64      `json:"popularity" db:"popularity"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// MovieStatus represents the current status of a movie
type MovieStatus string

const (
	// MovieStatusTBA indicates the movie has no release date announced
	MovieStatusTBA MovieStatus = "tba"
	// MovieStatusAnnounced indicates the movie has been announced but not yet in cinemas
	MovieStatusAnnounced MovieStatus = "announced"
	// MovieStatusInCinemas indicates the movie is currently in theaters
	MovieStatusInCinemas MovieStatus = "inCinemas"
	// MovieStatusReleased indicates the movie has been released for home viewing
	MovieStatusReleased MovieStatus = "released"
	// MovieStatusDeleted indicates the movie has been removed from the collection
	MovieStatusDeleted MovieStatus = "deleted"
)

// Availability represents when a movie becomes available for download
type Availability string

const (
	// AvailabilityTBA indicates the movie availability is to be announced
	AvailabilityTBA Availability = "tba"
	// AvailabilityAnnounced indicates the movie is announced but not yet available
	AvailabilityAnnounced Availability = "announced"
	// AvailabilityInCinemas indicates the movie becomes available when in cinemas
	AvailabilityInCinemas Availability = "inCinemas"
	// AvailabilityReleased indicates the movie becomes available when officially released
	AvailabilityReleased Availability = "released"
	// AvailabilityPreDB indicates the movie becomes available before official database listing
	AvailabilityPreDB Availability = "preDB"
)

// Language represents a movie's language information
type Language struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// StringArray is a custom type for handling JSON arrays of strings in the database
type StringArray []string

// Value implements the driver.Valuer interface for database storage
func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = StringArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// IntArray is a custom type for handling JSON arrays of integers in the database
type IntArray []int

// Value implements the driver.Valuer interface for database storage
func (i IntArray) Value() (driver.Value, error) {
	return json.Marshal(i)
}

// Scan implements the sql.Scanner interface for database retrieval
func (i *IntArray) Scan(value interface{}) error {
	if value == nil {
		*i = IntArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, i)
}

// MediaCover represents a collection of cover images for a movie
type MediaCover []MediaCoverImage

// Value implements the driver.Valuer interface for database storage
func (m MediaCover) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for database retrieval
func (m *MediaCover) Scan(value interface{}) error {
	if value == nil {
		*m = MediaCover{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, m)
}

// MediaCoverImage represents a single cover image with its metadata
type MediaCoverImage struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}

// AddOptions contains options for adding a new movie to the collection
type AddOptions struct {
	IgnoreEpisodesWithFiles    bool   `json:"ignoreEpisodesWithFiles"`
	IgnoreEpisodesWithoutFiles bool   `json:"ignoreEpisodesWithoutFiles"`
	Monitor                    bool   `json:"monitor"`
	SearchForMovie             bool   `json:"searchForMovie"`
	AddMethod                  string `json:"addMethod"`
}

// Value implements the driver.Valuer interface for database storage
func (a AddOptions) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for database retrieval
func (a *AddOptions) Scan(value interface{}) error {
	if value == nil {
		*a = AddOptions{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// Ratings contains rating information from various sources
type Ratings struct {
	Imdb           Rating `json:"imdb"`
	Tmdb           Rating `json:"tmdb"`
	Metacritic     Rating `json:"metacritic"`
	RottenTomatoes Rating `json:"rottenTomatoes"`
}

// Value implements the driver.Valuer interface for database storage
func (r *Ratings) Value() (driver.Value, error) {
	return json.Marshal(r)
}

// Scan implements the sql.Scanner interface for database retrieval
func (r *Ratings) Scan(value interface{}) error {
	if value == nil {
		*r = Ratings{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, r)
}

// Rating represents a single rating from a specific source
type Rating struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// Collection represents a movie collection or series
type Collection struct {
	Name   string            `json:"name"`
	TmdbID int               `json:"tmdbId"`
	Images []MediaCoverImage `json:"images"`
}

// Value implements the driver.Valuer interface for database storage
func (c Collection) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database retrieval
func (c *Collection) Scan(value interface{}) error {
	if value == nil {
		*c = Collection{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// BeforeCreate hook validates movie data before creation
func (m *Movie) BeforeCreate(_ *gorm.DB) error {
	if m.TmdbID == 0 {
		return errors.New("tmdb_id is required for movie creation")
	}
	if m.Title == "" {
		return errors.New("title is required for movie creation")
	}
	if m.TitleSlug == "" {
		return errors.New("title_slug is required for movie creation")
	}
	if m.QualityProfileID == 0 {
		m.QualityProfileID = 1 // Set default quality profile
	}
	return nil
}

// BeforeUpdate hook validates movie data before updates
func (m *Movie) BeforeUpdate(_ *gorm.DB) error {
	if m.TmdbID == 0 {
		return errors.New("tmdb_id cannot be empty")
	}
	if m.Title == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}

// AfterFind hook processes movie data after retrieval
func (m *Movie) AfterFind(_ *gorm.DB) error {
	// Set computed fields
	m.IsAvailable = m.computeAvailability()
	return nil
}

// computeAvailability determines if the movie is available based on its status and dates
func (m *Movie) computeAvailability() bool {
	now := time.Now()

	switch m.MinimumAvailability {
	case AvailabilityTBA:
		return false
	case AvailabilityAnnounced:
		return m.Status != MovieStatusTBA
	case AvailabilityInCinemas:
		return m.InCinemas != nil && m.InCinemas.Before(now)
	case AvailabilityReleased:
		return (m.PhysicalRelease != nil && m.PhysicalRelease.Before(now)) ||
			(m.DigitalRelease != nil && m.DigitalRelease.Before(now))
	case AvailabilityPreDB:
		return m.Status == MovieStatusReleased
	default:
		return false
	}
}
