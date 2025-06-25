package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ReleaseStatus represents the status of a release
type ReleaseStatus string

const (
	// ReleaseStatusAvailable indicates the release is available for download
	ReleaseStatusAvailable ReleaseStatus = "available"
	// ReleaseStatusGrabbed indicates the release has been grabbed/sent to download client
	ReleaseStatusGrabbed ReleaseStatus = "grabbed"
	// ReleaseStatusRejected indicates the release has been rejected
	ReleaseStatusRejected ReleaseStatus = "rejected"
	// ReleaseStatusFailed indicates the release failed to download
	ReleaseStatusFailed ReleaseStatus = "failed"
)

// ReleaseSource represents where the release was found
type ReleaseSource string

const (
	// ReleaseSourceRSS indicates the release was found via RSS
	ReleaseSourceRSS ReleaseSource = "rss"
	// ReleaseSourceSearch indicates the release was found via search
	ReleaseSourceSearch ReleaseSource = "search"
	// ReleaseSourceInteractiveSearch indicates the release was found via manual search
	ReleaseSourceInteractiveSearch ReleaseSource = "interactive"
)

// Protocol represents the download protocol
type Protocol string

const (
	// ProtocolTorrent represents torrent protocol
	ProtocolTorrent Protocol = "torrent"
	// ProtocolUsenet represents usenet protocol
	ProtocolUsenet Protocol = "usenet"
)

// ReleaseInfo contains detailed information about a release
type ReleaseInfo struct {
	Title                string   `json:"title"`
	Description          string   `json:"description,omitempty"`
	Year                 int      `json:"year,omitempty"`
	Edition              string   `json:"edition,omitempty"`
	Languages            []string `json:"languages,omitempty"`
	Subtitles            []string `json:"subtitles,omitempty"`
	Resolution           string   `json:"resolution,omitempty"`
	Source               string   `json:"source,omitempty"`
	Codec                string   `json:"codec,omitempty"`
	Container            string   `json:"container,omitempty"`
	ReleaseGroup         string   `json:"releaseGroup,omitempty"`
	Scene                bool     `json:"scene"`
	Freeleech            bool     `json:"freeleech"`
	DownloadVolumeFactor float64  `json:"downloadVolumeFactor"`
	UploadVolumeFactor   float64  `json:"uploadVolumeFactor"`
	MinimumRatio         *float64 `json:"minimumRatio,omitempty"`
	MinimumSeedTime      *int     `json:"minimumSeedTime,omitempty"`
}

// Value implements the driver.Valuer interface for database storage
func (ri ReleaseInfo) Value() (driver.Value, error) {
	return json.Marshal(ri)
}

// Scan implements the sql.Scanner interface for database retrieval
func (ri *ReleaseInfo) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ri)
	case string:
		return json.Unmarshal([]byte(v), ri)
	default:
		return fmt.Errorf("cannot scan %T into ReleaseInfo", value)
	}
}

