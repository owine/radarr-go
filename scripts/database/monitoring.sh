#!/bin/bash
# Database Monitoring and Health Check Script for Radarr Go
# Monitors database performance, replication lag, connection health, and alerting
#
# Usage:
#   ./monitoring.sh check
#   ./monitoring.sh replication-status
#   ./monitoring.sh performance-report
#   ./monitoring.sh alerts
#   ./monitoring.sh maintenance

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MONITORING_DIR="${PROJECT_ROOT}/monitoring"
LOG_FILE="${MONITORING_DIR}/monitoring_$(date +%Y%m%d).log"

# Alert thresholds
MAX_CONNECTIONS_PCT=${RADARR_MONITOR_MAX_CONN_PCT:-80}
MAX_REPLICATION_LAG_SECONDS=${RADARR_MONITOR_MAX_REP_LAG:-10}
MIN_FREE_DISK_GB=${RADARR_MONITOR_MIN_DISK_GB:-5}
MAX_QUERY_TIME_MS=${RADARR_MONITOR_MAX_QUERY_MS:-1000}
MAX_LOCK_WAIT_TIME_MS=${RADARR_MONITOR_MAX_LOCK_WAIT:-5000}

# Database configurations (same as backup script)
POSTGRES_HOST="${RADARR_DATABASE_HOST:-localhost}"
POSTGRES_PORT="${RADARR_DATABASE_PORT:-5432}"
POSTGRES_USER="${RADARR_DATABASE_USERNAME:-radarr}"
POSTGRES_DB="${RADARR_DATABASE_NAME:-radarr}"
POSTGRES_PASSWORD="${RADARR_DATABASE_PASSWORD:-password}"

MYSQL_HOST="${RADARR_DATABASE_HOST:-localhost}"
MYSQL_PORT="${RADARR_DATABASE_PORT:-3306}"
MYSQL_USER="${RADARR_DATABASE_USERNAME:-radarr}"
MYSQL_DB="${RADARR_DATABASE_NAME:-radarr}"
MYSQL_PASSWORD="${RADARR_DATABASE_PASSWORD:-password}"

# Notification configuration
ALERT_EMAIL="${RADARR_ALERT_EMAIL:-}"
SLACK_WEBHOOK="${RADARR_SLACK_WEBHOOK:-}"
DISCORD_WEBHOOK="${RADARR_DISCORD_WEBHOOK:-}"

# Setup monitoring directory
setup_monitoring() {
    mkdir -p "$MONITORING_DIR"
    mkdir -p "${MONITORING_DIR}/reports"
    mkdir -p "${MONITORING_DIR}/alerts"
}

# Logging with severity levels
log() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Send alert notification
send_alert() {
    local severity="$1"
    local title="$2"
    local message="$3"
    
    log "ALERT" "$severity: $title - $message"
    
    # Email notification
    if [ -n "$ALERT_EMAIL" ] && command -v mail >/dev/null 2>&1; then
        echo "$message" | mail -s "Radarr DB Alert: $title" "$ALERT_EMAIL"
    fi
    
    # Slack notification
    if [ -n "$SLACK_WEBHOOK" ] && command -v curl >/dev/null 2>&1; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"🚨 Radarr DB Alert: $title\\n$message\"}" \
            "$SLACK_WEBHOOK" >/dev/null 2>&1 || true
    fi
    
    # Discord notification
    if [ -n "$DISCORD_WEBHOOK" ] && command -v curl >/dev/null 2>&1; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"content\":\"🚨 **Radarr DB Alert: $title**\\n$message\"}" \
            "$DISCORD_WEBHOOK" >/dev/null 2>&1 || true
    fi
    
    # Write alert to file
    echo "$(date '+%Y-%m-%d %H:%M:%S'): [$severity] $title - $message" >> "${MONITORING_DIR}/alerts/alerts.log"
}

# Check database connectivity and basic health
check_connectivity() {
    local db_type="$1"
    
    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            if ! psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c 'SELECT 1;' >/dev/null 2>&1; then
                send_alert "CRITICAL" "PostgreSQL Connection Failed" "Cannot connect to PostgreSQL database at $POSTGRES_HOST:$POSTGRES_PORT"
                return 1
            fi
            log "INFO" "PostgreSQL connection healthy"
            ;;
        "mariadb")
            if ! mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" -e 'SELECT 1;' >/dev/null 2>&1; then
                send_alert "CRITICAL" "MariaDB Connection Failed" "Cannot connect to MariaDB database at $MYSQL_HOST:$MYSQL_PORT"
                return 1
            fi
            log "INFO" "MariaDB connection healthy"
            ;;
    esac
    
    return 0
}

