---
description: Analyze the codebase, scaffold IDP boilerplate structure, and generate initial documentation
---

# IDP Project – Initialization Workflow

## Purpose

This workflow initializes a new Internal Developer Platform (IDP) project. It performs three things:

1. **Scout** — Analyze what already exists in the repository
2. **Scaffold** — Create the IDP boilerplate folder structure (folders only, no code)
3. **Document** — Generate structured documentation in `docs/`

This process is strictly **analysis, scaffolding, and documentation** — no implementation code.

---

## 1. Workflow Overview

```
Phase 1: Scout     → Discover what exists (code, infra, config, templates)
Phase 2: Scaffold  → Create missing IDP directories from the boilerplate layout
Phase 3: Document  → Generate docs/ files based on what was discovered
Phase 4: Validate  → Verify completeness and quality
```

---

## 2. Phase 1: Project Scouting

### 2.1 Detect Project Type

Scan the repository root and identify which IDP components already exist:

- **Developer Portal** — Backstage app (`backstage/`, `app-config.yaml`)
- **Infrastructure Orchestration** — Crossplane (`crossplane/`, XRDs, compositions)
- **GitOps** — ArgoCD (`argocd/`, ApplicationSets)
- **Infrastructure** — K8s manifests, operators, monitoring (`infrastructure/`)
- **Service Templates** — Go/Node/Python boilerplates (e.g. `go-service-template/`)
- **CI/CD** — Pipeline definitions (`.github/workflows/`, `.gitlab-ci.yml`, `Jenkinsfile`)
- **Documentation** — Existing `docs/`, `README.md`, architecture docs

### 2.2 Analyze Existing Components

For each detected component:

- Identify technologies and versions (from `go.mod`, `package.json`, YAML configs)
- Assess maturity: **complete**, **partial** (stubbed/commented-out), or **placeholder**
- Map dependencies between components
- Note what is configured vs. hardcoded vs. missing

### 2.3 Scope Rules

**Scan:**
- All top-level directories
- Configuration files (YAML, JSON, TOML, `.env`)
- Makefiles, Dockerfiles, docker-compose files
- Proto definitions, API specs
- K8s manifests, Helm charts, Kustomize overlays
- CI/CD pipeline definitions

**Exclude:**
- `.git`, `.agent`, `.opencode`
- `node_modules`, `vendor`, `__pycache__`
- Build artifacts, binary outputs
- Secret files (`.env` values can be noted, but never copied into docs)

### 2.4 Adaptive Discovery

- Do NOT assume a fixed repository structure
- Detect existing directories dynamically
- Analyze only what actually exists
- If a service template exists, analyze its internal architecture

### 2.5 Expected Output

A structured summary covering:

- Detected IDP components and their maturity
- Technology stack inventory
- Service template architecture (if present)
- Infrastructure components (operators, providers, compositions)
- CI/CD pipeline status
- Gaps (what the architecture needs but doesn't have yet)

---

## 3. Phase 2: Boilerplate Scaffolding

Create the standard IDP directory structure. **Only create directories that don't already exist.** Add `.gitkeep` to empty directories so Git tracks them.

### 3.1 Standard IDP Layout

```
{project-root}/
├── backstage/                     # Developer Portal
│   ├── catalog/                   # Service catalog definitions
│   │   └── systems/               # System entity definitions
│   ├── templates/                 # Software templates
│   │   └── {template-name}/       # One dir per template
│   │       └── skeleton/          # Template skeleton files
│   └── packages/                  # Custom Backstage plugins
│       └── backend/
│           └── src/
│               └── plugins/
│
├── crossplane/                    # Infrastructure Orchestration
│   ├── providers/                 # Crossplane provider configs
│   ├── configurations/            # Platform configurations
│   ├── compositions/              # XRDs and compositions
│   └── claims/                    # Example claims
│
├── infrastructure/                # Kubernetes Infrastructure
│   ├── namespaces/                # Namespace definitions
│   ├── databases/                 # Database operator configs
│   ├── cache/                     # Cache operator configs
│   ├── messaging/                 # Message broker configs
│   ├── secrets/                   # Secret management
│   │   └── vault/                 # Vault configs
│   └── monitoring/                # Observability stack
│       ├── prometheus/
│       ├── grafana/
│       │   └── dashboards/
│       └── loki/
│
├── argocd/                        # GitOps Configuration
│   ├── platform/                  # Platform app-of-apps
│   └── services/                  # Service ApplicationSets
│
├── .github/                       # CI/CD
│   └── workflows/                 # Pipeline definitions
│
├── docs/                          # Documentation
│
└── {service-template}/            # Service template(s) — detect existing
```

### 3.2 Scaffolding Rules

- Never overwrite existing directories or files
- Only create what is missing from the standard layout
- If a service template already exists (e.g. `go-service-template/`), leave it as-is
- Name template directories under `backstage/templates/` to match existing service templates
- If the project has no service template yet, skip `backstage/templates/` skeleton

---

## 4. Phase 3: Documentation Generation

All generated documentation lives under `docs/`.

### 4.1 Required Files

| File | Description |
|------|-------------|
| `docs/idp-architecture-plan.md` | IDP architecture: components, tech stack, developer workflow, implementation phases |
| `docs/codebase-summary.md` | Repository structure, detected components, technology inventory |
| `docs/project-roadmap.md` | Phased implementation plan with checkboxes, timelines, deliverables |
| `docs/developer-workflow.md` | Step-by-step flow of what happens when a developer creates a new service |
| `docs/setup-guide.md` | Prerequisites, tool installation, step-by-step cluster + component setup |
| `docs/crossplane-guide.md` | XRDs, Compositions, Claims — how infrastructure-as-code works in the IDP |
| `docs/argocd-gitops-guide.md` | GitOps concepts, ArgoCD Applications, app-of-apps, ApplicationSets, sync policies |
| `docs/backstage-guide.md` | Software templates, catalog, scaffolder actions, plugins, developer portal setup |
| `docs/observability-guide.md` | Three pillars (metrics/logs/traces), Prometheus, Grafana, Loki, Jaeger, alerting |
| `docs/testing-guide.md` | 6-layer IDP testing strategy, chainsaw, kubeconform, E2E workflow tests |
| `README.md` | Project overview, quick start, component links |

### 4.2 Conditional Files

Generate only if relevant components exist:

| Condition | File | Description |
|-----------|------|-------------|
| Service template exists | `docs/{template}-guide.md` | Template architecture, local dev setup, conventions |
| Kafka/messaging detected | `docs/kafka-integration-guide.md` | Event-driven patterns, topic naming, producer/consumer usage |
| Multiple environments | `docs/deployment-guide.md` | Environment promotion, rollback procedures |

### 4.3 Documentation Content Rules

**Architecture Plan** must cover:
- High-level architecture diagram (ASCII)
- Technology stack table (component, technology, version, purpose)
- Component breakdown (Backstage, Crossplane, ArgoCD, infrastructure, service template)
- Boilerplate structure with directory descriptions
- Implementation phases (derived from what exists vs. what's needed)

**Project Roadmap** must:
- Align phases with the architecture plan
- Use checkboxes (`- [ ]`) for trackable tasks
- Include timeline estimates per phase
- Define deliverables per phase
- End with a milestone summary table

**Developer Workflow** must describe the target end-to-end experience:
- Developer accesses portal → selects template → fills form → submits
- Backstage scaffolds → creates repo → creates infrastructure claim → registers in catalog
- Crossplane provisions infrastructure (DB, cache, messaging)
- CI/CD builds and deploys
- Service is running with monitoring

**README.md** constraints:
- Maximum 300 lines
- Focus on: what this project is, how to get started, links to docs
- Do not repeat content from `docs/` files

### 4.4 General Documentation Rules

- Maximum **800 lines per file** — split if larger
- Write for the audience: platform engineers and service developers
- Use ASCII diagrams over image references
- Reference other docs with relative links (`./docs/...`)
- Never include secrets, credentials, or internal URLs
- Base content on what was actually discovered — do not fabricate components

---

## 5. Phase 4: Validation

### 5.1 Structure Check

Verify:
- [ ] All required `docs/` files were generated
- [ ] Boilerplate directories exist with `.gitkeep` where empty
- [ ] No existing files were overwritten
- [ ] README.md exists and is under 300 lines

### 5.2 Content Quality

Documentation must be:
- **Accurate** — matches what actually exists in the repo
- **Non-redundant** — no copy-paste between files
- **Actionable** — tells the reader what to do, not just what exists
- **Navigable** — cross-linked between related docs
- **Consistent** — same terminology throughout

### 5.3 Completeness Report

Output a summary:

```
=== IDP Init Complete ===

Detected components:
  - [list of what was found]

Scaffolded directories:
  - [list of new directories created]

Generated documentation:
  - [list of docs/ files created]

Gaps identified:
  - [list of components needed but not yet present]
```

---

## 6. Constraints

This workflow MUST NOT:

- Write implementation code (Go, TypeScript, Python, etc.)
- Modify existing source files
- Change infrastructure configurations
- Create Kubernetes manifests with actual resource definitions
- Introduce dependencies or install packages

It is exclusively for:

> **Discovery → Scaffolding → Documentation**

---

## 7. Reusability

This workflow is designed to be reused across IDP projects:

- The boilerplate layout is the **standard** — adapt directory names only if the project uses a different convention
- Documentation templates follow a consistent structure regardless of tech stack
- Phase 1 scouting adapts to whatever exists — it does not assume Go, Backstage, or any specific technology
- If a project uses a different service template language (Node, Python, Rust), the workflow still applies — just detect and document what's there
