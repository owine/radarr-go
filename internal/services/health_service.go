package services

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/radarr/radarr-go/internal/config"
	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// HealthService implements comprehensive health monitoring for the Radarr system
type HealthService struct {
	db     *database.Database
	config *config.Config
	logger *logger.Logger

	// Health check registry
	checkers     map[string]HealthChecker
	checkerTypes map[models.HealthCheckType][]string
	mu           sync.RWMutex

	// Background monitoring
	ctx        context.Context
	cancel     context.CancelFunc
	running    bool
	lastCheck  time.Time
	lastResult *models.HealthCheckResult

	// Dependencies
	systemChecker           SystemResourceCheckerInterface
	serviceChecker          ExternalServiceCheckerInterface
	performanceMonitor      PerformanceMonitorInterface
	notificationIntegration NotificationIntegrationInterface

	// Configuration
	healthConfig models.HealthCheckConfig
}

// NewHealthService creates a new health monitoring service
func NewHealthService(db *database.Database, cfg *config.Config, logger *logger.Logger) *HealthService {
	// Convert config values to models.HealthCheckConfig
	healthConfig := models.DefaultHealthCheckConfig()
	if cfg != nil {
		healthConfig.Enabled = cfg.Health.Enabled

		// Parse interval duration
		if interval, err := time.ParseDuration(cfg.Health.Interval); err == nil {
			healthConfig.Interval = interval
		}

		// Parse database timeout
		if dbTimeout, err := time.ParseDuration(cfg.Health.DatabaseTimeoutThreshold); err == nil {
			healthConfig.DatabaseTimeoutThreshold = dbTimeout
		}

		// Parse external service timeout
		if extTimeout, err := time.ParseDuration(cfg.Health.ExternalServiceTimeout); err == nil {
			healthConfig.ExternalServiceTimeout = extTimeout
		}

		healthConfig.DiskSpaceWarningThreshold = cfg.Health.DiskSpaceWarningThreshold
		healthConfig.DiskSpaceCriticalThreshold = cfg.Health.DiskSpaceCriticalThreshold
		healthConfig.MetricsRetentionDays = cfg.Health.MetricsRetentionDays
		healthConfig.NotifyCriticalIssues = cfg.Health.NotifyCriticalIssues
		healthConfig.NotifyWarningIssues = cfg.Health.NotifyWarningIssues
	}

	hs := &HealthService{
		db:           db,
		config:       cfg,
		logger:       logger,
		checkers:     make(map[string]HealthChecker),
		checkerTypes: make(map[models.HealthCheckType][]string),
		healthConfig: healthConfig,
	}

	// Initialize built-in health checkers
	hs.initializeBuiltinCheckers()

	return hs
}

// initializeBuiltinCheckers registers the built-in health checkers
func (hs *HealthService) initializeBuiltinCheckers() {
	// Database connectivity checker
	hs.RegisterChecker(&DatabaseHealthChecker{
		db:     hs.db,
		logger: hs.logger,
		config: hs.healthConfig,
	})

	// Disk space checker
	hs.RegisterChecker(&DiskSpaceHealthChecker{
		db:           hs.db,
		config:       hs.config,
		logger:       hs.logger,
		healthConfig: hs.healthConfig,
	})

	// System resources checker
	hs.RegisterChecker(&SystemResourcesHealthChecker{
		logger: hs.logger,
		config: hs.healthConfig,
	})

	// Root folder accessibility checker
	hs.RegisterChecker(&RootFolderHealthChecker{
		db:     hs.db,
		logger: hs.logger,
		config: hs.config,
	})
}

// RegisterChecker implements HealthServiceInterface
func (hs *HealthService) RegisterChecker(checker HealthChecker) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	name := checker.Name()
	checkType := checker.Type()

	hs.checkers[name] = checker

	if _, exists := hs.checkerTypes[checkType]; !exists {
		hs.checkerTypes[checkType] = make([]string, 0)
	}
	hs.checkerTypes[checkType] = append(hs.checkerTypes[checkType], name)

	hs.logger.Infow("Health checker registered", "name", name, "type", checkType)
}

