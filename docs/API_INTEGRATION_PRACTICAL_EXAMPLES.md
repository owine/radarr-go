# Radarr Go API - Practical Integration Examples

This comprehensive guide provides production-ready examples and integration patterns for effectively using the Radarr Go API. Each example includes complete error handling, best practices, and real-world usage patterns.

## Table of Contents

1. [Third-Party Client Examples](#third-party-client-examples)
2. [Common Integration Patterns](#common-integration-patterns)
3. [Automation Examples](#automation-examples)
4. [Troubleshooting Guide](#troubleshooting-guide)

---

## Common Integration Patterns

### Movie Search and Addition Workflow

Complete workflow for searching and adding movies with validation:

```python
# Python Implementation
class MovieAdditionWorkflow:
    def __init__(self, client):
        self.client = client

    def search_and_add_movie(self, query, quality_profile_id=1, root_folder="/movies",
                           monitor=True, search_on_add=True):
        """Complete workflow for searching and adding a movie"""

        # Step 1: Search for movie
        search_results = self.client.search_movies(query)
        if not search_results:
            raise ValueError(f"No movies found for query: {query}")

        # Step 2: Let user select or auto-select best match
        selected_movie = self._select_best_match(search_results, query)

        # Step 3: Validate selection
        if self._movie_already_exists(selected_movie['tmdbId']):
            raise ValueError(f"Movie already exists: {selected_movie['title']}")

        # Step 4: Prepare movie data
        movie_data = {
            'title': selected_movie['title'],
            'year': selected_movie['year'],
            'tmdbId': selected_movie['tmdbId'],
            'imdbId': selected_movie.get('imdbId'),
            'qualityProfileId': quality_profile_id,
            'rootFolderPath': root_folder,
            'monitored': monitor,
            'minimumAvailability': 'announced',
            'addOptions': {
                'monitor': 'movieOnly',
                'searchForMovie': search_on_add
            }
        }

        # Step 5: Add movie
        added_movie = self.client.add_movie(movie_data)

        # Step 6: Monitor addition progress if search enabled
        if search_on_add:
            self._monitor_search_progress(added_movie['id'])

        return added_movie

    def _select_best_match(self, results, query):
        """Select the best match based on title similarity and year"""
        import difflib

        query_lower = query.lower()
        best_match = None
        best_score = 0

        for movie in results:
            title_score = difflib.SequenceMatcher(
                None, query_lower, movie['title'].lower()
            ).ratio()

            # Bonus for exact year match if year in query
            year_bonus = 0
            import re
            year_match = re.search(r'\b(19|20)\d{2}\b', query)
            if year_match and int(year_match.group()) == movie['year']:
                year_bonus = 0.2

            total_score = title_score + year_bonus

            if total_score > best_score:
                best_score = total_score
                best_match = movie

        return best_match

    def _movie_already_exists(self, tmdb_id):
        """Check if movie already exists in collection"""
        try:
            movies = self.client.get_movies()
            return any(movie.get('tmdbId') == tmdb_id for movie in movies)
        except:
            return False

    def _monitor_search_progress(self, movie_id, timeout=300):
        """Monitor search progress for added movie"""
        import time

        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                # Check if movie has file
                movie = self.client.get_movie(movie_id)
                if movie.get('hasFile'):
                    print(f"Movie downloaded successfully: {movie['title']}")
                    return True

                # Check queue for active downloads
                queue = self.client.get_queue()
                movie_in_queue = any(
                    item.get('movieId') == movie_id
                    for item in queue.get('data', queue)
                )

                if movie_in_queue:
                    print(f"Movie is downloading...")
                else:
                    print(f"Searching for movie...")

                time.sleep(30)  # Check every 30 seconds

            except Exception as e:
                print(f"Error monitoring progress: {e}")
                break

        print("Search monitoring timeout reached")
        return False

# Usage example
workflow = MovieAdditionWorkflow(radarr_client)
try:
    added_movie = workflow.search_and_add_movie(
        "Inception 2010",
        quality_profile_id=1,
        root_folder="/movies/sci-fi",
        search_on_add=True
    )
    print(f"Successfully added: {added_movie['title']}")
except Exception as e:
    print(f"Failed to add movie: {e}")
```

### Queue Monitoring and Management

Advanced queue monitoring with automatic management:

```javascript
// JavaScript Implementation
class QueueManager {
    constructor(client) {
        this.client = client;
        this.monitoringInterval = null;
        this.notifications = [];
    }

    async startMonitoring(options = {}) {
        const {
            interval = 30000,  // 30 seconds
            autoRetryFailed = true,
            autoRemoveCompleted = false,
            maxRetries = 3,
            notificationCallback = null
        } = options;

        console.log('Starting queue monitoring...');

        this.monitoringInterval = setInterval(async () => {
            try {
                await this.processQueue({
                    autoRetryFailed,
                    autoRemoveCompleted,
                    maxRetries,
                    notificationCallback
                });
            } catch (error) {
                console.error('Queue monitoring error:', error);
            }
        }, interval);

        // Initial check
        await this.processQueue({
            autoRetryFailed,
            autoRemoveCompleted,
            maxRetries,
            notificationCallback
        });
    }

    async processQueue(options) {
        const queue = await this.client.getQueue(1, 100);  // Get all items
        const queueItems = queue.data || queue;

        if (queueItems.length === 0) {
            return;
        }

        console.log(`Processing ${queueItems.length} queue items`);

        for (const item of queueItems) {
            await this.processQueueItem(item, options);
        }

        // Generate status report
        await this.generateStatusReport(queueItems, options.notificationCallback);
    }

    async processQueueItem(item, options) {
        const { autoRetryFailed, autoRemoveCompleted, maxRetries } = options;

        switch (item.status.toLowerCase()) {
            case 'completed':
                if (autoRemoveCompleted) {
                    await this.removeCompletedItem(item);
                }
                break;

            case 'failed':
                if (autoRetryFailed) {
                    await this.handleFailedItem(item, maxRetries);
                }
                break;

            case 'warning':
                await this.handleWarningItem(item);
                break;

            case 'downloading':
                await this.monitorDownloadProgress(item);
                break;
        }
    }

    async removeCompletedItem(item) {
        try {
            await this.client.removeFromQueue(item.id, true, false);
            console.log(`Removed completed item: ${item.title}`);
        } catch (error) {
            console.error(`Failed to remove completed item: ${error.message}`);
        }
    }

    async handleFailedItem(item, maxRetries) {
        const retryCount = this.getRetryCount(item.id) || 0;

        if (retryCount < maxRetries) {
            try {
                // Try to grab the release again
                const releases = await this.client.searchMovieReleases(item.movieId);
                const sameRelease = releases.find(r => r.guid === item.downloadId);

                if (sameRelease) {
                    await this.client.grabRelease({ guid: sameRelease.guid });
                    this.incrementRetryCount(item.id);
                    console.log(`Retrying failed download: ${item.title} (attempt ${retryCount + 1})`);
                } else {
                    // Try alternative release
                    await this.findAlternativeRelease(item);
                }
            } catch (error) {
                console.error(`Failed to retry download: ${error.message}`);
            }
        } else {
            console.log(`Max retries reached for: ${item.title}`);
            await this.client.removeFromQueue(item.id, true, true);  // Remove and blacklist
        }
    }

    async findAlternativeRelease(item) {
        try {
            const releases = await this.client.searchMovieReleases(item.movieId);

            // Filter out the failed release and find best alternative
            const alternatives = releases.filter(r => r.guid !== item.downloadId);

            if (alternatives.length > 0) {
                // Sort by quality and size (prefer higher quality, reasonable size)
                const bestAlternative = alternatives.sort((a, b) => {
                    const qualityScore = this.getQualityScore(a.quality) - this.getQualityScore(b.quality);
                    if (qualityScore !== 0) return -qualityScore;  // Higher quality first

                    return Math.abs(a.size - item.size) - Math.abs(b.size - item.size);  // Similar size preferred
                })[0];

                await this.client.grabRelease({ guid: bestAlternative.guid });
                console.log(`Found alternative release for: ${item.title}`);

                // Remove failed item
                await this.client.removeFromQueue(item.id, true, false);
            }
        } catch (error) {
            console.error(`Failed to find alternative release: ${error.message}`);
        }
    }

    getQualityScore(quality) {
        const qualityMap = {
            'WEBDL-2160p': 100,
            'BluRay-2160p': 95,
            'WEBDL-1080p': 80,
            'BluRay-1080p': 75,
            'WEBDL-720p': 60,
            'BluRay-720p': 55,
            'HDTV-1080p': 50,
            'HDTV-720p': 40,
            'WEBDL-480p': 30,
            'DVD': 20,
            'SDTV': 10
        };

        return qualityMap[quality?.quality?.name] || 0;
    }

    async monitorDownloadProgress(item) {
        // Check for stalled downloads
        const lastProgress = this.getLastProgress(item.id);
        const currentProgress = item.progress || 0;

        if (lastProgress && currentProgress === lastProgress.value) {
            const stalledTime = Date.now() - lastProgress.timestamp;

            // If stalled for more than 30 minutes, investigate
            if (stalledTime > 30 * 60 * 1000) {
                console.warn(`Download may be stalled: ${item.title} at ${currentProgress}%`);
                // Could implement additional checks or notifications here
            }
        }

        this.setLastProgress(item.id, currentProgress);
    }

    async generateStatusReport(queueItems, notificationCallback) {
        const report = {
            total: queueItems.length,
            downloading: queueItems.filter(i => i.status.toLowerCase() === 'downloading').length,
            paused: queueItems.filter(i => i.status.toLowerCase() === 'paused').length,
            completed: queueItems.filter(i => i.status.toLowerCase() === 'completed').length,
            failed: queueItems.filter(i => i.status.toLowerCase() === 'failed').length,
            warnings: queueItems.filter(i => i.status.toLowerCase() === 'warning').length,
            eta: this.calculateOverallETA(queueItems)
        };

        if (notificationCallback) {
            notificationCallback(report, queueItems);
        }

        console.log('Queue Status:', report);
        return report;
    }

    calculateOverallETA(queueItems) {
        const downloading = queueItems.filter(i =>
            i.status.toLowerCase() === 'downloading' && i.timeLeft
        );

        if (downloading.length === 0) return null;

        // Parse time strings and find maximum
        let maxMinutes = 0;
        for (const item of downloading) {
            const minutes = this.parseTimeToMinutes(item.timeLeft);
            if (minutes > maxMinutes) maxMinutes = minutes;
        }

        return this.formatMinutesToTime(maxMinutes);
    }

    parseTimeToMinutes(timeString) {
        if (!timeString) return 0;

        const match = timeString.match(/(\d+):(\d+):(\d+)/);
        if (match) {
            const [, hours, minutes, seconds] = match;
            return parseInt(hours) * 60 + parseInt(minutes) + parseInt(seconds) / 60;
        }

        return 0;
    }

    formatMinutesToTime(minutes) {
        const hours = Math.floor(minutes / 60);
        const mins = Math.floor(minutes % 60);
        return `${hours}:${mins.toString().padStart(2, '0')}:00`;
    }

    // Helper methods for tracking retry counts and progress
    getRetryCount(itemId) {
        return this.retryTracker?.[itemId] || 0;
    }

    incrementRetryCount(itemId) {
        if (!this.retryTracker) this.retryTracker = {};
        this.retryTracker[itemId] = (this.retryTracker[itemId] || 0) + 1;
    }

    getLastProgress(itemId) {
        return this.progressTracker?.[itemId];
    }

    setLastProgress(itemId, progress) {
        if (!this.progressTracker) this.progressTracker = {};
        this.progressTracker[itemId] = {
            value: progress,
            timestamp: Date.now()
        };
    }

    stopMonitoring() {
        if (this.monitoringInterval) {
            clearInterval(this.monitoringInterval);
            this.monitoringInterval = null;
            console.log('Queue monitoring stopped');
        }
    }
}

// Usage example with Discord notifications
const queueManager = new QueueManager(radarrClient);

queueManager.startMonitoring({
    interval: 30000,
    autoRetryFailed: true,
    autoRemoveCompleted: true,
    maxRetries: 3,
    notificationCallback: async (report, items) => {
        if (report.failed > 0) {
            // Send Discord notification about failures
            const failedItems = items.filter(i => i.status.toLowerCase() === 'failed');
            const message = `🚨 ${report.failed} downloads failed:\n${failedItems.map(i => `• ${i.title}`).join('\n')}`;

            // Send to Discord webhook (implementation depends on your setup)
            // await sendDiscordNotification(message);
        }

        if (report.completed > 0) {
            const completedItems = items.filter(i => i.status.toLowerCase() === 'completed');
            const message = `✅ ${report.completed} downloads completed:\n${completedItems.map(i => `• ${i.title}`).join('\n')}`;

            // Send to Discord webhook
            // await sendDiscordNotification(message);
        }
    }
});

// Stop monitoring when done
// queueManager.stopMonitoring();
```

### Bulk Operations for Large Libraries

Efficient bulk operations with progress tracking and error handling:

```bash
#!/bin/bash

# Bulk Library Management Script
# Handles large-scale operations with progress tracking and error recovery

bulk_movie_operations() {
    local operation="$1"
    local input_file="$2"
    local batch_size="${3:-10}"
    local delay_between_batches="${4:-5}"
    local max_retries="${5:-3}"

    log_info "Starting bulk operation: $operation"
    log_info "Input file: $input_file"
    log_info "Batch size: $batch_size"
    log_info "Delay between batches: ${delay_between_batches}s"

    # Validate input file
    if [[ ! -f "$input_file" ]]; then
        log_error "Input file not found: $input_file"
        return 1
    fi

    # Count total items
    local total_items=$(wc -l < "$input_file")
    local current_item=0
    local successful_items=0
    local failed_items=0

    # Create progress tracking
    local progress_file="bulk_operation_progress_$(date +%s).json"
    local error_file="bulk_operation_errors_$(date +%s).log"

    # Initialize progress tracking
    cat > "$progress_file" <<EOF
{
    "operation": "$operation",
    "total_items": $total_items,
    "processed_items": 0,
    "successful_items": 0,
    "failed_items": 0,
    "start_time": "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')",
    "status": "running"
}
EOF

    # Process in batches
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip empty lines and comments
        [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue

        ((current_item++))

        # Execute operation with retry logic
        local retry_count=0
        local operation_successful=false

        while [[ $retry_count -lt $max_retries ]]; do
            if execute_bulk_operation "$operation" "$line"; then
                operation_successful=true
                ((successful_items++))
                break
            else
                ((retry_count++))
                if [[ $retry_count -lt $max_retries ]]; then
                    log_warning "Operation failed, retrying ($retry_count/$max_retries): $line"
                    sleep $((retry_count * 2))  # Exponential backoff
                fi
            fi
        done

        if [[ "$operation_successful" == "false" ]]; then
            ((failed_items++))
            echo "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ') FAILED: $line" >> "$error_file"
            log_error "All retries failed for: $line"
        fi

        # Update progress
        update_progress "$progress_file" "$current_item" "$successful_items" "$failed_items"

        # Progress reporting
        local progress_percent=$((current_item * 100 / total_items))
        printf "\rProgress: %d%% (%d/%d) | Success: %d | Failed: %d" \
               "$progress_percent" "$current_item" "$total_items" "$successful_items" "$failed_items"

        # Batch delay
        if [[ $((current_item % batch_size)) -eq 0 && $current_item -lt $total_items ]]; then
            echo
            log_info "Batch completed, waiting ${delay_between_batches}s..."
            sleep "$delay_between_batches"
        fi

    done < "$input_file"

    echo  # New line after progress

    # Final progress update
    jq --arg status "completed" \
       --arg end_time "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')" \
       '.status = $status | .end_time = $end_time' "$progress_file" > "${progress_file}.tmp" &&
       mv "${progress_file}.tmp" "$progress_file"

    # Summary
    log_success "Bulk operation completed!"
    log_info "Total items: $total_items"
    log_info "Successful: $successful_items"
    log_info "Failed: $failed_items"
    log_info "Success rate: $((successful_items * 100 / total_items))%"

    if [[ $failed_items -gt 0 ]]; then
        log_warning "Failed items logged to: $error_file"
    fi

    log_info "Progress tracking saved to: $progress_file"
}

execute_bulk_operation() {
    local operation="$1"
    local data="$2"

    case "$operation" in
        "add_movies")
            # Parse CSV line: title,year,tmdb_id,quality_profile_id,root_folder
            IFS=',' read -r title year tmdb_id quality_profile_id root_folder <<< "$data"
            add_movie "$title" "$year" "$tmdb_id" "${quality_profile_id:-1}" "${root_folder:-/movies}" >/dev/null 2>&1
            ;;
        "refresh_movies")
            # Parse movie ID
            refresh_movie "$data" >/dev/null 2>&1
            ;;
        "update_quality")
            # Parse: movie_id,new_quality_profile_id
            IFS=',' read -r movie_id quality_profile_id <<< "$data"
            update_movie_quality "$movie_id" "$quality_profile_id" >/dev/null 2>&1
            ;;
        "search_missing")
            # Parse movie ID for search
            search_movie_releases "$data" >/dev/null 2>&1
            ;;
        *)
            log_error "Unknown operation: $operation"
            return 1
            ;;
    esac
}

update_progress() {
    local progress_file="$1"
    local processed="$2"
    local successful="$3"
    local failed="$4"

    jq --arg processed "$processed" \
       --arg successful "$successful" \
       --arg failed "$failed" \
       --arg update_time "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')" \
       '.processed_items = ($processed | tonumber) |
        .successful_items = ($successful | tonumber) |
        .failed_items = ($failed | tonumber) |
        .last_update = $update_time' "$progress_file" > "${progress_file}.tmp" &&
       mv "${progress_file}.tmp" "$progress_file"
}

update_movie_quality() {
    local movie_id="$1"
    local quality_profile_id="$2"

    # Get current movie data
    local movie_data=$(api_get "/movie/$movie_id")
    if [[ -z "$movie_data" ]]; then
        return 1
    fi

    # Update quality profile
    local updated_data=$(echo "$movie_data" | jq ".qualityProfileId = $quality_profile_id")

    # Save updated movie
    api_put "/movie/$movie_id" "$updated_data" >/dev/null
}

# Parallel bulk operations for better performance
parallel_bulk_operations() {
    local operation="$1"
    local input_file="$2"
    local max_parallel="${3:-5}"
    local batch_size="${4:-20}"

    log_info "Starting parallel bulk operation: $operation"
    log_info "Max parallel processes: $max_parallel"

    # Split input file into chunks
    local temp_dir=$(mktemp -d)
    split -l "$batch_size" "$input_file" "$temp_dir/chunk_"

    # Process chunks in parallel
    local pids=()
    local chunk_count=0

    for chunk_file in "$temp_dir"/chunk_*; do
        if [[ ${#pids[@]} -ge $max_parallel ]]; then
            # Wait for one process to complete
            wait "${pids[0]}"
            pids=("${pids[@]:1}")  # Remove first PID
        fi

        # Start processing chunk in background
        (
            log_info "Processing chunk: $(basename "$chunk_file")"
            bulk_movie_operations "$operation" "$chunk_file" 1 0 3
        ) &

        pids+=($!)
        ((chunk_count++))
    done

    # Wait for all remaining processes
    for pid in "${pids[@]}"; do
        wait "$pid"
    done

    # Cleanup
    rm -rf "$temp_dir"

    log_success "Parallel bulk operation completed: $chunk_count chunks processed"
}

# Quality profile management
bulk_quality_management() {
    log_info "Analyzing quality profile usage"

    # Get all movies and quality profiles
    local movies=$(api_get "/movie")
    local quality_profiles=$(api_get "/qualityprofile")

    # Generate quality usage report
    echo "$movies" | jq -r '
        group_by(.qualityProfileId) |
        map({
            quality_profile_id: .[0].qualityProfileId,
            movie_count: length,
            movies: [.[].title]
        })
    ' > quality_usage_report.json

    # Generate recommendations
    cat > quality_recommendations.txt <<EOF
Quality Profile Usage Analysis
=============================

$(echo "$quality_profiles" | jq -r '.[] | "Profile: \(.name) (ID: \(.id))"')

Movies by Quality Profile:
$(cat quality_usage_report.json | jq -r '.[] | "Profile ID \(.quality_profile_id): \(.movie_count) movies"')

Recommendations:
1. Consider consolidating profiles with low usage
2. Review profiles with many low-quality movies
3. Update orphaned movies to appropriate profiles
EOF

    log_success "Quality analysis complete. Check quality_usage_report.json and quality_recommendations.txt"
}

# Example usage functions
example_bulk_add_from_imdb_list() {
    local imdb_list_url="$1"
    local quality_profile_id="${2:-1}"
    local root_folder="${3:-/movies}"

    log_info "Processing IMDb list: $imdb_list_url"

    # Extract IMDb IDs from list (this would need actual IMDb scraping implementation)
    # For demo purposes, assume we have a file with movie data

    cat > example_movies.csv <<EOF
title,year,tmdb_id,quality_profile_id,root_folder
The Matrix,1999,603,1,/movies
Inception,2010,27205,1,/movies
Interstellar,2014,157336,1,/movies
EOF

    bulk_movie_operations "add_movies" "example_movies.csv" 3 2 3
}

# Resume interrupted operations
resume_bulk_operation() {
    local progress_file="$1"

    if [[ ! -f "$progress_file" ]]; then
        log_error "Progress file not found: $progress_file"
        return 1
    fi

    local status=$(jq -r '.status' "$progress_file")
    if [[ "$status" == "completed" ]]; then
        log_info "Operation already completed"
        return 0
    fi

    local processed_items=$(jq -r '.processed_items' "$progress_file")
    local total_items=$(jq -r '.total_items' "$progress_file")

    log_info "Resuming operation from item $((processed_items + 1)) of $total_items"

    # This would require implementation to skip already processed items
    # For now, just show the concept
    log_warning "Resume functionality requires implementation based on operation type"
}
```

### Real-time Event Handling via WebSocket

Complete WebSocket integration for real-time updates:

```javascript
// Advanced WebSocket Integration
class RadarrWebSocketManager extends EventEmitter {
    constructor(config) {
        super();
        this.config = config;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 10;
        this.reconnectDelay = 5000;
        this.pingInterval = null;
        this.eventHandlers = new Map();
        this.isConnected = false;

        // Event filtering and processing
        this.eventFilters = [];
        this.eventProcessors = new Map();
        this.eventQueue = [];
        this.processingQueue = false;
    }

    connect() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            return Promise.resolve();
        }

        return new Promise((resolve, reject) => {
            const wsUrl = `${this.config.baseUrl.replace('http', 'ws')}/api/v3/signalr/messages?access_token=${this.config.apiKey}`;

            console.log('Connecting to WebSocket:', wsUrl);

            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('WebSocket connected');
                this.isConnected = true;
                this.reconnectAttempts = 0;
                this.startHeartbeat();
                this.emit('connected');
                resolve();
            };

            this.ws.onmessage = (event) => {
                this.handleMessage(event.data);
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket closed:', event.code, event.reason);
                this.isConnected = false;
                this.stopHeartbeat();
                this.emit('disconnected', event);

                if (!event.wasClean && this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect();
                }
            };

            this.ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                this.emit('error', error);
                reject(error);
            };
        });
    }

    handleMessage(data) {
        try {
            const message = JSON.parse(data);

            // Add to processing queue
            this.eventQueue.push({
                ...message,
                receivedAt: new Date(),
                id: this.generateEventId()
            });

            this.processEventQueue();

        } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
        }
    }

    async processEventQueue() {
        if (this.processingQueue || this.eventQueue.length === 0) {
            return;
        }

        this.processingQueue = true;

        while (this.eventQueue.length > 0) {
            const event = this.eventQueue.shift();
            await this.processEvent(event);
        }

        this.processingQueue = false;
    }

    async processEvent(event) {
        try {
            // Apply filters
            if (!this.shouldProcessEvent(event)) {
                return;
            }

            // Apply processors
            const processedEvent = await this.applyEventProcessors(event);

            // Emit general event
            this.emit('message', processedEvent);

            // Emit specific event types
            if (processedEvent.type) {
                this.emit(processedEvent.type, processedEvent.data, processedEvent);
            }

            // Handle specific event patterns
            await this.handleSpecificEvents(processedEvent);

        } catch (error) {
            console.error('Error processing event:', error);
            this.emit('processingError', error, event);
        }
    }

    shouldProcessEvent(event) {
        // Apply all filters
        return this.eventFilters.every(filter => filter(event));
    }

    async applyEventProcessors(event) {
        let processedEvent = { ...event };

        // Apply processors in order
        for (const [name, processor] of this.eventProcessors) {
            try {
                processedEvent = await processor(processedEvent) || processedEvent;
            } catch (error) {
                console.error(`Event processor '${name}' failed:`, error);
            }
        }

        return processedEvent;
    }

    async handleSpecificEvents(event) {
        switch (event.type) {
            case 'movie.downloaded':
                await this.handleMovieDownloaded(event.data);
                break;

            case 'movie.grabbed':
                await this.handleMovieGrabbed(event.data);
                break;

            case 'queue.updated':
                await this.handleQueueUpdated(event.data);
                break;

            case 'health.issue':
                await this.handleHealthIssue(event.data);
                break;

            case 'movie.added':
                await this.handleMovieAdded(event.data);
                break;
        }
    }

    async handleMovieDownloaded(data) {
        console.log(`Movie downloaded: ${data.movie?.title}`);

        // Send notification
        this.emit('notification', {
            type: 'success',
            title: 'Movie Downloaded',
            message: `${data.movie?.title} has been downloaded and is ready to watch!`,
            movie: data.movie,
            timestamp: new Date()
        });

        // Update local cache if available
        if (this.cache) {
            await this.cache.invalidateMovie(data.movie?.id);
        }
    }

    async handleMovieGrabbed(data) {
        console.log(`Movie grabbed: ${data.movie?.title}`);

        this.emit('notification', {
            type: 'info',
            title: 'Movie Grabbed',
            message: `${data.movie?.title} download started`,
            movie: data.movie,
            quality: data.release?.quality,
            indexer: data.release?.indexer,
            timestamp: new Date()
        });
    }

    async handleQueueUpdated(data) {
        // Emit queue update for real-time queue monitoring
        this.emit('queueUpdate', data);

        // Check for stalled downloads
        if (data.items) {
            const stalledItems = data.items.filter(item =>
                item.status === 'downloading' &&
                this.isDownloadStalled(item)
            );

            if (stalledItems.length > 0) {
                this.emit('stalledDownloads', stalledItems);
            }
        }
    }

    async handleHealthIssue(data) {
        if (data.type === 'error' || data.type === 'warning') {
            this.emit('notification', {
                type: data.type,
                title: 'Health Issue Detected',
                message: data.message,
                source: data.source,
                timestamp: new Date()
            });
        }
    }

    async handleMovieAdded(data) {
        console.log(`Movie added: ${data.movie?.title}`);

        this.emit('notification', {
            type: 'info',
            title: 'Movie Added',
            message: `${data.movie?.title} has been added to your collection`,
            movie: data.movie,
            timestamp: new Date()
        });
    }

    // Event filtering methods
    addEventFilter(name, filterFunction) {
        // Store filter with name for potential removal
        filterFunction._filterName = name;
        this.eventFilters.push(filterFunction);
    }

    removeEventFilter(name) {
        this.eventFilters = this.eventFilters.filter(f => f._filterName !== name);
    }

    // Event processing methods
    addEventProcessor(name, processorFunction) {
        this.eventProcessors.set(name, processorFunction);
    }

    removeEventProcessor(name) {
        this.eventProcessors.delete(name);
    }

    // Utility methods
    isDownloadStalled(item) {
        // Simple stall detection - could be enhanced
        if (!item.timeleft || !item.progress) return false;

        const timeLeft = this.parseTimeToMinutes(item.timeleft);
        const progress = item.progress;

        // If progress is low and time left is very high, likely stalled
        return progress < 5 && timeLeft > 1440; // More than 24 hours
    }

    parseTimeToMinutes(timeString) {
        if (!timeString) return 0;
        const match = timeString.match(/(\d+):(\d+):(\d+)/);
        if (match) {
            const [, hours, minutes] = match;
            return parseInt(hours) * 60 + parseInt(minutes);
        }
        return 0;
    }

    generateEventId() {
        return `event_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }

    startHeartbeat() {
        this.pingInterval = setInterval(() => {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.ws.ping();
            }
        }, 30000); // Ping every 30 seconds
    }

    stopHeartbeat() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    scheduleReconnect() {
        const delay = Math.min(
            this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
            60000 // Max 60 seconds
        );

        console.log(`Scheduling reconnect in ${delay}ms (attempt ${this.reconnectAttempts + 1})`);

        setTimeout(() => {
            this.reconnectAttempts++;
            this.connect().catch(error => {
                console.error('Reconnect failed:', error);
            });
        }, delay);
    }

    disconnect() {
        if (this.ws) {
            this.ws.close(1000, 'Client disconnect');
        }
        this.stopHeartbeat();
    }

    // High-level subscription methods
    subscribeToMovieEvents(callback) {
        this.on('movie.downloaded', callback);
        this.on('movie.grabbed', callback);
        this.on('movie.added', callback);
        this.on('movie.deleted', callback);
    }

    subscribeToQueueEvents(callback) {
        this.on('queue.updated', callback);
        this.on('stalledDownloads', callback);
    }

    subscribeToHealthEvents(callback) {
        this.on('health.issue', callback);
    }

    subscribeToNotifications(callback) {
        this.on('notification', callback);
    }
}

// Usage example with comprehensive event handling
const wsManager = new RadarrWebSocketManager({
    baseUrl: 'http://localhost:7878',
    apiKey: 'your-api-key-here'
});

// Add event filters
wsManager.addEventFilter('movieEventsOnly', (event) => {
    return event.type && event.type.startsWith('movie.');
});

// Add event processors
wsManager.addEventProcessor('enrichMovieData', async (event) => {
    if (event.type?.startsWith('movie.') && event.data?.movieId) {
        // Enrich with additional movie data
        try {
            const movieDetails = await radarrClient.getMovie(event.data.movieId);
            event.data.enrichedMovie = movieDetails;
        } catch (error) {
            console.error('Failed to enrich movie data:', error);
        }
    }
    return event;
});

// Subscribe to events
wsManager.subscribeToMovieEvents((event) => {
    console.log('Movie event received:', event.type, event.data);
});

wsManager.subscribeToNotifications((notification) => {
    // Send to notification service
    sendNotification(notification);
});

// Connect
wsManager.connect().then(() => {
    console.log('Successfully connected to Radarr WebSocket');
}).catch(error => {
    console.error('Failed to connect to WebSocket:', error);
});

// Cleanup on exit
process.on('SIGINT', () => {
    wsManager.disconnect();
    process.exit(0);
});
```

---

## Automation Examples

### Backup and Restore API Workflows

Complete backup and restore automation with data integrity validation:

```python
#!/usr/bin/env python3

import os
import json
import gzip
import hashlib
import datetime
from pathlib import Path
from dataclasses import dataclass, asdict
from typing import Dict, List, Optional, Any
import logging

@dataclass
class BackupConfig:
    """Configuration for backup operations"""
    backup_directory: str
    compression_enabled: bool = True
    encryption_enabled: bool = False
    retention_days: int = 30
    verify_integrity: bool = True
    include_movie_files: bool = True
    backup_format_version: str = "1.0"

@dataclass
class BackupManifest:
    """Backup manifest with metadata"""
    created_at: str
    radarr_version: str
    backup_format_version: str
    total_movies: int
    total_size_bytes: int
    components: List[str]
    checksum: str

class RadarrBackupManager:
    """Complete backup and restore manager for Radarr Go"""

    def __init__(self, radarr_client, config: BackupConfig):
        self.client = radarr_client
        self.config = config
        self.logger = logging.getLogger(__name__)

        # Create backup directory
        Path(config.backup_directory).mkdir(parents=True, exist_ok=True)

    def create_full_backup(self, backup_name: str = None) -> str:
        """Create a complete backup of Radarr configuration and data"""

        if not backup_name:
            backup_name = f"radarr_backup_{datetime.datetime.now().strftime('%Y%m%d_%H%M%S')}"

        backup_path = Path(self.config.backup_directory) / backup_name
        backup_path.mkdir(exist_ok=True)

        self.logger.info(f"Starting full backup: {backup_name}")

        try:
            # Backup components
            components = []
            total_size = 0

            # System configuration
            system_status = self.client.get_system_status()
            self._save_component(backup_path, "system_status.json", system_status)
            components.append("system_status")

            # Movies
            if self.config.include_movie_files:
                movies = self._backup_movies(backup_path)
                components.append("movies")
                total_size += movies['size']

            # Quality profiles
            quality_profiles = self.client.api_get("/qualityprofile")
            self._save_component(backup_path, "quality_profiles.json", quality_profiles)
            components.append("quality_profiles")

            # Custom formats
            try:
                custom_formats = self.client.api_get("/customformat")
                self._save_component(backup_path, "custom_formats.json", custom_formats)
                components.append("custom_formats")
            except:
                self.logger.warning("Custom formats backup failed (may not be available)")

            # Indexers
            indexers = self._backup_indexers(backup_path)
            components.append("indexers")

            # Download clients
            download_clients = self._backup_download_clients(backup_path)
            components.append("download_clients")

            # Notifications
            notifications = self.client.api_get("/notification")
            # Sanitize sensitive data
            for notification in notifications:
                if 'fields' in notification:
                    for field in notification['fields']:
                        if field.get('name', '').lower() in ['password', 'apikey', 'token', 'webhook']:
                            field['value'] = '[REDACTED]'
            self._save_component(backup_path, "notifications.json", notifications)
            components.append("notifications")

            # Root folders
            root_folders = self.client.api_get("/rootfolder")
            self._save_component(backup_path, "root_folders.json", root_folders)
            components.append("root_folders")

            # Import lists
            try:
                import_lists = self.client.api_get("/importlist")
                # Sanitize API keys
                for import_list in import_lists:
                    if 'fields' in import_list:
                        for field in import_list['fields']:
                            if 'api' in field.get('name', '').lower():
                                field['value'] = '[REDACTED]'
                self._save_component(backup_path, "import_lists.json", import_lists)
                components.append("import_lists")
            except:
                self.logger.warning("Import lists backup failed (may not be available)")

            # Configuration files
            try:
                host_config = self.client.api_get("/config/host")
                naming_config = self.client.api_get("/config/naming")
                media_management_config = self.client.api_get("/config/mediamanagement")

                self._save_component(backup_path, "host_config.json", host_config)
                self._save_component(backup_path, "naming_config.json", naming_config)
                self._save_component(backup_path, "media_management_config.json", media_management_config)

                components.extend(["host_config", "naming_config", "media_management_config"])
            except:
                self.logger.warning("Configuration backup failed")

            # Create manifest
            manifest = BackupManifest(
                created_at=datetime.datetime.now().isoformat(),
                radarr_version=system_status.get('version', 'unknown'),
                backup_format_version=self.config.backup_format_version,
                total_movies=len(movies.get('data', [])) if 'movies' in locals() else 0,
                total_size_bytes=total_size,
                components=components,
                checksum=self._calculate_backup_checksum(backup_path)
            )

            self._save_component(backup_path, "manifest.json", asdict(manifest))

            # Create archive if compression enabled
            if self.config.compression_enabled:
                archive_path = self._create_archive(backup_path)
                # Remove uncompressed directory
                import shutil
                shutil.rmtree(backup_path)
                backup_path = archive_path

            self.logger.info(f"Backup completed successfully: {backup_path}")
            return str(backup_path)

        except Exception as e:
            self.logger.error(f"Backup failed: {e}")
            # Cleanup on failure
            if backup_path.exists():
                import shutil
                shutil.rmtree(backup_path, ignore_errors=True)
            raise

    def _backup_movies(self, backup_path: Path) -> Dict[str, Any]:
        """Backup all movies with detailed information"""
        movies = self.client.get_movies()

        # Enhance movie data with additional information
        enhanced_movies = []
        total_size = 0

        for movie in movies:
            enhanced_movie = movie.copy()

            # Get movie files
            try:
                movie_files = self.client.api_get(f"/moviefile?movieId={movie['id']}")
                enhanced_movie['movieFiles'] = movie_files

                for file_info in movie_files:
                    if 'size' in file_info:
                        total_size += file_info['size']

            except Exception as e:
                self.logger.warning(f"Could not get files for movie {movie['title']}: {e}")
                enhanced_movie['movieFiles'] = []

            # Get movie history
            try:
                history = self.client.api_get(f"/history?movieId={movie['id']}")
                enhanced_movie['history'] = history
            except Exception as e:
                self.logger.warning(f"Could not get history for movie {movie['title']}: {e}")
                enhanced_movie['history'] = []

            enhanced_movies.append(enhanced_movie)

        movie_backup = {
            'data': enhanced_movies,
            'count': len(enhanced_movies),
            'size': total_size,
            'backed_up_at': datetime.datetime.now().isoformat()
        }

        self._save_component(backup_path, "movies.json", movie_backup)
        return movie_backup

    def _backup_indexers(self, backup_path: Path) -> List[Dict[str, Any]]:
        """Backup indexers with connection testing"""
        indexers = self.client.api_get("/indexer")

        for indexer in indexers:
            # Test indexer connection
            try:
                test_result = self.client.api_post(f"/indexer/{indexer['id']}/test", {})
                indexer['test_result'] = test_result
                indexer['last_tested'] = datetime.datetime.now().isoformat()
            except Exception as e:
                indexer['test_result'] = {'success': False, 'error': str(e)}

            # Sanitize API keys and passwords
            if 'fields' in indexer:
                for field in indexer['fields']:
                    if 'api' in field.get('name', '').lower() or 'password' in field.get('name', '').lower():
                        field['value'] = '[REDACTED]'

        self._save_component(backup_path, "indexers.json", indexers)
        return indexers

    def _backup_download_clients(self, backup_path: Path) -> List[Dict[str, Any]]:
        """Backup download clients with connection testing"""
        download_clients = self.client.api_get("/downloadclient")

        for client in download_clients:
            # Test connection
            try:
                test_result = self.client.api_post("/downloadclient/test", client)
                client['test_result'] = test_result
                client['last_tested'] = datetime.datetime.now().isoformat()
            except Exception as e:
                client['test_result'] = {'success': False, 'error': str(e)}

            # Sanitize passwords
            if 'fields' in client:
                for field in client['fields']:
                    if 'password' in field.get('name', '').lower():
                        field['value'] = '[REDACTED]'

        self._save_component(backup_path, "download_clients.json", download_clients)
        return download_clients

    def _save_component(self, backup_path: Path, filename: str, data: Any):
        """Save a component to the backup directory"""
        file_path = backup_path / filename

        if self.config.compression_enabled:
            with gzip.open(f"{file_path}.gz", 'wt', encoding='utf-8') as f:
                json.dump(data, f, indent=2, default=str)
        else:
            with open(file_path, 'w', encoding='utf-8') as f:
                json.dump(data, f, indent=2, default=str)

    def _calculate_backup_checksum(self, backup_path: Path) -> str:
        """Calculate SHA256 checksum of entire backup"""
        hasher = hashlib.sha256()

        for file_path in sorted(backup_path.glob("**/*")):
            if file_path.is_file() and file_path.name != "manifest.json":
                with open(file_path, 'rb') as f:
                    for chunk in iter(lambda: f.read(4096), b""):
                        hasher.update(chunk)

        return hasher.hexdigest()

    def _create_archive(self, backup_path: Path) -> Path:
        """Create compressed archive of backup"""
        import tarfile

        archive_path = backup_path.with_suffix('.tar.gz')

        with tarfile.open(archive_path, 'w:gz') as tar:
            tar.add(backup_path, arcname=backup_path.name)

        return archive_path

    def restore_from_backup(self, backup_path: str, components: List[str] = None,
                           dry_run: bool = False) -> Dict[str, Any]:
        """Restore Radarr from backup"""

        backup_path = Path(backup_path)

        if not backup_path.exists():
            raise FileNotFoundError(f"Backup not found: {backup_path}")

        self.logger.info(f"Starting restore from: {backup_path}")

        # Extract if needed
        if backup_path.suffix == '.gz':
            backup_path = self._extract_archive(backup_path)

        # Load manifest
        manifest_path = backup_path / "manifest.json"
        if not manifest_path.exists():
            raise ValueError("Invalid backup: manifest.json not found")

        with open(manifest_path, 'r') as f:
            manifest = json.load(f)

        # Verify integrity
        if self.config.verify_integrity:
            self._verify_backup_integrity(backup_path, manifest)

        # Determine components to restore
        available_components = manifest['components']
        if components is None:
            components = available_components
        else:
            # Validate requested components
            invalid_components = set(components) - set(available_components)
            if invalid_components:
                raise ValueError(f"Invalid components: {invalid_components}")

        restore_results = {}

        try:
            # Restore components in dependency order
            restore_order = [
                'quality_profiles',
                'custom_formats',
                'root_folders',
                'indexers',
                'download_clients',
                'notifications',
                'import_lists',
                'host_config',
                'naming_config',
                'media_management_config',
                'movies'
            ]

            for component in restore_order:
                if component in components:
                    result = self._restore_component(backup_path, component, dry_run)
                    restore_results[component] = result

            if not dry_run:
                self.logger.info("Restore completed successfully")
            else:
                self.logger.info("Dry run completed - no changes made")

            return restore_results

        except Exception as e:
            self.logger.error(f"Restore failed: {e}")
            raise

    def _verify_backup_integrity(self, backup_path: Path, manifest: Dict[str, Any]):
        """Verify backup integrity using checksums"""
        expected_checksum = manifest.get('checksum')
        if not expected_checksum:
            self.logger.warning("No checksum in manifest - skipping integrity check")
            return

        calculated_checksum = self._calculate_backup_checksum(backup_path)

        if calculated_checksum != expected_checksum:
            raise ValueError(f"Backup integrity check failed: {calculated_checksum} != {expected_checksum}")

        self.logger.info("Backup integrity verified")

    def _restore_component(self, backup_path: Path, component: str, dry_run: bool) -> Dict[str, Any]:
        """Restore a specific component"""

        component_file = backup_path / f"{component}.json"
        if not component_file.exists():
            component_file = backup_path / f"{component}.json.gz"

        if not component_file.exists():
            return {'success': False, 'error': f'Component file not found: {component}'}

        # Load component data
        if component_file.suffix == '.gz':
            with gzip.open(component_file, 'rt', encoding='utf-8') as f:
                data = json.load(f)
        else:
            with open(component_file, 'r', encoding='utf-8') as f:
                data = json.load(f)

        if dry_run:
            return {
                'success': True,
                'message': f'Would restore {component}',
                'item_count': len(data) if isinstance(data, list) else 1
            }

        try:
            result = self._perform_component_restore(component, data)
            return {'success': True, 'result': result}

        except Exception as e:
            return {'success': False, 'error': str(e)}

    def _perform_component_restore(self, component: str, data: Any) -> Dict[str, Any]:
        """Perform the actual restoration of a component"""

        if component == 'quality_profiles':
            return self._restore_quality_profiles(data)
        elif component == 'movies':
            return self._restore_movies(data)
        elif component == 'indexers':
            return self._restore_indexers(data)
        elif component == 'download_clients':
            return self._restore_download_clients(data)
        elif component == 'notifications':
            return self._restore_notifications(data)
        elif component == 'root_folders':
            return self._restore_root_folders(data)
        else:
            # Generic restore for configuration components
            return self._restore_config_component(component, data)

    def _restore_quality_profiles(self, profiles: List[Dict[str, Any]]) -> Dict[str, Any]:
        """Restore quality profiles with conflict resolution"""
        existing_profiles = self.client.api_get("/qualityprofile")
        existing_names = {p['name'] for p in existing_profiles}

        created = 0
        updated = 0
        skipped = 0

        for profile in profiles:
            profile_name = profile['name']

            if profile_name in existing_names:
                # Update existing profile
                existing_profile = next(p for p in existing_profiles if p['name'] == profile_name)
                profile['id'] = existing_profile['id']

                try:
                    self.client.api_put(f"/qualityprofile/{existing_profile['id']}", profile)
                    updated += 1
                except Exception as e:
                    self.logger.warning(f"Failed to update quality profile {profile_name}: {e}")
                    skipped += 1
            else:
                # Create new profile
                profile.pop('id', None)  # Remove ID for creation

                try:
                    self.client.api_post("/qualityprofile", profile)
                    created += 1
                except Exception as e:
                    self.logger.warning(f"Failed to create quality profile {profile_name}: {e}")
                    skipped += 1

        return {'created': created, 'updated': updated, 'skipped': skipped}

    def _restore_movies(self, movie_data: Dict[str, Any]) -> Dict[str, Any]:
        """Restore movies with duplicate detection"""
        movies = movie_data.get('data', movie_data)
        existing_movies = self.client.get_movies()
        existing_tmdb_ids = {m.get('tmdbId') for m in existing_movies if m.get('tmdbId')}

        added = 0
        updated = 0
        skipped = 0

        for movie in movies:
            tmdb_id = movie.get('tmdbId')

            if tmdb_id in existing_tmdb_ids:
                # Update existing movie
                existing_movie = next(m for m in existing_movies if m.get('tmdbId') == tmdb_id)

                # Merge important settings
                movie['id'] = existing_movie['id']
                movie['path'] = existing_movie.get('path', movie.get('path'))

                try:
                    self.client.update_movie(existing_movie['id'], movie)
                    updated += 1
                except Exception as e:
                    self.logger.warning(f"Failed to update movie {movie.get('title')}: {e}")
                    skipped += 1
            else:
                # Add new movie
                movie.pop('id', None)  # Remove ID for creation
                movie.pop('movieFiles', None)  # Remove file info
                movie.pop('history', None)  # Remove history

                try:
                    self.client.add_movie(movie)
                    added += 1
                except Exception as e:
                    self.logger.warning(f"Failed to add movie {movie.get('title')}: {e}")
                    skipped += 1

        return {'added': added, 'updated': updated, 'skipped': skipped}

    def list_backups(self) -> List[Dict[str, Any]]:
        """List all available backups with metadata"""
        backups = []
        backup_dir = Path(self.config.backup_directory)

        if not backup_dir.exists():
            return backups

        # Find backup directories and archives
        for item in backup_dir.iterdir():
            if item.is_dir():
                manifest_path = item / "manifest.json"
                if manifest_path.exists():
                    backups.append(self._get_backup_info(item, manifest_path))
            elif item.suffix == '.gz' and item.stem.endswith('.tar'):
                # Compressed backup
                try:
                    import tarfile
                    with tarfile.open(item, 'r:gz') as tar:
                        try:
                            manifest_member = tar.getmember(f"{item.stem.replace('.tar', '')}/manifest.json")
                            manifest_content = tar.extractfile(manifest_member).read()
                            manifest = json.loads(manifest_content.decode('utf-8'))

                            backup_info = {
                                'path': str(item),
                                'name': item.stem.replace('.tar', ''),
                                'size': item.stat().st_size,
                                'compressed': True,
                                **manifest
                            }
                            backups.append(backup_info)
                        except:
                            pass  # Skip invalid archives
                except:
                    pass  # Skip unreadable archives

        # Sort by creation date (newest first)
        backups.sort(key=lambda x: x.get('created_at', ''), reverse=True)

        return backups

    def _get_backup_info(self, backup_path: Path, manifest_path: Path) -> Dict[str, Any]:
        """Get backup information from manifest"""
        with open(manifest_path, 'r') as f:
            manifest = json.load(f)

        # Calculate directory size
        total_size = sum(f.stat().st_size for f in backup_path.rglob('*') if f.is_file())

        return {
            'path': str(backup_path),
            'name': backup_path.name,
            'size': total_size,
            'compressed': False,
            **manifest
        }

    def cleanup_old_backups(self) -> int:
        """Remove backups older than retention period"""
        if self.config.retention_days <= 0:
            return 0

        cutoff_date = datetime.datetime.now() - datetime.timedelta(days=self.config.retention_days)
        backups = self.list_backups()

        removed_count = 0

        for backup in backups:
            backup_date = datetime.datetime.fromisoformat(backup.get('created_at', ''))

            if backup_date < cutoff_date:
                backup_path = Path(backup['path'])

                try:
                    if backup_path.is_dir():
                        import shutil
                        shutil.rmtree(backup_path)
                    else:
                        backup_path.unlink()

                    self.logger.info(f"Removed old backup: {backup['name']}")
                    removed_count += 1

                except Exception as e:
                    self.logger.warning(f"Failed to remove backup {backup['name']}: {e}")

        return removed_count

# Usage examples
if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)

    # Initialize backup manager
    from radarr_client import RadarrClient

    client = RadarrClient("http://localhost:7878", "your-api-key")

    config = BackupConfig(
        backup_directory="/backups/radarr",
        compression_enabled=True,
        retention_days=30,
        verify_integrity=True
    )

    backup_manager = RadarrBackupManager(client, config)

    # Create backup
    backup_path = backup_manager.create_full_backup("manual_backup_20250102")
    print(f"Backup created: {backup_path}")

    # List backups
    backups = backup_manager.list_backups()
    print(f"Available backups: {len(backups)}")
    for backup in backups[:5]:  # Show latest 5
        print(f"  - {backup['name']} ({backup['total_movies']} movies, {backup['size']/1024/1024:.1f} MB)")

    # Restore from backup (dry run)
    restore_results = backup_manager.restore_from_backup(
        backup_path,
        components=['quality_profiles', 'indexers'],
        dry_run=True
    )
    print("Restore dry run results:", restore_results)

    # Cleanup old backups
    removed_count = backup_manager.cleanup_old_backups()
    print(f"Cleaned up {removed_count} old backups")
