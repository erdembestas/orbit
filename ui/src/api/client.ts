export type ApiError = {
  status: number;
  message: string;
};

export type NullableTime = { Time?: string; Valid?: boolean } | string | null;

export type InfoResponse = {
  app_name: string;
  version: string;
  environment: string;
  cluster_name: string;
  cluster_type: string;
  auth_mode: string;
};

export type LoginResponse = {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
};

export type MeResponse = {
  username: string;
  roles: string[];
  source: string;
};

export type Cluster = {
  id: string;
  name: string;
  type: string;
  mode: string;
  status: string;
  last_seen_at?: NullableTime;
};

export type Resource = {
  id: string;
  cluster_id: string;
  api_version: string;
  kind: string;
  namespace?: string;
  name: string;
  uid: string;
  resource_version?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  status?: string;
  observed_at: string;
};

export type Finding = {
  id: string;
  cluster_id: string;
  resource_id?: string;
  severity: string;
  category: string;
  title: string;
  description: string;
  status: string;
  evidence_json: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  resource_kind?: string;
  resource_namespace?: string;
  resource_name?: string;
};

export type EvidencePack = {
  id: string;
  findingId?: string;
  clusterId: string;
  resourceId?: string;
  scopeType: string;
  scopeNamespace?: string;
  scopeName?: string;
  tokenEstimate: number;
  packJson: Record<string, unknown>;
  createdAt: string;
};

export type ActionPlan = {
  id: string;
  findingId?: string;
  evidencePackId?: string;
  title: string;
  summary: string;
  riskLevel: string;
  status: string;
  planJson: Record<string, unknown>;
  createdAt: string;
  updatedAt: string;
};

export type ControllerStatus = {
  cluster_name: string;
  mode: string;
  last_seen_at?: NullableTime;
  resource_counts_by_kind?: Record<string, number>;
  open_finding_count?: number;
  latest_evidence_pack_count?: number;
  metrics_available?: boolean;
};

export type FindingRule = {
  name: string;
  resourceKind: string;
  category: string;
  severity: string;
  title: string;
  description: string;
  condition: string;
  evidenceFields: string[];
  limitations: string[];
};

export type ClusterHealthSnapshot = {
  id: string;
  cluster_id: string;
  observed_at: string;
  metrics_available: boolean;
  metrics_error?: string;
  health_status: string;
  health_score: number;
  node_count: number;
  ready_node_count: number;
  not_ready_node_count: number;
  cpu_usage_percent?: string;
  memory_usage_percent?: string;
  pod_count: number;
  running_pod_count: number;
  pending_pod_count: number;
  failed_pod_count: number;
  warning_event_count: number;
  summary_json?: Record<string, unknown>;
};

export type NodeHealthSnapshot = {
  id: string;
  node_name: string;
  observed_at: string;
  ready: boolean;
  cpu_usage_percent?: string;
  memory_usage_percent?: string;
  pod_count: number;
  running_pod_count: number;
  pending_pod_count: number;
  failed_pod_count: number;
  health_status: string;
  health_score: number;
};

export type NamespaceHealthSnapshot = {
  id: string;
  namespace: string;
  observed_at: string;
  pod_count: number;
  running_pod_count: number;
  pending_pod_count: number;
  failed_pod_count: number;
  restart_count: number;
  warning_event_count: number;
  used_cpu_millicores?: string;
  used_memory_bytes?: string;
  health_status: string;
  health_score: number;
};

export type ClusterHealthResponse = {
  cluster: Cluster;
  observedAt: string;
  metricsAvailable: boolean;
  metricsError?: string;
  healthStatus: string;
  healthScore: number;
  summary?: Record<string, unknown>;
  nodes: NodeHealthSnapshot[];
  namespaces: NamespaceHealthSnapshot[];
};

export type ReasoningResult = {
  rootCause: string;
  confidence: number;
  riskLevel: string;
  suggestedActionPlan: string[];
  validationSteps: string[];
  rollbackSteps: string[];
};

