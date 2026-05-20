#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-orbit-test}"
RELEASE="${RELEASE:-orbit}"
FORCE="${FORCE:-false}"

if helm status "$RELEASE" -n "$NAMESPACE" >/dev/null 2>&1; then
  helm uninstall "$RELEASE" -n "$NAMESPACE" || true
fi

if [[ "$FORCE" == "true" ]]; then
  kubectl -n "$NAMESPACE" delete pvc orbit-postgres-data --ignore-not-found || true
else
  echo "PVC retained. Remove explicitly with: kubectl -n $NAMESPACE delete pvc orbit-postgres-data"
fi
