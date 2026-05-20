package controller

type RuleDefinition struct {
	Name           string   `json:"name"`
	ResourceKind   string   `json:"resourceKind"`
	Category       string   `json:"category"`
	Severity       string   `json:"severity"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Condition      string   `json:"condition"`
	EvidenceFields []string `json:"evidenceFields"`
	Limitations    []string `json:"limitations"`
}

func FindingRules() []RuleDefinition {
	return []RuleDefinition{
		{
			Name:           "deployment_zero_replicas",
			ResourceKind:   "Deployment",
			Category:       "availability",
			Severity:       "low",
			Title:          "Deployment has zero replicas",
			Description:    "Flags deployments whose desired replica count is explicitly set to zero.",
			Condition:      "spec.replicas == 0",
			EvidenceFields: []string{"desiredReplicas"},
			Limitations: []string{
				"May be intentional for paused or manually scaled-down workloads.",
				"Does not distinguish scheduled scale-downs from accidental configuration changes.",
			},
		},
		{
			Name:           "deployment_unavailable",
			ResourceKind:   "Deployment",
			Category:       "availability",
			Severity:       "medium|high",
			Title:          "Deployment has unavailable replicas",
			Description:    "Flags deployments whose available replicas are below the desired replica count.",
			Condition:      "status.availableReplicas < spec.replicas",
			EvidenceFields: []string{"desiredReplicas", "availableReplicas", "updatedReplicas"},
			Limitations: []string{
				"May trigger during a normal rolling update before availability recovers.",
				"Does not currently inspect rollout conditions or progress deadlines.",
			},
		},
		{
			Name:           "pod_not_healthy",
			ResourceKind:   "Pod",
			Category:       "workload-health",
			Severity:       "medium|high",
			Title:          "Pod is not healthy",
			Description:    "Flags pods whose phase is neither Running nor Succeeded.",
			Condition:      "status.phase not in {Running, Succeeded}",
			EvidenceFields: []string{"phase", "waitingReason", "terminatedReason", "restartCount"},
			Limitations: []string{
				"Pending pods during short scheduling windows can be transient.",
				"Does not yet distinguish between benign startup delay and durable failure in all cases.",
			},
		},
		{
			Name:           "pod_container_restarts",
			ResourceKind:   "Pod",
			Category:       "workload-health",
			Severity:       "info|low",
			Title:          "Pod has container restarts",
			Description:    "Flags pods whose containers have restarted at least once.",
			Condition:      "sum(status.containerStatuses.restartCount) > 0",
			EvidenceFields: []string{"restartCount", "phase", "waitingReason", "terminatedReason"},
			Limitations: []string{
				"Can be noisy for long-lived pods with historical restarts that are now healthy.",
				"Does not yet model restart recency or decay older restart history.",
			},
		},
		{
			Name:           "probe_port_mismatch",
			ResourceKind:   "Deployment|Pod",
			Category:       "configuration",
			Severity:       "high",
			Title:          "Probe may target wrong port",
			Description:    "Flags readiness or liveness probes whose HTTP target port does not match declared container ports when that mismatch is detectable.",
			Condition:      "probe.httpGet.port not in declared container ports or named port map",
			EvidenceFields: []string{"probePort"},
			Limitations: []string{
				"Only evaluates HTTP probe ports today.",
				"Cannot detect cases where the container exposes the port indirectly without declaring it in the pod spec.",
			},
		},
	}
}
