#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

trap 'rm -f orbit-api orbit-controller; rm -rf ui/dist' EXIT

gofmt -w $(find ./cmd ./internal -name '*.go' -type f | sort)
go test ./...
go build ./cmd/orbit-api
go build ./cmd/orbit-controller
helm lint ./charts/orbit

if [[ -f ui/package.json ]]; then
  if [[ ! -d ui/node_modules ]]; then
    echo "Run ./scripts/ui-install.sh first"
    exit 1
  fi
  (
    cd ui
    npm run build
  )
fi
