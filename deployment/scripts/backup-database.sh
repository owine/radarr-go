#!/bin/bash
# scripts/backup-database.sh - Automated database backup script
# Creates compressed, timestamped backups with optional encryption

set -euo pipefail

# Configuration
POSTGRES_HOST="${POSTGRES_HOST:-postgres}"
POSTGRES_PORT="${POSTGRES_PORT:-5432}"
POSTGRES_USER="${POSTGRES_USER:-radarr}"
POSTGRES_DB="${POSTGRES_DB:-radarr}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
BACKUP_ENCRYPTION_PASSWORD="${BACKUP_ENCRYPTION_PASSWORD:-}"

# Logging
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$BACKUP_DIR/backup.log"
}

error() {
    log "ERROR: $1"
    exit 1
}

# Create backup
create_backup() {
    local timestamp=$(date +%Y%m%d_%H%M%S)
    local backup_file="$BACKUP_DIR/radarr_backup_$timestamp.sql"
    local compressed_file="$backup_file.gz"
    local encrypted_file="$compressed_file.enc"

    log "Starting database backup..."

    # Create backup directory if it doesn't exist
    mkdir -p "$BACKUP_DIR"

    # Create database dump
    if ! PGPASSWORD="$PGPASSWORD" pg_dump \
        -h "$POSTGRES_HOST" \
        -p "$POSTGRES_PORT" \
        -U "$POSTGRES_USER" \
        -d "$POSTGRES_DB" \
        --verbose \
        --clean \
        --create \
        --if-exists \
        --no-owner \
        --no-privileges > "$backup_file"; then
        error "Failed to create database dump"
    fi

    log "Database dump created: $backup_file"

    # Compress backup
    if ! gzip "$backup_file"; then
        error "Failed to compress backup"
    fi

    log "Backup compressed: $compressed_file"

    # Encrypt backup if password provided
    if [ -n "$BACKUP_ENCRYPTION_PASSWORD" ]; then
        if command -v openssl >/dev/null 2>&1; then
            if ! openssl enc -aes-256-cbc -salt -in "$compressed_file" -out "$encrypted_file" -k "$BACKUP_ENCRYPTION_PASSWORD"; then
                error "Failed to encrypt backup"
            fi

            # Remove unencrypted file
            rm -f "$compressed_file"
            log "Backup encrypted: $encrypted_file"
        else
            log "WARNING: openssl not available, backup not encrypted"
        fi
    fi

    # Get final backup file
    local final_backup_file="$compressed_file"
    if [ -f "$encrypted_file" ]; then
        final_backup_file="$encrypted_file"
    fi

    # Verify backup
    if [ -f "$final_backup_file" ] && [ -s "$final_backup_file" ]; then
        local backup_size=$(stat -f%z "$final_backup_file" 2>/dev/null || stat -c%s "$final_backup_file" 2>/dev/null || echo "unknown")
        log "Backup completed successfully: $final_backup_file (size: $backup_size bytes)"
    else
        error "Backup verification failed"
    fi

    # Create backup metadata
    cat > "$final_backup_file.meta" << EOF
backup_date=$(date -u '+%Y-%m-%d %H:%M:%S UTC')
postgres_version=$(PGPASSWORD="$PGPASSWORD" psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "SELECT version();" | head -1 | xargs)
database_name=$POSTGRES_DB
backup_size=$backup_size
encrypted=$( [ -f "$encrypted_file" ] && echo "true" || echo "false" )
checksum=$(sha256sum "$final_backup_file" | cut -d' ' -f1)
EOF

    log "Backup metadata created: $final_backup_file.meta"
}

# Clean old backups
cleanup_old_backups() {
    log "Cleaning up backups older than $BACKUP_RETENTION_DAYS days..."

    local deleted_count=0

    # Find and remove old backup files
    find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -mtime "+$BACKUP_RETENTION_DAYS" | while read -r old_backup; do
        log "Removing old backup: $old_backup"
        rm -f "$old_backup" "$old_backup.meta" 2>/dev/null || true
        ((deleted_count++))
    done

    log "Cleanup completed. Removed $deleted_count old backup files."

    # Show remaining backups
    local remaining_count=$(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f | wc -l)
    log "Remaining backups: $remaining_count"
}

# Verify database connection
verify_connection() {
    log "Verifying database connection..."

    if ! PGPASSWORD="$PGPASSWORD" pg_isready -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
        error "Cannot connect to database"
    fi

    log "Database connection verified"
}

# Test backup restore (optional)
test_restore() {
    local backup_file="$1"
    local test_db="radarr_test_restore_$(date +%s)"

    log "Testing backup restore with test database: $test_db"

    # Create test database
    if ! PGPASSWORD="$PGPASSWORD" createdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$test_db" >/dev/null 2>&1; then
        log "WARNING: Failed to create test database for restore test"
        return 1
    fi

    # Determine how to restore based on file extension
    local restore_cmd=""
    if [[ "$backup_file" == *.enc ]]; then
        if [ -z "$BACKUP_ENCRYPTION_PASSWORD" ]; then
            log "WARNING: Cannot test encrypted backup without password"
            PGPASSWORD="$PGPASSWORD" dropdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$test_db" >/dev/null 2>&1 || true
            return 1
        fi
        restore_cmd="openssl enc -aes-256-cbc -d -in '$backup_file' -k '$BACKUP_ENCRYPTION_PASSWORD' | gunzip | PGPASSWORD='$PGPASSWORD' psql -h '$POSTGRES_HOST' -p '$POSTGRES_PORT' -U '$POSTGRES_USER' -d '$test_db'"
    elif [[ "$backup_file" == *.gz ]]; then
        restore_cmd="gunzip -c '$backup_file' | PGPASSWORD='$PGPASSWORD' psql -h '$POSTGRES_HOST' -p '$POSTGRES_PORT' -U '$POSTGRES_USER' -d '$test_db'"
    else
        restore_cmd="PGPASSWORD='$PGPASSWORD' psql -h '$POSTGRES_HOST' -p '$POSTGRES_PORT' -U '$POSTGRES_USER' -d '$test_db' < '$backup_file'"
    fi

    # Perform restore test
    if eval "$restore_cmd" >/dev/null 2>&1; then
        log "Backup restore test passed"
    else
        log "WARNING: Backup restore test failed"
    fi

    # Clean up test database
    PGPASSWORD="$PGPASSWORD" dropdb -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" "$test_db" >/dev/null 2>&1 || true
}

