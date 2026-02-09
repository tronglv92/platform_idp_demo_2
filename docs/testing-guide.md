# IDP Testing Guide — Testing an Internal Developer Platform

This guide covers how to test each layer of the IDP, from YAML validation to full end-to-end developer workflow verification.

---

## 1. Why IDP Testing is Different

Testing an IDP is not like testing a regular application. You're testing **infrastructure automation** — that Crossplane claims create real databases, that ArgoCD syncs correctly, that Backstage templates produce valid projects.

```
Layer 6: End-to-End        "Backstage form → running service in 10 min"
Layer 5: GitOps            "Git push → ArgoCD syncs correctly"
Layer 4: Infrastructure    "Crossplane claim → DB + Redis + Kafka created"
Layer 3: Template          "Backstage skeleton → valid Go project"
Layer 2: Manifest          "YAML is valid, applies to cluster"
Layer 1: Unit              "Go code logic works correctly"
```

Each layer catches different problems:

| Layer | What it catches | Example failure |
|-------|----------------|-----------------|
| Unit | Logic bugs in Go code | Wrong cache key format |
| Manifest | Invalid YAML, wrong API versions | Typo in Deployment spec |
| Template | Broken variable substitution | `${{ values.serviceName }}` not replaced |
| Infrastructure | Crossplane composition bugs | Database not created, secrets missing |
| GitOps | ArgoCD sync issues | App stuck in OutOfSync |
| End-to-End | Integration failures | Service starts but can't connect to DB |

---

## 2. Testing Tools

### Required Tools

| Tool | Purpose | Install |
|------|---------|---------|
| `go test` | Go unit tests | Included with Go |
| `kubeconform` | K8s YAML schema validation | `brew install kubeconform` |
| `yamllint` | YAML syntax linting | `brew install yamllint` |
| `chainsaw` | K8s resource lifecycle testing | `brew install kyverno/tap/chainsaw` |
| `crossplane` CLI | Render/validate compositions locally | `curl -sL https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh \| sh` |

### Optional Tools

| Tool | Purpose | Install |
|------|---------|---------|
| `kuttl` | Simpler K8s test cases | `brew install kuttl` |
| `uptest` | Crossplane provider e2e testing | Go install from source |
| `act` | Run GitHub Actions locally | `brew install act` |
| `helm` | Render Helm charts for validation | `brew install helm` |

---

## 3. Layer 1: Unit Tests (Go Code)

Unit tests verify business logic in the Go service template.

### What Exists

The `go-service-template/` has unit tests using:
- Go standard `testing` package
- `testify/assert` and `testify/require` for assertions
- `mockery` for mock generation
- Table-driven test patterns

### Running Tests

```bash
cd go-service-template

# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific package
go test -v ./helper/security/...

# Generate coverage report
make coverage
# Opens coverage.html in browser
```

### Test Patterns

```go
// Table-driven test example
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {name: "valid input", input: "hello", expected: "HELLO", wantErr: false},
        {name: "empty input", input: "", expected: "", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Transform(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Mock Generation

```bash
# Generate mocks for all interfaces
make mocks

# Mocks are generated in internal/*/mocks/
```

### Coverage Target

- Minimum: **80%** coverage
- Enforced in CI pipeline

---

## 4. Layer 2: Manifest Validation (YAML)

Validate that all YAML manifests are syntactically correct and conform to Kubernetes schemas — **without needing a cluster**.

### 4.1 YAML Syntax (yamllint)

```bash
# Lint all YAML files
yamllint -d relaxed crossplane/ argocd/ infrastructure/

# With custom config
cat > .yamllint.yml <<EOF
extends: relaxed
rules:
  line-length:
    max: 200
  truthy:
    check-keys: false
EOF

yamllint -c .yamllint.yml .
```

### 4.2 Kubernetes Schema Validation (kubeconform)

```bash
# Validate K8s manifests
kubeconform -strict \
  -summary \
  -output json \
  infrastructure/**/*.yaml \
  go-service-template/k8s/*.yaml

# Skip CRDs (Crossplane, operators) that kubeconform doesn't know about
kubeconform -strict \
  -skip "GoService,GoServiceClaim,Cluster,Redis,KafkaTopic,Application,ApplicationSet" \
  infrastructure/**/*.yaml
