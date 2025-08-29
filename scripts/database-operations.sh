#!/bin/bash

# Radarr-Go Database Operations and Maintenance Script
# This script provides backup, monitoring, and maintenance utilities for Radarr-Go database

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BACKUP_DIR="${PROJECT_ROOT}/backups/database"
CONFIG_FILE="${PROJECT_ROOT}/config.yaml"
RETENTION_DAYS=30
DATE_FORMAT="%Y%m%d_%H%M%S"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1" >&2
}

log_debug() {
    if [[ "${DEBUG:-}" == "1" ]]; then
        echo -e "${BLUE}[DEBUG]${NC} $(date '+%Y-%m-%d %H:%M:%S') $1"
    fi
}

# Parse database configuration from config.yaml
parse_db_config() {
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Configuration file not found: $CONFIG_FILE"
        exit 1
    fi

    # Extract database configuration using yq or simple grep
    if command -v yq >/dev/null 2>&1; then
        DB_TYPE=$(yq eval '.database.type // "postgres"' "$CONFIG_FILE")
        DB_HOST=$(yq eval '.database.host // "localhost"' "$CONFIG_FILE")
        DB_PORT=$(yq eval '.database.port // (if .database.type == "postgres" then 5432 else 3306 end)' "$CONFIG_FILE")
        DB_NAME=$(yq eval '.database.database // "radarr"' "$CONFIG_FILE")
        DB_USER=$(yq eval '.database.username // "radarr"' "$CONFIG_FILE")
        DB_PASSWORD=$(yq eval '.database.password // ""' "$CONFIG_FILE")
    else
        log_warn "yq not found, using environment variables or defaults"
        DB_TYPE="${RADARR_DATABASE_TYPE:-postgres}"
        DB_HOST="${RADARR_DATABASE_HOST:-localhost}"
        DB_PORT="${RADARR_DATABASE_PORT:-$(if [[ "$DB_TYPE" == "postgres" ]]; then echo 5432; else echo 3306; fi)}"
        DB_NAME="${RADARR_DATABASE_DATABASE:-radarr}"
        DB_USER="${RADARR_DATABASE_USERNAME:-radarr}"
        DB_PASSWORD="${RADARR_DATABASE_PASSWORD:-}"
    fi

    log_debug "Database config: type=$DB_TYPE, host=$DB_HOST, port=$DB_PORT, database=$DB_NAME, user=$DB_USER"
}

# Create backup directory structure
setup_backup_dir() {
    mkdir -p "$BACKUP_DIR"/{daily,weekly,monthly}
    log_info "Backup directories created at $BACKUP_DIR"
}

# PostgreSQL backup functions
backup_postgres() {
    local backup_file="$1"
    local pg_dump_opts=(
        "--host=$DB_HOST"
        "--port=$DB_PORT"
        "--username=$DB_USER"
        "--dbname=$DB_NAME"
        "--no-password"
        "--verbose"
        "--clean"
        "--if-exists"
        "--create"
        "--format=custom"
        "--compress=9"
    )

    # Set PGPASSWORD for authentication
    export PGPASSWORD="$DB_PASSWORD"

    log_info "Starting PostgreSQL backup to $backup_file"

    if pg_dump "${pg_dump_opts[@]}" --file="$backup_file"; then
        log_info "PostgreSQL backup completed successfully"

        # Get backup file size for validation
        local backup_size=$(stat -c%s "$backup_file" 2>/dev/null || stat -f%z "$backup_file" 2>/dev/null || echo "unknown")
        log_info "Backup file size: $backup_size bytes"

        # Verify backup integrity
        if pg_restore --list "$backup_file" >/dev/null 2>&1; then
            log_info "Backup integrity verified"
        else
            log_error "Backup integrity check failed"
            return 1
        fi
    else
        log_error "PostgreSQL backup failed"
        return 1
    fi

    unset PGPASSWORD
}

