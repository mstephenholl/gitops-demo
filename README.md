# gitops-demo

A demonstration of deploying a containerized Go application to Kubernetes using GitOps with FluxCD.

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  GitHub Repository (single source of truth)             │
│                                                         │
│  ├── cmd/server/         Go HTTP service                │
│  ├── internal/           Handlers, middleware, version  │
│  ├── k8s/                K8s manifests (Kustomize)      │
│  ├── clusters/local/     FluxCD Kustomization           │
│  ├── .github/workflows/  CI/CD pipeline                 │
│  ├── .env.example        Environment variable template  │
│  └── Dockerfile          Multi-stage container build    │
└──────────────┬──────────────────────────────────────────┘
               │
               │  FluxCD watches (1m interval)
               ▼
┌─────────────────────────────────────────────────────────┐
│  k3d Cluster (local Kubernetes)                         │
│                                                         │
│  ┌─────────────────┐    ┌──────────────────────┐        │
│  │  flux-system ns │    │  gitops-demo ns      │        │
│  │                 │    │                      │        │
│  │  source-ctrl    │───▶│  Deployment (2 pods) │        │
│  │  kustomize-ctrl │    │  Service (ClusterIP) │        │
│  └─────────────────┘    └──────────────────────┘        │
│                                                         │
│  localhost:8080 ──▶ traefik LB ──▶ service ──▶ pods     │
└─────────────────────────────────────────────────────────┘
```

## Prerequisites

| Tool       | Install                                                                 |
|------------|-------------------------------------------------------------------------|
| Docker     | https://docs.docker.com/get-docker/                                     |
| k3d        | `curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh \| bash` |
| kubectl    | `curl -LO https://dl.k8s.io/release/$(curl -sL https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl` |
| Flux CLI   | `curl -s https://fluxcd.io/install.sh \| bash`                          |
| Go 1.25+   | https://go.dev/dl/                                                      |
| golangci-lint | `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest` |

You also need a **GitHub Personal Access Token** with `repo` scope.

## Quick Start

### 1. Clone and update the module path

```bash
git clone https://github.com/mstephenholl/gitops-demo.git
cd gitops-demo

# Replace mstephenholl with your GitHub username in all files:
find . -type f \( -name '*.go' -o -name 'go.mod' -o -name 'Makefile' -o -name 'Dockerfile' -o -name '.golangci.yml' \) \
  -exec sed -i '' 's|github.com/mstephenholl/gitops-demo|github.com/YOUR_USERNAME/gitops-demo|g' {} +
```

### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` and fill in the required values:

```dotenv
# Required for k3d + FluxCD setup
GITHUB_TOKEN=ghp_your_token_here
GITHUB_USER=your_username
```

> **Note:** `.env` is git-ignored and never copied into container images. All
> variables defined in `.env` are loaded by the Go server (via
> [godotenv](https://github.com/joho/godotenv)), the Makefile, and the helper
> scripts automatically. Environment variables set in your shell still take
> precedence over `.env` values.

### 3. Lint and test locally

```bash
make lint
make test-cover
```

### 4. Set up the k3d cluster and FluxCD

```bash
make k3d-setup
```

### 5. Watch Flux reconcile

```bash
# In one terminal: watch Flux
flux get kustomizations -w

# In another: watch pods come up
kubectl -n gitops-demo get pods -w
```

### 6. Hit the service

```bash
curl http://localhost:8080/info
```

Expected output:

```json
{
  "tag": "local-dev",
  "commit": "abc1234",
  "build_time": "2026-02-26T12:00:00Z",
  "go_version": "go1.25"
}
```

## GitOps Demo: Make a Change

To demonstrate the GitOps reconciliation loop:

```bash
# 1. Edit k8s/deployment.yaml — for example, change replicas from 2 to 3
sed -i 's/replicas: 2/replicas: 3/' k8s/deployment.yaml

# 2. Commit and push
git add k8s/deployment.yaml
git commit -m "scale to 3 replicas"
git push

# 3. Watch Flux detect and apply the change (within ~1 minute)
flux get kustomizations -w
kubectl -n gitops-demo get pods -w
```

## Endpoints

| Path      | Method | Description                     |
|-----------|--------|---------------------------------|
| `/healthz`| GET    | Liveness probe — returns `ok`   |
| `/readyz` | GET    | Readiness probe — returns `ready`|
| `/info`   | GET    | Build metadata (tag, commit, time, Go version) |

## Project Layout

```
.
├── cmd/server/          # Application entry point
├── internal/
│   ├── handlers/        # HTTP handlers and middleware
│   └── version/         # Build metadata (injected via ldflags)
├── k8s/                 # Kubernetes manifests (Kustomize)
├── clusters/local/      # FluxCD Kustomization for local cluster
├── scripts/             # Setup and teardown helpers
├── .github/workflows/   # CI/CD pipeline
├── .env.example         # Environment variable template (copy to .env)
├── .golangci.yml        # Linter and formatter configuration
├── Dockerfile           # Multi-stage container build
├── Makefile             # Build, test, lint, and cluster targets
└── go.mod
```

## Teardown

```bash
make k3d-teardown
```
