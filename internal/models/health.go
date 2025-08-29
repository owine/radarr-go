package models

import (
	"encoding/json"
	"time"
)

// Extended health status constants (adding to existing ones in notification.go)
const (
	HealthStatusHealthy  HealthStatus = "healthy"  // For consistency, maps to "ok"
	HealthStatusCritical HealthStatus = "critical" // New critical level
	HealthStatusUnknown  HealthStatus = "unknown"  // For uncertain states
)

// HealthSeverity represents the severity level of health issues
type HealthSeverity string

// Health severity level constants
const (
	HealthSeverityInfo     HealthSeverity = "info"     // Informational severity
	HealthSeverityWarning  HealthSeverity = "warning"  // Warning severity
	HealthSeverityError    HealthSeverity = "error"    // Error severity
	HealthSeverityCritical HealthSeverity = "critical" // Critical severity
)

// HealthCheckType represents different types of health checks
type HealthCheckType string

// Health check type constants
const (
	HealthCheckTypeDatabase        HealthCheckType = "database"        // Database check
	HealthCheckTypeDiskSpace       HealthCheckType = "diskSpace"       // Disk space check
	HealthCheckTypeDownloadClient  HealthCheckType = "downloadClient"  // Download client check
	HealthCheckTypeIndexer         HealthCheckType = "indexer"         // Indexer check
	HealthCheckTypeImportList      HealthCheckType = "importList"      // Import list check
	HealthCheckTypeRootFolder      HealthCheckType = "rootFolder"      // Root folder check
	HealthCheckTypeNetwork         HealthCheckType = "network"         // Network check
	HealthCheckTypeSystem          HealthCheckType = "system"          // System check
	HealthCheckTypePerformance     HealthCheckType = "performance"     // Performance check
	HealthCheckTypeConfiguration   HealthCheckType = "configuration"   // Configuration check
	HealthCheckTypeExternalService HealthCheckType = "externalService" // External service check
)

