import { alpha } from "@mui/material/styles";
import { Chip } from "@mui/material";

type StatusChipProps = {
  status: string | number | boolean | null | undefined;
};

const styles: Record<string, { color: string; background: string; border: string }> = {
  draft: { color: "#93C5FD", background: alpha("#60A5FA", 0.14), border: alpha("#60A5FA", 0.22) },
  waiting_approval: { color: "#FCD34D", background: alpha("#F59E0B", 0.12), border: alpha("#F59E0B", 0.22) },
  approved: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  rejected: { color: "#FCA5A5", background: alpha("#EF4444", 0.12), border: alpha("#EF4444", 0.22) },
  executed: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  failed: { color: "#FCA5A5", background: alpha("#EF4444", 0.12), border: alpha("#EF4444", 0.22) },
  open: { color: "#FCD34D", background: alpha("#F59E0B", 0.12), border: alpha("#F59E0B", 0.22) },
  resolved: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  active: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  running: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  healthy: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  succeeded: { color: "#86EFAC", background: alpha("#22C55E", 0.12), border: alpha("#22C55E", 0.22) },
  warning: { color: "#FCD34D", background: alpha("#F59E0B", 0.12), border: alpha("#F59E0B", 0.22) },
  critical: { color: "#FCA5A5", background: alpha("#EF4444", 0.12), border: alpha("#EF4444", 0.22) },
  unknown: { color: "#CBD5E1", background: alpha("#94A3B8", 0.1), border: alpha("#94A3B8", 0.2) },
  unavailable: { color: "#FCA5A5", background: alpha("#EF4444", 0.12), border: alpha("#EF4444", 0.22) },
  pending: { color: "#FCD34D", background: alpha("#F59E0B", 0.12), border: alpha("#F59E0B", 0.22) },
};

export default function StatusChip({ status }: StatusChipProps) {
  const label = String(status ?? "-");
  const normalized = label.toLowerCase();
  const style = styles[normalized] ?? styles.unknown;

  return (
    <Chip
      size="small"
      label={label.replaceAll("_", " ")}
      variant="outlined"
      sx={{
        textTransform: "capitalize",
        fontWeight: 600,
        color: style.color,
        bgcolor: style.background,
        borderColor: style.border,
      }}
    />
  );
}
