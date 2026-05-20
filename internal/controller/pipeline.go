package controller

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"orbit/internal/config"
	"orbit/internal/store"
)

type FindingCandidate struct {
	Severity        string
	Category        string
	Title           string
	Description     string
	ResourceUID     string
	ResourceKind    string
	ResourceName    string
	ResourceNS      string
	SuspectedCauses []string
	Evidence        map[string]any
}

type State struct {
	Deployments []appsv1.Deployment
	ReplicaSets []appsv1.ReplicaSet
	Pods        []corev1.Pod
	Events      []corev1.Event
	PodMetrics  []metricsv1beta1.PodMetrics
}

func EvaluateFindings(state State) []FindingCandidate {
	var findings []FindingCandidate

	for _, deployment := range state.Deployments {
		desired := int32(1)
		if deployment.Spec.Replicas != nil {
			desired = *deployment.Spec.Replicas
		}
		if desired == 0 {
			findings = append(findings, FindingCandidate{
				Severity:        "low",
				Category:        "availability",
				Title:           "Deployment has zero replicas",
				Description:     fmt.Sprintf("Deployment %s/%s is configured with zero desired replicas.", deployment.Namespace, deployment.Name),
				ResourceUID:     string(deployment.UID),
				ResourceKind:    "Deployment",
				ResourceName:    deployment.Name,
				ResourceNS:      deployment.Namespace,
				SuspectedCauses: []string{"The deployment spec explicitly sets replicas to zero.", "The workload may have been scaled down intentionally or by automation."},
				Evidence: map[string]any{
					"desiredReplicas": desired,
				},
			})
		}

		available := deployment.Status.AvailableReplicas
		if available < desired {
			severity := "medium"
			if available == 0 && desired > 0 {
				severity = "high"
			}
			findings = append(findings, FindingCandidate{
				Severity:        severity,
				Category:        "availability",
				Title:           "Deployment has unavailable replicas",
				Description:     fmt.Sprintf("Deployment %s/%s has %d available replicas out of %d desired replicas.", deployment.Namespace, deployment.Name, available, desired),
				ResourceUID:     string(deployment.UID),
				ResourceKind:    "Deployment",
				ResourceName:    deployment.Name,
				ResourceNS:      deployment.Namespace,
				SuspectedCauses: []string{"Pods may be failing readiness checks.", "The rollout may be blocked by scheduling or image issues.", "Recent warning events may explain why replicas are unavailable."},
				Evidence: map[string]any{
					"desiredReplicas":   desired,
					"availableReplicas": available,
					"updatedReplicas":   deployment.Status.UpdatedReplicas,
				},
			})
		}

		if suspect, ok := deploymentProbePortMismatch(deployment); ok {
			findings = append(findings, FindingCandidate{
				Severity:        "high",
				Category:        "configuration",
				Title:           "Probe may target wrong port",
				Description:     fmt.Sprintf("Deployment %s/%s contains a probe targeting port %s that does not match declared container ports.", deployment.Namespace, deployment.Name, suspect),
				ResourceUID:     string(deployment.UID),
				ResourceKind:    "Deployment",
				ResourceName:    deployment.Name,
				ResourceNS:      deployment.Namespace,
				SuspectedCauses: []string{"The readiness or liveness probe refers to a port that the container does not expose.", "A named probe port may not match the container port names."},
				Evidence: map[string]any{
					"probePort": suspect,
				},
			})
		}
	}

	for _, pod := range state.Pods {
		if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
			waitingReason := firstWaitingReason(pod)
			terminated := firstTerminatedReason(pod)
			restarts := sumContainerRestarts(pod)
			findings = append(findings, FindingCandidate{
				Severity:        podHealthSeverity(pod),
				Category:        "workload-health",
				Title:           "Pod is not healthy",
				Description:     fmt.Sprintf("Pod %s/%s is in phase %s.", pod.Namespace, pod.Name, pod.Status.Phase),
				ResourceUID:     string(pod.UID),
				ResourceKind:    "Pod",
				ResourceName:    pod.Name,
				ResourceNS:      pod.Namespace,
				SuspectedCauses: []string{"The pod may be Pending, Failed, or otherwise blocked from becoming healthy.", "Recent events and container state transitions usually explain the unhealthy pod phase."},
				Evidence: map[string]any{
					"phase":            string(pod.Status.Phase),
					"waitingReason":    waitingReason,
					"terminatedReason": terminated,
					"restartCount":     restarts,
				},
			})
		}

		restarts := sumContainerRestarts(pod)
		if restarts > 0 {
			waitingReason := firstWaitingReason(pod)
			terminated := firstTerminatedReason(pod)
			findings = append(findings, FindingCandidate{
				Severity:        podRestartSeverity(pod),
				Category:        "workload-health",
				Title:           "Pod has container restarts",
				Description:     fmt.Sprintf("Pod %s/%s has %d total container restarts.", pod.Namespace, pod.Name, restarts),
				ResourceUID:     string(pod.UID),
				ResourceKind:    "Pod",
				ResourceName:    pod.Name,
				ResourceNS:      pod.Namespace,
				SuspectedCauses: []string{"A container may be crashing repeatedly.", "Resource pressure or failing probes can drive restart loops."},
				Evidence: map[string]any{
					"restartCount":     restarts,
					"phase":            string(pod.Status.Phase),
					"waitingReason":    waitingReason,
					"terminatedReason": terminated,
				},
			})
		}

		if suspect, ok := podProbePortMismatch(pod); ok {
			findings = append(findings, FindingCandidate{
				Severity:        "high",
				Category:        "configuration",
				Title:           "Probe may target wrong port",
				Description:     fmt.Sprintf("Pod %s/%s contains a probe targeting port %s that does not match declared container ports.", pod.Namespace, pod.Name, suspect),
				ResourceUID:     string(pod.UID),
				ResourceKind:    "Pod",
				ResourceName:    pod.Name,
				ResourceNS:      pod.Namespace,
				SuspectedCauses: []string{"The readiness or liveness probe refers to a port that the container does not expose.", "A named probe port may not match the container port names."},
				Evidence: map[string]any{
					"probePort": suspect,
				},
			})
		}
	}

	return findings
}

