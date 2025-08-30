// Package services provides tests for the collection service functionality.
package services

import (
	"context"
	"testing"

	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectionService_CreateAndGet(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewCollectionService(db, logger)
	ctx := context.Background()

	collection := createTestCollection("Test Collection", 12345)

	// Create collection
	created, err := service.Create(ctx, collection)
	require.NoError(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, collection.Title, created.Title)
	assert.Equal(t, collection.TmdbID, created.TmdbID)

	// Test retrieval methods
	testCollectionRetrieval(ctx, t, service, created)
}

func TestCollectionService_Update(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewCollectionService(db, logger)
	ctx := context.Background()

	collection := createTestCollection("Original Title", 54321)
	collection.MinimumAvailability = "announced"

	created, err := service.Create(ctx, collection)
	require.NoError(t, err)

	// Update collection
	updates := &models.MovieCollectionV2{
		TmdbID:              54321,
		Monitored:           false,
		MinimumAvailability: "released",
	}

	updated, err := service.Update(ctx, created.ID, updates)
	require.NoError(t, err)
	assert.False(t, updated.Monitored)
	assert.Equal(t, "released", updated.MinimumAvailability)
}

func TestCollectionService_Delete(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewCollectionService(db, logger)
	ctx := context.Background()

	collection := createTestCollection("To Delete", 99999)

	created, err := service.Create(ctx, collection)
	require.NoError(t, err)

	// Delete collection without movies
	err = service.Delete(ctx, created.ID, false)
	require.NoError(t, err)

	// Verify it's deleted
	_, err = service.GetByID(ctx, created.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCollectionService_GetAll(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewCollectionService(db, logger)
	ctx := context.Background()

	// Create test collections
	testCollections := createMultipleTestCollections()

	for _, c := range testCollections {
		_, err := service.Create(ctx, c)
		require.NoError(t, err)
	}

	// Test getting all and filtered collections
	testGetAllCollections(ctx, t, service)
}

func TestCollectionService_Statistics(t *testing.T) {
	db, logger := setupTestDB(t)
	defer cleanupTestDB(db)

	service := NewCollectionService(db, logger)
	ctx := context.Background()

	collection := createTestCollection("Stats Test", 88888)

	created, err := service.Create(ctx, collection)
	require.NoError(t, err)

	stats, err := service.GetCollectionStatistics(ctx, created.ID)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Zero(t, stats.MovieCount) // No movies added yet
}

// createTestCollection creates a basic test collection
func createTestCollection(title string, tmdbID int) *models.MovieCollectionV2 {
	return &models.MovieCollectionV2{
		Title:               title,
		TmdbID:              tmdbID,
		Overview:            "A test movie collection",
		Monitored:           true,
		QualityProfileID:    1,
		MinimumAvailability: "released",
	}
}

// testCollectionRetrieval tests various collection retrieval methods
func testCollectionRetrieval(
	ctx context.Context, t *testing.T, service *CollectionService,
	created *models.MovieCollectionV2,
) {
	// Get by ID
	retrieved, err := service.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, retrieved.ID)
	assert.Equal(t, created.Title, retrieved.Title)

	// Get by TMDB ID
	byTmdb, err := service.GetByTmdbID(ctx, created.TmdbID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byTmdb.ID)
}

// createMultipleTestCollections creates multiple test collections
func createMultipleTestCollections() []*models.MovieCollectionV2 {
	return []*models.MovieCollectionV2{
		{Title: "Collection 1", TmdbID: 11111, QualityProfileID: 1, Monitored: true, MinimumAvailability: "announced"},
		{Title: "Collection 2", TmdbID: 22222, QualityProfileID: 1, Monitored: false, MinimumAvailability: "announced"},
	}
}

// testGetAllCollections tests getting all collections with filtering
func testGetAllCollections(ctx context.Context, t *testing.T, service *CollectionService) {
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
}
