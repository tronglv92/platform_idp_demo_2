# Developer Workflow: Creating a New Service

This document provides a detailed, step-by-step flow of what happens when a developer creates a new service using the Internal Developer Platform.

---

## Overview

**Total Time**: ~5 minutes  
**User Interaction**: Minimal (fill out a form)  
**Automation**: Complete infrastructure provisioning, deployment, and monitoring setup

---

## Flow Diagram

```
Developer → Backstage → GitHub → Crossplane → Kubernetes → Running Service
    ↓           ↓          ↓          ↓            ↓              ↓
  Form      Scaffold   Create    Provision    Deploy        Monitor
  Input     Code       Repo      Infra        Service       & Alert
```

---

## Detailed Step-by-Step Flow

### Step 1: Developer Accesses Backstage Portal

**Action**: Developer navigates to Backstage portal

```
URL: https://backstage.yourcompany.com
```

**What the developer sees**:
- Service catalog showing all existing services
- "Create Component" button
- Documentation and API references

**Time**: 10 seconds

---

### Step 2: Select Service Template

**Action**: Developer clicks "Create Component" and selects template

**Available templates**:
- ✅ **Go Service with PostgreSQL and Redis and Kafka** (most common)
- Go Service (basic, no database)
- Go Service with PostgreSQL only
- Go Service with Redis only
- Go Service with Kafka only

**What happens**:
- Backstage loads the template form
- Form is pre-populated with sensible defaults

**Time**: 10 seconds

---

### Step 3: Fill Out Service Information Form

**Action**: Developer fills out the form with service details

**Form Fields**:

#### Section 1: Service Information
```yaml
Service Name: my-awesome-api
  - Must be lowercase with hyphens
  - Example: user-service, payment-api, notification-service

Description: API for managing user profiles and preferences
  - Brief description of what the service does

Owner: backend-team
  - Team or person responsible
  - Used for notifications and access control
```

#### Section 2: Configuration
```yaml
Kubernetes Namespace: services
  - Where the service will be deployed
  - Options: services, staging, production

Number of Replicas: 3
  - How many pods to run
  - Range: 1-10
  - Default: 3 (for high availability)

Database Size: small
  - Options: small, medium, large
  - small: 20GB storage, 2 CPU, 4GB RAM
  - medium: 100GB storage, 4 CPU, 8GB RAM
  - large: 500GB storage, 8 CPU, 16GB RAM

Redis Size: small
  - Options: small, medium, large
  - small: 1GB memory
  - medium: 4GB memory
  - large: 16GB memory

Kafka Size: small
  - Options: small, medium, large
  - small: 3 nodes, 10GB storage
  - medium: 3 nodes, 100GB storage
  - large: 5 nodes, 500GB storage
```

#### Section 3: Repository
```yaml
Repository Location: github.com/yourorg/my-awesome-api
  - GitHub organization and repository name
  - Repository will be created automatically
```

**Time**: 1-2 minutes

---

### Step 4: Submit and Trigger Automation

**Action**: Developer clicks "Create" button

**What happens immediately**:
1. Backstage validates the form inputs
2. Shows a progress indicator
3. Starts the scaffolding process

**Time**: 5 seconds

---

### Step 5: Backstage Scaffolds the Service

**Action**: Backstage executes scaffolder actions

**Scaffolder Actions**:

#### 5.1: Fetch Template Skeleton
```
- Copies the Go service boilerplate
- Replaces template variables with user inputs
- Generates:
  - Go source code
  - Dockerfile
  - Kubernetes manifests
  - CI/CD pipelines
  - Documentation
```

#### 5.2: Create GitHub Repository
```
- Creates new repository: github.com/yourorg/my-awesome-api
- Sets repository description
- Adds default branch protection rules
- Configures GitHub Actions secrets
```

#### 5.3: Push Code to Repository
```
- Commits all scaffolded files
- Pushes to main branch
- Creates initial commit message:
  "Initial commit: Scaffolded by Backstage"
```

#### 5.4: Create Crossplane Claim
```yaml
# File: crossplane-claim.yaml
apiVersion: platform.example.com/v1alpha1
kind: GoService
metadata:
  name: my-awesome-api
  namespace: services
spec:
  serviceName: my-awesome-api
  namespace: services
  replicas: 3
  image: ghcr.io/yourorg/my-awesome-api:latest
  database:
    enabled: true
    size: small
  redis:
    enabled: true
    size: small
  kafka:
    enabled: true
    size: small
```

