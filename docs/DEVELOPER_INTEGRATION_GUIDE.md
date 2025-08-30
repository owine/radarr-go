# Radarr Go Developer Integration Guide

**Version**: v1.0.0-go
**API Version**: v3
**Last Updated**: 2024-01-01

## 🎯 Quick Start

Get your Radarr Go integration up and running in minutes with this comprehensive developer guide.

### 📋 Prerequisites

- Radarr Go instance running (default: `http://localhost:7878`)
- API key (found in Settings → General → Security)
- Basic knowledge of REST APIs and JSON

### 🚀 5-Minute Integration

```bash
# 1. Test connectivity
curl -H "X-API-Key: your-api-key" http://localhost:7878/api/v3/system/status

# 2. List your movies
curl -H "X-API-Key: your-api-key" http://localhost:7878/api/v3/movie

# 3. Search for new movies
curl -H "X-API-Key: your-api-key" "http://localhost:7878/api/v3/movie/lookup?term=inception"

# You're now connected! 🎉
```

## 🏗️ Integration Architecture Patterns

### 1. Simple Client Library Pattern

**Best for**: Basic automation, personal tools, simple integrations

```python
class RadarrClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({'X-API-Key': api_key})

    def get_movies(self, **filters):
        response = self.session.get(f"{self.base_url}/api/v3/movie", params=filters)
        response.raise_for_status()
        return response.json()

    def add_movie(self, movie_data):
        response = self.session.post(f"{self.base_url}/api/v3/movie", json=movie_data)
        response.raise_for_status()
        return response.json()
```

### 2. Service Layer Pattern

**Best for**: Production applications, complex business logic, multiple Radarr instances

```python
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import List, Optional

@dataclass
class Movie:
    id: int
    title: str
    year: int
    tmdb_id: int
    monitored: bool
    has_file: bool
    quality_profile_id: int

class MovieRepository(ABC):
    @abstractmethod
    def get_movies(self, filters: dict) -> List[Movie]:
        pass

    @abstractmethod
    def add_movie(self, movie: Movie) -> Movie:
        pass

class RadarrMovieRepository(MovieRepository):
    def __init__(self, client: RadarrClient):
        self.client = client

    def get_movies(self, filters: dict) -> List[Movie]:
        data = self.client.get_movies(**filters)
        return [self._to_movie(item) for item in data]

    def add_movie(self, movie: Movie) -> Movie:
        data = self.client.add_movie(self._from_movie(movie))
        return self._to_movie(data)

class MovieService:
    def __init__(self, repository: MovieRepository):
        self.repository = repository

    def sync_wanted_movies(self, external_list: List[dict]) -> dict:
        """Business logic for syncing movies from external sources"""
        results = {'added': [], 'skipped': [], 'errors': []}

        existing_movies = {m.tmdb_id: m for m in self.repository.get_movies({})}

        for movie_data in external_list:
            if movie_data['tmdb_id'] in existing_movies:
                results['skipped'].append(movie_data)
                continue

            try:
                new_movie = Movie(**movie_data)
                added = self.repository.add_movie(new_movie)
                results['added'].append(added)
            except Exception as e:
                results['errors'].append({'movie': movie_data, 'error': str(e)})

        return results
```

### 3. Event-Driven Pattern

**Best for**: Real-time applications, monitoring systems, reactive architectures

```python
import asyncio
import websockets
import json
from typing import Callable, Dict, Any

class RadarrEventManager:
    def __init__(self, ws_url: str, api_key: str):
        self.ws_url = f"{ws_url}/ws?apikey={api_key}"
        self.handlers: Dict[str, List[Callable]] = {}
        self.running = False

    def on(self, event_type: str, handler: Callable):
        """Register event handler"""
        if event_type not in self.handlers:
            self.handlers[event_type] = []
        self.handlers[event_type].append(handler)

    async def start(self):
        """Start listening for events"""
        self.running = True
        while self.running:
            try:
                async with websockets.connect(self.ws_url) as websocket:
                    print("Connected to Radarr Go WebSocket")
                    async for message in websocket:
                        await self._handle_message(json.loads(message))
            except Exception as e:
                print(f"WebSocket error: {e}")
                await asyncio.sleep(5)  # Reconnect delay

    async def _handle_message(self, data: dict):
        event_type = data.get('type', 'unknown')
        if event_type in self.handlers:
            for handler in self.handlers[event_type]:
                try:
                    if asyncio.iscoroutinefunction(handler):
                        await handler(data)
                    else:
                        handler(data)
                except Exception as e:
                    print(f"Handler error: {e}")

# Usage example
async def main():
    event_manager = RadarrEventManager('ws://localhost:7878', 'your-api-key')

    @event_manager.on('movieAdded')
    async def on_movie_added(data):
        print(f"New movie added: {data['movie']['title']}")
        # Trigger additional workflows
        await notify_team(f"New movie: {data['movie']['title']}")
        await update_external_database(data['movie'])

    @event_manager.on('healthAlert')
    def on_health_alert(data):
        print(f"Health alert: {data['message']}")
        # Send to monitoring system
        send_alert_to_slack(data['message'])

    await event_manager.start()
```

