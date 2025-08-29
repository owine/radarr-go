package services

import (
	"context"
	"testing"
	"time"

	"github.com/radarr/radarr-go/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMovieService for testing
type MockMovieService struct {
	mock.Mock
}

func (m *MockMovieService) GetByID(id int) (*models.Movie, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Movie), args.Error(1)
}

func (m *MockMovieService) GetAll() ([]models.Movie, error) {
	args := m.Called()
	return args.Get(0).([]models.Movie), args.Error(1)
}

func (m *MockMovieService) Update(movie *models.Movie) error {
	args := m.Called(movie)
	return args.Error(0)
}

// MockMetadataService for testing
type MockMetadataService struct {
	mock.Mock
}

func (m *MockMetadataService) RefreshMovieMetadata(movieID int) error {
	args := m.Called(movieID)
	return args.Error(0)
}

// MockImportListService for testing
type MockImportListService struct {
	mock.Mock
}

func (m *MockImportListService) GetImportListByID(id int) (*models.ImportList, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImportList), args.Error(1)
}

func (m *MockImportListService) GetEnabledImportLists() ([]models.ImportList, error) {
	args := m.Called()
	return args.Get(0).([]models.ImportList), args.Error(1)
}

func (m *MockImportListService) SyncImportList(id int) (*models.ImportListSyncResult, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ImportListSyncResult), args.Error(1)
}

func TestRefreshMovieHandler(t *testing.T) {
	// Setup mocks
	movieService := new(MockMovieService)
	metadataService := new(MockMetadataService)

	// Create handler
	handler := NewRefreshMovieHandler(movieService, metadataService)

	// Test basic properties
	assert.Equal(t, "RefreshMovie", handler.GetName())
	assert.Equal(t, "Refreshes metadata for a single movie from TMDB", handler.GetDescription())

	// Create test movie
	testMovie := &models.Movie{
		ID:     1,
		Title:  "Test Movie",
		TmdbID: 12345,
	}

	// Setup mock expectations
	movieService.On("GetByID", 1).Return(testMovie, nil)
	metadataService.On("RefreshMovieMetadata", 1).Return(nil)
	movieService.On("Update", testMovie).Return(nil)

	// Create task
	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{"movieId": 1},
	}

	// Track progress updates
	var progressUpdates []struct {
		percent int
		message string
	}
	updateProgress := func(percent int, message string) {
		progressUpdates = append(progressUpdates, struct {
			percent int
			message string
		}{percent, message})
	}

	// Execute handler
	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	// Assertions
	require.NoError(t, err)
	movieService.AssertExpectations(t)
	metadataService.AssertExpectations(t)

	// Verify progress was updated
	assert.GreaterOrEqual(t, len(progressUpdates), 3)
	assert.Equal(t, 0, progressUpdates[0].percent)
	assert.Equal(t, "Starting movie refresh", progressUpdates[0].message)
	assert.Equal(t, 100, progressUpdates[len(progressUpdates)-1].percent)
}

func TestRefreshMovieHandler_InvalidMovieID(t *testing.T) {
	movieService := new(MockMovieService)
	metadataService := new(MockMetadataService)
	handler := NewRefreshMovieHandler(movieService, metadataService)

	// Test with missing movieId
	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{},
	}

	updateProgress := func(percent int, message string) {}
	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "movieId not found")

	// Test with invalid movieId type
	task.Body = models.TaskBody{"movieId": "invalid"}
	err = handler.Execute(ctx, task, updateProgress)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid movieId")
}

func TestRefreshAllMoviesHandler(t *testing.T) {
	// Setup mocks
	movieService := new(MockMovieService)
	metadataService := new(MockMetadataService)

	// Create handler
	handler := NewRefreshAllMoviesHandler(movieService, metadataService)

	// Test basic properties
	assert.Equal(t, "RefreshAllMovies", handler.GetName())
	assert.Equal(t, "Refreshes metadata for all movies from TMDB", handler.GetDescription())

	// Create test movies
	testMovies := []models.Movie{
		{ID: 1, Title: "Movie 1", TmdbID: 1001},
		{ID: 2, Title: "Movie 2", TmdbID: 1002},
	}

	// Setup mock expectations
	movieService.On("GetAll").Return(testMovies, nil)
	metadataService.On("RefreshMovieMetadata", 1).Return(nil)
	metadataService.On("RefreshMovieMetadata", 2).Return(nil)
	movieService.On("Update", mock.AnythingOfType("*models.Movie")).Return(nil).Times(2)

	// Create task
	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{},
	}

	// Track progress updates
	var progressUpdates []struct {
		percent int
		message string
	}
	updateProgress := func(percent int, message string) {
		progressUpdates = append(progressUpdates, struct {
			percent int
			message string
		}{percent, message})
	}

	// Execute handler
	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	// Assertions
	require.NoError(t, err)
	movieService.AssertExpectations(t)
	metadataService.AssertExpectations(t)

	// Verify progress was updated
	assert.GreaterOrEqual(t, len(progressUpdates), 3)
	assert.Equal(t, 0, progressUpdates[0].percent)
	assert.Equal(t, "Starting bulk movie refresh", progressUpdates[0].message)
	assert.Equal(t, 100, progressUpdates[len(progressUpdates)-1].percent)
}

func TestRefreshAllMoviesHandler_NoMovies(t *testing.T) {
	movieService := new(MockMovieService)
	metadataService := new(MockMetadataService)
	handler := NewRefreshAllMoviesHandler(movieService, metadataService)

	// Setup mock - no movies
	movieService.On("GetAll").Return([]models.Movie{}, nil)

	task := &models.Task{ID: 1, Body: models.TaskBody{}}
	updateProgress := func(percent int, message string) {}

	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	require.NoError(t, err)
	movieService.AssertExpectations(t)
	// Should not call metadata service when no movies
	metadataService.AssertNotCalled(t, "RefreshMovieMetadata")
}