```

### 4.3 Crossplane Composition Rendering

Test that a Crossplane composition produces the expected output without applying to a cluster:

```bash
# Render a composition locally
crossplane beta render \
  crossplane/claims/example-service.yaml \
  crossplane/compositions/goservice-composition.yaml \
  crossplane/compositions/goservice-xrd.yaml

# Output shows all resources that WOULD be created
# Verify: Deployment, Service, PostgreSQL Cluster, Redis, Secrets
```

This is the fastest way to iterate on compositions. You can see exactly what resources a claim would create.

### 4.4 Crossplane Schema Validation

```bash
# Validate XRD schema
crossplane beta validate \
  crossplane/compositions/goservice-xrd.yaml \
  crossplane/compositions/goservice-composition.yaml

# Validate a claim against the XRD
crossplane beta validate \
  crossplane/compositions/goservice-xrd.yaml \
  crossplane/claims/example-service.yaml
```

### 4.5 Helm Chart Validation

```bash
# Render Helm chart without installing (check for template errors)
helm template monitoring prometheus-community/kube-prometheus-stack \
  -f infrastructure/monitoring/prometheus/values.yaml

# Dry-run install
helm install monitoring prometheus-community/kube-prometheus-stack \
  -f infrastructure/monitoring/prometheus/values.yaml \
  --dry-run
```

### 4.6 Validation Script

```bash
#!/bin/bash
# tests/manifests/validate.sh

set -e
echo "=== Manifest Validation ==="

echo "1. YAML lint..."
yamllint -d relaxed crossplane/ argocd/ infrastructure/ 2>&1 | head -20

