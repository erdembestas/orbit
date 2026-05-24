import Chip from "@mui/material/Chip";

type HealthStatusChipProps = {
  status: string | null | undefined;
};

const palette: Record<string, { color: string; background: string; border: string }> = {
  healthy: { color: "#86EFAC", background: "rgba(34, 197, 94, 0.12)", border: "rgba(34, 197, 94, 0.22)" },
  warning: { color: "#FCD34D", background: "rgba(245, 158, 11, 0.12)", border: "rgba(245, 158, 11, 0.22)" },
  critical: { color: "#FCA5A5", background: "rgba(239, 68, 68, 0.12)", border: "rgba(239, 68, 68, 0.22)" },
  unknown: { color: "#CBD5E1", background: "rgba(148, 163, 184, 0.1)", border: "rgba(148, 163, 184, 0.2)" },
};

export default function HealthStatusChip({ status }: HealthStatusChipProps) {
  const normalized = String(status ?? "unknown").toLowerCase();
  const style = palette[normalized] ?? palette.unknown;

  return (
    <Chip
      size="small"
      label={normalized}
      variant="outlined"
      sx={{
        height: 22,
        borderRadius: 1,
        textTransform: "capitalize",
        fontWeight: 700,
        fontSize: 11,
        color: style.color,
        bgcolor: style.background,
        borderColor: style.border,
      }}
    />
  );
}
