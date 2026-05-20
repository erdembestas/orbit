package reasoning

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"orbit/internal/store"
)

type Result struct {
	RootCause           string   `json:"rootCause"`
	Confidence          float64  `json:"confidence"`
	RiskLevel           string   `json:"riskLevel"`
	SuggestedActionPlan []string `json:"suggestedActionPlan"`
	ValidationSteps     []string `json:"validationSteps"`
	RollbackSteps       []string `json:"rollbackSteps"`
}

type MockProvider struct{}

func (MockProvider) Reason(finding *store.Finding, evidencePack store.EvidencePack) Result {
	title := ""
	if finding != nil {
		title = finding.Title
	}
	packText := strings.ToLower(string(evidencePack.PackJSON))

	switch {
	case strings.Contains(title, "Probe may target wrong port"):
		return Result{
			RootCause:           "The configured readiness or liveness probe appears to target a port that is not exposed by the workload containers.",
			Confidence:          0.93,
			RiskLevel:           "high",
			SuggestedActionPlan: []string{"Verify the container ports exposed by the workload.", "Update the probe to target a matching named or numeric port.", "Re-run a rollout status check after the probe fix."},
			ValidationSteps:     []string{"Check pod readiness after rollout.", "Confirm probe success events replace probe failures."},
			RollbackSteps:       []string{"Restore the previous probe port configuration if the updated probe causes new readiness failures."},
		}
	case strings.Contains(title, "Deployment has unavailable replicas"):
		return Result{
			RootCause:           "The deployment is not meeting its desired availability because one or more pods are failing to become Ready or are not scheduled successfully.",
			Confidence:          0.82,
			RiskLevel:           "medium",
			SuggestedActionPlan: []string{"Inspect rollout status and related pods.", "Review recent warning events affecting the deployment and pods.", "Resolve the blocking pod issue before retrying the rollout."},
			ValidationSteps:     []string{"Run kubectl rollout status for the deployment.", "Confirm available replicas match desired replicas."},
			RollbackSteps:       []string{"Pause the rollout or revert to the previous deployment revision if the new replica set remains unavailable."},
		}
	case strings.Contains(title, "Pod has container restarts"):
		return Result{
			RootCause:           "A container in the pod has restarted, which often indicates crashes, failed dependencies, or resource pressure.",
			Confidence:          0.74,
			RiskLevel:           "low",
			SuggestedActionPlan: []string{"Inspect recent container logs for the restarted container.", "Review resource requests and limits for signs of starvation or OOM pressure.", "Check warning events for repeated probe or crash loop failures."},
			ValidationSteps:     []string{"Confirm restart counts stop increasing.", "Confirm the pod remains Ready for multiple controller intervals."},
			RollbackSteps:       []string{"Revert the recent workload change if restart counts continue increasing after mitigation."},
		}
	case strings.Contains(packText, "probe may target wrong port"):
		return Result{
			RootCause:           "The compact evidence pack suggests a probe configuration mismatch and the health checks likely do not align with the exposed container port.",
			Confidence:          0.9,
			RiskLevel:           "high",
			SuggestedActionPlan: []string{"Check the workload probe configuration.", "Align the readiness or liveness probe port with the container port.", "Re-check pod readiness after the probe fix."},
			ValidationSteps:     []string{"Confirm readiness and liveness checks succeed.", "Confirm warning events for probe failures stop increasing."},
			RollbackSteps:       []string{"Restore the previous probe settings if the new probe change causes additional failures."},
		}
	case strings.Contains(packText, "deployment has unavailable replicas"):
		return Result{
			RootCause:           "The evidence pack shows the deployment is not meeting desired availability and likely needs rollout, pod, and warning-event inspection.",
			Confidence:          0.82,
			RiskLevel:           "medium",
			SuggestedActionPlan: []string{"Check rollout status for the deployment.", "Inspect related pods and warning events.", "Resolve the blocking health or scheduling issue before retrying the rollout."},
			ValidationSteps:     []string{"Confirm available replicas match desired replicas.", "Confirm warning events stop recurring."},
			RollbackSteps:       []string{"Revert the rollout if availability continues to decline after the change."},
		}
	case strings.Contains(packText, "pod has container restarts") || strings.Contains(packText, "\"restartcount\":") || strings.Contains(packText, "\"restartcount\": "):
		return Result{
			RootCause:           "The evidence pack indicates one or more containers have restarted, which points to crashes, resource pressure, or probe failures.",
			Confidence:          0.78,
			RiskLevel:           "low",
			SuggestedActionPlan: []string{"Inspect recent container logs for the affected pod.", "Review resource requests and limits.", "Check recent config or rollout changes affecting the pod."},
			ValidationSteps:     []string{"Confirm restart counts stop increasing.", "Confirm the pod stays Ready over multiple controller loops."},
			RollbackSteps:       []string{"Rollback the most recent workload change if restart behavior continues."},
		}
	case evidencePack.ScopeType == "namespace" && strings.Contains(packText, "unhealthypods") && strings.Count(packText, "\"name\"") > 2:
		return Result{
			RootCause:           "Multiple unhealthy workloads appear within the namespace, which suggests a namespace-level dependency, rollout, or capacity issue.",
			Confidence:          0.73,
			RiskLevel:           "medium",
			SuggestedActionPlan: []string{"Inspect shared deployments and services in the namespace.", "Review the recent namespace warning events.", "Prioritize the most restart-heavy or unavailable workloads first."},
			ValidationSteps:     []string{"Confirm unhealthy pod counts decrease on subsequent controller scans."},
			RollbackSteps:       []string{"Undo the most recent namespace-scoped change if multiple workloads degraded together."},
		}
	case evidencePack.ScopeType == "pod" && strings.Contains(packText, "\"restartcount\":") && !strings.Contains(packText, "\"restartcount\":0"):
		return Result{
			RootCause:           "The pod-specific evidence pack shows restart activity and warrants targeted pod-level investigation.",
			Confidence:          0.76,
			RiskLevel:           "low",
			SuggestedActionPlan: []string{"Inspect recent logs for the affected container.", "Review pod resource settings and probe behavior.", "Check related rollout and config changes for this workload."},
			ValidationSteps:     []string{"Confirm the pod remains Ready and restart counts remain stable."},
			RollbackSteps:       []string{"Revert the most recent workload change if the pod continues to restart."},
		}
	default:
		riskLevel := "medium"
		if finding != nil && finding.Severity != "" {
			riskLevel = strings.ToLower(finding.Severity)
		}
		return Result{
			RootCause:           "A deterministic finding was raised and requires operator review of the summarized workload evidence.",
			Confidence:          0.6,
			RiskLevel:           riskLevel,
			SuggestedActionPlan: []string{"Review the affected workload summary.", "Inspect the related events and container status data.", "Apply the smallest safe configuration or rollout correction after review."},
			ValidationSteps:     []string{"Confirm the finding no longer reproduces on the next controller scan."},
			RollbackSteps:       []string{"Undo the last corrective change if the workload health degrades further."},
		}
	}
}

