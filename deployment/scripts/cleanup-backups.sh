#!/bin/bash
# scripts/cleanup-backups.sh - Automated backup cleanup script
# Removes old backups and maintains storage limits

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-/backups}"
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
MAX_BACKUP_SIZE_GB="${MAX_BACKUP_SIZE_GB:-50}"
MIN_FREE_SPACE_GB="${MIN_FREE_SPACE_GB:-5}"

# Logging
log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] CLEANUP: $1" | tee -a "$BACKUP_DIR/cleanup.log"
}

error() {
    log "ERROR: $1"
    exit 1
}

# Clean old backups by age
cleanup_by_age() {
    log "Cleaning backups older than $BACKUP_RETENTION_DAYS days..."

    local deleted_count=0
    local deleted_size=0

    # Find old backup files
    while IFS= read -r -d '' backup_file; do
        if [ -f "$backup_file" ]; then
            local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
            log "Removing old backup: $(basename "$backup_file") (${file_size} bytes)"

            # Remove backup and metadata
            rm -f "$backup_file" "${backup_file}.meta" 2>/dev/null || true

            ((deleted_count++))
            deleted_size=$((deleted_size + file_size))
        fi
    done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -mtime "+$BACKUP_RETENTION_DAYS" -print0 2>/dev/null || true)

    if [ $deleted_count -gt 0 ]; then
        log "Removed $deleted_count old backup files ($(human_readable_size $deleted_size))"
    else
        log "No old backup files to remove"
    fi
}

# Clean backups by size limit
cleanup_by_size() {
    local max_size_bytes=$((MAX_BACKUP_SIZE_GB * 1024 * 1024 * 1024))

    log "Checking backup size limit (max: ${MAX_BACKUP_SIZE_GB}GB)..."

    # Calculate current backup size
    local current_size=0
    local backup_files=()

    while IFS= read -r -d '' backup_file; do
        if [ -f "$backup_file" ]; then
            local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
            current_size=$((current_size + file_size))
            backup_files+=("$backup_file:$file_size")
        fi
    done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -print0 2>/dev/null || true)

    log "Current backup size: $(human_readable_size $current_size)"

    if [ $current_size -le $max_size_bytes ]; then
        log "Backup size within limit"
        return 0
    fi

    log "Backup size exceeds limit, removing oldest backups..."

    # Sort by modification time (oldest first) and remove until within limit
    local deleted_count=0
    local deleted_size=0

    for backup_info in $(printf '%s\n' "${backup_files[@]}" | while read -r line; do
        local file_path="${line%:*}"
        local file_size="${line##*:}"
        local mod_time=$(stat -f%m "$file_path" 2>/dev/null || stat -c%Y "$file_path" 2>/dev/null || echo "0")
        echo "$mod_time:$file_path:$file_size"
    done | sort -n); do

        local file_path=$(echo "$backup_info" | cut -d: -f2)
        local file_size=$(echo "$backup_info" | cut -d: -f3)

        if [ $current_size -le $max_size_bytes ]; then
            break
        fi

        log "Removing backup to free space: $(basename "$file_path") ($(human_readable_size $file_size))"
        rm -f "$file_path" "${file_path}.meta" 2>/dev/null || true

        current_size=$((current_size - file_size))
        deleted_size=$((deleted_size + file_size))
        ((deleted_count++))
    done

    if [ $deleted_count -gt 0 ]; then
        log "Removed $deleted_count backups to stay within size limit (freed $(human_readable_size $deleted_size))"
    fi
}

# Clean backups by free space
cleanup_by_free_space() {
    local min_free_bytes=$((MIN_FREE_SPACE_GB * 1024 * 1024 * 1024))

    log "Checking free space requirement (min: ${MIN_FREE_SPACE_GB}GB)..."

    # Get available space
    local free_space_kb=$(df "$BACKUP_DIR" | awk 'NR==2 {print $4}')
    local free_space_bytes=$((free_space_kb * 1024))

    log "Current free space: $(human_readable_size $free_space_bytes)"

    if [ $free_space_bytes -ge $min_free_bytes ]; then
        log "Free space sufficient"
        return 0
    fi

    log "Free space below minimum, removing oldest backups..."

    local space_needed=$((min_free_bytes - free_space_bytes))
    local deleted_count=0
    local deleted_size=0

    # Get backup files sorted by age (oldest first)
    while IFS= read -r backup_file; do
        if [ ! -f "$backup_file" ]; then
            continue
        fi

        local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")

        log "Removing backup for space: $(basename "$backup_file") ($(human_readable_size $file_size))"
        rm -f "$backup_file" "${backup_file}.meta" 2>/dev/null || true

        deleted_size=$((deleted_size + file_size))
        ((deleted_count++))

        # Check if we've freed enough space
        if [ $deleted_size -ge $space_needed ]; then
            break
        fi
    done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -printf '%T@ %p\n' 2>/dev/null | sort -n | cut -d' ' -f2- ||
             find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -exec stat -f'%m %N' {} \; 2>/dev/null | sort -n | cut -d' ' -f2- ||
             find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" 2>/dev/null)

    if [ $deleted_count -gt 0 ]; then
        log "Removed $deleted_count backups to free space (freed $(human_readable_size $deleted_size))"
    fi
}

