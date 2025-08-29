package services

import (
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWantedMoviesService_GetMissingMovies(t *testing.T) {
	// Setup test database
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	movieService := NewMovieService(db, logger)
	qualityService := NewQualityService(db, logger)
	wantedService := NewWantedMoviesService(db, logger, movieService, qualityService)

	// Create test quality profile
	profile := &models.QualityProfile{
		Name:   "Test Profile",
		Cutoff: 7, // Bluray-1080p
		Items: models.QualityProfileItems{
			&models.QualityProfileItem{
				Quality: &models.QualityLevel{ID: 1, Title: "SDTV", Weight: 1},
				Allowed: true,
			},
			&models.QualityProfileItem{
				Quality: &models.QualityLevel{ID: 7, Title: "Bluray-1080p", Weight: 7},
				Allowed: true,
			},
		},
		UpgradeAllowed: true,
	}
	err := qualityService.CreateQualityProfile(profile)
	require.NoError(t, err)

	// Create test movie without file (missing)
	movie := &models.Movie{
		Title:               "Test Missing Movie",
		TmdbID:              12345,
		TitleSlug:           "test-missing-movie",
		QualityProfileID:    profile.ID,
		Monitored:           true,
		HasFile:             false,
		IsAvailable:         true,
		Year:                2023,
		Status:              models.MovieStatusReleased,
		MinimumAvailability: models.AvailabilityReleased,
	}
	err = movieService.Create(movie)
	require.NoError(t, err)

	// Create wanted movie entry
	wantedMovie := &models.WantedMovie{
		MovieID:         movie.ID,
		Status:          models.WantedStatusMissing,
		Reason:          "Movie has no file",
		TargetQualityID: profile.Cutoff,
		IsAvailable:     true,
		Priority:        models.PriorityHigh,
	}
	err = db.GORM.Create(wantedMovie).Error
	require.NoError(t, err)

	// Test GetMissingMovies
	filter := &models.WantedMovieFilter{
		Page:     1,
		PageSize: 10,
	}

	response, err := wantedService.GetMissingMovies(filter)
	require.NoError(t, err)

	assert.NotNil(t, response)
	assert.Equal(t, int64(1), response.TotalRecords)
	assert.Len(t, response.Records, 1)
	assert.Equal(t, models.WantedStatusMissing, response.Records[0].Status)
	assert.Equal(t, movie.ID, response.Records[0].MovieID)
}

func TestWantedMoviesService_RefreshWantedMovies(t *testing.T) {
	// Setup test database
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	movieService := NewMovieService(db, logger)
	qualityService := NewQualityService(db, logger)
	wantedService := NewWantedMoviesService(db, logger, movieService, qualityService)

	// Initialize default quality definitions
	err := qualityService.InitializeQualityDefinitions()
	require.NoError(t, err)

	// Create test quality profile
	profile := &models.QualityProfile{
		Name:   "Test Profile",
		Cutoff: 7, // Bluray-1080p
		Items: models.QualityProfileItems{
			&models.QualityProfileItem{
				Quality: &models.QualityLevel{ID: 1, Title: "SDTV", Weight: 1},
				Allowed: true,
			},
			&models.QualityProfileItem{
				Quality: &models.QualityLevel{ID: 7, Title: "Bluray-1080p", Weight: 7},
				Allowed: true,
			},
		},
		UpgradeAllowed: true,
	}
	err = qualityService.CreateQualityProfile(profile)
	require.NoError(t, err)

	// Create missing movie
	missingMovie := &models.Movie{
		Title:               "Missing Movie",
		TmdbID:              99901,
		TitleSlug:           "missing-movie",
		QualityProfileID:    profile.ID,
		Monitored:           true,
		HasFile:             false,
		IsAvailable:         true,
		Year:                2023,
		Status:              models.MovieStatusReleased,
		MinimumAvailability: models.AvailabilityReleased,
	}
	err = movieService.Create(missingMovie)
	require.NoError(t, err)

	// Create movie with adequate quality (should not be wanted)
	goodMovie := &models.Movie{
		Title:               "Good Quality Movie",
		TmdbID:              99902,
		TitleSlug:           "good-quality-movie",
		QualityProfileID:    profile.ID,
		Monitored:           true,
		HasFile:             true,
		IsAvailable:         true,
		Year:                2023,
		Status:              models.MovieStatusReleased,
		MinimumAvailability: models.AvailabilityReleased,
	}
	err = movieService.Create(goodMovie)
	require.NoError(t, err)

	// Create high quality file
	goodFile := &models.MovieFile{
		MovieID:      goodMovie.ID,
		RelativePath: "good-quality-movie.mkv",
		Path:         "/movies/good-quality-movie.mkv",
		Size:         5000000000,
		DateAdded:    time.Now(),
		Quality: models.Quality{
			Quality: models.QualityDefinition{
				ID:   7, // Bluray-1080p - meets cutoff
				Name: "Bluray-1080p",
			},
		},
	}
	err = db.GORM.Create(goodFile).Error
	require.NoError(t, err)

	goodMovie.MovieFileID = goodFile.ID
	err = movieService.Update(goodMovie)
	require.NoError(t, err)

	// Run refresh
	err = wantedService.RefreshWantedMovies()
	require.NoError(t, err)

	// Check stats
	stats, err := wantedService.GetWantedStats()
	require.NoError(t, err)

	assert.Equal(t, int64(1), stats.TotalWanted) // Only missing movie should be wanted
	assert.Equal(t, int64(1), stats.MissingCount)
	assert.Equal(t, int64(0), stats.CutoffUnmetCount) // Good movie shouldn't be wanted

	// Verify wanted movie was created for missing movie
	var wantedMovies []models.WantedMovie
	err = db.GORM.Find(&wantedMovies).Error
	require.NoError(t, err)

	assert.Len(t, wantedMovies, 1)
	assert.Equal(t, missingMovie.ID, wantedMovies[0].MovieID)
	assert.Equal(t, models.WantedStatusMissing, wantedMovies[0].Status)
}

func TestWantedMoviesService_GetEligibleForSearch(t *testing.T) {
	// Setup test database
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	movieService := NewMovieService(db, logger)
	qualityService := NewQualityService(db, logger)
	wantedService := NewWantedMoviesService(db, logger, movieService, qualityService)

	// Create test movies
	movie1 := &models.Movie{
		Title:     "Eligible Movie",
		TmdbID:    22221,
		TitleSlug: "eligible-movie",
		Monitored: true,
		HasFile:   false,
	}
	err := movieService.Create(movie1)
	require.NoError(t, err)

	movie2 := &models.Movie{
		Title:     "Max Attempts Movie",
		TmdbID:    22222,
		TitleSlug: "max-attempts-movie",
		Monitored: true,
		HasFile:   false,
	}
	err = movieService.Create(movie2)
	require.NoError(t, err)

	// Create wanted movie eligible for search
	eligibleWanted := &models.WantedMovie{
		MovieID:           movie1.ID,
		Status:            models.WantedStatusMissing,
		TargetQualityID:   7,
		IsAvailable:       true,
		SearchAttempts:    2,
		MaxSearchAttempts: 10,
		Priority:          models.PriorityHigh,
	}
	err = db.GORM.Create(eligibleWanted).Error
	require.NoError(t, err)

	// Create wanted movie that has reached max attempts
	maxAttemptsWanted := &models.WantedMovie{
		MovieID:           movie2.ID,
		Status:            models.WantedStatusMissing,
		TargetQualityID:   7,
		IsAvailable:       true,
		SearchAttempts:    10,
		MaxSearchAttempts: 10,
		Priority:          models.PriorityHigh,
	}
	err = db.GORM.Create(maxAttemptsWanted).Error
	require.NoError(t, err)

	// Test GetEligibleForSearch
	eligible, err := wantedService.GetEligibleForSearch(10)
	require.NoError(t, err)

	assert.Len(t, eligible, 1)
	assert.Equal(t, movie1.ID, eligible[0].MovieID)
	assert.True(t, eligible[0].IsEligibleForSearch())
}

func TestWantedMovie_IsEligibleForSearch(t *testing.T) {
	// Test eligible movie
	eligible := &models.WantedMovie{
		IsAvailable:       true,
		SearchAttempts:    5,
		MaxSearchAttempts: 10,
		NextSearchTime:    nil,
	}
	assert.True(t, eligible.IsEligibleForSearch())

	// Test max attempts reached
	maxAttempts := &models.WantedMovie{
		IsAvailable:       true,
		SearchAttempts:    10,
		MaxSearchAttempts: 10,
		NextSearchTime:    nil,
	}
	assert.False(t, maxAttempts.IsEligibleForSearch())

	// Test not available
	notAvailable := &models.WantedMovie{
		IsAvailable:       false,
		SearchAttempts:    5,
		MaxSearchAttempts: 10,
		NextSearchTime:    nil,
	}
	assert.False(t, notAvailable.IsEligibleForSearch())

	// Test future next search time
	future := time.Now().Add(time.Hour)
	futureSearch := &models.WantedMovie{
		IsAvailable:       true,
		SearchAttempts:    5,
		MaxSearchAttempts: 10,
		NextSearchTime:    &future,
	}
	assert.False(t, futureSearch.IsEligibleForSearch())

	// Test past next search time
	past := time.Now().Add(-time.Hour)
	pastSearch := &models.WantedMovie{
		IsAvailable:       true,
		SearchAttempts:    5,
		MaxSearchAttempts: 10,
		NextSearchTime:    &past,
	}
	assert.True(t, pastSearch.IsEligibleForSearch())
}

func TestWantedMovie_CalculateNextSearchTime(t *testing.T) {
	// Test first attempt
	wanted := &models.WantedMovie{
		SearchAttempts: 0,
		Priority:       models.PriorityNormal,
	}
	nextTime := wanted.CalculateNextSearchTime()
	assert.True(t, nextTime.After(time.Now()))
	assert.True(t, nextTime.Before(time.Now().Add(5*time.Hour))) // Should be around 2 hours

	// Test multiple attempts (exponential backoff)
	wanted.SearchAttempts = 3
	nextTime = wanted.CalculateNextSearchTime()
	assert.True(t, nextTime.After(time.Now().Add(10*time.Hour))) // Should be much longer

	// Test high priority (shorter delay)
	wanted.Priority = models.PriorityVeryHigh
	wanted.SearchAttempts = 1
	nextTime = wanted.CalculateNextSearchTime()
	highPriorityDelay := time.Until(nextTime)

	// Test low priority (longer delay)
	wanted.Priority = models.PriorityVeryLow
	nextTime = wanted.CalculateNextSearchTime()
	lowPriorityDelay := time.Until(nextTime)

	assert.True(t, lowPriorityDelay > highPriorityDelay)
}
