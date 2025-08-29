package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// DatabaseHealthChecker checks database connectivity and performance
type DatabaseHealthChecker struct {
	db     *database.Database
	logger *logger.Logger
	config models.HealthCheckConfig
}

// Name returns the human-readable name of this health checker
func (d *DatabaseHealthChecker) Name() string {
	return "Database Connectivity"
}

// Type returns the health check type identifier
func (d *DatabaseHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeDatabase
}

// IsEnabled returns whether this health checker is enabled
func (d *DatabaseHealthChecker) IsEnabled() bool {
	return d.config.Enabled
}

// GetInterval returns the check interval for this health checker
func (d *DatabaseHealthChecker) GetInterval() time.Duration {
	return d.config.Interval
}

// Check performs database connectivity health check
func (d *DatabaseHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
	start := time.Now()
	result := models.HealthCheckExecution{
		Type:      d.Type(),
		Source:    d.Name(),
		Status:    models.HealthStatusHealthy,
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Test database connection
	sqlDB, err := d.db.GORM.DB()
	if err != nil {
		return d.createConnectionErrorResult(result, err)
	}

	// Test database ping
	pingDuration, err := d.testDatabasePing(ctx, sqlDB)
	if err != nil {
		return d.createPingErrorResult(result, err)
	}

	// Test database query
	queryDuration, movieCount, err := d.testDatabaseQuery(ctx)
	if err != nil {
		return d.createQueryErrorResult(result, err)
	}

	// Populate performance details
	d.populatePerformanceDetails(&result, sqlDB, pingDuration, queryDuration, movieCount)

	// Evaluate health and create issues
	issues := d.evaluatePerformanceIssues(pingDuration, queryDuration, sqlDB.Stats())
	d.finalizeResult(&result, issues)

	return result
}

func (d *DatabaseHealthChecker) createConnectionErrorResult(
	result models.HealthCheckExecution, err error,
) models.HealthCheckExecution {
	result.Status = models.HealthStatusCritical
	result.Message = "Failed to get database connection"
	result.Error = err
	result.Issues = []models.HealthIssue{{
		Type:     d.Type(),
		Source:   d.Name(),
		Severity: models.HealthSeverityCritical,
		Message:  fmt.Sprintf("Database connection error: %v", err),
	}}
	return result
}

func (d *DatabaseHealthChecker) testDatabasePing(ctx context.Context, sqlDB *sql.DB) (time.Duration, error) {
	pingCtx, cancel := context.WithTimeout(ctx, d.config.DatabaseTimeoutThreshold)
	defer cancel()

	pingStart := time.Now()
	err := sqlDB.PingContext(pingCtx)
	return time.Since(pingStart), err
}

func (d *DatabaseHealthChecker) createPingErrorResult(
	result models.HealthCheckExecution, err error,
) models.HealthCheckExecution {
	result.Status = models.HealthStatusCritical
	result.Message = "Database ping failed"
	result.Error = err
	result.Issues = []models.HealthIssue{{
		Type:     d.Type(),
		Source:   d.Name(),
		Severity: models.HealthSeverityCritical,
		Message:  fmt.Sprintf("Database ping failed: %v", err),
	}}
	return result
}

func (d *DatabaseHealthChecker) testDatabaseQuery(ctx context.Context) (time.Duration, int64, error) {
	queryStart := time.Now()
	var count int64
	err := d.db.GORM.WithContext(ctx).Raw("SELECT COUNT(*) FROM movies").Scan(&count).Error
	return time.Since(queryStart), count, err
}

func (d *DatabaseHealthChecker) createQueryErrorResult(
	result models.HealthCheckExecution, err error,
) models.HealthCheckExecution {
	result.Status = models.HealthStatusError
	result.Message = "Database query test failed"
	result.Error = err
	result.Issues = []models.HealthIssue{{
		Type:     d.Type(),
		Source:   d.Name(),
		Severity: models.HealthSeverityError,
		Message:  fmt.Sprintf("Database query failed: %v", err),
	}}
	return result
}

func (d *DatabaseHealthChecker) populatePerformanceDetails(
	result *models.HealthCheckExecution, sqlDB *sql.DB, pingDuration, queryDuration time.Duration, movieCount int64,
) {
	stats := sqlDB.Stats()
	result.Details["ping_duration_ms"] = pingDuration.Milliseconds()
	result.Details["query_duration_ms"] = queryDuration.Milliseconds()
	result.Details["open_connections"] = stats.OpenConnections
	result.Details["in_use"] = stats.InUse
	result.Details["idle"] = stats.Idle
	result.Details["wait_count"] = stats.WaitCount
	result.Details["wait_duration_ms"] = stats.WaitDuration.Milliseconds()
	result.Details["max_idle_closed"] = stats.MaxIdleClosed
	result.Details["max_lifetime_closed"] = stats.MaxLifetimeClosed
	result.Details["movie_count"] = movieCount
}

func (d *DatabaseHealthChecker) evaluatePerformanceIssues(
	pingDuration, queryDuration time.Duration, stats sql.DBStats,
) []models.HealthIssue {
	var issues []models.HealthIssue

	if pingDuration > d.config.DatabaseTimeoutThreshold {
		issues = append(issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Database ping is slow (%v > %v)", pingDuration, d.config.DatabaseTimeoutThreshold),
		})
	}

	if queryDuration > 2*time.Second {
		issues = append(issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Database query is slow (%v)", queryDuration),
		})
	}

	if stats.WaitCount > 0 && stats.WaitDuration > 100*time.Millisecond {
		issues = append(issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityWarning,
			Message: fmt.Sprintf(
				"Database connection pool experiencing waits (count: %d, duration: %v)",
				stats.WaitCount, stats.WaitDuration),
		})
	}

	return issues
}

