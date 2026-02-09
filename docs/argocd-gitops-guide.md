# ArgoCD & GitOps Guide — Continuous Deployment for the IDP

This guide explains GitOps concepts and how ArgoCD automates deployment for all services in the platform.

---

## 1. What is GitOps?

GitOps means **Git is the single source of truth** for your infrastructure and applications. Instead of running `kubectl apply` manually, you push changes to Git and a controller automatically syncs them to the cluster.

### Traditional Deployment
```
Developer → Build → Push image → SSH to server → kubectl apply → Hope it works
```

### GitOps Deployment
```
Developer → Push code → CI builds image → Update manifest in Git → ArgoCD syncs → Done
```

### Core Principles

1. **Declarative** — Everything described in Git (YAML manifests)
2. **Versioned** — Git history = deployment history
3. **Automated** — Push to Git = deploy to cluster
4. **Self-healing** — If someone manually changes the cluster, ArgoCD reverts it

---

## 2. ArgoCD Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    ArgoCD (in cluster)                    │
│                                                          │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐ │
│  │  API Server   │    │  Repo Server │    │Application│ │
│  │  (UI + CLI)   │    │  (Git sync)  │    │Controller │ │
│  └──────┬───────┘    └──────┬───────┘    └─────┬─────┘ │
│         │                   │                   │       │
└─────────┼───────────────────┼───────────────────┼───────┘
          │                   │                   │
          ▼                   ▼                   ▼
     Dashboard           Git Repos          K8s Resources
     (port 8080)         (GitHub)           (Deployments,
                                             Services, etc.)
```

**Components:**
- **API Server** — UI dashboard and CLI interface
- **Repo Server** — Clones Git repos, renders manifests
- **Application Controller** — Compares desired state (Git) vs actual state (cluster), syncs differences

---

## 3. Key Concepts

### 3.1 Application

An ArgoCD Application defines:
- **Source**: Where to get manifests (Git repo + path)
- **Destination**: Where to deploy (cluster + namespace)
- **Sync Policy**: How to keep them in sync

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: user-service
  namespace: platform      # ArgoCD lives here
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/user-service.git
    targetRevision: main
    path: k8s              # Directory containing K8s manifests
  destination:
    server: https://kubernetes.default.svc
    namespace: services    # Deploy to this namespace
  syncPolicy:
    automated:
      prune: true          # Delete resources removed from Git
      selfHeal: true       # Revert manual changes
    syncOptions:
      - CreateNamespace=true
```

### 3.2 App-of-Apps Pattern

Instead of creating each Application manually, you create **one Application that manages other Applications**:

```
platform-apps (root Application)
├── crossplane-app
├── monitoring-app
├── vault-app
├── backstage-app
└── services-appset (ApplicationSet)
    ├── user-service
    ├── order-service
    └── ... (auto-discovered)
```

```yaml
# argocd/platform/app-of-apps.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: platform-apps
  namespace: platform
spec:
  project: default
  source:
    repoURL: https://github.com/myorg/platform-idp.git
    targetRevision: main
    path: argocd/platform   # Contains all platform Application YAMLs
  destination:
    server: https://kubernetes.default.svc
    namespace: platform
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 3.3 ApplicationSet

An ApplicationSet **auto-discovers** services and creates Applications dynamically. No manual Application creation needed per service.

```yaml
# argocd/services/service-appset.yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: services
  namespace: platform
spec:
  generators:
    # Auto-discover repos with catalog-info.yaml
    - git:
        repoURL: https://github.com/myorg/platform-idp.git
        revision: main
        directories:
          - path: "services/*"
  template:
    metadata:
      name: "{{path.basename}}"
    spec:
      project: default
      source:
        repoURL: https://github.com/myorg/{{path.basename}}.git
        targetRevision: main
        path: k8s
      destination:
        server: https://kubernetes.default.svc
        namespace: services
      syncPolicy:
        automated:
          prune: true
          selfHeal: true
```

---

## 4. Sync Policies

### 4.1 Manual Sync (Default)

Changes in Git are detected but NOT applied automatically. You must click "Sync" in the UI or run:
```bash
argocd app sync <app-name>
```

### 4.2 Automated Sync

Changes are applied automatically when detected:

```yaml
syncPolicy:
  automated:
    prune: true       # Remove resources deleted from Git
    selfHeal: true    # Revert manual cluster changes
```

### 4.3 Sync Options

```yaml
syncPolicy:
  syncOptions:
    - CreateNamespace=true        # Create namespace if missing
    - PruneLast=true              # Delete old resources after new ones are healthy
    - ApplyOutOfSyncOnly=true     # Only apply changed resources
    - ServerSideApply=true        # Use server-side apply for large resources
```

### 4.4 Sync Waves

Control the order resources are applied:

```yaml
# Deploy namespace first (wave 0)
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "0"

# Then ConfigMaps (wave 1)
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "1"

# Then Deployments (wave 2)
metadata:
  annotations:
    argocd.argoproj.io/sync-wave: "2"
```

---

## 5. Deployment Strategies

### 5.1 Rolling Update (Default)

```yaml
# In Deployment spec
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1           # Create 1 extra pod during update
      maxUnavailable: 0     # Never have fewer pods than desired
