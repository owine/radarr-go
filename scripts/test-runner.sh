#!/bin/bash

# test-runner.sh - Comprehensive test runner for radarr-go
# This script provides various testing modes for local development and CI

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_COMPOSE_FILE="$PROJECT_ROOT/docker-compose.test.yml"

# Default values
TEST_MODE="all"
DATABASE_TYPE="auto"
SKIP_DB_START=false
KEEP_DB_RUNNING=false
PARALLEL=false
VERBOSE=false
COVERAGE=false
BENCHMARKS=false
SHORT_MODE=false
CI_MODE=false

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Test runner for radarr-go with comprehensive database and integration testing.

OPTIONS:
    -m, --mode MODE         Test mode: unit|integration|benchmark|all (default: all)
    -d, --database TYPE     Database type: postgres|mariadb|auto (default: auto)
    -s, --skip-db-start     Skip starting test databases
    -k, --keep-db           Keep databases running after tests
    -p, --parallel          Run tests in parallel
    -v, --verbose           Verbose output
    -c, --coverage          Generate coverage report
    -b, --benchmarks        Include benchmark tests
    --short                 Run in short mode (skip long-running tests)
    --ci                    CI mode (additional validations)
    -h, --help              Show this help message

EXAMPLES:
    $0                      # Run all tests with auto database selection
    $0 -m integration -d postgres -v
                            # Run integration tests with PostgreSQL, verbose output
    $0 -m benchmark -c      # Run benchmarks with coverage
    $0 --ci --parallel      # CI mode with parallel execution
    $0 -m unit --short      # Quick unit tests only

DATABASE MANAGEMENT:
    The script automatically manages test database containers unless --skip-db-start
    is specified. Test databases run on non-standard ports (15432 for PostgreSQL,
    13306 for MariaDB) to avoid conflicts.

EXIT CODES:
    0 - All tests passed
    1 - Test failures or script errors
    2 - Database setup failures
    3 - Invalid arguments
EOF
}

# Function to check if Docker is available
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is required but not installed"
        exit 2
    fi

    if ! docker info &> /dev/null; then
        print_error "Docker daemon is not running"
        exit 2
    fi
}

# Function to check if docker-compose is available
check_docker_compose() {
    if ! command -v docker-compose &> /dev/null; then
        print_error "docker-compose is required but not installed"
        exit 2
    fi
}

# Function to start test databases
start_databases() {
    if [ "$SKIP_DB_START" = true ]; then
        print_info "Skipping database startup (--skip-db-start specified)"
        return 0
    fi

    print_info "Starting test databases..."

    if [ ! -f "$TEST_COMPOSE_FILE" ]; then
        print_error "Test compose file not found: $TEST_COMPOSE_FILE"
        exit 2
    fi

    # Start databases
    if ! docker-compose -f "$TEST_COMPOSE_FILE" up -d postgres-test mariadb-test; then
        print_error "Failed to start test databases"
        exit 2
    fi

    # Wait for databases to be healthy
    print_info "Waiting for databases to be ready..."
    local max_attempts=30
    local attempt=1

    while [ $attempt -le $max_attempts ]; do
        print_info "Health check attempt $attempt/$max_attempts..."

        # Check PostgreSQL
        if docker-compose -f "$TEST_COMPOSE_FILE" exec -T postgres-test pg_isready -U radarr_test -d radarr_test &> /dev/null; then
            postgres_ready=true
        else
            postgres_ready=false
        fi

        # Check MariaDB
        if docker-compose -f "$TEST_COMPOSE_FILE" exec -T mariadb-test mysql -u radarr_test -ptest_password -e "SELECT 1" radarr_test &> /dev/null; then
            mariadb_ready=true
        else
            mariadb_ready=false
        fi

        if [ "$postgres_ready" = true ] && [ "$mariadb_ready" = true ]; then
            print_success "Test databases are ready"
            return 0
        fi

        sleep 2
        ((attempt++))
    done

    print_warning "Databases may not be fully ready, continuing anyway..."
}

# Function to stop test databases
stop_databases() {
    if [ "$KEEP_DB_RUNNING" = true ]; then
        print_info "Keeping databases running (--keep-db specified)"
        print_info "To stop databases manually: docker-compose -f $TEST_COMPOSE_FILE down -v"
        return 0
    fi

    print_info "Stopping test databases..."
    docker-compose -f "$TEST_COMPOSE_FILE" down -v
    print_success "Test databases stopped"
}

# Function to determine available databases
get_available_databases() {
    local available=""

    # Check PostgreSQL
    if docker-compose -f "$TEST_COMPOSE_FILE" ps -q postgres-test | grep -q .; then
        if docker-compose -f "$TEST_COMPOSE_FILE" exec -T postgres-test pg_isready -U radarr_test -d radarr_test &> /dev/null; then
            available="$available postgres"
        fi
    fi

    # Check MariaDB
    if docker-compose -f "$TEST_COMPOSE_FILE" ps -q mariadb-test | grep -q .; then
        if docker-compose -f "$TEST_COMPOSE_FILE" exec -T mariadb-test mysql -u radarr_test -ptest_password -e "SELECT 1" radarr_test &> /dev/null; then
            available="$available mariadb"
        fi
    fi

    echo "$available"
}

# Function to build test flags
build_test_flags() {
    local flags=""

    if [ "$VERBOSE" = true ]; then
        flags="$flags -v"
    fi

    if [ "$SHORT_MODE" = true ]; then
        flags="$flags -short"
    fi

    if [ "$COVERAGE" = true ]; then
        flags="$flags -coverprofile=coverage.out"
    fi

    if [ "$PARALLEL" = true ]; then
        flags="$flags -parallel=$(nproc 2>/dev/null || echo 4)"
    fi

    if [ "$BENCHMARKS" = true ]; then
        flags="$flags -bench=."
        flags="$flags -benchmem"
    fi

    echo "$flags"
}

