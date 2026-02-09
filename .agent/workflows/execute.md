---
description: Execute an IDP implementation plan — build, validate, and finalize platform components
name: /execute
argument-hint: [path-to-plan-or-phase]
---

# Workflow: IDP Plan Execution & Validation

## Step 0: Plan Activation

1. **Locate Target Plan**
   - If argument provided → load the specified plan/phase file
   - If no argument → load the latest plan from `plans/`
   - Detect the first task with status: `pending` or `in_progress`

2. **Bind Context**
   - Read `CLAUDE.md` for project conventions
   - Read relevant component guides from `docs/`
   - Read `docs/project-roadmap.md` to understand phase context
   - Identify the IDP component being implemented:
     - Crossplane (XRDs, compositions, claims)
     - ArgoCD (Applications, ApplicationSets)
     - Backstage (templates, catalog, plugins)
     - Infrastructure (operators, namespaces, monitoring)
     - Service Template (go-service-template parameterization)
     - CI/CD (GitHub Actions workflows)

3. **Prerequisite Check**
   - Verify dependencies from the plan are met
   - Check that required operators/tools are available
   - If prerequisites are missing, report and stop

**Output:**
`Step 0: [Component] — [Task Name] Activated`

---

## Step 1: Analysis & Task Extraction

1. **Read the Plan**
   - Extract all tasks and sub-tasks from the plan file
   - Identify the implementation order (dependencies, sync waves)
   - Note acceptance criteria for each task

2. **Scan Current State**
   - Check what already exists in the target directories
   - Identify files to create vs. files to modify
   - Verify no conflicts with existing resources

3. **Identify Resource Types**

   | Component | Resource Types |
   |-----------|---------------|
   | **Crossplane** | XRD YAML, Composition YAML, Claim YAML, ProviderConfig YAML |
   | **ArgoCD** | Application YAML, ApplicationSet YAML, AppProject YAML |
   | **Backstage** | template.yaml, skeleton files, catalog-info.yaml, TypeScript plugins |
   | **Infrastructure** | Namespace YAML, Helm values, operator CRDs, ServiceMonitor, PrometheusRule |
   | **Service Template** | Go source, Dockerfile, Makefile, docker-compose.yaml, K8s manifests |
   | **CI/CD** | GitHub Actions YAML (.github/workflows/) |

4. **Dependency Mapping**
   - Which resources must be created first?
   - Which resources reference others (e.g., Composition references XRD)?
   - Which secrets/configs must exist before resources can be applied?

**Output:**
`Step 1: Found [N] tasks — Dependencies: [list or none]`

---

## Step 2: Implementation

### 2.1 Implementation Rules

- **Follow existing patterns** — check similar resources in the repo for conventions
- **Use the guides** — reference `docs/` guides for correct resource structure
- **One resource per file** — keep YAML files focused (exception: small related resources)
- **Add labels consistently** — all resources get `app.kubernetes.io/part-of: idp-platform`
- **Comment complex logic** — explain non-obvious configuration choices
- **No hardcoded secrets** — use Vault references, K8s secrets, or environment variables

### 2.2 Component-Specific Implementation

#### Crossplane Resources
```
Implementation order:
1. Provider installations (if not already installed)
2. ProviderConfigs
3. XRD definitions (CompositeResourceDefinition)
4. Compositions (map XRD → actual resources)
5. Example Claims (for testing)

Key rules:
- XRD schema must match what Backstage template form collects
- Compositions use patches to map XRD spec → resource spec
- Size tiers (small/medium/large) use transform maps
- Secrets must be created in the service namespace
- Use spec.writeConnectionSecretToRef for credential propagation
```

#### ArgoCD Resources
```
Implementation order:
1. AppProject definitions (RBAC boundaries)
2. Root Application (app-of-apps)
3. Platform Applications (Crossplane, monitoring, Vault, etc.)
4. Service ApplicationSet (auto-discovery)

Key rules:
- All Applications live in the ArgoCD namespace (platform)
- Use sync waves: namespaces (0) → operators (1) → configs (2) → apps (3)
- Enable automated sync with prune + selfHeal for platform apps
- Use manual sync for production service deployments (optional)
- ApplicationSet generators must match repo structure
```

