# Crossplane Guide — Infrastructure as Kubernetes Resources

This guide explains Crossplane concepts and how the IDP uses it to provision infrastructure (PostgreSQL, Redis, Kafka) through Kubernetes-native APIs.

---

## 1. What is Crossplane?

Crossplane extends Kubernetes so you can manage **any infrastructure** (databases, caches, cloud resources) using the same `kubectl` and YAML you already know.

### Without Crossplane (Traditional)
```
Developer → Jira ticket → Platform team → AWS Console → 3 days later → Database ready
```

### With Crossplane (IDP)
```
Developer → Backstage form → Crossplane Claim → 5 minutes → Database ready
```

### Key Concept

Crossplane lets you define **custom Kubernetes resources** (like `GoService`) that automatically create real infrastructure when applied.

---

## 2. Core Concepts

### 2.1 Providers

Providers connect Crossplane to external systems:

```
┌──────────────────────────────────────────────┐
│              Crossplane                        │
│                                                │
│  ┌──────────────┐  ┌──────────────────────┐  │
│  │   Provider    │  │     Provider         │  │
│  │  Kubernetes   │  │       Helm           │  │
│  └──────┬───────┘  └──────────┬───────────┘  │
│         │                     │               │
└─────────┼─────────────────────┼───────────────┘
          │                     │
          ▼                     ▼
   K8s Resources          Helm Releases
   (Deployments,          (Operators,
    Services,              Charts)
    ConfigMaps)
```

In this IDP, we use:
- **provider-kubernetes** — Creates K8s resources (Deployments, Services, Secrets)
- **provider-helm** — Installs Helm charts (operators, complex apps)

### 2.2 Composite Resource Definitions (XRDs)

An XRD is like a **custom API** you define. It specifies what fields developers can fill in.

Think of it as a **form definition**:

```yaml
# This defines WHAT developers can request
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: goservices.platform.example.com
spec:
  group: platform.example.com
  names:
    kind: GoService           # The custom resource name
    plural: goservices
  claimNames:
    kind: GoServiceClaim      # What developers actually use
    plural: goserviceclaims
  versions:
    - name: v1alpha1
      served: true
      reachable: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                serviceName:
                  type: string
                  description: "Name of the Go service"
                namespace:
                  type: string
                  default: services
                replicas:
                  type: integer
                  default: 2
                image:
                  type: string
                database:
                  type: object
                  properties:
                    enabled:
                      type: boolean
                      default: true
                    size:
                      type: string
                      enum: [small, medium, large]
                      default: small
                redis:
                  type: object
                  properties:
                    enabled:
                      type: boolean
                      default: false
                    size:
                      type: string
                      enum: [small, medium, large]
                      default: small
                kafka:
                  type: object
                  properties:
                    enabled:
                      type: boolean
                      default: false
                    topics:
                      type: array
                      items:
                        type: object
                        properties:
                          name:
                            type: string
                          partitions:
                            type: integer
                            default: 3
              required:
                - serviceName
                - image
```

### 2.3 Compositions

A Composition defines **HOW** to fulfill a request. It maps XRD fields to actual infrastructure resources.

Think of it as the **implementation** behind the form:

```
XRD (What)          →  Composition (How)
─────────────────      ─────────────────────────
serviceName: user   →  Deployment "user-service"
database.enabled    →  CloudNativePG Cluster
redis.enabled       →  Redis Instance
kafka.enabled       →  KafkaTopic CRDs
```

### 2.4 Claims

A Claim is what developers actually submit. It's an instance of the XRD:

```yaml
# This is what a developer (or Backstage) submits
apiVersion: platform.example.com/v1alpha1
kind: GoServiceClaim
metadata:
  name: user-service
  namespace: services
spec:
  serviceName: user-service
  image: ghcr.io/myorg/user-service:latest
  replicas: 2
  database:
    enabled: true
    size: small
  redis:
    enabled: true
    size: small
  kafka:
    enabled: true
    topics:
      - name: user-events
        partitions: 3
```

### 2.5 How It All Connects

