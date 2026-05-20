package evidence

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"orbit/internal/config"
	"orbit/internal/store"
)

type DataStore interface {
	GetFinding(ctx context.Context, findingID string) (store.Finding, error)
	GetEvidencePackByFindingID(ctx context.Context, findingID string) (store.EvidencePack, error)
	CreateEvidencePack(ctx context.Context, params store.CreateEvidencePackParams) (store.EvidencePack, error)
	CreateAdhocEvidencePack(ctx context.Context, params store.CreateEvidencePackParams) (store.EvidencePack, error)
	GetCluster(ctx context.Context, clusterID string) (store.Cluster, error)
	GetKubernetesResource(ctx context.Context, resourceID string) (store.KubernetesResource, error)
	FindResourceByKindNamespaceName(ctx context.Context, clusterID, kind, namespace, name string) (store.KubernetesResource, error)
	ListResourcesByNamespace(ctx context.Context, clusterID, namespace string) ([]store.KubernetesResource, error)
	ListEventsByNamespace(ctx context.Context, clusterID, namespace string) ([]store.KubernetesEvent, error)
	ListEventsForResource(ctx context.Context, clusterID, resourceUID string) ([]store.KubernetesEvent, error)
	ListFindingsByNamespace(ctx context.Context, clusterID, namespace string) ([]store.Finding, error)
	ListFindingsForResource(ctx context.Context, resourceID string) ([]store.Finding, error)
	ListLatestResourceMetrics(ctx context.Context, clusterID, namespace, podName string) ([]store.ResourceMetric, error)
}

type Service struct {
	cfg   config.Config
	store DataStore
	now   func() time.Time
	newID func() string
}

func NewService(cfg config.Config, dataStore DataStore) *Service {
	return &Service{
		cfg:   cfg,
		store: dataStore,
		now: func() time.Time {
			return time.Now().UTC()
		},
		newID: func() string {
			return storeID()
		},
	}
}

func (s *Service) BuildFindingEvidencePack(ctx context.Context, findingID string, persist bool) (store.EvidencePack, error) {
	finding, err := s.store.GetFinding(ctx, findingID)
	if err != nil {
		return store.EvidencePack{}, err
	}
	cluster, err := s.store.GetCluster(ctx, finding.ClusterID)
	if err != nil {
		return store.EvidencePack{}, err
	}

	resource := store.KubernetesResource{}
	if finding.ResourceID != "" {
		resource, err = s.store.GetKubernetesResource(ctx, finding.ResourceID)
		if err != nil {
			return store.EvidencePack{}, err
		}
	}

	events := []store.KubernetesEvent{}
	if resource.UID != "" {
		events, err = s.store.ListEventsForResource(ctx, finding.ClusterID, resource.UID)
		if err != nil {
			return store.EvidencePack{}, err
		}
	}

	namespaceResources := []store.KubernetesResource{}
	namespaceMetrics := []store.ResourceMetric{}
	if resource.Namespace != "" {
		namespaceResources, err = s.store.ListResourcesByNamespace(ctx, finding.ClusterID, resource.Namespace)
		if err != nil {
			return store.EvidencePack{}, err
		}
		namespaceMetrics, err = s.store.ListLatestResourceMetrics(ctx, finding.ClusterID, resource.Namespace, "")
		if err != nil {
			return store.EvidencePack{}, err
		}
	}

	pack, err := s.buildFindingPack(finding, cluster, resource, namespaceResources, events, namespaceMetrics)
	if err != nil {
		return store.EvidencePack{}, err
	}
	return s.persistPack(ctx, persist, store.CreateEvidencePackParams{
		ID:             s.newID(),
		FindingID:      finding.ID,
		ClusterID:      finding.ClusterID,
		ResourceID:     finding.ResourceID,
		ScopeType:      "finding",
		ScopeNamespace: resource.Namespace,
		ScopeName:      resource.Name,
		TokenEstimate:  pack.TokenEstimate,
		PackJSON:       pack.PackJSON,
		CreatedAt:      pack.CreatedAt,
	})
}

