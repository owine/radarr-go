package services

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/radarr/radarr-go/internal/database"
	"github.com/radarr/radarr-go/internal/logger"
	"github.com/radarr/radarr-go/internal/models"
)

// PerformanceMonitor implements performance metrics collection and monitoring
type PerformanceMonitor struct {
	db     *database.Database
	logger *logger.Logger
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(db *database.Database, logger *logger.Logger) *PerformanceMonitor {
	return &PerformanceMonitor{
		db:     db,
		logger: logger,
	}
}

// CollectMetrics implements PerformanceMonitorInterface
func (pm *PerformanceMonitor) CollectMetrics(ctx context.Context) (*models.PerformanceMetrics, error) {
	startTime := time.Now()

	// Collect memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryUsageMB := float64(m.Alloc) / 1024 / 1024
	memoryTotalMB := float64(m.Sys) / 1024 / 1024

	// Test database latency
	dbLatency, err := pm.measureDatabaseLatency(ctx)
	if err != nil {
		pm.logger.Warnw("Failed to measure database latency", "error", err)
		dbLatency = -1 // Indicate measurement failure
	}

	// TODO: Implement actual disk usage measurement
	diskUsagePercent := 50.0 // Placeholder
	diskAvailableGB := 100.0 // Placeholder
	diskTotalGB := 200.0     // Placeholder

	// TODO: Implement actual CPU usage measurement
	cpuUsagePercent := 25.0 // Placeholder

	// TODO: Implement API latency measurement
	apiLatencyMs := 50.0 // Placeholder

	// TODO: Implement active connections count
	activeConnections := 10 // Placeholder

	// TODO: Implement queue size measurement
	queueSize := 5 // Placeholder

	metrics := &models.PerformanceMetrics{
		CPUUsagePercent:   cpuUsagePercent,
		MemoryUsageMB:     memoryUsageMB,
		MemoryTotalMB:     memoryTotalMB,
		DiskUsagePercent:  diskUsagePercent,
		DiskAvailableGB:   diskAvailableGB,
		DiskTotalGB:       diskTotalGB,
		DatabaseLatencyMs: dbLatency,
		APILatencyMs:      apiLatencyMs,
		ActiveConnections: activeConnections,
		QueueSize:         queueSize,
		Timestamp:         startTime,
	}

	return metrics, nil
}

// measureDatabaseLatency measures the latency of a simple database operation
func (pm *PerformanceMonitor) measureDatabaseLatency(ctx context.Context) (float64, error) {
	start := time.Now()

	// Execute a simple query to measure latency
	var result int
	err := pm.db.GORM.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return -1, fmt.Errorf("database latency test failed: %w", err)
	}

	latency := time.Since(start)
	return float64(latency.Nanoseconds()) / 1e6, nil // Convert to milliseconds
}

// RecordMetrics implements PerformanceMonitorInterface
func (pm *PerformanceMonitor) RecordMetrics(metrics *models.PerformanceMetrics) error {
	if err := pm.db.GORM.Create(metrics).Error; err != nil {
		return fmt.Errorf("failed to record performance metrics: %w", err)
	}

	return nil
}

// GetMetrics implements PerformanceMonitorInterface
func (pm *PerformanceMonitor) GetMetrics(since, until *time.Time, limit int) ([]models.PerformanceMetrics, error) {
	query := pm.db.GORM.Model(&models.PerformanceMetrics{})

	if since != nil {
		query = query.Where("timestamp >= ?", *since)
	}
	if until != nil {
		query = query.Where("timestamp <= ?", *until)
	}

	if limit <= 0 {
		limit = 100 // Default limit
	}

	var metrics []models.PerformanceMetrics
	if err := query.Order("timestamp DESC").Limit(limit).Find(&metrics).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch performance metrics: %w", err)
	}

	return metrics, nil
}

// GetAverageMetrics implements PerformanceMonitorInterface
func (pm *PerformanceMonitor) GetAverageMetrics(since, until time.Time) (*models.PerformanceMetrics, error) {
	var result struct {
		AvgCPUUsage      float64 `gorm:"column:avg_cpu_usage"`
		AvgMemoryUsage   float64 `gorm:"column:avg_memory_usage"`
		AvgMemoryTotal   float64 `gorm:"column:avg_memory_total"`
		AvgDiskUsage     float64 `gorm:"column:avg_disk_usage"`
		AvgDiskAvailable float64 `gorm:"column:avg_disk_available"`
		AvgDiskTotal     float64 `gorm:"column:avg_disk_total"`
		AvgDBLatency     float64 `gorm:"column:avg_db_latency"`
		AvgAPILatency    float64 `gorm:"column:avg_api_latency"`
		AvgConnections   float64 `gorm:"column:avg_connections"`
		AvgQueueSize     float64 `gorm:"column:avg_queue_size"`
	}

	err := pm.db.GORM.Model(&models.PerformanceMetrics{}).
		Select(`
			AVG(cpu_usage_percent) as avg_cpu_usage,
			AVG(memory_usage_mb) as avg_memory_usage,
			AVG(memory_total_mb) as avg_memory_total,
			AVG(disk_usage_percent) as avg_disk_usage,
			AVG(disk_available_gb) as avg_disk_available,
			AVG(disk_total_gb) as avg_disk_total,
			AVG(database_latency_ms) as avg_db_latency,
			AVG(api_latency_ms) as avg_api_latency,
			AVG(active_connections) as avg_connections,
			AVG(queue_size) as avg_queue_size
		`).
		Where("timestamp BETWEEN ? AND ?", since, until).
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed to calculate average metrics: %w", err)
	}

	avgMetrics := &models.PerformanceMetrics{
		CPUUsagePercent:   result.AvgCPUUsage,
		MemoryUsageMB:     result.AvgMemoryUsage,
		MemoryTotalMB:     result.AvgMemoryTotal,
		DiskUsagePercent:  result.AvgDiskUsage,
		DiskAvailableGB:   result.AvgDiskAvailable,
		DiskTotalGB:       result.AvgDiskTotal,
		DatabaseLatencyMs: result.AvgDBLatency,
		APILatencyMs:      result.AvgAPILatency,
		ActiveConnections: int(result.AvgConnections),
		QueueSize:         int(result.AvgQueueSize),
		Timestamp:         since, // Use start time as reference
	}

	return avgMetrics, nil
}

