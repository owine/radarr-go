# Radarr Go Database Disaster Recovery Runbook

## Overview
This runbook provides step-by-step procedures for 3am emergencies, disaster recovery, and database failure scenarios for Radarr Go.

**RTO (Recovery Time Objective)**: 15 minutes for basic functionality  
**RPO (Recovery Point Objective)**: 4 hours maximum data loss

## Emergency Contact Information
- Database Administrator: [Your contact]
- System Administrator: [Your contact]  
- Development Team Lead: [Your contact]

## Quick Reference - Emergency Commands

### Immediate Assessment (< 2 minutes)
```bash
# Check database connectivity
./scripts/database/monitoring.sh check

# Check disk space
df -h

# Check container status
docker-compose ps

# Check application logs
docker-compose logs --tail=50 radarr-go
```

### Database Recovery (< 5 minutes)
```bash
# Option 1: Restart database containers
docker-compose down && docker-compose up -d

# Option 2: Restore from latest backup
./scripts/database/backup_restore.sh restore postgresql /path/to/latest/backup.sql

# Option 3: Force fix dirty migration state (if migration related)
docker-compose exec postgres-test psql -U radarr_test -d radarr_test \
  -c "UPDATE schema_migrations SET dirty = false WHERE dirty = true;"
```

## Detailed Recovery Procedures

### 1. Database Connection Failures

**Symptoms**: "Cannot connect to database", "Connection refused"

**Diagnosis**:
```bash
# Test connectivity
./scripts/database/monitoring.sh check

# Check container status
docker-compose ps

# Check database logs
docker-compose logs postgres-db
docker-compose logs mariadb-db
```

**Recovery Steps**:
1. **Container Issue**: Restart database container
   ```bash
   docker-compose restart postgres-db
   # or
   docker-compose restart mariadb-db
   ```

2. **Network Issue**: Check Docker network
   ```bash
   docker network ls
   docker-compose down && docker-compose up -d
   ```

3. **Configuration Issue**: Verify environment variables
   ```bash
   env | grep RADARR_DATABASE
   ```

### 2. Migration Failures

**Symptoms**: "Dirty database version X", "Failed to run migrations"

**Diagnosis**:
```bash
# Check migration state
docker-compose exec postgres-db psql -U radarr -d radarr \
  -c "SELECT * FROM schema_migrations ORDER BY version;"

# Check for partial migration artifacts
docker-compose exec postgres-db psql -U radarr -d radarr \
  -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';"
```

**Recovery Steps**:
1. **Force Fix Dirty State**:
   ```bash
   # PostgreSQL
   docker-compose exec postgres-db psql -U radarr -d radarr \
     -c "UPDATE schema_migrations SET dirty = false WHERE dirty = true;"
   
   # MySQL/MariaDB  
   docker-compose exec mariadb-db mysql -u radarr -p radarr \
     -e "UPDATE schema_migrations SET dirty = 0 WHERE dirty = 1;"
   ```

2. **Manual Migration Fix**:
   ```bash
   # Check which migration failed
   docker-compose logs radarr-go | grep migration
   
   # Apply specific migration manually
   docker-compose exec postgres-db psql -U radarr -d radarr \
     -f /path/to/specific_migration.up.sql
   ```

3. **Full Migration Reset** (DATA LOSS):
   ```bash
   # Backup current data first!
   ./scripts/database/backup_restore.sh backup postgresql
   
   # Drop and recreate database
   docker-compose exec postgres-db psql -U postgres -d postgres \
     -c "DROP DATABASE radarr; CREATE DATABASE radarr;"
   
   # Restart application (will auto-migrate)
   docker-compose restart radarr-go
   
   # Restore data if needed
   ./scripts/database/backup_restore.sh restore postgresql /path/to/backup.sql
   ```

### 3. Data Corruption

**Symptoms**: Foreign key violations, integrity constraint errors, inconsistent data

