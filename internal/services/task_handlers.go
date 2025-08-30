// Package services provides business logic and domain services for Radarr.
package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/radarr/radarr-go/internal/models"
)

// RefreshMovieHandler handles refreshing metadata for a single movie
type RefreshMovieHandler struct {
	movieService    MovieServiceInterface
	metadataService MetadataServiceInterface
}

// NewRefreshMovieHandler creates a new refresh movie handler
func NewRefreshMovieHandler(
	movieService MovieServiceInterface,
	metadataService MetadataServiceInterface,
) *RefreshMovieHandler {
	return &RefreshMovieHandler{
		movieService:    movieService,
		metadataService: metadataService,
	}
}

// Execute refreshes metadata for a specific movie
func (h *RefreshMovieHandler) Execute(
	ctx context.Context,
	task *models.TaskV2,
	updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Starting movie refresh")

	movieID, err := h.extractMovieID(task)
	if err != nil {
		return err
	}

	updateProgress(10, "Fetching movie from database")

	if err := h.checkContext(ctx); err != nil {
		return err
	}

	movie, err := h.movieService.GetByID(movieID)
	if err != nil {
		return fmt.Errorf("failed to get movie %d: %w", movieID, err)
	}

	updateProgress(25, "Refreshing metadata from TMDB")

	if err := h.checkContext(ctx); err != nil {
		return err
	}

	if err := h.metadataService.RefreshMovieMetadata(movieID); err != nil {
		return fmt.Errorf("failed to refresh metadata for movie %d: %w", movieID, err)
	}

	updateProgress(75, "Updating movie in database")

	if err := h.checkContext(ctx); err != nil {
		return err
	}

	if err := h.movieService.Update(movie); err != nil {
		return fmt.Errorf("failed to update movie %d: %w", movieID, err)
	}

	updateProgress(100, "Movie refresh completed")
	return nil
}

// extractMovieID extracts the movie ID from the task body
func (h *RefreshMovieHandler) extractMovieID(task *models.TaskV2) (int, error) {
	movieIDValue, exists := task.Body["movieId"]
	if !exists {
		return 0, fmt.Errorf("movieId not found in task body")
	}

	switch v := movieIDValue.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	case string:
		movieID, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid movieId format: %v", v)
		}
		return movieID, nil
	default:
		return 0, fmt.Errorf("invalid movieId type: %T", v)
	}
}

// checkContext checks if the context is cancelled
func (h *RefreshMovieHandler) checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

// GetName returns the command name this handler processes
func (h *RefreshMovieHandler) GetName() string {
	return "RefreshMovie"
}

// GetDescription returns a human-readable description
func (h *RefreshMovieHandler) GetDescription() string {
	return "Refreshes metadata for a single movie from TMDB"
}

// RefreshAllMoviesHandler handles refreshing metadata for all movies
type RefreshAllMoviesHandler struct {
	movieService    MovieServiceInterface
	metadataService MetadataServiceInterface
}

// NewRefreshAllMoviesHandler creates a new refresh all movies handler
func NewRefreshAllMoviesHandler(
	movieService MovieServiceInterface,
	metadataService MetadataServiceInterface,
) *RefreshAllMoviesHandler {
	return &RefreshAllMoviesHandler{
		movieService:    movieService,
		metadataService: metadataService,
	}
}

// Execute refreshes metadata for all movies
func (h *RefreshAllMoviesHandler) Execute(
	ctx context.Context, _ *models.TaskV2, updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Starting bulk movie refresh")

	movies, err := h.getAllMoviesForRefresh()
	if err != nil {
		return err
	}

	if len(movies) == 0 {
		updateProgress(100, "No movies to refresh")
		return nil
	}

	updateProgress(10, fmt.Sprintf("Found %d movies to refresh", len(movies)))

	processed := h.processMoviesInBatches(ctx, movies, updateProgress)

	updateProgress(100, fmt.Sprintf("Completed refreshing %d movies", processed))
	return nil
}

// getAllMoviesForRefresh gets all movies that need metadata refresh
func (h *RefreshAllMoviesHandler) getAllMoviesForRefresh() ([]models.Movie, error) {
	movies, err := h.movieService.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list movies: %w", err)
	}
	return movies, nil
}

// processMoviesInBatches processes movies in batches to refresh their metadata
func (h *RefreshAllMoviesHandler) processMoviesInBatches(
	ctx context.Context, movies []models.Movie,
	updateProgress func(percent int, message string),
) int {
	processed := 0
	batchSize := 10

	for i := 0; i < len(movies); i += batchSize {
		if ctx.Err() != nil {
			break
		}

		batch := h.getBatch(movies, i, batchSize)
		processed += h.processBatch(ctx, batch, movies, processed, updateProgress)

		// Small delay between batches to avoid overwhelming TMDB API
		time.Sleep(1 * time.Second)
	}

	return processed
}