## 🔧 Common Integration Patterns

### 1. Bulk Movie Import from External Sources

```python
import requests
import time
from typing import List, Dict, Any

class BulkMovieImporter:
    def __init__(self, radarr_client):
        self.radarr = radarr_client
        self.batch_size = 10
        self.delay_between_batches = 1.0

    def import_from_trakt_list(self, trakt_list_id: str, quality_profile_id: int,
                               root_folder: str) -> Dict[str, Any]:
        """Import movies from a Trakt list"""

        # 1. Fetch movies from Trakt
        trakt_movies = self._fetch_trakt_list(trakt_list_id)

        # 2. Get existing movies to avoid duplicates
        existing_movies = {m['tmdbId']: m for m in self.radarr.get_movies()}

        # 3. Filter new movies
        new_movies = [
            movie for movie in trakt_movies
            if movie['tmdb_id'] not in existing_movies
        ]

        # 4. Process in batches
        results = self._process_movie_batches(new_movies, quality_profile_id, root_folder)

        return {
            'total_processed': len(trakt_movies),
            'already_exists': len(trakt_movies) - len(new_movies),
            'imported': len(results['success']),
            'failed': len(results['errors']),
            'details': results
        }

    def _process_movie_batches(self, movies: List[dict], quality_profile_id: int,
                               root_folder: str) -> Dict[str, List]:
        results = {'success': [], 'errors': []}

        for i in range(0, len(movies), self.batch_size):
            batch = movies[i:i + self.batch_size]
            print(f"Processing batch {i//self.batch_size + 1}/{(len(movies) + self.batch_size - 1)//self.batch_size}")

            for movie in batch:
                try:
                    # Search for movie metadata
                    search_results = self.radarr.search_movies(movie['title'])
                    if not search_results:
                        results['errors'].append({
                            'movie': movie,
                            'error': 'Movie not found in search'
                        })
                        continue

                    # Find best match
                    best_match = self._find_best_match(search_results, movie)
                    if not best_match:
                        results['errors'].append({
                            'movie': movie,
                            'error': 'No suitable match found'
                        })
                        continue

                    # Prepare movie data
                    movie_data = {
                        'title': best_match['title'],
                        'tmdbId': best_match['tmdbId'],
                        'year': best_match['year'],
                        'qualityProfileId': quality_profile_id,
                        'rootFolderPath': root_folder,
                        'monitored': True,
                        'minimumAvailability': 'released',
                        'addOptions': {
                            'monitor': True,
                            'searchForMovie': True
                        }
                    }

                    # Add movie
                    added_movie = self.radarr.add_movie(movie_data)
                    results['success'].append(added_movie)
                    print(f"✅ Added: {added_movie['title']} ({added_movie['year']})")

                except Exception as e:
                    results['errors'].append({
                        'movie': movie,
                        'error': str(e)
                    })
                    print(f"❌ Failed to add {movie.get('title', 'Unknown')}: {e}")

            # Delay between batches to be respectful
            time.sleep(self.delay_between_batches)

        return results

    def _find_best_match(self, search_results: List[dict], target_movie: dict) -> dict:
        """Find the best matching movie from search results"""
        exact_matches = [
            movie for movie in search_results
            if movie['tmdbId'] == target_movie.get('tmdb_id')
        ]

        if exact_matches:
            return exact_matches[0]

        # Fallback to title and year matching
        title_matches = [
            movie for movie in search_results
            if movie['title'].lower() == target_movie['title'].lower() and
               movie['year'] == target_movie.get('year')
        ]

        return title_matches[0] if title_matches else None

# Usage
importer = BulkMovieImporter(radarr_client)
results = importer.import_from_trakt_list(
    trakt_list_id='popular-movies-2024',
    quality_profile_id=1,
    root_folder='/movies'
)
print(f"Import completed: {results['imported']} movies added, {results['failed']} failed")
```

### 2. Quality Profile Management