#### Backstage Resources
```
Implementation order:
1. app-config.yaml (main configuration)
2. Catalog entities (systems, groups)
3. Template definition (template.yaml)
4. Skeleton files (code that gets copied)
5. Custom scaffolder actions (TypeScript)
6. Plugin configuration

Key rules:
- Template form fields must map to Crossplane claim spec
- Skeleton uses ${{ values.xxx }} for variable substitution
- catalog-info.yaml must be in every skeleton for auto-registration
- Custom actions need unit tests
```

#### Infrastructure Resources
```
Implementation order:
1. Namespaces
2. RBAC (ServiceAccounts, Roles, RoleBindings)
3. Operators (via Helm or direct YAML)
4. Operator configurations (CRDs, custom resources)
5. Monitoring integration (ServiceMonitors, dashboards)

Key rules:
- Each operator goes in its designated namespace
- Use Helm for operator installation when available
- Pin operator versions explicitly
- Create ServiceMonitor for every operator
- Add Grafana dashboard JSON for key metrics
```

#### Service Template Resources
```
Implementation order:
1. Template variable substitution markers
2. Conditional file generation (database yes/no)
3. K8s manifest templates (deployment, service, configmap)
4. Crossplane claim template
5. CI/CD pipeline templates
6. catalog-info.yaml template
7. Local dev setup (docker-compose, Makefile targets)

Key rules:
- Follow go-zero framework conventions
- Maintain registry pattern (BaseContext → ServiceContext)
- Template variables use Backstage syntax: ${{ values.serviceName }}
- K8s manifests must reference secrets created by Crossplane
- CI/CD templates must be parameterized for service name and registry
```

#### CI/CD Pipelines
```
Implementation order:
1. CI workflow (lint → test → scan → build → push)
2. CD workflow (update manifests → trigger ArgoCD)
3. Platform CI (validate YAML, lint Crossplane compositions)

Key rules:
- Use GitHub Actions
- Pin action versions with SHA
- Store secrets in GitHub repository secrets
- Use multi-stage Docker builds
- Tag images with git SHA + semver
```

### 2.3 Static Verification

After creating resources, verify syntax and structure:

| Resource Type | Verification |
|---------------|-------------|
| YAML manifests | `kubectl apply --dry-run=client -f <file>` |
| Crossplane XRDs | `kubectl apply --dry-run=server -f <file>` (requires Crossplane) |
| Helm values | `helm template <chart> -f <values>` |
| Go code | `go build ./...` and `go vet ./...` |
| TypeScript | `npm run build` or `npx tsc --noEmit` |
| GitHub Actions | Validate YAML structure, check action versions exist |

**Output:**
`Step 2: Implemented [N] resources — Verification passed`

---

## Step 3: Validation

### 3.1 Component Validation

#### Crossplane
```bash
# Verify XRD is installed
kubectl get xrd

# Verify composition exists and is valid
kubectl get composition

# Test with an example claim
kubectl apply -f crossplane/claims/example.yaml

# Watch provisioning
kubectl get managed -w

# Verify resources were created
kubectl get all -n services -l app=<service-name>

# Verify secrets were created
kubectl get secrets -n services

# Clean up test claim
kubectl delete -f crossplane/claims/example.yaml
```

#### ArgoCD
```bash
# Verify Applications are synced
argocd app list

# Check app health
argocd app get <app-name>

# Verify ApplicationSet generates expected Applications
kubectl get applicationset -n platform
kubectl get application -n platform
```

#### Backstage
```bash
# Start Backstage locally
cd backstage && yarn dev

# Verify:
# - Template appears in "Create" page
# - Form fields render correctly
# - Catalog entities load
# - Plugins show data (K8s, ArgoCD)
```

