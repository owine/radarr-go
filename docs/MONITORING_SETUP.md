# Monitoring and Alerting Setup

This guide provides comprehensive monitoring and alerting configurations for Radarr Go production deployments using Prometheus, Grafana, and various alerting systems.

## Table of Contents

1. [Overview](#overview)
2. [Prometheus Configuration](#prometheus-configuration)
3. [Grafana Dashboard](#grafana-dashboard)
4. [Alerting Setup](#alerting-setup)
5. [Log Aggregation](#log-aggregation)
6. [Health Monitoring](#health-monitoring)
7. [Performance Monitoring](#performance-monitoring)
8. [Automated Monitoring Scripts](#automated-monitoring-scripts)

## Overview

Radarr Go provides built-in monitoring capabilities that integrate seamlessly with modern observability stacks:

### Built-in Metrics
- **System Metrics**: CPU, memory, disk usage, network I/O
- **Application Metrics**: API response times, database query performance, active connections
- **Business Metrics**: Movies processed, downloads completed, errors encountered
- **Health Metrics**: Component health status, external service availability

### Monitoring Stack Components
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization and dashboarding
- **AlertManager**: Alert routing and notification
- **Loki**: Log aggregation and analysis
- **Jaeger**: Distributed tracing (optional)

## Prometheus Configuration

### Prometheus Server Setup

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: 'radarr-production'
    replica: 'prometheus-1'

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  # Radarr Go Application Metrics
  - job_name: 'radarr-go'
    static_configs:
      - targets: ['radarr-go:7878']
    metrics_path: '/metrics'
    scrape_interval: 30s
    scrape_timeout: 10s
    honor_labels: true
    params:
      format: ['prometheus']

  # PostgreSQL Database Metrics
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    scrape_interval: 30s
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: postgres-exporter:9187

  # Node Metrics (System-level)
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']
    scrape_interval: 15s

  # Container Metrics (Docker)
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
    scrape_interval: 30s
    metrics_path: /metrics
    honor_labels: true

  # Redis Metrics (if using Redis)
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']
    scrape_interval: 30s

  # Nginx/Reverse Proxy Metrics
  - job_name: 'nginx'
    static_configs:
      - targets: ['nginx-exporter:9113']
    scrape_interval: 30s

# Recording rules for efficient querying
recording_rules:
  - name: radarr_aggregations
    rules:
      - record: radarr:api_request_duration_seconds:rate5m
        expr: rate(radarr_api_request_duration_seconds_sum[5m]) / rate(radarr_api_request_duration_seconds_count[5m])

      - record: radarr:database_query_duration_seconds:rate5m
        expr: rate(radarr_database_query_duration_seconds_sum[5m]) / rate(radarr_database_query_duration_seconds_count[5m])

      - record: radarr:movies_processed_total:rate5m
        expr: rate(radarr_movies_processed_total[5m])

      - record: radarr:error_rate:5m
        expr: rate(radarr_errors_total[5m]) / rate(radarr_requests_total[5m])

storage:
  tsdb:
    retention.time: 30d
    retention.size: 50GB
    min-block-duration: 2h
    max-block-duration: 25h
    wal-compression: true

# Performance optimizations
query:
  timeout: 2m
  max-concurrency: 20
  max-samples: 50000000

# Remote write configuration (optional)
# remote_write:
#   - url: "https://prometheus-remote-write-endpoint"
#     basic_auth:
#       username: "username"
#       password: "password"
```

### Alert Rules Configuration

```yaml
# monitoring/rules/radarr-alerts.yml
groups:
  - name: radarr-critical
    interval: 30s
    rules:
      - alert: RadarrDown
        expr: up{job="radarr-go"} == 0
        for: 1m
        labels:
          severity: critical
          service: radarr
        annotations:
          summary: "Radarr Go is down"
          description: "Radarr Go has been down for more than 1 minute"
          runbook_url: "https://docs.radarr.com/troubleshooting/service-down"

      - alert: RadarrHighMemoryUsage
        expr: (radarr_memory_usage_bytes / radarr_memory_limit_bytes) * 100 > 85
        for: 5m
        labels:
          severity: warning
          service: radarr
        annotations:
          summary: "Radarr Go high memory usage"
          description: "Radarr Go memory usage is above 85% for more than 5 minutes. Current: {{ $value }}%"

      - alert: RadarrHighCPUUsage
        expr: rate(radarr_cpu_seconds_total[5m]) * 100 > 80
        for: 10m
        labels:
          severity: warning
          service: radarr
        annotations:
          summary: "Radarr Go high CPU usage"
          description: "Radarr Go CPU usage is above 80% for more than 10 minutes. Current: {{ $value }}%"

      - alert: RadarrDatabaseConnections
        expr: radarr_database_connections_active / radarr_database_connections_max > 0.8
        for: 5m
        labels:
          severity: warning
          service: radarr
          component: database
        annotations:
          summary: "High database connection usage"
          description: "Database connection pool is {{ $value | humanizePercentage }} full"

      - alert: RadarrAPIResponseTime
        expr: radarr:api_request_duration_seconds:rate5m > 2
        for: 5m
        labels:
          severity: warning
          service: radarr
          component: api
        annotations:
          summary: "High API response time"
          description: "Average API response time is {{ $value }}s over the last 5 minutes"

      - alert: RadarrErrorRate
        expr: radarr:error_rate:5m > 0.05
        for: 5m
        labels:
          severity: critical
          service: radarr
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes"

  - name: radarr-database
    interval: 60s
    rules:
      - alert: PostgreSQLDown
        expr: up{job="postgres"} == 0
        for: 2m
        labels:
          severity: critical
          service: radarr
          component: database
        annotations:
          summary: "PostgreSQL is down"
          description: "PostgreSQL database has been down for more than 2 minutes"

      - alert: PostgreSQLTooManyConnections
        expr: pg_stat_activity_count / pg_settings_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
          service: radarr
          component: database
        annotations:
          summary: "PostgreSQL too many connections"
          description: "PostgreSQL connection usage is {{ $value | humanizePercentage }}"

      - alert: PostgreSQLSlowQueries
        expr: pg_stat_activity_max_tx_duration > 300
        for: 2m
        labels:
          severity: warning
          service: radarr
          component: database
        annotations:
          summary: "PostgreSQL slow queries detected"
          description: "PostgreSQL has queries running for more than 5 minutes"

  - name: radarr-business-metrics
    interval: 300s
    rules:
      - alert: RadarrNoMoviesProcessed
        expr: increase(radarr_movies_processed_total[1h]) == 0
        for: 2h
        labels:
          severity: warning
          service: radarr
        annotations:
          summary: "No movies processed in the last hour"
          description: "Radarr has not processed any movies in the last 2 hours"

      - alert: RadarrIndexerFailures
        expr: rate(radarr_indexer_requests_failed_total[5m]) / rate(radarr_indexer_requests_total[5m]) > 0.5
        for: 10m
        labels:
          severity: warning
          service: radarr
          component: indexers
        annotations:
          summary: "High indexer failure rate"
          description: "Indexer failure rate is {{ $value | humanizePercentage }} over the last 5 minutes"

      - alert: RadarrDiskSpaceLow
        expr: (radarr_disk_free_bytes / radarr_disk_total_bytes) < 0.1
        for: 5m
        labels:
          severity: critical
          service: radarr
          component: storage
        annotations:
          summary: "Low disk space"
          description: "Disk space is below 10%. Free: {{ $value | humanizePercentage }}"
```

### Exporters Configuration

```yaml
# monitoring/docker-compose.exporters.yml
version: '3.8'

services:
  # PostgreSQL Exporter
  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:latest
    container_name: postgres-exporter
    restart: unless-stopped
    environment:
      - DATA_SOURCE_NAME=postgresql://radarr:${POSTGRES_PASSWORD}@postgres:5432/radarr?sslmode=disable
      - PG_EXPORTER_EXTEND_QUERY_PATH=/etc/postgres_exporter/queries.yaml
    volumes:
      - ./monitoring/postgres-queries.yaml:/etc/postgres_exporter/queries.yaml:ro
    ports:
      - "9187:9187"
    depends_on:
      - postgres
    networks:
      - radarr-network

  # Node Exporter
  node-exporter:
    image: prom/node-exporter:latest
    container_name: node-exporter
    restart: unless-stopped
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
      - /etc/hostname:/etc/nodename:ro
      - /etc/localtime:/etc/localtime:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.sysfs=/host/sys'
      - '--path.rootfs=/rootfs'
      - '--collector.filesystem.ignored-mount-points=^/(sys|proc|dev|host|etc|rootfs/var/lib/docker/containers|rootfs/var/lib/docker/overlay2|rootfs/run/docker/netns|rootfs/var/lib/docker/aufs)($$|/)'
      - '--collector.textfile.directory=/var/lib/node_exporter/textfile_collector'
    ports:
      - "9100:9100"
    networks:
      - radarr-network

  # cAdvisor for Container Metrics
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:latest
    container_name: cadvisor
    restart: unless-stopped
    privileged: true
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:rw
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
      - /dev/disk/:/dev/disk:ro
    ports:
      - "8080:8080"
    devices:
      - /dev/kmsg:/dev/kmsg
    networks:
      - radarr-network

  # Redis Exporter (if using Redis)
  redis-exporter:
    image: oliver006/redis_exporter:latest
    container_name: redis-exporter
    restart: unless-stopped
    environment:
      - REDIS_ADDR=redis://redis:6379
    ports:
      - "9121:9121"
    depends_on:
      - redis
    networks:
      - radarr-network

  # Nginx Exporter (if using nginx reverse proxy)
  nginx-exporter:
    image: nginx/nginx-prometheus-exporter:latest
    container_name: nginx-exporter
    restart: unless-stopped
    command:
      - -nginx.scrape-uri=http://nginx:8080/nginx_status
    ports:
      - "9113:9113"
    depends_on:
      - nginx
    networks:
      - radarr-network

networks:
  radarr-network:
    external: true
```

### Custom PostgreSQL Queries

```yaml
# monitoring/postgres-queries.yaml
pg_database:
  query: "SELECT pg_database.datname, pg_database_size(pg_database.datname) as size_bytes FROM pg_database"
  master: true
  cache_seconds: 30
  metrics:
    - datname:
        usage: "LABEL"
        description: "Database name"
    - size_bytes:
        usage: "GAUGE"
        description: "Database size in bytes"

pg_stat_activity:
  query: "SELECT state, count(*) as count FROM pg_stat_activity GROUP BY state"
  master: true
  cache_seconds: 30
  metrics:
    - state:
        usage: "LABEL"
        description: "Connection state"
    - count:
        usage: "GAUGE"
        description: "Number of connections in this state"

pg_slow_queries:
  query: "SELECT count(*) as slow_queries FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '60 seconds'"
  master: true
  cache_seconds: 60
  metrics:
    - slow_queries:
        usage: "GAUGE"
        description: "Number of queries running for more than 60 seconds"

radarr_table_sizes:
  query: "SELECT schemaname, tablename, pg_total_relation_size(schemaname||'.'||tablename) as size_bytes FROM pg_tables WHERE schemaname NOT IN ('information_schema', 'pg_catalog')"
  master: true
  cache_seconds: 300
  metrics:
    - schemaname:
        usage: "LABEL"
        description: "Schema name"
    - tablename:
        usage: "LABEL"
        description: "Table name"
    - size_bytes:
        usage: "GAUGE"
        description: "Table size in bytes"
```

## Grafana Dashboard

### Dashboard Configuration

```json
{
  "dashboard": {
    "id": null,
    "title": "Radarr Go Production Monitoring",
    "tags": ["radarr", "go", "production"],
    "style": "dark",
    "timezone": "browser",
    "refresh": "30s",
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "panels": [
      {
        "id": 1,
        "title": "System Overview",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"radarr-go\"}",
            "legendFormat": "Radarr Status"
          },
          {
            "expr": "radarr_movies_total",
            "legendFormat": "Total Movies"
          },
          {
            "expr": "radarr_active_downloads",
            "legendFormat": "Active Downloads"
          },
          {
            "expr": "rate(radarr_api_requests_total[5m])",
            "legendFormat": "API Requests/sec"
          }
        ],
        "gridPos": {"h": 4, "w": 24, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "API Performance",
        "type": "timeseries",
        "targets": [
          {
            "expr": "radarr:api_request_duration_seconds:rate5m",
            "legendFormat": "Average Response Time"
          },
          {
            "expr": "histogram_quantile(0.95, rate(radarr_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th Percentile"
          },
          {
            "expr": "histogram_quantile(0.99, rate(radarr_api_request_duration_seconds_bucket[5m]))",
            "legendFormat": "99th Percentile"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 4}
      },
      {
        "id": 3,
        "title": "Error Rate",
        "type": "timeseries",
        "targets": [
          {
            "expr": "radarr:error_rate:5m * 100",
            "legendFormat": "Error Rate (%)"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 4}
      },
      {
        "id": 4,
        "title": "Resource Usage",
        "type": "timeseries",
        "targets": [
          {
            "expr": "radarr_memory_usage_bytes / 1024 / 1024",
            "legendFormat": "Memory (MB)"
          },
          {
            "expr": "rate(radarr_cpu_seconds_total[5m]) * 100",
            "legendFormat": "CPU (%)"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 12}
      },
      {
        "id": 5,
        "title": "Database Performance",
        "type": "timeseries",
        "targets": [
          {
            "expr": "radarr:database_query_duration_seconds:rate5m * 1000",
            "legendFormat": "Query Time (ms)"
          },
          {
            "expr": "radarr_database_connections_active",
            "legendFormat": "Active Connections"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 12}
      }
    ]
  }
}
```

### Grafana Provisioning

```yaml
# monitoring/grafana/provisioning/datasources/prometheus.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true
    jsonData:
      timeInterval: "15s"
      queryTimeout: "60s"
      httpMethod: "POST"

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: true
    jsonData:
      maxLines: 1000
      derivedFields:
        - datasourceUid: "prometheus"
          matcherRegex: "trace_id=(\\w+)"
          name: "TraceID"
          url: "$${__value.raw}"
```

```yaml
# monitoring/grafana/provisioning/dashboards/dashboard.yml
apiVersion: 1

providers:
  - name: 'radarr-dashboards'
    orgId: 1
    folder: 'Radarr'
    folderUid: 'radarr'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards/radarr
```

## Alerting Setup

### AlertManager Configuration

```yaml
# monitoring/alertmanager.yml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'radarr-alerts@yourdomain.com'
  smtp_auth_username: 'radarr-alerts@yourdomain.com'
  smtp_auth_password: 'your-email-password'
  slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'

route:
  group_by: ['alertname', 'severity', 'service']
  group_wait: 10s
  group_interval: 5m
  repeat_interval: 12h
  receiver: 'web.hook'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
      group_wait: 5s
      repeat_interval: 5m
    - match:
        severity: warning
      receiver: 'warning-alerts'
      group_wait: 30s
      repeat_interval: 1h
    - match:
        service: radarr
        component: database
      receiver: 'database-alerts'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://webhook-receiver:5000/webhook'
        send_resolved: true

  - name: 'critical-alerts'
    email_configs:
      - to: 'ops-team@yourdomain.com'
        subject: '[CRITICAL] Radarr Alert: {{ .GroupLabels.alertname }}'
        body: |
          {{ range .Alerts }}
          Alert: {{ .Annotations.summary }}
          Description: {{ .Annotations.description }}
          Instance: {{ .Labels.instance }}
          Severity: {{ .Labels.severity }}
          Time: {{ .StartsAt }}
          {{ end }}
    slack_configs:
      - channel: '#alerts-critical'
        color: 'danger'
        title: 'Critical Alert: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true

  - name: 'warning-alerts'
    slack_configs:
      - channel: '#alerts-warning'
        color: 'warning'
        title: 'Warning: {{ .GroupLabels.alertname }}'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
        send_resolved: true

  - name: 'database-alerts'
    email_configs:
      - to: 'dba-team@yourdomain.com'
        subject: '[DATABASE] Radarr Database Alert: {{ .GroupLabels.alertname }}'
    pagerduty_configs:
      - service_key: 'your-pagerduty-service-key'
        description: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']
```

### Notification Templates

```yaml
# monitoring/notification-templates/discord.yml
webhook_configs:
  - url: 'https://discord.com/api/webhooks/YOUR/WEBHOOK/URL'
    send_resolved: true
    http_config:
      proxy_url: 'http://proxy.company.com:8080'
    title: 'Radarr Alert'
    message: |
      **Alert**: {{ .GroupLabels.alertname }}
      **Severity**: {{ .CommonLabels.severity }}
      **Service**: {{ .CommonLabels.service }}

      {{ range .Alerts }}
      **Summary**: {{ .Annotations.summary }}
      **Description**: {{ .Annotations.description }}
      **Instance**: {{ .Labels.instance }}
      **Time**: {{ .StartsAt.Format "2006-01-02 15:04:05" }}
      {{ end }}
```

```bash
#!/bin/bash
# monitoring/scripts/custom-webhook.sh
# Custom webhook handler for complex alert processing

WEBHOOK_URL="$1"
ALERT_DATA="$2"

# Parse alert data
ALERT_NAME=$(echo "$ALERT_DATA" | jq -r '.alerts[0].labels.alertname')
SEVERITY=$(echo "$ALERT_DATA" | jq -r '.alerts[0].labels.severity')
DESCRIPTION=$(echo "$ALERT_DATA" | jq -r '.alerts[0].annotations.description')

# Send to multiple systems based on severity
case "$SEVERITY" in
  "critical")
    # Send to PagerDuty
    curl -X POST "https://events.pagerduty.com/v2/enqueue" \
      -H "Content-Type: application/json" \
      -d "{
        \"routing_key\": \"$PAGERDUTY_KEY\",
        \"event_action\": \"trigger\",
        \"payload\": {
          \"summary\": \"$ALERT_NAME: $DESCRIPTION\",
          \"source\": \"radarr-monitoring\",
          \"severity\": \"critical\"
        }
      }"

    # Send SMS via Twilio
    curl -X POST "https://api.twilio.com/2010-04-01/Accounts/$TWILIO_SID/Messages.json" \
      --data-urlencode "From=$TWILIO_FROM" \
      --data-urlencode "To=$ONCALL_PHONE" \
      --data-urlencode "Body=CRITICAL: Radarr - $DESCRIPTION" \
      -u "$TWILIO_SID:$TWILIO_TOKEN"
    ;;

  "warning")
    # Send to Slack only
    curl -X POST "$SLACK_WEBHOOK" \
      -H "Content-Type: application/json" \
      -d "{
        \"channel\": \"#alerts-warning\",
        \"username\": \"Radarr Monitor\",
        \"text\": \"⚠️ $ALERT_NAME: $DESCRIPTION\"
      }"
    ;;
esac
```

## Log Aggregation

### Loki Configuration

```yaml
# monitoring/loki.yml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  instance_addr: 127.0.0.1
  path_prefix: /tmp/loki
  storage:
    filesystem:
      chunks_directory: /tmp/loki/chunks
      rules_directory: /tmp/loki/rules
  replication_factor: 1
  ring:
    kvstore:
      store: inmemory

query_range:
  results_cache:
    cache:
      embedded_cache:
        enabled: true
        max_size_mb: 100

schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

ruler:
  alertmanager_url: http://alertmanager:9093

# Frontend limits
limits_config:
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  max_cache_freshness_per_query: 10m
  split_queries_by_interval: 15m
  query_timeout: 300s
  max_concurrent_tail_requests: 20
  max_query_parallelism: 32
  max_streams_per_user: 10000
  max_line_size: 256000
  increment_duplicate_timestamp: true

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: false
  retention_period: 0s

compactor:
  working_directory: /tmp/loki/boltdb-shipper-compactor
  shared_store: filesystem
```

### Promtail Configuration

```yaml
# monitoring/promtail.yml
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  # Radarr application logs
  - job_name: radarr-logs
    static_configs:
      - targets:
          - localhost
        labels:
          job: radarr
          service: radarr-go
          __path__: /var/log/radarr/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            timestamp: time
            message: msg
            component: component
            request_id: request_id
      - labels:
          level:
          component:
          request_id:
      - timestamp:
          source: timestamp
          format: RFC3339Nano

  # Container logs
  - job_name: containers
    static_configs:
      - targets:
          - localhost
        labels:
          job: containerlogs
          __path__: /var/lib/docker/containers/*/*.log
    pipeline_stages:
      - json:
          expressions:
            log: log
            stream: stream
            time: time
      - labels:
          stream:
      - timestamp:
          source: time
          format: RFC3339Nano

  # System logs
  - job_name: syslog
    static_configs:
      - targets:
          - localhost
        labels:
          job: syslog
          __path__: /var/log/syslog
    pipeline_stages:
      - regex:
          expression: '^(?P<timestamp>\S+\s+\d+\s+\d+:\d+:\d+)\s+(?P<hostname>\S+)\s+(?P<service>\S+)(?:\[(?P<pid>\d+)\])?:\s+(?P<message>.*)'
      - labels:
          hostname:
          service:
          pid:
      - timestamp:
          source: timestamp
          format: Jan 2 15:04:05

  # Nginx logs
  - job_name: nginx
    static_configs:
      - targets:
          - localhost
        labels:
          job: nginx
          service: nginx
          __path__: /var/log/nginx/*.log
    pipeline_stages:
      - regex:
          expression: '^(?P<remote_addr>\S+) - (?P<remote_user>\S+) \[(?P<timestamp>[^\]]+)\] "(?P<method>\S+) (?P<url>\S+) (?P<protocol>\S+)" (?P<status>\d+) (?P<body_bytes_sent>\d+) "(?P<referer>[^"]*)" "(?P<user_agent>[^"]*)"'
      - labels:
          method:
          status:
          remote_addr:
      - timestamp:
          source: timestamp
          format: 02/Jan/2006:15:04:05 -0700
```

## Health Monitoring

### Custom Health Checks

```bash
#!/bin/bash
# monitoring/scripts/health-check.sh
# Comprehensive health checking script

set -euo pipefail

# Configuration
RADARR_URL="http://localhost:7878"
API_KEY="${RADARR_API_KEY}"
POSTGRES_HOST="${POSTGRES_HOST:-localhost}"
POSTGRES_USER="${POSTGRES_USER:-radarr}"
POSTGRES_DB="${POSTGRES_DB:-radarr}"
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARNING: $1${NC}"; }
error() { echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1${NC}"; }

# Health check results
HEALTH_RESULTS=()

# Check API health
check_api_health() {
    log "Checking API health..."

    if curl -s -f "$RADARR_URL/ping" >/dev/null; then
        HEALTH_RESULTS+=("api:healthy")
        log "API health check passed"
    else
        HEALTH_RESULTS+=("api:unhealthy")
        error "API health check failed"
        return 1
    fi
}

# Check API authentication
check_api_auth() {
    log "Checking API authentication..."

    local response=$(curl -s -w "%{http_code}" -o /dev/null "$RADARR_URL/api/v3/system/status" -H "X-Api-Key: $API_KEY")

    if [ "$response" -eq 200 ]; then
        HEALTH_RESULTS+=("auth:healthy")
        log "API authentication check passed"
    else
        HEALTH_RESULTS+=("auth:unhealthy")
        error "API authentication check failed (HTTP $response)"
        return 1
    fi
}

# Check database connectivity
check_database() {
    log "Checking database connectivity..."

    if PGPASSWORD="$POSTGRES_PASSWORD" pg_isready -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1; then
        HEALTH_RESULTS+=("database:healthy")
        log "Database connectivity check passed"
    else
        HEALTH_RESULTS+=("database:unhealthy")
        error "Database connectivity check failed"
        return 1
    fi
}

# Check database performance
check_database_performance() {
    log "Checking database performance..."

    local query_time=$(PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -t -c "
        SELECT EXTRACT(EPOCH FROM (now() - query_start)) as seconds
        FROM pg_stat_activity
        WHERE state = 'active' AND query_start IS NOT NULL
        ORDER BY seconds DESC LIMIT 1;" 2>/dev/null | tr -d ' ')

    if [ -n "$query_time" ] && (( $(echo "$query_time > 60" | bc -l) )); then
        HEALTH_RESULTS+=("db_performance:degraded")
        warn "Long-running query detected: ${query_time}s"
    else
        HEALTH_RESULTS+=("db_performance:healthy")
        log "Database performance check passed"
    fi
}

# Check disk space
check_disk_space() {
    log "Checking disk space..."

    local usage=$(df / | awk 'NR==2 {print int($5)}')

    if [ "$usage" -gt 90 ]; then
        HEALTH_RESULTS+=("disk:critical")
        error "Disk usage critical: ${usage}%"
        return 1
    elif [ "$usage" -gt 80 ]; then
        HEALTH_RESULTS+=("disk:warning")
        warn "Disk usage high: ${usage}%"
    else
        HEALTH_RESULTS+=("disk:healthy")
        log "Disk space check passed: ${usage}% used"
    fi
}

# Check memory usage
check_memory() {
    log "Checking memory usage..."

    local mem_usage=$(free | awk '/^Mem:/ {print int($3/$2 * 100)}')

    if [ "$mem_usage" -gt 90 ]; then
        HEALTH_RESULTS+=("memory:critical")
        error "Memory usage critical: ${mem_usage}%"
        return 1
    elif [ "$mem_usage" -gt 80 ]; then
        HEALTH_RESULTS+=("memory:warning")
        warn "Memory usage high: ${mem_usage}%"
    else
        HEALTH_RESULTS+=("memory:healthy")
        log "Memory usage check passed: ${mem_usage}% used"
    fi
}

# Check external services (indexers, download clients)
check_external_services() {
    log "Checking external services..."

    # Get indexer statuses
    local indexers_response=$(curl -s "$RADARR_URL/api/v3/indexer" -H "X-Api-Key: $API_KEY" 2>/dev/null || echo "[]")
    local failed_indexers=$(echo "$indexers_response" | jq -r '.[] | select(.enable == true and .supportsRss == true) | select(.testAll == false) | .name' 2>/dev/null || echo "")

    if [ -n "$failed_indexers" ]; then
        HEALTH_RESULTS+=("indexers:degraded")
        warn "Failed indexers detected: $failed_indexers"
    else
        HEALTH_RESULTS+=("indexers:healthy")
        log "External services check passed"
    fi
}

# Send alert if needed
send_alert() {
    local status="$1"
    local message="$2"

    if [ -n "$ALERT_WEBHOOK" ]; then
        curl -s -X POST "$ALERT_WEBHOOK" \
            -H "Content-Type: application/json" \
            -d "{
                \"service\": \"radarr-go\",
                \"status\": \"$status\",
                \"message\": \"$message\",
                \"timestamp\": \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\",
                \"details\": $(printf '%s\n' "${HEALTH_RESULTS[@]}" | jq -R . | jq -s .)
            }" >/dev/null
    fi
}

# Main execution
main() {
    log "Starting Radarr Go health check..."

    local overall_status="healthy"
    local failed_checks=0

    # Run all health checks
    check_api_health || ((failed_checks++))
    check_api_auth || ((failed_checks++))
    check_database || ((failed_checks++))
    check_database_performance
    check_disk_space || ((failed_checks++))
    check_memory || ((failed_checks++))
    check_external_services

    # Determine overall status
    if [ $failed_checks -gt 0 ]; then
        if [ $failed_checks -gt 2 ]; then
            overall_status="critical"
        else
            overall_status="warning"
        fi
    fi

    # Generate summary
    local summary="Health check completed. Status: $overall_status (${#HEALTH_RESULTS[@]} checks)"

    case "$overall_status" in
        "healthy")
            log "$summary"
            ;;
        "warning")
            warn "$summary"
            send_alert "warning" "Radarr Go has $failed_checks failing health checks"
            ;;
        "critical")
            error "$summary"
            send_alert "critical" "Radarr Go has $failed_checks critical health check failures"
            exit 1
            ;;
    esac

    # Export metrics for Prometheus (optional)
    if [ -n "${NODE_EXPORTER_TEXTFILE_DIR:-}" ]; then
        {
            echo "# HELP radarr_health_check_status Health check status (1=healthy, 0=unhealthy)"
            echo "# TYPE radarr_health_check_status gauge"
            for result in "${HEALTH_RESULTS[@]}"; do
                local component="${result%:*}"
                local status="${result#*:}"
                local value=1
                [ "$status" != "healthy" ] && value=0
                echo "radarr_health_check_status{component=\"$component\",status=\"$status\"} $value"
            done
        } > "$NODE_EXPORTER_TEXTFILE_DIR/radarr_health.prom.$$" && \
        mv "$NODE_EXPORTER_TEXTFILE_DIR/radarr_health.prom.$$" "$NODE_EXPORTER_TEXTFILE_DIR/radarr_health.prom"
    fi
}

# Run if executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
```

## Performance Monitoring

### Performance Benchmarking Script

```bash
#!/bin/bash
# monitoring/scripts/performance-benchmark.sh
# Performance benchmarking and monitoring

set -euo pipefail

RADARR_URL="${RADARR_URL:-http://localhost:7878}"
API_KEY="${RADARR_API_KEY}"
BENCHMARK_DURATION="${BENCHMARK_DURATION:-300}"
CONCURRENT_USERS="${CONCURRENT_USERS:-10}"

# Test API endpoint performance
benchmark_api_endpoints() {
    local endpoints=(
        "/api/v3/system/status"
        "/api/v3/movie"
        "/api/v3/queue"
        "/api/v3/history"
        "/api/v3/indexer"
        "/api/v3/downloadclient"
    )

    echo "# API Endpoint Performance Benchmarks"
    echo "endpoint,avg_response_time,min_time,max_time,requests_per_second,error_rate"

    for endpoint in "${endpoints[@]}"; do
        echo "Testing $endpoint..."

        local results=$(ab -n 100 -c 10 -H "X-Api-Key: $API_KEY" \
            "$RADARR_URL$endpoint" 2>/dev/null | \
            awk '
            /Time taken for tests:/ { total_time = $5 }
            /Complete requests:/ { requests = $3 }
            /Failed requests:/ { failed = $3 }
            /Requests per second:/ { rps = $4 }
            /Time per request:.*\(mean\)/ { avg_time = $4 }
            /min.*mean.*max/ { getline; split($0, times, /\s+/); min_time = times[2]; max_time = times[4] }
            END {
                error_rate = (failed/requests) * 100
                printf "'"$endpoint"',%s,%s,%s,%s,%.2f\n", avg_time, min_time, max_time, rps, error_rate
            }')

        echo "$results"
    done
}

# Database query performance analysis
analyze_database_performance() {
    echo "# Database Performance Analysis"

    PGPASSWORD="$POSTGRES_PASSWORD" psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "
    SELECT
        query,
        calls,
        total_time,
        mean_time,
        stddev_time,
        rows,
        100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
    FROM pg_stat_statements
    WHERE query NOT LIKE '%pg_stat_statements%'
    ORDER BY mean_time DESC
    LIMIT 20;"
}

# Generate performance report
generate_report() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local report_file="performance_report_$(date '+%Y%m%d_%H%M%S').txt"

    {
        echo "Radarr Go Performance Report"
        echo "Generated: $timestamp"
        echo "=========================================="
        echo
        benchmark_api_endpoints
        echo
        analyze_database_performance
    } > "$report_file"

    echo "Performance report saved: $report_file"
}

# Main execution
case "${1:-benchmark}" in
    "benchmark") benchmark_api_endpoints ;;
    "database") analyze_database_performance ;;
    "report") generate_report ;;
    *) echo "Usage: $0 {benchmark|database|report}"; exit 1 ;;
esac
```

## Automated Monitoring Scripts

### Complete Monitoring Deployment

```bash
#!/bin/bash
# monitoring/deploy-monitoring.sh
# Complete monitoring stack deployment

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')] $1${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING: $1${NC}"; }
error() { echo -e "${RED}[$(date +'%H:%M:%S')] ERROR: $1${NC}"; exit 1; }

# Deploy monitoring stack
deploy_monitoring() {
    log "Deploying monitoring stack..."

    # Create monitoring directory structure
    mkdir -p {prometheus,grafana,alertmanager,loki,promtail}/{config,data}
    mkdir -p grafana/provisioning/{datasources,dashboards}

    # Set permissions
    sudo chown -R 472:472 grafana/  # Grafana user
    sudo chown -R 65534:65534 prometheus/  # Nobody user
    sudo chown -R 10001:10001 loki/  # Loki user

    # Deploy Prometheus
    docker run -d \
        --name prometheus \
        --network radarr-network \
        -p 9090:9090 \
        -v "$PWD/prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro" \
        -v "$PWD/prometheus/rules:/etc/prometheus/rules:ro" \
        -v prometheus_data:/prometheus \
        --restart unless-stopped \
        prom/prometheus:latest \
        --config.file=/etc/prometheus/prometheus.yml \
        --storage.tsdb.path=/prometheus \
        --web.console.libraries=/usr/share/prometheus/console_libraries \
        --web.console.templates=/usr/share/prometheus/consoles \
        --storage.tsdb.retention.time=30d \
        --web.enable-lifecycle

    # Deploy AlertManager
    docker run -d \
        --name alertmanager \
        --network radarr-network \
        -p 9093:9093 \
        -v "$PWD/alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro" \
        --restart unless-stopped \
        prom/alertmanager:latest

    # Deploy Grafana
    docker run -d \
        --name grafana \
        --network radarr-network \
        -p 3000:3000 \
        -v grafana_data:/var/lib/grafana \
        -v "$PWD/grafana/provisioning:/etc/grafana/provisioning:ro" \
        -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
        -e "GF_USERS_ALLOW_SIGN_UP=false" \
        --restart unless-stopped \
        grafana/grafana:latest

    # Deploy Loki
    docker run -d \
        --name loki \
        --network radarr-network \
        -p 3100:3100 \
        -v "$PWD/loki/loki.yml:/etc/loki/local-config.yaml:ro" \
        -v loki_data:/loki \
        --restart unless-stopped \
        grafana/loki:latest

    # Deploy Promtail
    docker run -d \
        --name promtail \
        --network radarr-network \
        -v "$PWD/promtail/promtail.yml:/etc/promtail/config.yml:ro" \
        -v /var/log:/var/log:ro \
        -v /var/lib/docker/containers:/var/lib/docker/containers:ro \
        --restart unless-stopped \
        grafana/promtail:latest

    # Deploy exporters
    docker-compose -f docker-compose.exporters.yml up -d

    log "Monitoring stack deployed successfully"
    log "Access URLs:"
    log "  Prometheus: http://localhost:9090"
    log "  Grafana: http://localhost:3000 (admin/admin)"
    log "  AlertManager: http://localhost:9093"
}

# Setup monitoring cron jobs
setup_cron_jobs() {
    log "Setting up monitoring cron jobs..."

    # Add health check every 5 minutes
    (crontab -l 2>/dev/null; echo "*/5 * * * * $PWD/scripts/health-check.sh >/dev/null 2>&1") | crontab -

    # Add performance benchmark daily
    (crontab -l 2>/dev/null; echo "0 2 * * * $PWD/scripts/performance-benchmark.sh report") | crontab -

    # Add log rotation weekly
    (crontab -l 2>/dev/null; echo "0 0 * * 0 docker system prune -f --volumes") | crontab -

    log "Cron jobs configured"
}

# Verify monitoring setup
verify_monitoring() {
    log "Verifying monitoring setup..."

    local services=("prometheus" "grafana" "alertmanager" "loki" "promtail")
    local failed=0

    for service in "${services[@]}"; do
        if docker ps | grep -q "$service"; then
            log "✓ $service is running"
        else
            error "✗ $service is not running"
            ((failed++))
        fi
    done

    # Test Prometheus targets
    local targets_up=$(curl -s http://localhost:9090/api/v1/targets | jq -r '.data.activeTargets[] | select(.health == "up") | .scrapeUrl' | wc -l)
    log "Prometheus targets up: $targets_up"

    if [ $failed -eq 0 ]; then
        log "All monitoring services are running correctly"
    else
        error "$failed monitoring services failed to start"
    fi
}

# Main execution
case "${1:-deploy}" in
    "deploy")
        deploy_monitoring
        setup_cron_jobs
        verify_monitoring
        ;;
    "verify")
        verify_monitoring
        ;;
    "clean")
        docker stop prometheus grafana alertmanager loki promtail 2>/dev/null || true
        docker rm prometheus grafana alertmanager loki promtail 2>/dev/null || true
        docker volume rm prometheus_data grafana_data loki_data 2>/dev/null || true
        log "Monitoring stack cleaned up"
        ;;
    *)
        echo "Usage: $0 {deploy|verify|clean}"
        exit 1
        ;;
esac
```

This comprehensive monitoring setup provides:

1. **Full observability stack** with Prometheus, Grafana, AlertManager, and Loki
2. **Complete alert rules** covering system, application, and business metrics
3. **Automated health checking** with multiple notification channels
4. **Performance benchmarking** and database query analysis
5. **Log aggregation** with structured logging support
6. **Deployment automation** with verification and cleanup scripts

The monitoring system is production-ready with proper error handling, security considerations, and scalability features.
