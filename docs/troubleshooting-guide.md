# Radarr Go API Troubleshooting Guide

This comprehensive troubleshooting guide covers common issues and solutions when working with the Radarr Go API.

## Table of Contents

1. [Authentication Issues](#authentication-issues)
2. [Connection and Network Problems](#connection-and-network-problems)
3. [Rate Limiting and Performance](#rate-limiting-and-performance)
4. [WebSocket Connection Issues](#websocket-connection-issues)
5. [CORS and Cross-Origin Requests](#cors-and-cross-origin-requests)
6. [Data Format and Serialization Issues](#data-format-and-serialization-issues)
7. [Database-Related Problems](#database-related-problems)
8. [Common Error Codes and Solutions](#common-error-codes-and-solutions)
9. [Debugging Tools and Techniques](#debugging-tools-and-techniques)
10. [Performance Optimization](#performance-optimization)

## Authentication Issues

### Problem: 401 Unauthorized Error

**Symptoms:**

- API requests return HTTP 401
- Error message: "Invalid API key" or "Authentication required"

**Common Causes and Solutions:**

#### 1. Missing or Invalid API Key

```bash
# Check if API key is set correctly
curl -I -H "X-API-Key: YOUR_API_KEY" http://localhost:7878/api/v3/system/status

# Expected: HTTP 200 OK
# If 401: API key is wrong or missing
```

**Solution:**

```python
# Python - Verify API key format and value
import requests

def test_api_key(base_url, api_key):
    headers = {
        'X-API-Key': api_key,
        'Content-Type': 'application/json'
    }

    try:
        response = requests.get(f"{base_url}/api/v3/system/status", headers=headers)
        if response.status_code == 401:
            error_data = response.json()
            print(f"Authentication failed: {error_data.get('error', 'Unknown error')}")
            return False
        elif response.status_code == 200:
            print("Authentication successful!")
            return True
        else:
            print(f"Unexpected status code: {response.status_code}")
            return False
    except Exception as e:
        print(f"Request failed: {e}")
        return False

# Test your API key
test_api_key("http://localhost:7878", "your-api-key-here")
```

#### 2. Incorrect Header Name or Format

```javascript
// JavaScript - Common header mistakes
const axios = require('axios');

// ❌ Wrong header name
const wrongHeaders = {
    'X-Api-Key': 'your-key',  // Should be 'X-API-Key' (capital API)
    'ApiKey': 'your-key'      // Wrong format entirely
};

// ✅ Correct header
const correctHeaders = {
    'X-API-Key': 'your-key'
};

// Test function
async function testHeaders(headers) {
    try {
        const response = await axios.get('http://localhost:7878/api/v3/system/status', {
            headers: headers
        });
        console.log('Success!', response.status);
    } catch (error) {
        console.log('Failed:', error.response?.status, error.response?.data);
    }
}
```

#### 3. Query Parameter Authentication

```python
# Alternative: Use query parameter instead of header
import requests

def test_query_auth(base_url, api_key):
    """Test authentication using query parameter"""
    params = {'apikey': api_key}  # lowercase 'apikey'

    response = requests.get(f"{base_url}/api/v3/system/status", params=params)
    return response.status_code == 200

# This works when headers are problematic
success = test_query_auth("http://localhost:7878", "your-api-key")
```

### Problem: API Key Not Found or Authentication Disabled

**Check Authentication Configuration:**

```bash
# Get system status to check auth method
curl "http://localhost:7878/api/v3/system/status?apikey=your-key" | jq '.authentication'

# Possible values:
# - "none" = Authentication disabled
# - "apikey" = API key required
# - "forms" = Form-based auth (web UI only)
```

**Solution for Disabled Authentication:**

```python
# When authentication is disabled, don't send API key
import requests

def adaptive_request(base_url, endpoint, api_key=None):
    """Make request with optional authentication"""
    headers = {}
    params = {}

    if api_key:
        # Try header first
        headers['X-API-Key'] = api_key

    try:
        response = requests.get(f"{base_url}/api/v3/{endpoint}",
                              headers=headers, params=params)

        if response.status_code == 401 and api_key:
            # Try query parameter
            headers.pop('X-API-Key', None)
            params['apikey'] = api_key
            response = requests.get(f"{base_url}/api/v3/{endpoint}",
                                  headers=headers, params=params)

        return response

    except Exception as e:
        print(f"Request failed: {e}")
        return None
```

## Connection and Network Problems

### Problem: Connection Refused or Timeout

**Symptoms:**

- `Connection refused` errors
- Request timeouts
- DNS resolution failures

**Diagnosis Steps:**

```bash
# 1. Test basic connectivity
ping your-radarr-host

# 2. Test port accessibility
telnet your-radarr-host 7878

# 3. Check if service is running
curl -I http://localhost:7878/ping

# 4. Test from same network
curl -I http://192.168.1.100:7878/ping
```

**Common Solutions:**

#### 1. Service Not Running

```bash
# Check if Radarr Go is running
ps aux | grep radarr
netstat -tlnp | grep :7878

# Start Radarr Go if not running
./radarr --data ./data --config config.yaml
```

#### 2. Firewall Issues

```bash
# Check firewall rules (Linux)
sudo iptables -L | grep 7878
sudo ufw status

# Allow Radarr port
sudo ufw allow 7878

# Windows PowerShell
New-NetFirewallRule -DisplayName "Radarr Go" -Direction Inbound -Port 7878 -Protocol TCP -Action Allow
```

#### 3. Binding/Listen Address Issues

```yaml
# config.yaml - Make sure server binds to correct address
server:
  port: 7878
  host: "0.0.0.0"  # Listen on all interfaces
  # host: "127.0.0.1"  # Only localhost (problematic for remote access)
```

### Problem: SSL/TLS Certificate Issues

**When using HTTPS:**

```python
import requests
import urllib3
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.retry import Retry

# Disable SSL warnings for testing (not recommended for production)
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)

class RadarrSSLClient:
    def __init__(self, base_url, api_key, verify_ssl=True):
        self.base_url = base_url
        self.api_key = api_key

        self.session = requests.Session()
        self.session.verify = verify_ssl
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json'
        })

        # Configure retries
        retry_strategy = Retry(
            total=3,
            status_forcelist=[429, 500, 502, 503, 504],
            method_whitelist=["HEAD", "GET", "OPTIONS"]
        )
        adapter = HTTPAdapter(max_retries=retry_strategy)
        self.session.mount("http://", adapter)
        self.session.mount("https://", adapter)

    def test_ssl_connection(self):
        """Test SSL connection with various configurations"""
        configurations = [
            ("Full SSL verification", True),
            ("No SSL verification", False),
        ]

        for desc, verify in configurations:
            try:
                self.session.verify = verify
                response = self.session.get(f"{self.base_url}/api/v3/ping", timeout=10)
                print(f"✓ {desc}: SUCCESS ({response.status_code})")
                return True
            except requests.exceptions.SSLError as e:
                print(f"✗ {desc}: SSL Error - {e}")
            except requests.exceptions.ConnectionError as e:
                print(f"✗ {desc}: Connection Error - {e}")
            except Exception as e:
                print(f"✗ {desc}: Unexpected Error - {e}")

        return False

# Usage
client = RadarrSSLClient("https://your-radarr-host:7878", "your-api-key")
client.test_ssl_connection()
```

## Rate Limiting and Performance

### Problem: HTTP 429 Too Many Requests

**Understanding Rate Limits:**

- Default: 100 requests per minute per API key
- Applies to all endpoints except `/ping`
- Rate limit headers included in responses

**Detection and Handling:**

```python
import time
import requests
from datetime import datetime, timedelta

class RateLimitedClient:
    def __init__(self, base_url, api_key, rate_limit=100):
        self.base_url = base_url
        self.api_key = api_key
        self.rate_limit = rate_limit
        self.requests_made = []

        self.session = requests.Session()
        self.session.headers.update({'X-API-Key': api_key})

    def _check_rate_limit(self):
        """Check if we're within rate limits"""
        now = datetime.now()
        minute_ago = now - timedelta(minutes=1)

        # Remove requests older than 1 minute
        self.requests_made = [req_time for req_time in self.requests_made if req_time > minute_ago]

        return len(self.requests_made) < self.rate_limit

    def _wait_for_rate_limit(self):
        """Wait if rate limit would be exceeded"""
        if not self._check_rate_limit():
            # Wait until oldest request is > 1 minute old
            oldest_request = min(self.requests_made)
            wait_time = 61 - (datetime.now() - oldest_request).total_seconds()

            if wait_time > 0:
                print(f"Rate limit reached. Waiting {wait_time:.1f} seconds...")
                time.sleep(wait_time)

    def request(self, method, endpoint, **kwargs):
        """Make rate-limited request"""
        self._wait_for_rate_limit()

        url = f"{self.base_url}/api/v3/{endpoint.lstrip('/')}"

        try:
            response = self.session.request(method, url, **kwargs)
            self.requests_made.append(datetime.now())

            # Handle rate limiting response
            if response.status_code == 429:
                retry_after = int(response.headers.get('Retry-After', 60))
                print(f"Rate limited by server. Waiting {retry_after} seconds...")
                time.sleep(retry_after)

                # Retry request
                response = self.session.request(method, url, **kwargs)
                self.requests_made.append(datetime.now())

            # Log rate limit status
            remaining = response.headers.get('X-RateLimit-Remaining')
            if remaining:
                print(f"Rate limit remaining: {remaining}")

            response.raise_for_status()
            return response

        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            raise

# Usage example
client = RateLimitedClient("http://localhost:7878", "your-api-key")

# Make multiple requests safely
for i in range(150):  # Exceeds rate limit
    try:
        response = client.request('GET', 'system/status')
        print(f"Request {i+1}: {response.status_code}")
    except Exception as e:
        print(f"Request {i+1} failed: {e}")
        break
```

### Problem: Slow API Responses

**Performance Optimization Strategies:**

```python
import asyncio
import aiohttp
import time
from concurrent.futures import ThreadPoolExecutor, as_completed

class OptimizedRadarrClient:
    def __init__(self, base_url, api_key, max_concurrent=10):
        self.base_url = base_url
        self.api_key = api_key
        self.max_concurrent = max_concurrent

    # 1. Async HTTP Client (fastest for many requests)
    async def async_batch_request(self, endpoints):
        """Make multiple async requests"""
        async with aiohttp.ClientSession(
            headers={'X-API-Key': self.api_key},
            timeout=aiohttp.ClientTimeout(total=30)
        ) as session:

            async def fetch(endpoint):
                url = f"{self.base_url}/api/v3/{endpoint.lstrip('/')}"
                try:
                    async with session.get(url) as response:
                        return await response.json()
                except Exception as e:
                    print(f"Async request failed for {endpoint}: {e}")
                    return None

            # Execute all requests concurrently
            tasks = [fetch(endpoint) for endpoint in endpoints]
            results = await asyncio.gather(*tasks, return_exceptions=True)

            return results

    # 2. Thread Pool (good for I/O bound operations)
    def threaded_batch_request(self, endpoints):
        """Make multiple requests using thread pool"""
        import requests

        def fetch_single(endpoint):
            url = f"{self.base_url}/api/v3/{endpoint.lstrip('/')}"
            headers = {'X-API-Key': self.api_key}

            try:
                response = requests.get(url, headers=headers, timeout=30)
                response.raise_for_status()
                return response.json()
            except Exception as e:
                print(f"Thread request failed for {endpoint}: {e}")
                return None

        with ThreadPoolExecutor(max_workers=self.max_concurrent) as executor:
            future_to_endpoint = {
                executor.submit(fetch_single, endpoint): endpoint
                for endpoint in endpoints
            }

            results = {}
            for future in as_completed(future_to_endpoint):
                endpoint = future_to_endpoint[future]
                try:
                    result = future.result()
                    results[endpoint] = result
                except Exception as e:
                    print(f"Thread execution failed for {endpoint}: {e}")
                    results[endpoint] = None

            return results

    # 3. Connection Pooling (for sequential requests)
    def session_based_requests(self, endpoints):
        """Use session for connection reuse"""
        import requests

        session = requests.Session()
        session.headers.update({'X-API-Key': self.api_key})

        # Configure connection pooling
        adapter = requests.adapters.HTTPAdapter(
            pool_connections=10,
            pool_maxsize=20,
            max_retries=3
        )
        session.mount('http://', adapter)
        session.mount('https://', adapter)

        results = []
        for endpoint in endpoints:
            url = f"{self.base_url}/api/v3/{endpoint.lstrip('/')}"
            try:
                response = session.get(url, timeout=30)
                response.raise_for_status()
                results.append(response.json())
            except Exception as e:
                print(f"Session request failed for {endpoint}: {e}")
                results.append(None)

        session.close()
        return results

# Performance comparison
async def compare_performance():
    client = OptimizedRadarrClient("http://localhost:7878", "your-api-key")

    endpoints = [
        'system/status',
        'movie?pageSize=100',
        'qualityprofile',
        'health',
        'wanted/missing?pageSize=50',
        'calendar'
    ] * 5  # 30 total requests

    # Test async method
    start_time = time.time()
    async_results = await client.async_batch_request(endpoints)
    async_time = time.time() - start_time
    print(f"Async requests: {async_time:.2f} seconds")

    # Test threaded method
    start_time = time.time()
    threaded_results = client.threaded_batch_request(endpoints)
    threaded_time = time.time() - start_time
    print(f"Threaded requests: {threaded_time:.2f} seconds")

    # Test session method
    start_time = time.time()
    session_results = client.session_based_requests(endpoints)
    session_time = time.time() - start_time
    print(f"Session requests: {session_time:.2f} seconds")

# Run comparison
# asyncio.run(compare_performance())
```

## WebSocket Connection Issues

### Problem: WebSocket Connection Failures

**Common WebSocket Issues and Solutions:**

```javascript
class RadarrWebSocketTroubleshooter {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl.replace(/^http/, 'ws');
        this.apiKey = apiKey;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.debugMode = true;
    }

    connect() {
        const wsUrl = `${this.baseUrl}/signalr/radarr`;

        if (this.debugMode) {
            console.log(`Attempting WebSocket connection to: ${wsUrl}`);
        }

        try {
            this.ws = new WebSocket(wsUrl, {
                headers: {
                    'X-API-Key': this.apiKey
                }
            });

            this.setupEventHandlers();

        } catch (error) {
            console.error('WebSocket connection failed:', error);
            this.handleConnectionError(error);
        }
    }

    setupEventHandlers() {
        this.ws.onopen = (event) => {
            console.log('✓ WebSocket connected successfully');
            this.reconnectAttempts = 0;

            // Send connection diagnostics
            this.sendDiagnostics();
        };

        this.ws.onclose = (event) => {
            console.log(`WebSocket closed: Code ${event.code}, Reason: ${event.reason}`);

            // Decode close codes
            this.diagnoseCloseCode(event.code);

            if (event.code !== 1000) { // Not normal closure
                this.attemptReconnection();
            }
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.diagnoseError(error);
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleMessage(data);
            } catch (e) {
                console.error('Failed to parse WebSocket message:', e);
                console.log('Raw message:', event.data);
            }
        };
    }

    diagnoseCloseCode(code) {
        const closeCodes = {
            1000: 'Normal Closure',
            1001: 'Going Away',
            1002: 'Protocol Error',
            1003: 'Unsupported Data',
            1006: 'Abnormal Closure (no close frame)',
            1007: 'Invalid frame payload data',
            1008: 'Policy Violation',
            1009: 'Message Too Big',
            1010: 'Mandatory Extension',
            1011: 'Internal Server Error',
            1012: 'Service Restart',
            1013: 'Try Again Later',
            1015: 'TLS Handshake'
        };

        const reason = closeCodes[code] || 'Unknown';
        console.log(`Close code ${code}: ${reason}`);

        // Provide specific troubleshooting advice
        switch (code) {
            case 1002:
                console.log('💡 Try: Check API key format and authentication');
                break;
            case 1006:
                console.log('💡 Try: Check network connectivity and firewall settings');
                break;
            case 1008:
                console.log('💡 Try: Verify API key has proper permissions');
                break;
            case 1011:
                console.log('💡 Try: Check Radarr server logs for errors');
                break;
            case 1015:
                console.log('💡 Try: Check SSL/TLS certificate configuration');
                break;
        }
    }

    diagnoseError(error) {
        console.log('WebSocket Error Details:');
        console.log('- Type:', error.type);
        console.log('- Message:', error.message);

        // Common error patterns
        if (error.message.includes('ECONNREFUSED')) {
            console.log('💡 Connection refused - check if Radarr is running');
        } else if (error.message.includes('timeout')) {
            console.log('💡 Connection timeout - check network and firewall');
        } else if (error.message.includes('certificate')) {
            console.log('💡 SSL certificate issue - check certificate validity');
        }
    }

    sendDiagnostics() {
        // Send handshake
        this.send({
            protocol: 'json',
            version: 1
        });

        // Test ping
        setTimeout(() => {
            this.send({ type: 6 }); // Ping
        }, 1000);
    }

    handleMessage(data) {
        if (this.debugMode) {
            console.log('WebSocket message received:', data);
        }

        // Handle pong response
        if (data.type === 6) {
            console.log('✓ WebSocket ping/pong working');
        }

        // Handle other message types...
    }

    send(data) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(data));
        } else {
            console.error('Cannot send: WebSocket not open');
        }
    }

    attemptReconnection() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.pow(2, this.reconnectAttempts) * 1000;

            console.log(`Reconnecting in ${delay/1000}s (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

            setTimeout(() => {
                this.connect();
            }, delay);
        } else {
            console.error('Max reconnection attempts reached');
        }
    }

    handleConnectionError(error) {
        console.log('WebSocket Connection Error Troubleshooting:');
        console.log('1. Verify Radarr is running and accessible');
        console.log('2. Check API key is correct');
        console.log('3. Verify WebSocket endpoint is available');
        console.log('4. Check firewall/proxy settings');

        // Test HTTP endpoint first
        this.testHttpConnectivity();
    }

    async testHttpConnectivity() {
        try {
            const response = await fetch(`${this.baseUrl.replace('ws', 'http')}/api/v3/ping`);
            if (response.ok) {
                console.log('✓ HTTP connectivity working - WebSocket issue is specific');
            } else {
                console.log('✗ HTTP connectivity also failing');
            }
        } catch (error) {
            console.log('✗ HTTP connectivity test failed:', error.message);
        }
    }
}

// Usage
const wsTroubleshooter = new RadarrWebSocketTroubleshooter('ws://localhost:7878', 'your-api-key');
wsTroubleshooter.connect();
```

## CORS and Cross-Origin Requests

### Problem: CORS Policy Violations

**Symptoms:**

- Browser console shows CORS errors
- Requests work in backend but fail in browser
- "Access-Control-Allow-Origin" errors

**Understanding CORS in Radarr Go:**

```yaml
# config.yaml - CORS configuration
server:
  cors:
    enabled: true
    allowed_origins:
      - "http://localhost:3000"
      - "https://your-app.com"
    allowed_methods:
      - GET
      - POST
      - PUT
      - DELETE
    allowed_headers:
      - "X-API-Key"
      - "Content-Type"
      - "Authorization"
    max_age: 86400
```

**Client-Side CORS Handling:**

```javascript
class CORSAwareRadarrClient {
    constructor(baseUrl, apiKey) {
        this.baseUrl = baseUrl;
        this.apiKey = apiKey;
    }

    async makeRequest(method, endpoint, data = null) {
        const url = `${this.baseUrl}/api/v3/${endpoint}`;

        const options = {
            method: method,
            headers: {
                'X-API-Key': this.apiKey,
                'Content-Type': 'application/json',
            },
            // CORS handling
            mode: 'cors',
            credentials: 'omit'
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(url, options);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();

        } catch (error) {
            if (error.message.includes('CORS')) {
                console.error('CORS Error - Troubleshooting steps:');
                console.error('1. Check Radarr CORS configuration');
                console.error('2. Verify your domain is in allowed_origins');
                console.error('3. Check if preflight request is being handled');

                // Try alternative approach
                return this.tryAlternativeRequest(method, endpoint, data);
            }
            throw error;
        }
    }

    async tryAlternativeRequest(method, endpoint, data) {
        console.log('Trying JSONP/proxy approach...');

        // Option 1: Use JSONP for GET requests
        if (method === 'GET') {
            return this.jsonpRequest(endpoint);
        }

        // Option 2: Use a CORS proxy (development only!)
        if (process.env.NODE_ENV === 'development') {
            return this.proxyRequest(method, endpoint, data);
        }

        throw new Error('CORS error and no alternatives available');
    }

    jsonpRequest(endpoint) {
        return new Promise((resolve, reject) => {
            const script = document.createElement('script');
            const callbackName = `radarr_callback_${Date.now()}`;

            window[callbackName] = (data) => {
                document.head.removeChild(script);
                delete window[callbackName];
                resolve(data);
            };

            script.src = `${this.baseUrl}/api/v3/${endpoint}?apikey=${this.apiKey}&callback=${callbackName}`;
            script.onerror = () => {
                document.head.removeChild(script);
                delete window[callbackName];
                reject(new Error('JSONP request failed'));
            };

            document.head.appendChild(script);
        });
    }

    async proxyRequest(method, endpoint, data) {
        // Using cors-anywhere or similar proxy (development only!)
        const proxyUrl = 'https://cors-anywhere.herokuapp.com/';
        const targetUrl = `${this.baseUrl}/api/v3/${endpoint}`;

        const options = {
            method: method,
            headers: {
                'X-API-Key': this.apiKey,
                'Content-Type': 'application/json',
                'X-Requested-With': 'XMLHttpRequest'
            }
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        const response = await fetch(proxyUrl + targetUrl, options);
        return await response.json();
    }
}

// Testing CORS functionality
async function testCORS() {
    const client = new CORSAwareRadarrClient('http://localhost:7878', 'your-api-key');

    try {
        const status = await client.makeRequest('GET', 'system/status');
        console.log('✓ CORS working correctly');
        return true;
    } catch (error) {
        console.error('✗ CORS test failed:', error.message);
        return false;
    }
}
```

**Server-Side CORS Proxy (Node.js):**

```javascript
// cors-proxy.js - Simple CORS proxy for development
const express = require('express');
const { createProxyMiddleware } = require('http-proxy-middleware');

const app = express();

// CORS proxy for Radarr
app.use('/radarr', createProxyMiddleware({
    target: 'http://localhost:7878',
    changeOrigin: true,
    pathRewrite: {
        '^/radarr': '', // Remove /radarr prefix
    },
    onProxyReq: (proxyReq, req, res) => {
        // Add CORS headers
        res.header('Access-Control-Allow-Origin', '*');
        res.header('Access-Control-Allow-Methods', 'GET,PUT,POST,DELETE,OPTIONS');
        res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-API-Key');
    },
    onError: (err, req, res) => {
        console.error('Proxy error:', err);
        res.status(500).send('Proxy error');
    }
}));

// Handle preflight requests
app.options('*', (req, res) => {
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', 'GET,PUT,POST,DELETE,OPTIONS');
    res.header('Access-Control-Allow-Headers', 'Content-Type, Authorization, X-API-Key');
    res.send();
});

app.listen(3001, () => {
    console.log('CORS proxy running on http://localhost:3001');
    console.log('Use http://localhost:3001/radarr/api/v3/... for Radarr API');
});
```

## Data Format and Serialization Issues

### Problem: JSON Parsing Errors

**Common JSON Issues:**

```python
import json
from datetime import datetime
from decimal import Decimal

class RadarrJSONHandler:
    """Handle common JSON serialization issues with Radarr API"""

    @staticmethod
    def safe_json_loads(json_str):
        """Safely parse JSON with error handling"""
        try:
            return json.loads(json_str)
        except json.JSONDecodeError as e:
            print(f"JSON Parse Error: {e}")
            print(f"Error at position {e.pos}")
            print(f"Context: ...{json_str[max(0, e.pos-20):e.pos+20]}...")

            # Try to fix common issues
            fixed_json = RadarrJSONHandler.fix_common_json_issues(json_str)
            if fixed_json != json_str:
                print("Attempting to fix JSON...")
                try:
                    return json.loads(fixed_json)
                except json.JSONDecodeError:
                    print("JSON fix failed")

            raise

    @staticmethod
    def fix_common_json_issues(json_str):
        """Fix common JSON formatting issues"""
        # Fix trailing commas
        import re
        fixed = re.sub(r',(\s*[}\]])', r'\1', json_str)

        # Fix single quotes to double quotes
        fixed = re.sub(r"'([^']*)':", r'"\1":', fixed)

        # Fix None to null
        fixed = fixed.replace('None', 'null')

        # Fix True/False to lowercase
        fixed = fixed.replace('True', 'true').replace('False', 'false')

        return fixed

    @staticmethod
    def custom_json_encoder(obj):
        """Custom JSON encoder for Python objects"""
        if isinstance(obj, datetime):
            return obj.isoformat()
        elif isinstance(obj, Decimal):
            return float(obj)
        elif hasattr(obj, '__dict__'):
            return obj.__dict__
        else:
            return str(obj)

    @staticmethod
    def safe_json_dumps(obj):
        """Safely serialize object to JSON"""
        try:
            return json.dumps(obj, default=RadarrJSONHandler.custom_json_encoder, indent=2)
        except (TypeError, ValueError) as e:
            print(f"JSON Serialization Error: {e}")

            # Try with string conversion
            try:
                return json.dumps(str(obj))
            except:
                return '{"error": "Could not serialize object"}'

# Example usage
def test_json_handling():
    # Test malformed JSON (common from API responses)
    malformed_json = '''
    {
        "title": "Inception",
        "year": 2010,
        "monitored": True,
        "tags": [1, 2, 3,],
        "added": None,
        "overview": 'A movie about dreams',
    }
    '''

    try:
        data = RadarrJSONHandler.safe_json_loads(malformed_json)
        print("Successfully parsed JSON:", data)
    except Exception as e:
        print(f"Failed to parse JSON: {e}")
```

### Problem: Date/Time Format Issues

**Handling Different Date Formats:**

```python
from datetime import datetime, timezone
import re

class RadarrDateHandler:
    """Handle various date/time formats from Radarr API"""

    # Common date patterns in Radarr
    DATE_PATTERNS = [
        r'(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z)',  # ISO 8601 UTC
        r'(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z)',  # ISO with milliseconds
        r'(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[+-]\d{2}:\d{2})',  # ISO with timezone
        r'(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})',  # SQL datetime
        r'(\d{4}-\d{2}-\d{2})',  # Date only
    ]

    @staticmethod
    def parse_radarr_date(date_str):
        """Parse various date formats from Radarr"""
        if not date_str:
            return None

        # Try ISO format first
        try:
            if date_str.endswith('Z'):
                return datetime.fromisoformat(date_str.replace('Z', '+00:00'))
            else:
                return datetime.fromisoformat(date_str)
        except ValueError:
            pass

        # Try other common formats
        formats = [
            '%Y-%m-%dT%H:%M:%SZ',
            '%Y-%m-%dT%H:%M:%S.%fZ',
            '%Y-%m-%d %H:%M:%S',
            '%Y-%m-%d',
        ]

        for fmt in formats:
            try:
                return datetime.strptime(date_str, fmt)
            except ValueError:
                continue

        print(f"Warning: Could not parse date: {date_str}")
        return None

    @staticmethod
    def format_for_radarr(dt):
        """Format datetime for Radarr API"""
        if dt is None:
            return None

        if dt.tzinfo is None:
            dt = dt.replace(tzinfo=timezone.utc)

        return dt.isoformat().replace('+00:00', 'Z')

    @staticmethod
    def normalize_movie_dates(movie_data):
        """Normalize all date fields in movie object"""
        date_fields = [
            'added', 'inCinemas', 'physicalRelease', 'digitalRelease',
            'createdAt', 'updatedAt'
        ]

        for field in date_fields:
            if field in movie_data and movie_data[field]:
                movie_data[field] = RadarrDateHandler.parse_radarr_date(movie_data[field])

        return movie_data

# Example usage
def test_date_handling():
    # Test various date formats
    test_dates = [
        "2024-01-15T14:30:00Z",
        "2024-01-15T14:30:00.123Z",
        "2024-01-15T14:30:00+01:00",
        "2024-01-15 14:30:00",
        "2024-01-15",
        None,
        "",
        "invalid-date"
    ]

    handler = RadarrDateHandler()

    for date_str in test_dates:
        parsed = handler.parse_radarr_date(date_str)
        formatted = handler.format_for_radarr(parsed)
        print(f"'{date_str}' -> {parsed} -> '{formatted}'")
```

## Database-Related Problems

### Problem: Database Connection Issues

**Diagnosis and Solutions:**

```python
import psycopg2
import mysql.connector
import logging

class RadarrDatabaseTroubleshooter:
    """Troubleshoot database connection issues"""

    def __init__(self, config):
        self.config = config
        self.logger = logging.getLogger(__name__)

    def test_postgresql_connection(self):
        """Test PostgreSQL connection"""
        try:
            conn = psycopg2.connect(
                host=self.config.get('host', 'localhost'),
                port=self.config.get('port', 5432),
                database=self.config.get('database', 'radarr'),
                user=self.config.get('username', 'radarr'),
                password=self.config.get('password', ''),
                connect_timeout=10
            )

            cursor = conn.cursor()
            cursor.execute('SELECT version();')
            version = cursor.fetchone()[0]

            cursor.close()
            conn.close()

            self.logger.info(f"✓ PostgreSQL connection successful: {version}")
            return True

        except psycopg2.OperationalError as e:
            self.logger.error(f"✗ PostgreSQL connection failed: {e}")
            self.diagnose_postgresql_error(str(e))
            return False
        except Exception as e:
            self.logger.error(f"✗ Unexpected PostgreSQL error: {e}")
            return False

    def test_mariadb_connection(self):
        """Test MariaDB connection"""
        try:
            conn = mysql.connector.connect(
                host=self.config.get('host', 'localhost'),
                port=self.config.get('port', 3306),
                database=self.config.get('database', 'radarr'),
                user=self.config.get('username', 'radarr'),
                password=self.config.get('password', ''),
                connection_timeout=10
            )

            cursor = conn.cursor()
            cursor.execute('SELECT VERSION();')
            version = cursor.fetchone()[0]

            cursor.close()
            conn.close()

            self.logger.info(f"✓ MariaDB connection successful: {version}")
            return True

        except mysql.connector.Error as e:
            self.logger.error(f"✗ MariaDB connection failed: {e}")
            self.diagnose_mariadb_error(str(e))
            return False
        except Exception as e:
            self.logger.error(f"✗ Unexpected MariaDB error: {e}")
            return False

    def diagnose_postgresql_error(self, error_msg):
        """Provide specific PostgreSQL troubleshooting advice"""
        if 'could not connect to server' in error_msg.lower():
            print("💡 PostgreSQL server connection failed:")
            print("   - Check if PostgreSQL is running")
            print("   - Verify host and port are correct")
            print("   - Check firewall settings")

        elif 'authentication failed' in error_msg.lower():
            print("💡 PostgreSQL authentication failed:")
            print("   - Verify username and password")
            print("   - Check pg_hba.conf configuration")
            print("   - Ensure user has database access")

        elif 'database does not exist' in error_msg.lower():
            print("💡 PostgreSQL database missing:")
            print("   - Create the database: CREATE DATABASE radarr;")
            print("   - Grant permissions: GRANT ALL ON DATABASE radarr TO radarr_user;")

        elif 'timeout' in error_msg.lower():
            print("💡 PostgreSQL connection timeout:")
            print("   - Check network connectivity")
            print("   - Increase connection timeout")
            print("   - Verify PostgreSQL is not overloaded")

    def diagnose_mariadb_error(self, error_msg):
        """Provide specific MariaDB troubleshooting advice"""
        if 'connection refused' in error_msg.lower():
            print("💡 MariaDB connection refused:")
            print("   - Check if MariaDB is running: systemctl status mariadb")
            print("   - Verify port 3306 is open")
            print("   - Check bind-address in my.cnf")

        elif 'access denied' in error_msg.lower():
            print("💡 MariaDB access denied:")
            print("   - Check username/password combination")
            print("   - Verify user permissions: SHOW GRANTS FOR 'user'@'host';")
            print("   - Check if user can connect from this host")

        elif 'unknown database' in error_msg.lower():
            print("💡 MariaDB database missing:")
            print("   - Create database: CREATE DATABASE radarr;")
            print("   - Grant permissions: GRANT ALL ON radarr.* TO 'radarr'@'%';")

    def test_connection_pool(self):
        """Test database connection pooling"""
        # This would test connection pool settings
        print("Testing connection pool configuration...")

        pool_tests = {
            'max_connections': 20,
            'idle_timeout': 300,
            'connection_lifetime': 3600
        }

        for setting, value in pool_tests.items():
            print(f"  {setting}: {value}")

        return True

# Usage
config = {
    'host': 'localhost',
    'port': 5432,
    'database': 'radarr',
    'username': 'radarr',
    'password': 'your_password'
}

troubleshooter = RadarrDatabaseTroubleshooter(config)

# Test PostgreSQL
if not troubleshooter.test_postgresql_connection():
    print("\nTroubleshooting PostgreSQL connection...")

# Test MariaDB (change port to 3306)
config['port'] = 3306
if not troubleshooter.test_mariadb_connection():
    print("\nTroubleshooting MariaDB connection...")
```

## Common Error Codes and Solutions

### HTTP Status Code Reference

```python
class RadarrErrorHandler:
    """Handle and explain common Radarr API errors"""

    ERROR_SOLUTIONS = {
        # 4xx Client Errors
        400: {
            'name': 'Bad Request',
            'causes': [
                'Invalid JSON in request body',
                'Missing required fields',
                'Invalid field values',
                'Malformed request parameters'
            ],
            'solutions': [
                'Validate JSON syntax',
                'Check required fields in API documentation',
                'Verify data types match expected values',
                'Use proper URL encoding for parameters'
            ]
        },

        401: {
            'name': 'Unauthorized',
            'causes': [
                'Missing API key',
                'Invalid API key',
                'API key in wrong format',
                'Authentication disabled but key provided'
            ],
            'solutions': [
                'Check API key is correct',
                'Verify API key format (no spaces/special chars)',
                'Use X-API-Key header or apikey parameter',
                'Check authentication settings in Radarr'
            ]
        },

        403: {
            'name': 'Forbidden',
            'causes': [
                'API key lacks required permissions',
                'Resource access denied',
                'Rate limit exceeded',
                'IP address blocked'
            ],
            'solutions': [
                'Check API key permissions',
                'Verify user role has API access',
                'Implement rate limiting in client',
                'Check IP whitelist settings'
            ]
        },

        404: {
            'name': 'Not Found',
            'causes': [
                'Endpoint does not exist',
                'Resource ID not found',
                'Wrong API version in URL',
                'Typo in endpoint URL'
            ],
            'solutions': [
                'Check endpoint exists in API documentation',
                'Verify resource ID is correct',
                'Use /api/v3/ prefix for all endpoints',
                'Double-check URL spelling'
            ]
        },

        409: {
            'name': 'Conflict',
            'causes': [
                'Resource already exists',
                'Duplicate movie in collection',
                'Conflicting operation in progress',
                'Database constraint violation'
            ],
            'solutions': [
                'Check if movie already exists before adding',
                'Use update instead of create for existing resources',
                'Wait for current operations to complete',
                'Verify unique constraints'
            ]
        },

        422: {
            'name': 'Unprocessable Entity',
            'causes': [
                'Validation errors',
                'Invalid TMDB ID',
                'Missing quality profile',
                'Invalid root folder path'
            ],
            'solutions': [
                'Check validation error details',
                'Verify TMDB ID exists',
                'Ensure quality profile exists',
                'Verify root folder is configured'
            ]
        },

        429: {
            'name': 'Too Many Requests',
            'causes': [
                'Rate limit exceeded',
                'Too many concurrent requests',
                'Bulk operations too fast'
            ],
            'solutions': [
                'Implement exponential backoff',
                'Reduce request frequency',
                'Add delays between requests',
                'Use batch endpoints where available'
            ]
        },

        # 5xx Server Errors
        500: {
            'name': 'Internal Server Error',
            'causes': [
                'Radarr server bug',
                'Database connection error',
                'Unhandled exception',
                'Configuration error'
            ],
            'solutions': [
                'Check Radarr server logs',
                'Verify database connectivity',
                'Restart Radarr service',
                'Report bug if reproducible'
            ]
        },

        502: {
            'name': 'Bad Gateway',
            'causes': [
                'Radarr server down',
                'Proxy/reverse proxy error',
                'Network connectivity issue'
            ],
            'solutions': [
                'Check if Radarr service is running',
                'Verify proxy configuration',
                'Test direct connection to Radarr'
            ]
        },

        503: {
            'name': 'Service Unavailable',
            'causes': [
                'Radarr maintenance mode',
                'Server overloaded',
                'Database maintenance',
                'Temporary service disruption'
            ],
            'solutions': [
                'Wait and retry request',
                'Check Radarr status page',
                'Implement retry with backoff',
                'Contact administrator'
            ]
        },

        504: {
            'name': 'Gateway Timeout',
            'causes': [
                'Request processing timeout',
                'Database query timeout',
                'External service timeout',
                'Network congestion'
            ],
            'solutions': [
                'Increase client timeout',
                'Optimize query parameters',
                'Retry with exponential backoff',
                'Check network connectivity'
            ]
        }
    }

    @classmethod
    def explain_error(cls, status_code, response_text=None):
        """Provide detailed explanation of error"""
        error_info = cls.ERROR_SOLUTIONS.get(status_code, {
            'name': 'Unknown Error',
            'causes': ['Unknown cause'],
            'solutions': ['Check API documentation']
        })

        print(f"\n🚨 HTTP {status_code}: {error_info['name']}")
        print("\n📋 Possible Causes:")
        for cause in error_info['causes']:
            print(f"   • {cause}")

        print("\n💡 Suggested Solutions:")
        for solution in error_info['solutions']:
            print(f"   • {solution}")

        if response_text:
            try:
                import json
                error_data = json.loads(response_text)
                if 'error' in error_data:
                    print(f"\n📄 Server Error Message: {error_data['error']}")
                if 'details' in error_data:
                    print(f"📄 Additional Details: {error_data['details']}")
            except:
                print(f"\n📄 Raw Response: {response_text[:200]}...")

    @classmethod
    def handle_request_error(cls, response):
        """Handle request errors with detailed diagnostics"""
        cls.explain_error(response.status_code, response.text)

        # Additional diagnostics based on status code
        if response.status_code == 429:
            retry_after = response.headers.get('Retry-After')
            if retry_after:
                print(f"⏰ Retry After: {retry_after} seconds")

        elif response.status_code in [500, 502, 503, 504]:
            print("🔧 Server-side issue detected. Consider:")
            print("   • Checking Radarr server logs")
            print("   • Monitoring system resources")
            print("   • Contacting administrator")
```

## Debugging Tools and Techniques

### Comprehensive API Debugging Suite

```python
import requests
import json
import time
import logging
from datetime import datetime
import traceback

class RadarrAPIDebugger:
    """Comprehensive debugging toolkit for Radarr API"""

    def __init__(self, base_url, api_key, debug_level='INFO'):
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key

        # Setup detailed logging
        logging.basicConfig(
            level=getattr(logging, debug_level),
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
            handlers=[
                logging.FileHandler('radarr_api_debug.log'),
                logging.StreamHandler()
            ]
        )
        self.logger = logging.getLogger('RadarrDebugger')

        # Request session with debugging
        self.session = requests.Session()
        self.session.headers.update({
            'X-API-Key': api_key,
            'Content-Type': 'application/json',
            'User-Agent': 'Radarr-API-Debugger/1.0'
        })

        # Debug statistics
        self.stats = {
            'requests_made': 0,
            'successful_requests': 0,
            'failed_requests': 0,
            'total_response_time': 0,
            'errors': []
        }

    def debug_request(self, method, endpoint, **kwargs):
        """Make request with comprehensive debugging"""
        url = f"{self.base_url}/api/v3/{endpoint.lstrip('/')}"

        # Pre-request logging
        self.logger.info(f"🚀 Starting {method} request to {endpoint}")
        self.logger.debug(f"Full URL: {url}")
        self.logger.debug(f"Headers: {dict(self.session.headers)}")

        if 'json' in kwargs:
            self.logger.debug(f"Request body: {json.dumps(kwargs['json'], indent=2)}")

        start_time = time.time()

        try:
            # Make request
            response = self.session.request(method, url, **kwargs)

            # Calculate timing
            response_time = time.time() - start_time
            self.stats['requests_made'] += 1
            self.stats['total_response_time'] += response_time

            # Log response details
            self.logger.info(f"📋 Response: {response.status_code} ({response_time:.3f}s)")
            self.logger.debug(f"Response headers: {dict(response.headers)}")

            # Log response body
            try:
                response_json = response.json()
                self.logger.debug(f"Response body: {json.dumps(response_json, indent=2, default=str)}")
            except:
                self.logger.debug(f"Response body (raw): {response.text[:500]}...")

            # Handle errors
            if response.status_code >= 400:
                self.stats['failed_requests'] += 1
                self.handle_error_response(response)
            else:
                self.stats['successful_requests'] += 1
                self.logger.info("✅ Request successful")

            return response

        except Exception as e:
            # Exception handling
            response_time = time.time() - start_time
            self.stats['requests_made'] += 1
            self.stats['failed_requests'] += 1

            error_info = {
                'timestamp': datetime.now().isoformat(),
                'method': method,
                'endpoint': endpoint,
                'error': str(e),
                'traceback': traceback.format_exc()
            }
            self.stats['errors'].append(error_info)

            self.logger.error(f"❌ Request failed: {e}")
            self.logger.debug(f"Full traceback:\n{traceback.format_exc()}")

            raise

    def handle_error_response(self, response):
        """Handle and analyze error responses"""
        self.logger.error(f"❌ HTTP {response.status_code}: {response.reason}")

        # Try to parse error details
        try:
            error_data = response.json()
            if 'error' in error_data:
                self.logger.error(f"Error message: {error_data['error']}")
            if 'details' in error_data:
                self.logger.error(f"Error details: {error_data['details']}")
        except:
            self.logger.error(f"Raw error response: {response.text[:200]}...")

        # Store error for analysis
        error_info = {
            'timestamp': datetime.now().isoformat(),
            'status_code': response.status_code,
            'url': response.url,
            'response_text': response.text
        }
        self.stats['errors'].append(error_info)

    def test_connectivity(self):
        """Comprehensive connectivity test"""
        self.logger.info("🔍 Starting connectivity diagnostics...")

        tests = [
            ('Ping endpoint', 'GET', 'ping'),
            ('System status', 'GET', 'system/status'),
            ('Health check', 'GET', 'health'),
            ('Quality profiles', 'GET', 'qualityprofile'),
        ]

        results = {}

        for test_name, method, endpoint in tests:
            self.logger.info(f"Testing: {test_name}")
            try:
                response = self.debug_request(method, endpoint)
                results[test_name] = {
                    'status': 'PASS',
                    'status_code': response.status_code,
                    'response_time': response.elapsed.total_seconds()
                }
            except Exception as e:
                results[test_name] = {
                    'status': 'FAIL',
                    'error': str(e)
                }

        # Print results
        self.logger.info("🏁 Connectivity test results:")
        for test_name, result in results.items():
            if result['status'] == 'PASS':
                self.logger.info(f"  ✅ {test_name}: {result['status_code']} ({result['response_time']:.3f}s)")
            else:
                self.logger.error(f"  ❌ {test_name}: {result['error']}")

        return results

    def analyze_performance(self, endpoint, iterations=10):
        """Analyze endpoint performance"""
        self.logger.info(f"📊 Performance analysis for {endpoint} ({iterations} iterations)")

        times = []
        errors = []

        for i in range(iterations):
            try:
                response = self.debug_request('GET', endpoint)
                times.append(response.elapsed.total_seconds())

                if response.status_code >= 400:
                    errors.append(response.status_code)

            except Exception as e:
                errors.append(str(e))

            # Small delay between requests
            time.sleep(0.1)

        # Calculate statistics
        if times:
            avg_time = sum(times) / len(times)
            min_time = min(times)
            max_time = max(times)

            self.logger.info(f"📈 Performance Results:")
            self.logger.info(f"  Average: {avg_time:.3f}s")
            self.logger.info(f"  Min: {min_time:.3f}s")
            self.logger.info(f"  Max: {max_time:.3f}s")
            self.logger.info(f"  Success Rate: {(len(times) - len(errors)) / iterations * 100:.1f}%")

        if errors:
            self.logger.warning(f"  Errors: {len(errors)} ({errors})")

        return {
            'times': times,
            'errors': errors,
            'avg_time': avg_time if times else None,
            'min_time': min_time if times else None,
            'max_time': max_time if times else None,
            'success_rate': (iterations - len(errors)) / iterations * 100
        }

    def debug_websocket_connection(self):
        """Debug WebSocket connectivity"""
        self.logger.info("🔗 Testing WebSocket connection...")

        # This would implement WebSocket debugging
        # For now, test if WebSocket endpoint is reachable
        try:
            import websocket

            ws_url = self.base_url.replace('http', 'ws') + '/signalr/radarr'
            self.logger.info(f"WebSocket URL: {ws_url}")

            def on_open(ws):
                self.logger.info("✅ WebSocket connection opened")
                ws.close()

            def on_error(ws, error):
                self.logger.error(f"❌ WebSocket error: {error}")

            def on_close(ws, close_status_code, close_msg):
                self.logger.info(f"🔒 WebSocket closed: {close_status_code} - {close_msg}")

            ws = websocket.WebSocketApp(ws_url,
                                      header={'X-API-Key': self.api_key},
                                      on_open=on_open,
                                      on_error=on_error,
                                      on_close=on_close)

            ws.run_forever(timeout=10)

        except ImportError:
            self.logger.warning("websocket-client not installed, skipping WebSocket test")
        except Exception as e:
            self.logger.error(f"WebSocket test failed: {e}")

    def generate_debug_report(self):
        """Generate comprehensive debug report"""
        report = f"""
RADARR GO API DEBUG REPORT
=========================
Generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
Base URL: {self.base_url}

REQUEST STATISTICS
------------------
Total Requests: {self.stats['requests_made']}
Successful: {self.stats['successful_requests']}
Failed: {self.stats['failed_requests']}
Success Rate: {(self.stats['successful_requests'] / max(1, self.stats['requests_made']) * 100):.1f}%
Average Response Time: {(self.stats['total_response_time'] / max(1, self.stats['requests_made'])):.3f}s

RECENT ERRORS
-------------
"""

        for error in self.stats['errors'][-5:]:  # Last 5 errors
            report += f"[{error['timestamp']}] {error.get('error', 'HTTP ' + str(error.get('status_code', 'Unknown')))}\n"

        if not self.stats['errors']:
            report += "No errors recorded.\n"

        return report

    def interactive_debug_session(self):
        """Interactive debugging session"""
        self.logger.info("🎛️  Starting interactive debug session")
        self.logger.info("Type 'help' for commands, 'quit' to exit")

        while True:
            try:
                command = input("\nradarr-debug> ").strip()

                if command == 'quit':
                    break
                elif command == 'help':
                    print("""
Available commands:
  test - Run connectivity tests
  perf <endpoint> - Performance analysis
  get <endpoint> - Make GET request
  post <endpoint> <json> - Make POST request
  stats - Show statistics
  report - Generate debug report
  websocket - Test WebSocket connection
  quit - Exit
""")
                elif command == 'test':
                    self.test_connectivity()

                elif command.startswith('perf '):
                    endpoint = command[5:]
                    self.analyze_performance(endpoint)

                elif command.startswith('get '):
                    endpoint = command[4:]
                    response = self.debug_request('GET', endpoint)
                    print(f"Status: {response.status_code}")

                elif command.startswith('post '):
                    parts = command[5:].split(' ', 1)
                    if len(parts) == 2:
                        endpoint, json_data = parts
                        data = json.loads(json_data)
                        response = self.debug_request('POST', endpoint, json=data)
                        print(f"Status: {response.status_code}")
                    else:
                        print("Usage: post <endpoint> <json>")

                elif command == 'stats':
                    print(json.dumps(self.stats, indent=2, default=str))

                elif command == 'report':
                    print(self.generate_debug_report())

                elif command == 'websocket':
                    self.debug_websocket_connection()

                else:
                    print(f"Unknown command: {command}")

            except KeyboardInterrupt:
                break
            except Exception as e:
                print(f"Error: {e}")

        self.logger.info("Debug session ended")

# Usage
if __name__ == "__main__":
    debugger = RadarrAPIDebugger("http://localhost:7878", "your-api-key", "DEBUG")

    # Run connectivity tests
    debugger.test_connectivity()

    # Performance analysis
    debugger.analyze_performance("system/status", 5)

    # Generate report
    print(debugger.generate_debug_report())

    # Interactive session
    # debugger.interactive_debug_session()
```

This completes the comprehensive Radarr Go API Troubleshooting Guide. The guide covers all major issues developers might encounter when working with the API, provides practical solutions, and includes debugging tools to help identify and resolve problems quickly.
