#!/bin/bash
# Database Backup and Restore Operations for Radarr Go
# Supports both PostgreSQL and MariaDB/MySQL with cross-database compatibility
#
# Usage:
#   ./backup_restore.sh backup postgresql
#   ./backup_restore.sh backup mariadb
#   ./backup_restore.sh restore postgresql /path/to/backup.sql
#   ./backup_restore.sh validate
#   ./backup_restore.sh performance-test

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
BACKUP_DIR="${PROJECT_ROOT}/backups"
LOG_FILE="${BACKUP_DIR}/backup_$(date +%Y%m%d_%H%M%S).log"

# Default database configurations
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

# Backup retention policy (days)
BACKUP_RETENTION_DAYS="${RADARR_BACKUP_RETENTION:-30}"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Check dependencies
check_dependencies() {
    case "$1" in
        "postgresql")
            command -v pg_dump >/dev/null 2>&1 || error_exit "pg_dump not found. Install PostgreSQL client tools."
            command -v psql >/dev/null 2>&1 || error_exit "psql not found. Install PostgreSQL client tools."
            ;;
        "mariadb"|"mysql")
            command -v mysqldump >/dev/null 2>&1 || error_exit "mysqldump not found. Install MySQL/MariaDB client tools."
            command -v mysql >/dev/null 2>&1 || error_exit "mysql not found. Install MySQL/MariaDB client tools."
            ;;
    esac
}

# Create backup directory
setup_backup_dir() {
    mkdir -p "$BACKUP_DIR"
    log "Backup directory: $BACKUP_DIR"
}

# Test database connectivity
test_connection() {
    log "Testing database connectivity..."

    case "$1" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            if ! psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' >/dev/null 2>&1; then
                error_exit "Cannot connect to PostgreSQL database"
            fi
            log "PostgreSQL connection successful"
            ;;
        "mariadb"|"mysql")
            if ! mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" -e 'SELECT 1;' >/dev/null 2>&1; then
                error_exit "Cannot connect to MariaDB/MySQL database"
            fi
            log "MariaDB/MySQL connection successful"
            ;;
    esac
}

# Generate backup filename
get_backup_filename() {
    local db_type="$1"
    local timestamp=$(date +%Y%m%d_%H%M%S)
    echo "${BACKUP_DIR}/radarr_${db_type}_${timestamp}.sql"
}

# Backup database with comprehensive options
backup_database() {
    local db_type="$1"
    local backup_file=$(get_backup_filename "$db_type")

    setup_backup_dir
    check_dependencies "$db_type"
    test_connection "$db_type"

    log "Starting backup of $db_type database to $backup_file"

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Create backup with full schema and data, optimized for large datasets
            if ! pg_dump \
                -h "$POSTGRES_HOST" \
                -p "$POSTGRES_PORT" \
                -U "$POSTGRES_USER" \
                -d "$POSTGRES_DB" \
                --verbose \
                --no-password \
                --format=plain \
                --no-tablespaces \
                --no-unlogged-table-data \
                --compress=6 \
                --lock-wait-timeout=300000 \
                --serializable-deferrable \
                --file="$backup_file" 2>>"$LOG_FILE"; then
                error_exit "PostgreSQL backup failed"
            fi
            ;;
        "mariadb"|"mysql")
            # Create backup with full schema and data, optimized for large datasets
            if ! mysqldump \
                -h "$MYSQL_HOST" \
                -P "$MYSQL_PORT" \
                -u "$MYSQL_USER" \
                -p"$MYSQL_PASSWORD" \
                --single-transaction \
                --routines \
                --triggers \
                --events \
                --set-gtid-purged=OFF \
                --compress \
                --lock-tables=false \
                --quick \
                --extended-insert \
                "$MYSQL_DB" > "$backup_file" 2>>"$LOG_FILE"; then
                error_exit "MariaDB/MySQL backup failed"
            fi
            ;;
    esac

    # Compress backup if large
    local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo 0)
    if [ "$file_size" -gt 52428800 ]; then  # 50MB
        log "Compressing backup (size: $file_size bytes)"
        gzip "$backup_file"
        backup_file="${backup_file}.gz"
    fi

    log "Backup completed successfully: $backup_file"
    log "Backup size: $(du -h "$backup_file" | cut -f1)"

    # Validate backup
    validate_backup "$backup_file" "$db_type"

    # Clean up old backups
    cleanup_old_backups

    echo "$backup_file"
}

