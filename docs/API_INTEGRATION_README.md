# Radarr Go API Integration Documentation

Welcome to the comprehensive API integration documentation for Radarr Go - a high-performance movie collection manager built in Go with 100% Radarr v3 API compatibility.

## 📚 Documentation Overview

This documentation suite provides everything you need to integrate with the Radarr Go API, from basic usage to advanced automation patterns.

### 📖 Available Guides

| Guide | Description | Best For |
|-------|-------------|-----------|
| **[Integration Guide](integration-guide.md)** | Complete client examples and basic integration patterns | Getting started, basic integration |
| **[Integration Patterns](integration-patterns.md)** | Advanced patterns for common scenarios | Production integrations |
| **[Automation Examples](automation-examples.md)** | Complete automation scripts and workflows | DevOps, automation engineers |
| **[Troubleshooting Guide](troubleshooting-guide.md)** | Comprehensive problem-solving reference | Debugging, issue resolution |

### 🔗 Related Resources

- **[OpenAPI Specification](openapi.yaml)** - Complete API reference with 150+ endpoints
- **[Swagger UI](swagger-ui/index.html)** - Interactive API documentation
- **[Main README](../README.md)** - Project overview and setup instructions

## 🚀 Quick Start

### 1. Basic Setup

```bash
# Start Radarr Go
./radarr --data ./data --config config.yaml

# Test API connectivity
curl -H "X-API-Key: your-api-key" http://localhost:7878/api/v3/ping
```

### 2. Choose Your Language

Select the client implementation that matches your stack:

#### Python (Most Complete)

```python
from integration_guide import RadarrClient

client = RadarrClient('http://localhost:7878', 'your-api-key')
movies = client.get_movies()
```

#### JavaScript/Node.js

```javascript
const RadarrClient = require('./integration_guide');

const client = new RadarrClient('http://localhost:7878', 'your-api-key');
const movies = await client.getMovies();
```

#### Shell Scripts

```bash
./integration-guide.sh status
./integration-guide.sh search "inception"
```

#### PowerShell

```powershell
Import-Module ./RadarrGoAPI.psm1 -ArgumentList 'http://localhost:7878', 'your-api-key'
Show-RadarrStatus
```

#### Go

```go
client, _ := NewClient("http://localhost:7878", "your-api-key")
status, _ := client.GetSystemStatus(context.Background())
```

### 3. Interactive Documentation

Visit the **Swagger UI** for interactive API testing:

```text
http://localhost:7878/swagger-ui/
```

## 🎯 Common Use Cases

### Movie Management

- **Search and Add Movies**: Find movies via TMDB and add to collection
- **Bulk Operations**: Mass import/export and library management
- **Quality Management**: Automated quality profile optimization

### Monitoring and Automation

- **Real-time Updates**: WebSocket integration for live status updates
- **Health Monitoring**: Automated system health checks and alerting
- **Queue Management**: Download progress tracking and management

### External Integration

- **Media Servers**: Plex, Jellyfin, and Emby integration
- **Notification Systems**: Discord, Slack, email notifications
- **Backup/Restore**: Complete configuration and library backup

## 🏗️ Architecture Overview

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Your App      │    │   Radarr Go     │    │   External      │
│                 │    │                 │    │   Services      │
│ • Python        │────│ • REST API      │────│ • TMDB          │
│ • JavaScript    │    │ • WebSocket     │    │ • Download      │
│ • Shell         │    │ • OpenAPI       │    │   Clients       │
│ • PowerShell    │    │ • Rate Limiting │    │ • Plex/Jellyfin │
│ • Go            │    │ • Auth (API Key)│    │ • Notifications │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Key Features

- **🔄 100% Radarr v3 Compatibility** - Drop-in replacement for existing integrations
- **⚡ High Performance** - Go-based implementation with optimized database operations
- **📡 Real-time Updates** - WebSocket support for live status monitoring
- **🛡️ Built-in Security** - API key authentication with configurable rate limiting
- **📊 Comprehensive API** - 150+ endpoints covering all functionality
- **🗄️ Multi-Database** - PostgreSQL and MariaDB support with connection pooling

## 📋 Integration Checklist

### Before You Start

- [ ] Radarr Go is running and accessible
- [ ] API key is configured and working
- [ ] Network connectivity is tested
- [ ] Rate limiting considerations are understood

### Development Phase

- [ ] Choose appropriate client language/framework
- [ ] Implement proper error handling and retries
- [ ] Add logging and debugging capabilities
- [ ] Handle authentication and rate limiting
- [ ] Test with both success and failure scenarios

### Production Deployment

- [ ] Implement proper security (HTTPS, API key rotation)
- [ ] Set up monitoring and alerting
- [ ] Configure backup and disaster recovery
- [ ] Load test integration points
- [ ] Document integration for team

### Ongoing Maintenance

- [ ] Monitor API usage and performance
- [ ] Keep client libraries updated
- [ ] Review and rotate API keys periodically
- [ ] Monitor for API changes and deprecations

## 🚨 Common Pitfalls and Solutions

### Authentication Issues

```python
# ❌ Wrong header case
headers = {'X-Api-Key': 'key'}  # Should be 'X-API-Key'

# ✅ Correct authentication
headers = {'X-API-Key': 'your-api-key'}
# OR
params = {'apikey': 'your-api-key'}
```

