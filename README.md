# Orbit

Single-cluster Kubernetes control plane for inventory, findings, evidence packs, and human-reviewed operational reasoning.

## Status

v0.0.1 preview.

Orbit is currently a local single-cluster preview. It is not production-ready.

## What Is Orbit?

Orbit runs inside a Kubernetes cluster and observes the same cluster where it is installed.

It collects Kubernetes inventory and events, creates deterministic findings, generates compact evidence packs, and provides mock reasoning with draft action plans.

Orbit does not apply changes to the cluster in v0.0.1.

## Current Capabilities

- Helm-based local install
- Orbit UI
- Orbit API
- Orbit controller
- PostgreSQL-backed local auth
- Single-cluster inventory collection
- Kubernetes resource and event collection
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

## Components

- `orbit-ui`
- `orbit-api`
- `orbit-controller`
- `orbit-postgres`

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

Cleanup:

```bash
kubectl delete namespace orbit-break-test
```

See [examples/README.md](/Users/erdem/Desktop/Orbit/examples/README.md) for details.

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

See [docs/SECURITY.md](/Users/erdem/Desktop/Orbit/docs/SECURITY.md).

## Repository Checks

```bash
./scripts/check-secrets.sh
./scripts/check-local.sh
./scripts/release-v0.0.1-check.sh
```

The current local preview was validated end to end against a fresh `orbit-test` install, API smoke checks, UI login, namespace analysis, pod analysis, findings, evidence packs, action plans, and controller status screens.

## API Documentation

- [docs/API.md](/Users/erdem/Desktop/Orbit/docs/API.md)
- [api/openapi.yaml](/Users/erdem/Desktop/Orbit/api/openapi.yaml)

## Roadmap

See [docs/ROADMAP.md](/Users/erdem/Desktop/Orbit/docs/ROADMAP.md).

## License

TBD