// UnregisterChecker implements HealthServiceInterface
func (hs *HealthService) UnregisterChecker(name string) {
	hs.mu.Lock()
	defer hs.mu.Unlock()

	if checker, exists := hs.checkers[name]; exists {
		checkType := checker.Type()
		delete(hs.checkers, name)

		// Remove from type mapping
		if checkers, typeExists := hs.checkerTypes[checkType]; typeExists {
			for i, checkerName := range checkers {
				if checkerName == name {
					hs.checkerTypes[checkType] = append(checkers[:i], checkers[i+1:]...)
					break
				}
			}
		}

		hs.logger.Infow("Health checker unregistered", "name", name, "type", checkType)
	}
}

// RunAllChecks implements HealthServiceInterface
func (hs *HealthService) RunAllChecks(ctx context.Context, types []models.HealthCheckType) models.HealthCheckResult {
	startTime := time.Now()

	hs.mu.RLock()
	defer hs.mu.RUnlock()

	checkersToRun := hs.determineCheckersToRun(types)
	checks, allIssues := hs.executeHealthChecks(ctx, checkersToRun)
	result := hs.buildHealthCheckResult(checks, allIssues, startTime)

	hs.finalizeHealthCheckResult(result, allIssues, startTime)

	return result
}

// determineCheckersToRun determines which health checkers should be executed
func (hs *HealthService) determineCheckersToRun(types []models.HealthCheckType) map[string]HealthChecker {
	checkersToRun := make(map[string]HealthChecker)

	if len(types) == 0 {
		// Run all checkers
		checkersToRun = hs.checkers
	} else {
		// Run only checkers for specified types
		for _, checkType := range types {
			if checkerNames, exists := hs.checkerTypes[checkType]; exists {
				for _, name := range checkerNames {
					if checker, exists := hs.checkers[name]; exists {
						checkersToRun[name] = checker
					}
				}
			}
		}
	}

	return checkersToRun
}

// executeHealthChecks runs health checks concurrently and collects results
func (hs *HealthService) executeHealthChecks(
	ctx context.Context, checkersToRun map[string]HealthChecker,
) ([]models.HealthCheckExecution, []models.HealthIssue) {
	checkResults := make(chan models.HealthCheckExecution, len(checkersToRun))
	var wg sync.WaitGroup

	checkCtx, checkCancel := context.WithTimeout(ctx, 30*time.Second)
	defer checkCancel()

	hs.runCheckersAsync(checkCtx, checkersToRun, checkResults, &wg)

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(checkResults)
	}()

	return hs.collectHealthCheckResults(checkResults)
}

// runCheckersAsync runs health checkers asynchronously
func (hs *HealthService) runCheckersAsync(
	ctx context.Context, checkersToRun map[string]HealthChecker,
	checkResults chan<- models.HealthCheckExecution, wg *sync.WaitGroup,
) {
	for name, checker := range checkersToRun {
		if !checker.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(name string, checker HealthChecker) {
			defer wg.Done()

			checkStart := time.Now()
			result := checker.Check(ctx)
			result.Duration = time.Since(checkStart)
			result.Timestamp = checkStart

			select {
			case checkResults <- result:
			case <-ctx.Done():
				hs.logger.Warnw("Health check result timeout", "checker", name)
			}
		}(name, checker)
	}
}

// collectHealthCheckResults collects and processes health check results
func (hs *HealthService) collectHealthCheckResults(
	checkResults <-chan models.HealthCheckExecution,
) ([]models.HealthCheckExecution, []models.HealthIssue) {
	var checks []models.HealthCheckExecution
	var allIssues []models.HealthIssue

	for result := range checkResults {
		checks = append(checks, result)
		allIssues = append(allIssues, result.Issues...)

		// Log significant issues
		if result.Status == models.HealthStatusError || result.Status == models.HealthStatusCritical {
			hs.logger.Warnw("Health check failed",
				"source", result.Source,
				"type", result.Type,
				"status", result.Status,
				"message", result.Message)
		}
	}

	return checks, allIssues
}

