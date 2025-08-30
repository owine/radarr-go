# Radarr Go API Endpoint Catalog

**Version**: v0.9.0-alpha
**API Version**: v3 (Radarr Compatible)
**Total Endpoints**: 150+

## Overview

Radarr Go provides a comprehensive REST API that maintains 100% compatibility with Radarr v3 API. All endpoints support JSON request/response format with optional API key authentication.

## Authentication

All API endpoints (except `/ping`) require authentication via API key:

- **Header**: `X-API-Key: your-api-key-here`
- **Query Parameter**: `?apikey=your-api-key-here`

## Base URL Structure

All API endpoints are prefixed with `/api/v3/`:
```
http://localhost:7878/api/v3/{endpoint}
```

## Endpoint Categories

### System and Health (15 endpoints)

#### System Status
- `GET /system/status` - Get system information and status
- `POST /system/health` - Trigger health check
- `POST /system/cleanup` - Trigger system cleanup

#### Health Monitoring
- `GET /health` - Overall health status summary
- `GET /health/dashboard` - Complete health dashboard with all metrics
- `GET /health/check/{name}` - Run specific health check by name

#### Health Issues Management
- `GET /health/issue` - List health issues with filtering options
- `GET /health/issue/{id}` - Get specific health issue details
- `POST /health/issue/{id}/dismiss` - Dismiss a health issue
- `POST /health/issue/{id}/resolve` - Mark health issue as resolved

#### System Resources
- `GET /health/system/resources` - Current system resource usage
- `GET /health/system/diskspace` - Disk space information for all monitored paths

#### Performance Metrics
- `GET /health/metrics` - Performance metrics with configurable time ranges
- `POST /health/metrics/record` - Manually record performance metrics
- `POST /health/monitoring/start` - Start background health monitoring
- `POST /health/monitoring/stop` - Stop background health monitoring
- `POST /health/monitoring/cleanup` - Cleanup old health monitoring data

### Movie Management (25 endpoints)

#### Core Movie Operations
- `GET /movie` - List all movies with filtering, sorting, and pagination
- `GET /movie/{id}` - Get specific movie by ID
- `POST /movie` - Add new movie to library
- `PUT /movie/{id}` - Update existing movie
- `DELETE /movie/{id}` - Remove movie from library

#### Movie Discovery and Metadata
- `GET /movie/lookup` - Search for movies (general lookup)
- `GET /movie/lookup/tmdb` - Search movies by TMDB ID
- `GET /movie/popular` - Get popular movies from TMDB
- `GET /movie/trending` - Get trending movies from TMDB
- `PUT /movie/{id}/refresh` - Refresh metadata for specific movie
- `POST /movie/refresh` - Refresh metadata for all movies
- `POST /movie/{id}/refresh` - Trigger movie refresh task

#### Movie File Management
- `GET /moviefile` - List all movie files with metadata
- `GET /moviefile/{id}` - Get specific movie file
- `DELETE /moviefile/{id}` - Delete movie file

#### Movie Collections
- `GET /collection` - Get all movie collections
- `GET /collection/{id}` - Get specific collection details
- `POST /collection` - Create new movie collection
- `PUT /collection/{id}` - Update collection information
- `DELETE /collection/{id}` - Delete movie collection
- `POST /collection/{id}/search` - Search for missing movies in collection
- `POST /collection/{id}/sync` - Sync collection from TMDB
- `GET /collection/{id}/statistics` - Get collection statistics and completion

#### Movie Search Operations
- `GET /search/movie` - Search external sources for movies
- `GET /search/movie/{id}` - Search for specific movie releases
- `GET /search/interactive` - Interactive search with manual selection

### Quality and Release Management (20 endpoints)

#### Quality Profiles
- `GET /qualityprofile` - List all quality profiles
- `GET /qualityprofile/{id}` - Get specific quality profile
- `POST /qualityprofile` - Create new quality profile
- `PUT /qualityprofile/{id}` - Update quality profile
- `DELETE /qualityprofile/{id}` - Delete quality profile

#### Quality Definitions
- `GET /qualitydefinition` - List quality definitions
- `GET /qualitydefinition/{id}` - Get specific quality definition
- `PUT /qualitydefinition/{id}` - Update quality definition settings

