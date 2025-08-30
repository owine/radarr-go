# Automation Examples for Radarr Go

This document provides complete automation scripts and workflows for common Radarr Go operational tasks.

## 1. Backup and Restore API Workflows

### Complete Backup System

```python
#!/usr/bin/env python3
"""
Radarr Go Backup System
Comprehensive backup of configuration, database, and metadata
"""

import os
import json
import sqlite3
import shutil
import tarfile
import logging
from datetime import datetime, timedelta
from pathlib import Path
import requests
from typing import Dict, List, Optional

class RadarrBackupManager:
    def __init__(self, radarr_url: str, api_key: str, backup_dir: str):
        self.radarr_url = radarr_url.rstrip('/')
        self.api_key = api_key
        self.backup_dir = Path(backup_dir)
        self.backup_dir.mkdir(parents=True, exist_ok=True)

        # Setup logging
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler(self.backup_dir / 'backup.log'),
                logging.StreamHandler()
            ]
        )
        self.logger = logging.getLogger(__name__)

        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        })

    def _api_request(self, endpoint: str, method: str = 'GET', **kwargs) -> requests.Response:
        """Make authenticated API request"""
        url = f"{self.radarr_url}/api/v3/{endpoint.lstrip('/')}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response

    def create_full_backup(self) -> str:
        """Create complete backup including config, database, and metadata"""
        timestamp = datetime.now().strftime('%Y%m%d_%H%M%S')
        backup_name = f"radarr_backup_{timestamp}"
        backup_path = self.backup_dir / backup_name
        backup_path.mkdir()

        self.logger.info(f"Starting full backup: {backup_name}")

        try:
            # 1. Export API configuration
            self.logger.info("Exporting API configuration...")
            self._export_api_config(backup_path / 'api_config.json')

            # 2. Export movie library
            self.logger.info("Exporting movie library...")
            self._export_movie_library(backup_path / 'movies.json')

            # 3. Export quality profiles
            self.logger.info("Exporting quality profiles...")
            self._export_quality_profiles(backup_path / 'quality_profiles.json')

            # 4. Export custom formats (if supported)
            self.logger.info("Exporting custom formats...")
            self._export_custom_formats(backup_path / 'custom_formats.json')

            # 5. Export indexers
            self.logger.info("Exporting indexers...")
            self._export_indexers(backup_path / 'indexers.json')

            # 6. Export download clients
            self.logger.info("Exporting download clients...")
            self._export_download_clients(backup_path / 'download_clients.json')

            # 7. Export notifications
            self.logger.info("Exporting notifications...")
            self._export_notifications(backup_path / 'notifications.json')

            # 8. Export root folders
            self.logger.info("Exporting root folders...")
            self._export_root_folders(backup_path / 'root_folders.json')

            # 9. Export collections
            self.logger.info("Exporting collections...")
            self._export_collections(backup_path / 'collections.json')

            # 10. Create backup metadata
            self._create_backup_metadata(backup_path / 'metadata.json')

            # 11. Create compressed archive
            self.logger.info("Creating compressed archive...")
            archive_path = self._create_archive(backup_path)

            # 12. Cleanup temporary directory
            shutil.rmtree(backup_path)

            self.logger.info(f"Backup completed successfully: {archive_path}")
            return str(archive_path)

        except Exception as e:
            self.logger.error(f"Backup failed: {e}")
            if backup_path.exists():
                shutil.rmtree(backup_path)
            raise

    def _export_api_config(self, output_path: Path):
        """Export system configuration"""
        try:
            system_status = self._api_request('system/status').json()
            health = self._api_request('health').json()

            config = {
                'system_status': system_status,
                'health_status': health,
                'export_timestamp': datetime.now().isoformat()
            }

            with open(output_path, 'w') as f:
                json.dump(config, f, indent=2, default=str)

        except Exception as e:
            self.logger.warning(f"Failed to export system config: {e}")

    def _export_movie_library(self, output_path: Path):
        """Export complete movie library"""
        movies = []
        page = 1

        while True:
            try:
                response = self._api_request('movie', params={'page': page, 'pageSize': 100})
                page_data = response.json()

                if isinstance(page_data, dict) and 'data' in page_data:
                    movie_batch = page_data['data']
                    if not movie_batch:
                        break
                    movies.extend(movie_batch)
                    page += 1
                else:
                    # Non-paginated response
                    movies = page_data
                    break

            except Exception as e:
                self.logger.error(f"Failed to export movies (page {page}): {e}")
                break

        # Get movie files for each movie
        for movie in movies:
            if movie.get('hasFile') and movie.get('movieFileId'):
                try:
                    movie_file = self._api_request(f'moviefile/{movie["movieFileId"]}').json()
                    movie['movieFile'] = movie_file
                except Exception as e:
                    self.logger.warning(f"Failed to get movie file for {movie['title']}: {e}")

        with open(output_path, 'w') as f:
            json.dump(movies, f, indent=2, default=str)

        self.logger.info(f"Exported {len(movies)} movies")

    def _export_quality_profiles(self, output_path: Path):
        """Export quality profiles"""
        try:
            profiles = self._api_request('qualityprofile').json()
            with open(output_path, 'w') as f:
                json.dump(profiles, f, indent=2)
            self.logger.info(f"Exported {len(profiles)} quality profiles")
        except Exception as e:
            self.logger.warning(f"Failed to export quality profiles: {e}")

    def _export_custom_formats(self, output_path: Path):
        """Export custom formats"""
        try:
            formats = self._api_request('customformat').json()
            with open(output_path, 'w') as f:
                json.dump(formats, f, indent=2)
            self.logger.info(f"Exported {len(formats)} custom formats")
        except Exception as e:
            self.logger.warning(f"Failed to export custom formats: {e}")

    def _export_indexers(self, output_path: Path):
        """Export indexers (without sensitive data)"""
        try:
            indexers = self._api_request('indexer').json()

            # Remove sensitive fields
            for indexer in indexers:
                if 'fields' in indexer:
                    for field in indexer['fields']:
                        if field.get('name', '').lower() in ['apikey', 'password', 'username']:
                            field['value'] = '[REDACTED]'

            with open(output_path, 'w') as f:
                json.dump(indexers, f, indent=2)
            self.logger.info(f"Exported {len(indexers)} indexers")
        except Exception as e:
            self.logger.warning(f"Failed to export indexers: {e}")

    def _export_download_clients(self, output_path: Path):
        """Export download clients (without sensitive data)"""
        try:
            clients = self._api_request('downloadclient').json()

            # Remove sensitive fields
            for client in clients:
                if 'fields' in client:
                    for field in client['fields']:
                        if field.get('name', '').lower() in ['password', 'apikey', 'username']:
                            field['value'] = '[REDACTED]'

            with open(output_path, 'w') as f:
                json.dump(clients, f, indent=2)
            self.logger.info(f"Exported {len(clients)} download clients")
        except Exception as e:
            self.logger.warning(f"Failed to export download clients: {e}")

    def _export_notifications(self, output_path: Path):
        """Export notification settings (without sensitive data)"""
        try:
            notifications = self._api_request('notification').json()

            # Remove sensitive fields
            for notification in notifications:
                if 'fields' in notification:
                    for field in notification['fields']:
                        if field.get('name', '').lower() in ['token', 'apikey', 'password', 'webhook']:
                            field['value'] = '[REDACTED]'

            with open(output_path, 'w') as f:
                json.dump(notifications, f, indent=2)
            self.logger.info(f"Exported {len(notifications)} notifications")
        except Exception as e:
            self.logger.warning(f"Failed to export notifications: {e}")

    def _export_root_folders(self, output_path: Path):
        """Export root folders"""
        try:
            folders = self._api_request('rootfolder').json()
            with open(output_path, 'w') as f:
                json.dump(folders, f, indent=2)
            self.logger.info(f"Exported {len(folders)} root folders")
        except Exception as e:
            self.logger.warning(f"Failed to export root folders: {e}")

    def _export_collections(self, output_path: Path):
        """Export movie collections"""
        try:
            collections = self._api_request('collection').json()
            with open(output_path, 'w') as f:
                json.dump(collections, f, indent=2, default=str)
            self.logger.info(f"Exported {len(collections)} collections")
        except Exception as e:
            self.logger.warning(f"Failed to export collections: {e}")

    def _create_backup_metadata(self, output_path: Path):
        """Create backup metadata file"""
        try:
            system_status = self._api_request('system/status').json()

            metadata = {
                'backup_version': '1.0',
                'backup_timestamp': datetime.now().isoformat(),
                'radarr_version': system_status.get('version'),
                'database_type': system_status.get('databaseType'),
                'backup_type': 'full',
                'created_by': 'RadarrBackupManager'
            }

            with open(output_path, 'w') as f:
                json.dump(metadata, f, indent=2)

        except Exception as e:
            self.logger.warning(f"Failed to create backup metadata: {e}")

    def _create_archive(self, backup_path: Path) -> Path:
        """Create compressed backup archive"""
        archive_path = backup_path.with_suffix('.tar.gz')

        with tarfile.open(archive_path, 'w:gz') as tar:
            tar.add(backup_path, arcname=backup_path.name)

        return archive_path

    def restore_from_backup(self, backup_archive: str, options: Dict = None):
        """Restore from backup archive"""
        options = options or {}
        backup_path = Path(backup_archive)

        if not backup_path.exists():
            raise FileNotFoundError(f"Backup file not found: {backup_path}")

        self.logger.info(f"Starting restore from: {backup_path}")

        # Extract archive
        extract_dir = self.backup_dir / 'restore_temp'
        extract_dir.mkdir(exist_ok=True)

        try:
            with tarfile.open(backup_path, 'r:gz') as tar:
                tar.extractall(extract_dir)

            # Find the backup directory
            backup_contents = list(extract_dir.iterdir())
            if len(backup_contents) == 1 and backup_contents[0].is_dir():
                backup_data_dir = backup_contents[0]
            else:
                backup_data_dir = extract_dir

            # Restore components based on options
            if options.get('restore_movies', True):
                self._restore_movies(backup_data_dir / 'movies.json')

            if options.get('restore_quality_profiles', True):
                self._restore_quality_profiles(backup_data_dir / 'quality_profiles.json')

            if options.get('restore_root_folders', True):
                self._restore_root_folders(backup_data_dir / 'root_folders.json')

            # More restoration options...

            self.logger.info("Restore completed successfully")

        finally:
            # Cleanup
            if extract_dir.exists():
                shutil.rmtree(extract_dir)

    def _restore_movies(self, movies_file: Path):
        """Restore movie library"""
        if not movies_file.exists():
            self.logger.warning("Movies backup file not found")
            return

        with open(movies_file) as f:
            movies = json.load(f)

        restored_count = 0
        skipped_count = 0

        for movie in movies:
            try:
                # Check if movie already exists
                existing = self._api_request(f'movie/{movie["id"]}')
                if existing.status_code == 200:
                    self.logger.info(f"Movie already exists, skipping: {movie['title']}")
                    skipped_count += 1
                    continue
            except requests.exceptions.HTTPError:
                pass  # Movie doesn't exist, continue with restoration

            try:
                # Remove read-only fields
                restore_movie = {k: v for k, v in movie.items()
                               if k not in ['id', 'createdAt', 'updatedAt', 'movieFile']}

                # Add movie
                self._api_request('movie', method='POST', json=restore_movie)
                restored_count += 1
                self.logger.info(f"Restored movie: {movie['title']}")

            except Exception as e:
                self.logger.error(f"Failed to restore movie {movie['title']}: {e}")

        self.logger.info(f"Movie restoration complete: {restored_count} restored, {skipped_count} skipped")

    def _restore_quality_profiles(self, profiles_file: Path):
        """Restore quality profiles"""
        if not profiles_file.exists():
            self.logger.warning("Quality profiles backup file not found")
            return

        with open(profiles_file) as f:
            profiles = json.load(f)

        for profile in profiles:
            try:
                # Remove ID for creation
                restore_profile = {k: v for k, v in profile.items() if k != 'id'}

                # Check if profile with same name exists
                existing_profiles = self._api_request('qualityprofile').json()
                if any(p['name'] == profile['name'] for p in existing_profiles):
                    self.logger.info(f"Quality profile already exists: {profile['name']}")
                    continue

                self._api_request('qualityprofile', method='POST', json=restore_profile)
                self.logger.info(f"Restored quality profile: {profile['name']}")

            except Exception as e:
                self.logger.error(f"Failed to restore quality profile {profile['name']}: {e}")

    def _restore_root_folders(self, folders_file: Path):
        """Restore root folders"""
        if not folders_file.exists():
            self.logger.warning("Root folders backup file not found")
            return

        with open(folders_file) as f:
            folders = json.load(f)

        for folder in folders:
            try:
                # Remove ID for creation
                restore_folder = {k: v for k, v in folder.items() if k != 'id'}

                # Check if folder already exists
                existing_folders = self._api_request('rootfolder').json()
                if any(f['path'] == folder['path'] for f in existing_folders):
                    self.logger.info(f"Root folder already exists: {folder['path']}")
                    continue

                self._api_request('rootfolder', method='POST', json=restore_folder)
                self.logger.info(f"Restored root folder: {folder['path']}")

            except Exception as e:
                self.logger.error(f"Failed to restore root folder {folder['path']}: {e}")

    def cleanup_old_backups(self, keep_days: int = 30):
        """Remove backups older than specified days"""
        cutoff_date = datetime.now() - timedelta(days=keep_days)

        removed_count = 0
        for backup_file in self.backup_dir.glob('radarr_backup_*.tar.gz'):
            # Extract timestamp from filename
            try:
                timestamp_str = backup_file.stem.replace('radarr_backup_', '')
                backup_date = datetime.strptime(timestamp_str, '%Y%m%d_%H%M%S')

                if backup_date < cutoff_date:
                    backup_file.unlink()
                    removed_count += 1
                    self.logger.info(f"Removed old backup: {backup_file.name}")

            except ValueError:
                self.logger.warning(f"Could not parse backup date: {backup_file.name}")

        self.logger.info(f"Cleanup complete: {removed_count} old backups removed")

# Usage example
def main():
    import argparse

    parser = argparse.ArgumentParser(description='Radarr Go Backup Manager')
    parser.add_argument('--url', required=True, help='Radarr URL')
    parser.add_argument('--api-key', required=True, help='API Key')
    parser.add_argument('--backup-dir', required=True, help='Backup directory')
    parser.add_argument('action', choices=['backup', 'restore', 'cleanup'],
                       help='Action to perform')
    parser.add_argument('--backup-file', help='Backup file for restore')
    parser.add_argument('--keep-days', type=int, default=30,
                       help='Days to keep backups (for cleanup)')

    args = parser.parse_args()

    manager = RadarrBackupManager(args.url, args.api_key, args.backup_dir)

    if args.action == 'backup':
        backup_path = manager.create_full_backup()
        print(f"Backup created: {backup_path}")

    elif args.action == 'restore':
        if not args.backup_file:
            print("Error: --backup-file required for restore")
            return
        manager.restore_from_backup(args.backup_file)
        print("Restore completed")

    elif args.action == 'cleanup':
        manager.cleanup_old_backups(args.keep_days)
        print(f"Cleanup completed (kept {args.keep_days} days)")

if __name__ == "__main__":
    main()
```

