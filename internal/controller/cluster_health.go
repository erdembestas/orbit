package controller

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"

	"orbit/internal/config"
	"orbit/internal/store"
)

type ClusterHealthComputation struct {
	ClusterSnapshot    store.CreateClusterHealthSnapshotParams
	NodeSnapshots      []store.CreateNodeHealthSnapshotParams
	NamespaceSnapshots []store.CreateNamespaceHealthSnapshotParams
}

func ComputeClusterHealth(cfg config.Config, clusterID string, observedAt time.Time, nodes []corev1.Node, pods []corev1.Pod, events []corev1.Event, nodeMetrics []metricsv1beta1.NodeMetrics, podMetrics []metricsv1beta1.PodMetrics, metricsAvailable bool, metricsError string) (ClusterHealthComputation, error) {
	nodeMetricByName := map[string]metricsv1beta1.NodeMetrics{}
	for _, metric := range nodeMetrics {
		nodeMetricByName[metric.Name] = metric
	}
	podMetricByKey := map[string]metricsv1beta1.PodMetrics{}
	for _, metric := range podMetrics {
		podMetricByKey[metric.Namespace+"/"+metric.Name] = metric
	}

	podsByNode := map[string][]corev1.Pod{}
	namespaceAgg := map[string]*namespaceAggregate{}
	runningPods := 0
	pendingPods := 0
	failedPods := 0
	warningEvents := 0

	for _, event := range events {
		if strings.EqualFold(event.Type, "Warning") {
			warningEvents++
			agg := ensureNamespaceAgg(namespaceAgg, event.Namespace)
			agg.WarningEventCount++
		}
	}

	for _, pod := range pods {
		namespace := pod.Namespace
		agg := ensureNamespaceAgg(namespaceAgg, namespace)
		agg.PodCount++
		switch pod.Status.Phase {
		case corev1.PodRunning:
			agg.RunningPodCount++
			runningPods++
		case corev1.PodPending:
			agg.PendingPodCount++
			pendingPods++
		case corev1.PodFailed:
			agg.FailedPodCount++
			failedPods++
		}
		agg.RestartCount += int(sumContainerRestarts(pod))
		if pod.Spec.NodeName != "" {
			podsByNode[pod.Spec.NodeName] = append(podsByNode[pod.Spec.NodeName], pod)
		}
		if metric, ok := podMetricByKey[pod.Namespace+"/"+pod.Name]; ok {
			cpu, memory := sumPodMetric(metric)
			agg.UsedCPUMillicores += cpu
			agg.UsedMemoryBytes += memory
		}
	}

	nodeSnapshots := make([]store.CreateNodeHealthSnapshotParams, 0, len(nodes))
	totalCPUCapacity := int64(0)
	totalCPUAllocatable := int64(0)
	totalCPUUsed := int64(0)
	totalMemoryCapacity := int64(0)
	totalMemoryAllocatable := int64(0)
	totalMemoryUsed := int64(0)
	readyNodes := 0
	notReadyNodes := 0
	clusterScore := 100
	clusterPressureCount := 0

	for _, node := range nodes {
		capacityCPU := node.Status.Capacity.Cpu().MilliValue()
		allocatableCPU := node.Status.Allocatable.Cpu().MilliValue()
		capacityMemory := node.Status.Capacity.Memory().Value()
		allocatableMemory := node.Status.Allocatable.Memory().Value()
		totalCPUCapacity += capacityCPU
		totalCPUAllocatable += allocatableCPU
		totalMemoryCapacity += capacityMemory
		totalMemoryAllocatable += allocatableMemory

		nodePods := podsByNode[node.Name]
		podCount, runningCount, pendingCount, failedCount := countPods(nodePods)

		ready := nodeReady(node)
		if ready {
			readyNodes++
		} else {
			notReadyNodes++
			clusterScore -= 30
		}

		conditions := nodeConditionMap(node)
		pressureFlags := pressureFlags(node)
		pressurePenalty := 0
		for _, key := range []string{"MemoryPressure", "DiskPressure", "PIDPressure"} {
			if pressureFlags[key] {
				pressurePenalty += 20
				clusterPressureCount++
			}
		}
		clusterScore -= pressurePenalty

		usedCPU := int64(0)
		usedMemory := int64(0)
		if metric, ok := nodeMetricByName[node.Name]; ok {
			usedCPU = metric.Usage.Cpu().MilliValue()
			usedMemory = metric.Usage.Memory().Value()
			totalCPUUsed += usedCPU
			totalMemoryUsed += usedMemory
		}

		nodeCPUPercent := percentString(usedCPU, allocatableCPU)
		nodeMemoryPercent := percentString(usedMemory, allocatableMemory)
		nodeScore := 100
		if !ready {
			nodeScore = minInt(nodeScore, 20)
		}
		nodeStatus := "healthy"
		if pressurePenalty > 0 {
			nodeScore -= pressurePenalty
		}
		if pct := percentFloat(usedCPU, allocatableCPU); pct > 0 {
			switch {
			case pct >= float64(cfg.NodeCPUCriticalPercent):
				nodeScore -= 25
			case pct >= float64(cfg.NodeCPUWarnPercent):
				nodeScore -= 10
			}
		}
		if pct := percentFloat(usedMemory, allocatableMemory); pct > 0 {
			switch {
			case pct >= float64(cfg.NodeMemoryCriticalPercent):
				nodeScore -= 25
			case pct >= float64(cfg.NodeMemoryWarnPercent):
				nodeScore -= 10
			}
		}
		if pendingCount > 0 {
			nodeScore -= minInt(10, pendingCount*3)
		}
		if failedCount > 0 {
			nodeScore -= minInt(20, failedCount*10)
		}
		nodeScore = clampScore(nodeScore)
		if !ready || hasCriticalPressure(pressureFlags) || percentOver(usedCPU, allocatableCPU, cfg.NodeCPUCriticalPercent) || percentOver(usedMemory, allocatableMemory, cfg.NodeMemoryCriticalPercent) {
			nodeStatus = "critical"
		} else if pressurePenalty > 0 || percentOver(usedCPU, allocatableCPU, cfg.NodeCPUWarnPercent) || percentOver(usedMemory, allocatableMemory, cfg.NodeMemoryWarnPercent) || pendingCount > 0 || failedCount > 0 {
			nodeStatus = "warning"
		}

		evidenceJSON, err := json.Marshal(map[string]any{
			"nodeName":       node.Name,
			"ready":          ready,
			"conditions":     conditions,
			"pressureFlags":  pressureFlags,
			"podCount":       podCount,
			"runningPods":    runningCount,
			"pendingPods":    pendingCount,
			"failedPods":     failedCount,
			"cpuPercent":     parsePercentString(nodeCPUPercent),
			"memoryPercent":  parsePercentString(nodeMemoryPercent),
			"metricsPresent": metricsAvailable,
		})
		if err != nil {
			return ClusterHealthComputation{}, err
		}
		conditionsJSON, err := json.Marshal(conditions)
		if err != nil {
			return ClusterHealthComputation{}, err
		}
		pressureJSON, err := json.Marshal(pressureFlags)
		if err != nil {
			return ClusterHealthComputation{}, err
		}

		nodeSnapshots = append(nodeSnapshots, store.CreateNodeHealthSnapshotParams{
			ClusterID:              clusterID,
			NodeName:               node.Name,
			ObservedAt:             observedAt,
			Ready:                  ready,
			ConditionsJSON:         conditionsJSON,
			CapacityCPUMillicores:  int64String(capacityCPU),
			AllocatableCPUMilli:    int64String(allocatableCPU),
			UsedCPUMillicores:      int64String(usedCPU),
			CPUUsagePercent:        nodeCPUPercent,
			CapacityMemoryBytes:    int64String(capacityMemory),
			AllocatableMemoryBytes: int64String(allocatableMemory),
			UsedMemoryBytes:        int64String(usedMemory),
			MemoryUsagePercent:     nodeMemoryPercent,
			PodCount:               podCount,
			RunningPodCount:        runningCount,
			PendingPodCount:        pendingCount,
			FailedPodCount:         failedCount,
			PressureFlagsJSON:      pressureJSON,
			HealthStatus:           nodeStatus,
			HealthScore:            nodeScore,
			EvidenceJSON:           evidenceJSON,
		})
	}

	namespaceNames := make([]string, 0, len(namespaceAgg))
	for namespace := range namespaceAgg {
		namespaceNames = append(namespaceNames, namespace)
	}
	sort.Strings(namespaceNames)

	namespaceSnapshots := make([]store.CreateNamespaceHealthSnapshotParams, 0, len(namespaceNames))
	for _, namespace := range namespaceNames {
		agg := namespaceAgg[namespace]
		score := 100
		if agg.PendingPodCount > 0 {
			score -= minInt(15, agg.PendingPodCount*5)
		}
		if agg.FailedPodCount > 0 {
			score -= minInt(30, agg.FailedPodCount*12)
		}
		if agg.RestartCount > 0 {
			score -= minInt(20, agg.RestartCount*2)
		}
		if agg.WarningEventCount > 0 {
			score -= minInt(15, agg.WarningEventCount)
		}
		score = clampScore(score)
		status := healthStatusFromScore(score)
		if agg.FailedPodCount > 0 && score < 60 {
			status = "critical"
		}
		if agg.PodCount == 0 && agg.WarningEventCount == 0 {
			status = "healthy"
		}
		evidenceJSON, err := json.Marshal(map[string]any{
			"namespace":         namespace,
			"podCount":          agg.PodCount,
			"runningPodCount":   agg.RunningPodCount,
			"pendingPodCount":   agg.PendingPodCount,
			"failedPodCount":    agg.FailedPodCount,
			"restartCount":      agg.RestartCount,
			"warningEventCount": agg.WarningEventCount,
			"usedCPUMillicores": agg.UsedCPUMillicores,
			"usedMemoryBytes":   agg.UsedMemoryBytes,
		})
		if err != nil {
			return ClusterHealthComputation{}, err
		}
		namespaceSnapshots = append(namespaceSnapshots, store.CreateNamespaceHealthSnapshotParams{
			ClusterID:         clusterID,
			Namespace:         namespace,
			ObservedAt:        observedAt,
			PodCount:          agg.PodCount,
			RunningPodCount:   agg.RunningPodCount,
			PendingPodCount:   agg.PendingPodCount,
			FailedPodCount:    agg.FailedPodCount,
			RestartCount:      agg.RestartCount,
			WarningEventCount: agg.WarningEventCount,
			UsedCPUMillicores: int64String(agg.UsedCPUMillicores),
			UsedMemoryBytes:   int64String(agg.UsedMemoryBytes),
			HealthStatus:      status,
			HealthScore:       score,
			EvidenceJSON:      evidenceJSON,
		})
	}

	if percentOver(totalCPUUsed, totalCPUAllocatable, cfg.NodeCPUCriticalPercent) {
		clusterScore -= 25
	} else if percentOver(totalCPUUsed, totalCPUAllocatable, cfg.NodeCPUWarnPercent) {
		clusterScore -= 10
	}
	if percentOver(totalMemoryUsed, totalMemoryAllocatable, cfg.NodeMemoryCriticalPercent) {
		clusterScore -= 25
	} else if percentOver(totalMemoryUsed, totalMemoryAllocatable, cfg.NodeMemoryWarnPercent) {
		clusterScore -= 10
	}
	if pendingPods > 0 {
		clusterScore -= minInt(15, pendingPods*3)
	}
	if failedPods > 0 {
		clusterScore -= minInt(25, failedPods*10)
	}
	if warningEvents > 0 {
		clusterScore -= minInt(10, warningEvents)
	}
	if !metricsAvailable {
		clusterScore -= 10
	}
	clusterScore = clampScore(clusterScore)

	status := healthStatusFromScore(clusterScore)
	if len(nodes) == 0 {
		status = "unknown"
		clusterScore = 0
	}

	summaryJSON, err := json.Marshal(map[string]any{
		"nodes": map[string]any{
			"count":             len(nodes),
			"readyCount":        readyNodes,
			"notReadyCount":     notReadyNodes,
			"pressureNodeCount": clusterPressureCount,
		},
		"cpu": map[string]any{
			"totalMillicores":       totalCPUCapacity,
			"allocatableMillicores": totalCPUAllocatable,
			"usedMillicores":        totalCPUUsed,
			"usagePercent":          parsePercentString(percentString(totalCPUUsed, totalCPUAllocatable)),
		},
		"memory": map[string]any{
			"totalBytes":       totalMemoryCapacity,
			"allocatableBytes": totalMemoryAllocatable,
			"usedBytes":        totalMemoryUsed,
			"usagePercent":     parsePercentString(percentString(totalMemoryUsed, totalMemoryAllocatable)),
		},
		"pods": map[string]any{
			"count":        len(pods),
			"runningCount": runningPods,
			"pendingCount": pendingPods,
			"failedCount":  failedPods,
		},
		"events": map[string]any{
			"warningCount": warningEvents,
		},
		"metricsAvailable": metricsAvailable,
		"metricsError":     metricsError,
	})
	if err != nil {
		return ClusterHealthComputation{}, err
	}

	return ClusterHealthComputation{
		ClusterSnapshot: store.CreateClusterHealthSnapshotParams{
			ClusterID:             clusterID,
			ObservedAt:            observedAt,
			MetricsAvailable:      metricsAvailable,
			MetricsError:          metricsError,
			NodeCount:             len(nodes),
			ReadyNodeCount:        readyNodes,
			NotReadyNodeCount:     notReadyNodes,
			TotalCPUMillicores:    int64String(totalCPUCapacity),
			AllocatableCPUMilli:   int64String(totalCPUAllocatable),
			UsedCPUMillicores:     int64String(totalCPUUsed),
			CPUUsagePercent:       percentString(totalCPUUsed, totalCPUAllocatable),
			TotalMemoryBytes:      int64String(totalMemoryCapacity),
			AllocatableMemoryByte: int64String(totalMemoryAllocatable),
			UsedMemoryBytes:       int64String(totalMemoryUsed),
			MemoryUsagePercent:    percentString(totalMemoryUsed, totalMemoryAllocatable),
			PodCount:              len(pods),
			RunningPodCount:       runningPods,
			PendingPodCount:       pendingPods,
			FailedPodCount:        failedPods,
			WarningEventCount:     warningEvents,
			HealthStatus:          status,
			HealthScore:           clusterScore,
			SummaryJSON:           summaryJSON,
		},
		NodeSnapshots:      nodeSnapshots,
		NamespaceSnapshots: namespaceSnapshots,
	}, nil
}