# Validate backup integrity
validate_backup() {
    local backup_file="$1"
    local db_type="$2"

    log "Validating backup integrity..."

    # Check if file is compressed
    if [[ "$backup_file" == *.gz ]]; then
        if ! gzip -t "$backup_file" >/dev/null 2>&1; then
            error_exit "Backup file is corrupted (gzip test failed)"
        fi
        # Test SQL syntax
        if ! zcat "$backup_file" | head -100 | grep -q "CREATE TABLE\|INSERT INTO" >/dev/null 2>&1; then
            error_exit "Backup file appears to be empty or invalid"
        fi
    else
        # Test SQL syntax
        if ! head -100 "$backup_file" | grep -q "CREATE TABLE\|INSERT INTO" >/dev/null 2>&1; then
            error_exit "Backup file appears to be empty or invalid"
        fi
    fi

    log "Backup validation successful"
}

# Restore database from backup
restore_database() {
    local db_type="$1"
    local backup_file="$2"

    if [ ! -f "$backup_file" ]; then
        error_exit "Backup file not found: $backup_file"
    fi

    check_dependencies "$db_type"
    test_connection "$db_type"

    log "Starting restore of $db_type database from $backup_file"

    # Confirm before restore
    read -p "WARNING: This will overwrite the current database. Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log "Restore cancelled by user"
        exit 0
    fi

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Drop and recreate database for clean restore
            log "Dropping and recreating PostgreSQL database"
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
                -c "DROP DATABASE IF EXISTS \"$POSTGRES_DB\";" \
                -c "CREATE DATABASE \"$POSTGRES_DB\";" 2>>"$LOG_FILE" || error_exit "Failed to recreate database"

            # Restore from backup
            if [[ "$backup_file" == *.gz ]]; then
                if ! zcat "$backup_file" | psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" 2>>"$LOG_FILE"; then
                    error_exit "PostgreSQL restore failed"
                fi
            else
                if ! psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" < "$backup_file" 2>>"$LOG_FILE"; then
                    error_exit "PostgreSQL restore failed"
                fi
            fi
            ;;
        "mariadb"|"mysql")
            # Drop and recreate database for clean restore
            log "Dropping and recreating MariaDB/MySQL database"
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
                -e "DROP DATABASE IF EXISTS \`$MYSQL_DB\`; CREATE DATABASE \`$MYSQL_DB\`;" 2>>"$LOG_FILE" || error_exit "Failed to recreate database"

            # Restore from backup
            if [[ "$backup_file" == *.gz ]]; then
                if ! zcat "$backup_file" | mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" 2>>"$LOG_FILE"; then
                    error_exit "MariaDB/MySQL restore failed"
                fi
            else
                if ! mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" < "$backup_file" 2>>"$LOG_FILE"; then
                    error_exit "MariaDB/MySQL restore failed"
                fi
            fi
            ;;
    esac

    log "Database restore completed successfully"

    # Run post-restore validation
    validate_restored_database "$db_type"
}

# Validate restored database
validate_restored_database() {
    local db_type="$1"

    log "Validating restored database..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Check critical tables exist
            local table_count=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
                -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('movies', 'quality_definitions', 'quality_profiles');" | tr -d ' ')

            if [ "$table_count" -ne 3 ]; then
                error_exit "Critical tables missing after restore (found $table_count/3)"
            fi
            ;;
        "mariadb"|"mysql")
            # Check critical tables exist
            local table_count=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" \
                -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = '$MYSQL_DB' AND table_name IN ('movies', 'quality_definitions', 'quality_profiles');" -s)

            if [ "$table_count" -ne 3 ]; then
                error_exit "Critical tables missing after restore (found $table_count/3)"
            fi
            ;;
    esac

    log "Database validation successful"
}

# Clean up old backups based on retention policy
cleanup_old_backups() {
    log "Cleaning up backups older than $BACKUP_RETENTION_DAYS days"

    if [ -d "$BACKUP_DIR" ]; then
        find "$BACKUP_DIR" -name "radarr_*.sql*" -mtime +$BACKUP_RETENTION_DAYS -type f -exec rm -f {} \;
        log "Old backups cleaned up"
    fi
}

