import AutoAwesomeRoundedIcon from "@mui/icons-material/AutoAwesomeRounded";
import PreviewRoundedIcon from "@mui/icons-material/PreviewRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import {
  Alert,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogContent,
  DialogTitle,
  Grid2,
  MenuItem,
  Stack,
  Tab,
  Tabs,
  TextField,
  Typography,
} from "@mui/material";
import { type ReactNode, useEffect, useMemo, useState } from "react";
import {
  formatDateTime,
  type ApiError,
  type EvidencePack,
  type OrbitApiClient,
  type ReasoningResponse,
  type Resource,
} from "../api/client";
import DataPanel from "../components/DataPanel";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import JsonPreview from "../components/JsonPreview";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";
import StatusChip from "../components/StatusChip";

type AnalyzePageProps = {
  client: OrbitApiClient;
};

type NamespaceOption = {
  name: string;
  status?: string;
  observedAt?: string;
};

type SummaryCard = {
  label: string;
  value: string | number;
};

export default function AnalyzePage({ client }: AnalyzePageProps) {
  const [tab, setTab] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [namespaces, setNamespaces] = useState<NamespaceOption[]>([]);
  const [pods, setPods] = useState<Resource[]>([]);
  const [namespaceValue, setNamespaceValue] = useState("");
  const [podNamespace, setPodNamespace] = useState("");
  const [podName, setPodName] = useState("");
  const [selectedPack, setSelectedPack] = useState<EvidencePack | null>(null);
  const [reasoning, setReasoning] = useState<ReasoningResponse | null>(null);
  const [jsonDialogOpen, setJsonDialogOpen] = useState(false);
  const [working, setWorking] = useState(false);

  async function loadNamespaces() {
    const namespaceResources = await client.listResources({ kind: "Namespace" });
    const allResources = await client.listResources();
    const options = deriveNamespaces(namespaceResources ?? [], allResources ?? []);
    setNamespaces(options);
    if (options.length > 0) {
      setNamespaceValue((current) => current || options[0].name);
      setPodNamespace((current) => current || options[0].name);
    }
  }

  async function loadPods(namespace: string) {
    if (!namespace) {
      setPods([]);
      setPodName("");
      return;
    }
    const podResources = await client.listResources({ kind: "Pod", namespace });
    const nextPods = podResources ?? [];
    setPods(nextPods);
    setPodName((current) => (current && nextPods.some((pod) => pod.name === current) ? current : nextPods[0]?.name ?? ""));
  }

  async function load() {
    setLoading(true);
    setError("");
    try {
      await loadNamespaces();
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load analysis inputs");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  useEffect(() => {
    void loadPods(podNamespace);
  }, [podNamespace]);

  async function handleAnalyzeNamespace() {
    if (!namespaceValue) {
      return;
    }
    setWorking(true);
    setError("");
    try {
      const pack = await client.generateEvidencePack({
        scopeType: "namespace",
        namespace: namespaceValue,
        persist: true,
      });
      if (pack) {
        setSelectedPack(pack);
        setReasoning(null);
      }
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to analyze namespace");
    } finally {
      setWorking(false);
    }
  }

  async function handleAnalyzePod() {
    if (!podNamespace || !podName) {
      return;
    }
    setWorking(true);
    setError("");
    try {
      const pack = await client.generateEvidencePack({
        scopeType: "pod",
        namespace: podNamespace,
        name: podName,
        persist: true,
      });
      if (pack) {
        setSelectedPack(pack);
        setReasoning(null);
      }
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to analyze pod");
    } finally {
      setWorking(false);
    }
  }

  async function handleReason() {
    if (!selectedPack) {
      return;
    }
    setWorking(true);
    setError("");
    try {
      const response = await client.reasonEvidencePack(selectedPack.id);
      if (response) {
        setReasoning(response);
      }
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to reason over evidence pack");
    } finally {
      setWorking(false);
    }
  }

  const packData = useMemo(() => asRecord(selectedPack?.packJson), [selectedPack]);
  const scope = asRecord(packData?.scope);
  const summary = asRecord(packData?.summary);
  const generatedAt = typeof packData?.generatedAt === "string" ? packData.generatedAt : selectedPack?.createdAt;
  const summaryCards: SummaryCard[] = scope?.type === "namespace"
    ? [
        { label: "Namespace", value: String(scope.namespace ?? "-") },
        { label: "Scope", value: selectedPack?.scopeType ?? "-" },
        { label: "Token estimate", value: selectedPack?.tokenEstimate ?? "-" },
        { label: "Created at", value: formatDateTime(selectedPack?.createdAt) },
        { label: "Evidence pack", value: selectedPack?.id ?? "-" },
        { label: "Pods", value: numeric(summary?.pods) },
        { label: "Deployments", value: numeric(summary?.deployments) },
        { label: "Services", value: numeric(summary?.services) },
        { label: "ConfigMaps", value: numeric(summary?.configmaps) },
        { label: "Warning events", value: numeric(summary?.warningEvents) },
        { label: "Open findings", value: numeric(summary?.openFindings) },
      ]
    : [
        { label: "Namespace", value: String(scope?.namespace ?? "-") },
        { label: "Pod", value: String(scope?.name ?? "-") },
        { label: "Scope", value: selectedPack?.scopeType ?? "-" },
        { label: "Token estimate", value: selectedPack?.tokenEstimate ?? "-" },
        { label: "Created at", value: formatDateTime(selectedPack?.createdAt) },
        { label: "Evidence pack", value: selectedPack?.id ?? "-" },
        { label: "Phase", value: String(asRecord(packData?.pod)?.phase ?? "-") },
        { label: "Ready", value: String(asRecord(packData?.pod)?.ready ?? "-") },
        { label: "Restart count", value: numeric(asRecord(packData?.pod)?.restartCount) },
      ];

  if (loading) {
    return <LoadingState label="Loading analysis inputs" />;
  }

  if (error && !selectedPack) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Analyze"
        subtitle="Generate compact namespace or pod context for reasoning."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      {error && <Alert severity="error">{error}</Alert>}
      {reasoning && <Alert severity="info">Mock reasoning only. No action is applied.</Alert>}

      <DataPanel>
          <Tabs value={tab} onChange={(_, next) => setTab(next)} sx={{ mb: 2 }}>
            <Tab label="Analyze Namespace" />
            <Tab label="Analyze Pod" />
          </Tabs>

          {tab === 0 ? (
            <Stack direction={{ xs: "column", lg: "row" }} spacing={1.5} alignItems={{ lg: "center" }}>
              <TextField
                select
                label="Namespace"
                value={namespaceValue}
                onChange={(event) => setNamespaceValue(event.target.value)}
                fullWidth
              >
                {namespaces.map((namespace) => (
                  <MenuItem key={namespace.name} value={namespace.name}>
                    {namespace.name}
                  </MenuItem>
                ))}
              </TextField>
              <Button type="button" variant="contained" onClick={() => void handleAnalyzeNamespace()} disabled={!namespaceValue || working}>
                Analyze namespace
              </Button>
            </Stack>
          ) : (
            <Stack direction={{ xs: "column", lg: "row" }} spacing={1.5} alignItems={{ lg: "center" }}>
              <TextField
                select
                label="Namespace"
                value={podNamespace}
                onChange={(event) => setPodNamespace(event.target.value)}
                fullWidth
              >
                {namespaces.map((namespace) => (
                  <MenuItem key={namespace.name} value={namespace.name}>
                    {namespace.name}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                select
                label="Pod"
                value={podName}
                onChange={(event) => setPodName(event.target.value)}
                fullWidth
              >
                {pods.map((pod) => (
                  <MenuItem key={pod.id} value={pod.name}>
                    {pod.name}
                  </MenuItem>
                ))}
              </TextField>
              <Button type="button" variant="contained" onClick={() => void handleAnalyzePod()} disabled={!podNamespace || !podName || working}>
                Analyze pod
              </Button>
            </Stack>
          )}
      </DataPanel>

      {!selectedPack ? (
        <EmptyState
          title="No analysis generated yet"
          message="Select a namespace or pod from live cluster inventory to generate a compact evidence summary."
        />
      ) : (
        <Stack spacing={2}>
          <Grid2 container spacing={1.5}>
            {summaryCards.map((card) => (
              <Grid2 key={card.label} size={{ xs: 12, sm: 6, lg: 3 }}>
                <StatCard title={card.label} value={card.value} />
              </Grid2>
            ))}
          </Grid2>

          <DataPanel>
              <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" spacing={1.5}>
                <Stack spacing={0.5}>
                  <Typography variant="h6">Generated summary</Typography>
                  <Typography color="text.secondary">
                    {selectedPack.scopeType === "namespace"
                      ? `Namespace evidence for ${selectedPack.scopeNamespace ?? "-"}`
                      : `Pod evidence for ${selectedPack.scopeNamespace ?? "-"}/${selectedPack.scopeName ?? "-"}`}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Generated at {formatDateTime(generatedAt)}
                  </Typography>
                </Stack>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  <Button type="button" variant="outlined" startIcon={<PreviewRoundedIcon />} onClick={() => setJsonDialogOpen(true)}>
                    View raw JSON
                  </Button>
                  <Button type="button" variant="contained" startIcon={<AutoAwesomeRoundedIcon />} onClick={() => void handleReason()} disabled={working}>
                    Reason over this {selectedPack.scopeType}
                  </Button>
                  </Stack>
                </Stack>
          </DataPanel>

          <EvidenceSections pack={selectedPack} reasoning={reasoning} />
        </Stack>
      )}

        <Dialog open={jsonDialogOpen} onClose={() => setJsonDialogOpen(false)} maxWidth="lg" fullWidth>
        <DialogTitle>Raw evidence pack JSON</DialogTitle>
        <DialogContent>
          {selectedPack && <JsonPreview value={selectedPack.packJson} />}
        </DialogContent>
      </Dialog>
    </Stack>
  );
}

function EvidenceSections({ pack, reasoning }: { pack: EvidencePack; reasoning: ReasoningResponse | null }) {
  const packData = asRecord(pack.packJson);
  const sections: Array<{ title: string; items: unknown[]; render: (item: unknown, index: number) => ReactNode }> = [
    {
      title: "Unhealthy pods",
      items: asArray(packData?.unhealthyPods),
      render: (item, index) => renderKeyValueRow(item, index),
    },
    {
      title: "Unavailable deployments",
      items: asArray(packData?.unavailableDeployments),
      render: (item, index) => renderKeyValueRow(item, index),
    },
    {
      title: "Recent warning events",
      items: asArray(packData?.recentWarningEvents),
      render: (item, index) => renderEventRow(item, index),
    },
    {
      title: "Restart-heavy pods",
      items: asArray(packData?.topRestartHeavyPods),
      render: (item, index) => renderKeyValueRow(item, index),
    },
    {
      title: "Related findings",
      items: asArray(packData?.relatedOpenFindings),
      render: (item, index) => renderKeyValueRow(item, index),
    },
    {
      title: "Suspected causes",
      items: asArray(packData?.suspectedDeterministicCauses),
      render: (item, index) => (
        <Typography key={`${item}-${index}`} variant="body2" color="text.secondary">
          • {String(item)}
        </Typography>
      ),
    },
  ];

  if (pack.scopeType === "pod") {
    const pod = asRecord(packData?.pod);
    const ownerRefs = asArray(packData?.ownerReferences);
    const services = asArray(packData?.relatedServiceCandidates);
    const containerStatuses = asArray(packData?.containerStatuses);
    sections.unshift(
      {
        title: "Pod summary",
        items: pod ? [pod] : [],
        render: (item, index) => renderKeyValueRow(item, index),
      },
      {
        title: "Owner references",
        items: ownerRefs,
        render: (item, index) => renderKeyValueRow(item, index),
      },
      {
        title: "Container statuses",
        items: containerStatuses,
        render: (item, index) => renderKeyValueRow(item, index),
      },
      {
        title: "Related service candidates",
        items: services,
        render: (item, index) => renderKeyValueRow(item, index),
      },
    );
  }

  return (
    <Stack spacing={2}>
      {reasoning && (
        <DataPanel title="Mock reasoning">
            <Stack spacing={1}>
              <Typography color="text.secondary">{reasoning.reasoning.rootCause}</Typography>
              <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                <StatusChip status={reasoning.reasoning.riskLevel} />
                <StatusChip status={`confidence ${Math.round(reasoning.reasoning.confidence * 100)}%`} />
                <StatusChip status={reasoning.actionPlan.status} />
              </Stack>
              <Typography variant="body2" color="text.secondary">
                Suggested action plan: {reasoning.actionPlan.title}
              </Typography>
              {reasoning.reasoning.suggestedActionPlan.map((step) => (
                <Typography key={step} variant="body2" color="text.secondary">
                  • {step}
                </Typography>
              ))}
            </Stack>
        </DataPanel>
      )}
      {sections
        .filter((section) => section.items.length > 0)
        .map((section) => (
          <DataPanel key={section.title} title={section.title}>
              <Stack spacing={1}>{section.items.map(section.render)}</Stack>
          </DataPanel>
        ))}
      {sections.every((section) => section.items.length === 0) && (
        <DataPanel>
            <Typography color="text.secondary">
              No structured sections were available in this evidence pack. Use View raw JSON for the fallback payload.
            </Typography>
        </DataPanel>
      )}
    </Stack>
  );
}

function deriveNamespaces(namespaceResources: Resource[], allResources: Resource[]): NamespaceOption[] {
  const byName = new Map<string, NamespaceOption>();
  for (const resource of namespaceResources) {
    if (resource.kind === "Namespace") {
      byName.set(resource.name, {
        name: resource.name,
        status: resource.status,
        observedAt: resource.observed_at,
      });
    }
  }
  for (const resource of allResources) {
    if (resource.namespace && !byName.has(resource.namespace)) {
      byName.set(resource.namespace, { name: resource.namespace });
    }
  }
  return [...byName.values()].sort((left, right) => left.name.localeCompare(right.name));
}

function asRecord(value: unknown): Record<string, unknown> | null {
  return value && typeof value === "object" && !Array.isArray(value) ? (value as Record<string, unknown>) : null;
}

function asArray(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function numeric(value: unknown): number | string {
  return typeof value === "number" ? value : "-";
}

function renderKeyValueRow(item: unknown, index: number) {
  const record = asRecord(item);
  if (!record) {
    return null;
  }
  return (
    <Card key={index} variant="outlined" sx={{ boxShadow: "none", bgcolor: "background.paper" }}>
      <CardContent sx={{ p: 1.5 }}>
        <Grid2 container spacing={1}>
          {Object.entries(record).map(([key, value]) => (
            <Grid2 key={key} size={{ xs: 12, sm: 6, lg: 4 }}>
              <Typography variant="body2" color="text.secondary">
                {key}
              </Typography>
              <Typography variant="body2" sx={{ wordBreak: "break-word" }}>
                {formatValue(value)}
              </Typography>
            </Grid2>
          ))}
        </Grid2>
      </CardContent>
    </Card>
  );
}

function renderEventRow(item: unknown, index: number) {
  const record = asRecord(item);
  if (!record) {
    return null;
  }
  return (
    <Card key={index} variant="outlined" sx={{ boxShadow: "none", bgcolor: "background.paper" }}>
      <CardContent sx={{ p: 1.5 }}>
        <Stack spacing={0.5}>
          <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
            {record.reason && <StatusChip status={record.reason as string} />}
            {record.kind && record.name && <Typography variant="body2">{String(record.kind)}/{String(record.name)}</Typography>}
          </Stack>
          <Typography variant="body2" color="text.secondary">
            {String(record.message ?? "-")}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
}

function formatValue(value: unknown): string {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  if (typeof value === "string" || typeof value === "number" || typeof value === "boolean") {
    return String(value);
  }
  if (Array.isArray(value)) {
    return value.length === 0 ? "-" : value.map((item) => formatValue(item)).join(", ");
  }
  if (typeof value === "object") {
    return JSON.stringify(value);
  }
  return String(value);
}
