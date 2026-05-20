#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

trap 'rm -f orbit-api orbit-controller' EXIT

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o orbit-api ./cmd/orbit-api
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o orbit-controller ./cmd/orbit-controller
eval "$(minikube docker-env)"
docker build -t orbit-api:local .
docker build -t orbit-ui:local ./ui
docker images | grep orbit-api
docker images | grep orbit-ui