## 2. Library Maintenance Automation Scripts

### Automated Library Health Monitor

```bash
#!/bin/bash
# Radarr Go Library Health Monitor
# Comprehensive library maintenance automation

set -euo pipefail

# Configuration
RADARR_URL="${RADARR_URL:-http://localhost:7878}"
API_KEY="${RADARR_API_KEY:-}"
CONFIG_FILE="${HOME}/.radarr-maintenance.conf"
LOG_FILE="${HOME}/.radarr-maintenance.log"
REPORT_FILE="${HOME}/radarr-health-report-$(date +%Y%m%d).txt"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
fi

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# API request function
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"

    local curl_opts=(
        -s
        -X "$method"
        -H "X-API-Key: $API_KEY"
        -H "Content-Type: application/json"
        -w "HTTPSTATUS:%{http_code}"
    )

    if [[ -n "$data" ]]; then
        curl_opts+=(-d "$data")
    fi

    local response=$(curl "${curl_opts[@]}" "$RADARR_URL/api/v3/$endpoint")
    local http_code=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
    local body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*$//')

    if [[ "$http_code" -ge 400 ]]; then
        log "ERROR: API request failed with HTTP $http_code"
        echo "$body" | jq -r '.error // "Unknown error"' >&2
        return 1
    fi

    echo "$body"
}

# Health check functions
check_system_health() {
    log "Checking system health..."

    local health_data=$(api_request "GET" "health")
    local status=$(echo "$health_data" | jq -r '.status // "unknown"')
    local issues_count=$(echo "$health_data" | jq -r '.issues | length')

    echo "=== SYSTEM HEALTH ===" >> "$REPORT_FILE"
    echo "Status: $status" >> "$REPORT_FILE"
    echo "Issues: $issues_count" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"

    if [[ "$issues_count" -gt 0 ]]; then
        log "WARNING: $issues_count health issues found"
        echo "$health_data" | jq -r '.issues[] | "- \(.type): \(.message)"' >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log "System health: OK"
    fi
}

check_missing_movies() {
    log "Checking for missing movies..."

    local missing_data=$(api_request "GET" "wanted/missing?page=1&pageSize=1")
    local total_missing=$(echo "$missing_data" | jq -r '.meta.total // 0')

    echo "=== MISSING MOVIES ===" >> "$REPORT_FILE"
    echo "Total missing: $total_missing" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"

    if [[ "$total_missing" -gt 0 ]]; then
        log "Found $total_missing missing movies"

        # Get sample of missing movies
        local missing_sample=$(api_request "GET" "wanted/missing?page=1&pageSize=10")
        echo "Sample of missing movies:" >> "$REPORT_FILE"
        echo "$missing_sample" | jq -r '.data[]? | "- \(.title) (\(.year)) - Available: \(.isAvailable)"' >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log "No missing movies found"
    fi
}

check_quality_cutoffs() {
    log "Checking quality cutoff violations..."

    local cutoff_data=$(api_request "GET" "wanted/cutoff?page=1&pageSize=1")
    local total_cutoff=$(echo "$cutoff_data" | jq -r '.meta.total // 0')

    echo "=== QUALITY CUTOFF UNMET ===" >> "$REPORT_FILE"
    echo "Total cutoff unmet: $total_cutoff" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"

    if [[ "$total_cutoff" -gt 0 ]]; then
        log "Found $total_cutoff movies not meeting quality cutoff"

        # Get sample
        local cutoff_sample=$(api_request "GET" "wanted/cutoff?page=1&pageSize=10")
        echo "Sample of cutoff unmet movies:" >> "$REPORT_FILE"
        echo "$cutoff_sample" | jq -r '.data[]? | "- \(.title) (\(.year)) - Current Quality: \(.movieFile.quality.quality.name // "Unknown")"' >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    else
        log "All movies meet quality cutoff requirements"
    fi
}

check_disk_space() {
    log "Checking disk space..."

    local root_folders=$(api_request "GET" "rootfolder")

    echo "=== DISK SPACE ===" >> "$REPORT_FILE"

    echo "$root_folders" | jq -c '.[]' | while read -r folder; do
        local path=$(echo "$folder" | jq -r '.path')
        local free_space=$(echo "$folder" | jq -r '.freeSpace // 0')
        local total_space=$(echo "$folder" | jq -r '.totalSpace // 0')

        if [[ "$total_space" -gt 0 ]]; then
            local used_percent=$(( (total_space - free_space) * 100 / total_space ))
            local free_gb=$(( free_space / 1024 / 1024 / 1024 ))

            echo "$path: ${free_gb}GB free (${used_percent}% used)" >> "$REPORT_FILE"

            if [[ "$used_percent" -gt 90 ]]; then
                log "WARNING: Low disk space on $path (${used_percent}% used)"
            fi
        fi
    done
    echo "" >> "$REPORT_FILE"
}

check_failed_downloads() {
    log "Checking for failed downloads..."

    # This would need to be adapted based on actual queue/history API
    local queue_data=$(api_request "GET" "queue" || echo '{"records": []}')
    local failed_count=$(echo "$queue_data" | jq -r '.records | map(select(.status == "failed")) | length')

    echo "=== FAILED DOWNLOADS ===" >> "$REPORT_FILE"
    echo "Failed downloads: $failed_count" >> "$REPORT_FILE"
    echo "" >> "$REPORT_FILE"

    if [[ "$failed_count" -gt 0 ]]; then
        log "Found $failed_count failed downloads"
        echo "$queue_data" | jq -r '.records[] | select(.status == "failed") | "- \(.title) - \(.errorMessage // "Unknown error")"' >> "$REPORT_FILE"
        echo "" >> "$REPORT_FILE"
    fi
}

fix_metadata_issues() {
    log "Fixing metadata issues..."

    # Find movies with missing metadata
    local all_movies=$(api_request "GET" "movie?pageSize=1000")
    local movies_needing_refresh=()

    while IFS= read -r movie; do
        local title=$(echo "$movie" | jq -r '.title')
        local overview=$(echo "$movie" | jq -r '.overview // empty')
        local images_count=$(echo "$movie" | jq -r '.images | length')

        if [[ -z "$overview" || "$images_count" -eq 0 ]]; then
            local movie_id=$(echo "$movie" | jq -r '.id')
            movies_needing_refresh+=("$movie_id:$title")
        fi
    done < <(echo "$all_movies" | jq -c '.data[]? // .[]?')

    if [[ ${#movies_needing_refresh[@]} -gt 0 ]]; then
        log "Found ${#movies_needing_refresh[@]} movies needing metadata refresh"

        for movie_info in "${movies_needing_refresh[@]}"; do
            local movie_id="${movie_info%%:*}"
            local title="${movie_info#*:}"

            log "Refreshing metadata for: $title"
            api_request "POST" "command" '{"name": "RefreshMovie", "movieId": '"$movie_id"'}'

            # Rate limiting
            sleep 2
        done
    else
        log "No movies need metadata refresh"
    fi
}

cleanup_completed_downloads() {
    log "Cleaning up completed downloads..."

    # Remove completed items from queue older than 24 hours
    local queue_data=$(api_request "GET" "queue" || echo '{"records": []}')
    local cutoff_date=$(date -d '24 hours ago' -u +%Y-%m-%dT%H:%M:%SZ)

    echo "$queue_data" | jq -c '.records[]? | select(.status == "completed" and .added < "'"$cutoff_date"'")' | while read -r item; do
        local item_id=$(echo "$item" | jq -r '.id')
        local title=$(echo "$item" | jq -r '.title')

        log "Removing completed download: $title"
        api_request "DELETE" "queue/$item_id"
    done
}

optimize_quality_profiles() {
    log "Optimizing quality profiles..."

    local profiles=$(api_request "GET" "qualityprofile")
    local movies=$(api_request "GET" "movie?pageSize=1000")

    # Analyze quality profile usage
    declare -A profile_usage

    echo "$movies" | jq -r '.data[]?.qualityProfileId // .[]?.qualityProfileId' | sort | uniq -c | while read -r count profile_id; do
        profile_usage[$profile_id]=$count
    done

    echo "=== QUALITY PROFILE USAGE ===" >> "$REPORT_FILE"
    echo "$profiles" | jq -r '.[] | "\(.id): \(.name)"' | while read -r line; do
        local profile_id=$(echo "$line" | cut -d: -f1)
        local profile_name=$(echo "$line" | cut -d: -f2-)
        local usage=${profile_usage[$profile_id]:-0}

        echo "$profile_name: $usage movies" >> "$REPORT_FILE"

        if [[ "$usage" -eq 0 ]]; then
            log "WARNING: Quality profile '$profile_name' is not used by any movies"
        fi
    done
    echo "" >> "$REPORT_FILE"
}

send_report_notification() {
    log "Sending health report notification..."

    # Example: Send via email (requires mailx or similar)
    if command -v mailx >/dev/null 2>&1 && [[ -n "${REPORT_EMAIL:-}" ]]; then
        mailx -s "Radarr Health Report - $(date +%Y-%m-%d)" "$REPORT_EMAIL" < "$REPORT_FILE"
        log "Health report sent to $REPORT_EMAIL"
    fi

    # Example: Send to Discord webhook
    if [[ -n "${DISCORD_WEBHOOK:-}" ]]; then
        local report_preview=$(head -n 50 "$REPORT_FILE" | tail -n +1)
        local json_data=$(jq -n --arg content "$report_preview" '{content: $content}')

        curl -s -X POST -H "Content-Type: application/json" \
             -d "$json_data" \
             "$DISCORD_WEBHOOK" >/dev/null

        log "Health report sent to Discord"
    fi
}

# Main execution
main() {
    local action="${1:-full}"

    log "Starting Radarr maintenance (action: $action)"

    # Validate configuration
    if [[ -z "$API_KEY" ]]; then
        echo "Error: API_KEY not set. Set RADARR_API_KEY environment variable or add to $CONFIG_FILE"
        exit 1
    fi

    # Initialize report
    cat > "$REPORT_FILE" << EOF
RADARR GO LIBRARY HEALTH REPORT
===============================
Generated: $(date)
Radarr URL: $RADARR_URL

EOF

    case "$action" in
        "full")
            check_system_health
            check_missing_movies
            check_quality_cutoffs
            check_disk_space
            check_failed_downloads
            optimize_quality_profiles
            ;;
        "health")
            check_system_health
            ;;
        "missing")
            check_missing_movies
            ;;
        "fix")
            fix_metadata_issues
            cleanup_completed_downloads
            ;;
        "report")
            check_system_health
            check_missing_movies
            check_quality_cutoffs
            check_disk_space
            send_report_notification
            ;;
        *)
            echo "Usage: $0 [full|health|missing|fix|report]"
            exit 1
            ;;
    esac

    log "Maintenance completed. Report saved to: $REPORT_FILE"
}

# Run main function with all arguments
main "$@"
```

