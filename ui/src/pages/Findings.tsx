import AutoAwesomeRoundedIcon from "@mui/icons-material/AutoAwesomeRounded";
import PreviewRoundedIcon from "@mui/icons-material/PreviewRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import {
  Alert,
  Box,
  Button,
  Card,
  Dialog,
  DialogContent,
  DialogTitle,
  MenuItem,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import {
  formatDateTime,
  type ApiError,
  type EvidencePack,
  type Finding,
  type OrbitApiClient,
  type ReasoningResponse,
} from "../api/client";
import DataPanel from "../components/DataPanel";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import JsonPreview from "../components/JsonPreview";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import ResponsiveDataView, { type ResponsiveColumn } from "../components/ResponsiveDataView";
import SeverityChip from "../components/SeverityChip";
import StatusChip from "../components/StatusChip";

type FindingsPageProps = {
  client: OrbitApiClient;
};

const severityOptions = ["", "critical", "high", "medium", "low", "info"];
const statusOptions = ["", "open", "resolved"];

export default function FindingsPage({ client }: FindingsPageProps) {
  const [findings, setFindings] = useState<Finding[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ severity: "", status: "", namespace: "", kind: "" });
  const [selectedPack, setSelectedPack] = useState<EvidencePack | null>(null);
  const [reasoning, setReasoning] = useState<ReasoningResponse | null>(null);
  const [dialogTitle, setDialogTitle] = useState("");

  async function load() {
    setLoading(true);
    setError("");
    try {
      const response = await client.listFindings();
      setFindings(response ?? null);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load findings");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const filtered = useMemo(() => {
    const source = findings ?? [];
    return source.filter((finding) => {
      return (
        (!filters.severity || finding.severity === filters.severity) &&
        (!filters.status || finding.status === filters.status) &&
        (!filters.namespace || (finding.resource_namespace ?? "").toLowerCase().includes(filters.namespace.toLowerCase())) &&
        (!filters.kind || (finding.resource_kind ?? "").toLowerCase().includes(filters.kind.toLowerCase()))
      );
    });
  }, [filters, findings]);

  async function handleViewEvidence(findingId: string) {
    try {
      const response = await client.getFindingEvidencePack(findingId);
      setDialogTitle("Finding evidence pack");
      setSelectedPack(response);
      setReasoning(null);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load evidence pack");
    }
  }

  async function handleReason(findingId: string) {
    try {
      const response = await client.reasonFinding(findingId);
      if (!response) {
        return;
      }
      setDialogTitle("Reasoning draft");
      setReasoning(response);
      setSelectedPack(response.evidencePack);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to generate reasoning");
    }
  }

  const columns: ResponsiveColumn<Finding>[] = [
    { key: "severity", label: "Severity", render: (finding) => <SeverityChip severity={finding.severity} /> },
    { key: "category", label: "Category", render: (finding) => finding.category },
    {
      key: "title",
      label: "Title",
      render: (finding) => (
        <Stack spacing={0.5}>
          <Typography fontWeight={700}>{finding.title}</Typography>
          <Typography variant="body2" color="text.secondary">
            {finding.description}
          </Typography>
        </Stack>
      ),
    },
    {
      key: "resource",
      label: "Resource",
      render: (finding) =>
        [finding.resource_kind, finding.resource_namespace, finding.resource_name].filter(Boolean).join(" / ") || "-",
      mobilePriority: false,
    },
    { key: "status", label: "Status", render: (finding) => <StatusChip status={finding.status} /> },
    { key: "updated", label: "Updated", render: (finding) => formatDateTime(finding.updated_at), mobilePriority: false },
    {
      key: "actions",
      label: "Actions",
      align: "right",
      render: (finding) => (
        <Stack direction={{ xs: "column", lg: "row" }} spacing={1} justifyContent="flex-end">
          <Button size="small" variant="outlined" startIcon={<PreviewRoundedIcon />} onClick={() => void handleViewEvidence(finding.id)}>
            View Evidence
          </Button>
          <Button size="small" variant="contained" startIcon={<AutoAwesomeRoundedIcon />} onClick={() => void handleReason(finding.id)}>
            Reason
          </Button>
        </Stack>
      ),
    },
  ];

  if (loading) {
    return <LoadingState label="Loading findings" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (findings === null) {
    return <EmptyState title="Findings not implemented" message="This backend does not currently expose findings." />;
  }

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Findings"
        subtitle="Deterministic issues detected from cluster state."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      <DataPanel>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1.5}>
            <TextField select label="Severity" value={filters.severity} onChange={(event) => setFilters({ ...filters, severity: event.target.value })} fullWidth>
              {severityOptions.map((option) => (
                <MenuItem key={option || "all"} value={option}>
                  {option || "All severities"}
                </MenuItem>
              ))}
            </TextField>
            <TextField select label="Status" value={filters.status} onChange={(event) => setFilters({ ...filters, status: event.target.value })} fullWidth>
              {statusOptions.map((option) => (
                <MenuItem key={option || "all"} value={option}>
                  {option || "All statuses"}
                </MenuItem>
              ))}
            </TextField>
            <TextField label="Namespace" value={filters.namespace} onChange={(event) => setFilters({ ...filters, namespace: event.target.value })} fullWidth />
            <TextField label="Kind" value={filters.kind} onChange={(event) => setFilters({ ...filters, kind: event.target.value })} fullWidth />
          </Stack>
      </DataPanel>

      {filtered.length === 0 ? (
        <EmptyState title="No findings match the current filters" message="Try broadening the filters or refresh the page." />
      ) : (
        <Card>
          <CardContent sx={{ p: 0 }}>
            <ResponsiveDataView
              rows={filtered}
              columns={columns}
              getRowId={(finding) => finding.id}
              renderMobileTitle={(finding) => finding.title}
              renderMobileSubtitle={(finding) => `${finding.category} • ${finding.resource_name ?? "cluster-scope"}`}
            />
          </CardContent>
        </Card>
      )}

      {reasoning && (
        <Alert severity="info">
          Draft action plan created: <strong>{reasoning.actionPlan.title}</strong>
        </Alert>
      )}

      <Dialog open={Boolean(selectedPack || reasoning)} onClose={() => { setSelectedPack(null); setReasoning(null); }} maxWidth="lg" fullWidth>
        <DialogTitle>{dialogTitle}</DialogTitle>
        <DialogContent>
          <Stack spacing={2}>
            {reasoning && (
              <DataPanel title={reasoning.actionPlan.title}>
                  <Stack spacing={1.5}>
                    <Typography color="text.secondary">{reasoning.reasoning.rootCause}</Typography>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                      <StatusChip status={reasoning.reasoning.riskLevel} />
                      <StatusChip status={`confidence ${Math.round(reasoning.reasoning.confidence * 100)}%`} />
                      <StatusChip status={reasoning.actionPlan.status} />
                    </Stack>
                    <Box>
                      <Typography fontWeight={700} sx={{ mb: 0.5 }}>
                        Suggested action plan
                      </Typography>
                      {reasoning.reasoning.suggestedActionPlan.map((step) => (
                        <Typography key={step} variant="body2" color="text.secondary">
                          • {step}
                        </Typography>
                      ))}
                    </Box>
                    <Box>
                      <Typography fontWeight={700} sx={{ mb: 0.5 }}>
                        Validation steps
                      </Typography>
                      {reasoning.reasoning.validationSteps.map((step) => (
                        <Typography key={step} variant="body2" color="text.secondary">
                          • {step}
                        </Typography>
                      ))}
                    </Box>
                    <Box>
                      <Typography fontWeight={700} sx={{ mb: 0.5 }}>
                        Rollback steps
                      </Typography>
                      {reasoning.reasoning.rollbackSteps.map((step) => (
                        <Typography key={step} variant="body2" color="text.secondary">
                          • {step}
                        </Typography>
                      ))}
                    </Box>
                  </Stack>
              </DataPanel>
            )}
            {selectedPack && <JsonPreview value={selectedPack.packJson} title="Evidence pack JSON" />}
          </Stack>
        </DialogContent>
      </Dialog>
    </Stack>
  );
}