# Clean corrupted or incomplete backups
cleanup_corrupted_backups() {
    log "Checking for corrupted or incomplete backups..."

    local corrupted_count=0

    while IFS= read -r -d '' backup_file; do
        if [ ! -f "$backup_file" ]; then
            continue
        fi

        local is_corrupted=false
        local reason=""

        # Check file size (should be > 1KB for valid backup)
        local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
        if [ $file_size -lt 1024 ]; then
            is_corrupted=true
            reason="file too small (${file_size} bytes)"
        fi

        # Check if gzipped file is valid
        if [ "$is_corrupted" = "false" ] && [[ "$backup_file" == *.gz* ]]; then
            if ! gzip -t "$backup_file" 2>/dev/null; then
                is_corrupted=true
                reason="corrupted gzip file"
            fi
        fi

        # Check if encrypted file is valid (basic check)
        if [ "$is_corrupted" = "false" ] && [[ "$backup_file" == *.enc ]]; then
            # Check if file starts with "Salted__" (OpenSSL encrypted files)
            local file_header=$(head -c 8 "$backup_file" 2>/dev/null | od -A n -t c | tr -d ' \n' || echo "")
            if [[ "$file_header" != *"Salted__"* ]]; then
                is_corrupted=true
                reason="invalid encrypted file header"
            fi
        fi

        # Remove corrupted backup
        if [ "$is_corrupted" = "true" ]; then
            log "Removing corrupted backup: $(basename "$backup_file") ($reason)"
            rm -f "$backup_file" "${backup_file}.meta" 2>/dev/null || true
            ((corrupted_count++))
        fi

    done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -print0 2>/dev/null || true)

    if [ $corrupted_count -gt 0 ]; then
        log "Removed $corrupted_count corrupted backup files"
    else
        log "No corrupted backup files found"
    fi
}

# Clean orphaned metadata files
cleanup_orphaned_metadata() {
    log "Cleaning orphaned metadata files..."

    local orphaned_count=0

    while IFS= read -r -d '' meta_file; do
        local backup_file="${meta_file%.meta}"

        if [ ! -f "$backup_file" ]; then
            log "Removing orphaned metadata: $(basename "$meta_file")"
            rm -f "$meta_file" 2>/dev/null || true
            ((orphaned_count++))
        fi
    done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*.meta" -type f -print0 2>/dev/null || true)

    if [ $orphaned_count -gt 0 ]; then
        log "Removed $orphaned_count orphaned metadata files"
    else
        log "No orphaned metadata files found"
    fi
}

# Clean log files
cleanup_logs() {
    log "Cleaning old log files..."

    local log_retention_days=$((BACKUP_RETENTION_DAYS / 2)) # Keep logs for half the backup retention
    local deleted_count=0

    # Clean backup logs
    while IFS= read -r -d '' log_file; do
        log "Removing old log: $(basename "$log_file")"
        rm -f "$log_file" 2>/dev/null || true
        ((deleted_count++))
    done < <(find "$BACKUP_DIR" -name "*.log" -type f -mtime "+$log_retention_days" -print0 2>/dev/null || true)

    # Clean old report files
    while IFS= read -r -d '' report_file; do
        log "Removing old report: $(basename "$report_file")"
        rm -f "$report_file" 2>/dev/null || true
        ((deleted_count++))
    done < <(find "$BACKUP_DIR" -name "backup_report_*.txt" -type f -mtime "+$log_retention_days" -print0 2>/dev/null || true)

    if [ $deleted_count -gt 0 ]; then
        log "Removed $deleted_count old log and report files"
    fi
}

# Generate human-readable size
human_readable_size() {
    local bytes="$1"

    if [ $bytes -lt 1024 ]; then
        echo "${bytes}B"
    elif [ $bytes -lt $((1024 * 1024)) ]; then
        echo "$((bytes / 1024))KB"
    elif [ $bytes -lt $((1024 * 1024 * 1024)) ]; then
        echo "$((bytes / 1024 / 1024))MB"
    else
        echo "$((bytes / 1024 / 1024 / 1024))GB"
    fi
}