export type ReasoningResponse = {
  reasoning: ReasoningResult;
  actionPlan: ActionPlan;
  evidencePack: EvidencePack;
};

export type EvidenceGenerateRequest = {
  scopeType: "namespace" | "pod";
  namespace: string;
  name?: string;
  persist: boolean;
};

type OrbitApiClientOptions = {
  getToken: () => string | null;
  onUnauthorized: () => void;
};

type RequestOptions = RequestInit & {
  optional404?: boolean;
};

export class OrbitApiClient {
  private getToken: OrbitApiClientOptions["getToken"];
  private onUnauthorized: OrbitApiClientOptions["onUnauthorized"];

  constructor(options: OrbitApiClientOptions) {
    this.getToken = options.getToken;
    this.onUnauthorized = options.onUnauthorized;
  }

  async login(username: string, password: string) {
    return this.request<LoginResponse>("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
  }

  async getInfo() {
    return this.request<InfoResponse>("/api/v1/info");
  }

  async getMe() {
    return this.request<MeResponse>("/api/v1/auth/me");
  }

  async getClusters() {
    return this.request<Cluster[]>("/api/v1/clusters", { optional404: true });
  }

  async getControllerStatus() {
    return this.request<ControllerStatus>("/api/v1/controller/status", { optional404: true });
  }

  async listFindingRules() {
    return this.request<FindingRule[]>("/api/v1/finding-rules", { optional404: true });
  }

  async getClusterHealth() {
    return this.request<ClusterHealthResponse>("/api/v1/cluster/health", { optional404: true });
  }

  async getNodeHealth() {
    return this.request<NodeHealthSnapshot[]>("/api/v1/cluster/health/nodes", { optional404: true });
  }

  async getNamespaceHealth() {
    return this.request<NamespaceHealthSnapshot[]>("/api/v1/cluster/health/namespaces", { optional404: true });
  }

  async getClusterHealthHistory(limit = 50) {
    return this.request<ClusterHealthSnapshot[]>(`/api/v1/cluster/health/history?limit=${limit}`, { optional404: true });
  }

  async listResources(filters?: { kind?: string; namespace?: string; name?: string }) {
    return this.request<Resource[]>(withQuery("/api/v1/inventory/resources", filters), { optional404: true });
  }

  async listFindings(filters?: { severity?: string; status?: string; namespace?: string; kind?: string }) {
    return this.request<Finding[]>(withQuery("/api/v1/findings", filters), { optional404: true });
  }

  async getFindingEvidencePack(id: string) {
    const response = await this.request<Record<string, unknown>>(`/api/v1/findings/${id}/evidence-pack`, { optional404: true });
    return response ? normalizeEvidencePack(response) : null;
  }

  async reasonFinding(id: string) {
    const response = await this.request<Record<string, unknown>>(`/api/v1/findings/${id}/reason`, {
      method: "POST",
      optional404: true,
    });
    return response ? normalizeReasoningResponse(response) : null;
  }

  async listEvidencePacks(filters?: { scopeType?: string; namespace?: string; name?: string; findingId?: string }) {
    const response = await this.request<Record<string, unknown>[]>(withQuery("/api/v1/evidence-packs", filters), { optional404: true });
    return response ? response.map(normalizeEvidencePack) : null;
  }

  async getEvidencePack(id: string) {
    const response = await this.request<Record<string, unknown>>(`/api/v1/evidence-packs/${id}`, { optional404: true });
    return response ? normalizeEvidencePack(response) : null;
  }

  async generateEvidencePack(payload: EvidenceGenerateRequest) {
    const response = await this.request<Record<string, unknown>>("/api/v1/evidence-packs/generate", {
      method: "POST",
      body: JSON.stringify(payload),
      optional404: true,
    });
    return response ? normalizeEvidencePack(response) : null;
  }

  async reasonEvidencePack(id: string) {
    const response = await this.request<Record<string, unknown>>(`/api/v1/evidence-packs/${id}/reason`, {
      method: "POST",
      optional404: true,
    });
    return response ? normalizeReasoningResponse(response) : null;
  }

  async listActionPlans() {
    const response = await this.request<Record<string, unknown>[]>("/api/v1/action-plans", { optional404: true });
    return response ? response.map(normalizeActionPlan) : null;
  }

  private async request<T>(path: string, options: RequestOptions = {}): Promise<T | null> {
    const headers = new Headers(options.headers ?? {});
    const token = this.getToken();

    if (!headers.has("Accept")) {
      headers.set("Accept", "application/json");
    }
    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }
    if (options.body && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    const response = await fetch(path, {
      ...options,
      headers,
    });

    const body = await parseResponseBody(response);

    if (response.status === 401) {
      this.onUnauthorized();
      throw {
        status: response.status,
        message: typeof body === "string" ? body : body?.error ?? "unauthorized",
      } satisfies ApiError;
    }

    if (response.status === 404 && options.optional404) {
      return null;
    }

    if (!response.ok) {
      throw {
        status: response.status,
        message: typeof body === "string" ? body : body?.error ?? "request failed",
      } satisfies ApiError;
    }

    return body as T;
  }
}

async function parseResponseBody(response: Response) {
  const contentType = response.headers.get("content-type") ?? "";
  if (contentType.includes("application/json")) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }

  try {
    return await response.text();
  } catch {
    return "";
  }
}

function withQuery(path: string, params?: Record<string, string | undefined>) {
  if (!params) {
    return path;
  }

  const query = new URLSearchParams();
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      query.set(key, value);
    }
  });

  const encoded = query.toString();
  return encoded ? `${path}?${encoded}` : path;
}

export function formatNullableTime(value?: NullableTime) {
  if (!value) {
    return "-";
  }
  if (typeof value === "string") {
    return formatDateTime(value);
  }
  if (value.Valid && value.Time) {
    return formatDateTime(value.Time);
  }
  return "-";
}

export function formatDateTime(value?: string) {
  if (!value) {
    return "-";
  }

  return new Date(value).toLocaleString();
}

function normalizeEvidencePack(input: Record<string, unknown>): EvidencePack {
  return {
    id: String(input.id ?? ""),
    findingId: optionalString(input.finding_id),
    clusterId: String(input.cluster_id ?? ""),
    resourceId: optionalString(input.resource_id),
    scopeType: String(input.scope_type ?? ""),
    scopeNamespace: optionalString(input.scope_namespace),
    scopeName: optionalString(input.scope_name),
    tokenEstimate: typeof input.token_estimate === "number" ? input.token_estimate : 0,
    packJson: isRecord(input.pack_json) ? input.pack_json : {},
    createdAt: String(input.created_at ?? ""),
  };
}

function normalizeActionPlan(input: Record<string, unknown>): ActionPlan {
  return {
    id: String(input.id ?? ""),
    findingId: optionalString(input.finding_id),
    evidencePackId: optionalString(input.evidence_pack_id),
    title: String(input.title ?? ""),
    summary: String(input.summary ?? ""),
    riskLevel: String(input.risk_level ?? ""),
    status: String(input.status ?? ""),
    planJson: isRecord(input.plan_json) ? input.plan_json : {},
    createdAt: String(input.created_at ?? ""),
    updatedAt: String(input.updated_at ?? ""),
  };
}

function normalizeReasoningResponse(input: Record<string, unknown>): ReasoningResponse {
  return {
    reasoning: (input.reasoning as ReasoningResult) ?? {
      rootCause: "",
      confidence: 0,
      riskLevel: "",
      suggestedActionPlan: [],
      validationSteps: [],
      rollbackSteps: [],
    },
    actionPlan: normalizeActionPlan(isRecord(input.actionPlan) ? input.actionPlan : {}),
    evidencePack: normalizeEvidencePack(isRecord(input.evidencePack) ? input.evidencePack : {}),
  };
}

function optionalString(value: unknown) {
  return typeof value === "string" && value !== "" ? value : undefined;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}