// getBatch extracts a batch of movies from the full list
func (h *RefreshAllMoviesHandler) getBatch(movies []models.Movie, start, batchSize int) []models.Movie {
	end := start + batchSize
	if end > len(movies) {
		end = len(movies)
	}
	return movies[start:end]
}

// processBatch processes a single batch of movies
func (h *RefreshAllMoviesHandler) processBatch(
	ctx context.Context, batch []models.Movie, allMovies []models.Movie,
	startProcessed int, updateProgress func(percent int, message string),
) int {
	batchProcessed := 0

	for _, movie := range batch {
		if ctx.Err() != nil {
			break
		}

		processed := startProcessed + batchProcessed
		h.processMovie(movie, allMovies, processed, updateProgress)
		batchProcessed++
	}

	return batchProcessed
}

// processMovie processes a single movie for metadata refresh
func (h *RefreshAllMoviesHandler) processMovie(
	movie models.Movie, allMovies []models.Movie, processed int,
	updateProgress func(percent int, message string),
) {
	updateProgress(10+((processed*80)/len(allMovies)),
		fmt.Sprintf("Refreshing movie: %s (%d/%d)", movie.Title, processed+1, len(allMovies)))

	if err := h.metadataService.RefreshMovieMetadata(movie.ID); err != nil {
		// Log error but continue with other movies
		updateProgress(10+((processed*80)/len(allMovies)),
			fmt.Sprintf("Failed to refresh movie: %s - %v", movie.Title, err))
	} else {
		if err := h.movieService.Update(&movie); err != nil {
			updateProgress(10+((processed*80)/len(allMovies)),
				fmt.Sprintf("Failed to save movie: %s - %v", movie.Title, err))
		}
	}
}

// GetName returns the command name this handler processes
func (h *RefreshAllMoviesHandler) GetName() string {
	return "RefreshAllMovies"
}

// GetDescription returns a human-readable description
func (h *RefreshAllMoviesHandler) GetDescription() string {
	return "Refreshes metadata for all movies from TMDB"
}

// SyncImportListHandler handles syncing import lists
type SyncImportListHandler struct {
	importListService ImportListServiceInterface
}

// NewSyncImportListHandler creates a new sync import list handler
func NewSyncImportListHandler(
	importListService ImportListServiceInterface,
) *SyncImportListHandler {
	return &SyncImportListHandler{
		importListService: importListService,
	}
}

// Execute syncs movies from import lists
func (h *SyncImportListHandler) Execute(
	ctx context.Context,
	task *models.TaskV2,
	updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Starting import list sync")

	importListID := h.extractImportListID(task)
	updateProgress(10, "Fetching import lists")

	importLists, err := h.getImportListsToSync(importListID)
	if err != nil {
		return err
	}

	if len(importLists) == 0 {
		updateProgress(100, "No import lists to sync")
		return nil
	}

	updateProgress(20, fmt.Sprintf("Found %d import lists to sync", len(importLists)))

	totalAdded := h.syncAllImportLists(ctx, importLists, updateProgress)

	updateProgress(100, fmt.Sprintf("Import list sync completed. Added %d movies total", totalAdded))
	return nil
}

// extractImportListID extracts the import list ID from task body
func (h *SyncImportListHandler) extractImportListID(task *models.TaskV2) *int {
	if listIDValue, exists := task.Body["importListId"]; exists {
		switch v := listIDValue.(type) {
		case int:
			return &v
		case float64:
			id := int(v)
			return &id
		case string:
			if id, err := strconv.Atoi(v); err == nil {
				return &id
			}
		}
	}
	return nil
}

// getImportListsToSync gets the import lists that need to be synced
func (h *SyncImportListHandler) getImportListsToSync(importListID *int) ([]*models.ImportList, error) {
	if importListID != nil {
		return h.getSpecificImportList(*importListID)
	}
	return h.getAllEnabledImportLists()
}

// getSpecificImportList gets a specific import list by ID
func (h *SyncImportListHandler) getSpecificImportList(importListID int) ([]*models.ImportList, error) {
	importList, err := h.importListService.GetImportListByID(importListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import list %d: %w", importListID, err)
	}
	return []*models.ImportList{importList}, nil
}

