import HubRoundedIcon from "@mui/icons-material/HubRounded";
import PriorityHighRoundedIcon from "@mui/icons-material/PriorityHighRounded";
import ReportProblemRoundedIcon from "@mui/icons-material/ReportProblemRounded";
import SyncRoundedIcon from "@mui/icons-material/SyncRounded";
import WarningAmberRoundedIcon from "@mui/icons-material/WarningAmberRounded";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Grid2,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useEffect, useMemo, useState } from "react";
import {
  formatDateTime,
  type ApiError,
  type ClusterHealthResponse,
  type ClusterHealthSnapshot,
  type ControllerStatus,
  type EvidencePack,
  type Finding,
  type InfoResponse,
  type MeResponse,
  type OrbitApiClient,
} from "../api/client";
import DataPanel from "../components/DataPanel";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import HealthStatusChip from "../components/HealthStatusChip";
import LoadingState from "../components/LoadingState";
import MetricBar from "../components/MetricBar";
import PageHeader from "../components/PageHeader";
import SeverityChip from "../components/SeverityChip";
import StatCard from "../components/StatCard";
import StatusChip from "../components/StatusChip";

type DashboardPageProps = {
  client: OrbitApiClient;
  me: MeResponse | null;
};

const timeRangeOptions = ["Last 24 hours"];

