import { Box, LinearProgress, Stack, Typography } from "@mui/material";

type MetricBarProps = {
  label: string;
  value: number | null | undefined;
  warn?: number;
  critical?: number;
};

export default function MetricBar({
  label,
  value,
  warn = 75,
  critical = 90,
}: MetricBarProps) {
  const normalized = typeof value === "number" ? Math.max(0, Math.min(100, value)) : null;
  const barColor =
    normalized == null
      ? "#475467"
      : normalized >= critical
        ? "#EF4444"
        : normalized >= warn
          ? "#F59E0B"
          : "#14B8A6";

  return (
    <Stack spacing={0.75}>
      <Stack direction="row" justifyContent="space-between" spacing={1}>
        <Typography variant="body2" color="text.secondary">
          {label}
        </Typography>
        <Typography variant="body2" fontWeight={700}>
          {normalized == null ? "Unavailable" : `${normalized.toFixed(1)}%`}
        </Typography>
      </Stack>
      <Box>
        <LinearProgress
          variant="determinate"
          value={normalized ?? 0}
          sx={{
            height: 7,
            borderRadius: 999,
            bgcolor: "#1B2330",
            "& .MuiLinearProgress-bar": {
              borderRadius: 999,
              backgroundColor: barColor,
            },
          }}
        />
      </Box>
    </Stack>
  );
}