#### 5.5: Register in Service Catalog
```
- Creates catalog-info.yaml in repository
- Registers service in Backstage catalog
- Links to GitHub repository
- Links to Kubernetes cluster
```

**Time**: 30 seconds

---

### Step 6: Crossplane Provisions Infrastructure

**Action**: Crossplane detects the new GoService claim and starts provisioning

**What Crossplane does**:

#### 6.1: Provision PostgreSQL Database
```
1. Creates PostgreSQL instance using CloudNativePG operator
2. Configures:
   - Database name: my_awesome_api
   - Storage: 20GB (small size)
   - CPU: 2 cores
   - Memory: 4GB
   - Backup schedule: Daily
3. Generates connection credentials
4. Creates Kubernetes Secret:
   - Name: my-awesome-api-postgres-conn
   - Keys: endpoint, username, password, database
```

#### 6.2: Provision Redis Cache
```
1. Creates Redis instance using Redis operator
2. Configures:
   - Memory: 1GB (small size)
   - Persistence: Enabled
   - High availability: Master + Replica
3. Generates connection credentials
4. Creates Kubernetes Secret:
   - Name: my-awesome-api-redis-conn
   - Keys: endpoint, password
```

#### 6.3: Provision Kafka Cluster
```
1. Creates Kafka cluster using Strimzi operator
2. Configures:
   - Replicas: 3 (small size)
   - Storage: 10GB per node
   - Retention: 7 days
3. Generates connection bootstrap servers
4. Creates Kubernetes Secret:
   - Name: my-awesome-api-kafka-conn
   - Keys: bootstrap-servers, user, password
```

#### 6.4: Create Kubernetes Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-awesome-api
  namespace: services
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-awesome-api
  template:
    spec:
      containers:
      - name: my-awesome-api
        image: ghcr.io/yourorg/my-awesome-api:latest
        env:
        - name: APP_DATABASE_HOST
          valueFrom:
            secretKeyRef:
              name: my-awesome-api-postgres-conn
              key: endpoint
        - name: APP_DATABASE_USER
          valueFrom:
            secretKeyRef:
              name: my-awesome-api-postgres-conn
              key: username
        - name: APP_DATABASE_PASSWORD
          valueFrom:
            secretKeyRef:
              name: my-awesome-api-postgres-conn
              key: password
        - name: APP_REDIS_HOST
          valueFrom:
            secretKeyRef:
              name: my-awesome-api-redis-conn
              key: endpoint
        - name: APP_REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: my-awesome-api-redis-conn
              key: password
        - name: APP_KAFKA_BOOTSTRAP_SERVERS
          valueFrom:
            secretKeyRef:
              name: my-awesome-api-kafka-conn
              key: bootstrap-servers
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
```

#### 6.5: Create Kubernetes Service
```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-awesome-api
  namespace: services
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: my-awesome-api
```

**Time**: 1-2 minutes

---

### Step 7: CI/CD Pipeline Builds and Deploys

**Action**: GitHub Actions workflow is triggered automatically

**CI/CD Pipeline Steps**:

#### 7.1: Lint and Format Check
```bash
- Run golangci-lint
- Check code formatting
- Verify imports
```

#### 7.2: Run Tests
```bash
- Unit tests with coverage
- Integration tests with testcontainers
- Generate coverage report
- Require >80% coverage
```

#### 7.3: Security Scanning
```bash
- Run gosec (Go security checker)
- Run Trivy (container vulnerability scanner)
- Check for known vulnerabilities
- Fail if critical issues found
```

#### 7.4: Build Docker Image
```bash
- Multi-stage Docker build
- Optimize image size
- Tag with git SHA: ghcr.io/yourorg/my-awesome-api:abc123
- Tag with version: ghcr.io/yourorg/my-awesome-api:v1.0.0
- Tag with latest: ghcr.io/yourorg/my-awesome-api:latest
```

#### 7.5: Push to Container Registry
```bash
- Push all tags to GitHub Container Registry
- Verify image push successful
```

#### 7.6: Update Kubernetes Manifests
```bash
- Update deployment.yaml with new image tag
- Commit changes to Git
- Push to repository
```

**Time**: 2-3 minutes

---

### Step 8: ArgoCD Deploys to Kubernetes

**Action**: ArgoCD detects changes and syncs to cluster

**ArgoCD Process**:

#### 8.1: Detect Changes
```
- ArgoCD monitors Git repository
- Detects new Kubernetes manifests
- Detects updated image tag
```

#### 8.2: Sync Application
```
- Applies Kubernetes manifests to cluster
- Creates/updates Deployment
- Creates/updates Service
- Waits for pods to be ready
```

#### 8.3: Health Check
```
- Verifies pods are running
- Checks liveness probe: GET /health
- Checks readiness probe: GET /ready
- Waits for all replicas to be healthy
```

#### 8.4: Update Status
```
- Marks deployment as "Synced"
- Updates Backstage with deployment status
- Sends notification to developer
```

**Time**: 1-2 minutes

---

### Step 9: Service is Running and Monitored

**Action**: Service is now live and automatically monitored

**What's automatically configured**:

#### 9.1: Service Endpoints
```
Internal URL: http://my-awesome-api.services.svc.cluster.local
Health Check: http://my-awesome-api.services.svc.cluster.local/health
Readiness: http://my-awesome-api.services.svc.cluster.local/ready
Metrics: http://my-awesome-api.services.svc.cluster.local/metrics
API: http://my-awesome-api.services.svc.cluster.local/api/v1
```

#### 9.2: Database Connection
```
PostgreSQL: my-awesome-api-postgres.services.svc.cluster.local:5432
Database: my_awesome_api
Credentials: Automatically injected via secrets
Connection Pool: 25 max connections
```

#### 9.3: Redis Connection
```
Redis: my-awesome-api-redis.services.svc.cluster.local:6379
Credentials: Automatically injected via secrets
Connection Pool: 10 connections
```

#### 9.4: Kafka Connection
```
Bootstrap Servers: kafka-cluster-kafka-bootstrap.infra.svc.cluster.local:9092
Topic naming: services.<service-name>.<event-type>
Credentials: Automatically injected via secrets
```

#### 9.5: Monitoring & Observability
```
Prometheus Metrics:
- HTTP request count and duration
- Database query count and duration
- Cache hit/miss ratio
- Go runtime metrics

