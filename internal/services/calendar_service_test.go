package services

import (
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/models"
)

func TestDefaultCalendarConfiguration(t *testing.T) {
	config := models.DefaultCalendarConfiguration()

	if config.DefaultView != models.CalendarViewMonth {
		t.Errorf("Expected default view to be 'month', got '%s'", config.DefaultView)
	}

	if !config.EnableEventCaching {
		t.Error("Expected event caching to be enabled by default")
	}

	if config.EventCacheDuration != 60 {
		t.Errorf("Expected default cache duration to be 60 minutes, got %d", config.EventCacheDuration)
	}

	if len(config.EnabledEventTypes) != 3 {
		t.Errorf("Expected 3 default event types, got %d", len(config.EnabledEventTypes))
	}

	expectedTypes := []models.CalendarEventType{
		models.CalendarEventCinemaRelease,
		models.CalendarEventPhysicalRelease,
		models.CalendarEventDigitalRelease,
	}

	for i, expectedType := range expectedTypes {
		if config.EnabledEventTypes[i] != expectedType {
			t.Errorf("Expected event type %s at index %d, got %s", expectedType, i, config.EnabledEventTypes[i])
		}
	}
}

func TestCalendarEventMethods(t *testing.T) {
	now := time.Now()
	futureDate := now.Add(24 * time.Hour)
	pastDate := now.Add(-24 * time.Hour)

	// Test future event
	futureEvent := models.CalendarEvent{
		EventDate: futureDate,
		EventType: models.CalendarEventCinemaRelease,
	}

	if !futureEvent.IsUpcoming() {
		t.Error("Expected future event to be upcoming")
	}

	if futureEvent.IsPast() {
		t.Error("Expected future event not to be past")
	}

	// Test past event
	pastEvent := models.CalendarEvent{
		EventDate: pastDate,
		EventType: models.CalendarEventPhysicalRelease,
	}

	if pastEvent.IsUpcoming() {
		t.Error("Expected past event not to be upcoming")
	}

	if !pastEvent.IsPast() {
		t.Error("Expected past event to be past")
	}

	// Test event priority
	cinemaEvent := models.CalendarEvent{EventType: models.CalendarEventCinemaRelease}
	physicalEvent := models.CalendarEvent{EventType: models.CalendarEventPhysicalRelease}

	if cinemaEvent.GetEventPriority() >= physicalEvent.GetEventPriority() {
		t.Error("Expected cinema release to have higher priority than physical release")
	}

	// Test event color
	color := cinemaEvent.GetEventColor()
	if color == "" {
		t.Error("Expected event to have a color")
	}
}

func TestCalendarEventUID(t *testing.T) {
	event := models.CalendarEvent{
		MovieID:   123,
		EventType: models.CalendarEventCinemaRelease,
	}

	uid := event.GenerateUID()
	expectedUID := "radarr-event-123-cinemaRelease@radarr-go"

	if uid != expectedUID {
		t.Errorf("Expected UID '%s', got '%s'", expectedUID, uid)
	}
}

func TestCalendarEventDisplayTitle(t *testing.T) {
	// Test with original title same as title
	event1 := models.CalendarEvent{
		Title:         "Test Movie",
		OriginalTitle: "Test Movie",
	}

	if event1.GetDisplayTitle() != "Test Movie" {
		t.Errorf("Expected display title 'Test Movie', got '%s'", event1.GetDisplayTitle())
	}

	// Test with different original title
	event2 := models.CalendarEvent{
		Title:         "Test Movie",
		OriginalTitle: "Original Test Movie",
	}

	expected := "Test Movie (Original Test Movie)"
	if event2.GetDisplayTitle() != expected {
		t.Errorf("Expected display title '%s', got '%s'", expected, event2.GetDisplayTitle())
	}

	// Test with empty original title
	event3 := models.CalendarEvent{
		Title:         "Test Movie",
		OriginalTitle: "",
	}

	if event3.GetDisplayTitle() != "Test Movie" {
		t.Errorf("Expected display title 'Test Movie', got '%s'", event3.GetDisplayTitle())
	}
}