func (s *Service) BuildNamespaceEvidencePack(ctx context.Context, clusterID, namespace string, persist bool) (store.EvidencePack, error) {
	cluster, err := s.store.GetCluster(ctx, clusterID)
	if err != nil {
		return store.EvidencePack{}, err
	}
	resources, err := s.store.ListResourcesByNamespace(ctx, clusterID, namespace)
	if err != nil {
		return store.EvidencePack{}, err
	}
	events, err := s.store.ListEventsByNamespace(ctx, clusterID, namespace)
	if err != nil {
		return store.EvidencePack{}, err
	}
	findings, err := s.store.ListFindingsByNamespace(ctx, clusterID, namespace)
	if err != nil {
		return store.EvidencePack{}, err
	}
	metrics, err := s.store.ListLatestResourceMetrics(ctx, clusterID, namespace, "")
	if err != nil {
		return store.EvidencePack{}, err
	}

	pack, err := s.buildNamespacePack(cluster, namespace, resources, events, findings, metrics)
	if err != nil {
		return store.EvidencePack{}, err
	}
	return s.persistPack(ctx, persist, store.CreateEvidencePackParams{
		ID:             s.newID(),
		ClusterID:      clusterID,
		ScopeType:      "namespace",
		ScopeNamespace: namespace,
		TokenEstimate:  pack.TokenEstimate,
		PackJSON:       pack.PackJSON,
		CreatedAt:      pack.CreatedAt,
	})
}

func (s *Service) BuildPodEvidencePack(ctx context.Context, clusterID, namespace, podName string, persist bool) (store.EvidencePack, error) {
	cluster, err := s.store.GetCluster(ctx, clusterID)
	if err != nil {
		return store.EvidencePack{}, err
	}
	resource, err := s.store.FindResourceByKindNamespaceName(ctx, clusterID, "Pod", namespace, podName)
	if err != nil {
		return store.EvidencePack{}, err
	}
	namespaceResources, err := s.store.ListResourcesByNamespace(ctx, clusterID, namespace)
	if err != nil {
		return store.EvidencePack{}, err
	}
	events, err := s.store.ListEventsForResource(ctx, clusterID, resource.UID)
	if err != nil {
		return store.EvidencePack{}, err
	}
	findings, err := s.store.ListFindingsForResource(ctx, resource.ID)
	if err != nil {
		return store.EvidencePack{}, err
	}
	metrics, err := s.store.ListLatestResourceMetrics(ctx, clusterID, namespace, podName)
	if err != nil {
		return store.EvidencePack{}, err
	}

	pack, err := s.buildPodPack(cluster, resource, namespaceResources, events, findings, metrics)
	if err != nil {
		return store.EvidencePack{}, err
	}
	return s.persistPack(ctx, persist, store.CreateEvidencePackParams{
		ID:             s.newID(),
		ClusterID:      clusterID,
		ResourceID:     resource.ID,
		ScopeType:      "pod",
		ScopeNamespace: namespace,
		ScopeName:      podName,
		TokenEstimate:  pack.TokenEstimate,
		PackJSON:       pack.PackJSON,
		CreatedAt:      pack.CreatedAt,
	})
}

