# Orbit Finding Rules

Orbit Phase 1 uses deterministic controller-side classification rules. These rules inspect persisted cluster inventory, events, and derived evidence. They do not call an LLM, they do not apply changes, and they do not mutate the cluster.

## What the rules do

- They create or update findings in PostgreSQL.
- They attach rule-specific evidence fields that explain why a finding fired.
- They feed compact evidence packs that can later be used for mock reasoning or a future real reasoning provider.

## Current rules

| Rule name | Resource kind | Condition | Category | Severity | Title | Evidence fields | Known limitations |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `deployment_zero_replicas` | Deployment | `spec.replicas == 0` | availability | low | Deployment has zero replicas | `desiredReplicas` | May be intentional for a paused or manually scaled-down workload. Does not distinguish deliberate scale-to-zero from accidental configuration drift. |
| `deployment_unavailable` | Deployment | `status.availableReplicas < spec.replicas` | availability | medium or high | Deployment has unavailable replicas | `desiredReplicas`, `availableReplicas`, `updatedReplicas` | Can fire during a normal rollout before availability recovers. Does not inspect rollout conditions or progress deadline state yet. |
| `pod_not_healthy` | Pod | `status.phase` is not `Running` or `Succeeded` | workload-health | medium or high | Pod is not healthy | `phase`, `waitingReason`, `terminatedReason`, `restartCount` | Pending pods can be transient. The rule improves severity for obvious crash reasons, but it still does not fully model startup grace periods. |
| `pod_container_restarts` | Pod | Sum of container `restartCount` values is greater than zero | workload-health | info or low | Pod has container restarts | `restartCount`, `phase`, `waitingReason`, `terminatedReason` | Historical restarts can still be noisy on otherwise healthy long-lived pods. Restart recency decay is not implemented yet. |
| `probe_port_mismatch` | Deployment or Pod | HTTP readiness/liveness probe port does not match declared container ports or named ports when detectable | configuration | high | Probe may target wrong port | `probePort` | Only HTTP probe ports are evaluated. The rule cannot detect cases where the application listens on a port that is not declared in the pod spec. |

## Current severity behavior

- Deployment unavailable becomes `high` when desired replicas are greater than zero and available replicas are zero.
- Deployment unavailable becomes `medium` when availability is only partially degraded.
- Pod not healthy becomes `high` for failed pods or severe waiting reasons such as `CrashLoopBackOff`, `ImagePullBackOff`, `ErrImagePull`, `CreateContainerConfigError`, `CreateContainerError`, or `RunContainerError`.
- Pod not healthy remains `medium` for other non-running phases.
- Pod container restarts becomes `info` when the pod is currently `Running` or `Succeeded`.
- Pod container restarts becomes `low` when the pod is still unhealthy.

## What evidence gets stored

Each finding stores a compact `evidence_json` payload with the specific fields that explain the rule trigger. Evidence packs later expand on that with related pods, warning events, probe summaries, resource summaries, and suspected deterministic causes.

## Limits and false positives

- Findings are intentionally simple and explainable. They are not predictive.
- Rollouts can temporarily trigger deployment availability findings.
- Restart-based findings still lack a recency window, so older restart history may remain visible.
- Probe mismatch detection only works when the declared probe target and declared container ports can be compared directly.
- Orbit does not auto-resolve findings yet, so open findings may persist until later lifecycle work is implemented.
