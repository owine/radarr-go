# Radarr Go API Integration Guide

This comprehensive guide provides practical examples and integration patterns for working with the Radarr Go API. All examples include complete error handling and demonstrate real-world usage patterns.

## Table of Contents

1. [Client Examples](#client-examples)
2. [Common Integration Patterns](#common-integration-patterns)
3. [Automation Examples](#automation-examples)
4. [Troubleshooting Guide](#troubleshooting-guide)

## Overview

Radarr Go provides a complete REST API compatible with Radarr v3. The API includes:

- **150+ endpoints** for comprehensive movie management
- **Real-time WebSocket updates**
- **Multi-database support** (PostgreSQL/MariaDB)
- **High performance** Go implementation
- **Complete iCal calendar integration**

### Base Configuration

All API calls require authentication and use these common settings:

- **Base URL**: `http://your-server:7878/api/v3/`
- **Authentication**: `X-API-Key` header or `apikey` query parameter
- **Content Type**: `application/json`
- **Rate Limit**: 100 requests/minute by default

## Client Examples

### Python Client with Requests

Complete Python client implementation with error handling:

```python
import requests
import json
from typing import Dict, List, Optional
from datetime import datetime, timedelta
import time

class RadarrClient:
    """Complete Python client for Radarr Go API"""

    def __init__(self, base_url: str, api_key: str, timeout: int = 30):
        self.base_url = base_url.rstrip('/') + '/api/v3'
        self.api_key = api_key
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json',
            'User-Agent': 'Radarr-Python-Client/1.0'
        })

    def _request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make authenticated request with error handling"""
        url = f"{self.base_url}/{endpoint.lstrip('/')}"

        try:
            response = self.session.request(method, url, timeout=self.timeout, **kwargs)

            # Handle rate limiting
            if response.status_code == 429:
                retry_after = int(response.headers.get('Retry-After', 60))
                print(f"Rate limited. Waiting {retry_after} seconds...")
                time.sleep(retry_after)
                return self._request(method, endpoint, **kwargs)

            response.raise_for_status()
            return response

        except requests.exceptions.Timeout:
            raise Exception(f"Request timeout after {self.timeout} seconds")
        except requests.exceptions.ConnectionError:
            raise Exception("Failed to connect to Radarr server")
        except requests.exceptions.HTTPError as e:
            try:
                error_data = e.response.json()
                error_msg = error_data.get('error', str(e))
                raise Exception(f"API Error: {error_msg}")
            except:
                raise Exception(f"HTTP {e.response.status_code}: {e.response.text}")

    def get_system_status(self) -> Dict:
        """Get system status information"""
        response = self._request('GET', '/system/status')
        return response.json()

    def search_movies(self, query: str) -> List[Dict]:
        """Search for movies to add"""
        response = self._request('GET', '/movie/lookup', params={'term': query})
        return response.json()

    def get_movies(self, **filters) -> List[Dict]:
        """Get movies with optional filtering"""
        response = self._request('GET', '/movie', params=filters)
        data = response.json()
        return data.get('data', data)  # Handle paginated vs non-paginated responses

    def add_movie(self, tmdb_id: int, quality_profile_id: int,
                  root_folder: str, monitored: bool = True,
                  search_on_add: bool = True) -> Dict:
        """Add movie to collection"""
        # First, get movie details from TMDB
        search_results = self._request('GET', '/movie/lookup/tmdb',
                                     params={'tmdbId': tmdb_id}).json()

        movie_data = {
            'title': search_results['title'],
            'tmdbId': tmdb_id,
            'year': search_results['year'],
            'qualityProfileId': quality_profile_id,
            'rootFolderPath': root_folder,
            'monitored': monitored,
            'minimumAvailability': 'released',
            'addOptions': {
                'monitor': monitored,
                'searchForMovie': search_on_add
            }
        }

        response = self._request('POST', '/movie', json=movie_data)
        return response.json()

    def get_quality_profiles(self) -> List[Dict]:
        """Get all quality profiles"""
        response = self._request('GET', '/qualityprofile')
        return response.json()

    def get_missing_movies(self, page: int = 1, page_size: int = 50) -> Dict:
        """Get movies missing files"""
        params = {
            'page': page,
            'pageSize': page_size,
            'includeAvailable': True
        }
        response = self._request('GET', '/wanted/missing', params=params)
        return response.json()

    def get_calendar_events(self, start_date: datetime, end_date: datetime,
                           include_unmonitored: bool = False) -> List[Dict]:
        """Get calendar events for date range"""
        params = {
            'start': start_date.isoformat(),
            'end': end_date.isoformat(),
            'unmonitored': include_unmonitored
        }
        response = self._request('GET', '/calendar', params=params)
        return response.json()

    def get_health_status(self) -> Dict:
        """Get comprehensive health information"""
        response = self._request('GET', '/health/dashboard')
        return response.json()

# Example usage
def main():
    # Initialize client
    client = RadarrClient('http://localhost:7878', 'your-api-key-here')

    try:
        # Check system status
        status = client.get_system_status()
        print(f"Radarr {status['version']} running on {status['osName']}")

        # Search for a movie
        search_results = client.search_movies('inception')
        if search_results:
            movie = search_results[0]
            print(f"Found: {movie['title']} ({movie['year']})")

            # Get quality profiles
            profiles = client.get_quality_profiles()
            if profiles:
                # Add the movie
                added_movie = client.add_movie(
                    tmdb_id=movie['tmdbId'],
                    quality_profile_id=profiles[0]['id'],
                    root_folder='/movies'
                )
                print(f"Added movie: {added_movie['title']}")

        # Get missing movies
        missing = client.get_missing_movies()
        print(f"Missing movies: {missing['meta']['total']}")

    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    main()
```

### JavaScript/Node.js Client with Axios

```javascript
const axios = require('axios');
const WebSocket = require('ws');

class RadarrClient {
    constructor(baseUrl, apiKey, timeout = 30000) {
        this.baseUrl = baseUrl.replace(/\/$/, '') + '/api/v3';
        this.apiKey = apiKey;

        // Configure axios instance
        this.api = axios.create({
            baseURL: this.baseUrl,
            timeout: timeout,
            headers: {
                'X-API-Key': apiKey,
                'Content-Type': 'application/json',
                'User-Agent': 'Radarr-Node-Client/1.0'
            }
        });

        // Add response interceptor for error handling
        this.api.interceptors.response.use(
            response => response,
            async error => {
                if (error.response?.status === 429) {
                    const retryAfter = parseInt(error.response.headers['retry-after'] || '60');
                    console.log(`Rate limited. Waiting ${retryAfter} seconds...`);
                    await this.sleep(retryAfter * 1000);
                    return this.api.request(error.config);
                }

                const errorMessage = error.response?.data?.error || error.message;
                throw new Error(`API Error: ${errorMessage}`);
            }
        );
    }

    sleep(ms) {
        return new Promise(resolve => setTimeout(resolve, ms));
    }

    async getSystemStatus() {
        const response = await this.api.get('/system/status');
        return response.data;
    }

    async searchMovies(query) {
        const response = await this.api.get('/movie/lookup', {
            params: { term: query }
        });
        return response.data;
    }

    async getMovies(filters = {}) {
        const response = await this.api.get('/movie', { params: filters });
        return response.data.data || response.data;
    }

    async addMovie(tmdbId, qualityProfileId, rootFolder, options = {}) {
        // Get movie details from TMDB
        const movieResponse = await this.api.get('/movie/lookup/tmdb', {
            params: { tmdbId }
        });
        const movieData = movieResponse.data;

        const payload = {
            title: movieData.title,
            tmdbId: tmdbId,
            year: movieData.year,
            qualityProfileId: qualityProfileId,
            rootFolderPath: rootFolder,
            monitored: options.monitored !== false,
            minimumAvailability: options.minimumAvailability || 'released',
            addOptions: {
                monitor: options.monitored !== false,
                searchForMovie: options.searchOnAdd !== false,
                addMethod: 'manual'
            }
        };

        const response = await this.api.post('/movie', payload);
        return response.data;
    }

    async getQualityProfiles() {
        const response = await this.api.get('/qualityprofile');
        return response.data;
    }

    async getMissingMovies(page = 1, pageSize = 50) {
        const response = await this.api.get('/wanted/missing', {
            params: { page, pageSize, includeAvailable: true }
        });
        return response.data;
    }

    async getCalendarEvents(startDate, endDate, includeUnmonitored = false) {
        const response = await this.api.get('/calendar', {
            params: {
                start: startDate.toISOString(),
                end: endDate.toISOString(),
                unmonitored: includeUnmonitored
            }
        });
        return response.data;
    }

    async getHealthDashboard() {
        const response = await this.api.get('/health/dashboard');
        return response.data;
    }

    // WebSocket connection for real-time updates
    connectWebSocket() {
        const wsUrl = this.baseUrl.replace(/^http/, 'ws') + '/signalr';
        const ws = new WebSocket(wsUrl, {
            headers: { 'X-API-Key': this.apiKey }
        });

        ws.on('open', () => {
            console.log('WebSocket connected');
            // Subscribe to movie updates
            ws.send(JSON.stringify({
                protocol: 'json',
                version: 1
            }));
        });

        ws.on('message', (data) => {
            try {
                const message = JSON.parse(data.toString());
                this.handleWebSocketMessage(message);
            } catch (error) {
                console.error('Error parsing WebSocket message:', error);
            }
        });

        ws.on('error', (error) => {
            console.error('WebSocket error:', error);
        });

        return ws;
    }

    handleWebSocketMessage(message) {
        // Override this method to handle real-time updates
        console.log('WebSocket message:', message);
    }
}

// Example usage with real-time monitoring
async function main() {
    const client = new RadarrClient('http://localhost:7878', 'your-api-key-here');

    try {
        // System check
        const status = await client.getSystemStatus();
        console.log(`Connected to Radarr ${status.version}`);

        // Movie workflow
        const searchResults = await client.searchMovies('dune 2021');
        if (searchResults.length > 0) {
            const movie = searchResults[0];
            console.log(`Found: ${movie.title} (${movie.year})`);

            const profiles = await client.getQualityProfiles();
            const hdProfile = profiles.find(p => p.name.includes('HD'));

            if (hdProfile) {
                const addedMovie = await client.addMovie(
                    movie.tmdbId,
                    hdProfile.id,
                    '/movies',
                    { searchOnAdd: true }
                );
                console.log(`Added: ${addedMovie.title}`);
            }
        }

        // Real-time monitoring
        const ws = client.connectWebSocket();

        // Graceful shutdown
        process.on('SIGINT', () => {
            ws.close();
            process.exit(0);
        });

    } catch (error) {
        console.error('Error:', error.message);
    }
}

// Run if called directly
if (require.main === module) {
    main();
}

module.exports = RadarrClient;
```

### Shell Script Examples with Curl

```bash
#!/bin/bash
# Radarr Go API Shell Script Examples

# Configuration
RADARR_URL="http://localhost:7878"
API_KEY="your-api-key-here"
BASE_URL="${RADARR_URL}/api/v3"

# Common curl function with error handling
radarr_api() {
    local method="$1"
    local endpoint="$2"
    local data="$3"

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

    local response=$(curl "${curl_opts[@]}" "$BASE_URL/$endpoint")
    local http_code=$(echo "$response" | grep -o "HTTPSTATUS:[0-9]*" | cut -d: -f2)
    local body=$(echo "$response" | sed 's/HTTPSTATUS:[0-9]*$//')

    if [[ "$http_code" -ge 400 ]]; then
        echo "ERROR: HTTP $http_code" >&2
        echo "$body" | jq -r '.error // "Unknown error"' >&2
        return 1
    fi

    echo "$body"
}

# System status check
check_system() {
    echo "=== System Status ==="
    local status=$(radarr_api "GET" "system/status")

    if [[ $? -eq 0 ]]; then
        echo "Version: $(echo "$status" | jq -r '.version')"
        echo "Database: $(echo "$status" | jq -r '.databaseType')"
        echo "OS: $(echo "$status" | jq -r '.osName')"
        echo "Uptime: $(echo "$status" | jq -r '.uptime // "Unknown"')"
        return 0
    else
        echo "Failed to get system status"
        return 1
    fi
}

# Search for movies
search_movie() {
    local query="$1"
    if [[ -z "$query" ]]; then
        echo "Usage: search_movie <query>"
        return 1
    fi

    echo "=== Searching for: $query ==="
    local results=$(radarr_api "GET" "movie/lookup?term=$(printf '%s' "$query" | jq -sRr @uri)")

    if [[ $? -eq 0 ]]; then
        echo "$results" | jq -r '.[] | "\(.tmdbId): \(.title) (\(.year)) - \(.overview[0:100])..."'
    else
        echo "Search failed"
        return 1
    fi
}

# Add movie by TMDB ID
add_movie() {
    local tmdb_id="$1"
    local quality_profile_id="$2"
    local root_folder="$3"

    if [[ -z "$tmdb_id" || -z "$quality_profile_id" || -z "$root_folder" ]]; then
        echo "Usage: add_movie <tmdb_id> <quality_profile_id> <root_folder>"
        return 1
    fi

    echo "=== Adding movie TMDB:$tmdb_id ==="

    # Get movie details
    local movie_info=$(radarr_api "GET" "movie/lookup/tmdb?tmdbId=$tmdb_id")
    if [[ $? -ne 0 ]]; then
        echo "Failed to get movie information"
        return 1
    fi

    local title=$(echo "$movie_info" | jq -r '.title')
    local year=$(echo "$movie_info" | jq -r '.year')

    # Prepare add request
    local add_data=$(jq -n \
        --arg title "$title" \
        --argjson tmdbId "$tmdb_id" \
        --argjson year "$year" \
        --argjson qualityProfileId "$quality_profile_id" \
        --arg rootFolderPath "$root_folder" \
        '{
            title: $title,
            tmdbId: $tmdbId,
            year: $year,
            qualityProfileId: $qualityProfileId,
            rootFolderPath: $rootFolderPath,
            monitored: true,
            minimumAvailability: "released",
            addOptions: {
                monitor: true,
                searchForMovie: true,
                addMethod: "manual"
            }
        }')

    local result=$(radarr_api "POST" "movie" "$add_data")
    if [[ $? -eq 0 ]]; then
        echo "Successfully added: $title ($year)"
        echo "Movie ID: $(echo "$result" | jq -r '.id')"
    else
        echo "Failed to add movie"
        return 1
    fi
}

# Get missing movies
get_missing_movies() {
    local page="${1:-1}"
    local page_size="${2:-20}"

    echo "=== Missing Movies (Page $page) ==="
    local missing=$(radarr_api "GET" "wanted/missing?page=$page&pageSize=$page_size")

    if [[ $? -eq 0 ]]; then
        local total=$(echo "$missing" | jq -r '.meta.total // 0')
        echo "Total missing: $total"
        echo
        echo "$missing" | jq -r '.data[]? | "\(.title) (\(.year)) - Available: \(.isAvailable)"'
    else
        echo "Failed to get missing movies"
        return 1
    fi
}

# Get calendar events
get_calendar() {
    local days_ahead="${1:-30}"
    local start_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    local end_date=$(date -u -d "+$days_ahead days" +"%Y-%m-%dT%H:%M:%SZ")

    echo "=== Calendar Events (Next $days_ahead days) ==="
    local events=$(radarr_api "GET" "calendar?start=$start_date&end=$end_date")

    if [[ $? -eq 0 ]]; then
        echo "$events" | jq -r '.[] | "\(.start[0:10]): \(.title) - \(.eventType)"'
    else
        echo "Failed to get calendar events"
        return 1
    fi
}

# Health monitoring
check_health() {
    echo "=== Health Dashboard ==="
    local health=$(radarr_api "GET" "health/dashboard")

    if [[ $? -eq 0 ]]; then
        local status=$(echo "$health" | jq -r '.status.status')
        local issues=$(echo "$health" | jq -r '.status.issues | length')

        echo "Overall Status: $status"
        echo "Issues Found: $issues"

        if [[ "$issues" -gt 0 ]]; then
            echo "Issues:"
            echo "$health" | jq -r '.status.issues[] | "  - \(.type): \(.message)"'
        fi

        # System resources
        local resources=$(echo "$health" | jq -r '.systemResources')
        echo "CPU Usage: $(echo "$resources" | jq -r '.cpuUsage')%"
        echo "Memory: $(echo "$resources" | jq -r '.memoryUsage / 1024 / 1024 | floor')MB / $(echo "$resources" | jq -r '.memoryTotal / 1024 / 1024 | floor')MB"
    else
        echo "Failed to get health status"
        return 1
    fi
}

# Bulk operations
bulk_add_from_list() {
    local movie_list="$1"
    local quality_profile_id="$2"
    local root_folder="$3"

    if [[ ! -f "$movie_list" ]]; then
        echo "Usage: bulk_add_from_list <movie_list_file> <quality_profile_id> <root_folder>"
        echo "Movie list file should contain one TMDB ID per line"
        return 1
    fi

    echo "=== Bulk Adding Movies ==="
    local success_count=0
    local error_count=0

    while IFS= read -r tmdb_id; do
        # Skip empty lines and comments
        [[ -z "$tmdb_id" || "$tmdb_id" =~ ^#.* ]] && continue

        echo "Adding TMDB:$tmdb_id..."
        if add_movie "$tmdb_id" "$quality_profile_id" "$root_folder" >/dev/null 2>&1; then
            ((success_count++))
            echo "  ✓ Success"
        else
            ((error_count++))
            echo "  ✗ Failed"
        fi

        # Rate limiting - wait 1 second between requests
        sleep 1
    done < "$movie_list"

    echo "Bulk add complete: $success_count successful, $error_count failed"
}

# Quality profiles management
list_quality_profiles() {
    echo "=== Quality Profiles ==="
    local profiles=$(radarr_api "GET" "qualityprofile")

    if [[ $? -eq 0 ]]; then
        echo "$profiles" | jq -r '.[] | "ID: \(.id) - \(.name) (Cutoff: \(.cutoff))"'
    else
        echo "Failed to get quality profiles"
        return 1
    fi
}

# Main script logic
case "${1:-help}" in
    "status")
        check_system
        ;;
    "search")
        search_movie "$2"
        ;;
    "add")
        add_movie "$2" "$3" "$4"
        ;;
    "missing")
        get_missing_movies "$2" "$3"
        ;;
    "calendar")
        get_calendar "$2"
        ;;
    "health")
        check_health
        ;;
    "bulk-add")
        bulk_add_from_list "$2" "$3" "$4"
        ;;
    "profiles")
        list_quality_profiles
        ;;
    "help"|*)
        echo "Radarr Go API Helper Script"
        echo
        echo "Usage: $0 <command> [arguments]"
        echo
        echo "Commands:"
        echo "  status                           - Get system status"
        echo "  search <query>                   - Search for movies"
        echo "  add <tmdb_id> <profile_id> <path> - Add movie"
        echo "  missing [page] [size]            - Get missing movies"
        echo "  calendar [days_ahead]            - Get calendar events"
        echo "  health                           - Check system health"
        echo "  bulk-add <file> <profile_id> <path> - Bulk add from file"
        echo "  profiles                         - List quality profiles"
        echo "  help                             - Show this help"
        ;;
esac
```

### PowerShell Examples for Windows

```powershell
# Radarr Go API PowerShell Module
# Save as RadarrGoAPI.psm1

param(
    [Parameter(Mandatory=$true)]
    [string]$BaseUrl,

    [Parameter(Mandatory=$true)]
    [string]$ApiKey
)

# Module variables
$Script:RadarrUrl = $BaseUrl.TrimEnd('/') + '/api/v3'
$Script:Headers = @{
    'X-API-Key' = $ApiKey
    'Content-Type' = 'application/json'
    'User-Agent' = 'Radarr-PowerShell-Client/1.0'
}

function Invoke-RadarrAPI {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$Endpoint,

        [Parameter(Mandatory=$false)]
        [Microsoft.PowerShell.Commands.WebRequestMethod]$Method = 'Get',

        [Parameter(Mandatory=$false)]
        [hashtable]$Body,

        [Parameter(Mandatory=$false)]
        [hashtable]$Query,

        [Parameter(Mandatory=$false)]
        [int]$TimeoutSec = 30
    )

    try {
        $Uri = "$Script:RadarrUrl/$($Endpoint.TrimStart('/'))"

        $RequestParams = @{
            Uri = $Uri
            Method = $Method
            Headers = $Script:Headers
            TimeoutSec = $TimeoutSec
        }

        if ($Query) {
            $QueryString = ($Query.GetEnumerator() | ForEach-Object {
                "$($_.Key)=$([System.Web.HttpUtility]::UrlEncode($_.Value))"
            }) -join '&'
            $RequestParams.Uri += "?$QueryString"
        }

        if ($Body) {
            $RequestParams.Body = $Body | ConvertTo-Json -Depth 10
        }

        $Response = Invoke-RestMethod @RequestParams
        return $Response

    } catch [System.Net.WebException] {
        $StatusCode = [int]$_.Exception.Response.StatusCode

        if ($StatusCode -eq 429) {
            $RetryAfter = $_.Exception.Response.Headers['Retry-After']
            if ($RetryAfter) {
                Write-Warning "Rate limited. Waiting $RetryAfter seconds..."
                Start-Sleep -Seconds ([int]$RetryAfter)
                return Invoke-RadarrAPI @PSBoundParameters
            }
        }

        $ErrorMessage = "HTTP $StatusCode"
        try {
            $ErrorResponse = $_.Exception.Response.GetResponseStream()
            $Reader = New-Object System.IO.StreamReader($ErrorResponse)
            $ErrorContent = $Reader.ReadToEnd() | ConvertFrom-Json
            $ErrorMessage = $ErrorContent.error
        } catch {
            # Use the original error if we can't parse the response
        }

        throw "API Error: $ErrorMessage"

    } catch {
        throw "Request failed: $($_.Exception.Message)"
    }
}

function Get-RadarrSystemStatus {
    <#
    .SYNOPSIS
    Gets Radarr system status information
    #>
    return Invoke-RadarrAPI -Endpoint 'system/status'
}

function Search-RadarrMovies {
    <#
    .SYNOPSIS
    Search for movies to add to collection
    .PARAMETER Query
    Search term (title, TMDB ID, or IMDB ID)
    #>
    param(
        [Parameter(Mandatory=$true)]
        [string]$Query
    )

    return Invoke-RadarrAPI -Endpoint 'movie/lookup' -Query @{ term = $Query }
}

function Get-RadarrMovies {
    <#
    .SYNOPSIS
    Get movies from collection with optional filtering
    #>
    param(
        [Parameter(Mandatory=$false)]
        [hashtable]$Filters = @{}
    )

    $Response = Invoke-RadarrAPI -Endpoint 'movie' -Query $Filters

    # Handle paginated response
    if ($Response.data) {
        return $Response.data
    } else {
        return $Response
    }
}

function Add-RadarrMovie {
    <#
    .SYNOPSIS
    Add movie to collection
    .PARAMETER TmdbId
    TMDB ID of the movie
    .PARAMETER QualityProfileId
    Quality profile ID to use
    .PARAMETER RootFolderPath
    Root folder path for the movie
    .PARAMETER Monitored
    Whether to monitor the movie
    .PARAMETER SearchOnAdd
    Whether to search for the movie after adding
    #>
    param(
        [Parameter(Mandatory=$true)]
        [int]$TmdbId,

        [Parameter(Mandatory=$true)]
        [int]$QualityProfileId,

        [Parameter(Mandatory=$true)]
        [string]$RootFolderPath,

        [Parameter(Mandatory=$false)]
        [bool]$Monitored = $true,

        [Parameter(Mandatory=$false)]
        [bool]$SearchOnAdd = $true
    )

    # Get movie details from TMDB
    $MovieInfo = Invoke-RadarrAPI -Endpoint 'movie/lookup/tmdb' -Query @{ tmdbId = $TmdbId }

    $MovieData = @{
        title = $MovieInfo.title
        tmdbId = $TmdbId
        year = $MovieInfo.year
        qualityProfileId = $QualityProfileId
        rootFolderPath = $RootFolderPath
        monitored = $Monitored
        minimumAvailability = 'released'
        addOptions = @{
            monitor = $Monitored
            searchForMovie = $SearchOnAdd
            addMethod = 'manual'
        }
    }

    return Invoke-RadarrAPI -Endpoint 'movie' -Method Post -Body $MovieData
}

function Get-RadarrQualityProfiles {
    <#
    .SYNOPSIS
    Get all quality profiles
    #>
    return Invoke-RadarrAPI -Endpoint 'qualityprofile'
}

function Get-RadarrMissingMovies {
    <#
    .SYNOPSIS
    Get movies that are missing files
    #>
    param(
        [Parameter(Mandatory=$false)]
        [int]$Page = 1,

        [Parameter(Mandatory=$false)]
        [int]$PageSize = 50
    )

    $Query = @{
        page = $Page
        pageSize = $PageSize
        includeAvailable = $true
    }

    return Invoke-RadarrAPI -Endpoint 'wanted/missing' -Query $Query
}

function Get-RadarrCalendarEvents {
    <#
    .SYNOPSIS
    Get calendar events for a date range
    #>
    param(
        [Parameter(Mandatory=$false)]
        [DateTime]$StartDate = (Get-Date),

        [Parameter(Mandatory=$false)]
        [DateTime]$EndDate = (Get-Date).AddDays(30),

        [Parameter(Mandatory=$false)]
        [bool]$IncludeUnmonitored = $false
    )

    $Query = @{
        start = $StartDate.ToString('yyyy-MM-ddTHH:mm:ssZ')
        end = $EndDate.ToString('yyyy-MM-ddTHH:mm:ssZ')
        unmonitored = $IncludeUnmonitored
    }

    return Invoke-RadarrAPI -Endpoint 'calendar' -Query $Query
}

function Get-RadarrHealthDashboard {
    <#
    .SYNOPSIS
    Get comprehensive health dashboard
    #>
    return Invoke-RadarrAPI -Endpoint 'health/dashboard'
}

function Import-RadarrMovieList {
    <#
    .SYNOPSIS
    Bulk import movies from a CSV file
    .PARAMETER Path
    Path to CSV file with columns: TmdbId, Title, Year
    .PARAMETER QualityProfileId
    Quality profile ID to use for all movies
    .PARAMETER RootFolderPath
    Root folder path for all movies
    #>
    param(
        [Parameter(Mandatory=$true)]
        [string]$Path,

        [Parameter(Mandatory=$true)]
        [int]$QualityProfileId,

        [Parameter(Mandatory=$true)]
        [string]$RootFolderPath
    )

    if (-not (Test-Path $Path)) {
        throw "File not found: $Path"
    }

    $Movies = Import-Csv $Path
    $Results = @()

    foreach ($Movie in $Movies) {
        Write-Progress -Activity "Adding Movies" -Status "Processing $($Movie.Title)" -PercentComplete (($Results.Count / $Movies.Count) * 100)

        try {
            $Result = Add-RadarrMovie -TmdbId $Movie.TmdbId -QualityProfileId $QualityProfileId -RootFolderPath $RootFolderPath
            $Results += [PSCustomObject]@{
                Title = $Movie.Title
                TmdbId = $Movie.TmdbId
                Status = 'Success'
                MovieId = $Result.id
                Error = $null
            }
            Write-Host "✓ Added: $($Movie.Title)" -ForegroundColor Green

        } catch {
            $Results += [PSCustomObject]@{
                Title = $Movie.Title
                TmdbId = $Movie.TmdbId
                Status = 'Failed'
                MovieId = $null
                Error = $_.Exception.Message
            }
            Write-Host "✗ Failed: $($Movie.Title) - $($_.Exception.Message)" -ForegroundColor Red
        }

        # Rate limiting
        Start-Sleep -Milliseconds 500
    }

    Write-Progress -Activity "Adding Movies" -Completed
    return $Results
}

# Example usage functions
function Show-RadarrStatus {
    <#
    .SYNOPSIS
    Display formatted system status
    #>
    $Status = Get-RadarrSystemStatus

    Write-Host "=== Radarr Go System Status ===" -ForegroundColor Cyan
    Write-Host "Version: $($Status.version)" -ForegroundColor White
    Write-Host "Database: $($Status.databaseType)" -ForegroundColor White
    Write-Host "OS: $($Status.osName)" -ForegroundColor White
    Write-Host "Authentication: $($Status.authentication)" -ForegroundColor White
}

function Show-RadarrHealth {
    <#
    .SYNOPSIS
    Display formatted health information
    #>
    $Health = Get-RadarrHealthDashboard

    Write-Host "=== Health Dashboard ===" -ForegroundColor Cyan

    $StatusColor = switch ($Health.status.status) {
        'healthy' { 'Green' }
        'warning' { 'Yellow' }
        'error' { 'Red' }
        default { 'White' }
    }

    Write-Host "Status: $($Health.status.status)" -ForegroundColor $StatusColor
    Write-Host "Issues: $($Health.status.issues.Count)" -ForegroundColor White

    if ($Health.status.issues.Count -gt 0) {
        Write-Host "`nIssues:" -ForegroundColor Yellow
        foreach ($Issue in $Health.status.issues) {
            $IssueColor = switch ($Issue.type) {
                'error' { 'Red' }
                'warning' { 'Yellow' }
                default { 'White' }
            }
            Write-Host "  $($Issue.type.ToUpper()): $($Issue.message)" -ForegroundColor $IssueColor
        }
    }

    # System resources
    if ($Health.systemResources) {
        Write-Host "`nSystem Resources:" -ForegroundColor Cyan
        Write-Host "  CPU Usage: $($Health.systemResources.cpuUsage)%" -ForegroundColor White
        $MemUsageMB = [math]::Round($Health.systemResources.memoryUsage / 1MB, 2)
        $MemTotalMB = [math]::Round($Health.systemResources.memoryTotal / 1MB, 2)
        Write-Host "  Memory: ${MemUsageMB}MB / ${MemTotalMB}MB" -ForegroundColor White
    }
}

# Export functions
Export-ModuleMember -Function @(
    'Invoke-RadarrAPI',
    'Get-RadarrSystemStatus',
    'Search-RadarrMovies',
    'Get-RadarrMovies',
    'Add-RadarrMovie',
    'Get-RadarrQualityProfiles',
    'Get-RadarrMissingMovies',
    'Get-RadarrCalendarEvents',
    'Get-RadarrHealthDashboard',
    'Import-RadarrMovieList',
    'Show-RadarrStatus',
    'Show-RadarrHealth'
)

# Example usage script (save separately as RadarrExample.ps1)
<#
# Import the module
Import-Module ./RadarrGoAPI.psm1 -ArgumentList 'http://localhost:7878', 'your-api-key-here'

# Check system status
Show-RadarrStatus

# Search for movies
$SearchResults = Search-RadarrMovies -Query "dune 2021"
$SearchResults | Select-Object title, year, tmdbId | Format-Table

# Get quality profiles
$Profiles = Get-RadarrQualityProfiles
$Profiles | Select-Object id, name, cutoff | Format-Table

# Add a movie
if ($SearchResults.Count -gt 0) {
    $Movie = $SearchResults[0]
    $Profile = $Profiles | Where-Object { $_.name -like "*HD*" } | Select-Object -First 1

    if ($Profile) {
        Add-RadarrMovie -TmdbId $Movie.tmdbId -QualityProfileId $Profile.id -RootFolderPath "C:\Movies"
    }
}

# Check health
Show-RadarrHealth

# Get missing movies
$Missing = Get-RadarrMissingMovies
Write-Host "Missing movies: $($Missing.meta.total)"
$Missing.data | Select-Object title, year, isAvailable | Format-Table
#>
```

### Go Client Library Example

```go
// radarr-go-client/client.go
package radarr

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "strconv"
    "time"
)

// Client represents a Radarr Go API client
type Client struct {
    BaseURL    *url.URL
    APIKey     string
    HTTPClient *http.Client
    UserAgent  string
}

// NewClient creates a new Radarr API client
func NewClient(baseURL, apiKey string) (*Client, error) {
    u, err := url.Parse(baseURL)
    if err != nil {
        return nil, fmt.Errorf("invalid base URL: %w", err)
    }

    return &Client{
        BaseURL: u,
        APIKey:  apiKey,
        HTTPClient: &http.Client{
            Timeout: 30 * time.Second,
        },
        UserAgent: "Radarr-Go-Client/1.0",
    }, nil
}

// doRequest executes an HTTP request with proper authentication and error handling
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
    u := *c.BaseURL
    u.Path = fmt.Sprintf("/api/v3/%s", endpoint)

    var reqBody *bytes.Buffer
    if body != nil {
        jsonData, err := json.Marshal(body)
        if err != nil {
            return fmt.Errorf("failed to marshal request body: %w", err)
        }
        reqBody = bytes.NewBuffer(jsonData)
    }

    req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("X-API-Key", c.APIKey)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("User-Agent", c.UserAgent)

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    // Handle rate limiting
    if resp.StatusCode == http.StatusTooManyRequests {
        retryAfter := resp.Header.Get("Retry-After")
        if retryAfter != "" {
            if seconds, err := strconv.Atoi(retryAfter); err == nil {
                select {
                case <-time.After(time.Duration(seconds) * time.Second):
                    return c.doRequest(ctx, method, endpoint, body, result)
                case <-ctx.Done():
                    return ctx.Err()
                }
            }
        }
    }

    if resp.StatusCode >= 400 {
        var apiErr APIError
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil {
            return &apiErr
        }
        return fmt.Errorf("API error: HTTP %d", resp.StatusCode)
    }

    if result != nil {
        if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
            return fmt.Errorf("failed to decode response: %w", err)
        }
    }

    return nil
}

