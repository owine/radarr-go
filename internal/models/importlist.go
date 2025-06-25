package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// ImportList represents an import list configuration for automatic movie discovery
type ImportList struct {
	ID                  int                    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                string                 `json:"name" gorm:"not null;size:255;uniqueIndex"`
	Implementation      ImportListType         `json:"implementation" gorm:"not null;size:50"`
	ConfigContract      string                 `json:"configContract" gorm:"size:100"`
	Settings            ImportListSettings     `json:"settings" gorm:"type:text"`
	EnableAuto          bool                   `json:"enableAuto" gorm:"default:true"`
	Enabled             bool                   `json:"enabled" gorm:"default:true"`
	EnableInteractive   bool                   `json:"enableInteractiveSearch" gorm:"default:false"`
	ListType            ImportListSourceType   `json:"listType" gorm:"size:20;default:'program'"`
	ListOrder           int                    `json:"listOrder" gorm:"default:0"`
	MinRefreshInterval  time.Duration          `json:"minRefreshInterval" gorm:"default:1440"` // minutes
	QualityProfileID    int                    `json:"qualityProfileId" gorm:"not null"`
	RootFolderPath      string                 `json:"rootFolderPath" gorm:"not null;size:500"`
	ShouldMonitor       bool                   `json:"shouldMonitor" gorm:"default:true"`
	MinimumAvailability Availability           `json:"minimumAvailability" gorm:"size:20;default:'released'"`
	Tags                IntArray               `json:"tags" gorm:"type:text"`
	Fields              ImportListFieldsArray  `json:"fields" gorm:"type:text"`
	LastSync            *time.Time             `json:"lastSync,omitempty"`
	CreatedAt           time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt           time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ImportListType represents different types of import list implementations
type ImportListType string

// Import list implementation types for different providers and sources
const (
	ImportListTypeTMDBCollection ImportListType = "TMDbCollectionImport"
	ImportListTypeTMDBCompany    ImportListType = "TMDbCompanyImport"
	ImportListTypeTMDBKeyword    ImportListType = "TMDbKeywordImport"
	ImportListTypeTMDBList       ImportListType = "TMDbListImport"
	ImportListTypeTMDBPerson     ImportListType = "TMDbPersonImport"
	ImportListTypeTMDBPopular    ImportListType = "TMDbPopularImport"
	ImportListTypeTMDBUser       ImportListType = "TMDbUserImport"
	ImportListTypeTrakt          ImportListType = "TraktImport"
	ImportListTypeTraktList      ImportListType = "TraktListImport"
	ImportListTypeTraktPopular   ImportListType = "TraktPopularImport"
	ImportListTypeTraktUser      ImportListType = "TraktUserImport"
	ImportListTypePlexWatchlist  ImportListType = "PlexImport"
	ImportListTypeRadarrList     ImportListType = "RadarrImport"
	ImportListTypeStevenLu       ImportListType = "StevenLuImport"
	ImportListTypeRSSImport      ImportListType = "RSSImport"
	ImportListTypeIMDbList       ImportListType = "IMDbListImport"
	ImportListTypeCouchPotato    ImportListType = "CouchPotatoImport"
)

// ImportListSourceType represents the source type of an import list
type ImportListSourceType string

// Import list source types for categorizing list origins
const (
	ImportListSourceTypeProgram ImportListSourceType = "program"
	ImportListSourceTypeOther   ImportListSourceType = "other"
	ImportListSourceTypeAdvanced ImportListSourceType = "advanced"
)

// ImportListSettings contains the configuration settings for an import list
type ImportListSettings struct {
	BaseURL            string            `json:"baseUrl,omitempty"`
	APIKey             string            `json:"apiKey,omitempty"`
	AccessToken        string            `json:"accessToken,omitempty"`
	RefreshToken       string            `json:"refreshToken,omitempty"`
	Username           string            `json:"username,omitempty"`
	Password           string            `json:"password,omitempty"`
	ListID             string            `json:"listId,omitempty"`
	UserID             string            `json:"userId,omitempty"`
	Limit              int               `json:"limit,omitempty"`
	MinVotes           int               `json:"minVotes,omitempty"`
	MinRating          float64           `json:"minRating,omitempty"`
	Certification      string            `json:"certification,omitempty"`
	IncludeGenreIDs    IntArray          `json:"includeGenreIds,omitempty"`
	ExcludeGenreIDs    IntArray          `json:"excludeGenreIds,omitempty"`
	RatingFrom         string            `json:"ratingFrom,omitempty"`
	LanguageCode       string            `json:"languageCode,omitempty"`
	CountryCode        string            `json:"countryCode,omitempty"`
	URL                string            `json:"url,omitempty"`
	AdditionalSettings map[string]string `json:"additionalSettings,omitempty"`
}

// Value implements the driver.Valuer interface for database storage
func (s ImportListSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for database retrieval
func (s *ImportListSettings) Scan(value interface{}) error {
	if value == nil {
		*s = ImportListSettings{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// ImportListField represents a configuration field for an import list
type ImportListField struct {
	Name         string      `json:"name"`
	Label        string      `json:"label"`
	Value        interface{} `json:"value"`
	Type         string      `json:"type"`
	Advanced     bool        `json:"advanced"`
	Privacy      string      `json:"privacy"`
	SelectOptions []SelectOption `json:"selectOptions,omitempty"`
	HelpText     string      `json:"helpText,omitempty"`
	HelpLink     string      `json:"helpLink,omitempty"`
	Order        int         `json:"order"`
}

// SelectOption represents an option in a select field
type SelectOption struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// ImportListFieldsArray is a custom type for handling JSON arrays of import list fields
type ImportListFieldsArray []ImportListField

// Value implements the driver.Valuer interface for database storage
func (f ImportListFieldsArray) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan implements the sql.Scanner interface for database retrieval
func (f *ImportListFieldsArray) Scan(value interface{}) error {
	if value == nil {
		*f = ImportListFieldsArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, f)
}

// ImportListMovie represents a movie discovered from an import list
type ImportListMovie struct {
	ID                  int               `json:"id" gorm:"primaryKey;autoIncrement"`
	ImportListID        int               `json:"importListId" gorm:"not null;index"`
	ImportList          *ImportList       `json:"importList,omitempty" gorm:"foreignKey:ImportListID"`
	TmdbID              int               `json:"tmdbId" gorm:"not null;index"`
	ImdbID              string            `json:"imdbId" gorm:"size:20;index"`
	Title               string            `json:"title" gorm:"not null;size:500"`
	OriginalTitle       string            `json:"originalTitle" gorm:"size:500"`
	Year                int               `json:"year" gorm:"index"`
	Overview            string            `json:"overview" gorm:"type:text"`
	Runtime             int               `json:"runtime"`
	Images              MediaCover        `json:"images" gorm:"type:text"`
	Genres              StringArray       `json:"genres" gorm:"type:text"`
	Ratings             Ratings           `json:"ratings" gorm:"type:text"`
	Certification       string            `json:"certification" gorm:"size:20"`
	Status              MovieStatus       `json:"status" gorm:"size:20"`
	InCinemas           *time.Time        `json:"inCinemas,omitempty"`
	PhysicalRelease     *time.Time        `json:"physicalRelease,omitempty"`
	DigitalRelease      *time.Time        `json:"digitalRelease,omitempty"`
	Website             string            `json:"website" gorm:"size:500"`
	YouTubeTrailerID    string            `json:"youTubeTrailerId" gorm:"size:50"`
	Studio              string            `json:"studio" gorm:"size:255"`
	MinimumAvailability Availability      `json:"minimumAvailability" gorm:"size:20"`
	IsExcluded          bool              `json:"isExcluded" gorm:"default:false;index"`
	IsExisting          bool              `json:"isExisting" gorm:"default:false;index"`
	IsRecommendation    bool              `json:"isRecommendation" gorm:"default:false;index"`
	ListPosition        int               `json:"listPosition" gorm:"default:0"`
	DiscoveredAt        time.Time         `json:"discoveredAt" gorm:"autoCreateTime"`
	CreatedAt           time.Time         `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt           time.Time         `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ImportListExclusion represents a movie that should be excluded from import lists
type ImportListExclusion struct {
	ID            int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TmdbID        int       `json:"tmdbId" gorm:"not null;uniqueIndex"`
	MovieTitle    string    `json:"movieTitle" gorm:"not null;size:500"`
	MovieYear     int       `json:"movieYear" gorm:"not null"`
	ImdbID        string    `json:"imdbId,omitempty" gorm:"size:20"`
	Reason        string    `json:"reason,omitempty" gorm:"size:255"`
	CreatedAt     time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ImportListSyncResult represents the result of syncing an import list
type ImportListSyncResult struct {
	ImportListID    int                   `json:"importListId"`
	ImportListName  string                `json:"importListName"`
	MoviesTotal     int                   `json:"moviesTotal"`
	MoviesAdded     int                   `json:"moviesAdded"`
	MoviesUpdated   int                   `json:"moviesUpdated"`
	MoviesExcluded  int                   `json:"moviesExcluded"`
	MoviesExisting  int                   `json:"moviesExisting"`
	Movies          []ImportListMovie     `json:"movies"`
	SyncTime        time.Time             `json:"syncTime"`
	Errors          []string              `json:"errors,omitempty"`
	Success         bool                  `json:"success"`
}

// ImportListTestResult represents the result of testing an import list configuration
type ImportListTestResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"errors"`
	Movies  []ImportListMovie `json:"movies,omitempty"`
}

// IsEnabled returns whether the import list is enabled
func (il *ImportList) IsEnabled() bool {
	return il.Enabled && il.EnableAuto
}

// GetListType returns the import list source type
func (il *ImportList) GetListType() ImportListSourceType {
	if il.ListType == "" {
		return ImportListSourceTypeProgram
	}
	return il.ListType
}

// RequiresAuthentication checks if the import list type requires authentication
func (il *ImportList) RequiresAuthentication() bool {
	switch il.Implementation {
	case ImportListTypeTrakt, ImportListTypeTraktList, ImportListTypeTraktPopular, 
		 ImportListTypeTraktUser, ImportListTypePlexWatchlist:
		return true
	case ImportListTypeTMDBCollection, ImportListTypeTMDBCompany, ImportListTypeTMDBKeyword,
		 ImportListTypeTMDBList, ImportListTypeTMDBPerson, ImportListTypeTMDBPopular,
		 ImportListTypeTMDBUser, ImportListTypeRadarrList, ImportListTypeStevenLu,
		 ImportListTypeRSSImport, ImportListTypeIMDbList, ImportListTypeCouchPotato:
		return false
	default:
		return false
	}
}

// GetBaseURL returns the base URL for the import list if applicable
func (il *ImportList) GetBaseURL() string {
	switch il.Implementation {
	case ImportListTypeTrakt, ImportListTypeTraktList, ImportListTypeTraktPopular, ImportListTypeTraktUser:
		return "https://api.trakt.tv"
	case ImportListTypeTMDBCollection, ImportListTypeTMDBCompany, ImportListTypeTMDBKeyword,
		 ImportListTypeTMDBList, ImportListTypeTMDBPerson, ImportListTypeTMDBPopular, ImportListTypeTMDBUser:
		return "https://api.themoviedb.org/3"
	case ImportListTypeIMDbList:
		return "https://www.imdb.com"
	case ImportListTypePlexWatchlist, ImportListTypeRadarrList, ImportListTypeStevenLu, 
		 ImportListTypeRSSImport, ImportListTypeCouchPotato:
		return il.Settings.BaseURL
	default:
		return il.Settings.BaseURL
	}
}

// ShouldAutoAdd determines if movies from this list should be automatically added
func (il *ImportList) ShouldAutoAdd() bool {
	return il.EnableAuto && il.Enabled
}

// GetMinRefreshIntervalMinutes returns the minimum refresh interval in minutes
func (il *ImportList) GetMinRefreshIntervalMinutes() int {
	if il.MinRefreshInterval <= 0 {
		return 1440 // 24 hours default
	}
	return int(il.MinRefreshInterval.Minutes())
}