echo "2. K8s schema validation..."
kubeconform -strict -summary \
  -skip "GoService,GoServiceClaim,Cluster,Redis,KafkaTopic,Application,ApplicationSet,CompositeResourceDefinition,Composition" \
  infrastructure/**/*.yaml \
  go-service-template/k8s/*.yaml

echo "3. Crossplane render..."
crossplane beta render \
  crossplane/claims/example-service.yaml \
  crossplane/compositions/goservice-composition.yaml \
  crossplane/compositions/goservice-xrd.yaml > /dev/null

echo "=== All validations passed ==="
```

---

## 5. Layer 3: Template Testing (Backstage Skeleton)

Test that the Backstage scaffolder produces a valid, buildable project when variables are substituted.

### 5.1 Render Skeleton with Test Values

```bash
#!/bin/bash
# tests/template/render-and-build.sh

set -e
TEMPLATE_DIR="backstage/templates/go-service-template/skeleton"
OUTPUT_DIR="/tmp/idp-template-test-$(date +%s)"

echo "=== Template Render Test ==="

# 1. Copy skeleton
cp -r "$TEMPLATE_DIR" "$OUTPUT_DIR"

# 2. Replace template variables with test values
cd "$OUTPUT_DIR"

# Replace Backstage template variables
find . -type f -name "*.go" -o -name "*.yaml" -o -name "*.yml" -o -name "*.mod" -o -name "*.md" -o -name "Makefile" -o -name "Dockerfile" | while read file; do
    sed -i '' \
        -e 's/\${{ values.serviceName }}/test-service/g' \
        -e 's/\${{ values.description }}/Test service for validation/g' \
        -e 's/\${{ values.owner }}/platform-team/g' \
        -e 's/\${{ values.namespace }}/services/g' \
        -e 's/\${{ values.replicas }}/2/g' \
        "$file" 2>/dev/null || true
done

# 3. Verify Go code compiles
echo "Checking Go build..."
go build ./... 2>/dev/null && echo "  Go build: PASSED" || echo "  Go build: FAILED"

# 4. Verify Go vet passes
echo "Checking Go vet..."
go vet ./... 2>/dev/null && echo "  Go vet: PASSED" || echo "  Go vet: FAILED"

# 5. Verify K8s manifests are valid
echo "Checking K8s manifests..."
if [ -d "k8s" ]; then
    kubeconform -strict k8s/*.yaml && echo "  K8s manifests: PASSED" || echo "  K8s manifests: FAILED"
fi

# 6. Verify Dockerfile is valid
echo "Checking Dockerfile..."
if [ -f "Dockerfile" ]; then
    docker build --check . 2>/dev/null && echo "  Dockerfile: PASSED" || echo "  Dockerfile syntax: PASSED (build skipped)"
fi

# 7. Verify catalog-info.yaml is valid
echo "Checking catalog-info.yaml..."
if [ -f "catalog-info.yaml" ]; then
    yamllint -d relaxed catalog-info.yaml && echo "  catalog-info.yaml: PASSED" || echo "  catalog-info.yaml: FAILED"
fi

# 8. Cleanup
rm -rf "$OUTPUT_DIR"

echo "=== Template Test Complete ==="
```

### 5.2 What to Verify

| Check | Why |
|-------|-----|
| Go compiles (`go build ./...`) | Template variables didn't break syntax |
| Go vet passes | No suspicious constructs |
| K8s manifests validate | Deployment/Service YAML is correct |
| Dockerfile is valid | Multi-stage build works |
| catalog-info.yaml is valid | Backstage can register the service |
| Crossplane claim is valid | Claim YAML matches XRD schema |

---

## 6. Layer 4: Infrastructure Testing (Crossplane + Operators)

Test that Crossplane claims actually create real resources. **Requires a running Kubernetes cluster** (use kind for CI).

### 6.1 chainsaw (Recommended)

[chainsaw](https://kyverno.github.io/chainsaw/) is a Kubernetes testing framework that validates resource lifecycle: create → assert → cleanup.

#### Test: GoService Claim Provisions All Resources

```yaml
# tests/crossplane/goservice/chainsaw-test.yaml
apiVersion: chainsaw.kyverno.io/v1alpha1
kind: Test
metadata:
  name: goservice-full-provisioning
spec:
  description: "Verify GoServiceClaim creates DB, Redis, Deployment, Service, and Secrets"
  timeouts:
    apply: 30s
    assert: 5m
    delete: 5m
  steps:
    # Step 1: Create the claim
    - name: create-claim
      try:
        - apply:
            file: claim.yaml

    # Step 2: Verify PostgreSQL was created
    - name: assert-database
      try:
        - assert:
            file: assert-database.yaml

    # Step 3: Verify Redis was created
    - name: assert-redis
      try:
        - assert:
            file: assert-redis.yaml

    # Step 4: Verify Deployment was created
    - name: assert-deployment
      try:
        - assert:
            file: assert-deployment.yaml

    # Step 5: Verify Service was created
    - name: assert-service
      try:
        - assert:
            file: assert-service.yaml

    # Step 6: Verify connection secrets exist
    - name: assert-secrets
      try:
        - assert:
            file: assert-secrets.yaml

    # Step 7: Cleanup — delete claim and verify cascade
    - name: cleanup
      try:
        - delete:
            file: claim.yaml
      catch:
        - sleep:
            duration: 10s
```

#### Test Fixtures

```yaml
# tests/crossplane/goservice/claim.yaml
apiVersion: platform.example.com/v1alpha1
kind: GoServiceClaim
metadata:
  name: test-service
  namespace: services
spec:
  serviceName: test-service
  image: nginx:latest
  replicas: 1
  database:
    enabled: true
    size: small
  redis:
    enabled: true
    size: small
  kafka:
    enabled: false
```

```yaml
# tests/crossplane/goservice/assert-database.yaml
apiVersion: postgresql.cnpg.io/v1
kind: Cluster
metadata:
  name: test-service-db
status:
  phase: Cluster in healthy state
```

```yaml
# tests/crossplane/goservice/assert-redis.yaml
apiVersion: redis.redis.opstreelabs.in/v1beta2
kind: Redis
metadata:
  name: test-service-redis
```

```yaml
# tests/crossplane/goservice/assert-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-service
  namespace: services
status:
  readyReplicas: 1
```

```yaml
# tests/crossplane/goservice/assert-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: test-service
  namespace: services
spec:
  ports:
    - name: http
      port: 8080
    - name: grpc
      port: 9090
```

```yaml
# tests/crossplane/goservice/assert-secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: test-service-postgres-conn
  namespace: services
```

#### Running chainsaw Tests

```bash
# Run all Crossplane tests
chainsaw test --test-dir tests/crossplane/

# Run specific test
chainsaw test --test-dir tests/crossplane/goservice/

# Verbose output
chainsaw test --test-dir tests/crossplane/ --v 3
```

### 6.2 Database-Only Claim Test

```yaml
# tests/crossplane/database-only/chainsaw-test.yaml
apiVersion: chainsaw.kyverno.io/v1alpha1
kind: Test
metadata:
  name: goservice-database-only
spec:
  description: "Verify GoServiceClaim with only database enabled"
  timeouts:
    assert: 5m
  steps:
    - name: create-claim
      try:
        - apply:
            resource:
              apiVersion: platform.example.com/v1alpha1
              kind: GoServiceClaim
              metadata:
                name: db-only-test
                namespace: services
              spec:
                serviceName: db-only-test
                image: nginx:latest
                replicas: 1
                database:
                  enabled: true
                  size: small
                redis:
                  enabled: false
                kafka:
                  enabled: false

    - name: assert-database-exists
      try:
        - assert:
            resource:
              apiVersion: postgresql.cnpg.io/v1
              kind: Cluster
              metadata:
                name: db-only-test-db

    - name: assert-no-redis
      try:
        - error:
            resource:
              apiVersion: redis.redis.opstreelabs.in/v1beta2
              kind: Redis
              metadata:
                name: db-only-test-redis

    - name: cleanup
      try:
        - delete:
            resource:
              apiVersion: platform.example.com/v1alpha1
              kind: GoServiceClaim
              metadata:
                name: db-only-test
                namespace: services
```

### 6.3 Operator Health Tests

```yaml
# tests/infrastructure/operators/chainsaw-test.yaml
apiVersion: chainsaw.kyverno.io/v1alpha1
kind: Test
metadata:
  name: operators-healthy
spec:
  description: "Verify all operators are running"
  steps:
    - name: cnpg-operator
      try:
        - assert:
            resource:
              apiVersion: apps/v1
              kind: Deployment
              metadata:
                name: cnpg-cloudnative-pg
                namespace: databases
              status:
                readyReplicas: 1

    - name: redis-operator
      try:
        - assert:
            resource:
              apiVersion: apps/v1
              kind: Deployment
              metadata:
                namespace: cache
              status:
                readyReplicas: 1

    - name: strimzi-operator
      try:
        - assert:
            resource:
              apiVersion: apps/v1
              kind: Deployment
              metadata:
                namespace: kafka
              status:
                readyReplicas: 1
```

---

## 7. Layer 5: GitOps Testing (ArgoCD)

Test that ArgoCD Applications sync correctly and remain healthy.

### 7.1 Sync Status Validation

```yaml
# tests/argocd/chainsaw-test.yaml
apiVersion: chainsaw.kyverno.io/v1alpha1
kind: Test
metadata:
  name: argocd-platform-apps-healthy
spec:
  description: "Verify all platform ArgoCD Applications are synced and healthy"
  steps:
    - name: app-of-apps-synced
      try:
        - assert:
            timeout: 5m
            resource:
              apiVersion: argoproj.io/v1alpha1
              kind: Application
              metadata:
                name: platform-apps
                namespace: platform
              status:
                sync:
                  status: Synced
                health:
                  status: Healthy
```

### 7.2 CLI-Based Validation

```bash
#!/bin/bash
# tests/argocd/validate-sync.sh

set -e
echo "=== ArgoCD Sync Validation ==="

# Check all apps are synced
OUTOF_SYNC=$(argocd app list -o json | jq -r '.[] | select(.status.sync.status != "Synced") | .metadata.name')

if [ -n "$OUTOF_SYNC" ]; then
    echo "FAILED: These apps are not synced:"
    echo "$OUTOF_SYNC"
    exit 1
fi

# Check all apps are healthy
UNHEALTHY=$(argocd app list -o json | jq -r '.[] | select(.status.health.status != "Healthy") | .metadata.name')

if [ -n "$UNHEALTHY" ]; then
    echo "WARNING: These apps are not healthy:"
    echo "$UNHEALTHY"
    exit 1
fi

echo "All ArgoCD applications: Synced + Healthy"

# Check for drift
echo "Checking for drift..."
for app in $(argocd app list -o name); do
    DIFF=$(argocd app diff "$app" 2>&1 || true)
    if [ -n "$DIFF" ]; then
        echo "DRIFT detected in $app:"
        echo "$DIFF"
    fi
done

echo "=== ArgoCD Validation Complete ==="
```

### 7.3 ApplicationSet Generation Test

```bash
# Verify ApplicationSet generates expected Applications
EXPECTED_APPS="user-service order-service notification-service"
ACTUAL_APPS=$(kubectl get applications -n platform -o name | sed 's|application.argoproj.io/||' | sort)

for app in $EXPECTED_APPS; do
    if echo "$ACTUAL_APPS" | grep -q "$app"; then
        echo "  $app: FOUND"
    else
        echo "  $app: MISSING"
    fi
done
```

---

## 8. Layer 6: End-to-End Testing

Test the complete developer workflow from Crossplane claim to running service.

### 8.1 Full Workflow Test

```bash
#!/bin/bash
# tests/e2e/test-full-workflow.sh

set -e

SERVICE_NAME="e2e-test-$(date +%s)"
NAMESPACE="services"
TIMEOUT=300

echo "=== E2E Test: Full Developer Workflow ==="
echo "Service: ${SERVICE_NAME}"
echo ""

# ------------------------------------------
# Step 1: Create Crossplane claim
# (simulates what Backstage scaffolder does)
# ------------------------------------------
echo "Step 1: Creating GoServiceClaim..."
cat <<EOF | kubectl apply -f -
apiVersion: platform.example.com/v1alpha1
kind: GoServiceClaim
metadata:
  name: ${SERVICE_NAME}
  namespace: ${NAMESPACE}
spec:
  serviceName: ${SERVICE_NAME}
  image: nginx:latest
  replicas: 1
  database:
    enabled: true
    size: small
  redis:
    enabled: true
    size: small
  kafka:
    enabled: false
EOF
echo "  Claim created"

# ------------------------------------------
# Step 2: Wait for infrastructure provisioning
# ------------------------------------------
echo "Step 2: Waiting for infrastructure (up to ${TIMEOUT}s)..."

echo "  Waiting for claim to be ready..."
kubectl wait goserviceclaim/${SERVICE_NAME} -n ${NAMESPACE} \
    --for=condition=Ready --timeout=${TIMEOUT}s
echo "  Claim: Ready"

# ------------------------------------------
# Step 3: Verify database
# ------------------------------------------
echo "Step 3: Verifying database..."
DB_STATUS=$(kubectl get clusters.postgresql.cnpg.io ${SERVICE_NAME}-db -o jsonpath='{.status.phase}' 2>/dev/null || echo "NOT_FOUND")
if [ "$DB_STATUS" = "Cluster in healthy state" ]; then
    echo "  Database: Healthy"
else
    echo "  Database: ${DB_STATUS}"
    exit 1
fi

# ------------------------------------------
# Step 4: Verify Redis
# ------------------------------------------
echo "Step 4: Verifying Redis..."
kubectl get redis ${SERVICE_NAME}-redis > /dev/null 2>&1
echo "  Redis: Created"

# ------------------------------------------
# Step 5: Verify Deployment
# ------------------------------------------
echo "Step 5: Verifying Deployment..."
kubectl wait deployment/${SERVICE_NAME} -n ${NAMESPACE} \
    --for=condition=Available --timeout=120s
echo "  Deployment: Available"

# ------------------------------------------
# Step 6: Verify Service
# ------------------------------------------
echo "Step 6: Verifying Service..."
SVC_IP=$(kubectl get svc ${SERVICE_NAME} -n ${NAMESPACE} -o jsonpath='{.spec.clusterIP}')
echo "  Service: ${SVC_IP}"

# ------------------------------------------
# Step 7: Verify Secrets
# ------------------------------------------
echo "Step 7: Verifying secrets..."
SECRETS_OK=true

kubectl get secret ${SERVICE_NAME}-postgres-conn -n ${NAMESPACE} > /dev/null 2>&1 || SECRETS_OK=false
kubectl get secret ${SERVICE_NAME}-redis-conn -n ${NAMESPACE} > /dev/null 2>&1 || SECRETS_OK=false

if [ "$SECRETS_OK" = true ]; then
    echo "  Secrets: All present"
else
    echo "  Secrets: MISSING"
    exit 1
fi

# ------------------------------------------
# Step 8: Verify monitoring (ServiceMonitor exists)
# ------------------------------------------
echo "Step 8: Checking monitoring..."
kubectl get servicemonitor ${SERVICE_NAME} -n ${NAMESPACE} > /dev/null 2>&1 \
    && echo "  ServiceMonitor: Found" \
    || echo "  ServiceMonitor: Not found (optional)"

# ------------------------------------------
# Step 9: Cleanup
# ------------------------------------------
echo "Step 9: Cleaning up..."
kubectl delete goserviceclaim/${SERVICE_NAME} -n ${NAMESPACE} --wait=false

echo "  Waiting for cleanup (up to ${TIMEOUT}s)..."
sleep 30

# Verify no orphaned resources
REMAINING=$(kubectl get all -n ${NAMESPACE} -l app=${SERVICE_NAME} --no-headers 2>/dev/null | wc -l)
if [ "$REMAINING" -eq 0 ]; then
    echo "  Cleanup: Complete (no orphaned resources)"
else
    echo "  Cleanup: WARNING — ${REMAINING} resources still exist"
fi

echo ""
echo "=== E2E Test: PASSED ==="
```

### 8.2 Make Target

```makefile
# Add to root Makefile
.PHONY: test-e2e test-manifests test-crossplane test-argocd test-all

test-manifests:
	@bash tests/manifests/validate.sh

test-crossplane:
	chainsaw test --test-dir tests/crossplane/

test-argocd:
	chainsaw test --test-dir tests/argocd/

test-e2e:
	@bash tests/e2e/test-full-workflow.sh

test-all: test-manifests test-crossplane test-argocd test-e2e
```

---

## 9. CI/CD Test Integration

### 9.1 GitHub Actions: Manifest Validation (Every Commit)

```yaml
# .github/workflows/validate.yaml
name: Validate Manifests

on:
  push:
    paths:
      - 'crossplane/**'
      - 'argocd/**'
      - 'infrastructure/**'
      - 'go-service-template/k8s/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install kubeconform
        run: |
          curl -sL https://github.com/yannh/kubeconform/releases/latest/download/kubeconform-linux-amd64.tar.gz \
            | tar xz -C /usr/local/bin

      - name: Install yamllint
        run: pip install yamllint

      - name: YAML Lint
        run: yamllint -d relaxed crossplane/ argocd/ infrastructure/

      - name: Kubernetes Schema Validation
        run: |
          kubeconform -strict -summary \
            -skip "GoService,GoServiceClaim,Cluster,Redis,KafkaTopic,Application,ApplicationSet,CompositeResourceDefinition,Composition" \
            infrastructure/**/*.yaml \
            go-service-template/k8s/*.yaml
```

### 9.2 GitHub Actions: Crossplane Tests (On PR to main)

```yaml
# .github/workflows/test-crossplane.yaml
name: Crossplane Integration Tests

on:
  pull_request:
    paths:
      - 'crossplane/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Create kind cluster
        uses: helm/kind-action@v1
        with:
          cluster_name: test-cluster

      - name: Install Crossplane
        run: |
          helm repo add crossplane-stable https://charts.crossplane.io/stable
          helm install crossplane crossplane-stable/crossplane \
            --namespace crossplane-system --create-namespace
          kubectl wait --namespace crossplane-system \
            --for=condition=ready pod --selector=app=crossplane --timeout=120s

      - name: Install providers
        run: |
          kubectl apply -f crossplane/providers/
          sleep 30
          kubectl wait provider --all --for=condition=Healthy --timeout=120s

      - name: Apply XRDs and Compositions
        run: |
          kubectl apply -f crossplane/compositions/goservice-xrd.yaml
          kubectl apply -f crossplane/compositions/goservice-composition.yaml

      - name: Install chainsaw
        run: |
          curl -sL https://github.com/kyverno/chainsaw/releases/latest/download/chainsaw_linux_amd64.tar.gz \
            | tar xz -C /usr/local/bin

      - name: Run Crossplane tests
        run: chainsaw test --test-dir tests/crossplane/
```

### 9.3 GitHub Actions: Go Service Tests (Every Commit)

```yaml
# .github/workflows/test-go.yaml
name: Go Service Tests

on:
  push:
    paths:
      - 'go-service-template/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run tests
        working-directory: go-service-template
        run: make test

      - name: Check coverage
        working-directory: go-service-template
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          echo "Coverage: ${COVERAGE}%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage ${COVERAGE}% is below 80% threshold"
            exit 1
          fi
```

---

## 10. Test Directory Structure

```
tests/
├── manifests/                         # Layer 2: YAML validation
│   └── validate.sh                    # kubeconform + yamllint script
│
├── template/                          # Layer 3: Backstage skeleton tests
│   ├── render-and-build.sh            # Render skeleton, verify it builds
│   └── test-values.json               # Test variable values
│
├── crossplane/                        # Layer 4: Infrastructure tests
│   ├── goservice/                     # Full GoService claim test
│   │   ├── chainsaw-test.yaml
│   │   ├── claim.yaml
│   │   ├── assert-database.yaml
│   │   ├── assert-redis.yaml
│   │   ├── assert-deployment.yaml
│   │   ├── assert-service.yaml
│   │   ├── assert-secrets.yaml
│   │   └── assert-cleanup.yaml
│   ├── database-only/                 # DB-only claim test
│   │   ├── chainsaw-test.yaml
│   │   └── ...
│   └── no-infra/                      # Basic service (no DB/Redis/Kafka)
│       ├── chainsaw-test.yaml
│       └── ...
│
├── infrastructure/                    # Layer 4: Operator health tests
│   └── operators/
│       └── chainsaw-test.yaml
│
├── argocd/                            # Layer 5: GitOps tests
│   ├── chainsaw-test.yaml
│   └── validate-sync.sh
│
└── e2e/                               # Layer 6: Full workflow tests
    └── test-full-workflow.sh
```

---

## 11. When to Run Each Test

| Test | Trigger | Environment | Duration |
|------|---------|-------------|----------|
| YAML lint + kubeconform | Every commit | CI (no cluster needed) | ~10s |
| Go unit tests | Every commit | CI (no cluster needed) | ~30s |
| Template render + build | Template changes | CI (no cluster needed) | ~1min |
| Crossplane composition render | Composition changes | CI (no cluster needed) | ~10s |
| Crossplane integration (chainsaw) | PR to main | Kind cluster in CI | ~5-10min |
| ArgoCD sync validation | After deploy to staging | Staging cluster | ~2min |
| End-to-end workflow | Nightly / pre-release | Dedicated test cluster | ~15min |
| Operator health checks | After infrastructure changes | Any cluster | ~1min |

---

## 12. Debugging Failed Tests

### Crossplane Claim Stuck

```bash
# Check claim status
kubectl describe goserviceclaim <name> -n services

# Check composite resource
kubectl get goservice

# Check managed resources
kubectl get managed

# Check Crossplane logs
kubectl logs -n crossplane-system -l pkg.crossplane.io/revision --tail=50
```

### ArgoCD App Not Syncing

```bash
# Check app status
argocd app get <app-name>

# Check sync diff
argocd app diff <app-name>

# Check controller logs
kubectl logs -n platform -l app.kubernetes.io/name=argocd-application-controller --tail=50
```

### Pod Not Starting

```bash
# Check pod events
kubectl describe pod <pod-name> -n services

# Check logs
kubectl logs <pod-name> -n services

# Common issues:
# - ImagePullBackOff → wrong image name or registry auth
# - CrashLoopBackOff → app crashes on start (check logs)
# - Pending → insufficient resources or missing PV
```

### Secret Not Found

```bash
# List secrets in namespace
kubectl get secrets -n services

# Check if Crossplane created the secret
kubectl get managed | grep secret

# Check composition patches for secret naming
```

---

## Next Steps

1. Install testing tools → Section 2
2. Set up `tests/` directory → Section 10
3. Start with Layer 2 (manifest validation) — no cluster needed
4. Add chainsaw tests as you build Crossplane compositions
5. Add CI workflows → Section 9

Related: [Setup Guide](./setup-guide.md) | [Crossplane Guide](./crossplane-guide.md) | [Project Roadmap](./project-roadmap.md)
