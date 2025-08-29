// Package services provides calendar functionality for tracking movie release dates and events.
package services

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// CalendarService provides calendar functionality for movie release tracking
type CalendarService struct {
	db     *database.Database
	logger *logger.Logger
}

// NewCalendarService creates a new calendar service instance
func NewCalendarService(db *database.Database, logger *logger.Logger) *CalendarService {
	return &CalendarService{
		db:     db,
		logger: logger,
	}
}

// GetCalendarEvents retrieves calendar events based on the provided request parameters
func (s *CalendarService) GetCalendarEvents(request *models.CalendarRequest) (*models.CalendarResponse, error) {
	// Set default date range if not provided
	if request.Start == nil || request.End == nil {
		request = s.setDefaultDateRange(request)
	}

	// Check cache first if caching is enabled
	cacheKey := s.generateCacheKey(request)
	if cached, err := s.getCachedEvents(cacheKey); err == nil && cached != nil {
		s.logger.Debug("Returning cached calendar events", "cacheKey", cacheKey)
		return cached, nil
	}

	// Generate events from movies
	events, err := s.generateCalendarEvents(request)
	if err != nil {
		return nil, fmt.Errorf("failed to generate calendar events: %w", err)
	}

	// Apply filters
	events = s.filterEvents(events, request)

	// Sort events by date and priority
	s.sortEvents(events)

	// Generate summary
	summary := s.generateEventSummary(events)

	response := &models.CalendarResponse{
		Events:      events,
		Summary:     summary,
		View:        request.View,
		DateRange:   models.CalendarDateRange{Start: *request.Start, End: *request.End},
		TotalEvents: len(events),
		Cached:      false,
	}

	// Cache the response if caching is enabled
	if err := s.cacheEvents(cacheKey, response); err != nil {
		s.logger.Warn("Failed to cache calendar events", "error", err)
	}

	return response, nil
}

// generateCalendarEvents creates calendar events from movies in the database
func (s *CalendarService) generateCalendarEvents(request *models.CalendarRequest) ([]models.CalendarEvent, error) {
	var movies []models.Movie

	query := s.db.GORM.Model(&models.Movie{}).
		Where("created_at IS NOT NULL")

	// Apply movie ID filter
	if len(request.MovieIDs) > 0 {
		query = query.Where("id IN (?)", request.MovieIDs)
	}

	// Apply tags filter
	if len(request.Tags) > 0 {
		query = query.Where("JSON_OVERLAPS(tags, ?)", request.Tags)
	}

	// Apply monitored filter
	if request.Monitored != nil {
		query = query.Where("monitored = ?", *request.Monitored)
	} else if !request.IncludeUnmonitored {
		query = query.Where("monitored = ?", true)
	}

	if err := query.Find(&movies).Error; err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}

	var events []models.CalendarEvent
	now := time.Now()

	for _, movie := range movies {
		// Generate events based on movie release dates
		movieEvents := s.generateEventsForMovie(&movie, request, now)
		events = append(events, movieEvents...)
	}

	return events, nil
}

// generateEventsForMovie creates calendar events for a specific movie
func (s *CalendarService) generateEventsForMovie(movie *models.Movie, request *models.CalendarRequest, now time.Time) []models.CalendarEvent {
	var events []models.CalendarEvent

	// Cinema release event
	if movie.InCinemas != nil && s.shouldIncludeEventType(models.CalendarEventCinemaRelease, request) {
		if s.isDateInRange(*movie.InCinemas, request) {
			event := s.createCalendarEvent(movie, models.CalendarEventCinemaRelease, *movie.InCinemas)
			events = append(events, event)
		}
	}

	// Physical release event
	if movie.PhysicalRelease != nil && s.shouldIncludeEventType(models.CalendarEventPhysicalRelease, request) {
		if s.isDateInRange(*movie.PhysicalRelease, request) {
			event := s.createCalendarEvent(movie, models.CalendarEventPhysicalRelease, *movie.PhysicalRelease)
			events = append(events, event)
		}
	}

	// Digital release event
	if movie.DigitalRelease != nil && s.shouldIncludeEventType(models.CalendarEventDigitalRelease, request) {
		if s.isDateInRange(*movie.DigitalRelease, request) {
			event := s.createCalendarEvent(movie, models.CalendarEventDigitalRelease, *movie.DigitalRelease)
			events = append(events, event)
		}
	}

	// Availability event (based on minimum availability and current status)
	if s.shouldIncludeEventType(models.CalendarEventAvailability, request) {
		if availabilityDate := s.calculateAvailabilityDate(movie, now); availabilityDate != nil {
			if s.isDateInRange(*availabilityDate, request) {
				event := s.createCalendarEvent(movie, models.CalendarEventAvailability, *availabilityDate)
				events = append(events, event)
			}
		}
	}

	return events
}

