import ErrorOutlineRoundedIcon from "@mui/icons-material/ErrorOutlineRounded";
import RefreshRoundedIcon from "@mui/icons-material/RefreshRounded";
import { Alert, AlertTitle, Button, Card, CardContent, Stack } from "@mui/material";

type ErrorStateProps = {
  title?: string;
  message: string;
  onRetry?: () => void;
};

export default function ErrorState({ title = "Request failed", message, onRetry }: ErrorStateProps) {
  return (
    <Card>
      <CardContent sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Alert
            severity="error"
            icon={<ErrorOutlineRoundedIcon fontSize="inherit" />}
            action={
              onRetry ? (
                <Button color="inherit" size="small" startIcon={<RefreshRoundedIcon />} onClick={onRetry}>
                  Retry
                </Button>
              ) : undefined
            }
          >
            <AlertTitle>{title}</AlertTitle>
            {message}
          </Alert>
        </Stack>
      </CardContent>
    </Card>
  );
}
