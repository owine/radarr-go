.PHONY: build run test clean docker-build docker-run deps fmt lint

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

# Build the binary
build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PATH)

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_NAME)-linux -v $(MAIN_PATH)

# Run the application
run: build
	./$(BINARY_NAME)

# Run with hot reload using air (install with: go install github.com/cosmtrek/air@latest)
dev:
	air

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux

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

# Build Docker image
docker-build:
	docker build -t radarr-go .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# View Docker logs
docker-logs:
	docker-compose logs -f radarr-go

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
all: deps fmt lint test build

# Development setup
setup:
	$(GOGET) github.com/cosmtrek/air@latest
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	mkdir -p data movies web/static web/templates