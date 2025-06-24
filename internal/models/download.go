// Package models provides data structures for the Radarr application.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// DownloadClientType represents the type of download client
type DownloadClientType string

const (
	// DownloadClientTypeQBittorrent represents qBittorrent client
	DownloadClientTypeQBittorrent DownloadClientType = "qbittorrent"
	// DownloadClientTypeTransmission represents Transmission client
	DownloadClientTypeTransmission DownloadClientType = "transmission"
	// DownloadClientTypeDeluge represents Deluge client
	DownloadClientTypeDeluge DownloadClientType = "deluge"
	// DownloadClientTypeSABnzbd represents SABnzbd client
	DownloadClientTypeSABnzbd DownloadClientType = "sabnzbd"
	// DownloadClientTypeNZBGet represents NZBGet client
	DownloadClientTypeNZBGet DownloadClientType = "nzbget"
	// DownloadClientTypeRTorrent represents rTorrent client
	DownloadClientTypeRTorrent DownloadClientType = "rtorrent"
	// DownloadClientTypeUtorrent represents uTorrent client
	DownloadClientTypeUtorrent DownloadClientType = "utorrent"
)

// DownloadProtocol represents the download protocol
type DownloadProtocol string

const (
	// DownloadProtocolTorrent represents torrent downloads
	DownloadProtocolTorrent DownloadProtocol = "torrent"
	// DownloadProtocolUsenet represents usenet downloads
	DownloadProtocolUsenet DownloadProtocol = "usenet"
)

// DownloadClientStatus represents the status of a download client
type DownloadClientStatus string

const (
	// DownloadClientStatusEnabled represents an enabled download client
	DownloadClientStatusEnabled DownloadClientStatus = "enabled"
	// DownloadClientStatusDisabled represents a disabled download client
	DownloadClientStatusDisabled DownloadClientStatus = "disabled"
	// DownloadClientStatusError represents a download client with errors
	DownloadClientStatusError DownloadClientStatus = "error"
)

// DownloadClientSettings represents flexible configuration for different download clients
type DownloadClientSettings map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (dcs DownloadClientSettings) Value() (driver.Value, error) {
	if dcs == nil {
		return nil, nil
	}
	return json.Marshal(dcs)
}

// Scan implements the sql.Scanner interface for database retrieval
func (dcs *DownloadClientSettings) Scan(value interface{}) error {
	if value == nil {
		*dcs = make(DownloadClientSettings)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, dcs)
	case string:
		return json.Unmarshal([]byte(v), dcs)
	default:
		return fmt.Errorf("cannot scan %T into DownloadClientSettings", value)
	}
}

// DownloadClient represents a download client configuration
type DownloadClient struct {
	ID                       int                    `json:"id" gorm:"primaryKey;autoIncrement"`
	Name                     string                 `json:"name" gorm:"not null;size:255"`
	Type                     DownloadClientType     `json:"implementation" gorm:"not null;size:50"`
	Protocol                 DownloadProtocol       `json:"protocol" gorm:"not null;size:20"`
	Host                     string                 `json:"host" gorm:"not null;size:255"`
	Port                     int                    `json:"port" gorm:"default:8080"`
	Username                 string                 `json:"username" gorm:"size:255"`
	Password                 string                 `json:"password" gorm:"size:255"`
	APIKey                   string                 `json:"apiKey" gorm:"size:255"`
	Category                 string                 `json:"category" gorm:"size:100"`
	RecentMoviePriority      string                 `json:"recentMoviePriority" gorm:"default:'Normal';size:20"`
	OlderMoviePriority       string                 `json:"olderMoviePriority" gorm:"default:'Normal';size:20"`
	AddPaused                bool                   `json:"addPaused" gorm:"default:false"`
	UseSsl                   bool                   `json:"useSsl" gorm:"default:false"`
	Enable                   bool                   `json:"enable" gorm:"default:true"`
	RemoveCompletedDownloads bool                   `json:"removeCompletedDownloads" gorm:"default:true"`
	RemoveFailedDownloads    bool                   `json:"removeFailedDownloads" gorm:"default:true"`
	Priority                 int                    `json:"priority" gorm:"default:1"`
	Settings                 DownloadClientSettings `json:"fields" gorm:"type:text"`
	Tags                     IntArray               `json:"tags" gorm:"type:text"`
	CreatedAt                time.Time              `json:"added" gorm:"autoCreateTime"`
	UpdatedAt                time.Time              `json:"updated" gorm:"autoUpdateTime"`
}

// TableName returns the database table name for the DownloadClient model
func (DownloadClient) TableName() string {
	return "download_clients"
}

// IsEnabled returns true if the download client is enabled
func (dc *DownloadClient) IsEnabled() bool {
	return dc.Enable
}

// SupportsProtocol returns true if the client supports the given protocol
func (dc *DownloadClient) SupportsProtocol(protocol DownloadProtocol) bool {
	return dc.Protocol == protocol
}

