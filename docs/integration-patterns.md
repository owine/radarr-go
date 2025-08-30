# Common Integration Patterns for Radarr Go API

This document provides battle-tested patterns for common Radarr Go integration scenarios.

## 1. Movie Search and Addition Workflow

### Complete Movie Addition Flow

```python
class MovieManager:
    def __init__(self, client):
        self.client = client
        self.default_quality_profile_id = None
        self.default_root_folder = None

    def setup_defaults(self):
        """Setup default quality profile and root folder"""
        # Get quality profiles and select a default
        profiles = self.client.get_quality_profiles()
        hd_profile = next((p for p in profiles if 'HD' in p['name']), profiles[0])
        self.default_quality_profile_id = hd_profile['id']

        # Set default root folder
        self.default_root_folder = '/movies'

        print(f"Using quality profile: {hd_profile['name']}")

    def smart_movie_search(self, query, year=None):
        """Enhanced movie search with year filtering and duplicate detection"""
        search_results = self.client.search_movies(query)

        if year:
            # Filter by year if specified
            search_results = [m for m in search_results if m.get('year') == year]

        # Check for existing movies to prevent duplicates
        existing_movies = self.client.get_movies()
        existing_tmdb_ids = {m['tmdbId'] for m in existing_movies}

        # Filter out movies already in collection
        new_movies = [m for m in search_results if m['tmdbId'] not in existing_tmdb_ids]

        return {
            'results': new_movies,
            'already_exists': len(search_results) - len(new_movies),
            'total_found': len(search_results)
        }

    def add_movie_with_validation(self, tmdb_id, quality_profile_id=None,
                                 root_folder=None, **options):
        """Add movie with comprehensive validation"""
        # Use defaults if not specified
        quality_profile_id = quality_profile_id or self.default_quality_profile_id
        root_folder = root_folder or self.default_root_folder

        try:
            # Validate movie exists and get details
            movie_info = self.client._request('GET', f'/movie/lookup/tmdb?tmdbId={tmdb_id}').json()

            # Check if already exists
            existing = self.client.get_movies(tmdbId=tmdb_id)
            if existing:
                return {
                    'success': False,
                    'error': 'Movie already exists',
                    'existing_movie': existing[0]
                }

            # Validate quality profile exists
            profiles = self.client.get_quality_profiles()
            if not any(p['id'] == quality_profile_id for p in profiles):
                return {
                    'success': False,
                    'error': f'Invalid quality profile ID: {quality_profile_id}'
                }

            # Add the movie
            added_movie = self.client.add_movie(
                tmdb_id=tmdb_id,
                quality_profile_id=quality_profile_id,
                root_folder=root_folder,
                monitored=options.get('monitored', True),
                search_on_add=options.get('search_on_add', True)
            )

            return {
                'success': True,
                'movie': added_movie,
                'message': f"Successfully added {movie_info['title']} ({movie_info['year']})"
            }

        except Exception as e:
            return {
                'success': False,
                'error': str(e)
            }

    def batch_add_movies(self, movie_list, progress_callback=None):
        """Add multiple movies with progress tracking and error handling"""
        results = {
            'successful': [],
            'failed': [],
            'skipped': [],
            'total': len(movie_list)
        }

        for i, movie_data in enumerate(movie_list):
            if progress_callback:
                progress_callback(i + 1, len(movie_list), movie_data.get('title', 'Unknown'))

            # Extract movie data
            tmdb_id = movie_data.get('tmdb_id') or movie_data.get('tmdbId')
            if not tmdb_id:
                results['failed'].append({
                    'movie': movie_data,
                    'error': 'Missing TMDB ID'
                })
                continue

            # Add movie
            result = self.add_movie_with_validation(
                tmdb_id=tmdb_id,
                quality_profile_id=movie_data.get('quality_profile_id'),
                root_folder=movie_data.get('root_folder'),
                monitored=movie_data.get('monitored', True),
                search_on_add=movie_data.get('search_on_add', False)  # Don't auto-search in batch
            )

            if result['success']:
                results['successful'].append(result)
            elif 'already exists' in result['error']:
                results['skipped'].append(result)
            else:
                results['failed'].append({
                    'movie': movie_data,
                    'error': result['error']
                })

            # Rate limiting - wait between requests
            time.sleep(0.5)

        return results

# Usage example
def main():
    client = RadarrClient('http://localhost:7878', 'your-api-key')
    manager = MovieManager(client)
    manager.setup_defaults()

    # Single movie addition
    result = manager.smart_movie_search("The Matrix", year=1999)
    print(f"Found {result['total_found']} movies, {result['already_exists']} already exist")

    if result['results']:
        movie = result['results'][0]
        add_result = manager.add_movie_with_validation(movie['tmdbId'])
        print(add_result['message'])

    # Batch addition from CSV
    import csv
    movie_list = []
    with open('movies_to_add.csv', 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            movie_list.append({
                'tmdb_id': int(row['tmdb_id']),
                'title': row['title'],
                'monitored': row.get('monitored', 'true').lower() == 'true'
            })

    def progress_callback(current, total, title):
        print(f"Progress: {current}/{total} - Adding {title}")

    batch_results = manager.batch_add_movies(movie_list, progress_callback)
    print(f"Batch complete: {len(batch_results['successful'])} added, "
          f"{len(batch_results['failed'])} failed, {len(batch_results['skipped'])} skipped")
```