```python
class QualityProfileManager:
    def __init__(self, radarr_client):
        self.radarr = radarr_client

    def create_profile_from_template(self, name: str, template: str = 'standard') -> dict:
        """Create a new quality profile from a predefined template"""
        templates = {
            'standard': {
                'name': name,
                'cutoff': 7,  # Bluray-1080p
                'items': [
                    {'quality': {'id': 1, 'name': 'SDTV'}, 'allowed': False},
                    {'quality': {'id': 2, 'name': 'DVD'}, 'allowed': True},
                    {'quality': {'id': 4, 'name': 'HDTV-720p'}, 'allowed': True},
                    {'quality': {'id': 5, 'name': 'WEBDL-720p'}, 'allowed': True},
                    {'quality': {'id': 6, 'name': 'Bluray-720p'}, 'allowed': True},
                    {'quality': {'id': 3, 'name': 'WEBDL-1080p'}, 'allowed': True},
                    {'quality': {'id': 7, 'name': 'Bluray-1080p'}, 'allowed': True},
                    {'quality': {'id': 30, 'name': 'Remux-1080p'}, 'allowed': False}
                ],
                'upgradeAllowed': True,
                'language': 'english'
            },
            'high_quality': {
                'name': name,
                'cutoff': 30,  # Remux-1080p
                'items': [
                    {'quality': {'id': 7, 'name': 'Bluray-1080p'}, 'allowed': True},
                    {'quality': {'id': 30, 'name': 'Remux-1080p'}, 'allowed': True},
                    {'quality': {'id': 19, 'name': 'Bluray-2160p'}, 'allowed': True},
                    {'quality': {'id': 31, 'name': 'Remux-2160p'}, 'allowed': True}
                ],
                'upgradeAllowed': True,
                'language': 'english'
            }
        }

        if template not in templates:
            raise ValueError(f"Unknown template: {template}")

        return self.radarr.create_quality_profile(templates[template])

    def optimize_profiles_for_storage(self, max_size_gb: float = 5.0):
        """Optimize quality profiles based on storage constraints"""
        profiles = self.radarr.get_quality_profiles()

        for profile in profiles:
            # Calculate average movie size for current profile
            movies_with_profile = self.radarr.get_movies(qualityProfileId=profile['id'])
            avg_size = self._calculate_average_movie_size(movies_with_profile)

            if avg_size > max_size_gb:
                # Adjust profile to limit quality
                optimized_items = self._optimize_quality_items(profile['items'], max_size_gb)

                updated_profile = {
                    **profile,
                    'items': optimized_items,
                    'name': f"{profile['name']} (Optimized)"
                }

                self.radarr.update_quality_profile(profile['id'], updated_profile)
                print(f"Optimized profile: {profile['name']} (avg size: {avg_size:.1f}GB -> target: {max_size_gb}GB)")

    def _calculate_average_movie_size(self, movies: List[dict]) -> float:
        """Calculate average movie file size in GB"""
        total_size = sum(movie.get('sizeOnDisk', 0) for movie in movies if movie.get('hasFile'))
        movie_count = sum(1 for movie in movies if movie.get('hasFile'))

        if movie_count == 0:
            return 0

        return (total_size / movie_count) / (1024**3)  # Convert to GB
```

### 3. Monitoring and Health Dashboard

