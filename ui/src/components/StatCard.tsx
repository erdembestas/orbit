import { alpha } from "@mui/material/styles";
import { Box, Card, CardContent, Typography, type SxProps, type Theme } from "@mui/material";
import type { ReactNode } from "react";

type StatCardProps = {
  title: string;
  value: string | number;
  subtitle?: string;
  icon?: ReactNode;
  accent?: string;
  sx?: SxProps<Theme>;
};

export default function StatCard({ title, value, subtitle, icon, accent = "#3157FF", sx }: StatCardProps) {
  return (
    <Card
      sx={{
        height: "100%",
        position: "relative",
        overflow: "hidden",
        ...sx,
      }}
    >
      <CardContent sx={{ p: 2 }}>
        <Box sx={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: 2 }}>
          <Box sx={{ minWidth: 0 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 0.75, fontSize: 12 }}>
              {title}
            </Typography>
            <Typography sx={{ lineHeight: 1.1, fontSize: { xs: 22, md: 24 }, fontWeight: 700 }}>
              {value}
            </Typography>
            {subtitle && (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 0.75 }}>
                {subtitle}
              </Typography>
            )}
          </Box>
          {icon && (
            <Box
              sx={{
                width: 28,
                height: 28,
                borderRadius: 1,
                display: "grid",
                placeItems: "center",
                color: accent,
                bgcolor: alpha(accent, 0.08),
                flexShrink: 0,
                "& svg": {
                  fontSize: 16,
                },
              }}
            >
              {icon}
            </Box>
          )}
        </Box>
      </CardContent>
    </Card>
  );
}
