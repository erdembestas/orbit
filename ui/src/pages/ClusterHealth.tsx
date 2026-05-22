import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import WarningAmberRoundedIcon from "@mui/icons-material/WarningAmberRounded";
import {
  Alert,
  Button,
  Card,
  CardContent,
  Grid2,
  Stack,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import {
  formatDateTime,
  type ApiError,
  type ClusterHealthResponse,
  type ClusterHealthSnapshot,
  type NamespaceHealthSnapshot,
  type NodeHealthSnapshot,
  type OrbitApiClient,
} from "../api/client";
import CompactStat from "../components/CompactStat";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import HealthStatusChip from "../components/HealthStatusChip";
import LoadingState from "../components/LoadingState";
import MetricBar from "../components/MetricBar";
import PageHeader from "../components/PageHeader";
import ResponsiveDataView, { type ResponsiveColumn } from "../components/ResponsiveDataView";

type ClusterHealthPageProps = {
  client: OrbitApiClient;
};

export default function ClusterHealthPage({ client }: ClusterHealthPageProps) {
  const [report, setReport] = useState<ClusterHealthResponse | null | undefined>(undefined);
  const [nodes, setNodes] = useState<NodeHealthSnapshot[] | null>(null);
  const [namespaces, setNamespaces] = useState<NamespaceHealthSnapshot[] | null>(null);
  const [history, setHistory] = useState<ClusterHealthSnapshot[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const [reportResponse, nodeResponse, namespaceResponse, historyResponse] = await Promise.all([
        client.getClusterHealth(),
        client.getClusterHealthNodes(),
        client.getClusterHealthNamespaces(),
        client.getClusterHealthHistory(20),
      ]);
      setReport(reportResponse);
      setNodes(nodeResponse ?? []);
      setNamespaces(namespaceResponse ?? []);
      setHistory(historyResponse ?? []);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load cluster health");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  if (loading) {
    return <LoadingState label="Loading cluster health" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (report === null) {
    return (
      <EmptyState
        title="Cluster health not implemented"
        message="The backend did not expose cluster health APIs."
      />
    );
  }

  if (!report?.observedAt) {
    return (
      <EmptyState
        title="No cluster health snapshot yet"
        message="No cluster health snapshot has been collected yet. Wait for the controller interval and refresh."
      />
    );
  }

  const nodeColumns: ResponsiveColumn<NodeHealthSnapshot>[] = [
    { key: "node", label: "Node", render: (node) => <Typography fontWeight={700}>{node.nodeName}</Typography> },
    { key: "ready", label: "Ready", render: (node) => <HealthStatusChip status={node.ready ? "healthy" : "critical"} /> },
    { key: "health", label: "Health", render: (node) => <HealthStatusChip status={node.healthStatus} /> },
    { key: "score", label: "Score", render: (node) => node.healthScore, align: "right" },
    { key: "cpu", label: "CPU %", render: (node) => formatPercent(node.cpuUsagePercent), align: "right" },
    { key: "memory", label: "Memory %", render: (node) => formatPercent(node.memoryUsagePercent), align: "right" },
    { key: "pods", label: "Pods", render: (node) => node.podCount, align: "right" },
    { key: "pressure", label: "Pressure", render: (node) => summarizePressure(node.pressureFlags) },
    { key: "observedAt", label: "Observed At", render: (node) => formatDateTime(node.observedAt), mobilePriority: false },
  ];

  const namespaceColumns: ResponsiveColumn<NamespaceHealthSnapshot>[] = [
    { key: "namespace", label: "Namespace", render: (item) => <Typography fontWeight={700}>{item.namespace}</Typography> },
    { key: "health", label: "Health", render: (item) => <HealthStatusChip status={item.healthStatus} /> },
    { key: "score", label: "Score", render: (item) => item.healthScore, align: "right" },
    { key: "pods", label: "Pods", render: (item) => item.podCount, align: "right" },
    { key: "running", label: "Running", render: (item) => item.runningPodCount, align: "right" },
    { key: "pending", label: "Pending", render: (item) => item.pendingPodCount, align: "right" },
    { key: "failed", label: "Failed", render: (item) => item.failedPodCount, align: "right" },
    { key: "restarts", label: "Restarts", render: (item) => item.restartCount, align: "right" },
    { key: "events", label: "Warning Events", render: (item) => item.warningEventCount, align: "right" },
    { key: "cpu", label: "CPU", render: (item) => formatMillicores(item.usedCpuMillicores), align: "right", mobilePriority: false },
    { key: "memory", label: "Memory", render: (item) => formatBytes(item.usedMemoryBytes), align: "right", mobilePriority: false },
  ];

  const historyColumns: ResponsiveColumn<ClusterHealthSnapshot>[] = [
    { key: "observedAt", label: "Observed At", render: (item) => formatDateTime(item.observedAt) },
    { key: "status", label: "Status", render: (item) => <HealthStatusChip status={item.healthStatus} /> },
    { key: "score", label: "Score", render: (item) => item.healthScore, align: "right" },
    { key: "cpu", label: "CPU %", render: (item) => formatPercent(item.cpuUsagePercent), align: "right" },
    { key: "memory", label: "Memory %", render: (item) => formatPercent(item.memoryUsagePercent), align: "right" },
    { key: "pods", label: "Pods", render: (item) => item.podCount, align: "right" },
    { key: "events", label: "Warning Events", render: (item) => item.warningEventCount, align: "right" },
    { key: "metrics", label: "Metrics", render: (item) => <HealthStatusChip status={item.metricsAvailable ? "healthy" : "unknown"} /> },
  ];

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Cluster Health"
        subtitle="Core node, namespace, and workload pressure summary."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      {!report.metricsAvailable && (
        <Alert icon={<WarningAmberRoundedIcon />} severity="warning">
          Metrics API is unavailable. Orbit is using node conditions, pod phases, and events only.
        </Alert>
      )}

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <CompactStat label="Health score" value={report.healthScore} hint={<HealthStatusChip status={report.healthStatus} />} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <CompactStat label="Observed at" value={formatDateTime(report.observedAt)} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <CompactStat label="Nodes ready" value={`${report.readyNodeCount} / ${report.nodeCount}`} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <CompactStat label="Pods" value={report.podCount} hint={`${report.runningPodCount} running`} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <CompactStat label="Metrics" value={report.metricsAvailable ? "Available" : "Unavailable"} />
        </Grid2>
      </Grid2>

      <Card>
        <CardContent sx={{ p: 2 }}>
          <Stack spacing={1.5}>
            <Typography variant="h6">Resource summary</Typography>
            <Grid2 container spacing={1.25}>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Node count" value={report.nodeCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Ready nodes" value={report.readyNodeCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Not ready" value={report.notReadyNodeCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Warning events" value={report.warningEventCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Running pods" value={report.runningPodCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Pending pods" value={report.pendingPodCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Failed pods" value={report.failedPodCount} /></Grid2>
              <Grid2 size={{ xs: 6, md: 3 }}><CompactStat label="Cluster" value={report.cluster.name || "-"} /></Grid2>
            </Grid2>
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2 }}>
          <Stack spacing={1.75}>
            <Typography variant="h6">CPU / Memory</Typography>
            <Grid2 container spacing={1.5}>
              <Grid2 size={{ xs: 12, md: 6 }}>
                <Stack spacing={1.25}>
                  <MetricBar label="CPU usage" value={report.cpuUsagePercent} />
                  <Typography variant="body2" color="text.secondary">
                    Allocatable: {formatMillicores(report.allocatableCpuMillicores)} · Used: {formatMillicores(report.usedCpuMillicores)}
                  </Typography>
                </Stack>
              </Grid2>
              <Grid2 size={{ xs: 12, md: 6 }}>
                <Stack spacing={1.25}>
                  <MetricBar label="Memory usage" value={report.memoryUsagePercent} />
                  <Typography variant="body2" color="text.secondary">
                    Allocatable: {formatBytes(report.allocatableMemoryBytes)} · Used: {formatBytes(report.usedMemoryBytes)}
                  </Typography>
                </Stack>
              </Grid2>
            </Grid2>
            {report.metricsError ? (
              <Typography variant="body2" color="text.secondary">
                Metrics detail: {report.metricsError}
              </Typography>
            ) : null}
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2 }}>
          <Stack spacing={1.5}>
            <Typography variant="h6">Node health</Typography>
            {nodes && nodes.length > 0 ? (
              <ResponsiveDataView
                rows={nodes}
                columns={nodeColumns}
                getRowId={(item) => item.id}
                renderMobileTitle={(item) => item.nodeName}
                renderMobileSubtitle={(item) => `${item.podCount} pods`}
              />
            ) : (
              <EmptyState title="No node health rows" message="No node health snapshots have been collected yet." />
            )}
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2 }}>
          <Stack spacing={1.5}>
            <Typography variant="h6">Namespace health</Typography>
            {namespaces && namespaces.length > 0 ? (
              <ResponsiveDataView
                rows={namespaces}
                columns={namespaceColumns}
                getRowId={(item) => item.id}
                renderMobileTitle={(item) => item.namespace}
                renderMobileSubtitle={(item) => `${item.podCount} pods • ${item.restartCount} restarts`}
              />
            ) : (
              <EmptyState title="No namespace health rows" message="No namespace health snapshots have been collected yet." />
            )}
          </Stack>
        </CardContent>
      </Card>

      <Card>
        <CardContent sx={{ p: 2 }}>
          <Stack spacing={1.5}>
            <Typography variant="h6">Health history</Typography>
            {history && history.length > 0 ? (
              <ResponsiveDataView
                rows={history}
                columns={historyColumns}
                getRowId={(item) => item.id}
                renderMobileTitle={(item) => formatDateTime(item.observedAt)}
                renderMobileSubtitle={(item) => `${item.healthStatus} • ${item.healthScore}`}
              />
            ) : (
              <EmptyState title="No health history yet" message="Wait for the next controller interval and refresh." />
            )}
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}

function formatPercent(value: number | null | undefined) {
  return typeof value === "number" ? `${value.toFixed(1)}%` : "-";
}

function formatMillicores(value: number | null | undefined) {
  return typeof value === "number" ? `${Math.round(value)}m` : "-";
}

function formatBytes(value: number | null | undefined) {
  if (typeof value !== "number") {
    return "-";
  }
  const units = ["B", "KiB", "MiB", "GiB", "TiB"];
  let current = value;
  let unit = 0;
  while (current >= 1024 && unit < units.length - 1) {
    current /= 1024;
    unit += 1;
  }
  return `${current.toFixed(current >= 10 || unit === 0 ? 0 : 1)} ${units[unit]}`;
}

function summarizePressure(value: Record<string, unknown>) {
  const active = Object.entries(value)
    .filter(([, enabled]) => enabled === true)
    .map(([key]) => key.replace(/Pressure$/, ""));
  return active.length > 0 ? active.join(", ") : "-";
}