```python
import time
from datetime import datetime, timedelta
from dataclasses import dataclass
from typing import Dict, List, Optional

@dataclass
class HealthMetrics:
    timestamp: datetime
    status: str
    cpu_usage: float
    memory_usage: float
    disk_usage: float
    active_tasks: int
    queue_size: int
    indexer_status: Dict[str, bool]

class RadarrMonitor:
    def __init__(self, radarr_client):
        self.radarr = radarr_client
        self.metrics_history: List[HealthMetrics] = []
        self.alert_thresholds = {
            'cpu_usage': 80.0,
            'memory_usage': 85.0,
            'disk_usage': 90.0,
            'queue_size': 50,
            'failed_indexers': 3
        }

    def collect_metrics(self) -> HealthMetrics:
        """Collect current health metrics"""
        try:
            # Get health dashboard
            health_data = self.radarr.get_health_dashboard()

            # Get system resources
            resources = health_data.get('systemResources', {})

            # Get queue status
            queue = self.radarr.get_queue()

            # Get indexer status
            indexers = self.radarr.get_indexers()
            indexer_status = {
                indexer['name']: indexer.get('enable', False) and
                               indexer.get('supportsSearch', False)
                for indexer in indexers
            }

            # Get active tasks
            tasks = self.radarr.get_tasks()
            active_tasks = len([task for task in tasks if task['status'] in ['queued', 'started']])

            metrics = HealthMetrics(
                timestamp=datetime.now(),
                status=health_data.get('status', {}).get('status', 'unknown'),
                cpu_usage=resources.get('cpuUsage', 0),
                memory_usage=(resources.get('memoryUsage', 0) / resources.get('memoryTotal', 1)) * 100,
                disk_usage=self._calculate_disk_usage(),
                active_tasks=active_tasks,
                queue_size=len(queue) if isinstance(queue, list) else queue.get('totalRecords', 0),
                indexer_status=indexer_status
            )

            self.metrics_history.append(metrics)

            # Keep only last 24 hours of data
            cutoff = datetime.now() - timedelta(hours=24)
            self.metrics_history = [m for m in self.metrics_history if m.timestamp > cutoff]

            return metrics

        except Exception as e:
            print(f"Failed to collect metrics: {e}")
            return None

    def check_alerts(self, metrics: HealthMetrics) -> List[Dict[str, Any]]:
        """Check for alert conditions"""
        alerts = []

        # CPU usage alert
        if metrics.cpu_usage > self.alert_thresholds['cpu_usage']:
            alerts.append({
                'type': 'cpu_high',
                'severity': 'warning',
                'message': f"High CPU usage: {metrics.cpu_usage:.1f}%",
                'value': metrics.cpu_usage,
                'threshold': self.alert_thresholds['cpu_usage']
            })

        # Memory usage alert
        if metrics.memory_usage > self.alert_thresholds['memory_usage']:
            alerts.append({
                'type': 'memory_high',
                'severity': 'warning',
                'message': f"High memory usage: {metrics.memory_usage:.1f}%",
                'value': metrics.memory_usage,
                'threshold': self.alert_thresholds['memory_usage']
            })

        # Disk usage alert
        if metrics.disk_usage > self.alert_thresholds['disk_usage']:
            alerts.append({
                'type': 'disk_high',
                'severity': 'critical',
                'message': f"High disk usage: {metrics.disk_usage:.1f}%",
                'value': metrics.disk_usage,
                'threshold': self.alert_thresholds['disk_usage']
            })

        # Queue size alert
        if metrics.queue_size > self.alert_thresholds['queue_size']:
            alerts.append({
                'type': 'queue_large',
                'severity': 'info',
                'message': f"Large download queue: {metrics.queue_size} items",
                'value': metrics.queue_size,
                'threshold': self.alert_thresholds['queue_size']
            })

        # Failed indexers alert
        failed_indexers = [name for name, status in metrics.indexer_status.items() if not status]
        if len(failed_indexers) > self.alert_thresholds['failed_indexers']:
            alerts.append({
                'type': 'indexers_failed',
                'severity': 'warning',
                'message': f"Multiple indexers failed: {', '.join(failed_indexers[:3])}...",
                'value': len(failed_indexers),
                'threshold': self.alert_thresholds['failed_indexers']
            })

        return alerts

    def generate_report(self, hours: int = 24) -> Dict[str, Any]:
        """Generate a health report for the specified time period"""
        cutoff = datetime.now() - timedelta(hours=hours)
        recent_metrics = [m for m in self.metrics_history if m.timestamp > cutoff]

        if not recent_metrics:
            return {'error': 'No metrics available for the specified period'}

        # Calculate averages
        avg_cpu = sum(m.cpu_usage for m in recent_metrics) / len(recent_metrics)
        avg_memory = sum(m.memory_usage for m in recent_metrics) / len(recent_metrics)
        avg_queue = sum(m.queue_size for m in recent_metrics) / len(recent_metrics)

        # Find peaks
        peak_cpu = max(m.cpu_usage for m in recent_metrics)
        peak_memory = max(m.memory_usage for m in recent_metrics)
        peak_queue = max(m.queue_size for m in recent_metrics)

        # Status distribution
        status_counts = {}
        for m in recent_metrics:
            status_counts[m.status] = status_counts.get(m.status, 0) + 1

        return {
            'period': f"Last {hours} hours",
            'total_samples': len(recent_metrics),
            'averages': {
                'cpu_usage': round(avg_cpu, 1),
                'memory_usage': round(avg_memory, 1),
                'queue_size': round(avg_queue, 1)
            },
            'peaks': {
                'cpu_usage': round(peak_cpu, 1),
                'memory_usage': round(peak_memory, 1),
                'queue_size': int(peak_queue)
            },
            'status_distribution': status_counts,
            'current_status': recent_metrics[-1].status if recent_metrics else 'unknown'
        }

    def _calculate_disk_usage(self) -> float:
        """Calculate disk usage percentage"""
        try:
            disk_info = self.radarr.get_disk_space()
            if disk_info:
                total_free = sum(disk.get('freeSpace', 0) for disk in disk_info)
                total_space = sum(disk.get('totalSpace', 0) for disk in disk_info)
                if total_space > 0:
                    return ((total_space - total_free) / total_space) * 100
        except:
            pass
        return 0.0

# Usage
monitor = RadarrMonitor(radarr_client)

# Collect metrics
metrics = monitor.collect_metrics()
if metrics:
    print(f"Status: {metrics.status}")
    print(f"CPU: {metrics.cpu_usage:.1f}%, Memory: {metrics.memory_usage:.1f}%")

    # Check for alerts
    alerts = monitor.check_alerts(metrics)
    for alert in alerts:
        print(f"🚨 ALERT [{alert['severity']}]: {alert['message']}")

# Generate daily report
report = monitor.generate_report(hours=24)
print(f"\\n24-Hour Report:")
print(f"Average CPU: {report['averages']['cpu_usage']}%")
print(f"Peak Memory: {report['peaks']['memory_usage']}%")
print(f"Current Status: {report['current_status']}")
```

### 4. Custom Release Processing Pipeline

