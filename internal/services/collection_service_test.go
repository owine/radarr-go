// Package services provides tests for the collection service functionality.
package services

import (
	"context"
	"testing"

	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionService(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewCollectionService(db, logger)

	t.Run("CreateAndGetCollection", func(t *testing.T) {
		ctx := context.Background()
		collection := &models.MovieCollection{
			Title:               "Test Collection",
			CleanTitle:          "testcollection",
			TmdbID:              12345,
			Overview:            "A test movie collection",
			Monitored:           true,
			QualityProfileID:    1,
			MinimumAvailability: models.AvailabilityReleased,
			SearchOnAdd:         true,
		}

		// Create collection
		created, err := service.Create(ctx, collection)
		require.NoError(t, err)
		assert.NotZero(t, created.ID)
		assert.Equal(t, collection.Title, created.Title)
		assert.Equal(t, collection.TmdbID, created.TmdbID)

		// Get by ID
		retrieved, err := service.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.ID)
		assert.Equal(t, created.Title, retrieved.Title)

		// Get by TMDB ID
		byTmdb, err := service.GetByTmdbID(ctx, created.TmdbID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, byTmdb.ID)
	})

	t.Run("UpdateCollection", func(t *testing.T) {
		ctx := context.Background()
		collection := &models.MovieCollection{
			Title:               "Original Title",
			TmdbID:              54321,
			QualityProfileID:    1,
			MinimumAvailability: models.AvailabilityAnnounced,
		}

		created, err := service.Create(ctx, collection)
		require.NoError(t, err)

		// Update collection
		updates := &models.MovieCollection{
			TmdbID:              54321,
			Monitored:           false,
			MinimumAvailability: models.AvailabilityReleased,
		}

		updated, err := service.Update(ctx, created.ID, updates)
		require.NoError(t, err)
		assert.False(t, updated.Monitored)
		assert.Equal(t, models.AvailabilityReleased, updated.MinimumAvailability)
	})

	t.Run("DeleteCollection", func(t *testing.T) {
		ctx := context.Background()
		collection := &models.MovieCollection{
			Title:            "To Delete",
			TmdbID:           99999,
			QualityProfileID: 1,
		}

		created, err := service.Create(ctx, collection)
		require.NoError(t, err)

		// Delete collection without movies
		err = service.Delete(ctx, created.ID, false)
		require.NoError(t, err)

		// Verify it's deleted
		_, err = service.GetByID(ctx, created.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("GetAllCollections", func(t *testing.T) {
		ctx := context.Background()

		// Create test collections
		collections := []*models.MovieCollection{
			{Title: "Collection 1", TmdbID: 11111, QualityProfileID: 1, Monitored: true},
			{Title: "Collection 2", TmdbID: 22222, QualityProfileID: 1, Monitored: false},
		}

		for _, c := range collections {
			_, err := service.Create(ctx, c)
			require.NoError(t, err)
		}

		// Get all collections
		all, err := service.GetAll(ctx, nil)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(all), 2)

		// Get monitored collections only
		monitored := true
		monitoredCollections, err := service.GetAll(ctx, &monitored)
		require.NoError(t, err)

		for _, c := range monitoredCollections {
			assert.True(t, c.Monitored)
		}
	})

	t.Run("GetCollectionStatistics", func(t *testing.T) {
		ctx := context.Background()
		collection := &models.MovieCollection{
			Title:            "Stats Test",
			TmdbID:           88888,
			QualityProfileID: 1,
		}

		created, err := service.Create(ctx, collection)
		require.NoError(t, err)

		stats, err := service.GetCollectionStatistics(ctx, created.ID)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Zero(t, stats.MovieCount) // No movies added yet
	})
}