## 2. Queue Monitoring and Management

### Real-time Queue Status Monitoring

```javascript
class QueueMonitor {
    constructor(client) {
        this.client = client;
        this.isMonitoring = false;
        this.monitorInterval = null;
        this.callbacks = {
            queueUpdate: [],
            downloadComplete: [],
            downloadFailed: [],
            queueEmpty: []
        };
    }

    // Subscribe to queue events
    on(event, callback) {
        if (this.callbacks[event]) {
            this.callbacks[event].push(callback);
        }
    }

    emit(event, data) {
        if (this.callbacks[event]) {
            this.callbacks[event].forEach(callback => callback(data));
        }
    }

    async startMonitoring(intervalSeconds = 10) {
        if (this.isMonitoring) {
            return;
        }

        this.isMonitoring = true;
        let lastQueueState = new Map();

        this.monitorInterval = setInterval(async () => {
            try {
                const queue = await this.getQueueStatus();
                const currentQueueState = new Map();

                // Process each queue item
                for (const item of queue.items) {
                    currentQueueState.set(item.id, item.status);

                    const lastStatus = lastQueueState.get(item.id);

                    // Detect status changes
                    if (lastStatus && lastStatus !== item.status) {
                        if (item.status === 'completed') {
                            this.emit('downloadComplete', item);
                        } else if (item.status === 'failed') {
                            this.emit('downloadFailed', item);
                        }
                    }
                }

                // Detect completed/removed items
                for (const [id, status] of lastQueueState.entries()) {
                    if (!currentQueueState.has(id) && status !== 'completed') {
                        // Item was removed from queue (likely completed)
                        this.emit('downloadComplete', { id, status: 'completed' });
                    }
                }

                // Check for empty queue
                if (queue.items.length === 0 && lastQueueState.size > 0) {
                    this.emit('queueEmpty', {});
                }

                lastQueueState = currentQueueState;
                this.emit('queueUpdate', queue);

            } catch (error) {
                console.error('Queue monitoring error:', error);
            }
        }, intervalSeconds * 1000);
    }

    stopMonitoring() {
        this.isMonitoring = false;
        if (this.monitorInterval) {
            clearInterval(this.monitorInterval);
            this.monitorInterval = null;
        }
    }

    async getQueueStatus() {
        // Get current queue items (this would be implemented based on actual API)
        const response = await this.client.api.get('/queue');

        return {
            items: response.data.records || [],
            totalCount: response.data.totalRecords || 0,
            isPaused: response.data.isPaused || false
        };
    }

    async getQueueDetails() {
        const queue = await this.getQueueStatus();

        const stats = {
            total: queue.totalCount,
            downloading: 0,
            queued: 0,
            completed: 0,
            failed: 0,
            paused: 0
        };

        const downloadInfo = {
            totalSize: 0,
            downloadedSize: 0,
            downloadSpeed: 0,
            timeRemaining: 0
        };

        queue.items.forEach(item => {
            stats[item.status] = (stats[item.status] || 0) + 1;

            if (item.status === 'downloading') {
                downloadInfo.totalSize += item.size || 0;
                downloadInfo.downloadedSize += (item.size || 0) * (item.progress || 0) / 100;
                downloadInfo.downloadSpeed += item.downloadRate || 0;
            }
        });

        // Calculate estimated time remaining
        if (downloadInfo.downloadSpeed > 0) {
            const remainingBytes = downloadInfo.totalSize - downloadInfo.downloadedSize;
            downloadInfo.timeRemaining = remainingBytes / downloadInfo.downloadSpeed;
        }

        return {
            queue,
            stats,
            downloadInfo
        };
    }

    // Queue management methods
    async pauseQueue() {
        return await this.client.api.post('/command', {
            name: 'pauseQueue'
        });
    }

    async resumeQueue() {
        return await this.client.api.post('/command', {
            name: 'resumeQueue'
        });
    }

    async removeFromQueue(queueId, blacklist = false) {
        return await this.client.api.delete(`/queue/${queueId}`, {
            params: { blacklist }
        });
    }

    async retryDownload(queueId) {
        return await this.client.api.post(`/queue/${queueId}/retry`);
    }
}

// Usage example
async function setupQueueMonitoring() {
    const client = new RadarrClient('http://localhost:7878', 'your-api-key');
    const monitor = new QueueMonitor(client);

    // Set up event handlers
    monitor.on('downloadComplete', (item) => {
        console.log(`✓ Download completed: ${item.title}`);
        // Send notification, update database, etc.
    });

    monitor.on('downloadFailed', (item) => {
        console.log(`✗ Download failed: ${item.title} - ${item.errorMessage}`);
        // Log failure, send alert, retry logic, etc.
    });

    monitor.on('queueUpdate', async (queue) => {
        const details = await monitor.getQueueDetails();
        console.log(`Queue: ${details.stats.downloading} downloading, ${details.stats.queued} queued`);

        if (details.downloadInfo.downloadSpeed > 0) {
            const speedMBps = (details.downloadInfo.downloadSpeed / 1024 / 1024).toFixed(2);
            const timeRemaining = Math.round(details.downloadInfo.timeRemaining / 60);
            console.log(`Download speed: ${speedMBps} MB/s, ETA: ${timeRemaining} minutes`);
        }
    });

    monitor.on('queueEmpty', () => {
        console.log('Queue is now empty!');
        // Trigger post-processing, cleanup, notifications, etc.
    });

    // Start monitoring
    await monitor.startMonitoring(5); // Check every 5 seconds

    // Example queue management
    const details = await monitor.getQueueDetails();
    if (details.stats.failed > 0) {
        console.log('Found failed downloads, attempting to retry...');
        for (const item of details.queue.items) {
            if (item.status === 'failed') {
                await monitor.retryDownload(item.id);
            }
        }
    }
}
```

