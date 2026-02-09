# Platform Roadmap — IDP Demo (Learn by Doing)

This is a simplified, hands-on roadmap for learning how to build an Internal Developer Platform. Each phase adds one concept, and every phase ends with a working demo you can see and touch.

**Approach:** Start small, get it working, understand it, then add the next layer.

**What we're building:** A developer fills out a Backstage form → a Go service with PostgreSQL is automatically created and running on Kubernetes.

**What we're NOT doing yet:** Production HA, Vault, Kafka, Loki, Jaeger, multi-environment, security scanning. Those come later after you understand the fundamentals.

---

## How to Read This Roadmap

Each phase follows this pattern:

```
What you'll learn  →  What you'll build  →  "It works!" moment
```

Phases are designed to be done in order. Don't skip ahead — each one builds on the previous.

---

## Phase 1 — Local Kubernetes Cluster (Day 1)

**What you'll learn:** How to run Kubernetes locally and deploy a simple app.

**Why this matters:** Everything in the IDP runs on Kubernetes. You need a cluster first.

### Tasks

- [ ] Install required tools: `kubectl`, `helm`, `kind`, `docker`
- [ ] Create a kind cluster with the config from [Setup Guide](./setup-guide.md)
- [ ] Verify the cluster is running: `kubectl get nodes`
- [ ] Create namespaces:
  ```bash
  kubectl create namespace platform
  kubectl create namespace services
  ```
- [ ] Deploy a test nginx pod to prove the cluster works:
  ```bash
  kubectl run test-nginx --image=nginx --namespace=services
  kubectl get pods -n services
  # You should see: test-nginx   1/1   Running
  ```
- [ ] Clean up the test: `kubectl delete pod test-nginx -n services`

### "It Works!" Moment

You have a running Kubernetes cluster on your laptop. You can deploy and delete pods.

### What You Learned

- `kind` creates a local K8s cluster inside Docker
- `kubectl` talks to the cluster
- Namespaces organize resources (like folders)

### Deliverable

- [ ] Kind cluster running with `platform` and `services` namespaces

---

## Phase 2 — Your First Operator: PostgreSQL (Day 2)

**What you'll learn:** What a Kubernetes operator is and how it automates database management.

**Why this matters:** Instead of manually installing PostgreSQL, an operator does it for you. This is the foundation of automated infrastructure.

### What is an Operator?

```
Without operator:  You SSH into a server → install PostgreSQL → configure it → maintain it
With operator:     You write a YAML file → operator creates and manages PostgreSQL for you
```

### Tasks

- [ ] Install CloudNativePG operator:
  ```bash
  helm repo add cnpg https://cloudnative-pg.github.io/charts
  helm repo update
  helm install cnpg cnpg/cloudnative-pg --namespace databases --create-namespace
  ```
- [ ] Verify operator is running:
  ```bash
  kubectl get pods -n databases
  # You should see: cnpg-cloudnative-pg-xxx   1/1   Running
  ```
- [ ] Create a PostgreSQL database by applying a YAML:
  ```yaml
  # test-database.yaml
  apiVersion: postgresql.cnpg.io/v1
  kind: Cluster
  metadata:
    name: my-first-db
    namespace: services
  spec:
    instances: 1
    storage:
      size: 1Gi
  ```
  ```bash
  kubectl apply -f test-database.yaml
  ```
- [ ] Watch the database come to life:
  ```bash
  kubectl get clusters -n services -w
  # Wait until: my-first-db   1   1   Cluster in healthy state
  ```
- [ ] Check what the operator created automatically:
  ```bash
  kubectl get pods -n services
  # my-first-db-1   (the actual PostgreSQL pod)

  kubectl get secrets -n services
  # my-first-db-app   (username + password — auto-generated!)
  ```
- [ ] Clean up: `kubectl delete cluster my-first-db -n services`

### "It Works!" Moment

You wrote 8 lines of YAML and got a running PostgreSQL with auto-generated credentials. No manual installation.

### What You Learned

