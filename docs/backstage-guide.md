# Backstage Guide — Developer Portal for the IDP

This guide explains Backstage concepts and how it serves as the self-service developer portal for creating and managing services.

---

## 1. What is Backstage?

Backstage is an **open-source developer portal** built by Spotify. It gives developers a single place to:

- **Create new services** via software templates (forms)
- **Browse all services** in a catalog
- **View documentation** (TechDocs)
- **Monitor services** (K8s status, ArgoCD sync, Grafana dashboards)

### Without Backstage
```
Developer → Slack: "Can someone create me a new service?"
           → Wait 2 days
           → Copy-paste from existing service
           → Manually set up CI/CD, database, monitoring
           → 1-2 weeks to first deploy
```

### With Backstage
```
Developer → Open portal → Fill form → Click Create → 10 minutes → Running service
```

---

## 2. Architecture

```
┌────────────────────────────────────────────────────────────┐
│                      Backstage                              │
│                                                             │
│  ┌─────────────┐  ┌─────────────┐  ┌───────────────────┐  │
│  │  Frontend    │  │  Backend    │  │  Plugins           │  │
│  │  (React)     │  │  (Node.js)  │  │  ┌───────────────┐│  │
│  │             │  │             │  │  │ Catalog       ││  │
│  │  - Catalog   │  │  - API      │  │  │ Scaffolder    ││  │
│  │  - Templates │  │  - Auth     │  │  │ TechDocs      ││  │
│  │  - TechDocs  │  │  - Catalog  │  │  │ Kubernetes    ││  │
│  │  - Dashboard │  │  - Scaffold │  │  │ ArgoCD        ││  │
│  └─────────────┘  └─────────────┘  │  └───────────────┘│  │
│                                     └───────────────────┘  │
└────────────────────────────────────────────────────────────┘
         │                │                    │
         ▼                ▼                    ▼
    Developer          PostgreSQL          GitHub, K8s,
    Browser            (catalog DB)        ArgoCD, Grafana
```

---

## 3. Core Concepts

### 3.1 Software Catalog

The catalog is a **registry of everything** in your organization:

| Entity Type | Example | Description |
|-------------|---------|-------------|
| **Component** | `user-service` | A microservice, library, or website |
| **API** | `user-api` | A gRPC or REST API |
| **System** | `backend-platform` | A group of related components |
| **Domain** | `commerce` | A business domain |
| **Group** | `platform-team` | A team |
| **User** | `john.doe` | A person |

Each entity is defined in a `catalog-info.yaml` file in its Git repo:

```yaml
# catalog-info.yaml (in service repo root)
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: user-service
  description: User management microservice
  annotations:
    github.com/project-slug: myorg/user-service
    backstage.io/techdocs-ref: dir:.
    argocd/app-name: user-service
    grafana/dashboard-selector: service=user-service
  tags:
    - go
    - grpc
    - postgresql
spec:
  type: service
  lifecycle: production
  owner: platform-team
  system: backend-platform
  providesApis:
    - user-api
  dependsOn:
    - resource:user-service-db
    - resource:user-service-redis
```

### 3.2 Software Templates

Templates are **forms that create new services**. When a developer fills out the form, Backstage:

1. Copies a skeleton (like `go-service-template`)
2. Replaces variables with form values
3. Creates a GitHub repository
4. Applies Crossplane claims
5. Registers the service in the catalog

