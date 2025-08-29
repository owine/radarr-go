// Package models defines calendar data structures and event models for Radarr.
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// CalendarEvent represents a calendar event for movie releases and monitoring
type CalendarEvent struct {
	ID               int                 `json:"id" db:"id" gorm:"primaryKey"`
	MovieID          int                 `json:"movieId" db:"movie_id" gorm:"not null;index"`
	Title            string              `json:"title" db:"title" gorm:"not null"`
	OriginalTitle    string              `json:"originalTitle" db:"original_title"`
	EventType        CalendarEventType   `json:"eventType" db:"event_type" gorm:"not null"`
	EventDate        time.Time           `json:"date" db:"event_date" gorm:"not null;index"`
	Status           CalendarEventStatus `json:"status" db:"status" gorm:"not null"`
	Monitored        bool                `json:"monitored" db:"monitored" gorm:"not null"`
	HasFile          bool                `json:"hasFile" db:"has_file"`
	Downloaded       bool                `json:"downloaded" db:"downloaded"`
	Unmonitored      bool                `json:"unmonitored" db:"unmonitored"`
	Year             int                 `json:"year" db:"year" gorm:"index"`
	Runtime          int                 `json:"runtime" db:"runtime"`
	Overview         string              `json:"overview" db:"overview" gorm:"type:text"`
	Images           MediaCover          `json:"images" db:"images" gorm:"type:text"`
	Genres           StringArray         `json:"genres" db:"genres" gorm:"type:text"`
	QualityProfileID int                 `json:"qualityProfileId" db:"quality_profile_id"`
	FolderName       string              `json:"folderName" db:"folder_name"`
	Path             string              `json:"path" db:"path"`
	TmdbID           int                 `json:"tmdbId" db:"tmdb_id" gorm:"index"`
	ImdbID           string              `json:"imdbId" db:"imdb_id" gorm:"index"`
	Tags             IntArray            `json:"tags" db:"tags" gorm:"type:text"`

	// Event-specific metadata
	EventDescription string         `json:"eventDescription" db:"event_description"`
	AllDay           bool           `json:"allDay" db:"all_day"`
	EndDate          *time.Time     `json:"endDate,omitempty" db:"end_date"`
	Location         string         `json:"location" db:"location"`
	Reminder         *time.Duration `json:"reminder,omitempty" db:"reminder"`

	// Associated movie data
	Movie *Movie `json:"movie,omitempty" gorm:"foreignKey:MovieID"`

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// CalendarEventType represents the type of calendar event
type CalendarEventType string

const (
	// CalendarEventCinemaRelease indicates movie is released in cinemas
	CalendarEventCinemaRelease CalendarEventType = "cinemaRelease"
	// CalendarEventPhysicalRelease indicates movie physical release
	CalendarEventPhysicalRelease CalendarEventType = "physicalRelease"
	// CalendarEventDigitalRelease indicates movie digital release
	CalendarEventDigitalRelease CalendarEventType = "digitalRelease"
	// CalendarEventAnnouncement indicates movie announcement
	CalendarEventAnnouncement CalendarEventType = "announcement"
	// CalendarEventMonitoring indicates monitoring status change
	CalendarEventMonitoring CalendarEventType = "monitoring"
	// CalendarEventAvailability indicates availability change
	CalendarEventAvailability CalendarEventType = "availability"
)

// CalendarEventStatus represents the current status of a calendar event
type CalendarEventStatus string

const (
	// CalendarEventStatusUpcoming indicates the event is upcoming
	CalendarEventStatusUpcoming CalendarEventStatus = "upcoming"
	// CalendarEventStatusCurrent indicates the event is happening now
	CalendarEventStatusCurrent CalendarEventStatus = "current"
	// CalendarEventStatusPast indicates the event has passed
	CalendarEventStatusPast CalendarEventStatus = "past"
	// CalendarEventStatusMissing indicates the event date is missing
	CalendarEventStatusMissing CalendarEventStatus = "missing"
	// CalendarEventStatusTBA indicates the event date is to be announced
	CalendarEventStatusTBA CalendarEventStatus = "tba"
)

// CalendarViewType represents different calendar view modes
type CalendarViewType string

const (
	// CalendarViewMonth shows events in monthly view
	CalendarViewMonth CalendarViewType = "month"
	// CalendarViewWeek shows events in weekly view
	CalendarViewWeek CalendarViewType = "week"
	// CalendarViewAgenda shows events in agenda/list view
	CalendarViewAgenda CalendarViewType = "agenda"
	// CalendarViewForecast shows upcoming events forecast
	CalendarViewForecast CalendarViewType = "forecast"
)

// CalendarConfiguration represents calendar settings and preferences
type CalendarConfiguration struct {
	ID                      int                `json:"id" db:"id" gorm:"primaryKey"`
	EnabledEventTypes       CalendarEventTypes `json:"enabledEventTypes" db:"enabled_event_types" gorm:"type:text"`
	DefaultView             CalendarViewType   `json:"defaultView" db:"default_view"`
	FirstDayOfWeek          int                `json:"firstDayOfWeek" db:"first_day_of_week"`
	ShowColoredEvents       bool               `json:"showColoredEvents" db:"show_colored_events"`
	ShowMovieInformation    bool               `json:"showMovieInformation" db:"show_movie_information"`
	FullCalendarEventFilter bool               `json:"fullCalendarEventFilter" db:"full_calendar_event_filter"`
	CollapseMultipleEvents  bool               `json:"collapseMultipleEvents" db:"collapse_multiple_events"`

	// iCal feed settings
	EnableICalFeed   bool     `json:"enableICalFeed" db:"enable_ical_feed"`
	ICalFeedAuth     bool     `json:"iCalFeedAuth" db:"ical_feed_auth"`
	ICalFeedPasskey  string   `json:"iCalFeedPasskey" db:"ical_feed_passkey"`
	ICalDaysInFuture int      `json:"iCalDaysInFuture" db:"ical_days_in_future"`
	ICalDaysInPast   int      `json:"iCalDaysInPast" db:"ical_days_in_past"`
	ICalTags         IntArray `json:"iCalTags" db:"ical_tags" gorm:"type:text"`

	// Event display settings
	EventTitleFormat       string `json:"eventTitleFormat" db:"event_title_format"`
	EventDescriptionFormat string `json:"eventDescriptionFormat" db:"event_description_format"`
	TimeZone               string `json:"timeZone" db:"time_zone"`

	// Caching settings
	EnableEventCaching bool `json:"enableEventCaching" db:"enable_event_caching"`
	EventCacheDuration int  `json:"eventCacheDuration" db:"event_cache_duration"` // minutes

	// Timestamps
	CreatedAt time.Time `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// CalendarEventTypes is a custom type for handling arrays of calendar event types
type CalendarEventTypes []CalendarEventType

// Value implements the driver.Valuer interface for database storage
func (c CalendarEventTypes) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database retrieval
func (c *CalendarEventTypes) Scan(value interface{}) error {
	if value == nil {
		*c = CalendarEventTypes{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// CalendarRequest represents request parameters for calendar data
type CalendarRequest struct {
	Start                   *time.Time          `json:"start,omitempty"`
	End                     *time.Time          `json:"end,omitempty"`
	View                    CalendarViewType    `json:"view,omitempty"`
	EventTypes              []CalendarEventType `json:"eventTypes,omitempty"`
	MovieIDs                []int               `json:"movieIds,omitempty"`
	Tags                    []int               `json:"tags,omitempty"`
	Monitored               *bool               `json:"monitored,omitempty"`
	IncludeUnmonitored      bool                `json:"includeUnmonitored"`
	IncludeMovieInformation bool                `json:"includeMovieInformation"`
}

// CalendarResponse represents the response structure for calendar requests
type CalendarResponse struct {
	Events      []CalendarEvent   `json:"events"`
	Summary     CalendarSummary   `json:"summary"`
	View        CalendarViewType  `json:"view"`
	DateRange   CalendarDateRange `json:"dateRange"`
	TotalEvents int               `json:"totalEvents"`
	Cached      bool              `json:"cached"`
	CacheExpiry *time.Time        `json:"cacheExpiry,omitempty"`
}

// CalendarSummary provides statistics about calendar events
type CalendarSummary struct {
	TotalEvents        int                         `json:"totalEvents"`
	UpcomingEvents     int                         `json:"upcomingEvents"`
	PastEvents         int                         `json:"pastEvents"`
	MonitoredMovies    int                         `json:"monitoredMovies"`
	UnmonitoredMovies  int                         `json:"unmonitoredMovies"`
	MoviesWithFiles    int                         `json:"moviesWithFiles"`
	MoviesWithoutFiles int                         `json:"moviesWithoutFiles"`
	EventsByType       map[CalendarEventType]int   `json:"eventsByType"`
	EventsByStatus     map[CalendarEventStatus]int `json:"eventsByStatus"`
}

// CalendarDateRange represents a date range for calendar queries
type CalendarDateRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// CalendarFeedConfig represents configuration for iCal feed generation
type CalendarFeedConfig struct {
	Title            string              `json:"title"`
	Description      string              `json:"description"`
	TimeZone         string              `json:"timeZone"`
	EventTypes       []CalendarEventType `json:"eventTypes"`
	Tags             []int               `json:"tags"`
	DaysInFuture     int                 `json:"daysInFuture"`
	DaysInPast       int                 `json:"daysInPast"`
	RequireAuth      bool                `json:"requireAuth"`
	PassKey          string              `json:"passKey,omitempty"`
	IncludeMonitored *bool               `json:"includeMonitored,omitempty"`
}

// ICalEvent represents an iCal calendar event
type ICalEvent struct {
	UID         string     `json:"uid"`
	Summary     string     `json:"summary"`
	Description string     `json:"description"`
	Start       time.Time  `json:"start"`
	End         *time.Time `json:"end,omitempty"`
	AllDay      bool       `json:"allDay"`
	Location    string     `json:"location,omitempty"`
	Categories  []string   `json:"categories,omitempty"`
	Status      string     `json:"status"`
	URL         string     `json:"url,omitempty"`
	Created     time.Time  `json:"created"`
	LastMod     time.Time  `json:"lastModified"`
	Sequence    int        `json:"sequence"`
}

// CalendarEventCache represents cached calendar event data
type CalendarEventCache struct {
	ID        string              `json:"id" db:"id" gorm:"primaryKey"`
	CacheKey  string              `json:"cacheKey" db:"cache_key" gorm:"uniqueIndex;not null"`
	Events    CalendarEventsData  `json:"events" db:"events" gorm:"type:text"`
	Summary   CalendarSummaryData `json:"summary" db:"summary" gorm:"type:text"`
	ExpiresAt time.Time           `json:"expiresAt" db:"expires_at" gorm:"not null;index"`
	CreatedAt time.Time           `json:"createdAt" db:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time           `json:"updatedAt" db:"updated_at" gorm:"autoUpdateTime"`
}

// CalendarEventsData is a custom type for handling JSON arrays of calendar events
type CalendarEventsData []CalendarEvent

// Value implements the driver.Valuer interface for database storage
func (c CalendarEventsData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database retrieval
func (c *CalendarEventsData) Scan(value interface{}) error {
	if value == nil {
		*c = CalendarEventsData{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// CalendarSummaryData is a custom type for handling JSON calendar summary data
type CalendarSummaryData CalendarSummary

// Value implements the driver.Valuer interface for database storage
func (c CalendarSummaryData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for database retrieval
func (c *CalendarSummaryData) Scan(value interface{}) error {
	if value == nil {
		*c = CalendarSummaryData{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, c)
}

// GetEventColor returns the color code for different event types
func (e *CalendarEvent) GetEventColor() string {
	switch e.EventType {
	case CalendarEventCinemaRelease:
		return "#007bff" // Blue
	case CalendarEventPhysicalRelease:
		return "#28a745" // Green
	case CalendarEventDigitalRelease:
		return "#17a2b8" // Cyan
	case CalendarEventAnnouncement:
		return "#ffc107" // Yellow
	case CalendarEventMonitoring:
		return "#6c757d" // Gray
	case CalendarEventAvailability:
		return "#dc3545" // Red
	default:
		return "#6c757d" // Default gray
	}
}

// GetEventPriority returns the priority level for event sorting
func (e *CalendarEvent) GetEventPriority() int {
	switch e.EventType {
	case CalendarEventCinemaRelease:
		return 1
	case CalendarEventDigitalRelease:
		return 2
	case CalendarEventPhysicalRelease:
		return 3
	case CalendarEventAvailability:
		return 4
	case CalendarEventAnnouncement:
		return 5
	case CalendarEventMonitoring:
		return 6
	default:
		return 10
	}
}

// IsUpcoming checks if the event is in the future
func (e *CalendarEvent) IsUpcoming() bool {
	return e.EventDate.After(time.Now())
}

// IsPast checks if the event is in the past
func (e *CalendarEvent) IsPast() bool {
	return e.EventDate.Before(time.Now())
}

// GetDisplayTitle returns a formatted title for display purposes
func (e *CalendarEvent) GetDisplayTitle() string {
	if e.OriginalTitle != "" && e.OriginalTitle != e.Title {
		return fmt.Sprintf("%s (%s)", e.Title, e.OriginalTitle)
	}
	return e.Title
}

// GenerateUID generates a unique identifier for iCal events
func (e *CalendarEvent) GenerateUID() string {
	return fmt.Sprintf("radarr-event-%d-%s@radarr-go", e.MovieID, e.EventType)
}

// ToICalEvent converts a calendar event to an iCal event format
func (e *CalendarEvent) ToICalEvent(baseURL string) ICalEvent {
	uid := e.GenerateUID()
	summary := e.GetDisplayTitle()

	if e.EventType != CalendarEventCinemaRelease {
		summary = fmt.Sprintf("%s - %s", summary, string(e.EventType))
	}

	description := e.EventDescription
	if description == "" {
		description = e.Overview
	}

	var categories []string
	switch e.EventType {
	case CalendarEventCinemaRelease:
		categories = append(categories, "Cinema Release", "Movies")
	case CalendarEventPhysicalRelease:
		categories = append(categories, "Physical Release", "Movies")
	case CalendarEventDigitalRelease:
		categories = append(categories, "Digital Release", "Movies")
	case CalendarEventAnnouncement:
		categories = append(categories, "Announcement", "Movies")
	}

	var url string
	if baseURL != "" {
		url = fmt.Sprintf("%s/movie/%d", baseURL, e.MovieID)
	}

	status := "CONFIRMED"
	if e.Status == CalendarEventStatusTBA || e.Status == CalendarEventStatusMissing {
		status = "TENTATIVE"
	}

	endDate := e.EndDate
	if e.AllDay && endDate == nil {
		nextDay := e.EventDate.AddDate(0, 0, 1)
		endDate = &nextDay
	}

	return ICalEvent{
		UID:         uid,
		Summary:     summary,
		Description: description,
		Start:       e.EventDate,
		End:         endDate,
		AllDay:      e.AllDay,
		Location:    e.Location,
		Categories:  categories,
		Status:      status,
		URL:         url,
		Created:     e.CreatedAt,
		LastMod:     e.UpdatedAt,
		Sequence:    0,
	}
}

// DefaultCalendarConfiguration returns default calendar configuration
func DefaultCalendarConfiguration() *CalendarConfiguration {
	return &CalendarConfiguration{
		EnabledEventTypes: CalendarEventTypes{
			CalendarEventCinemaRelease,
			CalendarEventPhysicalRelease,
			CalendarEventDigitalRelease,
		},
		DefaultView:             CalendarViewMonth,
		FirstDayOfWeek:          0, // Sunday
		ShowColoredEvents:       true,
		ShowMovieInformation:    true,
		FullCalendarEventFilter: false,
		CollapseMultipleEvents:  false,
		EnableICalFeed:          true,
		ICalFeedAuth:            false,
		ICalDaysInFuture:        365,
		ICalDaysInPast:          30,
		EventTitleFormat:        "{Movie Title}",
		EventDescriptionFormat:  "{Movie Overview}",
		TimeZone:                "UTC",
		EnableEventCaching:      true,
		EventCacheDuration:      60, // 1 hour
	}
}
