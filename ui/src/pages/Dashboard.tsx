import HubRoundedIcon from "@mui/icons-material/HubRounded";
import Inventory2RoundedIcon from "@mui/icons-material/Inventory2Rounded";
import LayersRoundedIcon from "@mui/icons-material/LayersRounded";
import PreviewRoundedIcon from "@mui/icons-material/PreviewRounded";
import ReportProblemRoundedIcon from "@mui/icons-material/ReportProblemRounded";
import ShowChartRoundedIcon from "@mui/icons-material/ShowChartRounded";
import SettingsEthernetRoundedIcon from "@mui/icons-material/SettingsEthernetRounded";
import ViewStreamRoundedIcon from "@mui/icons-material/ViewStreamRounded";
import {
  Alert,
  Card,
  CardContent,
  Grid2,
  Stack,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import {
  formatNullableTime,
  type ApiError,
  type ClusterHealthResponse,
  type ControllerStatus,
  type InfoResponse,
  type MeResponse,
  type OrbitApiClient,
} from "../api/client";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import HealthStatusChip from "../components/HealthStatusChip";
import LoadingState from "../components/LoadingState";
import MetricBar from "../components/MetricBar";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";
import StatusChip from "../components/StatusChip";

type DashboardPageProps = {
  client: OrbitApiClient;
  me: MeResponse | null;
};

export default function DashboardPage({ client, me }: DashboardPageProps) {
  const [info, setInfo] = useState<InfoResponse | null>(null);
  const [status, setStatus] = useState<ControllerStatus | null | undefined>(undefined);
  const [clusterHealth, setClusterHealth] = useState<ClusterHealthResponse | null | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const [infoResponse, statusResponse, clusterHealthResponse] = await Promise.all([
        client.getInfo(),
        client.getControllerStatus(),
        client.getClusterHealth(),
      ]);
      setInfo(infoResponse);
      setStatus(statusResponse);
      setClusterHealth(clusterHealthResponse);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load dashboard");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  if (loading) {
    return <LoadingState label="Loading overview" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  const counts = status?.resource_counts_by_kind ?? {};

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Overview"
        subtitle="Single-cluster control plane health and inventory summary."
      />

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Cluster" value={info?.cluster_name ?? "-"} icon={<HubRoundedIcon />} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Mode" value={status?.mode ?? "single-cluster"} icon={<LayersRoundedIcon />} accent="#6C5CFF" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Last seen" value={formatNullableTime(status?.last_seen_at)} icon={<SettingsEthernetRoundedIcon />} accent="#6172F3" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Open findings" value={status?.open_finding_count ?? 0} icon={<ReportProblemRoundedIcon />} accent="#F59E0B" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Evidence packs" value={status?.latest_evidence_pack_count ?? 0} icon={<PreviewRoundedIcon />} accent="#12B76A" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Pods" value={counts.Pod ?? 0} icon={<ViewStreamRoundedIcon />} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Deployments" value={counts.Deployment ?? 0} icon={<Inventory2RoundedIcon />} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Services" value={counts.Service ?? 0} icon={<SettingsEthernetRoundedIcon />} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 3 }}>
          <StatCard title="Health score" value={clusterHealth?.healthScore ?? "-"} icon={<ShowChartRoundedIcon />} accent="#1677ff" />
        </Grid2>
      </Grid2>

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, lg: 7 }}>
          <Card sx={{ height: "100%" }}>
            <CardContent sx={{ p: 2.25 }}>
              <Typography variant="h6" sx={{ mb: 1.5 }}>
                Runtime profile
              </Typography>
              <Grid2 container spacing={1.5}>
                <Grid2 size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">
                    Environment
                  </Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {info?.environment ?? "-"}
                  </Typography>
                </Grid2>
                <Grid2 size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">
                    Cluster type
                  </Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {info?.cluster_type ?? "-"}
                  </Typography>
                </Grid2>
                <Grid2 size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">
                    Auth mode
                  </Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {info?.auth_mode ?? "-"}
                  </Typography>
                </Grid2>
                <Grid2 size={{ xs: 12, sm: 6 }}>
                  <Typography variant="body2" color="text.secondary">
                    Logged-in user
                  </Typography>
                  <Stack direction="row" spacing={1} alignItems="center" sx={{ mt: 0.5 }}>
                    <Typography variant="body1" fontWeight={700}>
                      {me?.username ?? "-"}
                    </Typography>
                    {me?.roles?.[0] && <StatusChip status={me.roles[0]} />}
                  </Stack>
                </Grid2>
              </Grid2>
            </CardContent>
          </Card>
        </Grid2>
        <Grid2 size={{ xs: 12, lg: 5 }}>
          {status === null ? (
            <EmptyState
              title="Controller status unavailable"
              message="Orbit API is reachable, but controller status has not been exposed yet."
            />
          ) : (
            <Card sx={{ height: "100%" }}>
              <CardContent sx={{ p: 2.25 }}>
                <Typography variant="h6" sx={{ mb: 1 }}>
                  Controller summary
                </Typography>
                <Typography color="text.secondary" sx={{ mb: 1.75 }}>
                  Live counts are sourced from the in-cluster controller and refreshed from real APIs.
                </Typography>
                <Stack spacing={1.1}>
                  {[
                    ["Namespaces", counts.Namespace ?? 0],
                    ["ReplicaSets", counts.ReplicaSet ?? 0],
                    ["ConfigMaps", counts.ConfigMap ?? 0],
                    ["Findings", status?.open_finding_count ?? 0],
                  ].map(([label, value]) => (
                    <Stack key={label} direction="row" justifyContent="space-between" alignItems="center">
                      <Typography color="text.secondary">{label}</Typography>
                      <Typography fontWeight={800}>{value}</Typography>
                    </Stack>
                  ))}
                </Stack>
              </CardContent>
            </Card>
          )}
        </Grid2>
      </Grid2>

      {clusterHealth && (
        <Card>
          <CardContent sx={{ p: 2.25 }}>
            <Stack spacing={1.75}>
              <Stack
                direction={{ xs: "column", md: "row" }}
                justifyContent="space-between"
                alignItems={{ xs: "flex-start", md: "center" }}
                spacing={1}
              >
                <Typography variant="h6">Cluster health</Typography>
                <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                  <HealthStatusChip status={clusterHealth.healthStatus} />
                  <StatusChip status={clusterHealth.metricsAvailable ? "active" : "warning"} />
                </Stack>
              </Stack>

              {!clusterHealth.metricsAvailable && (
                <Alert severity="warning">
                  Metrics API is unavailable. Orbit is using node conditions, pod phases, and events only.
                </Alert>
              )}

              <Grid2 container spacing={1.5}>
                <Grid2 size={{ xs: 6, md: 3 }}>
                  <Typography variant="body2" color="text.secondary">Nodes ready</Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {clusterHealth.readyNodeCount} / {clusterHealth.nodeCount}
                  </Typography>
                </Grid2>
                <Grid2 size={{ xs: 6, md: 3 }}>
                  <Typography variant="body2" color="text.secondary">Running pods</Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {clusterHealth.runningPodCount}
                  </Typography>
                </Grid2>
                <Grid2 size={{ xs: 6, md: 3 }}>
                  <Typography variant="body2" color="text.secondary">Pending pods</Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {clusterHealth.pendingPodCount}
                  </Typography>
                </Grid2>
                <Grid2 size={{ xs: 6, md: 3 }}>
                  <Typography variant="body2" color="text.secondary">Failed pods</Typography>
                  <Typography variant="body1" fontWeight={700}>
                    {clusterHealth.failedPodCount}
                  </Typography>
                </Grid2>
              </Grid2>

              <Grid2 container spacing={1.5}>
                <Grid2 size={{ xs: 12, md: 6 }}>
                  <MetricBar label="CPU usage" value={clusterHealth.cpuUsagePercent} />
                </Grid2>
                <Grid2 size={{ xs: 12, md: 6 }}>
                  <MetricBar label="Memory usage" value={clusterHealth.memoryUsagePercent} />
                </Grid2>
              </Grid2>

              <Typography variant="body2" color="text.secondary">
                Warning events: {clusterHealth.warningEventCount} · Observed: {clusterHealth.observedAt ? new Date(clusterHealth.observedAt).toLocaleString() : "-"}
              </Typography>
            </Stack>
          </CardContent>
        </Card>
      )}
    </Stack>
  );
}