```yaml
# backstage/templates/go-service-template/template.yaml
apiVersion: scaffolder.backstage.io/v1beta3
kind: Template
metadata:
  name: go-service
  title: Go Microservice
  description: Create a production-ready Go service with optional PostgreSQL, Redis, and Kafka
  tags:
    - go
    - grpc
    - recommended
spec:
  owner: platform-team
  type: service

  # Form fields
  parameters:
    - title: Service Information
      required:
        - serviceName
        - description
        - owner
      properties:
        serviceName:
          title: Service Name
          type: string
          description: Unique name for this service (lowercase, hyphens)
          pattern: "^[a-z][a-z0-9-]*$"
          ui:autofocus: true
        description:
          title: Description
          type: string
          description: What does this service do?
        owner:
          title: Owner
          type: string
          description: Team that owns this service
          ui:field: OwnerPicker
          ui:options:
            catalogFilter:
              kind: Group

    - title: Infrastructure
      properties:
        database:
          title: PostgreSQL Database
          type: boolean
          default: true
        databaseSize:
          title: Database Size
          type: string
          enum: [small, medium, large]
          default: small
          description: "small: 2GB, medium: 4GB, large: 8GB"
        redis:
          title: Redis Cache
          type: boolean
          default: false
        redisSize:
          title: Redis Size
          type: string
          enum: [small, medium, large]
          default: small
        kafka:
          title: Kafka Messaging
          type: boolean
          default: false
        kafkaTopics:
          title: Kafka Topics (comma-separated)
          type: string
          description: "e.g., user-events,user-notifications"

    - title: Deployment
      properties:
        namespace:
          title: Namespace
          type: string
          default: services
        replicas:
          title: Replicas
          type: integer
          default: 2
          enum: [1, 2, 3, 5]

  # What happens when they click Create
  steps:
    # 1. Copy skeleton and replace variables
    - id: fetch-skeleton
      name: Fetch Skeleton
      action: fetch:template
      input:
        url: ./skeleton
        values:
          serviceName: ${{ parameters.serviceName }}
          description: ${{ parameters.description }}
          owner: ${{ parameters.owner }}
          database: ${{ parameters.database }}
          databaseSize: ${{ parameters.databaseSize }}
          redis: ${{ parameters.redis }}
          redisSize: ${{ parameters.redisSize }}
          kafka: ${{ parameters.kafka }}
          kafkaTopics: ${{ parameters.kafkaTopics }}
          namespace: ${{ parameters.namespace }}
          replicas: ${{ parameters.replicas }}

    # 2. Create GitHub repository
    - id: publish
      name: Create Repository
      action: publish:github
      input:
        allowedHosts: ["github.com"]
        repoUrl: github.com?owner=myorg&repo=${{ parameters.serviceName }}
        description: ${{ parameters.description }}
        defaultBranch: main
        repoVisibility: public

    # 3. Create Crossplane claim for infrastructure
    - id: create-infrastructure
      name: Provision Infrastructure
      action: crossplane:claim:create
      input:
        claimName: ${{ parameters.serviceName }}
        namespace: ${{ parameters.namespace }}
        spec:
          serviceName: ${{ parameters.serviceName }}
          image: ghcr.io/myorg/${{ parameters.serviceName }}:latest
          replicas: ${{ parameters.replicas }}
          database:
            enabled: ${{ parameters.database }}
            size: ${{ parameters.databaseSize }}
          redis:
            enabled: ${{ parameters.redis }}
            size: ${{ parameters.redisSize }}
          kafka:
            enabled: ${{ parameters.kafka }}

    # 4. Register in catalog
    - id: register
      name: Register in Catalog
      action: catalog:register
      input:
        repoContentsUrl: ${{ steps['publish'].output.repoContentsUrl }}
        catalogInfoPath: /catalog-info.yaml

  # Show links after creation
  output:
    links:
      - title: Repository
        url: ${{ steps['publish'].output.remoteUrl }}
      - title: Open in Catalog
        icon: catalog
        entityRef: ${{ steps['register'].output.entityRef }}
```

### 3.3 Skeleton Directory

The skeleton is the **template code** that gets copied for each new service:

```
backstage/templates/go-service-template/
├── template.yaml          # Template definition (form + steps)
└── skeleton/              # Code that gets copied
    ├── catalog-info.yaml  # Backstage catalog entry
    ├── go.mod
    ├── Makefile
    ├── Dockerfile
    ├── cmd/
    │   └── main/
    │       └── main.go
    ├── internal/
    │   ├── config/
    │   ├── handler/
    │   ├── service/
    │   └── registry/
    ├── k8s/
    │   ├── deploy.yaml
    │   └── service.yaml
    ├── crossplane-claim.yaml
    └── .github/
        └── workflows/
            ├── ci.yaml
            └── cd.yaml
```

Variables in skeleton files use `${{ values.serviceName }}` syntax:

```yaml
# skeleton/catalog-info.yaml
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  name: ${{ values.serviceName }}
  description: ${{ values.description }}
spec:
  type: service
  lifecycle: production
  owner: ${{ values.owner }}
```

---

## 4. Plugins

### 4.1 Kubernetes Plugin