# Monitor PostgreSQL specific metrics
monitor_postgresql() {
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    log "INFO" "Monitoring PostgreSQL metrics..."
    
    # Connection count and limits
    local conn_stats=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT 
            current_connections,
            max_connections,
            ROUND((current_connections::float / max_connections::float) * 100, 2) as usage_pct
        FROM (
            SELECT 
                (SELECT COUNT(*) FROM pg_stat_activity WHERE datname = current_database()) as current_connections,
                (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max_connections
        ) stats;
    " | tr -d ' ')
    
    local current_conn=$(echo "$conn_stats" | cut -d'|' -f1)
    local max_conn=$(echo "$conn_stats" | cut -d'|' -f2)
    local usage_pct=$(echo "$conn_stats" | cut -d'|' -f3)
    
    log "INFO" "PostgreSQL connections: $current_conn/$max_conn ($usage_pct%)"
    
    if (( $(echo "$usage_pct > $MAX_CONNECTIONS_PCT" | bc -l) )); then
        send_alert "WARNING" "High Connection Usage" "PostgreSQL using $usage_pct% of available connections ($current_conn/$max_conn)"
    fi
    
    # Long running queries
    local long_queries=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT COUNT(*) FROM pg_stat_activity 
        WHERE state = 'active' 
        AND query_start < NOW() - INTERVAL '$MAX_QUERY_TIME_MS milliseconds'
        AND query NOT LIKE '%pg_stat_activity%';
    " | tr -d ' ')
    
    if [ "$long_queries" -gt 0 ]; then
        send_alert "WARNING" "Long Running Queries" "Found $long_queries queries running longer than ${MAX_QUERY_TIME_MS}ms"
    fi
    
    # Lock monitoring
    local blocked_queries=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT COUNT(*) FROM pg_locks 
        WHERE NOT granted;
    " | tr -d ' ')
    
    if [ "$blocked_queries" -gt 0 ]; then
        send_alert "WARNING" "Database Locks" "Found $blocked_queries blocked queries"
    fi
    
    # Database size monitoring
    local db_size=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT ROUND(pg_database_size(current_database()) / (1024^3)::numeric, 2);
    " | tr -d ' ')
    
    log "INFO" "PostgreSQL database size: ${db_size}GB"
    
    # Replication status (if applicable)
    check_postgresql_replication
}

# Monitor MariaDB/MySQL specific metrics
monitor_mariadb() {
    log "INFO" "Monitoring MariaDB metrics..."
    
    # Connection count and limits
    local conn_current=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SHOW STATUS LIKE 'Threads_connected';
    " -s | awk '{print $2}')
    
    local conn_max=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SHOW VARIABLES LIKE 'max_connections';
    " -s | awk '{print $2}')
    
    local usage_pct=$(echo "scale=2; ($conn_current / $conn_max) * 100" | bc)
    
    log "INFO" "MariaDB connections: $conn_current/$conn_max ($usage_pct%)"
    
    if (( $(echo "$usage_pct > $MAX_CONNECTIONS_PCT" | bc -l) )); then
        send_alert "WARNING" "High Connection Usage" "MariaDB using $usage_pct% of available connections ($conn_current/$conn_max)"
    fi
    
    # Long running queries
    local long_queries=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SELECT COUNT(*) FROM information_schema.PROCESSLIST 
        WHERE COMMAND != 'Sleep' 
        AND TIME > $((MAX_QUERY_TIME_MS / 1000));
    " -s)
    
    if [ "$long_queries" -gt 0 ]; then
        send_alert "WARNING" "Long Running Queries" "Found $long_queries queries running longer than ${MAX_QUERY_TIME_MS}ms"
    fi
    
    # InnoDB lock monitoring
    local innodb_locks=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SELECT COUNT(*) FROM information_schema.INNODB_LOCKS;
    " -s 2>/dev/null || echo "0")
    
    if [ "$innodb_locks" -gt 0 ]; then
        send_alert "WARNING" "Database Locks" "Found $innodb_locks InnoDB locks"
    fi
    
    # Database size monitoring
    local db_size=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024 / 1024, 2) as size_gb
        FROM information_schema.tables
        WHERE table_schema = '$MYSQL_DB';
    " -s)
    
    log "INFO" "MariaDB database size: ${db_size}GB"
    
    # Replication status (if applicable)
    check_mariadb_replication
}

