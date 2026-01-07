.PHONY: build run test clean docker-build docker-run deps fmt lint \
	build-frontend dev-frontend clean-frontend install-frontend \
	build-all-with-frontend dev-full dev-env-start dev-env-stop \
	dev-env-restart dev-env-status dev-env-info dev-monitor dev-logs dev-perf \
	lint-go lint-go-ci lint-frontend lint-yaml lint-json lint-markdown lint-shell \
	lint-all lint-all-parallel lint-ci-fast lint-fix \
	setup-lint-tools setup-lint-tools-ci setup-lint-tools-minimal \
	check-lint-tools lint-cache-check lint-profile lint-benchmark \
	lint-performance-report ci ci-legacy dev-all dev-all-legacy \
	all all-legacy

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

# Linting tool commands
YAMLLINT_CMD=yamllint
JSONLINT_CMD=jsonlint-php
MARKDOWNLINT_CMD=markdownlint
SHELLCHECK_CMD=shellcheck
ESLINT_CMD=npx eslint

# Linting configuration files
YAMLLINT_CONFIG=.yamllint.yml
MARKDOWNLINT_CONFIG=.markdownlint.json

# File patterns for linting
YAML_FILES=$(shell find . -name '*.yml' -o -name '*.yaml' | grep -v node_modules | grep -v vendor)
JSON_FILES=$(shell find . -name '*.json' | grep -v node_modules | grep -v vendor | grep -v '.git' | grep -v 'tsconfig' | grep -v radarr-source)
MARKDOWN_FILES=$(shell find . -name '*.md' | grep -v node_modules | grep -v vendor | grep -v radarr-source)
SHELL_FILES=$(shell find . -name '*.sh' | grep -v node_modules | grep -v vendor)

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

# Comprehensive Linting Targets
# ===========================================

# Lint Go code (requires golangci-lint)
lint-go:
	@echo "Linting Go code..."
	@which golangci-lint > /dev/null || (echo "Error: golangci-lint not found. Run 'make setup-lint-tools'" && exit 1)
	golangci-lint run

# Lint frontend TypeScript/React code
lint-frontend:
	@echo "Linting frontend code..."
	@if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		cd $(FRONTEND_DIR) && $(NODE_CMD) run lint; \
	else \
		echo "Frontend not found or package.json missing. Skipping frontend linting."; \
	fi

# Lint YAML files
lint-yaml:
	@echo "Linting YAML files..."
	@if [ -z "$(YAML_FILES)" ]; then \
		echo "No YAML files found to lint."; \
	else \
		if which $(YAMLLINT_CMD) > /dev/null 2>&1; then \
			if [ -f "$(YAMLLINT_CONFIG)" ]; then \
				$(YAMLLINT_CMD) -c $(YAMLLINT_CONFIG) $(YAML_FILES); \
			else \
				$(YAMLLINT_CMD) $(YAML_FILES); \
			fi; \
		else \
			echo "Warning: yamllint not found. Install with 'make setup-lint-tools'"; \
		fi; \
	fi

# Lint JSON files
lint-json:
	@echo "Linting JSON files..."
	@if [ -z "$(JSON_FILES)" ]; then \
		echo "No JSON files found to lint."; \
	else \
		if which $(JSONLINT_CMD) > /dev/null 2>&1; then \
			for file in $(JSON_FILES); do \
				echo "Checking $$file"; \
				$(JSONLINT_CMD) "$$file" || exit 1; \
			done; \
		else \
			echo "Warning: jsonlint not found. Install with 'make setup-lint-tools'"; \
			for file in $(JSON_FILES); do \
				echo "Checking $$file with python"; \
				python3 -m json.tool "$$file" > /dev/null || exit 1; \
			done; \
		fi; \
	fi

# Lint Markdown files
lint-markdown:
	@echo "Linting Markdown files..."
	@if [ -z "$(MARKDOWN_FILES)" ]; then \
		echo "No Markdown files found to lint."; \
	else \
		if which $(MARKDOWNLINT_CMD) > /dev/null 2>&1; then \
			if [ -f "$(MARKDOWNLINT_CONFIG)" ]; then \
				$(MARKDOWNLINT_CMD) -c $(MARKDOWNLINT_CONFIG) $(MARKDOWN_FILES); \
			else \
				$(MARKDOWNLINT_CMD) $(MARKDOWN_FILES); \
			fi; \
		else \
			echo "Warning: markdownlint not found. Install with 'make setup-lint-tools'"; \
		fi; \
	fi