// DetectPerformanceIssues implements PerformanceMonitorInterface
func (pm *PerformanceMonitor) DetectPerformanceIssues(_ context.Context) ([]models.HealthIssue, error) {
	// Get recent metrics (last 15 minutes)
	since := time.Now().Add(-15 * time.Minute)
	metrics, err := pm.GetMetrics(&since, nil, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent metrics for issue detection: %w", err)
	}

	if len(metrics) == 0 {
		return []models.HealthIssue{}, nil
	}

	// Analyze latest metrics
	latest := metrics[0]
	var issues []models.HealthIssue

	pm.checkCPUIssues(latest, &issues)
	pm.checkMemoryIssues(latest, &issues)
	pm.checkDiskIssues(latest, &issues)
	pm.checkDatabaseLatencyIssues(latest, &issues)
	pm.checkAPILatencyIssues(latest, &issues)
	pm.checkQueueSizeIssues(latest, &issues)

	return issues, nil
}

func (pm *PerformanceMonitor) checkCPUIssues(latest models.PerformanceMetrics, issues *[]models.HealthIssue) {
	if latest.CPUUsagePercent > 90 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityCritical,
			Message:  fmt.Sprintf("Critical CPU usage: %.1f%%", latest.CPUUsagePercent),
		})
	} else if latest.CPUUsagePercent > 80 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("High CPU usage: %.1f%%", latest.CPUUsagePercent),
		})
	}
}

func (pm *PerformanceMonitor) checkMemoryIssues(latest models.PerformanceMetrics, issues *[]models.HealthIssue) {
	memoryUsagePercent := (latest.MemoryUsageMB / latest.MemoryTotalMB) * 100
	if memoryUsagePercent > 90 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityCritical,
			Message: fmt.Sprintf("Critical memory usage: %.1f%% (%.1f MB / %.1f MB)",
				memoryUsagePercent, latest.MemoryUsageMB, latest.MemoryTotalMB),
		})
	} else if memoryUsagePercent > 80 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityWarning,
			Message: fmt.Sprintf("High memory usage: %.1f%% (%.1f MB / %.1f MB)",
				memoryUsagePercent, latest.MemoryUsageMB, latest.MemoryTotalMB),
		})
	}
}

func (pm *PerformanceMonitor) checkDiskIssues(latest models.PerformanceMetrics, issues *[]models.HealthIssue) {
	if latest.DiskUsagePercent > 95 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityCritical,
			Message: fmt.Sprintf("Critical disk usage: %.1f%% (%.1f GB available)",
				latest.DiskUsagePercent, latest.DiskAvailableGB),
		})
	} else if latest.DiskUsagePercent > 85 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityWarning,
			Message: fmt.Sprintf("High disk usage: %.1f%% (%.1f GB available)",
				latest.DiskUsagePercent, latest.DiskAvailableGB),
		})
	}
}

func (pm *PerformanceMonitor) checkDatabaseLatencyIssues(
	latest models.PerformanceMetrics, issues *[]models.HealthIssue,
) {
	if latest.DatabaseLatencyMs > 1000 { // 1 second
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityError,
			Message:  fmt.Sprintf("High database latency: %.1f ms", latest.DatabaseLatencyMs),
		})
	} else if latest.DatabaseLatencyMs > 500 { // 500ms
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Elevated database latency: %.1f ms", latest.DatabaseLatencyMs),
		})
	}
}

func (pm *PerformanceMonitor) checkAPILatencyIssues(latest models.PerformanceMetrics, issues *[]models.HealthIssue) {
	if latest.APILatencyMs > 5000 { // 5 seconds
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityError,
			Message:  fmt.Sprintf("High API latency: %.1f ms", latest.APILatencyMs),
		})
	} else if latest.APILatencyMs > 2000 { // 2 seconds
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Elevated API latency: %.1f ms", latest.APILatencyMs),
		})
	}
}

func (pm *PerformanceMonitor) checkQueueSizeIssues(latest models.PerformanceMetrics, issues *[]models.HealthIssue) {
	if latest.QueueSize > 1000 {
		*issues = append(*issues, models.HealthIssue{
			Type:     models.HealthCheckTypePerformance,
			Source:   "Performance Monitor",
			Severity: models.HealthSeverityWarning,
			Message:  fmt.Sprintf("Large queue size: %d items", latest.QueueSize),
		})
	}
}

// CleanupOldMetrics implements PerformanceMonitorInterface
func (pm *PerformanceMonitor) CleanupOldMetrics(olderThan time.Time) error {
	result := pm.db.GORM.Where("timestamp < ?", olderThan).Delete(&models.PerformanceMetrics{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup old performance metrics: %w", result.Error)
	}

	pm.logger.Infow("Cleaned up old performance metrics", "count", result.RowsAffected, "older_than", olderThan)
	return nil
}