Shows pod status, resource usage, and logs directly in Backstage:

```
Service Page → Kubernetes Tab
├── Pods: 2/2 Running
├── CPU: 120m / 500m
├── Memory: 256Mi / 512Mi
└── Events: (recent pod events)
```

### 4.2 ArgoCD Plugin

Shows deployment and sync status:

```
Service Page → ArgoCD Tab
├── Sync Status: Synced
├── Health: Healthy
├── Last Sync: 5 minutes ago
└── History: (deployment history)
```

### 4.3 TechDocs Plugin

Auto-generates documentation from markdown in each repo:

```
Service Page → Docs Tab
└── (rendered from docs/ in the service repo)
```

Requires `mkdocs.yaml` in each service repo:
```yaml
site_name: user-service
nav:
  - Home: index.md
  - API: api.md
plugins:
  - techdocs-core
```

### 4.4 Grafana Plugin

Links to service dashboards:

```
Service Page → Monitoring Tab
└── Link to Grafana dashboard filtered by service name
```

---

## 5. Custom Scaffolder Actions

The IDP needs custom actions that don't come with Backstage out of the box.

### 5.1 Crossplane Claim Action

Creates a Crossplane GoService claim when a new service is scaffolded:

```typescript
// backstage/packages/backend/src/plugins/scaffolder-crossplane.ts
import { createTemplateAction } from '@backstage/plugin-scaffolder-node';
import { KubeConfig, CustomObjectsApi } from '@kubernetes/client-node';

export function createCrossplaneClaimAction() {
  return createTemplateAction({
    id: 'crossplane:claim:create',
    description: 'Create a Crossplane GoService claim',
    schema: {
      input: {
        type: 'object',
        required: ['claimName', 'namespace', 'spec'],
        properties: {
          claimName: { type: 'string' },
          namespace: { type: 'string' },
          spec: { type: 'object' },
        },
      },
    },
    async handler(ctx) {
      const kc = new KubeConfig();
      kc.loadFromCluster();
      const api = kc.makeApiClient(CustomObjectsApi);

      await api.createNamespacedCustomObject(
        'platform.example.com',
        'v1alpha1',
        ctx.input.namespace,
        'goserviceclaims',
        {
          apiVersion: 'platform.example.com/v1alpha1',
          kind: 'GoServiceClaim',
          metadata: {
            name: ctx.input.claimName,
            namespace: ctx.input.namespace,
          },
          spec: ctx.input.spec,
        },
      );

      ctx.logger.info(`Created GoServiceClaim: ${ctx.input.claimName}`);
    },
  });
}
```

### 5.2 ArgoCD Application Action

Registers a new ArgoCD Application:

```typescript
// backstage/packages/backend/src/plugins/scaffolder-argocd.ts
import { createTemplateAction } from '@backstage/plugin-scaffolder-node';

export function createArgoCDApplicationAction() {
  return createTemplateAction({
    id: 'argocd:application:create',
    description: 'Create an ArgoCD Application',
    schema: {
      input: {
        type: 'object',
        required: ['appName', 'repoUrl', 'path'],
        properties: {
          appName: { type: 'string' },
          repoUrl: { type: 'string' },
          path: { type: 'string', default: 'k8s' },
          namespace: { type: 'string', default: 'services' },
        },
      },
    },
    async handler(ctx) {
      // Create ArgoCD Application via ArgoCD API
      // Or apply Application YAML via Kubernetes API
      ctx.logger.info(`Created ArgoCD Application: ${ctx.input.appName}`);
    },
  });
}
```

---

## 6. Catalog Configuration

### 6.1 Root Catalog

```yaml
# backstage/catalog/catalog-info.yaml
apiVersion: backstage.io/v1alpha1
kind: Location
metadata:
  name: platform-catalog
  description: Platform service catalog
spec:
  targets:
    - ./systems/backend-platform.yaml
```

### 6.2 System Entity

```yaml
# backstage/catalog/systems/backend-platform.yaml
apiVersion: backstage.io/v1alpha1
kind: System
metadata:
  name: backend-platform
  description: Backend microservices platform
spec:
  owner: platform-team
  domain: engineering
```

### 6.3 Auto-Discovery

Configure Backstage to automatically discover services from GitHub:

```yaml
# app-config.yaml
catalog:
  providers:
    github:
      myOrg:
        organization: myorg
        catalogPath: /catalog-info.yaml
        filters:
          branch: main
          repository: '.*'
        schedule:
          frequency: { minutes: 30 }
          timeout: { minutes: 3 }
```

---

## 7. app-config.yaml

The main Backstage configuration file:

```yaml
app:
  title: IDP Developer Portal
  baseUrl: http://localhost:3000

backend:
  baseUrl: http://localhost:7007
  database:
    client: pg
    connection:
      host: localhost
      port: 5432
      user: backstage
      password: backstage

auth:
  providers:
    github:
      development:
        clientId: ${GITHUB_CLIENT_ID}
        clientSecret: ${GITHUB_CLIENT_SECRET}

integrations:
  github:
    - host: github.com
      token: ${GITHUB_TOKEN}

catalog:
  rules:
    - allow: [Component, System, API, Resource, Location, Template, Group, User]
  locations:
    - type: file
      target: ../../backstage/catalog/catalog-info.yaml
    - type: file
      target: ../../backstage/templates/go-service-template/template.yaml

kubernetes:
  serviceLocatorMethod:
    type: multiTenant
  clusterLocatorMethods:
    - type: config
      clusters:
        - url: https://kubernetes.default.svc
          name: local
          authProvider: serviceAccount
          serviceAccountToken: ${K8S_TOKEN}

argocd:
  baseUrl: https://argocd.platform.svc
  username: admin
  password: ${ARGOCD_PASSWORD}
```

---

## 8. Deployment to Kubernetes

```yaml
# backstage/k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backstage
  namespace: platform
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backstage
  template:
    metadata:
      labels:
        app: backstage
    spec:
      containers:
        - name: backstage
          image: backstage:latest
          ports:
            - containerPort: 7007
          env:
            - name: GITHUB_TOKEN
              valueFrom:
                secretKeyRef:
                  name: backstage-secrets
                  key: github-token
```

---

## 9. File Organization

```
backstage/
├── catalog/                          # Service catalog definitions
│   ├── catalog-info.yaml            # Root catalog location
│   └── systems/
│       └── backend-platform.yaml    # System entity
├── templates/                        # Software templates
│   └── go-service-template/
│       ├── template.yaml            # Form + scaffolding steps
│       └── skeleton/                # Template code
│           ├── catalog-info.yaml
│           ├── go.mod
│           ├── Makefile
│           ├── ...
│           ├── k8s/
│           └── .github/workflows/
└── packages/                         # Custom plugins
    └── backend/
        └── src/
            └── plugins/
                ├── scaffolder-crossplane.ts
                └── scaffolder-argocd.ts
```

---

## 10. Developer Experience Flow

```
Step 1: Developer opens portal
        → https://backstage.your-domain.com

Step 2: Click "Create" → Select "Go Microservice"

Step 3: Fill form:
        Service Name: user-service
        Description: User management service
        Owner: backend-team
        PostgreSQL: Yes (small)
        Redis: Yes (small)
        Kafka: No

Step 4: Click "Create"

Step 5: Backstage executes:
        ├── Copy go-service-template skeleton
        ├── Replace variables (service name, etc.)
        ├── Create GitHub repo: myorg/user-service
        ├── Push code
        ├── Create Crossplane GoServiceClaim
        └── Register in catalog

Step 6: Developer sees:
        ├── Link to GitHub repo
        ├── Link to service in catalog
        └── Status: Infrastructure provisioning...

Step 7: (2-5 minutes later)
        ├── PostgreSQL: Ready
        ├── Redis: Ready
        ├── CI: Building...
        └── ArgoCD: Syncing...

Step 8: (5-10 minutes later)
        ├── Service: Running (2/2 pods)
        ├── Health: Healthy
        └── Grafana: Dashboard available
```

---

## Next Steps

1. Initialize Backstage app → `npx @backstage/create-app@latest`
2. Configure `app-config.yaml`
3. Add software templates → `backstage/templates/`
4. Build custom scaffolder actions → `backstage/packages/backend/src/plugins/`
5. Deploy to Kubernetes

Related: [Setup Guide](./setup-guide.md) | [Crossplane Guide](./crossplane-guide.md) | [ArgoCD Guide](./argocd-gitops-guide.md) | [Developer Workflow](./developer-workflow.md)
