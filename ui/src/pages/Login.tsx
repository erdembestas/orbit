import LockRoundedIcon from "@mui/icons-material/LockRounded";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  InputAdornment,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { alpha } from "@mui/material/styles";
import { useState } from "react";
import { type ApiError, type OrbitApiClient } from "../api/client";

type LoginPageProps = {
  client: OrbitApiClient;
  onLogin: (token: string) => void;
};

export default function LoginPage({ client, onLogin }: LoginPageProps) {
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("admin");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      const response = await client.login(username, password);
      onLogin(response.accessToken);
    } catch (err) {
      setError((err as ApiError).message ?? "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Box
      sx={{
        minHeight: "100vh",
        display: "grid",
        placeItems: "center",
        px: 2,
        background:
          "radial-gradient(circle at top, rgba(20, 184, 166, 0.16), transparent 26%), #090D11",
      }}
    >
      <Card sx={{ width: "100%", maxWidth: 400, overflow: "hidden", bgcolor: "background.paper" }}>
        <Box
          sx={{
            px: { xs: 3, sm: 4 },
            pt: { xs: 3, sm: 3.5 },
            pb: 1.5,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <Stack spacing={1.5}>
            <Box
              sx={{
                width: 40,
                height: 40,
                borderRadius: 1,
                display: "grid",
                placeItems: "center",
                bgcolor: alpha("#14B8A6", 0.12),
                color: "primary.main",
              }}
            >
              <LockRoundedIcon />
            </Box>
            <Box>
              <Typography variant="overline" color="primary.main" sx={{ letterSpacing: "0.16em" }}>
                ORBIT
              </Typography>
              <Typography variant="h6">Control Plane</Typography>
              <Typography color="text.secondary">
                Local sign-in for the single-cluster operational console.
              </Typography>
            </Box>
          </Stack>
        </Box>

        <CardContent sx={{ p: { xs: 3, sm: 3.5 } }}>
          <Stack component="form" spacing={2} onSubmit={handleSubmit}>
            {error && <Alert severity="error">{error}</Alert>}
            <TextField
              label="Username"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              fullWidth
            />
            <TextField
              label="Password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              type="password"
              autoComplete="current-password"
              fullWidth
            />
            <TextField
              label="Local default"
              value="admin / admin"
              fullWidth
              InputProps={{
                readOnly: true,
                startAdornment: <InputAdornment position="start">Use</InputAdornment>,
              }}
              helperText="This local development account is intended only for Minikube installs."
            />
            <Button type="submit" variant="contained" disabled={loading} fullWidth>
              {loading ? "Signing in…" : "Sign in to Orbit"}
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Box>
  );
}