**Diagnosis**:
```bash
# Check for foreign key violations
./scripts/database/monitoring.sh validate

# Check database integrity
# PostgreSQL:
docker-compose exec postgres-db psql -U radarr -d radarr \
  -c "SELECT * FROM pg_catalog.pg_constraint WHERE contype = 'f';"

# MySQL:
docker-compose exec mariadb-db mysql -u radarr -p radarr \
  -e "SELECT * FROM information_schema.REFERENTIAL_CONSTRAINTS;"
```

**Recovery Steps**:
1. **Identify Corruption Scope**:
   ```bash
   # Find affected tables
   ./scripts/database/monitoring.sh performance-report
   
   # Check specific table integrity
   docker-compose exec postgres-db psql -U radarr -d radarr \
     -c "SELECT COUNT(*) FROM movies WHERE quality_profile_id NOT IN (SELECT id FROM quality_profiles);"
   ```

2. **Restore from Backup**:
   ```bash
   # Find latest clean backup
   ls -la backups/ | head -10
   
   # Restore from backup
   ./scripts/database/backup_restore.sh restore postgresql /path/to/clean/backup.sql
   ```

3. **Manual Data Repair** (if backup is too old):
   ```bash
   # Fix orphaned references
   docker-compose exec postgres-db psql -U radarr -d radarr -c "
     DELETE FROM wanted_movies 
     WHERE current_quality_id NOT IN (SELECT id FROM quality_definitions);
   "
   ```

### 4. Performance Issues

**Symptoms**: Slow queries, high CPU, connection timeouts

**Diagnosis**:
```bash
# Generate performance report
./scripts/database/monitoring.sh performance-report

# Check for long-running queries
./scripts/database/monitoring.sh check

# Check connection count
docker-compose exec postgres-db psql -U radarr -d radarr \
  -c "SELECT COUNT(*) FROM pg_stat_activity;"
```

**Recovery Steps**:
1. **Immediate Relief**:
   ```bash
   # Kill long-running queries (PostgreSQL)
   docker-compose exec postgres-db psql -U radarr -d radarr -c "
     SELECT pg_terminate_backend(pid) 
     FROM pg_stat_activity 
     WHERE state = 'active' 
     AND query_start < NOW() - INTERVAL '5 minutes'
     AND pid != pg_backend_pid();
   "
   ```

2. **Maintenance**:
   ```bash
   # Run database maintenance
   ./scripts/database/monitoring.sh maintenance
   ```

3. **Scale Up** (if needed):
   ```bash
   # Increase connection limits in docker-compose.yml
   # Add connection pooling configuration
   # Consider read replicas for reporting queries
   ```

### 5. Replication Failures

**Symptoms**: High replication lag, replica out of sync

**Diagnosis**:
```bash
# Check replication status
./scripts/database/monitoring.sh replication-status

# PostgreSQL specific
docker-compose exec postgres-master psql -U radarr -d radarr \
  -c "SELECT * FROM pg_stat_replication;"

# MySQL specific
docker-compose exec mysql-slave mysql -u radarr -p radarr \
  -e "SHOW SLAVE STATUS\G"
```

**Recovery Steps**:
1. **Restart Replication** (PostgreSQL):
   ```bash
   # On replica
   docker-compose exec postgres-replica psql -U postgres -d postgres \
     -c "SELECT pg_promote();"
   
   # Rebuild replica from master
   # (Complex procedure - see PostgreSQL docs)
   ```

2. **Restart Replication** (MySQL):
   ```bash
   # On replica
   docker-compose exec mysql-replica mysql -u root -p \
     -e "STOP SLAVE; RESET SLAVE; START SLAVE;"
   ```

### 6. Complete System Failure

**Symptoms**: Total database loss, corruption beyond repair

**Recovery Steps** (< 15 minutes):
1. **Assess Damage**:
   ```bash
   # Check what's accessible
   docker-compose ps
   ./scripts/database/monitoring.sh check
   ```