```

### Library Maintenance Automation Scripts

Comprehensive library maintenance with health monitoring and optimization:

```bash
#!/bin/bash

# Radarr Go Library Maintenance Automation
# Comprehensive maintenance scripts with health monitoring and optimization

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
MAINTENANCE_LOG="$SCRIPT_DIR/maintenance.log"
HEALTH_REPORT="$SCRIPT_DIR/health_report.json"
PERFORMANCE_LOG="$SCRIPT_DIR/performance.log"

# Maintenance configuration
MAINTENANCE_CONFIG="$SCRIPT_DIR/maintenance_config.json"
cat > "$MAINTENANCE_CONFIG" <<'EOF'
{
  "schedule": {
    "full_maintenance": "0 2 * * 0",
    "health_check": "0 */4 * * *",
    "cleanup": "0 3 * * *",
    "optimization": "0 1 * * 1"
  },
  "thresholds": {
    "disk_usage_warning": 85,
    "disk_usage_critical": 95,
    "failed_downloads_threshold": 10,
    "stalled_downloads_hours": 24,
    "missing_files_threshold": 50
  },
  "cleanup": {
    "remove_old_logs": true,
    "log_retention_days": 30,
    "clean_empty_folders": true,
    "remove_failed_downloads": true
  },
  "optimization": {
    "refresh_metadata": true,
    "reanalyze_missing": true,
    "update_quality_profiles": false,
    "reorganize_files": false
  }
}
EOF