export default function DashboardPage({ client, me }: DashboardPageProps) {
  const [info, setInfo] = useState<InfoResponse | null>(null);
  const [status, setStatus] = useState<ControllerStatus | null | undefined>(undefined);
  const [clusterHealth, setClusterHealth] = useState<ClusterHealthResponse | null | undefined>(undefined);
  const [history, setHistory] = useState<ClusterHealthSnapshot[] | null>(null);
  const [findings, setFindings] = useState<Finding[] | null>(null);
  const [evidencePacks, setEvidencePacks] = useState<EvidencePack[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [timeRange, setTimeRange] = useState(timeRangeOptions[0]);

  async function load() {
    setLoading(true);
    setError("");
    try {
      const [
        infoResponse,
        statusResponse,
        clusterHealthResponse,
        historyResponse,
        findingsResponse,
        packsResponse,
      ] = await Promise.all([
        client.getInfo(),
        client.getControllerStatus(),
        client.getClusterHealth(),
        client.getClusterHealthHistory(12),
        client.listFindings(),
        client.listEvidencePacks(),
      ]);
      setInfo(infoResponse);
      setStatus(statusResponse);
      setClusterHealth(clusterHealthResponse);
      setHistory(historyResponse ?? []);
      setFindings(findingsResponse ?? []);
      setEvidencePacks(packsResponse ?? []);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load dashboard");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const severityCounts = useMemo(() => {
    const base = { critical: 0, high: 0, medium: 0, low: 0, info: 0 };
    for (const finding of findings ?? []) {
      const key = finding.severity.toLowerCase() as keyof typeof base;
      if (key in base) {
        base[key] += 1;
      }
    }
    return base;
  }, [findings]);

  const criticalFindings = severityCounts.critical;
  const warningFindings = severityCounts.high + severityCounts.medium;

  const actionItems = useMemo(() => {
    const items: ActionItem[] = [];
    if (clusterHealth && !clusterHealth.metricsAvailable) {
      items.push({
        severity: "medium",
        title: "Metrics API unavailable",
        resource: info?.cluster_name ?? "cluster",
        description: "Orbit is using node conditions, pod phases, and events only.",
        timestamp: clusterHealth.observedAt,
        action: "Review",
      });
    }
    if (clusterHealth && clusterHealth.notReadyNodeCount > 0) {
      items.push({
        severity: "critical",
        title: "Nodes are not ready",
        resource: `${clusterHealth.notReadyNodeCount} node(s)`,
        description: "At least one node is reporting NotReady in the latest health snapshot.",
        timestamp: clusterHealth.observedAt,
        action: "Open",
      });
    }
    if (clusterHealth && clusterHealth.failedPodCount > 0) {
      items.push({
        severity: "high",
        title: "Failed pods detected",
        resource: `${clusterHealth.failedPodCount} failed`,
        description: "Cluster health includes failed workloads that need investigation.",
        timestamp: clusterHealth.observedAt,
        action: "Review",
      });
    }

    const rankedFindings = [...(findings ?? [])].sort((left, right) => {
      const severityDelta = severityRank(right.severity) - severityRank(left.severity);
      if (severityDelta !== 0) {
        return severityDelta;
      }
      return new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime();
    });

    for (const finding of rankedFindings.slice(0, 5)) {
      items.push({
        severity: finding.severity,
        title: finding.title,
        resource: [finding.resource_namespace, finding.resource_name].filter(Boolean).join("/") || finding.resource_kind || "cluster",
        description: finding.description,
        timestamp: finding.updated_at,
        action: severityRank(finding.severity) >= 4 ? "Open" : "Review",
      });
    }

    return items.slice(0, 6);
  }, [clusterHealth, findings, info?.cluster_name]);

  const recentFindings = useMemo(
    () =>
      [...(findings ?? [])]
        .sort((left, right) => new Date(right.updated_at).getTime() - new Date(left.updated_at).getTime())
        .slice(0, 5),
    [findings],
  );

  const recentEvidencePacks = useMemo(
    () =>
      [...(evidencePacks ?? [])]
        .sort((left, right) => new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime())
        .slice(0, 5),
    [evidencePacks],
  );

  const counts = status?.resource_counts_by_kind ?? {};

  if (loading) {
    return <LoadingState label="Loading dashboard" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (!clusterHealth) {
    return (
      <Stack spacing={2.5}>
        <PageHeader title="Dashboard" subtitle="Real-time overview of cluster health and operations." />
        <EmptyState
          title="No dashboard data"
          message="Orbit API is reachable, but cluster health and inventory data are not available yet."
          actionLabel="Refresh"
          onAction={() => void load()}
        />
      </Stack>
    );
  }

  return (
    <Stack spacing={2.25}>
      <PageHeader
        title="Dashboard"
        subtitle="Real-time overview of cluster health and operations."
        actions={
          <>
            <Typography variant="caption" sx={{ alignSelf: "center", mr: 0.5 }}>
              Last updated: {formatRelative(clusterHealth.observedAt)}
            </Typography>
            <TextField
              select
              size="small"
              value={timeRange}
              onChange={(event) => setTimeRange(event.target.value)}
              sx={{ minWidth: 148 }}
            >
              {timeRangeOptions.map((option) => (
                <MenuItem key={option} value={option}>
                  {option}
                </MenuItem>
              ))}
            </TextField>
            <Button variant="outlined" startIcon={<SyncRoundedIcon />} onClick={() => void load()}>
              Refresh
            </Button>
          </>
        }
      />

      {!clusterHealth.metricsAvailable && (
        <Alert severity="warning" variant="outlined">
          Metrics API is unavailable. Orbit is using node conditions, pod phases, and events only.
        </Alert>
      )}

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, md: 6, xl: 3 }}>
          <StatCard
            title="Health Score"
            value={clusterHealth.healthScore}
            subtitle={statusText(clusterHealth.healthStatus)}
            helper={`${clusterHealth.healthStatus} status`}
            accent={toneColor(clusterHealth.healthStatus)}
            visual={<ScoreRing score={clusterHealth.healthScore} color={toneColor(clusterHealth.healthStatus)} />}
          />
        </Grid2>
        <Grid2 size={{ xs: 12, md: 6, xl: 3 }}>
          <StatCard
            title="Healthy Nodes"
            value={`${clusterHealth.readyNodeCount} / ${clusterHealth.nodeCount}`}
            subtitle={`${Math.round(percent(clusterHealth.readyNodeCount, clusterHealth.nodeCount))}% healthy`}
            helper={`${clusterHealth.notReadyNodeCount} not ready`}
            icon={<HubRoundedIcon />}
            accent="#14B8A6"
          />
        </Grid2>
        <Grid2 size={{ xs: 12, md: 6, xl: 3 }}>
          <StatCard
            title="Warning Findings"
            value={warningFindings}
            subtitle={`${severityCounts.high} high • ${severityCounts.medium} medium`}
            helper={`${severityCounts.low} low findings`}
            icon={<WarningAmberRoundedIcon />}
            accent="#F59E0B"
          />
        </Grid2>
        <Grid2 size={{ xs: 12, md: 6, xl: 3 }}>
          <StatCard
            title="Critical Findings"
            value={criticalFindings}
            subtitle={criticalFindings > 0 ? "Requires immediate review" : "No critical findings"}
            helper={`${findings?.length ?? 0} findings total`}
            icon={<PriorityHighRoundedIcon />}
            accent="#EF4444"
          />
        </Grid2>
      </Grid2>

      <DataPanel
        title="Action Items"
        actions={
          <Typography variant="caption" color="text.secondary">
            {actionItems.length} items
          </Typography>
        }
      >
        {actionItems.length === 0 ? (
          <EmptyState title="No action items" message="No current cluster warnings or findings require review." />
        ) : (
          <Stack spacing={0}>
            {actionItems.map((item, index) => (
              <Box
                key={`${item.title}-${index}`}
                sx={{
                  display: "grid",
                  gridTemplateColumns: { xs: "1fr", lg: "120px 1.4fr 1.1fr 1.7fr 160px 80px" },
                  gap: 1.25,
                  py: 1.15,
                  borderTop: index === 0 ? "none" : "1px solid",
                  borderColor: "divider",
                  alignItems: "center",
                }}
              >
                <SeverityChip severity={item.severity} />
                <Typography fontWeight={600} sx={{ fontSize: 13 }}>
                  {item.title}
                </Typography>
                <Typography color="text.secondary" sx={{ fontSize: 12.5 }}>
                  {item.resource}
                </Typography>
                <Typography color="text.secondary" sx={{ fontSize: 12.5 }}>
                  {item.description}
                </Typography>
                <Typography color="text.secondary" sx={{ fontSize: 12.5 }}>
                  {formatRelative(item.timestamp)}
                </Typography>
                <Typography sx={{ color: "primary.main", fontSize: 12.5, fontWeight: 600 }}>{item.action}</Typography>
              </Box>
            ))}
          </Stack>
        )}
      </DataPanel>

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, xl: 3.5 }}>
          <DataPanel title="Findings by Severity">
            <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
              <SeverityDonut counts={severityCounts} />
              <Stack spacing={0.9} sx={{ minWidth: 180 }}>
                {severityEntries(severityCounts).map(([severity, count]) => (
                  <Stack key={severity} direction="row" justifyContent="space-between" spacing={2}>
                    <Stack direction="row" spacing={1} alignItems="center">
                      <Box sx={{ width: 8, height: 8, borderRadius: "50%", bgcolor: severityColor(severity) }} />
                      <Typography sx={{ fontSize: 12.5, textTransform: "capitalize" }}>{severity}</Typography>
                    </Stack>
                    <Typography sx={{ fontSize: 12.5, color: "text.secondary" }}>{count}</Typography>
                  </Stack>
                ))}
              </Stack>
            </Stack>
          </DataPanel>
        </Grid2>

        <Grid2 size={{ xs: 12, xl: 4.5 }}>
          <DataPanel title="Cluster Health Over Time">
            {history && history.length > 0 ? (
              <Stack spacing={1.5}>
                <HistorySparkline rows={history} />
                <Stack direction="row" spacing={2} flexWrap="wrap" useFlexGap>
                  <MetricKpi label="Latest" value={history[0]?.healthScore ?? "-"} />
                  <MetricKpi label="Max" value={Math.max(...history.map((item) => item.healthScore))} />
                  <MetricKpi label="Min" value={Math.min(...history.map((item) => item.healthScore))} />
                </Stack>
              </Stack>
            ) : (
              <EmptyState title="No health history yet" message="Wait for the next controller interval and refresh." />
            )}
          </DataPanel>
        </Grid2>

        <Grid2 size={{ xs: 12, xl: 4 }}>
          <DataPanel title="Resource Utilization">
            <Stack spacing={1.5}>
              <MetricBar label="CPU usage" value={clusterHealth.cpuUsagePercent} />
              <MetricBar label="Memory usage" value={clusterHealth.memoryUsagePercent} />
              <Stack direction="row" spacing={1.5} flexWrap="wrap" useFlexGap>
                <StatusChip status={`${clusterHealth.podCount} pods`} />
                <StatusChip status={`${clusterHealth.nodeCount} nodes`} />
                <StatusChip status={clusterHealth.metricsAvailable ? "metrics available" : "metrics unavailable"} />
              </Stack>
            </Stack>
          </DataPanel>
        </Grid2>
      </Grid2>

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, xl: 7 }}>
          <DataPanel title="Recent Findings" subtitle="Latest detected deterministic issues.">
            {recentFindings.length === 0 ? (
              <EmptyState title="No findings yet" message="Findings will appear here after controller scans and rule evaluation." />
            ) : (
              <Stack spacing={0}>
                {recentFindings.map((finding, index) => (
                  <Box
                    key={finding.id}
                    sx={{
                      display: "grid",
                      gridTemplateColumns: { xs: "1fr", lg: "110px 1.5fr 1.3fr 120px 80px" },
                      gap: 1.25,
                      py: 1.1,
                      borderTop: index === 0 ? "none" : "1px solid",
                      borderColor: "divider",
                      alignItems: "center",
                    }}
                  >
                    <SeverityChip severity={finding.severity} />
                    <Typography sx={{ fontSize: 13, fontWeight: 600 }}>{finding.title}</Typography>
                    <Typography color="text.secondary" sx={{ fontSize: 12.5 }}>
                      {[finding.resource_namespace, finding.resource_name].filter(Boolean).join("/") || finding.resource_kind || "-"}
                    </Typography>
                    <Typography color="text.secondary" sx={{ fontSize: 12.5 }}>
                      {formatRelative(finding.updated_at)}
                    </Typography>
                    <StatusChip status={finding.status} />
                  </Box>
                ))}
              </Stack>
            )}
          </DataPanel>
        </Grid2>

        <Grid2 size={{ xs: 12, xl: 5 }}>
          <DataPanel title="Evidence Packs" subtitle="Most recent compact reasoning contexts.">
            {recentEvidencePacks.length === 0 ? (
              <EmptyState title="No evidence packs yet" message="Generate a namespace or pod analysis to populate this list." />
            ) : (
              <Stack spacing={0}>
                {recentEvidencePacks.map((pack, index) => (
                  <Box
                    key={pack.id}
                    sx={{
                      display: "grid",
                      gridTemplateColumns: { xs: "1fr", lg: "1.4fr 120px 100px" },
                      gap: 1.25,
                      py: 1.1,
                      borderTop: index === 0 ? "none" : "1px solid",
                      borderColor: "divider",
                      alignItems: "center",
                    }}
                  >
                    <Stack spacing={0.35}>
                      <Typography sx={{ fontSize: 13, fontWeight: 600 }}>
                        {pack.scopeName ?? pack.scopeNamespace ?? pack.scopeType}
                      </Typography>
                      <Typography variant="caption">
                        {pack.scopeType} • {pack.scopeNamespace ?? "cluster-scope"}
                      </Typography>
                    </Stack>
                    <Typography sx={{ fontSize: 12.5, color: "text.secondary" }}>
                      {formatDateTime(pack.createdAt)}
                    </Typography>
                    <StatusChip status={`${pack.tokenEstimate} tokens`} />
                  </Box>
                ))}
              </Stack>
            )}
          </DataPanel>
        </Grid2>
      </Grid2>

      <DataPanel title="Runtime Summary">
        <Grid2 container spacing={1.25}>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Cluster" value={info?.cluster_name ?? "-"} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Mode" value={status?.mode ?? "single-cluster"} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Environment" value={info?.environment ?? "-"} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Logged-in User" value={me?.username ?? "-"} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Cluster Type" value={info?.cluster_type ?? "-"} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Auth Mode" value={info?.auth_mode ?? "-"} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Deployments" value={counts.Deployment ?? 0} />
          </Grid2>
          <Grid2 size={{ xs: 12, sm: 6, md: 3 }}>
            <MetricKpi label="Services" value={counts.Service ?? 0} />
          </Grid2>
        </Grid2>
      </DataPanel>
    </Stack>
  );
}

