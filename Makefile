.PHONY: build run test clean docker-build docker-run deps fmt lint \
	build-frontend dev-frontend clean-frontend install-frontend \
	build-all-with-frontend dev-full dev-env-start dev-env-stop \
	dev-env-restart dev-env-status dev-env-info dev-monitor dev-logs dev-perf

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
BINARY_NAME=radarr
MAIN_PATH=./cmd/radarr

# Frontend parameters
FRONTEND_DIR=web/frontend
NODE_CMD=npm
FRONTEND_BUILD_DIR=$(FRONTEND_DIR)/dist
STATIC_DIR=web/static

# Build variables
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS = -w -s -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(BUILD_DATE)'

# Build the binary
build:
	$(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME) -v $(MAIN_PATH)

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux -v $(MAIN_PATH)

# Build for specific platforms
build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 -v $(MAIN_PATH)

build-linux-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-linux-arm64 -v $(MAIN_PATH)

build-darwin-amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64 -v $(MAIN_PATH)

build-darwin-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64 -v $(MAIN_PATH)

build-windows-amd64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-windows-amd64.exe -v $(MAIN_PATH)

build-windows-arm64:
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-windows-arm64.exe -v $(MAIN_PATH)

build-freebsd-amd64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-freebsd-amd64 -v $(MAIN_PATH)

build-freebsd-arm64:
	CGO_ENABLED=0 GOOS=freebsd GOARCH=arm64 $(GOBUILD) -ldflags="$(LDFLAGS)" -o $(BINARY_NAME)-freebsd-arm64 -v $(MAIN_PATH)

# Build all platforms (matches CI pipeline)
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64 build-windows-arm64 build-freebsd-amd64 build-freebsd-arm64

# Build all platforms with frontend
build-all-with-frontend: build-frontend build-all

# Run the application
run: build
	./$(BINARY_NAME)

# Run with hot reload using air (install with: go install github.com/cosmtrek/air@latest)
dev-air:
	air

# Frontend Development Commands
# ===========================================

# Install frontend dependencies
install-frontend:
	@echo "Installing frontend dependencies..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && $(NODE_CMD) install; \
	else \
		echo "Frontend directory $(FRONTEND_DIR) not found. Creating placeholder structure..."; \
		mkdir -p $(FRONTEND_DIR)/src $(FRONTEND_DIR)/public $(STATIC_DIR); \
		echo "Frontend will be implemented in Phase 2"; \
	fi