// API Models
type SystemStatus struct {
    Version         string `json:"version"`
    BuildTime       string `json:"buildTime"`
    IsProduction    bool   `json:"isProduction"`
    OSName          string `json:"osName"`
    OSVersion       string `json:"osVersion"`
    DatabaseType    string `json:"databaseType"`
    DatabaseVersion string `json:"databaseVersion"`
    Authentication  string `json:"authentication"`
    RuntimeVersion  string `json:"runtimeVersion"`
}

type Movie struct {
    ID                  int                  `json:"id"`
    Title               string               `json:"title"`
    OriginalTitle       string               `json:"originalTitle,omitempty"`
    Year                int                  `json:"year"`
    TmdbID              int                  `json:"tmdbId"`
    ImdbID              string               `json:"imdbId,omitempty"`
    TitleSlug           string               `json:"titleSlug"`
    Status              string               `json:"status"`
    Overview            string               `json:"overview,omitempty"`
    InCinemas           *time.Time           `json:"inCinemas,omitempty"`
    PhysicalRelease     *time.Time           `json:"physicalRelease,omitempty"`
    DigitalRelease      *time.Time           `json:"digitalRelease,omitempty"`
    Images              []MediaImage         `json:"images,omitempty"`
    Path                string               `json:"path,omitempty"`
    RootFolderPath      string               `json:"rootFolderPath"`
    QualityProfileID    int                  `json:"qualityProfileId"`
    HasFile             bool                 `json:"hasFile"`
    MovieFileID         *int                 `json:"movieFileId,omitempty"`
    Monitored           bool                 `json:"monitored"`
    MinimumAvailability string               `json:"minimumAvailability"`
    IsAvailable         bool                 `json:"isAvailable"`
    Runtime             int                  `json:"runtime,omitempty"`
    Certification       string               `json:"certification,omitempty"`
    Genres              []string             `json:"genres,omitempty"`
    Tags                []int                `json:"tags,omitempty"`
    Added               time.Time            `json:"added"`
    CreatedAt           time.Time            `json:"createdAt"`
    UpdatedAt           time.Time            `json:"updatedAt"`
}