type ActionItem = {
  severity: string;
  title: string;
  resource: string;
  description: string;
  timestamp: string;
  action: string;
};

function MetricKpi({ label, value }: { label: string; value: string | number }) {
  return (
    <Stack spacing={0.35}>
      <Typography variant="caption" sx={{ textTransform: "uppercase", letterSpacing: "0.06em" }}>
        {label}
      </Typography>
      <Typography sx={{ fontSize: 15, fontWeight: 600 }}>{value}</Typography>
    </Stack>
  );
}

function ScoreRing({ score, color }: { score: number; color: string }) {
  const clamped = Math.max(0, Math.min(100, score));
  return (
    <Box
      sx={{
        width: 58,
        height: 58,
        borderRadius: "50%",
        display: "grid",
        placeItems: "center",
        background: `conic-gradient(${color} ${clamped * 3.6}deg, rgba(255,255,255,0.06) 0deg)`,
        position: "relative",
        flexShrink: 0,
      }}
    >
      <Box
        sx={{
          position: "absolute",
          inset: 5,
          borderRadius: "50%",
          bgcolor: "background.paper",
          border: "1px solid",
          borderColor: "divider",
        }}
      />
      <Stack spacing={0} alignItems="center" sx={{ position: "relative" }}>
        <Typography sx={{ fontSize: 18, fontWeight: 700, lineHeight: 1 }}>{score}</Typography>
        <Typography sx={{ fontSize: 10, color: "text.secondary", lineHeight: 1 }}>100</Typography>
      </Stack>
    </Box>
  );
}