# MySQL/MariaDB backup functions
backup_mysql() {
    local backup_file="$1"
    local mysqldump_opts=(
        "--host=$DB_HOST"
        "--port=$DB_PORT"
        "--user=$DB_USER"
        "--password=$DB_PASSWORD"
        "--single-transaction"
        "--routines"
        "--triggers"
        "--add-drop-database"
        "--create-options"
        "--disable-keys"
        "--extended-insert"
        "--quick"
        "--lock-tables=false"
        "--compress"
        "--result-file=$backup_file"
        "$DB_NAME"
    )

    log_info "Starting MySQL/MariaDB backup to $backup_file"

    if mysqldump "${mysqldump_opts[@]}"; then
        log_info "MySQL/MariaDB backup completed successfully"

        # Get backup file size for validation
        local backup_size=$(stat -c%s "$backup_file" 2>/dev/null || stat -f%z "$backup_file" 2>/dev/null || echo "unknown")
        log_info "Backup file size: $backup_size bytes"

        # Basic validation - check if file contains expected content
        if grep -q "CREATE DATABASE" "$backup_file" 2>/dev/null; then
            log_info "Backup content validated"
        else
            log_error "Backup validation failed - missing expected content"
            return 1
        fi
    else
        log_error "MySQL/MariaDB backup failed"
        return 1
    fi
}

# Main backup function with retention policy
create_backup() {
    local backup_type="${1:-daily}"
    local timestamp=$(date +"$DATE_FORMAT")
    local backup_file="$BACKUP_DIR/$backup_type/radarr_${DB_TYPE}_${backup_type}_${timestamp}.sql"

    setup_backup_dir
    parse_db_config

    log_info "Creating $backup_type backup for $DB_TYPE database"

    case "$DB_TYPE" in
        postgres|postgresql)
            backup_file="$BACKUP_DIR/$backup_type/radarr_${DB_TYPE}_${backup_type}_${timestamp}.dump"
            backup_postgres "$backup_file"
            ;;
        mysql|mariadb)
            backup_mysql "$backup_file"
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            exit 1
            ;;
    esac

    # Compress backup if it's a SQL file
    if [[ "$backup_file" == *.sql ]]; then
        log_info "Compressing backup file"
        gzip "$backup_file"
        backup_file="${backup_file}.gz"
    fi

    log_info "Backup created: $backup_file"

    # Apply retention policy
    cleanup_old_backups "$backup_type"
}

# Cleanup old backups based on retention policy
cleanup_old_backups() {
    local backup_type="$1"
    local retention_days="$RETENTION_DAYS"

    # Different retention for different backup types
    case "$backup_type" in
        daily)
            retention_days=7
            ;;
        weekly)
            retention_days=30
            ;;
        monthly)
            retention_days=365
            ;;
    esac

    log_info "Cleaning up $backup_type backups older than $retention_days days"

    if [[ -d "$BACKUP_DIR/$backup_type" ]]; then
        find "$BACKUP_DIR/$backup_type" -name "radarr_*" -type f -mtime +$retention_days -exec rm -f {} \;
        log_info "Cleanup completed for $backup_type backups"
    fi
}

# Database health check
health_check() {
    parse_db_config
    log_info "Performing database health check"

    case "$DB_TYPE" in
        postgres|postgresql)
            health_check_postgres
            ;;
        mysql|mariadb)
            health_check_mysql
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            exit 1
            ;;
    esac
}

# PostgreSQL health check
health_check_postgres() {
    export PGPASSWORD="$DB_PASSWORD"

    log_info "Checking PostgreSQL connection..."
    if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; then
        log_info "✓ Database is accepting connections"
    else
        log_error "✗ Database is not accepting connections"
        return 1
    fi

    log_info "Checking database size and stats..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT
            pg_size_pretty(pg_database_size('$DB_NAME')) as database_size,
            (SELECT count(*) FROM movies) as movie_count,
            (SELECT count(*) FROM tasks) as task_count,
            (SELECT count(*) FROM health_issues WHERE is_resolved = false) as open_health_issues;
    " 2>/dev/null | while read -r line; do
        log_info "Stats: $line"
    done

    log_info "Checking for long-running queries..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT
            pid,
            now() - pg_stat_activity.query_start AS duration,
            query
        FROM pg_stat_activity
        WHERE (now() - pg_stat_activity.query_start) > interval '5 minutes';
    " 2>/dev/null | while read -r line; do
        if [[ -n "$line" && "$line" != " " ]]; then
            log_warn "Long-running query detected: $line"
        fi
    done

    unset PGPASSWORD
}

