import Chip from "@mui/material/Chip";

type HealthStatusChipProps = {
  status: string | null | undefined;
};

const palette: Record<string, { color: string; background: string; border: string }> = {
  healthy: { color: "#166534", background: "#DCFCE7", border: "#BBF7D0" },
  warning: { color: "#B45309", background: "#FEF3C7", border: "#FDE68A" },
  critical: { color: "#B42318", background: "#FEE4E2", border: "#FECDCA" },
  unknown: { color: "#475467", background: "#F2F4F7", border: "#E4E7EC" },
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