function SeverityDonut({ counts }: { counts: Record<string, number> }) {
  const entries = severityEntries(counts);
  const total = entries.reduce((sum, [, count]) => sum + count, 0);
  const gradient = buildDonutGradient(entries, total);

  return (
    <Stack alignItems="center" spacing={1}>
      <Box
        sx={{
          width: 132,
          height: 132,
          borderRadius: "50%",
          background: gradient,
          display: "grid",
          placeItems: "center",
        }}
      >
        <Box
          sx={{
            width: 86,
            height: 86,
            borderRadius: "50%",
            bgcolor: "background.paper",
            border: "1px solid",
            borderColor: "divider",
            display: "grid",
            placeItems: "center",
          }}
        >
          <Stack spacing={0} alignItems="center">
            <Typography sx={{ fontSize: 26, fontWeight: 700, lineHeight: 1 }}>{total}</Typography>
            <Typography variant="caption">Total</Typography>
          </Stack>
        </Box>
      </Box>
    </Stack>
  );
}

function HistorySparkline({ rows }: { rows: ClusterHealthSnapshot[] }) {
  const ordered = [...rows].reverse();
  const width = 480;
  const height = 120;
  const values = ordered.map((item) => item.healthScore);
  const max = Math.max(...values, 100);
  const min = Math.min(...values, 0);
  const range = Math.max(max - min, 1);
  const points = ordered
    .map((item, index) => {
      const x = (index / Math.max(ordered.length - 1, 1)) * width;
      const y = height - ((item.healthScore - min) / range) * (height - 20) - 10;
      return `${x},${y}`;
    })
    .join(" ");

  return (
    <Box sx={{ position: "relative", width: "100%", overflow: "hidden" }}>
      <svg viewBox={`0 0 ${width} ${height}`} width="100%" height="140" preserveAspectRatio="none">
        <defs>
          <linearGradient id="healthArea" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="rgba(20, 184, 166, 0.28)" />
            <stop offset="100%" stopColor="rgba(20, 184, 166, 0.02)" />
          </linearGradient>
        </defs>
        <path
          d={`M ${points} L ${width},${height} L 0,${height} Z`}
          fill="url(#healthArea)"
          stroke="none"
        />
        <polyline
          fill="none"
          stroke="#14B8A6"
          strokeWidth="2.5"
          strokeLinejoin="round"
          strokeLinecap="round"
          points={points}
        />
      </svg>
      <Stack direction="row" justifyContent="space-between" sx={{ mt: -0.5 }}>
        <Typography variant="caption">{formatShortTime(ordered[0]?.observedAt)}</Typography>
        <Typography variant="caption">{formatShortTime(ordered[ordered.length - 1]?.observedAt)}</Typography>
      </Stack>
    </Box>
  );
}

