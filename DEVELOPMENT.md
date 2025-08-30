# Development Environment Setup Guide

This guide will help you set up a complete development environment for Radarr Go, including backend, frontend (Phase 2), databases, and monitoring tools.

## Quick Start

For experienced developers who want to get started immediately:

```bash
# Check if your environment is ready
make check-env

# Set up development tools
make setup

# Start the full development environment
make dev-full
```

## Prerequisites

### Required Tools

1. **Go** (1.23+)
   ```bash
   # macOS
   brew install go

   # Linux (Ubuntu/Debian)
   sudo apt update && sudo apt install golang-go

   # Verify installation
   go version
   ```

2. **Docker & Docker Compose**
   ```bash
   # macOS
   brew install docker docker-compose

   # Linux - Install Docker Engine
   curl -fsSL https://get.docker.com -o get-docker.sh
   sudo sh get-docker.sh

   # Verify installation
   docker --version
   docker-compose --version
   ```

3. **Node.js & npm** (for frontend development in Phase 2)
   ```bash
   # macOS
   brew install node

   # Linux (Ubuntu/Debian)
   curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
   sudo apt-get install -y nodejs

   # Verify installation
   node --version
   npm --version
   ```

### Optional Tools

- **Git** (for version control)
- **Make** (build automation)
- **curl** (API testing)

## Environment Setup

### 1. Check Your Environment

Run the environment check to see what's missing:

```bash
make check-env
```

This will show you which tools are installed and which need to be installed.

### 2. Install Development Tools

Install all Go development tools and create necessary directories:

```bash
make setup
```

This installs:
- `air` - Hot reload for Go applications
- `golangci-lint` - Go linter
- `migrate` - Database migration tool

### 3. Choose Your Development Approach

#### Option A: Full Docker Environment (Recommended)

Start everything with Docker (backend, databases, monitoring):

```bash
make dev-full
```

This starts:
- Backend with hot reload (port 7878)
- PostgreSQL database (port 5432)
- Development monitoring tools

#### Option B: Local Backend + Docker Databases

If you prefer running the backend locally but want Docker databases:

```bash
# Start test databases
make test-db-up

# In another terminal, run the backend with hot reload
make dev
```

#### Option C: Everything Local

For minimal setup without Docker:

```bash
# Install and start PostgreSQL locally
brew install postgresql
brew services start postgresql

# Create development database
createdb radarr_dev

# Run the backend
make dev
```

## Development Workflows

### Backend Development

#### Hot Reload Development
```bash
# Start with hot reload (recommended)
make dev

# Or with full Docker environment
make dev-full
```

The backend will restart automatically when you change Go files.

#### Testing
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmark tests
make test-bench

# Test specific package
go test ./internal/api -v

# Test with specific database
RADARR_DATABASE_TYPE=postgres make test
RADARR_DATABASE_TYPE=mariadb make test
```

#### Code Quality
```bash
# Format code
make fmt

# Lint code
make lint

# Run all quality checks
make all
```

### Frontend Development (Phase 2)

Currently, the frontend is in preparation phase. When React is implemented:

```bash
# Install frontend dependencies
make install-frontend

# Start frontend dev server
make dev-frontend

# Build frontend for production
make build-frontend
```

### Database Development

#### Using PostgreSQL (Default)
```bash
# Start PostgreSQL in Docker
docker-compose -f docker-compose.dev.yml up -d postgres-dev

# Or start test database
make test-db-up
```

#### Using MariaDB (Alternative)
```bash
# Start MariaDB profile
docker-compose -f docker-compose.dev.yml --profile mariadb up -d mariadb-dev

# Configure application to use MariaDB
export RADARR_DATABASE_TYPE=mariadb
make dev
```

#### Database Migrations
```bash
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
migrate create -ext sql -dir migrations add_new_feature
```

## Available Services

When running `make dev-full`, these services are available:

### Core Services
- **Backend API**: http://localhost:7878
  - Health check: http://localhost:7878/ping
  - API documentation: http://localhost:7878/api
- **Frontend** (Phase 2): http://localhost:3000
- **Database**: localhost:5432 (PostgreSQL) or localhost:3306 (MariaDB)

### Development Tools

#### Database Management
- **Adminer**: http://localhost:8081
  - Username: `radarr_dev`
  - Password: `dev_password`
  - Server: `postgres-dev`

#### Monitoring & Debugging
- **Prometheus**: http://localhost:9090 (Metrics collection)
- **Grafana**: http://localhost:3001 (Metrics visualization)
  - Username: `admin`
  - Password: `admin`
- **Jaeger**: http://localhost:16686 (Distributed tracing)
- **MailHog**: http://localhost:8025 (Email testing)

### Service Profiles

Different service combinations can be started using Docker Compose profiles:

```bash
# Start basic development (backend + database)
docker-compose -f docker-compose.dev.yml up

