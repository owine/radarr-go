package models_test

import (
	"fmt"
	"time"

	"github.com/radarr/radarr-go/internal/models"
)

// ExampleMovie_computeAvailability demonstrates how to check movie availability
func ExampleMovie_computeAvailability() {
	// Create a movie that's in cinemas
	inCinemasDate := time.Now().Add(-30 * 24 * time.Hour) // 30 days ago
	movie := &models.Movie{
		Title:               "Example Movie",
		Status:              models.MovieStatusInCinemas,
		InCinemas:           &inCinemasDate,
		MinimumAvailability: models.AvailabilityInCinemas,
	}

	// Check availability after finding (triggers AfterFind hook)
	_ = movie.AfterFind(nil)

	fmt.Printf("Movie is available: %t", movie.IsAvailable)
	// Output: Movie is available: true
}

// ExampleMovieStatus shows the different movie status constants
func ExampleMovieStatus() {
	statuses := []models.MovieStatus{
		models.MovieStatusTBA,
		models.MovieStatusAnnounced,
		models.MovieStatusInCinemas,
		models.MovieStatusReleased,
		models.MovieStatusDeleted,
	}

	for _, status := range statuses {
		fmt.Printf("Status: %s\n", status)
	}
	// Output:
	// Status: tba
	// Status: announced
	// Status: inCinemas
	// Status: released
	// Status: deleted
}

// ExampleAvailability shows the different availability options
func ExampleAvailability() {
	availabilities := []models.Availability{
		models.AvailabilityTBA,
		models.AvailabilityAnnounced,
		models.AvailabilityInCinemas,
		models.AvailabilityReleased,
		models.AvailabilityPreDB,
	}

	for _, availability := range availabilities {
		fmt.Printf("Availability: %s\n", availability)
	}
	// Output:
	// Availability: tba
	// Availability: announced
	// Availability: inCinemas
	// Availability: released
	// Availability: preDB
}

// ExampleStringArray_Value demonstrates JSON marshaling of string arrays
func ExampleStringArray_Value() {
	genres := models.StringArray{"Action", "Adventure", "Sci-Fi"}

	value, err := genres.Value()
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	fmt.Printf("JSON: %s", value)
	// Output: JSON: ["Action","Adventure","Sci-Fi"]
}
