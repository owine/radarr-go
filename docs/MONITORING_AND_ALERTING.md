# Radarr Go Monitoring and Alerting Setup

**Version**: v0.9.0-alpha

## Overview

This guide provides comprehensive monitoring and alerting solutions for production Radarr Go deployments. The monitoring stack includes:

- **Prometheus** - Metrics collection and storage
- **Grafana** - Visualization and dashboards
- **AlertManager** - Alert routing and notification
- **Loki** - Log aggregation and search
- **Node Exporter** - System metrics collection
- **PostgreSQL Exporter** - Database metrics

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Radarr Go     │───▶│   Prometheus     │───▶│    Grafana      │
│   (Metrics)     │    │  (Collection)    │    │ (Visualization) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         │                       ▼                       │
         │              ┌──────────────────┐             │
         │              │  AlertManager    │             │
         │              │ (Notifications)  │             │
         │              └──────────────────┘             │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│      Loki       │    │   Notifications   │    │   Dashboards    │
│ (Log Storage)   │    │ (Slack/Email/etc) │    │   (Metrics)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Prometheus Configuration

### Docker Compose Monitoring Stack

```yaml
# docker-compose.monitoring.yml
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: radarr-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - prometheus_data:/prometheus
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - ./monitoring/rules:/etc/prometheus/rules:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/usr/share/prometheus/console_libraries'
      - '--web.console.templates=/usr/share/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--storage.tsdb.retention.size=10GB'
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
      - '--web.external-url=http://prometheus.yourdomain.com'
    networks:
      - monitoring
    labels:
      - traefik.enable=true
      - traefik.http.routers.prometheus.rule=Host(`prometheus.yourdomain.com`)
      - traefik.http.routers.prometheus.entrypoints=websecure
      - traefik.http.routers.prometheus.tls=true
      - traefik.http.services.prometheus.loadbalancer.server.port=9090

  grafana:
    image: grafana/grafana:latest
    container_name: radarr-grafana
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
      - ./monitoring/grafana/provisioning:/etc/grafana/provisioning:ro
      - ./monitoring/grafana/dashboards:/var/lib/grafana/dashboards:ro
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
      - GF_SECURITY_ADMIN_USER=admin
      - GF_INSTALL_PLUGINS=grafana-clock-panel,grafana-simple-json-datasource,grafana-piechart-panel
      - GF_SMTP_ENABLED=true
      - GF_SMTP_HOST=${SMTP_HOST}:${SMTP_PORT}
      - GF_SMTP_USER=${SMTP_USER}
      - GF_SMTP_PASSWORD=${SMTP_PASSWORD}
      - GF_SMTP_FROM_ADDRESS=${SMTP_FROM}
      - GF_SERVER_ROOT_URL=https://grafana.yourdomain.com
      - GF_DATABASE_TYPE=postgres
      - GF_DATABASE_HOST=postgres:5432
      - GF_DATABASE_NAME=grafana
      - GF_DATABASE_USER=grafana
      - GF_DATABASE_PASSWORD=${GRAFANA_DB_PASSWORD}
    networks:
      - monitoring
    depends_on:
      - prometheus
    labels:
      - traefik.enable=true
      - traefik.http.routers.grafana.rule=Host(`grafana.yourdomain.com`)
      - traefik.http.routers.grafana.entrypoints=websecure
      - traefik.http.routers.grafana.tls=true
      - traefik.http.services.grafana.loadbalancer.server.port=3000

  alertmanager:
    image: prom/alertmanager:latest
    container_name: radarr-alertmanager
    restart: unless-stopped
    ports:
      - "9093:9093"
    volumes:
      - alertmanager_data:/alertmanager
      - ./monitoring/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
      - '--web.external-url=http://alertmanager.yourdomain.com'
      - '--web.route-prefix=/'
      - '--cluster.advertise-address=0.0.0.0:9093'
    networks:
      - monitoring
    labels:
      - traefik.enable=true
      - traefik.http.routers.alertmanager.rule=Host(`alertmanager.yourdomain.com`)
      - traefik.http.routers.alertmanager.entrypoints=websecure
      - traefik.http.routers.alertmanager.tls=true
      - traefik.http.services.alertmanager.loadbalancer.server.port=9093

  loki:
    image: grafana/loki:latest
    container_name: radarr-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    volumes:
      - loki_data:/loki
      - ./monitoring/loki.yml:/etc/loki/local-config.yaml:ro
    command: -config.file=/etc/loki/local-config.yaml
    networks:
      - monitoring

  promtail:
    image: grafana/promtail:latest
    container_name: radarr-promtail
    restart: unless-stopped
    volumes:
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - ./monitoring/promtail.yml:/etc/promtail/config.yml:ro
    command: -config.file=/etc/promtail/config.yml
    networks:
      - monitoring
    depends_on:
      - loki

  node-exporter:
    image: prom/node-exporter:latest
    container_name: radarr-node-exporter
    restart: unless-stopped
    ports:
      - "9100:9100"
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
      - '--collector.systemd'
      - '--collector.processes'
    networks:
      - monitoring

  postgres-exporter:
    image: prometheuscommunity/postgres-exporter:latest
    container_name: radarr-postgres-exporter
    restart: unless-stopped
    ports:
      - "9187:9187"
    environment:
      - DATA_SOURCE_NAME=postgresql://radarr:${POSTGRES_PASSWORD}@postgres:5432/radarr?sslmode=disable
      - PG_EXPORTER_EXCLUDE_DATABASES=template0,template1
    networks:
      - monitoring
    depends_on:
      - postgres

volumes:
  prometheus_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/monitoring/prometheus
  grafana_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/monitoring/grafana
  alertmanager_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/monitoring/alertmanager
  loki_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/monitoring/loki

networks:
  monitoring:
    driver: bridge
  default:
    external:
      name: radarr_default
```

