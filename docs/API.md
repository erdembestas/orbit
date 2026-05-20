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