type MovieCreateRequest struct {
    Title               string     `json:"title"`
    TmdbID              int        `json:"tmdbId"`
    Year                int        `json:"year"`
    QualityProfileID    int        `json:"qualityProfileId"`
    RootFolderPath      string     `json:"rootFolderPath"`
    Monitored           bool       `json:"monitored"`
    MinimumAvailability string     `json:"minimumAvailability"`
    Tags                []int      `json:"tags,omitempty"`
    AddOptions          AddOptions `json:"addOptions"`
}

type AddOptions struct {
    Monitor        bool   `json:"monitor"`
    SearchForMovie bool   `json:"searchForMovie"`
    AddMethod      string `json:"addMethod"`
}

type MovieSearchResult struct {
    Title           string       `json:"title"`
    OriginalTitle   string       `json:"originalTitle,omitempty"`
    Year            int          `json:"year"`
    TmdbID          int          `json:"tmdbId"`
    ImdbID          string       `json:"imdbId,omitempty"`
    TitleSlug       string       `json:"titleSlug"`
    Overview        string       `json:"overview,omitempty"`
    Images          []MediaImage `json:"images,omitempty"`
    Genres          []string     `json:"genres,omitempty"`
    InCinemas       *time.Time   `json:"inCinemas,omitempty"`
    PhysicalRelease *time.Time   `json:"physicalRelease,omitempty"`
    DigitalRelease  *time.Time   `json:"digitalRelease,omitempty"`
    Runtime         int          `json:"runtime,omitempty"`
    Certification   string       `json:"certification,omitempty"`
    Studio          string       `json:"studio,omitempty"`
}

