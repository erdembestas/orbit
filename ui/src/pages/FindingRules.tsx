import PrecisionManufacturingRoundedIcon from "@mui/icons-material/PrecisionManufacturingRounded";
import { Alert, Card, CardContent, Stack, Typography } from "@mui/material";
import { useEffect, useState } from "react";
import { type ApiError, type FindingRule, type OrbitApiClient } from "../api/client";
import EmptyState from "../components/EmptyState";
import ErrorState from "../components/ErrorState";
import LoadingState from "../components/LoadingState";
import PageHeader from "../components/PageHeader";
import SeverityChip from "../components/SeverityChip";
import StatusChip from "../components/StatusChip";

type FindingRulesPageProps = {
  client: OrbitApiClient;
};

export default function FindingRulesPage({ client }: FindingRulesPageProps) {
  const [rules, setRules] = useState<FindingRule[] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    async function load() {
      setLoading(true);
      setError("");
      try {
        const response = await client.listFindingRules();
        if (!cancelled) {
          setRules(response ?? null);
        }
      } catch (err) {
        if (!cancelled) {
          setError((err as ApiError).message ?? "Failed to load finding rules");
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }
    void load();
    return () => {
      cancelled = true;
    };
  }, [client]);

  if (loading) {
    return <LoadingState label="Loading finding rules" />;
  }

  if (error) {
    return <ErrorState message={error} />;
  }

  if (rules === null) {
    return <EmptyState title="Finding rules not implemented" message="The backend does not currently expose deterministic rule metadata." />;
  }

  return (
    <Stack spacing={2.5}>
      <PageHeader
        title="Finding Rules"
        subtitle="Deterministic classification rules currently used by Orbit controller."
      />

      <Alert severity="info" icon={<PrecisionManufacturingRoundedIcon fontSize="inherit" />}>
        These rules are deterministic and not LLM-based. They create findings and evidence only. No apply, remediation, approval, or execution happens here.
      </Alert>

      {rules.length === 0 ? (
        <EmptyState title="No rules defined" message="Orbit controller did not return any rule definitions." />
      ) : (
        <Stack spacing={1.5}>
          {rules.map((rule) => (
            <Card key={rule.name}>
              <CardContent sx={{ p: 2 }}>
                <Stack spacing={1.25}>
                  <Stack direction={{ xs: "column", md: "row" }} justifyContent="space-between" spacing={1.5}>
                    <Stack spacing={0.5}>
                      <Typography variant="h6">{rule.title}</Typography>
                      <Typography color="text.secondary">{rule.description}</Typography>
                    </Stack>
                    <Stack direction="row" spacing={1} flexWrap="wrap" useFlexGap>
                      <StatusChip status={rule.resourceKind} />
                      <StatusChip status={rule.category} />
                      <SeverityChip severity={rule.severity.split("|")[0]} />
                    </Stack>
                  </Stack>
                  <Typography variant="body2">
                    <strong>Condition:</strong> {rule.condition}
                  </Typography>
                  <Typography variant="body2">
                    <strong>Evidence fields:</strong> {rule.evidenceFields.join(", ")}
                  </Typography>
                  <Stack spacing={0.5}>
                    <Typography variant="body2" fontWeight={700}>
                      Known limitations
                    </Typography>
                    {rule.limitations.map((limitation) => (
                      <Typography key={limitation} variant="body2" color="text.secondary">
                        • {limitation}
                      </Typography>
                    ))}
                  </Stack>
                </Stack>
              </CardContent>
            </Card>
          ))}
        </Stack>
      )}
    </Stack>
  );
}