func BuildEvidencePack(candidate FindingCandidate, resource store.KubernetesResource, cluster store.Cluster, state State, cfg config.Config) (json.RawMessage, int, error) {
	replicaSetByUID := map[string]appsv1.ReplicaSet{}
	replicaSetsByName := map[string]appsv1.ReplicaSet{}
	for _, rs := range state.ReplicaSets {
		replicaSetByUID[string(rs.UID)] = rs
		replicaSetsByName[rs.Namespace+"/"+rs.Name] = rs
	}

	deploymentsByUID := map[string]appsv1.Deployment{}
	for _, deployment := range state.Deployments {
		deploymentsByUID[string(deployment.UID)] = deployment
	}

	podMetricsByKey := map[string]map[string]metricsv1beta1.ContainerMetrics{}
	for _, metric := range state.PodMetrics {
		key := metric.Namespace + "/" + metric.Name
		podMetricsByKey[key] = map[string]metricsv1beta1.ContainerMetrics{}
		for _, container := range metric.Containers {
			podMetricsByKey[key][container.Name] = container
		}
	}

	pack := map[string]any{
		"finding": map[string]any{
			"id":          "",
			"severity":    candidate.Severity,
			"category":    candidate.Category,
			"title":       candidate.Title,
			"description": candidate.Description,
		},
		"cluster": map[string]any{
			"id":   cluster.ID,
			"name": cluster.Name,
			"type": cluster.Type,
			"mode": cluster.Mode,
		},
		"affectedResource": map[string]any{
			"kind":        resource.Kind,
			"namespace":   resource.Namespace,
			"name":        resource.Name,
			"uid":         resource.UID,
			"status":      resource.Status,
			"labels":      resource.Labels,
			"annotations": resource.Annotations,
			"raw_json":    decodeRawJSON(resource.RawJSON),
		},
		"suspectedDeterministicCauses": candidate.SuspectedCauses,
		"generatedAt":                  time.Now().UTC().Format(time.RFC3339),
	}

	var relatedPods []corev1.Pod
	var ownerRefs []map[string]any
	var containerStatuses []map[string]any
	var probeSummaries []map[string]any
	var resourceSummaries []map[string]any

	if candidate.ResourceKind == "Deployment" {
		for _, deployment := range state.Deployments {
			if string(deployment.UID) != candidate.ResourceUID {
				continue
			}
			ownerRefs = ownerReferenceSummaries(deployment.OwnerReferences)
			relatedPods = podsForDeployment(deployment, state.Pods, cfg.EvidenceMaxRelatedResources)
			probeSummaries = append(probeSummaries, probeSummariesForContainers(deployment.Spec.Template.Spec.Containers)...)
			resourceSummaries = append(resourceSummaries, resourceSummariesForContainers(deployment.Spec.Template.Spec.Containers)...)
			break
		}
	}

	if candidate.ResourceKind == "Pod" {
		for _, pod := range state.Pods {
			if string(pod.UID) != candidate.ResourceUID {
				continue
			}
			ownerRefs = ownerReferenceSummaries(pod.OwnerReferences)
			relatedPods = append(relatedPods, pod)
			probeSummaries = append(probeSummaries, probeSummariesForContainers(pod.Spec.Containers)...)
			resourceSummaries = append(resourceSummaries, resourceSummariesForContainers(pod.Spec.Containers)...)
			if owner := resolveOwnerDeployment(pod, replicaSetsByName); owner != nil {
				pack["ownerDeployment"] = map[string]any{
					"namespace": pod.Namespace,
					"name":      owner.Name,
				}
			}
			break
		}
	}

	for _, pod := range relatedPods {
		for _, status := range pod.Status.ContainerStatuses {
			containerStatuses = append(containerStatuses, map[string]any{
				"pod":          pod.Name,
				"container":    status.Name,
				"ready":        status.Ready,
				"restartCount": status.RestartCount,
				"state":        containerStateSummary(status.State),
			})
		}
	}

	pack["ownerReferences"] = ownerRefs
	pack["relatedPods"] = relatedPodSummaries(relatedPods, podMetricsByKey)
	pack["containerStatuses"] = containerStatuses
	pack["probeSummary"] = probeSummaries
	pack["resourceSummary"] = resourceSummaries
	pack["recentWarningEvents"] = warningEventSummaries(candidate, relatedPods, state.Events, cfg.EvidenceMaxEvents)
	pack["metricsSummary"] = metricsSummary(relatedPods, podMetricsByKey)

	trimmedPack, estimate := enforceTokenBudget(pack, cfg)
	trimmedPack["tokenEstimate"] = estimate

	payload, err := json.Marshal(trimmedPack)
	if err != nil {
		return nil, 0, err
	}
	return payload, estimate, nil
}