# Validate current database schema
validate_schema() {
    log "Validating current database schema..."

    # Check both databases if available
    for db_type in postgresql mariadb; do
        if check_dependencies_quiet "$db_type" && test_connection_quiet "$db_type"; then
            validate_schema_specific "$db_type"
        fi
    done
}

# Quiet dependency check
check_dependencies_quiet() {
    case "$1" in
        "postgresql")
            command -v pg_dump >/dev/null 2>&1 && command -v psql >/dev/null 2>&1
            ;;
        "mariadb"|"mysql")
            command -v mysqldump >/dev/null 2>&1 && command -v mysql >/dev/null 2>&1
            ;;
    esac
}

# Quiet connection test
test_connection_quiet() {
    case "$1" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' >/dev/null 2>&1
            ;;
        "mariadb"|"mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" -e 'SELECT 1;' >/dev/null 2>&1
            ;;
    esac
}

# Validate specific database schema
validate_schema_specific() {
    local db_type="$1"
    log "Validating $db_type schema..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Check for foreign key constraint violations
            local fk_violations=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
                SELECT COUNT(*) FROM (
                    SELECT tc.table_name, tc.constraint_name
                    FROM information_schema.table_constraints tc
                    JOIN information_schema.referential_constraints rc ON tc.constraint_name = rc.constraint_name
                    WHERE tc.constraint_type = 'FOREIGN KEY'
                    AND NOT EXISTS (
                        SELECT 1 FROM information_schema.tables t
                        WHERE t.table_name = rc.unique_constraint_schema
                    )
                ) violations;" | tr -d ' ')

            if [ "$fk_violations" -gt 0 ]; then
                error_exit "$db_type has $fk_violations foreign key constraint violations"
            fi
            ;;
        "mariadb"|"mysql")
            # Check for foreign key constraint violations
            local fk_violations=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DB" -e "
                SELECT COUNT(*) FROM (
                    SELECT CONSTRAINT_NAME
                    FROM information_schema.REFERENTIAL_CONSTRAINTS
                    WHERE CONSTRAINT_SCHEMA = '$MYSQL_DB'
                    AND REFERENCED_TABLE_NAME NOT IN (
                        SELECT TABLE_NAME FROM information_schema.TABLES
                        WHERE TABLE_SCHEMA = '$MYSQL_DB'
                    )
                ) violations;" -s)

            if [ "$fk_violations" -gt 0 ]; then
                error_exit "$db_type has $fk_violations foreign key constraint violations"
            fi
            ;;
    esac

    log "$db_type schema validation successful"
}

# Performance test with synthetic data
performance_test() {
    log "Running database performance test..."

    for db_type in postgresql mariadb; do
        if check_dependencies_quiet "$db_type" && test_connection_quiet "$db_type"; then
            log "Testing $db_type performance..."
            performance_test_specific "$db_type"
        else
            log "Skipping $db_type performance test (not available)"
        fi
    done
}

