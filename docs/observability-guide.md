# Observability Guide — Monitoring, Logging, and Tracing

This guide covers the three pillars of observability for services running on the IDP platform.

---

## 1. The Three Pillars

```
┌──────────────────────────────────────────────────────────────┐
│                     Observability                              │
│                                                               │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │
│  │   Metrics     │  │    Logs      │  │     Traces       │   │
│  │  (Prometheus) │  │   (Loki)     │  │    (Jaeger)      │   │
│  │              │  │              │  │                  │   │
│  │  "What is    │  │  "What       │  │  "How does a     │   │
│  │   happening?"│  │   happened?" │  │   request flow?" │   │
│  └──────┬───────┘  └──────┬───────┘  └────────┬─────────┘   │
│         │                 │                    │              │
│         └─────────────────┼────────────────────┘              │
│                           ▼                                   │
│                    ┌──────────────┐                           │
│                    │   Grafana    │                           │
│                    │ (Dashboards) │                           │
│                    └──────────────┘                           │
└──────────────────────────────────────────────────────────────┘
```

| Pillar | Tool | Purpose | Question It Answers |
|--------|------|---------|-------------------|
| **Metrics** | Prometheus + Grafana | Numeric measurements over time | How many requests? What's the error rate? |
| **Logs** | Loki + Promtail | Structured event records | What happened at 3:42 PM? Why did this request fail? |
| **Traces** | Jaeger / Tempo | Request flow across services | Where is the bottleneck? Which service is slow? |

---

## 2. Metrics (Prometheus + Grafana)

### 2.1 How Prometheus Works

```
Go Service                     Prometheus                  Grafana
──────────                     ──────────                  ───────
Exposes /metrics  ◄───scrape──  Scrapes every 15s  ───────►  Dashboards
(port 9090)                    Stores time series           Queries PromQL
                               Evaluates alert rules        Shows graphs
```

### 2.2 Key Metrics for Go Services

Your Go service should expose these metrics (the go-service-template includes them via `go-zero` and custom middleware):

**RED Method** (for every service):

| Metric | Type | Description |
|--------|------|-------------|
| `http_requests_total` | Counter | Total HTTP requests |
| `http_request_duration_seconds` | Histogram | Request latency |
| `http_requests_errors_total` | Counter | Failed requests (5xx) |
| `grpc_server_handled_total` | Counter | Total gRPC calls |
| `grpc_server_handling_seconds` | Histogram | gRPC latency |

**USE Method** (for resources):

| Metric | Type | Description |
|--------|------|-------------|
| `go_goroutines` | Gauge | Number of goroutines |
| `go_memstats_alloc_bytes` | Gauge | Memory allocated |
| `process_cpu_seconds_total` | Counter | CPU usage |

**Business Metrics**:

| Metric | Type | Description |
|--------|------|-------------|
| `db_queries_total` | Counter | Database queries executed |
| `db_query_duration_seconds` | Histogram | Database query latency |
| `cache_hits_total` | Counter | Redis cache hits |
| `cache_misses_total` | Counter | Redis cache misses |
| `kafka_messages_produced_total` | Counter | Kafka messages sent |
| `kafka_messages_consumed_total` | Counter | Kafka messages processed |
| `kafka_consumer_lag` | Gauge | Kafka consumer lag |

### 2.3 Exposing Metrics in Go

```go
// internal/middleware/metrics.go
package middleware

import (
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
        },
        []string{"method", "path"},
    )
)

// MetricsHandler returns the Prometheus metrics endpoint
func MetricsHandler() http.Handler {
    return promhttp.Handler()
}
```

### 2.4 ServiceMonitor (Prometheus Discovery)

```yaml
# k8s/service-monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: user-service
  namespace: services
  labels:
    app: user-service
spec:
  selector:
    matchLabels:
      app: user-service
  endpoints:
    - port: http
      path: /metrics
      interval: 15s
```

### 2.5 Grafana Dashboards

#### Service Overview Dashboard

Key panels:
- **Request Rate** — `rate(http_requests_total[5m])`
- **Error Rate** — `rate(http_requests_errors_total[5m]) / rate(http_requests_total[5m])`
- **Latency P50/P95/P99** — `histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))`
- **Active Goroutines** — `go_goroutines`
- **Memory Usage** — `go_memstats_alloc_bytes`

#### Database Dashboard

- **Query Rate** — `rate(db_queries_total[5m])`
- **Query Latency** — `histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))`
- **Connection Pool** — `db_connections_active`

#### Cache Dashboard

- **Hit Rate** — `rate(cache_hits_total[5m]) / (rate(cache_hits_total[5m]) + rate(cache_misses_total[5m]))`
- **Latency** — `histogram_quantile(0.95, rate(cache_operation_duration_seconds_bucket[5m]))`

### 2.6 Alerting Rules

