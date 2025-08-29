package services

import (
	"strings"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/models"
)

func TestICalServiceGetDefaultFeedConfig(t *testing.T) {
	icalService := &ICalService{}
	config := icalService.GetDefaultFeedConfig()

	if config.Title != "Radarr Movie Calendar" {
		t.Errorf("Expected default title 'Radarr Movie Calendar', got '%s'", config.Title)
	}

	if config.DaysInFuture != 365 {
		t.Errorf("Expected default days in future to be 365, got %d", config.DaysInFuture)
	}

	if config.DaysInPast != 30 {
		t.Errorf("Expected default days in past to be 30, got %d", config.DaysInPast)
	}

	if len(config.EventTypes) != 3 {
		t.Errorf("Expected 3 default event types, got %d", len(config.EventTypes))
	}
}

func TestICalServiceValidateConfig(t *testing.T) {
	icalService := &ICalService{}

	// Test valid config
	validConfig := &models.CalendarFeedConfig{
		Title:        "Test Calendar",
		DaysInFuture: 365,
		DaysInPast:   30,
		TimeZone:     "UTC",
		EventTypes:   []models.CalendarEventType{models.CalendarEventCinemaRelease},
	}

	if err := icalService.ValidateICalFeedConfig(validConfig); err != nil {
		t.Errorf("Expected valid config to pass validation, got error: %v", err)
	}

	// Test config with empty title
	invalidConfig := &models.CalendarFeedConfig{
		Title: "",
	}

	if err := icalService.ValidateICalFeedConfig(invalidConfig); err == nil {
		t.Error("Expected config with empty title to fail validation")
	}

	// Test config with invalid days
	invalidDaysConfig := &models.CalendarFeedConfig{
		Title:        "Test Calendar",
		DaysInFuture: -1,
	}

	if err := icalService.ValidateICalFeedConfig(invalidDaysConfig); err == nil {
		t.Error("Expected config with negative days to fail validation")
	}
}

func TestICalServiceEscapeText(t *testing.T) {
	icalService := &ICalService{}

	testCases := []struct {
		input    string
		expected string
	}{
		{"Simple text", "Simple text"},
		{"Text with; semicolon", "Text with\\; semicolon"},
		{"Text with, comma", "Text with\\, comma"},
		{"Text with\\backslash", "Text with\\\\backslash"},
		{"Text with\nnewline", "Text with\\nnewline"},
		{"Text with\r\nCRLF", "Text with\\nCRLF"},
	}

	for _, tc := range testCases {
		result := icalService.escapeText(tc.input)
		if result != tc.expected {
			t.Errorf("Expected escaped text '%s', got '%s'", tc.expected, result)
		}
	}
}

func TestICalServiceFormatDateTime(t *testing.T) {
	icalService := &ICalService{}

	// Test with a known date/time
	testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
	formatted := icalService.formatDateTime(testTime)
	expected := "20231225T143045Z"

	if formatted != expected {
		t.Errorf("Expected formatted datetime '%s', got '%s'", expected, formatted)
	}
}

func TestICalServiceFormatDate(t *testing.T) {
	icalService := &ICalService{}

	// Test with a known date
	testTime := time.Date(2023, 12, 25, 14, 30, 45, 0, time.UTC)
	formatted := icalService.formatDate(testTime)
	expected := "20231225"

	if formatted != expected {
		t.Errorf("Expected formatted date '%s', got '%s'", expected, formatted)
	}
}

func TestICalServiceWrapText(t *testing.T) {
	icalService := &ICalService{}

	// Test text that doesn't need wrapping
	shortText := "Short text"
	wrapped := icalService.wrapText(shortText, 75)
	if wrapped != shortText {
		t.Errorf("Expected short text to remain unchanged, got '%s'", wrapped)
	}

	// Test text that needs wrapping
	longText := "This is a very long text that definitely needs to be wrapped at some point because it exceeds the maximum line length"
	wrapped = icalService.wrapText(longText, 75)

	// Should contain line continuation
	if !strings.Contains(wrapped, "\r\n ") {
		t.Error("Expected wrapped text to contain line continuation")
	}

	// Lines should not exceed the limit (accounting for continuation)
	lines := strings.Split(wrapped, "\r\n")
	for i, line := range lines {
		actualLength := len(line)
		if i > 0 && strings.HasPrefix(line, " ") {
			actualLength-- // Account for continuation space
		}
		if actualLength > 75 {
			t.Errorf("Line %d exceeds length limit: %d > 75", i, actualLength)
		}
	}
}

func TestToICalEvent(t *testing.T) {
	now := time.Now()
	event := models.CalendarEvent{
		ID:            1,
		MovieID:       123,
		Title:         "Test Movie",
		OriginalTitle: "Original Test Movie",
		EventType:     models.CalendarEventCinemaRelease,
		EventDate:     now,
		Overview:      "This is a test movie overview",
		AllDay:        true,
		CreatedAt:     now.Add(-time.Hour),
		UpdatedAt:     now,
	}

	baseURL := "https://radarr.example.com"
	icalEvent := event.ToICalEvent(baseURL)

	expectedUID := "radarr-event-123-cinemaRelease@radarr-go"
	if icalEvent.UID != expectedUID {
		t.Errorf("Expected UID '%s', got '%s'", expectedUID, icalEvent.UID)
	}

	expectedSummary := "Test Movie (Original Test Movie)"
	if icalEvent.Summary != expectedSummary {
		t.Errorf("Expected summary '%s', got '%s'", expectedSummary, icalEvent.Summary)
	}

	expectedURL := "https://radarr.example.com/movie/123"
	if icalEvent.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, icalEvent.URL)
	}

	if !icalEvent.AllDay {
		t.Error("Expected event to be all day")
	}

	if len(icalEvent.Categories) == 0 {
		t.Error("Expected event to have categories")
	}
}

func TestICalServiceGenerateURL(t *testing.T) {
	icalService := &ICalService{}

	config := &models.CalendarFeedConfig{
		Title:        "Test Calendar",
		DaysInFuture: 180,
		DaysInPast:   15,
		EventTypes:   []models.CalendarEventType{models.CalendarEventCinemaRelease},
		Tags:         []int{1, 2, 3},
		RequireAuth:  true,
		PassKey:      "secret123",
	}

	baseURL := "https://radarr.example.com"
	url := icalService.GenerateICalFeedURL(baseURL, config)

	if !strings.HasPrefix(url, "https://radarr.example.com/api/v3/calendar/feed.ics") {
		t.Errorf("Expected URL to start with base path, got '%s'", url)
	}

	if !strings.Contains(url, "daysInFuture=180") {
		t.Error("Expected URL to contain daysInFuture parameter")
	}

	if !strings.Contains(url, "daysInPast=15") {
		t.Error("Expected URL to contain daysInPast parameter")
	}

	if !strings.Contains(url, "eventTypes=cinemaRelease") {
		t.Error("Expected URL to contain eventTypes parameter")
	}

	if !strings.Contains(url, "tags=1,2,3") {
		t.Error("Expected URL to contain tags parameter")
	}

	if !strings.Contains(url, "passKey=secret123") {
		t.Error("Expected URL to contain passKey parameter")
	}
}