# Check PostgreSQL replication status
check_postgresql_replication() {
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    # Check if this is a master with replicas
    local replica_count=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT COUNT(*) FROM pg_stat_replication;
    " 2>/dev/null | tr -d ' ' || echo "0")
    
    if [ "$replica_count" -gt 0 ]; then
        log "INFO" "PostgreSQL master with $replica_count replica(s)"
        
        # Check replication lag
        local max_lag=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
            SELECT COALESCE(MAX(EXTRACT(EPOCH FROM (now() - backend_start))), 0)
            FROM pg_stat_replication;
        " | tr -d ' ')
        
        if (( $(echo "$max_lag > $MAX_REPLICATION_LAG_SECONDS" | bc -l) )); then
            send_alert "CRITICAL" "High Replication Lag" "PostgreSQL replication lag is ${max_lag}s (threshold: ${MAX_REPLICATION_LAG_SECONDS}s)"
        fi
    fi
    
    # Check if this is a replica
    local is_replica=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT pg_is_in_recovery();
    " 2>/dev/null | tr -d ' ' || echo "f")
    
    if [ "$is_replica" = "t" ]; then
        log "INFO" "PostgreSQL replica detected"
        
        # Check replica lag
        local lag=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
            SELECT CASE 
                WHEN pg_last_wal_receive_lsn() = pg_last_wal_replay_lsn() 
                THEN 0 
                ELSE EXTRACT(EPOCH FROM now() - pg_last_xact_replay_timestamp())::int 
            END;
        " | tr -d ' ')
        
        if [ "$lag" -gt "$MAX_REPLICATION_LAG_SECONDS" ]; then
            send_alert "CRITICAL" "Replica Lag" "PostgreSQL replica lag is ${lag}s (threshold: ${MAX_REPLICATION_LAG_SECONDS}s)"
        fi
    fi
}

# Check MariaDB/MySQL replication status
check_mariadb_replication() {
    # Check if this is a master with replicas
    local replica_count=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SHOW SLAVE HOSTS;
    " 2>/dev/null | wc -l || echo "1")
    
    # Subtract header line
    replica_count=$((replica_count - 1))
    
    if [ "$replica_count" -gt 0 ]; then
        log "INFO" "MariaDB master with $replica_count replica(s)"
    fi
    
    # Check if this is a replica
    local slave_status=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SHOW SLAVE STATUS\G
    " 2>/dev/null | grep -c "Slave_IO_State" || echo "0")
    
    if [ "$slave_status" -gt 0 ]; then
        log "INFO" "MariaDB replica detected"
        
        # Check replication lag
        local lag=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
            SHOW SLAVE STATUS\G
        " | grep "Seconds_Behind_Master:" | awk '{print $2}')
        
        if [ "$lag" != "NULL" ] && [ "$lag" -gt "$MAX_REPLICATION_LAG_SECONDS" ]; then
            send_alert "CRITICAL" "Replica Lag" "MariaDB replica lag is ${lag}s (threshold: ${MAX_REPLICATION_LAG_SECONDS}s)"
        fi
    fi
}

# Generate comprehensive performance report
generate_performance_report() {
    local db_type="$1"
    local report_file="${MONITORING_DIR}/reports/performance_report_${db_type}_$(date +%Y%m%d_%H%M%S).txt"
    
    log "INFO" "Generating $db_type performance report: $report_file"
    
    {
        echo "Radarr Go Database Performance Report"
        echo "======================================"
        echo "Database Type: $db_type"
        echo "Generated: $(date)"
        echo
        
        case "$db_type" in
            "postgresql")
                generate_postgresql_report
                ;;
            "mariadb")
                generate_mariadb_report
                ;;
        esac
        
    } > "$report_file"
    
    log "INFO" "Performance report saved: $report_file"
}

# PostgreSQL performance report
generate_postgresql_report() {
    export PGPASSWORD="$POSTGRES_PASSWORD"
    
    echo "Connection Statistics:"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
        SELECT 
            'Total Connections' as metric,
            COUNT(*) as value
        FROM pg_stat_activity
        UNION ALL
        SELECT 
            'Active Connections' as metric,
            COUNT(*) as value
        FROM pg_stat_activity WHERE state = 'active'
        UNION ALL
        SELECT 
            'Idle Connections' as metric,
            COUNT(*) as value
        FROM pg_stat_activity WHERE state = 'idle';
    "
    
    echo -e "\nDatabase Size:"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
        SELECT 
            schemaname,
            tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size,
            pg_total_relation_size(schemaname||'.'||tablename) as size_bytes
        FROM pg_tables 
        WHERE schemaname = 'public'
        ORDER BY size_bytes DESC;
    "
    
    echo -e "\nTop Query Performance (by total time):"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
        SELECT 
            SUBSTRING(query, 1, 60) as query_sample,
            calls,
            total_time,
            mean_time,
            rows
        FROM pg_stat_statements 
        ORDER BY total_time DESC 
        LIMIT 10;
    " 2>/dev/null || echo "pg_stat_statements extension not available"
    
    echo -e "\nIndex Usage Statistics:"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
        SELECT 
            schemaname,
            tablename,
            indexname,
            idx_scan as index_scans,
            idx_tup_read as tuples_read,
            idx_tup_fetch as tuples_fetched
        FROM pg_stat_user_indexes 
        WHERE idx_scan > 0
        ORDER BY idx_scan DESC
        LIMIT 20;
    "
    
    echo -e "\nTable Statistics:"
    psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
        SELECT 
            schemaname,
            relname as tablename,
            seq_scan,
            seq_tup_read,
            idx_scan,
            idx_tup_fetch,
            n_tup_ins as inserts,
            n_tup_upd as updates,
            n_tup_del as deletes
        FROM pg_stat_user_tables
        ORDER BY seq_scan + idx_scan DESC;
    "
}

