#!/usr/bin/env bash
set -euo pipefail

NAMESPACE="${NAMESPACE:-orbit-test}"

kubectl -n "$NAMESPACE" get pods -o wide
kubectl -n "$NAMESPACE" get svc
kubectl -n "$NAMESPACE" get endpoints
kubectl -n "$NAMESPACE" get pvc
kubectl -n "$NAMESPACE" logs deploy/orbit-postgres --tail=50
kubectl -n "$NAMESPACE" logs deploy/orbit-api --tail=100
kubectl -n "$NAMESPACE" logs deploy/orbit-controller --tail=100
kubectl -n "$NAMESPACE" logs deploy/orbit-ui --tail=50
kubectl -n "$NAMESPACE" exec deploy/orbit-api -- ls -l /etc/orbit/config
kubectl -n "$NAMESPACE" exec deploy/orbit-api -- ls -l /etc/orbit/secrets
kubectl -n "$NAMESPACE" exec deploy/orbit-api -- ls -l /etc/orbit/db
kubectl -n "$NAMESPACE" get deploy orbit-api -o jsonpath='{.spec.template.spec.containers[0].envFrom}{"\n"}'
kubectl -n "$NAMESPACE" get deploy orbit-api -o jsonpath='{.spec.template.spec.containers[0].env}{"\n"}'