```

### 5.2 Rollback

**Via UI**: Click "History" → select previous version → "Rollback"

**Via CLI**:
```bash
# See deployment history
argocd app history <app-name>

# Rollback to specific revision
argocd app rollback <app-name> <revision>

# Or revert the Git commit and let ArgoCD sync
git revert <commit-sha>
git push
```

**Via Git** (recommended — keeps Git as source of truth):
```bash
git revert <bad-commit>
git push origin main
# ArgoCD auto-syncs to the reverted state
```

---

## 6. IDP ArgoCD Structure

```
argocd/
├── platform/                          # Platform infrastructure apps
│   ├── app-of-apps.yaml              # Root application
│   ├── crossplane.yaml               # Crossplane Application
│   ├── monitoring.yaml               # Monitoring stack Application
│   ├── vault.yaml                    # Vault Application
│   ├── backstage.yaml                # Backstage Application
│   └── operators.yaml                # Database/Redis/Kafka operators
│
└── services/                          # Service deployment automation
    └── service-appset.yaml           # ApplicationSet for auto-discovery
```

### How It Works in the IDP

```
1. Developer creates service via Backstage
2. Backstage scaffolds code → pushes to GitHub
3. CI pipeline builds → pushes Docker image
4. CI updates k8s/deploy.yaml with new image tag
5. ArgoCD detects Git change
6. ArgoCD applies updated manifests to cluster
7. Kubernetes rolls out new version
8. ArgoCD reports sync status back to Backstage
```

---

## 7. Common Operations

### 7.1 Check Application Status

```bash
# List all apps
argocd app list

# Get details
argocd app get <app-name>

# Watch sync status
argocd app get <app-name> --refresh
```

### 7.2 Force Sync

```bash
# Sync specific app
argocd app sync <app-name>

# Force sync (ignore diff)
argocd app sync <app-name> --force

# Sync with prune
argocd app sync <app-name> --prune
```

### 7.3 Diff (Preview Changes)

```bash
# See what would change
argocd app diff <app-name>

# Live diff against a specific revision
argocd app diff <app-name> --revision <commit-sha>
```

### 7.4 Application Health States

| Status | Meaning |
|--------|---------|
| **Healthy** | All resources running correctly |
| **Progressing** | Deployment rolling out |
| **Degraded** | Some resources failed (e.g., pod crash) |
| **Suspended** | Paused (e.g., scaled to 0) |
| **Missing** | Resources exist in Git but not in cluster |
| **OutOfSync** | Cluster state differs from Git |

---

## 8. Environment Promotion

For multi-environment setups (dev → staging → prod):

### 8.1 Branch-Based

```
main branch      → Production
staging branch   → Staging
develop branch   → Development
```

### 8.2 Directory-Based (Recommended)

```
k8s/
├── base/                # Shared manifests
│   ├── deployment.yaml
│   └── service.yaml
├── overlays/
│   ├── dev/             # Dev overrides
│   │   └── kustomization.yaml
│   ├── staging/         # Staging overrides
│   │   └── kustomization.yaml
│   └── prod/            # Production overrides
│       └── kustomization.yaml
```

```yaml
# ArgoCD Application for each environment
# dev-app.yaml
spec:
  source:
    path: k8s/overlays/dev

# prod-app.yaml
spec:
  source:
    path: k8s/overlays/prod
```

---

## 9. Troubleshooting

### App shows "OutOfSync" but nothing changed
```bash
# Check for diff
argocd app diff <app-name>

# Common causes:
# - Kubernetes added default fields (resource requests, labels)
# - Helm rendered slightly different output
# Fix: add ignoreDifferences to Application spec
```

### App shows "Unknown" health
```bash
# Check events
kubectl describe application <app-name> -n platform

# Check ArgoCD logs
kubectl logs -n platform -l app.kubernetes.io/name=argocd-application-controller --tail=50
```

### Sync takes too long
```bash
# Check if resources are stuck
argocd app get <app-name> --show-operation

# Check pod events
kubectl get events -n services --sort-by=.metadata.creationTimestamp
```

---

## 10. Best Practices

1. **Git is the source of truth** — Never `kubectl apply` manually in production
2. **Use automated sync + self-heal** — Prevents drift
3. **Use sync waves** — Deploy dependencies before dependents
4. **Use ApplicationSets** — Auto-discover services, don't manage Applications manually
5. **Rollback via Git revert** — Keeps history clean
6. **Separate platform and service apps** — Different lifecycle and permissions
7. **Use health checks** — ArgoCD waits for resources to be healthy before marking sync complete

---

## Next Steps

1. Install ArgoCD → [Setup Guide](./setup-guide.md) Section 4
2. Create app-of-apps → `argocd/platform/app-of-apps.yaml`
3. Create service ApplicationSet → `argocd/services/service-appset.yaml`
4. Connect with Backstage → [Backstage Guide](./backstage-guide.md)

Related: [Architecture Plan](./idp-architecture-plan.md) | [Project Roadmap](./project-roadmap.md)
