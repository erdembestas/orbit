#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-orbit-test}"

kubectl -n "$NAMESPACE" port-forward svc/orbit-ui 8080:8080