// buildHealthCheckResult creates the final health check result
func (hs *HealthService) buildHealthCheckResult(
	checks []models.HealthCheckExecution, allIssues []models.HealthIssue,
	startTime time.Time,
) models.HealthCheckResult {
	overallStatus := hs.calculateOverallStatus(checks)
	summary := hs.calculateSummary(checks)

	return models.HealthCheckResult{
		OverallStatus: overallStatus,
		Checks:        checks,
		Issues:        allIssues,
		Summary:       summary,
		Timestamp:     startTime,
		Duration:      time.Since(startTime),
	}
}

// finalizeHealthCheckResult handles post-processing of health check results
func (hs *HealthService) finalizeHealthCheckResult(
	result models.HealthCheckResult, allIssues []models.HealthIssue,
	startTime time.Time,
) {
	// Store issues in database
	go hs.storeHealthIssues(allIssues)

	// Cache result
	hs.lastCheck = startTime
	hs.lastResult = &result

	hs.logger.Infow("Health check completed",
		"duration", result.Duration,
		"overall_status", result.OverallStatus,
		"total_checks", len(result.Checks),
		"issues", len(allIssues))
}

// RunCheck implements HealthServiceInterface
func (hs *HealthService) RunCheck(ctx context.Context, name string) (*models.HealthCheckExecution, error) {
	hs.mu.RLock()
	checker, exists := hs.checkers[name]
	hs.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("health checker %s not found", name)
	}

	if !checker.IsEnabled() {
		return nil, fmt.Errorf("health checker %s is disabled", name)
	}

	startTime := time.Now()
	result := checker.Check(ctx)
	result.Duration = time.Since(startTime)
	result.Timestamp = startTime

	return &result, nil
}

// GetHealthStatus implements HealthServiceInterface
func (hs *HealthService) GetHealthStatus(ctx context.Context) models.HealthStatus {
	// Return cached status if recent
	if hs.lastResult != nil && time.Since(hs.lastCheck) < time.Minute {
		return hs.lastResult.OverallStatus
	}

	// Run quick health check
	result := hs.RunAllChecks(ctx, nil)
	return result.OverallStatus
}

// GetHealthIssues implements HealthServiceInterface
func (hs *HealthService) GetHealthIssues(
	filter models.HealthIssueFilter, limit, offset int,
) ([]models.HealthIssue, int64, error) {
	query := hs.db.GORM.Model(&models.HealthIssue{})

	// Apply filters
	if len(filter.Types) > 0 {
		query = query.Where("type IN ?", filter.Types)
	}
	if len(filter.Severities) > 0 {
		query = query.Where("severity IN ?", filter.Severities)
	}
	if len(filter.Sources) > 0 {
		query = query.Where("source IN ?", filter.Sources)
	}
	if filter.Resolved != nil {
		query = query.Where("is_resolved = ?", *filter.Resolved)
	}
	if filter.Dismissed != nil {
		query = query.Where("is_dismissed = ?", *filter.Dismissed)
	}
	if filter.Since != nil {
		query = query.Where("created_at >= ?", *filter.Since)
	}
	if filter.Until != nil {
		query = query.Where("created_at <= ?", *filter.Until)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count health issues: %w", err)
	}

	// Get paginated results
	var issues []models.HealthIssue
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&issues).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to fetch health issues: %w", err)
	}

	return issues, total, nil
}