#### Infrastructure
```bash
# Verify operators are running
kubectl get pods -n <operator-namespace>

# Verify CRDs are installed
kubectl get crd | grep <operator>

# Test creating an instance
kubectl apply -f <test-instance.yaml>

# Verify monitoring integration
kubectl get servicemonitor -A
```

### 3.2 Integration Validation

If multiple components were implemented, test the integration:

```
Backstage template → creates Crossplane claim → provisions infrastructure → ArgoCD deploys
```

Test the developer workflow end-to-end if possible.

### 3.3 Quality Gate

- All resources apply without errors
- No `CrashLoopBackOff` or `Error` status on pods
- Crossplane claims reach `Ready` state
- ArgoCD apps reach `Synced` + `Healthy` state
- Monitoring shows metrics from new components

**Output:**
`Step 3: Validation [X/X passed] — Stable`

---

## Step 4: Review Gate

### 4.1 Platform Engineering Review Checks

| Category | Check |
|----------|-------|
| **Security** | No hardcoded secrets, RBAC scoped correctly, network policies considered |
| **Reliability** | Health checks defined, resource limits set, HA configured where needed |
| **Observability** | ServiceMonitors created, dashboards defined, alerting rules set |
| **Consistency** | Labels match conventions, naming follows patterns, versions pinned |
| **Documentation** | Inline comments on complex config, README updated if needed |
| **Idempotency** | Resources can be re-applied safely (no duplicate creation) |
| **Cleanup** | Deleting a claim/resource cascades correctly, no orphaned resources |

### 4.2 Scorecard

Generate a score (1-10) with issues categorized:

- **Critical** — Security risk, data loss risk, breaks existing resources
- **Warning** — Missing observability, suboptimal configuration, no HA
- **Suggestion** — Code style, documentation, optional improvements

### 4.3 User Decision

**BLOCKING STEP** — System must WAIT for user input:

`[Approve / Fix / Abort]`

**Output:**
`Step 4: Review complete — Score [N]/10 — Waiting approval`

---

## Step 5: Finalization

1. **Update Roadmap**
   - Mark completed tasks in `docs/project-roadmap.md` (`- [ ]` → `- [x]`)
   - Note completion date

2. **Update Documentation**
   - Update relevant guide in `docs/` if implementation diverged from the plan
   - Add any new learnings or gotchas

3. **Update Plan**
   - Mark the task/phase as `DONE` in the plan file

4. **Commit**
   - Use Conventional Commits format:
     - `feat(crossplane): add GoService XRD and compositions`
     - `feat(argocd): configure app-of-apps and service ApplicationSet`
     - `feat(backstage): create Go service software template`
     - `feat(infra): install CloudNativePG and Redis operators`
     - `feat(monitoring): add service dashboard and alert rules`
     - `feat(ci): add CI/CD pipeline templates`
     - `chore(docs): update roadmap with Phase X completion`

**Output:**
`Step 5: Phase finalized — Code committed`

---

## Execution Rules

### General
- Always read the plan before implementing
- Implement in dependency order (prerequisites first)
- Verify each resource after creation before moving to the next
- If a resource fails, diagnose before retrying
- Never delete existing working resources without user confirmation

### Platform Engineering Priorities
1. **Correctness** — resources must work as intended
2. **Security** — no exposed secrets, proper RBAC, least privilege
3. **Reliability** — health checks, resource limits, HA where needed
4. **Observability** — every component must be monitored
5. **Simplicity** — prefer straightforward solutions over clever ones

### What NOT to Do
- Do not install operators that aren't in the plan
- Do not change cluster-wide settings without user confirmation
- Do not modify existing services that are running in production
- Do not skip the review gate (Step 4)
- Do not commit without running validation (Step 3)

---

## Post-Execution Commands

- `/plan` — plan the next phase or component
- `/execute` — execute the next pending task
- `/clear` — reset context for a fresh start

---

## Reusability

This execution workflow adapts to any IDP project:
- Component types map to the project's specific technologies
- Validation commands adjust to the cluster environment
- Review checks apply universally to platform engineering work
- The 5-step structure (Activate → Analyze → Implement → Validate → Review → Finalize) works for any infrastructure component
