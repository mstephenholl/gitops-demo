#!/usr/bin/env bash
#
# teardown.sh — Remove the k3d cluster and clean up local resources.
#
# Usage:
#   ./scripts/teardown.sh
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
IMAGE_NAME="gitops-demo"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

log() { echo -e "${GREEN}[INFO]${NC}  $*"; }

log "Deleting k3d cluster '$CLUSTER_NAME'..."
k3d cluster delete "$CLUSTER_NAME" 2>/dev/null || true

log "Removing local Docker image '${IMAGE_NAME}:latest'..."
docker rmi "${IMAGE_NAME}:latest" 2>/dev/null || true

log "Cleanup complete."
