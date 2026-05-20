package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"orbit/internal/auth"
	"orbit/internal/config"
	"orbit/internal/controller"
	"orbit/internal/reasoning"
	"orbit/internal/store"
)

type ReadyChecker interface {
	Ready(ctx context.Context) error
}

type AuthService interface {
	Login(ctx context.Context, username, password, remoteAddr, userAgent string) (auth.LoginResponse, error)
	Authenticate(ctx context.Context, token, remoteAddr, userAgent string) (auth.Principal, error)
}

type DataStore interface {
	ListClusters(ctx context.Context) ([]store.Cluster, error)
	ListKubernetesResources(ctx context.Context, filters store.ListResourceFilters) ([]store.KubernetesResource, error)
	ListFindings(ctx context.Context, filters store.ListFindingFilters) ([]store.Finding, error)
	GetFinding(ctx context.Context, findingID string) (store.Finding, error)
	GetEvidencePack(ctx context.Context, evidencePackID string) (store.EvidencePack, error)
	GetEvidencePackByFindingID(ctx context.Context, findingID string) (store.EvidencePack, error)
	ListEvidencePacks(ctx context.Context, filters store.ListEvidencePackFilters) ([]store.EvidencePack, error)
	CreateLLMRun(ctx context.Context, params store.CreateLLMRunParams) (store.LLMRun, error)
	CreateActionPlanDraft(ctx context.Context, params store.CreateActionPlanDraftParams) (store.ActionPlan, error)
	ListActionPlans(ctx context.Context) ([]store.ActionPlan, error)
	GetActionPlan(ctx context.Context, actionPlanID string) (store.ActionPlan, error)
	ControllerStatus(ctx context.Context) (store.ControllerStatus, error)
}

type Reasoner interface {
	Reason(finding *store.Finding, evidencePack store.EvidencePack) reasoning.Result
}

type EvidenceService interface {
	BuildFindingEvidencePack(ctx context.Context, findingID string, persist bool) (store.EvidencePack, error)
	BuildNamespaceEvidencePack(ctx context.Context, clusterID, namespace string, persist bool) (store.EvidencePack, error)
	BuildPodEvidencePack(ctx context.Context, clusterID, namespace, podName string, persist bool) (store.EvidencePack, error)
}

type infoResponse struct {
	AppName     string `json:"app_name"`
	Version     string `json:"version"`
	Environment string `json:"environment"`
	ClusterName string `json:"cluster_name"`
	ClusterType string `json:"cluster_type"`
	AuthMode    string `json:"auth_mode"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type meResponse struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	Source   string   `json:"source"`
}

type inventoryResourceResponse struct {
	ID              string            `json:"id"`
	ClusterID       string            `json:"cluster_id"`
	APIVersion      string            `json:"api_version"`
	Kind            string            `json:"kind"`
	Namespace       string            `json:"namespace,omitempty"`
	Name            string            `json:"name"`
	UID             string            `json:"uid"`
	ResourceVersion string            `json:"resource_version,omitempty"`
	Labels          map[string]string `json:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty"`
	Status          string            `json:"status,omitempty"`
	ObservedAt      time.Time         `json:"observed_at"`
}

type protectedResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type evidenceGenerateRequest struct {
	ScopeType string `json:"scopeType"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Persist   bool   `json:"persist"`
}

func newHandler(cfg config.Config, readyChecker ReadyChecker, authService AuthService, dataStore DataStore, evidenceService EvidenceService, reasoner Reasoner, ready *uint32) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeText(w, http.StatusOK, "healthy")
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if atomic.LoadUint32(ready) == 0 {
			writeText(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		if err := readyChecker.Ready(r.Context()); err != nil {
			writeText(w, http.StatusServiceUnavailable, "not ready")
			return
		}
		writeText(w, http.StatusOK, "ready")
	})

	mux.HandleFunc("/api/v1/info", func(w http.ResponseWriter, r *http.Request) {
		payload := infoResponse{
			AppName:     "orbit",
			Version:     "dev",
			Environment: cfg.Environment,
			ClusterName: cfg.ClusterName,
			ClusterType: cfg.ClusterType,
			AuthMode:    cfg.AuthMode,
		}

		writeJSON(w, http.StatusOK, payload)
	})

	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var request loginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
			return
		}

		response, err := authService.Login(r.Context(), request.Username, request.Password, r.RemoteAddr, r.UserAgent())
		if err != nil {
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
				return
			}

			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "login failed"})
			return
		}

		writeJSON(w, http.StatusOK, response)
	})

	mux.Handle("/api/v1/auth/me", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal := principalFromContext(r.Context())
		writeJSON(w, http.StatusOK, meResponse{
			Username: principal.Username,
			Roles:    principal.Roles,
			Source:   principal.Source,
		})
	})))

	mux.Handle("/api/v1/protected", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, protectedResponse{Status: "ok"})
	})))

	mux.Handle("/api/v1/clusters", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clusters, err := dataStore.ListClusters(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list clusters"})
			return
		}
		writeJSON(w, http.StatusOK, clusters)
	})))

	mux.Handle("/api/v1/finding-rules", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, controller.FindingRules())
	})))

	mux.Handle("/api/v1/inventory/resources", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resources, err := dataStore.ListKubernetesResources(r.Context(), store.ListResourceFilters{
			Kind:      r.URL.Query().Get("kind"),
			Namespace: r.URL.Query().Get("namespace"),
			Name:      r.URL.Query().Get("name"),
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list resources"})
			return
		}
		response := make([]inventoryResourceResponse, 0, len(resources))
		for _, resource := range resources {
			response = append(response, inventoryResourceResponse{
				ID:              resource.ID,
				ClusterID:       resource.ClusterID,
				APIVersion:      resource.APIVersion,
				Kind:            resource.Kind,
				Namespace:       resource.Namespace,
				Name:            resource.Name,
				UID:             resource.UID,
				ResourceVersion: resource.ResourceVersion,
				Labels:          resource.Labels,
				Annotations:     resource.Annotations,
				Status:          resource.Status,
				ObservedAt:      resource.ObservedAt,
			})
		}
		writeJSON(w, http.StatusOK, response)
	})))

	mux.Handle("/api/v1/findings", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		findings, err := dataStore.ListFindings(r.Context(), store.ListFindingFilters{
			Severity:  r.URL.Query().Get("severity"),
			Status:    r.URL.Query().Get("status"),
			Namespace: r.URL.Query().Get("namespace"),
			Kind:      r.URL.Query().Get("kind"),
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list findings"})
			return
		}
		writeJSON(w, http.StatusOK, findings)
	})))

	mux.Handle("/api/v1/findings/", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleFindingRoutes(w, r, dataStore, evidenceService, reasoner)
	})))

	mux.Handle("/api/v1/evidence-packs/generate", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var request evidenceGenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
			return
		}

		clusters, err := dataStore.ListClusters(r.Context())
		if err != nil || len(clusters) == 0 {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to resolve cluster"})
			return
		}
		clusterID := clusters[0].ID

		var pack store.EvidencePack
		switch request.ScopeType {
		case "namespace":
			if request.Namespace == "" {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "namespace is required"})
				return
			}
			pack, err = evidenceService.BuildNamespaceEvidencePack(r.Context(), clusterID, request.Namespace, request.Persist)
		case "pod":
			if request.Namespace == "" || request.Name == "" {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "namespace and name are required for pod scope"})
				return
			}
			pack, err = evidenceService.BuildPodEvidencePack(r.Context(), clusterID, request.Namespace, request.Name, request.Persist)
		default:
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: "unsupported scopeType"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to build evidence pack"})
			return
		}
		writeJSON(w, http.StatusOK, pack)
	})))

	mux.Handle("/api/v1/evidence-packs", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		packs, err := dataStore.ListEvidencePacks(r.Context(), store.ListEvidencePackFilters{
			ScopeType: r.URL.Query().Get("scopeType"),
			Namespace: r.URL.Query().Get("namespace"),
			Name:      r.URL.Query().Get("name"),
			FindingID: r.URL.Query().Get("findingId"),
		})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list evidence packs"})
			return
		}
		writeJSON(w, http.StatusOK, packs)
	})))

	mux.Handle("/api/v1/evidence-packs/", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleEvidencePackRoutes(w, r, dataStore, reasoner)
	})))

	mux.Handle("/api/v1/action-plans", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		plans, err := dataStore.ListActionPlans(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to list action plans"})
			return
		}
		writeJSON(w, http.StatusOK, plans)
	})))

	mux.Handle("/api/v1/action-plans/", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actionPlanID := strings.TrimPrefix(r.URL.Path, "/api/v1/action-plans/")
		if actionPlanID == "" {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
			return
		}
		plan, err := dataStore.GetActionPlan(r.Context(), actionPlanID)
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "action plan not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get action plan"})
			return
		}
		writeJSON(w, http.StatusOK, plan)
	})))

	mux.Handle("/api/v1/controller/status", requireAuth(authService, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, err := dataStore.ControllerStatus(r.Context())
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "controller status not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get controller status"})
			return
		}
		writeJSON(w, http.StatusOK, status)
	})))

	return mux
}

func handleFindingRoutes(w http.ResponseWriter, r *http.Request, dataStore DataStore, evidenceService EvidenceService, reasoner Reasoner) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/findings/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	findingID := parts[0]

	if len(parts) == 1 {
		finding, err := dataStore.GetFinding(r.Context(), findingID)
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "finding not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get finding"})
			return
		}
		writeJSON(w, http.StatusOK, finding)
		return
	}

	switch parts[1] {
	case "evidence-pack":
		pack, err := dataStore.GetEvidencePackByFindingID(r.Context(), findingID)
		if errors.Is(err, store.ErrNotFound) {
			pack, err = evidenceService.BuildFindingEvidencePack(r.Context(), findingID, true)
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get evidence pack"})
			return
		}
		writeJSON(w, http.StatusOK, pack)
	case "reason":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		finding, err := dataStore.GetFinding(r.Context(), findingID)
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "finding not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get finding"})
			return
		}
		pack, err := dataStore.GetEvidencePackByFindingID(r.Context(), findingID)
		if errors.Is(err, store.ErrNotFound) {
			pack, err = evidenceService.BuildFindingEvidencePack(r.Context(), findingID, true)
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get evidence pack"})
			return
		}

		writeReasoningResponse(w, r, dataStore, reasoner, &finding, pack)
	default:
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
	}
}

func handleEvidencePackRoutes(w http.ResponseWriter, r *http.Request, dataStore DataStore, reasoner Reasoner) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/evidence-packs/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
		return
	}
	evidencePackID := parts[0]

	if len(parts) == 1 {
		pack, err := dataStore.GetEvidencePack(r.Context(), evidencePackID)
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "evidence pack not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get evidence pack"})
			return
		}
		writeJSON(w, http.StatusOK, pack)
		return
	}

	switch parts[1] {
	case "reason":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		pack, err := dataStore.GetEvidencePack(r.Context(), evidencePackID)
		if errors.Is(err, store.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, errorResponse{Error: "evidence pack not found"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to get evidence pack"})
			return
		}
		var finding *store.Finding
		if pack.FindingID != "" {
			found, err := dataStore.GetFinding(r.Context(), pack.FindingID)
			if err == nil {
				finding = &found
			}
		}
		writeReasoningResponse(w, r, dataStore, reasoner, finding, pack)
	default:
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "not found"})
	}
}

func writeReasoningResponse(w http.ResponseWriter, r *http.Request, dataStore DataStore, reasoner Reasoner, finding *store.Finding, pack store.EvidencePack) {
	now := time.Now().UTC()
	result := reasoner.Reason(finding, pack)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to encode reasoning"})
		return
	}

	findingID := ""
	if finding != nil {
		findingID = finding.ID
	}
	run, err := dataStore.CreateLLMRun(r.Context(), store.CreateLLMRunParams{
		FindingID:           findingID,
		EvidencePackID:      pack.ID,
		Provider:            "mock",
		Model:               "mock-phase1",
		PromptVersion:       "phase1-v1",
		InputTokenEstimate:  pack.TokenEstimate,
		OutputTokenEstimate: len(resultJSON) / 4,
		OutputJSON:          resultJSON,
		CreatedAt:           now,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create llm run"})
		return
	}
	planParams, err := reasoning.BuildActionPlanDraft(finding, pack, result, now)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to build action plan"})
		return
	}
	plan, err := dataStore.CreateActionPlanDraft(r.Context(), planParams)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to create action plan"})
		return
	}
	response := map[string]any{
		"evidencePack": pack,
		"reasoning":    result,
		"llmRun":       run,
		"actionPlan":   plan,
	}
	if finding != nil {
		response["finding"] = finding
	}
	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(payload)
}

func writeText(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(message + "\n"))
}

type contextKey string

const principalContextKey contextKey = "principal"

func requireAuth(authService AuthService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "missing bearer token"})
			return
		}

		principal, err := authService.Authenticate(r.Context(), token, r.RemoteAddr, r.UserAgent())
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalContextKey, principal)))
	})
}

func bearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}

	return token, true
}

func principalFromContext(ctx context.Context) auth.Principal {
	value, _ := ctx.Value(principalContextKey).(auth.Principal)
	return value
}