func enforceTokenBudget(pack map[string]any, cfg config.Config) (map[string]any, int) {
	estimate := tokenEstimate(pack)
	if estimate <= cfg.EvidenceMaxTokenEstimate {
		return pack, estimate
	}

	affected, ok := pack["affectedResource"].(map[string]any)
	if ok {
		delete(affected, "raw_json")
	}
	estimate = tokenEstimate(pack)
	if estimate <= cfg.EvidenceMaxTokenEstimate {
		return pack, estimate
	}

	if events, ok := pack["recentWarningEvents"].([]map[string]any); ok {
		for len(events) > 3 && estimate > cfg.EvidenceMaxTokenEstimate {
			events = events[:len(events)-1]
			pack["recentWarningEvents"] = events
			estimate = tokenEstimate(pack)
		}
	}

	if related, ok := pack["relatedPods"].([]map[string]any); ok {
		for len(related) > 1 && estimate > cfg.EvidenceMaxTokenEstimate {
			related = related[:len(related)-1]
			pack["relatedPods"] = related
			estimate = tokenEstimate(pack)
		}
	}

	if estimate > cfg.EvidenceMaxTokenEstimate {
		truncateNestedStrings(pack, 120)
		estimate = tokenEstimate(pack)
	}

	if estimate > cfg.EvidenceMaxTokenEstimate {
		delete(pack, "metricsSummary")
		estimate = tokenEstimate(pack)
	}

	if estimate > cfg.EvidenceMaxTokenEstimate {
		truncateNestedStrings(pack, 48)
		estimate = tokenEstimate(pack)
	}

	if estimate > cfg.EvidenceMaxTokenEstimate {
		pack["recentWarningEvents"] = []map[string]any{}
		estimate = tokenEstimate(pack)
	}

	return pack, estimate
}

func tokenEstimate(value any) int {
	payload, err := json.Marshal(value)
	if err != nil {
		return 0
	}
	return (len(payload) + 3) / 4
}

func podHealthSeverity(pod corev1.Pod) string {
	if pod.Status.Phase == corev1.PodFailed {
		return "high"
	}
	switch firstWaitingReason(pod) {
	case "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull", "CreateContainerConfigError", "CreateContainerError", "RunContainerError":
		return "high"
	default:
		return "medium"
	}
}

func podRestartSeverity(pod corev1.Pod) string {
	if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
		return "low"
	}
	return "info"
}