# MySQL/MariaDB health check
health_check_mysql() {
    log_info "Checking MySQL/MariaDB connection..."
    if mysqladmin -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" ping >/dev/null 2>&1; then
        log_info "✓ Database is accepting connections"
    else
        log_error "✗ Database is not accepting connections"
        return 1
    fi

    log_info "Checking database size and stats..."
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$DB_NAME" -sNe "
        SELECT
            CONCAT(ROUND(SUM(data_length + index_length) / 1024 / 1024, 2), ' MB') AS database_size,
            (SELECT COUNT(*) FROM movies) as movie_count,
            (SELECT COUNT(*) FROM tasks) as task_count,
            (SELECT COUNT(*) FROM health_issues WHERE is_resolved = 0) as open_health_issues;
    " 2>/dev/null | while read -r line; do
        log_info "Stats: $line"
    done

    log_info "Checking for long-running queries..."
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -sNe "
        SELECT
            ID,
            TIME,
            INFO
        FROM INFORMATION_SCHEMA.PROCESSLIST
        WHERE TIME > 300 AND COMMAND != 'Sleep';
    " 2>/dev/null | while read -r line; do
        if [[ -n "$line" ]]; then
            log_warn "Long-running query detected: $line"
        fi
    done
}

# Database optimization
optimize_database() {
    parse_db_config
    log_info "Optimizing database performance"

    case "$DB_TYPE" in
        postgres|postgresql)
            optimize_postgres
            ;;
        mysql|mariadb)
            optimize_mysql
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            exit 1
            ;;
    esac
}

# PostgreSQL optimization
optimize_postgres() {
    export PGPASSWORD="$DB_PASSWORD"

    log_info "Running PostgreSQL ANALYZE..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "ANALYZE;" >/dev/null 2>&1

    log_info "Running PostgreSQL VACUUM..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "VACUUM (ANALYZE, VERBOSE);" 2>&1 | grep -E "(INFO|DETAIL)" | head -10

    log_info "Checking for unused indexes..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT
            schemaname,
            tablename,
            indexname,
            idx_scan,
            idx_tup_read,
            idx_tup_fetch
        FROM pg_stat_user_indexes
        WHERE idx_scan = 0 AND schemaname = 'public'
        ORDER BY schemaname, tablename, indexname;
    " 2>/dev/null | while read -r line; do
        if [[ -n "$line" && "$line" != " " ]]; then
            log_warn "Unused index detected: $line"
        fi
    done

    unset PGPASSWORD
}

# MySQL/MariaDB optimization
optimize_mysql() {
    log_info "Running MySQL/MariaDB ANALYZE TABLE..."
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$DB_NAME" -e "
        SELECT CONCAT('ANALYZE TABLE ', table_name, ';')
        FROM information_schema.tables
        WHERE table_schema = '$DB_NAME'
    " -sN | mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$DB_NAME" >/dev/null 2>&1

    log_info "Running MySQL/MariaDB OPTIMIZE TABLE..."
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$DB_NAME" -e "
        SELECT CONCAT('OPTIMIZE TABLE ', table_name, ';')
        FROM information_schema.tables
        WHERE table_schema = '$DB_NAME' AND engine = 'InnoDB'
    " -sN | mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$DB_NAME" >/dev/null 2>&1
}

# Monitor database metrics
monitor_metrics() {
    parse_db_config
    log_info "Collecting database metrics"

    case "$DB_TYPE" in
        postgres|postgresql)
            monitor_postgres_metrics
            ;;
        mysql|mariadb)
            monitor_mysql_metrics
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            exit 1
            ;;
    esac
}

# PostgreSQL metrics monitoring
monitor_postgres_metrics() {
    export PGPASSWORD="$DB_PASSWORD"

    log_info "=== PostgreSQL Database Metrics ==="

    # Connection statistics
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT
            'Active Connections: ' || count(*)
        FROM pg_stat_activity
        WHERE state = 'active';

        SELECT
            'Total Connections: ' || count(*)
        FROM pg_stat_activity;

        SELECT
            'Max Connections: ' || setting
        FROM pg_settings
        WHERE name = 'max_connections';
    " 2>/dev/null | while read -r line; do
        log_info "$line"
    done

    # Database size and statistics
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "
        SELECT
            'Database Size: ' || pg_size_pretty(pg_database_size('$DB_NAME'));

        SELECT
            'Largest Tables:';

        SELECT
            '  ' || schemaname || '.' || tablename || ': ' || pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename))
        FROM pg_tables
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
        LIMIT 5;
    " 2>/dev/null | while read -r line; do
        log_info "$line"
    done

    unset PGPASSWORD
}