## 3. Bulk Operations for Large Libraries

### Efficient Library Management

```python
class LibraryManager:
    def __init__(self, client):
        self.client = client
        self.batch_size = 50  # Process in batches to avoid overwhelming the API

    def scan_library_issues(self):
        """Comprehensive library health scan"""
        print("Scanning library for issues...")

        issues = {
            'missing_files': [],
            'quality_cutoff_unmet': [],
            'unmonitored_available': [],
            'duplicate_movies': [],
            'path_issues': [],
            'metadata_issues': []
        }

        # Get all movies in batches
        all_movies = []
        page = 1
        while True:
            movies_page = self.client.get_movies(page=page, pageSize=self.batch_size)
            if not movies_page.get('data'):
                break
            all_movies.extend(movies_page['data'])
            if len(movies_page['data']) < self.batch_size:
                break
            page += 1

        print(f"Analyzing {len(all_movies)} movies...")

        # Track for duplicate detection
        seen_tmdb_ids = {}

        for movie in all_movies:
            # Check for missing files
            if movie['monitored'] and not movie['hasFile']:
                if movie.get('isAvailable', False):
                    issues['missing_files'].append(movie)

            # Check for quality cutoff not met
            if movie['hasFile'] and movie.get('movieFile'):
                movie_file = movie['movieFile']
                if movie_file.get('qualityCutoffNotMet', False):
                    issues['quality_cutoff_unmet'].append(movie)

            # Check for unmonitored but available movies
            if not movie['monitored'] and movie.get('isAvailable', False):
                issues['unmonitored_available'].append(movie)

            # Check for duplicates
            tmdb_id = movie['tmdbId']
            if tmdb_id in seen_tmdb_ids:
                issues['duplicate_movies'].append({
                    'original': seen_tmdb_ids[tmdb_id],
                    'duplicate': movie
                })
            else:
                seen_tmdb_ids[tmdb_id] = movie

            # Check for path issues
            if movie.get('path') and not os.path.exists(movie['path']):
                issues['path_issues'].append(movie)

            # Check for metadata issues
            if not movie.get('overview') or not movie.get('images'):
                issues['metadata_issues'].append(movie)

        return issues

    def fix_missing_metadata(self, movies_with_issues):
        """Refresh metadata for movies with missing information"""
        print(f"Refreshing metadata for {len(movies_with_issues)} movies...")

        for i, movie in enumerate(movies_with_issues):
            try:
                # Trigger metadata refresh
                self.client._request('PUT', f'/movie/{movie["id"]}/refresh')
                print(f"Progress: {i+1}/{len(movies_with_issues)} - Refreshed {movie['title']}")

                # Rate limiting
                time.sleep(1)

            except Exception as e:
                print(f"Failed to refresh {movie['title']}: {e}")

    def bulk_quality_update(self, quality_profile_id, filters=None):
        """Update quality profile for multiple movies"""
        filters = filters or {}
        movies = self.client.get_movies(**filters)

        if isinstance(movies, dict) and 'data' in movies:
            movies = movies['data']

        print(f"Updating quality profile for {len(movies)} movies...")

        successful = 0
        failed = 0

        for i, movie in enumerate(movies):
            try:
                # Update movie with new quality profile
                movie['qualityProfileId'] = quality_profile_id

                self.client._request('PUT', f'/movie/{movie["id"]}', json=movie)
                successful += 1

                if (i + 1) % 10 == 0:
                    print(f"Progress: {i+1}/{len(movies)} - {successful} successful, {failed} failed")

                # Rate limiting
                time.sleep(0.5)

            except Exception as e:
                failed += 1
                print(f"Failed to update {movie['title']}: {e}")

        print(f"Quality update complete: {successful} successful, {failed} failed")

    def bulk_monitoring_update(self, monitored_status, filters=None):
        """Bulk update monitoring status"""
        filters = filters or {}
        movies = self.client.get_movies(**filters)

        if isinstance(movies, dict) and 'data' in movies:
            movies = movies['data']

        print(f"Setting monitoring to {monitored_status} for {len(movies)} movies...")

        # Group movies into batches
        batches = [movies[i:i + self.batch_size] for i in range(0, len(movies), self.batch_size)]

        for batch_num, batch in enumerate(batches):
            try:
                # Prepare batch update
                movie_updates = []
                for movie in batch:
                    movie['monitored'] = monitored_status
                    movie_updates.append(movie)

                # Send batch update (if API supports it, otherwise update individually)
                for movie in movie_updates:
                    self.client._request('PUT', f'/movie/{movie["id"]}', json=movie)
                    time.sleep(0.1)  # Small delay between updates

                print(f"Batch {batch_num + 1}/{len(batches)} complete")

            except Exception as e:
                print(f"Batch {batch_num + 1} failed: {e}")

    def cleanup_failed_downloads(self):
        """Clean up failed downloads and blacklisted releases"""
        print("Cleaning up failed downloads...")

        # Get blacklist
        try:
            blacklist = self.client._request('GET', '/blacklist').json()

            print(f"Found {len(blacklist)} blacklisted releases")

            # Remove old blacklist entries (older than 30 days)
            cutoff_date = datetime.now() - timedelta(days=30)

            for item in blacklist:
                item_date = datetime.fromisoformat(item['date'].replace('Z', '+00:00'))
                if item_date < cutoff_date:
                    try:
                        self.client._request('DELETE', f'/blacklist/{item["id"]}')
                        print(f"Removed old blacklist entry: {item.get('title', 'Unknown')}")
                    except Exception as e:
                        print(f"Failed to remove blacklist entry {item['id']}: {e}")

                    time.sleep(0.2)

        except Exception as e:
            print(f"Failed to clean blacklist: {e}")

    def generate_library_report(self):
        """Generate comprehensive library report"""
        print("Generating library report...")

        # Get library statistics
        all_movies = []
        page = 1
        while True:
            movies_page = self.client.get_movies(page=page, pageSize=100)
            if isinstance(movies_page, dict) and 'data' in movies_page:
                movies_data = movies_page['data']
            else:
                movies_data = movies_page if isinstance(movies_page, list) else []

            if not movies_data:
                break
            all_movies.extend(movies_data)
            if len(movies_data) < 100:
                break
            page += 1

        # Calculate statistics
        stats = {
            'total_movies': len(all_movies),
            'monitored': sum(1 for m in all_movies if m['monitored']),
            'has_file': sum(1 for m in all_movies if m['hasFile']),
            'missing': sum(1 for m in all_movies if m['monitored'] and not m['hasFile']),
            'available_missing': sum(1 for m in all_movies
                                   if m['monitored'] and not m['hasFile'] and m.get('isAvailable')),
            'quality_profiles': {},
            'years': {},
            'genres': {},
            'file_sizes': 0
        }

        # Detailed analysis
        for movie in all_movies:
            # Quality profile distribution
            profile_id = movie.get('qualityProfileId')
            if profile_id:
                stats['quality_profiles'][profile_id] = stats['quality_profiles'].get(profile_id, 0) + 1

            # Year distribution
            year = movie.get('year')
            if year:
                stats['years'][year] = stats['years'].get(year, 0) + 1

            # Genre distribution
            for genre in movie.get('genres', []):
                stats['genres'][genre] = stats['genres'].get(genre, 0) + 1

            # File sizes
            if movie.get('movieFile', {}).get('size'):
                stats['file_sizes'] += movie['movieFile']['size']

        # Generate report
        report = f"""
RADARR LIBRARY REPORT
====================
Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}

OVERVIEW
--------
Total Movies: {stats['total_movies']}
Monitored: {stats['monitored']} ({stats['monitored']/stats['total_movies']*100:.1f}%)
With Files: {stats['has_file']} ({stats['has_file']/stats['total_movies']*100:.1f}%)
Missing: {stats['missing']} ({stats['missing']/stats['total_movies']*100:.1f}%)
Available Missing: {stats['available_missing']}

STORAGE
-------
Total Library Size: {stats['file_sizes'] / (1024**3):.2f} GB

TOP GENRES
----------
"""
        # Add top 10 genres
        top_genres = sorted(stats['genres'].items(), key=lambda x: x[1], reverse=True)[:10]
        for genre, count in top_genres:
            report += f"{genre}: {count} movies\n"

        report += "\nTOP YEARS\n---------\n"
        # Add top 10 years
        top_years = sorted(stats['years'].items(), key=lambda x: x[1], reverse=True)[:10]
        for year, count in top_years:
            report += f"{year}: {count} movies\n"

        return report

# Usage example
def main():
    client = RadarrClient('http://localhost:7878', 'your-api-key')
    manager = LibraryManager(client)

    # Comprehensive library scan
    print("Starting library health check...")
    issues = manager.scan_library_issues()

    print(f"\nLibrary Issues Found:")
    print(f"- Missing files: {len(issues['missing_files'])}")
    print(f"- Quality cutoff unmet: {len(issues['quality_cutoff_unmet'])}")
    print(f"- Unmonitored available: {len(issues['unmonitored_available'])}")
    print(f"- Duplicates: {len(issues['duplicate_movies'])}")
    print(f"- Path issues: {len(issues['path_issues'])}")
    print(f"- Metadata issues: {len(issues['metadata_issues'])}")

    # Fix metadata issues
    if issues['metadata_issues']:
        choice = input(f"\nFix {len(issues['metadata_issues'])} metadata issues? (y/n): ")
        if choice.lower() == 'y':
            manager.fix_missing_metadata(issues['metadata_issues'])

    # Generate and save report
    report = manager.generate_library_report()
    with open('radarr_library_report.txt', 'w') as f:
        f.write(report)
    print("\nLibrary report saved to radarr_library_report.txt")

    # Cleanup
    manager.cleanup_failed_downloads()

if __name__ == "__main__":
    main()
```