# Lint shell scripts
lint-shell:
	@echo "Linting shell scripts..."
	@if [ -z "$(SHELL_FILES)" ]; then \
		echo "No shell files found to lint."; \
	else \
		if which $(SHELLCHECK_CMD) > /dev/null 2>&1; then \
			$(SHELLCHECK_CMD) $(SHELL_FILES); \
		else \
			echo "Warning: shellcheck not found. Install with 'make setup-lint-tools'"; \
		fi; \
	fi

# Fast parallel linting for CI environments (critical checks only)
lint-ci-fast:
	@echo "⚡ Running fast parallel linting for CI..."
	@# Run critical linting in parallel
	@echo "Running Go linting..."
	@make lint-go &
	@# Frontend linting (if exists)
	@if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		echo "Running frontend linting (critical)..."; \
		make lint-frontend & \
	fi
	@# Wait for critical linting
	@wait
	@echo "✅ Critical CI linting completed"

# Run all linting checks in parallel (optimal for local development)
lint-all-parallel:
	@echo "⚡ Running all linting checks in parallel..."
	@# Start all linting processes in background
	@make lint-go > /tmp/lint-go.log 2>&1 & LINT_GO_PID=$$!; \
	if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		make lint-frontend > /tmp/lint-frontend.log 2>&1 & LINT_FRONTEND_PID=$$!; \
	fi; \
	make lint-yaml > /tmp/lint-yaml.log 2>&1 & LINT_YAML_PID=$$!; \
	make lint-json > /tmp/lint-json.log 2>&1 & LINT_JSON_PID=$$!; \
	make lint-markdown > /tmp/lint-markdown.log 2>&1 & LINT_MD_PID=$$!; \
	make lint-shell > /tmp/lint-shell.log 2>&1 & LINT_SHELL_PID=$$!; \
	echo "Waiting for all linting processes to complete..."; \
	wait $$LINT_GO_PID && echo "✅ Go linting passed" || (echo "❌ Go linting failed:" && cat /tmp/lint-go.log && FAILED=1); \
	if [ ! -z "$$LINT_FRONTEND_PID" ]; then \
		wait $$LINT_FRONTEND_PID && echo "✅ Frontend linting passed" || (echo "❌ Frontend linting failed:" && cat /tmp/lint-frontend.log && FAILED=1); \
	fi; \
	wait $$LINT_YAML_PID && echo "✅ YAML linting passed" || (echo "⚠️  YAML linting failed:" && cat /tmp/lint-yaml.log); \
	wait $$LINT_JSON_PID && echo "✅ JSON linting passed" || (echo "⚠️  JSON linting failed:" && cat /tmp/lint-json.log); \
	wait $$LINT_MD_PID && echo "✅ Markdown linting passed" || (echo "⚠️  Markdown linting failed:" && cat /tmp/lint-markdown.log); \
	wait $$LINT_SHELL_PID && echo "✅ Shell linting passed" || (echo "⚠️  Shell linting failed:" && cat /tmp/lint-shell.log); \
	rm -f /tmp/lint-*.log; \
	if [ "$$FAILED" = "1" ]; then exit 1; fi
	@echo "🎉 All parallel linting checks completed!"

# Run all linting checks (sequential)
lint-all: lint-go lint-frontend lint-yaml lint-json lint-markdown lint-shell
	@echo "All linting checks completed."

# Attempt to auto-fix linting issues where possible
lint-fix:
	@echo "Attempting to auto-fix linting issues..."
	@echo "Fixing Go code formatting..."
	$(GOFMT) ./...
	@if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		echo "Fixing frontend code..."; \
		cd $(FRONTEND_DIR) && $(NODE_CMD) run lint -- --fix 2>/dev/null || echo "Frontend auto-fix not available or failed"; \
	fi
	@if which $(MARKDOWNLINT_CMD) > /dev/null 2>&1; then \
		echo "Fixing Markdown files..."; \
		$(MARKDOWNLINT_CMD) --fix $(MARKDOWN_FILES) 2>/dev/null || echo "Markdown auto-fix completed with warnings"; \
	fi
	@echo "Auto-fix completed. Please review changes and re-run 'make lint-all'"

# Smart lint selection based on environment
lint:
	@if [ "$$CI" = "true" ]; then \
		echo "🤖 CI environment detected - using fast parallel linting"; \
		make lint-ci-fast; \
	else \
		echo "💻 Local environment detected - using Go linting"; \
		make lint-go; \
	fi

# Legacy lint target (for backward compatibility)
lint-legacy: lint-go

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

# All-in-one build and test (optimized)
all: deps fmt lint-all-parallel test test-bench build

# All-in-one build and test (legacy sequential)
all-legacy: deps fmt lint-all test test-bench build