type namespaceAggregate struct {
	PodCount          int
	RunningPodCount   int
	PendingPodCount   int
	FailedPodCount    int
	RestartCount      int
	WarningEventCount int
	UsedCPUMillicores int64
	UsedMemoryBytes   int64
}

func ensureNamespaceAgg(values map[string]*namespaceAggregate, namespace string) *namespaceAggregate {
	if namespace == "" {
		namespace = "default"
	}
	if agg, ok := values[namespace]; ok {
		return agg
	}
	agg := &namespaceAggregate{}
	values[namespace] = agg
	return agg
}

func countPods(pods []corev1.Pod) (count, running, pending, failed int) {
	count = len(pods)
	for _, pod := range pods {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			running++
		case corev1.PodPending:
			pending++
		case corev1.PodFailed:
			failed++
		}
	}
	return
}

func nodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func nodeConditionMap(node corev1.Node) map[string]any {
	conditions := map[string]any{}
	for _, condition := range node.Status.Conditions {
		conditions[string(condition.Type)] = map[string]any{
			"status":             string(condition.Status),
			"reason":             condition.Reason,
			"message":            condition.Message,
			"lastTransitionTime": condition.LastTransitionTime.Time.UTC().Format(time.RFC3339),
		}
	}
	return conditions
}

func pressureFlags(node corev1.Node) map[string]bool {
	flags := map[string]bool{
		"MemoryPressure": false,
		"DiskPressure":   false,
		"PIDPressure":    false,
	}
	for _, condition := range node.Status.Conditions {
		switch condition.Type {
		case corev1.NodeMemoryPressure:
			flags["MemoryPressure"] = condition.Status == corev1.ConditionTrue
		case corev1.NodeDiskPressure:
			flags["DiskPressure"] = condition.Status == corev1.ConditionTrue
		case corev1.NodePIDPressure:
			flags["PIDPressure"] = condition.Status == corev1.ConditionTrue
		}
	}
	return flags
}

