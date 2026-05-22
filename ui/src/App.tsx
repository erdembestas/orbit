import AssignmentRoundedIcon from "@mui/icons-material/AssignmentRounded";
import DashboardRoundedIcon from "@mui/icons-material/DashboardRounded";
import DescriptionRoundedIcon from "@mui/icons-material/DescriptionRounded";
import FactCheckRoundedIcon from "@mui/icons-material/FactCheckRounded";
import MonitorHeartRoundedIcon from "@mui/icons-material/MonitorHeartRounded";
import PolicyRoundedIcon from "@mui/icons-material/PolicyRounded";
import SearchRoundedIcon from "@mui/icons-material/SearchRounded";
import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import { Alert, Snackbar } from "@mui/material";
import { useEffect, useMemo, useState } from "react";
import {
  OrbitApiClient,
  type ApiError,
  type MeResponse,
} from "./api/client";
import AppShell, { type NavItem } from "./components/AppShell";
import ActionPlansPage from "./pages/ActionPlans";
import AnalyzePage from "./pages/Analyze";
import ClusterHealthPage from "./pages/ClusterHealth";
import ControllerStatusPage from "./pages/ControllerStatus";
import DashboardPage from "./pages/Dashboard";
import EvidencePacksPage from "./pages/EvidencePacks";
import FindingRulesPage from "./pages/FindingRules";
import FindingsPage from "./pages/Findings";
import InventoryPage from "./pages/Inventory";
import LoginPage from "./pages/Login";

const tokenStorageKey = "orbit-token";

const navItems: NavItem[] = [
  { key: "dashboard", label: "Dashboard", icon: <DashboardRoundedIcon /> },
  { key: "cluster-health", label: "Cluster Health", icon: <MonitorHeartRoundedIcon /> },
  { key: "analyze", label: "Analyze", icon: <SearchRoundedIcon /> },
  { key: "inventory", label: "Inventory", icon: <StorageRoundedIcon /> },
  { key: "findings", label: "Findings", icon: <FactCheckRoundedIcon /> },
  { key: "evidence", label: "Evidence Packs", icon: <DescriptionRoundedIcon /> },
  { key: "plans", label: "Action Plans", icon: <AssignmentRoundedIcon /> },
  { key: "rules", label: "Finding Rules", icon: <PolicyRoundedIcon /> },
  { key: "controller", label: "Controller Status", icon: <MonitorHeartRoundedIcon /> },
];

export default function App() {
  const [token, setToken] = useState<string>(() => localStorage.getItem(tokenStorageKey) ?? "");
  const [page, setPage] = useState("dashboard");
  const [me, setMe] = useState<MeResponse | null>(null);
  const [error, setError] = useState("");

  function clearSession() {
    localStorage.removeItem(tokenStorageKey);
    setToken("");
    setMe(null);
    setError("");
  }

  const client = useMemo(
    () =>
      new OrbitApiClient({
        getToken: () => localStorage.getItem(tokenStorageKey),
        onUnauthorized: clearSession,
      }),
    [],
  );

  useEffect(() => {
    if (!token) {
      return;
    }

    let cancelled = false;
    client
      .getMe()
      .then((response) => {
        if (!cancelled && response) {
          setMe(response);
          setError("");
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError((err as ApiError).message ?? "Failed to load profile");
        }
      });

    return () => {
      cancelled = true;
    };
  }, [client, token]);

  function handleLogin(nextToken: string) {
    localStorage.setItem(tokenStorageKey, nextToken);
    setToken(nextToken);
    setPage("dashboard");
  }

  if (!token) {
    return <LoginPage client={client} onLogin={handleLogin} />;
  }

  return (
    <>
      <AppShell
        title={navItems.find((item) => item.key === page)?.label ?? "Orbit"}
        username={me?.username ?? "Loading"}
        currentPage={page}
        navItems={navItems}
        onNavigate={setPage}
        onLogout={clearSession}
      >
        {page === "dashboard" && <DashboardPage client={client} me={me} />}
        {page === "cluster-health" && <ClusterHealthPage client={client} />}
        {page === "analyze" && <AnalyzePage client={client} />}
        {page === "inventory" && <InventoryPage client={client} />}
        {page === "findings" && <FindingsPage client={client} />}
        {page === "evidence" && <EvidencePacksPage client={client} />}
        {page === "plans" && <ActionPlansPage client={client} />}
        {page === "rules" && <FindingRulesPage client={client} />}
        {page === "controller" && <ControllerStatusPage client={client} />}
      </AppShell>

      <Snackbar open={Boolean(error)} autoHideDuration={5000} onClose={() => setError("")}>
        <Alert severity="error" onClose={() => setError("")} sx={{ width: "100%" }}>
          {error}
        </Alert>
      </Snackbar>
    </>
  );
}