```
Developer                    Crossplane                     Kubernetes
─────────                    ──────────                     ──────────

GoServiceClaim ──────────►  XRD validates  ──────────────►  Composition creates:
(user-service)               the request                    ├── Deployment
                                                            ├── Service
                                                            ├── CloudNativePG Cluster
                                                            ├── Redis Instance
                                                            ├── KafkaTopic
                                                            └── Secrets (connection strings)
```

---

## 3. Building the GoService Composition

### 3.1 Composition Structure

The GoService composition orchestrates multiple sub-resources:

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: goservice-composition
  labels:
    crossplane.io/xrd: goservices.platform.example.com
spec:
  compositeTypeRef:
    apiVersion: platform.example.com/v1alpha1
    kind: GoService
  resources:
    # 1. Kubernetes Deployment
    - name: deployment
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha2
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: apps/v1
              kind: Deployment
              spec:
                replicas: 2
                template:
                  spec:
                    containers:
                      - name: app
                        ports:
                          - containerPort: 8080
                            name: http
                          - containerPort: 9090
                            name: grpc
                        livenessProbe:
                          httpGet:
                            path: /health
                            port: 8080
                          initialDelaySeconds: 10
                        readinessProbe:
                          httpGet:
                            path: /ready
                            port: 8080
                          initialDelaySeconds: 5
                        envFrom:
                          - secretRef:
                              name: "" # patched
      patches:
        - fromFieldPath: spec.serviceName
          toFieldPath: spec.forProvider.manifest.metadata.name
        - fromFieldPath: spec.namespace
          toFieldPath: spec.forProvider.manifest.metadata.namespace
        - fromFieldPath: spec.replicas
          toFieldPath: spec.forProvider.manifest.spec.replicas
        - fromFieldPath: spec.image
          toFieldPath: spec.forProvider.manifest.spec.template.spec.containers[0].image

    # 2. Kubernetes Service
    - name: service
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha2
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: v1
              kind: Service
              spec:
                type: ClusterIP
                ports:
                  - name: http
                    port: 8080
                    targetPort: 8080
                  - name: grpc
                    port: 9090
                    targetPort: 9090
      patches:
        - fromFieldPath: spec.serviceName
          toFieldPath: spec.forProvider.manifest.metadata.name
        - fromFieldPath: spec.namespace
          toFieldPath: spec.forProvider.manifest.metadata.namespace
```

### 3.2 Database Composition (Conditional)

Only created when `database.enabled: true`:

```yaml
    # 3. PostgreSQL (CloudNativePG)
    - name: postgresql
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha2
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: postgresql.cnpg.io/v1
              kind: Cluster
              spec:
                instances: 2
                storage:
                  size: 20Gi
                postgresql:
                  parameters:
                    max_connections: "100"
      patches:
        - fromFieldPath: spec.serviceName
          toFieldPath: spec.forProvider.manifest.metadata.name
          transforms:
            - type: string
              string:
                fmt: "%s-db"
        - fromFieldPath: spec.database.size
          toFieldPath: spec.forProvider.manifest.spec.storage.size
          transforms:
            - type: map
              map:
                small: "20Gi"
                medium: "100Gi"
                large: "500Gi"
```

### 3.3 Redis Composition (Conditional)

```yaml
    # 4. Redis
    - name: redis
      base:
        apiVersion: kubernetes.crossplane.io/v1alpha2
        kind: Object
        spec:
          forProvider:
            manifest:
              apiVersion: redis.redis.opstreelabs.in/v1beta2
              kind: Redis
              spec:
                kubernetesConfig:
                  image: redis:7-alpine
                  resources:
                    limits:
                      memory: 1Gi
      patches:
        - fromFieldPath: spec.serviceName
          toFieldPath: spec.forProvider.manifest.metadata.name
          transforms:
            - type: string
              string:
                fmt: "%s-redis"
        - fromFieldPath: spec.redis.size
          toFieldPath: spec.forProvider.manifest.spec.kubernetesConfig.resources.limits.memory
          transforms:
            - type: map
              map:
                small: "1Gi"
                medium: "4Gi"
                large: "16Gi"