func sumContainerRestarts(pod corev1.Pod) int32 {
	restarts := int32(0)
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

func deploymentProbePortMismatch(deployment appsv1.Deployment) (string, bool) {
	return containersProbePortMismatch(deployment.Spec.Template.Spec.Containers)
}

func podProbePortMismatch(pod corev1.Pod) (string, bool) {
	return containersProbePortMismatch(pod.Spec.Containers)
}

func firstWaitingReason(pod corev1.Pod) string {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
			return status.State.Waiting.Reason
		}
	}
	for _, status := range pod.Status.InitContainerStatuses {
		if status.State.Waiting != nil && status.State.Waiting.Reason != "" {
			return status.State.Waiting.Reason
		}
	}
	return ""
}

func firstTerminatedReason(pod corev1.Pod) string {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated != nil && status.State.Terminated.Reason != "" {
			return status.State.Terminated.Reason
		}
	}
	for _, status := range pod.Status.InitContainerStatuses {
		if status.State.Terminated != nil && status.State.Terminated.Reason != "" {
			return status.State.Terminated.Reason
		}
	}
	return ""
}

func containersProbePortMismatch(containers []corev1.Container) (string, bool) {
	for _, container := range containers {
		portNames := map[string]struct{}{}
		portNumbers := map[int32]struct{}{}
		for _, port := range container.Ports {
			if port.Name != "" {
				portNames[port.Name] = struct{}{}
			}
			portNumbers[port.ContainerPort] = struct{}{}
		}

		for _, probe := range []*corev1.Probe{container.ReadinessProbe, container.LivenessProbe} {
			if probe == nil {
				continue
			}
			if probe.HTTPGet == nil {
				continue
			}
			switch probe.HTTPGet.Port.Type {
			case 0:
				if _, ok := portNumbers[probe.HTTPGet.Port.IntVal]; probe.HTTPGet.Port.IntVal != 0 && !ok {
					return fmt.Sprintf("%d", probe.HTTPGet.Port.IntVal), true
				}
			case 1:
				if _, ok := portNames[probe.HTTPGet.Port.StrVal]; probe.HTTPGet.Port.StrVal != "" && !ok {
					return probe.HTTPGet.Port.StrVal, true
				}
			}
		}
	}
	return "", false
}

func podsForDeployment(deployment appsv1.Deployment, pods []corev1.Pod, limit int) []corev1.Pod {
	var matched []corev1.Pod
	for _, pod := range pods {
		if pod.Namespace != deployment.Namespace {
			continue
		}
		if selectorMatches(deployment.Spec.Selector.MatchLabels, pod.Labels) {
			matched = append(matched, pod)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		return matched[i].Name < matched[j].Name
	})
	if len(matched) > limit {
		return matched[:limit]
	}
	return matched
}

func selectorMatches(selector, labels map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for key, expected := range selector {
		if labels[key] != expected {
			return false
		}
	}
	return true
}

func resolveOwnerDeployment(pod corev1.Pod, replicaSetsByName map[string]appsv1.ReplicaSet) *metav1.OwnerReference {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind != "ReplicaSet" {
			continue
		}
		rs, ok := replicaSetsByName[pod.Namespace+"/"+owner.Name]
		if !ok {
			continue
		}
		for _, rsOwner := range rs.OwnerReferences {
			if rsOwner.Kind == "Deployment" {
				return &rsOwner
			}
		}
	}
	return nil
}

func ownerReferenceSummaries(owners []metav1.OwnerReference) []map[string]any {
	summaries := make([]map[string]any, 0, len(owners))
	for _, owner := range owners {
		summaries = append(summaries, map[string]any{
			"apiVersion": owner.APIVersion,
			"kind":       owner.Kind,
			"name":       owner.Name,
			"uid":        string(owner.UID),
		})
	}
	return summaries
}

func probeSummariesForContainers(containers []corev1.Container) []map[string]any {
	var summaries []map[string]any
	for _, container := range containers {
		summaries = append(summaries, map[string]any{
			"container": container.Name,
			"readiness": probeSummary(container.ReadinessProbe),
			"liveness":  probeSummary(container.LivenessProbe),
			"ports":     containerPorts(container.Ports),
		})
	}
	return summaries
}

func probeSummary(probe *corev1.Probe) map[string]any {
	if probe == nil {
		return nil
	}
	summary := map[string]any{
		"initialDelaySeconds": probe.InitialDelaySeconds,
		"periodSeconds":       probe.PeriodSeconds,
	}
	if probe.HTTPGet != nil {
		summary["httpGet"] = map[string]any{
			"path": probe.HTTPGet.Path,
			"port": probe.HTTPGet.Port.String(),
		}
	}
	return summary
}

