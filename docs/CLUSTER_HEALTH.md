# Orbit Cluster Health

Orbit now stores cluster-level health snapshots in PostgreSQL and exposes them through authenticated API endpoints.

## What Orbit Collects

`orbit-controller` gathers:

- node readiness and node conditions
- node capacity and allocatable CPU or memory
- pod counts per node
- namespace pod pressure summaries
- warning event counts
- node metrics from `metrics.k8s.io` when available
- pod metrics from `metrics.k8s.io` when available

## Metrics API Behavior

Orbit tries to read:

- `metrics.k8s.io/v1beta1` `NodeMetrics`
- `metrics.k8s.io/v1beta1` `PodMetrics`

If the metrics API is unavailable:

- the controller does not crash
- the snapshot is still written
- `metricsAvailable` is `false`
- `metricsError` contains a compact error string
- cluster or node status is still derived from node readiness, node pressure, pod phases, and warning events

## Health Score Rules

Cluster scoring starts at `100` and subtracts points for:

- not ready nodes
- `MemoryPressure`, `DiskPressure`, or `PIDPressure`
- cluster CPU above warn or critical thresholds
- cluster memory above warn or critical thresholds
- pending pods
- failed pods
- warning events
- metrics unavailable

Score mapping:

- `>= 85` => `healthy`
- `>= 60` => `warning`
- `< 60` => `critical`
- stale or missing data may be reported as `unknown`

Node scoring is deterministic:

- `Ready=false` => critical
- pressure conditions => warning or critical
- CPU or memory threshold breach => warning or critical
- otherwise healthy

Namespace scoring is deterministic:

- failed pods => warning or critical
- pending pods => warning
- restart-heavy namespaces => warning
- warning events => warning
- otherwise healthy

## API Endpoints

- `GET /api/v1/cluster/health`
- `GET /api/v1/cluster/health/nodes`
- `GET /api/v1/cluster/health/namespaces`
- `GET /api/v1/cluster/health/history?limit=50`

All require `Authorization: Bearer <token>`.

## UI Display

The Orbit UI exposes cluster health in two places:

- `Dashboard` shows a compact summary with score, status, ready nodes, pod counts, and metrics availability.
- `Cluster Health` shows the latest cluster report, node health table, namespace pressure table, and recent history.

If metrics are unavailable, the UI shows a warning banner and treats CPU or memory percentages as unavailable instead of failing the page.

## RBAC

Only `orbit-controller` reads cluster health inputs.

Read-only permissions:

- core: `nodes`, `pods`, `namespaces`, `events`
- metrics.k8s.io: `nodes`, `pods`
- verbs: `get`, `list`, `watch`

`orbit-api` does not receive cluster-wide Kubernetes RBAC.

## Limitations

- health scoring is intentionally simple and deterministic
- metrics depend on `metrics-server`
- namespace health is pressure-oriented, not SLO-aware
- no remediation or automatic action is performed
