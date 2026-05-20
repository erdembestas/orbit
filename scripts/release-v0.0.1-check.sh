#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

./scripts/check-secrets.sh
./scripts/check-local.sh

helm template orbit ./charts/orbit -n orbit-test -f charts/orbit/values-local.yaml > /tmp/orbit-rendered.yaml
kubectl apply --dry-run=client -f /tmp/orbit-rendered.yaml
rm -f /tmp/orbit-rendered.yaml
