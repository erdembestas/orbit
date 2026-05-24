import { alpha } from "@mui/material/styles";
import { Box, Card, CardContent, Typography, type SxProps, type Theme } from "@mui/material";
import type { ReactNode } from "react";

type StatCardProps = {
  title: string;
  value: string | number;
  subtitle?: string;
  helper?: string;
  icon?: ReactNode;
  accent?: string;
  visual?: ReactNode;
  sx?: SxProps<Theme>;
};

export default function StatCard({
  title,
  value,
  subtitle,
  helper,
  icon,
  accent = "#14B8A6",
  visual,
  sx,
}: StatCardProps) {
  return (
    <Card
      sx={{
        height: "100%",
        position: "relative",
        overflow: "hidden",
        bgcolor: "background.paper",
        ...sx,
      }}
    >
      <CardContent sx={{ p: 1.75 }}>
        <Box sx={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: 2 }}>
          <Box sx={{ minWidth: 0, flex: 1 }}>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 0.7, fontSize: 12 }}>
              {title}
            </Typography>
            <Typography sx={{ lineHeight: 1.05, fontSize: { xs: 24, md: 28 }, fontWeight: 700 }}>
              {value}
            </Typography>
            {subtitle && (
              <Typography variant="body2" color="text.secondary" sx={{ mt: 0.65 }}>
                {subtitle}
              </Typography>
            )}
            {helper && (
              <Typography variant="caption" color="text.secondary" sx={{ mt: 0.65, display: "block" }}>
                {helper}
              </Typography>
            )}
          </Box>
          {visual ?? (icon && (
            <Box
              sx={{
                width: 30,
                height: 30,
                borderRadius: 1.25,
                display: "grid",
                placeItems: "center",
                color: accent,
                bgcolor: alpha(accent, 0.12),
                flexShrink: 0,
                "& svg": {
                  fontSize: 17,
                },
              }}
            >
              {icon}
            </Box>
          ))}
        </Box>
        <Box
          sx={{
            mt: 1.5,
            height: 2,
            borderRadius: 999,
            bgcolor: alpha("#FFFFFF", 0.05),
            overflow: "hidden",
          }}
        >
          <Box sx={{ width: "100%", height: "100%", bgcolor: alpha(accent, 0.85) }} />
        </Box>
      </CardContent>
    </Card>
  );
}
