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
import { formatNullableTime, type ApiError, type ControllerStatus, type OrbitApiClient } from "../api/client";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import StatCard from "../components/StatCard";

type ControllerStatusPageProps = {
  client: OrbitApiClient;
};

export default function ControllerStatusPage({ client }: ControllerStatusPageProps) {
  const [status, setStatus] = useState<ControllerStatus | null | undefined>(undefined);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const response = await client.getControllerStatus();
      setStatus(response);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load controller status");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  if (loading) {
    return <LoadingState label="Loading controller status" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (status === null) {
    return <EmptyState title="Controller status not implemented" message="The backend did not expose controller status." />;
  }

  const counts = status?.resource_counts_by_kind ?? {};
  const cards = [
    ["Pods", counts.Pod ?? 0],
    ["Deployments", counts.Deployment ?? 0],
    ["ReplicaSets", counts.ReplicaSet ?? 0],
    ["Services", counts.Service ?? 0],
    ["ConfigMaps", counts.ConfigMap ?? 0],
    ["Namespaces", counts.Namespace ?? 0],
  ];

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Controller Status"
        subtitle="Single-cluster inventory collector health."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      {status?.metrics_available === false && (
        <Alert icon={<WarningAmberRoundedIcon />} severity="warning">
          Metrics API is unavailable. Orbit continues without live pod metrics.
        </Alert>
      )}

      <Grid2 container spacing={1.5}>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <StatCard title="Cluster" value={status?.cluster_name ?? "-"} />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <StatCard title="Mode" value={status?.mode ?? "-"} accent="#6C5CFF" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <StatCard title="Last Seen" value={formatNullableTime(status?.last_seen_at)} accent="#6172F3" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <StatCard title="Open Findings" value={status?.open_finding_count ?? 0} accent="#F59E0B" />
        </Grid2>
        <Grid2 size={{ xs: 12, sm: 6, xl: 2.4 }}>
          <StatCard title="Evidence Packs" value={status?.latest_evidence_pack_count ?? 0} accent="#12B76A" />
        </Grid2>
      </Grid2>

      <Card>
        <CardContent sx={{ p: 2.25 }}>
          <Typography variant="h6" sx={{ mb: 1.75 }}>
            Resource counts
          </Typography>
          <Grid2 container spacing={1.5}>
            {cards.map(([label, value]) => (
              <Grid2 key={label} size={{ xs: 12, sm: 6, lg: 4 }}>
                <Card variant="outlined" sx={{ bgcolor: "#FBFCFF", boxShadow: "none" }}>
                  <CardContent sx={{ p: 1.75 }}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center">
                      <Typography color="text.secondary">{label}</Typography>
                      <Typography variant="h5">{value}</Typography>
                    </Stack>
                  </CardContent>
                </Card>
              </Grid2>
            ))}
          </Grid2>
        </CardContent>
      </Card>
    </Stack>
  );
}
