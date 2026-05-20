package controller

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"

	"orbit/internal/config"
	"orbit/internal/evidence"
	"orbit/internal/store"
)

type Logger interface {
	Info(msg string, attrs ...any)
	Error(msg string, attrs ...any)
}

type Runner struct {
	cfg      config.Config
	store    *store.Store
	evidence *evidence.Service
	logger   Logger
}

func NewRunner(cfg config.Config, dataStore *store.Store, logger Logger) *Runner {
	return &Runner{
		cfg:      cfg,
		store:    dataStore,
		evidence: evidence.NewService(cfg, dataStore),
		logger:   logger,
	}
}

func (r *Runner) Run(ctx context.Context) error {
	if !r.cfg.ControllerEnabled {
		r.logger.Info("controller disabled by config")
		<-ctx.Done()
		return nil
	}

	inClusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return err
	}

	kubeClient, err := kubernetes.NewForConfig(inClusterConfig)
	if err != nil {
		return err
	}

	metricsClient, err := metricsclient.NewForConfig(inClusterConfig)
	if err != nil {
		r.logger.Info("metrics client initialization failed", "error", err.Error())
	}

	if err := r.runOnce(ctx, kubeClient, metricsClient); err != nil {
		r.logger.Error("controller run failed", "error", err.Error())
	}

	ticker := time.NewTicker(time.Duration(r.cfg.ControllerIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.runOnce(ctx, kubeClient, metricsClient); err != nil {
				r.logger.Error("controller run failed", "error", err.Error())
			}
		}
	}
}

func (r *Runner) runOnce(ctx context.Context, kubeClient kubernetes.Interface, metricsClient metricsclient.Interface) error {
	cluster, err := r.store.EnsureCluster(ctx, r.cfg.ClusterName, r.cfg.ClusterType, r.cfg.Mode)
	if err != nil {
		return err
	}

	namespaces, err := kubeClient.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	deployments, err := kubeClient.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	replicaSets, err := kubeClient.AppsV1().ReplicaSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	pods, err := kubeClient.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	services, err := kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	configMaps, err := kubeClient.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	events, err := kubeClient.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	metricsList := State{}
	metricsAvailable := false
	if metricsClient != nil {
		podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
		if err != nil {
			r.logger.Info("metrics api unavailable", "error", err.Error())
		} else {
			metricsList.PodMetrics = podMetrics.Items
			metricsAvailable = true
		}
	}

	resourcesByUID := map[string]store.KubernetesResource{}
	now := time.Now().UTC()

	for _, namespace := range namespaces.Items {
		resource, err := r.upsertObject(ctx, cluster.ID, "v1", "Namespace", "", namespace.Name, string(namespace.UID), namespace.ResourceVersion, namespace.Labels, namespace.Annotations, string(namespace.Status.Phase), namespace, now)
		if err != nil {
			return err
		}
		resourcesByUID[resource.UID] = resource
	}
	for _, deployment := range deployments.Items {
		resource, err := r.upsertObject(ctx, cluster.ID, "apps/v1", "Deployment", deployment.Namespace, deployment.Name, string(deployment.UID), deployment.ResourceVersion, deployment.Labels, deployment.Annotations, deploymentStatus(deployment), deployment, now)
		if err != nil {
			return err
		}
		resourcesByUID[resource.UID] = resource
	}
	for _, replicaSet := range replicaSets.Items {
		resource, err := r.upsertObject(ctx, cluster.ID, "apps/v1", "ReplicaSet", replicaSet.Namespace, replicaSet.Name, string(replicaSet.UID), replicaSet.ResourceVersion, replicaSet.Labels, replicaSet.Annotations, replicaSetStatus(replicaSet), replicaSet, now)
		if err != nil {
			return err
		}
		resourcesByUID[resource.UID] = resource
	}
	for _, pod := range pods.Items {
		resource, err := r.upsertObject(ctx, cluster.ID, "v1", "Pod", pod.Namespace, pod.Name, string(pod.UID), pod.ResourceVersion, pod.Labels, pod.Annotations, string(pod.Status.Phase), pod, now)
		if err != nil {
			return err
		}
		resourcesByUID[resource.UID] = resource
	}
	for _, service := range services.Items {
		resource, err := r.upsertObject(ctx, cluster.ID, "v1", "Service", service.Namespace, service.Name, string(service.UID), service.ResourceVersion, service.Labels, service.Annotations, string(service.Spec.Type), service, now)
		if err != nil {
			return err
		}
		resourcesByUID[resource.UID] = resource
	}
	for _, configMap := range configMaps.Items {
		resource, err := r.upsertObject(ctx, cluster.ID, "v1", "ConfigMap", configMap.Namespace, configMap.Name, string(configMap.UID), configMap.ResourceVersion, configMap.Labels, configMap.Annotations, "active", configMap, now)
		if err != nil {
			return err
		}
		resourcesByUID[resource.UID] = resource
	}

	for _, event := range events.Items {
		firstSeen := nullableEventTime(event.FirstTimestamp)
		lastSeen := nullableEventTime(event.LastTimestamp)
		if _, err := r.store.UpsertKubernetesEvent(ctx, store.UpsertKubernetesEventParams{
			ClusterID:    cluster.ID,
			InvolvedUID:  string(event.InvolvedObject.UID),
			Namespace:    event.Namespace,
			InvolvedKind: event.InvolvedObject.Kind,
			InvolvedName: event.InvolvedObject.Name,
			Type:         event.Type,
			Reason:       event.Reason,
			Message:      event.Message,
			Count:        int(event.Count),
			FirstSeenAt:  firstSeen,
			LastSeenAt:   lastSeen,
			ObservedAt:   now,
		}); err != nil {
			return err
		}
	}

	if metricsAvailable {
		for _, podMetric := range metricsList.PodMetrics {
			for _, containerMetric := range podMetric.Containers {
				if err := r.store.InsertResourceMetric(ctx, store.InsertResourceMetricParams{
					ClusterID:     cluster.ID,
					Namespace:     podMetric.Namespace,
					PodName:       podMetric.Name,
					ContainerName: containerMetric.Name,
					CPUMillicores: int64String(containerMetric.Usage.Cpu().MilliValue()),
					MemoryBytes:   int64String(containerMetric.Usage.Memory().Value()),
					ObservedAt:    now,
				}); err != nil {
					return err
				}
			}
		}
	}

	state := State{
		Deployments: deployments.Items,
		ReplicaSets: replicaSets.Items,
		Pods:        pods.Items,
		Events:      events.Items,
		PodMetrics:  metricsList.PodMetrics,
	}
	findings := EvaluateFindings(state)
	evidencePackCount := 0

	for _, candidate := range findings {
		resource := resourcesByUID[candidate.ResourceUID]
		evidenceJSON, err := json.Marshal(candidate.Evidence)
		if err != nil {
			return err
		}
		finding, err := r.store.UpsertFinding(ctx, store.UpsertFindingParams{
			ClusterID:    cluster.ID,
			ResourceID:   resource.ID,
			Severity:     candidate.Severity,
			Category:     candidate.Category,
			Title:        candidate.Title,
			Description:  candidate.Description,
			Status:       "open",
			EvidenceJSON: evidenceJSON,
			CreatedAt:    now,
			UpdatedAt:    now,
		})
		if err != nil {
			return err
		}

		if _, err := r.evidence.BuildFindingEvidencePack(ctx, finding.ID, true); err != nil {
			return err
		}
		evidencePackCount++
	}

	if err := r.store.UpdateClusterLastSeen(ctx, cluster.ID, now); err != nil {
		return err
	}

	r.logger.Info(
		"controller scan completed",
		"cluster", cluster.Name,
		"namespaces", len(namespaces.Items),
		"deployments", len(deployments.Items),
		"replicasets", len(replicaSets.Items),
		"pods", len(pods.Items),
		"services", len(services.Items),
		"configmaps", len(configMaps.Items),
		"events", len(events.Items),
		"metrics_available", metricsAvailable,
		"findings", len(findings),
		"evidence_packs", evidencePackCount,
	)

	return nil
}