func hasCriticalPressure(flags map[string]bool) bool {
	return flags["MemoryPressure"] || flags["DiskPressure"] || flags["PIDPressure"]
}

func sumPodMetric(metric metricsv1beta1.PodMetrics) (int64, int64) {
	cpu := int64(0)
	memory := int64(0)
	for _, container := range metric.Containers {
		cpu += container.Usage.Cpu().MilliValue()
		memory += container.Usage.Memory().Value()
	}
	return cpu, memory
}

func percentString(used, total int64) string {
	pct := percentFloat(used, total)
	if pct <= 0 {
		return ""
	}
	return fmt.Sprintf("%.1f", pct)
}

func percentFloat(used, total int64) float64 {
	if used <= 0 || total <= 0 {
		return 0
	}
	return math.Round((float64(used)/float64(total))*1000) / 10
}

func parsePercentString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func percentOver(used, total int64, threshold int) bool {
	pct := percentFloat(used, total)
	return pct >= float64(threshold)
}

func clampScore(score int) int {
	switch {
	case score < 0:
		return 0
	case score > 100:
		return 100
	default:
		return score
	}
}

func healthStatusFromScore(score int) string {
	switch {
	case score >= 85:
		return "healthy"
	case score >= 60:
		return "warning"
	case score >= 0:
		return "critical"
	default:
		return "unknown"
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
