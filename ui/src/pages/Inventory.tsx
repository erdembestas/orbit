import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import {
  Button,
  Card,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import { formatDateTime, type ApiError, type OrbitApiClient, type Resource } from "../api/client";
import DataPanel from "../components/DataPanel";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import ResponsiveDataView, { type ResponsiveColumn } from "../components/ResponsiveDataView";
import StatusChip from "../components/StatusChip";

type InventoryPageProps = {
  client: OrbitApiClient;
};

export default function InventoryPage({ client }: InventoryPageProps) {
  const [resources, setResources] = useState<Resource[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [filters, setFilters] = useState({ kind: "", namespace: "", name: "" });

  async function load() {
    setLoading(true);
    setError("");
    try {
      const response = await client.listResources();
      setResources(response ?? null);
    } catch (err) {
      setError((err as ApiError).message ?? "Failed to load resources");
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  const filteredResources = useMemo(() => {
    const source = resources ?? [];
    return source.filter((resource) => {
      return (
        (!filters.kind || resource.kind.toLowerCase().includes(filters.kind.toLowerCase())) &&
        (!filters.namespace || (resource.namespace ?? "").toLowerCase().includes(filters.namespace.toLowerCase())) &&
        (!filters.name || resource.name.toLowerCase().includes(filters.name.toLowerCase()))
      );
    });
  }, [filters, resources]);

  const columns: ResponsiveColumn<Resource>[] = [
    { key: "kind", label: "Kind", render: (resource) => <StatusChip status={resource.kind} /> },
    { key: "namespace", label: "Namespace", render: (resource) => resource.namespace ?? "-" },
    {
      key: "name",
      label: "Name",
      render: (resource) => (
        <Typography fontWeight={700} sx={{ wordBreak: "break-word" }}>
          {resource.name}
        </Typography>
      ),
    },
    { key: "status", label: "Status", render: (resource) => <StatusChip status={resource.status ?? "unknown"} /> },
    { key: "observed", label: "Observed At", render: (resource) => formatDateTime(resource.observed_at), mobilePriority: false },
  ];

  if (loading) {
    return <LoadingState label="Loading inventory" />;
  }

  if (error) {
    return <ErrorState message={error} onRetry={() => void load()} />;
  }

  if (resources === null) {
    return <EmptyState title="Inventory not implemented" message="This backend does not currently expose the inventory endpoint." />;
  }

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Inventory"
        subtitle="Kubernetes resources observed by Orbit controller."
        actions={
          <Button variant="outlined" startIcon={<RefreshRoundedIcon />} onClick={() => void load()}>
            Refresh
          </Button>
        }
      />

      <DataPanel>
          <Stack direction={{ xs: "column", md: "row" }} spacing={1.5}>
            <TextField
              label="Kind"
              value={filters.kind}
              onChange={(event) => setFilters({ ...filters, kind: event.target.value })}
              fullWidth
              InputProps={{ startAdornment: <SearchRoundedIcon fontSize="small" sx={{ mr: 1, color: "text.secondary" }} /> }}
            />
            <TextField
              label="Namespace"
              value={filters.namespace}
              onChange={(event) => setFilters({ ...filters, namespace: event.target.value })}
              fullWidth
            />
            <TextField
              label="Name"
              value={filters.name}
              onChange={(event) => setFilters({ ...filters, name: event.target.value })}
              fullWidth
            />
          </Stack>
      </DataPanel>

      {filteredResources.length === 0 ? (
        <EmptyState title="No resources match the current filters" message="Try clearing one or more filters." />
      ) : (
        <Card>
          <CardContent sx={{ p: 0 }}>
            <ResponsiveDataView
              rows={filteredResources}
              columns={columns}
              getRowId={(resource) => resource.id}
              renderMobileTitle={(resource) => resource.name}
              renderMobileSubtitle={(resource) => `${resource.kind} • ${resource.namespace ?? "cluster-scope"}`}
            />
          </CardContent>
        </Card>
      )}
    </Stack>
  );
}