type MediaImage struct {
    CoverType string `json:"coverType"`
    URL       string `json:"url"`
    RemoteURL string `json:"remoteUrl,omitempty"`
}

type QualityProfile struct {
    ID             int                  `json:"id"`
    Name           string               `json:"name"`
    Cutoff         int                  `json:"cutoff"`
    Items          []QualityProfileItem `json:"items"`
    Language       string               `json:"language"`
    UpgradeAllowed bool                 `json:"upgradeAllowed"`
}

type QualityProfileItem struct {
    ID      int      `json:"id"`
    Name    string   `json:"name"`
    Allowed bool     `json:"allowed"`
    Quality *Quality `json:"quality,omitempty"`
}

type Quality struct {
    ID         int    `json:"id"`
    Name       string `json:"name"`
    Source     string `json:"source"`
    Resolution string `json:"resolution"`
}

type MovieList struct {
    Data []Movie        `json:"data"`
    Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
    Total      int `json:"total"`
    Page       int `json:"page"`
    PageSize   int `json:"pageSize"`
    TotalPages int `json:"totalPages"`
}

type CalendarEvent struct {
    ID        string    `json:"id"`
    MovieID   int       `json:"movieId"`
    Title     string    `json:"title"`
    Start     time.Time `json:"start"`
    End       time.Time `json:"end"`
    EventType string    `json:"eventType"`
    Movie     *Movie    `json:"movie,omitempty"`
}