## 4. Real-time Event Handling via WebSocket

### WebSocket Event Processing

```javascript
class RadarrWebSocketManager {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl.replace(/^http/, 'ws');
        this.apiKey = apiKey;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.eventHandlers = new Map();
        this.isConnected = false;

        // Built-in event handlers
        this.setupDefaultHandlers();
    }

    setupDefaultHandlers() {
        // Connection management
        this.on('connected', () => {
            console.log('WebSocket connected to Radarr');
            this.reconnectAttempts = 0;
            this.isConnected = true;
        });

        this.on('disconnected', () => {
            console.log('WebSocket disconnected from Radarr');
            this.isConnected = false;
            this.attemptReconnect();
        });

        this.on('error', (error) => {
            console.error('WebSocket error:', error);
        });
    }

    connect() {
        try {
            const wsUrl = `${this.baseUrl}/signalr/radarr`;
            this.ws = new WebSocket(wsUrl, {
                headers: {
                    'X-API-Key': this.apiKey
                }
            });

            this.ws.onopen = () => {
                this.emit('connected');

                // Send handshake
                this.send({
                    protocol: 'json',
                    version: 1
                });

                // Subscribe to all events
                this.send({
                    type: 1,
                    target: 'Subscribe',
                    arguments: ['movie', 'queue', 'health', 'system']
                });
            };

            this.ws.onclose = (event) => {
                this.emit('disconnected', event);
            };

            this.ws.onerror = (error) => {
                this.emit('error', error);
            };

            this.ws.onmessage = (event) => {
                this.handleMessage(event.data);
            };

        } catch (error) {
            this.emit('error', error);
        }
    }

    handleMessage(data) {
        try {
            const message = JSON.parse(data);

            // Handle different message types
            switch (message.type) {
                case 1: // Invocation
                    this.handleInvocation(message);
                    break;
                case 2: // StreamItem
                    this.handleStreamItem(message);
                    break;
                case 3: // Completion
                    this.handleCompletion(message);
                    break;
                case 6: // Ping
                    this.send({ type: 6 }); // Pong
                    break;
                default:
                    console.log('Unknown message type:', message);
            }
        } catch (error) {
            console.error('Error parsing WebSocket message:', error);
        }
    }

    handleInvocation(message) {
        const { target, arguments: args } = message;

        // Map SignalR events to our event system
        switch (target) {
            case 'movieUpdated':
                this.emit('movie.updated', args[0]);
                break;
            case 'movieAdded':
                this.emit('movie.added', args[0]);
                break;
            case 'movieDeleted':
                this.emit('movie.deleted', args[0]);
                break;
            case 'queueUpdated':
                this.emit('queue.updated', args[0]);
                break;
            case 'downloadComplete':
                this.emit('download.complete', args[0]);
                break;
            case 'downloadFailed':
                this.emit('download.failed', args[0]);
                break;
            case 'healthUpdated':
                this.emit('health.updated', args[0]);
                break;
            case 'systemUpdated':
                this.emit('system.updated', args[0]);
                break;
            default:
                this.emit('raw', { target, args });
        }
    }

    on(event, handler) {
        if (!this.eventHandlers.has(event)) {
            this.eventHandlers.set(event, []);
        }
        this.eventHandlers.get(event).push(handler);
    }

    off(event, handler) {
        if (this.eventHandlers.has(event)) {
            const handlers = this.eventHandlers.get(event);
            const index = handlers.indexOf(handler);
            if (index > -1) {
                handlers.splice(index, 1);
            }
        }
    }

    emit(event, data) {
        if (this.eventHandlers.has(event)) {
            this.eventHandlers.get(event).forEach(handler => {
                try {
                    handler(data);
                } catch (error) {
                    console.error(`Error in event handler for ${event}:`, error);
                }
            });
        }
    }

    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        }
    }

    attemptReconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.pow(2, this.reconnectAttempts) * 1000; // Exponential backoff

            console.log(`Attempting to reconnect in ${delay/1000} seconds... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Max reconnection attempts reached');
            this.emit('max-reconnects-reached');
        }
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
}

