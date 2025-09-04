#!/bin/bash
# Database User Management Script for Radarr Go
# Manages database users, permissions, and access control with least privilege principle
#
# Usage:
#   ./user_management.sh create-app-user postgresql
#   ./user_management.sh create-readonly-user postgresql monitoring_user
#   ./user_management.sh audit-permissions postgresql
#   ./user_management.sh rotate-password postgresql radarr_app

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
SECURITY_DIR="${PROJECT_ROOT}/security"
LOG_FILE="${SECURITY_DIR}/user_management_$(date +%Y%m%d_%H%M%S).log"

# Database configurations
POSTGRES_HOST="${RADARR_DATABASE_HOST:-localhost}"
POSTGRES_PORT="${RADARR_DATABASE_PORT:-5432}"
POSTGRES_ADMIN_USER="${RADARR_DATABASE_ADMIN_USER:-postgres}"
POSTGRES_ADMIN_PASSWORD="${RADARR_DATABASE_ADMIN_PASSWORD:-postgres}"
POSTGRES_DB="${RADARR_DATABASE_NAME:-radarr}"

MYSQL_HOST="${RADARR_DATABASE_HOST:-localhost}"
MYSQL_PORT="${RADARR_DATABASE_PORT:-3306}"
MYSQL_ADMIN_USER="${RADARR_DATABASE_ADMIN_USER:-root}"
MYSQL_ADMIN_PASSWORD="${RADARR_DATABASE_ADMIN_PASSWORD:-password}"
MYSQL_DB="${RADARR_DATABASE_NAME:-radarr}"

# Setup security directory
setup_security_dir() {
    mkdir -p "$SECURITY_DIR"
    chmod 700 "$SECURITY_DIR"  # Restrict access to security files
}

# Logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

error_exit() {
    log "ERROR: $1"
    exit 1
}

# Generate secure random password
generate_password() {
    local length="${1:-16}"
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 "$length" | tr -d "=+/" | cut -c1-"$length"
    elif [ -f /dev/urandom ]; then
        tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$length"
    else
        error_exit "Cannot generate secure password - openssl or /dev/urandom required"
    fi
}

# Create application user with minimal required permissions
create_app_user() {
    local db_type="$1"
    local username="${2:-radarr_app}"
    local password="${3:-$(generate_password 20)}"

    log "Creating application user '$username' for $db_type..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"

            # Create user with limited permissions
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d postgres << EOF
-- Create user if not exists
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = '$username') THEN
        CREATE USER "$username" WITH PASSWORD '$password';
    END IF;
END
\$\$;

-- Grant database connection
GRANT CONNECT ON DATABASE "$POSTGRES_DB" TO "$username";

-- Connect to the database to set up table permissions
\c "$POSTGRES_DB"

-- Grant schema usage
GRANT USAGE ON SCHEMA public TO "$username";

-- Grant table permissions (exactly what the app needs)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO "$username";
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL SEQUENCES IN SCHEMA public TO "$username";

-- Grant permissions on future tables (for migrations)
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO "$username";
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, UPDATE ON SEQUENCES TO "$username";

-- Remove superuser privileges if accidentally granted
ALTER USER "$username" NOSUPERUSER NOCREATEDB NOCREATEROLE;

EOF
            ;;
        "mariadb"|"mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" << EOF
-- Create user if not exists
CREATE USER IF NOT EXISTS '$username'@'%' IDENTIFIED BY '$password';
CREATE USER IF NOT EXISTS '$username'@'localhost' IDENTIFIED BY '$password';