### Prometheus Configuration

Create `monitoring/prometheus.yml`:

```yaml
# monitoring/prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    monitor: 'radarr-monitor'
    environment: 'production'

rule_files:
  - "rules/*.yml"

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

scrape_configs:
  # Radarr Go Application
  - job_name: 'radarr-go'
    static_configs:
      - targets: ['radarr-go:7878']
    metrics_path: /metrics
    scrape_interval: 30s
    scrape_timeout: 10s
    scheme: http
    basic_auth:
      username: ''
      password: ''
    params:
      format: ['prometheus']
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        replacement: 'radarr-go'

  # PostgreSQL Database
  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres-exporter:9187']
    scrape_interval: 30s
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        replacement: 'postgres'

  # System Metrics
  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']
    scrape_interval: 15s
    relabel_configs:
      - source_labels: [__address__]
        target_label: instance
        replacement: 'docker-host'

  # Prometheus Self-Monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
    scrape_interval: 15s

  # Docker Container Metrics (if using cAdvisor)
  - job_name: 'cadvisor'
    static_configs:
      - targets: ['cadvisor:8080']
    scrape_interval: 30s
    metrics_path: /metrics

  # AlertManager
  - job_name: 'alertmanager'
    static_configs:
      - targets: ['alertmanager:9093']
    scrape_interval: 30s

  # Grafana
  - job_name: 'grafana'
    static_configs:
      - targets: ['grafana:3000']
    scrape_interval: 30s
    metrics_path: /metrics

# Remote write configuration (optional - for long-term storage)
# remote_write:
#   - url: "https://prometheus.yourdomain.com/api/v1/write"
#     basic_auth:
#       username: "admin"
#       password: "password"
```

### Prometheus Alerting Rules

Create `monitoring/rules/radarr.yml`:

```yaml
# monitoring/rules/radarr.yml
groups:
  - name: radarr.rules
    rules:
      # Application Health Rules
      - alert: RadarrDown
        expr: up{job="radarr-go"} == 0
        for: 2m
        labels:
          severity: critical
          service: radarr
        annotations:
          summary: "Radarr Go is down"
          description: "Radarr Go has been down for more than 2 minutes."

      - alert: RadarrHighErrorRate
        expr: rate(radarr_http_requests_total{status=~"5.."}[5m]) / rate(radarr_http_requests_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
          service: radarr
        annotations:
          summary: "High error rate in Radarr Go"
          description: "Error rate is {{ $value | humanizePercentage }} over the last 5 minutes."

      - alert: RadarrHighResponseTime
        expr: histogram_quantile(0.95, rate(radarr_http_request_duration_seconds_bucket[5m])) > 2
        for: 5m
        labels:
          severity: warning
          service: radarr
        annotations:
          summary: "High response times in Radarr Go"
          description: "95th percentile response time is {{ $value }}s over the last 5 minutes."

      - alert: RadarrHighMemoryUsage
        expr: radarr_memory_usage_bytes / radarr_memory_limit_bytes > 0.8
        for: 10m
        labels:
          severity: warning
          service: radarr
        annotations:
          summary: "High memory usage in Radarr Go"
          description: "Memory usage is {{ $value | humanizePercentage }} of available memory."

      - alert: RadarrDatabaseConnectionFailure
        expr: radarr_database_connections_failed_total > 0
        for: 1m
        labels:
          severity: critical
          service: radarr
        annotations:
          summary: "Database connection failures in Radarr Go"
          description: "{{ $value }} database connection failures in the last minute."

  - name: database.rules
    rules:
      # PostgreSQL Rules
      - alert: PostgreSQLDown
        expr: up{job="postgres"} == 0
        for: 1m
        labels:
          severity: critical
          service: database
        annotations:
          summary: "PostgreSQL is down"
          description: "PostgreSQL has been down for more than 1 minute."

      - alert: PostgreSQLHighConnections
        expr: pg_stat_activity_count / pg_settings_max_connections > 0.8
        for: 5m
        labels:
          severity: warning
          service: database
        annotations:
          summary: "High database connections"
          description: "Database connection usage is {{ $value | humanizePercentage }}."

      - alert: PostgreSQLSlowQueries
        expr: rate(pg_stat_statements_mean_time_seconds[5m]) > 1
        for: 5m
        labels:
          severity: warning
          service: database
        annotations:
          summary: "Slow database queries detected"
          description: "Average query time is {{ $value }}s over the last 5 minutes."

      - alert: PostgreSQLReplicationLag
        expr: pg_stat_replication_lag_seconds > 30
        for: 2m
        labels:
          severity: critical
          service: database
        annotations:
          summary: "PostgreSQL replication lag"
          description: "Replication lag is {{ $value }} seconds."

      - alert: PostgreSQLDiskSpaceUsage
        expr: pg_database_size_bytes / (pg_database_size_bytes + node_filesystem_avail_bytes{mountpoint="/var/lib/postgresql/data"}) > 0.9
        for: 5m
        labels:
          severity: warning
          service: database
        annotations:
          summary: "High database disk usage"
          description: "Database disk usage is {{ $value | humanizePercentage }}."

  - name: system.rules
    rules:
      # System Resource Rules
      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
        for: 5m
        labels:
          severity: warning
          service: system
        annotations:
          summary: "High CPU usage on {{ $labels.instance }}"
          description: "CPU usage is {{ $value }}% for more than 5 minutes."

      - alert: HighMemoryUsage
        expr: (node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / node_memory_MemTotal_bytes > 0.9
        for: 5m
        labels:
          severity: warning
          service: system
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"
          description: "Memory usage is {{ $value | humanizePercentage }} for more than 5 minutes."

      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes < 10
        for: 5m
        labels:
          severity: critical
          service: system
        annotations:
          summary: "Disk space low on {{ $labels.instance }}"
          description: "Disk space is {{ $value }}% on device {{ $labels.device }}."

      - alert: DiskIOHigh
        expr: rate(node_disk_io_time_seconds_total[5m]) > 0.5
        for: 10m
        labels:
          severity: warning
          service: system
        annotations:
          summary: "High disk I/O on {{ $labels.instance }}"
          description: "Disk I/O utilization is {{ $value | humanizePercentage }}."

  - name: docker.rules
    rules:
      # Docker Container Rules
      - alert: ContainerHighCPU
        expr: sum(rate(container_cpu_usage_seconds_total{name!=""}[5m])) by (name) * 100 > 80
        for: 5m
        labels:
          severity: warning
          service: docker
        annotations:
          summary: "High CPU usage in container {{ $labels.name }}"
          description: "Container {{ $labels.name }} CPU usage is {{ $value }}%."

      - alert: ContainerHighMemory
        expr: (container_memory_usage_bytes{name!=""} / container_spec_memory_limit_bytes{name!=""}) * 100 > 90
        for: 5m
        labels:
          severity: warning
          service: docker
        annotations:
          summary: "High memory usage in container {{ $labels.name }}"
          description: "Container {{ $labels.name }} memory usage is {{ $value }}%."

      - alert: ContainerKilled
        expr: time() - container_last_seen{name!=""} > 60
        for: 1m
        labels:
          severity: critical
          service: docker
        annotations:
          summary: "Container {{ $labels.name }} killed"
          description: "Container {{ $labels.name }} has disappeared."
```