// Advanced event processing system
class RadarrEventProcessor {
    constructor(wsManager) {
        this.wsManager = wsManager;
        this.movieCache = new Map();
        this.queueCache = new Map();
        this.eventLog = [];
        this.maxEventLogSize = 1000;

        this.setupEventHandlers();
    }

    setupEventHandlers() {
        // Movie events
        this.wsManager.on('movie.added', (movie) => {
            this.movieCache.set(movie.id, movie);
            this.logEvent('movie.added', movie);

            console.log(`🎬 Movie added: ${movie.title} (${movie.year})`);

            // Trigger custom workflows
            this.onMovieAdded(movie);
        });

        this.wsManager.on('movie.updated', (movie) => {
            const oldMovie = this.movieCache.get(movie.id);
            this.movieCache.set(movie.id, movie);
            this.logEvent('movie.updated', { old: oldMovie, new: movie });

            // Detect specific changes
            if (oldMovie) {
                if (!oldMovie.hasFile && movie.hasFile) {
                    console.log(`📁 Movie file added: ${movie.title}`);
                    this.onMovieFileAdded(movie, oldMovie);
                }

                if (oldMovie.monitored !== movie.monitored) {
                    console.log(`👁 Monitoring changed for ${movie.title}: ${movie.monitored ? 'enabled' : 'disabled'}`);
                }

                if (oldMovie.qualityProfileId !== movie.qualityProfileId) {
                    console.log(`⚙️ Quality profile changed for ${movie.title}`);
                }
            }
        });

        this.wsManager.on('movie.deleted', (movieId) => {
            const movie = this.movieCache.get(movieId);
            this.movieCache.delete(movieId);
            this.logEvent('movie.deleted', { id: movieId, movie });

            if (movie) {
                console.log(`🗑 Movie deleted: ${movie.title}`);
                this.onMovieDeleted(movie);
            }
        });

        // Download events
        this.wsManager.on('download.complete', (download) => {
            console.log(`✅ Download completed: ${download.title}`);
            this.onDownloadComplete(download);
        });

        this.wsManager.on('download.failed', (download) => {
            console.log(`❌ Download failed: ${download.title} - ${download.errorMessage}`);
            this.onDownloadFailed(download);
        });

        // Queue events
        this.wsManager.on('queue.updated', (queue) => {
            this.updateQueueCache(queue);
            this.onQueueUpdated(queue);
        });

        // Health events
        this.wsManager.on('health.updated', (health) => {
            console.log(`🏥 Health status: ${health.status}`);
            if (health.issues && health.issues.length > 0) {
                console.warn(`⚠️ Health issues detected: ${health.issues.length}`);
                health.issues.forEach(issue => {
                    console.warn(`  - ${issue.type}: ${issue.message}`);
                });
            }
            this.onHealthUpdated(health);
        });
    }