```python
from typing import List, Dict, Any, Optional
import re
from dataclasses import dataclass

@dataclass
class ReleaseCandidate:
    title: str
    size: int
    quality: str
    indexer: str
    seeders: int
    age_hours: int
    download_url: str
    info_hash: Optional[str] = None

class CustomReleaseProcessor:
    def __init__(self, radarr_client):
        self.radarr = radarr_client
        self.quality_preferences = {
            'Remux-1080p': 100,
            'Bluray-1080p': 90,
            'WEBDL-1080p': 80,
            'WEBRip-1080p': 75,
            'HDTV-1080p': 60,
            'Bluray-720p': 50,
            'WEBDL-720p': 45,
            'WEBRip-720p': 40
        }

        self.size_preferences = {
            'min_size_gb': 1.0,
            'max_size_gb': 15.0,
            'preferred_range': (4.0, 8.0)  # Sweet spot for 1080p movies
        }

        self.indexer_trust_scores = {
            'high_trust': ['PTP', 'BTN', 'HDB'],
            'medium_trust': ['IPT', 'TL', 'PHD'],
            'low_trust': ['RARBG', 'YTS', '1337x']
        }

    def process_movie_releases(self, movie_id: int, custom_rules: Optional[Dict] = None) -> Dict[str, Any]:
        """Process and rank releases for a specific movie"""
        try:
            # Get movie details
            movie = self.radarr.get_movie(movie_id)
            if not movie:
                return {'error': 'Movie not found'}

            # Search for releases
            releases = self.radarr.search_movie_releases(movie_id)
            if not releases:
                return {'message': 'No releases found', 'candidates': []}

            # Convert to our internal format
            candidates = [self._convert_to_candidate(release) for release in releases]

            # Apply filtering rules
            filtered_candidates = self._apply_filters(candidates, movie, custom_rules)

            # Score and rank candidates
            scored_candidates = self._score_candidates(filtered_candidates, movie)

            # Sort by score (highest first)
            sorted_candidates = sorted(scored_candidates,
                                     key=lambda x: x['score'],
                                     reverse=True)

            return {
                'movie': {
                    'id': movie['id'],
                    'title': movie['title'],
                    'year': movie['year']
                },
                'total_releases': len(releases),
                'filtered_releases': len(filtered_candidates),
                'candidates': sorted_candidates[:10],  # Top 10
                'recommended': sorted_candidates[0] if sorted_candidates else None
            }

        except Exception as e:
            return {'error': f'Processing failed: {str(e)}'}

    def _convert_to_candidate(self, release: dict) -> ReleaseCandidate:
        """Convert Radarr release to our candidate format"""
        return ReleaseCandidate(
            title=release.get('title', ''),
            size=release.get('size', 0),
            quality=release.get('quality', {}).get('quality', {}).get('name', 'Unknown'),
            indexer=release.get('indexer', 'Unknown'),
            seeders=release.get('seeders', 0),
            age_hours=release.get('ageHours', 0),
            download_url=release.get('downloadUrl', ''),
            info_hash=release.get('infoHash')
        )

    def _apply_filters(self, candidates: List[ReleaseCandidate],
                      movie: dict, custom_rules: Optional[Dict]) -> List[ReleaseCandidate]:
        """Apply filtering rules to candidates"""
        filtered = []

        for candidate in candidates:
            # Size filtering
            size_gb = candidate.size / (1024**3)
            if size_gb < self.size_preferences['min_size_gb']:
                continue
            if size_gb > self.size_preferences['max_size_gb']:
                continue

            # Quality filtering - only allow configured qualities
            if candidate.quality not in self.quality_preferences:
                continue

            # Age filtering - reject releases older than 30 days
            if candidate.age_hours > (30 * 24):
                continue

            # Seeders filtering - require at least 5 seeders
            if candidate.seeders < 5:
                continue

            # Custom title filtering
            if self._has_blacklisted_terms(candidate.title):
                continue

            # Apply custom rules if provided
            if custom_rules and not self._apply_custom_rules(candidate, movie, custom_rules):
                continue

            filtered.append(candidate)

        return filtered

    def _score_candidates(self, candidates: List[ReleaseCandidate],
                         movie: dict) -> List[Dict[str, Any]]:
        """Score candidates based on multiple criteria"""
        scored = []

        for candidate in candidates:
            score = 0
            score_breakdown = {}

            # Quality score (0-100)
            quality_score = self.quality_preferences.get(candidate.quality, 0)
            score += quality_score
            score_breakdown['quality'] = quality_score

            # Size score (0-30)
            size_gb = candidate.size / (1024**3)
            size_score = self._calculate_size_score(size_gb)
            score += size_score
            score_breakdown['size'] = size_score

            # Indexer trust score (0-20)
            trust_score = self._calculate_trust_score(candidate.indexer)
            score += trust_score
            score_breakdown['indexer_trust'] = trust_score

            # Seeders score (0-20)
            seeders_score = min(candidate.seeders * 2, 20)
            score += seeders_score
            score_breakdown['seeders'] = seeders_score

            # Age score (0-15) - newer is better
            age_score = max(15 - (candidate.age_hours / 24), 0)
            score += age_score
            score_breakdown['age'] = age_score

            # Title quality indicators (0-15)
            title_score = self._calculate_title_score(candidate.title)
            score += title_score
            score_breakdown['title_quality'] = title_score

            scored.append({
                'candidate': candidate,
                'score': round(score, 1),
                'score_breakdown': score_breakdown,
                'size_gb': round(size_gb, 1)
            })

        return scored

    def _calculate_size_score(self, size_gb: float) -> float:
        """Calculate score based on file size preferences"""
        min_pref, max_pref = self.size_preferences['preferred_range']

        if min_pref <= size_gb <= max_pref:
            return 30  # Perfect size range
        elif size_gb < min_pref:
            # Penalty for too small
            return 30 * (size_gb / min_pref)
        else:
            # Penalty for too large
            max_allowed = self.size_preferences['max_size_gb']
            penalty = (size_gb - max_pref) / (max_allowed - max_pref)
            return max(30 * (1 - penalty), 0)

    def _calculate_trust_score(self, indexer: str) -> float:
        """Calculate score based on indexer trustworthiness"""
        if indexer in self.indexer_trust_scores['high_trust']:
            return 20
        elif indexer in self.indexer_trust_scores['medium_trust']:
            return 15
        elif indexer in self.indexer_trust_scores['low_trust']:
            return 5
        else:
            return 10  # Unknown indexer

    def _calculate_title_score(self, title: str) -> float:
        """Calculate score based on title quality indicators"""
        score = 0
        title_lower = title.lower()

        # Positive indicators
        positive_terms = [
            'proper', 'repack', 'internal', 'criterion', 'directors.cut',
            'extended.cut', 'unrated', 'imax', 'atmos', 'dts-hd'
        ]

        for term in positive_terms:
            if term in title_lower:
                score += 2

        # Negative indicators
        negative_terms = [
            'cam', 'ts', 'tc', 'workprint', 'screener', 'dvdscr',
            'r5', 'hdrip', 'webrip.xvid', 'mp4'
        ]

        for term in negative_terms:
            if term in title_lower:
                score -= 5

        return max(min(score, 15), 0)  # Cap between 0 and 15

    def _has_blacklisted_terms(self, title: str) -> bool:
        """Check for blacklisted terms in release title"""
        blacklist = [
            r'\.CAM\.',
            r'\.TS\.',
            r'\.HDCAM\.',
            r'\.KORSUB\.',
            r'\.HC\.',
            r'\.DUBBED\.',
            r'\.SUBBED\.'
        ]

        for pattern in blacklist:
            if re.search(pattern, title, re.IGNORECASE):
                return True

        return False

    def _apply_custom_rules(self, candidate: ReleaseCandidate,
                           movie: dict, rules: dict) -> bool:
        """Apply custom user-defined rules"""
        # Example custom rules implementation
        if 'min_seeders' in rules:
            if candidate.seeders < rules['min_seeders']:
                return False

        if 'preferred_groups' in rules:
            release_group = self._extract_release_group(candidate.title)
            if release_group and release_group not in rules['preferred_groups']:
                return False

        if 'max_age_hours' in rules:
            if candidate.age_hours > rules['max_age_hours']:
                return False

        return True

    def _extract_release_group(self, title: str) -> Optional[str]:
        """Extract release group from title"""
        # Common patterns for release groups
        patterns = [
            r'-([A-Za-z0-9]+)$',  # Group at end after dash
            r'\[([A-Za-z0-9]+)\]',  # Group in brackets
            r'\{([A-Za-z0-9]+)\}'   # Group in braces
        ]

        for pattern in patterns:
            match = re.search(pattern, title)
            if match:
                return match.group(1)

        return None

# Usage examples
processor = CustomReleaseProcessor(radarr_client)

# Process releases for a specific movie
result = processor.process_movie_releases(
    movie_id=123,
    custom_rules={
        'min_seeders': 10,
        'preferred_groups': ['SPARKS', 'FGT', 'AMIABLE'],
        'max_age_hours': 168  # 1 week
    }
)

if result.get('recommended'):
    rec = result['recommended']
    print(f"Recommended release: {rec['candidate'].title}")
    print(f"Score: {rec['score']}")
    print(f"Size: {rec['size_gb']}GB")
    print(f"Quality: {rec['candidate'].quality}")

    # Automatically grab the best release
    grab_result = radarr_client.grab_release(rec['candidate'].download_url)
    print(f"Grab result: {grab_result}")
```