## AlertManager Configuration

Create `monitoring/alertmanager.yml`:

```yaml
# monitoring/alertmanager.yml
global:
  smtp_smarthost: '${SMTP_HOST}:${SMTP_PORT}'
  smtp_from: '${SMTP_FROM}'
  smtp_auth_username: '${SMTP_USER}'
  smtp_auth_password: '${SMTP_PASSWORD}'
  smtp_require_tls: true

templates:
  - '/etc/alertmanager/templates/*.tmpl'

route:
  group_by: ['alertname', 'service']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'default-receiver'
  routes:
    - match:
        severity: critical
      receiver: 'critical-alerts'
      group_wait: 10s
      group_interval: 5m
      repeat_interval: 30m

    - match:
        severity: warning
      receiver: 'warning-alerts'
      group_wait: 30s
      group_interval: 10m
      repeat_interval: 2h

    - match:
        service: radarr
      receiver: 'radarr-alerts'
      group_by: ['alertname']
      group_wait: 10s

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'instance']

receivers:
  - name: 'default-receiver'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#alerts'
        username: 'AlertManager'
        title: '🚨 {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        text: >-
          {{ range .Alerts }}
          *Alert:* {{ .Annotations.summary }}
          *Description:* {{ .Annotations.description }}
          *Severity:* {{ .Labels.severity }}
          *Service:* {{ .Labels.service }}
          {{ end }}

  - name: 'critical-alerts'
    email_configs:
      - to: '${ADMIN_EMAIL}'
        subject: '🚨 CRITICAL ALERT: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        html: |
          <h2>Critical Alert</h2>
          {{ range .Alerts }}
          <p><strong>Alert:</strong> {{ .Annotations.summary }}</p>
          <p><strong>Description:</strong> {{ .Annotations.description }}</p>
          <p><strong>Severity:</strong> {{ .Labels.severity }}</p>
          <p><strong>Service:</strong> {{ .Labels.service }}</p>
          <p><strong>Started:</strong> {{ .StartsAt }}</p>
          {{ end }}

    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#critical-alerts'
        username: 'AlertManager'
        title: '🔥 CRITICAL: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        text: >-
          <!channel> Critical alert detected!
          {{ range .Alerts }}
          *Alert:* {{ .Annotations.summary }}
          *Description:* {{ .Annotations.description }}
          *Severity:* {{ .Labels.severity }}
          *Service:* {{ .Labels.service }}
          {{ end }}
        color: 'danger'

    webhook_configs:
      - url: '${DISCORD_WEBHOOK_URL}'
        send_resolved: true

  - name: 'warning-alerts'
    slack_configs:
      - api_url: '${SLACK_WEBHOOK_URL}'
        channel: '#warnings'
        username: 'AlertManager'
        title: '⚠️ WARNING: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        text: >-
          {{ range .Alerts }}
          *Alert:* {{ .Annotations.summary }}
          *Description:* {{ .Annotations.description }}
          *Severity:* {{ .Labels.severity }}
          *Service:* {{ .Labels.service }}
          {{ end }}
        color: 'warning'

  - name: 'radarr-alerts'
    email_configs:
      - to: '${RADARR_ADMIN_EMAIL}'
        subject: 'Radarr Alert: {{ range .Alerts }}{{ .Annotations.summary }}{{ end }}'
        html: |
          <h2>Radarr Go Alert</h2>
          {{ range .Alerts }}
          <p><strong>Alert:</strong> {{ .Annotations.summary }}</p>
          <p><strong>Description:</strong> {{ .Annotations.description }}</p>
          <p><strong>Severity:</strong> {{ .Labels.severity }}</p>
          <p><strong>Started:</strong> {{ .StartsAt }}</p>
          {{ end }}
          <p><a href="https://grafana.yourdomain.com/d/radarr-dashboard/radarr-go">View Dashboard</a></p>
```

