#!/bin/bash
# Development environment monitoring and debugging script
# Provides quick access to development tools and status checks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

print_header() {
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}"
}

# Function to check if a service is running
check_service() {
    local service=$1
    local port=$2
    local name=$3

    if curl -s -f http://localhost:$port > /dev/null 2>&1; then
        print_status $GREEN "✓ $name is running on port $port"
        return 0
    else
        print_status $RED "✗ $name is not responding on port $port"
        return 1
    fi
}

# Function to check Docker service
check_docker_service() {
    local container=$1
    local name=$2

    if docker ps --format "table {{.Names}}" | grep -q $container; then
        local status=$(docker inspect -f '{{.State.Status}}' $container 2>/dev/null)
        if [ "$status" = "running" ]; then
            print_status $GREEN "✓ $name container is running"
            return 0
        else
            print_status $YELLOW "⚠ $name container exists but status: $status"
            return 1
        fi
    else
        print_status $RED "✗ $name container not found"
        return 1
    fi
}

# Function to display service URLs
show_service_urls() {
    print_header "Development Service URLs"
    echo -e "${BLUE}Backend API:${NC}        http://localhost:7878"
    echo -e "${BLUE}API Health:${NC}         http://localhost:7878/ping"
    echo -e "${BLUE}API Documentation:${NC}  http://localhost:7878/api/v3"
    echo -e ""
    echo -e "${BLUE}Database Admin:${NC}     http://localhost:8081"
    echo -e "${BLUE}Prometheus:${NC}         http://localhost:9090"
    echo -e "${BLUE}Grafana:${NC}            http://localhost:3001 (admin/admin)"
    echo -e "${BLUE}Jaeger Tracing:${NC}     http://localhost:16686"
    echo -e "${BLUE}MailHog:${NC}            http://localhost:8025"
    echo -e ""
    echo -e "${BLUE}Frontend (Phase 2):${NC} http://localhost:3000"
    echo -e "${BLUE}Storybook (Future):${NC} http://localhost:3001"
}

# Function to check system status
check_system_status() {
    print_header "Development Environment Status"

    # Check core services
    check_service "radarr-backend" "7878" "Radarr Backend API"
    check_service "radarr-backend" "8080" "Debug/Profiling Port"

    # Check databases
    if check_docker_service "radarr-dev-postgres" "PostgreSQL"; then
        check_service "localhost" "5432" "PostgreSQL Database"
    fi

    if check_docker_service "radarr-dev-mariadb" "MariaDB"; then
        check_service "localhost" "3306" "MariaDB Database"
    fi

    # Check monitoring services
    if check_docker_service "radarr-dev-adminer" "Adminer (DB Admin)"; then
        check_service "localhost" "8081" "Database Admin Interface"
    fi

    if check_docker_service "radarr-dev-prometheus" "Prometheus"; then
        check_service "localhost" "9090" "Prometheus Metrics"
    fi

    if check_docker_service "radarr-dev-grafana" "Grafana"; then
        check_service "localhost" "3001" "Grafana Dashboard"
    fi

    if check_docker_service "radarr-dev-jaeger" "Jaeger"; then
        check_service "localhost" "16686" "Jaeger Tracing"
    fi

    if check_docker_service "radarr-dev-mailhog" "MailHog"; then
        check_service "localhost" "8025" "MailHog Email Testing"
    fi

    # Check frontend (when implemented)
    if check_docker_service "radarr-dev-frontend" "Frontend Dev Server"; then
        check_service "localhost" "3000" "React Development Server"
    fi
}

# Function to display Docker container status
show_docker_status() {
    print_header "Docker Container Status"

    if ! command -v docker >/dev/null 2>&1; then
        print_status $RED "Docker is not installed or not in PATH"
        return 1
    fi

    echo "Development containers:"
    docker ps --filter "name=radarr-dev-" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || {
        print_status $YELLOW "No development containers found"
        echo "Run 'make dev-full' to start the development environment"
    }
}

# Function to display logs
show_logs() {
    local service=${1:-"all"}

    print_header "Development Environment Logs"

    if [ "$service" = "all" ]; then
        echo "Showing logs for all services (press Ctrl+C to stop):"
        docker-compose -f docker-compose.dev.yml logs -f --tail=50
    else
        echo "Showing logs for $service (press Ctrl+C to stop):"
        docker-compose -f docker-compose.dev.yml logs -f --tail=50 $service
    fi
}

# Function to run development tests
run_dev_tests() {
    print_header "Running Development Tests"

    # Check if test databases are running
    if ! docker ps --format "table {{.Names}}" | grep -q "postgres-test\|mariadb-test"; then
        print_status $YELLOW "Starting test databases..."
        make test-db-up
        sleep 5
    fi

    # Run different types of tests
    echo -e "${BLUE}Running unit tests...${NC}"
    if make test-unit; then
        print_status $GREEN "✓ Unit tests passed"
    else
        print_status $RED "✗ Unit tests failed"
        return 1
    fi

    echo -e "${BLUE}Running integration tests...${NC}"
    if make test; then
        print_status $GREEN "✓ Integration tests passed"
    else
        print_status $RED "✗ Integration tests failed"
        return 1
    fi

    echo -e "${BLUE}Running benchmark tests...${NC}"
    if make test-bench; then
        print_status $GREEN "✓ Benchmark tests completed"
    else
        print_status $YELLOW "⚠ Benchmark tests had issues"
    fi
}

