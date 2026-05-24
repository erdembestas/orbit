import PreviewRoundedIcon from "@mui/icons-material/PreviewRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import {
  Button,
  Card,
  CardContent,
  Dialog,
  DialogContent,
  DialogTitle,
  Stack,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import { formatDateTime, type ActionPlan, type ApiError, type OrbitApiClient } from "../api/client";
import DataPanel from "../components/DataPanel";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import JsonPreview from "../components/JsonPreview";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import ResponsiveDataView, { type ResponsiveColumn } from "../components/ResponsiveDataView";
import StatusChip from "../components/StatusChip";

type ActionPlansPageProps = {
  client: OrbitApiClient;
};

export default function ActionPlansPage({ client }: ActionPlansPageProps) {
  const [plans, setPlans] = useState<ActionPlan[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedPlan, setSelectedPlan] = useState<ActionPlan | null>(null);

  async function load() {
    setLoading(true);
    setError("");
    try {
      const response = await client.listActionPlans();
      setPlans(response ?? null);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load action plans");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const columns: ResponsiveColumn<ActionPlan>[] = [
    {
      key: "title",
      label: "Title",
      render: (plan) => (
        <Stack spacing={0.5}>
          <Typography fontWeight={700}>{plan.title}</Typography>
          <Typography variant="body2" color="text.secondary">
            {plan.summary}
          </Typography>
        </Stack>
      ),
    },
    { key: "risk", label: "Risk", render: (plan) => <StatusChip status={plan.riskLevel} /> },
    { key: "status", label: "Status", render: (plan) => <StatusChip status={plan.status} /> },
    { key: "created", label: "Created At", render: (plan) => formatDateTime(plan.createdAt), mobilePriority: false },
    {
      key: "actions",
      label: "Actions",
      align: "right",
      render: (plan) => (
        <Button size="small" variant="outlined" startIcon={<PreviewRoundedIcon />} onClick={() => setSelectedPlan(plan)}>
          View Details
        </Button>
      ),
    },
  ];

  if (loading) {
    return <LoadingState label="Loading action plans" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (plans === null) {
    return <EmptyState title="Action plans not implemented" message="The backend does not currently expose action plan APIs." />;
  }

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Action Plans"
        subtitle="Draft recommendations generated from evidence packs."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      {plans.length === 0 ? (
        <EmptyState title="No action plans yet" message="Reason over a finding or evidence pack to create a draft plan." />
      ) : (
        <Card>
          <CardContent sx={{ p: 0 }}>
            <ResponsiveDataView
              rows={plans}
              columns={columns}
              getRowId={(plan) => plan.id}
              renderMobileTitle={(plan) => plan.title}
              renderMobileSubtitle={(plan) => `${plan.riskLevel} risk • ${plan.status}`}
            />
          </CardContent>
        </Card>
      )}

      <Dialog open={Boolean(selectedPlan)} onClose={() => setSelectedPlan(null)} maxWidth="lg" fullWidth>
        <DialogTitle>{selectedPlan?.title}</DialogTitle>
        <DialogContent>
          {selectedPlan && (
            <Stack spacing={2}>
              <DataPanel>
                    <Stack spacing={1.25}>
                    <Typography color="text.secondary">{selectedPlan.summary}</Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                      <StatusChip status={selectedPlan.riskLevel} />
                      <StatusChip status={selectedPlan.status} />
                      <StatusChip status={formatDateTime(selectedPlan.createdAt)} />
                    </Stack>
                  </Stack>
              </DataPanel>
              <JsonPreview value={selectedPlan.planJson} title="Action plan JSON" />
            </Stack>
          )}
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