func (d *DatabaseHealthChecker) finalizeResult(result *models.HealthCheckExecution, issues []models.HealthIssue) {
	result.Issues = issues
	if len(issues) == 0 {
		result.Message = "Database is healthy"
		return
	}

	result.Message = fmt.Sprintf("Database has %d issue(s)", len(issues))
	// Update status based on issues
	for _, issue := range issues {
		if issue.Severity == models.HealthSeverityCritical {
			result.Status = models.HealthStatusCritical
			return
		}
		if issue.Severity == models.HealthSeverityError && result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusError
		}
		if issue.Severity == models.HealthSeverityWarning && result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
	}
}

// DiskSpaceHealthChecker checks available disk space
type DiskSpaceHealthChecker struct {
	db           *database.Database
	config       *config.Config
	logger       *logger.Logger
	healthConfig models.HealthCheckConfig
}

// Name returns the human-readable name of this health checker
func (d *DiskSpaceHealthChecker) Name() string {
	return "Disk Space"
}

// Type returns the health check type identifier
func (d *DiskSpaceHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeDiskSpace
}

// IsEnabled returns whether this health checker is enabled
func (d *DiskSpaceHealthChecker) IsEnabled() bool {
	return d.healthConfig.Enabled
}

// GetInterval returns the check interval for this health checker
func (d *DiskSpaceHealthChecker) GetInterval() time.Duration {
	return d.healthConfig.Interval
}

// Check performs disk space health check
func (d *DiskSpaceHealthChecker) Check(_ context.Context) models.HealthCheckExecution {
	result := models.HealthCheckExecution{
		Type:      d.Type(),
		Source:    d.Name(),
		Status:    models.HealthStatusHealthy,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var issues []models.HealthIssue
	var pathsChecked []string

	// Check data directory
	dataDir := d.config.Storage.DataDirectory
	d.checkPath(dataDir, &result, &issues)
	pathsChecked = append(pathsChecked, dataDir)

	// TODO: Get root folders from config service and check them
	// For now, just check some common paths if they exist
	commonPaths := []string{
		"/tmp",
		"/var/log",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			d.checkPath(path, &result, &issues)
			pathsChecked = append(pathsChecked, path)
		}
	}

	result.Details["paths_checked"] = pathsChecked
	result.Issues = issues

	if len(issues) == 0 {
		result.Message = "All disk spaces are healthy"
	} else {
		result.Message = fmt.Sprintf("Found %d disk space issue(s)", len(issues))
	}

	return result
}

func (d *DiskSpaceHealthChecker) checkPath(
	path string, result *models.HealthCheckExecution, issues *[]models.HealthIssue,
) {
	diskInfo := d.getDiskSpaceInfo(path)

	result.Details[fmt.Sprintf("%s_free_bytes", filepath.Base(path))] = diskInfo.FreeBytes
	result.Details[fmt.Sprintf("%s_total_bytes", filepath.Base(path))] = diskInfo.TotalBytes
	result.Details[fmt.Sprintf("%s_usage_percent", filepath.Base(path))] = diskInfo.UsagePercent

	// Check against thresholds
	if diskInfo.FreeBytes < d.healthConfig.DiskSpaceCriticalThreshold {
		*issues = append(*issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityCritical,
			Message: fmt.Sprintf(
				"Critical disk space on %s: %.2f GB free (%.1f%% used)",
				path, float64(diskInfo.FreeBytes)/1024/1024/1024, diskInfo.UsagePercent,
			),
		})
		result.Status = models.HealthStatusCritical
	} else if diskInfo.FreeBytes < d.healthConfig.DiskSpaceWarningThreshold {
		*issues = append(*issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityWarning,
			Message: fmt.Sprintf(
				"Low disk space on %s: %.2f GB free (%.1f%% used)",
				path, float64(diskInfo.FreeBytes)/1024/1024/1024, diskInfo.UsagePercent,
			),
		})
		if result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
	}
}

