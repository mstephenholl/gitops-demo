#!/usr/bin/env bash
#
# setup.sh — Create a k3d cluster and prepare it for FluxCD GitOps.
#
# Usage:
#   ./scripts/setup.sh
#
# Prerequisites:
#   - docker, k3d, kubectl, flux CLI installed
#   - GITHUB_TOKEN env var set (PAT with 'repo' scope)
#   - GITHUB_USER env var set (your GitHub username)
#
set -euo pipefail

# Load .env file if present (supports standalone execution outside Make).
ENV_FILE="$(cd "$(dirname "$0")/.." && pwd)/.env"
if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "$ENV_FILE"
  set +a
fi

CLUSTER_NAME="${CLUSTER_NAME:-gitops-demo}"
REPO_NAME="${REPO_NAME:-gitops-demo}"
BRANCH="${BRANCH:-main}"
IMAGE_NAME="gitops-demo"
IMAGE_TAG="latest"

# ---- Colors ----
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[INFO]${NC}  $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC}  $*"; }
die()  { echo -e "${RED}[ERROR]${NC} $*" >&2; exit 1; }

# ---- Preflight checks ----
for cmd in docker k3d kubectl flux; do
  command -v "$cmd" >/dev/null 2>&1 || die "$cmd is not installed"
done

[[ -n "${GITHUB_TOKEN:-}" ]] || die "GITHUB_TOKEN is not set. Export a PAT with 'repo' scope."
[[ -n "${GITHUB_USER:-}"  ]] || die "GITHUB_USER is not set. Export your GitHub username."

# ---- Step 1: Create k3d cluster ----
if k3d cluster list | grep -q "$CLUSTER_NAME"; then
  warn "Cluster '$CLUSTER_NAME' already exists. Deleting and recreating."
  k3d cluster delete "$CLUSTER_NAME"
fi

log "Creating k3d cluster '$CLUSTER_NAME' with port mapping 8080->80..."
k3d cluster create "$CLUSTER_NAME" \
  --port "8080:80@loadbalancer" \
  --agents 2 \
  --wait

log "Cluster ready. Setting kubectl context..."
kubectl config use-context "k3d-${CLUSTER_NAME}"
kubectl cluster-info

# ---- Step 2: Build and import container image ----
log "Building Docker image '${IMAGE_NAME}:${IMAGE_TAG}'..."
docker build \
  --build-arg VERSION=local-dev \
  --build-arg COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)" \
  --build-arg BUILD_TIME="$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -t "${IMAGE_NAME}:${IMAGE_TAG}" .

log "Importing image into k3d cluster..."
k3d image import "${IMAGE_NAME}:${IMAGE_TAG}" -c "$CLUSTER_NAME"

# ---- Step 3: Bootstrap FluxCD ----
log "Running Flux pre-flight check..."
flux check --pre

log "Bootstrapping FluxCD..."
flux bootstrap github \
  --owner="$GITHUB_USER" \
  --repository="$REPO_NAME" \
  --branch="$BRANCH" \
  --path="clusters/local" \
  --personal \
  --token-auth

log "Flux bootstrap complete. Checking status..."
flux check

# ---- Done ----
echo ""
log "================================================================"
log "  Setup complete!"
log ""
log "  Flux is now watching: clusters/local/"
log "  App manifests:        k8s/"
log ""
log "  Next steps:"
log "    1. Commit & push clusters/local/apps.yaml"
log "    2. Watch Flux reconcile:  flux get kustomizations -w"
log "    3. Check pods:            kubectl -n gitops-demo get pods"
log "    4. Hit the app:           curl http://localhost:8080/info"
log "================================================================"
