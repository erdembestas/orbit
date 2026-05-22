import { Card, CardContent, Stack, Typography } from "@mui/material";
import type { ReactNode } from "react";

type CompactStatProps = {
  label: string;
  value: ReactNode;
  hint?: ReactNode;
};

export default function CompactStat({ label, value, hint }: CompactStatProps) {
  return (
    <Card variant="outlined">
      <CardContent sx={{ p: 1.5, "&:last-child": { pb: 1.5 } }}>
        <Stack spacing={0.5}>
          <Typography variant="caption" color="text.secondary" sx={{ letterSpacing: "0.04em", textTransform: "uppercase" }}>
            {label}
          </Typography>
          <Typography sx={{ fontSize: 22, fontWeight: 700, lineHeight: 1.15 }}>
            {value}
          </Typography>
          {hint ? (
            <Typography variant="caption" color="text.secondary">
              {hint}
            </Typography>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  );
}