// HealthIssue represents a specific health issue detected in the system
type HealthIssue struct {
	ID          int                    `json:"id" gorm:"primaryKey;autoIncrement"`
	Type        HealthCheckType        `json:"type" gorm:"not null;size:50"`
	Source      string                 `json:"source" gorm:"not null;size:100"` // Service/component name
	Severity    HealthSeverity         `json:"severity" gorm:"not null;size:20"`
	Message     string                 `json:"message" gorm:"not null;type:text"`
	Details     json.RawMessage        `json:"details,omitempty" gorm:"type:json"`
	WikiURL     *string                `json:"wikiUrl,omitempty" gorm:"size:255"`
	FirstSeen   time.Time              `json:"firstSeen" gorm:"not null"`
	LastSeen    time.Time              `json:"lastSeen" gorm:"not null"`
	ResolvedAt  *time.Time             `json:"resolvedAt,omitempty"`
	IsResolved  bool                   `json:"isResolved" gorm:"default:false"`
	IsDismissed bool                   `json:"isDismissed" gorm:"default:false"`
	Data        map[string]interface{} `json:"data,omitempty" gorm:"type:json"`
	CreatedAt   time.Time              `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time              `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (HealthIssue) TableName() string {
	return "health_issues"
}

// HealthCheckExecution represents a single health check execution result
type HealthCheckExecution struct {
	Type      HealthCheckType        `json:"type"`
	Source    string                 `json:"source"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Error     error                  `json:"-"`
	Issues    []HealthIssue          `json:"issues,omitempty"`
}

// HealthCheckResult represents the result of running health checks
type HealthCheckResult struct {
	OverallStatus HealthStatus           `json:"overallStatus"`
	Checks        []HealthCheckExecution `json:"checks"`
	Issues        []HealthIssue          `json:"issues"`
	Summary       HealthCheckSummary     `json:"summary"`
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration"`
}

// HealthCheckSummary provides a summary of health check results
type HealthCheckSummary struct {
	Total    int `json:"total"`
	Healthy  int `json:"healthy"`
	Warning  int `json:"warning"`
	Error    int `json:"error"`
	Critical int `json:"critical"`
	Unknown  int `json:"unknown"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	ID                int       `json:"id" gorm:"primaryKey;autoIncrement"`
	CPUUsagePercent   float64   `json:"cpuUsagePercent" gorm:"not null"`
	MemoryUsageMB     float64   `json:"memoryUsageMB" gorm:"not null"`
	MemoryTotalMB     float64   `json:"memoryTotalMB" gorm:"not null"`
	DiskUsagePercent  float64   `json:"diskUsagePercent" gorm:"not null"`
	DiskAvailableGB   float64   `json:"diskAvailableGB" gorm:"not null"`
	DiskTotalGB       float64   `json:"diskTotalGB" gorm:"not null"`
	DatabaseLatencyMs float64   `json:"databaseLatencyMs" gorm:"not null"`
	APILatencyMs      float64   `json:"apiLatencyMs" gorm:"not null"`
	ActiveConnections int       `json:"activeConnections" gorm:"not null"`
	QueueSize         int       `json:"queueSize" gorm:"not null"`
	Timestamp         time.Time `json:"timestamp" gorm:"not null;index"`
	CreatedAt         time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (PerformanceMetrics) TableName() string {
	return "performance_metrics"
}

// DiskSpaceInfo represents disk space information
type DiskSpaceInfo struct {
	Path         string  `json:"path"`
	FreeBytes    int64   `json:"freeBytes"`
	TotalBytes   int64   `json:"totalBytes"`
	UsedBytes    int64   `json:"usedBytes"`
	UsagePercent float64 `json:"usagePercent"`
	IsAccessible bool    `json:"isAccessible"`
	Warning      bool    `json:"warning"`
	Critical     bool    `json:"critical"`
}

// ServiceHealthInfo represents health information for external services
type ServiceHealthInfo struct {
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	URL          string        `json:"url,omitempty"`
	IsAvailable  bool          `json:"isAvailable"`
	ResponseTime time.Duration `json:"responseTime"`
	LastCheck    time.Time     `json:"lastCheck"`
	ErrorMessage string        `json:"errorMessage,omitempty"`
	Status       HealthStatus  `json:"status"`
}

// SystemResourceInfo represents system resource information
type SystemResourceInfo struct {
	CPUUsage      float64 `json:"cpuUsage"`
	MemoryUsage   float64 `json:"memoryUsage"`
	MemoryTotal   int64   `json:"memoryTotal"`
	DiskUsage     float64 `json:"diskUsage"`
	DiskAvailable int64   `json:"diskAvailable"`
	DiskTotal     int64   `json:"diskTotal"`
	Uptime        int64   `json:"uptime"`
	LoadAverage   float64 `json:"loadAverage,omitempty"`
}

// HealthCheckConfig represents configuration for health checks
type HealthCheckConfig struct {
	Enabled                    bool          `json:"enabled" yaml:"enabled"`
	Interval                   time.Duration `json:"interval" yaml:"interval"`
	DiskSpaceWarningThreshold  int64         `json:"diskSpaceWarningThreshold" yaml:"diskSpaceWarningThreshold"`
	DiskSpaceCriticalThreshold int64         `json:"diskSpaceCriticalThreshold" yaml:"diskSpaceCriticalThreshold"`
	DatabaseTimeoutThreshold   time.Duration `json:"databaseTimeoutThreshold" yaml:"databaseTimeoutThreshold"`
	ExternalServiceTimeout     time.Duration `json:"externalServiceTimeout" yaml:"externalServiceTimeout"`
	MetricsRetentionDays       int           `json:"metricsRetentionDays" yaml:"metricsRetentionDays"`
	NotifyCriticalIssues       bool          `json:"notifyCriticalIssues" yaml:"notifyCriticalIssues"`
	NotifyWarningIssues        bool          `json:"notifyWarningIssues" yaml:"notifyWarningIssues"`
}

// DefaultHealthCheckConfig returns default health check configuration
func DefaultHealthCheckConfig() HealthCheckConfig {
	return HealthCheckConfig{
		Enabled:                    true,
		Interval:                   15 * time.Minute,
		DiskSpaceWarningThreshold:  5 * 1024 * 1024 * 1024, // 5GB
		DiskSpaceCriticalThreshold: 1 * 1024 * 1024 * 1024, // 1GB
		DatabaseTimeoutThreshold:   5 * time.Second,
		ExternalServiceTimeout:     10 * time.Second,
		MetricsRetentionDays:       30,
		NotifyCriticalIssues:       true,
		NotifyWarningIssues:        false,
	}
}

// HealthCheckRequest represents a request for health checks
type HealthCheckRequest struct {
	Types         []HealthCheckType `json:"types,omitempty"`
	ForceRefresh  bool              `json:"forceRefresh"`
	IncludeIssues bool              `json:"includeIssues"`
}

// HealthIssueFilter represents filters for health issues
type HealthIssueFilter struct {
	Types      []HealthCheckType `json:"types,omitempty"`
	Severities []HealthSeverity  `json:"severities,omitempty"`
	Sources    []string          `json:"sources,omitempty"`
	Resolved   *bool             `json:"resolved,omitempty"`
	Dismissed  *bool             `json:"dismissed,omitempty"`
	Since      *time.Time        `json:"since,omitempty"`
	Until      *time.Time        `json:"until,omitempty"`
}

// HealthDashboard represents health dashboard data
type HealthDashboard struct {
	OverallStatus    HealthStatus         `json:"overallStatus"`
	Summary          HealthCheckSummary   `json:"summary"`
	CriticalIssues   []HealthIssue        `json:"criticalIssues"`
	RecentIssues     []HealthIssue        `json:"recentIssues"`
	SystemResources  SystemResourceInfo   `json:"systemResources"`
	ServiceHealth    []ServiceHealthInfo  `json:"serviceHealth"`
	DiskSpaceInfo    []DiskSpaceInfo      `json:"diskSpaceInfo"`
	PerformanceTrend []PerformanceMetrics `json:"performanceTrend"`
	LastUpdated      time.Time            `json:"lastUpdated"`
}

// IsHealthy returns true if the health status indicates a healthy system
func (s HealthStatus) IsHealthy() bool {
	return s == HealthStatusHealthy
}

// IsWarning returns true if the health status indicates a warning
func (s HealthStatus) IsWarning() bool {
	return s == HealthStatusWarning
}

// IsError returns true if the health status indicates an error
func (s HealthStatus) IsError() bool {
	return s == HealthStatusError
}

// IsCritical returns true if the health status indicates a critical issue
func (s HealthStatus) IsCritical() bool {
	return s == HealthStatusCritical
}

// Severity returns the severity level as an integer for comparison
func (s HealthSeverity) Severity() int {
	switch s {
	case HealthSeverityInfo:
		return 1
	case HealthSeverityWarning:
		return 2
	case HealthSeverityError:
		return 3
	case HealthSeverityCritical:
		return 4
	default:
		return 0
	}
}

// ToHealthStatus converts severity to health status
func (s HealthSeverity) ToHealthStatus() HealthStatus {
	switch s {
	case HealthSeverityInfo:
		return HealthStatusHealthy
	case HealthSeverityWarning:
		return HealthStatusWarning
	case HealthSeverityError:
		return HealthStatusError
	case HealthSeverityCritical:
		return HealthStatusCritical
	default:
		return HealthStatusUnknown
	}
}

// GetWorstStatus returns the worst health status from a list of statuses
func GetWorstStatus(statuses ...HealthStatus) HealthStatus {
	worst := HealthStatusHealthy
	for _, status := range statuses {
		switch status {
		case HealthStatusCritical:
			return HealthStatusCritical
		case HealthStatusError:
			if worst != HealthStatusCritical {
				worst = HealthStatusError
			}
		case HealthStatusWarning:
			if worst != HealthStatusCritical && worst != HealthStatusError {
				worst = HealthStatusWarning
			}
		case HealthStatusUnknown:
			if worst == HealthStatusHealthy {
				worst = HealthStatusUnknown
			}
		case HealthStatusHealthy:
			// Healthy status doesn't change the worst status
		case HealthStatusOK:
			// OK status doesn't change the worst status (equivalent to healthy)
		}
	}
	return worst
}