type HealthDashboard struct {
    Status          HealthStatus        `json:"status"`
    SystemResources *SystemResources    `json:"systemResources,omitempty"`
    DiskSpace       []DiskSpace         `json:"diskSpace,omitempty"`
    Services        []ServiceStatus     `json:"services,omitempty"`
    Metrics         *PerformanceMetrics `json:"metrics,omitempty"`
}

type HealthStatus struct {
    Status    string        `json:"status"`
    Version   string        `json:"version"`
    Uptime    int64         `json:"uptime"`
    Timestamp time.Time     `json:"timestamp"`
    Issues    []HealthIssue `json:"issues"`
    Warnings  int           `json:"warnings"`
    Errors    int           `json:"errors"`
}

type HealthIssue struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`
    Source    string    `json:"source"`
    Message   string    `json:"message"`
    Details   string    `json:"details,omitempty"`
    WikiURL   string    `json:"wikiUrl,omitempty"`
    Timestamp time.Time `json:"timestamp"`
}

type SystemResources struct {
    CPUUsage    float64   `json:"cpuUsage"`
    MemoryUsage int64     `json:"memoryUsage"`
    MemoryTotal int64     `json:"memoryTotal"`
    DiskSpace   int64     `json:"diskSpace"`
    DiskTotal   int64     `json:"diskTotal"`
    Goroutines  int       `json:"goroutines"`
    Timestamp   time.Time `json:"timestamp"`
}

type DiskSpace struct {
    Path        string  `json:"path"`
    Label       string  `json:"label"`
    FreeSpace   int64   `json:"freeSpace"`
    TotalSpace  int64   `json:"totalSpace"`
    PercentUsed float64 `json:"percentUsed"`
}

type ServiceStatus struct {
    Name      string    `json:"name"`
    Status    string    `json:"status"`
    Message   string    `json:"message,omitempty"`
    LastCheck time.Time `json:"lastCheck"`
}

type PerformanceMetrics struct {
    ResponseTime struct {
        Min float64 `json:"min"`
        Max float64 `json:"max"`
        Avg float64 `json:"avg"`
    } `json:"responseTime"`
    Throughput struct {
        RequestsPerSecond float64 `json:"requestsPerSecond"`
    } `json:"throughput"`
    ErrorRate           float64 `json:"errorRate"`
    DatabaseConnections struct {
        Active int `json:"active"`
        Idle   int `json:"idle"`
        Total  int `json:"total"`
    } `json:"databaseConnections"`
}

type APIError struct {
    Error   string `json:"error"`
    Code    string `json:"code"`
    Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    if e.Details != "" {
        return fmt.Sprintf("%s: %s (%s)", e.Code, e.Error, e.Details)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Error)
}

// API Methods

// GetSystemStatus retrieves system status information
func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
    var status SystemStatus
    err := c.doRequest(ctx, "GET", "system/status", nil, &status)
    return &status, err
}

// SearchMovies searches for movies to add
func (c *Client) SearchMovies(ctx context.Context, query string) ([]MovieSearchResult, error) {
    endpoint := fmt.Sprintf("movie/lookup?term=%s", url.QueryEscape(query))
    var results []MovieSearchResult
    err := c.doRequest(ctx, "GET", endpoint, nil, &results)
    return results, err
}

// GetMovies retrieves movies with optional filtering
func (c *Client) GetMovies(ctx context.Context, filters map[string]string) ([]Movie, error) {
    endpoint := "movie"
    if len(filters) > 0 {
        params := url.Values{}
        for key, value := range filters {
            params.Add(key, value)
        }
        endpoint += "?" + params.Encode()
    }

    var response interface{}
    err := c.doRequest(ctx, "GET", endpoint, nil, &response)
    if err != nil {
        return nil, err
    }

    // Handle both paginated and non-paginated responses
    if movieList, ok := response.(*MovieList); ok {
        return movieList.Data, nil
    }

    // Try to unmarshal as array directly
    jsonData, err := json.Marshal(response)
    if err != nil {
        return nil, err
    }

    var movies []Movie
    err = json.Unmarshal(jsonData, &movies)
    return movies, err
}

// AddMovie adds a movie to the collection
func (c *Client) AddMovie(ctx context.Context, req *MovieCreateRequest) (*Movie, error) {
    var movie Movie
    err := c.doRequest(ctx, "POST", "movie", req, &movie)
    return &movie, err
}

// GetMovie retrieves a specific movie by ID
func (c *Client) GetMovie(ctx context.Context, id int) (*Movie, error) {
    var movie Movie
    endpoint := fmt.Sprintf("movie/%d", id)
    err := c.doRequest(ctx, "GET", endpoint, nil, &movie)
    return &movie, err
}

// GetQualityProfiles retrieves all quality profiles
func (c *Client) GetQualityProfiles(ctx context.Context) ([]QualityProfile, error) {
    var profiles []QualityProfile
    err := c.doRequest(ctx, "GET", "qualityprofile", nil, &profiles)
    return profiles, err
}

// GetMissingMovies retrieves movies that are missing files
func (c *Client) GetMissingMovies(ctx context.Context, page, pageSize int) (*MovieList, error) {
    endpoint := fmt.Sprintf("wanted/missing?page=%d&pageSize=%d&includeAvailable=true",
        page, pageSize)
    var missing MovieList
    err := c.doRequest(ctx, "GET", endpoint, nil, &missing)
    return &missing, err
}

// GetCalendarEvents retrieves calendar events for a date range
func (c *Client) GetCalendarEvents(ctx context.Context, start, end time.Time, includeUnmonitored bool) ([]CalendarEvent, error) {
    endpoint := fmt.Sprintf("calendar?start=%s&end=%s&unmonitored=%t",
        start.Format(time.RFC3339),
        end.Format(time.RFC3339),
        includeUnmonitored)

    var events []CalendarEvent
    err := c.doRequest(ctx, "GET", endpoint, nil, &events)
    return events, err
}

// GetHealthDashboard retrieves comprehensive health information
func (c *Client) GetHealthDashboard(ctx context.Context) (*HealthDashboard, error) {
    var dashboard HealthDashboard
    err := c.doRequest(ctx, "GET", "health/dashboard", nil, &dashboard)
    return &dashboard, err
}

// Example usage
func ExampleUsage() {
    ctx := context.Background()

    client, err := NewClient("http://localhost:7878", "your-api-key-here")
    if err != nil {
        panic(err)
    }

    // Check system status
    status, err := client.GetSystemStatus(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Radarr %s running on %s\n", status.Version, status.OSName)

    // Search for movies
    movies, err := client.SearchMovies(ctx, "inception")
    if err != nil {
        panic(err)
    }

    if len(movies) > 0 {
        movie := movies[0]
        fmt.Printf("Found: %s (%d)\n", movie.Title, movie.Year)

        // Get quality profiles
        profiles, err := client.GetQualityProfiles(ctx)
        if err != nil {
            panic(err)
        }

        if len(profiles) > 0 {
            // Add the movie
            addRequest := &MovieCreateRequest{
                Title:               movie.Title,
                TmdbID:              movie.TmdbID,
                Year:                movie.Year,
                QualityProfileID:    profiles[0].ID,
                RootFolderPath:      "/movies",
                Monitored:           true,
                MinimumAvailability: "released",
                AddOptions: AddOptions{
                    Monitor:        true,
                    SearchForMovie: true,
                    AddMethod:      "manual",
                },
            }

            addedMovie, err := client.AddMovie(ctx, addRequest)
            if err != nil {
                panic(err)
            }
            fmt.Printf("Added: %s (ID: %d)\n", addedMovie.Title, addedMovie.ID)
        }
    }

    // Get health status
    health, err := client.GetHealthDashboard(ctx)
    if err != nil {
        panic(err)
    }
    fmt.Printf("System status: %s\n", health.Status.Status)
    if len(health.Status.Issues) > 0 {
        fmt.Printf("Issues found: %d\n", len(health.Status.Issues))
    }
}
```

## Common Integration Patterns

This section provides proven patterns for integrating with Radarr Go in real-world scenarios.
