import { Card, CardContent, Stack, Typography, type ReactNode } from "@mui/material";

type DataPanelProps = {
  title?: string;
  subtitle?: string;
  actions?: ReactNode;
  children: ReactNode;
  contentSx?: object;
};

export default function DataPanel({ title, subtitle, actions, children, contentSx }: DataPanelProps) {
  return (
    <Card>
      <CardContent sx={{ p: 2, "&:last-child": { pb: 2 }, ...contentSx }}>
        {(title || actions) && (
          <Stack
            direction={{ xs: "column", sm: "row" }}
            justifyContent="space-between"
            alignItems={{ xs: "flex-start", sm: "center" }}
            spacing={1.5}
            sx={{ mb: 1.75 }}
          >
            <Stack spacing={0.35}>
              {title ? <Typography variant="h6">{title}</Typography> : null}
              {subtitle ? (
                <Typography variant="body2" color="text.secondary">
                  {subtitle}
                </Typography>
              ) : null}
            </Stack>
            {actions}
          </Stack>
        )}
        {children}
      </CardContent>
    </Card>
  );
}
