# Orbit v0.0.1 Release Audit

## Summary

- Safe to push private repo: yes
- Safe to make public: no
- Blocking issues:
  - None after removing hardcoded runtime secret fallbacks from Go source and fixing the local PostgreSQL password mismatch in `values-local.yaml`
- Non-blocking warnings:
  - `values-local.yaml` contains local-only demo credentials
  - Bundled PostgreSQL is local preview only
  - No production secret-management flow yet

## Manual Review Coverage

- Go backend
- Helm values and templates
- Scripts
- UI
- Docs and examples

## Secret Scan Results

| File | Finding Type | Classification | Action Taken |
| --- | --- | --- | --- |
| `internal/secrets/secrets.go` | Hardcoded JWT and bootstrap fallbacks | MUST_REMOVE | Replaced with required mounted-file loading |
| `internal/db/config.go` | Hardcoded DB password fallback | MUST_REMOVE | Replaced with required mounted-file loading |
| `charts/orbit/values.yaml` | Local-looking credentials in default values | SAFE_PLACEHOLDER | Replaced with `CHANGE_ME_*` placeholders |
| `charts/orbit/values-local.yaml` | Local demo credentials | SAFE_LOCAL_DEMO | Retained and documented as local-only |

No raw secret values are recorded in this audit.

## Hardcoded Credential Review

- JWT handling:
  - JWT signing secret is now required from `/etc/orbit/secrets/ORBIT_JWT_SECRET`
  - Missing or empty file causes startup failure
- Bootstrap admin handling:
  - Username and password are now required from mounted secret files
  - No Go-source fallback remains
- DB secret handling:
  - DB password is now required from `/etc/orbit/db/ORBIT_DB_PASSWORD`
  - Non-secret DB host, port, name, user, and sslmode still use safe local defaults in Go
- Local-only values:
  - Helm `values-local.yaml` retains demo credentials for local Minikube usage only

## RBAC Review

- `orbit-api` permissions:
  - No cluster-wide RBAC
  - Release chart now grants no Kubernetes API read permissions because the API serves persisted PostgreSQL data only
- `orbit-controller` permissions:
  - Read-only cluster RBAC for namespaces, pods, services, configmaps, events, nodes, and selected workload resources
  - Metrics API reads are optional and read-only
- Secret access:
  - No component has cluster-wide secret access

## Generated Artifacts

Removed or ignored:

- `orbit-api`
- `orbit-controller`
- `ui/dist`
- `dist/`
- `build/`
- `.DS_Store`

## Validation Status

- Fresh `orbit-test` Helm deploy: passed
- `./scripts/helm-test-local.sh`: passed
- Direct API smoke test: passed
- Broken workload namespace and pod analysis: passed
- UI screen validation:
  - Dashboard: passed
  - Analyze namespace: passed
  - Analyze pod: passed
  - Inventory: passed
  - Findings: passed
  - Evidence Packs: passed
  - Action Plans: passed
  - Finding Rules: passed
  - Controller Status: passed

## Documentation Privacy Review

README, docs, and examples are safe for a private GitHub push.

They intentionally document local demo credentials and placeholders, but they do not contain production secrets.

## Git History Review

- Not available in this workspace because the repository has not been initialized with Git yet
- History-specific secret review must be repeated after Git import if earlier commits are added

## Final Recommendation

Safe for private GitHub push.

Public release should happen only after user review of:

- local-only demo credentials in `values-local.yaml`
- security notes
- roadmap and preview limitations

## Published Artifacts

- GitHub repository: `https://github.com/erdembestas/orbit`
- Preview release tag: `v0.0.1`
- GitHub release page: `https://github.com/erdembestas/orbit/releases/tag/v0.0.1`
- OCI image: `ghcr.io/erdembestas/orbit-api:v0.0.1`
- OCI image: `ghcr.io/erdembestas/orbit-api:latest`
- OCI image: `ghcr.io/erdembestas/orbit-ui:v0.0.1`
- OCI image: `ghcr.io/erdembestas/orbit-ui:latest`