func (s *Service) buildFindingPack(
	finding store.Finding,
	cluster store.Cluster,
	resource store.KubernetesResource,
	namespaceResources []store.KubernetesResource,
	events []store.KubernetesEvent,
	metrics []store.ResourceMetric,
) (store.EvidencePack, error) {
	pack := map[string]any{
		"scope": map[string]any{
			"type": "finding",
		},
		"cluster": clusterSummary(cluster),
		"finding": map[string]any{
			"id":          finding.ID,
			"severity":    finding.Severity,
			"category":    finding.Category,
			"title":       finding.Title,
			"description": finding.Description,
			"status":      finding.Status,
		},
		"generatedAt": s.now().Format(time.RFC3339),
	}

	if resource.ID != "" {
		pack["affectedResource"] = resourceSummary(resource, true)
	}

	var relatedPods []corev1.Pod
	var probeSummaries []map[string]any
	var resourceSummaries []map[string]any
	var containerStatuses []map[string]any
	var warningEvents []map[string]any
	var metricsSummary []map[string]any
	var causes []string
	var ownerRefs []map[string]any

	deploymentsByName := map[string]appsv1.Deployment{}
	replicaSetsByName := map[string]appsv1.ReplicaSet{}
	podsByName := map[string]corev1.Pod{}
	for _, item := range namespaceResources {
		switch item.Kind {
		case "Deployment":
			if deployment, ok := decodeDeployment(item.RawJSON); ok {
				deploymentsByName[item.Name] = deployment
			}
		case "ReplicaSet":
			if replicaSet, ok := decodeReplicaSet(item.RawJSON); ok {
				replicaSetsByName[item.Name] = replicaSet
			}
		case "Pod":
			if pod, ok := decodePod(item.RawJSON); ok {
				podsByName[item.Name] = pod
			}
		}
	}

	switch resource.Kind {
	case "Deployment":
		if deployment, ok := decodeDeployment(resource.RawJSON); ok {
			ownerRefs = ownerReferenceSummaries(deployment.OwnerReferences)
			relatedPods = podsForDeployment(deployment, namespaceResources, s.cfg.EvidenceMaxRelatedResources)
			probeSummaries = probeSummariesForContainers(deployment.Spec.Template.Spec.Containers)
			resourceSummaries = resourceSummariesForContainers(deployment.Spec.Template.Spec.Containers)
			causes = append(causes, causesFromFinding(finding)...)
		}
	case "Pod":
		if pod, ok := decodePod(resource.RawJSON); ok {
			relatedPods = []corev1.Pod{pod}
			ownerRefs = ownerReferenceSummaries(pod.OwnerReferences)
			probeSummaries = probeSummariesForContainers(pod.Spec.Containers)
			resourceSummaries = resourceSummariesForContainers(pod.Spec.Containers)
			if owner := resolveOwnerDeployment(pod, replicaSetsByName); owner != nil {
				pack["ownerDeployment"] = map[string]any{
					"namespace": namespaceOrValue(pod.Namespace, resource.Namespace),
					"name":      owner.Name,
				}
			}
			causes = append(causes, podSuspectedCauses(pod)...)
		}
	}

	for _, pod := range relatedPods {
		containerStatuses = append(containerStatuses, containerStatusesForPod(pod)...)
	}
	warningEvents = warningEventSummariesForFinding(finding, resource.UID, relatedPods, events, s.cfg.EvidenceMaxEvents)
	metricsSummary = metricsSummaryForPods(relatedPods, metrics)
	if len(causes) == 0 {
		causes = causesFromFinding(finding)
	}

	pack["ownerReferences"] = ownerRefs
	pack["relatedPods"] = relatedPodSummaries(relatedPods)
	pack["containerStatuses"] = containerStatuses
	pack["probeSummary"] = probeSummaries
	pack["resourceSummary"] = resourceSummaries
	pack["recentWarningEvents"] = warningEvents
	if len(metricsSummary) > 0 {
		pack["metricsSummary"] = metricsSummary
	}
	pack["suspectedDeterministicCauses"] = uniqueStrings(causes)

	trimmed, estimate := enforceFindingTokenBudget(pack, s.cfg)
	trimmed["tokenEstimate"] = estimate
	payload, err := json.Marshal(trimmed)
	if err != nil {
		return store.EvidencePack{}, err
	}
	return store.EvidencePack{
		FindingID:      finding.ID,
		ClusterID:      cluster.ID,
		ResourceID:     resource.ID,
		ScopeType:      "finding",
		ScopeNamespace: resource.Namespace,
		ScopeName:      resource.Name,
		TokenEstimate:  estimate,
		PackJSON:       payload,
		CreatedAt:      s.now(),
	}, nil
}

func (s *Service) buildNamespacePack(
	cluster store.Cluster,
	namespace string,
	resources []store.KubernetesResource,
	events []store.KubernetesEvent,
	findings []store.Finding,
	metrics []store.ResourceMetric,
) (store.EvidencePack, error) {
	var pods []corev1.Pod
	var deployments []appsv1.Deployment
	counts := map[string]int{}
	for _, resource := range resources {
		counts[resource.Kind]++
		switch resource.Kind {
		case "Pod":
			if pod, ok := decodePod(resource.RawJSON); ok {
				pods = append(pods, pod)
			}
		case "Deployment":
			if deployment, ok := decodeDeployment(resource.RawJSON); ok {
				deployments = append(deployments, deployment)
			}
		}
	}

	unhealthyPods := namespaceUnhealthyPodSummaries(pods)
	unavailableDeployments := namespaceUnavailableDeploymentSummaries(deployments)
	warningEvents := namespaceWarningEvents(events, s.cfg.EvidenceMaxEvents)
	openFindings := findingSummaries(findings)
	topRestartingPods := topRestartHeavyPods(pods, metrics, s.cfg.EvidenceMaxRelatedResources)
	causes := namespaceSuspectedCauses(unhealthyPods, unavailableDeployments, openFindings)

	pack := map[string]any{
		"scope": map[string]any{
			"type":      "namespace",
			"namespace": namespace,
		},
		"cluster":     clusterSummary(cluster),
		"generatedAt": s.now().Format(time.RFC3339),
		"summary": map[string]any{
			"pods":          counts["Pod"],
			"deployments":   counts["Deployment"],
			"services":      counts["Service"],
			"configmaps":    counts["ConfigMap"],
			"warningEvents": len(warningEvents),
			"openFindings":  len(openFindings),
		},
		"unhealthyPods":                unhealthyPods,
		"unavailableDeployments":       unavailableDeployments,
		"recentWarningEvents":          warningEvents,
		"topRestartHeavyPods":          topRestartingPods,
		"relatedOpenFindings":          openFindings,
		"suspectedDeterministicCauses": causes,
	}

	trimmed, estimate := enforceNamespaceTokenBudget(pack, s.cfg)
	trimmed["tokenEstimate"] = estimate
	payload, err := json.Marshal(trimmed)
	if err != nil {
		return store.EvidencePack{}, err
	}
	return store.EvidencePack{
		ClusterID:      cluster.ID,
		ScopeType:      "namespace",
		ScopeNamespace: namespace,
		TokenEstimate:  estimate,
		PackJSON:       payload,
		CreatedAt:      s.now(),
	}, nil
}

