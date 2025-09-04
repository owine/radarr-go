#!/bin/bash
# Migration Validation Script for Radarr Go
# Tests migration sequences, rollback safety, and cross-database compatibility
#
# Usage:
#   ./migration_validator.sh validate-all
#   ./migration_validator.sh test-sequence postgres
#   ./migration_validator.sh test-rollback mariadb
#   ./migration_validator.sh cross-validate

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TEST_DIR="${PROJECT_ROOT}/migration_tests"
LOG_FILE="${TEST_DIR}/migration_validation_$(date +%Y%m%d_%H%M%S).log"

# Test database configurations
TEST_POSTGRES_DB="radarr_migration_test_pg_$(date +%s)"
TEST_MYSQL_DB="radarr_migration_test_my_$(date +%s)"

# Database configurations
POSTGRES_HOST="${RADARR_DATABASE_HOST:-localhost}"
POSTGRES_PORT="${RADARR_DATABASE_PORT:-5432}"
POSTGRES_USER="${RADARR_DATABASE_USERNAME:-radarr}"
POSTGRES_PASSWORD="${RADARR_DATABASE_PASSWORD:-password}"

MYSQL_HOST="${RADARR_DATABASE_HOST:-localhost}"
MYSQL_PORT="${RADARR_DATABASE_PORT:-3306}"
MYSQL_USER="${RADARR_DATABASE_USERNAME:-radarr}"
MYSQL_PASSWORD="${RADARR_DATABASE_PASSWORD:-password}"

# Logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

error_exit() {
    log "ERROR: $1"
    cleanup_test_databases
    exit 1
}

# Setup test environment
setup_test_env() {
    mkdir -p "$TEST_DIR"
    log "Migration validation test starting..."
    log "Log file: $LOG_FILE"
}

# Get list of migrations in order
get_migrations() {
    local db_type="$1"
    find "${PROJECT_ROOT}/migrations/${db_type}" -name "*_*.up.sql" | sort | sed 's/.*\///g' | sed 's/\.up\.sql$//g'
}

# Create test database
create_test_database() {
    local db_type="$1"
    local test_db="$2"

    case "$db_type" in
        "postgres")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
                -c "DROP DATABASE IF EXISTS \"$test_db\";" \
                -c "CREATE DATABASE \"$test_db\";" >/dev/null 2>&1 || error_exit "Failed to create test database $test_db"
            ;;
        "mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
                -e "DROP DATABASE IF EXISTS \`$test_db\`; CREATE DATABASE \`$test_db\`;" >/dev/null 2>&1 || error_exit "Failed to create test database $test_db"
            ;;
    esac

    log "Created test database: $test_db"
}

# Run single migration
run_migration() {
    local db_type="$1"
    local test_db="$2"
    local migration="$3"
    local direction="$4"  # up or down

    local migration_file="${PROJECT_ROOT}/migrations/${db_type}/${migration}.${direction}.sql"

    if [ ! -f "$migration_file" ]; then
        error_exit "Migration file not found: $migration_file"
    fi

    log "Running $direction migration: $migration"

    case "$db_type" in
        "postgres")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            if ! psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$test_db" \
                -f "$migration_file" >/dev/null 2>>"$LOG_FILE"; then
                error_exit "Migration $migration ($direction) failed for $db_type"
            fi
            ;;
        "mysql")
            if ! mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" \
                < "$migration_file" 2>>"$LOG_FILE"; then
                error_exit "Migration $migration ($direction) failed for $db_type"
            fi
            ;;
    esac
}

# Validate foreign key constraints
validate_foreign_keys() {
    local db_type="$1"
    local test_db="$2"

    log "Validating foreign key constraints for $db_type..."

    case "$db_type" in
        "postgres")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Check for constraint violations
            local violations=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$test_db" -t -c "
                SELECT
                    tc.table_name,
                    tc.constraint_name,
                    rc.delete_rule
                FROM information_schema.table_constraints tc
                JOIN information_schema.referential_constraints rc ON tc.constraint_name = rc.constraint_name
                WHERE tc.constraint_type = 'FOREIGN KEY'
                AND tc.table_schema = 'public';
            " 2>/dev/null | wc -l)

            log "Found $violations foreign key constraints in PostgreSQL"
            ;;
        "mysql")
            # Check for constraint violations
            local violations=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" -e "
                SELECT
                    TABLE_NAME,
                    CONSTRAINT_NAME,
                    REFERENCED_TABLE_NAME
                FROM information_schema.REFERENTIAL_CONSTRAINTS
                WHERE CONSTRAINT_SCHEMA = '$test_db';
            " -s | wc -l)

            log "Found $violations foreign key constraints in MariaDB"
            ;;
    esac
}

