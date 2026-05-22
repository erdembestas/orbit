# Orbit API

Orbit v0.0.1 exposes a small authenticated API for local single-cluster operation.

Base URL:

- UI proxy: `http://localhost:8080`
- Direct API debug: `http://localhost:8081`

All protected endpoints require:

`Authorization: Bearer <token>`

## Auth Flow

1. Login with `POST /api/v1/auth/login`
2. Store the returned bearer token
3. Call protected endpoints with the `Authorization` header

## Public Endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /api/v1/info`
- `POST /api/v1/auth/login`

## Protected Endpoints

- `GET /api/v1/auth/me`
- `GET /api/v1/protected`
- `GET /api/v1/clusters`
- `GET /api/v1/finding-rules`
- `GET /api/v1/controller/status`
- `GET /api/v1/cluster/health`
- `GET /api/v1/cluster/health/nodes`
- `GET /api/v1/cluster/health/namespaces`
- `GET /api/v1/cluster/health/history`
- `GET /api/v1/inventory/resources`
- `GET /api/v1/findings`
- `GET /api/v1/findings/{id}`
- `GET /api/v1/findings/{id}/evidence-pack`
- `POST /api/v1/findings/{id}/reason`
- `GET /api/v1/evidence-packs`
- `POST /api/v1/evidence-packs/generate`
- `GET /api/v1/evidence-packs/{id}`
- `POST /api/v1/evidence-packs/{id}/reason`
- `GET /api/v1/action-plans`
- `GET /api/v1/action-plans/{id}`

## Login Example

```bash
curl -s -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' | jq .
```

## Current User Example

```bash
curl -s http://localhost:8081/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Controller Status Example

```bash
curl -s http://localhost:8081/api/v1/controller/status \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Cluster Health Example

```bash
curl -s http://localhost:8081/api/v1/cluster/health \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Example response shape:

```json
{
  "observedAt": "2026-05-21T09:07:17Z",
  "metricsAvailable": false,
  "metricsError": "metrics API unavailable",
  "healthStatus": "warning",
  "healthScore": 80,
  "summary": {
    "nodes": { "count": 1, "readyCount": 1, "notReadyCount": 0 },
    "pods": { "count": 19, "runningCount": 18, "pendingCount": 0, "failedCount": 1 }
  }
}
```

The same health report is shown in the UI on `Dashboard` and `Cluster Health`.

## Node Health Example

```bash
curl -s http://localhost:8081/api/v1/cluster/health/nodes \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Namespace Health Example

```bash
curl -s http://localhost:8081/api/v1/cluster/health/namespaces \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Cluster Health History Example

```bash
curl -s "http://localhost:8081/api/v1/cluster/health/history?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Inventory Example

```bash
curl -s "http://localhost:8081/api/v1/inventory/resources?kind=Pod&namespace=orbit-test" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Findings Example

```bash
curl -s "http://localhost:8081/api/v1/findings?status=open" \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Namespace Evidence Pack Example

```bash
curl -s -X POST http://localhost:8081/api/v1/evidence-packs/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"scopeType":"namespace","namespace":"orbit-break-test","persist":true}' | jq .
```

## Pod Evidence Pack Example

```bash
POD=$(kubectl -n orbit-test get pod -l app.kubernetes.io/name=orbit-api -o jsonpath='{.items[0].metadata.name}')

curl -s -X POST http://localhost:8081/api/v1/evidence-packs/generate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"scopeType\":\"pod\",\"namespace\":\"orbit-test\",\"name\":\"$POD\",\"persist\":true}" | jq .
```

## Reasoning Example

```bash
curl -s -X POST http://localhost:8081/api/v1/evidence-packs/$EVIDENCE_ID/reason \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Action Plans Example

```bash
curl -s http://localhost:8081/api/v1/action-plans \
  -H "Authorization: Bearer $TOKEN" | jq .
```

## Notes

- Reasoning is mock only in v0.0.1.
- Action plans are draft only.
- No apply, remediation, approval, or execution flow exists.
- Orbit returns compact evidence packs, not full-cluster dumps, for reasoning inputs.
- Cluster health returns `metricsAvailable=false` when `metrics-server` is absent, but still reports node conditions, pod counts, and warning-event-based health status.
- The UI reads all cluster health data through the same authenticated `/api/v1/cluster/health*` endpoints.