# Function to run unit tests
run_unit_tests() {
    print_info "Running unit tests..."

    local flags
    flags=$(build_test_flags)

    # Run tests that don't require database
    if go test $flags ./internal/models ./internal/config ./internal/logger; then
        print_success "Unit tests passed"
        return 0
    else
        print_error "Unit tests failed"
        return 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_info "Running integration tests..."

    local flags
    flags=$(build_test_flags)

    # Set database type environment variable
    local db_env=""
    if [ "$DATABASE_TYPE" != "auto" ]; then
        db_env="RADARR_TEST_DATABASE_TYPE=$DATABASE_TYPE"
    fi

    # Run integration tests
    if env $db_env go test $flags ./internal/services ./internal/database ./internal/api; then
        print_success "Integration tests passed"
        return 0
    else
        print_error "Integration tests failed"
        return 1
    fi
}

# Function to run benchmark tests
run_benchmark_tests() {
    print_info "Running benchmark tests..."

    local flags="-bench=. -benchmem"

    if [ "$VERBOSE" = true ]; then
        flags="$flags -v"
    fi

    # Set database type environment variable
    local db_env=""
    if [ "$DATABASE_TYPE" != "auto" ]; then
        db_env="RADARR_TEST_DATABASE_TYPE=$DATABASE_TYPE"
    fi

    # Run benchmark tests
    if env $db_env go test $flags ./internal/services; then
        print_success "Benchmark tests completed"
        return 0
    else
        print_error "Benchmark tests failed"
        return 1
    fi
}

# Function to run all tests
run_all_tests() {
    local failed=false

    # Run unit tests first (fastest)
    if ! run_unit_tests; then
        failed=true
    fi

    # Run integration tests
    if ! run_integration_tests; then
        failed=true
    fi

    # Run benchmarks if requested
    if [ "$BENCHMARKS" = true ]; then
        if ! run_benchmark_tests; then
            failed=true
        fi
    fi

    if [ "$failed" = true ]; then
        return 1
    fi

    return 0
}

# Function to generate coverage report
generate_coverage_report() {
    if [ "$COVERAGE" = true ] && [ -f "coverage.out" ]; then
        print_info "Generating coverage report..."

        if command -v go &> /dev/null; then
            go tool cover -html=coverage.out -o coverage.html
            print_success "Coverage report generated: coverage.html"

            # Print coverage summary
            local coverage_pct
            coverage_pct=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
            print_info "Total coverage: $coverage_pct"
        fi
    fi
}

# Function to cleanup on exit
cleanup() {
    local exit_code=$?

    if [ $exit_code -ne 0 ]; then
        print_error "Tests failed with exit code $exit_code"
    fi

    # Generate coverage report if requested
    generate_coverage_report

    # Stop databases unless requested to keep them running
    if [ "$SKIP_DB_START" = false ]; then
        stop_databases
    fi

    exit $exit_code
}

# Main function
main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -m|--mode)
                TEST_MODE="$2"
                shift 2
                ;;
            -d|--database)
                DATABASE_TYPE="$2"
                shift 2
                ;;
            -s|--skip-db-start)
                SKIP_DB_START=true
                shift
                ;;
            -k|--keep-db)
                KEEP_DB_RUNNING=true
                shift
                ;;
            -p|--parallel)
                PARALLEL=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -c|--coverage)
                COVERAGE=true
                shift
                ;;
            -b|--benchmarks)
                BENCHMARKS=true
                shift
                ;;
            --short)
                SHORT_MODE=true
                shift
                ;;
            --ci)
                CI_MODE=true
                PARALLEL=true
                COVERAGE=true
                shift
                ;;
            -h|--help)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 3
                ;;
        esac
    done

    # Validate arguments
    if [[ ! "$TEST_MODE" =~ ^(unit|integration|benchmark|all)$ ]]; then
        print_error "Invalid test mode: $TEST_MODE"
        show_usage
        exit 3
    fi

    if [[ ! "$DATABASE_TYPE" =~ ^(postgres|mariadb|auto)$ ]]; then
        print_error "Invalid database type: $DATABASE_TYPE"
        show_usage
        exit 3
    fi

    # Setup cleanup trap
    trap cleanup EXIT INT TERM

    # Change to project root
    cd "$PROJECT_ROOT"

    # Check prerequisites
    check_docker
    check_docker_compose

    # Start databases if needed
    if [[ "$TEST_MODE" =~ ^(integration|benchmark|all)$ ]]; then
        start_databases

        # Determine available databases
        available_dbs=$(get_available_databases)
        if [ -z "$available_dbs" ]; then
            print_error "No test databases are available"
            exit 2
        fi

        print_info "Available test databases:$available_dbs"

        # Set database type if auto-detection
        if [ "$DATABASE_TYPE" = "auto" ]; then
            if echo "$available_dbs" | grep -q "postgres"; then
                DATABASE_TYPE="postgres"
                print_info "Auto-selected PostgreSQL"
            elif echo "$available_dbs" | grep -q "mariadb"; then
                DATABASE_TYPE="mariadb"
                print_info "Auto-selected MariaDB"
            fi
        fi
    fi

    # Run tests based on mode
    case $TEST_MODE in
        unit)
            run_unit_tests
            ;;
        integration)
            run_integration_tests
            ;;
        benchmark)
            run_benchmark_tests
            ;;
        all)
            run_all_tests
            ;;
    esac
}

# Run main function
main "$@"