## 🔄 Integration Testing Framework

```python
import unittest
import requests_mock
from unittest.mock import Mock, patch

class RadarrIntegrationTestCase(unittest.TestCase):
    def setUp(self):
        self.base_url = 'http://localhost:7878'
        self.api_key = 'test-api-key'
        self.radarr = RadarrClient(self.base_url, self.api_key)

    @requests_mock.Mocker()
    def test_get_movies_integration(self, m):
        """Test movie retrieval with mocked responses"""
        # Mock the API response
        mock_movies = [
            {
                'id': 1,
                'title': 'Test Movie',
                'year': 2023,
                'tmdbId': 12345,
                'monitored': True,
                'hasFile': False
            }
        ]

        m.get(
            f'{self.base_url}/api/v3/movie',
            json=mock_movies,
            headers={'X-API-Key': self.api_key}
        )

        # Test the integration
        movies = self.radarr.get_movies()

        self.assertEqual(len(movies), 1)
        self.assertEqual(movies[0]['title'], 'Test Movie')

        # Verify the request was made correctly
        self.assertEqual(len(m.request_history), 1)
        request = m.request_history[0]
        self.assertEqual(request.headers['X-API-Key'], self.api_key)

    @requests_mock.Mocker()
    def test_error_handling(self, m):
        """Test error handling for API failures"""
        # Mock a 401 Unauthorized response
        m.get(
            f'{self.base_url}/api/v3/movie',
            status_code=401,
            json={'error': 'Invalid API key'}
        )

        with self.assertRaises(requests.exceptions.HTTPError):
            self.radarr.get_movies()

    def test_real_api_connectivity(self):
        """Integration test with real API (requires running instance)"""
        try:
            status = self.radarr.get_system_status()
            self.assertIn('version', status)
            self.assertIn('authentication', status)
            print(f"✅ Connected to Radarr Go v{status['version']}")
        except Exception as e:
            self.skipTest(f"Real API not available: {e}")

class BulkImportTestCase(unittest.TestCase):
    def setUp(self):
        self.mock_radarr = Mock()
        self.importer = BulkMovieImporter(self.mock_radarr)

    def test_duplicate_detection(self):
        """Test that duplicate movies are properly detected"""
        # Mock existing movies
        self.mock_radarr.get_movies.return_value = [
            {'tmdbId': 12345, 'title': 'Existing Movie'}
        ]

        # Mock search results
        self.mock_radarr.search_movies.return_value = [
            {'tmdbId': 12345, 'title': 'Existing Movie', 'year': 2023}
        ]

        # Test movies including a duplicate
        test_movies = [
            {'title': 'Existing Movie', 'tmdb_id': 12345, 'year': 2023},
            {'title': 'New Movie', 'tmdb_id': 67890, 'year': 2024}
        ]

        # Process movies
        results = self.importer._process_movie_batches(test_movies, 1, '/movies')

        # Should only attempt to add the new movie
        self.assertEqual(self.mock_radarr.search_movies.call_count, 1)

if __name__ == '__main__':
    # Run specific test suites
    suite = unittest.TestLoader().loadTestsFromTestCase(RadarrIntegrationTestCase)
    runner = unittest.TextTestRunner(verbosity=2)
    runner.run(suite)
```

