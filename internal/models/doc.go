// Package models defines the core data structures and database models for Radarr.
//
// This package contains all the domain models that represent the business entities
// in the Radarr application, including movies, quality profiles, indexers, and
// download clients. Each model includes appropriate GORM annotations for database
// mapping and custom serialization methods for JSON handling.
//
// Key Models:
//
//   - Movie: Represents a movie entity with metadata, files, and quality information
//   - QualityProfile: Defines quality settings and upgrade preferences
//   - QualityLevel: Individual quality definitions with size constraints
//   - MovieFile: Represents physical movie files with media information
//   - DownloadClient: Configuration for download automation services
//   - Indexer: Search provider configurations
//   - ImportList: Automatic movie discovery and import settings
//   - History: Audit trail of movie-related events
//   - Notification: Alert and notification configurations
//
// All models implement proper GORM hooks for validation and business logic,
// custom JSON serialization for complex types, and follow Go naming conventions.
//
// Example usage:
//
//	movie := &models.Movie{
//		TmdbID:              550,
//		Title:               "Fight Club",
//		TitleSlug:           "fight-club-1999",
//		Year:                1999,
//		QualityProfileID:    1,
//		MinimumAvailability: models.AvailabilityReleased,
//		Monitored:           true,
//	}
//
//	// GORM hooks will automatically validate the movie before creation
//	db.Create(movie)
package models
