package services

import (
	"testing"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/testhelpers"
)

// BenchmarkMovieService_Search benchmarks the movie search functionality
func BenchmarkMovieService_Search(b *testing.B) {
	testhelpers.SkipInShortMode(b)
	testhelpers.RequireAnyDatabase(b)

	testhelpers.RunBenchmarkWithTestDatabase(b, testhelpers.GetTestDatabaseType(), func(b *testing.B, db *database.Database, log *logger.Logger) {
		// Create test data factory
		factory := testhelpers.NewTestDataFactory(db.GORM)
		defer factory.Cleanup()

		// Create a quality profile first
		profile := factory.CreateQualityProfile()

		// Create some test movies for searching
		for i := 0; i < 10; i++ {
			factory.CreateMovie(func(m *models.Movie) {
				m.TmdbID = 12345 + i
				m.Title = "Search Test Movie " + string(rune('A'+i))
				m.TitleSlug = "search-test-movie-" + string(rune('a'+i))
				m.QualityProfileID = profile.ID
			})
		}

		// Create movie service
		movieService := NewMovieService(db, log)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = movieService.Search("test")
		}
	})
}

// BenchmarkMovieService_GetTotalCount benchmarks the count operation
func BenchmarkMovieService_GetTotalCount(b *testing.B) {
	testhelpers.SkipInShortMode(b)
	testhelpers.RequireAnyDatabase(b)

	testhelpers.RunBenchmarkWithTestDatabase(b, testhelpers.GetTestDatabaseType(), func(b *testing.B, db *database.Database, log *logger.Logger) {
		// Create test data factory
		factory := testhelpers.NewTestDataFactory(db.GORM)
		defer factory.Cleanup()

		// Create a quality profile first
		profile := factory.CreateQualityProfile()

		// Create test movies for counting
		for i := 0; i < 100; i++ {
			factory.CreateMovie(func(m *models.Movie) {
				m.TmdbID = 12345 + i
				m.Title = "Count Test Movie " + string(rune('A'+i%26))
				m.TitleSlug = "count-test-movie-" + string(rune('a'+i%26))
				m.QualityProfileID = profile.ID
			})
		}

		// Create movie service
		movieService := NewMovieService(db, log)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = movieService.GetTotalCount()
		}
	})
}

// BenchmarkMovieService_Create benchmarks movie creation
func BenchmarkMovieService_Create(b *testing.B) {
	testhelpers.SkipInShortMode(b)
	testhelpers.RequireAnyDatabase(b)

	testhelpers.RunBenchmarkWithTestDatabase(b, testhelpers.GetTestDatabaseType(), func(b *testing.B, db *database.Database, log *logger.Logger) {
		// Create test data factory
		factory := testhelpers.NewTestDataFactory(db.GORM)
		defer factory.Cleanup()

		// Create a quality profile first (required for movies)
		profile := factory.CreateQualityProfile()

		// Create movie service
		movieService := NewMovieService(db, log)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			movie := &models.Movie{
				TmdbID:           12345 + i,
				Title:            "Benchmark Movie",
				TitleSlug:        "benchmark-movie",
				Year:             2023,
				QualityProfileID: profile.ID,
				Monitored:        true,
			}
			b.StartTimer()

			_ = movieService.Create(movie)

			b.StopTimer()
			// Clean up the created movie to avoid unique constraint violations
			_ = db.GORM.Delete(movie)
			b.StartTimer()
		}
	})
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

// BenchmarkMovieService_GetByID benchmarks movie retrieval by ID
func BenchmarkMovieService_GetByID(b *testing.B) {
	testhelpers.SkipInShortMode(b)
	testhelpers.RequireAnyDatabase(b)

	testhelpers.RunBenchmarkWithTestDatabase(b, testhelpers.GetTestDatabaseType(), func(b *testing.B, db *database.Database, log *logger.Logger) {
		// Create test data factory
		factory := testhelpers.NewTestDataFactory(db.GORM)
		defer factory.Cleanup()

		// Create a quality profile first
		profile := factory.CreateQualityProfile()

		// Create a test movie
		movie := factory.CreateMovie(func(m *models.Movie) {
			m.QualityProfileID = profile.ID
		})

		// Create movie service
		movieService := NewMovieService(db, log)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = movieService.GetByID(movie.ID)
		}
	})
}

// BenchmarkMovieService_GetByTmdbID benchmarks movie retrieval by TMDB ID
func BenchmarkMovieService_GetByTmdbID(b *testing.B) {
	testhelpers.SkipInShortMode(b)
	testhelpers.RequireAnyDatabase(b)

	testhelpers.RunBenchmarkWithTestDatabase(b, testhelpers.GetTestDatabaseType(), func(b *testing.B, db *database.Database, log *logger.Logger) {
		// Create test data factory
		factory := testhelpers.NewTestDataFactory(db.GORM)
		defer factory.Cleanup()

		// Create a quality profile first
		profile := factory.CreateQualityProfile()

		// Create a test movie
		movie := factory.CreateMovie(func(m *models.Movie) {
			m.QualityProfileID = profile.ID
		})

		// Create movie service
		movieService := NewMovieService(db, log)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = movieService.GetByTmdbID(movie.TmdbID)
		}
	})
}