func BuildActionPlanDraft(finding *store.Finding, evidencePack store.EvidencePack, result Result, now time.Time) (store.CreateActionPlanDraftParams, error) {
	planJSON := map[string]any{
		"evidencePackId":      evidencePack.ID,
		"rootCause":           result.RootCause,
		"confidence":          result.Confidence,
		"suggestedActionPlan": result.SuggestedActionPlan,
		"validationSteps":     result.ValidationSteps,
		"rollbackSteps":       result.RollbackSteps,
	}
	title := draftTitle(finding, evidencePack)
	findingID := ""
	if finding != nil {
		findingID = finding.ID
		planJSON["findingId"] = finding.ID
	}
	payload, err := json.Marshal(planJSON)
	if err != nil {
		return store.CreateActionPlanDraftParams{}, err
	}

	return store.CreateActionPlanDraftParams{
		ID:             uuid.NewString(),
		FindingID:      findingID,
		EvidencePackID: evidencePack.ID,
		Title:          title,
		Summary:        result.RootCause,
		RiskLevel:      result.RiskLevel,
		Status:         "draft",
		PlanJSON:       payload,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func draftTitle(finding *store.Finding, evidencePack store.EvidencePack) string {
	if finding != nil && finding.Title != "" {
		return "Draft action plan for " + finding.Title
	}
	switch evidencePack.ScopeType {
	case "namespace":
		return "Draft action plan for namespace " + evidencePack.ScopeNamespace
	case "pod":
		return "Draft action plan for pod " + evidencePack.ScopeNamespace + "/" + evidencePack.ScopeName
	default:
		return "Draft action plan for evidence pack " + evidencePack.ID
	}
}