## 📚 Best Practices and Recommendations

### 1. Error Handling and Resilience

```python
import time
import random
from functools import wraps

def retry_on_failure(max_retries=3, delay_base=1.0, exceptions=(Exception,)):
    """Decorator for retrying failed operations with exponential backoff"""
    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            for attempt in range(max_retries):
                try:
                    return func(*args, **kwargs)
                except exceptions as e:
                    if attempt == max_retries - 1:
                        raise e

                    delay = delay_base * (2 ** attempt) + random.uniform(0, 1)
                    print(f"Attempt {attempt + 1} failed: {e}. Retrying in {delay:.1f}s...")
                    time.sleep(delay)

            return None
        return wrapper
    return decorator

class ResilientRadarrClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.api_key = api_key
        self.session = requests.Session()
        self.session.headers.update({'X-API-Key': api_key})

        # Configure timeouts
        self.session.timeout = 30

        # Configure retries for the session
        from requests.adapters import HTTPAdapter
        from urllib3.util.retry import Retry

        retry_strategy = Retry(
            total=3,
            backoff_factor=1,
            status_forcelist=[429, 500, 502, 503, 504],
        )

        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    @retry_on_failure(max_retries=3, exceptions=(requests.exceptions.RequestException,))
    def get_movies(self, **params):
        response = self.session.get(f"{self.base_url}/api/v3/movie", params=params)
        response.raise_for_status()
        return response.json()
```

### 2. Rate Limiting and API Etiquette

```python
import threading
import time
from collections import deque

class RateLimitedRadarrClient:
    def __init__(self, base_url, api_key, requests_per_minute=60):
        self.base_url = base_url
        self.api_key = api_key
        self.requests_per_minute = requests_per_minute
        self.request_times = deque()
        self.lock = threading.Lock()
        self.session = requests.Session()
        self.session.headers.update({'X-API-Key': api_key})

    def _wait_for_rate_limit(self):
        """Implement rate limiting"""
        with self.lock:
            now = time.time()
            minute_ago = now - 60

            # Remove requests older than 1 minute
            while self.request_times and self.request_times[0] < minute_ago:
                self.request_times.popleft()

            # Check if we need to wait
            if len(self.request_times) >= self.requests_per_minute:
                sleep_time = 60 - (now - self.request_times[0])
                if sleep_time > 0:
                    print(f"Rate limit reached, waiting {sleep_time:.1f}s...")
                    time.sleep(sleep_time)

            # Record this request
            self.request_times.append(now)

    def _request(self, method, endpoint, **kwargs):
        self._wait_for_rate_limit()

        response = self.session.request(
            method,
            f"{self.base_url}/api/v3{endpoint}",
            **kwargs
        )
        response.raise_for_status()
        return response.json() if response.content else None

    def get_movies(self, **params):
        return self._request('GET', '/movie', params=params)

    def add_movie(self, movie_data):
        return self._request('POST', '/movie', json=movie_data)
```

### 3. Configuration Management

