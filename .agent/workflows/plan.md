---
description: Analyze the IDP roadmap, assess current state, and generate an implementation plan for a specific phase or component
---

# Workflow: IDP Implementation Planning

## Purpose

This workflow helps a platform engineer plan the implementation of a specific IDP phase or component. It analyzes what exists, identifies gaps, and generates a detailed plan — without writing any implementation code.

---

## Phase 1: Context Initialization

### 1.1 Load Project State

1. **Read IDP documentation**:
   - `docs/idp-architecture-plan.md` — understand the target architecture
   - `docs/project-roadmap.md` — identify which phases are done, in progress, or pending
   - `docs/developer-workflow.md` — understand the end-to-end developer experience
   - `CLAUDE.md` — project conventions and patterns

2. **Read relevant component guides** (based on what's being planned):
   - `docs/setup-guide.md` — cluster + operator installation
   - `docs/crossplane-guide.md` — XRDs, compositions, claims
   - `docs/argocd-gitops-guide.md` — GitOps, app-of-apps, ApplicationSets
   - `docs/backstage-guide.md` — templates, catalog, plugins
   - `docs/observability-guide.md` — metrics, logs, traces
   - `docs/kafka-integration-guide.md` — event-driven messaging

3. **Check for existing plans**:
   - Search `plans/` directory for active plans
   - If found, ask: "Active plan detected. Continue, update, or create new?"

### 1.2 Determine Planning Scope

Identify what the user wants to plan. Common scopes:

| Scope | Example | Key Files |
|-------|---------|-----------|
| **Full phase** | "Plan Phase 2" | `docs/project-roadmap.md` Phase 2 section |
| **Single component** | "Plan Crossplane compositions" | `crossplane/`, `docs/crossplane-guide.md` |
| **Integration** | "Plan Backstage + Crossplane integration" | `backstage/`, `crossplane/`, both guides |
| **Improvement** | "Plan monitoring for Go services" | `infrastructure/monitoring/`, `docs/observability-guide.md` |

---

## Phase 2: Current State Assessment

### 2.1 Infrastructure Scan

Check what's already deployed or configured:

```
For each IDP component, assess:
├── Kubernetes cluster — nodes, namespaces, storage classes
├── Crossplane — providers installed? XRDs defined? compositions created?
├── ArgoCD — installed? apps configured? ApplicationSets?
├── Backstage — deployed? templates created? catalog populated?
├── Operators — CloudNativePG? Redis? Strimzi?
├── Monitoring — Prometheus? Grafana? Loki? Jaeger?
├── Vault — installed? auth configured? policies set?
└── CI/CD — GitHub Actions? container registry?
```

### 2.2 Analyze Existing Resources

For the target component(s):

1. **Scan existing files**:
   - YAML manifests in the relevant directory
   - Helm values files
   - Crossplane XRDs, compositions, claims
   - Backstage template.yaml, skeleton files
   - ArgoCD Application/ApplicationSet definitions

2. **Assess maturity**:
   - **Not started** — directory has only `.gitkeep`
   - **Scaffolded** — directory structure exists, no real content
   - **Partial** — some resources defined, incomplete
   - **Complete** — fully implemented and tested
   - **Needs update** — exists but outdated or misconfigured

3. **Check prerequisites**:
   - Does this component depend on another component that isn't ready?
   - Are required operators/CRDs installed?
   - Are required secrets/credentials available?

### 2.3 Gap Analysis

Compare current state against the target architecture:

```
Target (from architecture plan) vs. Current (from scan)
─────────────────────────────────────────────────────
Component A: Target = X, Current = Y → Gap: Z
Component B: Target = X, Current = X → Complete
Component C: Target = X, Current = 0 → Not started
```

---

## Phase 3: Plan Generation

Create: `plans/{YYYY-MM-DD-scope-name}/`

### 3.1 Master Plan (`plan.md`)

```markdown
---
title: [Plan title]
phase: [Roadmap phase number]
status: pending
components: [list of IDP components involved]
prerequisites: [what must be done first]
estimated_effort: [time estimate]
---

# Plan: [Title]

## Objective
[What this plan achieves, aligned with roadmap phase]

## Prerequisites
[What must be in place before starting]

## Tasks
[Ordered list of implementation steps with links to detailed files]

## Acceptance Criteria
[How to verify the plan is complete]

## Risks
[What could go wrong and mitigation strategies]
```

### 3.2 Task Details (`task-XX-name.md`)

For each major task, generate a file containing:

- **Objective**: What this task achieves
- **Component**: Which IDP component (Crossplane, ArgoCD, Backstage, etc.)
- **Resource type**: What kind of resources will be created (YAML manifests, Helm values, TypeScript plugins, Go code, etc.)
- **Implementation steps**: Step-by-step instructions
- **Files to create/modify**: Exact file paths and descriptions
- **Validation**: How to test that this task is complete
- **Dependencies**: What must be done before this task
- **Reference docs**: Links to relevant guides and upstream documentation

### 3.3 IDP-Specific Planning Considerations

When planning, always consider:

#### For Crossplane work:
- XRD schema design (what fields do developers need?)
- Composition resource mapping (XRD fields → actual K8s resources)
- Size tiers (small/medium/large for each resource)
- Secret generation and injection patterns
- Provider versions and compatibility

#### For ArgoCD work:
- App-of-apps hierarchy
- ApplicationSet generators (Git, list, cluster)
- Sync policies (automated vs. manual, prune, self-heal)
- Sync waves for resource ordering
- Health checks and resource tracking

#### For Backstage work:
- Template form UX (what fields, what order, what defaults)
- Scaffolder actions (fetch:template, publish:github, custom actions)
- Skeleton variable substitution patterns
- Catalog entity relationships (Component, System, API, Resource)
- Plugin configuration (Kubernetes, ArgoCD, TechDocs, Grafana)

#### For Infrastructure work:
- Operator versions and compatibility with cluster version
- Resource requests/limits for operators
- Storage class requirements
- Network policies
- RBAC and service accounts

#### For Monitoring work:
- ServiceMonitor definitions for Prometheus scraping
- Grafana dashboard JSON templates
- AlertManager rules and notification channels
- Loki log pipeline configuration
- Jaeger/Tempo sampling configuration

#### For Service Template work:
- go-zero framework conventions
- Registry pattern (BaseContext, ServiceContext, etc.)
- Template variable substitution (`${{ values.xxx }}`)
- Conditional file generation (database yes/no, kafka yes/no)
- CI/CD pipeline template parameterization

---

## Phase 4: Output & Handover

1. Present the plan summary to the user
2. Highlight key decisions that need user input:
   - Technology choices (e.g., which cloud provider, which registry)
   - Size/scale decisions (e.g., cluster size, default resource tiers)
   - Security decisions (e.g., auth method, network policies)
3. **Mandatory reminder**:
   > **Next step:** Review and approve this plan, then run `/execute {path-to-plan}` to begin implementation.
   > Run `/clear` before executing to refresh context.

---

## Constraints

- **NO IMPLEMENTATION**: Do not create YAML manifests, Helm charts, Go code, or TypeScript code during planning
- **NO MODIFICATIONS**: Do not modify existing files or configurations
- **RESEARCH ONLY**: Read files, scan directories, check upstream docs
- **DECISIONS DOCUMENTED**: All architectural decisions must be captured in the plan
- **ALIGNED WITH ROADMAP**: Plans must reference the specific roadmap phase they implement

---

## Planning Templates by Component

### Crossplane Plan Template
```
1. Define XRD schema (fields, types, defaults, validation)
2. Design composition (which K8s resources to create)
3. Map patches (XRD fields → resource fields)
4. Define size tiers (resource allocation per tier)
5. Plan secret management (how credentials flow)
6. Create example claims for testing
```

### ArgoCD Plan Template
```
1. Define Application hierarchy (app-of-apps structure)
2. Design ApplicationSet generators
3. Configure sync policies per environment
4. Plan sync wave ordering
5. Define health check customizations
6. Plan RBAC and project isolation
```

### Backstage Plan Template
```
1. Define template form fields and UX flow
2. Design skeleton directory structure
3. Plan scaffolder action sequence
4. Define catalog entity relationships
5. Plan plugin configuration
6. Design custom scaffolder actions (if needed)
```

### Infrastructure Plan Template
```
1. List operators to install (version, namespace, config)
2. Define Helm values for each operator
3. Plan namespace and RBAC structure
4. Design network policies
5. Plan storage requirements
6. Define monitoring integration (ServiceMonitors, dashboards)
```

---

## Reusability

This workflow adapts to any IDP project:
- Replace component names with project-specific ones
- Adjust technology stack in planning considerations
- Keep the 4-phase structure (Context → Assessment → Generation → Handover)
- Plans always reference the project's roadmap and architecture docs