func (s *Service) buildPodPack(
	cluster store.Cluster,
	resource store.KubernetesResource,
	resources []store.KubernetesResource,
	events []store.KubernetesEvent,
	findings []store.Finding,
	metrics []store.ResourceMetric,
) (store.EvidencePack, error) {
	pod, ok := decodePod(resource.RawJSON)
	if !ok {
		return store.EvidencePack{}, fmt.Errorf("resource %s is not a pod", resource.Name)
	}

	replicaSetsByName := map[string]appsv1.ReplicaSet{}
	services := []corev1.Service{}
	deployments := map[string]appsv1.Deployment{}
	for _, item := range resources {
		switch item.Kind {
		case "ReplicaSet":
			if replicaSet, ok := decodeReplicaSet(item.RawJSON); ok {
				replicaSetsByName[item.Name] = replicaSet
			}
		case "Service":
			if service, ok := decodeService(item.RawJSON); ok {
				services = append(services, service)
			}
		case "Deployment":
			if deployment, ok := decodeDeployment(item.RawJSON); ok {
				deployments[item.Name] = deployment
			}
		}
	}

	relatedFindings := findingSummaries(findings)
	relatedServices := relatedServiceSummaries(pod, services)
	metricsSummary := metricsSummaryForPod(metrics)
	ownerRefs := ownerReferenceSummaries(pod.OwnerReferences)
	probeSummary := probeSummariesForContainers(pod.Spec.Containers)
	resourceSummary := resourceSummariesForContainers(pod.Spec.Containers)
	containerStatuses := containerStatusesForPod(pod)
	causes := podSuspectedCauses(pod)
	warnings := warningEventSummaries(events, s.cfg.EvidenceMaxEvents)

	pack := map[string]any{
		"scope": map[string]any{
			"type":      "pod",
			"namespace": resource.Namespace,
			"name":      resource.Name,
		},
		"cluster":     clusterSummary(cluster),
		"generatedAt": s.now().Format(time.RFC3339),
		"pod": map[string]any{
			"namespace":    resource.Namespace,
			"name":         resource.Name,
			"phase":        string(pod.Status.Phase),
			"ready":        podReady(pod),
			"nodeName":     pod.Spec.NodeName,
			"restartCount": totalRestartCount(pod),
		},
		"ownerReferences":              ownerRefs,
		"relatedServiceCandidates":     relatedServices,
		"recentWarningEvents":          warnings,
		"containerStatuses":            containerStatuses,
		"probeSummary":                 probeSummary,
		"resourceSummary":              resourceSummary,
		"relatedOpenFindings":          relatedFindings,
		"suspectedDeterministicCauses": causes,
	}
	if len(metricsSummary) > 0 {
		pack["metricsSummary"] = metricsSummary
	}
	if owner := resolveOwnerDeployment(pod, replicaSetsByName); owner != nil {
		pack["relatedDeployment"] = deploymentSummaryFromName(resource.Namespace, owner.Name, deployments)
	}

	trimmed, estimate := enforcePodTokenBudget(pack, s.cfg)
	trimmed["tokenEstimate"] = estimate
	payload, err := json.Marshal(trimmed)
	if err != nil {
		return store.EvidencePack{}, err
	}
	return store.EvidencePack{
		ClusterID:      cluster.ID,
		ResourceID:     resource.ID,
		ScopeType:      "pod",
		ScopeNamespace: resource.Namespace,
		ScopeName:      resource.Name,
		TokenEstimate:  estimate,
		PackJSON:       payload,
		CreatedAt:      s.now(),
	}, nil
}