#### Custom Formats
- `GET /customformat` - List all custom formats
- `GET /customformat/{id}` - Get specific custom format
- `POST /customformat` - Create new custom format
- `PUT /customformat/{id}` - Update custom format
- `DELETE /customformat/{id}` - Delete custom format

#### Release Management
- `GET /release` - List available releases with filtering
- `GET /release/{id}` - Get specific release details
- `DELETE /release/{id}` - Remove release from consideration
- `GET /release/stats` - Release statistics and metrics
- `POST /release/grab` - Grab/download specific release
- `GET /search` - Search for releases across all indexers
- `GET /search/movie/{id}` - Search releases for specific movie

### Download and Queue Management (15 endpoints)

#### Download Clients
- `GET /downloadclient` - List configured download clients
- `GET /downloadclient/{id}` - Get specific download client
- `POST /downloadclient` - Add new download client
- `PUT /downloadclient/{id}` - Update download client configuration
- `DELETE /downloadclient/{id}` - Remove download client
- `POST /downloadclient/test` - Test download client connection
- `GET /downloadclient/stats` - Download client statistics

#### Queue Management
- `GET /queue` - List current download queue
- `GET /queue/{id}` - Get specific queue item
- `DELETE /queue/{id}` - Remove item from queue
- `DELETE /queue/bulk` - Bulk remove queue items
- `GET /queue/stats` - Queue statistics and metrics

#### Download History
- `GET /downloadhistory` - Download history with filtering

### Import and List Management (15 endpoints)

#### Import Lists
- `GET /importlist` - List configured import lists
- `GET /importlist/{id}` - Get specific import list
- `POST /importlist` - Create new import list
- `PUT /importlist/{id}` - Update import list configuration
- `DELETE /importlist/{id}` - Delete import list
- `POST /importlist/test` - Test import list connection
- `POST /importlist/{id}/sync` - Sync specific import list
- `POST /importlist/sync` - Sync all import lists
- `GET /importlist/stats` - Import list statistics

#### Import List Movies
- `GET /importlistmovies` - Movies discovered from import lists

#### Manual Import
- `POST /import/process` - Process import operation
- `GET /import/manual` - Get manual import candidates
- `POST /import/manual` - Process manual import with selections

### Task and Command Management (20 endpoints)

#### Task Management (Command API)
- `GET /command` - List all tasks/commands with status
- `GET /command/{id}` - Get specific task details
- `POST /command` - Queue new task/command
- `DELETE /command/{id}` - Cancel running task

#### System Tasks
- `GET /system/task` - List scheduled system tasks
- `POST /system/task` - Create new scheduled task
- `PUT /system/task/{id}` - Update scheduled task
- `DELETE /system/task/{id}` - Delete scheduled task
- `GET /system/task/status` - Get overall queue status

#### File Organization Tasks
- `GET /fileorganization` - List file organization history
- `GET /fileorganization/{id}` - Get specific organization attempt
- `POST /fileorganization/retry` - Retry failed organization attempts
- `POST /fileorganization/scan` - Trigger directory scan

#### File Operations
- `GET /fileoperation` - List file operations in progress
- `GET /fileoperation/{id}` - Get specific file operation status
- `DELETE /fileoperation/{id}` - Cancel file operation
- `GET /fileoperation/summary` - File operation summary statistics

#### Media Info Operations
- `POST /mediainfo/extract` - Extract media information from file

### Indexer and Search Provider Management (10 endpoints)

#### Indexer Configuration
- `GET /indexer` - List configured indexers
- `GET /indexer/{id}` - Get specific indexer configuration
- `POST /indexer` - Add new indexer
- `PUT /indexer/{id}` - Update indexer configuration
- `DELETE /indexer/{id}` - Remove indexer
- `POST /indexer/{id}/test` - Test indexer connection and functionality

### Configuration Management (25 endpoints)

#### Host Configuration
- `GET /config/host` - Get host configuration settings
- `PUT /config/host` - Update host configuration

#### Naming Configuration
- `GET /config/naming` - Get file naming configuration
- `PUT /config/naming` - Update naming configuration
- `GET /config/naming/tokens` - Get available naming tokens
- `GET /config/naming/preview/{movieId}` - Preview naming for specific movie

#### Media Management Configuration
- `GET /config/mediamanagement` - Get media management settings
- `PUT /config/mediamanagement` - Update media management settings

