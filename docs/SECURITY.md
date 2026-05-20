# Orbit Security Notes

## Preview Status

Orbit v0.0.1 is a local preview. It is not production-ready.

## Runtime Secrets

Orbit does not ship hardcoded runtime JWT, bootstrap password, or DB password values in Go source.

Required mounted secret files:

- `/etc/orbit/secrets/ORBIT_BOOTSTRAP_ADMIN_USERNAME`
- `/etc/orbit/secrets/ORBIT_BOOTSTRAP_ADMIN_PASSWORD`
- `/etc/orbit/secrets/ORBIT_JWT_SECRET`
- `/etc/orbit/db/ORBIT_DB_PASSWORD`

If any required secret file is missing or empty, startup fails fast.

## Local Values

- `charts/orbit/values-local.yaml` is local-development only.
- Local bootstrap credentials and JWT values in that file are not production-safe.
- Local PostgreSQL password values in that file are not production-safe.

Override all bootstrap, JWT, and DB credentials before any shared or production use.

## Kubernetes Access

- `orbit-api` does not require cluster-wide Kubernetes RBAC.
- `orbit-controller` uses read-only cluster inventory RBAC.
- No component is granted cluster-wide secret access.

## Execution Safety

- No remediation or apply path exists in v0.0.1.
- No approval or execution workflow exists yet.
- Mock reasoning outputs remain advisory and unexecuted.
- No real LLM or external provider calls are enabled in v0.0.1.

## Reporting Security Issues

Please report security issues privately and avoid opening public issues with sensitive details.