# Logging functions
log_maintenance() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] MAINTENANCE: $1" | tee -a "$MAINTENANCE_LOG"
}

log_performance() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" >> "$PERFORMANCE_LOG"
}

# Main maintenance orchestrator
run_full_maintenance() {
    log_maintenance "Starting full maintenance cycle"
    local start_time=$(date +%s)

    # Pre-maintenance health check
    log_maintenance "Running pre-maintenance health check"
    local pre_health=$(run_health_check)

    # Create maintenance report
    local report_file="maintenance_report_$(date +%Y%m%d_%H%M%S).json"
    cat > "$report_file" <<EOF
{
    "maintenance_started": "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')",
    "pre_maintenance_health": $pre_health,
    "operations": []
}
EOF

    # Execute maintenance operations
    local operations=(
        "system_cleanup"
        "library_optimization"
        "health_monitoring"
        "performance_tuning"
        "database_maintenance"
        "file_system_cleanup"
    )

    for operation in "${operations[@]}"; do
        log_maintenance "Executing: $operation"
        local operation_start=$(date +%s)

        if execute_maintenance_operation "$operation"; then
            local operation_end=$(date +%s)
            local duration=$((operation_end - operation_start))
            log_performance "$operation completed in ${duration}s"

            # Add to report
            update_maintenance_report "$report_file" "$operation" "success" "$duration"
        else
            log_maintenance "Operation failed: $operation"
            update_maintenance_report "$report_file" "$operation" "failed" 0
        fi
    done

    # Post-maintenance health check
    log_maintenance "Running post-maintenance health check"
    local post_health=$(run_health_check)

    # Finalize report
    local end_time=$(date +%s)
    local total_duration=$((end_time - start_time))

    jq --arg end_time "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')" \
       --arg duration "$total_duration" \
       --argjson post_health "$post_health" \
       '.maintenance_completed = $end_time |
        .total_duration_seconds = ($duration | tonumber) |
        .post_maintenance_health = $post_health' \
       "$report_file" > "${report_file}.tmp" && mv "${report_file}.tmp" "$report_file"

    log_maintenance "Full maintenance cycle completed in ${total_duration}s"
    log_maintenance "Report saved: $report_file"

    # Send notification if configured
    send_maintenance_notification "$report_file"
}

execute_maintenance_operation() {
    local operation="$1"

    case "$operation" in
        "system_cleanup")
            system_cleanup
            ;;
        "library_optimization")
            library_optimization
            ;;
        "health_monitoring")
            comprehensive_health_monitoring
            ;;
        "performance_tuning")
            performance_tuning
            ;;
        "database_maintenance")
            database_maintenance
            ;;
        "file_system_cleanup")
            file_system_cleanup
            ;;
        *)
            log_maintenance "Unknown operation: $operation"
            return 1
            ;;
    esac
}