-- Grant minimal required permissions
GRANT SELECT, INSERT, UPDATE, DELETE ON \`$MYSQL_DB\`.* TO '$username'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON \`$MYSQL_DB\`.* TO '$username'@'localhost';

-- Explicitly deny dangerous permissions
REVOKE FILE, PROCESS, SUPER, SHUTDOWN, RELOAD, LOCK TABLES ON *.* FROM '$username'@'%';
REVOKE FILE, PROCESS, SUPER, SHUTDOWN, RELOAD, LOCK TABLES ON *.* FROM '$username'@'localhost';

FLUSH PRIVILEGES;
EOF
            ;;
    esac

    # Save credentials securely
    local cred_file="${SECURITY_DIR}/${username}_${db_type}_credentials.txt"
    {
        echo "Database Type: $db_type"
        echo "Username: $username"
        echo "Password: $password"
        echo "Host: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_HOST" || echo "$MYSQL_HOST")"
        echo "Port: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_PORT" || echo "$MYSQL_PORT")"
        echo "Database: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_DB" || echo "$MYSQL_DB")"
        echo "Created: $(date)"
        echo
        echo "Environment variables:"
        echo "export RADARR_DATABASE_TYPE=$db_type"
        echo "export RADARR_DATABASE_HOST=$([ "$db_type" = "postgresql" ] && echo "$POSTGRES_HOST" || echo "$MYSQL_HOST")"
        echo "export RADARR_DATABASE_PORT=$([ "$db_type" = "postgresql" ] && echo "$POSTGRES_PORT" || echo "$MYSQL_PORT")"
        echo "export RADARR_DATABASE_USERNAME=$username"
        echo "export RADARR_DATABASE_PASSWORD=$password"
        echo "export RADARR_DATABASE_NAME=$([ "$db_type" = "postgresql" ] && echo "$POSTGRES_DB" || echo "$MYSQL_DB")"
    } > "$cred_file"
    chmod 600 "$cred_file"  # Restrict access to credentials

    log "Application user '$username' created for $db_type"
    log "Credentials saved to: $cred_file"
}

# Create read-only user for monitoring/reporting
create_readonly_user() {
    local db_type="$1"
    local username="${2:-radarr_readonly}"
    local password="${3:-$(generate_password 16)}"

    log "Creating read-only user '$username' for $db_type..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"

            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d postgres << EOF
-- Create read-only user
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = '$username') THEN
        CREATE USER "$username" WITH PASSWORD '$password';
    END IF;
END
\$\$;

-- Grant database connection
GRANT CONNECT ON DATABASE "$POSTGRES_DB" TO "$username";

\c "$POSTGRES_DB"

-- Grant schema usage
GRANT USAGE ON SCHEMA public TO "$username";

-- Grant SELECT only on all tables
GRANT SELECT ON ALL TABLES IN SCHEMA public TO "$username";

-- Grant SELECT on future tables
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO "$username";

-- Explicitly deny write permissions
REVOKE INSERT, UPDATE, DELETE, TRUNCATE ON ALL TABLES IN SCHEMA public FROM "$username";

-- Remove superuser privileges
ALTER USER "$username" NOSUPERUSER NOCREATEDB NOCREATEROLE;
EOF
            ;;
        "mariadb"|"mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" << EOF
-- Create read-only user
CREATE USER IF NOT EXISTS '$username'@'%' IDENTIFIED BY '$password';
CREATE USER IF NOT EXISTS '$username'@'localhost' IDENTIFIED BY '$password';

-- Grant SELECT permissions only
GRANT SELECT ON \`$MYSQL_DB\`.* TO '$username'@'%';
GRANT SELECT ON \`$MYSQL_DB\`.* TO '$username'@'localhost';

-- Explicitly deny write permissions
REVOKE INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, INDEX ON \`$MYSQL_DB\`.* FROM '$username'@'%';
REVOKE INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, INDEX ON \`$MYSQL_DB\`.* FROM '$username'@'localhost';

FLUSH PRIVILEGES;
EOF
            ;;
    esac

    # Save credentials
    local cred_file="${SECURITY_DIR}/${username}_${db_type}_readonly_credentials.txt"
    {
        echo "Database Type: $db_type (READ-ONLY)"
        echo "Username: $username"
        echo "Password: $password"
        echo "Host: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_HOST" || echo "$MYSQL_HOST")"
        echo "Port: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_PORT" || echo "$MYSQL_PORT")"
        echo "Database: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_DB" || echo "$MYSQL_DB")"
        echo "Permissions: SELECT only"
        echo "Created: $(date)"
    } > "$cred_file"
    chmod 600 "$cred_file"

    log "Read-only user '$username' created for $db_type"
    log "Credentials saved to: $cred_file"
}

# Audit current permissions
audit_permissions() {
    local db_type="$1"
    local audit_file="${SECURITY_DIR}/permission_audit_${db_type}_$(date +%Y%m%d_%H%M%S).txt"

    log "Auditing permissions for $db_type..."

    {
        echo "Radarr Go Database Permission Audit"
        echo "==================================="
        echo "Database Type: $db_type"
        echo "Audit Date: $(date)"
        echo

        case "$db_type" in
            "postgresql")
                export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"

                echo "Database Users:"
                psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -c "
                    SELECT
                        usename as username,
                        usesuper as is_superuser,
                        usecreatedb as can_create_db,
                        userepl as can_replicate,
                        valuntil as password_expiry
                    FROM pg_user
                    ORDER BY usename;
                "

                echo -e "\nTable Permissions:"
                psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" -c "
                    SELECT
                        grantee,
                        table_name,
                        privilege_type,
                        is_grantable
                    FROM information_schema.role_table_grants
                    WHERE table_schema = 'public'
                    ORDER BY grantee, table_name;
                "
                ;;
            "mariadb"|"mysql")
                echo "Database Users:"
                mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" -e "
                    SELECT
                        User,
                        Host,
                        Super_priv,
                        Create_priv,
                        Drop_priv,
                        Reload_priv,
                        File_priv,
                        password_expired
                    FROM mysql.user
                    WHERE User != ''
                    ORDER BY User;
                "

                echo -e "\nDatabase Permissions:"
                mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" -e "
                    SELECT
                        User,
                        Host,
                        Db,
                        Select_priv,
                        Insert_priv,
                        Update_priv,
                        Delete_priv,
                        Create_priv,
                        Drop_priv
                    FROM mysql.db
                    WHERE Db = '$MYSQL_DB' OR Db = '%'
                    ORDER BY User, Host;
                "
                ;;
        esac

    } > "$audit_file"

    log "Permission audit completed: $audit_file"
}

# Rotate user password
rotate_password() {
    local db_type="$1"
    local username="$2"
    local new_password="${3:-$(generate_password 20)}"

    log "Rotating password for user '$username' in $db_type..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d postgres -c "
                ALTER USER \"$username\" WITH PASSWORD '$new_password';
            " >/dev/null 2>&1 || error_exit "Failed to rotate password for PostgreSQL user $username"
            ;;
        "mariadb"|"mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" -e "
                SET PASSWORD FOR '$username'@'%' = PASSWORD('$new_password');
                SET PASSWORD FOR '$username'@'localhost' = PASSWORD('$new_password');
                FLUSH PRIVILEGES;
            " >/dev/null 2>&1 || error_exit "Failed to rotate password for MariaDB user $username"
            ;;
    esac

    # Update credentials file
    local cred_file="${SECURITY_DIR}/${username}_${db_type}_credentials.txt"
    if [ -f "$cred_file" ]; then
        # Create backup of old credentials
        cp "$cred_file" "${cred_file}.$(date +%Y%m%d_%H%M%S).bak"

        # Update password in credentials file
        sed -i.tmp "s/Password: .*/Password: $new_password/" "$cred_file"
        rm -f "${cred_file}.tmp"

        # Update environment variable line
        sed -i.tmp "s/export RADARR_DATABASE_PASSWORD=.*/export RADARR_DATABASE_PASSWORD=$new_password/" "$cred_file"
        rm -f "${cred_file}.tmp"
    fi

    log "Password rotated successfully for user '$username'"
    log "Updated credentials file: $cred_file"
}

# Create replication user
create_replication_user() {
    local db_type="$1"
    local username="${2:-radarr_repl}"
    local password="${3:-$(generate_password 20)}"

    log "Creating replication user '$username' for $db_type..."

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d postgres << EOF
-- Create replication user
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = '$username') THEN
        CREATE USER "$username" WITH REPLICATION PASSWORD '$password';
    END IF;
END
\$\$;

-- Grant minimal replication permissions
GRANT CONNECT ON DATABASE "$POSTGRES_DB" TO "$username";
EOF
            ;;
        "mariadb"|"mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" << EOF
-- Create replication user
CREATE USER IF NOT EXISTS '$username'@'%' IDENTIFIED BY '$password';

-- Grant replication permissions
GRANT REPLICATION SLAVE, REPLICATION CLIENT ON *.* TO '$username'@'%';
GRANT SELECT ON \`$MYSQL_DB\`.* TO '$username'@'%';

FLUSH PRIVILEGES;
EOF
            ;;
    esac

    # Save replication credentials
    local cred_file="${SECURITY_DIR}/${username}_${db_type}_replication_credentials.txt"
    {
        echo "Database Type: $db_type (REPLICATION)"
        echo "Username: $username"
        echo "Password: $password"
        echo "Host: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_HOST" || echo "$MYSQL_HOST")"
        echo "Port: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_PORT" || echo "$MYSQL_PORT")"
        echo "Database: $([ "$db_type" = "postgresql" ] && echo "$POSTGRES_DB" || echo "$MYSQL_DB")"
        echo "Permissions: Replication"
        echo "Created: $(date)"
    } > "$cred_file"
    chmod 600 "$cred_file"

    log "Replication user '$username' created for $db_type"
    log "Credentials saved to: $cred_file"
}

# Remove user
remove_user() {
    local db_type="$1"
    local username="$2"

    log "Removing user '$username' from $db_type..."

    # Confirm before removal
    read -p "WARNING: This will permanently remove user '$username'. Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log "User removal cancelled"
        return 0
    fi

    case "$db_type" in
        "postgresql")
            export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"

            # Revoke all permissions first
            psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d "$POSTGRES_DB" << EOF
-- Revoke all permissions
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM "$username";
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM "$username";
REVOKE ALL PRIVILEGES ON SCHEMA public FROM "$username";
REVOKE CONNECT ON DATABASE "$POSTGRES_DB" FROM "$username";

-- Drop user
DROP USER IF EXISTS "$username";
EOF
            ;;
        "mariadb"|"mysql")
            mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" << EOF
-- Drop user (automatically revokes permissions)
DROP USER IF EXISTS '$username'@'%';
DROP USER IF EXISTS '$username'@'localhost';

FLUSH PRIVILEGES;
EOF
            ;;
    esac

    # Archive credentials file
    local cred_file="${SECURITY_DIR}/${username}_${db_type}_credentials.txt"
    if [ -f "$cred_file" ]; then
        mv "$cred_file" "${cred_file}.removed.$(date +%Y%m%d_%H%M%S)"
        log "Credentials file archived"
    fi

    log "User '$username' removed from $db_type"
}

# Generate user permission matrix
generate_permission_matrix() {
    local matrix_file="${SECURITY_DIR}/permission_matrix_$(date +%Y%m%d_%H%M%S).txt"

    log "Generating permission matrix..."

    {
        echo "Radarr Go Database Permission Matrix"
        echo "====================================="
        echo "Generated: $(date)"
        echo
        echo "Recommended User Roles and Permissions:"
        echo

        cat << 'EOF'
1. APPLICATION USER (radarr_app)
   Purpose: Main application database access
   Tables: ALL (movies, quality_*, wanted_movies, etc.)
   Permissions: SELECT, INSERT, UPDATE, DELETE
   Constraints: NO DDL, NO ADMIN, NO REPLICATION

2. READ-ONLY USER (radarr_readonly)
   Purpose: Monitoring, reporting, analytics
   Tables: ALL (read-only access)
   Permissions: SELECT only
   Constraints: NO WRITE, NO DDL, NO ADMIN

3. REPLICATION USER (radarr_repl)
   Purpose: Database replication setup
   Permissions: REPLICATION SLAVE, REPLICATION CLIENT, SELECT on data
   Constraints: NO WRITE to application data

4. BACKUP USER (radarr_backup)
   Purpose: Automated backup operations
   Permissions: SELECT on all tables, LOCK TABLES (if needed)
   Constraints: NO WRITE, NO DDL

5. MIGRATION USER (radarr_migrate)
   Purpose: Schema migrations and maintenance
   Permissions: DDL operations (CREATE, ALTER, DROP)
   Constraints: Used only during maintenance windows

SECURITY PRINCIPLES:
- Principle of Least Privilege: Each user gets minimal required permissions
- Regular Password Rotation: Monthly rotation for production environments
- Network Security: Use SSL/TLS connections for remote access
- Monitoring: Log all administrative actions and permission changes
- Backup Strategy: Regular backups with tested restore procedures

CONNECTION SECURITY:
- Use SSL/TLS for all connections
- Restrict host access with specific IP ranges
- Use strong passwords (20+ characters with mixed case/numbers/symbols)
- Implement connection limits per user
- Monitor failed login attempts

MAINTENANCE SCHEDULE:
- Weekly: Permission audit and review
- Monthly: Password rotation for service accounts
- Quarterly: Full security review and penetration testing
- Annually: Disaster recovery testing and documentation update

EOF

        echo
        echo "Current Users (if databases are accessible):"
        echo

        # Try to list current users
        for db_type in postgresql mariadb; do
            echo "$db_type Users:"
            case "$db_type" in
                "postgresql")
                    if command -v psql >/dev/null 2>&1; then
                        export PGPASSWORD="$POSTGRES_ADMIN_PASSWORD"
                        psql -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_ADMIN_USER" -d postgres -c "
                            SELECT usename, usesuper, usecreatedb FROM pg_user ORDER BY usename;
                        " 2>/dev/null || echo "  Cannot connect to PostgreSQL"
                    else
                        echo "  PostgreSQL tools not available"
                    fi
                    ;;
                "mariadb")
                    if command -v mysql >/dev/null 2>&1; then
                        mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_ADMIN_USER" -p"$MYSQL_ADMIN_PASSWORD" -e "
                            SELECT User, Host, Super_priv FROM mysql.user WHERE User != '' ORDER BY User;
                        " 2>/dev/null || echo "  Cannot connect to MariaDB"
                    else
                        echo "  MySQL tools not available"
                    fi
                    ;;
            esac
            echo
        done

    } > "$matrix_file"

    log "Permission matrix generated: $matrix_file"
}

# Main function
main() {
    local command="$1"

    setup_security_dir

    case "$command" in
        "create-app-user")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 create-app-user <postgresql|mariadb> [username] [password]"
            fi
            create_app_user "$2" "${3:-radarr_app}" "${4:-}"
            ;;
        "create-readonly-user")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 create-readonly-user <postgresql|mariadb> [username] [password]"
            fi
            create_readonly_user "$2" "${3:-radarr_readonly}" "${4:-}"
            ;;
        "create-replication-user")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 create-replication-user <postgresql|mariadb> [username] [password]"
            fi
            create_replication_user "$2" "${3:-radarr_repl}" "${4:-}"
            ;;
        "audit-permissions")
            if [ $# -lt 2 ]; then
                error_exit "Usage: $0 audit-permissions <postgresql|mariadb>"
            fi
            audit_permissions "$2"
            ;;
        "rotate-password")
            if [ $# -lt 3 ]; then
                error_exit "Usage: $0 rotate-password <postgresql|mariadb> <username> [new_password]"
            fi
            rotate_password "$2" "$3" "${4:-}"
            ;;
        "remove-user")
            if [ $# -lt 3 ]; then
                error_exit "Usage: $0 remove-user <postgresql|mariadb> <username>"
            fi
            remove_user "$2" "$3"
            ;;
        "permission-matrix")
            generate_permission_matrix
            ;;
        *)
            echo "Usage: $0 <command>"
            echo
            echo "Commands:"
            echo "  create-app-user <db_type> [user] [pass]     Create application user"
            echo "  create-readonly-user <db_type> [user] [pass] Create read-only user"
            echo "  create-replication-user <db_type> [user] [pass] Create replication user"
            echo "  audit-permissions <db_type>                Audit current permissions"
            echo "  rotate-password <db_type> <user> [pass]    Rotate user password"
            echo "  remove-user <db_type> <user>               Remove user (with confirmation)"
            echo "  permission-matrix                          Generate permission matrix document"
            echo
            echo "Database types: postgresql, mariadb"
            echo
            echo "Environment variables:"
            echo "  RADARR_DATABASE_HOST           Database host (default: localhost)"
            echo "  RADARR_DATABASE_PORT           Database port"
            echo "  RADARR_DATABASE_NAME           Database name (default: radarr)"
            echo "  RADARR_DATABASE_ADMIN_USER     Admin username (default: postgres/root)"
            echo "  RADARR_DATABASE_ADMIN_PASSWORD Admin password"
            exit 1
            ;;
    esac
}

# Execute main function
main "$@"