    logEvent(type, data) {
        const event = {
            timestamp: new Date(),
            type,
            data
        };

        this.eventLog.push(event);

        // Keep log size manageable
        if (this.eventLog.length > this.maxEventLogSize) {
            this.eventLog.shift();
        }
    }

    updateQueueCache(queueItems) {
        // Update individual queue items
        if (Array.isArray(queueItems)) {
            queueItems.forEach(item => {
                this.queueCache.set(item.id, item);
            });
        } else if (queueItems.id) {
            this.queueCache.set(queueItems.id, queueItems);
        }
    }

    // Override these methods for custom behavior
    async onMovieAdded(movie) {
        // Example: Send notification
        await this.sendNotification(`New movie added: ${movie.title}`, 'info');
    }

    async onMovieFileAdded(movie, oldMovie) {
        // Example: Update external systems
        await this.notifyExternalSystems('movie_available', {
            tmdbId: movie.tmdbId,
            title: movie.title,
            year: movie.year,
            filePath: movie.movieFile?.path
        });
    }

    async onMovieDeleted(movie) {
        // Example: Cleanup external references
        await this.cleanupExternalReferences(movie.tmdbId);
    }

    async onDownloadComplete(download) {
        // Example: Trigger post-processing
        await this.triggerPostProcessing(download);
    }

