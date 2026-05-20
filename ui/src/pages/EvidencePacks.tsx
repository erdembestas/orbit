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
  Stack,
  Typography,
} from "@mui/material";
import { useEffect, useState } from "react";
import {
  formatDateTime,
  type ApiError,
  type EvidencePack,
  type OrbitApiClient,
  type ReasoningResponse,
} from "../api/client";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import JsonPreview from "../components/JsonPreview";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import ResponsiveDataView, { type ResponsiveColumn } from "../components/ResponsiveDataView";
import StatusChip from "../components/StatusChip";

type EvidencePacksPageProps = {
  client: OrbitApiClient;
};

export default function EvidencePacksPage({ client }: EvidencePacksPageProps) {
  const [packs, setPacks] = useState<EvidencePack[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedPack, setSelectedPack] = useState<EvidencePack | null>(null);
  const [reasoning, setReasoning] = useState<ReasoningResponse | null>(null);
  const [banner, setBanner] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const response = await client.listEvidencePacks();
      setPacks(response ?? null);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load evidence packs");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleView(packId: string) {
    try {
      const response = await client.getEvidencePack(packId);
      if (!response) {
        setBanner("Evidence pack detail is not implemented yet.");
        return;
      }
      setSelectedPack(response);
      setReasoning(null);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load evidence pack");
    }
  }

  async function handleReason(packId: string) {
    try {
      const response = await client.reasonEvidencePack(packId);
      if (!response) {
        setBanner("Mock reasoning is not implemented for evidence packs.");
        return;
      }
      setReasoning(response);
      setSelectedPack(response.evidencePack);
      await load();
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to reason over evidence pack");
    }
  }

  const columns: ResponsiveColumn<EvidencePack>[] = [
    { key: "scope", label: "Scope", render: (pack) => <StatusChip status={pack.scopeType} /> },
    { key: "namespace", label: "Namespace", render: (pack) => pack.scopeNamespace ?? "-" },
    {
      key: "name",
      label: "Name",
      render: (pack) => (
        <Typography fontWeight={700}>{pack.scopeName ?? pack.findingId ?? "finding scope"}</Typography>
      ),
    },
    { key: "tokens", label: "Token Estimate", render: (pack) => pack.tokenEstimate },
    { key: "created", label: "Created At", render: (pack) => formatDateTime(pack.createdAt), mobilePriority: false },
    {
      key: "actions",
      label: "Actions",
      align: "right",
      render: (pack) => (
        <Stack direction={{ xs: "column", lg: "row" }} spacing={1} justifyContent="flex-end">
          <Button size="small" variant="outlined" startIcon={<PreviewRoundedIcon />} onClick={() => void handleView(pack.id)}>
            View JSON
          </Button>
          <Button size="small" variant="contained" startIcon={<AutoAwesomeRoundedIcon />} onClick={() => void handleReason(pack.id)}>
            Reason
          </Button>
        </Stack>
      ),
    },
  ];

  if (loading) {
    return <LoadingState label="Loading evidence packs" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (packs === null) {
    return <EmptyState title="Evidence packs not implemented" message="The backend does not currently expose evidence pack APIs." />;
  }

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Evidence Packs"
        subtitle="Stored compact evidence contexts generated from findings, namespaces, and pods."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      {banner && <Alert severity="info">{banner}</Alert>}
      {reasoning && <Alert severity="success">Draft action plan created: {reasoning.actionPlan.title}</Alert>}

      {packs.length === 0 ? (
        <EmptyState title="No evidence packs yet" message="Use the Analyze page or finding actions to generate compact evidence." />
      ) : (
        <Card>
          <CardContent sx={{ p: 0 }}>
            <ResponsiveDataView
              rows={packs}
              columns={columns}
              getRowId={(pack) => pack.id}
              renderMobileTitle={(pack) => pack.scopeName ?? pack.findingId ?? pack.scopeType}
              renderMobileSubtitle={(pack) => `${pack.scopeType} • ${pack.scopeNamespace ?? "cluster-scope"}`}
            />
          </CardContent>
        </Card>
      )}

      <Dialog open={Boolean(selectedPack || reasoning)} onClose={() => { setSelectedPack(null); setReasoning(null); }} maxWidth="lg" fullWidth>
        <DialogTitle>Evidence pack details</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            {selectedPack && (
              <Card>
                <CardContent sx={{ p: 2 }}>
                  <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" spacing={2}>
                    <Stack spacing={1}>
                      <Typography variant="h6">
                        {selectedPack.scopeName ?? selectedPack.scopeNamespace ?? "Evidence pack"}
                      </Typography>
                      <Typography color="text.secondary">
                        Compact reasoning context for {selectedPack.scopeType} scope.
                      </Typography>
                    </Stack>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                      <StatusChip status={selectedPack.scopeType} />
                      <StatusChip status={`${selectedPack.tokenEstimate} tokens`} />
                    </Stack>
                  </Stack>
                </CardContent>
              </Card>
            )}
            {reasoning && (
              <Card>
                <CardContent sx={{ p: 2 }}>
                  <Stack spacing={1.5}>
                    <Stack direction="row" justifyContent="space-between" alignItems="center" spacing={2}>
                      <Typography variant="h6">Mock reasoning result</Typography>
                      <StatusChip status={reasoning.actionPlan.status} />
                    </Stack>
                    <Typography color="text.secondary">{reasoning.reasoning.rootCause}</Typography>
                    <Typography fontWeight={700}>Suggested action plan</Typography>
                    {reasoning.reasoning.suggestedActionPlan.map((step) => (
                      <Typography key={step} variant="body2" color="text.secondary">
                        • {step}
                      </Typography>
                    ))}
                  </Stack>
                </CardContent>
              </Card>
            )}
            {selectedPack && <JsonPreview value={selectedPack.packJson} title="Evidence pack JSON" />}
            {reasoning && <JsonPreview value={reasoning} title="Reasoning payload" />}
          </Stack>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