# MariaDB performance report  
generate_mariadb_report() {
    echo "Connection Statistics:"
    mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SELECT 
            'Total Connections' as metric,
            VARIABLE_VALUE as value
        FROM information_schema.GLOBAL_STATUS 
        WHERE VARIABLE_NAME = 'Threads_connected'
        UNION ALL
        SELECT 
            'Max Connections' as metric,
            VARIABLE_VALUE as value
        FROM information_schema.GLOBAL_VARIABLES 
        WHERE VARIABLE_NAME = 'max_connections'
        UNION ALL
        SELECT 
            'Running Threads' as metric,
            VARIABLE_VALUE as value
        FROM information_schema.GLOBAL_STATUS 
        WHERE VARIABLE_NAME = 'Threads_running';
    "
    
    echo -e "\nDatabase Size:"
    mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SELECT 
            table_name,
            ROUND((data_length + index_length) / 1024 / 1024, 2) as size_mb,
            table_rows
        FROM information_schema.tables
        WHERE table_schema = '$MYSQL_DB'
        ORDER BY (data_length + index_length) DESC;
    "
    
    echo -e "\nInnoDB Status:"
    mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SHOW ENGINE INNODB STATUS\G
    " | grep -E "(LATEST DETECTED DEADLOCK|BUFFER POOL|LOG|TRANSACTIONS)"
    
    echo -e "\nSlow Query Log (if enabled):"
    mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "
        SELECT 
            sql_text,
            exec_count,
            avg_timer_wait / 1000000000 as avg_time_ms,
            total_timer_wait / 1000000000 as total_time_ms
        FROM performance_schema.events_statements_summary_by_digest
        WHERE schema_name = '$MYSQL_DB'
        ORDER BY total_timer_wait DESC
        LIMIT 10;
    " 2>/dev/null || echo "Performance schema not fully enabled"
}

# Check disk space
check_disk_space() {
    log "INFO" "Checking disk space..."
    
    # Get disk usage for the data directory
    local data_dir="${RADARR_DATA_DIR:-./data}"
    local free_space_gb=$(df -BG "$data_dir" | awk 'NR==2 {print $4}' | sed 's/G//')
    
    log "INFO" "Free disk space: ${free_space_gb}GB"
    
    if [ "$free_space_gb" -lt "$MIN_FREE_DISK_GB" ]; then
        send_alert "CRITICAL" "Low Disk Space" "Only ${free_space_gb}GB free space remaining (threshold: ${MIN_FREE_DISK_GB}GB)"
    fi
}

# Database maintenance operations
run_maintenance() {
    local db_type="$1"
    
    log "INFO" "Running $db_type maintenance operations..."
    
    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            
            # Analyze tables for better query plans
            log "INFO" "Running ANALYZE on all tables..."
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
                ANALYZE;
            " >/dev/null 2>&1
            
            # Vacuum to reclaim space (non-blocking)
            log "INFO" "Running VACUUM on all tables..."
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
                VACUUM;
            " >/dev/null 2>&1
            
            # Check for bloated indexes
            local bloated_indexes=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
                SELECT COUNT(*) FROM (
                    SELECT 
                        schemaname,
                        tablename,
                        indexname,
                        pg_size_pretty(pg_relation_size(indexrelid)) as size
                    FROM pg_stat_user_indexes 
                    WHERE pg_relation_size(indexrelid) > 50000000  -- 50MB
                    AND idx_scan < 100  -- Used less than 100 times
                ) unused_large_indexes;
            " | tr -d ' ')
            
            if [ "$bloated_indexes" -gt 0 ]; then
                log "WARNING" "Found $bloated_indexes potentially unused large indexes"
            fi
            ;;
            
        "mariadb")
            # Optimize tables
            log "INFO" "Running OPTIMIZE TABLE on all tables..."
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" -e "
                SELECT CONCAT('OPTIMIZE TABLE ', table_name, ';') as stmt
                FROM information_schema.tables
                WHERE table_schema = '$MYSQL_DB'
                AND table_type = 'BASE TABLE';
            " -s | mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" >/dev/null 2>&1
            
            # Update table statistics
            log "INFO" "Running ANALYZE TABLE on all tables..."
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" -e "
                SELECT CONCAT('ANALYZE TABLE ', table_name, ';') as stmt
                FROM information_schema.tables
                WHERE table_schema = '$MYSQL_DB'
                AND table_type = 'BASE TABLE';
            " -s | mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" >/dev/null 2>&1
            ;;
    esac
    
    log "INFO" "Maintenance completed for $db_type"
}