    async onDownloadFailed(download) {
        // Example: Log failure and attempt retry logic
        await this.handleDownloadFailure(download);
    }

    onQueueUpdated(queue) {
        // Example: Update UI components
        this.updateQueueUI(queue);
    }

    async onHealthUpdated(health) {
        if (health.status === 'error') {
            await this.sendNotification('Radarr health issues detected!', 'error');
        }
    }

    // Helper methods (implement as needed)
    async sendNotification(message, type = 'info') {
        // Implement notification system (email, Slack, Discord, etc.)
        console.log(`NOTIFICATION [${type.toUpperCase()}]: ${message}`);
    }

    async notifyExternalSystems(event, data) {
        // Example: Notify Plex, Jellyfin, etc.
        console.log(`EXTERNAL NOTIFICATION: ${event}`, data);
    }

    async cleanupExternalReferences(tmdbId) {
        // Cleanup external system references
        console.log(`CLEANUP: ${tmdbId}`);
    }

    async triggerPostProcessing(download) {
        // Trigger custom post-processing scripts
        console.log(`POST-PROCESS: ${download.title}`);
    }

    async handleDownloadFailure(download) {
        // Implement retry logic, alternative source searching, etc.
        console.log(`FAILURE HANDLING: ${download.title}`);
    }

    updateQueueUI(queue) {
        // Update user interface components
        console.log(`UI UPDATE: Queue status changed`);
    }

