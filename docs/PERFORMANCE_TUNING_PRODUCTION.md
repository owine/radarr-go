# Production Performance Tuning Guide

This guide provides comprehensive performance optimization strategies for Radarr Go production deployments, covering database optimization, Go runtime tuning, memory management, and scaling considerations.

## Table of Contents

1. [Overview](#overview)
2. [Database Performance Optimization](#database-performance-optimization)
3. [Go Runtime Tuning](#go-runtime-tuning)
4. [Memory Management](#memory-management)
5. [Concurrent Processing Optimization](#concurrent-processing-optimization)
6. [Storage Performance](#storage-performance)
7. [Network Optimization](#network-optimization)
8. [API Performance Tuning](#api-performance-tuning)
9. [Monitoring and Profiling](#monitoring-and-profiling)
10. [Automated Performance Scripts](#automated-performance-scripts)

## Overview

Radarr Go achieves significant performance improvements over the original .NET implementation through:

### Performance Advantages

- **60-80% Lower Memory Usage**: Efficient garbage collection and memory pooling
- **3-5x Faster API Responses**: Optimized HTTP handling and database queries
- **Better Concurrency**: Native Go goroutines for parallel processing
- **Reduced CPU Overhead**: Compiled binary with no JIT compilation overhead
- **Optimized Database Queries**: Prepared statements and connection pooling

### Key Performance Metrics

- **API Response Time**: Target <200ms for 95th percentile
- **Memory Usage**: <500MB for libraries with 10,000+ movies
- **Database Connection Pool**: Optimal utilization at 70-80%
- **Concurrent Downloads**: Support for 50+ simultaneous downloads
- **CPU Usage**: <50% under normal load

## Database Performance Optimization

### PostgreSQL Optimization

```sql
-- postgresql.conf optimizations for Radarr Go

# Memory Settings
shared_buffers = 256MB                    # 25% of system RAM (for dedicated DB server)
effective_cache_size = 1GB                # 75% of system RAM
work_mem = 4MB                           # Per-query memory for sorting/hashing
maintenance_work_mem = 64MB              # Memory for VACUUM, CREATE INDEX, etc.

# Connection Settings
max_connections = 100                     # Match application pool size + overhead
superuser_reserved_connections = 3

# WAL Settings (Write-Ahead Logging)
wal_buffers = 16MB                       # WAL buffer size
checkpoint_completion_target = 0.9       # Spread checkpoints over time
checkpoint_timeout = 10min               # Maximum time between checkpoints
max_wal_size = 4GB                       # Maximum WAL size between checkpoints
min_wal_size = 1GB                       # Minimum WAL size to keep

# Query Planner Settings
random_page_cost = 1.1                   # SSD storage cost
effective_io_concurrency = 200           # Expected concurrent I/O operations
default_statistics_target = 100          # Statistics detail for query planning

# Autovacuum Settings
autovacuum = on
autovacuum_max_workers = 3
autovacuum_naptime = 1min
autovacuum_vacuum_threshold = 50
autovacuum_analyze_threshold = 50
autovacuum_vacuum_scale_factor = 0.2
autovacuum_analyze_scale_factor = 0.1

# Logging for Performance Analysis
log_min_duration_statement = 1000        # Log queries taking >1 second
log_checkpoints = on
log_connections = on
log_disconnections = on
log_lock_waits = on
log_temp_files = 10MB                    # Log temp files >10MB

# Performance Monitoring
shared_preload_libraries = 'pg_stat_statements'
pg_stat_statements.max = 10000
pg_stat_statements.track = all
```

### Database Index Optimization

```sql
-- Critical indexes for Radarr Go performance

-- Movies table indexes
CREATE INDEX CONCURRENTLY idx_movies_tmdb_id ON movies(tmdb_id);
CREATE INDEX CONCURRENTLY idx_movies_status ON movies(status);
CREATE INDEX CONCURRENTLY idx_movies_monitored ON movies(monitored);
CREATE INDEX CONCURRENTLY idx_movies_quality_profile_id ON movies(quality_profile_id);
CREATE INDEX CONCURRENTLY idx_movies_added ON movies(added);
CREATE INDEX CONCURRENTLY idx_movies_year ON movies(year);
CREATE INDEX CONCURRENTLY idx_movies_title_gin ON movies USING gin(to_tsvector('english', title));

-- Movie files table indexes
CREATE INDEX CONCURRENTLY idx_movie_files_movie_id ON movie_files(movie_id);
CREATE INDEX CONCURRENTLY idx_movie_files_relative_path ON movie_files(relative_path);
CREATE INDEX CONCURRENTLY idx_movie_files_date_added ON movie_files(date_added);
CREATE INDEX CONCURRENTLY idx_movie_files_quality ON movie_files USING gin(quality);

-- Download queue indexes
CREATE INDEX CONCURRENTLY idx_download_queue_status ON download_queue(status);
CREATE INDEX CONCURRENTLY idx_download_queue_movie_id ON download_queue(movie_id);
CREATE INDEX CONCURRENTLY idx_download_queue_added ON download_queue(added);

-- History table indexes (for large datasets)
CREATE INDEX CONCURRENTLY idx_history_date ON history(date);
CREATE INDEX CONCURRENTLY idx_history_movie_id ON history(movie_id);
CREATE INDEX CONCURRENTLY idx_history_event_type ON history(event_type);

-- Partial indexes for common queries
CREATE INDEX CONCURRENTLY idx_movies_wanted
ON movies(added) WHERE monitored = true AND status = 'wanted';

CREATE INDEX CONCURRENTLY idx_movies_missing_files
ON movies(id) WHERE monitored = true AND has_file = false;

-- Composite indexes for complex queries
CREATE INDEX CONCURRENTLY idx_movies_quality_status
ON movies(quality_profile_id, status) WHERE monitored = true;
```

### Connection Pool Configuration

```yaml
# config.yaml - Database connection pool settings
database:
  type: "postgres"
  host: "postgres"
  port: 5432
  database: "radarr"
  username: "radarr"

  # Connection Pool Settings
  max_connections: 25              # Maximum connections in pool
  max_idle_connections: 5          # Idle connections to maintain
  connection_timeout: "30s"        # Connection establishment timeout
  idle_timeout: "10m"              # Time before idle connection is closed
  max_lifetime: "1h"               # Maximum connection lifetime

  # Performance Settings
  enable_prepared_statements: true  # Use prepared statements for performance
  enable_query_logging: false      # Disable in production (enable for debugging)
  slow_query_threshold: "1s"       # Log queries slower than 1 second

  # PostgreSQL Specific
  ssl_mode: "prefer"               # SSL preference
  application_name: "radarr-go"    # Connection identifier
```

### Database Query Optimization

```bash
#!/bin/bash
# scripts/optimize-database.sh
# Database optimization and maintenance script

set -euo pipefail

POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_USER="${POSTGRES_USER:-radarr}"
POSTGRES_DB="${POSTGRES_DB:-radarr}"

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }

# Analyze query performance
analyze_slow_queries() {
    log "Analyzing slow queries..."

    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOF'
-- Top 20 slowest queries by average execution time
SELECT
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements
WHERE query NOT LIKE '%pg_stat_statements%'
  AND calls > 10
ORDER BY mean_time DESC
LIMIT 20;

-- Table sizes and bloat analysis
SELECT
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation,
    most_common_vals,
    most_common_freqs
FROM pg_stats
WHERE schemaname = 'public'
  AND tablename IN ('movies', 'movie_files', 'download_queue', 'history')
ORDER BY tablename, attname;

-- Index usage statistics
SELECT
    schemaname,
    tablename,
    indexname,
    idx_tup_read,
    idx_tup_fetch,
    idx_scan,
    idx_tup_read + idx_tup_fetch as total_reads
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY total_reads DESC;
EOF
}

# Update table statistics
update_statistics() {
    log "Updating table statistics..."

    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOF'
-- Analyze all tables for better query planning
ANALYZE movies;
ANALYZE movie_files;
ANALYZE download_queue;
ANALYZE history;
ANALYZE quality_profiles;
ANALYZE indexers;
ANALYZE notifications;

-- Update extended statistics if available (PostgreSQL 10+)
SELECT version();
EOF
}

# Optimize tables
optimize_tables() {
    log "Optimizing tables..."

    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOF'
-- Vacuum and reindex critical tables
VACUUM ANALYZE movies;
VACUUM ANALYZE movie_files;
VACUUM ANALYZE download_queue;

-- Full vacuum for heavily updated tables (schedule during maintenance window)
-- VACUUM FULL history;

-- Reindex if needed (causes locks, run during maintenance)
-- REINDEX INDEX CONCURRENTLY idx_movies_tmdb_id;
-- REINDEX INDEX CONCURRENTLY idx_movie_files_movie_id;
EOF
}

# Check for missing indexes
suggest_indexes() {
    log "Checking for missing indexes..."

    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" << 'EOF'
-- Find tables with sequential scans
SELECT
    schemaname,
    tablename,
    seq_scan,
    seq_tup_read,
    seq_tup_read / seq_scan as avg_seq_read,
    idx_scan,
    idx_tup_fetch
FROM pg_stat_user_tables
WHERE seq_scan > 0
ORDER BY seq_tup_read DESC;

-- Check for unused indexes (potential candidates for removal)
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch,
    pg_size_pretty(pg_relation_size(indexrelname::regclass)) as size
FROM pg_stat_user_indexes
WHERE idx_scan = 0
  AND schemaname = 'public'
ORDER BY pg_relation_size(indexrelname::regclass) DESC;
EOF
}

# Main execution
case "${1:-analyze}" in
    "analyze") analyze_slow_queries ;;
    "statistics") update_statistics ;;
    "optimize") optimize_tables ;;
    "indexes") suggest_indexes ;;
    "full")
        analyze_slow_queries
        update_statistics
        suggest_indexes
        ;;
    *)
        echo "Usage: $0 {analyze|statistics|optimize|indexes|full}"
        exit 1
        ;;
esac
```

## Go Runtime Tuning

### Environment Variables for Production

```bash
# .env - Go runtime optimization
# Memory Management
GOMEMLIMIT=2GiB                    # Set memory limit (Go 1.19+)
GOGC=100                          # GC percentage (default 100)
GOMEMLIMIT=soft                   # Soft memory limit

# CPU and Concurrency
GOMAXPROCS=4                      # Number of OS threads (default: CPU count)
GODEBUG=""                        # Disable debug features in production

# Network
GODEBUG=http2server=0            # Disable HTTP/2 if causing issues
GOTRACEBACK=none                 # Minimal stack traces in production

# Application-specific
RADARR_GO_BUFFER_SIZE=65536      # I/O buffer size (64KB)
RADARR_GO_WORKER_POOL_SIZE=50    # Background worker pool size
RADARR_GO_MAX_CONCURRENT_DOWNLOADS=25  # Download concurrency
```

### Runtime Configuration

```yaml
# config.yaml - Runtime performance settings
performance:
  # Memory settings
  gc_percentage: 100               # Garbage collection trigger percentage
  memory_limit: "2Gi"              # Memory limit (requires Go 1.19+)

  # Concurrency settings
  max_procs: 0                     # 0 = auto-detect CPU count
  worker_pool_size: 50             # Background worker pool
  api_concurrency: 100             # Maximum concurrent API requests

  # I/O settings
  buffer_size: 65536               # Default buffer size (64KB)
  read_timeout: "30s"              # Read timeout for operations
  write_timeout: "30s"             # Write timeout for operations

  # Download settings
  max_concurrent_downloads: 25     # Simultaneous downloads
  download_retry_count: 3          # Download retry attempts
  download_retry_delay: "5s"       # Delay between retries

  # File processing
  parallel_file_operations: 10     # Concurrent file operations
  file_scan_concurrency: 5         # Concurrent file scanning

  # Cache settings
  enable_response_caching: true    # Enable API response caching
  cache_ttl: "5m"                 # Cache time-to-live
  max_cache_size: "100MB"         # Maximum cache size
```

### Garbage Collection Tuning

```go
// Example: Fine-tuning GC in main.go
package main

import (
    "runtime"
    "runtime/debug"
    "time"
)

func optimizeRuntime() {
    // Set GC percentage based on memory availability
    memLimit := getMemoryLimit() // Custom function to detect memory
    if memLimit > 4*1024*1024*1024 { // 4GB
        debug.SetGCPercent(75)  // Less frequent GC for high memory
    } else {
        debug.SetGCPercent(100) // Default GC frequency
    }

    // Set memory limit if available (Go 1.19+)
    if memLimit > 0 {
        debug.SetMemoryLimit(int64(memLimit * 0.8)) // 80% of available memory
    }

    // Configure GOMAXPROCS if not set
    if runtime.GOMAXPROCS(0) == 1 && runtime.NumCPU() > 1 {
        runtime.GOMAXPROCS(runtime.NumCPU())
    }

    // Force initial GC to establish baseline
    runtime.GC()
}

func getMemoryLimit() int {
    // Implement memory detection logic
    // Could read from cgroups, /proc/meminfo, or environment variables
    return 0
}
```

### Performance Monitoring Integration

```go
// Example: Runtime metrics collection
package metrics

import (
    "runtime"
    "time"

    "github.com/prometheus/client_golang/prometheus"
)

var (
    // Go runtime metrics
    goGoroutines = prometheus.NewGaugeFunc(
        prometheus.GaugeOpts{
            Name: "go_goroutines",
            Help: "Number of goroutines that currently exist.",
        },
        func() float64 { return float64(runtime.NumGoroutine()) },
    )

    goMemStats = prometheus.NewGaugeFunc(
        prometheus.GaugeOpts{
            Name: "go_memstats_alloc_bytes",
            Help: "Number of bytes allocated and still in use.",
        },
        func() float64 {
            var m runtime.MemStats
            runtime.ReadMemStats(&m)
            return float64(m.Alloc)
        },
    )

    goGCDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "go_gc_duration_seconds",
            Help: "Time spent in garbage collection.",
        },
        []string{"phase"},
    )
)

func init() {
    prometheus.MustRegister(goGoroutines, goMemStats, goGCDuration)
}
```

## Memory Management

### Memory Pool Implementation

```go
// Example: Buffer pool for reducing allocations
package pool

import (
    "bytes"
    "sync"
)

var (
    // Buffer pools for different sizes
    smallBufferPool = &sync.Pool{
        New: func() interface{} {
            return bytes.NewBuffer(make([]byte, 0, 1024)) // 1KB
        },
    }

    largeBufferPool = &sync.Pool{
        New: func() interface{} {
            return bytes.NewBuffer(make([]byte, 0, 64*1024)) // 64KB
        },
    }
)

func GetSmallBuffer() *bytes.Buffer {
    return smallBufferPool.Get().(*bytes.Buffer)
}

func PutSmallBuffer(buf *bytes.Buffer) {
    buf.Reset()
    smallBufferPool.Put(buf)
}

func GetLargeBuffer() *bytes.Buffer {
    return largeBufferPool.Get().(*bytes.Buffer)
}

func PutLargeBuffer(buf *bytes.Buffer) {
    buf.Reset()
    largeBufferPool.Put(buf)
}
```

### Memory Profiling Script

```bash
#!/bin/bash
# scripts/memory-profile.sh
# Memory profiling and analysis

set -euo pipefail

RADARR_URL="${RADARR_URL:-http://localhost:7878}"
PROFILE_DURATION="${PROFILE_DURATION:-30}"
OUTPUT_DIR="profiles/$(date +%Y%m%d_%H%M%S)"

mkdir -p "$OUTPUT_DIR"

log() { echo "[$(date +'%H:%M:%S')] $1"; }

# Collect memory profile
collect_memory_profile() {
    log "Collecting memory profile for ${PROFILE_DURATION} seconds..."

    curl -o "$OUTPUT_DIR/heap.prof" \
        "$RADARR_URL/debug/pprof/heap"

    curl -o "$OUTPUT_DIR/allocs.prof" \
        "$RADARR_URL/debug/pprof/allocs"

    log "Memory profiles saved to $OUTPUT_DIR"
}

# Collect CPU profile
collect_cpu_profile() {
    log "Collecting CPU profile for ${PROFILE_DURATION} seconds..."

    curl -o "$OUTPUT_DIR/cpu.prof" \
        "$RADARR_URL/debug/pprof/profile?seconds=$PROFILE_DURATION"

    log "CPU profile saved to $OUTPUT_DIR"
}

# Collect goroutine profile
collect_goroutine_profile() {
    log "Collecting goroutine profile..."

    curl -o "$OUTPUT_DIR/goroutine.prof" \
        "$RADARR_URL/debug/pprof/goroutine"

    log "Goroutine profile saved to $OUTPUT_DIR"
}

# Analyze profiles
analyze_profiles() {
    log "Analyzing profiles..."

    if command -v go >/dev/null 2>&1; then
        # Heap analysis
        if [ -f "$OUTPUT_DIR/heap.prof" ]; then
            echo "=== Memory Usage (Top 20) ===" > "$OUTPUT_DIR/analysis.txt"
            go tool pprof -text -nodecount=20 "$OUTPUT_DIR/heap.prof" >> "$OUTPUT_DIR/analysis.txt" 2>/dev/null || true
            echo "" >> "$OUTPUT_DIR/analysis.txt"
        fi

        # CPU analysis
        if [ -f "$OUTPUT_DIR/cpu.prof" ]; then
            echo "=== CPU Usage (Top 20) ===" >> "$OUTPUT_DIR/analysis.txt"
            go tool pprof -text -nodecount=20 "$OUTPUT_DIR/cpu.prof" >> "$OUTPUT_DIR/analysis.txt" 2>/dev/null || true
            echo "" >> "$OUTPUT_DIR/analysis.txt"
        fi

        # Goroutine analysis
        if [ -f "$OUTPUT_DIR/goroutine.prof" ]; then
            echo "=== Goroutines ===" >> "$OUTPUT_DIR/analysis.txt"
            go tool pprof -text "$OUTPUT_DIR/goroutine.prof" >> "$OUTPUT_DIR/analysis.txt" 2>/dev/null || true
        fi

        log "Analysis saved to $OUTPUT_DIR/analysis.txt"
    else
        warn "Go toolchain not available for analysis"
    fi
}

# Generate memory usage report
memory_usage_report() {
    log "Generating memory usage report..."

    {
        echo "Radarr Go Memory Usage Report"
        echo "Generated: $(date)"
        echo "======================================"
        echo

        # System memory info
        echo "System Memory:"
        free -h
        echo

        # Process memory info
        echo "Process Memory (if running):"
        if pgrep -f radarr >/dev/null; then
            ps -p $(pgrep -f radarr) -o pid,ppid,cmd,%mem,vsz,rss,etime
        else
            echo "Radarr process not found"
        fi
        echo

        # Container memory (if using Docker)
        if command -v docker >/dev/null 2>&1 && docker ps | grep -q radarr; then
            echo "Container Memory Usage:"
            docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}" $(docker ps --filter "name=radarr" -q)
            echo
        fi

        # Go runtime stats (if available)
        if curl -s -f "$RADARR_URL/debug/vars" >/dev/null 2>&1; then
            echo "Go Runtime Stats:"
            curl -s "$RADARR_URL/debug/vars" | jq -r '
                "Goroutines: " + (.cmdline.Goroutines | tostring) + "\n" +
                "Memory Allocated: " + (.memstats.Alloc | tostring) + " bytes\n" +
                "Total Allocations: " + (.memstats.TotalAlloc | tostring) + " bytes\n" +
                "System Memory: " + (.memstats.Sys | tostring) + " bytes\n" +
                "GC Runs: " + (.memstats.NumGC | tostring) + "\n" +
                "Next GC: " + (.memstats.NextGC | tostring) + " bytes"
            ' 2>/dev/null || echo "Runtime stats not available"
        fi
    } > "$OUTPUT_DIR/memory_report.txt"

    log "Memory report saved to $OUTPUT_DIR/memory_report.txt"
}

# Main execution
case "${1:-all}" in
    "memory") collect_memory_profile ;;
    "cpu") collect_cpu_profile ;;
    "goroutine") collect_goroutine_profile ;;
    "analyze") analyze_profiles ;;
    "report") memory_usage_report ;;
    "all")
        collect_memory_profile
        collect_cpu_profile
        collect_goroutine_profile
        memory_usage_report
        analyze_profiles
        ;;
    *)
        echo "Usage: $0 {memory|cpu|goroutine|analyze|report|all}"
        exit 1
        ;;
esac
```

## Concurrent Processing Optimization

### Worker Pool Configuration

```go
// Example: Optimized worker pool implementation
package worker

import (
    "context"
    "sync"
    "time"
)

type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan Result
    wg         sync.WaitGroup
    ctx        context.Context
    cancel     context.CancelFunc
}

type Job interface {
    Execute() Result
}

type Result struct {
    ID    string
    Data  interface{}
    Error error
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
    ctx, cancel := context.WithCancel(context.Background())

    return &WorkerPool{
        workers:    workers,
        jobQueue:   make(chan Job, queueSize),
        resultChan: make(chan Result, queueSize),
        ctx:        ctx,
        cancel:     cancel,
    }
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        go wp.worker(i)
    }
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()

    for {
        select {
        case job := <-wp.jobQueue:
            result := job.Execute()
            select {
            case wp.resultChan <- result:
            case <-wp.ctx.Done():
                return
            }
        case <-wp.ctx.Done():
            return
        }
    }
}

func (wp *WorkerPool) Submit(job Job) error {
    select {
    case wp.jobQueue <- job:
        return nil
    case <-wp.ctx.Done():
        return wp.ctx.Err()
    default:
        return errors.New("job queue is full")
    }
}

func (wp *WorkerPool) Results() <-chan Result {
    return wp.resultChan
}

func (wp *WorkerPool) Stop() {
    wp.cancel()
    wp.wg.Wait()
    close(wp.jobQueue)
    close(wp.resultChan)
}
```

### Download Concurrency Tuning

```yaml
# config.yaml - Download optimization
downloads:
  # Concurrency settings
  max_concurrent_downloads: 25        # Total simultaneous downloads
  max_downloads_per_indexer: 5        # Per-indexer limit
  max_downloads_per_client: 10        # Per-download-client limit

  # Queue settings
  queue_size: 1000                    # Maximum queue size
  queue_processing_interval: "30s"    # Queue check interval

  # Retry settings
  max_retry_attempts: 3               # Maximum retries per download
  retry_delay: "30s"                  # Base retry delay
  exponential_backoff: true           # Use exponential backoff
  max_retry_delay: "5m"              # Maximum retry delay

  # Timeout settings
  connect_timeout: "30s"              # Connection timeout
  response_timeout: "5m"              # Response timeout
  total_timeout: "30m"                # Total download timeout

  # Bandwidth management
  rate_limit: "50MB/s"               # Overall rate limit (0 = unlimited)
  burst_size: "10MB"                 # Burst allowance

  # File handling
  temp_directory: "/tmp/radarr"       # Temporary download directory
  cleanup_failed_downloads: true     # Auto-cleanup failed downloads
  verify_downloads: true              # Verify download integrity
```

### Indexer Optimization

```yaml
# config.yaml - Indexer performance settings
indexers:
  # Search settings
  search_timeout: "30s"               # Per-indexer search timeout
  parallel_searches: true             # Enable parallel indexer searches
  max_concurrent_searches: 10         # Maximum concurrent searches

  # Rate limiting
  global_rate_limit: "100/min"        # Global indexer rate limit
  per_indexer_rate_limit: "30/min"    # Per-indexer rate limit

  # Retry settings
  max_search_retries: 2               # Search retry attempts
  retry_delay: "5s"                   # Retry delay

  # Caching
  enable_search_caching: true         # Enable search result caching
  search_cache_ttl: "5m"             # Search cache TTL

  # Health monitoring
  health_check_interval: "10m"        # Indexer health check interval
  failure_threshold: 5                # Failures before marking unhealthy
  recovery_threshold: 3               # Successes needed for recovery
```

## Storage Performance

### File System Optimization

```bash
#!/bin/bash
# scripts/optimize-storage.sh
# Storage performance optimization

set -euo pipefail

log() { echo "[$(date +'%H:%M:%S')] $1"; }

# Check file system performance
check_fs_performance() {
    log "Checking file system performance..."

    local test_dir="${1:-/movies}"
    local test_file="$test_dir/.radarr_perf_test"

    if [ ! -d "$test_dir" ]; then
        log "Directory $test_dir does not exist"
        return 1
    fi

    # Write performance test
    log "Testing write performance..."
    local write_speed=$(dd if=/dev/zero of="$test_file" bs=1M count=100 2>&1 | \
        grep -o '[0-9.]\+ MB/s' | tail -1)
    log "Write speed: $write_speed"

    # Read performance test
    log "Testing read performance..."
    local read_speed=$(dd if="$test_file" of=/dev/null bs=1M 2>&1 | \
        grep -o '[0-9.]\+ MB/s' | tail -1)
    log "Read speed: $read_speed"

    # Clean up
    rm -f "$test_file"

    # Check file system type and mount options
    log "File system info:"
    df -T "$test_dir"
    mount | grep "$(df "$test_dir" | tail -1 | awk '{print $1}')"
}

# Optimize mount options (requires root)
suggest_mount_optimizations() {
    log "Mount optimization suggestions:"

    cat << 'EOF'
For optimal performance, consider these mount options:

EXT4:
- noatime: Disable access time updates
- data=writeback: Faster but less safe
- commit=60: Increase commit interval

XFS:
- noatime: Disable access time updates
- largeio: Optimize for large I/O
- swalloc: Enable delayed allocation

BTRFS:
- noatime: Disable access time updates
- compress=lz4: Enable compression
- space_cache=v2: Faster free space caching

Example fstab entry:
/dev/sdb1 /movies ext4 defaults,noatime,data=writeback 0 2
EOF
}

# Check for storage issues
check_storage_health() {
    log "Checking storage health..."

    # Disk usage
    echo "=== Disk Usage ==="
    df -h
    echo

    # Inode usage
    echo "=== Inode Usage ==="
    df -i
    echo

    # Check for I/O errors
    echo "=== Recent I/O Errors ==="
    dmesg | grep -i "i/o error\|ata.*error" | tail -10
    echo

    # Check SMART status if available
    if command -v smartctl >/dev/null 2>&1; then
        echo "=== SMART Status ==="
        for disk in /dev/sd?; do
            if [ -e "$disk" ]; then
                echo "Disk: $disk"
                smartctl -H "$disk" 2>/dev/null || true
            fi
        done
    fi
}

# Optimize directory structure
optimize_directory_structure() {
    local movies_dir="${1:-/movies}"

    log "Optimizing directory structure for $movies_dir"

    # Check current structure
    local total_dirs=$(find "$movies_dir" -type d | wc -l)
    local total_files=$(find "$movies_dir" -type f | wc -l)

    log "Current structure: $total_dirs directories, $total_files files"

    # Check for directories with too many files
    echo "=== Directories with most files ==="
    find "$movies_dir" -type d -exec sh -c 'echo "$(find "$1" -maxdepth 1 -type f | wc -l) $1"' _ {} \; | \
        sort -nr | head -20

    echo
    echo "=== Directories with most subdirectories ==="
    find "$movies_dir" -type d -exec sh -c 'echo "$(find "$1" -maxdepth 1 -type d | wc -l) $1"' _ {} \; | \
        sort -nr | head -20

    # Suggestions
    cat << 'EOF'

Directory Structure Recommendations:
1. Keep files per directory under 10,000 for optimal performance
2. Use year-based organization: /movies/2023/Movie Name (2023)/
3. Avoid deeply nested structures (>6 levels)
4. Consider splitting large collections into subdirectories
EOF
}

# Main execution
case "${1:-check}" in
    "check") check_fs_performance "${2:-/movies}" ;;
    "mount") suggest_mount_optimizations ;;
    "health") check_storage_health ;;
    "structure") optimize_directory_structure "${2:-/movies}" ;;
    "all")
        check_fs_performance "${2:-/movies}"
        check_storage_health
        optimize_directory_structure "${2:-/movies}"
        suggest_mount_optimizations
        ;;
    *)
        echo "Usage: $0 {check|mount|health|structure|all} [directory]"
        exit 1
        ;;
esac
```

### Storage Configuration Optimization

```yaml
# config.yaml - Storage optimization
storage:
  # File operations
  parallel_file_operations: 10        # Concurrent file operations
  file_buffer_size: "64KB"           # File I/O buffer size
  use_sendfile: true                  # Use sendfile() for efficient copying

  # Directory scanning
  scan_recursively: true              # Recursive directory scanning
  follow_symlinks: false              # Don't follow symbolic links
  scan_concurrency: 5                 # Concurrent directory scans

  # File monitoring
  enable_file_watcher: true           # Enable file system monitoring
  watcher_buffer_size: 1000          # File event buffer size

  # Temporary files
  temp_directory: "/tmp/radarr"       # Temporary file location
  cleanup_temp_files: true           # Auto-cleanup temp files
  temp_file_retention: "1h"          # Temp file retention period

  # Performance settings
  disable_atime: true                 # Don't update access times
  use_direct_io: false               # Use direct I/O (bypass cache)
  prefetch_metadata: true            # Prefetch file metadata
```

## Network Optimization

### HTTP Client Tuning

```go
// Example: Optimized HTTP client configuration
package http

import (
    "crypto/tls"
    "net/http"
    "time"
)

func NewOptimizedClient() *http.Client {
    transport := &http.Transport{
        // Connection pooling
        MaxIdleConns:          100,              // Max idle connections total
        MaxIdleConnsPerHost:   20,               // Max idle connections per host
        MaxConnsPerHost:       50,               // Max connections per host
        IdleConnTimeout:       90 * time.Second, // Idle connection timeout

        // Connection timeouts
        DialTimeout:           10 * time.Second, // Connection timeout
        TLSHandshakeTimeout:   10 * time.Second, // TLS handshake timeout
        ResponseHeaderTimeout: 30 * time.Second, // Response header timeout

        // Keep-alive settings
        DisableKeepAlives:     false,            // Enable keep-alive

        // Compression
        DisableCompression:    false,            // Enable compression

        // TLS optimization
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: false,           // Verify certificates
            MinVersion:         tls.VersionTLS12, // Minimum TLS version
        },

        // HTTP/2 settings
        ForceAttemptHTTP2:     true,             // Attempt HTTP/2

        // Buffer sizes
        WriteBufferSize:       32 * 1024,        // 32KB write buffer
        ReadBufferSize:        32 * 1024,        // 32KB read buffer
    }

    return &http.Client{
        Transport: transport,
        Timeout:   5 * time.Minute, // Overall request timeout
    }
}
```

### API Server Optimization

```yaml
# config.yaml - API server performance
server:
  # Connection settings
  host: "0.0.0.0"
  port: 7878

  # Performance settings
  read_timeout: "30s"                 # Request read timeout
  write_timeout: "30s"                # Response write timeout
  idle_timeout: "120s"                # Idle connection timeout
  max_header_size: "1MB"              # Maximum header size

  # Concurrency settings
  max_concurrent_requests: 1000       # Maximum concurrent requests
  request_queue_size: 5000           # Request queue size

  # Keep-alive settings
  enable_keep_alive: true            # Enable HTTP keep-alive
  keep_alive_timeout: "90s"          # Keep-alive timeout

  # Compression settings
  enable_gzip: true                  # Enable gzip compression
  gzip_level: 6                      # Gzip compression level (1-9)
  min_compress_size: "1KB"           # Minimum size for compression

  # Security headers
  enable_security_headers: true      # Enable security headers
  enable_cors: true                  # Enable CORS

  # Rate limiting
  enable_rate_limiting: true         # Enable rate limiting
  rate_limit: "100/min"              # Requests per minute per IP
  burst_size: 20                     # Burst allowance
```

## API Performance Tuning

### Response Caching

```go
// Example: API response caching implementation
package cache

import (
    "crypto/md5"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "sync"
    "time"
)

type CacheItem struct {
    Data      interface{}
    ExpiresAt time.Time
}

type APICache struct {
    items map[string]CacheItem
    mutex sync.RWMutex
    ttl   time.Duration
}

func NewAPICache(ttl time.Duration) *APICache {
    cache := &APICache{
        items: make(map[string]CacheItem),
        ttl:   ttl,
    }

    // Start cleanup goroutine
    go cache.cleanup()

    return cache
}

func (c *APICache) generateKey(method, path string, params interface{}) string {
    h := md5.New()
    h.Write([]byte(fmt.Sprintf("%s:%s", method, path)))

    if params != nil {
        if paramBytes, err := json.Marshal(params); err == nil {
            h.Write(paramBytes)
        }
    }

    return hex.EncodeToString(h.Sum(nil))
}

func (c *APICache) Get(method, path string, params interface{}) (interface{}, bool) {
    key := c.generateKey(method, path, params)

    c.mutex.RLock()
    defer c.mutex.RUnlock()

    item, exists := c.items[key]
    if !exists || time.Now().After(item.ExpiresAt) {
        return nil, false
    }

    return item.Data, true
}

func (c *APICache) Set(method, path string, params interface{}, data interface{}) {
    key := c.generateKey(method, path, params)

    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.items[key] = CacheItem{
        Data:      data,
        ExpiresAt: time.Now().Add(c.ttl),
    }
}

func (c *APICache) cleanup() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        c.mutex.Lock()
        now := time.Now()
        for key, item := range c.items {
            if now.After(item.ExpiresAt) {
                delete(c.items, key)
            }
        }
        c.mutex.Unlock()
    }
}
```

### Database Query Optimization

```go
// Example: Query optimization techniques
package database

import (
    "database/sql"
    "fmt"
    "strings"
)

// Optimized movie queries with proper indexing
type MovieQueries struct {
    db *sql.DB

    // Prepared statements for common queries
    getMovieByID          *sql.Stmt
    getMoviesByStatus     *sql.Stmt
    getWantedMovies       *sql.Stmt
    updateMovieStatus     *sql.Stmt
    bulkInsertMovies      *sql.Stmt
}

func NewMovieQueries(db *sql.DB) (*MovieQueries, error) {
    mq := &MovieQueries{db: db}

    var err error

    // Prepare commonly used statements
    mq.getMovieByID, err = db.Prepare(`
        SELECT id, title, year, tmdb_id, status, monitored, quality_profile_id
        FROM movies
        WHERE id = $1
    `)
    if err != nil {
        return nil, err
    }

    mq.getMoviesByStatus, err = db.Prepare(`
        SELECT id, title, year, tmdb_id, status, monitored, quality_profile_id
        FROM movies
        WHERE status = $1 AND monitored = true
        ORDER BY added DESC
        LIMIT $2 OFFSET $3
    `)
    if err != nil {
        return nil, err
    }

    // Optimized query for wanted movies using partial index
    mq.getWantedMovies, err = db.Prepare(`
        SELECT m.id, m.title, m.year, m.tmdb_id, m.quality_profile_id
        FROM movies m
        WHERE m.monitored = true
          AND m.status = 'wanted'
          AND NOT EXISTS (
              SELECT 1 FROM movie_files mf
              WHERE mf.movie_id = m.id
                AND mf.quality_profile_id = m.quality_profile_id
          )
        ORDER BY m.added DESC
        LIMIT $1 OFFSET $2
    `)
    if err != nil {
        return nil, err
    }

    return mq, nil
}

// Bulk operations for better performance
func (mq *MovieQueries) BulkInsertMovies(movies []Movie) error {
    if len(movies) == 0 {
        return nil
    }

    // Build bulk insert query
    valueStrings := make([]string, 0, len(movies))
    valueArgs := make([]interface{}, 0, len(movies)*7)

    for i, movie := range movies {
        valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
            i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))
        valueArgs = append(valueArgs, movie.Title, movie.Year, movie.TmdbID,
            movie.Status, movie.Monitored, movie.QualityProfileID, movie.Added)
    }

    query := fmt.Sprintf(`
        INSERT INTO movies (title, year, tmdb_id, status, monitored, quality_profile_id, added)
        VALUES %s
        ON CONFLICT (tmdb_id) DO UPDATE SET
            title = EXCLUDED.title,
            year = EXCLUDED.year,
            status = EXCLUDED.status,
            monitored = EXCLUDED.monitored,
            quality_profile_id = EXCLUDED.quality_profile_id
    `, strings.Join(valueStrings, ","))

    _, err := mq.db.Exec(query, valueArgs...)
    return err
}

// Connection pool monitoring
func (mq *MovieQueries) GetPoolStats() sql.DBStats {
    return mq.db.Stats()
}
```

## Monitoring and Profiling

### Performance Benchmarking Suite

```bash
#!/bin/bash
# scripts/benchmark-suite.sh
# Comprehensive performance benchmarking

set -euo pipefail

RADARR_URL="${RADARR_URL:-http://localhost:7878}"
API_KEY="${RADARR_API_KEY}"
REPORT_DIR="benchmarks/$(date +%Y%m%d_%H%M%S)"
DURATION="${BENCHMARK_DURATION:-60}"

mkdir -p "$REPORT_DIR"

log() { echo "[$(date +'%H:%M:%S')] $1"; }

# API endpoint benchmarks
benchmark_api_endpoints() {
    log "Benchmarking API endpoints..."

    local endpoints=(
        "GET:/ping:0"
        "GET:/api/v3/system/status:1"
        "GET:/api/v3/movie:1"
        "GET:/api/v3/movie?monitored=true:1"
        "GET:/api/v3/queue:1"
        "GET:/api/v3/history:1"
        "GET:/api/v3/indexer:1"
        "GET:/api/v3/downloadclient:1"
        "GET:/api/v3/notification:1"
    )

    {
        echo "endpoint,method,rps,avg_latency_ms,p95_latency_ms,p99_latency_ms,error_rate"

        for endpoint in "${endpoints[@]}"; do
            IFS=':' read -r method path requires_auth <<< "$endpoint"

            log "Testing $method $path"

            local auth_header=""
            if [ "$requires_auth" = "1" ]; then
                auth_header="-H X-Api-Key:$API_KEY"
            fi

            # Run benchmark with wrk if available, otherwise use ab
            if command -v wrk >/dev/null 2>&1; then
                local result=$(wrk -t4 -c20 -d${DURATION}s --timeout 30s --latency \
                    $auth_header "$RADARR_URL$path" 2>/dev/null | \
                    awk '
                    /Requests\/sec:/ { rps = $2 }
                    /Latency.*50.00%/ { avg = $2 }
                    /Latency.*95.00%/ { p95 = $2 }
                    /Latency.*99.00%/ { p99 = $2 }
                    /Non-2xx/ { errors = $4 }
                    END {
                        gsub(/[^0-9.]/, "", avg)
                        gsub(/[^0-9.]/, "", p95)
                        gsub(/[^0-9.]/, "", p99)
                        if (!errors) errors = 0
                        printf "%.2f,%.2f,%.2f,%.2f,%.2f", rps, avg*1000, p95*1000, p99*1000, errors
                    }')
            else
                # Fallback to ab
                local result=$(ab -n 1000 -c 20 -t $DURATION $auth_header \
                    "$RADARR_URL$path" 2>/dev/null | \
                    awk '
                    /Requests per second:/ { rps = $4 }
                    /Time per request:.*mean/ { avg = $4 }
                    /95%/ { p95 = $2 }
                    /99%/ { p99 = $2 }
                    /Failed requests:/ { errors = $3 }
                    END {
                        if (!errors) errors = 0
                        error_rate = errors / 1000 * 100
                        if (!p99) p99 = p95
                        printf "%.2f,%.2f,%.2f,%.2f,%.2f", rps, avg, p95, p99, error_rate
                    }')
            fi

            echo "$path,$method,$result"
        done
    } > "$REPORT_DIR/api_benchmark.csv"

    log "API benchmark results saved to $REPORT_DIR/api_benchmark.csv"
}

# Database performance benchmark
benchmark_database() {
    log "Benchmarking database performance..."

    if [ -z "${POSTGRES_PASSWORD:-}" ]; then
        log "POSTGRES_PASSWORD not set, skipping database benchmark"
        return
    fi

    {
        echo "query_type,execution_time_ms,rows_returned"

        # Test common queries
        local queries=(
            "SELECT COUNT(*) FROM movies:movie_count"
            "SELECT * FROM movies WHERE monitored = true LIMIT 100:monitored_movies"
            "SELECT * FROM movies WHERE status = 'wanted' LIMIT 100:wanted_movies"
            "SELECT m.*, mf.* FROM movies m LEFT JOIN movie_files mf ON m.id = mf.movie_id LIMIT 100:movies_with_files"
            "SELECT * FROM history ORDER BY date DESC LIMIT 100:recent_history"
        )

        for query_def in "${queries[@]}"; do
            IFS=':' read -r query name <<< "$query_def"

            log "Testing query: $name"

            local result=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "${POSTGRES_HOST:-localhost}" \
                -U "${POSTGRES_USER:-radarr}" -d "${POSTGRES_DB:-radarr}" \
                -c "\\timing on" -c "$query" 2>&1 | \
                grep "Time:" | sed 's/Time: \([0-9.]*\) ms/\1/')

            local row_count=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "${POSTGRES_HOST:-localhost}" \
                -U "${POSTGRES_USER:-radarr}" -d "${POSTGRES_DB:-radarr}" \
                -t -c "$query" | wc -l)

            echo "$name,${result:-0},$row_count"
        done
    } > "$REPORT_DIR/database_benchmark.csv"

    log "Database benchmark results saved to $REPORT_DIR/database_benchmark.csv"
}

# System resource monitoring
monitor_resources() {
    log "Monitoring system resources for ${DURATION} seconds..."

    {
        echo "timestamp,cpu_percent,memory_mb,disk_io_read_mb,disk_io_write_mb,network_rx_mb,network_tx_mb"

        for i in $(seq 1 $DURATION); do
            local timestamp=$(date +%s)

            # CPU usage
            local cpu=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | sed 's/%us,//')

            # Memory usage (in MB)
            local memory=$(free -m | awk '/^Mem:/ {print $3}')

            # Disk I/O (approximate)
            local disk_stats=$(iostat -d 1 1 | tail -n +4 | head -1)
            local disk_read=$(echo "$disk_stats" | awk '{print $3}')
            local disk_write=$(echo "$disk_stats" | awk '{print $4}')

            # Network I/O (approximate)
            local network_stats=$(cat /proc/net/dev | grep eth0 || echo "0 0 0 0 0 0 0 0 0")
            local network_rx=$(echo "$network_stats" | awk '{print $2/1024/1024}')
            local network_tx=$(echo "$network_stats" | awk '{print $10/1024/1024}')

            echo "$timestamp,${cpu:-0},${memory:-0},${disk_read:-0},${disk_write:-0},$network_rx,$network_tx"
            sleep 1
        done
    } > "$REPORT_DIR/resource_monitoring.csv" &

    local monitor_pid=$!

    # Wait for monitoring to complete
    wait $monitor_pid

    log "Resource monitoring saved to $REPORT_DIR/resource_monitoring.csv"
}

# Generate performance report
generate_report() {
    log "Generating performance report..."

    {
        echo "Radarr Go Performance Benchmark Report"
        echo "Generated: $(date)"
        echo "Duration: ${DURATION} seconds"
        echo "======================================="
        echo

        # System information
        echo "System Information:"
        echo "OS: $(uname -a)"
        echo "CPU: $(nproc) cores"
        echo "Memory: $(free -h | awk '/^Mem:/ {print $2}')"
        echo "Go Version: $(go version 2>/dev/null || echo 'Not available')"
        echo

        # API performance summary
        if [ -f "$REPORT_DIR/api_benchmark.csv" ]; then
            echo "API Performance Summary:"
            echo "------------------------"
            awk -F, 'NR>1 {
                total_rps += $3
                total_latency += $4
                max_latency = ($4 > max_latency) ? $4 : max_latency
                count++
            } END {
                if (count > 0) {
                    printf "Average RPS: %.2f\n", total_rps/count
                    printf "Average Latency: %.2f ms\n", total_latency/count
                    printf "Max Latency: %.2f ms\n", max_latency
                }
            }' "$REPORT_DIR/api_benchmark.csv"
            echo
        fi

        # Database performance summary
        if [ -f "$REPORT_DIR/database_benchmark.csv" ]; then
            echo "Database Performance Summary:"
            echo "-----------------------------"
            awk -F, 'NR>1 {
                total_time += $2
                count++
                if ($2 > max_time) { max_time = $2; slowest = $1 }
            } END {
                if (count > 0) {
                    printf "Average Query Time: %.2f ms\n", total_time/count
                    printf "Slowest Query: %s (%.2f ms)\n", slowest, max_time
                }
            }' "$REPORT_DIR/database_benchmark.csv"
            echo
        fi

        # Resource usage summary
        if [ -f "$REPORT_DIR/resource_monitoring.csv" ]; then
            echo "Resource Usage Summary:"
            echo "----------------------"
            awk -F, 'NR>1 {
                cpu_sum += $2
                mem_sum += $3
                count++
                if ($2 > max_cpu) max_cpu = $2
                if ($3 > max_mem) max_mem = $3
            } END {
                if (count > 0) {
                    printf "Average CPU: %.2f%%\n", cpu_sum/count
                    printf "Average Memory: %.2f MB\n", mem_sum/count
                    printf "Peak CPU: %.2f%%\n", max_cpu
                    printf "Peak Memory: %.2f MB\n", max_mem
                }
            }' "$REPORT_DIR/resource_monitoring.csv"
        fi

        echo
        echo "Detailed results are available in CSV format:"
        echo "- API Benchmarks: $REPORT_DIR/api_benchmark.csv"
        echo "- Database Benchmarks: $REPORT_DIR/database_benchmark.csv"
        echo "- Resource Monitoring: $REPORT_DIR/resource_monitoring.csv"

    } > "$REPORT_DIR/performance_report.txt"

    log "Performance report saved to $REPORT_DIR/performance_report.txt"
}

# Performance optimization recommendations
generate_recommendations() {
    log "Generating optimization recommendations..."

    {
        echo "Performance Optimization Recommendations"
        echo "========================================"
        echo

        # Analyze API performance
        if [ -f "$REPORT_DIR/api_benchmark.csv" ]; then
            echo "API Optimization:"
            echo "----------------"

            local slow_endpoints=$(awk -F, 'NR>1 && $4 > 500 {print $1 ": " $4 "ms"}' "$REPORT_DIR/api_benchmark.csv")
            if [ -n "$slow_endpoints" ]; then
                echo "Slow endpoints (>500ms):"
                echo "$slow_endpoints"
                echo "Consider: Response caching, database query optimization, connection pooling"
                echo
            fi

            local low_throughput=$(awk -F, 'NR>1 && $3 < 10 {print $1 ": " $3 " RPS"}' "$REPORT_DIR/api_benchmark.csv")
            if [ -n "$low_throughput" ]; then
                echo "Low throughput endpoints (<10 RPS):"
                echo "$low_throughput"
                echo "Consider: Goroutine optimization, I/O improvements, resource allocation"
                echo
            fi
        fi

        # Analyze database performance
        if [ -f "$REPORT_DIR/database_benchmark.csv" ]; then
            echo "Database Optimization:"
            echo "---------------------"

            local slow_queries=$(awk -F, 'NR>1 && $2 > 1000 {print $1 ": " $2 "ms"}' "$REPORT_DIR/database_benchmark.csv")
            if [ -n "$slow_queries" ]; then
                echo "Slow queries (>1000ms):"
                echo "$slow_queries"
                echo "Consider: Index optimization, query rewriting, connection pool tuning"
                echo
            fi
        fi

        # Analyze resource usage
        if [ -f "$REPORT_DIR/resource_monitoring.csv" ]; then
            echo "Resource Optimization:"
            echo "---------------------"

            local high_cpu=$(awk -F, 'NR>1 && $2 > 80 {count++} END {print count+0}' "$REPORT_DIR/resource_monitoring.csv")
            if [ "$high_cpu" -gt 0 ]; then
                echo "High CPU usage detected ($high_cpu samples >80%)"
                echo "Consider: GOMAXPROCS tuning, algorithm optimization, concurrent processing limits"
                echo
            fi

            local high_memory=$(awk -F, 'NR>1 && $3 > 1000 {count++} END {print count+0}' "$REPORT_DIR/resource_monitoring.csv")
            if [ "$high_memory" -gt 0 ]; then
                echo "High memory usage detected ($high_memory samples >1GB)"
                echo "Consider: GOGC tuning, memory pooling, cache size limits"
                echo
            fi
        fi

        echo "General Recommendations:"
        echo "-----------------------"
        echo "1. Monitor metrics continuously with Prometheus/Grafana"
        echo "2. Set up alerting for performance degradation"
        echo "3. Regular performance regression testing"
        echo "4. Profile memory and CPU usage under load"
        echo "5. Optimize database schema and queries"
        echo "6. Implement proper caching strategies"
        echo "7. Use connection pooling for external services"
        echo "8. Monitor and tune garbage collection"

    } > "$REPORT_DIR/optimization_recommendations.txt"

    log "Optimization recommendations saved to $REPORT_DIR/optimization_recommendations.txt"
}

# Main execution
case "${1:-all}" in
    "api") benchmark_api_endpoints ;;
    "database") benchmark_database ;;
    "resources") monitor_resources ;;
    "report") generate_report ;;
    "recommendations") generate_recommendations ;;
    "all")
        log "Starting comprehensive performance benchmark..."

        # Run benchmarks in parallel where possible
        benchmark_api_endpoints &
        local api_pid=$!

        benchmark_database &
        local db_pid=$!

        monitor_resources &
        local monitor_pid=$!

        # Wait for all benchmarks to complete
        wait $api_pid $db_pid $monitor_pid

        # Generate reports
        generate_report
        generate_recommendations

        log "Benchmark suite completed. Results in $REPORT_DIR"
        ;;
    *)
        echo "Usage: $0 {api|database|resources|report|recommendations|all}"
        exit 1
        ;;
esac
```

This comprehensive performance tuning guide covers all major aspects of optimizing Radarr Go for production use, including database optimization, Go runtime tuning, memory management, concurrent processing, and comprehensive monitoring and profiling tools.