### Rate Limiting

```python
# ❌ No rate limiting handling
for movie in movies:
    api.add_movie(movie)  # Will hit rate limits

# ✅ Proper rate limiting
from time import sleep
for i, movie in enumerate(movies):
    api.add_movie(movie)
    if i % 10 == 0:  # Pause every 10 requests
        sleep(1)
```

### Error Handling

```python
# ❌ No error handling
response = requests.get(f"{url}/api/v3/movie")
data = response.json()

# ✅ Comprehensive error handling
try:
    response = requests.get(f"{url}/api/v3/movie", timeout=30)
    response.raise_for_status()
    data = response.json()
except requests.exceptions.Timeout:
    logger.error("Request timed out")
except requests.exceptions.HTTPError as e:
    logger.error(f"HTTP error: {e.response.status_code}")
except requests.exceptions.RequestException as e:
    logger.error(f"Request failed: {e}")
```

## 🔧 Development Tools

### API Testing Tools

```bash
# cURL examples
curl -H "X-API-Key: key" http://localhost:7878/api/v3/system/status | jq

# HTTPie (more user-friendly)
http GET localhost:7878/api/v3/movie X-API-Key:your-key

# Postman collection (import openapi.yaml)
```

### Debugging Utilities

```python
# Use the built-in debugger from troubleshooting guide
from troubleshooting_guide import RadarrAPIDebugger

debugger = RadarrAPIDebugger("http://localhost:7878", "your-key", "DEBUG")
debugger.test_connectivity()
debugger.interactive_debug_session()
```

### Performance Testing

```python
# Built-in performance analysis
debugger.analyze_performance("movie", iterations=100)
```

## 📊 Monitoring and Observability

### Health Checks

```python
# Basic health check
def check_radarr_health():
    try:
        response = requests.get(f"{radarr_url}/api/v3/health",
                              headers={'X-API-Key': api_key},
                              timeout=10)
        return response.status_code == 200
    except:
        return False
```

### Metrics Collection

- **Response Times**: Track API response latency
- **Error Rates**: Monitor failed requests and error types
- **Rate Limiting**: Track rate limit usage and throttling
- **Resource Usage**: Monitor system resources on Radarr server

### Alerting Scenarios

- API endpoint becomes unavailable
- Error rate exceeds threshold
- Response time degrades significantly
- Rate limits frequently exceeded
- WebSocket connections failing

## 🔐 Security Best Practices

### API Key Management

- **Rotation**: Regularly rotate API keys (monthly/quarterly)
- **Scope**: Use separate keys for different applications
- **Storage**: Never commit keys to version control
- **Environment**: Use environment variables or secure key stores

### Network Security

- **HTTPS**: Always use HTTPS in production
- **Firewall**: Restrict API access to necessary IP ranges
- **Proxy**: Consider using reverse proxy for additional security
- **VPN**: Use VPN for remote API access

### Data Protection

- **Sensitive Data**: Never log API keys or passwords
- **Encryption**: Encrypt stored configuration data
- **Audit**: Log API access for security monitoring
- **Compliance**: Follow relevant data protection regulations

## 🚀 Performance Optimization

### Client-Side Optimization

- **Connection Pooling**: Reuse HTTP connections
- **Compression**: Enable gzip compression
- **Caching**: Cache frequently accessed data
- **Batch Operations**: Use bulk endpoints where available
- **Async Operations**: Use async/await for concurrent requests

### Server-Side Considerations

- **Database**: Optimize database queries and indexes
- **Memory**: Monitor memory usage for large operations
- **CPU**: Consider CPU usage during bulk operations
- **Network**: Monitor bandwidth usage for large transfers

## 🤝 Contributing and Support

### Getting Help

1. **Check Documentation**: Review all guides thoroughly
2. **Search Issues**: Look for similar problems in GitHub issues
3. **Create Issue**: Provide detailed reproduction steps
4. **Community**: Join community discussions and forums

### Contributing Improvements

- **Documentation**: Submit improvements to guides
- **Examples**: Add new client implementations
- **Bug Reports**: Report integration issues
- **Feature Requests**: Suggest new API features

### Code Examples Repository

All code examples in this documentation are production-ready and include:

- ✅ Comprehensive error handling
- ✅ Rate limiting management
- ✅ Proper authentication
- ✅ Detailed logging
- ✅ Performance optimization
- ✅ Security best practices

## 📄 License and Attribution

This documentation is part of the Radarr Go project and is licensed under GPL-3.0.

When using code examples in your projects:

- Attribution is appreciated but not required
- Examples are provided as-is without warranty
- Modify and adapt as needed for your use case
- Consider contributing improvements back to the project

---

## 🎉 Ready to Build?

Choose your starting point:

- **🚀 New to Radarr API?** Start with [Integration Guide](integration-guide.md)
- **🔧 Building Production System?** Jump to [Integration Patterns](integration-patterns.md)
- **🤖 Need Automation?** Check out [Automation Examples](automation-examples.md)
- **🐛 Having Issues?** Visit [Troubleshooting Guide](troubleshooting-guide.md)

**Happy building! 🎬✨**
