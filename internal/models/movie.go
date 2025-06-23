package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Movie struct {
	ID                    int           `json:"id" db:"id" gorm:"primaryKey"`
	Title                 string        `json:"title" db:"title" gorm:"not null"`
	OriginalTitle         string        `json:"originalTitle" db:"original_title"`
	OriginalLanguage      Language      `json:"originalLanguage" db:"original_language" gorm:"type:text"`
	AlternateTitles       StringArray   `json:"alternateTitles" db:"alternate_titles" gorm:"type:text"`
	SecondaryYear         *int          `json:"secondaryYear,omitempty" db:"secondary_year"`
	SecondaryYearSourceID int           `json:"secondaryYearSourceId" db:"secondary_year_source_id"`
	SortTitle             string        `json:"sortTitle" db:"sort_title"`
	SizeOnDisk            int64         `json:"sizeOnDisk" db:"size_on_disk"`
	Status                MovieStatus   `json:"status" db:"status"`
	Overview              string        `json:"overview" db:"overview" gorm:"type:text"`
	InCinemas             *time.Time    `json:"inCinemas,omitempty" db:"in_cinemas"`
	PhysicalRelease       *time.Time    `json:"physicalRelease,omitempty" db:"physical_release"`
	DigitalRelease        *time.Time    `json:"digitalRelease,omitempty" db:"digital_release"`
	PhysicalReleaseNote   string        `json:"physicalReleaseNote" db:"physical_release_note"`
	Images                MediaCover    `json:"images" db:"images" gorm:"type:text"`
	Website               string        `json:"website" db:"website"`
	Year                  int           `json:"year" db:"year"`
	YouTubeTrailerID      string        `json:"youTubeTrailerId" db:"youtube_trailer_id"`
	Studio                string        `json:"studio" db:"studio"`
	Path                  string        `json:"path" db:"path"`
	QualityProfileID      int           `json:"qualityProfileId" db:"quality_profile_id"`
	HasFile               bool          `json:"hasFile" db:"has_file"`
	MovieFileID           int           `json:"movieFileId" db:"movie_file_id"`
	Monitored             bool          `json:"monitored" db:"monitored"`
	MinimumAvailability   Availability  `json:"minimumAvailability" db:"minimum_availability"`
	IsAvailable           bool          `json:"isAvailable" db:"is_available"`
	FolderName            string        `json:"folderName" db:"folder_name"`
	Runtime               int           `json:"runtime" db:"runtime"`
	CleanTitle            string        `json:"cleanTitle" db:"clean_title"`
	ImdbID                string        `json:"imdbId" db:"imdb_id"`
	TmdbID                int           `json:"tmdbId" db:"tmdb_id" gorm:"uniqueIndex"`
	TitleSlug             string        `json:"titleSlug" db:"title_slug" gorm:"uniqueIndex"`
	RootFolderPath        string        `json:"rootFolderPath" db:"root_folder_path"`
	Certification         string        `json:"certification" db:"certification"`
	Genres                StringArray   `json:"genres" db:"genres" gorm:"type:text"`
	Tags                  IntArray      `json:"tags" db:"tags" gorm:"type:text"`
	Added                 time.Time     `json:"added" db:"added"`
	AddOptions            AddOptions    `json:"addOptions" db:"add_options" gorm:"type:text"`
	Ratings               Ratings       `json:"ratings" db:"ratings" gorm:"type:text"`
	MovieFile             *MovieFile    `json:"movieFile,omitempty" gorm:"foreignKey:MovieFileID"`
	Collection            *Collection   `json:"collection,omitempty" db:"collection" gorm:"type:text"`
	Popularity            float64       `json:"popularity" db:"popularity"`
	
	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

type MovieStatus string

const (
	MovieStatusTBA       MovieStatus = "tba"
	MovieStatusAnnounced MovieStatus = "announced"
	MovieStatusInCinemas MovieStatus = "inCinemas"
	MovieStatusReleased  MovieStatus = "released"
	MovieStatusDeleted   MovieStatus = "deleted"
)

type Availability string

const (
	AvailabilityTBA          Availability = "tba"
	AvailabilityAnnounced    Availability = "announced"
	AvailabilityInCinemas    Availability = "inCinemas"
	AvailabilityReleased     Availability = "released"
	AvailabilityPreDB        Availability = "preDB"
)

type Language struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

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

type IntArray []int

func (i IntArray) Value() (driver.Value, error) {
	return json.Marshal(i)
}

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

type MediaCover []MediaCoverImage

func (m MediaCover) Value() (driver.Value, error) {
	return json.Marshal(m)
}

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

type MediaCoverImage struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
	RemoteURL string `json:"remoteUrl"`
}

type AddOptions struct {
	IgnoreEpisodesWithFiles      bool `json:"ignoreEpisodesWithFiles"`
	IgnoreEpisodesWithoutFiles   bool `json:"ignoreEpisodesWithoutFiles"`
	Monitor                      bool `json:"monitor"`
	SearchForMovie               bool `json:"searchForMovie"`
	AddMethod                    string `json:"addMethod"`
}

func (a AddOptions) Value() (driver.Value, error) {
	return json.Marshal(a)
}

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

type Ratings struct {
	Imdb           Rating `json:"imdb"`
	Tmdb           Rating `json:"tmdb"`
	Metacritic     Rating `json:"metacritic"`
	RottenTomatoes Rating `json:"rottenTomatoes"`
}

func (r Ratings) Value() (driver.Value, error) {
	return json.Marshal(r)
}

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

type Rating struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

type Collection struct {
	Name   string                `json:"name"`
	TmdbID int                   `json:"tmdbId"`
	Images []MediaCoverImage     `json:"images"`
}

func (c Collection) Value() (driver.Value, error) {
	return json.Marshal(c)
}

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