// GetHealthIssueByID implements HealthServiceInterface
func (hs *HealthService) GetHealthIssueByID(id int) (*models.HealthIssue, error) {
	var issue models.HealthIssue
	if err := hs.db.GORM.First(&issue, id).Error; err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("health issue not found")
		}
		return nil, fmt.Errorf("failed to fetch health issue: %w", err)
	}
	return &issue, nil
}

// DismissHealthIssue implements HealthServiceInterface
func (hs *HealthService) DismissHealthIssue(id int) error {
	result := hs.db.GORM.Model(&models.HealthIssue{}).Where("id = ?", id).Update("is_dismissed", true)
	if result.Error != nil {
		return fmt.Errorf("failed to dismiss health issue: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("health issue not found")
	}

	hs.logger.Infow("Health issue dismissed", "id", id)
	return nil
}

// ResolveHealthIssue implements HealthServiceInterface
func (hs *HealthService) ResolveHealthIssue(id int) error {
	now := time.Now()
	result := hs.db.GORM.Model(&models.HealthIssue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_resolved": true,
		"resolved_at": now,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to resolve health issue: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("health issue not found")
	}

	hs.logger.Infow("Health issue resolved", "id", id)
	return nil
}

// calculateOverallStatus determines the overall health status from individual checks
func (hs *HealthService) calculateOverallStatus(checks []models.HealthCheckExecution) models.HealthStatus {
	if len(checks) == 0 {
		return models.HealthStatusUnknown
	}

	var statuses []models.HealthStatus
	for _, check := range checks {
		statuses = append(statuses, check.Status)
	}

	return models.GetWorstStatus(statuses...)
}

// calculateSummary calculates summary statistics for health checks
func (hs *HealthService) calculateSummary(checks []models.HealthCheckExecution) models.HealthCheckSummary {
	summary := models.HealthCheckSummary{
		Total: len(checks),
	}

	for _, check := range checks {
		switch check.Status {
		case models.HealthStatusHealthy, models.HealthStatusOK:
			summary.Healthy++
		case models.HealthStatusWarning:
			summary.Warning++
		case models.HealthStatusError:
			summary.Error++
		case models.HealthStatusCritical:
			summary.Critical++
		case models.HealthStatusUnknown:
			summary.Unknown++
		default:
			summary.Unknown++
		}
	}

	return summary
}

// storeHealthIssues stores health issues in the database
func (hs *HealthService) storeHealthIssues(issues []models.HealthIssue) {
	for _, issue := range issues {
		// Check for existing similar issue
		var existing models.HealthIssue
		result := hs.db.GORM.Where("type = ? AND source = ? AND message = ? AND is_resolved = false",
			issue.Type, issue.Source, issue.Message).First(&existing)

		if result.Error == nil {
			// Update existing issue
			existing.LastSeen = time.Now()
			existing.Severity = issue.Severity // Update severity if changed
			if err := hs.db.GORM.Save(&existing).Error; err != nil {
				hs.logger.Errorw("Failed to update existing health issue", "error", err)
			}
		} else {
			// Create new issue
			issue.FirstSeen = time.Now()
			issue.LastSeen = issue.FirstSeen
			if err := hs.db.GORM.Create(&issue).Error; err != nil {
				hs.logger.Errorw("Failed to create health issue", "error", err)
			} else {
				hs.logger.Infow("New health issue created",
					"type", issue.Type,
					"source", issue.Source,
					"severity", issue.Severity)

				// Send notification for critical issues
				if hs.notificationIntegration != nil &&
					(issue.Severity == models.HealthSeverityCritical ||
						(issue.Severity == models.HealthSeverityWarning && hs.healthConfig.NotifyWarningIssues)) {
					go func(issue models.HealthIssue) {
						if err := hs.notificationIntegration.NotifyHealthIssue(&issue); err != nil {
							hs.logger.Errorw("Failed to send health issue notification", "error", err)
						}
					}(issue)
				}
			}
		}
	}
}

// GetSystemResources implements HealthServiceInterface
func (hs *HealthService) GetSystemResources(ctx context.Context) (*models.SystemResourceInfo, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Safely convert uint64 to int64 with overflow check
	memoryTotal := hs.safeUint64ToInt64(m.Sys)

	resources := &models.SystemResourceInfo{
		MemoryUsage: float64(m.Alloc) / 1024 / 1024, // MB
		MemoryTotal: memoryTotal / 1024 / 1024,      // MB
	}

	if hs.systemChecker != nil {
		if cpu, err := hs.systemChecker.GetCPUUsage(); err == nil {
			resources.CPUUsage = cpu
		}

		if uptime, err := hs.systemChecker.GetSystemUptime(); err == nil {
			resources.Uptime = uptime
		}

		if load, err := hs.systemChecker.GetLoadAverage(); err == nil {
			resources.LoadAverage = load
		}

		// Get disk usage for data directory
		if used, total, err := hs.systemChecker.GetDiskUsage(hs.config.Storage.DataDirectory); err == nil {
			resources.DiskAvailable = total - used
			resources.DiskTotal = total
			if total > 0 {
				resources.DiskUsage = float64(used) / float64(total) * 100
			}
		}
	}

	return resources, nil
}

// CheckDiskSpace implements HealthServiceInterface
func (hs *HealthService) CheckDiskSpace(ctx context.Context) ([]models.DiskSpaceInfo, error) {
	var diskSpaceInfo []models.DiskSpaceInfo

	if hs.systemChecker == nil {
		return diskSpaceInfo, fmt.Errorf("system checker not available")
	}

	// Check data directory
	if used, total, err := hs.systemChecker.GetDiskUsage(hs.config.Storage.DataDirectory); err == nil {
		free := total - used
		usagePercent := 0.0
		if total > 0 {
			usagePercent = float64(used) / float64(total) * 100
		}

		diskSpaceInfo = append(diskSpaceInfo, models.DiskSpaceInfo{
			Path:         hs.config.Storage.DataDirectory,
			FreeBytes:    free,
			TotalBytes:   total,
			UsedBytes:    used,
			UsagePercent: usagePercent,
			IsAccessible: hs.systemChecker.IsPathAccessible(hs.config.Storage.DataDirectory),
			Warning:      free < hs.healthConfig.DiskSpaceWarningThreshold,
			Critical:     free < hs.healthConfig.DiskSpaceCriticalThreshold,
		})
	}

	// TODO: Check root folders from configuration service

	return diskSpaceInfo, nil
}

// StartMonitoring implements HealthServiceInterface
func (hs *HealthService) StartMonitoring(ctx context.Context) error {
	if hs.running {
		return fmt.Errorf("health monitoring is already running")
	}

	hs.ctx, hs.cancel = context.WithCancel(ctx)
	hs.running = true

	go hs.monitoringLoop()

	hs.logger.Infow("Health monitoring started", "interval", hs.healthConfig.Interval)
	return nil
}

// StopMonitoring implements HealthServiceInterface
func (hs *HealthService) StopMonitoring() error {
	if !hs.running {
		return fmt.Errorf("health monitoring is not running")
	}

	hs.cancel()
	hs.running = false

	hs.logger.Info("Health monitoring stopped")
	return nil
}

// monitoringLoop runs the health monitoring in the background
func (hs *HealthService) monitoringLoop() {
	ticker := time.NewTicker(hs.healthConfig.Interval)
	defer ticker.Stop()

	// Run initial check
	hs.RunAllChecks(hs.ctx, nil)

	for {
		select {
		case <-hs.ctx.Done():
			return
		case <-ticker.C:
			if hs.healthConfig.Enabled {
				hs.RunAllChecks(hs.ctx, nil)
			}
		}
	}
}

// RecordPerformanceMetrics implements HealthServiceInterface
func (hs *HealthService) RecordPerformanceMetrics(ctx context.Context) error {
	if hs.performanceMonitor == nil {
		return fmt.Errorf("performance monitor not available")
	}

	metrics, err := hs.performanceMonitor.CollectMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to collect performance metrics: %w", err)
	}

	return hs.performanceMonitor.RecordMetrics(metrics)
}

