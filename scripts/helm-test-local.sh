#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-orbit-test}"
RELEASE="${RELEASE:-orbit}"

helm test "$RELEASE" -n "$NAMESPACE"