# Optimized development workflow (parallel linting)
dev-all: deps fmt lint-all-parallel test test-examples test-bench test-coverage build

# Legacy development workflow (sequential)
dev-all-legacy: deps fmt lint-all test test-examples test-bench test-coverage build

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

# Optimized CI workflow (uses fast parallel linting)
ci: deps fmt lint-ci-fast test-ci build-all

# Legacy CI workflow (sequential)
ci-legacy: deps fmt lint-all test-ci build-all

# Development setup
setup: setup-backend setup-frontend

# Backend development setup
setup-backend:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	mkdir -p data movies web/static web/templates

# Linting Tools Setup and Verification (Performance Optimized)
# ===========================================

# Fast parallel installation for CI environments
setup-lint-tools-ci:
	@echo "⚡ Installing critical linting tools for CI (parallel)..."
	@mkdir -p ~/.local/bin
	@# Install Go tools in parallel
	@echo "Installing Go linting tools in parallel..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest &
	@# Install Python tools (essential only)
	@echo "Installing Python yamllint..."
	@if which pip3 > /dev/null 2>&1; then \
		pip3 install --user --no-cache-dir yamllint & \
	elif which pip > /dev/null 2>&1; then \
		pip install --user --no-cache-dir yamllint & \
	fi
	@# Install Node.js tools (essential only)
	@echo "Installing Node.js markdownlint-cli..."
	@if which npm > /dev/null 2>&1; then \
		npm install -g --no-audit --no-fund markdownlint-cli & \
	fi
	@# Install shellcheck via package manager
	@echo "Installing shellcheck..."
	@if which apt-get > /dev/null 2>&1; then \
		sudo apt-get update -qq && sudo apt-get install -y -qq shellcheck & \
	elif which yum > /dev/null 2>&1; then \
		sudo yum install -y -q ShellCheck & \
	elif which apk > /dev/null 2>&1; then \
		sudo apk add --no-cache shellcheck & \
	fi
	@echo "Waiting for all installations to complete..."
	@wait
	@echo "✅ CI linting tools installation completed!"

# Full installation for local development (comprehensive)
setup-lint-tools:
	@echo "Installing linting tools for local development..."
	@echo "Installing Go linting tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installing YAML linting tools..."
	@if which pip3 > /dev/null 2>&1; then \
		pip3 install yamllint; \
	elif which pip > /dev/null 2>&1; then \
		pip install yamllint; \
	else \
		echo "Warning: pip/pip3 not found. Please install yamllint manually: pip install yamllint"; \
	fi
	@echo "Installing JSON linting tools..."
	@if which brew > /dev/null 2>&1; then \
		brew install jsonlint; \
	elif which apt-get > /dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y jsonlint; \
	elif which yum > /dev/null 2>&1; then \
		sudo yum install -y nodejs npm && sudo npm install -g jsonlint; \
	else \
		echo "Warning: Package manager not found. Using Python json.tool as fallback"; \
	fi
	@echo "Installing Markdown linting tools..."
	@if which npm > /dev/null 2>&1; then \
		npm install -g markdownlint-cli; \
	else \
		echo "Warning: npm not found. Please install Node.js and npm first"; \
	fi
	@echo "Installing Shell linting tools..."
	@if which brew > /dev/null 2>&1; then \
		brew install shellcheck; \
	elif which apt-get > /dev/null 2>&1; then \
		sudo apt-get update && sudo apt-get install -y shellcheck; \
	elif which yum > /dev/null 2>&1; then \
		sudo yum install -y ShellCheck; \
	else \
		echo "Warning: Package manager not found. Please install shellcheck manually"; \
	fi
	@echo "Linting tools installation completed!"
	@echo "Run 'make check-lint-tools' to verify installation"

# Check if all linting tools are installed
check-lint-tools:
	@echo "Checking linting tools installation..."
	@which golangci-lint > /dev/null && echo "golangci-lint: ✓ Installed" || echo "golangci-lint: ✗ Not installed"
	@which $(YAMLLINT_CMD) > /dev/null && echo "yamllint: ✓ Installed" || echo "yamllint: ✗ Not installed"
	@which $(JSONLINT_CMD) > /dev/null && echo "jsonlint: ✓ Installed" || echo "jsonlint: ✗ Not installed (falling back to python json.tool)"
	@which $(MARKDOWNLINT_CMD) > /dev/null && echo "markdownlint: ✓ Installed" || echo "markdownlint: ✗ Not installed"
	@which $(SHELLCHECK_CMD) > /dev/null && echo "shellcheck: ✓ Installed" || echo "shellcheck: ✗ Not installed"
	@echo "Linting tools check completed."