func (s *Service) persistPack(ctx context.Context, persist bool, params store.CreateEvidencePackParams) (store.EvidencePack, error) {
	if persist {
		if params.ScopeType == "finding" {
			return s.store.CreateEvidencePack(ctx, params)
		}
		return s.store.CreateAdhocEvidencePack(ctx, params)
	}
	return store.EvidencePack{
		ID:             params.ID,
		FindingID:      params.FindingID,
		ClusterID:      params.ClusterID,
		ResourceID:     params.ResourceID,
		ScopeType:      params.ScopeType,
		ScopeNamespace: params.ScopeNamespace,
		ScopeName:      params.ScopeName,
		TokenEstimate:  params.TokenEstimate,
		PackJSON:       params.PackJSON,
		CreatedAt:      params.CreatedAt,
	}, nil
}

func clusterSummary(cluster store.Cluster) map[string]any {
	return map[string]any{
		"id":   cluster.ID,
		"name": cluster.Name,
		"type": cluster.Type,
		"mode": cluster.Mode,
	}
}

func resourceSummary(resource store.KubernetesResource, includeRaw bool) map[string]any {
	summary := map[string]any{
		"kind":        resource.Kind,
		"namespace":   resource.Namespace,
		"name":        resource.Name,
		"uid":         resource.UID,
		"status":      resource.Status,
		"labels":      resource.Labels,
		"annotations": resource.Annotations,
	}
	if includeRaw {
		summary["raw_json"] = decodeRawJSON(resource.RawJSON)
	}
	return summary
}

func causesFromFinding(finding store.Finding) []string {
	switch finding.Title {
	case "Deployment has zero replicas":
		return []string{"The deployment spec explicitly sets replicas to zero.", "The workload may have been scaled down intentionally or by automation."}
	case "Deployment has unavailable replicas":
		return []string{"Pods may be failing readiness checks.", "The rollout may be blocked by scheduling or image issues.", "Recent warning events may explain why replicas are unavailable."}
	case "Pod is not healthy":
		return []string{"The pod may be Pending, Failed, or otherwise blocked from becoming healthy.", "Recent events and container state transitions usually explain the unhealthy pod phase."}
	case "Pod has container restarts":
		return []string{"A container may be crashing repeatedly.", "Resource pressure or failing probes can drive restart loops."}
	case "Probe may target wrong port":
		return []string{"The readiness or liveness probe refers to a port that the container does not expose.", "A named probe port may not match the container port names."}
	default:
		return []string{"Manual investigation is required."}
	}
}

func namespaceSuspectedCauses(unhealthyPods, unavailableDeployments, findings []map[string]any) []string {
	var causes []string
	if len(unhealthyPods) > 1 {
		causes = append(causes, "Multiple unhealthy pods were detected in the namespace and may share a common rollout, dependency, or capacity issue.")
	}
	if len(unavailableDeployments) > 0 {
		causes = append(causes, "One or more deployments are not meeting desired availability and should be checked together with warning events.")
	}
	if len(findings) > 0 {
		causes = append(causes, "Open findings already exist in the namespace and may explain the unhealthy resource counts.")
	}
	if len(causes) == 0 {
		causes = append(causes, "No deterministic namespace-level cause stood out; review the summarized warnings and workload health signals.")
	}
	return uniqueStrings(causes)
}

func podSuspectedCauses(pod corev1.Pod) []string {
	var causes []string
	if pod.Status.Phase != corev1.PodRunning && pod.Status.Phase != corev1.PodSucceeded {
		causes = append(causes, "The pod phase indicates it is not currently healthy.")
	}
	if totalRestartCount(pod) > 0 {
		causes = append(causes, "A restarted container suggests crashes, resource pressure, or repeated probe failures.")
	}
	if suspect, ok := podProbePortMismatch(pod); ok {
		causes = append(causes, "A probe targets port "+suspect+" which does not match the declared container ports.")
	}
	if len(causes) == 0 {
		causes = append(causes, "Manual pod-level investigation is required.")
	}
	return uniqueStrings(causes)
}