Grafana Dashboards:
- Service overview dashboard
- Database performance dashboard
- Redis performance dashboard
- Kafka performance dashboard

Logs:
- Structured JSON logs sent to Loki
- Searchable by service name, request ID, level

Alerts:
- High error rate (>5%)
- High latency (p95 > 1s)
- Database connection issues
- Redis connection issues
- Kafka consumer/producer issues
- Pod crashes
```

#### 9.6: Backstage Integration
```
Service Catalog Entry:
- Service name and description
- Owner and team
- Links to:
  - GitHub repository
  - API documentation
  - Grafana dashboards
  - ArgoCD application
  - Kubernetes resources

Kubernetes Plugin:
- Live pod status
- Resource usage (CPU, memory)
- Recent deployments
- Logs viewer
```

**Time**: Immediate (already running)

---

## Complete Timeline

| Step | Action | Time | Cumulative |
|------|--------|------|------------|
| 1 | Access Backstage | 10s | 10s |
| 2 | Select template | 10s | 20s |
| 3 | Fill out form | 1-2 min | 2m 20s |
| 4 | Submit form | 5s | 2m 25s |
| 5 | Backstage scaffolds | 30s | 2m 55s |
| 6 | Crossplane provisions | 1-2 min | 4m 55s |
| 7 | CI/CD builds | 2-3 min | 7m 55s |
| 8 | ArgoCD deploys | 1-2 min | 9m 55s |
| 9 | Service running | Immediate | **~10 min** |

**Total Time: ~10 minutes from form submission to running service** ⚡

---

## What the Developer Gets

After 10 minutes, the developer has:

✅ **Production-ready Go service** with:
- Clean architecture (handler → service → repository)
- PostgreSQL database with migrations
- Redis cache with connection pooling
- Kafka messaging for event-driven architecture
- Structured logging (JSON)
- Prometheus metrics
- Health checks
- Graceful shutdown

✅ **Fully provisioned infrastructure**:
- PostgreSQL database (managed, backed up)
- Redis cache (high availability)
- Kafka cluster (managed, event streaming)
- Kubernetes deployment (3 replicas)
- Load-balanced service endpoint
- Automatic secret management

✅ **Complete CI/CD pipeline**:
- Automated testing
- Security scanning
- Docker image building
- Automated deployment
- Rollback capability

✅ **Monitoring and observability**:
- Prometheus metrics collection
- Grafana dashboards (Service, DB, Redis, Kafka)
- Centralized logging
- Alerting rules

✅ **Documentation**:
- README with setup instructions
- API documentation (OpenAPI)
- Architecture diagrams
- Runbooks

✅ **Developer tools**:
- Local development setup (Docker Compose)
- Makefile with common commands
- VS Code configuration
- Git hooks

---

## Developer Experience Benefits

### Before IDP (Traditional Approach)
```
Time to production: 2-4 weeks
Steps:
1. Set up project structure (1 day)
2. Request database (3-5 days, waiting for ops)
3. Request Redis (2-3 days, waiting for ops)
4. Set up CI/CD (2-3 days)
5. Configure monitoring (1-2 days)
6. Write documentation (1 day)
7. Security review (3-5 days)
8. Deploy to production (1 day)

