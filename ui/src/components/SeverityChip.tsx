import { alpha } from "@mui/material/styles";
import { Chip } from "@mui/material";

type SeverityChipProps = {
  severity: string;
};

const styles: Record<string, { label: string; sx: object }> = {
  critical: { label: "critical", sx: { bgcolor: alpha("#D32F2F", 0.12), color: "#B3261E" } },
  high: { label: "high", sx: { bgcolor: alpha("#E65100", 0.12), color: "#C2410C" } },
  medium: { label: "medium", sx: { bgcolor: alpha("#F59E0B", 0.16), color: "#A16207" } },
  low: { label: "low", sx: { bgcolor: alpha("#1677FF", 0.1), color: "#0B74DE" } },
  info: { label: "info", sx: { bgcolor: alpha("#98A2B3", 0.18), color: "#475467" } },
};

export default function SeverityChip({ severity }: SeverityChipProps) {
  const normalized = severity.toLowerCase();
  const config = styles[normalized] ?? styles.info;

  return (
    <Chip
      size="small"
      label={config.label}
      sx={{ fontWeight: 700, textTransform: "capitalize", ...config.sx }}
    />
  );
}