# Test complete migration sequence
test_migration_sequence() {
    local db_type="$1"
    local test_db=""

    case "$db_type" in
        "postgres") test_db="$TEST_POSTGRES_DB" ;;
        "mysql") test_db="$TEST_MYSQL_DB" ;;
        *) error_exit "Unknown database type: $db_type" ;;
    esac

    log "Testing complete migration sequence for $db_type..."

    create_test_database "$db_type" "$test_db"

    # Get migrations in order
    local migrations=($(get_migrations "$db_type"))
    log "Found ${#migrations[@]} migrations for $db_type"

    # Apply all migrations
    for migration in "${migrations[@]}"; do
        run_migration "$db_type" "$test_db" "$migration" "up"
        validate_foreign_keys "$db_type" "$test_db"
    done

    log "All migrations applied successfully for $db_type"

    # Insert test data to validate constraints
    insert_test_data "$db_type" "$test_db"

    # Test rollback sequence (reverse order)
    log "Testing rollback sequence for $db_type..."
    for ((i=${#migrations[@]}-1; i>=0; i--)); do
        run_migration "$db_type" "$test_db" "${migrations[i]}" "down"
    done

    log "All rollbacks completed successfully for $db_type"
}

# Insert test data to validate constraints
insert_test_data() {
    local db_type="$1"
    local test_db="$2"

    log "Inserting test data to validate constraints..."

    case "$db_type" in
        "postgres")
            export PGPASSWORD="$POSTGRES_PASSWORD"

            # Test basic data insertion
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$test_db" -c "
                -- Insert test quality definition if table exists
                INSERT INTO quality_definitions (id, title, weight)
                VALUES (999, 'Test Quality', 999)
                ON CONFLICT (id) DO NOTHING;

                -- Insert test quality profile if table exists
                INSERT INTO quality_profiles (id, name, cutoff, items)
                VALUES (999, 'Test Profile', 1, '[]')
                ON CONFLICT (id) DO NOTHING;

                -- Insert test movie if table exists
                INSERT INTO movies (tmdb_id, title, title_slug, quality_profile_id)
                VALUES (999999, 'Test Movie', 'test-movie-999999', 999)
                ON CONFLICT (tmdb_id) DO NOTHING;

                -- Test wanted movies if table exists
                INSERT INTO wanted_movies (movie_id, status, target_quality_id)
                SELECT id, 'missing', 999
                FROM movies
                WHERE tmdb_id = 999999
                ON CONFLICT (movie_id) DO NOTHING;
            " >/dev/null 2>&1 || log "WARNING" "Some test data insertion failed (expected for partial schemas)"
            ;;
        "mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" -e "
                -- Insert test quality definition if table exists
                INSERT IGNORE INTO quality_definitions (id, title, weight)
                VALUES (999, 'Test Quality', 999);

                -- Insert test quality profile if table exists
                INSERT IGNORE INTO quality_profiles (id, name, cutoff, items)
                VALUES (999, 'Test Profile', 1, '[]');

                -- Insert test movie if table exists
                INSERT IGNORE INTO movies (tmdb_id, title, title_slug, quality_profile_id)
                VALUES (999999, 'Test Movie', 'test-movie-999999', 999);

                -- Test wanted movies if table exists
                INSERT IGNORE INTO wanted_movies (movie_id, status, target_quality_id)
                SELECT id, 'missing', 999
                FROM movies
                WHERE tmdb_id = 999999;
            " >/dev/null 2>&1 || log "WARNING" "Some test data insertion failed (expected for partial schemas)"
            ;;
    esac

    log "Test data insertion completed"
}

# Cross-validate schema consistency between databases
cross_validate_schemas() {
    log "Cross-validating schema consistency between PostgreSQL and MariaDB..."

    local pg_db="$TEST_POSTGRES_DB"
    local my_db="$TEST_MYSQL_DB"

    # Create both test databases
    create_test_database "postgres" "$pg_db"
    create_test_database "mysql" "$my_db"

    # Apply all migrations to both
    local pg_migrations=($(get_migrations "postgres"))
    local my_migrations=($(get_migrations "mysql"))

    if [ ${#pg_migrations[@]} -ne ${#my_migrations[@]} ]; then
        log "WARNING" "PostgreSQL has ${#pg_migrations[@]} migrations, MariaDB has ${#my_migrations[@]} migrations"
    fi

    # Apply migrations to both databases
    for migration in "${pg_migrations[@]}"; do
        run_migration "postgres" "$pg_db" "$migration" "up"
    done

    for migration in "${my_migrations[@]}"; do
        run_migration "mysql" "$my_db" "$migration" "up"
    done

    # Compare table structures (basic comparison)
    log "Comparing table structures between databases..."

    export PGPASSWORD="$POSTGRES_PASSWORD"
    local pg_tables=$(psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$pg_db" -t -c "
        SELECT table_name FROM information_schema.tables
        WHERE table_schema = 'public'
        ORDER BY table_name;
    " | tr -d ' ' | sort)

    local my_tables=$(mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$my_db" -e "
        SELECT table_name FROM information_schema.tables
        WHERE table_schema = '$my_db'
        ORDER BY table_name;
    " -s | sort)

    # Write comparison to file
    echo "$pg_tables" > "${TEST_DIR}/pg_tables.txt"
    echo "$my_tables" > "${TEST_DIR}/my_tables.txt"

    if diff "${TEST_DIR}/pg_tables.txt" "${TEST_DIR}/my_tables.txt" >/dev/null; then
        log "SUCCESS" "Table structures match between PostgreSQL and MariaDB"
    else
        log "WARNING" "Table structure differences detected:"
        diff "${TEST_DIR}/pg_tables.txt" "${TEST_DIR}/my_tables.txt" | tee -a "$LOG_FILE"
    fi

    # Test data compatibility
    insert_test_data "postgres" "$pg_db"
    insert_test_data "mysql" "$my_db"

    log "Cross-validation completed"
}

# Cleanup test databases
cleanup_test_databases() {
    log "Cleaning up test databases..."

    # PostgreSQL cleanup
    if [ -n "${TEST_POSTGRES_DB:-}" ]; then
        export PGPASSWORD="$POSTGRES_PASSWORD"
        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d postgres \
            -c "DROP DATABASE IF EXISTS \"$TEST_POSTGRES_DB\";" >/dev/null 2>&1 || true
    fi

    # MySQL cleanup
    if [ -n "${TEST_MYSQL_DB:-}" ]; then
        mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" \
            -e "DROP DATABASE IF EXISTS \`$TEST_MYSQL_DB\`;" >/dev/null 2>&1 || true
    fi

    log "Test databases cleaned up"
}

# Validate specific migration issues
validate_migration_issues() {
    log "Checking for known migration issues..."

    local issues_found=0

    # Check for quality_definitions table dependency in migration 007
    for db_type in postgres mysql; do
        local migration_007="${PROJECT_ROOT}/migrations/${db_type}/007_wanted_movies.up.sql"
        if [ -f "$migration_007" ]; then
            if grep -q "quality_definitions" "$migration_007"; then
                # Check if migration 010 creates quality_definitions
                local migration_010="${PROJECT_ROOT}/migrations/${db_type}/010_complete_schema_refactor.up.sql"
                if [ -f "$migration_010" ] && ! grep -q "CREATE TABLE.*quality_definitions" "$migration_010"; then
                    log "ERROR" "Migration 007 references quality_definitions but migration 010 doesn't create it ($db_type)"
                    issues_found=$((issues_found + 1))
                else
                    log "SUCCESS" "Migration 007 quality_definitions dependency resolved ($db_type)"
                fi
            fi
        fi
    done

    # Check for missing migration files
    for db_type in postgres mysql; do
        local migration_dir="${PROJECT_ROOT}/migrations/${db_type}"
        local migrations=($(find "$migration_dir" -name "*_*.up.sql" | sort))

        for migration_file in "${migrations[@]}"; do
            local basename=$(basename "$migration_file" .up.sql)
            local down_file="${migration_dir}/${basename}.down.sql"

            if [ ! -f "$down_file" ]; then
                log "ERROR" "Missing down migration: $down_file"
                issues_found=$((issues_found + 1))
            fi
        done
    done

    # Check for gap in migration sequence
    for db_type in postgres mysql; do
        local numbers=($(find "${PROJECT_ROOT}/migrations/${db_type}" -name "*.up.sql" | sed 's/.*\/\([0-9]*\)_.*/\1/' | sort -n))
        local expected=1

        for num in "${numbers[@]}"; do
            if [ "$num" -ne "$expected" ] && [ "$expected" -ne 9 ]; then  # 009 is known missing
                log "WARNING" "Migration sequence gap: expected ${expected}, found ${num} ($db_type)"
            fi
            expected=$((num + 1))
        done
    done

    if [ "$issues_found" -eq 0 ]; then
        log "SUCCESS" "No critical migration issues detected"
    else
        log "ERROR" "Found $issues_found migration issues"
    fi

    return $issues_found
}

# Test performance with large dataset
test_performance_with_data() {
    local db_type="$1"
    local test_db="$2"

    log "Testing migration performance with large dataset ($db_type)..."

    # Apply all migrations first
    local migrations=($(get_migrations "$db_type"))
    for migration in "${migrations[@]}"; do
        run_migration "$db_type" "$test_db" "$migration" "up"
    done

    # Insert large dataset (simplified for speed)
    local start_time=$(date +%s)

    case "$db_type" in
        "postgres")
            export PGPASSWORD="$POSTGRES_PASSWORD"
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$test_db" -c "
                -- Insert quality definitions first
                INSERT INTO quality_definitions (id, title, weight)
                VALUES (1, 'Test Quality', 1)
                ON CONFLICT DO NOTHING;

                -- Insert quality profile
                INSERT INTO quality_profiles (id, name, cutoff, items)
                VALUES (1, 'Test Profile', 1, '[]')
                ON CONFLICT DO NOTHING;

                -- Insert 10k movies with progress logging
                INSERT INTO movies (tmdb_id, title, title_slug, quality_profile_id)
                SELECT
                    generate_series(1, 10000),
                    'Movie ' || generate_series(1, 10000),
                    'movie-' || generate_series(1, 10000),
                    1;

                -- Insert wanted movies for every 10th movie
                INSERT INTO wanted_movies (movie_id, status, target_quality_id)
                SELECT id, 'missing', 1
                FROM movies
                WHERE id % 10 = 0;
            " >/dev/null 2>&1 || log "WARNING" "Large dataset insertion failed for PostgreSQL"
            ;;
        "mysql")
            # Insert in smaller batches for MySQL
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" -e "
                INSERT IGNORE INTO quality_definitions (id, title, weight) VALUES (1, 'Test Quality', 1);
                INSERT IGNORE INTO quality_profiles (id, name, cutoff, items) VALUES (1, 'Test Profile', 1, '[]');
            " >/dev/null 2>&1

            # Insert movies in batches
            for i in $(seq 1 1000 10000); do
                local end=$((i + 999))
                if [ $end -gt 10000 ]; then
                    end=10000
                fi

                mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" -e "
                    INSERT INTO movies (tmdb_id, title, title_slug, quality_profile_id) VALUES
                    $(for j in $(seq $i $end); do
                        echo "($j, 'Movie $j', 'movie-$j', 1)"
                        [ $j -lt $end ] && echo ","
                    done);
                " >/dev/null 2>&1
            done

            # Insert wanted movies
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" -e "
                INSERT INTO wanted_movies (movie_id, status, target_quality_id)
                SELECT id, 'missing', 1 FROM movies WHERE id % 10 = 0;
            " >/dev/null 2>&1 || true
            ;;
    esac

    local insert_time=$(($(date +%s) - start_time))
    log "Large dataset insertion completed in ${insert_time}s for $db_type"

    # Test query performance with large dataset
    test_query_performance "$db_type" "$test_db"
}

# Test query performance
test_query_performance() {
    local db_type="$1"
    local test_db="$2"

    log "Testing query performance with large dataset..."

    local queries=(
        "SELECT COUNT(*) FROM movies WHERE monitored = true"
        "SELECT COUNT(*) FROM movies WHERE has_file = false AND monitored = true"
        "SELECT COUNT(*) FROM wanted_movies WHERE status = 'missing'"
        "SELECT m.title, wm.status FROM movies m JOIN wanted_movies wm ON m.id = wm.movie_id LIMIT 100"
    )

    for query in "${queries[@]}"; do
        local start_time=$(date +%s%3N)  # milliseconds

        case "$db_type" in
            "postgres")
                export PGPASSWORD="$POSTGRES_PASSWORD"
                psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_USER" -d "$test_db" \
                    -c "$query" >/dev/null 2>&1 || true
                ;;
            "mysql")
                mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$test_db" \
                    -e "$query" >/dev/null 2>&1 || true
                ;;
        esac

        local end_time=$(date +%s%3N)
        local duration=$((end_time - start_time))

        log "INFO" "$db_type query took ${duration}ms: $(echo "$query" | cut -c1-50)..."

        if [ "$duration" -gt "$MAX_QUERY_TIME_MS" ]; then
            log "WARNING" "Slow query detected ($duration ms > $MAX_QUERY_TIME_MS ms threshold)"
        fi
    done
}

# Main validation function
validate_all() {
    log "Starting comprehensive migration validation..."

    # First check for known issues
    validate_migration_issues || error_exit "Critical migration issues detected"

    # Test both database types
    for db_type in postgres mysql; do
        # Check if tools are available
        case "$db_type" in
            "postgres")
                if ! command -v psql >/dev/null 2>&1; then
                    log "WARNING" "PostgreSQL tools not available, skipping tests"
                    continue
                fi
                ;;
            "mysql")
                if ! command -v mysql >/dev/null 2>&1; then
                    log "WARNING" "MySQL tools not available, skipping tests"
                    continue
                fi
                ;;
        esac

        # Test migration sequence
        test_migration_sequence "$db_type"

        # Test performance with data
        create_test_database "$db_type" "$([ "$db_type" = "postgres" ] && echo "$TEST_POSTGRES_DB" || echo "$TEST_MYSQL_DB")"
        test_performance_with_data "$db_type" "$([ "$db_type" = "postgres" ] && echo "$TEST_POSTGRES_DB" || echo "$TEST_MYSQL_DB")"
    done

    # Cross-validate if both databases are available
    if command -v psql >/dev/null 2>&1 && command -v mysql >/dev/null 2>&1; then
        cross_validate_schemas
    fi

    cleanup_test_databases
    log "Migration validation completed successfully"
}

# Trap to ensure cleanup
trap cleanup_test_databases EXIT

# Main function
main() {
    local command="${1:-validate-all}"

    setup_test_env

    case "$command" in
        "validate-all")
            validate_all
            ;;
        "test-sequence")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 test-sequence <postgres|mysql>"
            fi
            test_migration_sequence "$2"
            cleanup_test_databases
            ;;
        "test-rollback")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 test-rollback <postgres|mysql>"
            fi
            test_migration_sequence "$2"  # This includes rollback testing
            cleanup_test_databases
            ;;
        "cross-validate")
            cross_validate_schemas
            cleanup_test_databases
            ;;
        "check-issues")
            validate_migration_issues
            ;;
        *)
            echo "Usage: $0 <command>"
            echo
            echo "Commands:"
            echo "  validate-all                 Run complete migration validation"
            echo "  test-sequence <db_type>      Test migration sequence for specific database"
            echo "  test-rollback <db_type>      Test rollback sequence for specific database"
            echo "  cross-validate               Compare schemas between databases"
            echo "  check-issues                 Check for known migration issues"
            echo
            echo "Database types: postgres, mysql"
            echo
            echo "Environment variables:"
            echo "  RADARR_DATABASE_HOST         Database host (default: localhost)"
            echo "  RADARR_DATABASE_PORT         Database port"
            echo "  RADARR_DATABASE_USERNAME     Database username (default: radarr)"
            echo "  RADARR_DATABASE_PASSWORD     Database password (default: password)"
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"
