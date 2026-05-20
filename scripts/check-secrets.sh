#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

HIGH_RISK_PATTERN='sk-[A-Za-z0-9_-]+|github_pat_[A-Za-z0-9_]+|ghp_[A-Za-z0-9]+|glpat-[A-Za-z0-9_-]+|xox[baprs]-[A-Za-z0-9-]+|AKIA[0-9A-Z]{16}|ASIA[0-9A-Z]{16}|BEGIN RSA PRIVATE KEY|BEGIN OPENSSH PRIVATE KEY|BEGIN PRIVATE KEY'
INFO_PATTERN='CHANGE_ME|orbit-local-password|orbit-local-dev-jwt-secret|password: admin|username: admin'

echo "== Secret scan =="

if [[ -f orbit-api || -f orbit-controller ]]; then
  echo "WARN generated binaries present in repo root"
fi

HIGH_RISK_MATCHES="$(grep -RInE "$HIGH_RISK_PATTERN" . \
  --exclude-dir=.git \
  --exclude-dir=node_modules \
  --exclude-dir=dist \
  --exclude-dir=ui/dist \
  --exclude-dir=ui/node_modules \
  | grep -v '^./scripts/check-secrets.sh:' \
  | cut -d: -f1-2 || true)"

if [[ -n "$HIGH_RISK_MATCHES" ]]; then
  echo "FAIL high-risk secret-like patterns found:"
  echo "$HIGH_RISK_MATCHES"
  exit 1
fi

INFO_MATCHES="$(grep -RInE "$INFO_PATTERN" . \
  --exclude-dir=.git \
  --exclude-dir=node_modules \
  --exclude-dir=dist \
  --exclude-dir=ui/dist \
  --exclude-dir=ui/node_modules \
  | grep -v '^./scripts/check-secrets.sh:' \
  | cut -d: -f1-2 || true)"

if [[ -n "$INFO_MATCHES" ]]; then
  echo "INFO local demo or placeholder findings:"
  echo "$INFO_MATCHES"
fi

echo "PASS no high-risk secret patterns found"
