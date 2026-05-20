import InboxRoundedIcon from "@mui/icons-material/InboxRounded";
import { Box, Button, Card, CardContent, Typography, type ReactNode } from "@mui/material";

type EmptyStateProps = {
  title: string;
  message: string;
  actionLabel?: string;
  onAction?: () => void;
  icon?: ReactNode;
};

export default function EmptyState({ title, message, actionLabel, onAction, icon }: EmptyStateProps) {
  return (
    <Card>
      <CardContent
        sx={{
          minHeight: 180,
          display: "grid",
          placeItems: "center",
          textAlign: "center",
          px: { xs: 2.5, md: 4 },
        }}
      >
        <Box>
          <Box sx={{ color: "primary.main", mb: 1 }}>{icon ?? <InboxRoundedIcon sx={{ fontSize: 32 }} />}</Box>
          <Typography variant="h6" sx={{ mb: 0.75 }}>
            {title}
          </Typography>
          <Typography color="text.secondary" sx={{ maxWidth: 540, mx: "auto" }}>
            {message}
          </Typography>
          {actionLabel && onAction && (
            <Button sx={{ mt: 2 }} variant="outlined" onClick={onAction}>
              {actionLabel}
            </Button>
          )}
        </Box>
      </CardContent>
    </Card>
  );
}