func (r *Runner) upsertObject(
	ctx context.Context,
	clusterID, apiVersion, kind, namespace, name, uid, resourceVersion string,
	labels, annotations map[string]string,
	status string,
	obj any,
	observedAt time.Time,
) (store.KubernetesResource, error) {
	rawJSON, err := json.Marshal(obj)
	if err != nil {
		return store.KubernetesResource{}, err
	}
	return r.store.UpsertKubernetesResource(ctx, store.UpsertKubernetesResourceParams{
		ClusterID:       clusterID,
		APIVersion:      apiVersion,
		Kind:            kind,
		Namespace:       namespace,
		Name:            name,
		UID:             uid,
		ResourceVersion: resourceVersion,
		Labels:          labels,
		Annotations:     annotations,
		Status:          status,
		RawJSON:         rawJSON,
		ObservedAt:      observedAt,
	})
}

func nullableEventTime(value metav1.Time) sql.NullTime {
	if value.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: value.Time.UTC(), Valid: true}
}

func deploymentStatus(deployment appsv1.Deployment) string {
	return int32String(deployment.Status.AvailableReplicas) + "/" + int32String(pointerInt32Value(deployment.Spec.Replicas, 1))
}

func replicaSetStatus(replicaSet appsv1.ReplicaSet) string {
	return int32String(replicaSet.Status.ReadyReplicas) + "/" + int32String(pointerInt32Value(replicaSet.Spec.Replicas, 1))
}

func pointerInt32Value(value *int32, fallback int32) int32 {
	if value == nil {
		return fallback
	}
	return *value
}

func int32String(value int32) string {
	return int64String(int64(value))
}

func int64String(value int64) string {
	return fmt.Sprintf("%d", value)
}