// createCalendarEvent creates a calendar event from a movie and event type
func (s *CalendarService) createCalendarEvent(movie *models.Movie, eventType models.CalendarEventType, eventDate time.Time) models.CalendarEvent {
	status := s.determineEventStatus(eventDate, movie)

	event := models.CalendarEvent{
		MovieID:          movie.ID,
		Title:            movie.Title,
		OriginalTitle:    movie.OriginalTitle,
		EventType:        eventType,
		EventDate:        eventDate,
		Status:           status,
		Monitored:        movie.Monitored,
		HasFile:          movie.HasFile,
		Downloaded:       movie.HasFile,
		Unmonitored:      !movie.Monitored,
		Year:             movie.Year,
		Runtime:          movie.Runtime,
		Overview:         movie.Overview,
		Images:           movie.Images,
		Genres:           movie.Genres,
		QualityProfileID: movie.QualityProfileID,
		FolderName:       movie.FolderName,
		Path:             movie.Path,
		TmdbID:           movie.TmdbID,
		ImdbID:           movie.ImdbID,
		Tags:             movie.Tags,
		EventDescription: s.generateEventDescription(movie, eventType),
		AllDay:           s.isAllDayEvent(eventType),
		Movie:            movie,
	}

	return event
}

// generateEventDescription creates a description for the calendar event
func (s *CalendarService) generateEventDescription(movie *models.Movie, eventType models.CalendarEventType) string {
	var description strings.Builder

	switch eventType {
	case models.CalendarEventCinemaRelease:
		description.WriteString(fmt.Sprintf("%s is released in cinemas", movie.Title))
	case models.CalendarEventPhysicalRelease:
		description.WriteString(fmt.Sprintf("%s physical release", movie.Title))
		if movie.PhysicalReleaseNote != "" {
			description.WriteString(fmt.Sprintf(" - %s", movie.PhysicalReleaseNote))
		}
	case models.CalendarEventDigitalRelease:
		description.WriteString(fmt.Sprintf("%s digital release", movie.Title))
	case models.CalendarEventAvailability:
		description.WriteString(fmt.Sprintf("%s becomes available for download", movie.Title))
	case models.CalendarEventAnnouncement:
		description.WriteString(fmt.Sprintf("%s announcement", movie.Title))
	case models.CalendarEventMonitoring:
		if movie.Monitored {
			description.WriteString(fmt.Sprintf("%s is being monitored", movie.Title))
		} else {
			description.WriteString(fmt.Sprintf("%s monitoring disabled", movie.Title))
		}
	}

	if movie.Overview != "" && len(movie.Overview) < 200 {
		description.WriteString(fmt.Sprintf("\n\n%s", movie.Overview))
	}

	return description.String()
}

// determineEventStatus determines the status of an event based on its date and movie status
func (s *CalendarService) determineEventStatus(eventDate time.Time, movie *models.Movie) models.CalendarEventStatus {
	now := time.Now()

	// Check for missing or TBA dates
	if eventDate.IsZero() {
		return models.CalendarEventStatusMissing
	}

	// Check if date is in the future
	if eventDate.After(now) {
		return models.CalendarEventStatusUpcoming
	}

	// Check if date is today
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	eventDay := time.Date(eventDate.Year(), eventDate.Month(), eventDate.Day(), 0, 0, 0, 0, eventDate.Location())

	if eventDay.Equal(today) {
		return models.CalendarEventStatusCurrent
	}

	// Date is in the past
	return models.CalendarEventStatusPast
}

// calculateAvailabilityDate calculates when a movie becomes available for download
func (s *CalendarService) calculateAvailabilityDate(movie *models.Movie, now time.Time) *time.Time {
	switch movie.MinimumAvailability {
	case models.AvailabilityAnnounced:
		return &movie.Added
	case models.AvailabilityInCinemas:
		return movie.InCinemas
	case models.AvailabilityReleased:
		if movie.PhysicalRelease != nil {
			return movie.PhysicalRelease
		}
		return movie.DigitalRelease
	case models.AvailabilityPreDB:
		// For PreDB, assume available 1 week before physical release
		if movie.PhysicalRelease != nil {
			preDBDate := movie.PhysicalRelease.AddDate(0, 0, -7)
			return &preDBDate
		}
	}
	return nil
}

// shouldIncludeEventType checks if the event type should be included based on request filters
func (s *CalendarService) shouldIncludeEventType(eventType models.CalendarEventType, request *models.CalendarRequest) bool {
	if len(request.EventTypes) == 0 {
		return true // Include all event types if none specified
	}

	for _, allowedType := range request.EventTypes {
		if eventType == allowedType {
			return true
		}
	}
	return false
}

// isDateInRange checks if a date falls within the requested date range
func (s *CalendarService) isDateInRange(date time.Time, request *models.CalendarRequest) bool {
	if request.Start != nil && date.Before(*request.Start) {
		return false
	}
	if request.End != nil && date.After(*request.End) {
		return false
	}
	return true
}