// getDiskSpaceInfo returns disk space information for a path
func (d *DiskSpaceHealthChecker) getDiskSpaceInfo(path string) *models.DiskSpaceInfo {
	// This is a placeholder implementation
	// In production, use platform-specific implementations

	return &models.DiskSpaceInfo{
		Path:         path,
		FreeBytes:    10 * 1024 * 1024 * 1024,  // 10 GB free
		TotalBytes:   100 * 1024 * 1024 * 1024, // 100 GB total
		UsedBytes:    90 * 1024 * 1024 * 1024,  // 90 GB used
		UsagePercent: 90.0,
		IsAccessible: true,
		Warning:      false,
		Critical:     false,
	}
}

// SystemResourcesHealthChecker checks system resource usage
type SystemResourcesHealthChecker struct {
	logger *logger.Logger
	config models.HealthCheckConfig
}

// Name returns the human-readable name of this health checker
func (s *SystemResourcesHealthChecker) Name() string {
	return "System Resources"
}

// Type returns the health check type identifier
func (s *SystemResourcesHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeSystem
}

// IsEnabled returns whether this health checker is enabled
func (s *SystemResourcesHealthChecker) IsEnabled() bool {
	return s.config.Enabled
}

// GetInterval returns the check interval for this health checker
func (s *SystemResourcesHealthChecker) GetInterval() time.Duration {
	return s.config.Interval
}

// Check performs system resources health check
func (s *SystemResourcesHealthChecker) Check(_ context.Context) models.HealthCheckExecution {
	result := models.HealthCheckExecution{
		Type:      s.Type(),
		Source:    s.Name(),
		Status:    models.HealthStatusHealthy,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	// Collect memory statistics
	memoryStats := s.collectMemoryStats()
	s.populateMemoryDetails(&result, memoryStats)

	// Evaluate system resource issues
	issues := s.evaluateSystemIssues(memoryStats)
	s.finalizeSystemResult(&result, issues)

	return result
}

type memoryStats struct {
	usageMB      float64
	totalMB      float64
	usagePercent float64
	goroutines   int
	gcCycles     uint32
	nextGCMB     float64
}

func (s *SystemResourcesHealthChecker) collectMemoryStats() memoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUsageMB := float64(m.Alloc) / 1024 / 1024
	memoryTotalMB := float64(m.Sys) / 1024 / 1024
	memoryUsagePercent := (memoryUsageMB / memoryTotalMB) * 100

	return memoryStats{
		usageMB:      memoryUsageMB,
		totalMB:      memoryTotalMB,
		usagePercent: memoryUsagePercent,
		goroutines:   runtime.NumGoroutine(),
		gcCycles:     m.NumGC,
		nextGCMB:     float64(m.NextGC) / 1024 / 1024,
	}
}

func (s *SystemResourcesHealthChecker) populateMemoryDetails(result *models.HealthCheckExecution, stats memoryStats) {
	result.Details["memory_usage_mb"] = stats.usageMB
	result.Details["memory_total_mb"] = stats.totalMB
	result.Details["memory_usage_percent"] = stats.usagePercent
	result.Details["goroutines"] = stats.goroutines
	result.Details["gc_cycles"] = stats.gcCycles
	result.Details["next_gc_mb"] = stats.nextGCMB
}

func (s *SystemResourcesHealthChecker) evaluateSystemIssues(stats memoryStats) []models.HealthIssue {
	var issues []models.HealthIssue

	// Check memory usage thresholds
	if stats.usagePercent > 90 {
		issues = append(issues, models.HealthIssue{
			Type:     s.Type(),
			Source:   s.Name(),
			Severity: models.HealthSeverityCritical,
			Message: fmt.Sprintf(
				"Critical memory usage: %.1f%% (%.1f MB / %.1f MB)",
				stats.usagePercent, stats.usageMB, stats.totalMB,
			),
		})
	} else if stats.usagePercent > 80 {
		issues = append(issues, models.HealthIssue{
			Type:     s.Type(),
			Source:   s.Name(),
			Severity: models.HealthSeverityWarning,
			Message: fmt.Sprintf(
				"High memory usage: %.1f%% (%.1f MB / %.1f MB)",
				stats.usagePercent, stats.usageMB, stats.totalMB,
			),
		})
	}

	// Check goroutine count
	if stats.goroutines > 10000 {
		issues = append(issues, models.HealthIssue{
			Type:     s.Type(),
			Source:   s.Name(),
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("High goroutine count: %d", stats.goroutines),
		})
	}

	return issues
}