# Function to display performance metrics
show_performance() {
    print_header "Performance Metrics"

    # Check if backend is running
    if ! check_service "radarr-backend" "7878" "Backend" > /dev/null; then
        print_status $RED "Backend not running. Start with 'make dev' or 'make dev-full'"
        return 1
    fi

    echo -e "${BLUE}Memory Profile:${NC}"
    echo "Access: http://localhost:8080/debug/pprof/heap"
    echo "Command: go tool pprof http://localhost:8080/debug/pprof/heap"
    echo ""

    echo -e "${BLUE}CPU Profile:${NC}"
    echo "Access: http://localhost:8080/debug/pprof/profile"
    echo "Command: go tool pprof http://localhost:8080/debug/pprof/profile"
    echo ""

    echo -e "${BLUE}Goroutine Analysis:${NC}"
    echo "Access: http://localhost:8080/debug/pprof/goroutine"
    echo "Command: go tool pprof http://localhost:8080/debug/pprof/goroutine"
    echo ""

    # Show basic metrics if available
    if command -v curl >/dev/null 2>&1; then
        echo -e "${BLUE}Current Goroutines:${NC}"
        curl -s http://localhost:8080/debug/pprof/goroutine?debug=1 | head -5 2>/dev/null || echo "Metrics not available"
    fi
}

# Function to start specific development environment
start_environment() {
    local env_type=${1:-"full"}

    print_header "Starting Development Environment: $env_type"

    case $env_type in
        "backend-only")
            print_status $BLUE "Starting backend with hot reload..."
            make dev
            ;;
        "full")
            print_status $BLUE "Starting complete development environment..."
            make dev-full
            ;;
        "databases")
            print_status $BLUE "Starting development databases only..."
            docker-compose -f docker-compose.dev.yml up -d postgres-dev
            ;;
        "monitoring")
            print_status $BLUE "Starting with monitoring tools..."
            docker-compose -f docker-compose.dev.yml --profile monitoring up -d
            ;;
        "mariadb")
            print_status $BLUE "Starting with MariaDB instead of PostgreSQL..."
            docker-compose -f docker-compose.dev.yml --profile mariadb up -d
            ;;
        *)
            print_status $RED "Unknown environment type: $env_type"
            echo "Available types: backend-only, full, databases, monitoring, mariadb"
            return 1
            ;;
    esac
}

# Function to stop development environment
stop_environment() {
    print_header "Stopping Development Environment"

    print_status $YELLOW "Stopping all development services..."
    docker-compose -f docker-compose.dev.yml down

    if [ "$1" = "--clean" ]; then
        print_status $YELLOW "Cleaning up volumes and data..."
        docker-compose -f docker-compose.dev.yml down -v --remove-orphans
        make test-db-clean
        print_status $GREEN "Development environment cleaned"
    fi
}

# Function to display help
show_help() {
    print_header "Development Monitoring Script"

    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  status          - Check status of all development services"
    echo "  urls           - Display all service URLs"
    echo "  logs [service] - Show logs (all services or specific service)"
    echo "  docker         - Show Docker container status"
    echo "  test           - Run development tests"
    echo "  perf           - Show performance monitoring information"
    echo "  start [type]   - Start development environment"
    echo "  stop [--clean] - Stop development environment"
    echo "  help           - Show this help message"
    echo ""
    echo "Start types:"
    echo "  backend-only   - Backend with hot reload only"
    echo "  full          - Complete environment (default)"
    echo "  databases     - Databases only"
    echo "  monitoring    - With monitoring tools"
    echo "  mariadb       - Use MariaDB instead of PostgreSQL"
    echo ""
    echo "Examples:"
    echo "  $0 status                    # Check all services"
    echo "  $0 logs radarr-backend       # Show backend logs"
    echo "  $0 start monitoring          # Start with monitoring"
    echo "  $0 stop --clean              # Stop and clean volumes"
}

# Main script logic
case "${1:-status}" in
    "status")
        check_system_status
        ;;
    "urls")
        show_service_urls
        ;;
    "logs")
        show_logs "$2"
        ;;
    "docker")
        show_docker_status
        ;;
    "test")
        run_dev_tests
        ;;
    "perf")
        show_performance
        ;;
    "start")
        start_environment "$2"
        ;;
    "stop")
        stop_environment "$2"
        ;;
    "help"|"--help"|"-h")
        show_help
        ;;
    *)
        print_status $RED "Unknown command: $1"
        echo "Use '$0 help' for available commands"
        exit 1
        ;;
esac
