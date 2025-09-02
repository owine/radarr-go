package testhelpers_test

import (
	"testing"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
	"github.com/radarr/radarr-go/internal/testhelpers"
	"gorm.io/gorm"
)

// ExampleNewTestContext demonstrates basic test context usage
func ExampleNewTestContext() {
	// This would be inside a test function
	t := &testing.T{} // Placeholder for example

	// Create test context with default options
	ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

	// Use the context for testing
	_ = ctx.DB
	_ = ctx.Logger
	_ = ctx.Factory

	// Context automatically cleans up when test ends
}

// ExampleNewTestContext_withCustomOptions demonstrates custom options
func ExampleNewTestContext_withCustomOptions() {
	t := &testing.T{} // Placeholder for example

	options := testhelpers.TestOptions{
		DatabaseType:    "postgres", // Force PostgreSQL
		IsolateDatabase: true,       // Use isolated schema
		TempDir:         true,       // Create temp directory
		LogLevel:        "debug",    // Enable debug logging
	}

	ctx := testhelpers.NewTestContext(t, options)
	_ = ctx.TempDir // Available for file operations
}

// ExampleRunWithTestDatabase demonstrates the helper function
func ExampleRunWithTestDatabase() {
	t := &testing.T{} // Placeholder for example

	testhelpers.RunWithTestDatabase(t, "postgres", func(t *testing.T, db *database.Database, log *logger.Logger) {
		// Your test logic here
		// Database is automatically set up and cleaned up

		factory := testhelpers.NewTestDataFactory(db.GORM)
		movie := factory.CreateMovie()

		// Test operations with the movie
		_ = movie
	})
}

// ExampleSeedData_SeedBasicDataset demonstrates seeding test data
func ExampleSeedData_SeedBasicDataset() {
	t := &testing.T{} // Placeholder for example

	ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

	// Seed comprehensive test data
	dataset, err := ctx.SeedData.SeedBasicDataset()
	if err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	// Use the seeded data
	_ = dataset.ReleasedMovie
	_ = dataset.QualityProfile
	_ = dataset.PrimaryIndexer
}

// ExampleSeedData_SeedPerformanceDataset demonstrates performance testing setup
func ExampleSeedData_SeedPerformanceDataset() {
	t := &testing.T{} // Placeholder for example

	ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

	// Create large dataset for performance testing
	movieCount := 1000
	dataset, err := ctx.SeedData.SeedPerformanceDataset(movieCount)
	if err != nil {
		t.Fatalf("Failed to seed performance data: %v", err)
	}

	// Run performance tests
	_ = len(dataset.Movies) // Should be 1000
}

// ExampleTestContext_WithTransaction demonstrates transactional testing
func ExampleTestContext_WithTransaction() {
	t := &testing.T{} // Placeholder for example

	ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

	// Test within a transaction (automatically rolled back)
	ctx.WithTransaction(func(tx *gorm.DB) {
		// Create test data directly in transaction
		movie := &models.Movie{
			TmdbID:    123,
			Title:     "Test Movie",
			TitleSlug: "test-movie",
		}

		err := tx.Create(movie).Error
		if err != nil {
			t.Fatalf("Failed to create movie: %v", err)
		}

		// Test operations with the movie
		// Transaction is automatically rolled back at end
	})
}

// ExampleTestContext_AssertDatabaseCount demonstrates database assertions
func ExampleTestContext_AssertDatabaseCount() {
	t := &testing.T{} // Placeholder for example

	ctx := testhelpers.NewTestContext(t, testhelpers.DefaultTestOptions())

	// Create some test data
	ctx.Factory.CreateMovie()
	ctx.Factory.CreateMovie()

	// Assert the count
	ctx.AssertDatabaseCount("movies", 2)

	// Clean up
	ctx.Factory.Cleanup()

	// Assert database is empty
	ctx.AssertDatabaseEmpty("movies", "quality_profiles")
}

// ExampleRequireDatabase demonstrates database requirement checking
func ExampleRequireDatabase() {
	t := &testing.T{} // Placeholder for example

	// Skip test if PostgreSQL is not available
	testhelpers.RequireDatabase(t, "postgres")

	// Test continues only if PostgreSQL is running
}

// ExampleSkipInShortMode demonstrates conditional test skipping
func ExampleSkipInShortMode() {
	t := &testing.T{} // Placeholder for example

	// Skip this test when running with -short flag
	testhelpers.SkipInShortMode(t)

	// Long-running test logic here
}

// ExampleRunBenchmarkWithTestDatabase demonstrates benchmark setup
func ExampleRunBenchmarkWithTestDatabase() {
	b := &testing.B{} // Placeholder for example

	testhelpers.RunBenchmarkWithTestDatabase(b, "postgres", func(b *testing.B, db *database.Database, log *logger.Logger) {
		// Setup benchmark data outside timing
		factory := testhelpers.NewTestDataFactory(db.GORM)
		defer factory.Cleanup()

		// Create test movie for benchmark
		movie := factory.CreateMovie()

		// Reset timer before benchmark loop
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Benchmark operations here
			_ = movie.ID
		}
	})
}

// The following imports would be needed for actual usage:
// import (
//     "github.com/radarr/radarr-go/internal/database"
//     "github.com/radarr/radarr-go/internal/logger"
//     "gorm.io/gorm"
// )
