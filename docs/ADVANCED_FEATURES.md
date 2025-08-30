# Radarr-Go Advanced Features Documentation

This comprehensive guide covers the advanced capabilities that differentiate Radarr-Go from the original Radarr implementation, focusing on enhanced performance, sophisticated automation, and powerful management features.

## Table of Contents

1. [Task Scheduling System](#task-scheduling-system)
2. [Notification System](#notification-system)
3. [File Organization System](#file-organization-system)
4. [Health Monitoring System](#health-monitoring-system)
5. [Performance Features](#performance-features)
6. [Advanced Configuration](#advanced-configuration)

---

## Task Scheduling System

Radarr-Go features a sophisticated multi-threaded task scheduling system with priority queues, background processing, and comprehensive monitoring capabilities.

### Architecture Overview

The task system is built around three core worker pools:
- **High Priority Queue**: Critical operations like health checks and user-initiated tasks
- **Default Queue**: Standard operations like movie refreshes and import processing
- **Background Queue**: Low-priority maintenance tasks and scheduled operations

### Core Components

#### Task Creation and Management

```yaml
# Example task configuration in config.yaml
tasks:
  maxConcurrentTasks: 10
  queues:
    high-priority:
      maxWorkers: 5
      priority: high
    default:
      maxWorkers: 3
      priority: normal
    background:
      maxWorkers: 2
      priority: low
```

#### Task Types and Commands

**Built-in Task Commands:**
- `RefreshMovie`: Updates movie metadata from external sources
- `SearchMissing`: Searches for missing movies using configured indexers
- `ImportListSync`: Synchronizes import lists and adds new movies
- `HealthCheck`: Performs comprehensive system health verification
- `CleanupRecycleBin`: Removes old files from recycle bin
- `UpdateCheck`: Checks for application updates
- `BackupDatabase`: Creates database backups
- `OptimizeDatabase`: Optimizes database performance

### Configuration and Usage

#### Creating Scheduled Tasks

```bash
# Via API - Create recurring health check every 5 minutes
curl -X POST http://localhost:7878/api/v3/command \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "HealthCheck",
    "scheduledTask": {
      "name": "System Health Check",
      "interval": "5m",
      "priority": "high",
      "enabled": true
    }
  }'
```

#### Task Priority Configuration

Tasks are assigned priorities that determine execution order:

- **High**: Health checks, user-initiated operations, critical system tasks
- **Normal**: Movie refreshes, standard imports, API-triggered operations
- **Low**: Maintenance tasks, cleanup operations, background processing

#### Manual Task Execution

```bash
# Queue immediate movie refresh
curl -X POST http://localhost:7878/api/v3/command \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "RefreshMovie",
    "movieId": 123,
    "priority": "high"
  }'
```

### Monitoring and Management

#### Queue Status Monitoring

```bash
# Check queue status and active tasks
curl -X GET http://localhost:7878/api/v3/system/task \
  -H "X-API-Key: YOUR_API_KEY"
```

Response includes:
- Active worker counts per queue
- Queued task counts
- Currently executing tasks
- Queue performance metrics

#### Task History and Logging

All task executions are logged with:
- Start/end timestamps
- Execution duration
- Success/failure status
- Detailed error messages
- Progress tracking

### Performance Optimization

#### Worker Pool Tuning

Adjust worker counts based on system resources:

```yaml
# High-performance server configuration
tasks:
  queues:
    high-priority:
      maxWorkers: 8    # More workers for critical tasks
    default:
      maxWorkers: 6    # Balanced for normal operations
    background:
      maxWorkers: 4    # Sufficient for maintenance
```

#### Resource Management

- Tasks automatically respect system resource limits
- Database connection pooling prevents resource exhaustion
- Memory usage monitoring prevents system overload
- Automatic task retry with exponential backoff

### Custom Task Development

Create custom task handlers by implementing the `TaskHandler` interface:

```go
type CustomTaskHandler struct {
    name        string
    description string
}

func (h *CustomTaskHandler) Execute(ctx context.Context, task *models.TaskV2,
    updateProgress func(percent int, message string)) error {

    updateProgress(0, "Starting custom task")

    // Implement custom logic here

    updateProgress(100, "Custom task completed")
    return nil
}
```

### Troubleshooting Common Issues

#### Task Queue Backlog
- Monitor queue depths via `/api/v3/system/task`
- Increase worker counts if consistently backlogged
- Check for failing tasks causing retries

#### Long-Running Tasks
- Tasks automatically timeout after 30 minutes
- Monitor task progress via API
- Cancel stuck tasks if necessary

#### Memory Usage
- Tasks with large datasets may consume significant memory
- Monitor system resources during peak times
- Adjust concurrent task limits if needed

---

## Notification System

Radarr-Go provides a comprehensive notification system supporting 11+ notification providers with advanced templating, retry logic, and event-driven messaging.

### Supported Notification Providers

#### Chat and Messaging Platforms
- **Discord**: Rich embed notifications with custom webhooks
- **Slack**: Channel notifications with customizable formatting
- **Telegram**: Bot-based messaging with inline keyboards
- **Signal**: Secure messaging via Signal CLI API

#### Push Notification Services
- **Pushover**: Mobile push notifications with priority levels
- **Pushbullet**: Cross-platform push notifications
- **Gotify**: Self-hosted push notification server
- **Ntfy**: Simple pub-sub push notifications

#### Email Services
- **SMTP Email**: Direct SMTP server integration
- **Mailgun**: Transactional email service
- **SendGrid**: Cloud-based email delivery

#### Advanced Integrations
- **Webhook**: Custom HTTP POST notifications
- **Custom Scripts**: Execute custom scripts with notification data

### Configuration Guide

#### Discord Setup

```yaml
# config.yaml notification configuration
notifications:
  - name: "Discord Movies"
    implementation: "discord"
    enabled: true
    settings:
      webhook: "https://discord.com/api/webhooks/xxx/xxx"
      username: "Radarr-Go"
      avatar: "https://your-server.com/radarr-icon.png"
    # Event triggers
    onGrab: true
    onDownload: true
    onUpgrade: true
    onHealthIssue: true
    onMovieAdded: true
```

#### Email (SMTP) Setup

```yaml
notifications:
  - name: "Email Notifications"
    implementation: "email"
    enabled: true
    settings:
      server: "smtp.gmail.com"
      port: 587
      useSsl: true
      username: "your-email@gmail.com"
      password: "app-password"
      from: "radarr@your-domain.com"
      to: ["admin@your-domain.com", "user@your-domain.com"]
      cc: []
      bcc: []
    onDownload: true
    onHealthIssue: true
    includeHealthWarnings: false
```

#### Pushover Configuration

```yaml
notifications:
  - name: "Pushover Mobile"
    implementation: "pushover"
    enabled: true
    settings:
      userKey: "your-user-key"
      apiKey: "your-app-api-key"
      devices: []  # Empty for all devices
      priority: 0  # -2 to 2 priority levels
      sound: "pushover"
      retry: 60    # Retry interval for emergency priority
      expire: 3600 # Emergency expiration time
    onGrab: true
    onDownload: true
    onHealthIssue: true
```

### Event Types and Triggers

#### Core Movie Events
- **onGrab**: Movie grabbed from indexer
- **onDownload**: Movie successfully downloaded and imported
- **onUpgrade**: Movie file upgraded to better quality
- **onRename**: Movie files renamed
- **onMovieAdded**: New movie added to collection
- **onMovieDelete**: Movie removed from collection
- **onMovieFileDelete**: Movie file deleted

#### System Events
- **onHealthIssue**: System health problems detected
- **onApplicationUpdate**: New application version available
- **onManualInteractionRequired**: Manual intervention needed

### Template System

#### Built-in Templates

Radarr-Go includes sophisticated templates for each event type with rich variable substitution:

**Available Variables:**
- `{{movie.title}}` - Movie title
- `{{movie.year}}` - Movie release year
- `{{movie.imdbId}}` - IMDb identifier
- `{{movie.tmdbId}}` - TMDB identifier
- `{{quality.name}}` - Quality profile name
- `{{quality.resolution}}` - Video resolution
- `{{file.path}}` - File path
- `{{file.size}}` - Human-readable file size
- `{{downloadClient}}` - Download client name
- `{{indexer}}` - Indexer name
- `{{releaseGroup}}` - Release group
- `{{server.name}}` - Server name
- `{{server.url}}` - Server URL

#### Custom Template Example

```yaml
notifications:
  - name: "Discord Custom"
    implementation: "discord"
    settings:
      webhook: "https://discord.com/api/webhooks/xxx/xxx"
    templates:
      onDownload:
        title: "🎬 {{movie.title}} ({{movie.year}}) Downloaded!"
        description: |
          **Quality:** {{quality.name}}
          **Size:** {{file.size}}
          **Path:** {{file.relativePath}}
          **Download Client:** {{downloadClient}}

          [View on TMDB](https://www.themoviedb.org/movie/{{movie.tmdbId}})
        color: "3066993"  # Green color
        footer: "Radarr-Go"
        timestamp: true
```

### Advanced Features

#### Retry Logic and Error Handling

Each notification provider supports configurable retry logic:

```yaml
notifications:
  - name: "Reliable Webhook"
    implementation: "webhook"
    settings:
      url: "https://your-webhook-endpoint.com/notify"
      method: "POST"
      headers:
        Authorization: "Bearer your-token"
      retrySettings:
        maxRetries: 3
        initialDelay: "5s"
        maxDelay: "60s"
        backoffFactor: 2.0
        retryOnTimeout: true
        retryOnServerError: true
```

#### Conditional Notifications

Configure notifications to trigger only under specific conditions:

```yaml
notifications:
  - name: "Critical Only"
    implementation: "pushover"
    settings:
      priority: 2  # Emergency priority
    # Only notify for health issues and high-quality downloads
    onHealthIssue: true
    onDownload: true
    filters:
      minQuality: "HD-1080p"
      minSize: "2GB"
      excludePropers: false
```

#### Notification History and Analytics

Track notification delivery with detailed history:

```bash
# View notification history
curl -X GET http://localhost:7878/api/v3/notification/history \
  -H "X-API-Key: YOUR_API_KEY"
```

Response includes:
- Delivery timestamps
- Success/failure status
- Retry attempts
- Error messages
- Response codes

### Testing and Troubleshooting

#### Test Notifications

```bash
# Test specific notification provider
curl -X POST http://localhost:7878/api/v3/notification/test \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "id": 1,
    "eventType": "Test"
  }'
```

#### Common Issues

**Discord Webhook Failures:**
- Verify webhook URL is correct and active
- Check Discord server permissions
- Ensure rate limits aren't exceeded

**Email Delivery Issues:**
- Verify SMTP credentials and server settings
- Check firewall rules for SMTP ports
- Test with email provider's specific requirements

**Push Notification Problems:**
- Validate API keys and user tokens
- Check device registration
- Verify service-specific requirements

### Performance Considerations

#### Concurrent Delivery
- Notifications are sent concurrently to avoid blocking
- Failed notifications don't impact successful ones
- Delivery timeouts prevent indefinite waits

#### Rate Limiting
- Automatic rate limiting for provider APIs
- Queuing system for high-volume notifications
- Provider-specific throttling configuration

---

## File Organization System

The file organization system provides intelligent automated file management with configurable naming schemes, multiple operation modes, and sophisticated conflict resolution.

### Core Capabilities

#### File Operations
- **Move**: Transfer files to organized locations (default)
- **Copy**: Duplicate files while preserving originals
- **Hardlink**: Create hard links for space efficiency
- **Symlink**: Create symbolic links for advanced setups

#### Naming Templates

Highly flexible naming system with extensive token support:

```yaml
# config.yaml naming configuration
naming:
  movieFolder: "{Movie Title} ({Release Year})"
  movieFile: "{Movie Title} ({Release Year}) {Quality Full}"

  # Advanced template with quality and group info
  movieFileAdvanced: "{Movie Title} ({Release Year}) [{Quality Full}] [{MediaInfo VideoCodec} {MediaInfo AudioCodec}] - {Release Group}"

  # Series-style organization
  movieCollection: "{Movie Collection}/{Movie Title} ({Release Year})"
```

#### Available Naming Tokens

**Movie Information:**
- `{Movie Title}` - Movie title with cleaned characters
- `{Movie TitleThe}` - Movie title with "The" moved to end
- `{Movie OriginalTitle}` - Original movie title
- `{Release Year}` - Movie release year
- `{IMDb Id}` - IMDb identifier
- `{TMDb Id}` - TMDB identifier

**Quality and Media:**
- `{Quality Full}` - Full quality name (e.g., "Bluray-1080p")
- `{Quality Title}` - Quality title (e.g., "1080p")
- `{Quality Proper}` - "PROPER" if proper release
- `{Quality Repack}` - "REPACK" if repack release
- `{MediaInfo VideoCodec}` - Video codec (e.g., "x264")
- `{MediaInfo AudioCodec}` - Audio codec (e.g., "DTS")
- `{MediaInfo AudioChannels}` - Audio channel count

**Release Information:**
- `{Release Group}` - Release group name
- `{Edition Tags}` - Edition information (Director's Cut, etc.)
- `{Custom Formats}` - Custom format tags

**File Information:**
- `{Original Title}` - Original filename without extension
- `{Original Filename}` - Complete original filename

### Configuration Examples

#### Standard Home Media Setup

```yaml
naming:
  movieFolder: "{Movie Title} ({Release Year})"
  movieFile: "{Movie Title} ({Release Year}) - {Quality Title}"

# Results in:
# /movies/The Matrix (1999)/The Matrix (1999) - 1080p.mkv
```

#### Advanced Plex-Optimized Setup

```yaml
naming:
  movieFolder: "{Movie Title} ({Release Year})"
  movieFile: "{Movie Title} ({Release Year}) {Edition-{Edition Tags}} [{Quality Full}] [{MediaInfo VideoCodec}-{MediaInfo AudioCodec}]"

# Results in:
# /movies/The Matrix (1999)/The Matrix (1999) [Bluray-1080p] [x264-DTS].mkv
# /movies/Blade Runner (1982)/Blade Runner (1982) {Edition-Director's Cut} [Bluray-2160p] [x265-Atmos].mkv
```

#### Collection-Based Organization

```yaml
naming:
  movieFolder: "{Movie Collection}/{Movie Title} ({Release Year})"
  movieFile: "{Movie Title} ({Release Year}) - {Quality Title}"
  createMovieCollectionFolders: true

# Results in:
# /movies/The Matrix Collection/The Matrix (1999)/The Matrix (1999) - 1080p.mkv
# /movies/The Matrix Collection/The Matrix Reloaded (2003)/The Matrix Reloaded (2003) - 1080p.mkv
```

### File Operation Modes

#### Move (Default)
- Transfers files to new location
- Original file is removed
- Most space efficient
- Best for single-drive setups

```yaml
fileManagement:
  fileOperation: "move"
  deleteEmptyFolders: true
  createMovieFolders: true
```

#### Copy
- Duplicates files to new location
- Original files preserved
- Useful for backup scenarios
- Requires double storage space

```yaml
fileManagement:
  fileOperation: "copy"
  deleteOriginalFiles: false  # Keep originals
```

#### Hardlink
- Creates hard links instead of copies
- Same file accessible from multiple locations
- Space efficient alternative to copying
- Requires same filesystem

```yaml
fileManagement:
  fileOperation: "hardlink"
  # Automatic fallback to copy if cross-filesystem
  hardlinkFallback: "copy"
```

#### Symlink
- Creates symbolic links
- Useful for complex storage setups
- Requires proper permissions
- Can cross filesystem boundaries

```yaml
fileManagement:
  fileOperation: "symlink"
  createRelativeSymlinks: false  # Use absolute paths
```

### Import Decision Workflow

#### Automated Import Processing

1. **File Discovery**: Scan download folders for new files
2. **Movie Matching**: Match files to movies using filename parsing
3. **Quality Assessment**: Determine video/audio quality
4. **Import Decision**: Apply rules to determine import action
5. **File Organization**: Move/copy files to final location
6. **Database Update**: Update movie file records

#### Import Rules and Filters

```yaml
import:
  # Quality filtering
  minimumQuality: "SDTV"
  upgrades:
    enabled: true
    allowQualityUpgrade: true
    cutoffUnmet: true

  # File filtering
  skipFreeSpaceCheck: false
  minimumFreeSpace: "100MB"
  copyUsingHardlinks: true
  importExtraFiles: true
  extraFileExtensions: "srt,nfo,jpg"

  # Safety checks
  enableMediaInfo: true
  rescanAfterRefresh: "always"
  allowFingerprinting: true
```

### Conflict Resolution

#### Duplicate File Handling

```yaml
fileManagement:
  duplicateHandling:
    # Options: "skip", "replace", "upgrade", "rename"
    onExistingFile: "upgrade"

    # Quality comparison for upgrades
    allowQualityUpgrade: true
    allowSizeUpgrade: true

    # Backup settings
    createBackups: true
    backupFolder: "/backups/radarr"
    maxBackups: 5
```

#### Path Conflicts

- Automatic character replacement for filesystem compatibility
- Configurable path length limits
- Reserved word handling
- Unicode normalization

```yaml
naming:
  replacements:
    # Character replacements for filesystem compatibility
    ":": " -"
    "?": ""
    "<": ""
    ">": ""
    "|": ""
    "*": ""
    "\"": ""
    "/": " - "  # For Windows compatibility

  pathLengthLimit: 240  # Windows compatibility
  unicodeNormalization: true
```

### Monitoring and Troubleshooting

#### Organization History

Track all file organization operations:

```bash
# View recent file operations
curl -X GET http://localhost:7878/api/v3/history?eventType=downloadFolderImported \
  -H "X-API-Key: YOUR_API_KEY"
```

#### Failed Import Analysis

```bash
# Analyze failed imports
curl -X GET http://localhost:7878/api/v3/manualimport \
  -H "X-API-Key: YOUR_API_KEY"
```

Common failure reasons and solutions:

**File Permission Issues:**
- Ensure Radarr-Go has read/write access to all paths
- Check filesystem permissions
- Verify user/group ownership

**Path Length Limitations:**
- Reduce naming template complexity
- Enable path length limiting
- Use shorter folder structures

**Quality Detection Problems:**
- Enable MediaInfo for accurate quality detection
- Update quality definitions
- Check filename parsing rules

**Cross-Filesystem Operations:**
- Hardlinks fail across filesystems - use copy fallback
- Symlinks may have permission restrictions
- Consider mount point configurations

### Performance Optimization

#### Concurrent Processing
- Configurable concurrent import threads
- Parallel file operations for large imports
- Batch processing for efficiency

```yaml
import:
  concurrentImports: 4  # Number of parallel imports
  batchSize: 10        # Files per batch
  processTimeout: "30m" # Maximum processing time
```

#### Storage Efficiency
- Automatic cleanup of empty directories
- Duplicate file detection and removal
- Space usage monitoring and alerting

---

## Health Monitoring System

Radarr-Go includes a comprehensive health monitoring system that continuously watches system status, performance metrics, and configuration integrity.

### Health Check Categories

#### System Health Checks
- **Disk Space**: Monitor available storage across all configured paths
- **Database Connectivity**: Verify database connections and performance
- **Network Connectivity**: Test internet and indexer accessibility
- **Permission Validation**: Check file/folder access permissions
- **Memory Usage**: Monitor application memory consumption
- **CPU Usage**: Track processing load and bottlenecks

#### Application Health Checks
- **Indexer Status**: Verify indexer availability and authentication
- **Download Client Status**: Check download client connectivity
- **Import List Status**: Validate import list functionality
- **Notification Status**: Test notification provider configurations
- **API Endpoint Health**: Monitor API responsiveness

#### Configuration Health Checks
- **Quality Profile Validation**: Ensure quality profiles are properly configured
- **Naming Configuration**: Verify naming templates are valid
- **Path Configuration**: Check all configured paths exist and are accessible
- **SSL Certificate Validation**: Monitor certificate expiration
- **Database Schema**: Verify database structure integrity

### Configuration

#### Basic Health Configuration

```yaml
# config.yaml health monitoring settings
health:
  enabled: true
  interval: "5m"              # Check interval
  retentionDays: 30           # Keep health data for 30 days

  # Thresholds
  diskSpaceWarningThreshold: "5GB"    # Warn below 5GB
  diskSpaceCriticalThreshold: "1GB"   # Critical below 1GB
  databaseTimeoutThreshold: "10s"     # Database query timeout
  externalServiceTimeout: "15s"       # External service timeout

  # Notification settings
  notifyCriticalIssues: true          # Send notifications for critical issues
  notifyWarningIssues: false          # Don't notify for warnings

  # Performance monitoring
  enablePerformanceMonitoring: true
  performanceMetricInterval: "1m"
  memoryUsageThreshold: "1GB"
  cpuUsageThreshold: 80               # Percent
```

#### Advanced Health Monitoring

```yaml
health:
  # Custom health checks
  customChecks:
    - name: "Custom API Check"
      type: "http"
      url: "https://api.example.com/health"
      method: "GET"
      timeout: "30s"
      expectedStatus: 200
      interval: "10m"

    - name: "Custom Script Check"
      type: "script"
      script: "/scripts/custom-health-check.sh"
      timeout: "60s"
      interval: "15m"

  # Per-check configuration
  checks:
    diskSpace:
      enabled: true
      paths: ["/movies", "/downloads", "/data"]
      warningThreshold: "10GB"
      criticalThreshold: "2GB"

    database:
      enabled: true
      queryTimeout: "5s"
      connectionPoolCheck: true

    indexers:
      enabled: true
      timeout: "30s"
      checkFrequency: "15m"

    downloadClients:
      enabled: true
      timeout: "30s"
      testDownload: false
```

### Health Issue Management

#### Issue Severity Levels
- **OK**: All systems operating normally
- **Notice**: Informational messages, no action required
- **Warning**: Issues that may impact functionality but don't prevent operation
- **Error**: Problems that prevent or significantly impact functionality
- **Fatal**: Critical issues that prevent application startup or core functionality

#### Automatic Issue Resolution

Many health issues include automatic resolution capabilities:

```yaml
health:
  autoResolve:
    enabled: true

    # Automatic fixes
    createMissingFolders: true          # Create missing configured folders
    fixFilePermissions: false           # Don't auto-fix permissions (security)
    restartFailedServices: true         # Restart failed background services
    cleanupTempFiles: true              # Remove old temporary files
    optimizeDatabase: true              # Auto-optimize database when needed

    # Limits
    maxAutoActions: 10                  # Maximum automatic actions per hour
    requireConfirmation: false          # Auto-execute or require confirmation
```

#### Issue Notifications

Health issues automatically trigger notifications when configured:

```yaml
notifications:
  - name: "Health Alerts"
    implementation: "email"
    enabled: true
    settings:
      to: ["admin@example.com"]
    onHealthIssue: true
    healthNotificationSettings:
      minimumSeverity: "warning"        # Notify on warnings and above
      aggregateIssues: true             # Group related issues
      resolvedNotifications: true       # Notify when issues resolve
      maxNotificationsPerHour: 5        # Rate limiting
```

### Monitoring Dashboard

#### API Endpoints

```bash
# Get overall health status
curl -X GET http://localhost:7878/api/v3/health \
  -H "X-API-Key: YOUR_API_KEY"

# Get detailed health report
curl -X GET http://localhost:7878/api/v3/health/detailed \
  -H "X-API-Key: YOUR_API_KEY"

# Get health history
curl -X GET http://localhost:7878/api/v3/health/history?days=7 \
  -H "X-API-Key: YOUR_API_KEY"

# Run specific health check
curl -X POST http://localhost:7878/api/v3/health/check/diskspace \
  -H "X-API-Key: YOUR_API_KEY"
```

#### Health Status Response

```json
{
  "overallStatus": "warning",
  "lastChecked": "2024-01-15T10:30:00Z",
  "summary": {
    "total": 12,
    "ok": 9,
    "warning": 2,
    "error": 1,
    "fatal": 0
  },
  "checks": [
    {
      "type": "diskSpace",
      "source": "Movies Drive",
      "status": "warning",
      "message": "Disk space below warning threshold",
      "details": {
        "path": "/movies",
        "available": "3.2GB",
        "threshold": "5GB",
        "used": "89%"
      },
      "timestamp": "2024-01-15T10:30:00Z",
      "canAutoResolve": false
    }
  ]
}
```

### Performance Metrics

#### System Performance Monitoring

```yaml
health:
  performanceMonitoring:
    enabled: true

    # Metrics to collect
    systemMetrics:
      cpu: true
      memory: true
      diskIO: true
      networkIO: true

    # Application metrics
    applicationMetrics:
      requestLatency: true
      databaseQueryTime: true
      importProcessingTime: true
      notificationDeliveryTime: true

    # Retention
    metricsRetention: "7d"
    aggregationInterval: "5m"
```

#### Performance Alerting

```yaml
health:
  performanceAlerts:
    - metric: "memory_usage_percent"
      threshold: 85
      duration: "5m"
      severity: "warning"

    - metric: "database_query_time_ms"
      threshold: 5000
      duration: "1m"
      severity: "error"

    - metric: "disk_io_wait_percent"
      threshold: 50
      duration: "10m"
      severity: "warning"
```

### Custom Health Checks

#### HTTP Endpoint Checks

```yaml
health:
  customChecks:
    - name: "TMDB API"
      type: "http"
      url: "https://api.themoviedb.org/3/configuration"
      method: "GET"
      headers:
        Authorization: "Bearer YOUR_TMDB_TOKEN"
      expectedStatus: 200
      timeout: "30s"
      interval: "30m"

    - name: "Indexer Health"
      type: "http"
      url: "https://your-indexer.com/api"
      method: "GET"
      expectedContent: "ok"
      timeout: "15s"
      interval: "15m"
```

#### Script-Based Checks

```yaml
health:
  customChecks:
    - name: "Storage Array Status"
      type: "script"
      script: "/scripts/check-storage-array.sh"
      timeout: "60s"
      interval: "10m"
      expectedExitCode: 0

    - name: "Network Latency Check"
      type: "script"
      script: "/scripts/check-network-latency.py"
      arguments: ["--target", "8.8.8.8", "--timeout", "5"]
      timeout: "30s"
      interval: "5m"
```

### Troubleshooting Common Health Issues

#### Disk Space Issues
```bash
# Check disk usage
df -h /movies /downloads

# Find large files
find /movies -type f -size +1G -exec ls -lh {} \;

# Clean up old downloads
find /downloads -type f -mtime +7 -delete
```

#### Database Connection Issues
```bash
# Test database connectivity
pg_isready -h localhost -p 5432 -U radarr

# Check database locks
SELECT * FROM pg_locks WHERE NOT granted;

# Optimize database
VACUUM ANALYZE;
```

#### Permission Problems
```bash
# Check file ownership
ls -la /movies /downloads

# Fix ownership (adjust user/group as needed)
sudo chown -R radarr:radarr /movies /downloads

# Set proper permissions
sudo chmod -R 755 /movies /downloads
```

#### Memory Usage Issues
```bash
# Monitor memory usage
htop

# Check for memory leaks
ps aux --sort=-%mem | head

# Restart Radarr-Go if memory usage excessive
systemctl restart radarr-go
```

---

## Performance Features

Radarr-Go is architected from the ground up for superior performance compared to the original .NET implementation, with numerous optimizations and modern Go runtime benefits.

### Runtime Performance Advantages

#### Memory Efficiency
- **Native Compiled Binary**: No runtime overhead from JIT compilation or garbage collection pressure
- **Efficient Memory Usage**: Typically 50-80% lower memory consumption than .NET Radarr
- **Fast Startup**: Application starts in seconds, not minutes
- **Minimal Memory Footprint**: Base memory usage under 100MB vs 500MB+ for .NET version

#### Concurrency and Parallelism
- **Goroutine-based Concurrency**: Lightweight threading model enables thousands of concurrent operations
- **Non-blocking I/O**: Efficient handling of network requests and file operations
- **Parallel Processing**: Multiple movies/files processed simultaneously
- **Lock-free Data Structures**: Reduced contention in high-concurrency scenarios

### Database Optimizations

#### Connection Management
```yaml
database:
  postgres:
    # Connection pooling configuration
    maxOpenConns: 25        # Maximum open connections
    maxIdleConns: 10        # Maximum idle connections
    connMaxLifetime: "5m"   # Connection lifetime
    connMaxIdleTime: "30s"  # Idle connection timeout

    # Performance tuning
    sslMode: "disable"      # Disable SSL for internal networks (performance)
    timezone: "UTC"         # Consistent timezone handling

    # Query optimization
    preparedStatements: true    # Use prepared statements
    queryTimeout: "30s"         # Query timeout
    slowQueryThreshold: "1s"    # Log slow queries
```

#### Query Optimization
- **Prepared Statements**: All database queries use prepared statements for optimal performance
- **Index Optimization**: Carefully designed indexes for common query patterns
- **Batch Operations**: Multiple records processed in single transactions
- **Query Caching**: Frequently accessed data cached in memory

#### GORM Enhancements
- **Custom Hooks**: Validation and business logic implemented efficiently
- **Preloading Optimization**: Strategic relationship loading to minimize N+1 queries
- **Transaction Management**: Proper transaction boundaries for data consistency
- **Connection Pool Monitoring**: Real-time monitoring of connection pool usage

### HTTP Performance

#### Gin Framework Optimizations
```yaml
server:
  # Performance configuration
  maxHeaderSize: "1MB"
  readTimeout: "30s"
  writeTimeout: "30s"
  idleTimeout: "120s"

  # Request handling
  maxRequestSize: "100MB"      # For large import operations
  enableCompression: true      # Gzip compression
  compressionLevel: 6          # Balance compression vs CPU

  # Keep-alive settings
  keepAliveEnabled: true
  keepAliveTimeout: "60s"
```

#### Middleware Efficiency
- **Minimal Middleware Stack**: Only essential middleware enabled
- **Fast Routing**: Gin's radix tree router provides O(1) route matching
- **JSON Optimization**: Custom JSON encoding for better performance
- **Request Pooling**: Object reuse reduces garbage collection pressure

### I/O Performance

#### File Operations
- **Optimized File Copying**: Uses system calls for maximum throughput
- **Parallel File Processing**: Multiple files processed concurrently
- **Memory-Mapped Files**: For large file operations when beneficial
- **Efficient Disk Usage**: Smart allocation and cleanup strategies

```yaml
fileManagement:
  # I/O performance settings
  copyBufferSize: "1MB"        # Buffer size for file operations
  concurrentCopies: 4          # Parallel file operations
  checksumVerification: true   # Verify file integrity (slight performance cost)

  # Disk optimization
  preallocateFiles: true       # Reduce fragmentation
  syncWrites: false            # Async writes for performance
  directIO: false              # Use OS caching (usually better)
```

#### Network I/O
- **HTTP Client Pool**: Reuse connections for external API calls
- **Request Queuing**: Manage concurrent requests to external services
- **Timeout Management**: Prevent hanging requests from impacting performance
- **Rate Limiting**: Respect external service limits while maintaining throughput

### Memory Management

#### Garbage Collection Optimization
```bash
# Tuning Go GC for optimal performance
export GOGC=100              # Default GC target (adjust based on memory availability)
export GOMEMLIMIT=2GiB       # Memory limit (prevents OOM)
export GODEBUG=gctrace=1     # Enable GC tracing for monitoring
```

#### Memory Pool Usage
- **Object Pooling**: Reuse frequently allocated objects
- **Buffer Pools**: Efficient buffer management for I/O operations
- **String Optimization**: Minimize string allocations in hot paths
- **Slice Optimization**: Pre-allocate slices with appropriate capacity

### Caching Strategies

#### In-Memory Caching
```yaml
cache:
  # Movie metadata caching
  movieCache:
    enabled: true
    maxSize: 10000           # Maximum number of movies cached
    ttl: "1h"               # Time to live

  # Quality profile caching
  qualityCache:
    enabled: true
    maxSize: 100
    ttl: "24h"

  # Search result caching
  searchCache:
    enabled: true
    maxSize: 1000
    ttl: "15m"
```

#### Database Query Caching
- **Query Result Caching**: Cache frequently accessed data
- **Smart Invalidation**: Automatic cache invalidation on data changes
- **Memory-Safe Caching**: Bounded cache sizes prevent memory exhaustion
- **Cache Warming**: Preload commonly accessed data on startup

### Monitoring and Profiling

#### Performance Metrics Collection
```yaml
monitoring:
  enabled: true

  # Runtime metrics
  collectRuntimeMetrics: true
  metricsInterval: "30s"

  # Custom metrics
  httpRequestMetrics: true
  databaseQueryMetrics: true
  fileOperationMetrics: true

  # Export options
  prometheusExport: true
  prometheusPath: "/metrics"
```

#### Built-in Profiling
```bash
# Enable pprof endpoints for performance analysis
curl http://localhost:7878/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Memory profile
curl http://localhost:7878/debug/pprof/heap > mem.prof
go tool pprof mem.prof

# Goroutine analysis
curl http://localhost:7878/debug/pprof/goroutine > goroutine.prof
go tool pprof goroutine.prof
```

### Benchmarking and Testing

#### Performance Benchmarks
```bash
# Run performance benchmarks
make test-bench

# Specific benchmark categories
go test -bench=BenchmarkMovieService ./internal/services
go test -bench=BenchmarkDatabase ./internal/database
go test -bench=BenchmarkAPI ./internal/api
```

#### Load Testing
```bash
# API load testing with Apache Bench
ab -n 1000 -c 10 http://localhost:7878/api/v3/movie

# Database load testing
pgbench -c 10 -j 2 -t 1000 radarr_db

# File operation stress testing
./scripts/stress-test-file-operations.sh
```

### Production Performance Tuning

#### System-Level Optimizations
```bash
# Increase file descriptor limits
echo "radarr soft nofile 65536" >> /etc/security/limits.conf
echo "radarr hard nofile 65536" >> /etc/security/limits.conf

# Optimize TCP settings for high-throughput
echo "net.core.rmem_max = 16777216" >> /etc/sysctl.conf
echo "net.core.wmem_max = 16777216" >> /etc/sysctl.conf

# Disable swap for consistent performance
swapoff -a
```

#### Container Optimizations
```yaml
# Docker Compose performance settings
services:
  radarr-go:
    image: radarr/radarr-go:latest
    environment:
      - GOGC=80                    # More aggressive GC for containers
      - GOMEMLIMIT=1GiB           # Container memory limit
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '2'
        reservations:
          memory: 512M
          cpus: '1'
    sysctls:
      - net.core.somaxconn=1024   # Increase connection queue
    ulimits:
      nofile: 65536               # File descriptor limit
```

---

## Advanced Configuration

### Environment-Based Configuration

Radarr-Go supports comprehensive environment variable configuration with the `RADARR_` prefix:

```bash
# Core application settings
export RADARR_SERVER_PORT=7878
export RADARR_SERVER_BIND_ADDRESS="0.0.0.0"
export RADARR_SERVER_URL_BASE="/radarr"
export RADARR_LOG_LEVEL="info"

# Database configuration
export RADARR_DATABASE_TYPE="postgres"
export RADARR_DATABASE_HOST="localhost"
export RADARR_DATABASE_PORT=5432
export RADARR_DATABASE_USERNAME="radarr"
export RADARR_DATABASE_PASSWORD="secure_password"
export RADARR_DATABASE_NAME="radarr_main"

# Security settings
export RADARR_AUTH_METHOD="forms"
export RADARR_AUTH_API_KEY="your-secure-api-key"
export RADARR_AUTH_USERNAME="admin"
export RADARR_AUTH_PASSWORD="hashed_password"

# Feature toggles
export RADARR_FEATURES_ENABLE_CALENDAR=true
export RADARR_FEATURES_ENABLE_NOTIFICATIONS=true
export RADARR_FEATURES_ENABLE_HEALTH_CHECKS=true
```

### Multi-Environment Deployment

#### Development Configuration
```yaml
# config.dev.yaml
server:
  port: 7878
  logLevel: debug
  enableProfiling: true

database:
  type: postgres
  host: localhost
  username: radarr_dev

health:
  enabled: false  # Disable for development

features:
  enableTelemetry: false
```

#### Production Configuration
```yaml
# config.prod.yaml
server:
  port: 7878
  logLevel: info
  enableProfiling: false

database:
  type: postgres
  host: db-cluster.internal
  connectionPooling:
    maxOpenConns: 25
    maxIdleConns: 10

security:
  enableCSRF: true
  secureCookies: true

health:
  enabled: true
  interval: "5m"
```

### Configuration Validation

Radarr-Go performs comprehensive configuration validation on startup:

```bash
# Validate configuration without starting
./radarr --config-test

# Validate specific configuration file
./radarr --config config.prod.yaml --config-test

# Generate configuration template
./radarr --generate-config > config.yaml
```

This comprehensive feature documentation showcases the advanced capabilities that make Radarr-Go a superior choice for modern movie collection management, emphasizing performance, reliability, and extensive customization options.
