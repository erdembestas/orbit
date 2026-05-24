import { Box, Card, CardContent, CircularProgress, Stack, Typography } from "@mui/material";

type LoadingStateProps = {
  label?: string;
};

export default function LoadingState({ label = "Loading…" }: LoadingStateProps) {
  return (
    <Card>
      <CardContent sx={{ minHeight: 180, display: "grid", placeItems: "center" }}>
        <Stack spacing={2} alignItems="center" textAlign="center">
          <CircularProgress size={24} thickness={5} />
          <Typography color="text.secondary" variant="body2">
            {label}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
}