# Generate backup report
generate_backup_report() {
    local report_file="$BACKUP_DIR/backup_report_$(date +%Y%m%d).txt"

    {
        echo "Radarr Database Backup Report"
        echo "============================"
        echo "Generated: $(date)"
        echo "Database: $POSTGRES_DB"
        echo "Host: $POSTGRES_HOST:$POSTGRES_PORT"
        echo "User: $POSTGRES_USER"
        echo "Backup Directory: $BACKUP_DIR"
        echo "Retention Days: $BACKUP_RETENTION_DAYS"
        echo "Encryption: $( [ -n "$BACKUP_ENCRYPTION_PASSWORD" ] && echo "Enabled" || echo "Disabled" )"
        echo ""

        echo "Recent Backups:"
        echo "==============="
        find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" | \
            sort -r | head -10 | while read -r backup; do
            local size=$(stat -f%z "$backup" 2>/dev/null || stat -c%s "$backup" 2>/dev/null || echo "unknown")
            local date=$(basename "$backup" | sed 's/radarr_backup_\([0-9]\{8\}_[0-9]\{6\}\).*/\1/' | sed 's/_/ /' | sed 's/\([0-9]\{4\}\)\([0-9]\{2\}\)\([0-9]\{2\}\) \([0-9]\{2\}\)\([0-9]\{2\}\)\([0-9]\{2\}\)/\1-\2-\3 \4:\5:\6/')
            echo "$date - $(basename "$backup") ($size bytes)"
        done
        echo ""

        echo "Storage Usage:"
        echo "=============="
        echo "Total backup files: $(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" | wc -l)"
        local total_size=$(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -exec stat -f%z {} \; 2>/dev/null | awk '{sum+=$1} END {print sum+0}' || echo "0")
        echo "Total backup size: $total_size bytes"
        echo "Available space: $(df "$BACKUP_DIR" | awk 'NR==2 {print $4}')KB"
        echo ""

        echo "Backup Log (last 20 lines):"
        echo "============================"
        tail -20 "$BACKUP_DIR/backup.log" 2>/dev/null || echo "No log available"

    } > "$report_file"

    log "Backup report generated: $report_file"
}

# Main execution
main() {
    log "Starting backup process..."

    # Verify prerequisites
    if [ -z "${PGPASSWORD:-}" ]; then
        error "PGPASSWORD environment variable not set"
    fi

    if ! command -v pg_dump >/dev/null 2>&1; then
        error "pg_dump command not found"
    fi

    # Verify connection
    verify_connection

    # Create backup
    create_backup

    # Find the latest backup for testing
    local latest_backup=$(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" | sort -r | head -1)

    # Test restore (optional)
    if [ "${TEST_RESTORE:-false}" = "true" ] && [ -n "$latest_backup" ]; then
        test_restore "$latest_backup"
    fi

    # Cleanup old backups
    cleanup_old_backups

    # Generate report
    generate_backup_report

    log "Backup process completed successfully"
}

# Handle command line arguments
case "${1:-backup}" in
    "backup")
        main
        ;;
    "cleanup")
        cleanup_old_backups
        ;;
    "verify")
        verify_connection
        log "Database connection is healthy"
        ;;
    "report")
        generate_backup_report
        ;;
    "test")
        if [ -n "${2:-}" ] && [ -f "$2" ]; then
            test_restore "$2"
        else
            error "Please specify backup file to test"
        fi
        ;;
    *)
        echo "Usage: $0 {backup|cleanup|verify|report|test backup_file}"
        echo ""
        echo "Commands:"
        echo "  backup  - Create database backup (default)"
        echo "  cleanup - Remove old backup files"
        echo "  verify  - Test database connection"
        echo "  report  - Generate backup report"
        echo "  test    - Test restore from backup file"
        echo ""
        echo "Environment Variables:"
        echo "  POSTGRES_HOST               - Database host (default: postgres)"
        echo "  POSTGRES_PORT               - Database port (default: 5432)"
        echo "  POSTGRES_USER               - Database user (default: radarr)"
        echo "  POSTGRES_DB                 - Database name (default: radarr)"
        echo "  PGPASSWORD                  - Database password (required)"
        echo "  BACKUP_DIR                  - Backup directory (default: /backups)"
        echo "  BACKUP_RETENTION_DAYS       - Backup retention (default: 30)"
        echo "  BACKUP_ENCRYPTION_PASSWORD  - Encryption password (optional)"
        echo "  TEST_RESTORE                - Test restore after backup (default: false)"
        exit 1
        ;;
esac