```yaml
# infrastructure/monitoring/prometheus/alerts.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: service-alerts
  namespace: monitoring
spec:
  groups:
    - name: service.rules
      rules:
        # High error rate
        - alert: HighErrorRate
          expr: |
            rate(http_requests_errors_total[5m])
            / rate(http_requests_total[5m]) > 0.05
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "High error rate on {{ $labels.service }}"
            description: "Error rate is {{ $value | humanizePercentage }}"

        # High latency
        - alert: HighLatency
          expr: |
            histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High latency on {{ $labels.service }}"
            description: "P95 latency is {{ $value }}s"

        # Pod restarts
        - alert: PodRestarts
          expr: |
            increase(kube_pod_container_status_restarts_total[1h]) > 3
          labels:
            severity: warning
          annotations:
            summary: "Pod {{ $labels.pod }} restarting frequently"

        # Database connection failures
        - alert: DatabaseConnectionFailure
          expr: |
            increase(db_connection_errors_total[5m]) > 0
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "Database connection failures on {{ $labels.service }}"
```

---

## 3. Logs (Loki + Promtail)

### 3.1 How Loki Works

```
Go Service          Promtail             Loki              Grafana
──────────          ────────             ────              ───────
stdout/stderr ──►  Collects logs  ──►  Indexes labels  ──►  LogQL queries
(JSON format)      from pods           Stores chunks       Log panels
                   Adds labels         (like Prometheus
                   (namespace, pod)     but for logs)
```

### 3.2 Structured Logging in Go

Use `zerolog` for JSON-structured logs:

```go
// The go-service-template uses zerolog
import "github.com/rs/zerolog/log"

// Good: structured fields
log.Info().
    Str("user_id", userID).
    Str("action", "create").
    Dur("duration", elapsed).
    Msg("User created successfully")

// Output:
// {"level":"info","user_id":"123","action":"create","duration":45.2,"message":"User created successfully","time":"2024-01-15T10:30:00Z"}
```

### 3.3 Log Levels

| Level | When to Use | Example |
|-------|-------------|---------|
| **Error** | Something failed, needs attention | DB connection lost, API call failed |
| **Warn** | Unexpected but handled | Cache miss, retry succeeded |
| **Info** | Normal operations | Request processed, user created |
| **Debug** | Development details | SQL query, cache key, request body |

### 3.4 LogQL Queries (Grafana)

```
# All logs from a service
{namespace="services", app="user-service"}

# Only errors
{namespace="services", app="user-service"} |= "error"

# JSON parsing + filter
{namespace="services", app="user-service"} | json | level="error"

# Count errors per minute
count_over_time({namespace="services", app="user-service"} | json | level="error" [1m])

# Slow requests
{namespace="services", app="user-service"} | json | duration > 1000
```

### 3.5 Log Labels

Promtail automatically adds Kubernetes labels:

| Label | Source | Example |
|-------|--------|---------|
| `namespace` | K8s namespace | `services` |
| `pod` | K8s pod name | `user-service-abc123` |
| `container` | Container name | `app` |
| `app` | K8s label `app` | `user-service` |
| `stream` | stdout/stderr | `stdout` |

---

## 4. Traces (Jaeger)

### 4.1 How Distributed Tracing Works

When a request flows through multiple services, tracing connects the dots:

```
Browser → API Gateway → User Service → PostgreSQL
                     → Cache (Redis)
                     → Order Service → Kafka

Trace ID: abc-123
├── Span: API Gateway (2ms)
├── Span: User Service (45ms)
│   ├── Span: PostgreSQL Query (12ms)
│   └── Span: Redis GET (3ms)
└── Span: Order Service (30ms)
    └── Span: Kafka Produce (5ms)
```

### 4.2 OpenTelemetry in Go

```go
// internal/middleware/tracing.go
package middleware

import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/jaeger"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func InitTracer(serviceName, jaegerEndpoint string) (*sdktrace.TracerProvider, error) {
    exporter, err := jaeger.New(
        jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerEndpoint)),
    )
    if err != nil {
        return nil, err
    }

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource.NewWithAttributes(
            semconv.SchemaURL,
            semconv.ServiceNameKey.String(serviceName),
        )),
    )

    otel.SetTracerProvider(tp)
    return tp, nil
}
```

### 4.3 Adding Spans to Code

```go
import "go.opentelemetry.io/otel"

func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    ctx, span := otel.Tracer("user-service").Start(ctx, "GetUser")
    defer span.End()

    // Check cache
    ctx, cacheSpan := otel.Tracer("user-service").Start(ctx, "Redis.GET")
    user, err := s.cache.Get(ctx, id)
    cacheSpan.End()

    if err == nil {
        return user, nil
    }

    // Query database
    ctx, dbSpan := otel.Tracer("user-service").Start(ctx, "PostgreSQL.SELECT")
    user, err = s.repo.FindByID(ctx, id)
    dbSpan.End()

    return user, err
}
```

### 4.4 Viewing Traces in Jaeger

```bash
# Port-forward Jaeger UI
kubectl port-forward svc/jaeger-query -n monitoring 16686:16686
# Open: http://localhost:16686
```

