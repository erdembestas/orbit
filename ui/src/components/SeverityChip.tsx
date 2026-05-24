import { alpha } from "@mui/material/styles";
import { Chip } from "@mui/material";

type SeverityChipProps = {
  severity: string;
};

const styles: Record<string, { label: string; sx: object }> = {
  critical: { label: "critical", sx: { bgcolor: alpha("#EF4444", 0.14), color: "#FCA5A5", borderColor: alpha("#EF4444", 0.22) } },
  high: { label: "high", sx: { bgcolor: alpha("#F97316", 0.14), color: "#FDBA74", borderColor: alpha("#F97316", 0.22) } },
  medium: { label: "medium", sx: { bgcolor: alpha("#F59E0B", 0.14), color: "#FCD34D", borderColor: alpha("#F59E0B", 0.22) } },
  low: { label: "low", sx: { bgcolor: alpha("#14B8A6", 0.14), color: "#5EEAD4", borderColor: alpha("#14B8A6", 0.22) } },
  info: { label: "info", sx: { bgcolor: alpha("#60A5FA", 0.12), color: "#BFDBFE", borderColor: alpha("#60A5FA", 0.22) } },
};

export default function SeverityChip({ severity }: SeverityChipProps) {
  const normalized = severity.toLowerCase();
  const config = styles[normalized] ?? styles.info;

  return (
    <Chip
      size="small"
      label={config.label}
      variant="outlined"
      sx={{ fontWeight: 700, textTransform: "capitalize", ...config.sx }}
    />
  );
}
