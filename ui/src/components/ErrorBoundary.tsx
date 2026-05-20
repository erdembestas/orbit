import { Alert, Box, Button, Stack, Typography } from "@mui/material";
import { Component, type ErrorInfo, type ReactNode } from "react";

type ErrorBoundaryProps = {
  children: ReactNode;
};

type ErrorBoundaryState = {
  hasError: boolean;
  message: string;
};

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = {
    hasError: false,
    message: "",
  };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || "Unknown render error",
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("Orbit UI render error", error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        <Box
          sx={{
            minHeight: "100vh",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            bgcolor: "background.default",
            p: 3,
          }}
        >
          <Stack spacing={2} sx={{ maxWidth: 520, width: "100%" }}>
            <Alert severity="error">Something went wrong rendering this page</Alert>
            <Typography color="text.secondary">{this.state.message}</Typography>
            <Button variant="contained" onClick={() => window.location.reload()}>
              Reload
            </Button>
          </Stack>
        </Box>
      );
    }

    return this.props.children;
  }
}
