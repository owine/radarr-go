// Package services provides iCal feed generation functionality for calendar integration.
package services

import (
	"fmt"
	"strings"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// ICalService provides iCal feed generation functionality
type ICalService struct {
	db              *database.Database
	logger          *logger.Logger
	calendarService *CalendarService
}

// NewICalService creates a new iCal service instance
func NewICalService(db *database.Database, logger *logger.Logger, calendarService *CalendarService) *ICalService {
	return &ICalService{
		db:              db,
		logger:          logger,
		calendarService: calendarService,
	}
}

// GenerateICalFeed generates an iCal feed based on the provided configuration
func (s *ICalService) GenerateICalFeed(config *models.CalendarFeedConfig, baseURL string) (string, error) {
	// Create calendar request based on feed configuration
	request := s.createCalendarRequestFromConfig(config)

	// Get calendar events
	response, err := s.calendarService.GetCalendarEvents(request)
	if err != nil {
		return "", fmt.Errorf("failed to get calendar events: %w", err)
	}

	// Convert events to iCal format
	var icalBuilder strings.Builder

	// Write iCal header
	s.writeICalHeader(&icalBuilder, config)

	// Write events
	for _, event := range response.Events {
		icalEvent := event.ToICalEvent(baseURL)
		s.writeICalEvent(&icalBuilder, &icalEvent)
	}

	// Write iCal footer
	s.writeICalFooter(&icalBuilder)

	return icalBuilder.String(), nil
}

// createCalendarRequestFromConfig converts feed config to calendar request
func (s *ICalService) createCalendarRequestFromConfig(config *models.CalendarFeedConfig) *models.CalendarRequest {
	now := time.Now()
	start := now.AddDate(0, 0, -config.DaysInPast)
	end := now.AddDate(0, 0, config.DaysInFuture)

	request := &models.CalendarRequest{
		Start:                   &start,
		End:                     &end,
		EventTypes:              config.EventTypes,
		Tags:                    config.Tags,
		Monitored:               config.IncludeMonitored,
		IncludeUnmonitored:      config.IncludeMonitored == nil || !*config.IncludeMonitored,
		IncludeMovieInformation: true,
	}

	return request
}

// writeICalHeader writes the iCal calendar header
func (s *ICalService) writeICalHeader(builder *strings.Builder, config *models.CalendarFeedConfig) {
	now := time.Now().UTC()

	builder.WriteString("BEGIN:VCALENDAR\r\n")
	builder.WriteString("VERSION:2.0\r\n")
	builder.WriteString("PRODID:-//Radarr Go//Radarr Go Calendar//EN\r\n")
	builder.WriteString("CALSCALE:GREGORIAN\r\n")
	builder.WriteString("METHOD:PUBLISH\r\n")

	// Calendar properties
	builder.WriteString(fmt.Sprintf("X-WR-CALNAME:%s\r\n", s.escapeText(config.Title)))
	builder.WriteString(fmt.Sprintf("X-WR-CALDESC:%s\r\n", s.escapeText(config.Description)))
	builder.WriteString(fmt.Sprintf("X-WR-TIMEZONE:%s\r\n", config.TimeZone))
	builder.WriteString("X-WR-REFRESHINTERVAL:PT1H\r\n") // Refresh every hour

	// Publish information
	builder.WriteString(fmt.Sprintf("X-PUBLISHED-TTL:PT1H\r\n"))
	builder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", s.formatDateTime(now)))

	// Time zone information (if not UTC)
	if config.TimeZone != "UTC" {
		s.writeTimeZoneInfo(builder, config.TimeZone)
	}
}

// writeTimeZoneInfo writes time zone information to the iCal feed
func (s *ICalService) writeTimeZoneInfo(builder *strings.Builder, timeZone string) {
	// For simplicity, we'll use UTC. In a production system,
	// you would want to include proper VTIMEZONE components.
	// For now, we'll note that all times are in UTC.
	builder.WriteString("BEGIN:VTIMEZONE\r\n")
	builder.WriteString(fmt.Sprintf("TZID:%s\r\n", timeZone))
	builder.WriteString("BEGIN:STANDARD\r\n")
	builder.WriteString("DTSTART:19700101T000000\r\n")
	builder.WriteString("TZOFFSETFROM:+0000\r\n")
	builder.WriteString("TZOFFSETTO:+0000\r\n")
	builder.WriteString("TZNAME:UTC\r\n")
	builder.WriteString("END:STANDARD\r\n")
	builder.WriteString("END:VTIMEZONE\r\n")
}

// writeICalEvent writes a single event to the iCal feed
func (s *ICalService) writeICalEvent(builder *strings.Builder, event *models.ICalEvent) {
	builder.WriteString("BEGIN:VEVENT\r\n")

	// Required properties
	builder.WriteString(fmt.Sprintf("UID:%s\r\n", s.escapeText(event.UID)))
	builder.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", s.formatDateTime(event.Created)))

	// Event dates
	if event.AllDay {
		builder.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", s.formatDate(event.Start)))
		if event.End != nil {
			builder.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", s.formatDate(*event.End)))
		}
	} else {
		builder.WriteString(fmt.Sprintf("DTSTART:%s\r\n", s.formatDateTime(event.Start)))
		if event.End != nil {
			builder.WriteString(fmt.Sprintf("DTEND:%s\r\n", s.formatDateTime(*event.End)))
		}
	}

	// Event properties
	builder.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", s.escapeText(event.Summary)))

	if event.Description != "" {
		description := s.wrapText(s.escapeText(event.Description), 75)
		builder.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", description))
	}

	if event.Location != "" {
		builder.WriteString(fmt.Sprintf("LOCATION:%s\r\n", s.escapeText(event.Location)))
	}

	if event.URL != "" {
		builder.WriteString(fmt.Sprintf("URL:%s\r\n", event.URL))
	}

	// Categories
	if len(event.Categories) > 0 {
		categories := strings.Join(event.Categories, ",")
		builder.WriteString(fmt.Sprintf("CATEGORIES:%s\r\n", s.escapeText(categories)))
	}

	// Status
	builder.WriteString(fmt.Sprintf("STATUS:%s\r\n", event.Status))

	// Transparency (show as busy/free)
	if event.AllDay {
		builder.WriteString("TRANSP:TRANSPARENT\r\n") // Show as free time
	} else {
		builder.WriteString("TRANSP:OPAQUE\r\n") // Show as busy time
	}

	// Last modified
	builder.WriteString(fmt.Sprintf("LAST-MODIFIED:%s\r\n", s.formatDateTime(event.LastMod)))

	// Sequence
	builder.WriteString(fmt.Sprintf("SEQUENCE:%d\r\n", event.Sequence))

	// Priority (1-9, with 1 being highest)
	builder.WriteString("PRIORITY:5\r\n") // Medium priority

	// Classification
	builder.WriteString("CLASS:PUBLIC\r\n")

	builder.WriteString("END:VEVENT\r\n")
}

// writeICalFooter writes the iCal calendar footer
func (s *ICalService) writeICalFooter(builder *strings.Builder) {
	builder.WriteString("END:VCALENDAR\r\n")
}

// formatDateTime formats a time in iCal format (UTC)
func (s *ICalService) formatDateTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// formatDate formats a date in iCal date format
func (s *ICalService) formatDate(t time.Time) string {
	return t.Format("20060102")
}

// escapeText escapes text for iCal format
func (s *ICalService) escapeText(text string) string {
	// RFC 5545 text escaping
	text = strings.ReplaceAll(text, "\\", "\\\\") // Escape backslashes first
	text = strings.ReplaceAll(text, ";", "\\;")   // Escape semicolons
	text = strings.ReplaceAll(text, ",", "\\,")   // Escape commas
	text = strings.ReplaceAll(text, "\n", "\\n")  // Escape newlines
	text = strings.ReplaceAll(text, "\r", "")     // Remove carriage returns

	return text
}

// wrapText wraps text at the specified length for iCal format
func (s *ICalService) wrapText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	var result strings.Builder
	remaining := text

	for len(remaining) > maxLen {
		// Find the best place to break the line
		breakPoint := maxLen

		// Try to break at a space, but not too far back
		for i := maxLen; i > maxLen-20 && i > 0; i-- {
			if remaining[i] == ' ' {
				breakPoint = i
				break
			}
		}

		result.WriteString(remaining[:breakPoint])
		result.WriteString("\r\n ") // Line continuation in iCal
		remaining = remaining[breakPoint:]

		// Skip leading space on continued line
		if len(remaining) > 0 && remaining[0] == ' ' {
			remaining = remaining[1:]
		}
	}

	result.WriteString(remaining)
	return result.String()
}