# Generate cleanup report
generate_cleanup_report() {
    local report_file="$BACKUP_DIR/cleanup_report_$(date +%Y%m%d).txt"

    {
        echo "Backup Cleanup Report"
        echo "===================="
        echo "Generated: $(date)"
        echo "Backup Directory: $BACKUP_DIR"
        echo "Retention Days: $BACKUP_RETENTION_DAYS"
        echo "Max Size Limit: ${MAX_BACKUP_SIZE_GB}GB"
        echo "Min Free Space: ${MIN_FREE_SPACE_GB}GB"
        echo ""

        echo "Current Status:"
        echo "==============="
        local backup_count=$(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" | wc -l)
        echo "Total backups: $backup_count"

        local total_size=0
        while IFS= read -r -d '' backup_file; do
            local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
            total_size=$((total_size + file_size))
        done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -print0 2>/dev/null || true)

        echo "Total backup size: $(human_readable_size $total_size)"

        local free_space_kb=$(df "$BACKUP_DIR" | awk 'NR==2 {print $4}')
        local free_space_bytes=$((free_space_kb * 1024))
        echo "Free space: $(human_readable_size $free_space_bytes)"
        echo ""

        echo "Recent Backups:"
        echo "==============="
        find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" | \
            sort -r | head -5 | while read -r backup; do
            local size=$(stat -f%z "$backup" 2>/dev/null || stat -c%s "$backup" 2>/dev/null || echo "unknown")
            local date=$(basename "$backup" | sed 's/radarr_backup_\([0-9]\{8\}_[0-9]\{6\}\).*/\1/' | sed 's/_/ /' | sed 's/\([0-9]\{4\}\)\([0-9]\{2\}\)\([0-9]\{2\}\) \([0-9]\{2\}\)\([0-9]\{2\}\)\([0-9]\{2\}\)/\1-\2-\3 \4:\5:\6/')
            echo "$date - $(basename "$backup") ($(human_readable_size $size))"
        done
        echo ""

        echo "Cleanup Log (last 10 lines):"
        echo "============================="
        tail -10 "$BACKUP_DIR/cleanup.log" 2>/dev/null || echo "No cleanup log available"

    } > "$report_file"

    log "Cleanup report generated: $report_file"
}

# Main execution
main() {
    log "Starting backup cleanup process..."

    # Create backup directory if it doesn't exist
    mkdir -p "$BACKUP_DIR"

    # Run all cleanup operations
    cleanup_corrupted_backups
    cleanup_orphaned_metadata
    cleanup_by_age
    cleanup_by_size
    cleanup_by_free_space
    cleanup_logs

    # Generate report
    generate_cleanup_report

    # Final status
    local final_count=$(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" | wc -l)
    local final_size=0

    while IFS= read -r -d '' backup_file; do
        local file_size=$(stat -f%z "$backup_file" 2>/dev/null || stat -c%s "$backup_file" 2>/dev/null || echo "0")
        final_size=$((final_size + file_size))
    done < <(find "$BACKUP_DIR" -name "radarr_backup_*.sql.gz*" -type f -not -name "*.meta" -print0 2>/dev/null || true)

    log "Cleanup completed. Remaining: $final_count backups ($(human_readable_size $final_size))"
}

# Handle command line arguments
case "${1:-all}" in
    "all")
        main
        ;;
    "age")
        cleanup_by_age
        ;;
    "size")
        cleanup_by_size
        ;;
    "space")
        cleanup_by_free_space
        ;;
    "corrupted")
        cleanup_corrupted_backups
        ;;
    "metadata")
        cleanup_orphaned_metadata
        ;;
    "logs")
        cleanup_logs
        ;;
    "report")
        generate_cleanup_report
        ;;
    *)
        echo "Usage: $0 {all|age|size|space|corrupted|metadata|logs|report}"
        echo ""
        echo "Commands:"
        echo "  all       - Run all cleanup operations (default)"
        echo "  age       - Remove backups older than retention period"
        echo "  size      - Remove oldest backups if over size limit"
        echo "  space     - Remove backups if free space is low"
        echo "  corrupted - Remove corrupted or incomplete backups"
        echo "  metadata  - Remove orphaned metadata files"
        echo "  logs      - Remove old log files"
        echo "  report    - Generate cleanup report"
        echo ""
        echo "Environment Variables:"
        echo "  BACKUP_DIR              - Backup directory (default: /backups)"
        echo "  BACKUP_RETENTION_DAYS   - Days to keep backups (default: 30)"
        echo "  MAX_BACKUP_SIZE_GB      - Maximum total backup size in GB (default: 50)"
        echo "  MIN_FREE_SPACE_GB       - Minimum free space required in GB (default: 5)"
        exit 1
        ;;
esac
