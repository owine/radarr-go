package services

import (
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/tmdb"
	"github.com/stretchr/testify/assert"
)

func TestMetadataService_SearchMovies(t *testing.T) {
	// Skip if TMDB API key not provided
	cfg := &config.Config{
		TMDB: config.TMDBConfig{
			APIKey: "", // Empty for test
		},
	}

	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})

	service := NewMetadataService(nil, cfg, logger)

	// Test with empty API key should return error
	_, err := service.SearchMovies("test", 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TMDB API key not configured")
}

func TestMetadataService_LookupMovieByTMDBID(t *testing.T) {
	// Skip if TMDB API key not provided
	cfg := &config.Config{
		TMDB: config.TMDBConfig{
			APIKey: "", // Empty for test
		},
	}

	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})

	service := NewMetadataService(nil, cfg, logger)

	// Test with invalid TMDB ID
	_, err := service.LookupMovieByTMDBID(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid TMDB ID")

	// Test with empty API key should return error
	_, err = service.LookupMovieByTMDBID(550)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TMDB API key not configured")
}

func TestMetadataService_convertTMDBToMovie(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := &MetadataService{logger: logger}

	// Create test TMDB movie data
	tmdbMovie := &tmdb.Movie{
		ID:               550,
		Title:            "Fight Club",
		OriginalTitle:    "Fight Club",
		OriginalLanguage: "en",
		Overview:         "A ticking-time-bomb insomniac and a slippery soap salesman.",
		ReleaseDate:      "1999-10-15",
		Runtime:          139,
		VoteAverage:      8.4,
		VoteCount:        26280,
		Popularity:       61.416,
		Status:           "Released",
		Genres: []tmdb.Genre{
			{ID: 18, Name: "Drama"},
		},
		ProductionCompanies: []tmdb.ProductionCompany{
			{ID: 508, Name: "Regency Enterprises"},
		},
	}

	movie := service.convertTMDBToMovie(tmdbMovie, nil)

	assert.Equal(t, 550, movie.TmdbID)
	assert.Equal(t, "Fight Club", movie.Title)
	assert.Equal(t, "Fight Club", movie.OriginalTitle)
	assert.Equal(t, "en", movie.OriginalLanguage.Name)
	assert.Equal(t, 1999, movie.Year)
	assert.Equal(t, 139, movie.Runtime)
	assert.Equal(t, "fightclub", movie.CleanTitle)
	assert.Equal(t, "fight-club-1999", movie.TitleSlug)
	assert.Equal(t, "Regency Enterprises", movie.Studio)
	assert.True(t, movie.Monitored)

	// Test release date parsing
	expectedDate := time.Date(1999, 10, 15, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, &expectedDate, movie.InCinemas)
}

func TestMetadataService_generateTitleSlug(t *testing.T) {
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})
	service := &MetadataService{logger: logger}

	tests := []struct {
		title    string
		year     int
		expected string
	}{
		{"Fight Club", 1999, "fight-club-1999"},
		{"The Matrix", 1999, "the-matrix-1999"},
		{"Avengers: Endgame", 2019, "avengers-endgame-2019"},
		{"Spider-Man: No Way Home", 2021, "spider-man-no-way-home-2021"},
		{"Movie with    multiple   spaces", 2020, "movie-with-multiple-spaces-2020"},
		{"Movie with (special) [characters]!", 2020, "movie-with-special-characters-2020"},
		{"", 2020, "2020"},
		{"Test Movie", 0, "test-movie"},
	}

	for _, test := range tests {
		result := service.generateTitleSlug(test.title, test.year)
		assert.Equal(t, test.expected, result, "Failed for title: %s, year: %d", test.title, test.year)
	}
}

func TestMetadataService_RefreshMovieMetadata(t *testing.T) {
	cfg := &config.Config{
		TMDB: config.TMDBConfig{
			APIKey: "", // Empty for test
		},
	}
	logger := logger.New(config.LogConfig{Level: "debug", Format: "text", Output: "stdout"})

	service := NewMetadataService(nil, cfg, logger)

	// Test with nil database should fail
	err := service.RefreshMovieMetadata(1)
	assert.Error(t, err)
}