function severityEntries(counts: Record<string, number>) {
  return ["critical", "high", "medium", "low", "info"].map((severity) => [severity, counts[severity] ?? 0] as const);
}

function buildDonutGradient(entries: ReadonlyArray<readonly [string, number]>, total: number) {
  if (total === 0) {
    return "conic-gradient(rgba(255,255,255,0.06) 0deg 360deg)";
  }
  let cursor = 0;
  const segments = entries
    .filter(([, count]) => count > 0)
    .map(([severity, count]) => {
      const angle = (count / total) * 360;
      const segment = `${severityColor(severity)} ${cursor}deg ${cursor + angle}deg`;
      cursor += angle;
      return segment;
    });
  return `conic-gradient(${segments.join(", ")})`;
}

function severityColor(severity: string) {
  switch (severity) {
    case "critical":
      return "#EF4444";
    case "high":
      return "#F97316";
    case "medium":
      return "#F59E0B";
    case "low":
      return "#14B8A6";
    default:
      return "#60A5FA";
  }
}

function severityRank(severity: string) {
  switch (severity.toLowerCase()) {
    case "critical":
      return 5;
    case "high":
      return 4;
    case "medium":
      return 3;
    case "low":
      return 2;
    default:
      return 1;
  }
}

function toneColor(status: string) {
  switch (status) {
    case "healthy":
      return "#22C55E";
    case "warning":
      return "#F59E0B";
    case "critical":
      return "#EF4444";
    default:
      return "#60A5FA";
  }
}

function percent(part: number, total: number) {
  if (!total) {
    return 0;
  }
  return (part / total) * 100;
}

function statusText(status: string) {
  switch (status) {
    case "healthy":
      return "Good";
    case "warning":
      return "Warning";
    case "critical":
      return "Critical";
    default:
      return "Unknown";
  }
}

function formatRelative(value: string) {
  const timestamp = new Date(value).getTime();
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  const deltaMinutes = Math.max(0, Math.round((Date.now() - timestamp) / 60000));
  if (deltaMinutes < 1) {
    return "just now";
  }
  if (deltaMinutes < 60) {
    return `${deltaMinutes}m ago`;
  }
  const hours = Math.round(deltaMinutes / 60);
  if (hours < 24) {
    return `${hours}h ago`;
  }
  const days = Math.round(hours / 24);
  return `${days}d ago`;
}

function formatShortTime(value: string | undefined) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return "-";
  }
  return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}