Total: 2-4 weeks + lots of back-and-forth
```

### With IDP (This Platform)
```
Time to production: 10 minutes
Steps:
1. Fill out form (2 minutes)
2. Wait for automation (8 minutes)

Total: 10 minutes + zero back-and-forth
```

**Productivity Gain: 99% faster** 🚀

---

## What Happens Behind the Scenes

### Automation Components Working Together

```
┌─────────────┐
│  Developer  │
└──────┬──────┘
       │ (fills form)
       ▼
┌─────────────────────────────────────────────┐
│           Backstage Scaffolder              │
│  - Generates code from template             │
│  - Creates GitHub repository                │
│  - Commits and pushes code                  │
│  - Creates Crossplane claim                 │
│  - Registers in catalog                     │
└──────┬──────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│            Crossplane                        │
│  - Reads GoService claim                    │
│  - Provisions PostgreSQL (CloudNativePG)    │
│  - Provisions Redis (Redis Operator)        │
│  - Provisions Kafka (Strimzi Operator)       │
│  - Creates Kubernetes Deployment            │
│  - Creates Kubernetes Service               │
│  - Generates and stores secrets (Vault)     │
└──────┬──────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│          GitHub Actions (CI/CD)              │
│  - Triggered by code push                   │
│  - Runs tests and linting                   │
│  - Scans for security issues                │
│  - Builds Docker image                      │
│  - Pushes to container registry             │
│  - Updates Kubernetes manifests             │
└──────┬──────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│              ArgoCD (GitOps)                 │
│  - Detects manifest changes in Git          │
│  - Syncs to Kubernetes cluster              │
│  - Deploys pods with new image              │
│  - Waits for health checks                  │
│  - Updates deployment status                │
└──────┬──────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│         Kubernetes Cluster                   │
│  - Runs service pods (3 replicas)           │
│  - Manages PostgreSQL database              │
│  - Manages Redis cache                      │
│  - Manages Kafka cluster                    │
│  - Injects secrets as env vars              │
│  - Monitors health checks                   │
│  - Collects metrics and logs                │
└──────┬──────────────────────────────────────┘
       │
       ▼
┌─────────────────────────────────────────────┐
│      Observability Stack                     │
│  - Prometheus scrapes metrics               │
│  - Grafana displays dashboards              │
│  - Loki aggregates logs                     │
│  - Alertmanager sends alerts                │
└─────────────────────────────────────────────┘
```

---

## Error Handling and Rollback

### If Something Goes Wrong

#### During Scaffolding
```
- Backstage shows error message
- Developer can retry
- No infrastructure created yet
- No cleanup needed
```

#### During Infrastructure Provisioning
```
- Crossplane retries automatically
- If database fails: retries up to 3 times
- If Redis fails: retries up to 3 times
- If Kafka fails: retries up to 3 times
- Developer notified via Backstage
- Can view Crossplane logs for debugging
```

#### During Deployment
```
- ArgoCD shows sync status
- If health checks fail: deployment paused
- Previous version keeps running
- Developer can:
  - Check logs in Backstage
  - View metrics in Grafana
  - Rollback to previous version (1 click)
```

### Rollback Process
```
1. Developer clicks "Rollback" in Backstage or ArgoCD
2. ArgoCD reverts to previous Git commit
3. Kubernetes performs rolling update
4. Old version restored in ~30 seconds
5. Zero downtime
```

---

## Summary

This IDP transforms service creation from a **weeks-long manual process** into a **10-minute automated workflow**. Developers focus on business logic while the platform handles:

- Infrastructure provisioning
- Database and cache setup
- CI/CD pipeline configuration
- Monitoring and alerting
- Security and compliance
- Documentation

**Result**: Faster time to market, happier developers, more reliable services. 🎉
