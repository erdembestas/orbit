# Orbit

Single-cluster Kubernetes control plane for inventory, findings, evidence packs, and human-reviewed operational reasoning.

## Status

v0.0.1 preview.

Orbit is currently a local single-cluster preview. It is not production-ready.

## What Is Orbit?

Orbit runs inside a Kubernetes cluster and observes the same cluster where it is installed.

It collects Kubernetes inventory and events, creates deterministic findings, generates compact evidence packs, and provides mock reasoning with draft action plans.

Orbit does not apply changes to the cluster in v0.0.1.

![Orbit UI](https://raw.githubusercontent.com/erdembestas/orbit/main/docs/images/orbit-ui.png)

## How Orbit Works

Orbit is built as an in-cluster control plane.

1. `orbit-controller` connects to the Kubernetes API of the same cluster where Orbit is installed.
2. The controller collects inventory, recent events, node conditions and capacity, and optional metrics, then normalizes that state into PostgreSQL.
3. Deterministic finding rules evaluate the stored state and open findings such as unavailable deployments, unhealthy pods, restart-heavy workloads, and probe mismatches.
4. Orbit computes cluster, node, and namespace health snapshots from node conditions, pod phases, warning events, and metrics when `metrics-server` is available.
5. Orbit builds compact evidence packs for findings, namespaces, or pods so later reasoning only sees the relevant context instead of a full cluster dump.
6. `orbit-api` exposes authenticated APIs for inventory, findings, evidence packs, cluster health, mock reasoning, and draft action plans.
7. `orbit-ui` presents the operational workflow through the browser and proxies `/api` requests to `orbit-api`.

## Service Responsibilities

- `orbit-ui`
  - Serves the React UI on port `8080`
  - Proxies `/api/*` requests to `orbit-api`
  - Provides dashboard, analysis, findings, evidence pack, and draft action plan screens
- `orbit-api`
  - Handles login and JWT-based API access
  - Reads file-mounted runtime config, secrets, and DB settings
  - Serves inventory, findings, evidence packs, reasoning, and action plan APIs
  - Does not hold cluster-wide Kubernetes RBAC
- `orbit-controller`
  - Uses in-cluster Kubernetes credentials and read-only RBAC
  - Collects namespaces, workloads, pods, services, configmaps, and events
  - Collects node conditions, node capacity or allocatable, optional metrics-server data, and cluster health snapshots
  - Generates findings and evidence packs and stores them in PostgreSQL
- `orbit-postgres`
  - Stores auth data, inventory, events, findings, evidence packs, mock reasoning runs, and draft action plans

## Current Capabilities

- Helm-based local install
- Orbit UI
- Orbit API
- Orbit controller
- PostgreSQL-backed local auth
- Single-cluster inventory collection
- Kubernetes resource and event collection
- Cluster health snapshots and health report APIs
- Cluster Health dashboard page
- Deterministic findings
- Namespace evidence packs
- Pod evidence packs
- Mock reasoning
- Draft action plans
- API smoke script
- Example broken workloads

## Not Implemented Yet

- Real LLM provider
- OpenAI or Ollama integration
- LDAP
- Approval workflow
- Execution, apply, or remediation
- Multi-cluster agent mode
- Production-grade secret management
- Production-grade PostgreSQL

## Architecture

```text
Browser
  -> orbit-ui
  -> orbit-api
  -> orbit-postgres

orbit-controller
  -> Kubernetes API
  -> orbit-postgres
```

- `orbit-ui` serves the React UI and proxies `/api` to `orbit-api`.
- `orbit-api` handles auth and REST APIs and reads or writes PostgreSQL.
- `orbit-controller` observes the cluster with read-only RBAC.
- `orbit-postgres` stores auth, inventory, findings, evidence packs, mock reasoning runs, and draft action plans.

## Operational Flow

```text
User
  -> orbit-ui
  -> orbit-api
  -> orbit-postgres

orbit-controller
  -> Kubernetes API
  -> orbit-postgres

Stored cluster state
  -> deterministic finding rules
  -> cluster health snapshots
  -> evidence packs
  -> mock reasoning
  -> draft action plans
```

In practical use, the normal flow is:

- Deploy Orbit into a single cluster.
- Let `orbit-controller` collect current cluster state.
- Review findings or generate a namespace or pod analysis.
- Review Cluster Health for node readiness, namespace pressure, and recent health history.
- Review cluster health, node health, and namespace workload pressure.
- Inspect the compact evidence pack.
- Run mock reasoning to produce a draft action plan for human review.
- No execution or remediation occurs in `v0.0.1`.

## Components

- `orbit-ui`
- `orbit-api`
- `orbit-controller`
- `orbit-postgres`

## Published Images

- `ghcr.io/erdembestas/orbit-api:v0.0.1`
- `ghcr.io/erdembestas/orbit-api:latest`
- `ghcr.io/erdembestas/orbit-ui:v0.0.1`
- `ghcr.io/erdembestas/orbit-ui:latest`

`orbit-controller` runs from the same image as `orbit-api` with a different command.

## Local Install With Minikube

```bash
minikube start --driver=docker

./scripts/ui-install.sh
./scripts/build-local-image.sh
./scripts/helm-install-local.sh
./scripts/verify-local.sh
./scripts/helm-test-local.sh
./scripts/port-forward-ui.sh
```

Open:

`http://localhost:8080`

Login:

`admin / admin`

`admin / admin` is local-only and not production safe.

The chart defaults now point at the published GHCR images above. The local Minikube workflow still uses `charts/orbit/values-local.yaml`, which overrides those defaults to use the locally built `orbit-api:local` and `orbit-ui:local` images.

## Direct API Debug

```bash
./scripts/port-forward-api.sh

ORBIT_API_BASE_URL=http://localhost:8081 ./scripts/test-api-local.sh
```

API base URL:

`http://localhost:8081`

## Example Broken Workloads

```bash
kubectl apply -f examples/broken-workloads.yaml
kubectl get all -n orbit-break-test
```

Then in the UI:

- Open `Analyze`
- Select `orbit-break-test`
- Generate a namespace evidence pack
- Run mock reasoning
- Switch to pod analysis and inspect a failing pod in the same namespace
- Open `Cluster Health` to review node and namespace pressure after the controller interval

Cleanup:

```bash
kubectl delete namespace orbit-break-test
```

See [examples/README.md](./examples/README.md) for details.

## Runtime Config And Secrets

Orbit reads runtime configuration from mounted files, not normal environment variables.

Expected config files:

- `/etc/orbit/config/ORBIT_ENV`
- `/etc/orbit/config/ORBIT_LOG_LEVEL`
- `/etc/orbit/config/ORBIT_API_PORT`
- `/etc/orbit/config/ORBIT_AUTH_MODE`
- `/etc/orbit/config/ORBIT_CLUSTER_NAME`
- `/etc/orbit/config/ORBIT_CLUSTER_TYPE`
- `/etc/orbit/config/ORBIT_MODE`
- `/etc/orbit/config/ORBIT_CONTROLLER_ENABLED`
- `/etc/orbit/config/ORBIT_CONTROLLER_INTERVAL_SECONDS`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_ENABLED`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_INTERVAL_SECONDS`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_NODE_CPU_WARN_PERCENT`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_NODE_CPU_CRITICAL_PERCENT`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_NODE_MEMORY_WARN_PERCENT`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_NODE_MEMORY_CRITICAL_PERCENT`
- `/etc/orbit/config/ORBIT_CLUSTER_HEALTH_STALE_AFTER_SECONDS`
- `/etc/orbit/config/ORBIT_EVIDENCE_MAX_EVENTS`
- `/etc/orbit/config/ORBIT_EVIDENCE_MAX_RELATED_RESOURCES`
- `/etc/orbit/config/ORBIT_EVIDENCE_MAX_LOG_LINES`
- `/etc/orbit/config/ORBIT_EVIDENCE_MAX_TOKEN_ESTIMATE`

Required secret files:

- `/etc/orbit/secrets/ORBIT_BOOTSTRAP_ADMIN_USERNAME`
- `/etc/orbit/secrets/ORBIT_BOOTSTRAP_ADMIN_PASSWORD`
- `/etc/orbit/secrets/ORBIT_JWT_SECRET`
- `/etc/orbit/db/ORBIT_DB_PASSWORD`

Required secret files are fail-fast. If a required secret file is missing or empty, `orbit-api` will not start.

## Security Notes

- `charts/orbit/values-local.yaml` is for local development only.
- Local JWT secret values are not production-safe.
- Local PostgreSQL passwords are not production-safe.
- Override the bootstrap admin password, JWT secret, and DB password before any shared or production use.
- No remediation or apply path exists in v0.0.1.
- Mock reasoning suggestions are not executed.
- Read-only suggestions remain for human review only.
- `metrics-server` is optional. If it is missing, Cluster Health still works with `metricsAvailable=false`.

See [docs/SECURITY.md](./docs/SECURITY.md).

## Repository Checks

```bash
./scripts/check-secrets.sh
./scripts/check-local.sh
./scripts/release-v0.0.1-check.sh
```

The current local preview was validated end to end against a fresh `orbit-test` install, API smoke checks, UI login, namespace analysis, pod analysis, findings, evidence packs, action plans, and controller status screens.

## API Documentation

- [docs/API.md](./docs/API.md)
- [docs/CLUSTER_HEALTH.md](./docs/CLUSTER_HEALTH.md)
- [api/openapi.yaml](./api/openapi.yaml)

## GitHub Release

- GitHub repository: [erdembestas/orbit](https://github.com/erdembestas/orbit)
- Preview tag: [`v0.0.1`](https://github.com/erdembestas/orbit/releases/tag/v0.0.1)

## Roadmap

See [docs/ROADMAP.md](./docs/ROADMAP.md).

## License

This project is licensed under the Apache License 2.0. See [LICENSE](./LICENSE).