# MySQL/MariaDB metrics monitoring
monitor_mysql_metrics() {
    log_info "=== MySQL/MariaDB Database Metrics ==="

    # Connection statistics
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "
        SELECT
            CONCAT('Active Connections: ', COUNT(*)) as metric
        FROM INFORMATION_SCHEMA.PROCESSLIST
        WHERE COMMAND != 'Sleep'
        UNION ALL
        SELECT
            CONCAT('Total Connections: ', VARIABLE_VALUE) as metric
        FROM INFORMATION_SCHEMA.SESSION_STATUS
        WHERE VARIABLE_NAME = 'Threads_connected'
        UNION ALL
        SELECT
            CONCAT('Max Connections: ', @@max_connections) as metric;
    " -sN 2>/dev/null | while read -r line; do
        log_info "$line"
    done

    # Database size and statistics
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$DB_NAME" -e "
        SELECT
            CONCAT('Database Size: ', ROUND(SUM(data_length + index_length) / 1024 / 1024, 2), ' MB') AS metric
        FROM information_schema.tables
        WHERE table_schema = '$DB_NAME'
        UNION ALL
        SELECT 'Largest Tables:'
        UNION ALL
        SELECT
            CONCAT('  ', table_name, ': ', ROUND((data_length + index_length) / 1024 / 1024, 2), ' MB')
        FROM information_schema.tables
        WHERE table_schema = '$DB_NAME'
        ORDER BY (data_length + index_length) DESC
        LIMIT 5;
    " -sN 2>/dev/null | while read -r line; do
        log_info "$line"
    done
}

# Disaster recovery test
test_restore() {
    local backup_file="$1"
    local test_db_name="${DB_NAME}_restore_test"

    if [[ -z "$backup_file" ]]; then
        log_error "Backup file path required for restore test"
        exit 1
    fi

    if [[ ! -f "$backup_file" ]]; then
        log_error "Backup file not found: $backup_file"
        exit 1
    fi

    parse_db_config
    log_info "Testing restore from backup: $backup_file"

    case "$DB_TYPE" in
        postgres|postgresql)
            test_restore_postgres "$backup_file" "$test_db_name"
            ;;
        mysql|mariadb)
            test_restore_mysql "$backup_file" "$test_db_name"
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            exit 1
            ;;
    esac
}

# PostgreSQL restore test
test_restore_postgres() {
    local backup_file="$1"
    local test_db_name="$2"

    export PGPASSWORD="$DB_PASSWORD"

    # Create test database
    log_info "Creating test database: $test_db_name"
    createdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db_name" 2>/dev/null || true

    # Restore backup
    log_info "Restoring backup to test database..."
    if pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$test_db_name" --clean --if-exists "$backup_file" >/dev/null 2>&1; then
        log_info "✓ Restore test successful"

        # Validate restored data
        local table_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$test_db_name" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | tr -d ' ')
        log_info "✓ Restored $table_count tables"

        # Cleanup test database
        log_info "Cleaning up test database..."
        dropdb -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" "$test_db_name" 2>/dev/null || true
    else
        log_error "✗ Restore test failed"
        return 1
    fi

    unset PGPASSWORD
}

# MySQL/MariaDB restore test
test_restore_mysql() {
    local backup_file="$1"
    local test_db_name="$2"

    # Handle compressed backups
    local restore_cmd="mysql -h $DB_HOST -P $DB_PORT -u $DB_USER -p$DB_PASSWORD $test_db_name"
    if [[ "$backup_file" == *.gz ]]; then
        restore_cmd="gunzip -c '$backup_file' | $restore_cmd"
    else
        restore_cmd="$restore_cmd < '$backup_file'"
    fi

    # Create test database
    log_info "Creating test database: $test_db_name"
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS $test_db_name;" 2>/dev/null

    # Restore backup
    log_info "Restoring backup to test database..."
    if eval "$restore_cmd" >/dev/null 2>&1; then
        log_info "✓ Restore test successful"

        # Validate restored data
        local table_count=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -D "$test_db_name" -sNe "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$test_db_name';" 2>/dev/null)
        log_info "✓ Restored $table_count tables"

        # Cleanup test database
        log_info "Cleaning up test database..."
        mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "DROP DATABASE IF EXISTS $test_db_name;" 2>/dev/null
    else
        log_error "✗ Restore test failed"
        return 1
    fi
}

# Connection pool monitoring
monitor_connections() {
    parse_db_config
    log_info "Monitoring database connections"

    case "$DB_TYPE" in
        postgres|postgresql)
            monitor_postgres_connections
            ;;
        mysql|mariadb)
            monitor_mysql_connections
            ;;
        *)
            log_error "Unsupported database type: $DB_TYPE"
            exit 1
            ;;
    esac
}