// GetBaseURL returns the full base URL for the download client
func (dc *DownloadClient) GetBaseURL() string {
	scheme := "http"
	if dc.UseSsl {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s:%d", scheme, dc.Host, dc.Port)
}

// QueueStatus represents the status of a download in the queue
type QueueStatus string

const (
	// QueueStatusQueued represents a queued download
	QueueStatusQueued QueueStatus = "queued"
	// QueueStatusPaused represents a paused download
	QueueStatusPaused QueueStatus = "paused"
	// QueueStatusDownloading represents an active download
	QueueStatusDownloading QueueStatus = "downloading"
	// QueueStatusCompleted represents a completed download
	QueueStatusCompleted QueueStatus = "completed"
	// QueueStatusFailed represents a failed download
	QueueStatusFailed QueueStatus = "failed"
	// QueueStatusWarning represents a download with warnings
	QueueStatusWarning QueueStatus = "warning"
)

// QueueItem represents an item in the download queue
type QueueItem struct {
	ID                      int              `json:"id" gorm:"primaryKey;autoIncrement"`
	MovieID                 int              `json:"movieId" gorm:"index"`
	Movie                   *Movie           `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	DownloadClientID        int              `json:"downloadClientId" gorm:"index"`
	DownloadClient          *DownloadClient  `json:"downloadClient,omitempty" gorm:"foreignKey:DownloadClientID"`
	DownloadID              string           `json:"downloadId" gorm:"size:255;index"`
	Title                   string           `json:"title" gorm:"not null;size:500"`
	Size                    int64            `json:"size"`
	SizeLeft                int64            `json:"sizeleft"`
	Status                  QueueStatus      `json:"status" gorm:"not null;size:20"`
	TrackedDownloadStatus   string           `json:"trackedDownloadStatus" gorm:"size:50"`
	StatusMessages          StringArray      `json:"statusMessages" gorm:"type:text"`
	DownloadedInfo          DownloadedInfo   `json:"downloadedInfo" gorm:"type:text"`
	ErrorMessage            string           `json:"errorMessage" gorm:"type:text"`
	Added                   time.Time        `json:"added" gorm:"autoCreateTime"`
	Updated                 time.Time        `json:"updated" gorm:"autoUpdateTime"`
	TimeLeft                *time.Duration   `json:"timeleft,omitempty"`
	EstimatedCompletionTime *time.Time       `json:"estimatedCompletionTime,omitempty"`
	Protocol                DownloadProtocol `json:"protocol" gorm:"size:20"`
	OutputPath              string           `json:"outputPath" gorm:"size:500"`
}

// TableName returns the database table name for the QueueItem model
func (QueueItem) TableName() string {
	return "queue_items"
}

// DownloadedInfo contains information about downloaded files
type DownloadedInfo struct {
	Hash           string `json:"hash"`
	Category       string `json:"category"`
	DownloadedPath string `json:"downloadedPath"`
}

// Value implements the driver.Valuer interface for database storage
func (di DownloadedInfo) Value() (driver.Value, error) {
	return json.Marshal(di)
}

// Scan implements the sql.Scanner interface for database retrieval
func (di *DownloadedInfo) Scan(value interface{}) error {
	if value == nil {
		*di = DownloadedInfo{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, di)
	case string:
		return json.Unmarshal([]byte(v), di)
	default:
		return fmt.Errorf("cannot scan %T into DownloadedInfo", value)
	}
}

// IsCompleted returns true if the download is completed
func (qi *QueueItem) IsCompleted() bool {
	return qi.Status == QueueStatusCompleted
}

// IsFailed returns true if the download has failed
func (qi *QueueItem) IsFailed() bool {
	return qi.Status == QueueStatusFailed
}

// IsActive returns true if the download is actively downloading
func (qi *QueueItem) IsActive() bool {
	return qi.Status == QueueStatusDownloading
}

// GetProgress returns the download progress as a percentage (0-100)
func (qi *QueueItem) GetProgress() float64 {
	if qi.Size <= 0 {
		return 0
	}
	downloaded := qi.Size - qi.SizeLeft
	if downloaded <= 0 {
		return 0
	}
	return float64(downloaded) / float64(qi.Size) * 100
}

// DownloadClientTestResult represents the result of testing a download client connection
type DownloadClientTestResult struct {
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"validationFailures"`
}

// DownloadHistory represents a completed download from history
type DownloadHistory struct {
	ID               int              `json:"id" gorm:"primaryKey;autoIncrement"`
	MovieID          int              `json:"movieId" gorm:"index"`
	Movie            *Movie           `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	DownloadClientID int              `json:"downloadClientId" gorm:"index"`
	DownloadClient   *DownloadClient  `json:"downloadClient,omitempty" gorm:"foreignKey:DownloadClientID"`
	SourceTitle      string           `json:"sourceTitle" gorm:"not null;size:500"`
	Date             time.Time        `json:"date" gorm:"not null;index"`
	Protocol         DownloadProtocol `json:"protocol" gorm:"not null;size:20"`
	IndexerName      string           `json:"indexer" gorm:"size:255"`
	DownloadID       string           `json:"downloadId" gorm:"size:255"`
	Successful       bool             `json:"successful" gorm:"not null;index"`
	Data             HistoryData      `json:"data" gorm:"type:text"`
}

// TableName returns the database table name for the DownloadHistory model
func (DownloadHistory) TableName() string {
	return "download_history"
}

// HistoryData contains additional data about a download
type HistoryData map[string]interface{}

// Value implements the driver.Valuer interface for database storage
func (hd HistoryData) Value() (driver.Value, error) {
	if hd == nil {
		return nil, nil
	}
	return json.Marshal(hd)
}

// Scan implements the sql.Scanner interface for database retrieval
func (hd *HistoryData) Scan(value interface{}) error {
	if value == nil {
		*hd = make(HistoryData)
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, hd)
	case string:
		return json.Unmarshal([]byte(v), hd)
	default:
		return fmt.Errorf("cannot scan %T into HistoryData", value)
	}
}