```

---

## 4. Size Tiers Reference

| Component | Small | Medium | Large |
|-----------|-------|--------|-------|
| **PostgreSQL** | 20GB / 2CPU / 4GB RAM | 100GB / 4CPU / 8GB RAM | 500GB / 8CPU / 16GB RAM |
| **Redis** | 1GB RAM | 4GB RAM | 16GB RAM |
| **Kafka** | 3 brokers / 10GB | 3 brokers / 100GB | 5 brokers / 500GB |

---

## 5. Working with Claims

### 5.1 Create a Claim

```bash
kubectl apply -f crossplane/claims/user-service.yaml
```

### 5.2 Check Status

```bash
# See the claim
kubectl get goserviceclaim -n services

# See the composite resource
kubectl get goservice

# See all managed resources
kubectl get managed

# Detailed status
kubectl describe goserviceclaim user-service -n services
```

### 5.3 Check What Was Created

```bash
# Database
kubectl get clusters.postgresql.cnpg.io -A

# Redis
kubectl get redis -A

# Kafka topics
kubectl get kafkatopic -A

# Deployment
kubectl get deployment -n services

# Secrets (connection strings)
kubectl get secrets -n services
```

### 5.4 Delete a Claim

```bash
kubectl delete goserviceclaim user-service -n services
# Crossplane automatically deletes ALL resources it created
```

---

## 6. Debugging Crossplane

### 6.1 Common Issues

**Claim stuck in "Waiting"**:
```bash
# Check XRD is installed
kubectl get xrd

# Check composition exists
kubectl get composition

# Check events
kubectl describe goserviceclaim <name> -n services
```

**Resource not created**:
```bash
# Check provider logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/revision --tail=50

# Check managed resource status
kubectl describe <resource-type> <name>
```

**Provider not healthy**:
```bash
# Check provider status
kubectl get providers

# Check provider pod
kubectl get pods -n crossplane-system
kubectl logs -n crossplane-system <provider-pod>
```

### 6.2 Useful Commands

```bash
# List all Crossplane resources
kubectl get crossplane

# Watch composition progress
kubectl get managed -w

# Check Crossplane system health
kubectl get pods -n crossplane-system
kubectl get providers
kubectl get xrd
kubectl get composition
```

---

## 7. Crossplane vs. Alternatives

| Feature | Crossplane | Terraform | Pulumi |
|---------|-----------|-----------|--------|
| **Runs in** | Kubernetes | CLI / CI | CLI / CI |
| **State** | Kubernetes etcd | S3/remote | Pulumi Cloud |
| **Drift detection** | Continuous (reconciles) | On `plan` only | On `preview` only |
| **Self-healing** | Yes (controller loop) | No | No |
| **Developer interface** | `kubectl apply` claim | Run `terraform apply` | Run `pulumi up` |
| **IDP integration** | Native K8s API | Requires wrapper | Requires wrapper |

Crossplane is ideal for IDP because:
- Developers don't need new tools — just `kubectl` (or Backstage UI)
- Infrastructure self-heals if someone deletes a database accidentally
- Claims provide a simple API; complexity is hidden in compositions

---

## 8. File Organization

```
crossplane/
├── providers/           # Provider installations + configs
│   ├── provider-kubernetes.yaml
│   ├── provider-helm.yaml
│   └── provider-configs.yaml
├── configurations/      # Platform-level settings
│   └── platform-config.yaml
├── compositions/        # How to build infrastructure
│   ├── goservice-xrd.yaml          # XRD definition
│   ├── goservice-composition.yaml  # Main composition
│   ├── postgresql-composition.yaml # DB sub-composition
│   ├── redis-composition.yaml      # Cache sub-composition
│   └── kafka-composition.yaml      # Messaging sub-composition
└── claims/              # Example claims for testing
    ├── user-service.yaml
    └── order-service.yaml
```

---

## Next Steps

1. Install Crossplane and providers → [Setup Guide](./setup-guide.md) Section 3
2. Define the GoService XRD → `crossplane/compositions/goservice-xrd.yaml`
3. Build compositions → `crossplane/compositions/`
4. Test with example claims → `crossplane/claims/`
5. Integrate with Backstage → [Backstage Guide](./backstage-guide.md)

Related: [Architecture Plan](./idp-architecture-plan.md) | [Developer Workflow](./developer-workflow.md)
