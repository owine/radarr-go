# Health Monitoring System

Radarr-Go includes a comprehensive health monitoring and diagnostics system that provides proactive issue detection, performance metrics collection, and system reliability monitoring.

## Overview

The health monitoring system continuously monitors various aspects of your Radarr-Go installation:

- **Database connectivity and performance**
- **Disk space availability**
- **System resource usage (CPU, memory)**
- **External service connectivity (indexers, download clients)**
- **Configuration validation**
- **Performance metrics trending**

## Features

### Health Checks

The system runs various health checks at configurable intervals:

1. **Database Health Check**
   - Connection availability
   - Query performance
   - Connection pool status
   - Detects slow queries and connection issues

2. **Disk Space Health Check**
   - Free space monitoring for all configured paths
   - Warning and critical thresholds
   - Accessibility validation

3. **System Resources Health Check**
   - Memory usage monitoring
   - CPU usage tracking
   - Goroutine count monitoring
   - Performance impact detection

4. **Root Folder Health Check**
   - Path accessibility validation
   - Write permission verification
   - Directory structure integrity

### Issue Management

- **Severity Classification**: Issues are classified as Info, Warning, Error, or Critical
- **Automatic Resolution Detection**: Issues are automatically resolved when conditions improve
- **Issue Deduplication**: Prevents spam by updating existing issues instead of creating duplicates
- **User Actions**: Users can dismiss or manually resolve issues
- **Historical Tracking**: Maintains history of issues for trend analysis

### Performance Monitoring

- **Metrics Collection**: Automatically collects system performance metrics
- **Trend Analysis**: Tracks performance over time to detect degradation
- **Configurable Retention**: Customizable data retention periods
- **Performance Alerts**: Automatic alerts when performance thresholds are exceeded

## API Endpoints

### Health Status

```http
GET /api/v3/health
```

Returns overall system health status.

**Query Parameters:**
- `forceRefresh=true` - Force fresh health check execution
- `includeIssues=true` - Include detailed issue information
- `types=database,diskSpace` - Filter by specific check types

**Response:**
```json
{
  "status": "healthy",
  "summary": {
    "total": 4,
    "healthy": 3,
    "warning": 1,
    "error": 0,
    "critical": 0
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "duration": 1250,
  "issues": []
}
```

### Health Dashboard

```http
GET /api/v3/health/dashboard
```

Returns comprehensive health dashboard data including system resources, recent issues, and performance trends.

### Health Issues

```http
GET /api/v3/health/issue
```

List health issues with filtering and pagination.

**Query Parameters:**
- `types=database,diskSpace` - Filter by check types
- `severities=warning,critical` - Filter by severity levels
- `sources=Database Checker` - Filter by source components
- `resolved=false` - Filter by resolution status
- `dismissed=false` - Filter by dismissal status
- `page=1&pageSize=20` - Pagination

```http
GET /api/v3/health/issue/{id}
POST /api/v3/health/issue/{id}/dismiss
POST /api/v3/health/issue/{id}/resolve
```

### System Resources

```http
GET /api/v3/health/system/resources
```

Returns current system resource usage.

```http
GET /api/v3/health/system/diskspace
```

Returns disk space information for all monitored paths.

### Performance Metrics

```http
GET /api/v3/health/metrics
```

Retrieve performance metrics with optional time range filtering.

**Query Parameters:**
- `since=2024-01-01T00:00:00Z` - Start time (ISO 8601)
- `until=2024-01-31T23:59:59Z` - End time (ISO 8601)
- `limit=100` - Maximum number of records

### Monitoring Control

```http
POST /api/v3/health/monitoring/start
POST /api/v3/health/monitoring/stop
POST /api/v3/health/monitoring/cleanup
```

Control background health monitoring and data cleanup.

## Configuration

Health monitoring is configured in the `health` section of your `config.yaml`:

```yaml
health:
  enabled: true
  interval: "15m"
  disk_space_warning_threshold: 5368709120   # 5GB
  disk_space_critical_threshold: 1073741824  # 1GB
  database_timeout_threshold: "5s"
  external_service_timeout: "10s"
  metrics_retention_days: 30
  notify_critical_issues: true
  notify_warning_issues: false
```