## Loki Configuration

Create `monitoring/loki.yml`:

```yaml
# monitoring/loki.yml
auth_enabled: false

server:
  http_listen_port: 3100
  grpc_listen_port: 9096

common:
  path_prefix: /loki
  storage:
    filesystem:
      chunks_directory: /loki/chunks
      rules_directory: /loki/rules
  replication_factor: 1
  ring:
    instance_addr: 127.0.0.1
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

limits_config:
  retention_period: 744h  # 31 days
  enforce_metric_name: false
  reject_old_samples: true
  reject_old_samples_max_age: 168h
  ingestion_rate_mb: 16
  ingestion_burst_size_mb: 32

chunk_store_config:
  max_look_back_period: 0s

table_manager:
  retention_deletes_enabled: true
  retention_period: 744h
```

## Promtail Configuration

Create `monitoring/promtail.yml`:

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
  # Docker container logs
  - job_name: containers
    static_configs:
      - targets:
          - localhost
        labels:
          job: containerlogs
          __path__: /var/lib/docker/containers/*/*log

    pipeline_stages:
      - json:
          expressions:
            output: log
            stream: stream
            attrs:
      - json:
          source: attrs
          expressions:
            tag:
      - regex:
          source: tag
          expression: (?P<container_name>(?:[^|]*))\|
      - timestamp:
          source: time
          format: RFC3339Nano
      - labels:
          stream:
          container_name:
      - output:
          source: output

  # System logs
  - job_name: syslog
    static_configs:
      - targets:
          - localhost
        labels:
          job: syslog
          __path__: /var/log/syslog

  # Nginx access logs
  - job_name: nginx
    static_configs:
      - targets:
          - localhost
        labels:
          job: nginx
          __path__: /var/log/nginx/*log

    pipeline_stages:
      - match:
          selector: '{job="nginx"}'
          stages:
            - regex:
                expression: '^(?P<remote_addr>[\w\.]+) - (?P<remote_user>[^ ]*) \[(?P<timestamp>.*)\] "(?P<method>[^ ]*) (?P<path>[^ ]*) (?P<protocol>[^ ]*)" (?P<status>[\d]+) (?P<bytes_sent>[\d]+) "(?P<http_referer>[^"]*)" "(?P<http_user_agent>[^"]*)".*'
            - timestamp:
                source: timestamp
                format: 02/Jan/2006:15:04:05 -0700
            - labels:
                method:
                status:
                path:

  # Radarr Go application logs (if writing to files)
  - job_name: radarr
    static_configs:
      - targets:
          - localhost
        labels:
          job: radarr
          __path__: /opt/radarr/logs/*.log

    pipeline_stages:
      - json:
          expressions:
            timestamp: time
            level: level
            message: msg
            logger: logger
      - timestamp:
          source: timestamp
          format: RFC3339
      - labels:
          level:
          logger:
```

## Grafana Configuration

### Grafana Provisioning

Create `monitoring/grafana/provisioning/datasources/datasources.yml`:

```yaml
# monitoring/grafana/provisioning/datasources/datasources.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    orgId: 1
    uid: prometheus
    url: http://prometheus:9090
    basicAuth: false
    isDefault: true
    version: 1
    editable: false
    jsonData:
      httpMethod: POST
      queryTimeout: 60s
      timeInterval: 15s

  - name: Loki
    type: loki
    access: proxy
    orgId: 1
    uid: loki
    url: http://loki:3100
    basicAuth: false
    version: 1
    editable: false
    jsonData:
      maxLines: 1000
      derivedFields:
        - datasourceUid: prometheus
          matcherRegex: "trace_id=(\\w+)"
          name: TraceID
          url: "$${__value.raw}"
```

Create `monitoring/grafana/provisioning/dashboards/dashboards.yml`:

```yaml
# monitoring/grafana/provisioning/dashboards/dashboards.yml
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
```

### Radarr Go Dashboard

Create `monitoring/grafana/dashboards/radarr-dashboard.json`:

```json
{
  "dashboard": {
    "id": null,
    "title": "Radarr Go Monitoring",
    "tags": ["radarr", "go", "monitoring"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Application Status",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=\"radarr-go\"}",
            "legendFormat": "Status"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "mappings": [
              {
                "options": {
                  "0": {
                    "text": "DOWN",
                    "color": "red"
                  },
                  "1": {
                    "text": "UP",
                    "color": "green"
                  }
                },
                "type": "value"
              }
            ]
          }
        },
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 0,
          "y": 0
        }
      },
      {
        "id": 2,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(radarr_http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 6,
          "y": 0
        }
      },
      {
        "id": 3,
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(radarr_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(radarr_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 8
        }
      },
      {
        "id": 4,
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "radarr_memory_usage_bytes",
            "legendFormat": "Memory Usage"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 12,
          "y": 8
        }
      },
      {
        "id": 5,
        "title": "Database Connections",
        "type": "graph",
        "targets": [
          {
            "expr": "radarr_database_connections_active",
            "legendFormat": "Active Connections"
          },
          {
            "expr": "radarr_database_connections_idle",
            "legendFormat": "Idle Connections"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 12,
          "x": 0,
          "y": 16
        }
      },
      {
        "id": 6,
        "title": "Movies in Library",
        "type": "stat",
        "targets": [
          {
            "expr": "radarr_movies_total",
            "legendFormat": "Total Movies"
          }
        ],
        "gridPos": {
          "h": 8,
          "w": 6,
          "x": 12,
          "y": 16
        }
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

## Monitoring Setup Script

Create `scripts/setup-monitoring.sh`:

```bash
#!/bin/bash
# setup-monitoring.sh - Complete monitoring stack setup

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
MONITORING_DIR="${PROJECT_ROOT}/monitoring"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() { echo -e "${GREEN}[$(date +'%H:%M:%S')] $1${NC}"; }
warn() { echo -e "${YELLOW}[$(date +'%H:%M:%S')] WARNING: $1${NC}"; }
error() { echo -e "${RED}[$(date +'%H:%M:%S')] ERROR: $1${NC}"; exit 1; }

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."

    command -v docker >/dev/null 2>&1 || error "Docker not installed"
    command -v docker compose >/dev/null 2>&1 || command -v docker-compose >/dev/null 2>&1 || error "Docker Compose not installed"

    # Check disk space (minimum 10GB)
    local free_space=$(df / | awk 'NR==2 {print $4}')
    [ "$free_space" -gt 10000000 ] || warn "Low disk space: ${free_space}KB free"

    log "Prerequisites check passed"
}

# Setup monitoring directories
setup_directories() {
    log "Setting up monitoring directories..."

    mkdir -p "$MONITORING_DIR"/{prometheus,grafana,alertmanager,loki}/data
    mkdir -p "$MONITORING_DIR"/grafana/{dashboards,provisioning/{datasources,dashboards}}
    mkdir -p "$MONITORING_DIR"/prometheus/rules

    # Create bind mount directories
    sudo mkdir -p /opt/monitoring/{prometheus,grafana,alertmanager,loki}
    sudo chown -R 1000:1000 /opt/monitoring

    log "Monitoring directories created"
}

# Generate configuration files
generate_configs() {
    log "Generating monitoring configuration files..."

    # Check if configuration files exist
    if [[ -f "$MONITORING_DIR/prometheus.yml" ]]; then
        warn "Monitoring configuration already exists, skipping generation"
        return
    fi

    # Generate environment template
    cat > "$MONITORING_DIR/.env.template" << 'EOF'
# Monitoring Environment Configuration
GRAFANA_ADMIN_PASSWORD=admin-password-change-me
GRAFANA_DB_PASSWORD=grafana-db-password

# SMTP Configuration for Grafana
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=your-email@gmail.com

# AlertManager Configuration
ADMIN_EMAIL=admin@yourdomain.com
RADARR_ADMIN_EMAIL=radarr-admin@yourdomain.com

# Slack Notifications
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Discord Notifications (optional)
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/YOUR/WEBHOOK/URL

# Database
POSTGRES_PASSWORD=your-postgres-password
EOF

    log "Configuration template generated at $MONITORING_DIR/.env.template"
    log "Please copy this to .env and configure your settings:"
    log "  cp $MONITORING_DIR/.env.template $MONITORING_DIR/.env"
}

# Deploy monitoring stack
deploy_monitoring() {
    log "Deploying monitoring stack..."

    cd "$MONITORING_DIR"

    # Check if .env file exists
    if [[ ! -f ".env" ]]; then
        warn "Environment file not found, using template values"
        cp .env.template .env
    fi

    # Deploy the stack
    docker compose -f docker-compose.monitoring.yml up -d

    # Wait for services to be ready
    log "Waiting for services to start..."
    sleep 30

    # Check service health
    local services=("prometheus" "grafana" "alertmanager" "loki")
    local failed_services=()

    for service in "${services[@]}"; do
        if ! docker compose -f docker-compose.monitoring.yml ps "$service" | grep -q "healthy\|Up"; then
            failed_services+=("$service")
        fi
    done

    if [ ${#failed_services[@]} -gt 0 ]; then
        error "Failed services: ${failed_services[*]}"
    fi

    log "Monitoring stack deployed successfully!"
}

# Show service URLs
show_urls() {
    log "Monitoring services are available at:"
    echo ""
    echo "  Grafana:         http://localhost:3000 (admin/admin)"
    echo "  Prometheus:      http://localhost:9090"
    echo "  AlertManager:    http://localhost:9093"
    echo "  Loki:            http://localhost:3100"
    echo ""
    log "Import the Radarr Go dashboard from monitoring/grafana/dashboards/radarr-dashboard.json"
}

# Verify deployment
verify_deployment() {
    log "Verifying monitoring deployment..."

    local checks=0
    local passed=0

    # Check Prometheus targets
    ((checks++))
    if curl -s http://localhost:9090/api/v1/targets | grep -q '"health":"up"'; then
        log "✓ Prometheus targets are healthy"
        ((passed++))
    else
        warn "✗ Some Prometheus targets are down"
    fi

    # Check Grafana API
    ((checks++))
    if curl -s -u admin:admin http://localhost:3000/api/health | grep -q '"database":"ok"'; then
        log "✓ Grafana is healthy"
        ((passed++))
    else
        warn "✗ Grafana health check failed"
    fi

    # Check AlertManager API
    ((checks++))
    if curl -s http://localhost:9093/api/v1/status | grep -q '"status":"success"'; then
        log "✓ AlertManager is healthy"
        ((passed++))
    else
        warn "✗ AlertManager health check failed"
    fi

    log "Verification completed: $passed/$checks checks passed"

    if [ "$passed" -eq "$checks" ]; then
        log "All monitoring services are healthy!"
        return 0
    else
        warn "Some monitoring services have issues"
        return 1
    fi
}

# Main execution
main() {
    local command="${1:-deploy}"

    case "$command" in
        "deploy")
            check_prerequisites
            setup_directories
            generate_configs
            deploy_monitoring
            show_urls
            verify_deployment
            ;;
        "verify")
            verify_deployment
            ;;
        "status")
            cd "$MONITORING_DIR"
            docker compose -f docker-compose.monitoring.yml ps
            ;;
        "logs")
            cd "$MONITORING_DIR"
            docker compose -f docker-compose.monitoring.yml logs -f "${2:-}"
            ;;
        "stop")
            cd "$MONITORING_DIR"
            docker compose -f docker-compose.monitoring.yml down
            ;;
        "restart")
            cd "$MONITORING_DIR"
            docker compose -f docker-compose.monitoring.yml restart "${2:-}"
            ;;
        "clean")
            warn "This will remove all monitoring data!"
            read -p "Are you sure? (yes/no): " -r
            if [[ $REPLY == "yes" ]]; then
                cd "$MONITORING_DIR"
                docker compose -f docker-compose.monitoring.yml down -v
                sudo rm -rf /opt/monitoring
                log "Monitoring stack cleaned"
            fi
            ;;
        *)
            echo "Usage: $0 {deploy|verify|status|logs|stop|restart|clean}"
            echo ""
            echo "Commands:"
            echo "  deploy   - Deploy complete monitoring stack"
            echo "  verify   - Verify deployment health"
            echo "  status   - Show service status"
            echo "  logs     - Show logs (optionally specify service)"
            echo "  stop     - Stop all monitoring services"
            echo "  restart  - Restart services (optionally specify service)"
            echo "  clean    - Remove all monitoring data and containers"
            exit 1
            ;;
    esac
}

main "$@"
```

## Alert Testing

Create `scripts/test-alerts.sh`:

```bash
#!/bin/bash
# test-alerts.sh - Test monitoring alerts

set -euo pipefail

# Test configurations
PROMETHEUS_URL="http://localhost:9090"
ALERTMANAGER_URL="http://localhost:9093"

log() { echo "[$(date +'%H:%M:%S')] $1"; }
error() { echo "[$(date +'%H:%M:%S')] ERROR: $1"; exit 1; }

# Test Prometheus targets
test_prometheus_targets() {
    log "Testing Prometheus targets..."

    local targets_response=$(curl -s "${PROMETHEUS_URL}/api/v1/targets" | jq -r '.data.activeTargets[] | select(.health != "up") | .scrapeUrl')

    if [ -n "$targets_response" ]; then
        log "⚠️  Unhealthy targets found:"
        echo "$targets_response"
    else
        log "✅ All Prometheus targets are healthy"
    fi
}

# Test AlertManager rules
test_alert_rules() {
    log "Testing alert rules..."

    local rules_response=$(curl -s "${PROMETHEUS_URL}/api/v1/rules" | jq -r '.data.groups[].rules[] | select(.type == "alerting") | select(.state != "inactive") | .name')

    if [ -n "$rules_response" ]; then
        log "🔥 Active alerts:"
        echo "$rules_response"
    else
        log "✅ No active alerts"
    fi
}

# Send test alert
send_test_alert() {
    log "Sending test alert to AlertManager..."

    curl -XPOST "${ALERTMANAGER_URL}/api/v1/alerts" -H "Content-Type: application/json" -d '[
      {
        "labels": {
          "alertname": "TestAlert",
          "service": "test",
          "severity": "warning",
          "instance": "localhost"
        },
        "annotations": {
          "summary": "This is a test alert",
          "description": "This alert is generated for testing purposes"
        },
        "startsAt": "'"$(date -u +%Y-%m-%dT%H:%M:%SZ)"'",
        "endsAt": "'"$(date -u -d '+5 minutes' +%Y-%m-%dT%H:%M:%SZ)"'"
      }
    ]'

    log "Test alert sent. Check your notification channels."
}

# Test query performance
test_query_performance() {
    log "Testing query performance..."

    local queries=(
        "up"
        "rate(radarr_http_requests_total[5m])"
        "histogram_quantile(0.95, rate(radarr_http_request_duration_seconds_bucket[5m]))"
        "pg_stat_activity_count"
    )

    for query in "${queries[@]}"; do
        local start_time=$(date +%s%N)
        curl -s "${PROMETHEUS_URL}/api/v1/query?query=$(echo "$query" | sed 's/ /%20/g')" > /dev/null
        local end_time=$(date +%s%N)
        local duration=$(( (end_time - start_time) / 1000000 ))

        log "Query '$query': ${duration}ms"
    done
}

# Main execution
case "${1:-all}" in
    "targets") test_prometheus_targets ;;
    "rules") test_alert_rules ;;
    "alert") send_test_alert ;;
    "performance") test_query_performance ;;
    "all")
        test_prometheus_targets
        test_alert_rules
        test_query_performance
        ;;
    *)
        echo "Usage: $0 {targets|rules|alert|performance|all}"
        exit 1
        ;;
esac
```

## Next Steps

This completes the monitoring and alerting setup documentation. The monitoring stack includes:

- ✅ **Prometheus** for metrics collection
- ✅ **Grafana** for visualization and dashboards
- ✅ **AlertManager** for alert routing and notifications
- ✅ **Loki** for log aggregation
- ✅ **Automated setup scripts** for easy deployment
- ✅ **Comprehensive alerting rules** for application and infrastructure
- ✅ **Testing utilities** for validation

The next sections will cover:

1. **Performance Tuning Guide** - Database optimization, Go runtime tuning, and scaling
2. **Security Hardening** - Network security, authentication, and container security
3. **Automated Scripts** - Additional deployment automation and operational tools

This monitoring setup provides complete observability for production Radarr Go deployments with enterprise-grade alerting and visualization capabilities.
