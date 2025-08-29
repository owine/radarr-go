package services

import (
	"context"
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

func (d *DatabaseHealthChecker) Name() string {
	return "Database Connectivity"
}

func (d *DatabaseHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeDatabase
}

func (d *DatabaseHealthChecker) IsEnabled() bool {
	return d.config.Enabled
}

func (d *DatabaseHealthChecker) GetInterval() time.Duration {
	return d.config.Interval
}

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

	// Ping database with timeout
	pingCtx, cancel := context.WithTimeout(ctx, d.config.DatabaseTimeoutThreshold)
	defer cancel()

	pingStart := time.Now()
	if err := sqlDB.PingContext(pingCtx); err != nil {
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
	pingDuration := time.Since(pingStart)

	// Test a simple query
	queryStart := time.Now()
	var count int64
	if err := d.db.GORM.WithContext(ctx).Raw("SELECT COUNT(*) FROM movies").Scan(&count).Error; err != nil {
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
	queryDuration := time.Since(queryStart)

	// Check connection pool stats
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
	result.Details["movie_count"] = count

	// Evaluate performance and connection health
	var issues []models.HealthIssue

	if pingDuration > d.config.DatabaseTimeoutThreshold {
		issues = append(issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Database ping is slow (%v > %v)", pingDuration, d.config.DatabaseTimeoutThreshold),
		})
		result.Status = models.HealthStatusWarning
	}

	if queryDuration > 2*time.Second {
		issues = append(issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Database query is slow (%v)", queryDuration),
		})
		if result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
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
		if result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
	}

	result.Issues = issues
	if len(issues) == 0 {
		result.Message = "Database is healthy"
	} else {
		result.Message = fmt.Sprintf("Database has %d issue(s)", len(issues))
	}

	return result
}

// DiskSpaceHealthChecker checks available disk space
type DiskSpaceHealthChecker struct {
	db           *database.Database
	config       *config.Config
	logger       *logger.Logger
	healthConfig models.HealthCheckConfig
}

func (d *DiskSpaceHealthChecker) Name() string {
	return "Disk Space"
}

func (d *DiskSpaceHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeDiskSpace
}

func (d *DiskSpaceHealthChecker) IsEnabled() bool {
	return d.healthConfig.Enabled
}

func (d *DiskSpaceHealthChecker) GetInterval() time.Duration {
	return d.healthConfig.Interval
}

func (d *DiskSpaceHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
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
	if err := d.checkPath(dataDir, &result, &issues); err != nil {
		d.logger.Errorw("Failed to check data directory", "path", dataDir, "error", err)
	}
	pathsChecked = append(pathsChecked, dataDir)

	// TODO: Get root folders from config service and check them
	// For now, just check some common paths if they exist
	commonPaths := []string{
		"/tmp",
		"/var/log",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			if err := d.checkPath(path, &result, &issues); err != nil {
				d.logger.Errorw("Failed to check path", "path", path, "error", err)
			}
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
) error {
	diskInfo, err := d.getDiskSpaceInfo(path)
	if err != nil {
		*issues = append(*issues, models.HealthIssue{
			Type:     d.Type(),
			Source:   d.Name(),
			Severity: models.HealthSeverityError,
			Message:  fmt.Sprintf("Failed to check disk space for %s: %v", path, err),
		})
		if result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusError
		}
		return err
	}

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

	return nil
}

// getDiskSpaceInfo returns disk space information for a path
func (d *DiskSpaceHealthChecker) getDiskSpaceInfo(path string) (*models.DiskSpaceInfo, error) {
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
	}, nil
}

// SystemResourcesHealthChecker checks system resource usage
type SystemResourcesHealthChecker struct {
	logger *logger.Logger
	config models.HealthCheckConfig
}

func (s *SystemResourcesHealthChecker) Name() string {
	return "System Resources"
}

func (s *SystemResourcesHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeSystem
}

func (s *SystemResourcesHealthChecker) IsEnabled() bool {
	return s.config.Enabled
}

func (s *SystemResourcesHealthChecker) GetInterval() time.Duration {
	return s.config.Interval
}