# PostgreSQL connection monitoring
monitor_postgres_connections() {
    export PGPASSWORD="$DB_PASSWORD"

    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT
            'Connection Status' as metric,
            state,
            count(*) as count
        FROM pg_stat_activity
        GROUP BY state
        ORDER BY count DESC;

        SELECT
            'Waiting Connections' as metric,
            wait_event_type,
            wait_event,
            count(*) as count
        FROM pg_stat_activity
        WHERE wait_event IS NOT NULL
        GROUP BY wait_event_type, wait_event
        ORDER BY count DESC
        LIMIT 10;
    " 2>/dev/null

    unset PGPASSWORD
}

# MySQL/MariaDB connection monitoring
monitor_mysql_connections() {
    mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "
        SELECT
            'Connection Status' as metric,
            COMMAND as state,
            COUNT(*) as count
        FROM INFORMATION_SCHEMA.PROCESSLIST
        GROUP BY COMMAND
        ORDER BY count DESC;

        SHOW STATUS LIKE 'Threads_%';
        SHOW STATUS LIKE 'Connections';
        SHOW STATUS LIKE 'Max_used_connections';
    " 2>/dev/null
}

# Generate database report
generate_report() {
    local report_file="$BACKUP_DIR/database_report_$(date +"$DATE_FORMAT").txt"
    setup_backup_dir

    log_info "Generating database report: $report_file"

    {
        echo "=== Radarr-Go Database Report ==="
        echo "Generated: $(date)"
        echo "Database Type: $DB_TYPE"
        echo "Host: $DB_HOST:$DB_PORT"
        echo "Database: $DB_NAME"
        echo ""

        echo "=== Health Check ==="
        health_check 2>&1
        echo ""

        echo "=== Metrics ==="
        monitor_metrics 2>&1
        echo ""

        echo "=== Connection Status ==="
        monitor_connections 2>&1
        echo ""

        echo "=== Backup Status ==="
        if [[ -d "$BACKUP_DIR" ]]; then
            echo "Recent backups:"
            find "$BACKUP_DIR" -name "radarr_*" -type f -mtime -7 -ls 2>/dev/null | head -10
        fi
        echo ""

    } > "$report_file"

    log_info "Database report generated: $report_file"
}

# Show usage information
usage() {
    cat << EOF
Radarr-Go Database Operations Script

Usage: $0 [COMMAND] [OPTIONS]

Commands:
    backup [daily|weekly|monthly]   Create database backup with retention policy
    health                          Perform database health check
    optimize                        Optimize database performance (VACUUM/ANALYZE)
    monitor                         Display database metrics and statistics
    connections                     Monitor database connection pool status
    test-restore <backup_file>      Test restore functionality with backup file
    report                          Generate comprehensive database report
    cleanup                         Clean up old backup files

Examples:
    $0 backup daily                 # Create daily backup
    $0 health                       # Check database health
    $0 optimize                     # Optimize database
    $0 monitor                      # Show database metrics
    $0 test-restore backup.dump     # Test restore from backup
    $0 report                       # Generate full report

Environment Variables:
    DEBUG=1                         # Enable debug output
    RETENTION_DAYS=30              # Override backup retention period

Configuration:
    Uses config.yaml for database connection settings
    Falls back to RADARR_DATABASE_* environment variables

EOF
}

# Main script logic
main() {
    case "${1:-}" in
        backup)
            create_backup "${2:-daily}"
            ;;
        health)
            health_check
            ;;
        optimize)
            optimize_database
            ;;
        monitor)
            monitor_metrics
            ;;
        connections)
            monitor_connections
            ;;
        test-restore)
            test_restore "$2"
            ;;
        report)
            generate_report
            ;;
        cleanup)
            cleanup_old_backups "daily"
            cleanup_old_backups "weekly"
            cleanup_old_backups "monthly"
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            log_error "Unknown command: ${1:-}"
            echo ""
            usage
            exit 1
            ;;
    esac
}

# Check dependencies
check_dependencies() {
    local missing_deps=()

    case "$DB_TYPE" in
        postgres|postgresql)
            command -v pg_dump >/dev/null 2>&1 || missing_deps+=("postgresql-client")
            command -v pg_restore >/dev/null 2>&1 || missing_deps+=("postgresql-client")
            ;;
        mysql|mariadb)
            command -v mysqldump >/dev/null 2>&1 || missing_deps+=("mysql-client")
            command -v mysql >/dev/null 2>&1 || missing_deps+=("mysql-client")
            ;;
    esac

    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_error "Please install the required database client tools"
        exit 1
    fi
}

# Initialize and run
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Parse config to check dependencies
    parse_db_config
    check_dependencies

    # Run main function
    main "$@"
fi