### Configuration Options

- `enabled`: Enable/disable health monitoring (default: true)
- `interval`: How often to run health checks (default: "15m")
- `disk_space_warning_threshold`: Free space warning threshold in bytes (default: 5GB)
- `disk_space_critical_threshold`: Free space critical threshold in bytes (default: 1GB)
- `database_timeout_threshold`: Database query timeout threshold (default: "5s")
- `external_service_timeout`: External service check timeout (default: "10s")
- `metrics_retention_days`: How long to keep performance metrics (default: 30)
- `notify_critical_issues`: Send notifications for critical issues (default: true)
- `notify_warning_issues`: Send notifications for warning issues (default: false)

## Task Integration

The health monitoring system integrates with the task scheduling system:

### Scheduled Tasks

- **Health Check Task**: Runs comprehensive health checks
- **Performance Metrics Collection**: Collects and stores performance data
- **Health Maintenance**: Cleanup old data and optimize storage
- **Health Report Generation**: Generate health reports and summaries

### Manual Tasks

Health checks can be triggered manually via:

1. **API Endpoints**: Use the `/api/v3/health` endpoints
2. **Task Scheduler**: Queue health check tasks via `/api/v3/command`
3. **System Commands**: Use the `/api/v3/system/health` endpoint

## Database Schema

The health monitoring system uses two main tables:

### health_issues

Stores detected health issues with metadata:
- Issue type, source, and severity
- Detection and resolution timestamps
- User actions (dismissed/resolved)
- Additional structured data

### performance_metrics

Stores system performance data over time:
- CPU, memory, and disk usage
- Database and API latency
- Connection and queue statistics
- Collection timestamps

## Notification Integration

The system integrates with Radarr-Go's notification system to alert administrators of health issues:

- **Critical Issues**: Automatically send notifications when configured
- **Warning Issues**: Optional notifications for warning-level issues
- **Resolution Notifications**: Alert when critical issues are resolved
- **Configurable**: Control notification behavior via configuration

## Custom Health Checks

The system supports registering custom health checkers:

```go
type CustomHealthChecker struct {
    // Implementation
}

func (c *CustomHealthChecker) Name() string {
    return "Custom Service Check"
}

func (c *CustomHealthChecker) Type() models.HealthCheckType {
    return models.HealthCheckTypeExternalService
}

func (c *CustomHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
    // Perform health check
    return models.HealthCheckExecution{
        Type:      c.Type(),
        Source:    c.Name(),
        Status:    models.HealthStatusHealthy,
        Message:   "Service is healthy",
        Timestamp: time.Now(),
    }
}

// Register the checker
healthService.RegisterChecker(&CustomHealthChecker{})
```

## Best Practices

### Configuration

1. **Set appropriate thresholds** based on your system capacity
2. **Configure retention periods** to balance storage vs. historical data needs
3. **Enable notifications** for critical issues to ensure prompt response
4. **Adjust check intervals** based on your system's stability and performance needs

### Monitoring

1. **Review the health dashboard** regularly to identify trends
2. **Address warning issues** before they become critical
3. **Monitor performance metrics** to identify degradation over time
4. **Keep the health system itself healthy** by monitoring its resource usage

### Troubleshooting

1. **Check logs** for health check execution details
2. **Review issue history** to identify recurring problems
3. **Use performance metrics** to correlate issues with system changes
4. **Test external services** manually if health checks report problems

## Performance Impact

The health monitoring system is designed to have minimal performance impact:

- **Lightweight checks** that complete quickly
- **Configurable intervals** to control frequency
- **Background execution** doesn't block application operations
- **Efficient storage** with automatic cleanup of old data
- **Smart scheduling** to distribute load across time

## Security Considerations

- Health endpoints require API key authentication
- Issue data may contain sensitive system information
- Performance metrics could reveal system architecture details
- Access should be restricted to authorized administrators

## Migration and Upgrades

The health monitoring system includes database migrations:

- **Automatic schema updates** during application startup
- **Backward compatibility** with existing installations
- **Data preservation** during upgrades
- **Rollback support** for migration failures

For more detailed information about specific components, see the source code documentation in the `internal/services` package.