// ValidateICalFeedConfig validates the iCal feed configuration
func (s *ICalService) ValidateICalFeedConfig(config *models.CalendarFeedConfig) error {
	if config.Title == "" {
		return fmt.Errorf("feed title is required")
	}

	if config.DaysInFuture < 0 || config.DaysInFuture > 3650 { // Max 10 years
		return fmt.Errorf("days in future must be between 0 and 3650")
	}

	if config.DaysInPast < 0 || config.DaysInPast > 3650 { // Max 10 years
		return fmt.Errorf("days in past must be between 0 and 3650")
	}

	if config.TimeZone == "" {
		config.TimeZone = "UTC"
	}

	// Validate event types
	if len(config.EventTypes) == 0 {
		config.EventTypes = []models.CalendarEventType{
			models.CalendarEventCinemaRelease,
			models.CalendarEventPhysicalRelease,
			models.CalendarEventDigitalRelease,
		}
	}

	return nil
}

// GetDefaultFeedConfig returns a default iCal feed configuration
func (s *ICalService) GetDefaultFeedConfig() *models.CalendarFeedConfig {
	return &models.CalendarFeedConfig{
		Title:       "Radarr Movie Calendar",
		Description: "Movie release dates from Radarr",
		TimeZone:    "UTC",
		EventTypes: []models.CalendarEventType{
			models.CalendarEventCinemaRelease,
			models.CalendarEventPhysicalRelease,
			models.CalendarEventDigitalRelease,
		},
		Tags:             []int{},
		DaysInFuture:     365,
		DaysInPast:       30,
		RequireAuth:      false,
		IncludeMonitored: nil, // Include both monitored and unmonitored
	}
}