# Database-specific performance test
performance_test_specific() {
    local db_type="$1"
    local temp_db="radarr_perf_test_$(date +%s)"

    log "Creating temporary database $temp_db for performance testing..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Create temporary database
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
                -c "CREATE DATABASE \"$temp_db\";" >/dev/null 2>&1 || error_exit "Failed to create test database"

            # Run migrations
            cd "$PROJECT_ROOT"
            RADARR_DATABASE_NAME="$temp_db" ./radarr migrate-up >/dev/null 2>&1 || {
                psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres -c "DROP DATABASE \"$temp_db\";" >/dev/null 2>&1
                error_exit "Migration failed during performance test"
            }

            # Insert test data (10k movies)
            log "Inserting 10,000 test movies for performance testing..."
            local start_time=$(date +%s)

            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$temp_db" -c "
                INSERT INTO movies (tmdb_id, title, title_slug, year, monitored, quality_profile_id, added)
                SELECT
                    generate_series(1, 10000) as tmdb_id,
                    'Test Movie ' || generate_series(1, 10000) as title,
                    'test-movie-' || generate_series(1, 10000) as title_slug,
                    2000 + (generate_series(1, 10000) % 24) as year,
                    (generate_series(1, 10000) % 2)::boolean as monitored,
                    1 as quality_profile_id,
                    NOW() - (generate_series(1, 10000) || ' days')::interval as added;
            " >/dev/null 2>&1

            local insert_time=$(($(date +%s) - start_time))
            log "PostgreSQL: Inserted 10k movies in ${insert_time}s"

            # Test query performance
            local query_start=$(date +%s)
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$temp_db" -c "
                SELECT COUNT(*) FROM movies WHERE monitored = true AND year >= 2020;
            " >/dev/null 2>&1
            local query_time=$(($(date +%s) - query_start))
            log "PostgreSQL: Query performance test completed in ${query_time}s"

            # Cleanup
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
                -c "DROP DATABASE \"$temp_db\";" >/dev/null 2>&1
            ;;

        "mariadb"|"mysql")
            # Create temporary database
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
                -e "CREATE DATABASE \`$temp_db\`;" >/dev/null 2>&1 || error_exit "Failed to create test database"

            # Run migrations (note: this requires app to be built)
            cd "$PROJECT_ROOT"
            RADARR_DATABASE_NAME="$temp_db" ./radarr migrate-up >/dev/null 2>&1 || {
                mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" -e "DROP DATABASE \`$temp_db\`;" >/dev/null 2>&1
                error_exit "Migration failed during performance test"
            }

            # Insert test data (10k movies)
            log "Inserting 10,000 test movies for performance testing..."
            local start_time=$(date +%s)

            # Generate test data in batches for better performance
            for i in $(seq 1 100 10000); do
                local end=$((i + 99))
                if [ $end -gt 10000 ]; then
                    end=10000
                fi

                mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$temp_db" -e "
                    INSERT INTO movies (tmdb_id, title, title_slug, year, monitored, quality_profile_id, added) VALUES
                    $(for j in $(seq $i $end); do
                        echo "($j, 'Test Movie $j', 'test-movie-$j', $((2000 + j % 24)), $((j % 2)), 1, NOW() - INTERVAL $j DAY)"
                        [ $j -lt $end ] && echo ","
                    done)
                " >/dev/null 2>&1
            done

            local insert_time=$(($(date +%s) - start_time))
            log "MariaDB: Inserted 10k movies in ${insert_time}s"

            # Test query performance
            local query_start=$(date +%s)
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$temp_db" -e "
                SELECT COUNT(*) FROM movies WHERE monitored = true AND year >= 2020;
            " >/dev/null 2>&1
            local query_time=$(($(date +%s) - query_start))
            log "MariaDB: Query performance test completed in ${query_time}s"

            # Cleanup
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
                -e "DROP DATABASE \`$temp_db\`;" >/dev/null 2>&1
            ;;
    esac

    log "Performance test completed for $db_type"
}

# Main function
main() {
    local command="$1"

    setup_backup_dir

    case "$command" in
        "backup")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 backup <postgresql|mariadb>"
            fi
            backup_database "$2"
            ;;
        "restore")
            if [ $# -lt 3 ]; then
                error_exit "Usage: $0 restore <postgresql|mariadb> <backup_file>"
            fi
            restore_database "$2" "$3"
            ;;
        "validate")
            validate_schema
            ;;
        "performance-test")
            performance_test
            ;;
        "cleanup")
            cleanup_old_backups
            ;;
        *)
            echo "Usage: $0 <backup|restore|validate|performance-test|cleanup>"
            echo
            echo "Commands:"
            echo "  backup <db_type>              Create database backup"
            echo "  restore <db_type> <file>      Restore from backup file"
            echo "  validate                      Validate current schema"
            echo "  performance-test              Run performance tests"
            echo "  cleanup                       Remove old backups"
            echo
            echo "Database types: postgresql, mariadb"
            echo
            echo "Environment variables:"
            echo "  RADARR_DATABASE_HOST          Database host (default: localhost)"
            echo "  RADARR_DATABASE_PORT          Database port (default: 5432 for PostgreSQL, 3306 for MariaDB)"
            echo "  RADARR_DATABASE_USERNAME      Database username (default: radarr)"
            echo "  RADARR_DATABASE_NAME          Database name (default: radarr)"
            echo "  RADARR_DATABASE_PASSWORD      Database password (default: password)"
            echo "  RADARR_BACKUP_RETENTION       Backup retention days (default: 30)"
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"
