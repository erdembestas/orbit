import { Box, Stack, Typography, type ReactNode } from "@mui/material";

type PageHeaderProps = {
  title: string;
  subtitle: string;
  actions?: ReactNode;
};

export default function PageHeader({ title, subtitle, actions }: PageHeaderProps) {
  return (
    <Stack
      direction={{ xs: "column", md: "row" }}
      justifyContent="space-between"
      alignItems={{ xs: "flex-start", md: "center" }}
      spacing={2}
    >
      <Box sx={{ minWidth: 0 }}>
        <Typography variant="h4" sx={{ mb: 0.5 }}>
          {title}
        </Typography>
        <Typography variant="subtitle1">{subtitle}</Typography>
      </Box>
      {actions && (
        <Stack direction="row" spacing={1.5} flexWrap="wrap" useFlexGap>
          {actions}
        </Stack>
      )}
    </Stack>
  );
}