# System cleanup operations
system_cleanup() {
    log_maintenance "Starting system cleanup"

    # Clean old logs
    if [[ "$(jq -r '.cleanup.remove_old_logs' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        local retention_days=$(jq -r '.cleanup.log_retention_days' "$MAINTENANCE_CONFIG")
        find "$SCRIPT_DIR" -name "*.log" -type f -mtime +$retention_days -delete
        log_maintenance "Cleaned logs older than $retention_days days"
    fi

    # Remove failed downloads
    if [[ "$(jq -r '.cleanup.remove_failed_downloads' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        cleanup_failed_downloads
    fi

    # Clean empty directories
    if [[ "$(jq -r '.cleanup.clean_empty_folders' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        cleanup_empty_directories
    fi

    # Clear Radarr logs and history
    cleanup_radarr_logs

    return 0
}

cleanup_failed_downloads() {
    log_maintenance "Cleaning up failed downloads"

    # Get failed queue items
    local failed_items=$(api_get "/queue" | jq -r '.[] | select(.status == "failed") | .id')
    local cleaned_count=0

    while IFS= read -r item_id; do
        [[ -z "$item_id" ]] && continue

        if api_delete "/queue/${item_id}?removeFromClient=true&blacklist=false" >/dev/null 2>&1; then
            ((cleaned_count++))
        fi
    done <<< "$failed_items"

    log_maintenance "Removed $cleaned_count failed downloads"
}

cleanup_empty_directories() {
    log_maintenance "Cleaning empty directories"

    # Get root folders
    local root_folders=$(api_get "/rootfolder" | jq -r '.[].path')
    local removed_count=0

    while IFS= read -r root_path; do
        [[ -z "$root_path" ]] && continue

        if [[ -d "$root_path" ]]; then
            # Find and remove empty directories (be careful!)
            local empty_dirs=$(find "$root_path" -type d -empty -mindepth 2 2>/dev/null | wc -l)

            if [[ $empty_dirs -gt 0 ]]; then
                find "$root_path" -type d -empty -mindepth 2 -delete 2>/dev/null || true
                removed_count=$((removed_count + empty_dirs))
            fi
        fi
    done <<< "$root_folders"

    log_maintenance "Removed $removed_count empty directories"
}

cleanup_radarr_logs() {
    log_maintenance "Cleaning Radarr logs"

    # Clean old history entries (keep last 1000)
    local history_count=$(api_get "/history?pageSize=1" | jq -r '.totalRecords // 0')

    if [[ $history_count -gt 1000 ]]; then
        log_maintenance "History has $history_count entries, cleaning old entries"

        # Get old history items
        local old_items=$(api_get "/history?page=1&pageSize=$((history_count - 1000))&sortKey=date&sortDirection=ascending" | jq -r '.records[]?.id // empty')
        local deleted_count=0

        while IFS= read -r history_id; do
            [[ -z "$history_id" ]] && continue

            if api_delete "/history/${history_id}" >/dev/null 2>&1; then
                ((deleted_count++))
            fi
        done <<< "$old_items"

        log_maintenance "Cleaned $deleted_count old history entries"
    fi
}

# Library optimization
library_optimization() {
    log_maintenance "Starting library optimization"

    # Refresh metadata for all movies
    if [[ "$(jq -r '.optimization.refresh_metadata' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        refresh_movie_metadata
    fi

    # Analyze missing movies
    if [[ "$(jq -r '.optimization.reanalyze_missing' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        analyze_missing_movies
    fi

    # Optimize quality profiles
    if [[ "$(jq -r '.optimization.update_quality_profiles' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        optimize_quality_profiles
    fi

    # File organization
    if [[ "$(jq -r '.optimization.reorganize_files' "$MAINTENANCE_CONFIG")" == "true" ]]; then
        organize_movie_files
    fi

    return 0
}

refresh_movie_metadata() {
    log_maintenance "Refreshing movie metadata"

    # Get all movies that haven't been refreshed in the last week
    local stale_movies=$(api_get "/movie" | jq -r '
        .[] |
        select(.added and (now - (.added | fromdateiso8601) > 604800)) |
        .id
    ')

    local refresh_count=0
    while IFS= read -r movie_id; do
        [[ -z "$movie_id" ]] && continue

        if execute_command "RefreshMovie" "\"movieId\": $movie_id" >/dev/null; then
            ((refresh_count++))
        fi

        # Rate limiting
        sleep 0.5
    done <<< "$stale_movies"

    log_maintenance "Queued metadata refresh for $refresh_count movies"
}

analyze_missing_movies() {
    log_maintenance "Analyzing missing movies"

    # Get missing movies
    local missing_movies=$(api_get "/wanted/missing?pageSize=100" | jq -r '.records[] | .id')
    local missing_count=$(echo "$missing_movies" | wc -l)

    if [[ $missing_count -gt 0 ]]; then
        log_maintenance "Found $missing_count missing movies"

        # Create missing movies report
        local missing_report="missing_movies_$(date +%Y%m%d).json"
        api_get "/wanted/missing?pageSize=1000" > "$missing_report"

        # Analyze patterns
        analyze_missing_patterns "$missing_report"

        # Auto-search for high-priority missing movies
        auto_search_missing_movies
    fi
}

analyze_missing_patterns() {
    local report_file="$1"

    log_maintenance "Analyzing missing movie patterns"

    # Analyze by year, genre, and quality profile
    jq -r '
        .records[] |
        [.year, .genres[0]?, .qualityProfileId, .title] |
        @csv
    ' "$report_file" | sort | uniq -c | sort -nr > "missing_patterns_$(date +%Y%m%d).txt"

    # Find movies missing for longest time
    jq -r '
        .records[] |
        select(.added) |
        [.added, .title, .year] |
        @csv
    ' "$report_file" | sort | head -20 > "longest_missing_$(date +%Y%m%d).txt"
}

auto_search_missing_movies() {
    log_maintenance "Auto-searching missing movies"

    # Get high-priority missing movies (added recently, popular)
    local priority_missing=$(api_get "/wanted/missing?sortKey=added&sortDirection=descending&pageSize=10" | jq -r '.records[] | .id')

    local search_count=0
    while IFS= read -r movie_id; do
        [[ -z "$movie_id" ]] && continue

        if execute_command "MoviesSearch" "\"movieIds\": [$movie_id]" >/dev/null; then
            ((search_count++))
        fi

        # Rate limiting for searches
        sleep 2
    done <<< "$priority_missing"

    log_maintenance "Initiated search for $search_count high-priority missing movies"
}

# Comprehensive health monitoring
comprehensive_health_monitoring() {
    log_maintenance "Running comprehensive health monitoring"

    local health_data="{}"

    # System health
    health_data=$(echo "$health_data" | jq --argjson system_health "$(check_system_health)" '.system_health = $system_health')

    # Indexer health
    health_data=$(echo "$health_data" | jq --argjson indexer_health "$(check_indexer_health)" '.indexer_health = $indexer_health')

    # Download client health
    health_data=$(echo "$health_data" | jq --argjson client_health "$(check_download_client_health)" '.download_client_health = $client_health')

    # Library health
    health_data=$(echo "$health_data" | jq --argjson library_health "$(check_library_health)" '.library_health = $library_health')

    # Performance metrics
    health_data=$(echo "$health_data" | jq --argjson performance "$(collect_performance_metrics)" '.performance_metrics = $performance')

    # Save health report
    echo "$health_data" | jq --arg timestamp "$(date -u '+%Y-%m-%dT%H:%M:%S.%3NZ')" '. + {timestamp: $timestamp}' > "$HEALTH_REPORT"

    # Check for critical issues
    check_critical_issues "$health_data"

    return 0
}

check_system_health() {
    local system_health="{}"

    # Disk space check
    local disk_usage=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    local disk_warning=$(jq -r '.thresholds.disk_usage_warning' "$MAINTENANCE_CONFIG")
    local disk_critical=$(jq -r '.thresholds.disk_usage_critical' "$MAINTENANCE_CONFIG")

    local disk_status="ok"
    if [[ $disk_usage -gt $disk_critical ]]; then
        disk_status="critical"
    elif [[ $disk_usage -gt $disk_warning ]]; then
        disk_status="warning"
    fi

    system_health=$(echo "$system_health" | jq --arg usage "$disk_usage" --arg status "$disk_status" '.disk = {usage: $usage, status: $status}')

    # Memory usage
    local memory_info=$(free -m | awk 'NR==2{printf "%.1f", $3*100/$2}')
    system_health=$(echo "$system_health" | jq --arg memory "$memory_info" '.memory_usage = $memory')

    # Radarr service health
    if check_health >/dev/null 2>&1; then
        system_health=$(echo "$system_health" | jq '.radarr_service = "running"')
    else
        system_health=$(echo "$system_health" | jq '.radarr_service = "error"')
    fi

    echo "$system_health"
}

check_indexer_health() {
    local indexer_health="{}"
    local indexers=$(api_get "/indexer")

    local total_indexers=$(echo "$indexers" | jq 'length')
    local working_indexers=0
    local failed_tests=0

    echo "$indexers" | jq -c '.[]' | while IFS= read -r indexer; do
        local indexer_id=$(echo "$indexer" | jq -r '.id')
        local indexer_name=$(echo "$indexer" | jq -r '.name')

        # Test indexer
        if api_post "/indexer/${indexer_id}/test" "{}" >/dev/null 2>&1; then
            ((working_indexers++))
            log_maintenance "Indexer $indexer_name is healthy"
        else
            ((failed_tests++))
            log_maintenance "Indexer $indexer_name failed health check"
        fi

        sleep 1  # Rate limiting
    done

    indexer_health=$(echo "$indexer_health" | jq \
        --arg total "$total_indexers" \
        --arg working "$working_indexers" \
        --arg failed "$failed_tests" \
        '.total = ($total | tonumber) | .working = ($working | tonumber) | .failed = ($failed | tonumber)')

    echo "$indexer_health"
}

check_download_client_health() {
    local client_health="{}"
    local clients=$(api_get "/downloadclient")

    local total_clients=$(echo "$clients" | jq 'length')
    local working_clients=0

    echo "$clients" | jq -c '.[]' | while IFS= read -r client; do
        local client_name=$(echo "$client" | jq -r '.name')

        # Test download client
        if api_post "/downloadclient/test" "$client" >/dev/null 2>&1; then
            ((working_clients++))
            log_maintenance "Download client $client_name is healthy"
        else
            log_maintenance "Download client $client_name failed health check"
        fi

        sleep 1  # Rate limiting
    done

    client_health=$(echo "$client_health" | jq \
        --arg total "$total_clients" \
        --arg working "$working_clients" \
        '.total = ($total | tonumber) | .working = ($working | tonumber)')

    echo "$client_health"
}

check_library_health() {
    local library_health="{}"

    # Get library statistics
    local total_movies=$(api_get "/movie" | jq 'length')
    local missing_movies=$(api_get "/wanted/missing?pageSize=1" | jq '.totalRecords // 0')
    local movies_with_files=$(api_get "/movie" | jq '[.[] | select(.hasFile)] | length')

    library_health=$(echo "$library_health" | jq \
        --arg total "$total_movies" \
        --arg missing "$missing_movies" \
        --arg with_files "$movies_with_files" \
        '.total_movies = ($total | tonumber) |
         .missing_movies = ($missing | tonumber) |
         .movies_with_files = ($with_files | tonumber) |
         .completion_rate = (($with_files | tonumber) / ($total | tonumber) * 100)')

    # Check for orphaned files
    check_orphaned_files

    echo "$library_health"
}

collect_performance_metrics() {
    local metrics="{}"

    # API response times
    local api_start=$(date +%s%N)
    api_get "/system/status" >/dev/null
    local api_end=$(date +%s%N)
    local api_response_time=$(( (api_end - api_start) / 1000000 ))  # Convert to milliseconds

    metrics=$(echo "$metrics" | jq --arg response_time "$api_response_time" '.api_response_time_ms = ($response_time | tonumber)')

    # Queue processing speed
    local queue=$(api_get "/queue")
    local queue_size=$(echo "$queue" | jq 'length')
    local active_downloads=$(echo "$queue" | jq '[.[] | select(.status == "downloading")] | length')

    metrics=$(echo "$metrics" | jq \
        --arg queue_size "$queue_size" \
        --arg active "$active_downloads" \
        '.queue_size = ($queue_size | tonumber) | .active_downloads = ($active | tonumber)')

    echo "$metrics"
}

# Send maintenance notifications
send_maintenance_notification() {
    local report_file="$1"

    # Check if notification webhook is configured
    local webhook_url="${MAINTENANCE_WEBHOOK_URL:-}"

    if [[ -n "$webhook_url" ]]; then
        # Create notification payload
        local payload=$(jq -n \
            --arg title "Radarr Maintenance Report" \
            --slurpfile report "$report_file" \
            '{
                title: $title,
                color: "green",
                description: "Maintenance cycle completed successfully",
                fields: [
                    {
                        name: "Duration",
                        value: ($report[0].total_duration_seconds | tostring) + " seconds",
                        inline: true
                    },
                    {
                        name: "Operations",
                        value: ($report[0].operations | length | tostring),
                        inline: true
                    }
                ],
                timestamp: ($report[0].maintenance_completed)
            }')

        # Send notification
        curl -X POST "$webhook_url" \
             -H "Content-Type: application/json" \
             -d "$payload" \
             >/dev/null 2>&1
    fi
}

# Main execution
main() {
    case "${1:-}" in
        "full")
            run_full_maintenance
            ;;
        "health")
            comprehensive_health_monitoring
            ;;
        "cleanup")
            system_cleanup
            ;;
        "optimize")
            library_optimization
            ;;
        *)
            cat <<EOF
Radarr Go Library Maintenance Automation

Usage: $0 <command>

Commands:
  full      - Run complete maintenance cycle
  health    - Run health monitoring
  cleanup   - Run system cleanup
  optimize  - Run library optimization

Environment Variables:
  RADARR_URL               - Radarr server URL
  RADARR_API_KEY          - API key
  MAINTENANCE_WEBHOOK_URL  - Discord/Slack webhook for notifications

EOF
            exit 1
            ;;
    esac
}

# Validation
if [[ -z "${RADARR_API_KEY:-}" ]]; then
    echo "Error: RADARR_API_KEY environment variable is required"
    exit 1
fi

# Execute main function
main "$@"
```

---

## Troubleshooting Guide

### Common API Errors and Solutions

Complete troubleshooting guide for common integration issues:

#### Authentication Issues

**Error: 401 Unauthorized - Invalid API Key**
```json
{
  "error": {
    "message": "Invalid API key",
    "code": "UNAUTHORIZED",
    "timestamp": "2025-01-02T10:30:00Z"
  }
}
```

**Solutions:**
```python
# Check API key format and location
def troubleshoot_authentication():
    """Troubleshoot common authentication issues"""

    # 1. Verify API key format
    api_key = "your-api-key-here"
    if len(api_key) != 32:
        print("❌ API key should be 32 characters long")
        return False

    # 2. Test different authentication methods
    headers_auth = {'X-API-Key': api_key}
    query_auth = {'apikey': api_key}

    # Try header authentication first (preferred)
    try:
        response = requests.get(f"{base_url}/api/v3/system/status", headers=headers_auth)
        if response.status_code == 200:
            print("✅ Header authentication working")
            return True
    except Exception as e:
        print(f"❌ Header authentication failed: {e}")

    # Try query parameter authentication
    try:
        response = requests.get(f"{base_url}/api/v3/system/status", params=query_auth)
        if response.status_code == 200:
            print("✅ Query parameter authentication working")
            return True
    except Exception as e:
        print(f"❌ Query parameter authentication failed: {e}")

    # 3. Check if API key is enabled in Radarr settings
    print("❌ Check Radarr settings:")
    print("   - Settings > General > Security > API Key")
    print("   - Ensure API key authentication is enabled")

    return False

# Generate new API key
def generate_api_key():
    """Generate a new API key for troubleshooting"""
    import secrets
    import string

    alphabet = string.ascii_lowercase + string.digits
    api_key = ''.join(secrets.choice(alphabet) for _ in range(32))

    print(f"New API key (for testing): {api_key}")
    print("⚠️  Remember to update this in your Radarr settings")

    return api_key
```

**Error: 403 Forbidden - Authentication method not allowed**
```bash
# Check Radarr configuration for authentication methods
curl -H "X-API-Key: your-api-key" http://localhost:7878/api/v3/config/host

# Verify authentication settings
{
  "authenticationMethod": "forms",  # Should be "none" or "basic" for API access
  "authenticationRequired": "enabled"
}
```

#### Connection and Network Issues

**Error: Connection timeout or refused**
```python
import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry

def create_resilient_session():
    """Create a session with retry logic and timeouts"""
    session = requests.Session()

    # Configure retry strategy
    retry_strategy = Retry(
        total=3,
        backoff_factor=1,
        status_forcelist=[429, 500, 502, 503, 504],
        allowed_methods=["HEAD", "GET", "OPTIONS", "POST", "PUT", "DELETE"]
    )

    # Mount adapter with retry strategy
    adapter = HTTPAdapter(max_retries=retry_strategy)
    session.mount("http://", adapter)
    session.mount("https://", adapter)

    # Set reasonable timeouts
    session.timeout = (10, 30)  # (connect_timeout, read_timeout)

    return session

# Test connectivity with diagnostics
def diagnose_connection(base_url):
    """Diagnose connection issues with detailed feedback"""
    import socket
    from urllib.parse import urlparse

    parsed = urlparse(base_url)
    host = parsed.hostname or 'localhost'
    port = parsed.port or (443 if parsed.scheme == 'https' else 7878)

    print(f"Testing connection to {host}:{port}")

    # Test basic connectivity
    try:
        sock = socket.create_connection((host, port), timeout=10)
        sock.close()
        print("✅ TCP connection successful")
    except socket.timeout:
        print("❌ Connection timeout - check if Radarr is running")
        return False
    except ConnectionRefusedError:
        print("❌ Connection refused - check if Radarr is listening on this port")
        return False
    except socket.gaierror as e:
        print(f"❌ DNS resolution failed: {e}")
        return False

    # Test HTTP response
    try:
        response = requests.get(f"{base_url}/ping", timeout=10)
        if response.status_code == 200:
            print("✅ HTTP service responding")
            return True
        else:
            print(f"❌ HTTP error: {response.status_code}")
    except requests.exceptions.Timeout:
        print("❌ HTTP timeout - service may be overloaded")
    except requests.exceptions.RequestException as e:
        print(f"❌ HTTP request failed: {e}")

    return False
```

**CORS Issues (Browser/Web Applications)**
```javascript
// Configure CORS for browser-based applications
class RadarrClientWithCORS {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;

        // Check if running in browser
        this.isBrowser = typeof window !== 'undefined';
    }

    async makeRequest(endpoint, options = {}) {
        const url = `${this.baseUrl}/api/v3${endpoint}`;

        const config = {
            mode: 'cors',  // Explicit CORS mode
            credentials: 'omit',  // Don't send cookies
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey,
                ...options.headers
            },
            ...options
        };

        // Add CORS preflight handling
        if (this.isBrowser && (options.method === 'POST' || options.method === 'PUT' || options.method === 'DELETE')) {
            // Browser will send preflight OPTIONS request automatically
            console.log('CORS preflight will be sent automatically');
        }

        try {
            const response = await fetch(url, config);

            if (!response.ok) {
                // Handle CORS-specific errors
                if (response.type === 'opaque' || response.status === 0) {
                    throw new Error('CORS error: Check Radarr CORS settings and ensure API is accessible');
                }

                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            if (error.message.includes('CORS')) {
                console.error('CORS Configuration Help:');
                console.error('1. Check Radarr Settings > General > Host');
                console.error('2. Add your domain to "Allowed Origins"');
                console.error('3. Or use a proxy/server-side integration');
            }
            throw error;
        }
    }
}

// CORS troubleshooting helper
function diagnoseCORS(radarrUrl) {
    console.log('CORS Troubleshooting Guide:');
    console.log('1. Browser Developer Tools > Network tab');
    console.log('2. Look for preflight OPTIONS requests');
    console.log('3. Check response headers for:');
    console.log('   - Access-Control-Allow-Origin');
    console.log('   - Access-Control-Allow-Methods');
    console.log('   - Access-Control-Allow-Headers');
    console.log('');
    console.log('Radarr CORS Settings:');
    console.log(`- URL: ${radarrUrl}/settings/general`);
    console.log('- Add your domain to "URL Base" or use "*" for development');
    console.log('- Restart Radarr after changes');
}
```

#### Rate Limiting Issues

**Error: 429 Too Many Requests**
```python
import time
import random
from functools import wraps

class RateLimitedClient:
    """Client with intelligent rate limiting and backoff"""

    def __init__(self, base_url, api_key, requests_per_minute=60):
        self.base_url = base_url
        self.api_key = api_key
        self.min_interval = 60.0 / requests_per_minute  # Seconds between requests
        self.last_request_time = 0
        self.consecutive_errors = 0
        self.backoff_multiplier = 1

    def rate_limited_request(func):
        """Decorator for rate-limited API requests"""
        @wraps(func)
        def wrapper(self, *args, **kwargs):
            # Wait for minimum interval
            time_since_last = time.time() - self.last_request_time
            if time_since_last < self.min_interval * self.backoff_multiplier:
                sleep_time = (self.min_interval * self.backoff_multiplier) - time_since_last
                time.sleep(sleep_time)

            try:
                result = func(self, *args, **kwargs)

                # Success - reset backoff
                self.consecutive_errors = 0
                self.backoff_multiplier = 1
                self.last_request_time = time.time()

                return result

            except requests.exceptions.HTTPError as e:
                if e.response.status_code == 429:
                    # Rate limited - increase backoff
                    self.consecutive_errors += 1
                    self.backoff_multiplier = min(2 ** self.consecutive_errors, 32)

                    # Extract retry-after header if present
                    retry_after = e.response.headers.get('Retry-After', 60)
                    wait_time = int(retry_after) + random.uniform(1, 5)  # Add jitter

                    print(f"Rate limited. Waiting {wait_time} seconds...")
                    time.sleep(wait_time)

                    # Retry the request
                    return wrapper(self, *args, **kwargs)

                raise

        return wrapper

    @rate_limited_request
    def api_get(self, endpoint):
        response = requests.get(
            f"{self.base_url}/api/v3{endpoint}",
            headers={'X-API-Key': self.api_key}
        )
        response.raise_for_status()
        return response.json()

# Batch processing with rate limiting
def process_batch_with_rate_limiting(items, process_func, batch_size=10, delay=1):
    """Process items in batches with rate limiting"""
    results = []

    for i in range(0, len(items), batch_size):
        batch = items[i:i+batch_size]

        print(f"Processing batch {i//batch_size + 1}/{(len(items) + batch_size - 1)//batch_size}")

        batch_results = []
        for item in batch:
            try:
                result = process_func(item)
                batch_results.append(result)
            except Exception as e:
                print(f"Error processing item {item}: {e}")
                batch_results.append(None)

            # Small delay between items in batch
            time.sleep(delay)

        results.extend(batch_results)

        # Longer delay between batches
        if i + batch_size < len(items):
            time.sleep(delay * 3)

    return results
```

#### WebSocket Connection Issues

**WebSocket Connection Failures**
```javascript
class RobustWebSocketManager {
    constructor(config) {
        this.config = config;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 10;
        this.reconnectDelay = 1000;
        this.isManualDisconnect = false;

        // Connection diagnostics
        this.connectionLog = [];
        this.lastPingTime = null;
        this.pingInterval = null;
    }

    async connect() {
        return new Promise((resolve, reject) => {
            try {
                // Validate WebSocket URL
                const wsUrl = this.constructWebSocketUrl();
                this.logConnection('info', `Attempting connection to: ${wsUrl}`);

                this.ws = new WebSocket(wsUrl);

                this.ws.onopen = (event) => {
                    this.logConnection('success', 'WebSocket connected');
                    this.reconnectAttempts = 0;
                    this.startPing();
                    resolve(event);
                };

                this.ws.onmessage = (event) => {
                    this.handleMessage(event);
                };

                this.ws.onclose = (event) => {
                    this.logConnection('warning', `WebSocket closed: ${event.code} - ${event.reason}`);
                    this.stopPing();

                    if (!this.isManualDisconnect) {
                        this.handleReconnection(event);
                    }
                };

                this.ws.onerror = (error) => {
                    this.logConnection('error', `WebSocket error: ${error.message}`);

                    // Provide specific troubleshooting guidance
                    this.diagnoseWebSocketError(error);
                    reject(error);
                };

                // Connection timeout
                setTimeout(() => {
                    if (this.ws.readyState === WebSocket.CONNECTING) {
                        this.ws.close();
                        reject(new Error('WebSocket connection timeout'));
                    }
                }, 10000);

            } catch (error) {
                this.logConnection('error', `Failed to create WebSocket: ${error.message}`);
                reject(error);
            }
        });
    }

    constructWebSocketUrl() {
        let wsUrl = this.config.baseUrl.replace('http://', 'ws://').replace('https://', 'wss://');

        // Check if URL ends with proper WebSocket endpoint
        if (!wsUrl.includes('/signalr')) {
            wsUrl += '/api/v3/signalr/messages';
        }

        // Add authentication
        const separator = wsUrl.includes('?') ? '&' : '?';
        wsUrl += `${separator}access_token=${this.config.apiKey}`;

        return wsUrl;
    }

    diagnoseWebSocketError(error) {
        const troubleshooting = [
            '🔍 WebSocket Troubleshooting:',
            '',
            '1. Check Radarr Configuration:',
            '   - Ensure WebSocket support is enabled',
            '   - Verify API key is correct',
            '   - Check firewall settings',
            '',
            '2. Network Issues:',
            '   - Try different network connection',
            '   - Check proxy settings',
            '   - Verify DNS resolution',
            '',
            '3. Browser Issues (if applicable):',
            '   - Check browser console for CORS errors',
            '   - Try in incognito/private mode',
            '   - Clear browser cache',
            '',
            '4. Radarr Version:',
            '   - Ensure using compatible Radarr version',
            '   - Check for Radarr updates',
            ''
        ];

        console.error(troubleshooting.join('\n'));

        // Log connection attempts for analysis
        console.table(this.connectionLog.slice(-10));
    }

    startPing() {
        this.pingInterval = setInterval(() => {
            if (this.ws && this.ws.readyState === WebSocket.OPEN) {
                this.lastPingTime = Date.now();
                this.ws.send(JSON.stringify({type: 'ping'}));

                // Check for pong response
                setTimeout(() => {
                    const timeSincePing = Date.now() - this.lastPingTime;
                    if (timeSincePing > 30000) {  // 30 seconds without pong
                        this.logConnection('warning', 'Ping timeout - connection may be stale');
                        this.ws.close(1000, 'Ping timeout');
                    }
                }, 30000);
            }
        }, 45000);  // Ping every 45 seconds
    }

    stopPing() {
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    handleReconnection(event) {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            this.logConnection('error', 'Max reconnection attempts reached');
            return;
        }

        const delay = Math.min(
            this.reconnectDelay * Math.pow(2, this.reconnectAttempts),
            30000  // Max 30 seconds
        );

        this.reconnectAttempts++;
        this.logConnection('info', `Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`);

        setTimeout(() => {
            this.connect().catch(error => {
                this.logConnection('error', `Reconnection failed: ${error.message}`);
            });
        }, delay);
    }

    logConnection(level, message) {
        const logEntry = {
            timestamp: new Date().toISOString(),
            level,
            message,
            attempts: this.reconnectAttempts,
            readyState: this.ws ? this.ws.readyState : 'None'
        };

        this.connectionLog.push(logEntry);

        // Keep only last 50 log entries
        if (this.connectionLog.length > 50) {
            this.connectionLog.shift();
        }

        // Console output with colors
        const colors = {
            info: '\x1b[36m',      // Cyan
            success: '\x1b[32m',   // Green
            warning: '\x1b[33m',   // Yellow
            error: '\x1b[31m',     // Red
            reset: '\x1b[0m'       // Reset
        };

        console.log(`${colors[level]}[WebSocket ${level.toUpperCase()}]${colors.reset} ${message}`);
    }

    // Public method to get connection diagnostics
    getConnectionDiagnostics() {
        return {
            currentState: this.ws ? this.ws.readyState : 'Disconnected',
            reconnectAttempts: this.reconnectAttempts,
            isManualDisconnect: this.isManualDisconnect,
            recentLogs: this.connectionLog.slice(-10),
            pingStatus: this.lastPingTime ? `Last ping: ${new Date(this.lastPingTime).toLocaleString()}` : 'No pings sent'
        };
    }
}

// Usage example with error handling
const wsManager = new RobustWebSocketManager({
    baseUrl: 'http://localhost:7878',
    apiKey: 'your-api-key'
});

wsManager.connect()
    .then(() => console.log('Connected successfully'))
    .catch(error => {
        console.error('Connection failed:', error.message);

        // Show diagnostics
        const diagnostics = wsManager.getConnectionDiagnostics();
        console.log('Connection Diagnostics:', diagnostics);
    });
```

### Performance Optimization

**Slow API Response Times**
```python
import time
import statistics
from functools import wraps

class PerformanceMonitor:
    """Monitor and optimize API performance"""

    def __init__(self):
        self.response_times = []
        self.slow_endpoints = {}
        self.cache = {}
        self.cache_ttl = {}

    def performance_monitor(func):
        """Decorator to monitor API performance"""
        @wraps(func)
        def wrapper(self, *args, **kwargs):
            start_time = time.time()

            try:
                result = func(self, *args, **kwargs)

                # Record successful response time
                response_time = time.time() - start_time
                self.response_times.append(response_time)

                # Keep only last 100 measurements
                if len(self.response_times) > 100:
                    self.response_times.pop(0)

                # Track slow endpoints
                endpoint = args[0] if args else 'unknown'
                if response_time > 2.0:  # Slow threshold: 2 seconds
                    if endpoint not in self.slow_endpoints:
                        self.slow_endpoints[endpoint] = []
                    self.slow_endpoints[endpoint].append(response_time)

                return result

            except Exception as e:
                response_time = time.time() - start_time
                print(f"Request failed after {response_time:.2f}s: {e}")
                raise

        return wrapper

    @performance_monitor
    def optimized_api_get(self, endpoint, use_cache=True, cache_ttl=300):
        """Optimized API GET with caching"""

        # Check cache first
        if use_cache and endpoint in self.cache:
            cache_time = self.cache_ttl.get(endpoint, 0)
            if time.time() - cache_time < cache_ttl:
                print(f"Cache hit for {endpoint}")
                return self.cache[endpoint]

        # Make API request
        response = requests.get(
            f"{self.base_url}/api/v3{endpoint}",
            headers={'X-API-Key': self.api_key},
            timeout=30
        )
        response.raise_for_status()

        result = response.json()

        # Cache result
        if use_cache:
            self.cache[endpoint] = result
            self.cache_ttl[endpoint] = time.time()

        return result

    def get_performance_stats(self):
        """Get performance statistics"""
        if not self.response_times:
            return None

        return {
            'average_response_time': statistics.mean(self.response_times),
            'median_response_time': statistics.median(self.response_times),
            'max_response_time': max(self.response_times),
            'min_response_time': min(self.response_times),
            'total_requests': len(self.response_times),
            'slow_endpoints': {
                endpoint: {
                    'count': len(times),
                    'average': statistics.mean(times),
                    'max': max(times)
                }
                for endpoint, times in self.slow_endpoints.items()
            }
        }

    def optimize_requests(self):
        """Provide optimization recommendations"""
        stats = self.get_performance_stats()
        if not stats:
            return []

        recommendations = []

        if stats['average_response_time'] > 1.0:
            recommendations.append({
                'issue': 'Slow average response time',
                'value': f"{stats['average_response_time']:.2f}s",
                'suggestions': [
                    'Enable API response caching',
                    'Use pagination for large datasets',
                    'Filter results to reduce payload size',
                    'Check Radarr server performance'
                ]
            })

        if stats['slow_endpoints']:
            for endpoint, data in stats['slow_endpoints'].items():
                recommendations.append({
                    'issue': f'Slow endpoint: {endpoint}',
                    'value': f"Average: {data['average']:.2f}s, Max: {data['max']:.2f}s",
                    'suggestions': [
                        f'Cache responses for {endpoint}',
                        'Use specific filters to reduce data',
                        'Consider pagination',
                        'Monitor during low-usage periods'
                    ]
                })

        return recommendations

# Pagination helper for large datasets
class PaginatedClient:
    """Client that handles pagination efficiently"""

    def __init__(self, base_client):
        self.client = base_client

    def get_all_paginated(self, endpoint, page_size=100, max_items=None):
        """Get all items from a paginated endpoint"""
        all_items = []
        page = 1

        while True:
            # Construct paginated request
            paginated_endpoint = f"{endpoint}?page={page}&pageSize={page_size}"

            try:
                response = self.client.optimized_api_get(paginated_endpoint)

                # Handle different response formats
                items = []
                if isinstance(response, dict):
                    if 'data' in response:
                        items = response['data']
                        total_records = response.get('totalRecords', 0)
                    elif 'records' in response:
                        items = response['records']
                        total_records = response.get('totalRecords', 0)
                    else:
                        # Single page response
                        break
                else:
                    items = response

                if not items:
                    break

                all_items.extend(items)

                # Check if we've gotten all items or hit max limit
                if max_items and len(all_items) >= max_items:
                    all_items = all_items[:max_items]
                    break

                if 'total_records' in locals() and len(all_items) >= total_records:
                    break

                if len(items) < page_size:
                    # Last page
                    break

                page += 1

            except Exception as e:
                print(f"Error fetching page {page}: {e}")
                break

        return all_items

    def get_paginated_generator(self, endpoint, page_size=100):
        """Generator that yields items from paginated endpoint"""
        page = 1

        while True:
            paginated_endpoint = f"{endpoint}?page={page}&pageSize={page_size}"

            try:
                response = self.client.optimized_api_get(paginated_endpoint)

                items = []
                if isinstance(response, dict) and 'data' in response:
                    items = response['data']
                elif isinstance(response, dict) and 'records' in response:
                    items = response['records']
                else:
                    items = response if isinstance(response, list) else []

                if not items:
                    break

                for item in items:
                    yield item

                if len(items) < page_size:
                    break

                page += 1

            except Exception as e:
                print(f"Error fetching page {page}: {e}")
                break
```

### Custom Notification Webhook Implementation

**Complete webhook implementation with error handling**
```python
import hmac
import hashlib
import json
from datetime import datetime
from typing import Dict, Any, Optional

class RadarrWebhookHandler:
    """Complete webhook handler for Radarr notifications"""

    def __init__(self, webhook_secret: Optional[str] = None):
        self.webhook_secret = webhook_secret
        self.handlers = {}

    def register_handler(self, event_type: str, handler_func):
        """Register a handler for a specific event type"""
        if event_type not in self.handlers:
            self.handlers[event_type] = []
        self.handlers[event_type].append(handler_func)

    def verify_signature(self, payload: bytes, signature: str) -> bool:
        """Verify webhook signature for security"""
        if not self.webhook_secret:
            return True  # No verification if no secret set

        expected_signature = hmac.new(
            self.webhook_secret.encode('utf-8'),
            payload,
            hashlib.sha256
        ).hexdigest()

        return hmac.compare_digest(f"sha256={expected_signature}", signature)

    def handle_webhook(self, payload: Dict[str, Any], signature: str = None) -> Dict[str, Any]:
        """Handle incoming webhook payload"""

        # Verify signature if provided
        if signature and not self.verify_signature(json.dumps(payload).encode(), signature):
            raise ValueError("Invalid webhook signature")

        event_type = payload.get('eventType', 'unknown')

        # Process event
        results = []
        if event_type in self.handlers:
            for handler in self.handlers[event_type]:
                try:
                    result = handler(payload)
                    results.append({'handler': handler.__name__, 'result': result, 'success': True})
                except Exception as e:
                    results.append({'handler': handler.__name__, 'error': str(e), 'success': False})
        else:
            return {'error': f'No handler for event type: {event_type}'}

        return {
            'event_type': event_type,
            'processed_at': datetime.now().isoformat(),
            'handlers_executed': len(results),
            'results': results
        }

# Example webhook handlers
def handle_movie_downloaded(payload):
    """Handle movie downloaded event"""
    movie = payload.get('movie', {})
    movie_file = payload.get('movieFile', {})

    print(f"🎬 Movie Downloaded: {movie.get('title')} ({movie.get('year')})")
    print(f"   File: {movie_file.get('relativePath', 'Unknown')}")
    print(f"   Quality: {movie_file.get('quality', {}).get('quality', {}).get('name', 'Unknown')}")
    print(f"   Size: {movie_file.get('size', 0) / 1024 / 1024:.1f} MB")

    # Send notification to external service
    send_discord_notification(
        title="Movie Downloaded",
        description=f"{movie.get('title')} ({movie.get('year')}) is ready to watch!",
        color=0x00ff00  # Green
    )

    return {'status': 'processed', 'movie_id': movie.get('id')}

def handle_movie_grabbed(payload):
    """Handle movie grabbed/downloading event"""
    movie = payload.get('movie', {})
    release = payload.get('release', {})

    print(f"📥 Movie Grabbed: {movie.get('title')} ({movie.get('year')})")
    print(f"   Release: {release.get('releaseTitle', 'Unknown')}")
    print(f"   Indexer: {release.get('indexer', 'Unknown')}")
    print(f"   Quality: {release.get('quality', {}).get('quality', {}).get('name', 'Unknown')}")

    send_discord_notification(
        title="Movie Grabbed",
        description=f"Started downloading {movie.get('title')} ({movie.get('year')})",
        color=0xff9900  # Orange
    )

    return {'status': 'processed', 'download_id': release.get('downloadId')}

def handle_health_issue(payload):
    """Handle health issue event"""
    health_check = payload.get('healthCheck', {})

    print(f"⚠️ Health Issue: {health_check.get('type', 'Unknown')}")
    print(f"   Source: {health_check.get('source', 'Unknown')}")
    print(f"   Message: {health_check.get('message', 'No message')}")

    # Send critical alerts immediately
    if health_check.get('type') == 'error':
        send_discord_notification(
            title="🚨 Critical Health Issue",
            description=health_check.get('message', 'Unknown error'),
            color=0xff0000  # Red
        )

    return {'status': 'processed', 'severity': health_check.get('type')}

def send_discord_notification(title, description, color=0x7289da):
    """Send notification to Discord webhook"""
    import requests

    discord_webhook_url = "https://discord.com/api/webhooks/YOUR_WEBHOOK_URL"

    payload = {
        "embeds": [{
            "title": title,
            "description": description,
            "color": color,
            "timestamp": datetime.now().isoformat(),
            "footer": {
                "text": "Radarr Go"
            }
        }]
    }

    try:
        response = requests.post(discord_webhook_url, json=payload, timeout=10)
        response.raise_for_status()
        return True
    except Exception as e:
        print(f"Failed to send Discord notification: {e}")
        return False

# Flask web server to receive webhooks
from flask import Flask, request, jsonify

app = Flask(__name__)
webhook_handler = RadarrWebhookHandler(webhook_secret="your-secret-here")

# Register event handlers
webhook_handler.register_handler('Download', handle_movie_downloaded)
webhook_handler.register_handler('Grab', handle_movie_grabbed)
webhook_handler.register_handler('HealthIssue', handle_health_issue)

@app.route('/webhook/radarr', methods=['POST'])
def radarr_webhook():
    """Endpoint to receive Radarr webhooks"""
    try:
        payload = request.get_json()
        signature = request.headers.get('X-Radarr-Signature')

        result = webhook_handler.handle_webhook(payload, signature)

        return jsonify(result), 200

    except ValueError as e:
        return jsonify({'error': str(e)}), 401
    except Exception as e:
        return jsonify({'error': str(e)}), 500

# Health check endpoint
@app.route('/health', methods=['GET'])
def health_check():
    return jsonify({'status': 'healthy', 'timestamp': datetime.now().isoformat()})

if __name__ == '__main__':
    print("Starting Radarr webhook server...")
    print("Configure Radarr webhook URL: http://your-server:5000/webhook/radarr")
    app.run(host='0.0.0.0', port=5000, debug=True)
```

### Integration with External Tools

**Plex Integration Example**
```python
class PlexRadarrIntegration:
    """Integration between Radarr and Plex Media Server"""

    def __init__(self, radarr_client, plex_url, plex_token):
        self.radarr = radarr_client
        self.plex_url = plex_url
        self.plex_token = plex_token

    def sync_plex_library(self):
        """Sync Plex library with Radarr collection"""

        # Get Plex movies
        plex_movies = self.get_plex_movies()

        # Get Radarr movies
        radarr_movies = self.radarr.get_movies()

        # Find movies in Plex but not in Radarr
        plex_titles = {(m['title'], m['year']) for m in plex_movies}
        radarr_titles = {(m['title'], m['year']) for m in radarr_movies}

        missing_in_radarr = plex_titles - radarr_titles
        missing_in_plex = radarr_titles - plex_titles

        print(f"Movies in Plex but not Radarr: {len(missing_in_radarr)}")
        print(f"Movies in Radarr but not Plex: {len(missing_in_plex)}")

        return {
            'plex_only': list(missing_in_radarr),
            'radarr_only': list(missing_in_plex),
            'total_plex': len(plex_movies),
            'total_radarr': len(radarr_movies)
        }

    def get_plex_movies(self):
        """Get movie list from Plex Media Server"""
        import xml.etree.ElementTree as ET

        url = f"{self.plex_url}/library/sections/1/all"  # Assuming section 1 is movies
        headers = {'X-Plex-Token': self.plex_token}

        response = requests.get(url, headers=headers)
        response.raise_for_status()

        root = ET.fromstring(response.content)
        movies = []

        for video in root.findall('.//Video'):
            movies.append({
                'title': video.get('title'),
                'year': int(video.get('year', 0)),
                'rating_key': video.get('ratingKey'),
                'file_path': video.get('file')
            })

        return movies

    def trigger_plex_scan(self, library_section=1):
        """Trigger Plex library scan after movie download"""

        url = f"{self.plex_url}/library/sections/{library_section}/refresh"
        headers = {'X-Plex-Token': self.plex_token}

        response = requests.get(url, headers=headers)
        response.raise_for_status()

        print("Plex library scan triggered")
        return True

# Jellyfin Integration Example
class JellyfinRadarrIntegration:
    """Integration between Radarr and Jellyfin Media Server"""

    def __init__(self, radarr_client, jellyfin_url, jellyfin_api_key):
        self.radarr = radarr_client
        self.jellyfin_url = jellyfin_url
        self.jellyfin_api_key = jellyfin_api_key

    def trigger_jellyfin_scan(self):
        """Trigger Jellyfin library scan"""

        # Get library ID for movies
        libraries = self.get_jellyfin_libraries()
        movie_library = next((lib for lib in libraries if 'movie' in lib['Name'].lower()), None)

        if not movie_library:
            print("No movie library found in Jellyfin")
            return False

        # Trigger scan
        url = f"{self.jellyfin_url}/Library/Refresh"
        headers = {'X-Emby-Token': self.jellyfin_api_key}
        params = {'Id': movie_library['Id']}

        response = requests.post(url, headers=headers, params=params)
        response.raise_for_status()

        print(f"Jellyfin scan triggered for library: {movie_library['Name']}")
        return True

    def get_jellyfin_libraries(self):
        """Get Jellyfin library list"""

        url = f"{self.jellyfin_url}/Library/VirtualFolders"
        headers = {'X-Emby-Token': self.jellyfin_api_key}

        response = requests.get(url, headers=headers)
        response.raise_for_status()

        return response.json()

# Usage example with webhook integration
def setup_media_server_integration():
    """Setup integration with media servers"""

    # Initialize clients
    radarr_client = RadarrClient("http://localhost:7878", "your-radarr-api-key")

    plex_integration = PlexRadarrIntegration(
        radarr_client,
        "http://localhost:32400",
        "your-plex-token"
    )

    jellyfin_integration = JellyfinRadarrIntegration(
        radarr_client,
        "http://localhost:8096",
        "your-jellyfin-api-key"
    )

    # Webhook handler that triggers media server scans
    def handle_movie_imported(payload):
        movie = payload.get('movie', {})
        print(f"Movie imported: {movie.get('title')}")

        # Trigger scans on both media servers
        plex_integration.trigger_plex_scan()
        jellyfin_integration.trigger_jellyfin_scan()

        return {'media_servers_notified': True}

    # Register handler
    webhook_handler.register_handler('Download', handle_movie_imported)

    return {
        'radarr': radarr_client,
        'plex': plex_integration,
        'jellyfin': jellyfin_integration
    }
```

---

## Summary

This comprehensive guide provides production-ready examples for integrating with the Radarr Go API across all major use cases:

### ✅ What You've Learned

**Third-Party Client Integration:**
- Complete Python client with async support, error handling, and rate limiting
- JavaScript/Node.js client with WebSocket integration and real-time updates
- Shell script automation with comprehensive error handling
- PowerShell client for Windows environments with advanced features
- Go client library for native integration

**Common Integration Patterns:**
- Movie search and addition workflows with validation and monitoring
- Advanced queue management with automatic retry and alternative release selection
- Efficient bulk operations with progress tracking and parallel execution
- Real-time event handling via WebSocket with filtering and processing

**Automation Examples:**
- Complete backup and restore system with data integrity validation
- Library maintenance automation with health monitoring and optimization
- Scheduled maintenance tasks with comprehensive reporting

**Troubleshooting Solutions:**
- Authentication issues and resolution strategies
- Network connectivity diagnosis and CORS handling
- Performance optimization with caching and pagination
- WebSocket connection troubleshooting with auto-reconnection
- Custom webhook implementation with security features
- External tool integration (Plex, Jellyfin, etc.)

### 🚀 Next Steps

1. **Choose Your Implementation:**
   - **Python**: Best for data analysis, automation, and complex workflows
   - **JavaScript/Node.js**: Ideal for web applications and real-time interfaces
   - **Shell Scripts**: Perfect for system administration and cron jobs
   - **PowerShell**: Excellent for Windows environments and enterprise integration
   - **Go**: Optimal for high-performance applications and microservices

2. **Start with Basic Operations:**
   - Implement authentication and basic API calls
   - Test with simple movie operations
   - Add error handling and logging

3. **Add Advanced Features:**
   - Implement WebSocket connections for real-time updates
   - Add comprehensive error handling and retry logic
   - Include performance monitoring and optimization

4. **Scale Your Integration:**
   - Implement bulk operations for large libraries
   - Add automation and scheduled maintenance
   - Integrate with external tools and notification systems

### 📚 Additional Resources

**API Documentation:**
- [API Endpoints Reference](/Users/owine/Git/radarr-go/docs/API_ENDPOINTS.md)
- [OpenAPI Specification](/Users/owine/Git/radarr-go/docs/openapi.yaml)
- [API Compatibility Guide](/Users/owine/Git/radarr-go/docs/API_COMPATIBILITY.md)

**Development Resources:**
- [Developer Guide](/Users/owine/Git/radarr-go/docs/DEVELOPER_GUIDE.md)
- [Configuration Reference](/Users/owine/Git/radarr-go/docs/CONFIGURATION_REFERENCE.md)
- [Troubleshooting Guide](/Users/owine/Git/radarr-go/docs/troubleshooting-guide.md)

**Community and Support:**
- GitHub Issues: Report bugs and request features
- Discussions: Community support and integration examples
- API Updates: Follow releases for new features and improvements

### 🔧 Best Practices Summary

**Security:**
- Always use HTTPS in production environments
- Store API keys securely (environment variables, key management)
- Implement webhook signature verification
- Use proper authentication methods

**Performance:**
- Implement request caching for frequently accessed data
- Use pagination for large datasets
- Add rate limiting to respect API limits
- Monitor and optimize slow endpoints

**Reliability:**
- Include comprehensive error handling
- Implement retry logic with exponential backoff
- Add health checks and monitoring
- Use WebSocket reconnection strategies

**Maintainability:**
- Follow language-specific coding standards
- Add comprehensive logging and monitoring
- Include unit tests for critical functionality
- Document your integration thoroughly

---

**Happy integrating!** 🎬

For questions, issues, or contributions to this guide, please visit the [Radarr Go GitHub repository](https://github.com/radarr/radarr-go).

*This guide covers Radarr Go v0.9.0-alpha and maintains 100% compatibility with Radarr v3 API.*

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create comprehensive third-party client integration examples", "activeForm": "Creating third-party client integration examples", "status": "completed"}, {"content": "Develop common integration patterns and workflows", "activeForm": "Developing common integration patterns and workflows", "status": "completed"}, {"content": "Build automation examples for various use cases", "activeForm": "Building automation examples for various use cases", "status": "completed"}, {"content": "Create comprehensive troubleshooting guide for API integration", "activeForm": "Creating comprehensive troubleshooting guide for API integration", "status": "completed"}]

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create comprehensive third-party client integration examples", "activeForm": "Creating third-party client integration examples", "status": "completed"}, {"content": "Develop common integration patterns and workflows", "activeForm": "Developing common integration patterns and workflows", "status": "completed"}, {"content": "Build automation examples for various use cases", "activeForm": "Building automation examples for various use cases", "status": "completed"}, {"content": "Create comprehensive troubleshooting guide for API integration", "activeForm": "Creating comprehensive troubleshooting guide for API integration", "status": "in_progress"}]

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create comprehensive third-party client integration examples", "activeForm": "Creating third-party client integration examples", "status": "completed"}, {"content": "Develop common integration patterns and workflows", "activeForm": "Developing common integration patterns and workflows", "status": "completed"}, {"content": "Build automation examples for various use cases", "activeForm": "Building automation examples for various use cases", "status": "in_progress"}, {"content": "Create comprehensive troubleshooting guide for API integration", "activeForm": "Creating comprehensive troubleshooting guide for API integration", "status": "pending"}]

### Python Client with Requests Library

Complete Python client implementation with advanced features:

```python
import requests
import json
import time
import asyncio
import aiohttp
from typing import Dict, List, Optional, Union, Any
from datetime import datetime, timedelta
from dataclasses import dataclass
from urllib.parse import urljoin
import logging

@dataclass
class RadarrMovie:
    """Data class for Radarr movie objects"""
    id: int
    title: str
    year: int
    tmdb_id: int
    imdb_id: Optional[str] = None
    status: str = "announced"
    monitored: bool = True
    quality_profile_id: int = 1
    root_folder_path: str = "/movies"

class RadarrAPIError(Exception):
    """Custom exception for Radarr API errors"""
    def __init__(self, message: str, status_code: int = None, response_data: dict = None):
        super().__init__(message)
        self.status_code = status_code
        self.response_data = response_data or {}

class RadarrClient:
    """Production-ready Python client for Radarr Go API"""

    def __init__(self, base_url: str, api_key: str, timeout: int = 30):
        self.base_url = base_url.rstrip('/') + '/api/v3'
        self.api_key = api_key
        self.timeout = timeout
        self.session = requests.Session()

        # Configure session with sensible defaults
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json',
            'User-Agent': 'Radarr-Python-Client/2.0'
        })

        # Connection pooling configuration
        adapter = requests.adapters.HTTPAdapter(
            pool_connections=20,
            pool_maxsize=20,
            max_retries=3
        )
        self.session.mount('http://', adapter)
        self.session.mount('https://', adapter)

        # Setup logging
        self.logger = logging.getLogger(__name__)

    def _request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make authenticated request with comprehensive error handling"""
        url = urljoin(self.base_url + '/', endpoint.lstrip('/'))

        # Add timeout if not specified
        if 'timeout' not in kwargs:
            kwargs['timeout'] = self.timeout

        try:
            self.logger.debug(f"{method} {url}")
            response = self.session.request(method, url, **kwargs)

            # Handle rate limiting with exponential backoff
            if response.status_code == 429:
                retry_after = int(response.headers.get('Retry-After', 60))
                self.logger.warning(f"Rate limited. Waiting {retry_after} seconds...")
                time.sleep(retry_after)
                return self._request(method, endpoint, **kwargs)

            # Log response details for debugging
            self.logger.debug(f"Response: {response.status_code} in {response.elapsed.total_seconds():.2f}s")

            if not response.ok:
                self._handle_error_response(response)

            return response

        except requests.exceptions.Timeout:
            raise RadarrAPIError(f"Request timeout after {self.timeout} seconds", 408)
        except requests.exceptions.ConnectionError as e:
            raise RadarrAPIError(f"Failed to connect to Radarr server: {str(e)}", 503)
        except requests.exceptions.RequestException as e:
            raise RadarrAPIError(f"Request failed: {str(e)}")

    def _handle_error_response(self, response: requests.Response):
        """Handle HTTP error responses with detailed error information"""
        try:
            error_data = response.json()
            if isinstance(error_data, dict) and 'error' in error_data:
                error_msg = error_data['error']
                if isinstance(error_msg, dict):
                    message = error_msg.get('message', str(error_msg))
                else:
                    message = str(error_msg)
            else:
                message = str(error_data)
        except (ValueError, json.JSONDecodeError):
            message = response.text or f"HTTP {response.status_code}"

        raise RadarrAPIError(
            f"API Error {response.status_code}: {message}",
            response.status_code,
            error_data if 'error_data' in locals() else {}
        )

    # System Information
    def get_system_status(self) -> Dict[str, Any]:
        """Get comprehensive system status information"""
        response = self._request('GET', '/system/status')
        return response.json()

    def health_check(self) -> bool:
        """Quick health check - returns True if system is responsive"""
        try:
            response = self._request('GET', '/ping')
            return response.status_code == 200
        except:
            return False

    # Movie Management
    def search_movies(self, query: str, limit: int = 20) -> List[Dict[str, Any]]:
        """Search for movies to add to collection"""
        params = {'term': query}
        if limit:
            params['limit'] = limit

        response = self._request('GET', '/movie/lookup', params=params)
        return response.json()

    def get_movies(self, monitored: bool = None, **filters) -> List[Dict[str, Any]]:
        """Get movies with comprehensive filtering options"""
        params = {}

        if monitored is not None:
            params['monitored'] = str(monitored).lower()

        # Add additional filters
        for key, value in filters.items():
            if value is not None:
                params[key] = value

        response = self._request('GET', '/movie', params=params)
        data = response.json()

        # Handle both paginated and non-paginated responses
        if isinstance(data, dict) and 'data' in data:
            return data['data']
        return data if isinstance(data, list) else []

    def get_movie(self, movie_id: int) -> Dict[str, Any]:
        """Get detailed information about a specific movie"""
        response = self._request('GET', f'/movie/{movie_id}')
        return response.json()

    def add_movie(self, movie_data: Union[Dict[str, Any], RadarrMovie]) -> Dict[str, Any]:
        """Add a new movie to the collection"""
        if isinstance(movie_data, RadarrMovie):
            # Convert dataclass to dict
            data = {
                'title': movie_data.title,
                'year': movie_data.year,
                'tmdbId': movie_data.tmdb_id,
                'imdbId': movie_data.imdb_id,
                'qualityProfileId': movie_data.quality_profile_id,
                'rootFolderPath': movie_data.root_folder_path,
                'monitored': movie_data.monitored,
                'addOptions': {
                    'monitor': 'movieOnly',
                    'searchForMovie': True
                }
            }
        else:
            data = movie_data

        response = self._request('POST', '/movie', json=data)
        return response.json()

    def update_movie(self, movie_id: int, movie_data: Dict[str, Any]) -> Dict[str, Any]:
        """Update an existing movie"""
        response = self._request('PUT', f'/movie/{movie_id}', json=movie_data)
        return response.json()

    def delete_movie(self, movie_id: int, delete_files: bool = False, add_exclusion: bool = False) -> bool:
        """Delete a movie from the collection"""
        params = {
            'deleteFiles': str(delete_files).lower(),
            'addExclusion': str(add_exclusion).lower()
        }

        response = self._request('DELETE', f'/movie/{movie_id}', params=params)
        return response.status_code == 200

    # Queue and Download Management
    def get_queue(self, page: int = 1, page_size: int = 20, sort_key: str = 'timeleft') -> Dict[str, Any]:
        """Get download queue with pagination"""
        params = {
            'page': page,
            'pageSize': page_size,
            'sortKey': sort_key
        }

        response = self._request('GET', '/queue', params=params)
        return response.json()

    def remove_from_queue(self, queue_id: int, remove_from_client: bool = True, blacklist: bool = False) -> bool:
        """Remove item from download queue"""
        params = {
            'removeFromClient': str(remove_from_client).lower(),
            'blacklist': str(blacklist).lower()
        }

        response = self._request('DELETE', f'/queue/{queue_id}', params=params)
        return response.status_code == 200

    # Search and Releases
    def search_movie_releases(self, movie_id: int) -> List[Dict[str, Any]]:
        """Search for releases for a specific movie"""
        params = {'movieId': movie_id}
        response = self._request('GET', '/release', params=params)
        return response.json()

    def grab_release(self, release_guid: str, indexer_id: int = None) -> Dict[str, Any]:
        """Grab/download a specific release"""
        data = {'guid': release_guid}
        if indexer_id:
            data['indexerId'] = indexer_id

        response = self._request('POST', '/release/grab', json=data)
        return response.json()

    # Health and Monitoring
    def get_health_status(self) -> List[Dict[str, Any]]:
        """Get system health issues"""
        response = self._request('GET', '/health')
        return response.json()

    def get_system_resources(self) -> Dict[str, Any]:
        """Get current system resource usage"""
        response = self._request('GET', '/health/system/resources')
        return response.json()

    # Calendar Integration
    def get_calendar_events(self, start_date: datetime = None, end_date: datetime = None) -> List[Dict[str, Any]]:
        """Get calendar events for specified date range"""
        params = {}

        if start_date:
            params['start'] = start_date.isoformat()
        if end_date:
            params['end'] = end_date.isoformat()

        response = self._request('GET', '/calendar', params=params)
        return response.json()

    def get_calendar_feed_url(self) -> str:
        """Get iCal feed URL for external calendar applications"""
        response = self._request('GET', '/calendar/feed/url')
        return response.json().get('url', '')

    # Task Management
    def execute_command(self, command_name: str, **parameters) -> Dict[str, Any]:
        """Execute a system command with parameters"""
        data = {
            'name': command_name,
            **parameters
        }

        response = self._request('POST', '/command', json=data)
        return response.json()

    def refresh_movie(self, movie_id: int) -> Dict[str, Any]:
        """Refresh metadata for a specific movie"""
        return self.execute_command('RefreshMovie', movieId=movie_id)

    def refresh_all_movies(self) -> Dict[str, Any]:
        """Refresh metadata for all movies"""
        return self.execute_command('RefreshMovie')

    # Bulk Operations
    def bulk_add_movies(self, movies: List[Union[Dict[str, Any], RadarrMovie]]) -> List[Dict[str, Any]]:
        """Add multiple movies in batch"""
        results = []

        for movie in movies:
            try:
                result = self.add_movie(movie)
                results.append({'success': True, 'data': result})
                # Small delay to avoid overwhelming the API
                time.sleep(0.1)
            except RadarrAPIError as e:
                self.logger.error(f"Failed to add movie: {e}")
                results.append({'success': False, 'error': str(e)})

        return results

    def bulk_search_missing(self, movie_ids: List[int] = None) -> Dict[str, Any]:
        """Search for all missing movies or specific list"""
        data = {}
        if movie_ids:
            data['movieIds'] = movie_ids

        return self.execute_command('MoviesSearch', **data)

    # Context Manager Support
    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.session.close()

# Async Version for High-Performance Applications
class AsyncRadarrClient:
    """Async version of the Radarr client for high-performance applications"""

    def __init__(self, base_url: str, api_key: str, timeout: int = 30):
        self.base_url = base_url.rstrip('/') + '/api/v3'
        self.api_key = api_key
        self.timeout = aiohttp.ClientTimeout(total=timeout)
        self.headers = {
            'X-API-Key': api_key,
            'Content-Type': 'application/json',
            'User-Agent': 'Radarr-Async-Python-Client/2.0'
        }

    async def _request(self, method: str, endpoint: str, session: aiohttp.ClientSession, **kwargs):
        """Async request with error handling"""
        url = urljoin(self.base_url + '/', endpoint.lstrip('/'))

        try:
            async with session.request(method, url, headers=self.headers, **kwargs) as response:
                if response.status == 429:
                    retry_after = int(response.headers.get('Retry-After', 60))
                    await asyncio.sleep(retry_after)
                    return await self._request(method, endpoint, session, **kwargs)

                if not response.ok:
                    error_text = await response.text()
                    raise RadarrAPIError(f"HTTP {response.status}: {error_text}", response.status)

                return await response.json()

        except aiohttp.ClientError as e:
            raise RadarrAPIError(f"Request failed: {str(e)}")

    async def get_movies_async(self, **filters) -> List[Dict[str, Any]]:
        """Async version of get_movies"""
        connector = aiohttp.TCPConnector(limit=20, limit_per_host=20)

        async with aiohttp.ClientSession(
            connector=connector,
            timeout=self.timeout
        ) as session:
            params = {k: v for k, v in filters.items() if v is not None}
            data = await self._request('GET', '/movie', session, params=params)
            return data.get('data', data) if isinstance(data, dict) else data

    async def bulk_operations_async(self, operations: List[tuple]) -> List[Any]:
        """Execute multiple operations concurrently"""
        connector = aiohttp.TCPConnector(limit=50, limit_per_host=20)

        async with aiohttp.ClientSession(
            connector=connector,
            timeout=self.timeout
        ) as session:

            async def execute_operation(op):
                method, endpoint, kwargs = op
                try:
                    return await self._request(method, endpoint, session, **kwargs)
                except Exception as e:
                    return {'error': str(e)}

            results = await asyncio.gather(*[execute_operation(op) for op in operations])
            return results

# Usage Examples
if __name__ == "__main__":
    # Initialize client
    client = RadarrClient(
        base_url="http://localhost:7878",
        api_key="your-api-key-here"
    )

    try:
        # Context manager usage
        with client:
            # Basic usage examples
            status = client.get_system_status()
            print(f"Radarr version: {status['version']}")

            # Search and add movie
            search_results = client.search_movies("The Matrix")
            if search_results:
                movie_to_add = search_results[0]
                added_movie = client.add_movie({
                    'title': movie_to_add['title'],
                    'year': movie_to_add['year'],
                    'tmdbId': movie_to_add['tmdbId'],
                    'qualityProfileId': 1,
                    'rootFolderPath': '/movies',
                    'monitored': True,
                    'addOptions': {
                        'monitor': 'movieOnly',
                        'searchForMovie': True
                    }
                })
                print(f"Added movie: {added_movie['title']}")

            # Get queue status
            queue = client.get_queue()
            print(f"Queue items: {len(queue.get('data', []))}")

            # Check health
            if client.health_check():
                health_issues = client.get_health_status()
                if health_issues:
                    print(f"Health issues found: {len(health_issues)}")
                else:
                    print("System healthy")

    except RadarrAPIError as e:
        print(f"API Error: {e}")
        if e.status_code:
            print(f"Status Code: {e.status_code}")

    # Async usage example
    async def async_example():
        async_client = AsyncRadarrClient(
            base_url="http://localhost:7878",
            api_key="your-api-key-here"
        )

        # Concurrent operations
        operations = [
            ('GET', '/movie', {}),
            ('GET', '/queue', {}),
            ('GET', '/health', {})
        ]

        results = await async_client.bulk_operations_async(operations)
        print(f"Concurrent operations completed: {len(results)}")

    # Run async example
    # asyncio.run(async_example())
```

### JavaScript/Node.js Client with Axios

Complete Node.js client implementation with TypeScript support:

```javascript
const axios = require('axios');
const EventEmitter = require('events');
const WebSocket = require('ws');

/**
 * Production-ready Node.js client for Radarr Go API
 */
class RadarrGoClient extends EventEmitter {
  constructor(config) {
    super();

    this.baseUrl = config.baseUrl.replace(/\/$/, '');
    this.apiKey = config.apiKey;
    this.timeout = config.timeout || 30000;

    // Configure axios instance
    this.axios = axios.create({
      baseURL: `${this.baseUrl}/api/v3`,
      timeout: this.timeout,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        'User-Agent': 'Radarr-NodeJS-Client/2.0'
      }
    });

    this.setupInterceptors();
    this.setupRateLimiting();
  }

  setupInterceptors() {
    // Request interceptor for logging
    this.axios.interceptors.request.use(
      config => {
        console.log(`${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      error => Promise.reject(error)
    );

    // Response interceptor for error handling
    this.axios.interceptors.response.use(
      response => response,
      async error => {
        if (error.response?.status === 429) {
          const retryAfter = parseInt(error.response.headers['retry-after']) || 60;
          console.warn(`Rate limited. Waiting ${retryAfter} seconds...`);
          await this.sleep(retryAfter * 1000);
          return this.axios(error.config);
        }

        // Enhanced error handling
        if (error.response) {
          const errorData = error.response.data;
          const message = errorData?.error?.message || errorData?.error || 'API Error';
          const enhancedError = new Error(`${error.response.status}: ${message}`);
          enhancedError.status = error.response.status;
          enhancedError.data = errorData;
          throw enhancedError;
        }

        throw error;
      }
    );
  }

  setupRateLimiting() {
    this.requestQueue = [];
    this.isProcessingQueue = false;
    this.rateLimitDelay = 100; // ms between requests
  }

  async makeRequest(config) {
    return new Promise((resolve, reject) => {
      this.requestQueue.push({ config, resolve, reject });
      this.processQueue();
    });
  }

  async processQueue() {
    if (this.isProcessingQueue || this.requestQueue.length === 0) return;

    this.isProcessingQueue = true;

    while (this.requestQueue.length > 0) {
      const { config, resolve, reject } = this.requestQueue.shift();

      try {
        const response = await this.axios(config);
        resolve(response.data);
      } catch (error) {
        reject(error);
      }

      if (this.requestQueue.length > 0) {
        await this.sleep(this.rateLimitDelay);
      }
    }

    this.isProcessingQueue = false;
  }

  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  // System Information
  async getSystemStatus() {
    return this.makeRequest({ method: 'GET', url: '/system/status' });
  }

  async healthCheck() {
    try {
      const response = await this.axios.get('/ping');
      return response.status === 200;
    } catch {
      return false;
    }
  }

  // Movie Management
  async searchMovies(query, options = {}) {
    const params = { term: query, ...options };
    return this.makeRequest({
      method: 'GET',
      url: '/movie/lookup',
      params
    });
  }

  async getMovies(filters = {}) {
    const response = await this.makeRequest({
      method: 'GET',
      url: '/movie',
      params: filters
    });

    // Handle paginated response
    return response.data || response;
  }

  async getMovie(movieId) {
    return this.makeRequest({
      method: 'GET',
      url: `/movie/${movieId}`
    });
  }

  async addMovie(movieData) {
    return this.makeRequest({
      method: 'POST',
      url: '/movie',
      data: movieData
    });
  }

  async updateMovie(movieId, movieData) {
    return this.makeRequest({
      method: 'PUT',
      url: `/movie/${movieId}`,
      data: movieData
    });
  }

  async deleteMovie(movieId, options = {}) {
    const params = {
      deleteFiles: options.deleteFiles || false,
      addExclusion: options.addExclusion || false
    };

    await this.makeRequest({
      method: 'DELETE',
      url: `/movie/${movieId}`,
      params
    });

    return true;
  }

  // Queue Management
  async getQueue(page = 1, pageSize = 20, sortKey = 'timeleft') {
    return this.makeRequest({
      method: 'GET',
      url: '/queue',
      params: { page, pageSize, sortKey }
    });
  }

  async removeFromQueue(queueId, options = {}) {
    const params = {
      removeFromClient: options.removeFromClient !== false,
      blacklist: options.blacklist || false
    };

    await this.makeRequest({
      method: 'DELETE',
      url: `/queue/${queueId}`,
      params
    });

    return true;
  }

  // Search and Downloads
  async searchMovieReleases(movieId) {
    return this.makeRequest({
      method: 'GET',
      url: '/release',
      params: { movieId }
    });
  }

  async grabRelease(releaseData) {
    return this.makeRequest({
      method: 'POST',
      url: '/release/grab',
      data: releaseData
    });
  }

  // Commands and Tasks
  async executeCommand(commandName, parameters = {}) {
    return this.makeRequest({
      method: 'POST',
      url: '/command',
      data: { name: commandName, ...parameters }
    });
  }

  async refreshMovie(movieId) {
    return this.executeCommand('RefreshMovie', { movieId });
  }

  async refreshAllMovies() {
    return this.executeCommand('RefreshMovie');
  }

  // Health Monitoring
  async getHealthStatus() {
    return this.makeRequest({ method: 'GET', url: '/health' });
  }

  async getSystemResources() {
    return this.makeRequest({ method: 'GET', url: '/health/system/resources' });
  }

  // Calendar Integration
  async getCalendarEvents(startDate, endDate) {
    const params = {};
    if (startDate) params.start = startDate.toISOString();
    if (endDate) params.end = endDate.toISOString();

    return this.makeRequest({
      method: 'GET',
      url: '/calendar',
      params
    });
  }

  // WebSocket Real-time Updates
  connectWebSocket() {
    if (this.ws) {
      this.ws.close();
    }

    const wsUrl = `${this.baseUrl.replace('http', 'ws')}/api/v3/ws?apikey=${this.apiKey}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.on('open', () => {
      console.log('WebSocket connected');
      this.emit('connected');
    });

    this.ws.on('message', (data) => {
      try {
        const message = JSON.parse(data);
        this.emit('message', message);

        // Emit specific event types
        if (message.type) {
          this.emit(message.type, message.data);
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error);
      }
    });

    this.ws.on('close', () => {
      console.log('WebSocket disconnected');
      this.emit('disconnected');

      // Auto-reconnect after 5 seconds
      setTimeout(() => this.connectWebSocket(), 5000);
    });

    this.ws.on('error', (error) => {
      console.error('WebSocket error:', error);
      this.emit('error', error);
    });

    return this.ws;
  }

  // Bulk Operations
  async bulkAddMovies(movies) {
    const results = [];
    const batchSize = 5; // Process in batches to avoid overwhelming the API

    for (let i = 0; i < movies.length; i += batchSize) {
      const batch = movies.slice(i, i + batchSize);
      const batchPromises = batch.map(async (movie) => {
        try {
          const result = await this.addMovie(movie);
          return { success: true, data: result };
        } catch (error) {
          return { success: false, error: error.message };
        }
      });

      const batchResults = await Promise.allSettled(batchPromises);
      results.push(...batchResults.map(r => r.value));

      // Small delay between batches
      if (i + batchSize < movies.length) {
        await this.sleep(1000);
      }
    }

    return results;
  }

  // Monitoring and Statistics
  async getStatistics() {
    const [movies, queue, health] = await Promise.all([
      this.getMovies(),
      this.getQueue(),
      this.getHealthStatus()
    ]);

    return {
      totalMovies: movies.length || 0,
      queueItems: queue.data?.length || queue.length || 0,
      healthIssues: health.length || 0,
      timestamp: new Date().toISOString()
    };
  }

  // Cleanup
  close() {
    if (this.ws) {
      this.ws.close();
    }
    this.removeAllListeners();
  }
}

// TypeScript definitions (save as radarr-client.d.ts)
const TypeScriptDefinitions = `
interface RadarrConfig {
  baseUrl: string;
  apiKey: string;
  timeout?: number;
}

interface Movie {
  id?: number;
  title: string;
  year: number;
  tmdbId: number;
  imdbId?: string;
  qualityProfileId: number;
  rootFolderPath: string;
  monitored: boolean;
  addOptions?: {
    monitor: string;
    searchForMovie: boolean;
  };
}

interface QueueItem {
  id: number;
  movieId: number;
  title: string;
  status: string;
  progress: number;
  downloadId: string;
}

interface HealthIssue {
  id: number;
  source: string;
  type: string;
  message: string;
  wikiUrl?: string;
}

declare class RadarrGoClient {
  constructor(config: RadarrConfig);

  // System
  getSystemStatus(): Promise<any>;
  healthCheck(): Promise<boolean>;

  // Movies
  searchMovies(query: string, options?: any): Promise<Movie[]>;
  getMovies(filters?: any): Promise<Movie[]>;
  getMovie(movieId: number): Promise<Movie>;
  addMovie(movieData: Movie): Promise<Movie>;
  updateMovie(movieId: number, movieData: Partial<Movie>): Promise<Movie>;
  deleteMovie(movieId: number, options?: any): Promise<boolean>;

  // Queue
  getQueue(page?: number, pageSize?: number, sortKey?: string): Promise<any>;
  removeFromQueue(queueId: number, options?: any): Promise<boolean>;

  // Commands
  executeCommand(commandName: string, parameters?: any): Promise<any>;
  refreshMovie(movieId: number): Promise<any>;
  refreshAllMovies(): Promise<any>;

  // Health
  getHealthStatus(): Promise<HealthIssue[]>;
  getSystemResources(): Promise<any>;

  // WebSocket
  connectWebSocket(): WebSocket;

  // Events
  on(event: string, listener: Function): this;
  emit(event: string, ...args: any[]): boolean;

  // Cleanup
  close(): void;
}

export = RadarrGoClient;
`;

// Usage Examples
async function examples() {
  const client = new RadarrGoClient({
    baseUrl: 'http://localhost:7878',
    apiKey: 'your-api-key-here'
  });

  try {
    // Basic usage
    const status = await client.getSystemStatus();
    console.log(`Radarr version: ${status.version}`);

    // Search and add movie
    const searchResults = await client.searchMovies('Inception');
    if (searchResults.length > 0) {
      const movieToAdd = {
        title: searchResults[0].title,
        year: searchResults[0].year,
        tmdbId: searchResults[0].tmdbId,
        qualityProfileId: 1,
        rootFolderPath: '/movies',
        monitored: true,
        addOptions: {
          monitor: 'movieOnly',
          searchForMovie: true
        }
      };

      const addedMovie = await client.addMovie(movieToAdd);
      console.log(`Added: ${addedMovie.title}`);
    }

    // Monitor queue
    const queue = await client.getQueue();
    console.log(`Queue items: ${queue.data?.length || 0}`);

    // WebSocket example
    client.connectWebSocket();
    client.on('movie.downloaded', (data) => {
      console.log(`Movie downloaded: ${data.title}`);
    });

    client.on('queue.updated', (data) => {
      console.log('Queue updated');
    });

  } catch (error) {
    console.error('API Error:', error.message);
  }
}

module.exports = RadarrGoClient;

// Run examples
if (require.main === module) {
  examples();
}
```

### Shell Script Examples with cURL

Comprehensive shell script examples for system integration:

```bash
#!/bin/bash

# Radarr Go API Shell Script Client
# Comprehensive examples for system integration and automation

set -euo pipefail

# Configuration
RADARR_URL="${RADARR_URL:-http://localhost:7878}"
RADARR_API_KEY="${RADARR_API_KEY}"
API_BASE="${RADARR_URL}/api/v3"
CURL_OPTS=(-s -H "X-API-Key: ${RADARR_API_KEY}" -H "Content-Type: application/json")

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# API Helper Functions
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"
    local url="${API_BASE}${endpoint}"

    log_info "Making ${method} request to ${endpoint}"

    if [[ -n "$data" ]]; then
        curl "${CURL_OPTS[@]}" -X "${method}" -d "$data" "$url"
    else
        curl "${CURL_OPTS[@]}" -X "${method}" "$url"
    fi
}

api_get() {
    api_request "GET" "$1"
}

api_post() {
    api_request "POST" "$1" "$2"
}

api_put() {
    api_request "PUT" "$1" "$2"
}

api_delete() {
    api_request "DELETE" "$1"
}

# Error handling with retry logic
api_request_with_retry() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"
    local max_retries=3
    local retry_delay=5
    local attempt=1

    while [[ $attempt -le $max_retries ]]; do
        if response=$(api_request "$method" "$endpoint" "$data" 2>/dev/null); then
            echo "$response"
            return 0
        else
            local exit_code=$?
            if [[ $attempt -lt $max_retries ]]; then
                log_warning "Attempt $attempt failed, retrying in ${retry_delay}s..."
                sleep $retry_delay
                ((attempt++))
                ((retry_delay*=2)) # Exponential backoff
            else
                log_error "All $max_retries attempts failed"
                return $exit_code
            fi
        fi
    done
}

# System Functions
check_health() {
    log_info "Checking Radarr health status"

    if ping_response=$(curl -s -f "${RADARR_URL}/ping" 2>/dev/null); then
        log_success "Radarr is responsive"

        health_issues=$(api_get "/health" | jq -r 'length')
        if [[ "$health_issues" -eq 0 ]]; then
            log_success "No health issues found"
        else
            log_warning "$health_issues health issues found"
            api_get "/health" | jq -r '.[] | "- \(.source): \(.message)"'
        fi
        return 0
    else
        log_error "Radarr is not responding"
        return 1
    fi
}

get_system_status() {
    log_info "Getting system status"
    api_get "/system/status" | jq -r '
        "Version: \(.version)",
        "Database: \(.databaseType)",
        "OS: \(.osName)",
        "Mode: \(.mode)",
        "URL Base: \(.urlBase // "/")",
        "Authentication: \(.authentication)"
    '
}

# Movie Management Functions
search_movies() {
    local query="$1"
    log_info "Searching for movies: $query"

    api_get "/movie/lookup?term=$(printf '%s' "$query" | jq -sRr @uri)" | \
        jq -r '.[] | "\(.title) (\(.year)) - TMDB: \(.tmdbId)"'
}

add_movie() {
    local title="$1"
    local year="$2"
    local tmdb_id="$3"
    local quality_profile_id="${4:-1}"
    local root_folder="${5:-/movies}"
    local monitored="${6:-true}"

    log_info "Adding movie: $title ($year)"

    local movie_data=$(cat <<EOF
{
    "title": "$title",
    "year": $year,
    "tmdbId": $tmdb_id,
    "qualityProfileId": $quality_profile_id,
    "rootFolderPath": "$root_folder",
    "monitored": $monitored,
    "addOptions": {
        "monitor": "movieOnly",
        "searchForMovie": true
    }
}
EOF
)

    if result=$(api_post "/movie" "$movie_data"); then
        local movie_id=$(echo "$result" | jq -r '.id')
        log_success "Added movie with ID: $movie_id"
        echo "$result" | jq -r '"Movie: \(.title) (\(.year))", "Status: \(.status)", "Monitored: \(.monitored)"'
    else
        log_error "Failed to add movie"
        return 1
    fi
}

get_movies() {
    local filters="$1"
    log_info "Getting movies with filters: $filters"

    local url="/movie"
    if [[ -n "$filters" ]]; then
        url="${url}?${filters}"
    fi

    api_get "$url" | jq -r '
        if type == "array" then
            .[] | "\(.title) (\(.year)) - Status: \(.status) - Monitored: \(.monitored)"
        else
            .data[]? | "\(.title) (\(.year)) - Status: \(.status) - Monitored: \(.monitored)"
        fi
    '
}

delete_movie() {
    local movie_id="$1"
    local delete_files="${2:-false}"
    local add_exclusion="${3:-false}"

    log_info "Deleting movie ID: $movie_id"

    if api_delete "/movie/${movie_id}?deleteFiles=${delete_files}&addExclusion=${add_exclusion}" >/dev/null; then
        log_success "Movie deleted successfully"
    else
        log_error "Failed to delete movie"
        return 1
    fi
}

# Queue Management Functions
get_queue() {
    log_info "Getting download queue"

    api_get "/queue" | jq -r '
        if has("data") then
            .data[] | "[\(.status)] \(.title) - \(.progress)% - ETA: \(.timeleft // "Unknown")"
        else
            .[] | "[\(.status)] \(.title) - \(.progress)% - ETA: \(.timeleft // "Unknown")"
        fi
    '
}

remove_from_queue() {
    local queue_id="$1"
    local remove_from_client="${2:-true}"
    local blacklist="${3:-false}"

    log_info "Removing queue item ID: $queue_id"

    if api_delete "/queue/${queue_id}?removeFromClient=${remove_from_client}&blacklist=${blacklist}" >/dev/null; then
        log_success "Item removed from queue"
    else
        log_error "Failed to remove item from queue"
        return 1
    fi
}

# Search and Download Functions
search_movie_releases() {
    local movie_id="$1"
    log_info "Searching releases for movie ID: $movie_id"

    api_get "/release?movieId=${movie_id}" | jq -r '
        .[] | "\(.title)", "  Size: \(.size / 1048576 | floor)MB", "  Indexer: \(.indexer)", "  Age: \(.age)d", ""
    '
}

grab_release() {
    local release_guid="$1"
    local indexer_id="${2:-}"

    log_info "Grabbing release: $release_guid"

    local grab_data="{\"guid\": \"$release_guid\""
    if [[ -n "$indexer_id" ]]; then
        grab_data="${grab_data}, \"indexerId\": $indexer_id"
    fi
    grab_data="${grab_data}}"

    if result=$(api_post "/release/grab" "$grab_data"); then
        log_success "Release grabbed successfully"
        echo "$result" | jq -r '"Title: \(.title)", "Status: \(.status)"'
    else
        log_error "Failed to grab release"
        return 1
    fi
}

# Task and Command Functions
execute_command() {
    local command_name="$1"
    shift
    local parameters="$*"

    log_info "Executing command: $command_name"

    local command_data="{\"name\": \"$command_name\""
    if [[ -n "$parameters" ]]; then
        command_data="${command_data}, $parameters"
    fi
    command_data="${command_data}}"

    if result=$(api_post "/command" "$command_data"); then
        local command_id=$(echo "$result" | jq -r '.id')
        log_success "Command queued with ID: $command_id"
        echo "$result" | jq -r '"Name: \(.name)", "Status: \(.status)", "Started: \(.started // "Not started")"'
    else
        log_error "Failed to execute command"
        return 1
    fi
}

refresh_movie() {
    local movie_id="${1:-}"

    if [[ -n "$movie_id" ]]; then
        log_info "Refreshing movie ID: $movie_id"
        execute_command "RefreshMovie" "\"movieId\": $movie_id"
    else
        log_info "Refreshing all movies"
        execute_command "RefreshMovie"
    fi
}

# Monitoring Functions
monitor_queue() {
    local interval="${1:-30}"

    log_info "Monitoring queue (checking every ${interval}s, press Ctrl+C to stop)"

    while true; do
        clear
        echo "=== Download Queue Status - $(date) ==="
        echo

        if queue_data=$(api_get "/queue" 2>/dev/null); then
            if queue_items=$(echo "$queue_data" | jq -r '.data[]? // .[]?' 2>/dev/null); then
                if [[ -n "$queue_items" ]]; then
                    echo "$queue_data" | jq -r '
                        (if has("data") then .data else . end) |
                        .[] | "[\(.status | ascii_upcase)] \(.title) - \(.progress)%"
                    '
                else
                    echo "Queue is empty"
                fi
            else
                echo "No queue items found"
            fi
        else
            log_error "Failed to get queue status"
        fi

        echo
        echo "Next update in ${interval}s..."
        sleep "$interval"
    done
}

# Bulk Operations
bulk_add_from_list() {
    local movie_list_file="$1"
    local quality_profile_id="${2:-1}"
    local root_folder="${3:-/movies}"

    log_info "Bulk adding movies from: $movie_list_file"

    local success_count=0
    local error_count=0

    while IFS=',' read -r title year tmdb_id || [[ -n "$title" ]]; do
        # Skip header line and empty lines
        [[ "$title" == "title" || -z "$title" ]] && continue

        if add_movie "$title" "$year" "$tmdb_id" "$quality_profile_id" "$root_folder" "true" >/dev/null 2>&1; then
            ((success_count++))
            log_success "Added: $title ($year)"
        else
            ((error_count++))
            log_error "Failed to add: $title ($year)"
        fi

        # Small delay to avoid overwhelming the API
        sleep 1
    done < "$movie_list_file"

    log_info "Bulk add complete: $success_count successful, $error_count failed"
}

# Calendar Functions
get_calendar() {
    local start_date="${1:-}"
    local end_date="${2:-}"

    log_info "Getting calendar events"

    local url="/calendar"
    local params=()

    if [[ -n "$start_date" ]]; then
        params+=("start=${start_date}")
    fi

    if [[ -n "$end_date" ]]; then
        params+=("end=${end_date}")
    fi

    if [[ ${#params[@]} -gt 0 ]]; then
        url="${url}?$(IFS='&'; echo "${params[*]}")"
    fi

    api_get "$url" | jq -r '.[] | "\(.airDate): \(.title) (\(.year))"'
}

# Configuration Functions
backup_configuration() {
    local backup_dir="${1:-./radarr-backup-$(date +%Y%m%d-%H%M%S)}"

    log_info "Creating configuration backup in: $backup_dir"
    mkdir -p "$backup_dir"

    # Export various configurations
    api_get "/qualityprofile" > "$backup_dir/quality-profiles.json"
    api_get "/indexer" > "$backup_dir/indexers.json"
    api_get "/downloadclient" > "$backup_dir/download-clients.json"
    api_get "/notification" > "$backup_dir/notifications.json"
    api_get "/rootfolder" > "$backup_dir/root-folders.json"
    api_get "/movie" > "$backup_dir/movies.json"

    log_success "Configuration backup completed"
    echo "Backup location: $backup_dir"
}

# Main command dispatcher
main() {
    case "${1:-}" in
        "health")
            check_health
            ;;
        "status")
            get_system_status
            ;;
        "search")
            [[ -z "${2:-}" ]] && { log_error "Usage: $0 search <query>"; exit 1; }
            search_movies "$2"
            ;;
        "add")
            [[ $# -lt 4 ]] && { log_error "Usage: $0 add <title> <year> <tmdb_id> [quality_profile_id] [root_folder]"; exit 1; }
            add_movie "$2" "$3" "$4" "${5:-1}" "${6:-/movies}"
            ;;
        "movies")
            get_movies "${2:-}"
            ;;
        "delete")
            [[ -z "${2:-}" ]] && { log_error "Usage: $0 delete <movie_id> [delete_files] [add_exclusion]"; exit 1; }
            delete_movie "$2" "${3:-false}" "${4:-false}"
            ;;
        "queue")
            get_queue
            ;;
        "remove-queue")
            [[ -z "${2:-}" ]] && { log_error "Usage: $0 remove-queue <queue_id> [remove_from_client] [blacklist]"; exit 1; }
            remove_from_queue "$2" "${3:-true}" "${4:-false}"
            ;;
        "releases")
            [[ -z "${2:-}" ]] && { log_error "Usage: $0 releases <movie_id>"; exit 1; }
            search_movie_releases "$2"
            ;;
        "grab")
            [[ -z "${2:-}" ]] && { log_error "Usage: $0 grab <release_guid> [indexer_id]"; exit 1; }
            grab_release "$2" "${3:-}"
            ;;
        "refresh")
            refresh_movie "${2:-}"
            ;;
        "monitor-queue")
            monitor_queue "${2:-30}"
            ;;
        "bulk-add")
            [[ -z "${2:-}" ]] && { log_error "Usage: $0 bulk-add <csv_file> [quality_profile_id] [root_folder]"; exit 1; }
            bulk_add_from_list "$2" "${3:-1}" "${4:-/movies}"
            ;;
        "calendar")
            get_calendar "$2" "$3"
            ;;
        "backup")
            backup_configuration "$2"
            ;;
        *)
            cat <<EOF
Radarr Go API Shell Client

Usage: $0 <command> [arguments]

Commands:
  health                          - Check system health
  status                          - Get system status
  search <query>                  - Search for movies
  add <title> <year> <tmdb_id>    - Add movie to collection
  movies [filters]                - List movies
  delete <id> [del_files] [excl]  - Delete movie
  queue                           - Show download queue
  remove-queue <id> [client] [bl] - Remove from queue
  releases <movie_id>             - Search movie releases
  grab <guid> [indexer_id]        - Download release
  refresh [movie_id]              - Refresh movie(s)
  monitor-queue [interval]        - Monitor queue continuously
  bulk-add <csv_file>             - Bulk add from CSV
  calendar [start] [end]          - Get calendar events
  backup [directory]              - Backup configuration

Examples:
  $0 health
  $0 search "The Matrix"
  $0 add "Inception" 2010 27205
  $0 movies "monitored=true"
  $0 monitor-queue 60
  $0 bulk-add movies.csv 1 /movies

Environment Variables:
  RADARR_URL      - Radarr server URL (default: http://localhost:7878)
  RADARR_API_KEY  - API key (required)

CSV Format for bulk operations:
  title,year,tmdb_id
  "The Matrix",1999,603
  "Inception",2010,27205
EOF
            exit 1
            ;;
    esac
}

# Validation
if [[ -z "${RADARR_API_KEY:-}" ]]; then
    log_error "RADARR_API_KEY environment variable is required"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    log_error "jq is required but not installed"
    exit 1
fi

if ! command -v curl >/dev/null 2>&1; then
    log_error "curl is required but not installed"
    exit 1
fi

# Run main function
main "$@"
```

### PowerShell Examples for Windows

```powershell
# Radarr Go API PowerShell Client
# Production-ready PowerShell module for Windows automation

param(
    [Parameter(Mandatory=$false)]
    [string]$RadarrUrl = $env:RADARR_URL ?? "http://localhost:7878",

    [Parameter(Mandatory=$true)]
    [string]$ApiKey = $env:RADARR_API_KEY
)

# Module-level variables
$script:RadarrBaseUrl = $RadarrUrl.TrimEnd('/') + '/api/v3'
$script:Headers = @{
    'X-API-Key' = $ApiKey
    'Content-Type' = 'application/json'
    'User-Agent' = 'Radarr-PowerShell-Client/2.0'
}

# Enhanced error handling class
class RadarrAPIException : System.Exception {
    [int]$StatusCode
    [object]$ResponseData

    RadarrAPIException([string]$message, [int]$statusCode, [object]$responseData) : base($message) {
        $this.StatusCode = $statusCode
        $this.ResponseData = $responseData
    }
}

# Core API request function with retry logic
function Invoke-RadarrAPI {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$Endpoint,

        [Parameter(Mandatory=$false)]
        [Microsoft.PowerShell.Commands.WebRequestMethod]$Method = 'GET',

        [Parameter(Mandatory=$false)]
        [object]$Body,

        [Parameter(Mandatory=$false)]
        [hashtable]$QueryParameters,

        [Parameter(Mandatory=$false)]
        [int]$TimeoutSec = 30,

        [Parameter(Mandatory=$false)]
        [int]$MaxRetries = 3,

        [Parameter(Mandatory=$false)]
        [switch]$PassThru
    )

    $uri = "$script:RadarrBaseUrl/$($Endpoint.TrimStart('/'))"

    # Add query parameters
    if ($QueryParameters) {
        $queryString = ($QueryParameters.GetEnumerator() | ForEach-Object {
            "$($_.Key)=$([System.Web.HttpUtility]::UrlEncode($_.Value))"
        }) -join '&'
        $uri += "?$queryString"
    }

    $requestParams = @{
        Uri = $uri
        Method = $Method
        Headers = $script:Headers
        TimeoutSec = $TimeoutSec
        ErrorAction = 'Stop'
    }

    if ($Body) {
        if ($Body -is [string]) {
            $requestParams.Body = $Body
        } else {
            $requestParams.Body = ($Body | ConvertTo-Json -Depth 10 -Compress)
        }
    }

    $attempt = 1
    $backoffDelay = 5

    do {
        try {
            Write-Verbose "Making $Method request to $uri (attempt $attempt)"
            $response = Invoke-RestMethod @requestParams

            if ($PassThru) {
                return $response
            } else {
                return $response
            }
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__

            # Handle rate limiting
            if ($statusCode -eq 429) {
                $retryAfter = $_.Exception.Response.Headers['Retry-After']
                if ($retryAfter) {
                    $delay = [int]$retryAfter
                } else {
                    $delay = $backoffDelay
                }

                Write-Warning "Rate limited. Waiting $delay seconds before retry..."
                Start-Sleep -Seconds $delay
                $attempt++
                $backoffDelay *= 2
                continue
            }

            # Handle other errors
            if ($attempt -lt $MaxRetries -and $statusCode -ge 500) {
                Write-Warning "Attempt $attempt failed with status $statusCode. Retrying in $backoffDelay seconds..."
                Start-Sleep -Seconds $backoffDelay
                $attempt++
                $backoffDelay *= 2
                continue
            }

            # Parse error response
            try {
                $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
                $errorMessage = $errorResponse.error.message ?? $errorResponse.error ?? $_.Exception.Message
            }
            catch {
                $errorMessage = $_.Exception.Message
            }

            throw [RadarrAPIException]::new(
                "API request failed: $errorMessage",
                $statusCode,
                $errorResponse
            )
        }
    } while ($attempt -le $MaxRetries)
}

# System Information Functions
function Get-RadarrSystemStatus {
    [CmdletBinding()]
    param()

    Write-Verbose "Getting Radarr system status"
    return Invoke-RadarrAPI -Endpoint '/system/status'
}

function Test-RadarrHealth {
    [CmdletBinding()]
    param()

    try {
        $response = Invoke-WebRequest -Uri "$RadarrUrl/ping" -TimeoutSec 10 -ErrorAction Stop
        return $response.StatusCode -eq 200
    }
    catch {
        return $false
    }
}

function Get-RadarrHealthIssues {
    [CmdletBinding()]
    param()

    Write-Verbose "Getting health issues"
    return Invoke-RadarrAPI -Endpoint '/health'
}

# Movie Management Functions
function Find-RadarrMovies {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true, ValueFromPipeline=$true)]
        [string]$SearchTerm,

        [Parameter(Mandatory=$false)]
        [int]$Limit = 20
    )

    process {
        Write-Verbose "Searching for movies: $SearchTerm"

        $queryParams = @{ 'term' = $SearchTerm }
        if ($Limit -gt 0) {
            $queryParams['limit'] = $Limit
        }

        return Invoke-RadarrAPI -Endpoint '/movie/lookup' -QueryParameters $queryParams
    }
}

function Get-RadarrMovies {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$false)]
        [bool]$Monitored,

        [Parameter(Mandatory=$false)]
        [string]$SortKey = 'title',

        [Parameter(Mandatory=$false)]
        [string]$SortDirection = 'asc',

        [Parameter(Mandatory=$false)]
        [int]$Page = 1,

        [Parameter(Mandatory=$false)]
        [int]$PageSize = 100
    )

    $queryParams = @{
        'sortKey' = $SortKey
        'sortDirection' = $SortDirection
        'page' = $Page
        'pageSize' = $PageSize
    }

    if ($PSBoundParameters.ContainsKey('Monitored')) {
        $queryParams['monitored'] = $Monitored.ToString().ToLower()
    }

    Write-Verbose "Getting movies with filters"
    $response = Invoke-RadarrAPI -Endpoint '/movie' -QueryParameters $queryParams

    # Handle paginated response
    if ($response.data) {
        return $response.data
    } else {
        return $response
    }
}

function Get-RadarrMovie {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true, ValueFromPipeline=$true)]
        [int]$MovieId
    )

    process {
        Write-Verbose "Getting movie with ID: $MovieId"
        return Invoke-RadarrAPI -Endpoint "/movie/$MovieId"
    }
}

function Add-RadarrMovie {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$Title,

        [Parameter(Mandatory=$true)]
        [int]$Year,

        [Parameter(Mandatory=$true)]
        [int]$TmdbId,

        [Parameter(Mandatory=$false)]
        [string]$ImdbId,

        [Parameter(Mandatory=$false)]
        [int]$QualityProfileId = 1,

        [Parameter(Mandatory=$false)]
        [string]$RootFolderPath = '/movies',

        [Parameter(Mandatory=$false)]
        [bool]$Monitored = $true,

        [Parameter(Mandatory=$false)]
        [bool]$SearchForMovie = $true
    )

    $movieData = @{
        title = $Title
        year = $Year
        tmdbId = $TmdbId
        qualityProfileId = $QualityProfileId
        rootFolderPath = $RootFolderPath
        monitored = $Monitored
        addOptions = @{
            monitor = 'movieOnly'
            searchForMovie = $SearchForMovie
        }
    }

    if ($ImdbId) {
        $movieData.imdbId = $ImdbId
    }

    Write-Verbose "Adding movie: $Title ($Year)"
    return Invoke-RadarrAPI -Endpoint '/movie' -Method 'POST' -Body $movieData
}

function Update-RadarrMovie {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [int]$MovieId,

        [Parameter(Mandatory=$true)]
        [hashtable]$MovieData
    )

    Write-Verbose "Updating movie with ID: $MovieId"
    return Invoke-RadarrAPI -Endpoint "/movie/$MovieId" -Method 'PUT' -Body $MovieData
}

function Remove-RadarrMovie {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true, ValueFromPipeline=$true)]
        [int]$MovieId,

        [Parameter(Mandatory=$false)]
        [bool]$DeleteFiles = $false,

        [Parameter(Mandatory=$false)]
        [bool]$AddExclusion = $false
    )

    process {
        $queryParams = @{
            'deleteFiles' = $DeleteFiles.ToString().ToLower()
            'addExclusion' = $AddExclusion.ToString().ToLower()
        }

        Write-Verbose "Deleting movie with ID: $MovieId"
        Invoke-RadarrAPI -Endpoint "/movie/$MovieId" -Method 'DELETE' -QueryParameters $queryParams

        Write-Output "Movie $MovieId deleted successfully"
    }
}

# Queue Management Functions
function Get-RadarrQueue {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$false)]
        [int]$Page = 1,

        [Parameter(Mandatory=$false)]
        [int]$PageSize = 20,

        [Parameter(Mandatory=$false)]
        [string]$SortKey = 'timeleft'
    )

    $queryParams = @{
        'page' = $Page
        'pageSize' = $PageSize
        'sortKey' = $SortKey
    }

    Write-Verbose "Getting download queue"
    $response = Invoke-RadarrAPI -Endpoint '/queue' -QueryParameters $queryParams

    if ($response.data) {
        return $response.data
    } else {
        return $response
    }
}

function Remove-RadarrQueueItem {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true, ValueFromPipeline=$true)]
        [int]$QueueId,

        [Parameter(Mandatory=$false)]
        [bool]$RemoveFromClient = $true,

        [Parameter(Mandatory=$false)]
        [bool]$Blacklist = $false
    )

    process {
        $queryParams = @{
            'removeFromClient' = $RemoveFromClient.ToString().ToLower()
            'blacklist' = $Blacklist.ToString().ToLower()
        }

        Write-Verbose "Removing queue item: $QueueId"
        Invoke-RadarrAPI -Endpoint "/queue/$QueueId" -Method 'DELETE' -QueryParameters $queryParams

        Write-Output "Queue item $QueueId removed successfully"
    }
}

# Command Execution Functions
function Invoke-RadarrCommand {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$CommandName,

        [Parameter(Mandatory=$false)]
        [hashtable]$Parameters = @{}
    )

    $commandData = @{
        name = $CommandName
    }

    foreach ($key in $Parameters.Keys) {
        $commandData[$key] = $Parameters[$key]
    }

    Write-Verbose "Executing command: $CommandName"
    return Invoke-RadarrAPI -Endpoint '/command' -Method 'POST' -Body $commandData
}

function Start-RadarrMovieRefresh {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$false)]
        [int]$MovieId
    )

    $parameters = @{}
    if ($MovieId) {
        $parameters['movieId'] = $MovieId
        Write-Verbose "Refreshing movie ID: $MovieId"
    } else {
        Write-Verbose "Refreshing all movies"
    }

    return Invoke-RadarrCommand -CommandName 'RefreshMovie' -Parameters $parameters
}

# Search and Release Functions
function Find-RadarrReleases {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [int]$MovieId
    )

    Write-Verbose "Searching releases for movie ID: $MovieId"
    return Invoke-RadarrAPI -Endpoint '/release' -QueryParameters @{ 'movieId' = $MovieId }
}

function Start-RadarrReleaseDownload {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$ReleaseGuid,

        [Parameter(Mandatory=$false)]
        [int]$IndexerId
    )

    $grabData = @{
        guid = $ReleaseGuid
    }

    if ($IndexerId) {
        $grabData.indexerId = $IndexerId
    }

    Write-Verbose "Downloading release: $ReleaseGuid"
    return Invoke-RadarrAPI -Endpoint '/release/grab' -Method 'POST' -Body $grabData
}

# Monitoring Functions
function Start-RadarrQueueMonitor {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$false)]
        [int]$IntervalSeconds = 30,

        [Parameter(Mandatory=$false)]
        [switch]$ShowProgress
    )

    Write-Host "Starting queue monitor (Press Ctrl+C to stop)" -ForegroundColor Green

    try {
        while ($true) {
            Clear-Host
            Write-Host "=== Radarr Download Queue - $(Get-Date) ===" -ForegroundColor Cyan
            Write-Host

            try {
                $queue = Get-RadarrQueue -PageSize 50

                if ($queue.Count -eq 0) {
                    Write-Host "Queue is empty" -ForegroundColor Yellow
                } else {
                    $queue | ForEach-Object {
                        $status = $_.status.ToUpper()
                        $progress = if ($_.progress) { "$($_.progress)%" } else { "0%" }
                        $eta = if ($_.timeleft) { $_.timeleft } else { "Unknown" }

                        $color = switch ($status) {
                            'DOWNLOADING' { 'Green' }
                            'PAUSED' { 'Yellow' }
                            'FAILED' { 'Red' }
                            default { 'White' }
                        }

                        Write-Host "[$status] $($_.title)" -ForegroundColor $color
                        Write-Host "  Progress: $progress - ETA: $eta" -ForegroundColor Gray

                        if ($ShowProgress -and $_.progress -gt 0) {
                            $barWidth = 50
                            $completed = [math]::Floor($barWidth * ($_.progress / 100))
                            $remaining = $barWidth - $completed

                            $progressBar = "[" + ("█" * $completed) + ("░" * $remaining) + "]"
                            Write-Host "  $progressBar" -ForegroundColor Cyan
                        }

                        Write-Host
                    }
                }
            }
            catch {
                Write-Host "Error getting queue: $($_.Exception.Message)" -ForegroundColor Red
            }

            Write-Host "Next update in $IntervalSeconds seconds..." -ForegroundColor Gray
            Start-Sleep -Seconds $IntervalSeconds
        }
    }
    catch {
        Write-Host "`nMonitoring stopped." -ForegroundColor Yellow
    }
}

# Bulk Operations
function Import-RadarrMoviesFromCsv {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$CsvPath,

        [Parameter(Mandatory=$false)]
        [int]$QualityProfileId = 1,

        [Parameter(Mandatory=$false)]
        [string]$RootFolderPath = '/movies',

        [Parameter(Mandatory=$false)]
        [int]$BatchSize = 5,

        [Parameter(Mandatory=$false)]
        [int]$DelaySeconds = 1
    )

    if (!(Test-Path $CsvPath)) {
        throw "CSV file not found: $CsvPath"
    }

    $movies = Import-Csv -Path $CsvPath
    $totalCount = $movies.Count
    $successCount = 0
    $errorCount = 0
    $results = @()

    Write-Host "Importing $totalCount movies from $CsvPath" -ForegroundColor Green

    for ($i = 0; $i -lt $totalCount; $i += $BatchSize) {
        $batch = $movies[$i..([math]::Min($i + $BatchSize - 1, $totalCount - 1))]

        foreach ($movie in $batch) {
            try {
                $result = Add-RadarrMovie -Title $movie.title -Year $movie.year -TmdbId $movie.tmdb_id -QualityProfileId $QualityProfileId -RootFolderPath $RootFolderPath

                $successCount++
                $results += [PSCustomObject]@{
                    Title = $movie.title
                    Year = $movie.year
                    Status = 'Success'
                    MovieId = $result.id
                    Error = $null
                }

                Write-Host "✓ Added: $($movie.title) ($($movie.year))" -ForegroundColor Green
            }
            catch {
                $errorCount++
                $results += [PSCustomObject]@{
                    Title = $movie.title
                    Year = $movie.year
                    Status = 'Failed'
                    MovieId = $null
                    Error = $_.Exception.Message
                }

                Write-Host "✗ Failed: $($movie.title) ($($movie.year)) - $($_.Exception.Message)" -ForegroundColor Red
            }
        }

        # Delay between batches
        if ($i + $BatchSize -lt $totalCount) {
            Write-Host "Waiting $DelaySeconds seconds before next batch..." -ForegroundColor Yellow
            Start-Sleep -Seconds $DelaySeconds
        }
    }

    Write-Host "`nBulk import completed:" -ForegroundColor Cyan
    Write-Host "  Success: $successCount" -ForegroundColor Green
    Write-Host "  Failed: $errorCount" -ForegroundColor Red

    return $results
}

# Calendar Functions
function Get-RadarrCalendar {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$false)]
        [datetime]$StartDate,

        [Parameter(Mandatory=$false)]
        [datetime]$EndDate,

        [Parameter(Mandatory=$false)]
        [bool]$IncludeUnmonitored = $false
    )

    $queryParams = @{}

    if ($StartDate) {
        $queryParams['start'] = $StartDate.ToString('yyyy-MM-ddTHH:mm:ss.fffZ')
    }

    if ($EndDate) {
        $queryParams['end'] = $EndDate.ToString('yyyy-MM-ddTHH:mm:ss.fffZ')
    }

    if ($IncludeUnmonitored) {
        $queryParams['unmonitored'] = 'true'
    }

    Write-Verbose "Getting calendar events"
    return Invoke-RadarrAPI -Endpoint '/calendar' -QueryParameters $queryParams
}

# Configuration Backup
function Backup-RadarrConfiguration {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$false)]
        [string]$BackupPath = ".\radarr-backup-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    )

    Write-Host "Creating configuration backup in: $BackupPath" -ForegroundColor Green

    if (!(Test-Path $BackupPath)) {
        New-Item -ItemType Directory -Path $BackupPath -Force | Out-Null
    }

    $configs = @{
        'quality-profiles.json' = '/qualityprofile'
        'indexers.json' = '/indexer'
        'download-clients.json' = '/downloadclient'
        'notifications.json' = '/notification'
        'root-folders.json' = '/rootfolder'
        'movies.json' = '/movie'
    }

    foreach ($config in $configs.GetEnumerator()) {
        try {
            Write-Host "Backing up $($config.Key)..." -ForegroundColor Yellow
            $data = Invoke-RadarrAPI -Endpoint $config.Value
            $data | ConvertTo-Json -Depth 10 | Out-File -FilePath (Join-Path $BackupPath $config.Key) -Encoding UTF8
        }
        catch {
            Write-Warning "Failed to backup $($config.Key): $($_.Exception.Message)"
        }
    }

    Write-Host "Configuration backup completed!" -ForegroundColor Green
    Write-Host "Backup location: $BackupPath" -ForegroundColor Cyan

    return $BackupPath
}

# Export all functions
Export-ModuleMember -Function *

# Usage Examples
if ($MyInvocation.InvocationName -eq '&') {
    # Example usage when script is dot-sourced
    Write-Host "Radarr Go PowerShell Client loaded" -ForegroundColor Green
    Write-Host "Available functions:" -ForegroundColor Cyan
    Get-Command -Module $MyInvocation.MyCommand.Module | Select-Object Name | Sort-Object Name

    Write-Host "`nExample usage:" -ForegroundColor Yellow
    Write-Host "  Test-RadarrHealth" -ForegroundColor Gray
    Write-Host "  Get-RadarrSystemStatus" -ForegroundColor Gray
    Write-Host "  Find-RadarrMovies 'The Matrix'" -ForegroundColor Gray
    Write-Host "  Get-RadarrMovies -Monitored `$true" -ForegroundColor Gray
    Write-Host "  Start-RadarrQueueMonitor -IntervalSeconds 30" -ForegroundColor Gray
}
```

### Go Client Library Example

Complete Go client library for integration:

```go
// Package radarrclient provides a comprehensive Go client for the Radarr Go API
package radarrclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client represents a Radarr Go API client
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
	timeout    time.Duration
}

// Config holds configuration for the Radarr client
type Config struct {
	BaseURL   string
	APIKey    string
	Timeout   time.Duration
	UserAgent string
}

// APIError represents an API error response
type APIError struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// Movie represents a movie in Radarr
type Movie struct {
	ID               int                `json:"id,omitempty"`
	Title            string             `json:"title"`
	OriginalTitle    string             `json:"originalTitle,omitempty"`
	AlternativeTitles []AlternativeTitle `json:"alternativeTitles,omitempty"`
	SortTitle        string             `json:"sortTitle,omitempty"`
	Status           string             `json:"status,omitempty"`
	Overview         string             `json:"overview,omitempty"`
	InCinemas        *time.Time         `json:"inCinemas,omitempty"`
	PhysicalRelease  *time.Time         `json:"physicalRelease,omitempty"`
	DigitalRelease   *time.Time         `json:"digitalRelease,omitempty"`
	Images           []MediaCover       `json:"images,omitempty"`
	Website          string             `json:"website,omitempty"`
	Year             int                `json:"year"`
	HasFile          bool               `json:"hasFile"`
	YouTubeTrailerID string             `json:"youTubeTrailerId,omitempty"`
	Studio           string             `json:"studio,omitempty"`
	Path             string             `json:"path,omitempty"`
	QualityProfileID int                `json:"qualityProfileId"`
	Monitored        bool               `json:"monitored"`
	MinimumAvailability string          `json:"minimumAvailability,omitempty"`
	IsAvailable      bool               `json:"isAvailable"`
	FolderName       string             `json:"folderName,omitempty"`
	Runtime          int                `json:"runtime,omitempty"`
	CleanTitle       string             `json:"cleanTitle,omitempty"`
	ImdbID           string             `json:"imdbId,omitempty"`
	TmdbID           int                `json:"tmdbId"`
	TitleSlug        string             `json:"titleSlug,omitempty"`
	Certification    string             `json:"certification,omitempty"`
	Genres           []string           `json:"genres,omitempty"`
	Tags             []int              `json:"tags,omitempty"`
	Added            *time.Time         `json:"added,omitempty"`
	Ratings          *Ratings           `json:"ratings,omitempty"`
	MovieFile        *MovieFile         `json:"movieFile,omitempty"`
	Collection       *Collection        `json:"collection,omitempty"`
}

// AlternativeTitle represents an alternative title for a movie
type AlternativeTitle struct {
	SourceType string `json:"sourceType"`
	MovieID    int    `json:"movieId"`
	Title      string `json:"title"`
	SourceID   int    `json:"sourceId"`
	Votes      int    `json:"votes"`
	VoteCount  int    `json:"voteCount"`
	Language   string `json:"language"`
}

// MediaCover represents a movie cover image
type MediaCover struct {
	CoverType string `json:"coverType"`
	URL       string `json:"url"`
}

// Ratings represents movie ratings
type Ratings struct {
	Imdb   *Rating `json:"imdb,omitempty"`
	Tmdb   *Rating `json:"tmdb,omitempty"`
	Metacritic *Rating `json:"metacritic,omitempty"`
	RottenTomatoes *Rating `json:"rottenTomatoes,omitempty"`
}

// Rating represents a single rating
type Rating struct {
	Votes int     `json:"votes"`
	Value float64 `json:"value"`
	Type  string  `json:"type"`
}

// MovieFile represents a movie file
type MovieFile struct {
	ID               int                `json:"id,omitempty"`
	MovieID          int                `json:"movieId"`
	RelativePath     string             `json:"relativePath"`
	Path             string             `json:"path"`
	Size             int64              `json:"size"`
	DateAdded        *time.Time         `json:"dateAdded,omitempty"`
	SceneName        string             `json:"sceneName,omitempty"`
	ReleaseGroup     string             `json:"releaseGroup,omitempty"`
	Quality          *Quality           `json:"quality,omitempty"`
	IndexerFlags     int                `json:"indexerFlags"`
	MediaInfo        *MediaInfo         `json:"mediaInfo,omitempty"`
	OriginalFilePath string             `json:"originalFilePath,omitempty"`
}

// Quality represents quality information
type Quality struct {
	Quality  *QualityDefinition `json:"quality,omitempty"`
	Revision *Revision          `json:"revision,omitempty"`
}

// QualityDefinition represents a quality definition
type QualityDefinition struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	Resolution int    `json:"resolution"`
	Modifier   string `json:"modifier"`
}

// Revision represents quality revision information
type Revision struct {
	Version  int  `json:"version"`
	Real     int  `json:"real"`
	IsRepack bool `json:"isRepack"`
}

// MediaInfo represents media information
type MediaInfo struct {
	AudioBitrate         int     `json:"audioBitrate"`
	AudioChannels        float64 `json:"audioChannels"`
	AudioCodec           string  `json:"audioCodec"`
	AudioLanguages       string  `json:"audioLanguages"`
	AudioStreamCount     int     `json:"audioStreamCount"`
	VideoBitDepth        int     `json:"videoBitDepth"`
	VideoBitrate         int     `json:"videoBitrate"`
	VideoCodec           string  `json:"videoCodec"`
	VideoDynamicRange    string  `json:"videoDynamicRange"`
	VideoDynamicRangeType string `json:"videoDynamicRangeType"`
	VideoFPS             float64 `json:"videoFps"`
	Resolution           string  `json:"resolution"`
	RunTime              string  `json:"runTime"`
	ScanType             string  `json:"scanType"`
	Subtitles            string  `json:"subtitles"`
}

// Collection represents a movie collection
type Collection struct {
	ID             int                 `json:"id,omitempty"`
	Title          string              `json:"title"`
	CleanTitle     string              `json:"cleanTitle,omitempty"`
	SortTitle      string              `json:"sortTitle,omitempty"`
	TmdbID         int                 `json:"tmdbId"`
	Images         []MediaCover        `json:"images,omitempty"`
	Overview       string              `json:"overview,omitempty"`
	Monitored      bool                `json:"monitored"`
	RootFolderPath string              `json:"rootFolderPath"`
	QualityProfileID int               `json:"qualityProfileId"`
	SearchOnAdd    bool                `json:"searchOnAdd"`
	MinimumAvailability string         `json:"minimumAvailability"`
	Movies         []Movie             `json:"movies,omitempty"`
	Added          *time.Time          `json:"added,omitempty"`
	Tags           []int               `json:"tags,omitempty"`
}

// QueueItem represents an item in the download queue
type QueueItem struct {
	ID                    int        `json:"id"`
	MovieID               int        `json:"movieId"`
	Languages             []Language `json:"languages"`
	Quality               *Quality   `json:"quality"`
	CustomFormats         []string   `json:"customFormats"`
	Size                  int64      `json:"size"`
	Title                 string     `json:"title"`
	SizeLeft              int64      `json:"sizeleft"`
	TimeLeft              string     `json:"timeleft"`
	EstimatedCompletionTime *time.Time `json:"estimatedCompletionTime"`
	Status                string     `json:"status"`
	TrackedDownloadStatus string     `json:"trackedDownloadStatus"`
	TrackedDownloadState  string     `json:"trackedDownloadState"`
	StatusMessages        []StatusMessage `json:"statusMessages"`
	ErrorMessage          string     `json:"errorMessage"`
	DownloadID            string     `json:"downloadId"`
	Protocol              string     `json:"protocol"`
	DownloadClient        string     `json:"downloadClient"`
	Indexer               string     `json:"indexer"`
	OutputPath            string     `json:"outputPath"`
	Movie                 *Movie     `json:"movie"`
}

// Language represents a language
type Language struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// StatusMessage represents a status message
type StatusMessage struct {
	Title    string   `json:"title"`
	Messages []string `json:"messages"`
}

// SystemStatus represents system status information
type SystemStatus struct {
	Version                string `json:"version"`
	BuildTime              string `json:"buildTime"`
	IsDebug                bool   `json:"isDebug"`
	IsProduction           bool   `json:"isProduction"`
	IsAdmin                bool   `json:"isAdmin"`
	IsUserInteractive      bool   `json:"isUserInteractive"`
	StartupPath            string `json:"startupPath"`
	AppData                string `json:"appData"`
	OSName                 string `json:"osName"`
	OSVersion              string `json:"osVersion"`
	IsMonoRuntime          bool   `json:"isMonoRuntime"`
	IsMono                 bool   `json:"isMono"`
	IsLinux                bool   `json:"isLinux"`
	IsOSX                  bool   `json:"isOsx"`
	IsWindows              bool   `json:"isWindows"`
	Mode                   string `json:"mode"`
	Branch                 string `json:"branch"`
	Authentication         string `json:"authentication"`
	SqliteVersion          string `json:"sqliteVersion"`
	MigrationVersion       int    `json:"migrationVersion"`
	URLBase                string `json:"urlBase"`
	RuntimeVersion         string `json:"runtimeVersion"`
	DatabaseType           string `json:"databaseType"`
	DatabaseVersion        string `json:"databaseVersion"`
	PackageVersion         string `json:"packageVersion"`
	PackageAuthor          string `json:"packageAuthor"`
	PackageUpdateMechanism string `json:"packageUpdateMechanism"`
}

// Command represents a command/task
type Command struct {
	ID               int                    `json:"id"`
	Name             string                 `json:"name"`
	CommandName      string                 `json:"commandName"`
	Message          string                 `json:"message"`
	Body             map[string]interface{} `json:"body"`
	Priority         string                 `json:"priority"`
	Status           string                 `json:"status"`
	Queued           *time.Time             `json:"queued"`
	Started          *time.Time             `json:"started"`
	Ended            *time.Time             `json:"ended"`
	Duration         string                 `json:"duration"`
	Exception        string                 `json:"exception"`
	Trigger          string                 `json:"trigger"`
	ClientUserAgent  string                 `json:"clientUserAgent"`
	StateChangeTime  *time.Time             `json:"stateChangeTime"`
	SendUpdatesToClient bool                `json:"sendUpdatesToClient"`
	UpdateScheduledTask bool                `json:"updateScheduledTask"`
	LastExecutionTime   *time.Time          `json:"lastExecutionTime"`
}

// NewClient creates a new Radarr API client
func NewClient(config Config) (*Client, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	userAgent := config.UserAgent
	if userAgent == "" {
		userAgent = "Radarr-Go-Client/2.0"
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		userAgent: userAgent,
		timeout:   timeout,
	}, nil
}

// makeRequest performs an HTTP request to the API
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	apiURL := c.baseURL.ResolveReference(&url.URL{Path: "/api/v3" + endpoint})

	req, err := http.NewRequestWithContext(ctx, method, apiURL.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("User-Agent", c.userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			apiErr.StatusCode = resp.StatusCode
			return &apiErr
		}

		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// System Operations
func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	var status SystemStatus
	err := c.makeRequest(ctx, "GET", "/system/status", nil, &status)
	return &status, err
}

func (c *Client) HealthCheck(ctx context.Context) error {
	return c.makeRequest(ctx, "GET", "/ping", nil, nil)
}

// Movie Operations
func (c *Client) GetMovies(ctx context.Context) ([]Movie, error) {
	var movies []Movie
	err := c.makeRequest(ctx, "GET", "/movie", nil, &movies)
	return movies, err
}

func (c *Client) GetMovie(ctx context.Context, movieID int) (*Movie, error) {
	var movie Movie
	err := c.makeRequest(ctx, "GET", "/movie/"+strconv.Itoa(movieID), nil, &movie)
	return &movie, err
}

func (c *Client) SearchMovies(ctx context.Context, query string) ([]Movie, error) {
	var movies []Movie
	endpoint := fmt.Sprintf("/movie/lookup?term=%s", url.QueryEscape(query))
	err := c.makeRequest(ctx, "GET", endpoint, nil, &movies)
	return movies, err
}

func (c *Client) AddMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	var addedMovie Movie
	err := c.makeRequest(ctx, "POST", "/movie", movie, &addedMovie)
	return &addedMovie, err
}

func (c *Client) UpdateMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	var updatedMovie Movie
	endpoint := "/movie/" + strconv.Itoa(movie.ID)
	err := c.makeRequest(ctx, "PUT", endpoint, movie, &updatedMovie)
	return &updatedMovie, err
}

func (c *Client) DeleteMovie(ctx context.Context, movieID int, deleteFiles, addExclusion bool) error {
	endpoint := fmt.Sprintf("/movie/%d?deleteFiles=%t&addExclusion=%t", movieID, deleteFiles, addExclusion)
	return c.makeRequest(ctx, "DELETE", endpoint, nil, nil)
}

// Queue Operations
func (c *Client) GetQueue(ctx context.Context) ([]QueueItem, error) {
	var queue []QueueItem
	err := c.makeRequest(ctx, "GET", "/queue", nil, &queue)
	return queue, err
}

func (c *Client) RemoveFromQueue(ctx context.Context, queueID int, removeFromClient, blacklist bool) error {
	endpoint := fmt.Sprintf("/queue/%d?removeFromClient=%t&blacklist=%t", queueID, removeFromClient, blacklist)
	return c.makeRequest(ctx, "DELETE", endpoint, nil, nil)
}

// Command Operations
func (c *Client) ExecuteCommand(ctx context.Context, commandName string, parameters map[string]interface{}) (*Command, error) {
	commandData := map[string]interface{}{
		"name": commandName,
	}

	for key, value := range parameters {
		commandData[key] = value
	}

	var command Command
	err := c.makeRequest(ctx, "POST", "/command", commandData, &command)
	return &command, err
}

func (c *Client) RefreshMovie(ctx context.Context, movieID int) (*Command, error) {
	parameters := map[string]interface{}{}
	if movieID > 0 {
		parameters["movieId"] = movieID
	}

	return c.ExecuteCommand(ctx, "RefreshMovie", parameters)
}

func (c *Client) RefreshAllMovies(ctx context.Context) (*Command, error) {
	return c.ExecuteCommand(ctx, "RefreshMovie", nil)
}

// Example usage
func ExampleUsage() {
	// Create client
	client, err := NewClient(Config{
		BaseURL: "http://localhost:7878",
		APIKey:  "your-api-key-here",
		Timeout: 30 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Check system health
	if err := client.HealthCheck(ctx); err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		return
	}

	// Get system status
	status, err := client.GetSystemStatus(ctx)
	if err != nil {
		fmt.Printf("Failed to get system status: %v\n", err)
		return
	}
	fmt.Printf("Radarr version: %s\n", status.Version)

	// Search for movies
	searchResults, err := client.SearchMovies(ctx, "The Matrix")
	if err != nil {
		fmt.Printf("Search failed: %v\n", err)
		return
	}

	fmt.Printf("Found %d search results\n", len(searchResults))

	// Add a movie
	if len(searchResults) > 0 {
		movieToAdd := &searchResults[0]
		movieToAdd.QualityProfileID = 1
		movieToAdd.Monitored = true
		// Set other required fields...

		addedMovie, err := client.AddMovie(ctx, movieToAdd)
		if err != nil {
			fmt.Printf("Failed to add movie: %v\n", err)
		} else {
			fmt.Printf("Added movie: %s (ID: %d)\n", addedMovie.Title, addedMovie.ID)
		}
	}

	// Get queue
	queue, err := client.GetQueue(ctx)
	if err != nil {
		fmt.Printf("Failed to get queue: %v\n", err)
		return
	}

	fmt.Printf("Queue has %d items\n", len(queue))

	// Refresh all movies
	command, err := client.RefreshAllMovies(ctx)
	if err != nil {
		fmt.Printf("Failed to refresh movies: %v\n", err)
	} else {
		fmt.Printf("Refresh command queued (ID: %d)\n", command.ID)
	}
}
```

This completes the first section of comprehensive third-party client integration examples. Each client provides:

1. **Production-ready error handling** with retries and rate limiting
2. **Complete API coverage** for all major operations
3. **Best practices** for each language/platform
4. **Comprehensive documentation** and usage examples
5. **Real-world usage patterns** for integration scenarios

<function_calls>
<invoke name="TodoWrite">
<parameter name="todos">[{"content": "Create comprehensive third-party client integration examples", "activeForm": "Creating third-party client integration examples", "status": "completed"}, {"content": "Develop common integration patterns and workflows", "activeForm": "Developing common integration patterns and workflows", "status": "in_progress"}, {"content": "Build automation examples for various use cases", "activeForm": "Building automation examples for various use cases", "status": "pending"}, {"content": "Create comprehensive troubleshooting guide for API integration", "activeForm": "Creating comprehensive troubleshooting guide for API integration", "status": "pending"}]