# Performance Analysis and Profiling
# ===========================================

# Profile linting performance to identify bottlenecks
lint-profile:
	@echo "🔍 Profiling linting performance..."
	@mkdir -p tmp/profile
	@# Time individual linting steps
	@echo "Timing Go linting..." && time make lint-go > tmp/profile/go.log 2>&1 || true
	@if [ -d "$(FRONTEND_DIR)" ] && [ -f "$(FRONTEND_DIR)/package.json" ]; then \
		echo "Timing frontend linting..." && time make lint-frontend > tmp/profile/frontend.log 2>&1 || true; \
	fi
	@echo "Timing YAML linting..." && time make lint-yaml > tmp/profile/yaml.log 2>&1 || true
	@echo "Timing JSON linting..." && time make lint-json > tmp/profile/json.log 2>&1 || true
	@echo "Timing Markdown linting..." && time make lint-markdown > tmp/profile/markdown.log 2>&1 || true
	@echo "Timing Shell linting..." && time make lint-shell > tmp/profile/shell.log 2>&1 || true
	@echo "📊 Linting performance profile complete. Check tmp/profile/ for detailed logs."

# Benchmark parallel vs sequential linting
lint-benchmark:
	@echo "🏁 Benchmarking linting approaches..."
	@mkdir -p tmp/benchmark
	@echo "Testing sequential linting..."
	@time make lint-all > tmp/benchmark/sequential.log 2>&1 || echo "Sequential completed with errors"
	@echo "Testing parallel linting..."
	@time make lint-all-parallel > tmp/benchmark/parallel.log 2>&1 || echo "Parallel completed with errors"
	@echo "Testing CI fast linting..."
	@time make lint-ci-fast > tmp/benchmark/ci-fast.log 2>&1 || echo "CI fast completed with errors"
	@echo "📊 Benchmark complete. Check tmp/benchmark/ for results."

# Verify lint tools are cached and available
lint-cache-check:
	@echo "🔍 Checking lint tool cache status..."
	@which golangci-lint > /dev/null && echo "✅ golangci-lint: cached" || echo "❌ golangci-lint: not cached"
	@which yamllint > /dev/null && echo "✅ yamllint: cached" || echo "❌ yamllint: not cached"
	@which markdownlint > /dev/null && echo "✅ markdownlint: cached" || echo "❌ markdownlint: not cached"
	@which shellcheck > /dev/null && echo "✅ shellcheck: cached" || echo "❌ shellcheck: not cached"
	@ls -la ~/.local/bin/ 2>/dev/null | head -5 || echo "Local bin directory empty"
	@ls -la ~/go/bin/golangci-lint 2>/dev/null || echo "golangci-lint not in ~/go/bin/"

# Fast CI tool installation with aggressive caching
setup-lint-tools-minimal:
	@echo "⚡ Minimal linting tools for CI (essential only)..."
	@# Only install absolutely critical tools
	@if [ ! -f "$(shell go env GOPATH)/bin/golangci-lint" ]; then \
		echo "Installing golangci-lint..."; \
		$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	else \
		echo "✅ golangci-lint already cached"; \
	fi
	@# Skip non-critical tools in CI-minimal mode
	@echo "✅ Minimal CI linting tools ready"

# Run comprehensive linting performance analysis
lint-performance-report:
	@echo "📊 Running comprehensive linting performance analysis..."
	@chmod +x ./scripts/lint-performance-report.sh
	@./scripts/lint-performance-report.sh
	@echo "✅ Performance analysis complete"

# Check development environment
check-env: check-lint-tools
	@echo "Checking development environment..."
	@echo "Go version: $(shell go version)"
	@which air > /dev/null && echo "Air (hot reload): ✓ Installed" || echo "Air (hot reload): ✗ Not installed (run 'make setup-backend')"
	@which migrate > /dev/null && echo "migrate: ✓ Installed" || echo "migrate: ✗ Not installed (run 'make setup-backend')"
	@which node > /dev/null && echo "Node.js: ✓ Installed ($(shell node --version))" || echo "Node.js: ✗ Not installed (required for frontend)"
	@which npm > /dev/null && echo "npm: ✓ Installed ($(shell npm --version))" || echo "npm: ✗ Not installed (required for frontend)"
	@which docker > /dev/null && echo "Docker: ✓ Installed" || echo "Docker: ✗ Not installed (required for development databases)"
	@which docker compose > /dev/null && echo "Docker Compose: ✓ Installed" || echo "Docker Compose: ✗ Not installed (required for development databases)"