# Main monitoring check
health_check() {
    log "INFO" "Starting database health check..."
    
    setup_monitoring
    check_disk_space
    
    # Check each available database
    for db_type in postgresql mariadb; do
        if check_connectivity "$db_type"; then
            case "$db_type" in
                "postgresql")
                    monitor_postgresql
                    ;;
                "mariadb")
                    monitor_mariadb
                    ;;
            esac
        fi
    done
    
    log "INFO" "Health check completed"
}

# Run performance tests with migration validation
performance_test_migrations() {
    log "INFO" "Testing migration performance on large datasets..."
    
    for db_type in postgresql mariadb; do
        if command -v ${db_type%db} >/dev/null 2>&1; then
            log "INFO" "Testing $db_type migration performance..."
            
            # This would integrate with the backup_restore.sh performance test
            "${SCRIPT_DIR}/backup_restore.sh" performance-test || log "WARNING" "$db_type performance test failed"
        fi
    done
}

# Main function
main() {
    local command="${1:-check}"
    
    setup_monitoring
    
    case "$command" in
        "check"|"health")
            health_check
            ;;
        "replication-status")
            log "INFO" "Checking replication status..."
            for db_type in postgresql mariadb; do
                if check_connectivity "$db_type" 2>/dev/null; then
                    case "$db_type" in
                        "postgresql") check_postgresql_replication ;;
                        "mariadb") check_mariadb_replication ;;
                    esac
                fi
            done
            ;;
        "performance-report")
            for db_type in postgresql mariadb; do
                if check_connectivity "$db_type" 2>/dev/null; then
                    generate_performance_report "$db_type"
                fi
            done
            ;;
        "alerts")
            log "INFO" "Recent alerts:"
            if [ -f "${MONITORING_DIR}/alerts/alerts.log" ]; then
                tail -20 "${MONITORING_DIR}/alerts/alerts.log"
            else
                log "INFO" "No alerts found"
            fi
            ;;
        "maintenance")
            for db_type in postgresql mariadb; do
                if check_connectivity "$db_type" 2>/dev/null; then
                    run_maintenance "$db_type"
                fi
            done
            ;;
        "migration-test")
            performance_test_migrations
            ;;
        *)
            echo "Usage: $0 <command>"
            echo
            echo "Commands:"
            echo "  check, health            Run comprehensive health check"
            echo "  replication-status       Check replication status"
            echo "  performance-report       Generate detailed performance report"
            echo "  alerts                   Show recent alerts"
            echo "  maintenance             Run database maintenance tasks"
            echo "  migration-test          Test migration performance"
            echo
            echo "Environment variables:"
            echo "  RADARR_DATABASE_HOST            Database host (default: localhost)"
            echo "  RADARR_DATABASE_PORT            Database port"
            echo "  RADARR_DATABASE_USERNAME        Database username (default: radarr)"
            echo "  RADARR_DATABASE_NAME            Database name (default: radarr)"
            echo "  RADARR_DATABASE_PASSWORD        Database password (default: password)"
            echo "  RADARR_MONITOR_MAX_CONN_PCT     Max connection percentage (default: 80)"
            echo "  RADARR_MONITOR_MAX_REP_LAG      Max replication lag seconds (default: 10)"
            echo "  RADARR_MONITOR_MIN_DISK_GB      Min free disk space GB (default: 5)"
            echo "  RADARR_MONITOR_MAX_QUERY_MS     Max query time ms (default: 1000)"
            echo "  RADARR_ALERT_EMAIL              Email for alerts"
            echo "  RADARR_SLACK_WEBHOOK            Slack webhook URL"
            echo "  RADARR_DISCORD_WEBHOOK          Discord webhook URL"
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"