func namespaceUnhealthyPodSummaries(pods []corev1.Pod) []map[string]any {
	var summaries []map[string]any
	for _, pod := range pods {
		if pod.Status.Phase == corev1.PodRunning && totalRestartCount(pod) == 0 {
			continue
		}
		summaries = append(summaries, map[string]any{
			"name":          pod.Name,
			"phase":         string(pod.Status.Phase),
			"ready":         podReady(pod),
			"restartCount":  totalRestartCount(pod),
			"waitingReason": firstWaitingReason(pod),
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return fmt.Sprintf("%s/%s", summaries[i]["phase"], summaries[i]["name"]) < fmt.Sprintf("%s/%s", summaries[j]["phase"], summaries[j]["name"])
	})
	return summaries
}

func namespaceUnavailableDeploymentSummaries(deployments []appsv1.Deployment) []map[string]any {
	var summaries []map[string]any
	for _, deployment := range deployments {
		desired := int32(1)
		if deployment.Spec.Replicas != nil {
			desired = *deployment.Spec.Replicas
		}
		if deployment.Status.AvailableReplicas >= desired {
			continue
		}
		summaries = append(summaries, map[string]any{
			"name":              deployment.Name,
			"desiredReplicas":   desired,
			"availableReplicas": deployment.Status.AvailableReplicas,
			"updatedReplicas":   deployment.Status.UpdatedReplicas,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return fmt.Sprintf("%s", summaries[i]["name"]) < fmt.Sprintf("%s", summaries[j]["name"])
	})
	return summaries
}

func topRestartHeavyPods(pods []corev1.Pod, metrics []store.ResourceMetric, limit int) []map[string]any {
	type podScore struct {
		pod   corev1.Pod
		score int32
	}
	var scored []podScore
	for _, pod := range pods {
		score := totalRestartCount(pod)
		if score == 0 && pod.Status.Phase == corev1.PodRunning {
			continue
		}
		scored = append(scored, podScore{pod: pod, score: score})
	}
	sort.Slice(scored, func(i, j int) bool {
		if scored[i].score == scored[j].score {
			return scored[i].pod.Name < scored[j].pod.Name
		}
		return scored[i].score > scored[j].score
	})
	if len(scored) > limit {
		scored = scored[:limit]
	}
	var summaries []map[string]any
	for _, item := range scored {
		summary := map[string]any{
			"name":         item.pod.Name,
			"phase":        string(item.pod.Status.Phase),
			"restartCount": item.score,
		}
		if podMetrics := metricsSummaryForPods([]corev1.Pod{item.pod}, metrics); len(podMetrics) > 0 {
			summary["metrics"] = podMetrics[0]
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

func findingSummaries(findings []store.Finding) []map[string]any {
	var summaries []map[string]any
	for _, finding := range findings {
		if finding.Status != "" && finding.Status != "open" {
			continue
		}
		summaries = append(summaries, map[string]any{
			"id":        finding.ID,
			"severity":  finding.Severity,
			"category":  finding.Category,
			"title":     finding.Title,
			"status":    finding.Status,
			"kind":      finding.ResourceKind,
			"namespace": finding.ResourceNamespace,
			"name":      finding.ResourceName,
		})
	}
	return summaries
}

func relatedPodSummaries(pods []corev1.Pod) []map[string]any {
	var summaries []map[string]any
	for _, pod := range pods {
		summaries = append(summaries, map[string]any{
			"namespace": pod.Namespace,
			"name":      pod.Name,
			"phase":     string(pod.Status.Phase),
			"nodeName":  pod.Spec.NodeName,
			"ready":     podReady(pod),
		})
	}
	return summaries
}

func containerStatusesForPod(pod corev1.Pod) []map[string]any {
	var statuses []map[string]any
	for _, status := range pod.Status.ContainerStatuses {
		statuses = append(statuses, map[string]any{
			"container":        status.Name,
			"ready":            status.Ready,
			"restartCount":     status.RestartCount,
			"waitingReason":    waitingReason(status.State),
			"terminatedReason": terminatedReason(status.State),
			"state":            containerStateSummary(status.State),
		})
	}
	return statuses
}

func warningEventSummariesForFinding(finding store.Finding, resourceUID string, relatedPods []corev1.Pod, events []store.KubernetesEvent, limit int) []map[string]any {
	relatedUIDs := map[string]struct{}{}
	if resourceUID != "" {
		relatedUIDs[resourceUID] = struct{}{}
	}
	for _, pod := range relatedPods {
		relatedUIDs[string(pod.UID)] = struct{}{}
	}
	var filtered []store.KubernetesEvent
	for _, event := range events {
		if event.Type != corev1.EventTypeWarning {
			continue
		}
		if _, ok := relatedUIDs[event.InvolvedUID]; !ok {
			continue
		}
		filtered = append(filtered, event)
	}
	_ = finding
	return warningEventSummaries(filtered, limit)
}

func warningEventSummaries(events []store.KubernetesEvent, limit int) []map[string]any {
	sort.Slice(events, func(i, j int) bool {
		left := events[i].ObservedAt
		if events[i].LastSeenAt.Valid {
			left = events[i].LastSeenAt.Time
		}
		right := events[j].ObservedAt
		if events[j].LastSeenAt.Valid {
			right = events[j].LastSeenAt.Time
		}
		return left.After(right)
	})
	if len(events) > limit {
		events = events[:limit]
	}
	var summaries []map[string]any
	for _, event := range events {
		summary := map[string]any{
			"namespace": event.Namespace,
			"kind":      event.InvolvedKind,
			"name":      event.InvolvedName,
			"reason":    event.Reason,
			"message":   event.Message,
			"count":     event.Count,
		}
		if event.FirstSeenAt.Valid {
			summary["firstSeen"] = event.FirstSeenAt.Time.UTC().Format(time.RFC3339)
		}
		if event.LastSeenAt.Valid {
			summary["lastSeen"] = event.LastSeenAt.Time.UTC().Format(time.RFC3339)
		}
		summaries = append(summaries, summary)
	}
	return summaries
}

func namespaceWarningEvents(events []store.KubernetesEvent, limit int) []map[string]any {
	var filtered []store.KubernetesEvent
	for _, event := range events {
		if event.Type == corev1.EventTypeWarning {
			filtered = append(filtered, event)
		}
	}
	return warningEventSummaries(filtered, limit)
}

func metricsSummaryForPods(pods []corev1.Pod, metrics []store.ResourceMetric) []map[string]any {
	allowed := map[string]struct{}{}
	for _, pod := range pods {
		allowed[pod.Namespace+"/"+pod.Name] = struct{}{}
	}
	var summaries []map[string]any
	for _, metric := range metrics {
		key := metric.Namespace + "/" + metric.PodName
		if _, ok := allowed[key]; !ok {
			continue
		}
		summaries = append(summaries, map[string]any{
			"namespace":     metric.Namespace,
			"pod":           metric.PodName,
			"container":     metric.ContainerName,
			"cpuMillicores": metric.CPUMillicores,
			"memoryBytes":   metric.MemoryBytes,
		})
	}
	return summaries
}

func metricsSummaryForPod(metrics []store.ResourceMetric) []map[string]any {
	var summaries []map[string]any
	for _, metric := range metrics {
		summaries = append(summaries, map[string]any{
			"container":     metric.ContainerName,
			"cpuMillicores": metric.CPUMillicores,
			"memoryBytes":   metric.MemoryBytes,
		})
	}
	return summaries
}

func deploymentSummaryFromName(namespace, name string, deployments map[string]appsv1.Deployment) map[string]any {
	deployment, ok := deployments[name]
	if !ok {
		return map[string]any{
			"namespace": namespace,
			"name":      name,
		}
	}
	desired := int32(1)
	if deployment.Spec.Replicas != nil {
		desired = *deployment.Spec.Replicas
	}
	return map[string]any{
		"namespace":         namespace,
		"name":              name,
		"desiredReplicas":   desired,
		"availableReplicas": deployment.Status.AvailableReplicas,
		"updatedReplicas":   deployment.Status.UpdatedReplicas,
	}
}

func relatedServiceSummaries(pod corev1.Pod, services []corev1.Service) []map[string]any {
	var summaries []map[string]any
	for _, service := range services {
		if len(service.Spec.Selector) == 0 {
			continue
		}
		if !selectorMatches(service.Spec.Selector, pod.Labels) {
			continue
		}
		summaries = append(summaries, map[string]any{
			"name":     service.Name,
			"type":     string(service.Spec.Type),
			"selector": service.Spec.Selector,
		})
	}
	return summaries
}

func enforceFindingTokenBudget(pack map[string]any, cfg config.Config) (map[string]any, int) {
	estimate := tokenEstimate(pack)
	if estimate <= cfg.EvidenceMaxTokenEstimate {
		return pack, estimate
	}
	if affected, ok := pack["affectedResource"].(map[string]any); ok {
		delete(affected, "raw_json")
	}
	return trimPack(pack, cfg, []string{"recentWarningEvents", "relatedPods", "metricsSummary"})
}

func enforceNamespaceTokenBudget(pack map[string]any, cfg config.Config) (map[string]any, int) {
	return trimPack(pack, cfg, []string{"recentWarningEvents", "topRestartHeavyPods", "unhealthyPods", "unavailableDeployments"})
}

func enforcePodTokenBudget(pack map[string]any, cfg config.Config) (map[string]any, int) {
	return trimPack(pack, cfg, []string{"relatedServiceCandidates", "recentWarningEvents", "metricsSummary"})
}

func trimPack(pack map[string]any, cfg config.Config, reductionOrder []string) (map[string]any, int) {
	estimate := tokenEstimate(pack)
	if estimate <= cfg.EvidenceMaxTokenEstimate {
		return pack, estimate
	}

	for _, key := range reductionOrder {
		switch items := pack[key].(type) {
		case []map[string]any:
			for len(items) > 1 && estimate > cfg.EvidenceMaxTokenEstimate {
				items = items[:len(items)-1]
				pack[key] = items
				estimate = tokenEstimate(pack)
			}
		case []any:
			for len(items) > 1 && estimate > cfg.EvidenceMaxTokenEstimate {
				items = items[:len(items)-1]
				pack[key] = items
				estimate = tokenEstimate(pack)
			}
		}
	}

	if estimate > cfg.EvidenceMaxTokenEstimate {
		truncateNestedStrings(pack, 120)
		estimate = tokenEstimate(pack)
	}
	if estimate > cfg.EvidenceMaxTokenEstimate {
		truncateNestedStrings(pack, 60)
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

func podsForDeployment(deployment appsv1.Deployment, resources []store.KubernetesResource, limit int) []corev1.Pod {
	var matched []corev1.Pod
	for _, resource := range resources {
		if resource.Kind != "Pod" {
			continue
		}
		pod, ok := decodePod(resource.RawJSON)
		if !ok {
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

func resolveOwnerDeployment(pod corev1.Pod, replicaSetsByName map[string]appsv1.ReplicaSet) *metav1.OwnerReference {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind != "ReplicaSet" {
			continue
		}
		rs, ok := replicaSetsByName[owner.Name]
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

func totalRestartCount(pod corev1.Pod) int32 {
	var total int32
	for _, status := range pod.Status.ContainerStatuses {
		total += status.RestartCount
	}
	return total
}

func podReady(pod corev1.Pod) bool {
	if len(pod.Status.ContainerStatuses) == 0 {
		return false
	}
	for _, status := range pod.Status.ContainerStatuses {
		if !status.Ready {
			return false
		}
	}
	return true
}

func firstWaitingReason(pod corev1.Pod) string {
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil {
			return status.State.Waiting.Reason
		}
	}
	return ""
}

func waitingReason(state corev1.ContainerState) string {
	if state.Waiting == nil {
		return ""
	}
	return state.Waiting.Reason
}

func terminatedReason(state corev1.ContainerState) string {
	if state.Terminated == nil {
		return ""
	}
	return state.Terminated.Reason
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

func namespaceOrValue(primary, fallback string) string {
	if primary != "" {
		return primary
	}
	return fallback
}

func uniqueStrings(items []string) []string {
	seen := map[string]struct{}{}
	var result []string
	for _, item := range items {
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
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

func decodeDeployment(payload json.RawMessage) (appsv1.Deployment, bool) {
	var deployment appsv1.Deployment
	if err := json.Unmarshal(payload, &deployment); err != nil {
		return appsv1.Deployment{}, false
	}
	return deployment, true
}

func decodeReplicaSet(payload json.RawMessage) (appsv1.ReplicaSet, bool) {
	var replicaSet appsv1.ReplicaSet
	if err := json.Unmarshal(payload, &replicaSet); err != nil {
		return appsv1.ReplicaSet{}, false
	}
	return replicaSet, true
}

func decodePod(payload json.RawMessage) (corev1.Pod, bool) {
	var pod corev1.Pod
	if err := json.Unmarshal(payload, &pod); err != nil {
		return corev1.Pod{}, false
	}
	return pod, true
}

func decodeService(payload json.RawMessage) (corev1.Service, bool) {
	var service corev1.Service
	if err := json.Unmarshal(payload, &service); err != nil {
		return corev1.Service{}, false
	}
	return service, true
}

func podProbePortMismatch(pod corev1.Pod) (string, bool) {
	return containersProbePortMismatch(pod.Spec.Containers)
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
			if probe == nil || probe.HTTPGet == nil {
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

func storeID() string {
	return uuid.NewString()
}

func PackText(payload json.RawMessage) string {
	return strings.ToLower(string(payload))
}