    // Utility methods
    getRecentEvents(minutes = 60) {
        const cutoff = new Date(Date.now() - minutes * 60 * 1000);
        return this.eventLog.filter(event => event.timestamp > cutoff);
    }

    getMovieById(id) {
        return this.movieCache.get(id);
    }

    getQueueItem(id) {
        return this.queueCache.get(id);
    }

    getEventStats(hours = 24) {
        const cutoff = new Date(Date.now() - hours * 60 * 60 * 1000);
        const recentEvents = this.eventLog.filter(event => event.timestamp > cutoff);

        const stats = {};
        recentEvents.forEach(event => {
            stats[event.type] = (stats[event.type] || 0) + 1;
        });

        return {
            total: recentEvents.length,
            breakdown: stats,
            timespan: `${hours} hours`
        };
    }
}

// Usage example
async function setupRealTimeMonitoring() {
    const wsManager = new RadarrWebSocketManager('http://localhost:7878', 'your-api-key');
    const processor = new RadarrEventProcessor(wsManager);

    // Custom event handlers
    processor.onMovieAdded = async (movie) => {
        // Send to Discord/Slack
        await sendDiscordMessage(`📽️ **New Movie Added**\n${movie.title} (${movie.year})\nTMDB: ${movie.tmdbId}`);

        // Update external database
        await updateMovieDatabase({
            tmdbId: movie.tmdbId,
            title: movie.title,
            year: movie.year,
            status: 'monitored',
            addedDate: new Date()
        });
    };

    processor.onDownloadComplete = async (download) => {
        // Notify Plex to refresh library
        await refreshPlexLibrary();

        // Send completion notification
        await sendDiscordMessage(`✅ **Download Complete**\n${download.title}`);

        // Trigger custom post-processing
        await runPostProcessingScript(download.path);
    };

    // Connect and start monitoring
    wsManager.connect();

    // Graceful shutdown handling
    process.on('SIGINT', () => {
        console.log('Shutting down WebSocket connection...');
        wsManager.disconnect();
        process.exit(0);
    });

    // Monitor connection health
    setInterval(() => {
        if (!wsManager.isConnected) {
            console.warn('WebSocket not connected, attempting reconnection...');
            wsManager.connect();
        }
    }, 30000); // Check every 30 seconds
}

// Run the real-time monitoring
setupRealTimeMonitoring().catch(console.error);
```

This completes the Common Integration Patterns section. The patterns shown here provide production-ready solutions for the most common Radarr Go API integration scenarios, with comprehensive error handling, rate limiting, and real-world considerations.
