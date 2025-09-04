# Migration Guide for Radarr Go

## Version Migration Overview

This guide helps users navigate the versioning changes in Radarr Go and provides migration paths between different versions.

## 🚨 Critical Version Information

### Version Timeline

- **v0.0.x Series** (Deprecated): Experimental releases - discontinued
- **v0.9.0-alpha+**: Current active development - production-ready alpha
- **v1.0.0** (Future): Stable production release (Q2 2025)

### Current Recommendation

**Use v0.9.0-alpha or later** for all new installations and testing.

## Migration Scenarios

### Scenario 1: New Installation

**Recommended**: Start with v0.9.0-alpha

```bash
# Docker installation (recommended)
docker pull ghcr.io/radarr/radarr-go:v0.9.0-alpha
docker-compose -f docker-compose.yml up -d

# Or direct binary
wget https://github.com/radarr/radarr-go/releases/download/v0.9.0-alpha/radarr-linux-amd64
chmod +x radarr-linux-amd64
./radarr-linux-amd64 --data ./data
```

### Scenario 2: Upgrading from v0.0.x (Experimental Series)

**⚠️ No Automatic Migration Available**

The v0.0.x series was experimental and the architecture has significantly evolved. A fresh installation is required.

#### Migration Steps

1. **Backup Current Data** (if desired):

   ```bash
   # Backup your current data directory
   cp -r ./data ./data-v0.0.x-backup

   # Export any custom configurations
   cp config.yaml config-v0.0.x-backup.yaml
   ```

2. **Fresh Installation**:

   ```bash
   # Stop old version
   docker-compose down  # or stop your binary

   # Remove old containers/images (optional)
   docker system prune -a

   # Install v0.9.0-alpha
   docker pull ghcr.io/radarr/radarr-go:v0.9.0-alpha
   ```

3. **Reconfigure**:
   - Set up database connections (PostgreSQL recommended)
   - Reconfigure indexers, download clients, and notifications
   - Re-add movie library paths and quality profiles
   - Import movies will require re-scanning

#### What You'll Lose

- Movie history and statistics
- Download history
- Custom quality profiles (need manual recreation)
- Notification configurations (need manual recreation)

#### What You Can Preserve

- Movie files themselves (just re-scan library)
- Custom scripts and configurations (manual port)

### Scenario 3: Moving Between v0.9.x Versions

**✅ Standard Upgrade Process**

Future upgrades within the v0.9.x series will support standard migration:

```bash
# Stop current version
docker-compose down

# Pull new version
docker pull ghcr.io/radarr/radarr-go:v0.9.1-alpha

# Update docker-compose.yml with new tag
# Start new version - automatic database migration
docker-compose up -d
```

## Database Migration Strategy

### Supported Databases

- **PostgreSQL** (Recommended): Best performance and feature support
- **MariaDB**: Full compatibility with MySQL ecosystem

### Database Setup for v0.9.0-alpha+

#### PostgreSQL Setup

```bash
# Using Docker
docker run --name radarr-postgres \
  -e POSTGRES_DB=radarr \
  -e POSTGRES_USER=radarr \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -d postgres:17
```

#### MariaDB Setup

```bash
# Using Docker
docker run --name radarr-mariadb \
  -e MYSQL_DATABASE=radarr \
  -e MYSQL_USER=radarr \
  -e MYSQL_PASSWORD=password \
  -e MYSQL_ROOT_PASSWORD=rootpassword \
  -p 3306:3306 \
  -d mariadb:11
```

### Configuration Example

```yaml
server:
  port: 7878
  host: "0.0.0.0"

database:
  type: "postgres"  # or "mariadb"
  host: "localhost"
  port: 5432         # or 3306 for mariadb
  username: "radarr"
  password: "password"
  name: "radarr"

log:
  level: "info"
  format: "json"
```

## Docker Migration

### Docker Tag Strategy

Understanding the new Docker tagging strategy:

#### Production Tags (Future)

- `latest`: Latest stable release
- `stable`: Alias for latest stable
- `v1.2.3`: Specific version pinning

#### Testing Tags (Current)

- `testing`: Latest pre-release
- `v0.9.0-alpha`: Specific alpha version
- `prerelease`: Latest pre-release alias

#### Database-Specific Tags

- `postgres`: Latest with PostgreSQL optimizations
- `mariadb`: Latest with MariaDB optimizations
- `multi-db`: Supports both databases

### Docker Upgrade Process

```bash
# Current alpha users
docker pull ghcr.io/radarr/radarr-go:testing

# Pin to specific version (recommended)
docker pull ghcr.io/radarr/radarr-go:v0.9.0-alpha

# Update docker-compose.yml
services:
  radarr:
    image: ghcr.io/radarr/radarr-go:v0.9.0-alpha
    # ... rest of config
```

## Configuration Migration

### Environment Variables

The new version supports comprehensive environment variable configuration:

```bash
# Database configuration
RADARR_DATABASE_TYPE=postgres
RADARR_DATABASE_HOST=localhost
RADARR_DATABASE_PORT=5432
RADARR_DATABASE_USERNAME=radarr
RADARR_DATABASE_PASSWORD=password
RADARR_DATABASE_NAME=radarr

# Server configuration
RADARR_SERVER_PORT=7878
RADARR_SERVER_HOST=0.0.0.0

# Logging
RADARR_LOG_LEVEL=info
RADARR_LOG_FORMAT=json
```

### Configuration File Evolution

The configuration format has evolved but maintains backward compatibility where possible:

```yaml
# v0.9.0+ format
server:
  port: 7878
  host: "0.0.0.0"
  url_base: ""

database:
  type: "postgres"
  host: "localhost"
  port: 5432
  username: "radarr"
  password: "password"
  name: "radarr"
  max_connections: 10

log:
  level: "info"
  format: "json"
  output: "stdout"
```

## Performance Considerations

### Resource Requirements

v0.9.0-alpha has significantly improved performance:

| Component | v0.0.x (Est.) | v0.9.0-alpha | Improvement |
|-----------|---------------|--------------|-------------|
| Memory Usage | ~200MB | ~50-80MB | 60-75% reduction |
| Startup Time | ~10-15s | ~2-3s | 80% faster |
| API Response | Variable | <100ms avg | Consistent performance |
| Database Connections | Basic | Pooled | Optimized |

### Recommended Resources

- **CPU**: 1-2 cores minimum
- **RAM**: 512MB minimum, 1GB recommended
- **Storage**: 10GB minimum for database and logs
- **Database**: PostgreSQL recommended for best performance

## Troubleshooting Common Migration Issues

### Issue: Database Connection Failures

**Symptoms**: Cannot connect to database
**Solutions**:

```bash
# Check database is running
docker ps | grep postgres

# Verify connection settings
docker exec -it radarr-postgres psql -U radarr -d radarr -c "\dt"

# Check network connectivity
docker network ls
```

### Issue: Configuration Not Loading

**Symptoms**: Default values being used instead of config
**Solutions**:

```bash
# Verify config file path
ls -la config.yaml

# Check environment variables
env | grep RADARR_

# Validate YAML syntax
python -c "import yaml; yaml.safe_load(open('config.yaml'))"
```

### Issue: Port Conflicts

**Symptoms**: Cannot bind to port 7878
**Solutions**:

```bash
# Check what's using the port
lsof -i :7878
netstat -tulpn | grep 7878

# Use different port
RADARR_SERVER_PORT=8989 ./radarr
```

## Pre-Release Upgrade Path

### Following Alpha/Beta Releases

```bash
# Subscribe to releases
gh repo set-default radarr/radarr-go
gh release list --watch

# Upgrade to latest alpha
docker pull ghcr.io/radarr/radarr-go:testing
docker-compose down && docker-compose up -d
```

### When v1.0.0 Releases

```bash
# Upgrade to stable
docker pull ghcr.io/radarr/radarr-go:latest
docker-compose down && docker-compose up -d

# Or pin to specific version
docker pull ghcr.io/radarr/radarr-go:v1.0.0
```

## Rollback Procedures

### Rolling Back from v0.9.x

If you need to rollback (not recommended):

```bash
# Stop current version
docker-compose down

# Restore previous version (if available)
docker pull ghcr.io/radarr/radarr-go:v0.0.10

# Restore database backup
# Note: Database schemas may be incompatible
```

**⚠️ Warning**: Rollbacks may not work due to database schema changes.

## Support and Community

### Getting Help

- **GitHub Issues**: Report bugs and request features
- **Documentation**: Check README.md and VERSIONING.md
- **Discord/Community**: Links in main repository

### Before Reporting Issues

1. Check current version: `./radarr --version`
2. Review logs: `docker logs radarr-container`
3. Verify configuration: Check config.yaml and environment variables
4. Test with clean database: Isolate configuration vs. data issues

## FAQ

### Q: Why was v0.0.x deprecated?

A: The v0.0.x series was experimental. The project has evolved significantly and v0.9.0-alpha represents the true maturity level with 95% feature parity.

### Q: When will v1.0.0 be released?

A: Target is Q2 2025 after achieving 100% feature parity and completing production testing.

### Q: Can I use this in production now?

A: v0.9.0-alpha is near production-ready but still in alpha. Use in testing environments first. Production use is at your own risk.

### Q: What's the difference between PostgreSQL and MariaDB support?

A: Both are fully supported. PostgreSQL is recommended for best performance due to native Go driver and advanced features.

### Q: How do I migrate my existing Radarr (.NET) data?

A: Direct migration from original Radarr is not currently supported. This is a rewrite focused on API compatibility, not data compatibility.

---

**Last Updated**: September 2025
**Applies to**: Radarr Go v0.9.0-alpha and later

For the most current information, check the [GitHub repository](https://github.com/radarr/radarr-go).