2. **Emergency Restore**:
   ```bash
   # Find latest backup
   ls -la backups/ | head -5
   
   # Complete rebuild
   docker-compose down -v
   docker volume prune -f
   docker-compose up -d postgres-db mariadb-db
   sleep 30
   
   # Restore from backup
   ./scripts/database/backup_restore.sh restore postgresql /path/to/latest/backup.sql
   
   # Start application
   docker-compose up -d radarr-go
   ```

3. **Validate Recovery**:
   ```bash
   # Test application functionality
   curl http://localhost:7878/api/v3/system/status
   
   # Validate data integrity
   ./scripts/database/monitoring.sh validate
   ```

## Prevention and Maintenance

### Daily Automated Tasks
```bash
# Setup cron jobs for:
0 2 * * * /path/to/radarr-go/scripts/database/backup_restore.sh backup postgresql
0 3 * * * /path/to/radarr-go/scripts/database/backup_restore.sh backup mariadb
0 4 * * * /path/to/radarr-go/scripts/database/monitoring.sh maintenance
*/15 * * * * /path/to/radarr-go/scripts/database/monitoring.sh check
```

### Weekly Tasks
```bash
# Run every Sunday at 1 AM
0 1 * * 0 /path/to/radarr-go/scripts/database/backup_restore.sh validate
0 1 * * 0 /path/to/radarr-go/scripts/database/user_management.sh audit-permissions postgresql
0 1 * * 0 /path/to/radarr-go/scripts/database/user_management.sh audit-permissions mariadb
```

### Monthly Tasks
```bash
# Password rotation
./scripts/database/user_management.sh rotate-password postgresql radarr_app

# Performance review
./scripts/database/monitoring.sh performance-report

# Backup validation (restore to test environment)
./scripts/database/migration_validator.sh validate-all
```

### Quarterly Tasks
```bash
# Disaster recovery testing
./scripts/database/migration_validator.sh cross-validate

# Full backup restore test
./scripts/database/backup_restore.sh performance-test

# Security audit
./scripts/database/user_management.sh permission-matrix
```

## Configuration Templates

### PostgreSQL High Availability
```yaml
# docker-compose.yml additions for HA
postgres-master:
  image: postgres:17-alpine
  environment:
    - POSTGRES_DB=radarr
    - POSTGRES_USER=radarr
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    - POSTGRES_REPLICATION_MODE=master
    - POSTGRES_REPLICATION_USER=radarr_repl
    - POSTGRES_REPLICATION_PASSWORD=${REPL_PASSWORD}
  volumes:
    - postgres_master_data:/var/lib/postgresql/data
    - ./postgresql.conf:/etc/postgresql/postgresql.conf

postgres-replica:
  image: postgres:17-alpine
  environment:
    - POSTGRES_DB=radarr
    - POSTGRES_USER=radarr
    - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
    - POSTGRES_REPLICATION_MODE=replica
    - POSTGRES_MASTER_HOST=postgres-master
    - POSTGRES_REPLICATION_USER=radarr_repl
    - POSTGRES_REPLICATION_PASSWORD=${REPL_PASSWORD}
  volumes:
    - postgres_replica_data:/var/lib/postgresql/data
  depends_on:
    - postgres-master
```

### MariaDB Master-Slave Setup
```yaml
mariadb-master:
  image: mariadb:11.4
  environment:
    - MARIADB_DATABASE=radarr
    - MARIADB_USER=radarr
    - MARIADB_PASSWORD=${MYSQL_PASSWORD}
    - MARIADB_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
    - MARIADB_REPLICATION_MODE=master
    - MARIADB_REPLICATION_USER=radarr_repl
    - MARIADB_REPLICATION_PASSWORD=${REPL_PASSWORD}
  volumes:
    - mariadb_master_data:/var/lib/mysql
    - ./my.cnf:/etc/mysql/my.cnf

mariadb-slave:
  image: mariadb:11.4
  environment:
    - MARIADB_DATABASE=radarr
    - MARIADB_USER=radarr
    - MARIADB_PASSWORD=${MYSQL_PASSWORD}
    - MARIADB_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD}
    - MARIADB_REPLICATION_MODE=slave
    - MARIADB_MASTER_HOST=mariadb-master
    - MARIADB_REPLICATION_USER=radarr_repl
    - MARIADB_REPLICATION_PASSWORD=${REPL_PASSWORD}
  volumes:
    - mariadb_slave_data:/var/lib/mysql
  depends_on:
    - mariadb-master
```

