#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

NAMESPACE="${NAMESPACE:-orbit-test}"
RELEASE="${RELEASE:-orbit}"

command -v helm >/dev/null
command -v kubectl >/dev/null

if [[ "$(kubectl config current-context)" != "minikube" ]]; then
  echo "kubectl context must be minikube" >&2
  exit 1
fi

helm upgrade --install "$RELEASE" ./charts/orbit \
  -n "$NAMESPACE" \
  --create-namespace \
  -f charts/orbit/values-local.yaml

kubectl -n "$NAMESPACE" rollout restart deployment/orbit-api
kubectl -n "$NAMESPACE" rollout restart deployment/orbit-controller
kubectl -n "$NAMESPACE" rollout restart deployment/orbit-ui

kubectl -n "$NAMESPACE" rollout status deployment/orbit-postgres --timeout=240s
kubectl -n "$NAMESPACE" rollout status deployment/orbit-api --timeout=240s
kubectl -n "$NAMESPACE" rollout status deployment/orbit-controller --timeout=240s
kubectl -n "$NAMESPACE" rollout status deployment/orbit-ui --timeout=240s