// Release represents a movie release found by indexers
type Release struct {
	ID               int             `json:"id" gorm:"primaryKey;autoIncrement"`
	GUID             string          `json:"guid" gorm:"not null;size:500;uniqueIndex"`
	Title            string          `json:"title" gorm:"not null;size:500"`
	SortTitle        string          `json:"sortTitle" gorm:"size:500;index"`
	Overview         string          `json:"overview" gorm:"type:text"`
	Quality          Quality         `json:"quality" gorm:"type:text"`
	QualityWeight    int             `json:"qualityWeight" gorm:"index"`
	Age              int             `json:"age"`
	AgeHours         float64         `json:"ageHours"`
	AgeMinutes       float64         `json:"ageMinutes"`
	Size             int64           `json:"size"`
	IndexerID        int             `json:"indexerId" gorm:"not null;index"`
	Indexer          *Indexer        `json:"indexer,omitempty" gorm:"foreignKey:IndexerID"`
	MovieID          *int            `json:"movieId,omitempty" gorm:"index"`
	Movie            *Movie          `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	ImdbID           string          `json:"imdbId" gorm:"size:20;index"`
	TmdbID           *int            `json:"tmdbId,omitempty" gorm:"index"`
	Protocol         Protocol        `json:"protocol" gorm:"not null;size:20"`
	DownloadURL      string          `json:"downloadUrl" gorm:"not null;size:2000"`
	InfoURL          string          `json:"infoUrl" gorm:"size:2000"`
	CommentURL       string          `json:"commentUrl" gorm:"size:2000"`
	Seeders          *int            `json:"seeders,omitempty"`
	Leechers         *int            `json:"leechers,omitempty"`
	PeerCount        int             `json:"peers"`
	PublishDate      time.Time       `json:"publishDate" gorm:"not null;index"`
	Status           ReleaseStatus   `json:"status" gorm:"default:'available';index"`
	Source           ReleaseSource   `json:"source" gorm:"not null;size:20"`
	ReleaseInfo      ReleaseInfo     `json:"releaseInfo" gorm:"type:text"`
	Categories       IntArray        `json:"categories" gorm:"type:text"`
	DownloadClientID *int            `json:"downloadClientId,omitempty"`
	DownloadClient   *DownloadClient `json:"downloadClient,omitempty" gorm:"foreignKey:DownloadClientID"`
	RejectionReasons StringArray     `json:"rejectionReasons" gorm:"type:text"`
	IndexerFlags     int             `json:"indexerFlags" gorm:"default:0"`
	SceneMapping     bool            `json:"sceneMapping" gorm:"default:false"`
	MagnetURL        string          `json:"magnetUrl" gorm:"size:2000"`
	CreatedAt        time.Time       `json:"added" gorm:"autoCreateTime;index"`
	UpdatedAt        time.Time       `json:"updated" gorm:"autoUpdateTime"`
	GrabbedAt        *time.Time      `json:"grabbedAt,omitempty"`
	FailedAt         *time.Time      `json:"failedAt,omitempty"`
}

// TableName returns the database table name for the Release model
func (Release) TableName() string {
	return "releases"
}

// IsGrabbable returns true if the release can be grabbed
func (r *Release) IsGrabbable() bool {
	return r.Status == ReleaseStatusAvailable && len(r.RejectionReasons) == 0
}

// IsTorrent returns true if the release uses torrent protocol
func (r *Release) IsTorrent() bool {
	return r.Protocol == ProtocolTorrent
}

// IsUsenet returns true if the release uses usenet protocol
func (r *Release) IsUsenet() bool {
	return r.Protocol == ProtocolUsenet
}

// GetSizeString returns formatted size string
func (r *Release) GetSizeString() string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	size := float64(r.Size)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", size/GB)
	case size >= MB:
		return fmt.Sprintf("%.1f MB", size/MB)
	case size >= KB:
		return fmt.Sprintf("%.1f KB", size/KB)
	default:
		return fmt.Sprintf("%d B", r.Size)
	}
}

// GetAgeString returns formatted age string
func (r *Release) GetAgeString() string {
	switch {
	case r.Age >= 365:
		years := r.Age / 365
		return fmt.Sprintf("%d year(s)", years)
	case r.Age >= 30:
		months := r.Age / 30
		return fmt.Sprintf("%d month(s)", months)
	case r.Age >= 1:
		return fmt.Sprintf("%d day(s)", r.Age)
	case r.AgeHours >= 1:
		return fmt.Sprintf("%.1f hour(s)", r.AgeHours)
	default:
		return fmt.Sprintf("%.1f minute(s)", r.AgeMinutes)
	}
}

// HasRejections returns true if the release has rejection reasons
func (r *Release) HasRejections() bool {
	return len(r.RejectionReasons) > 0
}

// GetRejectionString returns formatted rejection reasons
func (r *Release) GetRejectionString() string {
	if len(r.RejectionReasons) == 0 {
		return ""
	}
	return strings.Join(r.RejectionReasons, ", ")
}

// SearchRequest represents a search request for releases
type SearchRequest struct {
	MovieID    *int          `json:"movieId,omitempty"`
	ImdbID     string        `json:"imdbId,omitempty"`
	TmdbID     *int          `json:"tmdbId,omitempty"`
	Title      string        `json:"title,omitempty"`
	Year       *int          `json:"year,omitempty"`
	Categories []int         `json:"categories,omitempty"`
	IndexerIDs []int         `json:"indexerIds,omitempty"`
	Limit      int           `json:"limit,omitempty"`
	Offset     int           `json:"offset,omitempty"`
	SortBy     string        `json:"sortBy,omitempty"`
	SortOrder  string        `json:"sortOrder,omitempty"`
	Protocol   *Protocol     `json:"protocol,omitempty"`
	Source     ReleaseSource `json:"source"`
}

// SearchResponse represents the response from a search request
type SearchResponse struct {
	Releases   []Release `json:"releases"`
	Total      int       `json:"total"`
	Offset     int       `json:"offset"`
	Limit      int       `json:"limit"`
	IndexerID  int       `json:"indexerId"`
	SearchTime float64   `json:"searchTime"`
}

// GrabRequest represents a request to grab a release
type GrabRequest struct {
	GUID             string `json:"guid" binding:"required"`
	IndexerID        int    `json:"indexerId" binding:"required"`
	MovieID          *int   `json:"movieId,omitempty"`
	DownloadClientID *int   `json:"downloadClientId,omitempty"`
}

// GrabResponse represents the response from grabbing a release
type GrabResponse struct {
	ID               int    `json:"id"`
	GUID             string `json:"guid"`
	Title            string `json:"title"`
	Status           string `json:"status"`
	DownloadClientID *int   `json:"downloadClientId,omitempty"`
	Message          string `json:"message,omitempty"`
}

// ReleaseFilter represents filters for release queries
type ReleaseFilter struct {
	Status        []ReleaseStatus `json:"status,omitempty"`
	Source        []ReleaseSource `json:"source,omitempty"`
	Protocol      []Protocol      `json:"protocol,omitempty"`
	IndexerIDs    []int           `json:"indexerIds,omitempty"`
	MovieIDs      []int           `json:"movieIds,omitempty"`
	MinSize       *int64          `json:"minSize,omitempty"`
	MaxSize       *int64          `json:"maxSize,omitempty"`
	MinAge        *int            `json:"minAge,omitempty"`
	MaxAge        *int            `json:"maxAge,omitempty"`
	HasSeeders    *bool           `json:"hasSeeders,omitempty"`
	MinSeeders    *int            `json:"minSeeders,omitempty"`
	Categories    []int           `json:"categories,omitempty"`
	Languages     []string        `json:"languages,omitempty"`
	Resolutions   []string        `json:"resolutions,omitempty"`
	Sources       []string        `json:"sources,omitempty"`
	Codecs        []string        `json:"codecs,omitempty"`
	Freeleech     *bool           `json:"freeleech,omitempty"`
	Scene         *bool           `json:"scene,omitempty"`
	CreatedAfter  *time.Time      `json:"createdAfter,omitempty"`
	CreatedBefore *time.Time      `json:"createdBefore,omitempty"`
}

// ReleaseStats represents statistics about releases
type ReleaseStats struct {
	TotalReleases     int                   `json:"totalReleases"`
	AvailableReleases int                   `json:"availableReleases"`
	GrabbedReleases   int                   `json:"grabbedReleases"`
	RejectedReleases  int                   `json:"rejectedReleases"`
	FailedReleases    int                   `json:"failedReleases"`
	ProtocolBreakdown map[Protocol]int      `json:"protocolBreakdown"`
	SourceBreakdown   map[ReleaseSource]int `json:"sourceBreakdown"`
	IndexerBreakdown  map[int]int           `json:"indexerBreakdown"`
	AverageSize       float64               `json:"averageSize"`
	TotalSize         int64                 `json:"totalSize"`
	AverageAge        float64               `json:"averageAge"`
	LastUpdated       time.Time             `json:"lastUpdated"`
}
