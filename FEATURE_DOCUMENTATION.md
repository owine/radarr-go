# Radarr-Go Advanced Features Documentation

This comprehensive guide covers Radarr-Go's advanced features that differentiate it from the original Radarr, including sophisticated task scheduling, multi-provider notifications, intelligent health monitoring, and automated file organization systems.

## Table of Contents

1. [Task Scheduling System](#task-scheduling-system)
2. [Notification System](#notification-system)
3. [File Organization System](#file-organization-system)
4. [Health Monitoring](#health-monitoring)
5. [Performance Optimization](#performance-optimization)
6. [Troubleshooting](#troubleshooting)

---

## Task Scheduling System

Radarr-Go's task scheduling system provides a sophisticated, multi-threaded approach to background job management with real-time monitoring and priority-based execution.

### Architecture Overview

The task system operates with three distinct worker pools:

- **High-Priority Pool**: 2 concurrent workers for urgent tasks
- **Default Pool**: 3 concurrent workers for standard operations
- **Background Pool**: 1 worker for low-priority maintenance tasks

### Task Types and Commands

#### Core Movie Management Tasks

- `RefreshMovieCommand`: Updates movie metadata from TMDB
- `MovieSearchCommand`: Searches indexers for specific movies
- `RssSearchCommand`: Performs RSS sync across all enabled indexers
- `ImportListSyncCommand`: Synchronizes import lists with configured providers

#### System Maintenance Tasks

- `HealthCheckCommand`: Runs comprehensive system health checks
- `CleanupCommand`: Removes old logs, metrics, and temporary files
- `BackupCommand`: Creates database and configuration backups
- `UpdateCheckCommand`: Checks for application updates

#### File Management Tasks

- `OrganizeFilesCommand`: Organizes downloaded files according to naming rules
- `RenameMovieCommand`: Renames existing movie files
- `RescanMovieCommand`: Rescans movie directories for changes

### Configuration

Tasks are configured through the API or configuration files:

```yaml
tasks:
  default_priority: "normal"
  max_concurrent_tasks: 3
  timeout: 3600  # seconds
  retry_count: 3
  retry_delay: 300  # seconds
```

### Creating Scheduled Tasks

#### Via API

```http
POST /api/v3/system/task/scheduled
Content-Type: application/json

{
  "name": "Weekly Library Refresh",
  "commandName": "RefreshMovieCommand",
  "body": {
    "movieIds": [],
    "isNewMovie": false
  },
  "intervalMs": 604800000,
  "priority": "low",
  "enabled": true
}
```

#### Via Configuration

```yaml
scheduled_tasks:
  - name: "Nightly Cleanup"
    command: "CleanupCommand"
    interval: "24h"
    priority: "low"
    enabled: true
  - name: "RSS Sync"
    command: "RssSearchCommand"
    interval: "15m"
    priority: "normal"
    enabled: true
```

### Task Management Operations

#### Manual Task Execution

```bash
# Queue a movie refresh task
curl -X POST "http://localhost:7878/api/v3/system/task" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Manual Movie Refresh",
    "commandName": "RefreshMovieCommand",
    "body": {"movieId": 123},
    "priority": "high"
  }'
```

#### Task Monitoring

```bash
# View all tasks
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/task"

# View specific task
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/task/123"

# View queue status
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/task/queue"
```

#### Task Cancellation

```bash
# Cancel a running task
curl -X DELETE -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/task/123"
```

### Advanced Task Features

#### Custom Task Handlers

Create custom task handlers by implementing the `TaskHandler` interface:

```go
type CustomTaskHandler struct {
    logger *logger.Logger
}

func (h *CustomTaskHandler) Execute(ctx context.Context, task *models.TaskV2, updateProgress func(int, string)) error {
    updateProgress(0, "Starting custom operation")

    // Perform custom logic here

    updateProgress(100, "Custom operation completed")
    return nil
}

func (h *CustomTaskHandler) GetName() string {
    return "CustomCommand"
}

func (h *CustomTaskHandler) GetDescription() string {
    return "Executes custom business logic"
}
```

#### Task Dependencies

Tasks can be configured with dependencies to ensure proper execution order:

```json
{
  "name": "Post-Download Processing",
  "commandName": "OrganizeFilesCommand",
  "dependencies": ["DownloadCompletedCommand"],
  "priority": "normal"
}
```

### Performance Optimization

#### Worker Pool Tuning

Adjust worker pools based on system resources:

```yaml
task_pools:
  high_priority:
    max_workers: 4
    queue_size: 50
  default:
    max_workers: 6
    queue_size: 100
  background:
    max_workers: 2
    queue_size: 25
```

#### Task Timeout Configuration

Configure appropriate timeouts for different task types:

```yaml
task_timeouts:
  RefreshMovieCommand: 300    # 5 minutes
  MovieSearchCommand: 1800    # 30 minutes
  OrganizeFilesCommand: 600   # 10 minutes
  HealthCheckCommand: 120     # 2 minutes
```

---

## Notification System

Radarr-Go features an extensive notification system supporting 21+ providers with advanced templating, retry logic, and event filtering.

### Supported Providers

#### Chat & Messaging

- **Discord**: Rich embed notifications with custom colors and thumbnails
- **Slack**: Channel-based notifications with threading support
- **Telegram**: Bot-based messaging with markdown formatting
- **Matrix**: Decentralized messaging with E2E encryption support
- **Signal**: Privacy-focused messaging via Signal-CLI REST API

#### Email Services

- **SMTP**: Direct SMTP server integration
- **Mailgun**: Transactional email service
- **SendGrid**: Cloud email delivery platform

#### Push Notifications

- **Pushover**: Mobile and desktop push notifications
- **Pushbullet**: Cross-device notification synchronization
- **Gotify**: Self-hosted push notification server
- **Join**: Android device integration via Joaoapps
- **Ntfy**: Simple HTTP-based push notifications

#### Media Server Integration

- **Plex**: Media server library updates and webhooks
- **Emby**: Media server notifications
- **Jellyfin**: Open-source media server integration
- **Kodi**: Direct JSON-RPC notifications

#### Advanced Integration

- **Webhook**: Custom HTTP POST notifications
- **Custom Script**: Execute custom scripts with notification data
- **Apprise**: Universal notification library integration
- **Notifiarr**: Community-focused notification service

### Provider Configuration

#### Discord Setup

```json
{
  "name": "Discord Movies Channel",
  "implementation": "discord",
  "enabled": true,
  "settings": {
    "webHookUrl": "https://discord.com/api/webhooks/...",
    "username": "Radarr",
    "avatar": "https://your-server.com/radarr-icon.png",
    "onGrab": true,
    "onDownload": true,
    "onUpgrade": true,
    "onHealthIssue": true,
    "includeHealthWarnings": false
  }
}
```

#### Email/SMTP Setup

```json
{
  "name": "System Administrator Email",
  "implementation": "email",
  "enabled": true,
  "settings": {
    "server": "smtp.gmail.com",
    "port": 587,
    "ssl": false,
    "requireAuthentication": true,
    "username": "your-email@gmail.com",
    "password": "app-specific-password",
    "from": "radarr@yourdomain.com",
    "to": ["admin@yourdomain.com", "team@yourdomain.com"],
    "subject": "Radarr - {EventType}",
    "onGrab": false,
    "onDownload": true,
    "onUpgrade": true,
    "onHealthIssue": true
  }
}
```

#### Webhook Setup

```json
{
  "name": "Custom Webhook Integration",
  "implementation": "webhook",
  "enabled": true,
  "settings": {
    "url": "https://your-api.com/radarr-webhook",
    "method": "POST",
    "username": "",
    "password": "",
    "headers": {
      "Authorization": "Bearer YOUR_TOKEN",
      "Content-Type": "application/json"
    },
    "customTemplate": "{\"event\": \"{EventType}\", \"movie\": \"{MovieTitle}\", \"quality\": \"{QualityName}\"}",
    "onGrab": true,
    "onDownload": true,
    "onUpgrade": true,
    "onHealthIssue": true
  }
}
```

#### Pushover Setup

```json
{
  "name": "Mobile Push Notifications",
  "implementation": "pushover",
  "enabled": true,
  "settings": {
    "userKey": "your-user-key",
    "apiToken": "your-app-token",
    "devices": ["phone", "tablet"],
    "priority": 0,
    "sound": "pushover",
    "retry": 60,
    "expire": 3600,
    "onGrab": true,
    "onDownload": true,
    "onHealthIssue": true
  }
}
```

### Event Types and Triggers

#### Available Events

- `grab`: Movie grabbed from indexer
- `download`: Movie downloaded and imported
- `upgrade`: Movie upgraded to better quality
- `rename`: Movie file renamed
- `movieAdded`: New movie added to library
- `movieDelete`: Movie removed from library
- `movieFileDelete`: Movie file deleted
- `health`: System health issue detected
- `applicationUpdate`: New Radarr version available
- `manualInteractionRequired`: Manual intervention needed

#### Event Configuration

Configure which events trigger notifications for each provider:

```json
{
  "onGrab": true,
  "onDownload": true,
  "onUpgrade": true,
  "onRename": false,
  "onMovieAdded": true,
  "onMovieDelete": false,
  "onMovieFileDelete": false,
  "onHealthIssue": true,
  "onApplicationUpdate": true,
  "onManualInteractionRequired": true
}
```

### Advanced Template System

#### Template Variables

Available variables for custom templates:

- `{MovieTitle}`: Movie title
- `{MovieYear}`: Release year
- `{QualityName}`: Quality profile name
- `{SourceTitle}`: Original release name
- `{DownloadClient}`: Download client name
- `{EventType}`: Event type identifier
- `{HealthIssueType}`: Health issue category
- `{HealthMessage}`: Health issue description
- `{ServerName}`: Radarr server name
- `{Timestamp}`: Event timestamp

#### Custom Discord Template

```json
{
  "customTemplate": {
    "embeds": [{
      "title": "{MovieTitle} ({MovieYear})",
      "description": "**Quality:** {QualityName}\n**Source:** {SourceTitle}",
      "color": 3066993,
      "thumbnail": {
        "url": "{MoviePosterUrl}"
      },
      "fields": [{
        "name": "Download Client",
        "value": "{DownloadClient}",
        "inline": true
      }, {
        "name": "Size",
        "value": "{FileSize}",
        "inline": true
      }],
      "timestamp": "{Timestamp}"
    }]
  }
}
```

#### Custom Email Template

```html
<!DOCTYPE html>
<html>
<head>
    <title>Radarr Notification</title>
</head>
<body>
    <h2>Movie {EventType}: {MovieTitle} ({MovieYear})</h2>
    <p><strong>Quality:</strong> {QualityName}</p>
    <p><strong>Source:</strong> {SourceTitle}</p>
    <p><strong>Download Client:</strong> {DownloadClient}</p>
    <p><strong>Time:</strong> {Timestamp}</p>
</body>
</html>
```

### Notification Testing

#### Test Notification via API

```bash
curl -X POST "http://localhost:7878/api/v3/notification/test" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Notification",
    "implementation": "discord",
    "settings": {
      "webHookUrl": "https://discord.com/api/webhooks/...",
      "onGrab": true
    }
  }'
```

### Retry Configuration

#### Automatic Retry Logic

```json
{
  "retryConfig": {
    "maxRetries": 3,
    "initialDelay": "30s",
    "maxDelay": "5m",
    "backoffFactor": 2.0,
    "retryCondition": "network_error"
  }
}
```

### Notification History

#### View Notification History

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/notification/history?limit=50&offset=0"
```

---

## File Organization System

Radarr-Go's file organization system provides automated file processing with multiple operation modes, conflict resolution, and comprehensive naming templates.

### File Operations

#### Operation Types

1. **Move**: Relocates files from source to destination (default)
2. **Copy**: Creates a copy while preserving original
3. **Hard Link**: Creates hard links to save disk space
4. **Symbolic Link**: Creates symbolic links for advanced setups

### Naming Configuration

#### Template Variables

Basic movie information:

- `{MovieTitle}`: Movie title
- `{MovieTitleClean}`: Movie title with special characters removed
- `{MovieYear}`: Release year
- `{ImdbId}`: IMDb identifier
- `{TmdbId}`: TMDB identifier

Quality and technical details:

- `{Quality}`: Quality name (e.g., "Bluray-1080p")
- `{QualityProper}`: Quality with proper/repack indicators
- `{Resolution}`: Video resolution (e.g., "1080p")
- `{Source}`: Media source (e.g., "Bluray", "Web")
- `{Codec}`: Video codec information

Advanced metadata:

- `{MediaInfo}`: Comprehensive media information
- `{Edition}`: Movie edition (Director's Cut, Extended, etc.)
- `{ReleaseGroup}`: Release group name
- `{OriginalTitle}`: Original movie title
- `{Collection}`: Movie collection name

#### Naming Examples

##### Basic Template

```
{MovieTitle} ({MovieYear}) [{Quality}]
```

Result: `Avatar (2009) [Bluray-1080p].mkv`

##### Advanced Template with Media Info

```
{MovieTitle} ({MovieYear}) [{Quality} {MediaInfo.VideoCodec} {MediaInfo.AudioCodec}]-{ReleaseGroup}
```

Result: `Avatar (2009) [Bluray-1080p x264 DTS]-SPARKS.mkv`

##### Collection-Based Organization

```
{Collection}/{MovieTitle} ({MovieYear}) [{Quality}]
```

Result: `Avatar Collection/Avatar (2009) [Bluray-1080p].mkv`

### Organization Strategies

#### Folder Structure Templates

##### Year-Based Organization

```yaml
naming:
  movie_folder_format: "{MovieTitle} ({MovieYear})"
  movie_file_format: "{MovieTitle} ({MovieYear}) [{Quality}]"
```

##### Collection-Based Organization

```yaml
naming:
  movie_folder_format: "{Collection}/{MovieTitle} ({MovieYear})"
  movie_file_format: "{MovieTitle} ({MovieYear}) [{Quality}]"
```

##### Genre-Based Organization

```yaml
naming:
  movie_folder_format: "{Genre}/{MovieTitle} ({MovieYear})"
  movie_file_format: "{MovieTitle} ({MovieYear}) [{Quality}]"
```

### Automatic File Processing

#### Configuration

```yaml
file_management:
  auto_organize: true
  operation: "move"  # move, copy, hardlink, symlink
  create_empty_folders: false
  delete_empty_folders: true
  skip_free_space_check: false
  minimum_free_space: 100  # MB
  chmod_folder: "755"
  chown_user: ""
  chown_group: ""
  enable_media_info: true
  import_extra_files: true
  extra_file_extensions: ["srt", "sub", "nfo", "jpg"]
```

#### Import Decision Workflow

1. **File Detection**: Scan for importable video files
2. **Quality Parsing**: Extract quality information from filename
3. **Movie Matching**: Match against existing movies in library
4. **Duplicate Handling**: Check for existing files and handle conflicts
5. **Organization**: Apply naming rules and move/copy/link files
6. **Verification**: Validate file integrity and update database

### Conflict Resolution

#### Handling Existing Files

```yaml
file_management:
  file_conflict_action: "rename"  # replace, rename, skip
  rename_pattern: "{OriginalName} - {Quality}"
  backup_existing: true
  backup_location: "/backups/radarr"
```

#### Quality Upgrades

```yaml
quality_management:
  auto_upgrade: true
  upgrade_conditions:
    - better_quality: true
    - proper_repack: true
    - size_difference_threshold: 10  # percent
  delete_replaced_files: true
  recycle_bin: "/recycle/radarr"
```

### Manual Import Processing

#### Scan Directory for Imports

```bash
curl -X POST "http://localhost:7878/api/v3/manual-import/scan" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "path": "/downloads/movies",
    "recursive": true,
    "replaceExistingFiles": false
  }'
```

#### Process Manual Import

```bash
curl -X POST "http://localhost:7878/api/v3/manual-import" \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[{
    "path": "/downloads/movies/Avatar.2009.1080p.BluRay.x264-SPARKS.mkv",
    "movieId": 123,
    "qualityId": 6,
    "replaceExistingFiles": false
  }]'
```

### File Organization Monitoring

#### View Organization History

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/file-organization?limit=50&offset=0"
```

#### Retry Failed Organizations

```bash
curl -X POST "http://localhost:7878/api/v3/file-organization/retry" \
  -H "X-API-Key: YOUR_API_KEY"
```

### Advanced Features

#### Extra File Management

Automatically handle associated files (subtitles, NFO files, artwork):

```yaml
extra_files:
  enabled: true
  extensions: ["srt", "sub", "idx", "nfo", "jpg", "png"]
  naming_patterns:
    subtitles: "{MovieTitle}.{Language}.{Extension}"
    nfo: "{MovieTitle}.nfo"
    artwork: "poster.jpg"
```

#### Sample File Detection

Automatically skip sample files:

```yaml
sample_detection:
  enabled: true
  minimum_size: 50  # MB
  keywords: ["sample", "trailer", "preview"]
  max_duration: 300  # seconds
```

---

## Health Monitoring

Radarr-Go provides comprehensive health monitoring with real-time issue detection, automated resolution suggestions, and integration with the notification system.

### Health Check Types

#### System Health Checks

- **Database Connectivity**: Monitors database response times and connection status
- **Disk Space**: Tracks available space on root folders and data directories
- **System Resources**: Monitors CPU, memory, and system load
- **File Permissions**: Verifies read/write access to critical directories
- **Network Connectivity**: Tests internet connectivity and DNS resolution

#### Application Health Checks

- **Indexer Status**: Monitors indexer availability and response times
- **Download Client Status**: Checks download client connectivity and queue status
- **Import List Status**: Validates import list functionality
- **Notification Status**: Tests notification provider connectivity
- **Task Queue Status**: Monitors background task execution

#### Configuration Health Checks

- **Root Folder Accessibility**: Verifies all root folders are accessible
- **Quality Profile Validation**: Ensures quality profiles are properly configured
- **Naming Configuration**: Validates naming templates and token usage
- **API Key Security**: Checks API key configuration and security

### Health Check Configuration

#### Basic Configuration

```yaml
health:
  enabled: true
  interval: "5m"
  database_timeout_threshold: "30s"
  external_service_timeout: "10s"
  disk_space_warning_threshold: 1073741824  # 1GB in bytes
  disk_space_critical_threshold: 536870912   # 512MB in bytes
  metrics_retention_days: 30
  notify_critical_issues: true
  notify_warning_issues: false
```

#### Advanced Health Monitoring

```yaml
health_monitoring:
  continuous_monitoring: true
  check_intervals:
    database: "1m"
    disk_space: "5m"
    system_resources: "2m"
    external_services: "10m"
  thresholds:
    cpu_usage_warning: 80      # percent
    cpu_usage_critical: 95     # percent
    memory_usage_warning: 85   # percent
    memory_usage_critical: 95  # percent
    response_time_warning: 5   # seconds
    response_time_critical: 15 # seconds
```

### Health Issue Types

#### Critical Issues (Red)

- Database connectivity lost
- Disk space critically low (< 512MB)
- All indexers failing
- Download clients unreachable
- System resources exhausted

#### Warning Issues (Yellow)

- Disk space low (< 1GB)
- Some indexers failing
- High system resource usage
- Network connectivity issues
- Configuration problems

#### Information Issues (Blue)

- System updates available
- Configuration suggestions
- Performance recommendations
- Maintenance reminders

### Health Dashboard

#### System Overview

The health dashboard provides:

- Overall system status indicator
- Critical issue count and details
- System resource utilization graphs
- Service availability status
- Performance metrics trends
- Recent health events timeline

#### API Endpoints

##### Get Health Status

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health"
```

##### Get Detailed Health Dashboard

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/dashboard"
```

##### Run Manual Health Check

```bash
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/check"
```

##### Get System Resources

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/resources"
```

### Health Issue Management

#### Issue Resolution

##### Dismiss Issues

```bash
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/issues/123/dismiss"
```

##### Mark Issues as Resolved

```bash
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/issues/123/resolve"
```

#### Automated Issue Resolution

The system can automatically resolve certain types of issues:

```yaml
auto_resolution:
  enabled: true
  actions:
    cleanup_temp_files: true
    restart_failed_services: true
    clear_stuck_tasks: true
    rotate_logs: true
  max_attempts: 3
  cooldown_period: "30m"
```

### Performance Metrics

#### Metric Collection

```yaml
performance_monitoring:
  enabled: true
  collection_interval: "30s"
  metrics:
    - cpu_usage
    - memory_usage
    - disk_io
    - network_io
    - database_performance
    - api_response_times
  retention_period: "30d"
```

#### Available Metrics

- **System Metrics**: CPU usage, memory consumption, disk I/O, network traffic
- **Application Metrics**: API response times, task execution times, error rates
- **Database Metrics**: Query execution times, connection pool status, transaction rates
- **Service Metrics**: Indexer response times, download client performance

#### Performance Trend Analysis

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/metrics?since=2024-01-01T00:00:00Z&until=2024-01-31T23:59:59Z&limit=1000"
```

### Health Notifications

#### Integration with Notification System

Health issues automatically trigger notifications based on severity:

```json
{
  "onHealthIssue": true,
  "includeHealthWarnings": false,
  "healthNotificationSettings": {
    "criticalOnly": false,
    "aggregateIssues": true,
    "throttleMinutes": 60,
    "customTemplate": {
      "subject": "Radarr Health Alert - {HealthIssueType}",
      "body": "A {HealthIssueSeverity} health issue has been detected:\n\n{HealthIssueMessage}\n\nTime: {Timestamp}"
    }
  }
}
```

### Custom Health Checks

#### Creating Custom Health Checkers

```go
type CustomHealthChecker struct {
    logger *logger.Logger
    config CustomConfig
}

func (c *CustomHealthChecker) Name() string {
    return "custom-service-check"
}

func (c *CustomHealthChecker) Type() models.HealthCheckType {
    return models.HealthCheckTypeExternalService
}

func (c *CustomHealthChecker) Check(ctx context.Context) models.HealthCheckExecution {
    // Implement custom health check logic
    return models.HealthCheckExecution{
        Source:    "CustomService",
        Type:      models.HealthCheckTypeExternalService,
        Status:    models.HealthStatusHealthy,
        Message:   "Custom service is operational",
        Timestamp: time.Now(),
    }
}
```

---

## Performance Optimization

### System Tuning

#### Database Optimization

```yaml
database:
  connection_pool:
    max_open_connections: 25
    max_idle_connections: 10
    connection_max_lifetime: "5m"
    connection_max_idle_time: "1m"
  query_optimization:
    use_prepared_statements: true
    enable_query_cache: true
    cache_size: 100
  maintenance:
    auto_vacuum: true
    vacuum_schedule: "0 2 * * *"  # Daily at 2 AM
```

#### API Performance

```yaml
api:
  rate_limiting:
    enabled: true
    requests_per_minute: 100
    burst_size: 20
  response_caching:
    enabled: true
    cache_duration: "5m"
    max_cache_size: 100
```

#### File System Optimization

```yaml
file_system:
  io_optimization:
    read_buffer_size: 65536      # 64KB
    write_buffer_size: 65536     # 64KB
    concurrent_operations: 4
  monitoring:
    track_io_performance: true
    log_slow_operations: true
    slow_operation_threshold: "5s"
```

---

## Troubleshooting

### Common Issues and Solutions

#### Task System Issues

**Issue**: Tasks stuck in "started" status

**Solution**:

```bash
# Check task queue status
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/task/queue"

# Cancel stuck tasks
curl -X DELETE -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/task/{task-id}"

# Restart task service if needed
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/system/restart"
```

#### Notification Issues

**Issue**: Notifications not being sent

**Troubleshooting Steps**:

1. Test notification configuration
2. Check notification history for error messages
3. Verify network connectivity
4. Review notification logs
5. Validate provider-specific settings

```bash
# Test specific notification
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/notification/{id}/test"

# Check notification history
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/notification/history"
```

#### File Organization Issues

**Issue**: Files not being organized automatically

**Diagnosis**:

```bash
# Check file organization logs
tail -f /data/logs/radarr.log | grep "file-organization"

# View failed organization attempts
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/file-organization?status=failed"

# Retry failed organizations
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/file-organization/retry"
```

#### Health Monitoring Issues

**Issue**: Health checks failing or missing

**Resolution**:

```bash
# Manually trigger health check
curl -X POST -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/check"

# Review health check configuration
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/config"

# Check system resources
curl -H "X-API-Key: YOUR_API_KEY" \
  "http://localhost:7878/api/v3/health/resources"
```

### Performance Issues

#### High CPU Usage

1. Check active tasks and their resource consumption
2. Review database query performance
3. Optimize indexer check intervals
4. Consider reducing concurrent operations

#### Memory Issues

1. Monitor memory usage trends
2. Check for memory leaks in long-running tasks
3. Adjust database connection pool settings
4. Review cache configurations

#### Disk I/O Issues

1. Monitor disk usage and free space
2. Check file organization operation frequency
3. Consider using hard links instead of copying
4. Optimize database location and settings

### Log Analysis

#### Key Log Locations

- Application logs: `/data/logs/radarr.log`
- Task logs: `/data/logs/tasks/`
- Health check logs: `/data/logs/health/`
- File organization logs: `/data/logs/file-org/`

#### Log Level Configuration

```yaml
logging:
  level: "info"  # debug, info, warn, error
  file_logging: true
  console_logging: true
  max_log_files: 10
  max_log_size: "10MB"
```

---

This comprehensive documentation covers Radarr-Go's advanced features. For additional support, consult the official documentation or community forums. Regular updates to this documentation reflect new features and improvements in the Radarr-Go system.