#### Root Folder Management
- `GET /rootfolder` - List configured root folders
- `GET /rootfolder/{id}` - Get specific root folder details
- `POST /rootfolder` - Add new root folder
- `PUT /rootfolder/{id}` - Update root folder configuration
- `DELETE /rootfolder/{id}` - Remove root folder

#### Configuration Statistics
- `GET /config/stats` - Configuration statistics and health

### Notification Management (10 endpoints)

#### Notification Configuration
- `GET /notification` - List configured notification providers
- `GET /notification/{id}` - Get specific notification configuration
- `POST /notification` - Add new notification provider
- `PUT /notification/{id}` - Update notification configuration
- `DELETE /notification/{id}` - Remove notification provider
- `POST /notification/test` - Test notification delivery

#### Notification Schema and Providers
- `GET /notification/schema` - List available notification provider types
- `GET /notification/schema/{type}` - Get provider-specific configuration fields

#### Notification History
- `GET /notification/history` - Notification delivery history

### Calendar and Scheduling (10 endpoints)

#### Calendar Events
- `GET /calendar` - Get calendar events with date filtering
- `GET /calendar/feed.ics` - iCal feed for external calendar applications
- `GET /calendar/feed/url` - Generate shareable iCal feed URL
- `POST /calendar/refresh` - Force refresh calendar events
- `GET /calendar/stats` - Calendar event statistics

#### Calendar Configuration
- `GET /calendar/config` - Get calendar configuration settings
- `PUT /calendar/config` - Update calendar configuration

### Wanted Movies Management (15 endpoints)

#### Wanted Movies Tracking
- `GET /wanted/missing` - List missing movies
- `GET /wanted/cutoff` - List movies that haven't met quality cutoff
- `GET /wanted` - List all wanted movies with comprehensive filtering
- `GET /wanted/stats` - Wanted movies statistics and metrics
- `GET /wanted/{id}` - Get specific wanted movie details

#### Wanted Movies Operations
- `POST /wanted/search` - Trigger search for wanted movies
- `POST /wanted/bulk` - Perform bulk operations on wanted movies
- `POST /wanted/refresh` - Refresh wanted movies analysis
- `PUT /wanted/{id}/priority` - Update wanted movie search priority
- `DELETE /wanted/{id}` - Remove movie from wanted list

### History and Activity (10 endpoints)

#### History Management
- `GET /history` - List activity history with filtering
- `GET /history/{id}` - Get specific history record
- `DELETE /history/{id}` - Delete history record
- `GET /history/stats` - History statistics and metrics

#### Activity Tracking
- `GET /activity` - List current activities and tasks
- `GET /activity/{id}` - Get specific activity details
- `DELETE /activity/{id}` - Cancel/remove activity
- `GET /activity/running` - List currently running activities

### Parse and Rename Operations (10 endpoints)

#### Release Parsing
- `GET /parse` - Parse single release title/name
- `POST /parse` - Parse multiple release titles in batch
- `DELETE /parse/cache` - Clear release name parsing cache

#### File and Folder Renaming
- `GET /rename/preview` - Preview file rename operations
- `POST /rename` - Execute file rename operations
- `GET /rename/preview/folder` - Preview folder rename operations
- `POST /rename/folder` - Execute folder rename operations

## Response Formats

### Standard Success Response
```json
{
  "data": [...],
  "meta": {
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```

### Error Response
```json
{
  "error": "Resource not found",
  "code": "MOVIE_NOT_FOUND",
  "details": "Movie with ID 123 does not exist"
}
```

## Pagination

Most list endpoints support pagination via query parameters:
- `page` - Page number (default: 1)
- `pageSize` - Items per page (default: 20, max: 100)

## Filtering and Sorting

Many endpoints support advanced filtering and sorting:
- `sortBy` - Field to sort by
- `sortDirection` - `asc` or `desc`
- Various field-specific filters (see individual endpoint documentation)

## Rate Limiting

- Default: 100 requests per minute per API key
- Health check endpoint (`/ping`) is not rate limited
- Rate limit headers included in responses

## WebSocket Support

Real-time updates available via WebSocket connection for:
- Task/command status updates
- Health monitoring alerts
- Queue status changes
- File organization progress

Connect to: `ws://localhost:7878/ws` with API key authentication.