func containerPorts(ports []corev1.ContainerPort) []map[string]any {
	var summaries []map[string]any
	for _, port := range ports {
		summaries = append(summaries, map[string]any{
			"name": port.Name,
			"port": port.ContainerPort,
		})
	}
	return summaries
}

func resourceSummariesForContainers(containers []corev1.Container) []map[string]any {
	var summaries []map[string]any
	for _, container := range containers {
		summaries = append(summaries, map[string]any{
			"container": container.Name,
			"requests":  fmt.Sprintf("%v", container.Resources.Requests),
			"limits":    fmt.Sprintf("%v", container.Resources.Limits),
		})
	}
	return summaries
}

func relatedPodSummaries(pods []corev1.Pod, metrics map[string]map[string]metricsv1beta1.ContainerMetrics) []map[string]any {
	var summaries []map[string]any
	for _, pod := range pods {
		item := map[string]any{
			"namespace": pod.Namespace,
			"name":      pod.Name,
			"phase":     string(pod.Status.Phase),
			"nodeName":  pod.Spec.NodeName,
		}
		if podMetric, ok := metrics[pod.Namespace+"/"+pod.Name]; ok {
			item["metrics"] = metricSetSummary(podMetric)
		}
		summaries = append(summaries, item)
	}
	return summaries
}

func metricsSummary(pods []corev1.Pod, metrics map[string]map[string]metricsv1beta1.ContainerMetrics) []map[string]any {
	var summaries []map[string]any
	for _, pod := range pods {
		podMetric, ok := metrics[pod.Namespace+"/"+pod.Name]
		if !ok {
			continue
		}
		summaries = append(summaries, map[string]any{
			"namespace":  pod.Namespace,
			"pod":        pod.Name,
			"containers": metricSetSummary(podMetric),
		})
	}
	return summaries
}

func metricSetSummary(metrics map[string]metricsv1beta1.ContainerMetrics) []map[string]any {
	var items []map[string]any
	for name, metric := range metrics {
		items = append(items, map[string]any{
			"container":     name,
			"cpuMillicores": metric.Usage.Cpu().MilliValue(),
			"memoryBytes":   metric.Usage.Memory().Value(),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i]["container"].(string) < items[j]["container"].(string)
	})
	return items
}

func warningEventSummaries(candidate FindingCandidate, relatedPods []corev1.Pod, events []corev1.Event, limit int) []map[string]any {
	relatedUIDs := map[string]struct{}{candidate.ResourceUID: {}}
	for _, pod := range relatedPods {
		relatedUIDs[string(pod.UID)] = struct{}{}
	}

	var summaries []map[string]any
	for _, event := range events {
		if event.Type != corev1.EventTypeWarning {
			continue
		}
		if _, ok := relatedUIDs[string(event.InvolvedObject.UID)]; !ok {
			continue
		}
		summaries = append(summaries, map[string]any{
			"namespace": event.Namespace,
			"kind":      event.InvolvedObject.Kind,
			"name":      event.InvolvedObject.Name,
			"reason":    event.Reason,
			"message":   event.Message,
			"count":     event.Count,
			"firstSeen": event.FirstTimestamp.UTC().Format(time.RFC3339),
			"lastSeen":  event.LastTimestamp.UTC().Format(time.RFC3339),
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return fmt.Sprintf("%v", summaries[i]["lastSeen"]) > fmt.Sprintf("%v", summaries[j]["lastSeen"])
	})
	if len(summaries) > limit {
		return summaries[:limit]
	}
	return summaries
}

func containerStateSummary(state corev1.ContainerState) string {
	switch {
	case state.Running != nil:
		return "Running"
	case state.Waiting != nil:
		return "Waiting:" + state.Waiting.Reason
	case state.Terminated != nil:
		return "Terminated:" + state.Terminated.Reason
	default:
		return "Unknown"
	}
}

func decodeRawJSON(payload json.RawMessage) any {
	if len(payload) == 0 {
		return map[string]any{}
	}
	var value any
	if err := json.Unmarshal(payload, &value); err != nil {
		return map[string]any{}
	}
	return value
}

func truncateNestedStrings(value any, maxLen int) {
	switch typed := value.(type) {
	case map[string]any:
		for key, item := range typed {
			if text, ok := item.(string); ok && len(text) > maxLen {
				typed[key] = text[:maxLen] + "..."
				continue
			}
			truncateNestedStrings(item, maxLen)
		}
	case []map[string]any:
		for _, item := range typed {
			truncateNestedStrings(item, maxLen)
		}
	case []any:
		for _, item := range typed {
			truncateNestedStrings(item, maxLen)
		}
	}
}