### Connection Pooling Configuration
```yaml
# config.yaml additions
database:
  connection_pool:
    max_open_conns: 25      # Maximum open connections
    max_idle_conns: 5       # Maximum idle connections  
    conn_max_lifetime: 300  # Connection max lifetime (seconds)
    conn_max_idle_time: 60  # Connection max idle time (seconds)
    
monitoring:
  health_check_interval: 30   # Health check interval (seconds)
  performance_report_interval: 3600  # Performance report interval (seconds)
  alert_thresholds:
    max_connections_pct: 80
    max_query_time_ms: 1000
    max_replication_lag_seconds: 10
    min_free_disk_gb: 5
```

## Alert Notifications

### Slack Integration
```bash
export RADARR_SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
./scripts/database/monitoring.sh check
```

### Email Alerts
```bash
export RADARR_ALERT_EMAIL="admin@yourcompany.com"
./scripts/database/monitoring.sh check
```

### Discord Integration
```bash
export RADARR_DISCORD_WEBHOOK="https://discord.com/api/webhooks/YOUR/WEBHOOK"
./scripts/database/monitoring.sh check
```

## Monitoring Queries

### Critical Health Checks
```sql
-- PostgreSQL
-- Check connection count
SELECT COUNT(*) as active_connections, 
       (SELECT setting::int FROM pg_settings WHERE name = 'max_connections') as max_connections
FROM pg_stat_activity;

-- Check long running queries
SELECT pid, query_start, state, query 
FROM pg_stat_activity 
WHERE state = 'active' 
AND query_start < NOW() - INTERVAL '5 minutes';

-- Check database size
SELECT pg_size_pretty(pg_database_size(current_database())) as database_size;

-- Check table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

```sql
-- MySQL/MariaDB
-- Check connection count
SELECT VARIABLE_VALUE as current_connections
FROM information_schema.GLOBAL_STATUS 
WHERE VARIABLE_NAME = 'Threads_connected';

-- Check long running queries  
SELECT ID, USER, HOST, DB, COMMAND, TIME, STATE, INFO
FROM information_schema.PROCESSLIST
WHERE COMMAND != 'Sleep' AND TIME > 300;

-- Check database size
SELECT table_schema as database_name,
       ROUND(SUM(data_length + index_length) / 1024 / 1024, 2) as size_mb
FROM information_schema.tables
WHERE table_schema = 'radarr'
GROUP BY table_schema;

-- Check InnoDB status
SHOW ENGINE INNODB STATUS;
```

## Backup Strategy

### Automated Backup Schedule
- **Full Backup**: Daily at 2 AM local time
- **Incremental**: Every 4 hours (if supported)
- **Retention**: 30 days local, 90 days offsite
- **Testing**: Weekly backup restore validation

### Backup Verification
```bash
# Weekly backup validation
for backup in $(find backups/ -name "*.sql*" -mtime -7); do
  echo "Testing backup: $backup"
  ./scripts/database/backup_restore.sh validate "$backup"
done
```

### Offsite Backup
```bash
# Example: AWS S3 sync
aws s3 sync backups/ s3://your-backup-bucket/radarr-backups/ --exclude "*.log"

