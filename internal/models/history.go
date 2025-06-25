package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// History represents a historical event in the system
type History struct {
	ID           int                    `json:"id" gorm:"primaryKey;autoIncrement"`
	MovieID      *int                   `json:"movieId,omitempty" gorm:"index"`
	Movie        *Movie                 `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	EventType    HistoryEventType       `json:"eventType" gorm:"not null;size:50;index"`
	Date         time.Time              `json:"date" gorm:"not null;index"`
	Quality      QualityDefinition      `json:"quality" gorm:"type:text"`
	SourceTitle  string                 `json:"sourceTitle" gorm:"size:500"`
	Language     Language               `json:"language" gorm:"type:text"`
	DownloadID   string                 `json:"downloadId,omitempty" gorm:"size:100;index"`
	Data         HistoryEventData       `json:"data" gorm:"type:text"`
	Message      string                 `json:"message,omitempty" gorm:"type:text"`
	Successful   bool                   `json:"successful" gorm:"default:true;index"`
	CreatedAt    time.Time              `json:"createdAt" gorm:"autoCreateTime"`
}

// HistoryEventType represents different types of historical events
type HistoryEventType string

// History event types for tracking different system activities
const (
	HistoryEventTypeGrabbed              HistoryEventType = "grabbed"
	HistoryEventTypeDownloadFolderImported HistoryEventType = "downloadFolderImported"
	HistoryEventTypeDownloadFailed       HistoryEventType = "downloadFailed"
	HistoryEventTypeMovieFileDeleted     HistoryEventType = "movieFileDeleted"
	HistoryEventTypeMovieFileRenamed     HistoryEventType = "movieFileRenamed"
	HistoryEventTypeMovieAdded           HistoryEventType = "movieAdded"
	HistoryEventTypeMovieDeleted         HistoryEventType = "movieDeleted"
	HistoryEventTypeMovieSearched        HistoryEventType = "movieSearched"
	HistoryEventTypeMovieRefreshed       HistoryEventType = "movieRefreshed"
	HistoryEventTypeQualityUpgraded      HistoryEventType = "qualityUpgraded"
	HistoryEventTypeMovieImported        HistoryEventType = "movieImported"
	HistoryEventTypeMovieUnmonitored     HistoryEventType = "movieUnmonitored"
	HistoryEventTypeMovieMonitored       HistoryEventType = "movieMonitored"
	HistoryEventTypeIgnoredDownload      HistoryEventType = "ignoredDownload"
)

// HistoryEventData contains additional data specific to the event type
type HistoryEventData struct {
	Indexer           string                 `json:"indexer,omitempty"`
	NzbInfoURL        string                 `json:"nzbInfoUrl,omitempty"`
	ReleaseGroup      string                 `json:"releaseGroup,omitempty"`
	Age               int                    `json:"age,omitempty"`
	AgeHours          float64                `json:"ageHours,omitempty"`
	AgeMinutes        float64                `json:"ageMinutes,omitempty"`
	PublishedDate     *time.Time             `json:"publishedDate,omitempty"`
	DownloadClient    string                 `json:"downloadClient,omitempty"`
	Size              int64                  `json:"size,omitempty"`
	DownloadURL       string                 `json:"downloadUrl,omitempty"`
	GUID              string                 `json:"guid,omitempty"`
	TvdbID            int                    `json:"tvdbId,omitempty"`
	TvRageID          int                    `json:"tvRageId,omitempty"`
	Protocol          string                 `json:"protocol,omitempty"`
	TorrentInfoHash   string                 `json:"torrentInfoHash,omitempty"`
	DroppedPath       string                 `json:"droppedPath,omitempty"`
	ImportedPath      string                 `json:"importedPath,omitempty"`
	DownloadedPath    string                 `json:"downloadedPath,omitempty"`
	Reason            string                 `json:"reason,omitempty"`
	StatusMessages    []HistoryStatusMessage `json:"statusMessages,omitempty"`
	FileID            int                    `json:"fileId,omitempty"`
	PreferredWordScore int                   `json:"preferredWordScore,omitempty"`
	CustomFormatScore int                    `json:"customFormatScore,omitempty"`
	CustomFormats     []string               `json:"customFormats,omitempty"`
	IndexerFlags      []string               `json:"indexerFlags,omitempty"`
	MovieMetadata     map[string]interface{} `json:"movieMetadata,omitempty"`
}

// Value implements the driver.Valuer interface for database storage
func (h HistoryEventData) Value() (driver.Value, error) {
	return json.Marshal(h)
}

// Scan implements the sql.Scanner interface for database retrieval
func (h *HistoryEventData) Scan(value interface{}) error {
	if value == nil {
		*h = HistoryEventData{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, h)
}

// HistoryStatusMessage represents a status message in history data
type HistoryStatusMessage struct {
	Title    string             `json:"title"`
	Messages []HistoryMessage   `json:"messages"`
}

// HistoryMessage represents an individual message
type HistoryMessage struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// Activity represents current system activity
type Activity struct {
	ID           int                    `json:"id" gorm:"primaryKey;autoIncrement"`
	Type         ActivityType           `json:"type" gorm:"not null;size:50;index"`
	Title        string                 `json:"title" gorm:"not null;size:255"`
	Message      string                 `json:"message,omitempty" gorm:"type:text"`
	MovieID      *int                   `json:"movieId,omitempty" gorm:"index"`
	Movie        *Movie                 `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	Progress     float64                `json:"progress" gorm:"default:0"`
	Status       ActivityStatus         `json:"status" gorm:"not null;size:20;index"`
	StartTime    time.Time              `json:"startTime" gorm:"not null;index"`
	EndTime      *time.Time             `json:"endTime,omitempty"`
	Data         ActivityData           `json:"data" gorm:"type:text"`
	CreatedAt    time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt    time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// ActivityType represents different types of system activities
type ActivityType string

// Activity types for tracking ongoing system operations
const (
	ActivityTypeMovieSearch     ActivityType = "movieSearch"
	ActivityTypeMovieRefresh    ActivityType = "movieRefresh"
	ActivityTypeDownload        ActivityType = "download"
	ActivityTypeImport          ActivityType = "import"
	ActivityTypeRename          ActivityType = "rename"
	ActivityTypeMetadataRefresh ActivityType = "metadataRefresh"
	ActivityTypeHealthCheck     ActivityType = "healthCheck"
	ActivityTypeBackup          ActivityType = "backup"
	ActivityTypeImportListSync  ActivityType = "importListSync"
	ActivityTypeIndexerTest     ActivityType = "indexerTest"
	ActivityTypeQueueProcess    ActivityType = "queueProcess"
	ActivityTypeSystemUpdate    ActivityType = "systemUpdate"
)

// ActivityStatus represents the current status of an activity
type ActivityStatus string

// Activity status values for tracking operation progress
const (
	ActivityStatusRunning   ActivityStatus = "running"
	ActivityStatusCompleted ActivityStatus = "completed"
	ActivityStatusFailed    ActivityStatus = "failed"
	ActivityStatusCancelled ActivityStatus = "cancelled"
	ActivityStatusQueued    ActivityStatus = "queued"
)

// ActivityData contains additional data specific to the activity type
type ActivityData struct {
	Command           string                 `json:"command,omitempty"`
	CommandID         int                    `json:"commandId,omitempty"`
	TotalItems        int                    `json:"totalItems,omitempty"`
	ProcessedItems    int                    `json:"processedItems,omitempty"`
	FailedItems       int                    `json:"failedItems,omitempty"`
	SuccessfulItems   int                    `json:"successfulItems,omitempty"`
	Errors            []string               `json:"errors,omitempty"`
	Warnings          []string               `json:"warnings,omitempty"`
	DownloadClient    string                 `json:"downloadClient,omitempty"`
	IndexerName       string                 `json:"indexerName,omitempty"`
	SearchQuery       string                 `json:"searchQuery,omitempty"`
	ResultCount       int                    `json:"resultCount,omitempty"`
	Duration          time.Duration          `json:"duration,omitempty"`
	EstimatedTime     time.Duration          `json:"estimatedTime,omitempty"`
	AdditionalData    map[string]interface{} `json:"additionalData,omitempty"`
}

// Value implements the driver.Valuer interface for database storage
func (a ActivityData) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for database retrieval
func (a *ActivityData) Scan(value interface{}) error {
	if value == nil {
		*a = ActivityData{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// HistoryRequest represents a request for history data with filtering options
type HistoryRequest struct {
	Page         int                `json:"page" form:"page"`
	PageSize     int                `json:"pageSize" form:"pageSize"`
	SortKey      string             `json:"sortKey" form:"sortKey"`
	SortDir      string             `json:"sortDir" form:"sortDir"`
	MovieID      *int               `json:"movieId" form:"movieId"`
	EventType    *HistoryEventType  `json:"eventType" form:"eventType"`
	Successful   *bool              `json:"successful" form:"successful"`
	DownloadID   string             `json:"downloadId" form:"downloadId"`
	Since        *time.Time         `json:"since" form:"since"`
	Until        *time.Time         `json:"until" form:"until"`
}

// HistoryResponse represents a paginated response of history records
type HistoryResponse struct {
	Page         int       `json:"page"`
	PageSize     int       `json:"pageSize"`
	SortKey      string    `json:"sortKey"`
	SortDir      string    `json:"sortDirection"`
	TotalRecords int64     `json:"totalRecords"`
	Records      []History `json:"records"`
}

// ActivityRequest represents a request for activity data with filtering options
type ActivityRequest struct {
	Page     int            `json:"page" form:"page"`
	PageSize int            `json:"pageSize" form:"pageSize"`
	Type     *ActivityType  `json:"type" form:"type"`
	Status   *ActivityStatus `json:"status" form:"status"`
	MovieID  *int           `json:"movieId" form:"movieId"`
	Since    *time.Time     `json:"since" form:"since"`
	Until    *time.Time     `json:"until" form:"until"`
}

// ActivityResponse represents a paginated response of activity records
type ActivityResponse struct {
	Page         int        `json:"page"`
	PageSize     int        `json:"pageSize"`
	TotalRecords int64      `json:"totalRecords"`
	Records      []Activity `json:"records"`
}

// IsCompleted returns whether the activity has completed (successfully or not)
func (a *Activity) IsCompleted() bool {
	return a.Status == ActivityStatusCompleted || a.Status == ActivityStatusFailed || a.Status == ActivityStatusCancelled
}

// IsRunning returns whether the activity is currently running
func (a *Activity) IsRunning() bool {
	return a.Status == ActivityStatusRunning
}

// GetDuration returns the duration of the activity if completed
func (a *Activity) GetDuration() time.Duration {
	if a.EndTime != nil {
		return a.EndTime.Sub(a.StartTime)
	}
	if a.IsRunning() {
		return time.Since(a.StartTime)
	}
	return 0
}

// UpdateProgress updates the activity progress and estimated completion time
func (a *Activity) UpdateProgress(processed, total int) {
	if total > 0 {
		a.Progress = float64(processed) / float64(total) * 100
		
		// Update data
		a.Data.ProcessedItems = processed
		a.Data.TotalItems = total
		
		// Estimate remaining time if we have progress
		if a.Progress > 0 && a.Progress < 100 {
			elapsed := time.Since(a.StartTime)
			estimatedTotal := time.Duration(float64(elapsed) / (a.Progress / 100))
			a.Data.EstimatedTime = estimatedTotal - elapsed
		}
	}
}

// Complete marks the activity as completed
func (a *Activity) Complete(successful bool) {
	now := time.Now()
	a.EndTime = &now
	a.Progress = 100
	a.Data.Duration = a.GetDuration()
	
	if successful {
		a.Status = ActivityStatusCompleted
	} else {
		a.Status = ActivityStatusFailed
	}
}

// Cancel marks the activity as cancelled
func (a *Activity) Cancel() {
	now := time.Now()
	a.EndTime = &now
	a.Status = ActivityStatusCancelled
	a.Data.Duration = a.GetDuration()
}

// Fail marks the activity as failed with an error message
func (a *Activity) Fail(errorMsg string) {
	now := time.Now()
	a.EndTime = &now
	a.Status = ActivityStatusFailed
	a.Data.Duration = a.GetDuration()
	
	if errorMsg != "" {
		a.Data.Errors = append(a.Data.Errors, errorMsg)
		a.Message = errorMsg
	}
}