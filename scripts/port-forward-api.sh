#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-orbit-test}"

kubectl -n "$NAMESPACE" port-forward svc/orbit-api 8081:8080