// GetPerformanceMetrics implements HealthServiceInterface
func (hs *HealthService) GetPerformanceMetrics(
	since, until *time.Time, limit int,
) ([]models.PerformanceMetrics, error) {
	if hs.performanceMonitor == nil {
		return nil, fmt.Errorf("performance monitor not available")
	}

	return hs.performanceMonitor.GetMetrics(since, until, limit)
}

// CleanupOldMetrics implements HealthServiceInterface
func (hs *HealthService) CleanupOldMetrics() error {
	if hs.performanceMonitor == nil {
		return fmt.Errorf("performance monitor not available")
	}

	cutoff := time.Now().AddDate(0, 0, -hs.healthConfig.MetricsRetentionDays)
	return hs.performanceMonitor.CleanupOldMetrics(cutoff)
}

// GetHealthDashboard implements HealthServiceInterface
func (hs *HealthService) GetHealthDashboard(ctx context.Context) (*models.HealthDashboard, error) {
	// Get recent health check results
	result := hs.RunAllChecks(ctx, nil)

	// Get critical issues
	criticalIssues, _, err := hs.GetHealthIssues(models.HealthIssueFilter{
		Severities: []models.HealthSeverity{models.HealthSeverityCritical},
		Resolved:   boolPtr(false),
	}, 10, 0)
	if err != nil {
		hs.logger.Errorw("Failed to fetch critical issues", "error", err)
	}

	// Get recent issues
	recentIssues, _, err := hs.GetHealthIssues(models.HealthIssueFilter{
		Since: timePtr(time.Now().Add(-24 * time.Hour)),
	}, 20, 0)
	if err != nil {
		hs.logger.Errorw("Failed to fetch recent issues", "error", err)
	}

	// Get system resources
	systemResources, err := hs.GetSystemResources(ctx)
	if err != nil {
		hs.logger.Errorw("Failed to get system resources", "error", err)
		systemResources = &models.SystemResourceInfo{}
	}

	// Get disk space info
	diskSpaceInfo, err := hs.CheckDiskSpace(ctx)
	if err != nil {
		hs.logger.Errorw("Failed to check disk space", "error", err)
	}

	// Get service health (if available)
	var serviceHealth []models.ServiceHealthInfo
	if hs.serviceChecker != nil {
		// These methods would need to be implemented
		// serviceHealth = append(serviceHealth, hs.serviceChecker.CheckDownloadClients(ctx)...)
		// serviceHealth = append(serviceHealth, hs.serviceChecker.CheckIndexers(ctx)...)
	}

	// Get performance trend (last 24 hours)
	since := time.Now().Add(-24 * time.Hour)
	performanceTrend, err := hs.GetPerformanceMetrics(&since, nil, 288) // Every 5 minutes for 24h
	if err != nil {
		hs.logger.Errorw("Failed to get performance trend", "error", err)
	}

	dashboard := &models.HealthDashboard{
		OverallStatus:    result.OverallStatus,
		Summary:          result.Summary,
		CriticalIssues:   criticalIssues,
		RecentIssues:     recentIssues,
		SystemResources:  *systemResources,
		ServiceHealth:    serviceHealth,
		DiskSpaceInfo:    diskSpaceInfo,
		PerformanceTrend: performanceTrend,
		LastUpdated:      time.Now(),
	}

	return dashboard, nil
}

// Helper functions

// safeUint64ToInt64 safely converts uint64 to int64 with overflow protection
func (hs *HealthService) safeUint64ToInt64(value uint64) int64 {
	if value > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(value)
}

func boolPtr(b bool) *bool {
	return &b
}

func timePtr(t time.Time) *time.Time {
	return &t
}
