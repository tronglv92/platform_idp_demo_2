# Platform Setup Guide — From Zero to Running IDP

This guide walks you through setting up the entire Internal Developer Platform from scratch. By the end, you'll have a working Kubernetes cluster with Crossplane, ArgoCD, monitoring stack, and Vault — ready for Phase 2 (data infrastructure).

**Prerequisites knowledge**: Basic Kubernetes (`kubectl`), Helm, YAML. No prior IDP experience needed.

---

## 1. Prerequisites

### 1.1 Required Tools

Install these on your local machine:

| Tool | Purpose | Install |
|------|---------|---------|
| `kubectl` | Kubernetes CLI | `brew install kubectl` |
| `helm` | Package manager for K8s | `brew install helm` |
| `kind` | Local K8s cluster | `brew install kind` |
| `argocd` | ArgoCD CLI | `brew install argocd` |
| `crossplane` CLI | Crossplane helper | `curl -sL https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh \| sh` |
| `docker` | Container runtime | [Docker Desktop](https://www.docker.com/products/docker-desktop/) |
| `go` (1.21+) | Go compiler | `brew install go` |
| `make` | Build automation | Pre-installed on macOS |

### 1.2 Verify Installation

```bash
kubectl version --client
helm version
kind version
argocd version --client
docker version
go version
make --version
```

---

## 2. Create Kubernetes Cluster

### 2.1 Kind Cluster (Local Development)

Create a multi-node cluster for realistic testing:

```yaml
# kind-config.yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: idp-platform
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
  - role: worker
  - role: worker
```

```bash
# Create the cluster
kind create cluster --config kind-config.yaml

# Verify
kubectl cluster-info
kubectl get nodes
```

Expected output: 3 nodes (1 control-plane, 2 workers) in `Ready` state.

### 2.2 Create Namespaces

```bash
kubectl create namespace platform      # Backstage, ArgoCD, Crossplane
kubectl create namespace services      # Deployed microservices
kubectl create namespace monitoring    # Prometheus, Grafana, Loki, Jaeger
kubectl create namespace kafka         # Strimzi Kafka operator + clusters
kubectl create namespace databases     # CloudNativePG PostgreSQL instances
kubectl create namespace cache         # Redis instances
```

### 2.3 Install Ingress Controller

```bash
# NGINX ingress for kind
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# Ensure the controller runs on control-plane node (where ports 80/443 are mapped)
kubectl patch deployment ingress-nginx-controller -n ingress-nginx --type=json -p='[
  {"op": "add", "path": "/spec/template/spec/nodeSelector", "value": {"ingress-ready": "true"}},
  {"op": "add", "path": "/spec/template/spec/tolerations", "value": [{"key": "node-role.kubernetes.io/control-plane", "operator": "Exists", "effect": "NoSchedule"}]}
]'

# Wait for it to be ready
kubectl rollout status deployment ingress-nginx-controller -n ingress-nginx --timeout=120s
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s

# Verify controller is on control-plane node
kubectl get pods -n ingress-nginx -o wide
```

> **Important**: The ingress controller pod **must** run on the `control-plane` node. Kind only maps host ports 80/443 to the control-plane node. If the pod schedules on a worker node, ingress traffic will not reach it.

---

## 3. Install Crossplane

Crossplane turns your Kubernetes cluster into a universal control plane for infrastructure.

### 3.1 Install Crossplane Operator

```bash
# Add Helm repo
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

# Install Crossplane
helm install crossplane \
  crossplane-stable/crossplane \
  --namespace crossplane-system \
  --create-namespace \
  --set args='{"--enable-usages"}'

# Wait for pods
kubectl wait --namespace crossplane-system \
  --for=condition=ready pod \
  --selector=app=crossplane \
  --timeout=120s
```

### 3.2 Install Providers

```bash
# Kubernetes provider (manages K8s resources)
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-kubernetes
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.11.0
EOF

# Helm provider (manages Helm releases)
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-helm
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-helm:v0.16.0
EOF

# Wait for providers to be healthy
kubectl wait provider provider-kubernetes --for=condition=Healthy --timeout=120s
kubectl wait provider provider-helm --for=condition=Healthy --timeout=120s
```

### 3.3 Configure Provider Credentials

```bash
# Create ProviderConfig for Kubernetes provider
cat <<EOF | kubectl apply -f -
apiVersion: kubernetes.crossplane.io/v1alpha1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: InjectedIdentity
EOF

# Create ProviderConfig for Helm provider
cat <<EOF | kubectl apply -f -
apiVersion: helm.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: InjectedIdentity
EOF
```

### 3.4 Verify

```bash
kubectl get providers
kubectl get providerconfigs
```

All providers should show `INSTALLED=True`, `HEALTHY=True`.

---

## 4. Install ArgoCD

ArgoCD provides GitOps-based continuous deployment.

### 4.1 Install ArgoCD

```bash
# Add Helm repo
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

# Install ArgoCD with custom values (ingress at argocd.test.com)
helm install argocd argo/argo-cd \
  --namespace platform \
  -f argocd/values-argocd.yaml

# Wait for pods
kubectl wait --namespace platform \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/name=argocd-server \
  --timeout=180s
```

### 4.2 Create Self-Signed TLS Certificate

Generate a self-signed certificate with SANs and create a Kubernetes secret for the ingress:

```bash
# Generate self-signed cert (must include SANs, not just CN)
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /tmp/argocd-tls.key -out /tmp/argocd-tls.crt \
  -subj "/CN=argocd.test.com" \
  -addext "subjectAltName=DNS:argocd.test.com"

# Create TLS secret in platform namespace
kubectl create secret tls argocd-server-tls \
  --cert=/tmp/argocd-tls.crt --key=/tmp/argocd-tls.key \
  -n platform
```

> **Note**: The `-addext "subjectAltName=..."` is required. NGINX ingress rejects certificates that only use the legacy Common Name field without SANs.

### 4.3 Add DNS Entry

Add the following to `/etc/hosts` (for local development with kind):

```bash
echo "127.0.0.1 argocd.test.com" | sudo tee -a /etc/hosts
```

### 4.4 Access ArgoCD UI

```bash
# Get initial admin password
kubectl -n platform get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d; echo
```

Open: `https://argocd.test.com` — Login: `admin` / (password from above)

> **Note**: The browser will show a certificate warning because of the self-signed cert. Click "Advanced" → "Accept the Risk and Continue" (Firefox) or "Proceed to argocd.test.com" (Chrome).

### 4.5 Configure Git Repository

```bash
# Login via CLI
argocd login argocd.test.com --username admin --password <password> --insecure

# Add your Git repository
argocd repo add https://github.com/<your-org>/platform-idp.git \
  --username <github-user> \
  --password <github-token>
```

### 4.6 Create App-of-Apps

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
    repoURL: https://github.com/tronglv92/platform_idp_demo_2
    targetRevision: main
    path: argocd/platform
  destination:
    server: https://kubernetes.default.svc
    namespace: platform
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

---

## 5. Install Monitoring Stack

### 5.1 Prometheus + Grafana (kube-prometheus-stack)

```bash
# Add Helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install
helm install monitoring prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --set grafana.adminPassword=admin \
  --set grafana.service.type=ClusterIP \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
```

### 5.2 Loki (Log Aggregation)

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install loki grafana/loki-stack \
  --namespace monitoring \
  --set grafana.enabled=false \
  --set promtail.enabled=true
```

### 5.3 Jaeger (Distributed Tracing)

```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm repo update

helm install jaeger jaegertracing/jaeger \
  --namespace monitoring \
  --set provisionDataStore.cassandra=false \
  --set allInOne.enabled=true \
  --set storage.type=memory \
  --set agent.enabled=false \
  --set collector.enabled=false \
  --set query.enabled=false
```

### 5.4 Access Dashboards

```bash
# Grafana (admin/admin)
kubectl port-forward svc/monitoring-grafana -n monitoring 3000:80

# Prometheus
kubectl port-forward svc/monitoring-kube-prometheus-prometheus -n monitoring 9090:9090

# Jaeger
kubectl port-forward svc/jaeger-query -n monitoring 16686:16686
```

---

## 6. Install Vault (Secrets Management)

### 6.1 Install Vault

```bash
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update

helm install vault hashicorp/vault \
  --namespace platform \
  --set server.dev.enabled=true \
  --set injector.enabled=true
```

> **Note**: `dev.enabled=true` runs Vault in dev mode (not for production). For production, configure HA with Raft storage.

### 6.2 Configure Kubernetes Auth

```bash
# Exec into Vault pod
kubectl exec -it vault-0 -n platform -- /bin/sh

# Inside the pod:
vault auth enable kubernetes

vault write auth/kubernetes/config \
  kubernetes_host="https://kubernetes.default.svc:443"

vault write auth/kubernetes/role/service-role \
  bound_service_account_names=default \
  bound_service_account_namespaces=services \
  policies=service-policy \
  ttl=24h

# Create a policy
vault policy write service-policy - <<EOF
path "secret/data/services/*" {
  capabilities = ["read"]
}
EOF

exit
```

---

## 7. Install Data Infrastructure Operators

These operators will be used by Crossplane compositions to provision databases, caches, and messaging.

### 7.1 CloudNativePG (PostgreSQL)

```bash
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm repo update

helm install cnpg cnpg/cloudnative-pg \
  --namespace databases \
  --set monitoring.podMonitorEnabled=true
```

### 7.2 Redis Operator

```bash
helm repo add ot-helm https://ot-container-kit.github.io/helm-charts
helm repo update

helm install redis-operator ot-helm/redis-operator \
  --namespace cache
```

### 7.3 Strimzi (Kafka)

```bash
helm repo add strimzi https://strimzi.io/charts
helm repo update

helm install strimzi strimzi/strimzi-kafka-operator \
  --namespace kafka
```

### 7.4 Verify Operators

```bash
kubectl get pods -n databases    # cnpg controller running
kubectl get pods -n cache        # redis-operator running
kubectl get pods -n kafka        # strimzi-cluster-operator running
```

---

## 8. Verification Checklist

Run this to verify everything is healthy:

```bash
echo "=== Nodes ==="
kubectl get nodes

echo "=== Namespaces ==="
kubectl get namespaces

echo "=== Crossplane ==="
kubectl get providers
kubectl get pods -n crossplane-system

echo "=== ArgoCD ==="
kubectl get pods -n platform -l app.kubernetes.io/name=argocd-server

echo "=== Monitoring ==="
kubectl get pods -n monitoring

echo "=== Vault ==="
kubectl get pods -n platform -l app.kubernetes.io/name=vault

echo "=== Operators ==="
kubectl get pods -n databases
kubectl get pods -n cache
kubectl get pods -n kafka
```

Expected: All pods in `Running` state, all providers `HEALTHY`.

---

## 9. Resource Requirements

### Local (kind) Minimum

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 4 cores | 8 cores |
| RAM | 8 GB | 16 GB |
| Disk | 20 GB | 50 GB |
| Docker | 6 GB memory | 10 GB memory |

### Production (Cloud)

| Component | Nodes | CPU/Node | RAM/Node |
|-----------|-------|----------|----------|
| Control plane | 3 | 2 vCPU | 4 GB |
| Workers | 3-5 | 4 vCPU | 16 GB |
| Monitoring | 1-2 | 4 vCPU | 16 GB |

---

## 10. Troubleshooting

### Pods stuck in Pending
```bash
kubectl describe pod <pod-name> -n <namespace>
# Check: insufficient resources, missing PV, image pull errors
```

### Crossplane provider not healthy
```bash
kubectl describe provider <provider-name>
kubectl logs -n crossplane-system -l pkg.crossplane.io/revision
```

### ArgoCD sync failed
```bash
argocd app get <app-name>
argocd app sync <app-name> --force
```

### Kind cluster issues
```bash
# Delete and recreate
kind delete cluster --name idp-platform
kind create cluster --config kind-config.yaml
```

---

## Next Steps

After completing this setup:
1. **Phase 2**: Define Crossplane XRDs and Compositions → see [Crossplane Guide](./crossplane-guide.md)
2. **Phase 3**: Parameterize go-service-template → see [Go Service Template Guide](./go-service-template-guide.md)
3. **Phase 4**: Deploy Backstage → see [Backstage Guide](./backstage-guide.md)

Full roadmap: [Project Roadmap](./project-roadmap.md)
