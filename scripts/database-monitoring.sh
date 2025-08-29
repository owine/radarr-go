#!/bin/bash

# Radarr-Go Database Monitoring and Alerting Script
# Provides continuous monitoring, alerting, and performance tracking

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
MONITORING_DIR="${PROJECT_ROOT}/monitoring"
LOG_FILE="${MONITORING_DIR}/database-monitoring.log"
ALERTS_FILE="${MONITORING_DIR}/alerts.log"
METRICS_FILE="${MONITORING_DIR}/metrics.json"

# Thresholds (configurable)
CPU_THRESHOLD=80
MEMORY_THRESHOLD=80
DISK_THRESHOLD=85
CONNECTION_THRESHOLD=80  # % of max connections
QUERY_TIME_THRESHOLD=300 # seconds
REPLICATION_LAG_THRESHOLD=10 # seconds

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Source the database operations script for common functions
source "${SCRIPT_DIR}/database-operations.sh"

# Initialize monitoring directory
init_monitoring() {
    mkdir -p "$MONITORING_DIR"
    touch "$LOG_FILE" "$ALERTS_FILE" "$METRICS_FILE"
}

# Log monitoring events
log_monitoring() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') $1" >> "$LOG_FILE"
}

# Send alert
send_alert() {
    local severity="$1"
    local message="$2"
    local alert_message="$(date '+%Y-%m-%d %H:%M:%S') [$severity] $message"

    echo "$alert_message" >> "$ALERTS_FILE"
    log_monitoring "ALERT [$severity]: $message"

    # Color coding for terminal output
    case "$severity" in
        CRITICAL) echo -e "${RED}🚨 $alert_message${NC}" ;;
        WARNING)  echo -e "${YELLOW}⚠️  $alert_message${NC}" ;;
        INFO)     echo -e "${GREEN}ℹ️  $alert_message${NC}" ;;
        *)        echo "$alert_message" ;;
    esac

    # Here you could add additional alerting mechanisms:
    # - Send email
    # - Post to Slack/Discord webhook
    # - Send to monitoring system (Prometheus/Grafana)
    # - Trigger PagerDuty/OpsGenie
}

# Collect database metrics
collect_metrics() {
    local timestamp=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    local metrics="{\"timestamp\":\"$timestamp\""

    case "$DB_TYPE" in
        postgres|postgresql)
            metrics="$metrics,$(collect_postgres_metrics)"
            ;;
        mysql|mariadb)
            metrics="$metrics,$(collect_mysql_metrics)"
            ;;
    esac

    metrics="$metrics}"

    # Append to metrics file
    echo "$metrics" >> "$METRICS_FILE"

    # Keep only last 1000 entries to prevent file growth
    tail -n 1000 "$METRICS_FILE" > "${METRICS_FILE}.tmp" && mv "${METRICS_FILE}.tmp" "$METRICS_FILE"

    echo "$metrics"
}