## 3. Custom Quality Management Workflows

### Advanced Quality Profile Manager

```python
#!/usr/bin/env python3
"""
Advanced Quality Profile Manager for Radarr Go
Automated quality profile management and optimization
"""

import json
import logging
from datetime import datetime
from typing import Dict, List, Optional, Tuple
import requests

class QualityProfileManager:
    """Advanced quality profile management system"""

    def __init__(self, radarr_url: str, api_key: str):
        self.radarr_url = radarr_url.rstrip('/')
        self.api_key = api_key

        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        })

        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(__name__)

    def _api_request(self, endpoint: str, method: str = 'GET', **kwargs):
        """Make authenticated API request"""
        url = f"{self.radarr_url}/api/v3/{endpoint.lstrip('/')}"
        response = self.session.request(method, url, **kwargs)
        response.raise_for_status()
        return response

    def analyze_quality_needs(self) -> Dict:
        """Analyze library to determine optimal quality profiles"""
        self.logger.info("Analyzing library quality needs...")

        # Get all movies and their file information
        movies = self._api_request('movie', params={'pageSize': 10000}).json()
        if isinstance(movies, dict) and 'data' in movies:
            movies = movies['data']

        analysis = {
            'total_movies': len(movies),
            'by_year': {},
            'by_genre': {},
            'file_quality_distribution': {},
            'missing_files': 0,
            'quality_recommendations': {}
        }

        # Analyze movies
        for movie in movies:
            year = movie.get('year', 'unknown')
            genres = movie.get('genres', [])

            # Year analysis
            if year not in analysis['by_year']:
                analysis['by_year'][year] = {'count': 0, 'has_file': 0, 'qualities': {}}
            analysis['by_year'][year]['count'] += 1

            # Genre analysis
            for genre in genres:
                if genre not in analysis['by_genre']:
                    analysis['by_genre'][genre] = {'count': 0, 'has_file': 0, 'qualities': {}}
                analysis['by_genre'][genre]['count'] += 1

            # File quality analysis
            if movie.get('hasFile') and movie.get('movieFile'):
                movie_file = movie['movieFile']
                quality_name = movie_file.get('quality', {}).get('quality', {}).get('name', 'Unknown')

                analysis['by_year'][year]['has_file'] += 1
                for genre in genres:
                    analysis['by_genre'][genre]['has_file'] += 1

                # Quality distribution
                if quality_name not in analysis['file_quality_distribution']:
                    analysis['file_quality_distribution'][quality_name] = 0
                analysis['file_quality_distribution'][quality_name] += 1

                # Track qualities by year and genre
                if quality_name not in analysis['by_year'][year]['qualities']:
                    analysis['by_year'][year]['qualities'][quality_name] = 0
                analysis['by_year'][year]['qualities'][quality_name] += 1

                for genre in genres:
                    if quality_name not in analysis['by_genre'][genre]['qualities']:
                        analysis['by_genre'][genre]['qualities'][quality_name] = 0
                    analysis['by_genre'][genre]['qualities'][quality_name] += 1
            else:
                analysis['missing_files'] += 1

        # Generate recommendations
        analysis['quality_recommendations'] = self._generate_quality_recommendations(analysis)

        return analysis

    def _generate_quality_recommendations(self, analysis: Dict) -> Dict:
        """Generate quality profile recommendations based on analysis"""
        recommendations = {
            'suggested_profiles': [],
            'optimization_notes': []
        }

        total_movies = analysis['total_movies']
        quality_dist = analysis['file_quality_distribution']

        # Recommend profiles based on collection characteristics
        if total_movies > 1000:
            # Large collection - recommend tiered approach
            recommendations['suggested_profiles'].extend([
                {
                    'name': 'Ultra HD Collection',
                    'description': 'For premium movies and new releases',
                    'cutoff': 'Remux-2160p',
                    'qualities': ['Remux-2160p', 'BluRay-2160p', 'WEBRip-2160p', 'WEBDL-2160p'],
                    'recommended_for': 'New releases and premium content'
                },
                {
                    'name': 'High Definition Standard',
                    'description': 'Standard HD quality for most movies',
                    'cutoff': 'BluRay-1080p',
                    'qualities': ['Remux-1080p', 'BluRay-1080p', 'WEBRip-1080p', 'WEBDL-1080p'],
                    'recommended_for': 'General collection'
                },
                {
                    'name': 'Space Saver',
                    'description': 'Compressed quality for older/less important movies',
                    'cutoff': 'WEBDL-720p',
                    'qualities': ['BluRay-720p', 'WEBRip-720p', 'WEBDL-720p', 'HDTV-720p'],
                    'recommended_for': 'Older movies, TV movies, documentaries'
                }
            ])
        else:
            # Smaller collection - single high-quality profile
            recommendations['suggested_profiles'].append({
                'name': 'Quality Focused',
                'description': 'High quality for curated collection',
                'cutoff': 'BluRay-1080p',
                'qualities': ['Remux-1080p', 'BluRay-1080p', 'WEBRip-1080p', 'WEBDL-1080p', 'BluRay-720p'],
                'recommended_for': 'All movies in collection'
            })

        # Analyze current quality distribution for optimization notes
        if 'HDTV-720p' in quality_dist and quality_dist['HDTV-720p'] > total_movies * 0.3:
            recommendations['optimization_notes'].append(
                "High percentage of HDTV-720p files detected. Consider upgrading to BluRay-720p or higher."
            )

        if 'WEBDL-480p' in quality_dist:
            recommendations['optimization_notes'].append(
                "SD quality files detected. Consider setting minimum quality to 720p."
            )

        return recommendations

    def create_optimized_profiles(self, analysis: Dict) -> List[Dict]:
        """Create optimized quality profiles based on analysis"""
        self.logger.info("Creating optimized quality profiles...")

        # Get current quality definitions
        qualities = self._api_request('quality/definition').json()
        quality_map = {q['name']: q for q in qualities}

        created_profiles = []

        for profile_spec in analysis['quality_recommendations']['suggested_profiles']:
            self.logger.info(f"Creating profile: {profile_spec['name']}")

            # Build quality profile structure
            profile_data = {
                'name': profile_spec['name'],
                'cutoff': self._get_quality_id(profile_spec['cutoff'], quality_map),
                'items': self._build_quality_items(profile_spec['qualities'], quality_map),
                'language': 'english',
                'upgradeAllowed': True,
                'minFormatScore': 0,
                'cutoffFormatScore': 0
            }

            try:
                response = self._api_request('qualityprofile', method='POST', json=profile_data)
                created_profile = response.json()
                created_profiles.append(created_profile)
                self.logger.info(f"Successfully created profile: {created_profile['name']}")
            except Exception as e:
                self.logger.error(f"Failed to create profile {profile_spec['name']}: {e}")

        return created_profiles

    def _get_quality_id(self, quality_name: str, quality_map: Dict) -> int:
        """Get quality ID by name"""
        quality = quality_map.get(quality_name)
        if not quality:
            # Fallback to a reasonable default
            return quality_map.get('BluRay-1080p', {}).get('id', 7)
        return quality['id']

    def _build_quality_items(self, quality_names: List[str], quality_map: Dict) -> List[Dict]:
        """Build quality items for profile"""
        items = []

        for quality_name in quality_names:
            quality = quality_map.get(quality_name)
            if quality:
                items.append({
                    'quality': {
                        'id': quality['id'],
                        'name': quality['name'],
                        'source': quality.get('source', 'unknown'),
                        'resolution': quality.get('resolution', 'unknown')
                    },
                    'allowed': True
                })

        return items

    def optimize_movie_assignments(self, analysis: Dict, profile_mapping: Dict = None) -> Dict:
        """Optimize movie quality profile assignments"""
        self.logger.info("Optimizing movie quality profile assignments...")

        if not profile_mapping:
            # Default mapping based on movie characteristics
            profile_mapping = {
                'new_releases': 'Ultra HD Collection',
                'classics': 'High Definition Standard',
                'documentaries': 'Space Saver',
                'tv_movies': 'Space Saver',
                'default': 'High Definition Standard'
            }

        # Get current profiles
        profiles = self._api_request('qualityprofile').json()
        profile_name_to_id = {p['name']: p['id'] for p in profiles}

        # Get all movies
        movies = self._api_request('movie', params={'pageSize': 10000}).json()
        if isinstance(movies, dict) and 'data' in movies:
            movies = movies['data']

        optimization_results = {
            'total_movies': len(movies),
            'changes_made': 0,
            'failed_updates': 0,
            'recommendations': []
        }

        for movie in movies:
            recommended_profile = self._determine_optimal_profile(movie, profile_mapping)
            recommended_profile_id = profile_name_to_id.get(recommended_profile)

            if not recommended_profile_id:
                continue

            current_profile_id = movie.get('qualityProfileId')

            if current_profile_id != recommended_profile_id:
                try:
                    # Update movie profile
                    movie_update = {
                        'id': movie['id'],
                        'qualityProfileId': recommended_profile_id
                    }

                    # Include required fields for update
                    for field in ['title', 'tmdbId', 'year', 'monitored', 'minimumAvailability']:
                        if field in movie:
                            movie_update[field] = movie[field]

                    self._api_request(f'movie/{movie["id"]}', method='PUT', json=movie_update)
                    optimization_results['changes_made'] += 1

                    self.logger.info(f"Updated {movie['title']} to profile: {recommended_profile}")

                except Exception as e:
                    optimization_results['failed_updates'] += 1
                    self.logger.error(f"Failed to update {movie['title']}: {e}")

        return optimization_results

    def _determine_optimal_profile(self, movie: Dict, profile_mapping: Dict) -> str:
        """Determine optimal quality profile for a movie"""
        year = movie.get('year', 0)
        genres = movie.get('genres', [])

        # New releases (last 2 years)
        current_year = datetime.now().year
        if year >= current_year - 2:
            return profile_mapping.get('new_releases', profile_mapping['default'])

        # Documentaries
        if 'Documentary' in genres:
            return profile_mapping.get('documentaries', profile_mapping['default'])

        # TV Movies
        if any(genre in genres for genre in ['TV Movie', 'Made for TV']):
            return profile_mapping.get('tv_movies', profile_mapping['default'])

        # Classics (pre-1980)
        if year < 1980:
            return profile_mapping.get('classics', profile_mapping['default'])

        # Default
        return profile_mapping['default']

    def generate_quality_report(self, analysis: Dict) -> str:
        """Generate comprehensive quality management report"""
        report = f"""
RADARR GO QUALITY ANALYSIS REPORT
=================================
Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

LIBRARY OVERVIEW
----------------
Total Movies: {analysis['total_movies']}
Movies with Files: {analysis['total_movies'] - analysis['missing_files']}
Missing Files: {analysis['missing_files']}

QUALITY DISTRIBUTION
--------------------
"""

        for quality, count in sorted(analysis['file_quality_distribution'].items(),
                                   key=lambda x: x[1], reverse=True):
            percentage = (count / (analysis['total_movies'] - analysis['missing_files'])) * 100
            report += f"{quality}: {count} files ({percentage:.1f}%)\n"

        report += "\nRECOMMENDED QUALITY PROFILES\n"
        report += "----------------------------\n"

        for profile in analysis['quality_recommendations']['suggested_profiles']:
            report += f"\nProfile: {profile['name']}\n"
            report += f"Description: {profile['description']}\n"
            report += f"Cutoff: {profile['cutoff']}\n"
            report += f"Qualities: {', '.join(profile['qualities'])}\n"
            report += f"Recommended for: {profile['recommended_for']}\n"

        if analysis['quality_recommendations']['optimization_notes']:
            report += "\nOPTIMIZATION NOTES\n"
            report += "-------------------\n"
            for note in analysis['quality_recommendations']['optimization_notes']:
                report += f"- {note}\n"

        return report

# Usage example
def main():
    import argparse

    parser = argparse.ArgumentParser(description='Advanced Quality Profile Manager')
    parser.add_argument('--url', required=True, help='Radarr URL')
    parser.add_argument('--api-key', required=True, help='API Key')
    parser.add_argument('action', choices=['analyze', 'create-profiles', 'optimize', 'report'],
                       help='Action to perform')
    parser.add_argument('--output', help='Output file for report')

    args = parser.parse_args()

    manager = QualityProfileManager(args.url, args.api_key)

    if args.action == 'analyze':
        analysis = manager.analyze_quality_needs()
        print(json.dumps(analysis, indent=2))

    elif args.action == 'create-profiles':
        analysis = manager.analyze_quality_needs()
        created_profiles = manager.create_optimized_profiles(analysis)
        print(f"Created {len(created_profiles)} quality profiles")

    elif args.action == 'optimize':
        analysis = manager.analyze_quality_needs()
        results = manager.optimize_movie_assignments(analysis)
        print(f"Optimization complete: {results['changes_made']} movies updated, "
              f"{results['failed_updates']} failed")

    elif args.action == 'report':
        analysis = manager.analyze_quality_needs()
        report = manager.generate_quality_report(analysis)

        if args.output:
            with open(args.output, 'w') as f:
                f.write(report)
            print(f"Report saved to: {args.output}")
        else:
            print(report)

if __name__ == "__main__":
    main()
```