// isAllDayEvent determines if an event type should be displayed as all-day
func (s *CalendarService) isAllDayEvent(eventType models.CalendarEventType) bool {
	switch eventType {
	case models.CalendarEventCinemaRelease, models.CalendarEventPhysicalRelease, models.CalendarEventDigitalRelease:
		return true
	default:
		return false
	}
}

// filterEvents applies additional filters to the events list
func (s *CalendarService) filterEvents(events []models.CalendarEvent, request *models.CalendarRequest) []models.CalendarEvent {
	if len(events) == 0 {
		return events
	}

	var filtered []models.CalendarEvent

	for _, event := range events {
		// Additional filtering logic can be added here
		if s.passesFilters(&event, request) {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// passesFilters checks if an event passes all the applied filters
func (s *CalendarService) passesFilters(event *models.CalendarEvent, request *models.CalendarRequest) bool {
	// Add any additional filter logic here
	return true
}

// sortEvents sorts events by date and priority
func (s *CalendarService) sortEvents(events []models.CalendarEvent) {
	sort.Slice(events, func(i, j int) bool {
		// First sort by date
		if !events[i].EventDate.Equal(events[j].EventDate) {
			return events[i].EventDate.Before(events[j].EventDate)
		}

		// Then sort by priority (lower number = higher priority)
		iPriority := events[i].GetEventPriority()
		jPriority := events[j].GetEventPriority()
		if iPriority != jPriority {
			return iPriority < jPriority
		}

		// Finally sort by movie title
		return events[i].Title < events[j].Title
	})
}

// generateEventSummary creates a summary of calendar events
func (s *CalendarService) generateEventSummary(events []models.CalendarEvent) models.CalendarSummary {
	summary := models.CalendarSummary{
		TotalEvents:    len(events),
		EventsByType:   make(map[models.CalendarEventType]int),
		EventsByStatus: make(map[models.CalendarEventStatus]int),
	}

	now := time.Now()
	monitoredMovies := make(map[int]bool)
	moviesWithFiles := make(map[int]bool)

	for _, event := range events {
		// Count events by type
		summary.EventsByType[event.EventType]++

		// Count events by status
		summary.EventsByStatus[event.Status]++

		// Count upcoming and past events
		if event.EventDate.After(now) {
			summary.UpcomingEvents++
		} else {
			summary.PastEvents++
		}

		// Track unique movies
		if event.Monitored {
			monitoredMovies[event.MovieID] = true
		} else {
			monitoredMovies[event.MovieID] = false
		}

		if event.HasFile {
			moviesWithFiles[event.MovieID] = true
		}
	}

	// Count movie statistics
	for movieID, monitored := range monitoredMovies {
		if monitored {
			summary.MonitoredMovies++
		} else {
			summary.UnmonitoredMovies++
		}

		if moviesWithFiles[movieID] {
			summary.MoviesWithFiles++
		} else {
			summary.MoviesWithoutFiles++
		}
	}

	return summary
}

// setDefaultDateRange sets default start and end dates if not provided
func (s *CalendarService) setDefaultDateRange(request *models.CalendarRequest) *models.CalendarRequest {
	now := time.Now()

	if request.Start == nil {
		start := now.AddDate(0, 0, -30) // 30 days ago
		request.Start = &start
	}

	if request.End == nil {
		end := now.AddDate(0, 0, 90) // 90 days from now
		request.End = &end
	}

	return request
}

// generateCacheKey creates a cache key for the calendar request
func (s *CalendarService) generateCacheKey(request *models.CalendarRequest) string {
	data, _ := json.Marshal(request)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// getCachedEvents retrieves cached calendar events
func (s *CalendarService) getCachedEvents(cacheKey string) (*models.CalendarResponse, error) {
	var cached models.CalendarEventCache

	if err := s.db.GORM.Where("cache_key = ? AND expires_at > ?", cacheKey, time.Now()).
		First(&cached).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Cache miss, not an error
		}
		return nil, fmt.Errorf("failed to query cache: %w", err)
	}

	response := &models.CalendarResponse{
		Events:      []models.CalendarEvent(cached.Events),
		Summary:     models.CalendarSummary(cached.Summary),
		TotalEvents: len(cached.Events),
		Cached:      true,
		CacheExpiry: &cached.ExpiresAt,
	}

	return response, nil
}

// cacheEvents stores calendar events in the cache
func (s *CalendarService) cacheEvents(cacheKey string, response *models.CalendarResponse) error {
	config, err := s.GetCalendarConfiguration()
	if err != nil || !config.EnableEventCaching {
		return nil // Caching disabled
	}

	expiresAt := time.Now().Add(time.Duration(config.EventCacheDuration) * time.Minute)

	cache := models.CalendarEventCache{
		ID:        cacheKey,
		CacheKey:  cacheKey,
		Events:    models.CalendarEventsData(response.Events),
		Summary:   models.CalendarSummaryData(response.Summary),
		ExpiresAt: expiresAt,
	}

	// Use UPSERT to handle cache updates
	if err := s.db.GORM.Save(&cache).Error; err != nil {
		return fmt.Errorf("failed to cache events: %w", err)
	}

	return nil
}

// ClearExpiredCache removes expired cache entries
func (s *CalendarService) ClearExpiredCache() error {
	result := s.db.GORM.Where("expires_at < ?", time.Now()).
		Delete(&models.CalendarEventCache{})

	if result.Error != nil {
		return fmt.Errorf("failed to clear expired cache: %w", result.Error)
	}

	if result.RowsAffected > 0 {
		s.logger.Info("Cleared expired calendar cache entries", "count", result.RowsAffected)
	}

	return nil
}

// GetCalendarConfiguration retrieves the calendar configuration
func (s *CalendarService) GetCalendarConfiguration() (*models.CalendarConfiguration, error) {
	var config models.CalendarConfiguration

	if err := s.db.GORM.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default configuration if none exists
			defaultConfig := models.DefaultCalendarConfiguration()
			if err := s.db.GORM.Create(defaultConfig).Error; err != nil {
				return nil, fmt.Errorf("failed to create default calendar configuration: %w", err)
			}
			return defaultConfig, nil
		}
		return nil, fmt.Errorf("failed to get calendar configuration: %w", err)
	}

	return &config, nil
}