# Collect PostgreSQL specific metrics
collect_postgres_metrics() {
    export PGPASSWORD="$DB_PASSWORD"

    local metrics=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -A -c "
        WITH connection_stats AS (
            SELECT
                count(*) as total_connections,
                count(*) FILTER (WHERE state = 'active') as active_connections,
                count(*) FILTER (WHERE state = 'idle') as idle_connections
            FROM pg_stat_activity
        ),
        database_stats AS (
            SELECT
                pg_database_size('$DB_NAME') as database_size,
                (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max_connections
            FROM pg_database WHERE datname = '$DB_NAME'
        ),
        query_stats AS (
            SELECT
                count(*) FILTER (WHERE now() - query_start > interval '$QUERY_TIME_THRESHOLD seconds') as long_queries,
                max(EXTRACT(EPOCH FROM (now() - query_start))) as max_query_duration
            FROM pg_stat_activity
            WHERE state = 'active' AND query_start IS NOT NULL
        ),
        lock_stats AS (
            SELECT count(*) as active_locks
            FROM pg_locks
            WHERE NOT granted
        ),
        replication_stats AS (
            SELECT
                coalesce(max(EXTRACT(EPOCH FROM (now() - backend_start))), 0) as replication_lag
            FROM pg_stat_replication
        )
        SELECT
            '\"connections\": {' ||
            '\"total\": ' || cs.total_connections || ', ' ||
            '\"active\": ' || cs.active_connections || ', ' ||
            '\"idle\": ' || cs.idle_connections || ', ' ||
            '\"max\": ' || ds.max_connections || ', ' ||
            '\"usage_percent\": ' || ROUND((cs.total_connections::float / ds.max_connections::float) * 100, 2) ||
            '}, ' ||
            '\"database\": {' ||
            '\"size_bytes\": ' || ds.database_size ||
            '}, ' ||
            '\"queries\": {' ||
            '\"long_running\": ' || coalesce(qs.long_queries, 0) || ', ' ||
            '\"max_duration\": ' || coalesce(qs.max_query_duration, 0) ||
            '}, ' ||
            '\"locks\": {' ||
            '\"waiting\": ' || ls.active_locks ||
            '}, ' ||
            '\"replication\": {' ||
            '\"lag_seconds\": ' || rs.replication_lag ||
            '}'
        FROM connection_stats cs
        CROSS JOIN database_stats ds
        CROSS JOIN query_stats qs
        CROSS JOIN lock_stats ls
        CROSS JOIN replication_stats rs;
    " 2>/dev/null | head -1)

    unset PGPASSWORD
    echo "$metrics"
}

# Collect MySQL/MariaDB specific metrics
collect_mysql_metrics() {
    local metrics=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -sN -e "
        SELECT CONCAT(
            '\"connections\": {',
            '\"total\": ', (SELECT VARIABLE_VALUE FROM INFORMATION_SCHEMA.SESSION_STATUS WHERE VARIABLE_NAME = 'Threads_connected'), ', ',
            '\"active\": ', (SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE COMMAND != 'Sleep'), ', ',
            '\"max\": ', @@max_connections, ', ',
            '\"usage_percent\": ', ROUND((SELECT VARIABLE_VALUE FROM INFORMATION_SCHEMA.SESSION_STATUS WHERE VARIABLE_NAME = 'Threads_connected') / @@max_connections * 100, 2),
            '}, ',
            '\"database\": {',
            '\"size_bytes\": ', (SELECT SUM(data_length + index_length) FROM information_schema.tables WHERE table_schema = '$DB_NAME'),
            '}, ',
            '\"queries\": {',
            '\"long_running\": ', (SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE TIME > $QUERY_TIME_THRESHOLD AND COMMAND != 'Sleep'), ', ',
            '\"max_duration\": ', COALESCE((SELECT MAX(TIME) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE COMMAND != 'Sleep'), 0),
            '}, ',
            '\"locks\": {',
            '\"waiting\": ', (SELECT COUNT(*) FROM INFORMATION_SCHEMA.PROCESSLIST WHERE STATE LIKE '%Waiting%'),
            '}, ',
            '\"replication\": {',
            '\"lag_seconds\": 0',
            '}'
        ) as metrics;
    " 2>/dev/null | head -1)

    echo "$metrics"
}

# Check database health and generate alerts
check_health_and_alert() {
    log_monitoring "Starting health check"

    # Collect current metrics
    local current_metrics=$(collect_metrics)
    log_monitoring "Collected metrics: $current_metrics"

    # Parse metrics for alerting (requires jq for proper JSON parsing)
    if command -v jq >/dev/null 2>&1; then
        check_connection_threshold "$current_metrics"
        check_query_performance "$current_metrics"
        check_database_size "$current_metrics"
        check_replication_lag "$current_metrics"
        check_locks "$current_metrics"
    else
        # Fallback to basic monitoring without JSON parsing
        log_monitoring "jq not available, using basic monitoring"
        basic_health_check
    fi

    log_monitoring "Health check completed"
}

# Check connection pool usage
check_connection_threshold() {
    local metrics="$1"
    local usage_percent=$(echo "$metrics" | jq -r '.connections.usage_percent // 0')

    if (( $(echo "$usage_percent > $CONNECTION_THRESHOLD" | bc -l) )); then
        send_alert "WARNING" "Connection pool usage high: ${usage_percent}% (threshold: ${CONNECTION_THRESHOLD}%)"
    fi

    local total_connections=$(echo "$metrics" | jq -r '.connections.total // 0')
    local max_connections=$(echo "$metrics" | jq -r '.connections.max // 100')

    if (( total_connections >= max_connections - 5 )); then
        send_alert "CRITICAL" "Connection pool nearly exhausted: $total_connections/$max_connections"
    fi
}

# Check query performance
check_query_performance() {
    local metrics="$1"
    local long_queries=$(echo "$metrics" | jq -r '.queries.long_running // 0')
    local max_duration=$(echo "$metrics" | jq -r '.queries.max_duration // 0')

    if (( long_queries > 0 )); then
        send_alert "WARNING" "Long running queries detected: $long_queries queries (max duration: ${max_duration}s)"
    fi

    if (( $(echo "$max_duration > $(($QUERY_TIME_THRESHOLD * 2))" | bc -l) )); then
        send_alert "CRITICAL" "Very long query detected: ${max_duration}s duration"
    fi
}

# Check database size growth
check_database_size() {
    local metrics="$1"
    local size_bytes=$(echo "$metrics" | jq -r '.database.size_bytes // 0')
    local size_gb=$(echo "scale=2; $size_bytes / 1024 / 1024 / 1024" | bc)

    # Check if size is growing rapidly (comparison with previous measurements)
    if [[ -f "$METRICS_FILE" ]]; then
        local prev_size=$(tail -2 "$METRICS_FILE" | head -1 | jq -r '.database.size_bytes // 0' 2>/dev/null || echo "0")
        if [[ "$prev_size" != "0" ]] && (( size_bytes > prev_size )); then
            local growth_mb=$(echo "scale=2; ($size_bytes - $prev_size) / 1024 / 1024" | bc)
            if (( $(echo "$growth_mb > 100" | bc -l) )); then
                send_alert "INFO" "Database size grew by ${growth_mb}MB (current size: ${size_gb}GB)"
            fi
        fi
    fi
}

# Check replication lag
check_replication_lag() {
    local metrics="$1"
    local lag_seconds=$(echo "$metrics" | jq -r '.replication.lag_seconds // 0')

    if (( $(echo "$lag_seconds > $REPLICATION_LAG_THRESHOLD" | bc -l) )); then
        send_alert "WARNING" "Replication lag detected: ${lag_seconds}s (threshold: ${REPLICATION_LAG_THRESHOLD}s)"
    fi
}

# Check for waiting locks
check_locks() {
    local metrics="$1"
    local waiting_locks=$(echo "$metrics" | jq -r '.locks.waiting // 0')

    if (( waiting_locks > 0 )); then
        send_alert "WARNING" "Database locks detected: $waiting_locks waiting locks"
    fi
}

# Basic health check without JSON parsing
basic_health_check() {
    # Simple connection test
    case "$DB_TYPE" in
        postgres|postgresql)
            export PGPASSWORD="$DB_PASSWORD"
            if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" >/dev/null 2>&1; then
                send_alert "CRITICAL" "Database connection failed"
            fi
            unset PGPASSWORD
            ;;
        mysql|mariadb)
            if ! mysqladmin -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" ping >/dev/null 2>&1; then
                send_alert "CRITICAL" "Database connection failed"
            fi
            ;;
    esac
}