- Operators watch for custom resources (like `Cluster`) and create real infrastructure
- The operator auto-generated secrets (username/password)
- Deleting the custom resource cleans up everything

### Deliverable

- [ ] CloudNativePG operator running
- [ ] You understand: YAML in → database out

---

## Phase 3 — Crossplane: Automate Everything (Day 3-4)

**What you'll learn:** How Crossplane lets you define a custom API (`GoService`) that creates multiple resources with one YAML.

**Why this matters:** This is the magic of the IDP. Instead of creating a Deployment, Service, and Database separately, Crossplane creates ALL of them from a single claim.

### What is Crossplane?

```
Without Crossplane:
  kubectl apply -f deployment.yaml    (manual)
  kubectl apply -f service.yaml       (manual)
  kubectl apply -f database.yaml      (manual)
  kubectl apply -f secrets.yaml       (manual)

With Crossplane:
  kubectl apply -f goservice-claim.yaml   (one file → everything created)
```

### Tasks

#### 3.1 Install Crossplane

- [ ] Install Crossplane:
  ```bash
  helm repo add crossplane-stable https://charts.crossplane.io/stable
  helm repo update
  helm install crossplane crossplane-stable/crossplane \
    --namespace crossplane-system --create-namespace
  ```
- [ ] Install the Kubernetes provider (lets Crossplane create K8s resources):
  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: pkg.crossplane.io/v1
  kind: Provider
  metadata:
    name: provider-kubernetes
  spec:
    package: xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.14.1
  EOF
  ```
- [ ] Wait for provider to be healthy:
  ```bash
  kubectl wait provider provider-kubernetes --for=condition=Healthy --timeout=120s
  ```
- [ ] Configure provider credentials:
  ```bash
  cat <<EOF | kubectl apply -f -
  apiVersion: kubernetes.crossplane.io/v1alpha1
  kind: ProviderConfig
  metadata:
    name: default
  spec:
    credentials:
      source: InjectedIdentity
  EOF
  ```

#### 3.2 Create a Simple XRD (Your Custom API)

Start simple — a `SimpleService` that only creates a Deployment + Service:

- [ ] Create the XRD:
  ```yaml
  # crossplane/compositions/simple-service-xrd.yaml
  apiVersion: apiextensions.crossplane.io/v1
  kind: CompositeResourceDefinition
  metadata:
    name: simpleservices.demo.example.com
  spec:
    group: demo.example.com
    names:
      kind: SimpleService
      plural: simpleservices
    claimNames:
      kind: SimpleServiceClaim
      plural: simpleserviceclaims
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
                  image:
                    type: string
                    default: nginx:latest
                  replicas:
                    type: integer
                    default: 1
                required:
                  - serviceName
  ```

- [ ] Create the Composition (what happens when someone creates a claim):
  ```yaml
  # crossplane/compositions/simple-service-composition.yaml
  apiVersion: apiextensions.crossplane.io/v1
  kind: Composition
  metadata:
    name: simple-service
  spec:
    compositeTypeRef:
      apiVersion: demo.example.com/v1alpha1
      kind: SimpleService
    resources:
      - name: deployment
        base:
          apiVersion: kubernetes.crossplane.io/v1alpha2
          kind: Object
          spec:
            forProvider:
              manifest:
                apiVersion: apps/v1
                kind: Deployment
                metadata:
                  namespace: services
                spec:
                  selector:
                    matchLabels: {}
                  template:
                    spec:
                      containers:
                        - name: app
                          ports:
                            - containerPort: 80
        patches:
          - fromFieldPath: spec.serviceName
            toFieldPath: spec.forProvider.manifest.metadata.name
          - fromFieldPath: spec.serviceName
            toFieldPath: spec.forProvider.manifest.spec.selector.matchLabels.app
          - fromFieldPath: spec.serviceName
            toFieldPath: spec.forProvider.manifest.spec.template.metadata.labels.app
          - fromFieldPath: spec.image
            toFieldPath: spec.forProvider.manifest.spec.template.spec.containers[0].image
          - fromFieldPath: spec.replicas
            toFieldPath: spec.forProvider.manifest.spec.replicas

      - name: service
        base:
          apiVersion: kubernetes.crossplane.io/v1alpha2
          kind: Object
          spec:
            forProvider:
              manifest:
                apiVersion: v1
                kind: Service
                metadata:
                  namespace: services
                spec:
                  type: ClusterIP
                  ports:
                    - port: 80
                      targetPort: 80
        patches:
          - fromFieldPath: spec.serviceName
            toFieldPath: spec.forProvider.manifest.metadata.name
          - fromFieldPath: spec.serviceName
            toFieldPath: spec.forProvider.manifest.spec.selector.app
  ```

- [ ] Apply both:
  ```bash
  kubectl apply -f crossplane/compositions/simple-service-xrd.yaml
  kubectl apply -f crossplane/compositions/simple-service-composition.yaml
  ```

#### 3.3 Test It — Create a Claim

- [ ] Create a claim (this is what a developer would submit):
  ```yaml
  # crossplane/claims/test-simple.yaml
  apiVersion: demo.example.com/v1alpha1
  kind: SimpleServiceClaim
  metadata:
    name: hello-world
    namespace: services
  spec:
    serviceName: hello-world
    image: nginx:latest
    replicas: 2
  ```
  ```bash
  kubectl apply -f crossplane/claims/test-simple.yaml
  ```
- [ ] Watch resources appear:
  ```bash
  kubectl get pods -n services -w
  # hello-world-xxx   1/1   Running   (2 pods!)

  kubectl get svc -n services
  # hello-world   ClusterIP   ...
  ```
- [ ] Clean up: `kubectl delete simpleserviceclaim hello-world -n services`
- [ ] Verify cleanup: `kubectl get pods -n services` — should be empty

### "It Works!" Moment

You wrote ONE claim YAML (5 lines of spec) and Crossplane created a Deployment with 2 pods AND a Service. Delete the claim → everything disappears.

### What You Learned

- **XRD** = your custom API definition (what fields are available)
- **Composition** = the implementation (what resources to create)
- **Claim** = what a developer submits (simple YAML)
- Crossplane uses **patches** to map claim fields → resource fields

### Deliverable

- [ ] Crossplane installed with Kubernetes provider
- [ ] SimpleService XRD + Composition working
- [ ] Claim creates Deployment + Service; delete claim cleans up

---

## Phase 4 — Add Database to Composition (Day 4-5)

**What you'll learn:** How to extend a Crossplane composition to create multiple types of infrastructure (app + database).

**Why this matters:** This is where the IDP becomes powerful — ONE claim creates app AND database AND secrets, all wired together.

### Tasks

#### 4.1 Upgrade XRD — Add Database Fields

- [ ] Create a new XRD `GoService` that extends SimpleService with database options:
  ```yaml
  # crossplane/compositions/goservice-xrd.yaml
  ```
  Add these fields to the spec:
  - `database.enabled` (boolean, default: true)
  - `database.size` (string enum: small/medium/large, default: small)

  See [Crossplane Guide](./crossplane-guide.md) Section 2.2 for the full XRD schema.

#### 4.2 Create GoService Composition

- [ ] New composition that creates:
  1. Kubernetes Deployment (with env vars from DB secret)
  2. Kubernetes Service
  3. CloudNativePG Cluster (when database.enabled = true)
- [ ] Map database size to storage:
  - small → 1Gi (for demo, keep small)
  - medium → 5Gi
  - large → 10Gi

  See [Crossplane Guide](./crossplane-guide.md) Section 3 for composition patterns.

#### 4.3 Test the Full Claim

- [ ] Create a GoService claim:
  ```yaml
  # crossplane/claims/test-goservice.yaml
  apiVersion: platform.example.com/v1alpha1
  kind: GoServiceClaim
  metadata:
    name: demo-api
    namespace: services
  spec:
    serviceName: demo-api
    image: nginx:latest
    replicas: 1
    database:
      enabled: true
      size: small
  ```
- [ ] Apply and watch:
  ```bash
  kubectl apply -f crossplane/claims/test-goservice.yaml

  # Watch pods (both app pod AND database pod)
  kubectl get pods -n services -w

  # Check database
  kubectl get clusters.postgresql.cnpg.io -n services

  # Check secrets (auto-generated DB credentials)
  kubectl get secrets -n services | grep demo-api
  ```
- [ ] Verify the app pod has database env vars injected
- [ ] Clean up and verify cascade deletion

### "It Works!" Moment

ONE claim created: a running app + a PostgreSQL database + auto-generated credentials + secrets injected into the app. Delete the claim → everything gone.

### What You Learned

- Compositions can create different types of resources (K8s + CRDs)
- Patches with `transforms: map` convert friendly names (small/medium/large) to real values
- Secrets flow from operator → Crossplane → pod environment variables

### Deliverable

- [ ] GoService XRD with database fields
- [ ] GoService Composition creates Deployment + Service + PostgreSQL
- [ ] Claim tested end-to-end with cleanup verification

---

## Phase 5 — ArgoCD: GitOps Deployment (Day 5-6)

**What you'll learn:** How ArgoCD watches a Git repo and automatically deploys changes to the cluster.

**Why this matters:** In the IDP, no one runs `kubectl apply` manually. ArgoCD syncs everything from Git.

### What is GitOps?

```
Traditional:  Developer → kubectl apply → hope it works
GitOps:       Developer → git push → ArgoCD sees change → applies to cluster automatically
```

### Tasks

#### 5.1 Install ArgoCD

- [ ] Install ArgoCD:
  ```bash
  helm repo add argo https://argoproj.github.io/argo-helm
  helm repo update
  helm install argocd argo/argo-cd \
    --namespace platform \
    --set server.service.type=ClusterIP \
    --set configs.params."server\.insecure"=true
  ```
- [ ] Get the admin password:
  ```bash
  kubectl -n platform get secret argocd-initial-admin-secret \
    -o jsonpath="{.data.password}" | base64 -d; echo
  ```
- [ ] Access the UI:
  ```bash
  kubectl port-forward svc/argocd-server -n platform 8080:443
  # Open: https://localhost:8080
  # Login: admin / <password from above>
  ```

#### 5.2 Create Your First ArgoCD Application

- [ ] Put your Crossplane compositions in a Git repo (this repo!)
- [ ] Create an ArgoCD Application that watches the `crossplane/` directory:
  ```yaml
  # argocd/platform/crossplane-compositions.yaml
  apiVersion: argoproj.io/v1alpha1
  kind: Application
  metadata:
    name: crossplane-compositions
    namespace: platform
  spec:
    project: default
    source:
      repoURL: <your-git-repo-url>
      targetRevision: main
      path: crossplane/compositions
    destination:
      server: https://kubernetes.default.svc
    syncPolicy:
      automated:
        prune: true
        selfHeal: true
  ```
- [ ] Apply it:
  ```bash
  kubectl apply -f argocd/platform/crossplane-compositions.yaml
  ```

#### 5.3 Test GitOps Flow

- [ ] Change a composition in Git (e.g., change default replicas)
- [ ] Push to main
- [ ] Watch ArgoCD UI — it auto-detects the change and syncs
- [ ] Verify the change applied to the cluster

#### 5.4 Test Self-Healing

- [ ] Manually delete a resource that ArgoCD manages:
  ```bash
  kubectl delete composition simple-service
  ```
- [ ] Watch ArgoCD recreate it within seconds (self-heal)

### "It Works!" Moment

You push a change to Git → ArgoCD automatically applies it. You manually delete a resource → ArgoCD recreates it. Git is the source of truth.

### What You Learned

- ArgoCD continuously watches Git and syncs to the cluster
- `selfHeal: true` means manual changes get reverted
- `prune: true` means resources deleted from Git get deleted from cluster
- The ArgoCD UI shows sync status, diffs, and history

### Deliverable

- [ ] ArgoCD installed and accessible
- [ ] At least one Application syncing from Git
- [ ] Self-healing tested

---

## Phase 6 — Backstage: Developer Portal (Day 7-9)

**What you'll learn:** How to create a self-service form that triggers the entire IDP pipeline.

**Why this matters:** This is the user-facing part. Developers don't write YAML — they fill out a form.

### Tasks

#### 6.1 Install Backstage Locally

- [ ] Create a Backstage app:
  ```bash
  npx @backstage/create-app@latest --name idp-portal
  cd idp-portal
  ```
- [ ] Start in development mode:
  ```bash
  yarn dev
  # Opens http://localhost:3000
  ```
- [ ] Explore the default UI — catalog, create page, docs

#### 6.2 Create a Simple Template

Start with a minimal template that just shows a form:

- [ ] Create the template definition:
  ```yaml
  # backstage/templates/go-service-template/template.yaml
  apiVersion: scaffolder.backstage.io/v1beta3
  kind: Template
  metadata:
    name: go-service-demo
    title: Go Service (Demo)
    description: Create a Go service with PostgreSQL
  spec:
    owner: platform-team
    type: service
    parameters:
      - title: Service Information
        required:
          - serviceName
        properties:
          serviceName:
            title: Service Name
            type: string
            pattern: "^[a-z][a-z0-9-]*$"
          description:
            title: Description
            type: string
          database:
            title: Include PostgreSQL?
            type: boolean
            default: true
          databaseSize:
            title: Database Size
            type: string
            enum: [small, medium, large]
            default: small
            description: "small: 1GB, medium: 5GB, large: 10GB"

    steps:
      - id: log
        name: Log Inputs
        action: debug:log
        input:
          message: "Creating service: ${{ parameters.serviceName }}"

    output:
      links:
        - title: Check cluster
          url: https://localhost:8080
  ```

- [ ] Register the template in Backstage:
  Add to `app-config.yaml`:
  ```yaml
  catalog:
    locations:
      - type: file
        target: ../backstage/templates/go-service-template/template.yaml
  ```

- [ ] Verify the template appears in the "Create" page

#### 6.3 Add Real Scaffolder Actions

Replace the `debug:log` step with real actions:

- [ ] **Step 1**: Fetch skeleton (copy go-service-template and replace variables)
- [ ] **Step 2**: Create GitHub repository and push code
- [ ] **Step 3**: Apply Crossplane claim (using `kubectl` or custom action)
- [ ] **Step 4**: Register in catalog

See [Backstage Guide](./backstage-guide.md) Section 3.2 for the full template definition.

#### 6.4 Demo the Full Flow

- [ ] Open Backstage → Click "Create" → Select "Go Service (Demo)"
- [ ] Fill in: name=`demo-api`, database=yes, size=small
- [ ] Click Create
- [ ] Watch:
  - GitHub repo created (if configured)
  - Crossplane claim applied
  - Database + Deployment appear in cluster
  - Service registered in Backstage catalog

### "It Works!" Moment

You filled out a form and a running service with a database appeared on your cluster. No YAML, no kubectl, no tickets.

### What You Learned

- Backstage templates define a form (parameters) and automation (steps)
- Scaffolder actions execute the steps (create repo, apply claim, register)
- The catalog tracks all services
- Backstage connects the developer experience to the platform automation

### Deliverable

- [ ] Backstage running locally
- [ ] At least one software template registered
- [ ] Form submission triggers Crossplane claim creation

---

## Phase 7 — Polish & Demo Day (Day 10)

**What you'll learn:** How all the pieces fit together in a complete IDP demo.

**Why this matters:** Seeing the full loop work end-to-end proves you understand how an IDP works.

### Tasks

- [ ] Run the full demo flow 3 times with different service names
- [ ] Verify each service has its own database
- [ ] Delete a service via claim deletion → verify full cleanup
- [ ] Show the Backstage catalog with all created services
- [ ] Show the ArgoCD dashboard with all synced apps
- [ ] Take notes on what was confusing, what broke, what you'd change

### Demo Script

```
1. Open Backstage (http://localhost:3000)
2. Show empty catalog
3. Click "Create" → "Go Service (Demo)"
4. Fill form: user-service, database=yes, size=small
5. Click "Create"
6. Switch to terminal:
   - kubectl get pods -n services -w  (watch pods appear)
   - kubectl get clusters.postgresql.cnpg.io -n services  (database created)
   - kubectl get secrets -n services  (credentials auto-generated)
7. Switch to ArgoCD UI (https://localhost:8080)
   - Show app synced and healthy
8. Switch back to Backstage
   - Show user-service in catalog
9. Repeat with order-service
10. Delete user-service claim
    - Show cascade cleanup (pods, DB, secrets all gone)
```

### "It Works!" Moment

You can demo the entire IDP to someone in 5 minutes. They see: form → running service with database.

### Deliverable

- [ ] Full IDP demo working end-to-end
- [ ] At least 2 services created and visible in catalog
- [ ] Cleanup verified

---

## What Comes After the Demo

Once the demo works, you can gradually add production features. Here's the priority order:

### Next: Add More Infrastructure

| Feature | Complexity | Guide |
|---------|-----------|-------|
| Add Redis to composition | Medium | [Crossplane Guide](./crossplane-guide.md) |
| Add Kafka to composition | Hard | [Kafka Guide](./kafka-integration-guide.md) |
| Add monitoring (Prometheus + Grafana) | Medium | [Observability Guide](./observability-guide.md) |
| Add CI/CD pipeline templates | Medium | [ArgoCD Guide](./argocd-gitops-guide.md) |

### Later: Production Hardening

| Feature | Why |
|---------|-----|
| HashiCorp Vault | Proper secret management (not just K8s secrets) |
| Loki + Jaeger | Log aggregation and distributed tracing |
| Grafana dashboards per service | Monitoring visibility |
| Alerting rules | Get notified when things break |
| Multiple environments (dev/staging/prod) | Environment promotion |
| RBAC and network policies | Security |
| Backstage plugins (K8s, ArgoCD, Grafana) | Rich developer portal |
| ApplicationSet for auto-discovery | ArgoCD auto-creates apps for new services |

### Much Later: Scale

| Feature | Why |
|---------|-----|
| Multiple service templates (Node, Python) | Support more languages |
| Multi-cluster deployment | Scale across regions |
| Cost tracking per service | FinOps |
| Service mesh (Istio/Linkerd) | Advanced networking |
| Policy engine (Kyverno/OPA) | Governance |

---

## Milestone Summary

| Phase | Focus | Time | You'll See |
|-------|-------|------|-----------|
| **Phase 1** | Kind cluster | Day 1 | Pods running on your laptop |
| **Phase 2** | PostgreSQL operator | Day 2 | YAML → database appears |
| **Phase 3** | Crossplane basics | Day 3-4 | One claim → Deployment + Service |
| **Phase 4** | Crossplane + database | Day 4-5 | One claim → App + DB + Secrets |
| **Phase 5** | ArgoCD GitOps | Day 5-6 | Git push → auto-deploy |
| **Phase 6** | Backstage portal | Day 7-9 | Form → running service |
| **Phase 7** | Full demo | Day 10 | End-to-end IDP working |

---

## Tips for Learning

1. **Don't rush.** Understand each phase before moving to the next
2. **Break things on purpose.** Delete a pod, a secret, a composition — see what happens
3. **Read the error messages.** `kubectl describe` is your best friend
4. **Keep notes.** Write down what confused you — it helps others learning too
5. **It's OK to restart.** `kind delete cluster` and start fresh is perfectly fine
6. **Refer to the guides.** Each phase links to a detailed guide with more context

Good luck! By Day 10, you'll have built something that most companies take months to set up.