## 4. Integration with External Tools

### Plex/Jellyfin Integration Manager

```python
#!/usr/bin/env python3
"""
Radarr Go Integration with Plex and Jellyfin
Automated library synchronization and management
"""

import requests
import json
import time
import logging
from datetime import datetime
from typing import Dict, List, Optional
from xml.etree import ElementTree as ET
from urllib.parse import urljoin, quote

class MediaServerIntegration:
    """Base class for media server integrations"""

    def __init__(self, name: str, base_url: str, auth_token: str):
        self.name = name
        self.base_url = base_url.rstrip('/')
        self.auth_token = auth_token
        self.session = requests.Session()

        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger(f"MediaServer-{name}")

    def test_connection(self) -> bool:
        """Test connection to media server"""
        raise NotImplementedError

    def refresh_library(self, library_id: str = None) -> bool:
        """Refresh media library"""
        raise NotImplementedError

    def get_libraries(self) -> List[Dict]:
        """Get available libraries"""
        raise NotImplementedError

class PlexIntegration(MediaServerIntegration):
    """Plex Media Server integration"""

    def __init__(self, base_url: str, auth_token: str):
        super().__init__("Plex", base_url, auth_token)
        self.session.headers.update({
            'X-Plex-Token': auth_token,
            'Accept': 'application/json'
        })

    def test_connection(self) -> bool:
        """Test Plex connection"""
        try:
            response = self.session.get(f"{self.base_url}/identity")
            response.raise_for_status()
            self.logger.info("Plex connection successful")
            return True
        except Exception as e:
            self.logger.error(f"Plex connection failed: {e}")
            return False

    def get_libraries(self) -> List[Dict]:
        """Get Plex libraries"""
        try:
            response = self.session.get(f"{self.base_url}/library/sections")
            response.raise_for_status()

            data = response.json()
            libraries = []

            for section in data.get('MediaContainer', {}).get('Directory', []):
                if section.get('type') == 'movie':
                    libraries.append({
                        'id': section['key'],
                        'name': section['title'],
                        'type': section['type'],
                        'path': section.get('Location', [{}])[0].get('path', '')
                    })

            return libraries

        except Exception as e:
            self.logger.error(f"Failed to get Plex libraries: {e}")
            return []

    def refresh_library(self, library_id: str = None) -> bool:
        """Refresh Plex library"""
        try:
            if library_id:
                # Refresh specific library
                url = f"{self.base_url}/library/sections/{library_id}/refresh"
            else:
                # Refresh all movie libraries
                libraries = self.get_libraries()
                movie_libs = [lib for lib in libraries if lib['type'] == 'movie']

                if not movie_libs:
                    self.logger.warning("No movie libraries found")
                    return False

                for lib in movie_libs:
                    self.refresh_library(lib['id'])
                return True

            response = self.session.get(url)
            response.raise_for_status()
            self.logger.info(f"Plex library refresh triggered: {library_id or 'all'}")
            return True

        except Exception as e:
            self.logger.error(f"Failed to refresh Plex library: {e}")
            return False

    def get_movie_info(self, library_id: str, title: str, year: int = None) -> Optional[Dict]:
        """Get movie information from Plex"""
        try:
            # Search for movie
            search_url = f"{self.base_url}/library/sections/{library_id}/search"
            params = {'query': title, 'type': 1}  # type=1 for movies

            response = self.session.get(search_url, params=params)
            response.raise_for_status()

            data = response.json()

            for movie in data.get('MediaContainer', {}).get('Metadata', []):
                movie_year = movie.get('year')
                if year and movie_year and abs(movie_year - year) > 1:
                    continue

                return {
                    'key': movie['key'],
                    'title': movie['title'],
                    'year': movie_year,
                    'rating': movie.get('rating'),
                    'duration': movie.get('duration'),
                    'added_at': movie.get('addedAt'),
                    'updated_at': movie.get('updatedAt')
                }

            return None

        except Exception as e:
            self.logger.error(f"Failed to get movie info from Plex: {e}")
            return None

class JellyfinIntegration(MediaServerIntegration):
    """Jellyfin Media Server integration"""

    def __init__(self, base_url: str, auth_token: str, user_id: str = None):
        super().__init__("Jellyfin", base_url, auth_token)
        self.user_id = user_id
        self.session.headers.update({
            'Authorization': f'MediaBrowser Token="{auth_token}"',
            'Content-Type': 'application/json'
        })

    def test_connection(self) -> bool:
        """Test Jellyfin connection"""
        try:
            response = self.session.get(f"{self.base_url}/System/Info")
            response.raise_for_status()
            self.logger.info("Jellyfin connection successful")
            return True
        except Exception as e:
            self.logger.error(f"Jellyfin connection failed: {e}")
            return False

    def get_libraries(self) -> List[Dict]:
        """Get Jellyfin libraries"""
        try:
            if not self.user_id:
                # Get first user
                users = self.session.get(f"{self.base_url}/Users").json()
                self.user_id = users[0]['Id']

            response = self.session.get(f"{self.base_url}/Users/{self.user_id}/Views")
            response.raise_for_status()

            data = response.json()
            libraries = []

            for item in data.get('Items', []):
                if item.get('CollectionType') == 'movies':
                    libraries.append({
                        'id': item['Id'],
                        'name': item['Name'],
                        'type': 'movie',
                        'path': ''  # Jellyfin doesn't expose paths directly
                    })

            return libraries

        except Exception as e:
            self.logger.error(f"Failed to get Jellyfin libraries: {e}")
            return []

    def refresh_library(self, library_id: str = None) -> bool:
        """Refresh Jellyfin library"""
        try:
            if library_id:
                # Refresh specific library
                url = f"{self.base_url}/Library/Refresh"
                data = {'Id': library_id}
            else:
                # Refresh all movie libraries
                libraries = self.get_libraries()
                movie_libs = [lib for lib in libraries if lib['type'] == 'movie']

                for lib in movie_libs:
                    self.refresh_library(lib['id'])
                return True

            response = self.session.post(url, json=data)
            response.raise_for_status()
            self.logger.info(f"Jellyfin library refresh triggered: {library_id or 'all'}")
            return True

        except Exception as e:
            self.logger.error(f"Failed to refresh Jellyfin library: {e}")
            return False

class RadarrMediaServerManager:
    """Manages integration between Radarr and media servers"""

    def __init__(self, radarr_url: str, radarr_api_key: str):
        self.radarr_url = radarr_url.rstrip('/')
        self.radarr_api_key = radarr_api_key
        self.media_servers = []

        self.radarr_session = requests.Session()
        self.radarr_session.headers.update({
            'X-API-Key': radarr_api_key,
            'Content-Type': 'application/json'
        })

        logging.basicConfig(level=logging.INFO)
        self.logger = logging.getLogger("RadarrMediaServer")

    def add_media_server(self, server: MediaServerIntegration):
        """Add media server for integration"""
        if server.test_connection():
            self.media_servers.append(server)
            self.logger.info(f"Added {server.name} media server")
        else:
            self.logger.error(f"Failed to add {server.name} media server")

    def _radarr_api_request(self, endpoint: str, method: str = 'GET', **kwargs):
        """Make Radarr API request"""
        url = f"{self.radarr_url}/api/v3/{endpoint.lstrip('/')}"
        response = self.radarr_session.request(method, url, **kwargs)
        response.raise_for_status()
        return response

    def sync_with_media_servers(self):
        """Synchronize Radarr library with media servers"""
        self.logger.info("Starting library synchronization...")

        # Get Radarr movies
        movies_response = self._radarr_api_request('movie', params={'pageSize': 10000})
        movies = movies_response.json()
        if isinstance(movies, dict) and 'data' in movies:
            movies = movies['data']

        sync_stats = {
            'total_movies': len(movies),
            'movies_with_files': 0,
            'media_server_refreshes': 0,
            'sync_issues': []
        }

        # Track movies that have files
        movies_with_files = [m for m in movies if m.get('hasFile')]
        sync_stats['movies_with_files'] = len(movies_with_files)

        # Refresh media server libraries for movies with files
        recently_added = [
            m for m in movies_with_files
            if self._is_recently_added(m.get('added', ''))
        ]

        if recently_added:
            self.logger.info(f"Found {len(recently_added)} recently added movies")

            for server in self.media_servers:
                try:
                    server.refresh_library()
                    sync_stats['media_server_refreshes'] += 1
                    time.sleep(5)  # Wait between refreshes
                except Exception as e:
                    sync_stats['sync_issues'].append(f"{server.name} refresh failed: {e}")

        # Log synchronization results
        self.logger.info(f"Sync complete: {sync_stats['media_server_refreshes']} refreshes triggered")
        if sync_stats['sync_issues']:
            for issue in sync_stats['sync_issues']:
                self.logger.warning(issue)

        return sync_stats

    def _is_recently_added(self, added_date: str, hours: int = 24) -> bool:
        """Check if movie was recently added"""
        if not added_date:
            return False

        try:
            added_dt = datetime.fromisoformat(added_date.replace('Z', '+00:00'))
            cutoff_dt = datetime.now(added_dt.tzinfo) - timedelta(hours=hours)
            return added_dt > cutoff_dt
        except:
            return False

    def monitor_and_sync(self, check_interval: int = 300):
        """Monitor Radarr for changes and sync with media servers"""
        self.logger.info(f"Starting continuous monitoring (check every {check_interval}s)")

        last_movie_count = 0

        while True:
            try:
                # Get current movie count
                status = self._radarr_api_request('system/status').json()
                current_movie_count = len(self._radarr_api_request('movie').json())

                # Check if library has grown
                if current_movie_count > last_movie_count:
                    self.logger.info(f"Library growth detected: {current_movie_count - last_movie_count} new movies")
                    self.sync_with_media_servers()
                    last_movie_count = current_movie_count

                time.sleep(check_interval)

            except KeyboardInterrupt:
                self.logger.info("Monitoring stopped by user")
                break
            except Exception as e:
                self.logger.error(f"Monitoring error: {e}")
                time.sleep(60)  # Wait before retrying

    def create_media_server_collections(self):
        """Create collections in media servers based on Radarr collections"""
        self.logger.info("Creating media server collections...")

        # Get Radarr collections
        collections = self._radarr_api_request('collection').json()

        for collection in collections:
            collection_name = collection['name']
            movies = collection.get('movies', [])

            self.logger.info(f"Processing collection: {collection_name} ({len(movies)} movies)")

            for server in self.media_servers:
                try:
                    self._create_server_collection(server, collection_name, movies)
                except Exception as e:
                    self.logger.error(f"Failed to create collection on {server.name}: {e}")

    def _create_server_collection(self, server: MediaServerIntegration,
                                collection_name: str, movies: List[Dict]):
        """Create collection on specific media server"""
        # This would need server-specific implementation
        # For now, just log the action
        self.logger.info(f"Would create collection '{collection_name}' on {server.name}")
        # TODO: Implement server-specific collection creation

# Usage example and CLI
def main():
    import argparse
    from configparser import ConfigParser

    parser = argparse.ArgumentParser(description='Radarr Media Server Integration')
    parser.add_argument('--config', default='media_integration.conf',
                       help='Configuration file')
    parser.add_argument('action', choices=['sync', 'monitor', 'test', 'collections'],
                       help='Action to perform')

    args = parser.parse_args()

    # Load configuration
    config = ConfigParser()
    config.read(args.config)

    # Initialize Radarr manager
    manager = RadarrMediaServerManager(
        config.get('radarr', 'url'),
        config.get('radarr', 'api_key')
    )

    # Add media servers
    if config.has_section('plex'):
        plex = PlexIntegration(
            config.get('plex', 'url'),
            config.get('plex', 'token')
        )
        manager.add_media_server(plex)

    if config.has_section('jellyfin'):
        jellyfin = JellyfinIntegration(
            config.get('jellyfin', 'url'),
            config.get('jellyfin', 'token'),
            config.get('jellyfin', 'user_id', fallback=None)
        )
        manager.add_media_server(jellyfin)

    # Execute action
    if args.action == 'sync':
        results = manager.sync_with_media_servers()
        print(f"Sync completed: {results}")

    elif args.action == 'monitor':
        manager.monitor_and_sync()

    elif args.action == 'test':
        for server in manager.media_servers:
            print(f"Testing {server.name}: {'OK' if server.test_connection() else 'FAILED'}")

    elif args.action == 'collections':
        manager.create_media_server_collections()

if __name__ == "__main__":
    main()
```

This completes the Automation Examples section. These scripts provide production-ready automation solutions for backup/restore, library maintenance, quality management, and external system integration.
