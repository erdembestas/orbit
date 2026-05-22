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
      ? "#D0D5DD"
      : normalized >= critical
        ? "#D92D20"
        : normalized >= warn
          ? "#F79009"
          : "#1677FF";

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
            height: 6,
            borderRadius: 999,
            bgcolor: "#EEF2F6",
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
