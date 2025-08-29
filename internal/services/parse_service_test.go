// Package services provides tests for the parse service functionality.
package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseService(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewParseService(db, logger)

	t.Run("ParseReleaseTitle", func(t *testing.T) {
		ctx := context.Background()
		testCases := []struct {
			title           string
			expectedMovie   string
			expectedYear    int
			expectedQuality string
		}{
			{
				title:           "The Matrix 1999 1080p BluRay x264-GROUP",
				expectedMovie:   "The Matrix",
				expectedYear:    1999,
				expectedQuality: "1080p Bluray",
			},
			{
				title:           "Inception.2010.720p.WEB-DL.H264-FGT",
				expectedMovie:   "Inception",
				expectedYear:    2010,
				expectedQuality: "720p WEB-DL",
			},
			{
				title:           "Avatar.2009.2160p.4K.BluRay.x265-REMUX",
				expectedMovie:   "Avatar",
				expectedYear:    2009,
				expectedQuality: "2160p",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				result, err := service.ParseReleaseTitle(ctx, tc.title)
				require.NoError(t, err)
				assert.NotNil(t, result.ParsedMovieInfo)
				assert.Equal(t, tc.expectedMovie, result.ParsedMovieInfo.PrimaryMovieTitle)
				assert.Equal(t, tc.expectedYear, result.ParsedMovieInfo.Year)
				assert.Equal(t, tc.expectedQuality, result.ParsedMovieInfo.Quality.Quality.Name)
			})
		}
	})

	t.Run("ParseMultipleTitles", func(t *testing.T) {
		ctx := context.Background()
		titles := []string{
			"The Dark Knight 2008 1080p BluRay",
			"Interstellar 2014 720p WEB-DL",
			"Invalid.Title.Without.Year",
		}

		results, err := service.ParseMultipleTitles(ctx, titles)
		require.NoError(t, err)
		assert.Len(t, results, 3)

		// First two should parse successfully
		assert.Equal(t, "The Dark Knight", results[0].ParsedMovieInfo.PrimaryMovieTitle)
		assert.Equal(t, 2008, results[0].ParsedMovieInfo.Year)

		assert.Equal(t, "Interstellar", results[1].ParsedMovieInfo.PrimaryMovieTitle)
		assert.Equal(t, 2014, results[1].ParsedMovieInfo.Year)
	})

	t.Run("CacheResults", func(t *testing.T) {
		ctx := context.Background()
		title := "Cached Movie 2021 1080p BluRay x264-TEST"

		// Parse once
		result1, err := service.ParseReleaseTitle(ctx, title)
		require.NoError(t, err)

		// Parse again (should come from cache)
		result2, err := service.ParseReleaseTitle(ctx, title)
		require.NoError(t, err)

		assert.Equal(t, result1.ParsedMovieInfo.PrimaryMovieTitle, result2.ParsedMovieInfo.PrimaryMovieTitle)
		assert.Equal(t, result1.ParsedMovieInfo.Year, result2.ParsedMovieInfo.Year)
	})

	t.Run("ClearCache", func(t *testing.T) {
		ctx := context.Background()

		// Parse a title to populate cache
		_, err := service.ParseReleaseTitle(ctx, "Test Movie 2020 720p WEB-DL")
		require.NoError(t, err)

		// Clear cache
		err = service.ClearCache(ctx)
		require.NoError(t, err)

		// Parse again (should work without cache)
		result, err := service.ParseReleaseTitle(ctx, "Another Movie 2021 1080p BluRay")
		require.NoError(t, err)
		assert.Equal(t, "Another Movie", result.ParsedMovieInfo.PrimaryMovieTitle)
	})

	t.Run("QualityExtraction", func(t *testing.T) {
		ctx := context.Background()
		testCases := []struct {
			title           string
			expectedQuality string
			expectedSource  string
		}{
			{
				title:           "Movie 2020 2160p BluRay UHD x265",
				expectedQuality: "2160p",
				expectedSource:  "",
			},
			{
				title:           "Movie 2020 1080p WEB-DL H264",
				expectedQuality: "1080p WEB-DL",
				expectedSource:  "webdl",
			},
			{
				title:           "Movie 2020 720p HDTV x264",
				expectedQuality: "720p HDTV",
				expectedSource:  "hdtv",
			},
			{
				title:           "Movie 2020 480p DVDRip XviD",
				expectedQuality: "DVD-Rip",
				expectedSource:  "dvd",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.title, func(t *testing.T) {
				result, err := service.ParseReleaseTitle(ctx, tc.title)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedQuality, result.ParsedMovieInfo.Quality.Quality.Name)
				if tc.expectedSource != "" {
					assert.Equal(t, tc.expectedSource, result.ParsedMovieInfo.Quality.Quality.Source)
				}
			})
		}
	})
}