# Example: rsync to remote server  
rsync -av --delete backups/ user@backup-server:/backups/radarr/
```

## Performance Optimization

### Database Tuning (PostgreSQL)
```sql
-- postgresql.conf optimizations
shared_buffers = 256MB
effective_cache_size = 1GB  
maintenance_work_mem = 64MB
work_mem = 4MB
wal_buffers = 16MB
checkpoint_completion_target = 0.9
random_page_cost = 1.1
```

### Database Tuning (MariaDB)
```ini
# my.cnf optimizations  
[mysqld]
innodb_buffer_pool_size = 512M
innodb_log_file_size = 128M
innodb_flush_log_at_trx_commit = 2
innodb_file_per_table = 1
query_cache_size = 128M
query_cache_type = 1
max_connections = 100
```

### Index Optimization
```sql
-- Identify unused indexes
SELECT schemaname, tablename, indexname, idx_scan
FROM pg_stat_user_indexes 
WHERE idx_scan < 10
AND pg_relation_size(indexrelid) > 1000000;

-- Identify missing indexes (slow queries)
SELECT query, calls, total_time, mean_time
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;
```

## Security Hardening

### Access Control
- Use principle of least privilege for all database users
- Regular password rotation (monthly for production)
- SSL/TLS encryption for all connections
- IP-based access restrictions
- Regular permission audits

### Network Security
```yaml
# docker-compose.yml security
services:
  postgres-db:
    networks:
      - radarr-internal
    # Don't expose ports externally in production
    # ports:
    #   - "5432:5432"  # Remove this line
```

### Authentication Security
```bash
# Generate strong passwords
./scripts/database/user_management.sh create-app-user postgresql radarr_prod

# Regular security audit
./scripts/database/user_management.sh audit-permissions postgresql
./scripts/database/user_management.sh permission-matrix
```

## Testing and Validation

### Pre-deployment Testing
```bash
# Test all migrations
./scripts/database/migration_validator.sh validate-all

# Performance testing
./scripts/database/migration_validator.sh migration-test

# Cross-database validation
./scripts/database/migration_validator.sh cross-validate
```

### Post-deployment Validation
```bash
# Health check
./scripts/database/monitoring.sh check

# Performance baseline
./scripts/database/monitoring.sh performance-report

# Backup validation
./scripts/database/backup_restore.sh validate
```

## Escalation Procedures

### Level 1: Automated Recovery
- Automated health checks detect issues
- Automatic alerting via Slack/Discord/Email
- Automated restart attempts
- Self-healing for transient issues

### Level 2: Manual Intervention
- Database administrator responds to alerts
- Use this runbook for standard recovery procedures
- Implement temporary fixes if needed
- Document all actions taken

### Level 3: Emergency Response
- Major data loss or corruption detected
- Multiple recovery attempts failed
- Business impact significant
- Contact development team lead
- Consider emergency maintenance window

### Level 4: Disaster Declaration
- Complete system failure
- Data center or infrastructure issues
- Implement disaster recovery plan
- Activate backup data center (if available)
- Public communication about outage

## Contact and Documentation

### Key Files
- Migration files: `/migrations/postgres/` and `/migrations/mysql/`
- Backup scripts: `/scripts/database/backup_restore.sh`
- Monitoring: `/scripts/database/monitoring.sh`
- User management: `/scripts/database/user_management.sh`
- This runbook: `/scripts/database/disaster_recovery.md`

### Documentation Updates
When recovering from incidents:
1. Document what went wrong
2. Update this runbook with new procedures
3. Add monitoring for the failure mode
4. Test the updated procedures
5. Share lessons learned with team

### Emergency Decision Tree
```
Database Issue Detected
│
├─ Application Still Working? 
│  ├─ Yes → Monitor, Schedule Maintenance
│  └─ No → Immediate Recovery Required
│
├─ Data Accessible?
│  ├─ Yes → Connection/Performance Issue → Restart Services
│  └─ No → Data Recovery Required
│
├─ Recent Backup Available?
│  ├─ Yes → Restore from Backup (< 5 min)
│  └─ No → Manual Recovery/Rebuild (< 30 min)
│
└─ Recovery Successful?
   ├─ Yes → Document Incident, Update Procedures
   └─ No → Escalate to Level 4, Consider External Help
```

**Remember**: In a crisis, documentation is your friend. Follow procedures step by step, document everything you try, and don't hesitate to escalate when needed.