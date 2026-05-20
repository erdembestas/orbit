#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-orbit-test}"
RELEASE="${RELEASE:-orbit}"

helm uninstall "$RELEASE" -n "$NAMESPACE"
echo "If the PVC was retained, remove it with: kubectl -n $NAMESPACE delete pvc orbit-postgres-data"