func TestSyncImportListHandler(t *testing.T) {
	// Setup mock
	importListService := new(MockImportListService)

	// Create handler
	handler := NewSyncImportListHandler(importListService)

	// Test basic properties
	assert.Equal(t, "SyncImportList", handler.GetName())
	assert.Equal(t, "Syncs movies from configured import lists", handler.GetDescription())

	// Create test import lists
	testImportLists := []models.ImportList{
		{ID: 1, Name: "Test List 1", Enabled: true},
		{ID: 2, Name: "Test List 2", Enabled: true},
	}

	syncResult := &models.ImportListSyncResult{
		ImportListID:   1,
		ImportListName: "Test List 1",
		MoviesAdded:    5,
		MoviesTotal:    10,
	}

	// Setup mock expectations for syncing all lists
	importListService.On("GetEnabledImportLists").Return(testImportLists, nil)
	importListService.On("SyncImportList", 1).Return(syncResult, nil)
	importListService.On("SyncImportList", 2).Return(syncResult, nil)

	// Create task without specific import list ID
	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{},
	}

	// Track progress updates
	var progressUpdates []struct {
		percent int
		message string
	}
	updateProgress := func(percent int, message string) {
		progressUpdates = append(progressUpdates, struct {
			percent int
			message string
		}{percent, message})
	}

	// Execute handler
	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	// Assertions
	require.NoError(t, err)
	importListService.AssertExpectations(t)

	// Verify progress was updated
	assert.GreaterOrEqual(t, len(progressUpdates), 3)
	assert.Equal(t, 0, progressUpdates[0].percent)
	assert.Equal(t, "Starting import list sync", progressUpdates[0].message)
	assert.Equal(t, 100, progressUpdates[len(progressUpdates)-1].percent)
}

func TestSyncImportListHandler_SpecificList(t *testing.T) {
	importListService := new(MockImportListService)
	handler := NewSyncImportListHandler(importListService)

	// Create test import list
	testImportList := &models.ImportList{
		ID:      1,
		Name:    "Specific List",
		Enabled: true,
	}

	syncResult := &models.ImportListSyncResult{
		ImportListID:   1,
		ImportListName: "Specific List",
		MoviesAdded:    3,
		MoviesTotal:    5,
	}

	// Setup mock expectations for specific list
	importListService.On("GetImportListByID", 1).Return(testImportList, nil)
	importListService.On("SyncImportList", 1).Return(syncResult, nil)

	// Create task with specific import list ID
	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{"importListId": 1},
	}

	updateProgress := func(percent int, message string) {}

	// Execute handler
	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	// Assertions
	require.NoError(t, err)
	importListService.AssertExpectations(t)
}

// testTaskHandler is a helper function to test task handlers with common functionality
func testTaskHandler(t *testing.T, handler TaskHandler, expectedName, expectedDescription, expectedStartMessage string) {
	// Test basic properties
	assert.Equal(t, expectedName, handler.GetName())
	assert.Equal(t, expectedDescription, handler.GetDescription())

	// Create task
	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{},
	}

	// Track progress updates
	var progressUpdates []struct {
		percent int
		message string
	}
	updateProgress := func(percent int, message string) {
		progressUpdates = append(progressUpdates, struct {
			percent int
			message string
		}{percent, message})
	}

	// Execute handler
	ctx := context.Background()
	err := handler.Execute(ctx, task, updateProgress)

	// We expect no error
	require.NoError(t, err)

	// Verify progress was updated
	assert.GreaterOrEqual(t, len(progressUpdates), 2)
	assert.Equal(t, 0, progressUpdates[0].percent)
	assert.Equal(t, expectedStartMessage, progressUpdates[0].message)
	assert.Equal(t, 100, progressUpdates[len(progressUpdates)-1].percent)
}

func TestHealthCheckHandler(t *testing.T) {
	container := &Container{
		DB:     nil, // DB will be mocked if needed
		Config: nil,
		Logger: nil,
	}
	handler := NewHealthCheckHandler(container)
	testTaskHandler(t, handler, "HealthCheck", "Performs system health checks", "Starting health check")
}

func TestCleanupHandler(t *testing.T) {
	container := &Container{
		DB:     nil, // DB will be mocked if needed
		Config: nil,
		Logger: nil,
	}
	handler := NewCleanupHandler(container)
	testTaskHandler(t, handler, "Cleanup", "Performs cleanup tasks like removing old logs and completed downloads", "Starting cleanup")
}

func TestTaskHandlerCancellation(t *testing.T) {
	movieService := new(MockMovieService)
	metadataService := new(MockMetadataService)
	handler := NewRefreshMovieHandler(movieService, metadataService)

	// Create context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	task := &models.Task{
		ID:   1,
		Body: models.TaskBody{"movieId": 1},
	}

	updateProgress := func(percent int, message string) {
		// Simulate slow progress update
		time.Sleep(20 * time.Millisecond)
	}

	// Setup mocks - but execution should be cancelled before they're called
	testMovie := &models.Movie{ID: 1, Title: "Test Movie"}
	movieService.On("GetByID", 1).Return(testMovie, nil).Maybe()
	metadataService.On("RefreshMovieMetadata", 1).Return(nil).Maybe()
	movieService.On("Update", testMovie).Return(nil).Maybe()

	// Execute handler - should be cancelled
	err := handler.Execute(ctx, task, updateProgress)

	// Should return context cancellation error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context")
}