# Start with MariaDB instead of PostgreSQL
docker-compose -f docker-compose.dev.yml --profile mariadb up

# Start with Redis caching
docker-compose -f docker-compose.dev.yml --profile redis up

# Start with monitoring tools
docker-compose -f docker-compose.dev.yml --profile monitoring up

# Start with frontend development (Phase 2)
docker-compose -f docker-compose.dev.yml --profile frontend up

# Start everything
docker-compose -f docker-compose.dev.yml --profile mariadb --profile redis --profile monitoring --profile frontend up
```

## Configuration

### Environment Variables

The application supports these development environment variables:

```bash
# Server configuration
export RADARR_SERVER_PORT=7878
export RADARR_LOG_LEVEL=debug

# Database configuration
export RADARR_DATABASE_TYPE=postgres  # or mariadb
export RADARR_DATABASE_HOST=localhost
export RADARR_DATABASE_PORT=5432      # or 3306 for MariaDB
export RADARR_DATABASE_USERNAME=radarr_dev
export RADARR_DATABASE_PASSWORD=dev_password
export RADARR_DATABASE_NAME=radarr_dev

# Development features
export RADARR_DEV_MODE=true
export RADARR_ENABLE_PROFILING=true
```

### Configuration Files

- `config.yaml` - Main application configuration
- `.air.toml` - Hot reload configuration (auto-generated)
- `docker-compose.dev.yml` - Development environment
- `scripts/` - Database initialization and monitoring configs

## Common Development Tasks

### Adding a New API Endpoint

1. Add handler in `internal/api/handlers.go`
2. Register route in `internal/api/server.go`
3. Add tests in `internal/api/*_test.go`
4. Test with curl or Postman

```bash
# Test the new endpoint
curl -X GET http://localhost:7878/api/your-endpoint
```

### Adding a New Database Model

1. Define struct in `internal/models/`
2. Add GORM annotations
3. Create migration:
   ```bash
   migrate create -ext sql -dir migrations add_your_model
   ```
4. Apply migration:
   ```bash
   make migrate-up
   ```

### Frontend Integration (Phase 2)

When the React frontend is ready:

1. Initialize React app in `web/frontend/`
2. Configure proxy to backend API
3. Start development servers:
   ```bash
   # Terminal 1: Backend
   make dev

   # Terminal 2: Frontend
   make dev-frontend
   ```

## Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Find what's using the port
lsof -ti:7878
kill -9 <PID>

# Or use different ports
export RADARR_SERVER_PORT=7879
```

#### Database Connection Issues
```bash
# Check if database is running
docker-compose -f docker-compose.dev.yml ps

# View database logs
docker-compose -f docker-compose.dev.yml logs postgres-dev

# Reset database
make test-db-clean
make test-db-up
```

#### Hot Reload Not Working
```bash
# Reinstall air
go install github.com/cosmtrek/air@latest

# Check air configuration
cat .air.toml

# Run without air
make build && ./radarr
```

#### Docker Issues
```bash
# Clean Docker system
docker system prune -a

# Rebuild development image
docker-compose -f docker-compose.dev.yml build --no-cache

# View all containers
docker-compose -f docker-compose.dev.yml ps -a
```

### Performance Issues

#### Slow Database Queries
1. Check logs in Adminer or database container
2. Use database-specific monitoring:
   ```bash
   # PostgreSQL
   docker exec -it radarr-dev-postgres psql -U radarr_dev -c "SELECT * FROM pg_stat_activity;"

   # MariaDB
   docker exec -it radarr-dev-mariadb mysql -u radarr_dev -p -e "SHOW PROCESSLIST;"
   ```

#### Memory Issues
```bash
# Check container resource usage
docker stats

# Limit container resources in docker-compose.dev.yml
services:
  radarr-backend:
    deploy:
      resources:
        limits:
          memory: 512M
```

## Production Considerations

While this is a development guide, keep these production aspects in mind:

- Use environment-specific configurations
- Enable TLS/SSL in production
- Use production-grade databases with proper backups
- Implement proper logging and monitoring
- Use secrets management for passwords and API keys
- Configure proper resource limits

## Getting Help

- **Documentation**: Check `CLAUDE.md` for detailed project information
- **API Documentation**: Available at http://localhost:7878/docs (when implemented)
- **Database Schema**: Check `migrations/` directory
- **Configuration**: See `config.yaml` for all options

## Next Steps

Once your development environment is running:

1. Explore the API at http://localhost:7878
2. Check the database schema in Adminer
3. Run the test suite to understand the codebase
4. Start developing new features
5. Prepare for React frontend integration in Phase 2

Happy coding! 🚀