func (s *SystemResourcesHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
	result := models.HealthCheckExecution{
		Type:      s.Type(),
		Source:    s.Name(),
		Status:    models.HealthStatusHealthy,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var issues []models.HealthIssue

	// Memory statistics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUsageMB := float64(m.Alloc) / 1024 / 1024
	memoryTotalMB := float64(m.Sys) / 1024 / 1024
	memoryUsagePercent := (memoryUsageMB / memoryTotalMB) * 100

	result.Details["memory_usage_mb"] = memoryUsageMB
	result.Details["memory_total_mb"] = memoryTotalMB
	result.Details["memory_usage_percent"] = memoryUsagePercent
	result.Details["goroutines"] = runtime.NumGoroutine()
	result.Details["gc_cycles"] = m.NumGC
	result.Details["next_gc_mb"] = float64(m.NextGC) / 1024 / 1024

	// Check memory usage thresholds
	if memoryUsagePercent > 90 {
		issues = append(issues, models.HealthIssue{
			Type:     s.Type(),
			Source:   s.Name(),
			Severity: models.HealthSeverityCritical,
			Message: fmt.Sprintf(
				"Critical memory usage: %.1f%% (%.1f MB / %.1f MB)",
				memoryUsagePercent, memoryUsageMB, memoryTotalMB,
			),
		})
		result.Status = models.HealthStatusCritical
	} else if memoryUsagePercent > 80 {
		issues = append(issues, models.HealthIssue{
			Type:     s.Type(),
			Source:   s.Name(),
			Severity: models.HealthSeverityWarning,
			Message: fmt.Sprintf(
				"High memory usage: %.1f%% (%.1f MB / %.1f MB)",
				memoryUsagePercent, memoryUsageMB, memoryTotalMB,
			),
		})
		if result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
	}

	// Check goroutine count
	goroutines := runtime.NumGoroutine()
	if goroutines > 10000 {
		issues = append(issues, models.HealthIssue{
			Type:     s.Type(),
			Source:   s.Name(),
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("High goroutine count: %d", goroutines),
		})
		if result.Status == models.HealthStatusHealthy {
			result.Status = models.HealthStatusWarning
		}
	}

	result.Issues = issues
	if len(issues) == 0 {
		result.Message = "System resources are healthy"
	} else {
		result.Message = fmt.Sprintf("Found %d system resource issue(s)", len(issues))
	}

	return result
}

// RootFolderHealthChecker checks accessibility of configured root folders
type RootFolderHealthChecker struct {
	db     *database.Database
	logger *logger.Logger
	config *config.Config
}

func (r *RootFolderHealthChecker) Name() string {
	return "Root Folder Accessibility"
}

func (r *RootFolderHealthChecker) Type() models.HealthCheckType {
	return models.HealthCheckTypeRootFolder
}

func (r *RootFolderHealthChecker) IsEnabled() bool {
	return true // Always enabled
}

func (r *RootFolderHealthChecker) GetInterval() time.Duration {
	return 30 * time.Minute
}

func (r *RootFolderHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
	result := models.HealthCheckExecution{
		Type:      r.Type(),
		Source:    r.Name(),
		Status:    models.HealthStatusHealthy,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}

	var issues []models.HealthIssue
	var checkedPaths []string

	// TODO: Get root folders from config service
	// For now, check the data directory
	paths := []string{r.config.Storage.DataDirectory}

	for _, path := range paths {
		checkedPaths = append(checkedPaths, path)

		// Check if path exists and is accessible
		info, err := os.Stat(path)
		if err != nil {
			issues = append(issues, models.HealthIssue{
				Type:     r.Type(),
				Source:   r.Name(),
				Severity: models.HealthSeverityCritical,
				Message:  fmt.Sprintf("Root folder not accessible: %s - %v", path, err),
			})
			result.Status = models.HealthStatusCritical
			continue
		}

		// Check if it's a directory
		if !info.IsDir() {
			issues = append(issues, models.HealthIssue{
				Type:     r.Type(),
				Source:   r.Name(),
				Severity: models.HealthSeverityError,
				Message:  fmt.Sprintf("Root folder path is not a directory: %s", path),
			})
			if result.Status == models.HealthStatusHealthy {
				result.Status = models.HealthStatusError
			}
			continue
		}

		// Check write permissions by creating a temporary file
		testFile := filepath.Join(path, ".radarr-health-check")
		if err := os.WriteFile(testFile, []byte("test"), 0400); err != nil {
			issues = append(issues, models.HealthIssue{
				Type:     r.Type(),
				Source:   r.Name(),
				Severity: models.HealthSeverityError,
				Message:  fmt.Sprintf("Root folder not writable: %s - %v", path, err),
			})
			if result.Status == models.HealthStatusHealthy {
				result.Status = models.HealthStatusError
			}
		} else {
			// Clean up test file
			if err := os.Remove(testFile); err != nil {
				// Not critical if cleanup fails, just log it
				r.logger.Debug("Failed to clean up test file", "file", testFile, "error", err)
			}
		}
	}

	result.Details["checked_paths"] = checkedPaths
	result.Issues = issues

	if len(issues) == 0 {
		result.Message = "All root folders are accessible"
	} else {
		result.Message = fmt.Sprintf("Found %d root folder issue(s)", len(issues))
	}

	return result
}