// UpdateCalendarConfiguration updates the calendar configuration
func (s *CalendarService) UpdateCalendarConfiguration(config *models.CalendarConfiguration) error {
	if err := s.db.GORM.Save(config).Error; err != nil {
		return fmt.Errorf("failed to update calendar configuration: %w", err)
	}

	s.logger.Info("Calendar configuration updated", "configId", config.ID)
	return nil
}

// GetCalendarStats returns statistics about calendar events
func (s *CalendarService) GetCalendarStats() (map[string]interface{}, error) {
	now := time.Now()

	// Count total movies
	var totalMovies int64
	if err := s.db.GORM.Model(&models.Movie{}).Count(&totalMovies).Error; err != nil {
		return nil, fmt.Errorf("failed to count movies: %w", err)
	}

	// Count monitored movies
	var monitoredMovies int64
	if err := s.db.GORM.Model(&models.Movie{}).Where("monitored = ?", true).Count(&monitoredMovies).Error; err != nil {
		return nil, fmt.Errorf("failed to count monitored movies: %w", err)
	}

	// Count movies with files
	var moviesWithFiles int64
	if err := s.db.GORM.Model(&models.Movie{}).Where("has_file = ?", true).Count(&moviesWithFiles).Error; err != nil {
		return nil, fmt.Errorf("failed to count movies with files: %w", err)
	}

	// Count upcoming releases (next 30 days)
	futureDate := now.AddDate(0, 0, 30)
	var upcomingInCinemas int64
	s.db.GORM.Model(&models.Movie{}).
		Where("in_cinemas > ? AND in_cinemas <= ? AND monitored = ?", now, futureDate, true).
		Count(&upcomingInCinemas)

	var upcomingPhysical int64
	s.db.GORM.Model(&models.Movie{}).
		Where("physical_release > ? AND physical_release <= ? AND monitored = ?", now, futureDate, true).
		Count(&upcomingPhysical)

	var upcomingDigital int64
	s.db.GORM.Model(&models.Movie{}).
		Where("digital_release > ? AND digital_release <= ? AND monitored = ?", now, futureDate, true).
		Count(&upcomingDigital)

	stats := map[string]interface{}{
		"totalMovies":        totalMovies,
		"monitoredMovies":    monitoredMovies,
		"unmonitoredMovies":  totalMovies - monitoredMovies,
		"moviesWithFiles":    moviesWithFiles,
		"moviesWithoutFiles": totalMovies - moviesWithFiles,
		"upcomingInCinemas":  upcomingInCinemas,
		"upcomingPhysical":   upcomingPhysical,
		"upcomingDigital":    upcomingDigital,
		"upcomingTotal":      upcomingInCinemas + upcomingPhysical + upcomingDigital,
	}

	return stats, nil
}

// RefreshCalendarEvents forces a refresh of calendar events and clears cache
func (s *CalendarService) RefreshCalendarEvents() error {
	// Clear all cached events
	if err := s.db.GORM.Delete(&models.CalendarEventCache{}, "1=1").Error; err != nil {
		s.logger.Warn("Failed to clear calendar cache during refresh", "error", err)
	}

	s.logger.Info("Calendar events refreshed and cache cleared")
	return nil
}
