package services

import (
	"testing"

	"github.com/radarr/radarr-go/internal/models"
)

// BenchmarkMovieService_Search benchmarks the movie search functionality
func BenchmarkMovieService_Search(b *testing.B) {
	// Skip this benchmark since it requires a real database connection
	b.Skip("Skipping benchmark that requires database connection")
}

// BenchmarkMovieService_GetTotalCount benchmarks the count operation
func BenchmarkMovieService_GetTotalCount(b *testing.B) {
	// Skip this benchmark since it requires a real database connection
	b.Skip("Skipping benchmark that requires database connection")
}

// BenchmarkMovieService_Create benchmarks movie creation
func BenchmarkMovieService_Create(b *testing.B) {
	// Skip this benchmark since it requires a real database connection
	b.Skip("Skipping benchmark that requires database connection")
}

// BenchmarkMovieService_ValidationHooks benchmarks the GORM validation hooks
func BenchmarkMovieService_ValidationHooks(b *testing.B) {
	movie := &models.Movie{
		TmdbID:    12345,
		Title:     "Test Movie",
		TitleSlug: "test-movie",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Benchmark the validation hook execution
		_ = movie.BeforeCreate(nil)
		_ = movie.BeforeUpdate(nil)
		_ = movie.AfterFind(nil)
	}
}