// GenerateICalFeedURL generates a URL for accessing the iCal feed
func (s *ICalService) GenerateICalFeedURL(baseURL string, config *models.CalendarFeedConfig) string {
	url := fmt.Sprintf("%s/api/v3/calendar/feed.ics", baseURL)

	var params []string

	if len(config.EventTypes) > 0 {
		var eventTypes []string
		for _, eventType := range config.EventTypes {
			eventTypes = append(eventTypes, string(eventType))
		}
		params = append(params, fmt.Sprintf("eventTypes=%s", strings.Join(eventTypes, ",")))
	}

	if len(config.Tags) > 0 {
		var tags []string
		for _, tag := range config.Tags {
			tags = append(tags, fmt.Sprintf("%d", tag))
		}
		params = append(params, fmt.Sprintf("tags=%s", strings.Join(tags, ",")))
	}

	if config.DaysInFuture != 365 {
		params = append(params, fmt.Sprintf("daysInFuture=%d", config.DaysInFuture))
	}

	if config.DaysInPast != 30 {
		params = append(params, fmt.Sprintf("daysInPast=%d", config.DaysInPast))
	}

	if config.IncludeMonitored != nil {
		params = append(params, fmt.Sprintf("monitored=%t", *config.IncludeMonitored))
	}

	if config.RequireAuth && config.PassKey != "" {
		params = append(params, fmt.Sprintf("passKey=%s", config.PassKey))
	}

	if len(params) > 0 {
		url += "?" + strings.Join(params, "&")
	}

	return url
}

// ParseICalFeedParams parses URL parameters for iCal feed configuration
func (s *ICalService) ParseICalFeedParams(params map[string]string) *models.CalendarFeedConfig {
	config := s.GetDefaultFeedConfig()

	if title := params["title"]; title != "" {
		config.Title = title
	}

	if description := params["description"]; description != "" {
		config.Description = description
	}

	if timeZone := params["timeZone"]; timeZone != "" {
		config.TimeZone = timeZone
	}

	if eventTypesStr := params["eventTypes"]; eventTypesStr != "" {
		eventTypeStrs := strings.Split(eventTypesStr, ",")
		var eventTypes []models.CalendarEventType
		for _, eventTypeStr := range eventTypeStrs {
			eventTypes = append(eventTypes, models.CalendarEventType(strings.TrimSpace(eventTypeStr)))
		}
		config.EventTypes = eventTypes
	}

	if tagsStr := params["tags"]; tagsStr != "" {
		tagStrs := strings.Split(tagsStr, ",")
		var tags []int
		for _, tagStr := range tagStrs {
			if tag := parseInt(strings.TrimSpace(tagStr)); tag != 0 {
				tags = append(tags, tag)
			}
		}
		config.Tags = tags
	}

	if daysInFutureStr := params["daysInFuture"]; daysInFutureStr != "" {
		if days := parseInt(daysInFutureStr); days > 0 {
			config.DaysInFuture = days
		}
	}

	if daysInPastStr := params["daysInPast"]; daysInPastStr != "" {
		if days := parseInt(daysInPastStr); days >= 0 {
			config.DaysInPast = days
		}
	}

	if monitoredStr := params["monitored"]; monitoredStr != "" {
		monitored := strings.ToLower(monitoredStr) == "true"
		config.IncludeMonitored = &monitored
	}

	if passKey := params["passKey"]; passKey != "" {
		config.PassKey = passKey
		config.RequireAuth = true
	}

	return config
}

// Helper function to parse integer from string
func parseInt(s string) int {
	var result int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		result = result*10 + int(r-'0')
	}
	return result
}