func (s *SystemResourcesHealthChecker) finalizeSystemResult(
	result *models.HealthCheckExecution, issues []models.HealthIssue,
) {
	result.Issues = issues
	if len(issues) == 0 {
		result.Message = "System resources are healthy"
		return
	}

	result.Message = fmt.Sprintf("Found %d system resource issue(s)", len(issues))
	// Update status based on issues
	for _, issue := range issues {
		if issue.Severity == models.HealthSeverityCritical {
			result.Status = models.HealthStatusCritical
			return
		}
		if issue.Severity == models.HealthSeverityWarning && result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
	}
}

// RootFolderHealthChecker checks accessibility of configured root folders
type RootFolderHealthChecker struct {
	db     *database.Database
	logger *logger.Logger
	config *config.Config
}

// Name returns the human-readable name of this health checker
func (r *RootFolderHealthChecker) Name() string {
	return "Root Folder Accessibility"
}

// Type returns the health check type identifier
func (r *RootFolderHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeRootFolder
}

// IsEnabled returns whether this health checker is enabled
func (r *RootFolderHealthChecker) IsEnabled() bool {
	return true // Always enabled
}

// GetInterval returns the check interval for this health checker
func (r *RootFolderHealthChecker) GetInterval() time.Duration {
	return 30 * time.Minute
}

// Check performs root folder accessibility health check
func (r *RootFolderHealthChecker) Check(_ context.Context) models.HealthCheckExecution {
	result := models.HealthCheckExecution{
		Type:      r.Type(),
		Source:    r.Name(),
		Status:    models.HealthStatusHealthy,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var issues []models.HealthIssue
	// TODO: Get root folders from config service
	// For now, check the data directory
	paths := []string{r.config.Storage.DataDirectory}
	checkedPaths := r.checkAllPaths(paths, &issues)

	result.Details["checked_paths"] = checkedPaths
	r.finalizeRootFolderResult(&result, issues)

	return result
}

func (r *RootFolderHealthChecker) checkAllPaths(paths []string, issues *[]models.HealthIssue) []string {
	var checkedPaths []string
	for _, path := range paths {
		checkedPaths = append(checkedPaths, path)
		r.checkSinglePath(path, issues)
	}
	return checkedPaths
}

func (r *RootFolderHealthChecker) checkSinglePath(path string, issues *[]models.HealthIssue) {
	// Check if path exists and is accessible
	info, err := os.Stat(path)
	if err != nil {
		*issues = append(*issues, models.HealthIssue{
			Type:     r.Type(),
			Source:   r.Name(),
			Severity: models.HealthSeverityCritical,
			Message:  fmt.Sprintf("Root folder not accessible: %s - %v", path, err),
		})
		return
	}

	// Check if it's a directory
	if !info.IsDir() {
		*issues = append(*issues, models.HealthIssue{
			Type:     r.Type(),
			Source:   r.Name(),
			Severity: models.HealthSeverityError,
			Message:  fmt.Sprintf("Root folder path is not a directory: %s", path),
		})
		return
	}

	// Check write permissions
	r.checkWritePermissions(path, issues)
}

func (r *RootFolderHealthChecker) checkWritePermissions(path string, issues *[]models.HealthIssue) {
	testFile := filepath.Join(path, ".radarr-health-check")
	if err := os.WriteFile(testFile, []byte("test"), 0400); err != nil {
		*issues = append(*issues, models.HealthIssue{
			Type:     r.Type(),
			Source:   r.Name(),
			Severity: models.HealthSeverityError,
			Message:  fmt.Sprintf("Root folder not writable: %s - %v", path, err),
		})
		return
	}

	// Clean up test file
	if err := os.Remove(testFile); err != nil {
		// Not critical if cleanup fails, just log it
		r.logger.Debug("Failed to clean up test file", "file", testFile, "error", err)
	}
}

func (r *RootFolderHealthChecker) finalizeRootFolderResult(
	result *models.HealthCheckExecution, issues []models.HealthIssue,
) {
	result.Issues = issues
	if len(issues) == 0 {
		result.Message = "All root folders are accessible"
		return
	}

	result.Message = fmt.Sprintf("Found %d root folder issue(s)", len(issues))
	// Update status based on issues
	for _, issue := range issues {
		if issue.Severity == models.HealthSeverityCritical {
			result.Status = models.HealthStatusCritical
			return
		}
		if issue.Severity == models.HealthSeverityError && result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusError
		}
	}
}
