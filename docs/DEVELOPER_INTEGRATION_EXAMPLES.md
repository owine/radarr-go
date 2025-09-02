# Radarr Go API - Developer Integration Guide

This comprehensive guide provides practical examples for integrating with the Radarr Go API, covering all major features and use cases.

## Table of Contents

- [Getting Started](#getting-started)
- [Authentication](#authentication)
- [Movie Management](#movie-management)
- [Search and Downloads](#search-and-downloads)
- [Quality Management](#quality-management)
- [Health Monitoring](#health-monitoring)
- [Real-time Updates](#real-time-updates)
- [Error Handling](#error-handling)
- [Client Libraries](#client-libraries)
- [Advanced Patterns](#advanced-patterns)

## Getting Started

### Quick Start Example

```javascript
// Basic API client setup
class RadarrGoClient {
  constructor(baseUrl, apiKey) {
    this.baseUrl = baseUrl.replace(/\/$/, '');
    this.apiKey = apiKey;
  }

  async request(endpoint, options = {}) {
    const url = `${this.baseUrl}/api/v3${endpoint}`;
    const config = {
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': this.apiKey,
        ...options.headers
      },
      ...options
    };

    const response = await fetch(url, config);

    if (!response.ok) {
      const error = await response.json().catch(() => ({}));
      throw new Error(`API Error ${response.status}: ${error.error || response.statusText}`);
    }

    return response.json();
  }

  // Health check
  async ping() {
    return this.request('/ping');
  }

  // System status
  async getSystemStatus() {
    return this.request('/system/status');
  }
}

// Usage
const radarr = new RadarrGoClient('http://localhost:7878', 'your_api_key_here');

// Test connection
try {
  const status = await radarr.getSystemStatus();
  console.log(`Connected to Radarr Go ${status.version}`);
} catch (error) {
  console.error('Connection failed:', error.message);
}
```

## Authentication

### API Key Authentication

```python
import requests
import json

class RadarrAPI:
    def __init__(self, base_url, api_key):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        })

    def get(self, endpoint, params=None):
        response = self.session.get(f"{self.base_url}/api/v3{endpoint}", params=params)
        response.raise_for_status()
        return response.json()

    def post(self, endpoint, data=None):
        response = self.session.post(f"{self.base_url}/api/v3{endpoint}", json=data)
        response.raise_for_status()
        return response.json()

    def put(self, endpoint, data):
        response = self.session.put(f"{self.base_url}/api/v3{endpoint}", json=data)
        response.raise_for_status()
        return response.json()

    def delete(self, endpoint):
        response = self.session.delete(f"{self.base_url}/api/v3{endpoint}")
        response.raise_for_status()
        return response.status_code == 204

# Example usage
radarr = RadarrAPI('http://localhost:7878', 'your_api_key_here')

# Test authentication
try:
    status = radarr.get('/system/status')
    print(f"✅ Authentication successful - Radarr Go {status['version']}")
except requests.exceptions.HTTPError as e:
    if e.response.status_code == 401:
        print("❌ Authentication failed - Check your API key")
    else:
        print(f"❌ API Error: {e}")
```

### Environment-Based Configuration

```bash
# .env file
RADARR_URL=http://localhost:7878
RADARR_API_KEY=your_api_key_here
RADARR_TIMEOUT=30
```

```go
// Go example with environment configuration
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"
)

type RadarrClient struct {
    BaseURL string
    APIKey  string
    Client  *http.Client
}

type SystemStatus struct {
    Version      string `json:"version"`
    RuntimeName  string `json:"runtimeName"`
    StartTime    string `json:"startTime"`
    Authentication string `json:"authentication"`
}

func NewRadarrClient() *RadarrClient {
    return &RadarrClient{
        BaseURL: os.Getenv("RADARR_URL"),
        APIKey:  os.Getenv("RADARR_API_KEY"),
        Client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (r *RadarrClient) GetSystemStatus() (*SystemStatus, error) {
    req, err := http.NewRequest("GET", r.BaseURL+"/api/v3/system/status", nil)
    if err != nil {
        return nil, err
    }

    req.Header.Set("X-API-Key", r.APIKey)
    req.Header.Set("Accept", "application/json")

    resp, err := r.Client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API error: %d", resp.StatusCode)
    }

    var status SystemStatus
    err = json.NewDecoder(resp.Body).Decode(&status)
    return &status, err
}

func main() {
    client := NewRadarrClient()

    status, err := client.GetSystemStatus()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Connected to Radarr Go %s (%s)\n", status.Version, status.RuntimeName)
}
```

## Movie Management

### Adding Movies

```javascript
// Add a movie from TMDB
async function addMovie(radarr, tmdbId, options = {}) {
  // First, lookup movie details
  const searchResults = await radarr.request(`/movie/lookup/tmdb?tmdbId=${tmdbId}`);

  if (!searchResults) {
    throw new Error(`Movie with TMDB ID ${tmdbId} not found`);
  }

  // Prepare movie data
  const movieData = {
    tmdbId: tmdbId,
    title: searchResults.title,
    titleSlug: searchResults.titleSlug,
    year: searchResults.year,
    qualityProfileId: options.qualityProfileId || 1,
    monitored: options.monitored !== false,
    minimumAvailability: options.minimumAvailability || 'released',
    rootFolderPath: options.rootFolderPath || '/movies',
    tags: options.tags || [],
    addOptions: {
      searchForMovie: options.searchForMovie !== false,
      monitor: options.monitor || 'movieOnly'
    }
  };

  // Add the movie
  const movie = await radarr.request('/movie', {
    method: 'POST',
    body: JSON.stringify(movieData)
  });

  console.log(`✅ Added movie: ${movie.title} (${movie.year})`);
  return movie;
}

// Example usage
const newMovie = await addMovie(radarr, 603, {  // The Matrix TMDB ID
  qualityProfileId: 1,
  monitored: true,
  searchForMovie: true,
  rootFolderPath: '/movies'
});
```

### Bulk Movie Operations

```python
def bulk_movie_operations(radarr):
    """Demonstrate bulk movie operations"""

    # Get all movies
    all_movies = radarr.get('/movie')
    print(f"Total movies: {len(all_movies)}")

    # Filter unmonitored movies
    unmonitored = [m for m in all_movies if not m['monitored']]
    print(f"Unmonitored movies: {len(unmonitored)}")

    # Bulk monitor movies from a specific year
    movies_2023 = [m for m in all_movies if m['year'] == 2023]

    for movie in movies_2023:
        if not movie['monitored']:
            movie['monitored'] = True
            updated = radarr.put(f"/movie/{movie['id']}", movie)
            print(f"✅ Now monitoring: {updated['title']}")

    # Get movies missing files
    missing_movies = radarr.get('/wanted/missing', {
        'page': 1,
        'pageSize': 50,
        'sortKey': 'title',
        'sortDirection': 'asc'
    })

    print(f"Missing movies: {missing_movies['totalRecords']}")

    return {
        'total': len(all_movies),
        'unmonitored': len(unmonitored),
        'missing': missing_movies['totalRecords']
    }

# Example usage
stats = bulk_movie_operations(radarr)
print(f"Movie statistics: {stats}")
```

### Movie File Management

```javascript
// Movie file operations
class MovieFileManager {
  constructor(radarrClient) {
    this.radarr = radarrClient;
  }

  async getMovieFiles(movieId = null) {
    const endpoint = movieId ? `/moviefile?movieId=${movieId}` : '/moviefile';
    return this.radarr.request(endpoint);
  }

  async deleteMovieFile(fileId, deleteFromDisk = false) {
    return this.radarr.request(`/moviefile/${fileId}?deleteFromDisk=${deleteFromDisk}`, {
      method: 'DELETE'
    });
  }

  async getFileInfo(fileId) {
    return this.radarr.request(`/moviefile/${fileId}`);
  }

  async organizeMovieFiles() {
    // Get file organization status
    const organizations = await this.radarr.request('/fileorganization');

    // Trigger scan for new files
    await this.radarr.request('/fileorganization/scan', {
      method: 'POST',
      body: JSON.stringify({ path: '/downloads/completed' })
    });

    return organizations;
  }

  async renameMovieFiles(movieIds) {
    // Preview renames
    const preview = await this.radarr.request('/rename/preview', {
      method: 'GET'
    });

    console.log(`Previewing ${preview.length} renames`);

    // Execute renames
    const result = await this.radarr.request('/rename', {
      method: 'POST',
      body: JSON.stringify({ movieIds })
    });

    return result;
  }
}

// Usage example
const fileManager = new MovieFileManager(radarr);

// Get all movie files
const files = await fileManager.getMovieFiles();
console.log(`Total movie files: ${files.length}`);

// Get files for specific movie
const movieFiles = await fileManager.getMovieFiles(1);
console.log(`Files for movie 1: ${movieFiles.length}`);

// Organize and rename files
await fileManager.organizeMovieFiles();
```

## Search and Downloads

### Release Search and Management

```python
class SearchManager:
    def __init__(self, radarr_client):
        self.radarr = radarr_client

    def search_movie_releases(self, movie_id, indexer_ids=None):
        """Search for releases of a specific movie"""
        endpoint = f'/search/movie/{movie_id}'
        params = {}

        if indexer_ids:
            params['indexerIds'] = indexer_ids

        releases = self.radarr.get(endpoint, params)

        # Sort by quality and size
        def quality_score(release):
            quality_weights = {
                'Bluray-2160p': 100,
                'Bluray-1080p': 90,
                'Bluray-720p': 80,
                'WEB-2160p': 85,
                'WEB-1080p': 75,
                'WEB-720p': 65
            }
            return quality_weights.get(release.get('quality', {}).get('quality', {}).get('name'), 0)

        releases.sort(key=quality_score, reverse=True)
        return releases

    def interactive_search(self, movie_id):
        """Perform interactive search with manual selection"""
        releases = self.radarr.get(f'/search/interactive?movieId={movie_id}')

        print(f"Found {len(releases)} releases:")
        for i, release in enumerate(releases[:10]):  # Show top 10
            quality = release.get('quality', {}).get('quality', {}).get('name', 'Unknown')
            size_gb = release.get('size', 0) / (1024**3)
            age = release.get('ageHours', 0) / 24

            print(f"{i+1:2d}. {release.get('title', 'Unknown')[:60]}")
            print(f"    Quality: {quality} | Size: {size_gb:.2f}GB | Age: {age:.1f}d")
            print(f"    Indexer: {release.get('indexer', 'Unknown')}")
            print()

        return releases

    def grab_release(self, release):
        """Grab a specific release"""
        grab_data = {
            'guid': release['guid'],
            'indexerId': release['indexerId'],
            'movieId': release.get('movieId')
        }

        result = self.radarr.post('/release/grab', grab_data)
        print(f"✅ Grabbed: {release.get('title', 'Unknown')}")
        return result

    def auto_search(self, movie_id):
        """Automatically search and grab best release"""
        releases = self.search_movie_releases(movie_id)

        if not releases:
            print("❌ No releases found")
            return None

        # Filter for approved releases
        approved_releases = [r for r in releases if r.get('approved', False)]

        if approved_releases:
            best_release = approved_releases[0]  # Already sorted by quality
            return self.grab_release(best_release)
        else:
            print("❌ No approved releases found")
            return None

# Usage example
search = SearchManager(radarr)

# Interactive search
releases = search.interactive_search(movie_id=1)

# Grab the first release
if releases:
    search.grab_release(releases[0])

# Auto search for best quality
search.auto_search(movie_id=2)
```

### Queue Management

```javascript
// Download queue monitoring and management
class QueueManager {
  constructor(radarrClient) {
    this.radarr = radarrClient;
  }

  async getQueue(options = {}) {
    const params = new URLSearchParams({
      page: options.page || 1,
      pageSize: options.pageSize || 50,
      sortKey: options.sortKey || 'timeleft',
      sortDirection: options.sortDirection || 'asc',
      includeUnknownMovieItems: options.includeUnknownMovieItems || false
    });

    return this.radarr.request(`/queue?${params}`);
  }

  async getQueueStats() {
    return this.radarr.request('/queue/stats');
  }

  async removeQueueItem(id, options = {}) {
    const params = new URLSearchParams({
      removeFromClient: options.removeFromClient !== false,
      blocklist: options.blocklist || false
    });

    return this.radarr.request(`/queue/${id}?${params}`, {
      method: 'DELETE'
    });
  }

  async bulkRemoveQueueItems(ids, options = {}) {
    return this.radarr.request('/queue/bulk', {
      method: 'DELETE',
      body: JSON.stringify({
        ids: ids,
        removeFromClient: options.removeFromClient !== false,
        blocklist: options.blocklist || false
      })
    });
  }

  async monitorQueue() {
    const queue = await this.getQueue();
    const stats = await this.getQueueStats();

    console.log(`📊 Queue Statistics:`);
    console.log(`   Total: ${stats.totalCount}`);
    console.log(`   Active: ${stats.count}`);
    console.log(`   Errors: ${stats.errors ? 'Yes' : 'No'}`);
    console.log(`   Warnings: ${stats.warnings ? 'Yes' : 'No'}`);

    console.log(`\n📥 Current Downloads:`);
    queue.records.forEach(item => {
      const progress = item.sizeleft === 0 ? 100 :
        ((item.size - item.sizeleft) / item.size * 100);

      console.log(`   ${item.title}`);
      console.log(`   Progress: ${progress.toFixed(1)}% | ETA: ${item.timeleft || 'Unknown'}`);
      console.log(`   Status: ${item.status} | Client: ${item.downloadClient}`);
      console.log();
    });

    return { queue, stats };
  }

  async cleanupFailedItems() {
    const queue = await this.getQueue({ pageSize: 250 });
    const failedItems = queue.records.filter(item =>
      item.status === 'failed' || item.trackedDownloadStatus === 'warning'
    );

    console.log(`🧹 Found ${failedItems.length} failed/problematic items`);

    for (const item of failedItems) {
      console.log(`Removing failed item: ${item.title}`);
      await this.removeQueueItem(item.id, { removeFromClient: true, blocklist: false });

      // Small delay to avoid rate limiting
      await new Promise(resolve => setTimeout(resolve, 500));
    }

    return failedItems.length;
  }
}

// Usage
const queueManager = new QueueManager(radarr);

// Monitor queue
await queueManager.monitorQueue();

// Cleanup failed downloads
const removedCount = await queueManager.cleanupFailedItems();
console.log(`✅ Removed ${removedCount} failed items`);
```

## Quality Management

### Quality Profile Management

```python
def manage_quality_profiles(radarr):
    """Comprehensive quality profile management"""

    # Get all quality profiles
    profiles = radarr.get('/qualityprofile')
    print(f"Current quality profiles: {len(profiles)}")

    for profile in profiles:
        print(f"  - {profile['name']} (ID: {profile['id']})")

    # Create a new 4K profile
    new_profile = {
        'name': '4K Ultra HD',
        'upgradeAllowed': True,
        'cutoff': 18,  # Bluray-2160p
        'items': [
            {
                'id': 18,
                'name': 'Bluray-2160p',
                'allowed': True
            },
            {
                'id': 19,
                'name': 'WEB-2160p',
                'allowed': True
            },
            {
                'id': 16,
                'name': 'HDTV-2160p',
                'allowed': False
            }
        ],
        'minFormatScore': 0,
        'cutoffFormatScore': 0,
        'formatItems': []
    }

    try:
        created_profile = radarr.post('/qualityprofile', new_profile)
        print(f"✅ Created profile: {created_profile['name']}")
    except Exception as e:
        print(f"❌ Failed to create profile: {e}")

    # Get quality definitions
    definitions = radarr.get('/qualitydefinition')
    print(f"\nQuality definitions: {len(definitions)}")

    # Update quality definition sizes
    for definition in definitions:
        if definition['name'] == 'Bluray-1080p':
            definition['minSize'] = 8000  # 8GB minimum
            definition['maxSize'] = 25000  # 25GB maximum

            updated = radarr.put(f"/qualitydefinition/{definition['id']}", definition)
            print(f"✅ Updated {definition['name']} size limits")

    return profiles

# Custom format management
def manage_custom_formats(radarr):
    """Create and manage custom formats"""

    # HDR Custom Format
    hdr_format = {
        'name': 'HDR',
        'includeCustomFormatWhenRenaming': True,
        'specifications': [
            {
                'name': 'HDR',
                'implementation': 'ReleaseTitleSpecification',
                'negate': False,
                'required': True,
                'fields': {
                    'value': r'\b(HDR|HDR10|HDR10Plus|DolbyVision)\b'
                }
            }
        ]
    }

    try:
        created_format = radarr.post('/customformat', hdr_format)
        print(f"✅ Created custom format: {created_format['name']}")
        return created_format
    except Exception as e:
        print(f"❌ Failed to create custom format: {e}")
        return None

# Usage
profiles = manage_quality_profiles(radarr)
hdr_format = manage_custom_formats(radarr)
```

## Health Monitoring

### Comprehensive Health Monitoring

```javascript
// Advanced health monitoring system
class HealthMonitor {
  constructor(radarrClient) {
    this.radarr = radarrClient;
  }

  async getHealthDashboard() {
    return this.radarr.request('/health/dashboard');
  }

  async getHealthChecks() {
    return this.radarr.request('/health');
  }

  async getSystemResources() {
    return this.radarr.request('/health/system/resources');
  }

  async getDiskSpace() {
    return this.radarr.request('/health/system/diskspace');
  }

  async getPerformanceMetrics(timeRange = '1h') {
    const endTime = new Date();
    const startTime = new Date(endTime.getTime() - this.parseTimeRange(timeRange));

    return this.radarr.request('/health/metrics', {
      method: 'GET',
      headers: {
        'Accept': 'application/json'
      }
    }, {
      startTime: startTime.toISOString(),
      endTime: endTime.toISOString(),
      interval: 'minute'
    });
  }

  parseTimeRange(range) {
    const units = { h: 3600000, m: 60000, d: 86400000 };
    const match = range.match(/^(\d+)([hmd])$/);
    return match ? parseInt(match[1]) * units[match[2]] : 3600000;
  }

  async generateHealthReport() {
    console.log('🏥 Radarr Go Health Report');
    console.log('='.repeat(50));

    try {
      // System resources
      const resources = await this.getSystemResources();
      console.log(`\n💻 System Resources:`);
      console.log(`   CPU Usage: ${resources.cpuUsage.toFixed(1)}%`);
      console.log(`   Memory Usage: ${resources.memoryUsage.toFixed(1)}% (${this.formatBytes(resources.usedMemory)}/${this.formatBytes(resources.totalMemory)})`);
      console.log(`   Uptime: ${resources.uptime}`);

      // Disk space
      const diskSpace = await this.getDiskSpace();
      console.log(`\n💾 Disk Space:`);
      diskSpace.forEach(disk => {
        console.log(`   ${disk.label} (${disk.path}): ${disk.freeSpacePercentage.toFixed(1)}% free (${this.formatBytes(disk.freeSpace)})`);
        if (disk.freeSpacePercentage < 10) {
          console.log(`   ⚠️  WARNING: Low disk space on ${disk.label}`);
        }
      });

      // Health checks
      const healthChecks = await this.getHealthChecks();
      console.log(`\n🔍 Health Checks:`);

      const healthSummary = healthChecks.reduce((acc, check) => {
        acc[check.type] = (acc[check.type] || 0) + 1;
        return acc;
      }, {});

      console.log(`   ✅ OK: ${healthSummary.ok || 0}`);
      console.log(`   ℹ️  Notice: ${healthSummary.notice || 0}`);
      console.log(`   ⚠️  Warning: ${healthSummary.warning || 0}`);
      console.log(`   ❌ Error: ${healthSummary.error || 0}`);

      // Show critical issues
      const criticalIssues = healthChecks.filter(check => check.type === 'error');
      if (criticalIssues.length > 0) {
        console.log(`\n🚨 Critical Issues:`);
        criticalIssues.forEach(issue => {
          console.log(`   - ${issue.message}`);
          if (issue.wikiUrl) {
            console.log(`     Help: ${issue.wikiUrl}`);
          }
        });
      }

      // Performance metrics
      const metrics = await this.getPerformanceMetrics();
      console.log(`\n📈 Performance Metrics:`);
      console.log(`   Average Response Time: ${metrics.averageResponseTime.toFixed(1)}ms`);
      console.log(`   Requests/Second: ${metrics.requestsPerSecond.toFixed(2)}`);
      console.log(`   Error Rate: ${metrics.errorRate.toFixed(2)}%`);
      console.log(`   DB Connection Time: ${metrics.databaseConnectionTime.toFixed(1)}ms`);

      return {
        resources,
        diskSpace,
        healthSummary,
        criticalIssues,
        metrics
      };

    } catch (error) {
      console.error('❌ Failed to generate health report:', error);
      throw error;
    }
  }

  formatBytes(bytes) {
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = bytes;
    let unitIndex = 0;

    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }

    return `${size.toFixed(1)} ${units[unitIndex]}`;
  }

  async startHealthMonitoring(interval = 300000) { // 5 minutes
    console.log(`🔄 Starting health monitoring (interval: ${interval/1000}s)`);

    const monitor = async () => {
      try {
        const report = await this.generateHealthReport();

        // Alert on critical issues
        if (report.criticalIssues.length > 0) {
          console.log(`\n🚨 ALERT: ${report.criticalIssues.length} critical issues detected!`);
          // Here you could send notifications via webhook, email, etc.
        }

        // Alert on low disk space
        const lowSpaceDisks = report.diskSpace.filter(disk => disk.freeSpacePercentage < 10);
        if (lowSpaceDisks.length > 0) {
          console.log(`\n⚠️  ALERT: Low disk space on ${lowSpaceDisks.map(d => d.label).join(', ')}`);
        }

        // Alert on high resource usage
        if (report.resources.memoryUsage > 90) {
          console.log(`\n⚠️  ALERT: High memory usage (${report.resources.memoryUsage.toFixed(1)}%)`);
        }

      } catch (error) {
        console.error('Health monitoring error:', error);
      }
    };

    // Initial check
    await monitor();

    // Schedule regular checks
    setInterval(monitor, interval);
  }
}

// Usage
const healthMonitor = new HealthMonitor(radarr);

// Generate one-time report
await healthMonitor.generateHealthReport();

// Start continuous monitoring
await healthMonitor.startHealthMonitoring(300000); // Every 5 minutes
```

## Real-time Updates

### WebSocket Integration

```python
import asyncio
import json
import websockets
import logging

class RadarrWebSocketClient:
    def __init__(self, ws_url, api_key):
        self.ws_url = ws_url.replace('http', 'ws').rstrip('/') + '/signalr'
        self.api_key = api_key
        self.websocket = None
        self.handlers = {}

    def on(self, event_type, handler):
        """Register event handler"""
        if event_type not in self.handlers:
            self.handlers[event_type] = []
        self.handlers[event_type].append(handler)

    def emit(self, event_type, data):
        """Emit event to all registered handlers"""
        if event_type in self.handlers:
            for handler in self.handlers[event_type]:
                try:
                    handler(data)
                except Exception as e:
                    logging.error(f"Error in event handler: {e}")

    async def connect(self):
        """Connect to WebSocket"""
        headers = {'X-API-Key': self.api_key}

        try:
            self.websocket = await websockets.connect(self.ws_url, extra_headers=headers)
            logging.info("✅ WebSocket connected")

            # Send initial handshake
            await self.websocket.send(json.dumps({
                'protocol': 'json',
                'version': 1
            }))

            return True

        except Exception as e:
            logging.error(f"❌ WebSocket connection failed: {e}")
            return False

    async def listen(self):
        """Listen for WebSocket messages"""
        if not self.websocket:
            raise Exception("WebSocket not connected")

        try:
            async for message in self.websocket:
                try:
                    data = json.loads(message)

                    # Handle different message types
                    if 'type' in data:
                        if data['type'] == 1:  # Hub message
                            self.handle_hub_message(data)
                        elif data['type'] == 3:  # Close message
                            logging.info("WebSocket closed by server")
                            break

                except json.JSONDecodeError:
                    logging.warning(f"Invalid JSON received: {message}")
                except Exception as e:
                    logging.error(f"Error processing message: {e}")

        except websockets.exceptions.ConnectionClosed:
            logging.info("WebSocket connection closed")
        except Exception as e:
            logging.error(f"WebSocket error: {e}")

    def handle_hub_message(self, data):
        """Handle SignalR hub messages"""
        if 'target' in data and 'arguments' in data:
            target = data['target']
            arguments = data.get('arguments', [])

            # Map SignalR targets to our events
            event_mapping = {
                'movieUpdated': 'movie_updated',
                'movieDeleted': 'movie_deleted',
                'movieFileUpdated': 'movie_file_updated',
                'queueUpdated': 'queue_updated',
                'healthUpdated': 'health_updated',
                'taskUpdated': 'task_updated'
            }

            event_type = event_mapping.get(target, target)
            self.emit(event_type, arguments[0] if arguments else {})

    async def close(self):
        """Close WebSocket connection"""
        if self.websocket:
            await self.websocket.close()
            self.websocket = None

# Usage example
async def setup_realtime_monitoring():
    ws_client = RadarrWebSocketClient('ws://localhost:7878', 'your_api_key')

    # Register event handlers
    def on_movie_updated(data):
        print(f"🎬 Movie updated: {data.get('title', 'Unknown')} ({data.get('id')})")

    def on_queue_updated(data):
        print(f"📥 Queue updated: {data.get('title', 'Unknown')} - {data.get('status', 'Unknown')}")

    def on_health_updated(data):
        health_type = data.get('type', 'unknown')
        message = data.get('message', 'Unknown issue')

        if health_type == 'error':
            print(f"🚨 Health Error: {message}")
        elif health_type == 'warning':
            print(f"⚠️  Health Warning: {message}")
        else:
            print(f"ℹ️  Health Update: {message}")

    def on_task_updated(data):
        task_name = data.get('name', 'Unknown')
        status = data.get('status', 'Unknown')

        if status == 'completed':
            print(f"✅ Task completed: {task_name}")
        elif status == 'failed':
            print(f"❌ Task failed: {task_name}")
            if 'exception' in data:
                print(f"   Error: {data['exception']}")
        else:
            print(f"🔄 Task {status}: {task_name}")

    # Register handlers
    ws_client.on('movie_updated', on_movie_updated)
    ws_client.on('queue_updated', on_queue_updated)
    ws_client.on('health_updated', on_health_updated)
    ws_client.on('task_updated', on_task_updated)

    # Connect and listen
    if await ws_client.connect():
        print("🔄 Starting real-time monitoring...")
        try:
            await ws_client.listen()
        except KeyboardInterrupt:
            print("\n⏹️  Stopping monitoring...")
        finally:
            await ws_client.close()

# Run the monitoring
if __name__ == "__main__":
    asyncio.run(setup_realtime_monitoring())
```

## Error Handling

### Robust Error Handling Patterns

```javascript
// Comprehensive error handling for API operations
class RadarrErrorHandler {
  static async withRetry(operation, maxRetries = 3, backoffMs = 1000) {
    let lastError;

    for (let attempt = 1; attempt <= maxRetries; attempt++) {
      try {
        return await operation();
      } catch (error) {
        lastError = error;

        // Don't retry on client errors (4xx)
        if (error.status >= 400 && error.status < 500) {
          throw error;
        }

        if (attempt < maxRetries) {
          console.warn(`Attempt ${attempt} failed, retrying in ${backoffMs}ms...`);
          await new Promise(resolve => setTimeout(resolve, backoffMs));
          backoffMs *= 2; // Exponential backoff
        }
      }
    }

    throw lastError;
  }

  static handleApiError(error) {
    if (error.status === 401) {
      return {
        type: 'AUTHENTICATION_ERROR',
        message: 'Invalid API key or authentication failed',
        recoverable: true,
        action: 'Check your API key configuration'
      };
    }

    if (error.status === 404) {
      return {
        type: 'NOT_FOUND',
        message: 'Resource not found',
        recoverable: false,
        action: 'Verify the resource ID or endpoint'
      };
    }

    if (error.status === 429) {
      return {
        type: 'RATE_LIMITED',
        message: 'Rate limit exceeded',
        recoverable: true,
        action: 'Wait before making more requests'
      };
    }

    if (error.status >= 500) {
      return {
        type: 'SERVER_ERROR',
        message: 'Server error occurred',
        recoverable: true,
        action: 'Retry the request or check server status'
      };
    }

    return {
      type: 'UNKNOWN_ERROR',
      message: error.message || 'Unknown error occurred',
      recoverable: false,
      action: 'Check the error details and try again'
    };
  }
}

// Enhanced API client with error handling
class RobustRadarrClient {
  constructor(baseUrl, apiKey, options = {}) {
    this.baseUrl = baseUrl.replace(/\/$/, '');
    this.apiKey = apiKey;
    this.options = {
      maxRetries: options.maxRetries || 3,
      timeout: options.timeout || 30000,
      ...options
    };
  }

  async request(endpoint, options = {}) {
    const operation = async () => {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), this.options.timeout);

      try {
        const response = await fetch(`${this.baseUrl}/api/v3${endpoint}`, {
          headers: {
            'Content-Type': 'application/json',
            'X-API-Key': this.apiKey,
            ...options.headers
          },
          signal: controller.signal,
          ...options
        });

        clearTimeout(timeoutId);

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          const error = new Error(errorData.error || response.statusText);
          error.status = response.status;
          error.response = response;
          error.data = errorData;
          throw error;
        }

        return response.json();
      } catch (error) {
        clearTimeout(timeoutId);

        if (error.name === 'AbortError') {
          const timeoutError = new Error('Request timed out');
          timeoutError.status = 408;
          throw timeoutError;
        }

        throw error;
      }
    };

    try {
      return await RadarrErrorHandler.withRetry(operation, this.options.maxRetries);
    } catch (error) {
      const errorInfo = RadarrErrorHandler.handleApiError(error);

      console.error(`API Error: ${errorInfo.message}`);
      console.error(`Action: ${errorInfo.action}`);

      // Add error context
      error.errorInfo = errorInfo;
      throw error;
    }
  }

  async safeRequest(endpoint, options = {}) {
    try {
      return {
        success: true,
        data: await this.request(endpoint, options)
      };
    } catch (error) {
      return {
        success: false,
        error: error.errorInfo || RadarrErrorHandler.handleApiError(error),
        originalError: error
      };
    }
  }
}

// Usage examples
const radarr = new RobustRadarrClient('http://localhost:7878', 'your_api_key', {
  maxRetries: 3,
  timeout: 30000
});

// With automatic retry and error handling
try {
  const movies = await radarr.request('/movie');
  console.log(`Retrieved ${movies.length} movies`);
} catch (error) {
  console.error('Failed to retrieve movies:', error.errorInfo);
}

// Safe request that doesn't throw
const result = await radarr.safeRequest('/movie/999'); // Non-existent movie
if (result.success) {
  console.log('Movie:', result.data);
} else {
  console.log('Error:', result.error.message);
  console.log('Recoverable:', result.error.recoverable);
}
```

## Client Libraries

### TypeScript SDK Example

```typescript
// Type definitions for Radarr Go API
interface Movie {
  id: number;
  title: string;
  originalTitle: string;
  year: number;
  tmdbId: number;
  imdbId: string;
  titleSlug: string;
  monitored: boolean;
  hasFile: boolean;
  movieFileId?: number;
  qualityProfileId: number;
  minimumAvailability: 'tba' | 'announced' | 'inCinemas' | 'released' | 'preDB';
  status: 'tba' | 'announced' | 'inCinemas' | 'released' | 'deleted';
  overview: string;
  inCinemas?: string;
  physicalRelease?: string;
  digitalRelease?: string;
  runtime: number;
  genres: string[];
  tags: number[];
  rootFolderPath: string;
  certification: string;
  website?: string;
  youTubeTrailerId?: string;
  studio: string;
  path: string;
  folderName: string;
  sizeOnDisk: number;
  isAvailable: boolean;
  popularity: number;
  createdAt: string;
  updatedAt: string;
}

interface SystemStatus {
  version: string;
  buildTime: string;
  isDebug: boolean;
  isProduction: boolean;
  startupPath: string;
  appData: string;
  osName: string;
  osVersion: string;
  isLinux: boolean;
  isOsx: boolean;
  isWindows: boolean;
  mode: string;
  branch: string;
  authentication: 'none' | 'apikey' | 'basic';
  migrationVersion: number;
  urlBase: string;
  runtimeVersion: string;
  runtimeName: string;
  startTime: string;
}

interface QueueItem {
  id: number;
  movieId: number;
  movie?: Movie;
  size: number;
  title: string;
  sizeleft: number;
  timeleft?: string;
  estimatedCompletionTime?: string;
  added: string;
  status: string;
  trackedDownloadStatus: string;
  trackedDownloadState: string;
  errorMessage?: string;
  downloadId: string;
  protocol: 'usenet' | 'torrent';
  downloadClient: string;
  indexer: string;
  outputPath: string;
}

interface HealthCheck {
  source: string;
  type: 'ok' | 'notice' | 'warning' | 'error';
  message: string;
  wikiUrl?: string;
}

// Main SDK class
export class RadarrGoSDK {
  private baseUrl: string;
  private apiKey: string;
  private timeout: number;

  constructor(baseUrl: string, apiKey: string, timeout: number = 30000) {
    this.baseUrl = baseUrl.replace(/\/$/, '');
    this.apiKey = apiKey;
    this.timeout = timeout;
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.timeout);

    try {
      const response = await fetch(`${this.baseUrl}/api/v3${endpoint}`, {
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': this.apiKey,
          ...options.headers,
        },
        signal: controller.signal,
        ...options,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || response.statusText);
      }

      return response.json();
    } finally {
      clearTimeout(timeoutId);
    }
  }

  // System methods
  async getSystemStatus(): Promise<SystemStatus> {
    return this.request<SystemStatus>('/system/status');
  }

  async ping(): Promise<{ message: string }> {
    return this.request<{ message: string }>('/ping');
  }

  // Movie methods
  async getMovies(params?: {
    tmdbId?: number;
    monitored?: boolean;
    hasFile?: boolean;
    qualityProfileId?: number;
  }): Promise<Movie[]> {
    const searchParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, value.toString());
        }
      });
    }

    const endpoint = `/movie${searchParams.toString() ? `?${searchParams}` : ''}`;
    return this.request<Movie[]>(endpoint);
  }

  async getMovie(id: number): Promise<Movie> {
    return this.request<Movie>(`/movie/${id}`);
  }

  async addMovie(movieData: Partial<Movie>): Promise<Movie> {
    return this.request<Movie>('/movie', {
      method: 'POST',
      body: JSON.stringify(movieData),
    });
  }

  async updateMovie(id: number, movieData: Partial<Movie>): Promise<Movie> {
    return this.request<Movie>(`/movie/${id}`, {
      method: 'PUT',
      body: JSON.stringify(movieData),
    });
  }

  async deleteMovie(
    id: number,
    options?: { deleteFiles?: boolean; addImportListExclusion?: boolean }
  ): Promise<void> {
    const params = new URLSearchParams();
    if (options?.deleteFiles) params.append('deleteFiles', 'true');
    if (options?.addImportListExclusion) params.append('addImportListExclusion', 'true');

    await this.request<void>(`/movie/${id}?${params}`, { method: 'DELETE' });
  }

  async searchMovies(term: string, year?: number): Promise<Movie[]> {
    const params = new URLSearchParams({ term });
    if (year) params.append('year', year.toString());

    return this.request<Movie[]>(`/movie/lookup?${params}`);
  }

  // Queue methods
  async getQueue(params?: {
    page?: number;
    pageSize?: number;
    sortKey?: string;
    sortDirection?: 'asc' | 'desc';
    includeUnknownMovieItems?: boolean;
  }): Promise<{
    page: number;
    pageSize: number;
    sortKey: string;
    sortDirection: string;
    totalRecords: number;
    records: QueueItem[];
  }> {
    const searchParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, value.toString());
        }
      });
    }

    return this.request(`/queue?${searchParams}`);
  }

  async removeQueueItem(
    id: number,
    options?: { removeFromClient?: boolean; blocklist?: boolean }
  ): Promise<void> {
    const params = new URLSearchParams();
    if (options?.removeFromClient !== undefined) {
      params.append('removeFromClient', options.removeFromClient.toString());
    }
    if (options?.blocklist) {
      params.append('blocklist', 'true');
    }

    await this.request<void>(`/queue/${id}?${params}`, { method: 'DELETE' });
  }

  // Health methods
  async getHealth(): Promise<HealthCheck[]> {
    return this.request<HealthCheck[]>('/health');
  }

  async getHealthDashboard(): Promise<{
    healthChecks: HealthCheck[];
    systemResources: any;
    diskSpace: any[];
    performanceMetrics: any;
    healthIssues: any[];
  }> {
    return this.request('/health/dashboard');
  }

  // Task methods
  async getTasks(params?: {
    status?: 'queued' | 'started' | 'completed' | 'failed' | 'cancelled';
    commandType?: string;
    includeCompletedCommands?: boolean;
  }): Promise<any[]> {
    const searchParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          searchParams.append(key, value.toString());
        }
      });
    }

    return this.request(`/command?${searchParams}`);
  }

  async queueTask(taskData: {
    name: string;
    movieId?: number;
    movieIds?: number[];
    sendUpdatesToClient?: boolean;
  }): Promise<any> {
    return this.request('/command', {
      method: 'POST',
      body: JSON.stringify(taskData),
    });
  }
}

// Usage example
const radarr = new RadarrGoSDK('http://localhost:7878', 'your_api_key');

async function example() {
  try {
    // Get system status
    const status = await radarr.getSystemStatus();
    console.log(`Connected to Radarr Go ${status.version}`);

    // Get all monitored movies
    const movies = await radarr.getMovies({ monitored: true });
    console.log(`Found ${movies.length} monitored movies`);

    // Search for a movie
    const searchResults = await radarr.searchMovies('The Matrix', 1999);
    console.log(`Found ${searchResults.length} results for The Matrix`);

    // Check health
    const healthChecks = await radarr.getHealth();
    const errors = healthChecks.filter(h => h.type === 'error');
    if (errors.length > 0) {
      console.warn(`${errors.length} health errors detected`);
    }

  } catch (error) {
    console.error('API Error:', error);
  }
}

export default RadarrGoSDK;
```

## Advanced Patterns

### Batch Operations with Rate Limiting

```python
import asyncio
from typing import List, Callable, Any
import time

class BatchProcessor:
    def __init__(self, radarr_client, rate_limit_per_second=10):
        self.radarr = radarr_client
        self.rate_limit = rate_limit_per_second
        self.last_request_time = 0

    async def rate_limited_request(self, func: Callable, *args, **kwargs):
        """Execute request with rate limiting"""
        current_time = time.time()
        time_since_last = current_time - self.last_request_time
        min_interval = 1.0 / self.rate_limit

        if time_since_last < min_interval:
            sleep_time = min_interval - time_since_last
            await asyncio.sleep(sleep_time)

        self.last_request_time = time.time()
        return await func(*args, **kwargs)

    async def batch_process_movies(self, movie_ids: List[int], operation: str, **kwargs):
        """Process movies in batches with rate limiting"""
        results = []
        errors = []

        print(f"Processing {len(movie_ids)} movies with operation: {operation}")

        for i, movie_id in enumerate(movie_ids):
            try:
                if operation == 'refresh':
                    result = await self.rate_limited_request(
                        self.radarr.put, f'/movie/{movie_id}/refresh'
                    )
                elif operation == 'update':
                    result = await self.rate_limited_request(
                        self.radarr.put, f'/movie/{movie_id}', kwargs.get('data', {})
                    )
                elif operation == 'search':
                    result = await self.rate_limited_request(
                        self.radarr.get, f'/search/movie/{movie_id}'
                    )
                else:
                    raise ValueError(f"Unknown operation: {operation}")

                results.append({'movie_id': movie_id, 'result': result})
                print(f"✅ Processed movie {movie_id} ({i+1}/{len(movie_ids)})")

            except Exception as e:
                errors.append({'movie_id': movie_id, 'error': str(e)})
                print(f"❌ Failed to process movie {movie_id}: {e}")

        return {
            'successful': len(results),
            'failed': len(errors),
            'results': results,
            'errors': errors
        }

# Usage example
async def bulk_operations_example():
    processor = BatchProcessor(radarr_client, rate_limit_per_second=5)

    # Get all unmonitored movies
    all_movies = await radarr_client.get('/movie')
    unmonitored_ids = [m['id'] for m in all_movies if not m['monitored']]

    # Batch update to monitor them
    update_data = {'monitored': True}
    result = await processor.batch_process_movies(
        unmonitored_ids[:10],  # Process first 10
        'update',
        data=update_data
    )

    print(f"Batch update completed: {result['successful']} successful, {result['failed']} failed")

# Run the example
# asyncio.run(bulk_operations_example())
```

### Comprehensive Integration Example

```javascript
// Complete integration example showing various patterns
class RadarrIntegration {
  constructor(baseUrl, apiKey) {
    this.radarr = new RobustRadarrClient(baseUrl, apiKey);
    this.healthMonitor = new HealthMonitor(this.radarr);
    this.queueManager = new QueueManager(this.radarr);
  }

  async setupNewInstance() {
    console.log('🚀 Setting up new Radarr Go instance...');

    try {
      // 1. Verify connection
      const status = await this.radarr.request('/system/status');
      console.log(`✅ Connected to Radarr Go ${status.version}`);

      // 2. Setup quality profiles
      await this.setupQualityProfiles();

      // 3. Setup root folders
      await this.setupRootFolders();

      // 4. Configure indexers (example with test data)
      await this.setupIndexers();

      // 5. Configure download clients
      await this.setupDownloadClients();

      // 6. Setup notifications
      await this.setupNotifications();

      // 7. Import existing movies (if any)
      await this.importExistingMovies();

      console.log('✅ Radarr Go instance setup completed!');

    } catch (error) {
      console.error('❌ Setup failed:', error);
      throw error;
    }
  }

  async setupQualityProfiles() {
    console.log('📊 Setting up quality profiles...');

    const profiles = [
      {
        name: 'HD-1080p',
        cutoff: 7, // Bluray-1080p
        items: [
          { quality: { id: 7, name: 'Bluray-1080p' }, allowed: true },
          { quality: { id: 9, name: 'HDTV-1080p' }, allowed: true },
          { quality: { id: 3, name: 'WEB-1080p' }, allowed: true }
        ],
        upgradeAllowed: true,
        minFormatScore: 0,
        cutoffFormatScore: 0
      },
      {
        name: '4K-UHD',
        cutoff: 18, // Bluray-2160p
        items: [
          { quality: { id: 18, name: 'Bluray-2160p' }, allowed: true },
          { quality: { id: 19, name: 'WEB-2160p' }, allowed: true }
        ],
        upgradeAllowed: true,
        minFormatScore: 0,
        cutoffFormatScore: 0
      }
    ];

    for (const profile of profiles) {
      try {
        const result = await this.radarr.safeRequest('/qualityprofile', {
          method: 'POST',
          body: JSON.stringify(profile)
        });

        if (result.success) {
          console.log(`✅ Created quality profile: ${profile.name}`);
        } else {
          console.log(`⚠️  Quality profile might already exist: ${profile.name}`);
        }
      } catch (error) {
        console.error(`Failed to create quality profile ${profile.name}:`, error);
      }
    }
  }

  async setupRootFolders() {
    console.log('📁 Setting up root folders...');

    const rootFolders = [
      { path: '/movies' },
      { path: '/movies-4k' }
    ];

    for (const folder of rootFolders) {
      try {
        const result = await this.radarr.safeRequest('/rootfolder', {
          method: 'POST',
          body: JSON.stringify(folder)
        });

        if (result.success) {
          console.log(`✅ Added root folder: ${folder.path}`);
        } else {
          console.log(`⚠️  Root folder might already exist: ${folder.path}`);
        }
      } catch (error) {
        console.error(`Failed to add root folder ${folder.path}:`, error);
      }
    }
  }

  async setupIndexers() {
    console.log('🔍 Setting up indexers...');

    // This is just an example - real indexers would need actual configuration
    const exampleIndexers = [
      {
        name: 'Example Indexer',
        implementation: 'Newznab',
        settings: {
          baseUrl: 'https://api.example.com',
          apiKey: 'your_indexer_api_key',
          categories: [2000, 2010, 2020, 2030, 2040, 2045, 2050, 2060]
        },
        enableRss: true,
        enableAutomaticSearch: true,
        enableInteractiveSearch: true,
        priority: 25
      }
    ];

    for (const indexer of exampleIndexers) {
      console.log(`⚠️  Example indexer configuration - update with real settings: ${indexer.name}`);
    }
  }

  async setupDownloadClients() {
    console.log('⬇️ Setting up download clients...');

    // Example download client configurations
    console.log('⚠️  Configure download clients manually or add configuration here');
  }

  async setupNotifications() {
    console.log('📢 Setting up notifications...');

    // Example Discord notification
    const discordNotification = {
      name: 'Discord',
      implementation: 'Discord',
      settings: {
        webhookUrl: 'https://discord.com/api/webhooks/your/webhook/url',
        username: 'Radarr Go',
        avatar: 'https://github.com/radarr/radarr-go/raw/main/web/static/logo.png'
      },
      onGrab: true,
      onDownload: true,
      onUpgrade: true,
      onHealthIssue: true,
      onApplicationUpdate: true
    };

    console.log('⚠️  Configure notification settings with your actual webhook URLs');
  }

  async importExistingMovies() {
    console.log('📥 Checking for existing movies to import...');

    // This would scan root folders for existing movies
    const result = await this.radarr.safeRequest('/import/manual');

    if (result.success && result.data.length > 0) {
      console.log(`Found ${result.data.length} potential imports`);
      // Process imports here
    } else {
      console.log('No existing movies found for import');
    }
  }

  async dailyMaintenance() {
    console.log('🧹 Running daily maintenance...');

    try {
      // 1. Generate health report
      const healthReport = await this.healthMonitor.generateHealthReport();

      // 2. Clean up failed queue items
      const cleanupCount = await this.queueManager.cleanupFailedItems();
      console.log(`Cleaned up ${cleanupCount} failed queue items`);

      // 3. Check for movies without files
      const missingMovies = await this.radarr.request('/wanted/missing?pageSize=10');
      console.log(`Movies missing files: ${missingMovies.totalRecords}`);

      // 4. Update movie metadata for recently added movies
      const recentMovies = await this.radarr.request('/movie?sortKey=added&sortDirection=desc&pageSize=5');
      for (const movie of recentMovies) {
        if (this.isRecentlyAdded(movie.added)) {
          await this.radarr.request(`/movie/${movie.id}/refresh`, { method: 'PUT' });
          console.log(`Refreshed metadata for: ${movie.title}`);
        }
      }

      // 5. Check disk space
      if (healthReport.diskSpace) {
        const lowSpaceDisks = healthReport.diskSpace.filter(disk => disk.freeSpacePercentage < 15);
        if (lowSpaceDisks.length > 0) {
          console.warn(`⚠️  Low disk space on: ${lowSpaceDisks.map(d => d.label).join(', ')}`);
        }
      }

      console.log('✅ Daily maintenance completed');

    } catch (error) {
      console.error('❌ Daily maintenance failed:', error);
    }
  }

  isRecentlyAdded(addedDate) {
    const added = new Date(addedDate);
    const now = new Date();
    const diffHours = (now - added) / (1000 * 60 * 60);
    return diffHours < 24; // Within last 24 hours
  }

  async startAutomation() {
    console.log('🤖 Starting automated monitoring...');

    // Health monitoring every 5 minutes
    setInterval(async () => {
      try {
        const health = await this.radarr.request('/health');
        const errors = health.filter(h => h.type === 'error');
        if (errors.length > 0) {
          console.warn(`🚨 Health errors detected: ${errors.length}`);
        }
      } catch (error) {
        console.error('Health check failed:', error);
      }
    }, 5 * 60 * 1000);

    // Queue monitoring every minute
    setInterval(async () => {
      try {
        const stats = await this.queueManager.getQueueStats();
        if (stats.errors) {
          console.warn('⚠️  Queue has errors');
        }
      } catch (error) {
        console.error('Queue check failed:', error);
      }
    }, 60 * 1000);

    // Daily maintenance at 2 AM
    const scheduleDaily = () => {
      const now = new Date();
      const next2AM = new Date(now);
      next2AM.setHours(2, 0, 0, 0);

      if (next2AM <= now) {
        next2AM.setDate(next2AM.getDate() + 1);
      }

      const msUntil2AM = next2AM.getTime() - now.getTime();

      setTimeout(() => {
        this.dailyMaintenance();
        setInterval(() => this.dailyMaintenance(), 24 * 60 * 60 * 1000);
      }, msUntil2AM);

      console.log(`⏰ Daily maintenance scheduled for ${next2AM.toLocaleString()}`);
    };

    scheduleDaily();
  }
}

// Usage
const integration = new RadarrIntegration('http://localhost:7878', 'your_api_key');

// Setup new instance
await integration.setupNewInstance();

// Start automation
await integration.startAutomation();
```

## Summary

This comprehensive developer integration guide provides:

✅ **Complete API Coverage**: Examples for all major Radarr Go API features
✅ **Multiple Languages**: JavaScript, Python, TypeScript, and Go examples
✅ **Production-Ready Patterns**: Error handling, rate limiting, retry logic
✅ **Real-time Integration**: WebSocket support for live updates
✅ **Batch Operations**: Efficient bulk processing with rate limiting
✅ **Health Monitoring**: Comprehensive system monitoring and alerting
✅ **TypeScript SDK**: Full type-safe SDK with complete API coverage
✅ **Advanced Patterns**: Complex integration scenarios and automation

The examples are designed to be practical, production-ready, and demonstrate best practices for integrating with the Radarr Go API. Each pattern can be adapted to your specific use case and requirements.

For interactive testing and exploration, use the built-in Swagger UI at `/static/swagger-ui.html` with your Radarr Go instance.