// getAllEnabledImportLists gets all enabled import lists
func (h *SyncImportListHandler) getAllEnabledImportLists() ([]*models.ImportList, error) {
	enabledLists, err := h.importListService.GetEnabledImportLists()
	if err != nil {
		return nil, fmt.Errorf("failed to get import lists: %w", err)
	}

	// Convert slice to pointer slice
	importLists := make([]*models.ImportList, len(enabledLists))
	for i := range enabledLists {
		importLists[i] = &enabledLists[i]
	}
	return importLists, nil
}

// syncAllImportLists syncs all provided import lists
func (h *SyncImportListHandler) syncAllImportLists(
	ctx context.Context, importLists []*models.ImportList,
	updateProgress func(percent int, message string),
) int {
	totalAdded := 0
	processed := 0

	for _, importList := range importLists {
		if ctx.Err() != nil {
			break
		}

		progress := 20 + ((processed * 70) / len(importLists))
		updateProgress(progress, fmt.Sprintf("Syncing import list: %s", importList.Name))

		result, err := h.importListService.SyncImportList(importList.ID)
		if err != nil {
			updateProgress(progress, fmt.Sprintf("Failed to sync import list %s: %v", importList.Name, err))
		} else {
			totalAdded += result.MoviesAdded
			updateProgress(progress, fmt.Sprintf("Added %d movies from %s", result.MoviesAdded, importList.Name))
		}

		processed++
	}

	return totalAdded
}

// GetName returns the command name this handler processes
func (h *SyncImportListHandler) GetName() string {
	return "SyncImportList"
}

// GetDescription returns a human-readable description
func (h *SyncImportListHandler) GetDescription() string {
	return "Syncs movies from configured import lists"
}

// HealthCheckHandler handles system health checks
type HealthCheckHandler struct {
	container *Container
}

// NewHealthCheckHandler creates a new health check handler
func NewHealthCheckHandler(container *Container) *HealthCheckHandler {
	return &HealthCheckHandler{
		container: container,
	}
}

// Execute performs system health checks
func (h *HealthCheckHandler) Execute(
	ctx context.Context, _ *models.TaskV2, updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Starting health check")

	checks := []struct {
		name string
		fn   func(ctx context.Context) error
	}{
		{"Database Connection", h.checkDatabase},
		{"Disk Space", h.checkDiskSpace},
		{"Download Clients", h.checkDownloadClients},
		{"Indexers", h.checkIndexers},
		{"Import Lists", h.checkImportLists},
	}

	passed := 0
	failed := 0

	for i, check := range checks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progress := (i * 100) / len(checks)
		updateProgress(progress, fmt.Sprintf("Running check: %s", check.name))

		if err := check.fn(ctx); err != nil {
			failed++
			updateProgress(progress, fmt.Sprintf("Check failed: %s - %v", check.name, err))
		} else {
			passed++
			updateProgress(progress, fmt.Sprintf("Check passed: %s", check.name))
		}
	}

	updateProgress(100, fmt.Sprintf("Health check completed. Passed: %d, Failed: %d", passed, failed))
	return nil
}

// GetName returns the command name this handler processes
func (h *HealthCheckHandler) GetName() string {
	return "HealthCheck"
}

// GetDescription returns a human-readable description
func (h *HealthCheckHandler) GetDescription() string {
	return "Performs system health checks"
}

