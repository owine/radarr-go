package services

import (
	"context"
	"time"

	"github.com/radarr/radarr-go/internal/models"
)

// HealthChecker interface defines a health check component
type HealthChecker interface {
	// Name returns the name of the health check
	Name() string

	// Type returns the type of health check
	Type() models.HealthCheckType

	// Check performs the health check and returns the result
	Check(ctx context.Context) models.HealthCheckExecution

	// IsEnabled returns true if this health check is enabled
	IsEnabled() bool

	// GetInterval returns how often this check should run
	GetInterval() time.Duration
}

// HealthServiceInterface defines the interface for the health monitoring service
type HealthServiceInterface interface {
	// RegisterChecker registers a new health checker
	RegisterChecker(checker HealthChecker)

	// UnregisterChecker removes a health checker
	UnregisterChecker(name string)

	// RunAllChecks runs all registered health checks
	RunAllChecks(ctx context.Context, types []string) models.HealthCheckResultV2

	// RunCheck runs a specific health check by name
	RunCheck(ctx context.Context, name string) (*models.HealthCheckExecution, error)

	// GetHealthStatus returns the current overall health status
	GetHealthStatus(ctx context.Context) models.HealthStatus

	// GetHealthIssues returns current health issues with optional filtering
	GetHealthIssues(filter models.HealthIssueFilterV2, limit, offset int) ([]models.HealthIssueV2, int64, error)

	// GetHealthIssueByID returns a specific health issue
	GetHealthIssueByID(id int) (*models.HealthIssueV2, error)

	// DismissHealthIssue dismisses a health issue
	DismissHealthIssue(id int) error

	// ResolveHealthIssue marks a health issue as resolved
	ResolveHealthIssue(id int) error

	// GetHealthDashboard returns comprehensive health dashboard data
	GetHealthDashboard(ctx context.Context) (*models.HealthDashboard, error)

	// RecordPerformanceMetrics records system performance metrics
	RecordPerformanceMetrics(ctx context.Context) error

	// GetPerformanceMetrics returns performance metrics with optional time range
	GetPerformanceMetrics(since, until *time.Time, limit int) ([]models.PerformanceMetrics, error)

	// CleanupOldMetrics removes old performance metrics based on retention policy
	CleanupOldMetrics() error

	// GetSystemResources returns current system resource usage
	GetSystemResources(ctx context.Context) (*models.SystemResourceInfo, error)

	// CheckDiskSpace checks disk space for all configured paths
	CheckDiskSpace(ctx context.Context) ([]models.DiskSpaceInfo, error)

	// StartMonitoring starts the health monitoring background process
	StartMonitoring(ctx context.Context) error

	// StopMonitoring stops the health monitoring background process
	StopMonitoring() error
}

// HealthIssueServiceInterface defines the interface for health issue management
type HealthIssueServiceInterface interface {
	// CreateIssue creates a new health issue
	CreateIssue(issue *models.HealthIssueV2) error

	// UpdateIssue updates an existing health issue
	UpdateIssue(issue *models.HealthIssueV2) error

	// GetIssues returns health issues with filtering
	GetIssues(filter models.HealthIssueFilterV2, limit, offset int) ([]models.HealthIssueV2, int64, error)

	// GetIssueByID returns a specific health issue
	GetIssueByID(id int) (*models.HealthIssueV2, error)

	// ResolveIssue marks an issue as resolved
	ResolveIssue(id int) error

	// DismissIssue dismisses an issue
	DismissIssue(id int) error

	// CheckForDuplicates checks if a similar issue already exists
	CheckForDuplicates(issue *models.HealthIssueV2) (*models.HealthIssueV2, error)

	// CleanupResolvedIssues removes old resolved issues
	CleanupResolvedIssues(olderThan time.Time) error

	// GetIssueHistory returns the history of issues for trending analysis
	GetIssueHistory(since time.Time, groupBy string) (map[string]int, error)
}

// PerformanceMonitorInterface defines the interface for performance monitoring
type PerformanceMonitorInterface interface {
	// CollectMetrics collects current system performance metrics
	CollectMetrics(ctx context.Context) (*models.PerformanceMetrics, error)

	// RecordMetrics stores performance metrics in the database
	RecordMetrics(metrics *models.PerformanceMetrics) error

	// GetMetrics retrieves performance metrics with optional filtering
	GetMetrics(since, until *time.Time, limit int) ([]models.PerformanceMetrics, error)

	// GetAverageMetrics returns average metrics over a time period
	GetAverageMetrics(since, until time.Time) (*models.PerformanceMetrics, error)

	// DetectPerformanceIssues analyzes metrics to detect performance issues
	DetectPerformanceIssues(ctx context.Context) ([]models.HealthIssueV2, error)

	// CleanupOldMetrics removes old metrics based on retention policy
	CleanupOldMetrics(olderThan time.Time) error
}

// SystemResourceCheckerInterface defines the interface for system resource checking
type SystemResourceCheckerInterface interface {
	// GetCPUUsage returns current CPU usage percentage
	GetCPUUsage() (float64, error)

	// GetMemoryUsage returns current memory usage information
	GetMemoryUsage() (used, total int64, err error)

	// GetDiskUsage returns disk usage information for a path
	GetDiskUsage(path string) (used, total int64, err error)

	// GetSystemUptime returns system uptime in seconds
	GetSystemUptime() (int64, error)

	// GetLoadAverage returns system load average
	GetLoadAverage() (float64, error)

	// IsPathAccessible checks if a path is accessible
	IsPathAccessible(path string) bool
}

// ExternalServiceCheckerInterface defines the interface for checking external services
type ExternalServiceCheckerInterface interface {
	// CheckService checks the health of an external service
	CheckService(ctx context.Context, serviceType, name, url string) models.ServiceHealthInfo

	// CheckDownloadClients checks all configured download clients
	CheckDownloadClients(ctx context.Context) ([]models.ServiceHealthInfo, error)

	// CheckIndexers checks all configured indexers
	CheckIndexers(ctx context.Context) ([]models.ServiceHealthInfo, error)

	// CheckImportLists checks all configured import lists
	CheckImportLists(ctx context.Context) ([]models.ServiceHealthInfo, error)

	// CheckNotifications checks all configured notification services
	CheckNotifications(ctx context.Context) ([]models.ServiceHealthInfo, error)
}

// NotificationIntegrationInterface defines the interface for health notifications
type NotificationIntegrationInterface interface {
	// NotifyHealthIssue sends a notification for a health issue
	NotifyHealthIssue(issue *models.HealthIssue) error

	// NotifyHealthRecovery sends a notification when an issue is resolved
	NotifyHealthRecovery(issue *models.HealthIssue) error

	// NotifyPerformanceAlert sends a notification for performance issues
	NotifyPerformanceAlert(metrics *models.PerformanceMetrics, message string) error

	// ShouldNotify determines if a notification should be sent for an issue
	ShouldNotify(issue *models.HealthIssue) bool
}