```python
import os
import yaml
from dataclasses import dataclass
from typing import Optional, List

@dataclass
class RadarrConfig:
    base_url: str
    api_key: str
    timeout: int = 30
    max_retries: int = 3
    requests_per_minute: int = 60
    default_quality_profile_id: int = 1
    default_root_folder: str = "/movies"
    preferred_indexers: List[str] = None

    @classmethod
    def from_file(cls, config_path: str):
        """Load configuration from YAML file"""
        with open(config_path, 'r') as f:
            data = yaml.safe_load(f)

        return cls(
            base_url=data['radarr']['base_url'],
            api_key=data['radarr']['api_key'],
            timeout=data.get('timeout', 30),
            max_retries=data.get('max_retries', 3),
            requests_per_minute=data.get('requests_per_minute', 60),
            default_quality_profile_id=data.get('default_quality_profile_id', 1),
            default_root_folder=data.get('default_root_folder', '/movies'),
            preferred_indexers=data.get('preferred_indexers', [])
        )

    @classmethod
    def from_env(cls):
        """Load configuration from environment variables"""
        return cls(
            base_url=os.getenv('RADARR_URL', 'http://localhost:7878'),
            api_key=os.getenv('RADARR_API_KEY'),
            timeout=int(os.getenv('RADARR_TIMEOUT', '30')),
            max_retries=int(os.getenv('RADARR_MAX_RETRIES', '3')),
            requests_per_minute=int(os.getenv('RADARR_REQUESTS_PER_MINUTE', '60')),
            default_quality_profile_id=int(os.getenv('RADARR_DEFAULT_QUALITY_PROFILE', '1')),
            default_root_folder=os.getenv('RADARR_DEFAULT_ROOT_FOLDER', '/movies')
        )

# Example config.yaml
config_yaml = """
radarr:
  base_url: "http://localhost:7878"
  api_key: "your-api-key-here"
  timeout: 30
  max_retries: 3
  requests_per_minute: 60
  default_quality_profile_id: 1
  default_root_folder: "/movies"
  preferred_indexers:
    - "PTP"
    - "BTN"
    - "IPT"
"""

# Usage
config = RadarrConfig.from_file('config.yaml')
# or
config = RadarrConfig.from_env()
```

### 4. Logging and Observability

```python
import logging
import json
import time
from functools import wraps

# Configure structured logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(levelname)s %(name)s %(message)s',
    handlers=[
        logging.StreamHandler(),
        logging.FileHandler('radarr_integration.log')
    ]
)

def log_api_calls(func):
    """Decorator to log API calls with timing"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.time()
        logger = logging.getLogger(func.__module__)

        try:
            result = func(*args, **kwargs)
            duration = time.time() - start_time

            logger.info("API call successful", extra={
                'function': func.__name__,
                'duration_ms': round(duration * 1000, 2),
                'args_count': len(args),
                'kwargs_keys': list(kwargs.keys())
            })

            return result

        except Exception as e:
            duration = time.time() - start_time

            logger.error("API call failed", extra={
                'function': func.__name__,
                'duration_ms': round(duration * 1000, 2),
                'error': str(e),
                'error_type': type(e).__name__
            })

            raise

    return wrapper

class ObservableRadarrClient:
    def __init__(self, base_url, api_key):
        self.base_url = base_url
        self.api_key = api_key
        self.logger = logging.getLogger(__name__)
        self.session = requests.Session()
        self.session.headers.update({'X-API-Key': api_key})

    @log_api_calls
    def get_movies(self, **params):
        self.logger.debug("Getting movies", extra={'filters': params})
        response = self.session.get(f"{self.base_url}/api/v3/movie", params=params)
        response.raise_for_status()

        movies = response.json()
        self.logger.info("Retrieved movies", extra={'count': len(movies)})
        return movies

    @log_api_calls
    def add_movie(self, movie_data):
        self.logger.info("Adding movie", extra={
            'title': movie_data.get('title'),
            'year': movie_data.get('year'),
            'tmdb_id': movie_data.get('tmdbId')
        })

        response = self.session.post(f"{self.base_url}/api/v3/movie", json=movie_data)
        response.raise_for_status()

        added_movie = response.json()
        self.logger.info("Movie added successfully", extra={
            'id': added_movie['id'],
            'title': added_movie['title']
        })

        return added_movie
```

## 📝 Summary

This comprehensive integration guide provides:

- **🏗️ Multiple Architecture Patterns** - From simple clients to event-driven systems
- **🔧 Real-world Examples** - Bulk import, monitoring, custom processing
- **✅ Best Practices** - Error handling, rate limiting, testing
- **📊 Production-ready Code** - Logging, configuration, resilience

### Next Steps

1. **Start Small**: Begin with a simple client library
2. **Add Resilience**: Implement error handling and retries
3. **Scale Up**: Add monitoring and observability
4. **Optimize**: Fine-tune performance for your use case

### Resources

- 📖 [Interactive API Documentation](/docs/swagger)
- 📄 [OpenAPI Specification](/docs/openapi.yaml)
- 🔄 [API Compatibility Guide](/docs/api-compatibility)
- 🏥 [Health Dashboard](/api/v3/health/dashboard)

Ready to build amazing integrations with Radarr Go! 🎬✨