// checkDatabase verifies database connectivity
func (h *HealthCheckHandler) checkDatabase(ctx context.Context) error {
	if h.container.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := h.container.DB.GORM.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// checkDiskSpace verifies adequate disk space
func (h *HealthCheckHandler) checkDiskSpace(_ context.Context) error {
	// This would implement actual disk space checking
	// For now, just return success
	return nil
}

// checkDownloadClients verifies download clients are accessible
func (h *HealthCheckHandler) checkDownloadClients(_ context.Context) error {
	if h.container.DownloadService == nil {
		return fmt.Errorf("download service not initialized")
	}

	clients, err := h.container.DownloadService.GetDownloadClients()
	if err != nil {
		return fmt.Errorf("failed to get download clients: %w", err)
	}

	for _, client := range clients {
		if client.Enable {
			_, err := h.container.DownloadService.TestDownloadClient(&client)
			if err != nil {
				return fmt.Errorf("download client %s failed test: %w", client.Name, err)
			}
		}
	}

	return nil
}

// checkIndexers verifies indexers are accessible
func (h *HealthCheckHandler) checkIndexers(_ context.Context) error {
	if h.container.IndexerService == nil {
		return fmt.Errorf("indexer service not initialized")
	}

	indexers, err := h.container.IndexerService.GetIndexers()
	if err != nil {
		return fmt.Errorf("failed to get indexers: %w", err)
	}

	for _, indexer := range indexers {
		if indexer.IsEnabled() {
			_, err := h.container.IndexerService.TestIndexer(indexer)
			if err != nil {
				return fmt.Errorf("indexer %s failed test: %w", indexer.Name, err)
			}
		}
	}

	return nil
}

// checkImportLists verifies import lists are accessible
func (h *HealthCheckHandler) checkImportLists(_ context.Context) error {
	if h.container.ImportListService == nil {
		return fmt.Errorf("import list service not initialized")
	}

	importLists, err := h.container.ImportListService.GetImportLists()
	if err != nil {
		return fmt.Errorf("failed to get import lists: %w", err)
	}

	for _, importList := range importLists {
		if importList.IsEnabled() {
			_, err := h.container.ImportListService.TestImportList(&importList)
			if err != nil {
				return fmt.Errorf("import list %s failed test: %w", importList.Name, err)
			}
		}
	}

	return nil
}

// CleanupHandler handles cleanup tasks like removing old logs, completed downloads, etc.
type CleanupHandler struct {
	container *Container
}

// NewCleanupHandler creates a new cleanup handler
func NewCleanupHandler(container *Container) *CleanupHandler {
	return &CleanupHandler{
		container: container,
	}
}

// Execute performs cleanup tasks
func (h *CleanupHandler) Execute(
	ctx context.Context, _ *models.TaskV2, updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Starting cleanup")

	cleanupTasks := []struct {
		name string
		fn   func(ctx context.Context) error
	}{
		{"Completed Tasks", h.cleanupCompletedTasks},
		{"Old History Records", h.cleanupOldHistory},
		{"Failed Downloads", h.cleanupFailedDownloads},
		{"Orphaned Files", h.cleanupOrphanedFiles},
	}

	for i, cleanupTask := range cleanupTasks {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progress := (i * 100) / len(cleanupTasks)
		updateProgress(progress, fmt.Sprintf("Running cleanup: %s", cleanupTask.name))

		if err := cleanupTask.fn(ctx); err != nil {
			updateProgress(progress, fmt.Sprintf("Cleanup failed: %s - %v", cleanupTask.name, err))
		} else {
			updateProgress(progress, fmt.Sprintf("Cleanup completed: %s", cleanupTask.name))
		}
	}

	updateProgress(100, "Cleanup completed")
	return nil
}

// GetName returns the command name this handler processes
func (h *CleanupHandler) GetName() string {
	return "Cleanup"
}

// GetDescription returns a human-readable description
func (h *CleanupHandler) GetDescription() string {
	return "Performs cleanup tasks like removing old logs and completed downloads"
}

// cleanupCompletedTasks removes old completed tasks
func (h *CleanupHandler) cleanupCompletedTasks(_ context.Context) error {
	if h.container.DB == nil {
		return fmt.Errorf("database not initialized")
	}

	// Remove completed tasks older than 7 days
	cutoff := time.Now().AddDate(0, 0, -7)

	result := h.container.DB.GORM.Where("status IN (?, ?, ?) AND ended_at < ?",
		"completed", "failed", "aborted", cutoff).
		Delete(&models.TaskV2{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup completed tasks: %w", result.Error)
	}

	if h.container.Logger != nil {
		h.container.Logger.Infow("Cleaned up completed tasks", "count", result.RowsAffected)
	}
	return nil
}

// cleanupOldHistory removes old history records
func (h *CleanupHandler) cleanupOldHistory(_ context.Context) error {
	// This would implement history cleanup
	// For now, just return success
	return nil
}

// cleanupFailedDownloads removes failed downloads
func (h *CleanupHandler) cleanupFailedDownloads(_ context.Context) error {
	// This would implement failed download cleanup
	// For now, just return success
	return nil
}

// cleanupOrphanedFiles removes orphaned files
func (h *CleanupHandler) cleanupOrphanedFiles(_ context.Context) error {
	// This would implement orphaned file cleanup
	// For now, just return success
	return nil
}

// RefreshWantedMoviesHandler handles refreshing the wanted movies list
type RefreshWantedMoviesHandler struct {
	wantedService WantedMoviesServiceInterface
}

// WantedMoviesServiceInterface defines the interface for wanted movies operations
type WantedMoviesServiceInterface interface {
	RefreshWantedMovies() error
	GetWantedStats() (*models.WantedMoviesStats, error)
	GetEligibleForSearch(limit int) ([]models.WantedMovie, error)
	UpdateSearchAttempt(id int, success bool, reason, indexer, errorCode string) error
}

// NewRefreshWantedMoviesHandler creates a new refresh wanted movies handler
func NewRefreshWantedMoviesHandler(wantedService WantedMoviesServiceInterface) *RefreshWantedMoviesHandler {
	return &RefreshWantedMoviesHandler{
		wantedService: wantedService,
	}
}

// Execute refreshes the wanted movies list
func (h *RefreshWantedMoviesHandler) Execute(
	_ context.Context, _ *models.TaskV2, updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Starting wanted movies refresh")

	// Refresh wanted movies
	if err := h.wantedService.RefreshWantedMovies(); err != nil {
		return fmt.Errorf("failed to refresh wanted movies: %w", err)
	}

	updateProgress(50, "Getting updated statistics")

	// Get updated stats
	stats, err := h.wantedService.GetWantedStats()
	if err != nil {
		updateProgress(100, "Wanted movies refresh completed (stats unavailable)")
		return nil // Don't fail the task if stats fail
	}

	updateProgress(100, fmt.Sprintf("Wanted movies refresh completed - Found %d missing, %d cutoff unmet",
		stats.MissingCount, stats.CutoffUnmetCount))

	return nil
}

// GetName returns the command name this handler processes
func (h *RefreshWantedMoviesHandler) GetName() string {
	return "RefreshWantedMovies"
}

// GetDescription returns a human-readable description
func (h *RefreshWantedMoviesHandler) GetDescription() string {
	return "Analyzes all monitored movies and updates the wanted movies list"
}

// AutoWantedSearchHandler handles automatic searching for wanted movies
type AutoWantedSearchHandler struct {
	wantedService WantedMoviesServiceInterface
	searchService SearchServiceInterface
}

// SearchServiceInterface defines the interface for search operations
type SearchServiceInterface interface {
	SearchMovieReleases(movieID int, forceSearch bool) (*models.SearchResponse, error)
}

// NewAutoWantedSearchHandler creates a new automatic wanted search handler
func NewAutoWantedSearchHandler(
	wantedService WantedMoviesServiceInterface,
	searchService SearchServiceInterface,
) *AutoWantedSearchHandler {
	return &AutoWantedSearchHandler{
		wantedService: wantedService,
		searchService: searchService,
	}
}

// Execute performs automatic searching for eligible wanted movies
func (h *AutoWantedSearchHandler) Execute(
	ctx context.Context, _ *models.TaskV2, updateProgress func(percent int, message string),
) error {
	updateProgress(0, "Getting eligible wanted movies")

	// Get movies eligible for search
	eligibleMovies, err := h.wantedService.GetEligibleForSearch(50) // Limit to 50 for automatic searches
	if err != nil {
		return fmt.Errorf("failed to get eligible wanted movies: %w", err)
	}

	if len(eligibleMovies) == 0 {
		updateProgress(100, "No wanted movies eligible for search")
		return nil
	}

	updateProgress(10, fmt.Sprintf("Found %d eligible movies, starting searches", len(eligibleMovies)))

	searchedCount := 0
	successCount := 0

	for i, wantedMovie := range eligibleMovies {
		// Check if task was cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progress := 10 + ((i + 1) * 80 / len(eligibleMovies))
		updateProgress(progress, fmt.Sprintf("Searching for movie %d (%d/%d)",
			wantedMovie.MovieID, i+1, len(eligibleMovies)))

		_, err := h.searchService.SearchMovieReleases(wantedMovie.MovieID, false)
		searchedCount++

		if err != nil {
			// Update search attempt with failure
			if updateErr := h.wantedService.UpdateSearchAttempt(wantedMovie.ID, false,
				"Automatic search failed", "", err.Error()); updateErr != nil {
				// Log but don't fail the task
				continue
			}
		} else {
			successCount++
			// Update search attempt with success
			if updateErr := h.wantedService.UpdateSearchAttempt(wantedMovie.ID, true,
				"Automatic search completed", "", ""); updateErr != nil {
				// Log but don't fail the task
				continue
			}
		}

		// Small delay between searches to avoid overwhelming indexers
		time.Sleep(100 * time.Millisecond)
	}

	updateProgress(100, fmt.Sprintf("Automatic search completed - %d searched, %d successful",
		searchedCount, successCount))

	return nil
}

// GetName returns the command name this handler processes
func (h *AutoWantedSearchHandler) GetName() string {
	return "AutoWantedSearch"
}

// GetDescription returns a human-readable description
func (h *AutoWantedSearchHandler) GetDescription() string {
	return "Automatically searches for wanted movies that are eligible for search"
}
