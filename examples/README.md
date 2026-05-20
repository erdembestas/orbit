# Example Broken Workloads

Apply the demo workloads:

```bash
kubectl apply -f examples/broken-workloads.yaml
kubectl get all -n orbit-break-test
```

What each workload demonstrates:

- `crashloop-demo`
  - exits repeatedly and should trigger restart-related and availability findings
- `probe-port-mismatch`
  - serves on port `80` while probes target `8080`
  - should trigger probe and availability findings
- `zero-replica-demo`
  - desired replicas set to `0`
  - should trigger the zero-replica finding

Suggested UI flow:

1. Open the Orbit UI
2. Go to `Analyze`
3. Select namespace `orbit-break-test`
4. Generate a namespace evidence pack
5. Run mock reasoning over that evidence pack

Cleanup:

```bash
kubectl delete namespace orbit-break-test
```
