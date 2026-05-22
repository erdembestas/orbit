import { Chip, type ChipProps } from "@mui/material";

type StatusChipProps = {
  status: string | number | boolean | null | undefined;
};

const styles: Record<string, ChipProps["color"]> = {
  draft: "info",
  waiting_approval: "warning",
  approved: "success",
  rejected: "error",
  executed: "success",
  failed: "error",
  open: "warning",
  resolved: "success",
  active: "success",
  running: "success",
  healthy: "success",
  succeeded: "success",
  warning: "warning",
  critical: "error",
  unknown: "default",
  unavailable: "error",
  pending: "warning",
};

export default function StatusChip({ status }: StatusChipProps) {
  const label = String(status ?? "-");
  const normalized = label.toLowerCase();
  const color = styles[normalized] ?? "default";

  return (
    <Chip
      size="small"
      label={label.replaceAll("_", " ")}
      color={color}
      variant={color === "default" ? "outlined" : "filled"}
      sx={{ textTransform: "capitalize", fontWeight: 600 }}
    />
  );
}