# Build frontend for production
build-frontend: install-frontend
	@echo "Building frontend..."
	@if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd $(FRONTEND_DIR) && $(NODE_CMD) run build; \
	else \
		echo "Frontend not yet implemented. Creating placeholder static files..."; \
		mkdir -p $(STATIC_DIR); \
		echo "<!DOCTYPE html><html><head><title>Radarr Go</title></head><body><h1>Radarr Go - Frontend Coming Soon</h1><p>The React frontend will be available in Phase 2.</p></body></html>" > $(STATIC_DIR)/index.html; \
		echo "Frontend placeholder created in $(STATIC_DIR)/"; \
		exit 0; \
	fi
	@echo "Copying frontend build to static directory..."
	@mkdir -p $(STATIC_DIR)
	@cp -r $(FRONTEND_BUILD_DIR)/* $(STATIC_DIR)/

# Start frontend development server (legacy - use dev-frontend below)
dev-frontend-legacy:
	@echo "Starting frontend development server..."
	@if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd $(FRONTEND_DIR) && $(NODE_CMD) run dev; \
	else \
		echo "Frontend not yet implemented. Use 'make setup-frontend' to create initial structure."; \
		echo "For now, you can develop the backend with 'make dev'"; \
	fi

# Clean frontend build artifacts
clean-frontend:
	@echo "Cleaning frontend build artifacts..."
	@if [ -d "$(FRONTEND_BUILD_DIR)" ]; then rm -rf $(FRONTEND_BUILD_DIR); fi
	@if [ -d "$(STATIC_DIR)" ]; then rm -rf $(STATIC_DIR); fi
	@if [ -d "$(FRONTEND_DIR)/node_modules" ]; then rm -rf $(FRONTEND_DIR)/node_modules; fi

# Setup initial frontend structure (for Phase 2 preparation)
setup-frontend:
	@echo "Setting up frontend structure for React development..."
	mkdir -p $(FRONTEND_DIR)/src/components $(FRONTEND_DIR)/src/pages $(FRONTEND_DIR)/src/hooks
	mkdir -p $(FRONTEND_DIR)/src/services $(FRONTEND_DIR)/src/utils $(FRONTEND_DIR)/public
	mkdir -p $(STATIC_DIR)
	@echo "Frontend structure created. Ready for React implementation in Phase 2."

# Development Environment Commands (Consolidated)
# ===========================================

# Quick development with defaults (uses docker compose.override.yml automatically)
dev:
	@echo "Starting basic development environment..."
	@echo "Backend with hot reload + PostgreSQL + Admin tools"
	docker compose up --build

# Full development environment with all services
dev-full:
	@echo "Starting full development environment..."
	@echo "Backend + PostgreSQL + Database Admin + Frontend"
	docker compose --profile frontend up -d --build

# Development with specific database
dev-postgres:
	@echo "Starting development with PostgreSQL only..."
	docker compose up radarr-go postgres adminer --build

dev-mariadb:
	@echo "Starting development with MariaDB..."
	RADARR_DATABASE_TYPE=mariadb docker compose up radarr-go mariadb adminer --profile mariadb --build

# Development with monitoring stack
dev-monitoring:
	@echo "Starting development with monitoring..."
	docker compose up --profile admin --profile monitoring --build

# Frontend development
dev-frontend:
	@echo "Starting frontend development..."
	docker compose up --profile frontend --build

# Test Environment Commands (Consolidated)
# ===========================================

# Start test databases
test-db-up:
	@echo "Starting test databases..."
	docker compose up --profile test -d postgres-test mariadb-test
	@echo "Waiting for test databases to be ready..."
	@sleep 10
	@echo "Test databases should be ready!"

# Alternative: Use dedicated test override
test-env-up:
	@echo "Starting complete test environment..."
	docker compose -f docker-compose.yml -f docker-compose.test.yml.new up --profile test -d

test-db-down:
	docker compose down postgres-test mariadb-test -v

test-db-logs:
	docker compose logs -f postgres-test mariadb-test

test-db-clean:
	docker compose down --remove-orphans -v
	docker volume prune -f

# Run all tests using the comprehensive test runner
test:
	./scripts/test-runner.sh --mode all

# Run tests with coverage
test-coverage:
	./scripts/test-runner.sh --mode all --coverage

# Run benchmark tests
test-bench:
	./scripts/test-runner.sh --mode benchmark --benchmarks

# Run example tests
test-examples:
	$(GOTEST) -run Example ./...

# Run tests with PostgreSQL only
test-postgres:
	./scripts/test-runner.sh --mode all --database postgres

# Run tests with MariaDB only
test-mariadb:
	./scripts/test-runner.sh --mode all --database mariadb

# Run integration tests only
test-integration:
	./scripts/test-runner.sh --mode integration

# Run unit tests only (no database required) - legacy
test-unit-legacy:
	./scripts/test-runner.sh --mode unit

# Run tests in CI mode (parallel, with coverage)
test-ci:
	./scripts/test-runner.sh --ci

# Run quick tests (short mode)
test-quick:
	./scripts/test-runner.sh --mode unit --short

# Legacy test targets (for compatibility)
test-legacy: test-db-up
	$(GOTEST) -v ./...

test-coverage-legacy: test-db-up
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

test-bench-legacy: test-db-up
	$(GOTEST) -bench=. -benchmem ./...

# Run tests in Docker container (full isolation)
test-docker:
	docker compose -f docker-compose.test.yml up --build --abort-on-container-exit test-runner

# Run tests without database (unit tests only)
test-unit:
	$(GOTEST) -v -short ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-*
	rm -f coverage.out

# Clean everything including frontend and test databases
clean-all: clean clean-frontend test-db-clean

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOFMT) ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Production Environment Commands
# ===========================================

# Production deployment
prod-up:
	@echo "Starting production environment..."
	docker compose -f docker-compose.yml -f docker-compose.prod.yml.new up -d
	@echo "Production services started. Check logs with: make prod-logs"

prod-down:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml.new down

prod-logs:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml.new logs -f

prod-status:
	docker compose -f docker-compose.yml -f docker-compose.prod.yml.new ps

# Docker Commands (Legacy/Simple)
# ===========================================

# Build Docker image
docker-build:
	docker build -t radarr-go .

# Run with Docker Compose
docker-run:
	docker compose up -d

# Stop Docker Compose
docker-stop:
	docker compose down

# View Docker logs
docker-logs:
	docker compose logs -f radarr-go

# Docker Compose Management
# ===========================================

# Clean up all containers and volumes
docker-clean:
	@echo "Cleaning up all Docker containers and volumes..."
	docker compose down --remove-orphans -v
	docker compose -f docker-compose.dev.yml.new down --remove-orphans -v 2>/dev/null || true
	docker compose -f docker-compose.test.yml.new down --remove-orphans -v 2>/dev/null || true
	docker compose -f docker-compose.prod.yml.new down --remove-orphans -v 2>/dev/null || true
	docker system prune -f

# Show all running services
docker-ps:
	@echo "=== Main Services ==="
	docker compose ps
	@echo ""
	@echo "=== All Docker Containers ==="
	docker ps -a

# Database migrations
migrate-up:
	migrate -path migrations -database "mysql://radarr:password@tcp(localhost:3306)/radarr" up

migrate-down:
	migrate -path migrations -database "mysql://radarr:password@tcp(localhost:3306)/radarr" down

# Initialize project
init: deps
	mkdir -p data movies
	cp config.yaml data/

# All-in-one build and test
all: deps fmt lint test test-bench build

# Development workflow
dev-all: deps fmt lint test test-examples test-bench test-coverage build

# Enhanced Development Environment Management
dev-env-start:
	./scripts/dev-environment.sh start

dev-env-stop:
	./scripts/dev-environment.sh stop

dev-env-restart:
	./scripts/dev-environment.sh restart

dev-env-status:
	./scripts/dev-environment.sh status

dev-env-info:
	./scripts/dev-environment.sh info

# Development monitoring and debugging
dev-monitor:
	./scripts/dev-monitor.sh status

dev-logs:
	./scripts/dev-monitor.sh logs

dev-perf:
	./scripts/dev-monitor.sh perf

# CI/CD workflow (includes database testing)
ci: deps fmt lint test-ci build-all

# Development setup
setup: setup-backend setup-frontend

# Backend development setup
setup-backend:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	mkdir -p data movies web/static web/templates

# Check development environment
check-env:
	@echo "Checking development environment..."
	@echo "Go version: $(shell go version)"
	@which air > /dev/null && echo "Air (hot reload): ✓ Installed" || echo "Air (hot reload): ✗ Not installed (run 'make setup-backend')"
	@which golangci-lint > /dev/null && echo "golangci-lint: ✓ Installed" || echo "golangci-lint: ✗ Not installed (run 'make setup-backend')"
	@which migrate > /dev/null && echo "migrate: ✓ Installed" || echo "migrate: ✗ Not installed (run 'make setup-backend')"
	@which node > /dev/null && echo "Node.js: ✓ Installed ($(shell node --version))" || echo "Node.js: ✗ Not installed (required for frontend)"
	@which npm > /dev/null && echo "npm: ✓ Installed ($(shell npm --version))" || echo "npm: ✗ Not installed (required for frontend)"
	@which docker > /dev/null && echo "Docker: ✓ Installed" || echo "Docker: ✗ Not installed (required for development databases)"
	@which docker compose > /dev/null && echo "Docker Compose: ✓ Installed" || echo "Docker Compose: ✗ Not installed (required for development databases)"