In Jaeger UI:
1. Select service: `user-service`
2. Click "Find Traces"
3. Click on a trace to see the full request flow
4. See latency breakdown per span

---

## 5. Grafana Dashboard Setup

### 5.1 Data Sources

| Source | URL | Purpose |
|--------|-----|---------|
| Prometheus | `http://monitoring-kube-prometheus-prometheus.monitoring:9090` | Metrics |
| Loki | `http://loki.monitoring:3100` | Logs |
| Jaeger | `http://jaeger-query.monitoring:16686` | Traces |

### 5.2 Service Overview Dashboard Template

Create a dashboard with variables:
- `$namespace` — Kubernetes namespace
- `$service` — Service name

Panels:
```
┌──────────────────────────────────────────────────────┐
│                Service: $service                      │
├──────────────┬──────────────┬────────────────────────┤
│ Request Rate │  Error Rate  │  P95 Latency           │
│  1.2k/s      │  0.1%        │  45ms                  │
├──────────────┴──────────────┴────────────────────────┤
│           Request Rate Over Time (graph)              │
├──────────────────────────────────────────────────────┤
│           Latency Distribution (heatmap)              │
├──────────────┬───────────────────────────────────────┤
│  Pod Status  │  Resource Usage (CPU/Memory)           │
├──────────────┴───────────────────────────────────────┤
│           Recent Logs (Loki panel)                    │
└──────────────────────────────────────────────────────┘
```

### 5.3 Dashboard JSON Template

```json
{
  "dashboard": {
    "title": "Service Overview - $service",
    "templating": {
      "list": [
        {
          "name": "namespace",
          "type": "query",
          "query": "label_values(kube_pod_info, namespace)"
        },
        {
          "name": "service",
          "type": "query",
          "query": "label_values(http_requests_total{namespace=\"$namespace\"}, service)"
        }
      ]
    },
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{namespace=\"$namespace\", service=\"$service\"}[5m]))"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{namespace=\"$namespace\", service=\"$service\", status=~\"5..\"}[5m])) / sum(rate(http_requests_total{namespace=\"$namespace\", service=\"$service\"}[5m]))"
          }
        ]
      },
      {
        "title": "P95 Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{namespace=\"$namespace\", service=\"$service\"}[5m])) by (le))"
          }
        ]
      }
    ]
  }
}
```

---

## 6. Troubleshooting Scenarios

### Scenario: Service returning 500 errors

```
Step 1: Check Grafana dashboard
        → Error rate spiked at 14:30

Step 2: Query Loki for errors
        {app="user-service"} | json | level="error" | ts >= "14:25"
        → "database connection refused"

Step 3: Check database metrics
        → db_connections_active = 0, db_connection_errors > 0

Step 4: Check database pod
        kubectl get pods -n databases -l app=user-service-db
        → Pod in CrashLoopBackOff

Step 5: Check Jaeger traces
        → All traces show database span failing

Root cause: Database pod crashed due to OOM
Fix: Increase database memory limits
```

### Scenario: High latency

```
Step 1: Grafana shows P95 > 2s

Step 2: Jaeger trace shows:
        User Service (2.1s)
        ├── Redis GET (2ms) ✓
        ├── PostgreSQL SELECT (1.8s) ← slow
        └── Response (5ms) ✓

Step 3: Check database dashboard
        → Query duration spike correlates with missing index

Fix: Add database index on the slow query column
```

---

## 7. Configuration Summary

### Environment Variables for Go Services

```bash
# Metrics
APP_METRICS_ENABLED=true
APP_METRICS_PORT=9090
APP_METRICS_PATH=/metrics

# Tracing
APP_TRACING_ENABLED=true
APP_TRACING_ENDPOINT=http://jaeger-collector.monitoring:14268/api/traces
APP_TRACING_SAMPLE_RATE=1.0

# Logging
APP_LOG_LEVEL=info
APP_LOG_FORMAT=json
```

---

## 8. File Organization

```
infrastructure/monitoring/
├── prometheus/
│   ├── alerts.yaml              # Alert rules
│   └── service-monitors/        # Per-service monitors
│       └── template.yaml
├── grafana/
│   └── dashboards/
│       ├── service-overview.json
│       ├── postgresql.json
│       ├── redis.json
│       └── kafka.json
└── loki/
    └── loki-config.yaml
```

---

## Next Steps

1. Install monitoring stack → [Setup Guide](./setup-guide.md) Section 5
2. Add metrics to Go service → use Prometheus client
3. Configure ServiceMonitor → `k8s/service-monitor.yaml`
4. Create Grafana dashboards → `infrastructure/monitoring/grafana/dashboards/`
5. Set up alerting rules → `infrastructure/monitoring/prometheus/alerts.yaml`

Related: [Architecture Plan](./idp-architecture-plan.md) | [Setup Guide](./setup-guide.md) | [Developer Workflow](./developer-workflow.md)
