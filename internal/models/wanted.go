// Package models provides data structures for the Radarr application.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// WantedStatus represents the reason why a movie is wanted
type WantedStatus string

const (
	// WantedStatusMissing indicates the movie has no file at all
	WantedStatusMissing WantedStatus = "missing"
	// WantedStatusCutoffUnmet indicates the movie has a file but quality is below cutoff
	WantedStatusCutoffUnmet WantedStatus = "cutoffUnmet"
	// WantedStatusUpgrade indicates the movie could be upgraded to better quality
	WantedStatusUpgrade WantedStatus = "upgrade"
)

// WantedMovie represents a movie that needs to be downloaded or upgraded
type WantedMovie struct {
	ID                int            `json:"id" gorm:"primaryKey"`
	MovieID           int            `json:"movieId" gorm:"not null;uniqueIndex"`
	Status            WantedStatus   `json:"status" gorm:"not null"`
	Reason            string         `json:"reason"`
	CurrentQualityID  *int           `json:"currentQualityId,omitempty"`
	TargetQualityID   int            `json:"targetQualityId"`
	IsAvailable       bool           `json:"isAvailable"`
	LastSearchTime    *time.Time     `json:"lastSearchTime,omitempty"`
	NextSearchTime    *time.Time     `json:"nextSearchTime,omitempty"`
	SearchAttempts    int            `json:"searchAttempts" gorm:"default:0"`
	MaxSearchAttempts int            `json:"maxSearchAttempts" gorm:"default:10"`
	Priority          WantedPriority `json:"priority" gorm:"default:3"`
	SearchFailures    SearchFailures `json:"searchFailures" gorm:"type:text"`
	CreatedAt         time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`

	// Relationships
	Movie          *Movie        `json:"movie,omitempty" gorm:"foreignKey:MovieID"`
	CurrentQuality *QualityLevel `json:"currentQuality,omitempty" gorm:"foreignKey:CurrentQualityID"`
	TargetQuality  *QualityLevel `json:"targetQuality,omitempty" gorm:"foreignKey:TargetQualityID"`
}

// TableName returns the database table name for the WantedMovie model
func (WantedMovie) TableName() string {
	return "wanted_movies"
}

// WantedPriority represents the priority level for wanted movie searches
type WantedPriority int

const (
	// PriorityVeryLow for movies that can wait
	PriorityVeryLow WantedPriority = 1
	// PriorityLow for movies with low priority
	PriorityLow WantedPriority = 2
	// PriorityNormal for standard priority movies
	PriorityNormal WantedPriority = 3
	// PriorityHigh for important movies
	PriorityHigh WantedPriority = 4
	// PriorityVeryHigh for urgent movies
	PriorityVeryHigh WantedPriority = 5
)

// SearchFailure represents a single search failure for a wanted movie
type SearchFailure struct {
	Timestamp time.Time `json:"timestamp"`
	Reason    string    `json:"reason"`
	Indexer   string    `json:"indexer,omitempty"`
	ErrorCode string    `json:"errorCode,omitempty"`
}

// SearchFailures represents a collection of search failures
type SearchFailures []SearchFailure

// Value implements the driver.Valuer interface for database storage
func (sf SearchFailures) Value() (driver.Value, error) {
	if sf == nil {
		return nil, nil
	}
	return json.Marshal(sf)
}

// Scan implements the sql.Scanner interface for database retrieval
func (sf *SearchFailures) Scan(value interface{}) error {
	if value == nil {
		*sf = SearchFailures{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, sf)
	case string:
		return json.Unmarshal([]byte(v), sf)
	default:
		return fmt.Errorf("cannot scan %T into SearchFailures", value)
	}
}

// AddFailure adds a new search failure to the collection
func (sf *SearchFailures) AddFailure(reason, indexer, errorCode string) {
	failure := SearchFailure{
		Timestamp: time.Now(),
		Reason:    reason,
		Indexer:   indexer,
		ErrorCode: errorCode,
	}
	*sf = append(*sf, failure)
}

// GetRecentFailures returns failures within the specified duration
func (sf *SearchFailures) GetRecentFailures(since time.Duration) []SearchFailure {
	cutoff := time.Now().Add(-since)
	var recent []SearchFailure

	for _, failure := range *sf {
		if failure.Timestamp.After(cutoff) {
			recent = append(recent, failure)
		}
	}

	return recent
}

// WantedMovieFilter represents filtering criteria for wanted movies
type WantedMovieFilter struct {
	Status           *WantedStatus   `json:"status,omitempty"`
	Priority         *WantedPriority `json:"priority,omitempty"`
	MinPriority      *WantedPriority `json:"minPriority,omitempty"`
	MaxPriority      *WantedPriority `json:"maxPriority,omitempty"`
	IsAvailable      *bool           `json:"isAvailable,omitempty"`
	QualityProfileID *int            `json:"qualityProfileId,omitempty"`
	SearchRequired   *bool           `json:"searchRequired,omitempty"`
	LastSearchBefore *time.Time      `json:"lastSearchBefore,omitempty"`
	LastSearchAfter  *time.Time      `json:"lastSearchAfter,omitempty"`
	Monitored        *bool           `json:"monitored,omitempty"`
	Year             *int            `json:"year,omitempty"`
	Genre            *string         `json:"genre,omitempty"`
	SortBy           string          `json:"sortBy,omitempty"`
	SortDir          string          `json:"sortDir,omitempty"`
	Page             int             `json:"page"`
	PageSize         int             `json:"pageSize"`
}

// WantedMoviesResponse represents a paginated response of wanted movies
type WantedMoviesResponse struct {
	Records       []WantedMovie `json:"records"`
	Page          int           `json:"page"`
	PageSize      int           `json:"pageSize"`
	SortKey       string        `json:"sortKey"`
	SortDirection string        `json:"sortDirection"`
	TotalRecords  int64         `json:"totalRecords"`
	FilteredCount int64         `json:"filteredCount"`
}

// WantedMoviesBulkOperation represents a bulk operation on wanted movies
type WantedMoviesBulkOperation struct {
	MovieIDs  []int                     `json:"movieIds"`
	Operation WantedMoviesBulkOpType    `json:"operation"`
	Options   WantedMoviesBulkOpOptions `json:"options,omitempty"`
}

// WantedMoviesBulkOpType represents the type of bulk operation
type WantedMoviesBulkOpType string

const (
	// BulkOpSearch triggers searches for selected wanted movies
	BulkOpSearch WantedMoviesBulkOpType = "search"
	// BulkOpSetPriority changes priority for selected wanted movies
	BulkOpSetPriority WantedMoviesBulkOpType = "setPriority"
	// BulkOpRemove removes movies from wanted list (if they have files)
	BulkOpRemove WantedMoviesBulkOpType = "remove"
	// BulkOpResetSearchAttempts resets search attempt counters
	BulkOpResetSearchAttempts WantedMoviesBulkOpType = "resetSearchAttempts"
)

// WantedMoviesBulkOpOptions contains options for bulk operations
type WantedMoviesBulkOpOptions struct {
	Priority    *WantedPriority `json:"priority,omitempty"`
	ForceSearch bool            `json:"forceSearch,omitempty"`
	SearchType  string          `json:"searchType,omitempty"`
}

// WantedSearchTrigger represents a request to trigger searches for wanted movies
type WantedSearchTrigger struct {
	MovieIDs      []int  `json:"movieIds,omitempty"`   // Empty means all wanted movies
	FilterMissing bool   `json:"filterMissing"`        // Search only missing movies
	FilterCutoff  bool   `json:"filterCutoff"`         // Search only cutoff unmet movies
	ForceSearch   bool   `json:"forceSearch"`          // Ignore search cooldowns
	SearchType    string `json:"searchType,omitempty"` // "automatic" or "manual"
}

// IsEligibleForSearch determines if a wanted movie is ready for another search
func (w *WantedMovie) IsEligibleForSearch() bool {
	// Check if max search attempts reached
	if w.SearchAttempts >= w.MaxSearchAttempts {
		return false
	}

	// Check if we should wait before next search
	if w.NextSearchTime != nil && time.Now().Before(*w.NextSearchTime) {
		return false
	}

	// Must be available for search
	if !w.IsAvailable {
		return false
	}

	return true
}

// CalculateNextSearchTime calculates when the next search should occur based on attempts and priority
func (w *WantedMovie) CalculateNextSearchTime() time.Time {
	baseDelay := time.Hour * 2 // Base delay of 2 hours

	// Increase delay based on search attempts (exponential backoff)
	multiplier := 1
	if w.SearchAttempts > 0 {
		multiplier = w.SearchAttempts * w.SearchAttempts
		if multiplier > 24 { // Cap at 24x delay (48 hours max with 2h base)
			multiplier = 24
		}
	}

	// Adjust delay based on priority
	priorityMultiplier := float64(6-int(w.Priority)) / 2.0 // Higher priority = shorter delay
	if priorityMultiplier < 0.5 {
		priorityMultiplier = 0.5
	}

	finalDelay := time.Duration(float64(baseDelay) * float64(multiplier) * priorityMultiplier)
	return time.Now().Add(finalDelay)
}

// IncrementSearchAttempts increments the search attempt counter and updates next search time
func (w *WantedMovie) IncrementSearchAttempts() {
	w.SearchAttempts++
	nextSearch := w.CalculateNextSearchTime()
	w.NextSearchTime = &nextSearch
	w.LastSearchTime = &[]time.Time{time.Now()}[0]
}

// ResetSearchAttempts resets the search attempt counter and clears next search time
func (w *WantedMovie) ResetSearchAttempts() {
	w.SearchAttempts = 0
	w.NextSearchTime = nil
	w.SearchFailures = SearchFailures{}
}

// GetPriorityString returns a human-readable priority string
func (w *WantedMovie) GetPriorityString() string {
	switch w.Priority {
	case PriorityVeryLow:
		return "Very Low"
	case PriorityLow:
		return "Low"
	case PriorityNormal:
		return "Normal"
	case PriorityHigh:
		return "High"
	case PriorityVeryHigh:
		return "Very High"
	default:
		return "Normal"
	}
}

// WantedMoviesStats represents statistics about wanted movies
type WantedMoviesStats struct {
	TotalWanted       int64 `json:"totalWanted"`
	MissingCount      int64 `json:"missingCount"`
	CutoffUnmetCount  int64 `json:"cutoffUnmetCount"`
	UpgradeCount      int64 `json:"upgradeCount"`
	AvailableCount    int64 `json:"availableCount"`
	SearchingCount    int64 `json:"searchingCount"`
	HighPriorityCount int64 `json:"highPriorityCount"`
}