# Monitor system resources
monitor_system_resources() {
    log_monitoring "Monitoring system resources"

    # CPU usage
    if command -v top >/dev/null 2>&1; then
        local cpu_usage=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
        if (( $(echo "$cpu_usage > $CPU_THRESHOLD" | bc -l) )); then
            send_alert "WARNING" "High CPU usage: ${cpu_usage}%"
        fi
    fi

    # Memory usage
    if command -v free >/dev/null 2>&1; then
        local mem_usage=$(free | grep Mem | awk '{printf "%.1f", $3/$2 * 100.0}')
        if (( $(echo "$mem_usage > $MEMORY_THRESHOLD" | bc -l) )); then
            send_alert "WARNING" "High memory usage: ${mem_usage}%"
        fi
    fi

    # Disk usage for data directory
    if [[ -d "$PROJECT_ROOT/data" ]]; then
        local disk_usage=$(df "$PROJECT_ROOT/data" | tail -1 | awk '{print $5}' | sed 's/%//')
        if (( disk_usage > DISK_THRESHOLD )); then
            send_alert "WARNING" "High disk usage: ${disk_usage}%"
        fi
    fi
}

# Generate monitoring dashboard
generate_dashboard() {
    local dashboard_file="${MONITORING_DIR}/dashboard.html"

    log_monitoring "Generating monitoring dashboard"

    cat > "$dashboard_file" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <title>Radarr-Go Database Monitoring</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .metrics { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .metric-title { font-size: 18px; font-weight: bold; color: #2c3e50; margin-bottom: 10px; }
        .metric-value { font-size: 24px; font-weight: bold; }
        .status-ok { color: #27ae60; }
        .status-warning { color: #f39c12; }
        .status-critical { color: #e74c3c; }
        .logs { margin-top: 20px; background: white; padding: 20px; border-radius: 5px; }
        .log-entry { margin: 5px 0; padding: 5px; font-family: monospace; background: #f8f9fa; border-radius: 3px; }
        .refresh-info { text-align: center; color: #7f8c8d; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔍 Radarr-Go Database Monitoring</h1>
            <p>Real-time database health and performance monitoring</p>
        </div>

        <div class="metrics" id="metrics">
            <!-- Metrics will be populated by JavaScript -->
        </div>

        <div class="logs">
            <h2>Recent Alerts</h2>
            <div id="alerts">
                <!-- Alerts will be populated by JavaScript -->
            </div>
        </div>

        <div class="refresh-info">
            <p>Last updated: <span id="last-updated"></span></p>
            <p>Auto-refresh every 30 seconds</p>
        </div>
    </div>

    <script>
        function updateDashboard() {
            fetch('metrics.json')
                .then(response => response.text())
                .then(data => {
                    const lines = data.trim().split('\n');
                    const latestMetrics = JSON.parse(lines[lines.length - 1]);
                    displayMetrics(latestMetrics);
                })
                .catch(error => console.error('Error loading metrics:', error));

            fetch('alerts.log')
                .then(response => response.text())
                .then(data => {
                    const lines = data.trim().split('\n').slice(-10); // Last 10 alerts
                    displayAlerts(lines);
                })
                .catch(error => console.error('Error loading alerts:', error));

            document.getElementById('last-updated').textContent = new Date().toLocaleString();
        }

        function displayMetrics(metrics) {
            const container = document.getElementById('metrics');
            const connections = metrics.connections || {};
            const database = metrics.database || {};
            const queries = metrics.queries || {};

            container.innerHTML = `
                <div class="metric-card">
                    <div class="metric-title">Database Connections</div>
                    <div class="metric-value ${getStatusClass(connections.usage_percent, 50, 80)}">
                        ${connections.total || 0} / ${connections.max || 0}
                    </div>
                    <p>${connections.usage_percent || 0}% usage</p>
                    <p>Active: ${connections.active || 0}, Idle: ${connections.idle || 0}</p>
                </div>

                <div class="metric-card">
                    <div class="metric-title">Database Size</div>
                    <div class="metric-value">
                        ${formatBytes(database.size_bytes || 0)}
                    </div>
                </div>

                <div class="metric-card">
                    <div class="metric-title">Query Performance</div>
                    <div class="metric-value ${queries.long_running > 0 ? 'status-warning' : 'status-ok'}">
                        ${queries.long_running || 0} long queries
                    </div>
                    <p>Max duration: ${queries.max_duration || 0}s</p>
                </div>

                <div class="metric-card">
                    <div class="metric-title">System Status</div>
                    <div class="metric-value status-ok">
                        ✓ Operational
                    </div>
                    <p>Last check: ${metrics.timestamp}</p>
                </div>
            `;
        }

        function displayAlerts(alerts) {
            const container = document.getElementById('alerts');
            container.innerHTML = alerts.map(alert =>
                `<div class="log-entry">${alert}</div>`
            ).join('');
        }

        function getStatusClass(value, warning, critical) {
            if (value >= critical) return 'status-critical';
            if (value >= warning) return 'status-warning';
            return 'status-ok';
        }

        function formatBytes(bytes) {
            const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
            if (bytes === 0) return '0 Bytes';
            const i = parseInt(Math.floor(Math.log(bytes) / Math.log(1024)));
            return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
        }

        // Initial load and auto-refresh
        updateDashboard();
        setInterval(updateDashboard, 30000);
    </script>
</body>
</html>
EOF

    log_monitoring "Dashboard generated: $dashboard_file"
    echo "Dashboard available at: $dashboard_file"
}

# Continuous monitoring mode
run_continuous_monitoring() {
    local interval="${1:-60}" # Default 60 seconds

    log_monitoring "Starting continuous monitoring (interval: ${interval}s)"
    send_alert "INFO" "Database monitoring started"

    while true; do
        check_health_and_alert
        monitor_system_resources

        # Generate dashboard every 10 iterations
        if (( $(date +%s) % 600 == 0 )); then
            generate_dashboard
        fi

        sleep "$interval"
    done
}

# Performance analysis
analyze_performance() {
    log_monitoring "Starting performance analysis"

    if [[ ! -f "$METRICS_FILE" ]] || [[ ! -s "$METRICS_FILE" ]]; then
        echo "No metrics data available for analysis"
        return 1
    fi

    echo "=== Database Performance Analysis ==="
    echo "Generated: $(date)"
    echo ""

    # Analyze connection patterns
    echo "Connection Usage Patterns:"
    if command -v jq >/dev/null 2>&1; then
        cat "$METRICS_FILE" | jq -r '.connections.usage_percent' | tail -n 100 | \
        awk '{
            sum += $1; count++;
            if($1 > max) max = $1;
            if(min == "" || $1 < min) min = $1
        }
        END {
            print "  Average: " sum/count "%"
            print "  Maximum: " max "%"
            print "  Minimum: " min "%"
        }'
    fi
    echo ""

    # Analyze query performance
    echo "Query Performance:"
    if command -v jq >/dev/null 2>&1; then
        cat "$METRICS_FILE" | jq -r '.queries.max_duration' | tail -n 100 | \
        awk '{
            sum += $1; count++;
            if($1 > max) max = $1
        }
        END {
            print "  Average max duration: " sum/count "s"
            print "  Peak duration: " max "s"
        }'
    fi
    echo ""

    # Recent alerts summary
    echo "Recent Alerts (last 24 hours):"
    if [[ -f "$ALERTS_FILE" ]]; then
        grep "$(date -d '24 hours ago' '+%Y-%m-%d' 2>/dev/null || date -v-24H '+%Y-%m-%d')" "$ALERTS_FILE" | \
        awk '{print $3}' | sort | uniq -c | sort -nr | head -5
    fi
    echo ""
}

# Usage information
usage() {
    cat << EOF
Radarr-Go Database Monitoring Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    monitor [interval]      Start continuous monitoring (default: 60s interval)
    check                   Run single health check with alerting
    metrics                 Collect and display current metrics
    dashboard               Generate HTML dashboard
    analyze                 Analyze historical performance data
    alerts                  Show recent alerts
    system                  Check system resources

Examples:
    $0 monitor              # Start continuous monitoring
    $0 monitor 30           # Monitor every 30 seconds
    $0 check                # Single health check
    $0 dashboard            # Generate dashboard
    $0 analyze              # Performance analysis

Configuration:
    Set thresholds via environment variables:
    CPU_THRESHOLD=80
    MEMORY_THRESHOLD=80
    DISK_THRESHOLD=85
    CONNECTION_THRESHOLD=80
    QUERY_TIME_THRESHOLD=300

EOF
}

# Main script execution
main() {
    init_monitoring
    parse_db_config

    case "${1:-}" in
        monitor)
            run_continuous_monitoring "${2:-60}"
            ;;
        check)
            check_health_and_alert
            ;;
        metrics)
            collect_metrics
            ;;
        dashboard)
            generate_dashboard
            ;;
        analyze)
            analyze_performance
            ;;
        alerts)
            echo "=== Recent Alerts ==="
            tail -n 20 "$ALERTS_FILE" 2>/dev/null || echo "No alerts found"
            ;;
        system)
            monitor_system_resources
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            echo "Starting single health check..."
            check_health_and_alert
            ;;
    esac
}

# Initialize and run
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
