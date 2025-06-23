# Radarr Go

A Go-based implementation of Radarr, a movie collection manager for Usenet and BitTorrent users.

## Features

- **Movie Management**: Add, monitor, and organize your movie collection
- **API Compatibility**: RESTful API compatible with Radarr's v3 API structure
- **Multiple Databases**: Support for SQLite and PostgreSQL
- **Configurable**: YAML-based configuration with environment variable overrides
- **Docker Support**: Ready-to-run Docker containers
- **Lightweight**: Built with Go for performance and low resource usage

## Quick Start

### Using Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/radarr/radarr-go.git
cd radarr-go

# Start the application
docker-compose up -d

# View logs
docker-compose logs -f radarr-go
```

The application will be available at `http://localhost:7878`

### Building from Source

```bash
# Install dependencies
make deps

# Build the application
make build

# Run the application
make run
```

### Development Setup

```bash
# Install development tools
make setup

# Run with hot reload
make dev

# Run tests
make test

# Format and lint
make fmt lint
```

## Configuration

Configuration is handled through a YAML file (`config.yaml`) and environment variables.

### Configuration File

```yaml
server:
  port: 7878
  host: "0.0.0.0"
  url_base: ""
  enable_ssl: false

database:
  type: "sqlite"  # or "postgres"
  connection_url: "./data/radarr.db"
  max_connections: 10

log:
  level: "info"
  format: "json"
  output: "stdout"

auth:
  method: "none"
  api_key: ""

storage:
  data_directory: "./data"
  movie_directory: "./data/movies"
  backup_directory: "./data/backups"
```

### Environment Variables

All configuration options can be overridden with environment variables using the `RADARR_` prefix:

- `RADARR_SERVER_PORT=7878`
- `RADARR_DATABASE_TYPE=sqlite`
- `RADARR_LOG_LEVEL=info`
- `RADARR_AUTH_API_KEY=your-api-key`

## API Endpoints

The application implements Radarr's v3 API structure:

### System
- `GET /api/v3/system/status` - System status information

### Movies
- `GET /api/v3/movie` - Get all movies
- `GET /api/v3/movie/:id` - Get specific movie
- `POST /api/v3/movie` - Add new movie
- `PUT /api/v3/movie/:id` - Update movie
- `DELETE /api/v3/movie/:id` - Delete movie

### Movie Files
- `GET /api/v3/moviefile` - Get all movie files
- `GET /api/v3/moviefile/:id` - Get specific movie file
- `DELETE /api/v3/moviefile/:id` - Delete movie file

### Search
- `GET /api/v3/search/movie?term=query` - Search for movies

## Database Migrations

The application uses database migrations to manage schema changes:

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down
```

## Development

### Project Structure

```
radarr-go/
├── cmd/radarr/           # Main application entry point
├── internal/
│   ├── api/              # HTTP API handlers and routing
│   ├── config/           # Configuration management
│   ├── database/         # Database connectivity and migrations
│   ├── logger/           # Logging utilities
│   ├── models/           # Data models
│   └── services/         # Business logic services
├── migrations/           # Database migration files
├── web/                  # Static web assets (if any)
├── config.yaml           # Default configuration
├── docker-compose.yml    # Docker Compose setup
├── Dockerfile           # Docker image definition
└── Makefile            # Build and development tasks
```

### Adding New Features

1. **Models**: Add new data structures in `internal/models/`
2. **Services**: Implement business logic in `internal/services/`
3. **API**: Add HTTP handlers in `internal/api/`
4. **Migrations**: Create database migrations in `migrations/`

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

## Production Deployment

### Docker (Recommended)

```yaml
version: '3.8'
services:
  radarr-go:
    image: radarr/radarr-go:latest
    container_name: radarr-go
    restart: unless-stopped
    ports:
      - "7878:7878"
    volumes:
      - /path/to/data:/data
      - /path/to/movies:/movies
    environment:
      - RADARR_AUTH_API_KEY=your-secure-api-key
      - RADARR_LOG_LEVEL=info
```

### Binary Deployment

1. Build the binary: `make build-linux`
2. Copy the binary and configuration to your server
3. Create a systemd service file
4. Start and enable the service

## Migration from Original Radarr

This Go implementation aims to be API-compatible with the original Radarr. To migrate:

1. Export your current Radarr configuration and database
2. Import the data using the migration tools (coming soon)
3. Update your integrations to point to the new API endpoints

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run `make all` to ensure everything passes
6. Submit a pull request

## License

This project is licensed under the GPL-3.0 License - see the LICENSE file for details.

## Acknowledgments

- Original Radarr project and team
- Go community for excellent libraries
- Contributors and testers